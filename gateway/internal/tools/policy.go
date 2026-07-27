package tools

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
		return Evaluation{
			Tool:     call.Name,
			Decision: Allow,
			Reason:   "read-only operation",
			Risk:     10,
		}

	case "search_documents":
		return Evaluation{
			Tool:     call.Name,
			Decision: Allow,
			Reason:   "read-only search operation",
			Risk:     10,
		}

	case "send_email":
		return Evaluation{
			Tool:     call.Name,
			Decision: RequireApproval,
			Reason:   "external side effect requires human approval",
			Risk:     60,
		}

	case "update_customer":
		return Evaluation{
			Tool:     call.Name,
			Decision: RequireApproval,
			Reason:   "customer data mutation requires human approval",
			Risk:     70,
		}

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
