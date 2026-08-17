package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/namtran1812/sentrymesh/gateway/internal/abuse"
	"github.com/namtran1812/sentrymesh/gateway/internal/api"
	"github.com/namtran1812/sentrymesh/gateway/internal/metrics"
	"github.com/namtran1812/sentrymesh/gateway/internal/middleware"
	"github.com/namtran1812/sentrymesh/gateway/internal/ratelimit"
	"github.com/namtran1812/sentrymesh/gateway/internal/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Version string `json:"version"`
}

type ReadinessResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Backend string `json:"backend"`
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	_ = json.NewEncoder(w).Encode(HealthResponse{
		Status:  "ok",
		Service: "sentrymesh-gateway",
		Version: "0.1.0",
	})
}

func readinessHandler(
	deps *Dependencies,
) http.HandlerFunc {
	return func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		w.Header().Set(
			"Content-Type",
			"application/json",
		)

		ctx, cancel := context.WithTimeout(
			r.Context(),
			2*time.Second,
		)
		defer cancel()

		if err := deps.Ready(ctx); err != nil {
			w.WriteHeader(
				http.StatusServiceUnavailable,
			)

			_ = json.NewEncoder(w).Encode(
				ReadinessResponse{
					Status:  "not_ready",
					Service: "sentrymesh-gateway",
					Backend: deps.Backend,
				},
			)

			return
		}

		_ = json.NewEncoder(w).Encode(
			ReadinessResponse{
				Status:  "ready",
				Service: "sentrymesh-gateway",
				Backend: deps.Backend,
			},
		)
	}
}

func main() {
	traceCtx := context.Background()

	tracing, err := telemetry.NewTracing(traceCtx)
	if err != nil {
		log.Fatalf(
			"initialize tracing: %v",
			err,
		)
	}

	defer func() {
		ctx, cancel := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cancel()

		if err := tracing.Shutdown(ctx); err != nil {
			log.Printf(
				"shutdown tracing: %v",
				err,
			)
		}
	}()

	startupCtx, startupCancel :=
		context.WithTimeout(
			context.Background(),
			15*time.Second,
		)

	deps, err :=
		NewDependencies(startupCtx)

	startupCancel()

	if err != nil {
		log.Fatalf(
			"initialize dependencies: %v",
			err,
		)
	}

	defer deps.Close()

	asyncAudit, err :=
		configureAuditMode(deps)
	if err != nil {
		log.Fatalf(
			"configure audit persistence: %v",
			err,
		)
	}

	if asyncAudit != nil {
		defer func() {
			ctx, cancel :=
				context.WithTimeout(
					context.Background(),
					10*time.Second,
				)
			defer cancel()

			if err := asyncAudit.Close(
				ctx,
			); err != nil {
				log.Printf(
					"drain audit writer: %v",
					err,
				)
			}
		}()
	}

	installDependencies(deps)

	seedAPIKeys()

	rateLimitCapacity := 20
	rateLimitRefill := 5.0

	if os.Getenv("SENTRYMESH_BENCHMARK_MODE") == "1" {
		rateLimitCapacity = 1_000_000
		rateLimitRefill = 1_000_000.0
		log.Println("benchmark mode enabled: API rate limits relaxed")
	}

	apiLimiter := ratelimit.New(
		rateLimitCapacity,
		rateLimitRefill,
	)

	abuseTracker := abuse.NewPersistent(
		5,
		30*time.Second,
		30*time.Second,
		deps.Abuse,
	)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", healthHandler)
	mux.Handle(
		"GET /metrics",
		middleware.Auth(
			middleware.RequireScope(
				"audit:read",
				metrics.Handler(),
			),
		),
	)
	mux.HandleFunc(
		"GET /ready",
		readinessHandler(deps),
	)
	mux.Handle("POST /v1/chat/completions", middleware.BodyLimit(1<<20, middleware.Auth(middleware.TrafficGuard(abuseTracker, apiLimiter, deps.Audit, http.HandlerFunc(api.ChatHandler)))))
	mux.Handle("GET /v1/audit/events", middleware.Auth(middleware.RequireScope("audit:read", http.HandlerFunc(api.AuditEventsHandler))))
	mux.Handle("GET /v1/audit/stats", middleware.Auth(middleware.RequireScope("audit:read", http.HandlerFunc(api.AuditStatsHandler))))
	mux.Handle("POST /v1/tools/evaluate", middleware.BodyLimit(1<<20, middleware.Auth(middleware.TrafficGuard(abuseTracker, apiLimiter, deps.Audit, middleware.RequireScope("tools:evaluate", http.HandlerFunc(api.ToolEvaluationHandler))))))
	mux.Handle("GET /v1/approvals", middleware.Auth(middleware.RequireScope("audit:read", http.HandlerFunc(api.ListApprovalsHandler))))
	mux.Handle("POST /v1/approvals/{id}/approve", middleware.Auth(middleware.RequireScope("approvals:write", http.HandlerFunc(api.ApproveHandler))))
	mux.Handle("POST /v1/approvals/{id}/reject", middleware.Auth(middleware.RequireScope("approvals:write", http.HandlerFunc(api.RejectHandler))))
	mux.Handle("POST /v1/approvals/{id}/execute", middleware.Auth(middleware.RequireScope("tools:execute", http.HandlerFunc(api.ExecuteApprovalHandler))))
	mux.Handle("GET /v1/keys", middleware.Auth(middleware.RequireScope("keys:manage", http.HandlerFunc(api.ListKeysHandler))))
	mux.Handle("GET /v1/audit/auth-events", middleware.Auth(middleware.RequireScope("audit:read", http.HandlerFunc(api.AuthEventsHandler))))
	mux.Handle("GET /v1/audit/abuse-events", middleware.Auth(middleware.RequireScope("audit:read", http.HandlerFunc(api.AbuseEventsHandler))))
	mux.Handle("GET /v1/security/posture", middleware.Auth(middleware.RequireScope("audit:read", http.HandlerFunc(api.SecurityPostureHandler))))
	mux.Handle("GET /v1/evals/latest", middleware.Auth(middleware.RequireScope("evals:read", http.HandlerFunc(api.EvalResultsHandler))))
	mux.Handle("POST /v1/rag/inspect", middleware.Auth(middleware.RequireScope("rag:inspect", http.HandlerFunc(api.RAGInspectHandler))))
	mux.Handle("POST /v1/rag/context", middleware.BodyLimit(1<<20, middleware.Auth(middleware.RequireScope("rag:context", http.HandlerFunc(api.RAGContextHandler)))))
	mux.Handle("POST /v1/rag/chat", middleware.BodyLimit(1<<20, middleware.Auth(middleware.TrafficGuard(abuseTracker, apiLimiter, deps.Audit, middleware.RequireScope("rag:chat", http.HandlerFunc(api.RAGChatHandler))))))
	mux.Handle("GET /v1/rag/requests/{request_id}/provenance", middleware.Auth(middleware.RequireScope("audit:read", http.HandlerFunc(api.RAGEventsHandler))))
	mux.Handle("POST /v1/keys", middleware.BodyLimit(1<<20, middleware.Auth(middleware.RequireScope("keys:manage", http.HandlerFunc(api.CreateKeyHandler)))))
	mux.Handle("POST /v1/keys/{id}/revoke", middleware.Auth(middleware.RequireScope("keys:manage", http.HandlerFunc(api.RevokeKeyHandler))))
	mux.Handle("GET /v1/approvals/{id}/events", middleware.Auth(middleware.RequireScope("audit:read", http.HandlerFunc(api.ToolEventsHandler))))

	handler := http.Handler(
		middleware.CORS(mux),
	)

	disableAccessLog :=
		os.Getenv(
			"SENTRYMESH_DISABLE_ACCESS_LOG",
		) == "1"

	if disableAccessLog &&
		os.Getenv(
			"SENTRYMESH_BENCHMARK_MODE",
		) != "1" {
		log.Fatal(
			"SENTRYMESH_DISABLE_ACCESS_LOG requires benchmark mode",
		)
	}

	if !disableAccessLog {
		handler = middleware.AccessLog(
			handler,
		)
	} else {
		log.Println(
			"benchmark mode: access logging disabled",
		)
	}

	handler = middleware.TraceRequest(
		handler,
	)

	handler = otelhttp.NewHandler(
		handler,
		"sentrymesh.http",
		otelhttp.WithSpanNameFormatter(
			func(
				_ string,
				r *http.Request,
			) string {
				return r.Method +
					" " +
					r.URL.Path
			},
		),
	)

	handler = middleware.RequestContext(
		handler,
	)

	server := &http.Server{
		Addr:              ":8080",
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)

	go func() {
		log.Println("SentryMesh Gateway listening on http://localhost:8080")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	signalCh := make(chan os.Signal, 1)

	signal.Notify(
		signalCh,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	select {
	case err := <-errCh:
		log.Fatal(err)

	case sig := <-signalCh:
		log.Printf("received signal %s; shutting down", sig)
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}

	log.Println("SentryMesh Gateway stopped")
}
