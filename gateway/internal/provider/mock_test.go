package provider

import (
	"context"
	"strings"
	"testing"
)

func TestMockProvider(t *testing.T) {
	p := NewMockProvider()

	res, err := p.Chat(
		context.Background(),
		Request{
			Model: "test-model",
			Messages: []Message{
				{
					Role:    "user",
					Content: "hello",
				},
			},
		},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(res.Content, "hello") {
		t.Fatalf("unexpected response: %s", res.Content)
	}
}
