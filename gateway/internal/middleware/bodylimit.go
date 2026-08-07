package middleware

import (
	"encoding/json"
	"net/http"
)

func BodyLimit(
	maxBytes int64,
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		r.Body = http.MaxBytesReader(
			w,
			r.Body,
			maxBytes,
		)

		next.ServeHTTP(w, r)
	})
}

func WritePayloadTooLarge(
	w http.ResponseWriter,
) {
	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(
		http.StatusRequestEntityTooLarge,
	)

	_ = json.NewEncoder(w).Encode(
		map[string]string{
			"error": "request body too large",
		},
	)
}
