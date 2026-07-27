package tools

import "testing"

func TestReadCustomerAllowed(t *testing.T) {
	result := Evaluate(ToolCall{
		Name: "read_customer",
	})

	if result.Decision != Allow {
		t.Fatalf("expected ALLOW, got %s", result.Decision)
	}
}

func TestSendEmailRequiresApproval(t *testing.T) {
	result := Evaluate(ToolCall{
		Name: "send_email",
	})

	if result.Decision != RequireApproval {
		t.Fatalf(
			"expected REQUIRE_APPROVAL, got %s",
			result.Decision,
		)
	}
}

func TestDeleteCustomerDenied(t *testing.T) {
	result := Evaluate(ToolCall{
		Name: "delete_customer",
	})

	if result.Decision != Deny {
		t.Fatalf("expected DENY, got %s", result.Decision)
	}
}

func TestUnknownToolDenied(t *testing.T) {
	result := Evaluate(ToolCall{
		Name: "launch_missiles",
	})

	if result.Decision != Deny {
		t.Fatalf("expected DENY, got %s", result.Decision)
	}
}
