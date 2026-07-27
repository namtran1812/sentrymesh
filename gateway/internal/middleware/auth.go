package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/namtran1812/sentrymesh/gateway/internal/auth"
	"github.com/namtran1812/sentrymesh/gateway/internal/identity"
)

type identityKey struct{}

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")

		if header == "" {
			writeAuthError(w, "missing Authorization header")
			return
		}

		const prefix = "Bearer "

		if !strings.HasPrefix(header, prefix) {
			writeAuthError(w, "invalid Authorization header")
			return
		}

		apiKey := strings.TrimSpace(strings.TrimPrefix(header, prefix))

		principal, err := auth.Resolve(r.Context(), apiKey)
		if err != nil {
			writeAuthError(w, "invalid API key")
			return
		}

		ctx := context.WithValue(
			r.Context(),
			identityKey{},
			principal,
		)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func IdentityFromContext(
	ctx context.Context,
) (identity.Identity, bool) {
	value := ctx.Value(identityKey{})

	principal, ok := value.(identity.Identity)

	return principal, ok
}

func writeAuthError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)

	_ = json.NewEncoder(w).Encode(
		map[string]string{
			"error": message,
		},
	)
}
