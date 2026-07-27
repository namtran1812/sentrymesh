package risk

import "testing"

func TestSecretsCauseBlock(t *testing.T) {
	result := Evaluate(Input{
		SecretCount: 1,
	})

	if result.Action != "BLOCK" {
		t.Fatalf("expected BLOCK, got %s", result.Action)
	}
}

func TestPIICausesRedactionDecision(t *testing.T) {
	result := Evaluate(Input{
		PIICount: 2,
	})

	if result.Action != "ALLOW_WITH_REDACTION" {
		t.Fatalf("expected ALLOW_WITH_REDACTION, got %s", result.Action)
	}
}

func TestBenignRequestAllowed(t *testing.T) {
	result := Evaluate(Input{})

	if result.Action != "ALLOW" {
		t.Fatalf("expected ALLOW, got %s", result.Action)
	}
}
