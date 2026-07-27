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
		score += 20
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

	decision := Decision{
		Score:    score,
		Severity: "LOW",
		Action:   "ALLOW",
	}

	switch {
	case score >= 80:
		decision.Severity = "CRITICAL"
		decision.Action = "BLOCK"
	case score >= 60:
		decision.Severity = "HIGH"
		decision.Action = "REVIEW"
	case score >= 30:
		decision.Severity = "MEDIUM"
		decision.Action = "ALLOW_WITH_REDACTION"
	}

	return decision
}
