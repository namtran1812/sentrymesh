package provider

import (
	"context"
	"fmt"
)

type MockProvider struct{}

func NewMockProvider() *MockProvider {
	return &MockProvider{}
}

func (p *MockProvider) Chat(
	ctx context.Context,
	req Request,
) (Response, error) {
	select {
	case <-ctx.Done():
		return Response{}, ctx.Err()
	default:
	}

	if len(req.Messages) == 0 {
		return Response{}, fmt.Errorf("messages cannot be empty")
	}

	last := req.Messages[len(req.Messages)-1]

	return Response{
		ID:      "mock_completion",
		Model:   req.Model,
		Content: "Mock model response to: " + last.Content,
	}, nil
}
