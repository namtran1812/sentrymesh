package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/namtran1812/sentrymesh/gateway/internal/audit"
	"github.com/namtran1812/sentrymesh/gateway/internal/tools"
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

	response := ToolEvaluationResponse{
		Evaluation: result,
	}

	status := http.StatusOK

	if result.Decision == tools.RequireApproval {
		item, err := approvalStore.Create(
			r.Context(),
			req.Name,
			req.Arguments,
			result.Risk,
			result.Reason,
		)
		if err != nil {
			http.Error(
				w,
				`{"error":"failed to create approval request"}`,
				http.StatusInternalServerError,
			)
			return
		}

		response.ApprovalID = &item.ID

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
			log.Printf(
				"failed to write TOOL_APPROVAL_REQUESTED approval=%d: %v",
				item.ID,
				err,
			)
		}

		status = http.StatusAccepted
	}

	if result.Decision == tools.Deny {
		status = http.StatusForbidden
	}

	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(response)
}
