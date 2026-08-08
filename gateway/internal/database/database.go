package database

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func OpenProduction(
	ctx context.Context,
) (*pgxpool.Pool, error) {
	databaseURL := os.Getenv(
		"DATABASE_URL",
	)

	if databaseURL == "" {
		return nil, fmt.Errorf(
			"DATABASE_URL is required",
		)
	}

	pool, err := OpenPostgres(
		ctx,
		databaseURL,
	)
	if err != nil {
		return nil, err
	}

	if err := RunMigrations(
		ctx,
		pool,
	); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}
