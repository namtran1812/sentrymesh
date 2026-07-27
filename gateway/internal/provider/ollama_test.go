package provider

import "testing"

func TestNewOllamaProvider(t *testing.T) {
	p := NewOllamaProvider()

	if p == nil {
		t.Fatal("expected provider")
	}

	if p.baseURL != "http://localhost:11434" {
		t.Fatalf("unexpected baseURL: %s", p.baseURL)
	}
}
