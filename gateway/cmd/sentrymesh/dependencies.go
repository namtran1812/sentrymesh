package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/namtran1812/sentrymesh/gateway/internal/abuse"
	"github.com/namtran1812/sentrymesh/gateway/internal/approval"
	"github.com/namtran1812/sentrymesh/gateway/internal/audit"
	"github.com/namtran1812/sentrymesh/gateway/internal/auth"
	"github.com/namtran1812/sentrymesh/gateway/internal/database"
)

type Dependencies struct {
	Auth      auth.Repository
	Approvals approval.Repository
	Audit     audit.Repository
	Abuse     abuse.Repository

	close func()
}

func NewDependencies(
	ctx context.Context,
) (*Dependencies, error) {
	if os.Getenv("DATABASE_URL") != "" {
		return newPostgresDependencies(ctx)
	}

	return newSQLiteDependencies()
}

func newPostgresDependencies(
	ctx context.Context,
) (*Dependencies, error) {
	pool, err := database.OpenProduction(ctx)
	if err != nil {
		return nil, fmt.Errorf(
			"open postgres persistence: %w",
			err,
		)
	}

	log.Println("primary persistence: postgres")

	return &Dependencies{
		Auth: auth.NewPostgresStore(
			pool,
		),
		Approvals: approval.NewPostgresStore(
			pool,
		),
		Audit: audit.NewPostgresStore(
			pool,
		),
		Abuse: abuse.NewPostgresStore(
			pool,
		),

		close: func() {
			pool.Close()
		},
	}, nil
}

func newSQLiteDependencies() (
	*Dependencies,
	error,
) {
	authPath := envOrDefault(
		"SENTRYMESH_AUTH_DB",
		"sentrymesh-auth.db",
	)

	approvalPath := envOrDefault(
		"SENTRYMESH_APPROVAL_DB",
		"sentrymesh-approvals.db",
	)

	auditPath := envOrDefault(
		"SENTRYMESH_AUDIT_DB",
		"sentrymesh-audit.db",
	)

	abusePath := envOrDefault(
		"SENTRYMESH_ABUSE_DB",
		"sentrymesh-abuse.db",
	)

	authStore, err := auth.NewStore(
		authPath,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"open auth store: %w",
			err,
		)
	}

	approvalStore, err :=
		approval.NewStore(
			approvalPath,
		)
	if err != nil {
		_ = authStore.Close()

		return nil, fmt.Errorf(
			"open approval store: %w",
			err,
		)
	}

	auditStore, err :=
		audit.NewStore(
			auditPath,
		)
	if err != nil {
		_ = approvalStore.Close()
		_ = authStore.Close()

		return nil, fmt.Errorf(
			"open audit store: %w",
			err,
		)
	}

	if err := auditStore.EnsureToolEvents(); err != nil {
		_ = auditStore.Close()
		_ = approvalStore.Close()
		_ = authStore.Close()

		return nil, err
	}

	if err := auditStore.EnsureAuthEvents(); err != nil {
		_ = auditStore.Close()
		_ = approvalStore.Close()
		_ = authStore.Close()

		return nil, err
	}

	if err := auditStore.EnsureRAGEvents(); err != nil {
		_ = auditStore.Close()
		_ = approvalStore.Close()
		_ = authStore.Close()

		return nil, err
	}

	if err := auditStore.EnsureAbuseEvents(); err != nil {
		_ = auditStore.Close()
		_ = approvalStore.Close()
		_ = authStore.Close()

		return nil, err
	}

	abuseStore, err :=
		abuse.NewStore(
			abusePath,
		)
	if err != nil {
		_ = auditStore.Close()
		_ = approvalStore.Close()
		_ = authStore.Close()

		return nil, fmt.Errorf(
			"open abuse store: %w",
			err,
		)
	}

	log.Println("primary persistence: sqlite")

	return &Dependencies{
		Auth:      authStore,
		Approvals: approvalStore,
		Audit:     auditStore,
		Abuse:     abuseStore,

		close: func() {
			_ = abuseStore.Close()
			_ = auditStore.Close()
			_ = approvalStore.Close()
			_ = authStore.Close()
		},
	}, nil
}

func envOrDefault(
	name string,
	fallback string,
) string {
	value := os.Getenv(name)

	if value == "" {
		return fallback
	}

	return value
}

func (d *Dependencies) Close() {
	if d == nil || d.close == nil {
		return
	}

	d.close()
}
