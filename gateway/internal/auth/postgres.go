package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/namtran1812/sentrymesh/gateway/internal/identity"
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
	name string,
	rawKey string,
	principal identity.Identity,
	expiresAt *time.Time,
) error {
	scopes := strings.Join(principal.Scopes, ",")

	_, err := s.pool.Exec(
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
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		`,
		name,
		Hash(rawKey),
		principal.UserID,
		string(principal.Role),
		principal.Team,
		scopes,
		expiresAt,
	)

	if err != nil {
		return fmt.Errorf("create API key: %w", err)
	}

	return nil
}

func (s *PostgresStore) Resolve(
	ctx context.Context,
	rawKey string,
) (identity.Identity, error) {
	var (
		id        int64
		name      string
		userID    string
		role      string
		team      string
		scopes    string
		expiresAt *time.Time
		revokedAt *time.Time
	)

	err := s.pool.QueryRow(
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
			revoked_at
		FROM api_keys
		WHERE key_hash = $1
		`,
		Hash(rawKey),
	).Scan(
		&id,
		&name,
		&userID,
		&role,
		&team,
		&scopes,
		&expiresAt,
		&revokedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return identity.Identity{}, fmt.Errorf("invalid API key")
	}

	if err != nil {
		return identity.Identity{}, fmt.Errorf("resolve API key: %w", err)
	}

	if revokedAt != nil {
		return identity.Identity{}, fmt.Errorf("API key revoked")
	}

	if expiresAt != nil && time.Now().After(*expiresAt) {
		return identity.Identity{}, fmt.Errorf("API key expired")
	}

	return identity.Identity{
		UserID:  userID,
		Role:    identity.Role(role),
		Team:    team,
		Scopes:  splitScopes(scopes),
		KeyID:   id,
		KeyName: name,
	}, nil
}

func (s *PostgresStore) Revoke(
	ctx context.Context,
	rawKey string,
) error {
	tag, err := s.pool.Exec(
		ctx,
		`
		UPDATE api_keys
		SET revoked_at = NOW()
		WHERE key_hash = $1
		  AND revoked_at IS NULL
		`,
		Hash(rawKey),
	)

	if err != nil {
		return fmt.Errorf("revoke API key: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("API key not found or already revoked")
	}

	return nil
}

func (s *PostgresStore) List(
	ctx context.Context,
) ([]KeyRecord, error) {
	rows, err := s.pool.Query(
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
		return nil, fmt.Errorf("list API keys: %w", err)
	}
	defer rows.Close()

	results := make([]KeyRecord, 0)

	for rows.Next() {
		var (
			item      KeyRecord
			expiresAt *time.Time
			revokedAt *time.Time
			createdAt time.Time
		)

		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.UserID,
			&item.Role,
			&item.Team,
			&item.Scopes,
			&expiresAt,
			&revokedAt,
			&createdAt,
		); err != nil {
			return nil, err
		}

		item.CreatedAt = createdAt.UTC().Format(time.RFC3339Nano)

		if expiresAt != nil {
			value := expiresAt.UTC().Format(time.RFC3339Nano)
			item.ExpiresAt = &value
		}

		if revokedAt != nil {
			value := revokedAt.UTC().Format(time.RFC3339Nano)
			item.RevokedAt = &value
		}

		results = append(results, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (s *PostgresStore) RevokeByID(
	ctx context.Context,
	id int64,
) error {
	tag, err := s.pool.Exec(
		ctx,
		`
		UPDATE api_keys
		SET revoked_at = NOW()
		WHERE id = $1
		  AND revoked_at IS NULL
		`,
		id,
	)

	if err != nil {
		return fmt.Errorf("revoke API key: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("API key not found or already revoked")
	}

	return nil
}

func (s *PostgresStore) FindByName(
	ctx context.Context,
	name string,
) (KeyRecord, error) {
	return s.findOne(
		ctx,
		`WHERE name = $1 ORDER BY id DESC LIMIT 1`,
		name,
	)
}

func (s *PostgresStore) FindByID(
	ctx context.Context,
	id int64,
) (KeyRecord, error) {
	return s.findOne(
		ctx,
		`WHERE id = $1`,
		id,
	)
}

func (s *PostgresStore) findOne(
	ctx context.Context,
	where string,
	arg any,
) (KeyRecord, error) {
	var (
		item      KeyRecord
		expiresAt *time.Time
		revokedAt *time.Time
		createdAt time.Time
	)

	err := s.pool.QueryRow(
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
		`+where,
		arg,
	).Scan(
		&item.ID,
		&item.Name,
		&item.UserID,
		&item.Role,
		&item.Team,
		&item.Scopes,
		&expiresAt,
		&revokedAt,
		&createdAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return KeyRecord{}, fmt.Errorf("API key not found")
	}

	if err != nil {
		return KeyRecord{}, err
	}

	item.CreatedAt = createdAt.UTC().Format(time.RFC3339Nano)

	if expiresAt != nil {
		value := expiresAt.UTC().Format(time.RFC3339Nano)
		item.ExpiresAt = &value
	}

	if revokedAt != nil {
		value := revokedAt.UTC().Format(time.RFC3339Nano)
		item.RevokedAt = &value
	}

	return item, nil
}

func (s *PostgresStore) Close() error {
	// Pool lifetime belongs to the application/database layer.
	return nil
}
