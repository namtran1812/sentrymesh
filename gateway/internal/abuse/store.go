package abuse

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

type StoredState struct {
	KeyID         int64
	Score         int
	LastUpdated   time.Time
	CooldownUntil *time.Time
}

type Store struct {
	db *sql.DB
}

func NewStore(path string) (*Store, error) {
	db, err := sql.Open(
		"sqlite",
		path+"?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)",
	)
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
		CREATE TABLE IF NOT EXISTS abuse_state (
			key_id INTEGER PRIMARY KEY,
			score INTEGER NOT NULL,
			last_updated TEXT NOT NULL,
			cooldown_until TEXT
		);
	`)

	return err
}

func (s *Store) Save(
	ctx context.Context,
	state StoredState,
) error {
	var cooldown any

	if state.CooldownUntil != nil {
		cooldown = state.CooldownUntil.UTC().
			Format(time.RFC3339Nano)
	}

	_, err := s.db.ExecContext(
		ctx,
		`
		INSERT INTO abuse_state (
			key_id,
			score,
			last_updated,
			cooldown_until
		)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(key_id)
		DO UPDATE SET
			score = excluded.score,
			last_updated = excluded.last_updated,
			cooldown_until = excluded.cooldown_until
		`,
		state.KeyID,
		state.Score,
		state.LastUpdated.UTC().
			Format(time.RFC3339Nano),
		cooldown,
	)

	return err
}

func (s *Store) Load(
	ctx context.Context,
	keyID int64,
) (StoredState, error) {
	var item StoredState
	var lastUpdated string
	var cooldown sql.NullString

	err := s.db.QueryRowContext(
		ctx,
		`
		SELECT
			key_id,
			score,
			last_updated,
			cooldown_until
		FROM abuse_state
		WHERE key_id = ?
		`,
		keyID,
	).Scan(
		&item.KeyID,
		&item.Score,
		&lastUpdated,
		&cooldown,
	)
	if err != nil {
		return StoredState{}, err
	}

	parsedLastUpdated, err := time.Parse(
		time.RFC3339Nano,
		lastUpdated,
	)
	if err != nil {
		return StoredState{}, err
	}

	item.LastUpdated = parsedLastUpdated

	if cooldown.Valid {
		parsedCooldown, err := time.Parse(
			time.RFC3339Nano,
			cooldown.String,
		)
		if err != nil {
			return StoredState{}, err
		}

		item.CooldownUntil = &parsedCooldown
	}

	return item, nil
}

func (s *Store) List(
	ctx context.Context,
) ([]StoredState, error) {
	rows, err := s.db.QueryContext(
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
		return nil, err
	}
	defer rows.Close()

	results := make([]StoredState, 0)

	for rows.Next() {
		var item StoredState
		var lastUpdated string
		var cooldown sql.NullString

		if err := rows.Scan(
			&item.KeyID,
			&item.Score,
			&lastUpdated,
			&cooldown,
		); err != nil {
			return nil, err
		}

		parsedLastUpdated, err := time.Parse(
			time.RFC3339Nano,
			lastUpdated,
		)
		if err != nil {
			return nil, err
		}

		item.LastUpdated = parsedLastUpdated

		if cooldown.Valid {
			parsedCooldown, err := time.Parse(
				time.RFC3339Nano,
				cooldown.String,
			)
			if err != nil {
				return nil, err
			}

			item.CooldownUntil = &parsedCooldown
		}

		results = append(results, item)
	}

	return results, rows.Err()
}

func (s *Store) Delete(
	ctx context.Context,
	keyID int64,
) error {
	result, err := s.db.ExecContext(
		ctx,
		`DELETE FROM abuse_state WHERE key_id = ?`,
		keyID,
	)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return fmt.Errorf("abuse state not found")
	}

	return nil
}

func (s *Store) Close() error {
	return s.db.Close()
}
