package identity

type Role string

const (
	Analyst Role = "analyst"
	Sales   Role = "sales"
	Admin   Role = "admin"
)

type Identity struct {
	UserID string   `json:"user_id"`
	Role   Role     `json:"role"`
	Team   string   `json:"team,omitempty"`
	Scopes []string `json:"scopes,omitempty"`

	KeyID   int64  `json:"key_id,omitempty"`
	KeyName string `json:"key_name,omitempty"`
}
