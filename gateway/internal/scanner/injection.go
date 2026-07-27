package scanner

import (
	"regexp"
	"strings"
)

type InjectionFinding struct {
	Type       string `json:"type"`
	Severity   string `json:"severity"`
	Confidence int    `json:"confidence"`
	Action     string `json:"action"`
	Matched    string `json:"matched,omitempty"`
}

var injectionPatterns = []struct {
	name       string
	severity   string
	confidence int
	pattern    *regexp.Regexp
}{
	{
		name:       "IGNORE_PREVIOUS_INSTRUCTIONS",
		severity:   "HIGH",
		confidence: 95,
		pattern: regexp.MustCompile(
			`(?i)\b(ignore|disregard|forget)\b.{0,40}\b(previous|prior|earlier|above)\b.{0,30}\b(instructions?|rules?|prompt)\b`,
		),
	},
	{
		name:       "SYSTEM_PROMPT_EXTRACTION",
		severity:   "HIGH",
		confidence: 90,
		pattern: regexp.MustCompile(
			`(?i)\b(reveal|show|print|output|repeat|expose)\b.{0,40}\b(system prompt|developer message|hidden instructions?)\b`,
		),
	},
	{
		name:       "ROLE_OVERRIDE",
		severity:   "MEDIUM",
		confidence: 75,
		pattern: regexp.MustCompile(
			`(?i)\b(you are now|act as|pretend to be)\b.{0,60}\b(unrestricted|developer mode|jailbroken|no rules)\b`,
		),
	},
	{
		name:       "DATA_EXFILTRATION",
		severity:   "CRITICAL",
		confidence: 95,
		pattern: regexp.MustCompile(
			`(?i)\b(export|dump|retrieve|send|list)\b.{0,60}\b(all|every)\b.{0,30}\b(customers?|users?|records?|credentials?|secrets?)\b`,
		),
	},
}

func ScanPromptInjection(input string) []InjectionFinding {
	findings := make([]InjectionFinding, 0)

	normalized := strings.Join(strings.Fields(input), " ")

	for _, candidate := range injectionPatterns {
		match := candidate.pattern.FindString(normalized)
		if match == "" {
			continue
		}

		findings = append(findings, InjectionFinding{
			Type:       candidate.name,
			Severity:   candidate.severity,
			Confidence: candidate.confidence,
			Action:     "BLOCK",
			Matched:    match,
		})
	}

	return findings
}
