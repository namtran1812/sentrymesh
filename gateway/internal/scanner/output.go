package scanner

type OutputScan struct {
	Safe           bool         `json:"safe"`
	Redacted       string       `json:"redacted"`
	PIIFindings    []PIIFinding `json:"pii_findings,omitempty"`
	SecretFindings []Finding    `json:"secret_findings,omitempty"`
}

func ScanOutput(input string) OutputScan {
	secretFindings := ScanSecrets(input)
	redacted, piiFindings := ScanAndRedactPII(input)

	return OutputScan{
		Safe:           len(secretFindings) == 0,
		Redacted:       redacted,
		PIIFindings:    piiFindings,
		SecretFindings: secretFindings,
	}
}
