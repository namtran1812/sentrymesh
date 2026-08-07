package rag

import "time"

type ContextEntry struct {
	DocumentID     string     `json:"document_id"`
	Source         string     `json:"source"`
	TrustLevel     TrustLevel `json:"trust_level"`
	Classification string     `json:"classification"`
	OwnerTeam      string     `json:"owner_team"`
	Decision       Decision   `json:"decision"`
	Included       bool       `json:"included"`
	Reason         string     `json:"reason"`
}

type ContextTrace struct {
	RequestID string         `json:"request_id"`
	Timestamp time.Time      `json:"timestamp"`
	Entries   []ContextEntry `json:"entries"`
}
