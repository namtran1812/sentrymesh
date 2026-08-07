package api

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func AuthEventsHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	limit := 50

	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			limit = parsed
		}
	}

	events, err := auditStore.ListAuthEvents(
		r.Context(),
		limit,
	)
	if err != nil {
		http.Error(
			w,
			`{"error":"failed to list auth events"}`,
			http.StatusInternalServerError,
		)
		return
	}

	_ = json.NewEncoder(w).Encode(events)
}
