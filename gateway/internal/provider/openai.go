package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type OpenAIProvider struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

type openAIChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type openAIChatResponse struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func NewOpenAIProvider() (*OpenAIProvider, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")

	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY is not set")
	}

	baseURL := os.Getenv("OPENAI_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	return &OpenAIProvider{
		apiKey:  apiKey,
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}, nil
}

func (p *OpenAIProvider) Chat(
	ctx context.Context,
	req Request,
) (Response, error) {
	payload := openAIChatRequest{
		Model:    req.Model,
		Messages: req.Messages,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return Response{}, fmt.Errorf("marshal OpenAI request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		p.baseURL+"/chat/completions",
		bytes.NewReader(body),
	)
	if err != nil {
		return Response{}, fmt.Errorf("create OpenAI request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	httpRes, err := p.httpClient.Do(httpReq)
	if err != nil {
		return Response{}, fmt.Errorf("OpenAI request failed: %w", err)
	}
	defer httpRes.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(httpRes.Body, 2<<20))
	if err != nil {
		return Response{}, fmt.Errorf("read OpenAI response: %w", err)
	}

	if httpRes.StatusCode < 200 || httpRes.StatusCode >= 300 {
		return Response{}, fmt.Errorf(
			"OpenAI returned status %d: %s",
			httpRes.StatusCode,
			string(responseBody),
		)
	}

	var decoded openAIChatResponse

	if err := json.Unmarshal(responseBody, &decoded); err != nil {
		return Response{}, fmt.Errorf("decode OpenAI response: %w", err)
	}

	if len(decoded.Choices) == 0 {
		return Response{}, fmt.Errorf("OpenAI response contained no choices")
	}

	return Response{
		ID:      decoded.ID,
		Model:   decoded.Model,
		Content: decoded.Choices[0].Message.Content,
	}, nil
}
