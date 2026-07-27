package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/namtran1812/sentrymesh/gateway/internal/scanner"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type BlockResponse struct {
	RequestID string            `json:"request_id"`
	Decision  string            `json:"decision"`
	RiskScore int               `json:"risk_score"`
	Findings  []scanner.Finding `json:"findings"`
}

type AllowedResponse struct {
	RequestID string `json:"request_id"`
	Decision  string `json:"decision"`
	Message   string `json:"message"`
}

func ChatHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req ChatRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	var content strings.Builder

	for _, message := range req.Messages {
		content.WriteString(message.Content)
		content.WriteString("\n")
	}

	findings := scanner.ScanSecrets(content.String())

	if len(findings) > 0 {
		w.WriteHeader(http.StatusForbidden)

		_ = json.NewEncoder(w).Encode(BlockResponse{
			RequestID: "req_dev_001",
			Decision:  "BLOCK",
			RiskScore: 95,
			Findings:  findings,
		})

		return
	}

	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(AllowedResponse{
		RequestID: "req_dev_001",
		Decision:  "ALLOW",
		Message:   "request passed SentryMesh security checks",
	})
}
