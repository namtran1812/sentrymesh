package api

import (
	"encoding/json"
	"net/http"

	"github.com/namtran1812/sentrymesh/gateway/internal/middleware"
	"github.com/namtran1812/sentrymesh/gateway/internal/rag"
)

type RAGContextRequest struct {
	RequestID string         `json:"request_id"`
	Documents []rag.Document `json:"documents"`
}

func RAGContextHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	principal, ok := middleware.IdentityFromContext(r.Context())
	if !ok {
		http.Error(
			w,
			`{"error":"authenticated identity unavailable"}`,
			http.StatusUnauthorized,
		)
		return
	}

	var req RAGContextRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(
			w,
			`{"error":"invalid request body"}`,
			http.StatusBadRequest,
		)
		return
	}

	if req.RequestID == "" {
		req.RequestID = newRequestID()
	}

	result := rag.BuildContext(
		req.RequestID,
		principal,
		req.Documents,
	)

	_ = json.NewEncoder(w).Encode(result)
}
