package approval

import "context"

type Repository interface {
	Create(
		ctx context.Context,
		tool string,
		arguments any,
		risk int,
		reason string,
	) (Request, error)

	ListPending(ctx context.Context) ([]Request, error)
	ListActive(ctx context.Context) ([]Request, error)

	Get(
		ctx context.Context,
		id int64,
	) (Request, error)

	SetStatus(
		ctx context.Context,
		id int64,
		status Status,
	) error

	ClaimExecution(
		ctx context.Context,
		id int64,
	) (bool, error)

	FinishExecution(
		ctx context.Context,
		id int64,
	) error

	FailExecution(
		ctx context.Context,
		id int64,
	) error

	Close() error
}

var _ Repository = (*Store)(nil)
