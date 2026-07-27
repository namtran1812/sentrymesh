package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestChatBlocksAWSKey(t *testing.T) {
	body := `{
		"model":"test-model",
		"messages":[
			{
				"role":"user",
				"content":"my key is AKIA1234567890ABCDEF"
			}
		]
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/v1/chat/completions",
		strings.NewReader(body),
	)

	rec := httptest.NewRecorder()

	ChatHandler(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}

	if !strings.Contains(rec.Body.String(), `"decision":"BLOCK"`) {
		t.Fatalf("unexpected response: %s", rec.Body.String())
	}
}

func TestChatAllowsBenignRequest(t *testing.T) {
	body := `{
		"model":"test-model",
		"messages":[
			{
				"role":"user",
				"content":"summarize this document"
			}
		]
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/v1/chat/completions",
		strings.NewReader(body),
	)

	rec := httptest.NewRecorder()

	ChatHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	if !strings.Contains(rec.Body.String(), `"decision":"ALLOW"`) {
		t.Fatalf("unexpected response: %s", rec.Body.String())
	}
}
