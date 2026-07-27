package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

type Event struct {
	RequestID         string
	Timestamp         time.Time
	Provider          string
	Model             string
	Decision          string
	RiskScore         int
	Severity          string
	LatencyMS         int64
	SecretFindings    any
	PIIFindings       any
	InjectionFindings any
	OutputFindings    any
}

type Store struct {
	db *sql.DB
}

func NewStore(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	store := &Store{db: db}

	if err := store.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return store, nil
}

func (s *Store) migrate() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS audit_events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			request_id TEXT NOT NULL,
			timestamp TEXT NOT NULL,
			provider TEXT NOT NULL,
			model TEXT NOT NULL,
			decision TEXT NOT NULL,
			risk_score INTEGER NOT NULL,
			severity TEXT NOT NULL,
			latency_ms INTEGER NOT NULL,
			secret_findings TEXT NOT NULL,
			pii_findings TEXT NOT NULL,
			injection_findings TEXT NOT NULL,
			output_findings TEXT NOT NULL
		);

		CREATE INDEX IF NOT EXISTS idx_audit_request_id
		ON audit_events(request_id);

		CREATE INDEX IF NOT EXISTS idx_audit_timestamp
		ON audit_events(timestamp);
	`)

	return err
}

func marshal(v any) string {
	if v == nil {
		return "null"
	}

	data, err := json.Marshal(v)
	if err != nil {
		return "null"
	}

	return string(data)
}

func (s *Store) Write(ctx context.Context, event Event) error {
	_, err := s.db.ExecContext(
		ctx,
		`
		INSERT INTO audit_events (
			request_id,
			timestamp,
			provider,
			model,
			decision,
			risk_score,
			severity,
			latency_ms,
			secret_findings,
			pii_findings,
			injection_findings,
			output_findings
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`,
		event.RequestID,
		event.Timestamp.UTC().Format(time.RFC3339Nano),
		event.Provider,
		event.Model,
		event.Decision,
		event.RiskScore,
		event.Severity,
		event.LatencyMS,
		marshal(event.SecretFindings),
		marshal(event.PIIFindings),
		marshal(event.InjectionFindings),
		marshal(event.OutputFindings),
	)

	if err != nil {
		return fmt.Errorf("write audit event: %w", err)
	}

	return nil
}

func (s *Store) Close() error {
	return s.db.Close()
}
