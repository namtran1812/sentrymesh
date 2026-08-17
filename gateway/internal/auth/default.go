package auth

import (
	"log"
	"os"
)

var DefaultStore Repository = mustDefaultStore()

func mustDefaultStore() Repository {
	path := os.Getenv(
		"SENTRYMESH_AUTH_DB",
	)

	if path == "" {
		path = "sentrymesh-auth.db"
	}

	store, err := NewStore(path)
	if err != nil {
		log.Fatalf(
			"initialize auth store: %v",
			err,
		)
	}

	return store
}

func SetDefaultStore(
	store Repository,
) {
	if store == nil {
		panic("auth repository cannot be nil")
	}

	DefaultStore = store
}
