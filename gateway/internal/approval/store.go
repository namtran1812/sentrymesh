package approval

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

type Status string

const (
	Pending  Status = "PENDING"
	Approved Status = "APPROVED"
	Rejected Status = "REJECTED"
)

type Request struct {
	ExecutedAt *string         `json:"executed_at,omitempty"`
	ID         int64           `json:"id"`
	CreatedAt  string          `json:"created_at"`
	Tool       string          `json:"tool"`
	Arguments  json.RawMessage `json:"arguments"`
	Risk       int             `json:"risk"`
	Reason     string          `json:"reason"`
	Status     Status          `json:"status"`
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
		CREATE TABLE IF NOT EXISTS approvals (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			created_at TEXT NOT NULL,
			tool TEXT NOT NULL,
			arguments TEXT NOT NULL,
			risk INTEGER NOT NULL,
			reason TEXT NOT NULL,
			status TEXT NOT NULL,
			executed_at TEXT
		);
	`)

	return err
}

func (s *Store) Create(
	ctx context.Context,
	tool string,
	arguments any,
	risk int,
	reason string,
) (Request, error) {
	data, err := json.Marshal(arguments)
	if err != nil {
		return Request{}, err
	}

	createdAt := time.Now().UTC().Format(time.RFC3339Nano)

	result, err := s.db.ExecContext(
		ctx,
		`
		INSERT INTO approvals (
			created_at,
			tool,
			arguments,
			risk,
			reason,
			status
		)
		VALUES (?, ?, ?, ?, ?, ?)
		`,
		createdAt,
		tool,
		string(data),
		risk,
		reason,
		Pending,
	)
	if err != nil {
		return Request{}, fmt.Errorf("create approval: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return Request{}, err
	}

	return Request{
		ID:        id,
		CreatedAt: createdAt,
		Tool:      tool,
		Arguments: data,
		Risk:      risk,
		Reason:    reason,
		Status:    Pending,
	}, nil
}

func (s *Store) ListPending(ctx context.Context) ([]Request, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`
		SELECT
			id,
			created_at,
			tool,
			arguments,
			risk,
			reason,
			status,
			executed_at
		FROM approvals
		WHERE status = ?
		ORDER BY id DESC
		`,
		Pending,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]Request, 0)

	for rows.Next() {
		var item Request
		var arguments string

		if err := rows.Scan(
			&item.ID,
			&item.CreatedAt,
			&item.Tool,
			&arguments,
			&item.Risk,
			&item.Reason,
			&item.Status,
		); err != nil {
			return nil, err
		}

		item.Arguments = json.RawMessage(arguments)
		results = append(results, item)
	}

	return results, rows.Err()
}

func (s *Store) SetStatus(
	ctx context.Context,
	id int64,
	status Status,
) error {
	result, err := s.db.ExecContext(
		ctx,
		`
		UPDATE approvals
		SET status = ?
		WHERE id = ? AND status = ?
		`,
		status,
		id,
		Pending,
	)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return fmt.Errorf("approval not found or already resolved")
	}

	return nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Get(
	ctx context.Context,
	id int64,
) (Request, error) {
	var item Request
	var arguments string

	err := s.db.QueryRowContext(
		ctx,
		`
		SELECT
			id,
			created_at,
			tool,
			arguments,
			risk,
			reason,
			status
		FROM approvals
		WHERE id = ?
		`,
		id,
	).Scan(
		&item.ID,
		&item.CreatedAt,
		&item.Tool,
		&arguments,
		&item.Risk,
		&item.Reason,
		&item.Status,
		&item.ExecutedAt,
	)

	if err != nil {
		return Request{}, err
	}

	item.Arguments = json.RawMessage(arguments)

	return item, nil
}

func (s *Store) DB() *sql.DB {
	return s.db
}
