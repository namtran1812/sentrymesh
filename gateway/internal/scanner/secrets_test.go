package scanner

import "testing"

func TestDetectAWSAccessKey(t *testing.T) {
	input := "my key is AKIA1234567890ABCDEF"

	findings := ScanSecrets(input)

	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}

	if findings[0].Type != "AWS_ACCESS_KEY" {
		t.Fatalf("expected AWS_ACCESS_KEY, got %s", findings[0].Type)
	}
}

func TestBenignInput(t *testing.T) {
	input := "summarize this support ticket"

	findings := ScanSecrets(input)

	if len(findings) != 0 {
		t.Fatalf("expected no findings, got %d", len(findings))
	}
}
