package item

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"

	"github.com/aliskhannn/warehouse-control/internal/api/response"
	"github.com/aliskhannn/warehouse-control/internal/model"
	repoitem "github.com/aliskhannn/warehouse-control/internal/repository/item"
)

// service defines the interface for item service used by the handler.
type service interface {
	// Create adds a new item with the specified fields.
	Create(ctx context.Context, userID uuid.UUID, name, description string, quantity int, price decimal.Decimal) (uuid.UUID, error)

	// GetByID retrieves an item by its ID.
	GetByID(ctx context.Context, itemID uuid.UUID) (*model.Item, error)

	// GetAll retrieves all items, optionally filtered by name.
	GetAll(ctx context.Context, nameFilter string) ([]*model.Item, error)

	// Update modifies an existing item.
	Update(ctx context.Context, userID, itemID uuid.UUID, name, description string, quantity int, price decimal.Decimal) error

	// Delete removes an item by its ID.
	Delete(ctx context.Context, userID, itemID uuid.UUID) error
}

// Handler provides HTTP handlers for item endpoints.
type Handler struct {
	service   service
	validator *validator.Validate
}

// NewHandler creates a new item handler.
func NewHandler(s service, v *validator.Validate) *Handler {
	return &Handler{
		service:   s,
		validator: v,
	}
}

// CreateRequest represents the JSON request body for creating an item.
type CreateRequest struct {
	Name        string          `json:"name" validate:"required"`
	Description string          `json:"description"`
	Quantity    int             `json:"quantity" validate:"required,min=0"`
	Price       decimal.Decimal `json:"price" validate:"required"`
}

// UpdateRequest represents the JSON request body for updating an item.
type UpdateRequest struct {
	Name        string          `json:"name" validate:"required"`
	Description string          `json:"description"`
	Quantity    int             `json:"quantity" validate:"required,min=0"`
	Price       decimal.Decimal `json:"price" validate:"required"`
}

// Create handles creating a new item.
func (h *Handler) Create(c *ginext.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		zlog.Logger.Error().Err(err).Msg("failed to bind JSON")
		response.Fail(c, http.StatusBadRequest, fmt.Errorf("invalid request body"))
		return
	}

	if err := h.validator.Struct(req); err != nil {
		zlog.Logger.Error().Err(err).Msg("validation failed")
		response.Fail(c, http.StatusBadRequest, fmt.Errorf("validation error: %s", err.Error()))
		return
	}

	userIDVal, exists := c.Get("userID")
	if !exists {
		response.Fail(c, http.StatusUnauthorized, fmt.Errorf("userID not found in context"))
		return
	}

	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, fmt.Errorf("invalid userID type"))
		return
	}

	id, err := h.service.Create(c.Request.Context(), userID, req.Name, req.Description, req.Quantity, req.Price)
	if err != nil {
		zlog.Logger.Error().Err(err).Msg("failed to create item")
		response.Fail(c, http.StatusInternalServerError, fmt.Errorf("failed to create item"))
		return
	}

	response.Created(c, map[string]string{"id": id.String()})
}

// Update handles updating an existing item.
func (h *Handler) Update(c *ginext.Context) {
	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		zlog.Logger.Error().Err(err).Msg("failed to bind JSON")
		response.Fail(c, http.StatusBadRequest, fmt.Errorf("invalid request body"))
		return
	}

	if err := h.validator.Struct(req); err != nil {
		zlog.Logger.Error().Err(err).Msg("validation failed")
		response.Fail(c, http.StatusBadRequest, fmt.Errorf("validation error: %s", err.Error()))
		return
	}

	userID, itemID, ok := h.getUserAndItemIDFromContext(c)
	if !ok {
		return
	}

	if err := h.service.Update(c.Request.Context(), userID, itemID, req.Name, req.Description, req.Quantity, req.Price); err != nil {
		if errors.Is(err, repoitem.ErrItemNotFound) {
			zlog.Logger.Error().Err(err).Msg("failed to update item")
			response.Fail(c, http.StatusNotFound, err)
			return
		}

		zlog.Logger.Error().Err(err).Msg("failed to update item")
		response.Fail(c, http.StatusInternalServerError, fmt.Errorf("failed to update item"))
		return
	}

	response.OK(c, map[string]string{"id": itemID.String()})
}

// Delete handles deleting an item.
func (h *Handler) Delete(c *ginext.Context) {
	userID, itemID, ok := h.getUserAndItemIDFromContext(c)
	if !ok {
		return
	}

	if err := h.service.Delete(c.Request.Context(), userID, itemID); err != nil {
		if errors.Is(err, repoitem.ErrItemNotFound) {
			zlog.Logger.Error().Err(err).Msg("failed to delete item")
			response.Fail(c, http.StatusNotFound, err)
			return
		}

		zlog.Logger.Error().Err(err).Msg("failed to delete item")
		response.Fail(c, http.StatusInternalServerError, fmt.Errorf("failed to delete item"))
		return
	}

	response.OK(c, map[string]string{"id": itemID.String()})
}

// GetByID handles retrieving an item by ID.
func (h *Handler) GetByID(c *ginext.Context) {
	itemIDStr := c.Param("id")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, fmt.Errorf("invalid item ID"))
		return
	}

	item, err := h.service.GetByID(c.Request.Context(), itemID)
	if err != nil {
		if errors.Is(err, repoitem.ErrItemNotFound) {
			zlog.Logger.Error().Err(err).Str("itemID", itemIDStr).Msg("failed to get item by id")
			response.Fail(c, http.StatusNotFound, err)
			return
		}

		zlog.Logger.Error().Err(err).Msg("failed to get item")
		response.Fail(c, http.StatusInternalServerError, fmt.Errorf("failed to get item"))
		return
	}

	response.OK(c, item)
}

// GetAll handles retrieving all items, optionally filtered by name query param.
func (h *Handler) GetAll(c *ginext.Context) {
	nameFilter := c.Query("name")

	items, err := h.service.GetAll(c.Request.Context(), nameFilter)
	if err != nil {
		if errors.Is(err, repoitem.ErrNoItemsFound) {
			zlog.Logger.Error().Err(err).Msg("failed to get items")
			response.Fail(c, http.StatusNotFound, err)
			return
		}

		zlog.Logger.Error().Err(err).Msg("failed to get all items")
		response.Fail(c, http.StatusInternalServerError, fmt.Errorf("failed to get items"))
		return
	}

	response.OK(c, items)
}

// getUserAndItemIDFromContext retrieves the userID from the context and the itemID from the request parameters.
// Returns an error and automatically sends a response if something goes wrong.
func (h *Handler) getUserAndItemIDFromContext(c *ginext.Context) (uuid.UUID, uuid.UUID, bool) {
	itemIDStr := c.Param("id")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, fmt.Errorf("invalid item ID"))
		return uuid.Nil, uuid.Nil, false
	}

	userIDVal, exists := c.Get("userID")
	if !exists {
		response.Fail(c, http.StatusUnauthorized, fmt.Errorf("userID not found in context"))
		return uuid.Nil, uuid.Nil, false
	}

	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, fmt.Errorf("invalid userID type"))
		return uuid.Nil, uuid.Nil, false
	}

	return userID, itemID, true
}
