package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/namtran1812/sentrymesh/gateway/internal/audit"
	"github.com/namtran1812/sentrymesh/gateway/internal/provider"
	"github.com/namtran1812/sentrymesh/gateway/internal/risk"
	"github.com/namtran1812/sentrymesh/gateway/internal/scanner"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string    `json:"model"`
	Provider string    `json:"provider,omitempty"`
	Messages []Message `json:"messages"`
}

type SecurityResponse struct {
	RequestID         string                     `json:"request_id"`
	Decision          string                     `json:"decision"`
	RiskScore         int                        `json:"risk_score"`
	Severity          string                     `json:"severity"`
	Message           string                     `json:"message"`
	SanitizedPrompt   string                     `json:"sanitized_prompt,omitempty"`
	ModelResponse     string                     `json:"model_response,omitempty"`
	SecretFindings    []scanner.Finding          `json:"secret_findings,omitempty"`
	PIIFindings       []scanner.PIIFinding       `json:"pii_findings,omitempty"`
	InjectionFindings []scanner.InjectionFinding `json:"injection_findings,omitempty"`
	OutputFindings    *scanner.OutputScan        `json:"output_scan,omitempty"`
}

var providerRouter = provider.NewDefaultRouter()

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
	started := time.Now()

	var req ChatRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, requestID, "invalid request body")
		return
	}

	if len(req.Messages) == 0 {
		writeError(w, http.StatusBadRequest, requestID, "messages cannot be empty")
		return
	}

	if req.Provider == "" {
		req.Provider = "mock"
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

	if riskDecision.Action == "BLOCK" {
		w.WriteHeader(http.StatusForbidden)

		_ = json.NewEncoder(w).Encode(SecurityResponse{
			RequestID:         requestID,
			Decision:          riskDecision.Action,
			RiskScore:         riskDecision.Score,
			Severity:          riskDecision.Severity,
			Message:           "request blocked by SentryMesh security policy",
			SanitizedPrompt:   sanitizedPrompt,
			SecretFindings:    secretFindings,
			PIIFindings:       piiFindings,
			InjectionFindings: injectionFindings,
		})

		return
	}

	providerMessages := make([]provider.Message, 0, len(req.Messages))

	for _, message := range req.Messages {
		sanitizedContent, _ := scanner.ScanAndRedactPII(message.Content)

		providerMessages = append(
			providerMessages,
			provider.Message{
				Role:    message.Role,
				Content: sanitizedContent,
			},
		)
	}

	modelResponse, err := providerRouter.Chat(
		r.Context(),
		req.Provider,
		provider.Request{
			Model:    req.Model,
			Messages: providerMessages,
		},
	)

	if err != nil {
		writeError(
			w,
			http.StatusBadGateway,
			requestID,
			"model provider request failed",
		)
		return
	}

	outputScan := scanner.ScanOutput(modelResponse.Content)

	if !outputScan.Safe {
		w.WriteHeader(http.StatusForbidden)

		_ = json.NewEncoder(w).Encode(SecurityResponse{
			RequestID:         requestID,
			Decision:          "BLOCK",
			RiskScore:         100,
			Severity:          "CRITICAL",
			Message:           "model output blocked by SentryMesh security policy",
			SanitizedPrompt:   sanitizedPrompt,
			SecretFindings:    secretFindings,
			PIIFindings:       piiFindings,
			InjectionFindings: injectionFindings,
			OutputFindings:    &outputScan,
		})

		return
	}

	finalOutput := outputScan.Redacted

	message := "request passed SentryMesh security checks"

	if riskDecision.Action == "ALLOW_WITH_REDACTION" {
		message = "request allowed after sensitive data redaction"
	}

	if len(outputScan.PIIFindings) > 0 {
		message = "request allowed after output redaction"
	}

	latency := time.Since(started).Milliseconds()

	_ = auditStore.Write(
		r.Context(),
		audit.Event{
			RequestID:         requestID,
			Timestamp:         time.Now(),
			Provider:          req.Provider,
			Model:             req.Model,
			Decision:          riskDecision.Action,
			RiskScore:         riskDecision.Score,
			Severity:          riskDecision.Severity,
			LatencyMS:         latency,
			SecretFindings:    secretFindings,
			PIIFindings:       piiFindings,
			InjectionFindings: injectionFindings,
			OutputFindings:    outputScan,
		},
	)

	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(SecurityResponse{
		RequestID:         requestID,
		Decision:          riskDecision.Action,
		RiskScore:         riskDecision.Score,
		Severity:          riskDecision.Severity,
		Message:           message,
		SanitizedPrompt:   sanitizedPrompt,
		ModelResponse:     finalOutput,
		SecretFindings:    secretFindings,
		PIIFindings:       piiFindings,
		InjectionFindings: injectionFindings,
		OutputFindings:    &outputScan,
	})
}

func writeError(
	w http.ResponseWriter,
	status int,
	requestID string,
	message string,
) {
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(map[string]string{
		"error":      message,
		"request_id": requestID,
	})
}
