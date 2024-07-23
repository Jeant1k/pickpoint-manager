-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS outbox (
    outbox_id SERIAL PRIMARY KEY,
    event_time TIMESTAMPTZ NOT NULL,
    method_name TEXT NOT NULL,
    raw_request JSONB NOT NULL,
    processed BOOLEAN DEFAULT FALSE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS outbox;
-- +goose StatementEnd
