package scanner

import (
	"strings"
	"testing"
)

func TestRedactsEmail(t *testing.T) {
	input := "Contact alice@example.com"

	redacted, findings := ScanAndRedactPII(input)

	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}

	if strings.Contains(redacted, "alice@example.com") {
		t.Fatalf("email was not redacted")
	}

	if !strings.Contains(redacted, "<EMAIL_REDACTED>") {
		t.Fatalf("expected email placeholder, got %s", redacted)
	}
}

func TestRedactsSSN(t *testing.T) {
	input := "SSN is 123-45-6789"

	redacted, findings := ScanAndRedactPII(input)

	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}

	if strings.Contains(redacted, "123-45-6789") {
		t.Fatalf("SSN was not redacted")
	}
}

func TestNoPII(t *testing.T) {
	input := "summarize this document"

	redacted, findings := ScanAndRedactPII(input)

	if redacted != input {
		t.Fatalf("expected unchanged input")
	}

	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}
