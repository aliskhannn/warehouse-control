-- +goose Up
-- +goose StatementBegin
CREATE TABLE items
(
    id          UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    name        TEXT           NOT NULL,
    description TEXT,
    quantity    INT            NOT NULL  DEFAULT 0,
    price       NUMERIC(12, 2) NOT NULL  DEFAULT 0.0,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_items_name ON items (name);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS items;
DROP INDEX idx_items_name;
-- +goose StatementEnd
