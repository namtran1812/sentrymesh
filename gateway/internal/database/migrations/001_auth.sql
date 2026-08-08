CREATE TABLE IF NOT EXISTS api_keys (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    key_hash TEXT NOT NULL UNIQUE,
    user_id TEXT NOT NULL,
    role TEXT NOT NULL,
    team TEXT NOT NULL,
    scopes TEXT NOT NULL DEFAULT '',
    expires_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_api_keys_hash
    ON api_keys(key_hash);

CREATE INDEX IF NOT EXISTS idx_api_keys_user_id
    ON api_keys(user_id);

CREATE INDEX IF NOT EXISTS idx_api_keys_status
    ON api_keys(revoked_at, expires_at);
