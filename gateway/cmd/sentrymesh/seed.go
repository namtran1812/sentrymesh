package main

import (
	"context"
	"log"

	"github.com/namtran1812/sentrymesh/gateway/internal/auth"
	"github.com/namtran1812/sentrymesh/gateway/internal/identity"
)

func seedDevelopmentKeys() {
	ctx := context.Background()

	keys := []struct {
		name      string
		raw       string
		principal identity.Identity
	}{
		{
			name: "analyst-dev",
			raw:  "sm_analyst_dev",
			principal: identity.Identity{
				UserID: "u_analyst_1",
				Role:   identity.Analyst,
				Team:   "risk",
			},
		},
		{
			name: "sales-dev",
			raw:  "sm_sales_dev",
			principal: identity.Identity{
				UserID: "u_sales_1",
				Role:   identity.Sales,
				Team:   "enterprise",
			},
		},
		{
			name: "admin-dev",
			raw:  "sm_admin_dev",
			principal: identity.Identity{
				UserID: "u_admin_1",
				Role:   identity.Admin,
				Team:   "security",
			},
		},
	}

	for _, key := range keys {
		err := auth.DefaultStore.Create(
			ctx,
			key.name,
			key.raw,
			key.principal,
			nil,
		)

		if err != nil {
			continue
		}

		log.Printf("created development API key: %s", key.name)
	}
}
