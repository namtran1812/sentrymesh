package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type AuthEvent struct {
	Timestamp time.Time
	EventType string
	KeyID     int64
	KeyName   string
	Actor     any
	Details   any
}

type AuthEventRecord struct {
	ID        int64           `json:"id"`
	Timestamp string          `json:"timestamp"`
	EventType string          `json:"event_type"`
	KeyID     int64           `json:"key_id"`
	KeyName   string          `json:"key_name"`
	Actor     json.RawMessage `json:"actor"`
	Details   json.RawMessage `json:"details"`
}

func (s *Store) EnsureAuthEvents() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS auth_events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp TEXT NOT NULL,
			event_type TEXT NOT NULL,
			key_id INTEGER NOT NULL,
			key_name TEXT NOT NULL,
			actor TEXT NOT NULL,
			details TEXT NOT NULL
		);

		CREATE INDEX IF NOT EXISTS idx_auth_events_timestamp
		ON auth_events(timestamp);

		CREATE INDEX IF NOT EXISTS idx_auth_events_key_id
		ON auth_events(key_id);
	`)

	return err
}

func (s *Store) WriteAuthEvent(
	ctx context.Context,
	event AuthEvent,
) error {
	actor, err := json.Marshal(event.Actor)
	if err != nil {
		return err
	}

	details, err := json.Marshal(event.Details)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(
		ctx,
		`
		INSERT INTO auth_events (
			timestamp,
			event_type,
			key_id,
			key_name,
			actor,
			details
		)
		VALUES (?, ?, ?, ?, ?, ?)
		`,
		event.Timestamp.UTC().Format(time.RFC3339Nano),
		event.EventType,
		event.KeyID,
		event.KeyName,
		string(actor),
		string(details),
	)
	if err != nil {
		return fmt.Errorf("write auth event: %w", err)
	}

	return nil
}

func (s *Store) ListAuthEvents(
	ctx context.Context,
	limit int,
) ([]AuthEventRecord, error) {
	if limit <= 0 || limit > 100 {
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
			actor,
			details
		FROM auth_events
		ORDER BY id DESC
		LIMIT ?
		`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]AuthEventRecord, 0)

	for rows.Next() {
		var item AuthEventRecord
		var actor string
		var details string

		if err := rows.Scan(
			&item.ID,
			&item.Timestamp,
			&item.EventType,
			&item.KeyID,
			&item.KeyName,
			&actor,
			&details,
		); err != nil {
			return nil, err
		}

		item.Actor = json.RawMessage(actor)
		item.Details = json.RawMessage(details)

		results = append(results, item)
	}

	return results, rows.Err()
}
