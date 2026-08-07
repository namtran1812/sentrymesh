package auth

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/namtran1812/sentrymesh/gateway/internal/identity"

	_ "modernc.org/sqlite"
)

type Key struct {
	Scopes    string
	ID        int64
	Name      string
	KeyHash   string
	UserID    string
	Role      identity.Role
	Team      string
	ExpiresAt *string
	RevokedAt *string
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
		CREATE TABLE IF NOT EXISTS api_keys (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			key_hash TEXT NOT NULL UNIQUE,
			user_id TEXT NOT NULL,
			role TEXT NOT NULL,
			team TEXT NOT NULL,
			scopes TEXT NOT NULL DEFAULT '',
			expires_at TEXT,
			revoked_at TEXT,
			created_at TEXT NOT NULL
		);

		CREATE INDEX IF NOT EXISTS idx_api_keys_hash
		ON api_keys(key_hash);
	`)

	return err
}

func Hash(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
func (s *Store) Create(
	ctx context.Context,
	name string,
	rawKey string,
	principal identity.Identity,
	expiresAt *time.Time,
) error {
	var expires *string

	if expiresAt != nil {
		value := expiresAt.UTC().Format(time.RFC3339Nano)
		expires = &value
	}

	scopes := strings.Join(principal.Scopes, ",")

	_, err := s.db.ExecContext(
		ctx,
		`
		INSERT INTO api_keys (
			name,
			key_hash,
			user_id,
			role,
			team,
			scopes,
			expires_at,
			created_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`,
		name,
		Hash(rawKey),
		principal.UserID,
		principal.Role,
		principal.Team,
		scopes,
		expires,
		time.Now().UTC().Format(time.RFC3339Nano),
	)

	return err
}

func (s *Store) Resolve(
	ctx context.Context,
	rawKey string,
) (identity.Identity, error) {
	var key Key

	err := s.db.QueryRowContext(
		ctx,
		`
		SELECT
			id,
			name,
			key_hash,
			user_id,
			role,
			team,
			scopes,
			expires_at,
			revoked_at
		FROM api_keys
		WHERE key_hash = ?
		`,
		Hash(rawKey),
	).Scan(
		&key.ID,
		&key.Name,
		&key.KeyHash,
		&key.UserID,
		&key.Role,
		&key.Team,
		&key.Scopes,
		&key.ExpiresAt,
		&key.RevokedAt,
	)
	if err != nil {
		return identity.Identity{}, fmt.Errorf("invalid API key")
	}

	if key.RevokedAt != nil {
		return identity.Identity{}, fmt.Errorf("API key revoked")
	}

	if key.ExpiresAt != nil {
		expires, err := time.Parse(time.RFC3339Nano, *key.ExpiresAt)
		if err != nil {
			return identity.Identity{}, fmt.Errorf("invalid expiry")
		}

		if time.Now().After(expires) {
			return identity.Identity{}, fmt.Errorf("API key expired")
		}
	}

	return identity.Identity{
		UserID: key.UserID,
		Role:   key.Role,
		Team:   key.Team,
		Scopes: splitScopes(key.Scopes),
	}, nil
}

func (s *Store) Revoke(
	ctx context.Context,
	rawKey string,
) error {
	result, err := s.db.ExecContext(
		ctx,
		`
		UPDATE api_keys
		SET revoked_at = ?
		WHERE key_hash = ?
		  AND revoked_at IS NULL
		`,
		time.Now().UTC().Format(time.RFC3339Nano),
		Hash(rawKey),
	)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return fmt.Errorf("API key not found or already revoked")
	}

	return nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func splitScopes(raw string) []string {
	if raw == "" {
		return nil
	}

	parts := strings.Split(raw, ",")

	result := make([]string, 0, len(parts))

	for _, part := range parts {
		scope := strings.TrimSpace(part)
		if scope != "" {
			result = append(result, scope)
		}
	}

	return result
}

type KeyRecord struct {
	ID        int64   `json:"id"`
	Name      string  `json:"name"`
	UserID    string  `json:"user_id"`
	Role      string  `json:"role"`
	Team      string  `json:"team"`
	Scopes    string  `json:"scopes"`
	ExpiresAt *string `json:"expires_at,omitempty"`
	RevokedAt *string `json:"revoked_at,omitempty"`
	CreatedAt string  `json:"created_at"`
}

func (s *Store) List(
	ctx context.Context,
) ([]KeyRecord, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`
		SELECT
			id,
			name,
			user_id,
			role,
			team,
			scopes,
			expires_at,
			revoked_at,
			created_at
		FROM api_keys
		ORDER BY id DESC
		`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]KeyRecord, 0)

	for rows.Next() {
		var item KeyRecord

		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.UserID,
			&item.Role,
			&item.Team,
			&item.Scopes,
			&item.ExpiresAt,
			&item.RevokedAt,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}

		results = append(results, item)
	}

	return results, rows.Err()
}

func (s *Store) RevokeByID(
	ctx context.Context,
	id int64,
) error {
	result, err := s.db.ExecContext(
		ctx,
		`
		UPDATE api_keys
		SET revoked_at = ?
		WHERE id = ?
		  AND revoked_at IS NULL
		`,
		time.Now().UTC().Format(time.RFC3339Nano),
		id,
	)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return fmt.Errorf("API key not found or already revoked")
	}

	return nil
}

func (s *Store) FindByName(
	ctx context.Context,
	name string,
) (KeyRecord, error) {
	var item KeyRecord

	err := s.db.QueryRowContext(
		ctx,
		`
		SELECT
			id,
			name,
			user_id,
			role,
			team,
			scopes,
			expires_at,
			revoked_at,
			created_at
		FROM api_keys
		WHERE name = ?
		ORDER BY id DESC
		LIMIT 1
		`,
		name,
	).Scan(
		&item.ID,
		&item.Name,
		&item.UserID,
		&item.Role,
		&item.Team,
		&item.Scopes,
		&item.ExpiresAt,
		&item.RevokedAt,
		&item.CreatedAt,
	)

	return item, err
}

func (s *Store) FindByID(
	ctx context.Context,
	id int64,
) (KeyRecord, error) {
	var item KeyRecord

	err := s.db.QueryRowContext(
		ctx,
		`
		SELECT
			id,
			name,
			user_id,
			role,
			team,
			scopes,
			expires_at,
			revoked_at,
			created_at
		FROM api_keys
		WHERE id = ?
		`,
		id,
	).Scan(
		&item.ID,
		&item.Name,
		&item.UserID,
		&item.Role,
		&item.Team,
		&item.Scopes,
		&item.ExpiresAt,
		&item.RevokedAt,
		&item.CreatedAt,
	)

	return item, err
}
