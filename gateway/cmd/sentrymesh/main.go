package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/namtran1812/sentrymesh/gateway/internal/api"
	"github.com/namtran1812/sentrymesh/gateway/internal/middleware"
)

type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Version string `json:"version"`
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	_ = json.NewEncoder(w).Encode(HealthResponse{
		Status:  "ok",
		Service: "sentrymesh-gateway",
		Version: "0.1.0",
	})
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("POST /v1/chat/completions", api.ChatHandler)
	mux.HandleFunc("GET /v1/audit/events", api.AuditEventsHandler)
	mux.HandleFunc("GET /v1/audit/stats", api.AuditStatsHandler)
	mux.Handle("POST /v1/tools/evaluate", middleware.Auth(http.HandlerFunc(api.ToolEvaluationHandler)))
	mux.HandleFunc("GET /v1/approvals", api.ListApprovalsHandler)
	mux.Handle("POST /v1/approvals/{id}/approve", middleware.Auth(http.HandlerFunc(api.ApproveHandler)))
	mux.Handle("POST /v1/approvals/{id}/reject", middleware.Auth(http.HandlerFunc(api.RejectHandler)))
	mux.Handle("POST /v1/approvals/{id}/execute", middleware.Auth(http.HandlerFunc(api.ExecuteApprovalHandler)))
	mux.HandleFunc("GET /v1/approvals/{id}/events", api.ToolEventsHandler)

	server := &http.Server{
		Addr:    ":8080",
		Handler: middleware.CORS(mux),
	}

	log.Println("SentryMesh Gateway listening on http://localhost:8080")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
