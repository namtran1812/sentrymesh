package audit

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

var ErrAsyncClosed = errors.New(
	"asynchronous audit repository closed",
)

type AsyncOptions struct {
	QueueSize     int
	BatchSize     int
	FlushInterval time.Duration
	WriteTimeout  time.Duration
}

func DefaultAsyncOptions() AsyncOptions {
	return AsyncOptions{
		QueueSize:     16_384,
		BatchSize:     128,
		FlushInterval: 10 * time.Millisecond,
		WriteTimeout:  5 * time.Second,
	}
}

type AsyncRepository struct {
	Repository

	options AsyncOptions

	queue chan Event
	stop  chan struct{}
	done  chan struct{}

	mu     sync.RWMutex
	closed bool
}

func NewAsyncRepository(
	repository Repository,
	options AsyncOptions,
) (*AsyncRepository, error) {
	if repository == nil {
		return nil, fmt.Errorf(
			"audit: nil repository",
		)
	}

	if options.QueueSize <= 0 {
		return nil, fmt.Errorf(
			"audit: queue size must be positive",
		)
	}

	if options.BatchSize <= 0 {
		return nil, fmt.Errorf(
			"audit: batch size must be positive",
		)
	}

	if options.FlushInterval <= 0 {
		return nil, fmt.Errorf(
			"audit: flush interval must be positive",
		)
	}

	if options.WriteTimeout <= 0 {
		return nil, fmt.Errorf(
			"audit: write timeout must be positive",
		)
	}

	repositoryWrapper :=
		&AsyncRepository{
			Repository: repository,
			options:    options,
			queue: make(
				chan Event,
				options.QueueSize,
			),
			stop: make(chan struct{}),
			done: make(chan struct{}),
		}

	go repositoryWrapper.run()

	return repositoryWrapper, nil
}

// Write acknowledges an event once it has entered the bounded queue.
//
// The queue deliberately blocks when full. Security audit events are never
// silently dropped to preserve backpressure semantics.
func (a *AsyncRepository) Write(
	ctx context.Context,
	event Event,
) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.closed {
		return ErrAsyncClosed
	}

	select {
	case a.queue <- event:
		return nil

	case <-ctx.Done():
		return ctx.Err()
	}
}

func (a *AsyncRepository) Pending() int {
	return len(a.queue)
}

func (a *AsyncRepository) run() {
	defer close(a.done)

	ticker :=
		time.NewTicker(
			a.options.FlushInterval,
		)
	defer ticker.Stop()

	batch := make(
		[]Event,
		0,
		a.options.BatchSize,
	)

	flush := func() {
		if len(batch) == 0 {
			return
		}

		a.persistWithRetry(batch)

		batch = batch[:0]
	}

	for {
		select {
		case event := <-a.queue:
			batch = append(
				batch,
				event,
			)

			if len(batch) >=
				a.options.BatchSize {
				flush()
			}

		case <-ticker.C:
			flush()

		case <-a.stop:
			// No new writers can enter after Close acquires the
			// exclusive lock. Drain everything already accepted.
			for {
				select {
				case event := <-a.queue:
					batch = append(
						batch,
						event,
					)

					if len(batch) >=
						a.options.BatchSize {
						flush()
					}

				default:
					flush()
					return
				}
			}
		}
	}
}

func (a *AsyncRepository) persistWithRetry(
	events []Event,
) {
	delay := 25 * time.Millisecond

	for {
		ctx, cancel :=
			context.WithTimeout(
				context.Background(),
				a.options.WriteTimeout,
			)

		err := writeBatch(
			ctx,
			a.Repository,
			events,
		)

		cancel()

		if err == nil {
			return
		}

		log.Printf(
			"audit batch write failed; retrying: events=%d err=%v",
			len(events),
			err,
		)

		time.Sleep(delay)

		if delay < time.Second {
			delay *= 2

			if delay > time.Second {
				delay = time.Second
			}
		}
	}
}

func (a *AsyncRepository) Close(
	ctx context.Context,
) error {
	a.mu.Lock()

	if !a.closed {
		a.closed = true
		close(a.stop)
	}

	a.mu.Unlock()

	select {
	case <-a.done:
		return nil

	case <-ctx.Done():
		return fmt.Errorf(
			"drain async audit queue: %w",
			ctx.Err(),
		)
	}
}
