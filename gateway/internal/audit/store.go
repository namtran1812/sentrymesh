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
	LatencyUS         int64
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
			latency_us INTEGER NOT NULL DEFAULT 0,
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

	if err != nil {
		return err
	}

	rows, err := s.db.Query(
		`PRAGMA table_info(audit_events)`,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	hasLatencyUS := false

	for rows.Next() {
		var cid int
		var name string
		var dataType string
		var notNull int
		var defaultValue any
		var primaryKey int

		if err := rows.Scan(
			&cid,
			&name,
			&dataType,
			&notNull,
			&defaultValue,
			&primaryKey,
		); err != nil {
			return err
		}

		if name == "latency_us" {
			hasLatencyUS = true
		}
	}

	if err := rows.Err(); err != nil {
		return err
	}

	if !hasLatencyUS {
		_, err = s.db.Exec(
			`ALTER TABLE audit_events
			 ADD COLUMN latency_us INTEGER NOT NULL DEFAULT 0`,
		)
		if err != nil {
			return err
		}
	}

	return nil
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
			latency_us,
			secret_findings,
			pii_findings,
			injection_findings,
			output_findings
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`,
		event.RequestID,
		event.Timestamp.UTC().Format(time.RFC3339Nano),
		event.Provider,
		event.Model,
		event.Decision,
		event.RiskScore,
		event.Severity,
		event.LatencyMS,
		event.LatencyUS,
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
