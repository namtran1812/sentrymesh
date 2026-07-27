package api

import (
	"encoding/json"
	"net/http"
)

func AuditStatsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	stats, err := auditStore.Stats(r.Context())
	if err != nil {
		http.Error(
			w,
			`{"error":"failed to query audit stats"}`,
			http.StatusInternalServerError,
		)
		return
	}

	_ = json.NewEncoder(w).Encode(stats)
}
