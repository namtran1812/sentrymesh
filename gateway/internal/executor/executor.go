package executor

import (
	"context"
	"fmt"
)

type Result struct {
	Tool   string `json:"tool"`
	Status string `json:"status"`
	Output any    `json:"output,omitempty"`
}

func Execute(
	ctx context.Context,
	tool string,
	arguments map[string]any,
) (Result, error) {
	select {
	case <-ctx.Done():
		return Result{}, ctx.Err()
	default:
	}

	switch tool {
	case "send_email":
		to, _ := arguments["to"].(string)
		subject, _ := arguments["subject"].(string)

		return Result{
			Tool:   tool,
			Status: "EXECUTED",
			Output: map[string]any{
				"to":      to,
				"subject": subject,
				"message": "simulated email sent",
			},
		}, nil

	case "update_customer":
		customerID, _ := arguments["customer_id"].(string)

		return Result{
			Tool:   tool,
			Status: "EXECUTED",
			Output: map[string]any{
				"customer_id": customerID,
				"message":     "simulated customer update applied",
			},
		}, nil

	default:
		return Result{}, fmt.Errorf(
			"tool %q has no executor",
			tool,
		)
	}
}
