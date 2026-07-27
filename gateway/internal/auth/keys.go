package auth

import (
	"fmt"

	"github.com/namtran1812/sentrymesh/gateway/internal/identity"
)

type Principal struct {
	APIKey   string
	Identity identity.Identity
}

var principals = map[string]identity.Identity{
	"sm_analyst_dev": {
		UserID: "u_analyst_1",
		Role:   identity.Analyst,
		Team:   "risk",
	},
	"sm_sales_dev": {
		UserID: "u_sales_1",
		Role:   identity.Sales,
		Team:   "enterprise",
	},
	"sm_admin_dev": {
		UserID: "u_admin_1",
		Role:   identity.Admin,
		Team:   "security",
	},
}

func Resolve(apiKey string) (identity.Identity, error) {
	principal, ok := principals[apiKey]
	if !ok {
		return identity.Identity{}, fmt.Errorf("invalid API key")
	}

	return principal, nil
}
