package audit

import (
	"context"
	"errors"
	"path/filepath"
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
