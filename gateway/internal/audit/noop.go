package audit

import "context"

type NoopRepository struct {
	Repository
}

func NewNoopRepository(
	fallback Repository,
) Repository {
	return &NoopRepository{
		Repository: fallback,
	}
}

func (n *NoopRepository) Write(
	context.Context,
	Event,
) error {
	return nil
}
