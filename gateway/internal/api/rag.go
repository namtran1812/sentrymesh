package api

import (
	"encoding/json"
	"net/http"

	"github.com/namtran1812/sentrymesh/gateway/internal/rag"
)

type RAGInspectRequest struct {
	Documents []rag.Document `json:"documents"`
}

func RAGInspectHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	var req RAGInspectRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(
			w,
			`{"error":"invalid request body"}`,
			http.StatusBadRequest,
		)
		return
	}

	results := make([]rag.Result, 0, len(req.Documents))

	for _, document := range req.Documents {
		results = append(
			results,
			rag.Inspect(document),
		)
	}

	_ = json.NewEncoder(w).Encode(results)
}
