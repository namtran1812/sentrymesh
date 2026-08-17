package api

import "github.com/namtran1812/sentrymesh/gateway/internal/audit"

var auditStore audit.Repository

func SetAuditStore(
	store audit.Repository,
) {
	if store == nil {
		panic(
			"audit repository cannot be nil",
		)
	}

	auditStore = store
}
