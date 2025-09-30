package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"

	"github.com/aliskhannn/warehouse-control/internal/api/response"
	repouser "github.com/aliskhannn/warehouse-control/internal/repository/user"
	serviceuser "github.com/aliskhannn/warehouse-control/internal/service/user"
)

// service defines the user service interface used by the auth handler.
type service interface {
	// Register creates a new user account with the given username, role, and password.
	// Returns the created user's ID or an error if the user already exists.
	Register(ctx context.Context, username, role, password string) (uuid.UUID, error)

	// Login authenticates a user by username and password, returning a JWT if successful.
	Login(ctx context.Context, username, password string) (string, error)
}

// Handler provides HTTP handlers for authentication endpoints.
type Handler struct {
	service   service
	validator *validator.Validate
}

// NewHandler creates a new authentication handler.
func NewHandler(s service, v *validator.Validate) *Handler {
	return &Handler{
		service:   s,
		validator: v,
	}
}

// RegisterRequest represents the JSON request body for user registration.
type RegisterRequest struct {
	Username string `json:"username" validate:"required"`
	Role     string `json:"role" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// LoginRequest represents the JSON request body for user login.
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// Register handles user registration.
func (h *Handler) Register(c *ginext.Context) {
	var req RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		zlog.Logger.Error().Err(err).Msg("failed to bind json")
		response.Fail(c, http.StatusBadRequest, fmt.Errorf("invalid request body"))
		return
	}

	if err := h.validator.Struct(req); err != nil {
		zlog.Logger.Error().Err(err).Msg("failed to validate request")
		response.Fail(c, http.StatusBadRequest, fmt.Errorf("validation error: %s", err.Error()))
		return
	}

	id, err := h.service.Register(c.Request.Context(), req.Username, req.Role, req.Password)
	if err != nil {
		if errors.Is(err, serviceuser.ErrUserAlreadyExists) {
			zlog.Logger.Error().Err(err).Msg("user already exists")
			response.Fail(c, http.StatusConflict, fmt.Errorf("user already exists"))
			return
		}

		zlog.Logger.Error().Err(err).Msg("failed to register user")
		response.Fail(c, http.StatusInternalServerError, fmt.Errorf("internal server error"))
		return
	}

	response.Created(c, map[string]string{
		"id": id.String(),
	})
}

// Login handles user authentication.
func (h *Handler) Login(c *ginext.Context) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		zlog.Logger.Error().Err(err).Msg("failed to bind json")
		response.Fail(c, http.StatusBadRequest, fmt.Errorf("invalid request body"))
		return
	}

	if err := h.validator.Struct(req); err != nil {
		zlog.Logger.Error().Err(err).Msg("failed to validate request")
		response.Fail(c, http.StatusBadRequest, fmt.Errorf("validation error: %s", err.Error()))
		return
	}

	token, err := h.service.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, serviceuser.ErrInvalidCredentials) {
			zlog.Logger.Error().Err(err).Msg("invalid credentials")
			response.Fail(c, http.StatusUnauthorized, fmt.Errorf("invalid credentials"))
			return
		}

		if errors.Is(err, repouser.ErrUserNotFound) {
			zlog.Logger.Error().Err(err).Msg("user not found")
			response.Fail(c, http.StatusNotFound, fmt.Errorf("user not found"))
			return
		}

		zlog.Logger.Error().Err(err).Msg("failed to login")
		response.Fail(c, http.StatusInternalServerError, fmt.Errorf("internal server error"))
		return
	}

	response.OK(c, map[string]string{
		"token": token,
	})
}
