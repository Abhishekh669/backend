package lib

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/Abhishekh669/backend/internals/database"
	"github.com/Abhishekh669/backend/internals/models"
	"github.com/Abhishekh669/backend/internals/rbac"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func GetUserInfoByEmail(email string, ctx context.Context) (*models.UserType, error) {
	if email == "" {
		return nil, errors.New("email cannot be empty")
	}

	pool, err := database.GetPostgresPool()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
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
	// Get user email from context
	currentUserEmail := c.GetString("user_email")
	if currentUserEmail == "" {
		log.Println("ERROR: user_email not found in context")
		return nil, fmt.Errorf("unauthorized: user email not found in context")
	}

	fmt.Println("this is the current user email : ", currentUserEmail)

	// Get last password reset at from context
	val, exists := c.Get("last_password_reset_at")
	var lastPasswordResetAt int64
	if exists {
		// Type assertion
		ts, ok := val.(int64)
		if !ok {
			log.Printf("WARNING: last_password_reset_at has wrong type: %T, expected int64", val)
			lastPasswordResetAt = 0
		} else {
			lastPasswordResetAt = ts
		}
	} else {
		lastPasswordResetAt = 0
	}

	// Get user info by email
	currentUser, err := GetUserInfoByEmail(currentUserEmail, c.Request.Context())
	if err != nil {
		log.Printf("ERROR: failed to get user info for email %s: %v", currentUserEmail, err)
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// FIX: Check if currentUser is nil
	if currentUser == nil {
		log.Printf("ERROR: user not found for email: %s", currentUserEmail)
		return nil, fmt.Errorf("unauthorized: user not found")
	}

	// FIX: Check if Role is nil before accessing
	if currentUser.Role == "" {
		log.Printf("ERROR: user role is nil for user: %s", currentUserEmail)
		return nil, fmt.Errorf("unauthorized: user role not assigned")
	}

	// Check permission
	hasPermission := rbac.HasPermission(&currentUser.Role, action)

	// Check password reset timestamp
	if !hasPermission {
		log.Printf("WARNING: user %s does not have permission %v", currentUserEmail, action)
		return nil, fmt.Errorf("unauthorized: insufficient permissions")
	}

	if currentUser.LastPasswordResetAt != lastPasswordResetAt {
		log.Printf("WARNING: password reset timestamp mismatch for user %s", currentUserEmail)
		return nil, fmt.Errorf("unauthorized: session expired, please login again")
	}

	return &currentUser.Role, nil
}
