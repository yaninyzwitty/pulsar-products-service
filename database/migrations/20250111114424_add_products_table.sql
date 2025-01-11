-- +goose Up
-- +goose StatementBegin
CREATE TABLE products (
    id          BIGINT PRIMARY KEY,                 -- Unique identifier for the product
    name        TEXT NOT NULL,                      -- Name of the product
    description TEXT,                               -- Detailed description of the product
    price       NUMERIC(10, 2) NOT NULL,            -- Price of the product with two decimal precision
    stock       INTEGER NOT NULL DEFAULT 0,         -- Number of items available in stock
    category    TEXT,                               -- Category the product belongs to
    created_at  TIMESTAMP NOT NULL DEFAULT NOW(),   -- Timestamp when the product was created
    updated_at  TIMESTAMP NOT NULL DEFAULT NOW()    -- Timestamp when the product was last updated
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS products;
-- +goose StatementEnd
