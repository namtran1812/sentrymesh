package api

import (
	"log"
	"os"

	"github.com/namtran1812/sentrymesh/gateway/internal/audit"
)

var auditStore = func() *audit.Store {
	path := os.Getenv("SENTRYMESH_AUDIT_DB")

	if path == "" {
		path = "sentrymesh-audit.db"
	}

	store, err := audit.NewStore(path)
	if err != nil {
		log.Fatalf("initialize audit store: %v", err)
	}

	if err := store.EnsureToolEvents(); err != nil {
		log.Fatalf("initialize tool audit events: %v", err)
	}

	if err := store.EnsureAuthEvents(); err != nil {
		log.Fatalf("initialize auth audit events: %v", err)
	}

	return store
}()
