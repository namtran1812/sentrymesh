package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/namtran1812/sentrymesh/gateway/internal/abuse"
	"github.com/namtran1812/sentrymesh/gateway/internal/api"
	"github.com/namtran1812/sentrymesh/gateway/internal/middleware"
	"github.com/namtran1812/sentrymesh/gateway/internal/ratelimit"
	"github.com/namtran1812/sentrymesh/gateway/internal/runtime"
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
	seedDevelopmentKeys()

	apiLimiter := ratelimit.New(20, 5)

	abuseTracker := abuse.NewPersistent(
		5,
		30*time.Second,
		30*time.Second,
		runtime.AbuseStore,
	)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", healthHandler)
	mux.Handle("POST /v1/chat/completions", middleware.Auth(middleware.TrafficGuard(abuseTracker, apiLimiter, runtime.AuditStore, http.HandlerFunc(api.ChatHandler))))
	mux.HandleFunc("GET /v1/audit/events", api.AuditEventsHandler)
	mux.HandleFunc("GET /v1/audit/stats", api.AuditStatsHandler)
	mux.Handle("POST /v1/tools/evaluate", middleware.Auth(middleware.TrafficGuard(abuseTracker, apiLimiter, runtime.AuditStore, middleware.RequireScope("tools:evaluate", http.HandlerFunc(api.ToolEvaluationHandler)))))
	mux.HandleFunc("GET /v1/approvals", api.ListApprovalsHandler)
	mux.Handle("POST /v1/approvals/{id}/approve", middleware.Auth(middleware.RequireScope("approvals:write", http.HandlerFunc(api.ApproveHandler))))
	mux.Handle("POST /v1/approvals/{id}/reject", middleware.Auth(middleware.RequireScope("approvals:write", http.HandlerFunc(api.RejectHandler))))
	mux.Handle("POST /v1/approvals/{id}/execute", middleware.Auth(middleware.RequireScope("tools:execute", http.HandlerFunc(api.ExecuteApprovalHandler))))
	mux.Handle("GET /v1/keys", middleware.Auth(middleware.RequireScope("keys:manage", http.HandlerFunc(api.ListKeysHandler))))
	mux.Handle("GET /v1/audit/auth-events", middleware.Auth(middleware.RequireScope("audit:read", http.HandlerFunc(api.AuthEventsHandler))))
	mux.Handle("GET /v1/audit/abuse-events", middleware.Auth(middleware.RequireScope("audit:read", http.HandlerFunc(api.AbuseEventsHandler))))
	mux.Handle("GET /v1/security/posture", middleware.Auth(middleware.RequireScope("audit:read", http.HandlerFunc(api.SecurityPostureHandler))))
	mux.Handle("GET /v1/evals/latest", middleware.Auth(middleware.RequireScope("evals:read", http.HandlerFunc(api.EvalResultsHandler))))
	mux.Handle("POST /v1/rag/inspect", middleware.Auth(middleware.RequireScope("rag:inspect", http.HandlerFunc(api.RAGInspectHandler))))
	mux.Handle("POST /v1/rag/context", middleware.Auth(middleware.RequireScope("rag:context", http.HandlerFunc(api.RAGContextHandler))))
	mux.Handle("POST /v1/rag/chat", middleware.Auth(middleware.TrafficGuard(abuseTracker, apiLimiter, runtime.AuditStore, middleware.RequireScope("rag:chat", http.HandlerFunc(api.RAGChatHandler)))))
	mux.Handle("GET /v1/rag/requests/{request_id}/provenance", middleware.Auth(middleware.RequireScope("audit:read", http.HandlerFunc(api.RAGEventsHandler))))
	mux.Handle("POST /v1/keys", middleware.Auth(middleware.RequireScope("keys:manage", http.HandlerFunc(api.CreateKeyHandler))))
	mux.Handle("POST /v1/keys/{id}/revoke", middleware.Auth(middleware.RequireScope("keys:manage", http.HandlerFunc(api.RevokeKeyHandler))))
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
