CREATE TABLE IF NOT EXISTS abuse_state (
    key_id BIGINT PRIMARY KEY,
    score INTEGER NOT NULL DEFAULT 0,
    last_updated TIMESTAMPTZ NOT NULL,
    cooldown_until TIMESTAMPTZ,

    CONSTRAINT abuse_score_nonnegative
    CHECK (score >= 0)
);

CREATE INDEX IF NOT EXISTS idx_abuse_cooldown
    ON abuse_state(cooldown_until);
