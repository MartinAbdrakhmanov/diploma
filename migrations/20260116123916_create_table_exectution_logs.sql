-- +goose Up
-- +goose StatementBegin
CREATE TABLE execution_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    function_id UUID NOT NULL
        REFERENCES functions(id) ON DELETE CASCADE,

    started_at  TIMESTAMPTZ NOT NULL,
    finished_at TIMESTAMPTZ NOT NULL,
    duration_ms BIGINT NOT NULL,
    coldstart INT,

    status TEXT NOT NULL CHECK (status IN ('success', 'error', 'timeout')),

    exit_code INT,          -- docker only
    error_code TEXT,       
    error_message TEXT,     

    memory_bytes INT,
    cpu_percent INT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_execution_logs_fn_time
ON execution_logs (function_id, started_at DESC);
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS execution_logs CASCADE;
DROP INDEX IF EXISTS idx_execution_logs_fn_time;
-- +goose StatementEnd
