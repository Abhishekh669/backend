package lib

import (
	"context"
	"errors"
	"fmt"

	"github.com/Abhishekh669/backend/internals/database"
	"github.com/Abhishekh669/backend/internals/models"
	"github.com/Abhishekh669/backend/internals/rbac"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func GetUserInfoByEmail(email string, ctx context.Context) (*models.UserType, error) {

	pool, err := database.GetPostgresPool()

	if err != nil {
		return nil, errors.New("failed to connect to database")
	}

	query := `
		SELECT id, email, gender, image, is_active, last_password_reset_at,
		       role, name, phone, password, salary, created_at, updated_at
		FROM users
		WHERE email=$1
		LIMIT 1
	`

	var user models.UserType

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire DB connection: %w", err)
	}
	defer conn.Release()

	err = conn.QueryRow(ctx, query, email).Scan(
		&user.Id,
		&user.Email,
		&user.Gender,
		&user.Image,
		&user.IsActive,
		&user.LastPasswordResetAt,
		&user.Role,
		&user.Name,
		&user.Phone,
		&user.Password,
		&user.Salary,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // user not found
		}
		return nil, fmt.Errorf("error querying user: %w", err)
	}

	return &user, nil

}

func HasPermissionCheck(c *gin.Context, action rbac.Permission) (*models.Role, error) {
	currentUserEmail := c.GetString("user_email")
	if currentUserEmail == "" {
		return nil, fmt.Errorf("error not found")
	}
	val, exists := c.Get("last_password_reset_at")
	var lastPasswordResetAt int64
	if exists {
		// Type assertion
		ts, ok := val.(int64)
		if !ok {
			// fallback in case type is wrong
			lastPasswordResetAt = 0
		} else {
			lastPasswordResetAt = ts
		}
	} else {
		// fallback if key does not exist
		lastPasswordResetAt = 0
	}
	currentUser, err := GetUserInfoByEmail(currentUserEmail, c.Request.Context())
	if err != nil {
		return &currentUser.Role, errors.New(err.Error())
	}

	if currentUser == nil {
		return &currentUser.Role, fmt.Errorf("unauthorized user")
	}

	hasPermission := rbac.HasPermission(&currentUser.Role, action)

	if !hasPermission || currentUser.LastPasswordResetAt != lastPasswordResetAt {
		return &currentUser.Role, fmt.Errorf("unauthorized user")
	}
	return &currentUser.Role, nil

}
