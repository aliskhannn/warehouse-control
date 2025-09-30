package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/wb-go/wbf/ginext"

	"github.com/aliskhannn/warehouse-control/internal/api/response"
	"github.com/aliskhannn/warehouse-control/internal/model"
)

// service defines the interface for item service used by the handler.
type service interface {
	// GetHistory retrieves the change history for a given item.
	GetHistory(ctx context.Context, itemID uuid.UUID) ([]*model.ItemHistory, error)

	// CompareVersions decodes old and new JSONB data from history and returns them as maps.
	CompareVersions(oldData, newData json.RawMessage) (map[string]interface{}, map[string]interface{}, error)
}

// Handler provides HTTP handlers for item audit operations.
type Handler struct {
	service service
}

// NewHandler creates a new item audit handler.
func NewHandler(s service) *Handler {
	return &Handler{
		service: s,
	}
}

// GetHistory returns the change history for a specific item.
func (h *Handler) GetHistory(c *ginext.Context) {
	itemIDStr := c.Param("id")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, fmt.Errorf("invalid item ID"))
		return
	}

	history, err := h.service.GetHistory(c, itemID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, fmt.Errorf("failed to get history: %w", err))
		return
	}

	response.OK(c, history)
}

// CompareVersions compares old and new versions of item data.
func (h *Handler) CompareVersions(c *ginext.Context) {
	// Accept JSON payload with oldData and newData.
	var req struct {
		Old json.RawMessage `json:"old"`
		New json.RawMessage `json:"new"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, fmt.Errorf("invalid request body"))
		return
	}

	oldMap, newMap, err := h.service.CompareVersions(req.Old, req.New)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, fmt.Errorf("failed to compare versions: %w", err))
		return
	}

	response.OK(c, map[string]interface{}{
		"old": oldMap,
		"new": newMap,
	})
}
