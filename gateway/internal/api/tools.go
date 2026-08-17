package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/namtran1812/sentrymesh/gateway/internal/approval"
	"github.com/namtran1812/sentrymesh/gateway/internal/audit"
	"github.com/namtran1812/sentrymesh/gateway/internal/middleware"
	"github.com/namtran1812/sentrymesh/gateway/internal/tools"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type ToolEvaluationRequest struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

type ToolEvaluationResponse struct {
	tools.Evaluation
	ApprovalID *int64 `json:"approval_id,omitempty"`
}

func ToolEvaluationHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	var req ToolEvaluationRequest

	if err := decodeJSON(w, r, &req); err != nil {
		return
	}

	if req.Name == "" {
		http.Error(
			w,
			`{"error":"tool name is required"}`,
			http.StatusBadRequest,
		)
		return
	}

	principal, ok := middleware.IdentityFromContext(r.Context())
	if !ok {
		http.Error(
			w,
			`{"error":"authenticated identity unavailable"}`,
			http.StatusUnauthorized,
		)
		return
	}

	pipelineCtx, pipelineSpan := otel.Tracer(
		"sentrymesh/api",
	).Start(
		r.Context(),
		"tool.security_pipeline",
	)
	defer pipelineSpan.End()

	pipelineSpan.SetAttributes(
		attribute.String(
			"sentrymesh.request_id",
			requestID(r),
		),
		attribute.String(
			"sentrymesh.tool.name",
			req.Name,
		),
		attribute.Int64(
			"sentrymesh.key_id",
			principal.KeyID,
		),
		attribute.String(
			"sentrymesh.user_id",
			principal.UserID,
		),
		attribute.String(
			"sentrymesh.role",
			string(principal.Role),
		),
		attribute.String(
			"sentrymesh.team",
			principal.Team,
		),
	)

	r = r.WithContext(
		pipelineCtx,
	)

	var result tools.Evaluation

	func() {
		_, span := otel.Tracer(
			"sentrymesh/api",
		).Start(
			pipelineCtx,
			"tool.policy_evaluation",
		)
		defer span.End()

		result = tools.Evaluate(
			tools.ToolCall{
				Name:      req.Name,
				Arguments: req.Arguments,
				Identity:  principal,
			},
		)

		span.SetAttributes(
			attribute.String(
				"sentrymesh.tool.decision",
				string(result.Decision),
			),
			attribute.Int(
				"sentrymesh.tool.risk",
				result.Risk,
			),
		)
	}()

	pipelineSpan.SetAttributes(
		attribute.String(
			"sentrymesh.tool.decision",
			string(result.Decision),
		),
		attribute.Int(
			"sentrymesh.tool.risk",
			result.Risk,
		),
	)

	response := ToolEvaluationResponse{
		Evaluation: result,
	}

	status := http.StatusOK

	if result.Decision == tools.RequireApproval {
		var item approval.Request

		_, approvalSpan := otel.Tracer(
			"sentrymesh/api",
		).Start(
			pipelineCtx,
			"approval.create",
		)

		item, err := approvalStore.Create(
			pipelineCtx,
			req.Name,
			req.Arguments,
			result.Risk,
			result.Reason,
		)

		if err != nil {
			approvalSpan.RecordError(err)
			approvalSpan.SetStatus(
				codes.Error,
				"approval creation failed",
			)
			approvalSpan.End()

			pipelineSpan.RecordError(err)
			pipelineSpan.SetStatus(
				codes.Error,
				"approval creation failed",
			)

			http.Error(
				w,
				`{"error":"failed to create approval request"}`,
				http.StatusInternalServerError,
			)
			return
		}

		approvalSpan.SetAttributes(
			attribute.Int64(
				"sentrymesh.approval_id",
				item.ID,
			),
		)
		approvalSpan.End()

		response.ApprovalID = &item.ID

		_, auditSpan := otel.Tracer(
			"sentrymesh/api",
		).Start(
			pipelineCtx,
			"audit.enqueue",
		)

		err = auditStore.WriteToolEvent(
			r.Context(),
			audit.ToolEvent{
				Timestamp:  time.Now(),
				ApprovalID: item.ID,
				EventType:  "TOOL_APPROVAL_REQUESTED",
				Tool:       req.Name,
				Risk:       result.Risk,
				Status:     "PENDING",
				Details:    req.Arguments,
			},
		)

		if err != nil {
			auditSpan.RecordError(err)
			auditSpan.SetStatus(
				codes.Error,
				"tool audit write failed",
			)

			log.Printf(
				"failed to write TOOL_APPROVAL_REQUESTED approval=%d: %v",
				item.ID,
				err,
			)
		}

		auditSpan.SetAttributes(
			attribute.Int64(
				"sentrymesh.approval_id",
				item.ID,
			),
			attribute.String(
				"sentrymesh.audit.event_type",
				"TOOL_APPROVAL_REQUESTED",
			),
		)
		auditSpan.End()

		pipelineSpan.SetAttributes(
			attribute.Int64(
				"sentrymesh.approval_id",
				item.ID,
			),
		)

		status = http.StatusAccepted
	}

	if result.Decision == tools.Deny {
		status = http.StatusForbidden
	}

	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(response)
}
