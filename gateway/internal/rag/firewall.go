package rag

import (
	"strings"

	"github.com/namtran1812/sentrymesh/gateway/internal/scanner"
)

type Decision string

const (
	Allow  Decision = "ALLOW"
	Filter Decision = "FILTER"
	Block  Decision = "BLOCK"
)

type Result struct {
	DocumentID        string                     `json:"document_id"`
	Decision          Decision                   `json:"decision"`
	SanitizedContent  string                     `json:"sanitized_content,omitempty"`
	InjectionFindings []scanner.InjectionFinding `json:"injection_findings,omitempty"`
	Reason            string                     `json:"reason"`
}

func Inspect(document Document) Result {
	content := strings.TrimSpace(document.Content)

	findings := scanner.ScanPromptInjection(content)

	if len(findings) == 0 {
		return Result{
			DocumentID:       document.ID,
			Decision:         Allow,
			SanitizedContent: content,
			Reason:           "document passed retrieval security checks",
		}
	}

	if document.TrustLevel == UntrustedExternal {
		return Result{
			DocumentID:        document.ID,
			Decision:          Block,
			InjectionFindings: findings,
			Reason:            "untrusted document contains prompt injection",
		}
	}

	return Result{
		DocumentID:        document.ID,
		Decision:          Filter,
		SanitizedContent:  removeSuspiciousLines(content),
		InjectionFindings: findings,
		Reason:            "trusted document contained suspicious instructions and was filtered",
	}
}

func removeSuspiciousLines(content string) string {
	lines := strings.Split(content, "\n")
	safe := make([]string, 0, len(lines))

	for _, line := range lines {
		if len(scanner.ScanPromptInjection(line)) > 0 {
			continue
		}

		safe = append(safe, line)
	}

	return strings.TrimSpace(strings.Join(safe, "\n"))
}
