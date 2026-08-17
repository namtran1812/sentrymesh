package abuse

import "context"

type Repository interface {
	Save(
		ctx context.Context,
		state StoredState,
	) error

	Load(
		ctx context.Context,
		keyID int64,
	) (StoredState, error)

	List(
		ctx context.Context,
	) ([]StoredState, error)

	Delete(
		ctx context.Context,
		keyID int64,
	) error
}

var _ Repository = (*Store)(nil)
