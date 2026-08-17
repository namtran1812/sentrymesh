package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/namtran1812/sentrymesh/gateway/internal/api"
	"github.com/namtran1812/sentrymesh/gateway/internal/approval"
	"github.com/namtran1812/sentrymesh/gateway/internal/auth"
	"github.com/namtran1812/sentrymesh/gateway/internal/database"
)

type persistenceCloser func()

func configurePrimaryPersistence() (
	persistenceCloser,
	error,
) {
	databaseURL := os.Getenv(
		"DATABASE_URL",
	)

	if databaseURL == "" {
		log.Println(
			"primary persistence: sqlite",
		)

		return func() {}, nil
	}

	ctx, cancel :=
		context.WithTimeout(
			context.Background(),
			15*time.Second,
		)
	defer cancel()

	pool, err :=
		database.OpenProduction(ctx)
	if err != nil {
		return nil, fmt.Errorf(
			"initialize postgres: %w",
			err,
		)
	}

	auth.SetDefaultStore(
		auth.NewPostgresStore(pool),
	)

	api.SetApprovalStore(
		approval.NewPostgresStore(pool),
	)

	log.Println(
		"primary persistence: postgres",
	)

	return func() {
		pool.Close()
	}, nil
}
