package auth

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/namtran1812/sentrymesh/gateway/internal/identity"

	_ "modernc.org/sqlite"
)

type Key struct {
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

	_, err := s.db.ExecContext(
		ctx,
		`
		INSERT INTO api_keys (
			name,
			key_hash,
			user_id,
			role,
			team,
			expires_at,
			created_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		`,
		name,
		Hash(rawKey),
		principal.UserID,
		principal.Role,
		principal.Team,
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
