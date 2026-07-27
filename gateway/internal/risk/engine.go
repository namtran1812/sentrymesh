package risk

type Input struct {
	SecretCount    int
	PIICount       int
	InjectionCount int
	MaxInjection   int
}

type Decision struct {
	Score    int    `json:"score"`
	Severity string `json:"severity"`
	Action   string `json:"action"`
}

func Evaluate(input Input) Decision {
	score := 0

	if input.PIICount > 0 {
		score += 30
	}

	if input.SecretCount > 0 {
		score += 95
	}

	if input.InjectionCount > 0 {
		score += input.MaxInjection
	}

	if score > 100 {
		score = 100
	}

	switch {
	case score >= 80:
		return Decision{
			Score:    score,
			Severity: "CRITICAL",
			Action:   "BLOCK",
		}

	case score >= 60:
		return Decision{
			Score:    score,
			Severity: "HIGH",
			Action:   "REVIEW",
		}

	case score >= 30:
		return Decision{
			Score:    score,
			Severity: "MEDIUM",
			Action:   "ALLOW_WITH_REDACTION",
		}

	default:
		return Decision{
			Score:    score,
			Severity: "LOW",
			Action:   "ALLOW",
		}
	}
}
