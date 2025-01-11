-- +goose Up
-- +goose StatementBegin
CREATE TABLE outbox (
    id SERIAL PRIMARY KEY,              -- Unique identifier for the outbox entry
    event_type TEXT NOT NULL,           -- Type of the event (e.g., "product_created", "product_updated")
    payload TEXT NOT NULL,              -- JSON-encoded event data
    status TEXT NOT NULL DEFAULT 'pending', -- Status of the event ("pending", "processed", "failed")
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, -- Timestamp when the event was created
    processed_at TIMESTAMP             -- Timestamp when the event was successfully processed
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS outbox;
-- +goose StatementEnd
