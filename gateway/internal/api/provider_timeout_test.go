package api

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"
	"time"
)

func TestProviderContextTimesOut(
	t *testing.T,
) {
	req := httptest.NewRequest(
		"POST",
		"/test",
		nil,
	)

	ctx, cancel := context.WithTimeout(
		req.Context(),
		time.Millisecond,
	)
	defer cancel()

	<-ctx.Done()

	if !errors.Is(
		ctx.Err(),
		context.DeadlineExceeded,
	) {
		t.Fatalf(
			"expected deadline exceeded, got %v",
			ctx.Err(),
		)
	}
}

func TestIsProviderTimeout(
	t *testing.T,
) {
	ctx, cancel := context.WithCancel(
		context.Background(),
	)

	cancel()

	if isProviderTimeout(
		ctx,
		context.DeadlineExceeded,
	) != true {
		t.Fatal(
			"expected deadline exceeded to be recognized",
		)
	}
}
