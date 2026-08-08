CREATE TABLE IF NOT EXISTS audit_events (
    id BIGSERIAL PRIMARY KEY,
    request_id TEXT NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    provider TEXT NOT NULL,
    model TEXT NOT NULL,
    decision TEXT NOT NULL,
    risk_score INTEGER NOT NULL,
    severity TEXT NOT NULL,
    latency_ms BIGINT NOT NULL,
    secret_findings JSONB NOT NULL DEFAULT 'null'::jsonb,
    pii_findings JSONB NOT NULL DEFAULT 'null'::jsonb,
    injection_findings JSONB NOT NULL DEFAULT 'null'::jsonb,
    output_findings JSONB NOT NULL DEFAULT 'null'::jsonb
);

CREATE INDEX IF NOT EXISTS idx_audit_request_id
    ON audit_events(request_id);

CREATE INDEX IF NOT EXISTS idx_audit_timestamp
    ON audit_events(timestamp DESC);

CREATE INDEX IF NOT EXISTS idx_audit_decision
    ON audit_events(decision);

CREATE INDEX IF NOT EXISTS idx_audit_severity
    ON audit_events(severity);


CREATE TABLE IF NOT EXISTS auth_events (
    id BIGSERIAL PRIMARY KEY,
    timestamp TIMESTAMPTZ NOT NULL,
    event_type TEXT NOT NULL,
    key_id BIGINT NOT NULL,
    key_name TEXT NOT NULL,
    actor JSONB NOT NULL,
    details JSONB NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_auth_events_timestamp
    ON auth_events(timestamp DESC);

CREATE INDEX IF NOT EXISTS idx_auth_events_key_id
    ON auth_events(key_id);


CREATE TABLE IF NOT EXISTS tool_events (
    id BIGSERIAL PRIMARY KEY,
    timestamp TIMESTAMPTZ NOT NULL,
    approval_id BIGINT NOT NULL,
    event_type TEXT NOT NULL,
    tool TEXT NOT NULL,
    risk INTEGER NOT NULL,
    status TEXT NOT NULL,
    details JSONB NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_tool_events_approval
    ON tool_events(approval_id);

CREATE INDEX IF NOT EXISTS idx_tool_events_timestamp
    ON tool_events(timestamp DESC);


CREATE TABLE IF NOT EXISTS rag_events (
    id BIGSERIAL PRIMARY KEY,
    request_id TEXT NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    user_id TEXT NOT NULL,
    role TEXT NOT NULL,
    team TEXT NOT NULL,
    trace JSONB NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_rag_request_id
    ON rag_events(request_id);

CREATE INDEX IF NOT EXISTS idx_rag_timestamp
    ON rag_events(timestamp DESC);


CREATE TABLE IF NOT EXISTS abuse_events (
    id BIGSERIAL PRIMARY KEY,
    timestamp TIMESTAMPTZ NOT NULL,
    event_type TEXT NOT NULL,
    key_id BIGINT NOT NULL,
    key_name TEXT NOT NULL,
    user_id TEXT NOT NULL,
    role TEXT NOT NULL,
    team TEXT NOT NULL,
    path TEXT NOT NULL,
    details JSONB NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_abuse_events_key
    ON abuse_events(key_id);

CREATE INDEX IF NOT EXISTS idx_abuse_events_timestamp
    ON abuse_events(timestamp DESC);
