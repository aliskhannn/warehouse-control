package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/wb-go/wbf/dbpg"

	"github.com/aliskhannn/warehouse-control/internal/model"
)

var ErrUserNotFound = errors.New("user not found")

// Repository provides methods to interact with users table.
type Repository struct {
	db *dbpg.DB
}

// NewRepository creates a new user repository.
func NewRepository(db *dbpg.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, user *model.User) (uuid.UUID, error) {
	query := `
		INSERT INTO users (username, password_hash, role)
		VALUES ($1, $2, $3)
		RETURNING id;
	`

	err := r.db.Master.QueryRowContext(
		ctx, query, user.Username, user.PasswordHash, user.Role,
	).Scan(&user.ID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user.ID, nil
}

// GetUserByID retrieves a user by id.
func (r *Repository) GetUserByID(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	query := `
        SELECT id, username, role, created_at
        FROM users
        WHERE id = $1
    `
	var u model.User
	err := r.db.Master.QueryRowContext(ctx, query, userID).Scan(
		&u.ID, &u.Username, &u.Role, &u.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("query user by id: %w", err)
	}
	return &u, nil
}

// GetUserByUsername retrieves a user by username.
func (r *Repository) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	query := `
		SELECT id, username, password_hash, role, created_at
		FROM users
		WHERE username = $1
	`

	var user model.User
	err := r.db.Master.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return &user, nil
}

// CheckUserExistsByUsername checks if a user with the given username already exists in the database.
func (r *Repository) CheckUserExistsByUsername(ctx context.Context, username string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`

	var exists bool
	err := r.db.Master.QueryRowContext(ctx, query, username).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if user exists: %w", err)
	}

	return exists, nil
}
