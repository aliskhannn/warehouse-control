-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION log_item_insert() RETURNS TRIGGER AS
$$
BEGIN
    INSERT INTO item_history(item_id, action, changed_by, old_data, new_data)
    VALUES (NEW.id,
            'INSERT',
            current_setting('app.current_user_id')::UUID,
            NULL,
            to_jsonb(NEW));
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION log_item_update() RETURNS TRIGGER AS
$$
BEGIN
    INSERT INTO item_history(item_id, action, changed_by, old_data, new_data)
    VALUES (NEW.id,
            'UPDATE',
            current_setting('app.current_user_id')::UUID,
            to_jsonb(OLD),
            to_jsonb(NEW));
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION log_item_delete() RETURNS TRIGGER AS
$$
BEGIN
    INSERT INTO item_history(item_id, action, changed_by, old_data, new_data)
    VALUES (OLD.id,
            'DELETE',
            current_setting('app.current_user_id')::UUID,
            to_jsonb(OLD),
            NULL);
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_item_insert ON items;
CREATE TRIGGER trg_item_insert
    AFTER INSERT
    ON items
    FOR EACH ROW
EXECUTE FUNCTION log_item_insert();

DROP TRIGGER IF EXISTS trg_item_update ON items;
CREATE TRIGGER trg_item_update
    AFTER UPDATE
    ON items
    FOR EACH ROW
EXECUTE FUNCTION log_item_update();

DROP TRIGGER IF EXISTS trg_item_delete ON items;
CREATE TRIGGER trg_item_delete
    BEFORE DELETE
    ON items
    FOR EACH ROW
EXECUTE FUNCTION log_item_delete();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trg_item_insert ON items;
DROP TRIGGER IF EXISTS trg_item_update ON items;
DROP TRIGGER IF EXISTS trg_item_delete ON items;

DROP FUNCTION IF EXISTS log_item_insert();
DROP FUNCTION IF EXISTS log_item_update();
DROP FUNCTION IF EXISTS log_item_delete();
-- +goose StatementEnd
