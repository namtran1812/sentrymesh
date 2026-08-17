package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/namtran1812/sentrymesh/gateway/internal/approval"
	"github.com/namtran1812/sentrymesh/gateway/internal/audit"
	"github.com/namtran1812/sentrymesh/gateway/internal/executor"
	"github.com/namtran1812/sentrymesh/gateway/internal/identity"
	"github.com/namtran1812/sentrymesh/gateway/internal/middleware"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

func ExecuteApprovalHandler(
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

	if principal.Role != identity.Admin {
		http.Error(
			w,
			`{"error":"admin role required"}`,
			http.StatusForbidden,
		)
		return
	}

	id, err := strconv.ParseInt(
		r.PathValue("id"),
		10,
		64,
	)
	if err != nil {
		http.Error(
			w,
			`{"error":"invalid approval id"}`,
			http.StatusBadRequest,
		)
		return
	}

	item, err := approvalStore.Get(r.Context(), id)
	if err != nil {
		http.Error(
			w,
			`{"error":"approval not found"}`,
			http.StatusNotFound,
		)
		return
	}

	if item.ExecutedAt != nil || item.Status == "EXECUTED" {
		http.Error(
			w,
			`{"error":"approval already executed"}`,
			http.StatusConflict,
		)
		return
	}

	if item.Status == "EXECUTING" {
		http.Error(
			w,
			`{"error":"approval execution already in progress"}`,
			http.StatusConflict,
		)
		return
	}

	if item.Status != approval.Approved {
		http.Error(
			w,
			`{"error":"approval must be APPROVED before execution"}`,
			http.StatusConflict,
		)
		return
	}

	pipelineCtx, pipelineSpan := otel.Tracer(
		"sentrymesh/api",
	).Start(
		r.Context(),
		"tool.execution_pipeline",
	)
	defer pipelineSpan.End()

	pipelineSpan.SetAttributes(
		attribute.String(
			"sentrymesh.request_id",
			requestID(r),
		),
		attribute.Int64(
			"sentrymesh.approval_id",
			id,
		),
		attribute.String(
			"sentrymesh.tool.name",
			item.Tool,
		),
		attribute.Int(
			"sentrymesh.tool.risk",
			item.Risk,
		),
	)

	r = r.WithContext(
		pipelineCtx,
	)

	_, claimSpan := otel.Tracer(
		"sentrymesh/api",
	).Start(
		pipelineCtx,
		"approval.claim",
	)

	claimed, err := approvalStore.ClaimExecution(
		r.Context(),
		id,
	)
	if err != nil {
		http.Error(
			w,
			`{"error":"execution claim conflict"}`,
			http.StatusConflict,
		)
		return
	}

	if !claimed {
		claimSpan.SetStatus(
			codes.Error,
			"execution not claimed",
		)
		claimSpan.End()

		http.Error(
			w,
			`{"error":"approval already claimed or executed"}`,
			http.StatusConflict,
		)
		return
	}

	claimSpan.End()

	_, startedAuditSpan := otel.Tracer(
		"sentrymesh/api",
	).Start(
		pipelineCtx,
		"audit.execution_started",
	)

	_ = auditStore.WriteToolEvent(
		r.Context(),
		audit.ToolEvent{
			Timestamp:  time.Now(),
			ApprovalID: id,
			EventType:  "TOOL_EXECUTION_STARTED",
			Tool:       item.Tool,
			Risk:       item.Risk,
			Status:     "EXECUTING",
			Details: map[string]any{
				"actor": map[string]any{
					"user_id": principal.UserID,
					"role":    principal.Role,
					"team":    principal.Team,
				},
			},
		},
	)

	startedAuditSpan.End()

	var arguments map[string]any

	if err := json.Unmarshal(item.Arguments, &arguments); err != nil {
		_ = approvalStore.FailExecution(r.Context(), id)

		http.Error(
			w,
			`{"error":"invalid stored arguments"}`,
			http.StatusInternalServerError,
		)
		return
	}

	executeCtx, executeSpan := otel.Tracer(
		"sentrymesh/api",
	).Start(
		pipelineCtx,
		"tool.execute",
	)

	result, err := executor.Execute(
		executeCtx,
		item.Tool,
		arguments,
	)

	executeSpan.SetAttributes(
		attribute.String(
			"sentrymesh.tool.name",
			item.Tool,
		),
	)
	if err != nil {
		executeSpan.RecordError(err)
		executeSpan.SetStatus(
			codes.Error,
			"tool execution failed",
		)
		executeSpan.End()

		pipelineSpan.RecordError(err)
		pipelineSpan.SetStatus(
			codes.Error,
			"tool execution failed",
		)

		_ = approvalStore.FailExecution(
			pipelineCtx,
			id,
		)

		http.Error(
			w,
			`{"error":"tool execution failed"}`,
			http.StatusInternalServerError,
		)
		return
	}

	executeSpan.End()

	_, finishSpan := otel.Tracer(
		"sentrymesh/api",
	).Start(
		pipelineCtx,
		"approval.finish",
	)

	if err := approvalStore.FinishExecution(
		r.Context(),
		id,
	); err != nil {
		finishSpan.RecordError(err)
		finishSpan.SetStatus(
			codes.Error,
			"execution finalization failed",
		)
		finishSpan.End()

		pipelineSpan.RecordError(err)
		pipelineSpan.SetStatus(
			codes.Error,
			"execution finalization failed",
		)

		http.Error(
			w,
			`{"error":"failed to finalize execution"}`,
			http.StatusInternalServerError,
		)
		return
	}

	finishSpan.End()

	_, succeededAuditSpan := otel.Tracer(
		"sentrymesh/api",
	).Start(
		pipelineCtx,
		"audit.execution_succeeded",
	)

	_ = auditStore.WriteToolEvent(
		r.Context(),
		audit.ToolEvent{
			Timestamp:  time.Now(),
			ApprovalID: id,
			EventType:  "TOOL_EXECUTION_SUCCEEDED",
			Tool:       item.Tool,
			Risk:       item.Risk,
			Status:     "EXECUTED",
			Details: map[string]any{
				"result": result,
				"actor": map[string]any{
					"user_id": principal.UserID,
					"role":    principal.Role,
					"team":    principal.Team,
				},
			},
		},
	)

	succeededAuditSpan.End()

	pipelineSpan.SetAttributes(
		attribute.String(
			"sentrymesh.execution.status",
			"EXECUTED",
		),
	)

	_ = json.NewEncoder(w).Encode(result)
}
