package audit

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestWriteAuditEvent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.db")

	store, err := NewStore(path)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	defer store.Close()

	err = store.Write(
		context.Background(),
		Event{
			RequestID: "req_test",
			Timestamp: time.Now(),
			Provider:  "mock",
			Model:     "test",
			Decision:  "ALLOW",
			RiskScore: 0,
			Severity:  "LOW",
			LatencyMS: 5,
		},
	)

	if err != nil {
		t.Fatalf("write event: %v", err)
	}
}
