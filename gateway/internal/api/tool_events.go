package api

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func ToolEventsHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	id, err := strconv.ParseInt(
		r.PathValue("id"),
		10,
		64,
	)
	if err != nil {
		http.Error(
			w,
			`{"error":"invalid approval id"}`,
			http.StatusBadRequest,
		)
		return
	}

	events, err := auditStore.ListToolEvents(
		r.Context(),
		id,
	)
	if err != nil {
		http.Error(
			w,
			`{"error":"failed to query tool events"}`,
			http.StatusInternalServerError,
		)
		return
	}

	_ = json.NewEncoder(w).Encode(events)
}
