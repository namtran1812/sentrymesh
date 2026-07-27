package auth

import (
	"log"
	"os"
)

var DefaultStore = func() *Store {
	path := os.Getenv("SENTRYMESH_AUTH_DB")

	if path == "" {
		path = "sentrymesh-auth.db"
	}

	store, err := NewStore(path)
	if err != nil {
		log.Fatalf("initialize auth store: %v", err)
	}

	return store
}()
