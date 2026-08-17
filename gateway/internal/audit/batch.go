package audit

import "context"

// BatchWriter is an optional optimization implemented by repositories that
// can persist multiple audit events more efficiently than individual writes.
type BatchWriter interface {
	WriteBatch(
		context.Context,
		[]Event,
	) error
}

func writeBatch(
	ctx context.Context,
	repository Repository,
	events []Event,
) error {
	if len(events) == 0 {
		return nil
	}

	if writer, ok := repository.(BatchWriter); ok {
		return writer.WriteBatch(
			ctx,
			events,
		)
	}

	for _, event := range events {
		if err := repository.Write(
			ctx,
			event,
		); err != nil {
			return err
		}
	}

	return nil
}
