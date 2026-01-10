-- +goose Up


-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TYPE runtime_type as ENUM('docker', 'wasm');
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS functions (
    id UUID PRIMARY KEY      DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    "name" TEXT NOT NULL,
    runtime runtime_type NOT NULL,
    wasm_path TEXT,  
    "image" TEXT NOT NULL,
    "timeout" INT DEFAULT 2, -- seconds
    max_memory INT DEFAULT 0,   -- 0 = unspecified, bytes
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_called_at TIMESTAMPTZ,
    UNIQUE (user_id, "name")
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS functions CASCADE;
DROP TYPE IF EXISTS runtime_type;
DROP EXTENSION IF EXISTS "uuid-ossp";
-- +goose StatementEnd
