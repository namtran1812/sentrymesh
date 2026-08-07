package middleware

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/namtran1812/sentrymesh/gateway/internal/abuse"
	"github.com/namtran1812/sentrymesh/gateway/internal/audit"
	"github.com/namtran1812/sentrymesh/gateway/internal/ratelimit"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(
	status int,
) {
	if r.status == 0 {
		r.status = status
	}

	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(
	data []byte,
) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}

	return r.ResponseWriter.Write(data)
}

func TrafficGuard(
	tracker *abuse.Tracker,
	limiter *ratelimit.Limiter,
	auditStore *audit.Store,
	next http.Handler,
) http.Handler {
	return AbuseGuard(
		tracker,
		auditStore,
		RateLimit(
			limiter,
			auditStore,
			next,
		),
	)
}

func AbuseGuard(
	tracker *abuse.Tracker,
	auditStore *audit.Store,
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		principal, ok :=
			IdentityFromContext(r.Context())

		if !ok {
			w.Header().Set(
				"Content-Type",
				"application/json",
			)

			w.WriteHeader(
				http.StatusUnauthorized,
			)

			_ = json.NewEncoder(w).Encode(
				map[string]string{
					"error": "authenticated identity unavailable",
				},
			)

			return
		}

		key := fmt.Sprintf(
			"api_key:%d",
			principal.KeyID,
		)

		tracker.RegisterKey(
			r.Context(),
			key,
			principal.KeyID,
		)

		blocked, retryAfter, score :=
			tracker.Check(key)

		if blocked {
			seconds := retrySeconds(retryAfter)

			writeAbuseEvent(
				r,
				auditStore,
				principal.KeyID,
				principal.KeyName,
				principal.UserID,
				string(principal.Role),
				principal.Team,
				"ABUSE_COOLDOWN_BLOCKED",
				map[string]any{
					"score":               score,
					"retry_after_seconds": seconds,
				},
			)

			w.Header().Set(
				"Content-Type",
				"application/json",
			)

			w.Header().Set(
				"Retry-After",
				strconv.Itoa(seconds),
			)

			w.WriteHeader(
				http.StatusTooManyRequests,
			)

			_ = json.NewEncoder(w).Encode(
				map[string]any{
					"error":               "API key temporarily blocked for abusive traffic",
					"abuse_score":         score,
					"retry_after_seconds": seconds,
				},
			)

			return
		}

		recorder := &statusRecorder{
			ResponseWriter: w,
		}

		next.ServeHTTP(
			recorder,
			r,
		)

		points, reason := abusePoints(
			r.URL.Path,
			recorder.status,
		)

		if points == 0 {
			return
		}

		score, cooldownUntil, entered :=
			tracker.Add(
				key,
				points,
			)

		writeAbuseEvent(
			r,
			auditStore,
			principal.KeyID,
			principal.KeyName,
			principal.UserID,
			string(principal.Role),
			principal.Team,
			"ABUSE_SCORE_UPDATED",
			map[string]any{
				"score":  score,
				"points": points,
				"reason": reason,
				"status": recorder.status,
			},
		)

		if entered {
			writeAbuseEvent(
				r,
				auditStore,
				principal.KeyID,
				principal.KeyName,
				principal.UserID,
				string(principal.Role),
				principal.Team,
				"ABUSE_COOLDOWN_STARTED",
				map[string]any{
					"score": score,
					"cooldown_until": cooldownUntil.UTC().Format(
						time.RFC3339Nano,
					),
					"trigger": reason,
				},
			)
		}
	})
}

func abusePoints(
	path string,
	status int,
) (int, string) {
	if status == http.StatusTooManyRequests {
		return 1, "rate_limit_violation"
	}

	if status != http.StatusForbidden {
		return 0, ""
	}

	switch path {
	case "/v1/chat/completions":
		return 3, "chat_security_block"

	case "/v1/rag/chat":
		return 3, "rag_security_block"
	}

	return 0, ""
}

func retrySeconds(
	retryAfter time.Duration,
) int {
	seconds := int(
		retryAfter.Round(
			time.Second,
		) / time.Second,
	)

	if seconds < 1 {
		seconds = 1
	}

	return seconds
}

func writeAbuseEvent(
	r *http.Request,
	store *audit.Store,
	keyID int64,
	keyName string,
	userID string,
	role string,
	team string,
	eventType string,
	details any,
) {
	if store == nil {
		return
	}

	err := store.WriteAbuseEvent(
		r.Context(),
		audit.AbuseEvent{
			Timestamp: time.Now(),
			EventType: eventType,
			KeyID:     keyID,
			KeyName:   keyName,
			UserID:    userID,
			Role:      role,
			Team:      team,
			Path:      r.URL.Path,
			Details:   details,
		},
	)

	if err != nil {
		log.Printf(
			"failed to write abuse event: %v",
			err,
		)
	}
}
