package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/namtran1812/sentrymesh/gateway/internal/approval"
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

	_ = json.NewEncoder(w).Encode(
		map[string]any{
			"id":     id,
			"status": status,
		},
	)
}
