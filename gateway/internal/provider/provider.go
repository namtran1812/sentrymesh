package provider

import (
	"context"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Request struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Response struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Content string `json:"content"`
}

type Provider interface {
	Chat(
		ctx context.Context,
		req Request,
	) (Response, error)
}
