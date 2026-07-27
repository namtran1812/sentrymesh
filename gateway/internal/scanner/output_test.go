package scanner

import (
	"strings"
	"testing"
)

func TestOutputPIIRedaction(t *testing.T) {
	result := ScanOutput("Contact alice@example.com")

	if strings.Contains(result.Redacted, "alice@example.com") {
		t.Fatal("expected email redaction")
	}

	if len(result.PIIFindings) != 1 {
		t.Fatalf("expected 1 PII finding, got %d", len(result.PIIFindings))
	}
}

func TestOutputSecretUnsafe(t *testing.T) {
	result := ScanOutput("key: AKIA1234567890ABCDEF")

	if result.Safe {
		t.Fatal("expected unsafe output")
	}

	if len(result.SecretFindings) != 1 {
		t.Fatalf("expected 1 secret finding, got %d", len(result.SecretFindings))
	}
}
