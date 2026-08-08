package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func OpenPostgres(
	ctx context.Context,
	databaseURL string,
) (*pgxpool.Pool, error) {
	config, err :=
		pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf(
			"parse postgres config: %w",
			err,
		)
	}

	config.MaxConns = 20
	config.MinConns = 2
	config.MaxConnLifetime =
		30 * time.Minute
	config.MaxConnIdleTime =
		5 * time.Minute
	config.HealthCheckPeriod =
		30 * time.Second

	pool, err :=
		pgxpool.NewWithConfig(
			ctx,
			config,
		)
	if err != nil {
		return nil, fmt.Errorf(
			"open postgres pool: %w",
			err,
		)
	}

	pingCtx, cancel :=
		context.WithTimeout(
			ctx,
			5*time.Second,
		)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()

		return nil, fmt.Errorf(
			"ping postgres: %w",
			err,
		)
	}

	return pool, nil
}
