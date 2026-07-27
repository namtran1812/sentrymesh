package tools

import "testing"

func TestReadCustomerAllowed(t *testing.T) {
	result := Evaluate(ToolCall{
		Name: "read_customer",
		Arguments: map[string]any{
			"fields": []any{"name", "email"},
		},
	})

	if result.Decision != Allow {
		t.Fatalf("expected ALLOW, got %s", result.Decision)
	}
}

func TestReadCustomerSensitiveFieldDenied(t *testing.T) {
	result := Evaluate(ToolCall{
		Name: "read_customer",
		Arguments: map[string]any{
			"fields": []any{"name", "ssn"},
		},
	})

	if result.Decision != Deny {
		t.Fatalf("expected DENY, got %s", result.Decision)
	}
}

func TestInternalEmailAllowed(t *testing.T) {
	result := Evaluate(ToolCall{
		Name: "send_email",
		Arguments: map[string]any{
			"to": "security@sentrymesh.local",
		},
	})

	if result.Decision != Allow {
		t.Fatalf("expected ALLOW, got %s", result.Decision)
	}
}

func TestExternalEmailRequiresApproval(t *testing.T) {
	result := Evaluate(ToolCall{
		Name: "send_email",
		Arguments: map[string]any{
			"to": "customer@example.com",
		},
	})

	if result.Decision != RequireApproval {
		t.Fatalf(
			"expected REQUIRE_APPROVAL, got %s",
			result.Decision,
		)
	}
}

func TestProtectedCustomerUpdateDenied(t *testing.T) {
	result := Evaluate(ToolCall{
		Name: "update_customer",
		Arguments: map[string]any{
			"fields": []any{"role"},
		},
	})

	if result.Decision != Deny {
		t.Fatalf("expected DENY, got %s", result.Decision)
	}
}

func TestUnknownToolDenied(t *testing.T) {
	result := Evaluate(ToolCall{
		Name: "unknown_tool",
	})

	if result.Decision != Deny {
		t.Fatalf("expected DENY, got %s", result.Decision)
	}
}
