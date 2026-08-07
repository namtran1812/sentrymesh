package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type AbuseEvent struct {
	Timestamp time.Time
	EventType string
	KeyID     int64
	KeyName   string
	UserID    string
	Role      string
	Team      string
	Path      string
	Details   any
}

type AbuseEventRecord struct {
	ID        int64           `json:"id"`
	Timestamp string          `json:"timestamp"`
	EventType string          `json:"event_type"`
	KeyID     int64           `json:"key_id"`
	KeyName   string          `json:"key_name"`
	UserID    string          `json:"user_id"`
	Role      string          `json:"role"`
	Team      string          `json:"team"`
	Path      string          `json:"path"`
	Details   json.RawMessage `json:"details"`
}

func (s *Store) EnsureAbuseEvents() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS abuse_events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp TEXT NOT NULL,
			event_type TEXT NOT NULL,
			key_id INTEGER NOT NULL,
			key_name TEXT NOT NULL,
			user_id TEXT NOT NULL,
			role TEXT NOT NULL,
			team TEXT NOT NULL,
			path TEXT NOT NULL,
			details TEXT NOT NULL
		);

		CREATE INDEX IF NOT EXISTS idx_abuse_events_key
		ON abuse_events(key_id);

		CREATE INDEX IF NOT EXISTS idx_abuse_events_timestamp
		ON abuse_events(timestamp);
	`)

	return err
}

func (s *Store) WriteAbuseEvent(
	ctx context.Context,
	event AbuseEvent,
) error {
	details, err := json.Marshal(event.Details)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(
		ctx,
		`
		INSERT INTO abuse_events (
			timestamp,
			event_type,
			key_id,
			key_name,
			user_id,
			role,
			team,
			path,
			details
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`,
		event.Timestamp.UTC().Format(time.RFC3339Nano),
		event.EventType,
		event.KeyID,
		event.KeyName,
		event.UserID,
		event.Role,
		event.Team,
		event.Path,
		string(details),
	)

	if err != nil {
		return fmt.Errorf("write abuse event: %w", err)
	}

	return nil
}

func (s *Store) ListAbuseEvents(
	ctx context.Context,
	limit int,
) ([]AbuseEventRecord, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	rows, err := s.db.QueryContext(
		ctx,
		`
		SELECT
			id,
			timestamp,
			event_type,
			key_id,
			key_name,
			user_id,
			role,
			team,
			path,
			details
		FROM abuse_events
		ORDER BY id DESC
		LIMIT ?
		`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]AbuseEventRecord, 0)

	for rows.Next() {
		var item AbuseEventRecord
		var details string

		if err := rows.Scan(
			&item.ID,
			&item.Timestamp,
			&item.EventType,
			&item.KeyID,
			&item.KeyName,
			&item.UserID,
			&item.Role,
			&item.Team,
			&item.Path,
			&details,
		); err != nil {
			return nil, err
		}

		item.Details = json.RawMessage(details)

		results = append(results, item)
	}

	return results, rows.Err()
}
