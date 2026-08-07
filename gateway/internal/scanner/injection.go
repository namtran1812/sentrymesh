package scanner

import (
	"encoding/base64"
	"regexp"
	"strings"
	"unicode"
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

	// Spanish
	{
		name:       "MULTILINGUAL_INSTRUCTION_OVERRIDE",
		severity:   "HIGH",
		confidence: 90,
		pattern: regexp.MustCompile(
			`(?i)\b(ignora|olvida)\b.{0,50}\b(instrucciones|reglas)\b.{0,30}\b(anteriores|previas)\b`,
		),
	},
	{
		name:       "MULTILINGUAL_PROMPT_EXTRACTION",
		severity:   "HIGH",
		confidence: 90,
		pattern: regexp.MustCompile(
			`(?i)\b(revela|muestra|imprime)\b.{0,50}\b(prompt del sistema|instrucciones ocultas)\b`,
		),
	},

	// French
	{
		name:       "MULTILINGUAL_INSTRUCTION_OVERRIDE",
		severity:   "HIGH",
		confidence: 90,
		pattern: regexp.MustCompile(
			`(?i)\b(ignore|oublie)\b.{0,50}\b(instructions|règles|regles)\b.{0,30}\b(précédentes|precedentes|antérieures|anterieures)\b`,
		),
	},
	{
		name:       "MULTILINGUAL_PROMPT_EXTRACTION",
		severity:   "HIGH",
		confidence: 90,
		pattern: regexp.MustCompile(
			`(?i)\b(révèle|revele|montre|affiche)\b.{0,50}\b(prompt système|prompt systeme|instructions cachées|instructions cachees)\b`,
		),
	},
}

var base64CandidatePattern = regexp.MustCompile(
	`\b[A-Za-z0-9+/]{24,}={0,2}\b`,
)

func normalizeInjectionInput(input string) string {
	var b strings.Builder

	for _, r := range input {
		// Remove zero-width / formatting characters.
		if unicode.Is(unicode.Cf, r) {
			continue
		}

		// Normalize a few common Cyrillic homoglyphs used to bypass ASCII rules.
		switch r {
		case 'а':
			r = 'a'
		case 'е':
			r = 'e'
		case 'о':
			r = 'o'
		case 'р':
			r = 'p'
		case 'с':
			r = 'c'
		case 'х':
			r = 'x'
		case 'А':
			r = 'A'
		case 'Е':
			r = 'E'
		case 'О':
			r = 'O'
		case 'Р':
			r = 'P'
		case 'С':
			r = 'C'
		case 'Х':
			r = 'X'
		}

		b.WriteRune(r)
	}

	return strings.Join(strings.Fields(b.String()), " ")
}

func scanPromptInjectionRaw(input string) []InjectionFinding {
	normalized := normalizeInjectionInput(input)

	findings := make([]InjectionFinding, 0)

	for _, candidate := range injectionPatterns {
		match := candidate.pattern.FindString(normalized)
		if match == "" {
			continue
		}

		findings = append(
			findings,
			InjectionFinding{
				Type:       candidate.name,
				Severity:   candidate.severity,
				Confidence: candidate.confidence,
				Action:     "BLOCK",
				Matched:    match,
			},
		)
	}

	return findings
}

func scanEncodedInjection(input string) []InjectionFinding {
	candidates := base64CandidatePattern.FindAllString(input, -1)

	for _, candidate := range candidates {
		decoded, err := base64.StdEncoding.DecodeString(candidate)
		if err != nil {
			continue
		}

		decodedText := string(decoded)

		if len(scanPromptInjectionRaw(decodedText)) == 0 {
			continue
		}

		return []InjectionFinding{
			{
				Type:       "ENCODED_PROMPT_INJECTION",
				Severity:   "HIGH",
				Confidence: 90,
				Action:     "BLOCK",
				Matched:    candidate,
			},
		}
	}

	return nil
}

func ScanPromptInjection(input string) []InjectionFinding {
	if isBenignInjectionDiscussion(input) {
		return nil
	}

	findings := scanPromptInjectionRaw(input)

	encoded := scanEncodedInjection(input)

	return append(findings, encoded...)
}
