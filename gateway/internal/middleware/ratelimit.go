package middleware

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/namtran1812/sentrymesh/gateway/internal/audit"
	"github.com/namtran1812/sentrymesh/gateway/internal/ratelimit"
)

func RateLimit(
	limiter *ratelimit.Limiter,
	auditStore audit.Repository,
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

		allowed, retryAfter :=
			limiter.Allow(key)

		if allowed {
			next.ServeHTTP(w, r)
			return
		}

		seconds := int(
			retryAfter.Round(
				time.Second,
			) / time.Second,
		)

		if seconds < 1 {
			seconds = 1
		}

		if auditStore != nil {
			err := auditStore.WriteAbuseEvent(
				r.Context(),
				audit.AbuseEvent{
					Timestamp: time.Now(),
					EventType: "RATE_LIMIT_EXCEEDED",
					KeyID:     principal.KeyID,
					KeyName:   principal.KeyName,
					UserID:    principal.UserID,
					Role:      string(principal.Role),
					Team:      principal.Team,
					Path:      r.URL.Path,
					Details: map[string]any{
						"retry_after_seconds": seconds,
					},
				},
			)

			if err != nil {
				log.Printf(
					"failed to write rate limit event: %v",
					err,
				)
			}
		}

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
				"error":               "rate limit exceeded",
				"retry_after_seconds": seconds,
			},
		)
	})
}
