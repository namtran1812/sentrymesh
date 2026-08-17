package main

import (
	"github.com/namtran1812/sentrymesh/gateway/internal/api"
	"github.com/namtran1812/sentrymesh/gateway/internal/auth"
)

func installDependencies(
	deps *Dependencies,
) {
	auth.SetDefaultStore(
		deps.Auth,
	)

	api.SetApprovalStore(
		deps.Approvals,
	)

	api.SetAuditStore(
		deps.Audit,
	)

	api.SetAbuseStore(
		deps.Abuse,
	)
}
