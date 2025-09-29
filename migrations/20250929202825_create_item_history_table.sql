-- +goose Up
-- +goose StatementBegin
CREATE TABLE item_history
(
    id         UUID PRIMARY KEY         DEFAULT uuid_generate_v4(),
    item_id    UUID        NOT NULL REFERENCES items (id) ON DELETE CASCADE,
    action     item_action NOT NULL,
    changed_by UUID        NOT NULL REFERENCES users (id),
    changed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    old_data   JSONB,
    new_data   JSONB
);

CREATE INDEX idx_item_history_item_id ON item_history (item_id);
CREATE INDEX idx_item_history_changed_at ON item_history (changed_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS item_history;
DROP INDEX idx_item_history_item_id;
DROP INDEX idx_item_history_changed_at;
-- +goose StatementEnd
