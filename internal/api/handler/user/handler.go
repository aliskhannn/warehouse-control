package user

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"

	"github.com/aliskhannn/warehouse-control/internal/api/response"
	"github.com/aliskhannn/warehouse-control/internal/model"
)

// service defines the user service interface used by the auth handler.
type service interface {
	// GetUserByID returns user info by ID.
	GetUserByID(ctx context.Context, userID uuid.UUID) (*model.User, error)
}

// Handler provides HTTP handlers for user endpoints.
type Handler struct {
	service service
}

// NewHandler creates a new user handler.
func NewHandler(s service) *Handler {
	return &Handler{
		service: s,
	}
}

// GetByID returns user information by ID.
func (h *Handler) GetByID(c *ginext.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, fmt.Errorf("invalid user ID"))
		return
	}

	user, err := h.service.GetUserByID(c, userID)
	if err != nil {
		zlog.Logger.Error().Err(err).Str("userID", userIDStr).Msg("failed to get user")
		response.Fail(c, http.StatusInternalServerError, fmt.Errorf("failed to get user: %w", err))
		return
	}

	// Отправляем только публичные поля, не отдаём пароль
	response.OK(c, map[string]interface{}{
		"id":       user.ID,
		"username": user.Username,
		"role":     user.Role,
	})
}
