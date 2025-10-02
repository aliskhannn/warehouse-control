package item

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/wb-go/wbf/dbpg"

	"github.com/aliskhannn/warehouse-control/internal/model"
)

var (
	ErrItemNotFound = errors.New("item not found")
	ErrNoItemsFound = errors.New("no items found")
)

// Repository provides methods to interact with items table.
type Repository struct {
	db *dbpg.DB
}

// NewRepository creates a new item repository.
func NewRepository(db *dbpg.DB) *Repository {
	return &Repository{db: db}
}

// CreateItem adds a new item to the database.
func (r *Repository) CreateItem(ctx context.Context, item *model.Item) (uuid.UUID, error) {
	query := `
		INSERT INTO items (name, description, quantity, price)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(
		ctx, query, item.Name, item.Description, item.Quantity, item.Price,
	).Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create item: %w", err)
	}

	return item.ID, nil
}

// GetItemByID retrieves an item by id.
func (r *Repository) GetItemByID(ctx context.Context, itemID uuid.UUID) (*model.Item, error) {
	query := `
        SELECT id, name, description, quantity, price, created_at, updated_at
        FROM items
        WHERE id = $1
    `

	var i model.Item
	err := r.db.QueryRowContext(ctx, query, itemID).Scan(
		&i.ID, &i.Name, &i.Description, &i.Quantity, &i.Price, &i.CreatedAt, &i.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrItemNotFound
		}

		return nil, fmt.Errorf("query item by id: %w", err)
	}

	return &i, nil
}

// GetAllItems retrieves all items, optionally filtered by name.
func (r *Repository) GetAllItems(ctx context.Context, nameFilter string) ([]*model.Item, error) {
	query := `
		SELECT id, name, description, quantity, price, created_at, updated_at
		FROM items
		WHERE ($1 = '' OR name ILIKE '%' || $1 || '%')
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, nameFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to query items: %w", err)
	}
	defer rows.Close()

	var items []*model.Item
	for rows.Next() {
		var i model.Item
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Description,
			&i.Quantity,
			&i.Price,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan item: %w", err)
		}

		items = append(items, &i)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate items: %w", err)
	}

	if len(items) == 0 {
		return nil, ErrNoItemsFound
	}

	return items, nil
}

// UpdateItem updates an existing item in the database.
func (r *Repository) UpdateItem(ctx context.Context, item *model.Item) error {
	query := `
		UPDATE items
		SET name = $1, description = $2, quantity = $3, price = $4, updated_at = NOW()
		WHERE id = $5
	`

	res, err := r.db.ExecContext(ctx, query, item.Name, item.Description, item.Quantity, item.Price, item.ID)
	if err != nil {
		return fmt.Errorf("failed to update item: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrItemNotFound
	}

	return nil
}

// DeleteItem deletes an item by id.
func (r *Repository) DeleteItem(ctx context.Context, itemID uuid.UUID) error {
	query := `DELETE FROM items WHERE id = $1`

	res, err := r.db.ExecContext(ctx, query, itemID)
	if err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrItemNotFound
	}

	return nil
}

// GetItemHistory retrieves change history for an item.
func (r *Repository) GetItemHistory(ctx context.Context, itemID uuid.UUID) ([]*model.ItemHistory, error) {
	query := `
		SELECT id, item_id, action, changed_by, changed_at, old_data, new_data
		FROM item_history
		WHERE item_id = $1
		ORDER BY changed_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, itemID)
	if err != nil {
		return nil, fmt.Errorf("failed to query item history: %w", err)
	}
	defer rows.Close()

	var history []*model.ItemHistory
	for rows.Next() {
		var h model.ItemHistory
		var oldData, newData sql.NullString

		if err := rows.Scan(
			&h.ID, &h.ItemID, &h.Action, &h.ChangedBy, &h.ChangedAt, &oldData, &newData,
		); err != nil {
			return nil, fmt.Errorf("failed to scan item history: %w", err)
		}

		// Convert into json.RawMessage
		if oldData.Valid {
			h.OldData = json.RawMessage(oldData.String)
		} else {
			h.OldData = json.RawMessage(`{}`)
		}

		if newData.Valid {
			h.NewData = json.RawMessage(newData.String)
		} else {
			h.NewData = json.RawMessage(`{}`)
		}

		history = append(history, &h)
	}

	return history, nil
}

// CompareVersions decodes old and new JSONB data from history and returns them as maps.
func (r *Repository) CompareVersions(oldData, newData json.RawMessage) (map[string]interface{}, map[string]interface{}, error) {
	var oldMap, newMap map[string]interface{}

	if oldData != nil {
		if err := json.Unmarshal(oldData, &oldMap); err != nil {
			return nil, nil, fmt.Errorf("failed to unmarshal old data: %w", err)
		}
	}

	if newData != nil {
		if err := json.Unmarshal(newData, &newMap); err != nil {
			return nil, nil, fmt.Errorf("failed to unmarshal new data: %w", err)
		}
	}

	return oldMap, newMap, nil
}

// SetCurrentUser sets the current user in the PostgreSQL session for auditing.
func (r *Repository) SetCurrentUser(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "SELECT set_config('app.current_user_id', $1, false)", userID.String())
	if err != nil {
		return fmt.Errorf("failed to set current_user_id: %w", err)
	}

	return nil
}
