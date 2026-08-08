package middleware

import (
	"net/http"
	"os"
	"strings"
)

func CORS(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		allowedOrigin := strings.TrimSpace(
			os.Getenv("SENTRYMESH_ALLOWED_ORIGIN"),
		)

		if allowedOrigin == "" {
			allowedOrigin = "http://localhost:5173"
		}

		origin := r.Header.Get("Origin")

		if origin == allowedOrigin {
			w.Header().Set(
				"Access-Control-Allow-Origin",
				allowedOrigin,
			)

			w.Header().Set(
				"Vary",
				"Origin",
			)
		}

		w.Header().Set(
			"Access-Control-Allow-Headers",
			"Content-Type, Authorization",
		)

		w.Header().Set(
			"Access-Control-Allow-Methods",
			"GET, POST, OPTIONS",
		)

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
