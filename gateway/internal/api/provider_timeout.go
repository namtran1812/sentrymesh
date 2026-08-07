package api

import (
	"context"
	"errors"
	"net/http"
	"time"
)

const providerTimeout = 20 * time.Second

func providerContext(
	r *http.Request,
) (
	context.Context,
	context.CancelFunc,
) {
	return context.WithTimeout(
		r.Context(),
		providerTimeout,
	)
}

func isProviderTimeout(
	ctx context.Context,
	err error,
) bool {
	return errors.Is(
		ctx.Err(),
		context.DeadlineExceeded,
	) || errors.Is(
		err,
		context.DeadlineExceeded,
	)
}
