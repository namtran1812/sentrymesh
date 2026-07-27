package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/namtran1812/sentrymesh/gateway/internal/approval"
	"github.com/namtran1812/sentrymesh/gateway/internal/executor"
)

func ExecuteApprovalHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
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

	claimed, err := approvalStore.ClaimExecution(r.Context(), id)
	if err != nil {
		http.Error(
			w,
			`{"error":"failed to claim execution"}`,
			http.StatusInternalServerError,
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

	if err := approvalStore.FinishExecution(r.Context(), id); err != nil {
		http.Error(
			w,
			`{"error":"failed to finalize execution"}`,
			http.StatusInternalServerError,
		)
		return
	}

	_ = json.NewEncoder(w).Encode(result)
}
