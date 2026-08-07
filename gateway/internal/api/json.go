package api

import (
	"encoding/json"
	"errors"
	"net/http"
)

func decodeJSON(
	w http.ResponseWriter,
	r *http.Request,
	dst any,
) error {
	err := json.NewDecoder(r.Body).Decode(dst)

	if err == nil {
		return nil
	}

	var maxBytesErr *http.MaxBytesError

	if errors.As(err, &maxBytesErr) {
		w.Header().Set(
			"Content-Type",
			"application/json",
		)

		w.WriteHeader(
			http.StatusRequestEntityTooLarge,
		)

		_ = json.NewEncoder(w).Encode(
			map[string]any{
				"error":       "request body too large",
				"limit_bytes": maxBytesErr.Limit,
			},
		)

		return err
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(
		http.StatusBadRequest,
	)

	_ = json.NewEncoder(w).Encode(
		map[string]string{
			"error": "invalid request body",
		},
	)

	return err
}
