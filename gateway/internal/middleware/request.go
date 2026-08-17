package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"os"

	"github.com/namtran1812/sentrymesh/gateway/internal/metrics"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"time"
)

type requestIDKey struct{}

const requestIDHeader = "X-Request-ID"

func newRequestID() string {
	buf := make([]byte, 8)

	if _, err := rand.Read(buf); err != nil {
		return "req_unknown"
	}

	return "req_" + hex.EncodeToString(buf)
}

func RequestIDFromContext(ctx context.Context) string {
	value, _ := ctx.Value(requestIDKey{}).(string)
	return value
}

func RequestContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		requestID := r.Header.Get(requestIDHeader)

		if requestID == "" {
			requestID = newRequestID()
		}

		ctx := context.WithValue(
			r.Context(),
			requestIDKey{},
			requestID,
		)

		w.Header().Set(
			requestIDHeader,
			requestID,
		)

		next.ServeHTTP(
			w,
			r.WithContext(ctx),
		)
	})
}

type accessRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (r *accessRecorder) WriteHeader(status int) {
	if r.status == 0 {
		r.status = status
	}

	r.ResponseWriter.WriteHeader(status)
}

func (r *accessRecorder) Write(data []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}

	n, err := r.ResponseWriter.Write(data)
	r.bytes += n

	return n, err
}

func AccessLog(next http.Handler) http.Handler {
	logger := slog.New(
		slog.NewJSONHandler(
			os.Stdout,
			&slog.HandlerOptions{},
		),
	)

	return http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		started := time.Now()

		recorder := &accessRecorder{
			ResponseWriter: w,
		}

		next.ServeHTTP(recorder, r)

		duration := time.Since(started)

		switch r.URL.Path {
		case "/health", "/ready", "/metrics":
		default:
			metrics.ObserveRequest(duration)
		}

		status := recorder.status
		if status == 0 {
			status = http.StatusOK
		}

		attrs := []any{
			"event", "http_request",
			"request_id", RequestIDFromContext(r.Context()),
			"method", r.Method,
			"path", r.URL.Path,
			"status", status,
			"bytes", recorder.bytes,
			"latency_us", duration.Microseconds(),
			"remote_addr", r.RemoteAddr,
		}

		spanContext :=
			trace.SpanFromContext(
				r.Context(),
			).SpanContext()

		if spanContext.IsValid() {
			attrs = append(
				attrs,
				"trace_id",
				spanContext.TraceID().
					String(),
				"span_id",
				spanContext.SpanID().
					String(),
			)
		}

		if principal, ok := IdentityFromContext(r.Context()); ok {
			attrs = append(
				attrs,
				"key_id", principal.KeyID,
				"user_id", principal.UserID,
				"role", string(principal.Role),
				"team", principal.Team,
			)
		}

		logger.Info(
			"http request",
			attrs...,
		)
	})
}

func TraceRequest(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(
		func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			span :=
				trace.SpanFromContext(
					r.Context(),
				)

			if span.IsRecording() {
				span.SetAttributes(
					attribute.String(
						"sentrymesh.request_id",
						RequestIDFromContext(
							r.Context(),
						),
					),
				)
			}

			next.ServeHTTP(
				w,
				r,
			)
		},
	)
}
