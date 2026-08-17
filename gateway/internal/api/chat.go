package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/namtran1812/sentrymesh/gateway/internal/audit"
	"github.com/namtran1812/sentrymesh/gateway/internal/metrics"
	"github.com/namtran1812/sentrymesh/gateway/internal/provider"
	"github.com/namtran1812/sentrymesh/gateway/internal/risk"
	"github.com/namtran1812/sentrymesh/gateway/internal/scanner"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
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

var chatTracer = otel.Tracer(
	"github.com/namtran1812/sentrymesh/gateway/internal/api/chat",
)

func writeAuditWithTrace(
	ctx context.Context,
	event audit.Event,
) error {
	ctx, span :=
		chatTracer.Start(
			ctx,
			"audit.enqueue",
		)
	defer span.End()

	span.SetAttributes(
		attribute.String(
			"sentrymesh.request_id",
			event.RequestID,
		),
		attribute.String(
			"sentrymesh.audit.decision",
			event.Decision,
		),
	)

	err := auditStore.Write(
		ctx,
		event,
	)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(
			codes.Error,
			err.Error(),
		)

		return err
	}

	return nil
}

func ChatHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	requestID := requestID(r)
	started := time.Now()

	pipelineCtx, pipelineSpan :=
		chatTracer.Start(
			r.Context(),
			"chat.security_pipeline",
		)
	defer pipelineSpan.End()

	r = r.WithContext(
		pipelineCtx,
	)

	pipelineSpan.SetAttributes(
		attribute.String(
			"sentrymesh.request_id",
			requestID,
		),
	)

	var req ChatRequest

	if err := json.NewDecoder(
		r.Body,
	).Decode(&req); err != nil {
		pipelineSpan.RecordError(err)
		pipelineSpan.SetStatus(
			codes.Error,
			"invalid request body",
		)

		writeError(
			w,
			http.StatusBadRequest,
			requestID,
			"invalid request body",
		)
		return
	}

	if len(req.Messages) == 0 {
		pipelineSpan.SetStatus(
			codes.Error,
			"messages cannot be empty",
		)

		writeError(
			w,
			http.StatusBadRequest,
			requestID,
			"messages cannot be empty",
		)
		return
	}

	if req.Provider == "" {
		req.Provider = "mock"
	}

	pipelineSpan.SetAttributes(
		attribute.String(
			"sentrymesh.provider",
			req.Provider,
		),
		attribute.String(
			"sentrymesh.model",
			req.Model,
		),
		attribute.Int(
			"sentrymesh.message_count",
			len(req.Messages),
		),
	)

	var content strings.Builder

	for _, message := range req.Messages {
		content.WriteString(
			message.Content,
		)
		content.WriteString("\n")
	}

	rawPrompt :=
		strings.TrimSpace(
			content.String(),
		)

	_, scanSpan :=
		chatTracer.Start(
			pipelineCtx,
			"security.input_scan",
		)

	secretFindings :=
		scanner.ScanSecrets(
			rawPrompt,
		)

	injectionFindings :=
		scanner.ScanPromptInjection(
			rawPrompt,
		)

	sanitizedPrompt, piiFindings :=
		scanner.ScanAndRedactPII(
			rawPrompt,
		)

	scanSpan.SetAttributes(
		attribute.Int(
			"sentrymesh.findings.secrets",
			len(secretFindings),
		),
		attribute.Int(
			"sentrymesh.findings.pii",
			len(piiFindings),
		),
		attribute.Int(
			"sentrymesh.findings.injection",
			len(injectionFindings),
		),
	)

	scanSpan.End()

	maxInjectionScore := 0

	for _, finding := range injectionFindings {
		if finding.Confidence >
			maxInjectionScore {
			maxInjectionScore =
				finding.Confidence
		}
	}

	_, riskSpan :=
		chatTracer.Start(
			pipelineCtx,
			"security.risk_evaluation",
		)

	riskDecision :=
		risk.Evaluate(
			risk.Input{
				SecretCount:    len(secretFindings),
				PIICount:       len(piiFindings),
				InjectionCount: len(injectionFindings),
				MaxInjection:   maxInjectionScore,
			},
		)

	riskSpan.SetAttributes(
		attribute.String(
			"sentrymesh.decision",
			riskDecision.Action,
		),
		attribute.Int(
			"sentrymesh.risk_score",
			riskDecision.Score,
		),
		attribute.String(
			"sentrymesh.severity",
			riskDecision.Severity,
		),
	)

	riskSpan.End()

	pipelineSpan.SetAttributes(
		attribute.String(
			"sentrymesh.decision",
			riskDecision.Action,
		),
		attribute.Int(
			"sentrymesh.risk_score",
			riskDecision.Score,
		),
		attribute.String(
			"sentrymesh.severity",
			riskDecision.Severity,
		),
	)

	if riskDecision.Action ==
		"BLOCK" {
		metrics.IncSecurityBlock()

		duration :=
			time.Since(started)

		_ = writeAuditWithTrace(
			pipelineCtx,
			audit.Event{
				RequestID:         requestID,
				Timestamp:         time.Now(),
				Provider:          req.Provider,
				Model:             req.Model,
				Decision:          riskDecision.Action,
				RiskScore:         riskDecision.Score,
				Severity:          riskDecision.Severity,
				LatencyMS:         duration.Milliseconds(),
				LatencyUS:         duration.Microseconds(),
				SecretFindings:    secretFindings,
				PIIFindings:       piiFindings,
				InjectionFindings: injectionFindings,
			},
		)

		w.WriteHeader(
			http.StatusForbidden,
		)

		_ = json.NewEncoder(
			w,
		).Encode(
			SecurityResponse{
				RequestID:         requestID,
				Decision:          riskDecision.Action,
				RiskScore:         riskDecision.Score,
				Severity:          riskDecision.Severity,
				Message:           "request blocked by SentryMesh security policy",
				SanitizedPrompt:   sanitizedPrompt,
				SecretFindings:    secretFindings,
				PIIFindings:       piiFindings,
				InjectionFindings: injectionFindings,
			},
		)

		return
	}

	providerMessages :=
		make(
			[]provider.Message,
			0,
			len(req.Messages),
		)

	for _, message := range req.Messages {
		sanitizedContent, _ :=
			scanner.ScanAndRedactPII(
				message.Content,
			)

		providerMessages =
			append(
				providerMessages,
				provider.Message{
					Role:    message.Role,
					Content: sanitizedContent,
				},
			)
	}

	providerRequest :=
		r.WithContext(
			pipelineCtx,
		)

	providerCtx, cancelProvider :=
		providerContext(
			providerRequest,
		)
	defer cancelProvider()

	providerCtx, providerSpan :=
		chatTracer.Start(
			providerCtx,
			"provider.generate",
		)

	providerSpan.SetAttributes(
		attribute.String(
			"sentrymesh.provider",
			req.Provider,
		),
		attribute.String(
			"sentrymesh.model",
			req.Model,
		),
	)

	modelResponse, err :=
		providerRouter.Chat(
			providerCtx,
			req.Provider,
			provider.Request{
				Model:    req.Model,
				Messages: providerMessages,
			},
		)

	if err != nil {
		providerSpan.RecordError(err)
		providerSpan.SetStatus(
			codes.Error,
			err.Error(),
		)
		providerSpan.End()

		metrics.IncProviderError()

		if isProviderTimeout(
			providerCtx,
			err,
		) {
			writeJSONError(
				w,
				http.StatusGatewayTimeout,
				"model provider timed out",
			)
			return
		}

		writeJSONError(
			w,
			http.StatusBadGateway,
			"model provider request failed",
		)
		return
	}

	providerSpan.End()

	_, outputSpan :=
		chatTracer.Start(
			pipelineCtx,
			"security.output_scan",
		)

	outputScan :=
		scanner.ScanOutput(
			modelResponse.Content,
		)

	outputSpan.SetAttributes(
		attribute.Bool(
			"sentrymesh.output.safe",
			outputScan.Safe,
		),
		attribute.Int(
			"sentrymesh.output.pii_findings",
			len(
				outputScan.PIIFindings,
			),
		),
	)

	outputSpan.End()

	if !outputScan.Safe {
		pipelineSpan.SetAttributes(
			attribute.String(
				"sentrymesh.decision",
				"BLOCK",
			),
			attribute.Int(
				"sentrymesh.risk_score",
				100,
			),
			attribute.String(
				"sentrymesh.severity",
				"CRITICAL",
			),
		)

		w.WriteHeader(
			http.StatusForbidden,
		)

		_ = json.NewEncoder(
			w,
		).Encode(
			SecurityResponse{
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
			},
		)

		return
	}

	finalOutput :=
		outputScan.Redacted

	message :=
		"request passed SentryMesh security checks"

	if riskDecision.Action ==
		"ALLOW_WITH_REDACTION" {
		metrics.IncRedaction()

		message =
			"request allowed after sensitive data redaction"
	}

	if len(
		outputScan.PIIFindings,
	) > 0 {
		message =
			"request allowed after output redaction"
	}

	duration :=
		time.Since(started)

	_ = writeAuditWithTrace(
		pipelineCtx,
		audit.Event{
			RequestID:         requestID,
			Timestamp:         time.Now(),
			Provider:          req.Provider,
			Model:             req.Model,
			Decision:          riskDecision.Action,
			RiskScore:         riskDecision.Score,
			Severity:          riskDecision.Severity,
			LatencyMS:         duration.Milliseconds(),
			LatencyUS:         duration.Microseconds(),
			SecretFindings:    secretFindings,
			PIIFindings:       piiFindings,
			InjectionFindings: injectionFindings,
			OutputFindings:    outputScan,
		},
	)

	w.WriteHeader(
		http.StatusOK,
	)

	_ = json.NewEncoder(
		w,
	).Encode(
		SecurityResponse{
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
		},
	)
}

func writeError(
	w http.ResponseWriter,
	status int,
	requestID string,
	message string,
) {
	w.WriteHeader(status)

	_ = json.NewEncoder(
		w,
	).Encode(
		map[string]string{
			"error":      message,
			"request_id": requestID,
		},
	)
}
