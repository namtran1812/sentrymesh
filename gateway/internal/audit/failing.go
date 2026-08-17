package audit

import (
	"context"
	"errors"
)

var ErrInjectedAuditFailure = errors.New("injected audit persistence failure")

type FailingRepository struct {
	Repository

	FailWrites bool
}

func NewFailingRepository(
	repository Repository,
	failWrites bool,
) Repository {
	return &FailingRepository{
		Repository: repository,
		FailWrites: failWrites,
	}
}

func (f *FailingRepository) Write(
	ctx context.Context,
	event Event,
) error {
	if f.FailWrites {
		return ErrInjectedAuditFailure
	}

	return f.Repository.Write(
		ctx,
		event,
	)
}

func (f *FailingRepository) WriteBatch(
	ctx context.Context,
	events []Event,
) error {
	if f.FailWrites {
		return ErrInjectedAuditFailure
	}

	if writer, ok :=
		f.Repository.(BatchWriter); ok {
		return writer.WriteBatch(
			ctx,
			events,
		)
	}

	return writeBatch(
		ctx,
		f.Repository,
		events,
	)
}
