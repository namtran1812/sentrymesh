package runtime

import (
	"log"
	"os"

	"github.com/namtran1812/sentrymesh/gateway/internal/abuse"
	"github.com/namtran1812/sentrymesh/gateway/internal/audit"
)

var AuditStore audit.Repository = func() audit.Repository {
	path := os.Getenv("SENTRYMESH_AUDIT_DB")

	if path == "" {
		path = "sentrymesh-audit.db"
	}

	store, err := audit.NewStore(path)
	if err != nil {
		log.Fatalf(
			"initialize audit store: %v",
			err,
		)
	}

	if err := store.EnsureToolEvents(); err != nil {
		log.Fatalf(
			"initialize tool audit events: %v",
			err,
		)
	}

	if err := store.EnsureAuthEvents(); err != nil {
		log.Fatalf(
			"initialize auth audit events: %v",
			err,
		)
	}

	if err := store.EnsureRAGEvents(); err != nil {
		log.Fatalf(
			"initialize RAG audit events: %v",
			err,
		)
	}

	if err := store.EnsureAbuseEvents(); err != nil {
		log.Fatalf(
			"initialize abuse audit events: %v",
			err,
		)
	}

	return store
}()

func SetAuditStore(store audit.Repository) {
	if store == nil {
		panic("runtime: nil audit repository")
	}

	AuditStore = store
}

var AbuseStore abuse.Repository = func() abuse.Repository {
	path := os.Getenv("SENTRYMESH_ABUSE_DB")

	if path == "" {
		path = "sentrymesh-abuse.db"
	}

	store, err := abuse.NewStore(path)
	if err != nil {
		log.Fatalf(
			"initialize abuse store: %v",
			err,
		)
	}

	return store
}()

func SetAbuseStore(
	store abuse.Repository,
) {
	if store == nil {
		panic(
			"runtime: nil abuse repository",
		)
	}

	AbuseStore = store
}
