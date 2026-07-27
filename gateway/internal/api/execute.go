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
		http.Error(
			w,
			`{"error":"approval already claimed or executed"}`,
			http.StatusConflict,
		)
		return
	}

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

	result, err := executor.Execute(
		r.Context(),
		item.Tool,
		arguments,
	)
	if err != nil {
		_ = approvalStore.FailExecution(r.Context(), id)

		http.Error(
			w,
			`{"error":"tool execution failed"}`,
			http.StatusInternalServerError,
		)
		return
	}

	if err := approvalStore.FinishExecution(
		r.Context(),
		id,
	); err != nil {
		http.Error(
			w,
			`{"error":"failed to finalize execution"}`,
			http.StatusInternalServerError,
		)
		return
	}

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

	_ = json.NewEncoder(w).Encode(result)
}
