ALTER TABLE audit_events
ADD COLUMN IF NOT EXISTS latency_us BIGINT NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_audit_latency_us
ON audit_events(latency_us);
