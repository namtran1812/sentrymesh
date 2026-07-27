package tools

import (
	"fmt"
	"strings"
)

type Decision string

const (
	Allow           Decision = "ALLOW"
	RequireApproval Decision = "REQUIRE_APPROVAL"
	Deny            Decision = "DENY"
)

type ToolCall struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

type Evaluation struct {
	Tool     string   `json:"tool"`
	Decision Decision `json:"decision"`
	Reason   string   `json:"reason"`
	Risk     int      `json:"risk"`
}

func Evaluate(call ToolCall) Evaluation {
	switch call.Name {
	case "read_customer":
		return evaluateReadCustomer(call)

	case "search_documents":
		return Evaluation{
			Tool:     call.Name,
			Decision: Allow,
			Reason:   "read-only search operation",
			Risk:     10,
		}

	case "send_email":
		return evaluateSendEmail(call)

	case "update_customer":
		return evaluateUpdateCustomer(call)

	case "delete_customer":
		return Evaluation{
			Tool:     call.Name,
			Decision: Deny,
			Reason:   "destructive customer operation is prohibited",
			Risk:     95,
		}

	case "export_database":
		return Evaluation{
			Tool:     call.Name,
			Decision: Deny,
			Reason:   "bulk data export is prohibited",
			Risk:     100,
		}

	case "execute_shell":
		return Evaluation{
			Tool:     call.Name,
			Decision: Deny,
			Reason:   "arbitrary shell execution is prohibited",
			Risk:     100,
		}

	default:
		return Evaluation{
			Tool:     call.Name,
			Decision: Deny,
			Reason:   "unknown tools are denied by default",
			Risk:     90,
		}
	}
}

func evaluateReadCustomer(call ToolCall) Evaluation {
	fields := stringSlice(call.Arguments["fields"])

	sensitive := map[string]bool{
		"ssn":          true,
		"password":     true,
		"api_key":      true,
		"credit_card":  true,
		"health_data":  true,
		"bank_account": true,
	}

	for _, field := range fields {
		if sensitive[strings.ToLower(field)] {
			return Evaluation{
				Tool:     call.Name,
				Decision: Deny,
				Reason:   fmt.Sprintf("sensitive field %q cannot be accessed", field),
				Risk:     90,
			}
		}
	}

	return Evaluation{
		Tool:     call.Name,
		Decision: Allow,
		Reason:   "requested customer fields are permitted",
		Risk:     10,
	}
}

func evaluateSendEmail(call ToolCall) Evaluation {
	to, _ := call.Arguments["to"].(string)

	if to == "" {
		return Evaluation{
			Tool:     call.Name,
			Decision: Deny,
			Reason:   "email recipient is required",
			Risk:     80,
		}
	}

	if strings.HasSuffix(
		strings.ToLower(to),
		"@sentrymesh.local",
	) {
		return Evaluation{
			Tool:     call.Name,
			Decision: Allow,
			Reason:   "internal email recipient",
			Risk:     20,
		}
	}

	return Evaluation{
		Tool:     call.Name,
		Decision: RequireApproval,
		Reason:   "external email requires human approval",
		Risk:     60,
	}
}

func evaluateUpdateCustomer(call ToolCall) Evaluation {
	fields := stringSlice(call.Arguments["fields"])

	protected := map[string]bool{
		"role":          true,
		"permissions":   true,
		"account_owner": true,
		"billing_plan":  true,
	}

	for _, field := range fields {
		if protected[strings.ToLower(field)] {
			return Evaluation{
				Tool:     call.Name,
				Decision: Deny,
				Reason:   fmt.Sprintf("protected field %q cannot be modified", field),
				Risk:     90,
			}
		}
	}

	return Evaluation{
		Tool:     call.Name,
		Decision: RequireApproval,
		Reason:   "customer mutation requires human approval",
		Risk:     70,
	}
}

func stringSlice(value any) []string {
	items, ok := value.([]any)
	if !ok {
		return nil
	}

	result := make([]string, 0, len(items))

	for _, item := range items {
		if s, ok := item.(string); ok {
			result = append(result, s)
		}
	}

	return result
}
