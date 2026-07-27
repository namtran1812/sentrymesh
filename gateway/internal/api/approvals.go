package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/namtran1812/sentrymesh/gateway/internal/approval"
	"github.com/namtran1812/sentrymesh/gateway/internal/audit"
	"github.com/namtran1812/sentrymesh/gateway/internal/identity"
	"github.com/namtran1812/sentrymesh/gateway/internal/middleware"
)

func ListApprovalsHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	items, err := approvalStore.ListActive(r.Context())
	if err != nil {
		http.Error(
			w,
			`{"error":"failed to list approvals"}`,
			http.StatusInternalServerError,
		)
		return
	}

	_ = json.NewEncoder(w).Encode(items)
}

func ApproveHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	resolveApproval(w, r, approval.Approved)
}

func RejectHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	resolveApproval(w, r, approval.Rejected)
}

func resolveApproval(
	w http.ResponseWriter,
	r *http.Request,
	status approval.Status,
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

	if err := approvalStore.SetStatus(
		r.Context(),
		id,
		status,
	); err != nil {
		http.Error(
			w,
			`{"error":"approval not found or already resolved"}`,
			http.StatusNotFound,
		)
		return
	}

	item, _ := approvalStore.Get(r.Context(), id)

	eventType := "TOOL_APPROVED"

	if status == approval.Rejected {
		eventType = "TOOL_REJECTED"
	}

	_ = auditStore.WriteToolEvent(
		r.Context(),
		audit.ToolEvent{
			Timestamp:  time.Now(),
			ApprovalID: id,
			EventType:  eventType,
			Tool:       item.Tool,
			Risk:       item.Risk,
			Status:     string(status),
			Details: map[string]any{
				"arguments": item.Arguments,
				"actor": map[string]any{
					"user_id": principal.UserID,
					"role":    principal.Role,
					"team":    principal.Team,
				},
			},
		},
	)

	_ = json.NewEncoder(w).Encode(
		map[string]any{
			"id":     id,
			"status": status,
		},
	)
}
