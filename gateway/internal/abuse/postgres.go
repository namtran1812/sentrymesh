package abuse

import (
	"context"
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

func NewPostgresStore(
	pool *pgxpool.Pool,
) *PostgresStore {
	return &PostgresStore{
		pool: pool,
	}
}

func (s *PostgresStore) Save(
	ctx context.Context,
	state StoredState,
) error {
	_, err := s.pool.Exec(
		ctx,
		`
		INSERT INTO abuse_state (
			key_id,
			score,
			last_updated,
			cooldown_until
		)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (key_id)
		DO UPDATE SET
			score = EXCLUDED.score,
			last_updated = EXCLUDED.last_updated,
			cooldown_until = EXCLUDED.cooldown_until
		`,
		state.KeyID,
		state.Score,
		state.LastUpdated.UTC(),
		state.CooldownUntil,
	)

	if err != nil {
		return fmt.Errorf(
			"save abuse state: %w",
			err,
		)
	}

	return nil
}

func (s *PostgresStore) Load(
	ctx context.Context,
	keyID int64,
) (StoredState, error) {
	var item StoredState
	var cooldown *time.Time

	err := s.pool.QueryRow(
		ctx,
		`
		SELECT
			key_id,
			score,
			last_updated,
			cooldown_until
		FROM abuse_state
		WHERE key_id = $1
		`,
		keyID,
	).Scan(
		&item.KeyID,
		&item.Score,
		&item.LastUpdated,
		&cooldown,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return StoredState{},
			fmt.Errorf(
				"abuse state not found",
			)
	}

	if err != nil {
		return StoredState{},
			fmt.Errorf(
				"load abuse state: %w",
				err,
			)
	}

	item.LastUpdated =
		item.LastUpdated.UTC()

	if cooldown != nil {
		value := cooldown.UTC()
		item.CooldownUntil = &value
	}

	return item, nil
}

func (s *PostgresStore) List(
	ctx context.Context,
) ([]StoredState, error) {
	rows, err := s.pool.Query(
		ctx,
		`
		SELECT
			key_id,
			score,
			last_updated,
			cooldown_until
		FROM abuse_state
		ORDER BY score DESC, key_id ASC
		`,
	)
	if err != nil {
		return nil,
			fmt.Errorf(
				"list abuse state: %w",
				err,
			)
	}
	defer rows.Close()

	results :=
		make([]StoredState, 0)

	for rows.Next() {
		var item StoredState
		var cooldown *time.Time

		if err := rows.Scan(
			&item.KeyID,
			&item.Score,
			&item.LastUpdated,
			&cooldown,
		); err != nil {
			return nil,
				fmt.Errorf(
					"scan abuse state: %w",
					err,
				)
		}

		item.LastUpdated =
			item.LastUpdated.UTC()

		if cooldown != nil {
			value := cooldown.UTC()
			item.CooldownUntil =
				&value
		}

		results = append(
			results,
			item,
		)
	}

	if err := rows.Err(); err != nil {
		return nil,
			fmt.Errorf(
				"iterate abuse state: %w",
				err,
			)
	}

	return results, nil
}

func (s *PostgresStore) Delete(
	ctx context.Context,
	keyID int64,
) error {
	tag, err := s.pool.Exec(
		ctx,
		`
		DELETE FROM abuse_state
		WHERE key_id = $1
		`,
		keyID,
	)
	if err != nil {
		return fmt.Errorf(
			"delete abuse state: %w",
			err,
		)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf(
			"abuse state not found",
		)
	}

	return nil
}
