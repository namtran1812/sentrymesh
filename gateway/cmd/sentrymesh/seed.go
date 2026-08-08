package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/namtran1812/sentrymesh/gateway/internal/auth"
	"github.com/namtran1812/sentrymesh/gateway/internal/identity"
)

func seedAPIKeys() {
	env := strings.ToLower(
		strings.TrimSpace(
			os.Getenv("SENTRYMESH_ENV"),
		),
	)

	if env == "production" {
		seedProductionKeys()
		return
	}

	seedDevelopmentKeys()
}

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
				Scopes: []string{
					"tools:evaluate",
					"rag:context",
					"rag:chat",
				},
			},
		},
		{
			name: "sales-dev",
			raw:  "sm_sales_dev",
			principal: identity.Identity{
				UserID: "u_sales_1",
				Role:   identity.Sales,
				Team:   "enterprise",
				Scopes: []string{
					"tools:evaluate",
					"rag:context",
				},
			},
		},
		{
			name: "admin-dev",
			raw:  "sm_admin_dev",
			principal: identity.Identity{
				UserID: "u_admin_1",
				Role:   identity.Admin,
				Team:   "security",
				Scopes: []string{
					"tools:evaluate",
					"approvals:write",
					"tools:execute",
					"audit:read",
					"keys:manage",
					"rag:inspect",
					"rag:context",
					"rag:chat",
					"evals:read",
				},
			},
		},
	}

	for _, key := range keys {
		if err := ensureAPIKey(
			ctx,
			key.name,
			key.raw,
			key.principal,
		); err != nil {
			log.Printf(
				"development API key %s: %v",
				key.name,
				err,
			)
			continue
		}

		log.Printf(
			"development API key ready: %s",
			key.name,
		)
	}
}

func seedProductionKeys() {
	ctx := context.Background()

	rawKey := strings.TrimSpace(
		os.Getenv("SENTRYMESH_ADMIN_KEY"),
	)

	if rawKey == "" {
		log.Fatal(
			"SENTRYMESH_ADMIN_KEY is required in production",
		)
	}

	if len(rawKey) < 32 {
		log.Fatal(
			"SENTRYMESH_ADMIN_KEY must be at least 32 characters",
		)
	}

	if rawKey == "sm_admin_dev" ||
		rawKey == "sm_sales_dev" ||
		rawKey == "sm_analyst_dev" {

		log.Fatal(
			"development API keys cannot be used in production",
		)
	}

	principal := identity.Identity{
		UserID: "u_admin_prod",
		Role:   identity.Admin,
		Team:   "security",
		Scopes: []string{
			"tools:evaluate",
			"approvals:write",
			"tools:execute",
			"audit:read",
			"keys:manage",
			"rag:inspect",
			"rag:context",
			"rag:chat",
			"evals:read",
		},
	}

	if err := ensureAPIKey(
		ctx,
		"admin-prod",
		rawKey,
		principal,
	); err != nil {
		log.Fatalf(
			"initialize production admin key: %v",
			err,
		)
	}

	// Deliberately never print the raw API key.
	log.Println(
		"production admin API key ready",
	)
}

func ensureAPIKey(
	ctx context.Context,
	name string,
	rawKey string,
	principal identity.Identity,
) error {
	_, err := auth.DefaultStore.Resolve(
		ctx,
		rawKey,
	)

	if err == nil {
		return nil
	}

	err = auth.DefaultStore.Create(
		ctx,
		name,
		rawKey,
		principal,
		nil,
	)

	if err != nil {
		return fmt.Errorf(
			"create API key %s: %w",
			name,
			err,
		)
	}

	return nil
}
