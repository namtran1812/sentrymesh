package audit

import (
	"context"
	"errors"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestAsyncRepositoryDrainsOnClose(
	t *testing.T,
) {
	store, err := NewStore(
		filepath.Join(
			t.TempDir(),
			"audit.db",
		),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	async, err :=
		NewAsyncRepository(
			store,
			AsyncOptions{
				QueueSize:     32,
				BatchSize:     4,
				FlushInterval: time.Hour,
				WriteTimeout:  time.Second,
			},
		)
	if err != nil {
		t.Fatal(err)
	}

	const total = 10

	for i := 0; i < total; i++ {
		err := async.Write(
			context.Background(),
			Event{
				RequestID: "async-test-" +
					string(rune('a'+i)),
				Timestamp: time.Now(),
				Provider:  "mock",
				Model:     "test",
				Decision:  "ALLOW",
				Severity:  "LOW",
			},
		)
		if err != nil {
			t.Fatal(err)
		}
	}

	ctx, cancel :=
		context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
	defer cancel()

	if err := async.Close(ctx); err != nil {
		t.Fatal(err)
	}

	records, err :=
		store.List(
			context.Background(),
			20,
		)
	if err != nil {
		t.Fatal(err)
	}

	if len(records) != total {
		t.Fatalf(
			"expected %d persisted events, got %d",
			total,
			len(records),
		)
	}
}

func TestAsyncRepositoryRejectsAfterClose(
	t *testing.T,
) {
	store, err := NewStore(
		filepath.Join(
			t.TempDir(),
			"audit.db",
		),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	options := DefaultAsyncOptions()

	async, err :=
		NewAsyncRepository(
			store,
			options,
		)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel :=
		context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
	defer cancel()

	if err := async.Close(ctx); err != nil {
		t.Fatal(err)
	}

	err = async.Write(
		context.Background(),
		Event{},
	)

	if !errors.Is(
		err,
		ErrAsyncClosed,
	) {
		t.Fatalf(
			"expected ErrAsyncClosed, got %v",
			err,
		)
	}
}

type blockingRepository struct {
	Repository

	started chan struct{}
	release chan struct{}
	once    sync.Once
}

func (b *blockingRepository) Write(
	ctx context.Context,
	event Event,
) error {
	b.once.Do(func() {
		close(b.started)
	})

	select {
	case <-b.release:
		return nil

	case <-ctx.Done():
		return ctx.Err()
	}
}

func TestAsyncRepositoryAppliesBackpressure(
	t *testing.T,
) {
	base, err := NewStore(
		filepath.Join(
			t.TempDir(),
			"audit.db",
		),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer base.Close()

	blocker := &blockingRepository{
		Repository: base,
		started:    make(chan struct{}),
		release:    make(chan struct{}),
	}

	async, err := NewAsyncRepository(
		blocker,
		AsyncOptions{
			QueueSize:     1,
			BatchSize:     1,
			FlushInterval: time.Hour,
			WriteTimeout:  5 * time.Second,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	// First event is consumed by the worker and blocks in persistence.
	if err := async.Write(
		context.Background(),
		Event{
			RequestID: "event-1",
			Timestamp: time.Now(),
			Decision:  "ALLOW",
			Severity:  "LOW",
		},
	); err != nil {
		t.Fatal(err)
	}

	select {
	case <-blocker.started:
	case <-time.After(time.Second):
		t.Fatal("writer did not begin persistence")
	}

	// Second event occupies the single queue slot.
	if err := async.Write(
		context.Background(),
		Event{
			RequestID: "event-2",
			Timestamp: time.Now(),
			Decision:  "ALLOW",
			Severity:  "LOW",
		},
	); err != nil {
		t.Fatal(err)
	}

	writeDone := make(chan error, 1)

	// Third event must wait because the worker is blocked and the queue is full.
	go func() {
		writeDone <- async.Write(
			context.Background(),
			Event{
				RequestID: "event-3",
				Timestamp: time.Now(),
				Decision:  "ALLOW",
				Severity:  "LOW",
			},
		)
	}()

	select {
	case err := <-writeDone:
		t.Fatalf(
			"expected backpressure, write completed early: %v",
			err,
		)

	case <-time.After(50 * time.Millisecond):
		// Expected: producer remains blocked.
	}

	stats := async.AsyncStats()

	if stats.Saturated == 0 {
		t.Fatal(
			"expected saturation counter to increment",
		)
	}

	close(blocker.release)

	select {
	case err := <-writeDone:
		if err != nil {
			t.Fatal(err)
		}

	case <-time.After(2 * time.Second):
		t.Fatal(
			"blocked producer did not resume",
		)
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	if err := async.Close(ctx); err != nil {
		t.Fatal(err)
	}

	stats = async.AsyncStats()

	if stats.Enqueued != 3 {
		t.Fatalf(
			"expected 3 enqueued events, got %d",
			stats.Enqueued,
		)
	}

	if stats.Flushed != 3 {
		t.Fatalf(
			"expected 3 flushed events, got %d",
			stats.Flushed,
		)
	}
}

func TestAsyncRepositoryBackpressureHonorsContext(
	t *testing.T,
) {
	base, err := NewStore(
		filepath.Join(
			t.TempDir(),
			"audit.db",
		),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer base.Close()

	blocker := &blockingRepository{
		Repository: base,
		started:    make(chan struct{}),
		release:    make(chan struct{}),
	}

	async, err := NewAsyncRepository(
		blocker,
		AsyncOptions{
			QueueSize:     1,
			BatchSize:     1,
			FlushInterval: time.Hour,
			WriteTimeout:  time.Second,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if err := async.Write(
		context.Background(),
		Event{
			RequestID: "event-1",
			Timestamp: time.Now(),
		},
	); err != nil {
		t.Fatal(err)
	}

	select {
	case <-blocker.started:
	case <-time.After(time.Second):
		t.Fatal("writer did not start")
	}

	if err := async.Write(
		context.Background(),
		Event{
			RequestID: "event-2",
			Timestamp: time.Now(),
		},
	); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		50*time.Millisecond,
	)
	defer cancel()

	err = async.Write(
		ctx,
		Event{
			RequestID: "event-3",
			Timestamp: time.Now(),
		},
	)

	if !errors.Is(
		err,
		context.DeadlineExceeded,
	) {
		t.Fatalf(
			"expected context deadline, got %v",
			err,
		)
	}

	close(blocker.release)

	closeCtx, closeCancel :=
		context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
	defer closeCancel()

	if err := async.Close(
		closeCtx,
	); err != nil {
		t.Fatal(err)
	}

	stats := async.AsyncStats()

	if stats.Saturated == 0 {
		t.Fatal(
			"expected saturation to be observed",
		)
	}

	// Only the first two requests were accepted.
	if stats.Enqueued != 2 {
		t.Fatalf(
			"expected 2 accepted events, got %d",
			stats.Enqueued,
		)
	}
}
