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

	content := "Mock model response to: " + last.Content

	if last.Content == "TEST_OUTPUT_PII" {
		content = "Contact alice@example.com for assistance."
	}

	if last.Content == "TEST_OUTPUT_SECRET" {
		content = "Leaked key: AKIA1234567890ABCDEF"
	}

	return Response{
		ID:      "mock_completion",
		Model:   req.Model,
		Content: content,
	}, nil
}
