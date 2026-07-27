package scanner

import (
	"regexp"
)

type Finding struct {
	Type     string `json:"type"`
	Severity string `json:"severity"`
	Action   string `json:"action"`
}

var patterns = []struct {
	name     string
	severity string
	pattern  *regexp.Regexp
}{
	{
		name:     "AWS_ACCESS_KEY",
		severity: "CRITICAL",
		pattern:  regexp.MustCompile(`AKIA[0-9A-Z]{16}`),
	},
	{
		name:     "GITHUB_TOKEN",
		severity: "CRITICAL",
		pattern:  regexp.MustCompile(`gh[pousr]_[A-Za-z0-9_]{20,}`),
	},
	{
		name:     "PRIVATE_KEY",
		severity: "CRITICAL",
		pattern:  regexp.MustCompile(`-----BEGIN (RSA |EC |OPENSSH )?PRIVATE KEY-----`),
	},
}

func ScanSecrets(input string) []Finding {
	var findings []Finding

	for _, p := range patterns {
		if p.pattern.MatchString(input) {
			findings = append(findings, Finding{
				Type:     p.name,
				Severity: p.severity,
				Action:   "BLOCK",
			})
		}
	}

	return findings
}
