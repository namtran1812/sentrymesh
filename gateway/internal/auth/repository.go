package auth

import (
	"context"
	"time"

	"github.com/namtran1812/sentrymesh/gateway/internal/identity"
)

type Repository interface {
	Create(
		ctx context.Context,
		name string,
		rawKey string,
		principal identity.Identity,
		expiresAt *time.Time,
	) error

	Resolve(
		ctx context.Context,
		rawKey string,
	) (identity.Identity, error)

	Revoke(
		ctx context.Context,
		rawKey string,
	) error

	List(
		ctx context.Context,
	) ([]KeyRecord, error)

	RevokeByID(
		ctx context.Context,
		id int64,
	) error

	FindByName(
		ctx context.Context,
		name string,
	) (KeyRecord, error)

	FindByID(
		ctx context.Context,
		id int64,
	) (KeyRecord, error)

	Close() error
}

var _ Repository = (*Store)(nil)
