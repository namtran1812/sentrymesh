package api

import (
	"github.com/namtran1812/sentrymesh/gateway/internal/audit"
	"github.com/namtran1812/sentrymesh/gateway/internal/runtime"
)

var auditStore audit.Repository = runtime.AuditStore

func SetAuditStore(store audit.Repository) {
	if store == nil {
		panic("api: nil audit repository")
	}

	auditStore = store
}
