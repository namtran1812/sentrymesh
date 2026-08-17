package approval

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

var _ Repository = (*PostgresStore)(nil)

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

func (s *PostgresStore) Create(
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

	var (
		id        int64
		createdAt time.Time
	)

	err = s.pool.QueryRow(
		ctx,
		`
		INSERT INTO approvals (
			tool,
			arguments,
			risk,
			reason,
			status,
			created_at
		)
		VALUES ($1, $2::jsonb, $3, $4, $5, NOW())
		RETURNING id, created_at
		`,
		tool,
		string(data),
		risk,
		reason,
		Pending,
	).Scan(
		&id,
		&createdAt,
	)

	if err != nil {
		return Request{}, fmt.Errorf("create approval: %w", err)
	}

	return Request{
		ID:        id,
		CreatedAt: createdAt.UTC().Format(time.RFC3339Nano),
		Tool:      tool,
		Arguments: data,
		Risk:      risk,
		Reason:    reason,
		Status:    Pending,
	}, nil
}

func (s *PostgresStore) ListPending(
	ctx context.Context,
) ([]Request, error) {
	return s.list(
		ctx,
		`
		WHERE status = $1
		ORDER BY id DESC
		`,
		Pending,
	)
}

func (s *PostgresStore) ListActive(
	ctx context.Context,
) ([]Request, error) {
	return s.list(
		ctx,
		`
		WHERE status IN ($1, $2, $3)
		  AND executed_at IS NULL
		ORDER BY id DESC
		`,
		Pending,
		Approved,
		"EXECUTING",
	)
}

func (s *PostgresStore) list(
	ctx context.Context,
	where string,
	args ...any,
) ([]Request, error) {
	rows, err := s.pool.Query(
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
		`+where,
		args...,
	)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]Request, 0)

	for rows.Next() {
		item, err := scanRequest(rows)
		if err != nil {
			return nil, err
		}

		results = append(results, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (s *PostgresStore) Get(
	ctx context.Context,
	id int64,
) (Request, error) {
	row := s.pool.QueryRow(
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
		WHERE id = $1
		`,
		id,
	)

	item, err := scanRequest(row)

	if errors.Is(err, pgx.ErrNoRows) {
		return Request{}, fmt.Errorf("get approval: not found")
	}

	if err != nil {
		return Request{}, fmt.Errorf("get approval: %w", err)
	}

	return item, nil
}

func (s *PostgresStore) SetStatus(
	ctx context.Context,
	id int64,
	status Status,
) error {
	tag, err := s.pool.Exec(
		ctx,
		`
		UPDATE approvals
		SET status = $1
		WHERE id = $2
		  AND status = $3
		`,
		status,
		id,
		Pending,
	)

	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("approval not found or already resolved")
	}

	return nil
}

func (s *PostgresStore) ClaimExecution(
	ctx context.Context,
	id int64,
) (bool, error) {
	tag, err := s.pool.Exec(
		ctx,
		`
		UPDATE approvals
		SET status = $1
		WHERE id = $2
		  AND status = $3
		  AND executed_at IS NULL
		`,
		"EXECUTING",
		id,
		Approved,
	)

	if err != nil {
		return false, err
	}

	return tag.RowsAffected() == 1, nil
}

func (s *PostgresStore) FinishExecution(
	ctx context.Context,
	id int64,
) error {
	tag, err := s.pool.Exec(
		ctx,
		`
		UPDATE approvals
		SET
			status = $1,
			executed_at = NOW()
		WHERE id = $2
		  AND status = $3
		`,
		"EXECUTED",
		id,
		"EXECUTING",
	)

	if err != nil {
		return err
	}

	if tag.RowsAffected() != 1 {
		return fmt.Errorf("approval is not executing")
	}

	return nil
}

func (s *PostgresStore) FailExecution(
	ctx context.Context,
	id int64,
) error {
	tag, err := s.pool.Exec(
		ctx,
		`
		UPDATE approvals
		SET status = $1
		WHERE id = $2
		  AND status = $3
		`,
		Approved,
		id,
		"EXECUTING",
	)

	if err != nil {
		return err
	}

	if tag.RowsAffected() != 1 {
		return fmt.Errorf("approval is not executing")
	}

	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanRequest(row rowScanner) (Request, error) {
	var (
		item       Request
		createdAt  time.Time
		executedAt *time.Time
		arguments  []byte
	)

	err := row.Scan(
		&item.ID,
		&createdAt,
		&item.Tool,
		&arguments,
		&item.Risk,
		&item.Reason,
		&item.Status,
		&executedAt,
	)

	if err != nil {
		return Request{}, err
	}

	item.CreatedAt = createdAt.UTC().Format(time.RFC3339Nano)
	item.Arguments = json.RawMessage(arguments)

	if executedAt != nil {
		value := executedAt.UTC().Format(time.RFC3339Nano)
		item.ExecutedAt = &value
	}

	return item, nil
}

func (s *PostgresStore) Close() error {
	return nil
}
