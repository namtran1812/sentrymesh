package api

import (
	"log"
	"os"

	"github.com/namtran1812/sentrymesh/gateway/internal/approval"
)

var approvalStore = func() *approval.Store {
	path := os.Getenv("SENTRYMESH_APPROVAL_DB")

	if path == "" {
		path = "sentrymesh-approvals.db"
	}

	store, err := approval.NewStore(path)
	if err != nil {
		log.Fatalf("initialize approval store: %v", err)
	}

	return store
}()
