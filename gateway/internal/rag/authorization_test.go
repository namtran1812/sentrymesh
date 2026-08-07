package rag

import (
	"testing"

	"github.com/namtran1812/sentrymesh/gateway/internal/identity"
)

func TestCrossTeamAccessDenied(t *testing.T) {
	document := Document{
		ID:        "doc_1",
		OwnerTeam: "sales",
	}

	principal := identity.Identity{
		UserID: "u_1",
		Role:   identity.Analyst,
		Team:   "risk",
	}

	result := Authorize(document, principal)

	if result.Allowed {
		t.Fatal("expected cross-team access denial")
	}
}

func TestAdminCanAccessCrossTeamDocument(t *testing.T) {
	document := Document{
		ID:        "doc_1",
		OwnerTeam: "sales",
	}

	principal := identity.Identity{
		UserID: "u_admin",
		Role:   identity.Admin,
		Team:   "security",
	}

	result := Authorize(document, principal)

	if !result.Allowed {
		t.Fatalf("expected access, got %s", result.Reason)
	}
}

func TestRestrictedDocumentDeniedToSales(t *testing.T) {
	document := Document{
		ID:             "doc_restricted",
		Classification: "RESTRICTED",
	}

	principal := identity.Identity{
		UserID: "u_sales",
		Role:   identity.Sales,
		Team:   "sales",
	}

	result := Authorize(document, principal)

	if result.Allowed {
		t.Fatal("expected restricted document denial")
	}
}
