package auth

import (
	"context"

	"github.com/namtran1812/sentrymesh/gateway/internal/identity"
)

func Resolve(
	ctx context.Context,
	apiKey string,
) (identity.Identity, error) {
	return DefaultStore.Resolve(ctx, apiKey)
}
