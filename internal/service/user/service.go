package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/aliskhannn/warehouse-control/internal/config"
	"github.com/aliskhannn/warehouse-control/internal/model"
	repouser "github.com/aliskhannn/warehouse-control/internal/repository/user"
)

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// repository defines the interface for user-related data access.
type repository interface {
	// CreateUser add a new user to database.
	CreateUser(ctx context.Context, user *model.User) (uuid.UUID, error)

	// GetUserByID retrieves a user by id.
	GetUserByID(ctx context.Context, userID uuid.UUID) (*model.User, error)

	// GetUserByUsername retrieves a user by username.
	GetUserByUsername(ctx context.Context, username string) (*model.User, error)

	// CheckUserExistsByUsername checks if a user with the given username already exists in the database.
	CheckUserExistsByUsername(ctx context.Context, username string) (bool, error)
}

// Service contains business logic for user management such as registration and authentication.
type Service struct {
	repository repository
	cfg        *config.Config
}

// NewService creates a new user service with the provided repository and configuration.
func NewService(r repository, cfg *config.Config) *Service {
	return &Service{
		repository: r,
		cfg:        cfg,
	}
}

// Register creates a new user account with the given username, role, and password.
// Returns the created user's ID or an error if the user already exists.
func (s *Service) Register(ctx context.Context, username, role, password string) (uuid.UUID, error) {
	// Check if user already exists.
	exists, err := s.repository.CheckUserExistsByUsername(ctx, username)
	if err != nil {
		return uuid.Nil, fmt.Errorf("check user exists: %w", err)
	}
	if exists {
		return uuid.Nil, ErrUserAlreadyExists
	}

	// Hash password.
	hashedPassword, err := hashPassword(password)
	if err != nil {
		return uuid.Nil, fmt.Errorf("hash password: %w", err)
	}

	user := &model.User{
		Username:     username,
		Role:         role,
		PasswordHash: hashedPassword,
	}

	id, err := s.repository.CreateUser(ctx, user)
	if err != nil {
		return uuid.Nil, fmt.Errorf("create user: %w", err)
	}

	return id, nil
}

// Login authenticates a user by username and password, returning a JWT if successful.
func (s *Service) Login(ctx context.Context, username, password string) (string, error) {
	user, err := s.repository.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, repouser.ErrUserNotFound) {
			return "", ErrInvalidCredentials
		}

		return "", fmt.Errorf("get user by username: %w", err)
	}

	// Verify password.
	if err := verifyPassword(password, user.PasswordHash); err != nil {
		return "", ErrInvalidCredentials
	}

	// Generate JWT token.
	token, err := generateToken(user, s.cfg.JWT.Secret, s.cfg.JWT.TTL)
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}

	return token, nil
}

// GetUserByID returns user info by ID.
func (s *Service) GetUserByID(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	user, err := s.repository.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return user, nil
}

// hashPassword generates a bcrypt hash for the given password.
func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

// verifyPassword checks if the given password matches the stored hash.
func verifyPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// generateToken creates a signed JWT token containing the user's ID, username, and role.
func generateToken(user *model.User, secret string, ttl time.Duration) (string, error) {
	expTime := time.Now().Add(ttl)

	claims := jwt.MapClaims{
		"user_id":  user.ID.String(),
		"username": user.Username,
		"role":     user.Role,
		"exp":      expTime.Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
