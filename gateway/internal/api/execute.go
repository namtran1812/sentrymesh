package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/namtran1812/sentrymesh/gateway/internal/approval"
	"github.com/namtran1812/sentrymesh/gateway/internal/executor"
)

func ExecuteApprovalHandler(
	w http.ResponseWriter,
	r *http.Request,
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

	item, err := approvalStore.Get(r.Context(), id)
	if err != nil {
		http.Error(
			w,
			`{"error":"approval not found"}`,
			http.StatusNotFound,
		)
		return
	}

	if item.ExecutedAt != nil {
		http.Error(
			w,
			`{"error":"approval already executed"}`,
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

	var arguments map[string]any

	if err := json.Unmarshal(item.Arguments, &arguments); err != nil {
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
		http.Error(
			w,
			`{"error":"tool execution failed"}`,
			http.StatusInternalServerError,
		)
		return
	}

	_, err = approvalStore.DB().ExecContext(
		r.Context(),
		`
		UPDATE approvals
		SET executed_at = ?
		WHERE id = ?
		  AND executed_at IS NULL
		`,
		time.Now().UTC().Format(time.RFC3339Nano),
		id,
	)
	if err != nil {
		http.Error(
			w,
			`{"error":"failed to record execution"}`,
			http.StatusInternalServerError,
		)
		return
	}

	_ = json.NewEncoder(w).Encode(result)
}
