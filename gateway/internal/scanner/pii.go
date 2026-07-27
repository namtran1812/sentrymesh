package scanner

import (
	"regexp"
)

type PIIFinding struct {
	Type       string `json:"type"`
	Severity   string `json:"severity"`
	Action     string `json:"action"`
	RedactedAs string `json:"redacted_as"`
}

var emailPattern = regexp.MustCompile(
	`[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}`,
)

var phonePattern = regexp.MustCompile(
	`\b(?:\+?1[-.\s]?)?(?:\(?\d{3}\)?[-.\s]?)\d{3}[-.\s]?\d{4}\b`,
)

var ssnPattern = regexp.MustCompile(
	`\b\d{3}-\d{2}-\d{4}\b`,
)

func ScanAndRedactPII(input string) (string, []PIIFinding) {
	redacted := input
	findings := make([]PIIFinding, 0)

	if emailPattern.MatchString(redacted) {
		findings = append(findings, PIIFinding{
			Type:       "EMAIL_ADDRESS",
			Severity:   "MEDIUM",
			Action:     "REDACT",
			RedactedAs: "<EMAIL_REDACTED>",
		})

		redacted = emailPattern.ReplaceAllString(
			redacted,
			"<EMAIL_REDACTED>",
		)
	}

	if phonePattern.MatchString(redacted) {
		findings = append(findings, PIIFinding{
			Type:       "PHONE_NUMBER",
			Severity:   "MEDIUM",
			Action:     "REDACT",
			RedactedAs: "<PHONE_REDACTED>",
		})

		redacted = phonePattern.ReplaceAllString(
			redacted,
			"<PHONE_REDACTED>",
		)
	}

	if ssnPattern.MatchString(redacted) {
		findings = append(findings, PIIFinding{
			Type:       "SSN",
			Severity:   "HIGH",
			Action:     "REDACT",
			RedactedAs: "<SSN_REDACTED>",
		})

		redacted = ssnPattern.ReplaceAllString(
			redacted,
			"<SSN_REDACTED>",
		)
	}

	return redacted, findings
}
