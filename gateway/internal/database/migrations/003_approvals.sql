CREATE TABLE IF NOT EXISTS approvals (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    tool TEXT NOT NULL,
    arguments JSONB NOT NULL,
    risk INTEGER NOT NULL,
    reason TEXT NOT NULL,
    status TEXT NOT NULL,
    executed_at TIMESTAMPTZ,

    CONSTRAINT approvals_status_check
    CHECK (
        status IN (
            'PENDING',
            'APPROVED',
            'REJECTED',
            'EXECUTING',
            'EXECUTED'
        )
    )
);

CREATE INDEX IF NOT EXISTS idx_approvals_status
    ON approvals(status);

CREATE INDEX IF NOT EXISTS idx_approvals_created_at
    ON approvals(created_at DESC);
