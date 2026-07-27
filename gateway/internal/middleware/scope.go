package middleware

import (
	"encoding/json"
	"net/http"
)

func RequireScope(
	scope string,
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		principal, ok := IdentityFromContext(r.Context())
		if !ok {
			writeScopeError(
				w,
				http.StatusUnauthorized,
				"authenticated identity unavailable",
			)
			return
		}

		for _, candidate := range principal.Scopes {
			if candidate == scope || candidate == "*" {
				next.ServeHTTP(w, r)
				return
			}
		}

		writeScopeError(
			w,
			http.StatusForbidden,
			"required scope: "+scope,
		)
	})
}

func writeScopeError(
	w http.ResponseWriter,
	status int,
	message string,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(
		map[string]string{
			"error": message,
		},
	)
}
