package api

import (
	"encoding/json"
	"net/http"
)

func RAGEventsHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	requestID := r.PathValue("request_id")

	if requestID == "" {
		http.Error(
			w,
			`{"error":"request id required"}`,
			http.StatusBadRequest,
		)
		return
	}

	events, err := auditStore.GetRAGEvents(
		r.Context(),
		requestID,
	)
	if err != nil {
		http.Error(
			w,
			`{"error":"failed to query RAG provenance"}`,
			http.StatusInternalServerError,
		)
		return
	}

	_ = json.NewEncoder(w).Encode(events)
}
