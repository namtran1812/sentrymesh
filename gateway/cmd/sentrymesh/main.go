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

	server := &http.Server{
		Addr:    ":8080",
		Handler: middleware.CORS(mux),
	}

	log.Println("SentryMesh Gateway listening on http://localhost:8080")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
