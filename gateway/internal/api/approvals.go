package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/namtran1812/sentrymesh/gateway/internal/approval"
	"github.com/namtran1812/sentrymesh/gateway/internal/audit"
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

	rawID := r.PathValue("id")

	id, err := strconv.ParseInt(rawID, 10, 64)
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
			Details:    item.Arguments,
		},
	)

	_ = json.NewEncoder(w).Encode(
		map[string]any{
			"id":     id,
			"status": status,
		},
	)
}
