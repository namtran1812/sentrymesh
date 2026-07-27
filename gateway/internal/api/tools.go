package api

import (
	"encoding/json"
	"net/http"

	"github.com/namtran1812/sentrymesh/gateway/internal/tools"
)

type ToolEvaluationRequest struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

func ToolEvaluationHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	var req ToolEvaluationRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(
			w,
			`{"error":"invalid request body"}`,
			http.StatusBadRequest,
		)
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

	result := tools.Evaluate(
		tools.ToolCall{
			Name:      req.Name,
			Arguments: req.Arguments,
		},
	)

	status := http.StatusOK

	if result.Decision == tools.RequireApproval {
		status = http.StatusAccepted
	}

	if result.Decision == tools.Deny {
		status = http.StatusForbidden
	}

	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(result)
}
