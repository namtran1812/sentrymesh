package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/namtran1812/sentrymesh/gateway/internal/risk"
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

type SecurityResponse struct {
	RequestID         string                     `json:"request_id"`
	Decision          string                     `json:"decision"`
	RiskScore         int                        `json:"risk_score"`
	Severity          string                     `json:"severity"`
	Message           string                     `json:"message"`
	SanitizedPrompt   string                     `json:"sanitized_prompt,omitempty"`
	SecretFindings    []scanner.Finding          `json:"secret_findings,omitempty"`
	PIIFindings       []scanner.PIIFinding       `json:"pii_findings,omitempty"`
	InjectionFindings []scanner.InjectionFinding `json:"injection_findings,omitempty"`
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
		writeError(w, http.StatusBadRequest, requestID, "invalid request body")
		return
	}

	if len(req.Messages) == 0 {
		writeError(w, http.StatusBadRequest, requestID, "messages cannot be empty")
		return
	}

	var content strings.Builder

	for _, message := range req.Messages {
		content.WriteString(message.Content)
		content.WriteString("\n")
	}

	rawPrompt := strings.TrimSpace(content.String())

	secretFindings := scanner.ScanSecrets(rawPrompt)
	injectionFindings := scanner.ScanPromptInjection(rawPrompt)
	sanitizedPrompt, piiFindings := scanner.ScanAndRedactPII(rawPrompt)

	maxInjectionScore := 0
	for _, finding := range injectionFindings {
		if finding.Confidence > maxInjectionScore {
			maxInjectionScore = finding.Confidence
		}
	}

	riskDecision := risk.Evaluate(risk.Input{
		SecretCount:    len(secretFindings),
		PIICount:       len(piiFindings),
		InjectionCount: len(injectionFindings),
		MaxInjection:   maxInjectionScore,
	})

	status := http.StatusOK
	message := "request passed SentryMesh security checks"

	switch riskDecision.Action {
	case "BLOCK":
		status = http.StatusForbidden
		message = "request blocked by SentryMesh security policy"
	case "REVIEW":
		status = http.StatusAccepted
		message = "request requires security review"
	case "ALLOW_WITH_REDACTION":
		message = "request allowed after sensitive data redaction"
	}

	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(SecurityResponse{
		RequestID:         requestID,
		Decision:          riskDecision.Action,
		RiskScore:         riskDecision.Score,
		Severity:          riskDecision.Severity,
		Message:           message,
		SanitizedPrompt:   sanitizedPrompt,
		SecretFindings:    secretFindings,
		PIIFindings:       piiFindings,
		InjectionFindings: injectionFindings,
	})
}

func writeError(w http.ResponseWriter, status int, requestID string, message string) {
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(map[string]string{
		"error":      message,
		"request_id": requestID,
	})
}
