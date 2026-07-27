package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type OllamaProvider struct {
	baseURL    string
	httpClient *http.Client
}

type ollamaRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type ollamaResponse struct {
	Model   string `json:"model"`
	Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message"`
}

func NewOllamaProvider() *OllamaProvider {
	return &OllamaProvider{
		baseURL: "http://localhost:11434",
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

func (p *OllamaProvider) Chat(
	ctx context.Context,
	req Request,
) (Response, error) {
	payload := ollamaRequest{
		Model:    req.Model,
		Messages: req.Messages,
		Stream:   false,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return Response{}, fmt.Errorf("marshal ollama request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		p.baseURL+"/api/chat",
		bytes.NewReader(body),
	)
	if err != nil {
		return Response{}, fmt.Errorf("create ollama request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	httpRes, err := p.httpClient.Do(httpReq)
	if err != nil {
		return Response{}, fmt.Errorf("ollama request failed: %w", err)
	}
	defer httpRes.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(httpRes.Body, 2<<20))
	if err != nil {
		return Response{}, fmt.Errorf("read ollama response: %w", err)
	}

	if httpRes.StatusCode < 200 || httpRes.StatusCode >= 300 {
		return Response{}, fmt.Errorf(
			"ollama returned status %d: %s",
			httpRes.StatusCode,
			string(responseBody),
		)
	}

	var decoded ollamaResponse

	if err := json.Unmarshal(responseBody, &decoded); err != nil {
		return Response{}, fmt.Errorf("decode ollama response: %w", err)
	}

	if decoded.Message.Content == "" {
		return Response{}, fmt.Errorf("ollama returned empty content")
	}

	return Response{
		ID:      "ollama_completion",
		Model:   decoded.Model,
		Content: decoded.Message.Content,
	}, nil
}
