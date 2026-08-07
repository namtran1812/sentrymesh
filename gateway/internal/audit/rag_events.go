package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/namtran1812/sentrymesh/gateway/internal/rag"
)

type RAGEvent struct {
	RequestID string
	Timestamp time.Time
	UserID    string
	Role      string
	Team      string
	Trace     rag.ContextTrace
}

type RAGEventRecord struct {
	ID        int64           `json:"id"`
	RequestID string          `json:"request_id"`
	Timestamp string          `json:"timestamp"`
	UserID    string          `json:"user_id"`
	Role      string          `json:"role"`
	Team      string          `json:"team"`
	Trace     json.RawMessage `json:"trace"`
}

func (s *Store) EnsureRAGEvents() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS rag_events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			request_id TEXT NOT NULL,
			timestamp TEXT NOT NULL,
			user_id TEXT NOT NULL,
			role TEXT NOT NULL,
			team TEXT NOT NULL,
			trace TEXT NOT NULL
		);

		CREATE INDEX IF NOT EXISTS idx_rag_request_id
		ON rag_events(request_id);

		CREATE INDEX IF NOT EXISTS idx_rag_timestamp
		ON rag_events(timestamp);
	`)

	return err
}

func (s *Store) WriteRAGEvent(
	ctx context.Context,
	event RAGEvent,
) error {
	trace, err := json.Marshal(event.Trace)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(
		ctx,
		`
		INSERT INTO rag_events (
			request_id,
			timestamp,
			user_id,
			role,
			team,
			trace
		)
		VALUES (?, ?, ?, ?, ?, ?)
		`,
		event.RequestID,
		event.Timestamp.UTC().Format(time.RFC3339Nano),
		event.UserID,
		event.Role,
		event.Team,
		string(trace),
	)

	if err != nil {
		return fmt.Errorf("write RAG event: %w", err)
	}

	return nil
}

func (s *Store) GetRAGEvents(
	ctx context.Context,
	requestID string,
) ([]RAGEventRecord, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`
		SELECT
			id,
			request_id,
			timestamp,
			user_id,
			role,
			team,
			trace
		FROM rag_events
		WHERE request_id = ?
		ORDER BY id ASC
		`,
		requestID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]RAGEventRecord, 0)

	for rows.Next() {
		var item RAGEventRecord
		var trace string

		if err := rows.Scan(
			&item.ID,
			&item.RequestID,
			&item.Timestamp,
			&item.UserID,
			&item.Role,
			&item.Team,
			&trace,
		); err != nil {
			return nil, err
		}

		item.Trace = json.RawMessage(trace)
		results = append(results, item)
	}

	return results, rows.Err()
}
