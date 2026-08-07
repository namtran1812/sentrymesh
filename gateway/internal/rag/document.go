package rag

type TrustLevel string

const (
	TrustedInternal   TrustLevel = "TRUSTED_INTERNAL"
	UntrustedExternal TrustLevel = "UNTRUSTED_EXTERNAL"
)

type Document struct {
	ID             string     `json:"id"`
	Source         string     `json:"source"`
	OwnerTeam      string     `json:"owner_team"`
	Classification string     `json:"classification"`
	TrustLevel     TrustLevel `json:"trust_level"`
	Content        string     `json:"content"`
	AllowedRoles   []string   `json:"allowed_roles,omitempty"`
}
