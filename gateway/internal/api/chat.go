package api

import (
	"crypto/rand"
	"encoding/hex"
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
	RequestID       string               `json:"request_id"`
	Decision        string               `json:"decision"`
	Message         string               `json:"message"`
	SanitizedPrompt string               `json:"sanitized_prompt,omitempty"`
	PIIFindings     []scanner.PIIFinding `json:"pii_findings,omitempty"`
}

func newRequestID() string {
	bytes := make([]byte, 8)

	if _, err := rand.Read(bytes); err != nil {
		return "req_unknown"
	}

	return "req_" + hex.EncodeToString(bytes)
}

func ChatHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	requestID := newRequestID()

	var req ChatRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error":      "invalid request body",
			"request_id": requestID,
		})
		return
	}

	if len(req.Messages) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error":      "messages cannot be empty",
			"request_id": requestID,
		})
		return
	}

	var content strings.Builder

	for _, message := range req.Messages {
		content.WriteString(message.Content)
		content.WriteString("\n")
	}

	rawPrompt := content.String()

	secretFindings := scanner.ScanSecrets(rawPrompt)

	if len(secretFindings) > 0 {
		w.WriteHeader(http.StatusForbidden)

		_ = json.NewEncoder(w).Encode(BlockResponse{
			RequestID: requestID,
			Decision:  "BLOCK",
			RiskScore: 95,
			Findings:  secretFindings,
		})

		return
	}

	sanitizedPrompt, piiFindings := scanner.ScanAndRedactPII(rawPrompt)

	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(AllowedResponse{
		RequestID:       requestID,
		Decision:        "ALLOW",
		Message:         "request passed SentryMesh security checks",
		SanitizedPrompt: strings.TrimSpace(sanitizedPrompt),
		PIIFindings:     piiFindings,
	})
}
