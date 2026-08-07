package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/namtran1812/sentrymesh/gateway/internal/audit"
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

	err := auditStore.WriteRAGEvent(
		r.Context(),
		audit.RAGEvent{
			RequestID: req.RequestID,
			Timestamp: time.Now(),
			UserID:    principal.UserID,
			Role:      string(principal.Role),
			Team:      principal.Team,
			Trace:     result.Trace,
		},
	)

	if err != nil {
		log.Printf(
			"failed to write RAG provenance request=%s: %v",
			req.RequestID,
			err,
		)
	}

	_ = json.NewEncoder(w).Encode(result)
}
