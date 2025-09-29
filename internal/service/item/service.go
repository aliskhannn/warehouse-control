package item

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/aliskhannn/warehouse-control/internal/model"
)

// repository defines the interface for item-related data access.
type repository interface {
	// CreateItem adds a new item to the database and returns its ID.
	CreateItem(ctx context.Context, item *model.Item) (uuid.UUID, error)

	// GetItemByID retrieves an item by its ID.
	GetItemByID(ctx context.Context, itemID uuid.UUID) (*model.Item, error)

	// GetAllItems retrieves all items, optionally filtered by name.
	GetAllItems(ctx context.Context, nameFilter string) ([]*model.Item, error)

	// UpdateItem updates an existing item in the database.
	UpdateItem(ctx context.Context, item *model.Item) error

	// DeleteItem removes an item by its ID.
	DeleteItem(ctx context.Context, itemID uuid.UUID) error

	// GetItemHistory retrieves change history for an item.
	GetItemHistory(ctx context.Context, itemID uuid.UUID) ([]*model.ItemHistory, error)

	// CompareVersions decodes old and new JSONB data from history and returns them as maps.
	CompareVersions(oldData, newData json.RawMessage) (map[string]interface{}, map[string]interface{}, error)
}

// Service provides business logic for items and item history.
type Service struct {
	repository repository
}

// NewService creates a new item service.
func NewService(r repository) *Service {
	return &Service{repository: r}
}

// Create adds a new item with the specified fields.
func (s *Service) Create(ctx context.Context, name, description string, quantity int, price decimal.Decimal) (uuid.UUID, error) {
	item := &model.Item{
		Name:        name,
		Description: description,
		Quantity:    quantity,
		Price:       price,
	}

	id, err := s.repository.CreateItem(ctx, item)
	if err != nil {
		return uuid.Nil, fmt.Errorf("create item: %w", err)
	}

	return id, nil
}

// GetByID retrieves an item by its ID.
func (s *Service) GetByID(ctx context.Context, itemID uuid.UUID) (*model.Item, error) {
	item, err := s.repository.GetItemByID(ctx, itemID)
	if err != nil {
		return nil, fmt.Errorf("get item by id: %w", err)
	}

	return item, nil
}

// GetAll retrieves all items, optionally filtered by name.
func (s *Service) GetAll(ctx context.Context, nameFilter string) ([]*model.Item, error) {
	items, err := s.repository.GetAllItems(ctx, nameFilter)
	if err != nil {
		return nil, fmt.Errorf("get all items: %w", err)
	}

	return items, nil
}

// Update modifies an existing item.
func (s *Service) Update(ctx context.Context, itemID uuid.UUID, name, description string, quantity int, price decimal.Decimal) error {
	item := &model.Item{
		ID:          itemID,
		Name:        name,
		Description: description,
		Quantity:    quantity,
		Price:       price,
	}

	if err := s.repository.UpdateItem(ctx, item); err != nil {
		return fmt.Errorf("update item: %w", err)
	}

	return nil
}

// Delete removes an item by its ID.
func (s *Service) Delete(ctx context.Context, itemID uuid.UUID) error {
	if err := s.repository.DeleteItem(ctx, itemID); err != nil {
		return fmt.Errorf("delete item: %w", err)
	}

	return nil
}

// GetHistory retrieves the change history for a given item.
func (s *Service) GetHistory(ctx context.Context, itemID uuid.UUID) ([]*model.ItemHistory, error) {
	history, err := s.repository.GetItemHistory(ctx, itemID)
	if err != nil {
		return nil, fmt.Errorf("get item history: %w", err)
	}

	return history, nil
}

// CompareVersions decodes old and new JSONB data from history and returns them as maps.
func (s *Service) CompareVersions(oldData, newData json.RawMessage) (map[string]interface{}, map[string]interface{}, error) {
	oldMap, newMap, err := s.repository.CompareVersions(oldData, newData)
	if err != nil {
		return nil, nil, fmt.Errorf("compare versions: %w", err)
	}

	return oldMap, newMap, nil
}
