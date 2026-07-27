package provider

import (
	"context"
	"testing"
)

func TestRouterUsesRegisteredProvider(t *testing.T) {
	router := NewRouter()
	router.Register("mock", NewMockProvider())

	res, err := router.Chat(
		context.Background(),
		"mock",
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

	if res.Content == "" {
		t.Fatal("expected non-empty provider response")
	}
}

func TestRouterRejectsUnknownProvider(t *testing.T) {
	router := NewRouter()

	_, err := router.Chat(
		context.Background(),
		"missing",
		Request{},
	)

	if err == nil {
		t.Fatal("expected error for unknown provider")
	}
}
