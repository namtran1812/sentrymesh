package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type ToolEvent struct {
	Timestamp  time.Time
	ApprovalID int64
	EventType  string
	Tool       string
	Risk       int
	Status     string
	Details    any
}

func (s *Store) EnsureToolEvents() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS tool_events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp TEXT NOT NULL,
			approval_id INTEGER NOT NULL,
			event_type TEXT NOT NULL,
			tool TEXT NOT NULL,
			risk INTEGER NOT NULL,
			status TEXT NOT NULL,
			details TEXT NOT NULL
		);

		CREATE INDEX IF NOT EXISTS idx_tool_events_approval
		ON tool_events(approval_id);

		CREATE INDEX IF NOT EXISTS idx_tool_events_timestamp
		ON tool_events(timestamp);
	`)

	return err
}

func (s *Store) WriteToolEvent(
	ctx context.Context,
	event ToolEvent,
) error {
	data, err := json.Marshal(event.Details)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(
		ctx,
		`
		INSERT INTO tool_events (
			timestamp,
			approval_id,
			event_type,
			tool,
			risk,
			status,
			details
		)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		`,
		event.Timestamp.UTC().Format(time.RFC3339Nano),
		event.ApprovalID,
		event.EventType,
		event.Tool,
		event.Risk,
		event.Status,
		string(data),
	)

	if err != nil {
		return fmt.Errorf("write tool event: %w", err)
	}

	return nil
}

type ToolEventRecord struct {
	ID         int64           `json:"id"`
	Timestamp  string          `json:"timestamp"`
	ApprovalID int64           `json:"approval_id"`
	EventType  string          `json:"event_type"`
	Tool       string          `json:"tool"`
	Risk       int             `json:"risk"`
	Status     string          `json:"status"`
	Details    json.RawMessage `json:"details"`
}

func (s *Store) ListToolEvents(
	ctx context.Context,
	approvalID int64,
) ([]ToolEventRecord, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`
		SELECT
			id,
			timestamp,
			approval_id,
			event_type,
			tool,
			risk,
			status,
			details
		FROM tool_events
		WHERE approval_id = ?
		ORDER BY id ASC
		`,
		approvalID,
	)
	if err != nil {
		return nil, fmt.Errorf("query tool events: %w", err)
	}
	defer rows.Close()

	results := make([]ToolEventRecord, 0)

	for rows.Next() {
		var item ToolEventRecord
		var details string

		if err := rows.Scan(
			&item.ID,
			&item.Timestamp,
			&item.ApprovalID,
			&item.EventType,
			&item.Tool,
			&item.Risk,
			&item.Status,
			&details,
		); err != nil {
			return nil, err
		}

		item.Details = json.RawMessage(details)
		results = append(results, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}
