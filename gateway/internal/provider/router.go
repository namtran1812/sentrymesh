package provider

import (
	"context"
	"fmt"
)

type Router struct {
	providers map[string]Provider
}

func NewRouter() *Router {
	return &Router{
		providers: map[string]Provider{},
	}
}

func (r *Router) Register(name string, p Provider) {
	r.providers[name] = p
}

func (r *Router) Chat(
	ctx context.Context,
	providerName string,
	req Request,
) (Response, error) {
	p, ok := r.providers[providerName]
	if !ok {
		return Response{}, fmt.Errorf("unknown provider: %s", providerName)
	}

	return p.Chat(ctx, req)
}
