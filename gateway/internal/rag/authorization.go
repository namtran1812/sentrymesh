package rag

import (
	"strings"

	"github.com/namtran1812/sentrymesh/gateway/internal/identity"
)

type AuthorizationResult struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason"`
}

func Authorize(
	document Document,
	principal identity.Identity,
) AuthorizationResult {
	if strings.EqualFold(document.Classification, "RESTRICTED") &&
		principal.Role != identity.Admin {
		return AuthorizationResult{
			Allowed: false,
			Reason:  "restricted document requires admin role",
		}
	}

	if len(document.AllowedRoles) > 0 {
		allowed := false

		for _, role := range document.AllowedRoles {
			if strings.EqualFold(
				role,
				string(principal.Role),
			) {
				allowed = true
				break
			}
		}

		if !allowed {
			return AuthorizationResult{
				Allowed: false,
				Reason:  "caller role is not authorized for document",
			}
		}
	}

	if document.OwnerTeam != "" &&
		principal.Team != document.OwnerTeam &&
		principal.Role != identity.Admin {

		return AuthorizationResult{
			Allowed: false,
			Reason:  "cross-team document access denied",
		}
	}

	return AuthorizationResult{
		Allowed: true,
		Reason:  "document access permitted",
	}
}
