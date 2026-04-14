package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Abhishekh669/backend/internals/config"
	"github.com/Abhishekh669/backend/internals/database"
	"github.com/Abhishekh669/backend/internals/lib"
	"github.com/Abhishekh669/backend/internals/models"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserListResponse struct {
	Users      []models.SafeUserType `json:"users"`
	Total      int                   `json:"total"`
	HasMore    bool                  `json:"has_more"`
	NextOffset int                   `json:"next_offset"`
	UserData   UserStats             `json:"user_data"`
}

type UserStats struct {
	TotalUsers  int `json:"total_users"`
	ActiveUsers int `json:"active_users"`

	// Gender counts
	MaleCount   int `json:"male_count"`
	FemaleCount int `json:"female_count"`
	OtherCount  int `json:"other_count"`

	// Role counts
	AdminCount         int `json:"admin_count"`
	ChefCount          int `json:"chef_count"`
	WaiterCount        int `json:"waiter_count"`
	CashierCount       int `json:"cashier_count"`
	DeliveryStaffCount int `json:"delivery_staff_count"`
	ManagerCount       int `json:"manager_count"`
	CustomerCount      int `json:"customer_count"`

	// Recent users
	RecentUsersWeekly int `json:"recent_users_weekly"`
}

// If you want to include actual users from last week

type UserRepo interface {
	UpdateExistingPasswordSession(ctx context.Context, sessionId uuid.UUID, token, pin string) error
	GetForgetPasswordSessionByEmail(ctx context.Context, email string) (*models.PasswordResetRequest, error)
	CleanupExpiredNUsedForgetPasswordSessions(ctx context.Context) error
	MarkForgetPasswordSessionUsed(ctx context.Context, sessionId uuid.UUID) error
	GetForgetPasswordSession(ctx context.Context, email string, token string) (*models.PasswordResetRequest, error)
	CreateForgetPasswordSession(ctx context.Context, email string, token string, pin string) error
	UpdateUserPassword(ctx context.Context, userId uuid.UUID, newPassword string) error
	GetUserDataByName(ctx context.Context, userName string) ([]models.UserTypeForAttendance, error)
	UpdateUser(ctx context.Context, user *models.UpdateUserType) error
	DeleteUser(c context.Context, userIds []string, requesterRole models.Role) error
	CreateNewUser(ctx context.Context, user *models.CreateUserType) error
	GetUserStats(ctx context.Context) (*UserStats, error)
	GetAllUsers(ctx context.Context, searchTerm string, limit, offset int, oldFirstBool bool) (*UserListResponse, error)
	EnsureAdminUserExists(ctx context.Context) error
	GetUserInfoByEmail(email string, ctx context.Context) (*models.UserType, error)
	GetUserById(id string, ctx context.Context) (*models.UserType, error)
	LoginUser(email, password string, ctx context.Context) (*models.UserType, error)
}

type userRepo struct {
	pool *pgxpool.Pool
}

var (
	ErrEmailExists = errors.New("email already exists")
	ErrPhoneExists = errors.New("phone number already exists")
)

func (r *userRepo) CleanupExpiredNUsedForgetPasswordSessions(ctx context.Context) error {

	query := `
		DELETE FROM password_reset_requests
		WHERE created_at < NOW() - INTERVAL '15 minutes'
		   OR is_used = TRUE
	`

	result, err := r.pool.Exec(ctx, query)
	if err != nil {
		return err
	}

	rows := result.RowsAffected()

	log.Printf("🧹 cleaned up %d expired/used password reset sessions", rows)

	return nil
}

func (r *userRepo) MarkForgetPasswordSessionUsed(ctx context.Context, sessionId uuid.UUID) error {
	query := `
		UPDATE password_reset_requests
		SET is_used = true, updated_at = NOW()
		WHERE id = $1 AND is_used = FALSE
	`

	result, err := r.pool.Exec(ctx, query, sessionId)
	if err != nil {
		return fmt.Errorf("failed to mark session as used: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("session not found")
	}

	return nil
}

func (r *userRepo) UpdateExistingPasswordSession(ctx context.Context, sessionId uuid.UUID, token, pin string) error {
	query := `
		UPDATE password_reset_requests
		SET session_token =$1, pin_code = $2, updated_at = NOW()
		WHERE id = $3
	`

	result, err := r.pool.Exec(ctx, query, token, pin, sessionId)
	if err != nil {
		return fmt.Errorf("failed to mark session as used: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("session not found")
	}

	return nil
}

func (r *userRepo) GetForgetPasswordSessionByEmail(ctx context.Context, email string) (*models.PasswordResetRequest, error) {

	query := `
		SELECT
			id,
			email,
			session_token,
			pin_code,
			is_used,
			updated_at,
			created_at
		FROM password_reset_requests
		WHERE email = $1
 	 AND is_used = FALSE
	ORDER BY created_at DESC
	LIMIT 1
	`

	var session models.PasswordResetRequest

	err := r.pool.QueryRow(ctx, query, email).Scan(
		&session.ID,
		&session.Email,
		&session.SessionToken,
		&session.PinCode,
		&session.IsUsed,
		&session.UpdatedAt,
		&session.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // not found (clean handling)
		}
		return nil, err
	}

	return &session, nil
}

func (r *userRepo) GetForgetPasswordSession(ctx context.Context, email string, token string) (*models.PasswordResetRequest, error) {
	query := `
		SELECT
			id,
			email,
			session_token,
			pin_code,
			is_used,
			updated_at,
			created_at
		FROM password_reset_requests
		WHERE email = $1
		  AND session_token = $2
	
		LIMIT 1
	`

	var session models.PasswordResetRequest

	err := r.pool.QueryRow(ctx, query, email, token).Scan(
		&session.ID,
		&session.Email,
		&session.SessionToken,
		&session.PinCode,
		&session.IsUsed,
		&session.UpdatedAt,
		&session.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // not found (clean handling)
		}
		return nil, err
	}

	return &session, nil
}

func (r *userRepo) CreateForgetPasswordSession(ctx context.Context, email string, token string, pin string) error {
	fmt.Println("thisis hte pin hoitw : ", pin)
	query := `
		INSERT INTO password_reset_requests (
			email,
			session_token,
			pin_code,
			created_at
		)
		VALUES ($1, $2, $3, NOW())
	`

	_, err := r.pool.Exec(ctx, query,
		email,
		token,
		pin, // ensures 6-digit format (e.g., 000123)
	)

	if err != nil {
		log.Printf("❌ failed to create forget password session: %v", err)
		return errors.New("failed to create password reset session")
	}

	return nil
}

func (r *userRepo) UpdateUserPassword(ctx context.Context, userId uuid.UUID, newPassword string) error {

	query := `
		UPDATE users
		SET 
			password = $1,
			last_password_reset_at = $2,
			updated_at = NOW()
		WHERE id = $3
	`

	result, err := r.pool.Exec(ctx, query, newPassword, time.Now().UnixMilli(), userId)
	if err != nil {
		return fmt.Errorf("failed to update user password: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func (r *userRepo) GetUserDataByName(ctx context.Context, userName string) ([]models.UserTypeForAttendance, error) {
	// Define the query to fetch user data
	fmt.Println("this is search term: ", userName)

	query := `
        SELECT 
			id,
            name,
            email,
            phone,
            is_active,
            image
        FROM users
        WHERE 
            name ILIKE $1 OR 
            email ILIKE $1 OR 
            phone ILIKE $1
        ORDER BY name
    `

	// Execute the query with wildcard search
	searchPattern := "%" + userName + "%"
	rows, err := r.pool.Query(ctx, query, searchPattern)
	if err != nil {
		return nil, fmt.Errorf("error querying users by name, email, or phone: %w", err)
	}
	defer rows.Close()

	// Slice to hold the results
	var users []models.UserTypeForAttendance

	// Iterate through the rows
	for rows.Next() {
		var user models.UserTypeForAttendance

		// Scan the row into the user struct
		err := rows.Scan(
			&user.Id,
			&user.Name,
			&user.Email,
			&user.Phone,
			&user.IsActive,
			&user.Image,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning user row: %w", err)
		}

		users = append(users, user)
	}

	fmt.Println("this is users found: ", users)

	// Check for errors from iterating over rows
	if rows.Err() != nil {
		return nil, fmt.Errorf("error iterating user rows: %w", rows.Err())
	}

	return users, nil
}

func (r *userRepo) UpdateUser(ctx context.Context, user *models.UpdateUserType) error {
	var (
		ErrEmailExists = errors.New("email already exists")
		ErrPhoneExists = errors.New("phone number already exists")
	)
	if user.Id == "" {
		return fmt.Errorf("user ID is required")
	}
	if user.Email == "" {
		return fmt.Errorf("email is required")
	}
	if user.Gender == "" {
		return fmt.Errorf("gender is required")
	}
	if user.Phone == "" {
		return fmt.Errorf("phone number is required")
	}
	if user.Role == "" {
		return fmt.Errorf("role is required")
	}

	query := `
		UPDATE users
		SET
			name       = $1,
			email      = $2,
			phone      = $3,
			role       = $4,
			gender     = $5,
			salary     = $6,
			is_active  = $7,
			image      = $8,
			updated_at = $9
		WHERE id = $10
	`

	res, err := r.pool.Exec(ctx, query,
		user.Name,
		user.Email,
		user.Phone,
		user.Role,
		user.Gender,
		user.Salary,
		user.IsActive,
		user.Image,
		time.Now(),
		user.Id,
	)

	if err != nil {
		// 🔥 Postgres error handling
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == "23505" {
				switch pgErr.ConstraintName {
				case "users_email_key":
					return ErrEmailExists
				case "users_phone_key":
					return ErrPhoneExists
				default:
					return fmt.Errorf("duplicate value violates unique constraint")
				}
			}
		}
		return fmt.Errorf("failed to update user")
	}

	if res.RowsAffected() == 0 {
		return fmt.Errorf("no user updated, invalid user ID")
	}

	return nil
}

func (r *userRepo) DeleteUser(ctx context.Context, userIDs []string, requesterRole models.Role) error {
	if len(userIDs) == 0 {
		return fmt.Errorf("no user selected")
	}

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		log.Printf("failed to begin transaction: %v", err)
		return fmt.Errorf("failed to delete user")
	}
	defer tx.Rollback(ctx)

	// 🔎 Check roles of users being deleted
	rows, err := tx.Query(ctx, `SELECT id, role FROM users WHERE id = ANY($1)`, userIDs)
	if err != nil {
		log.Printf("failed to fetch user roles: %v", err)
		return fmt.Errorf("failed to delete user")
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		var role models.Role
		if err := rows.Scan(&id, &role); err != nil {
			return fmt.Errorf("failed to delete user")
		}

		// 🚫 Nobody can delete an admin
		if role == models.RoleAdmin {
			return fmt.Errorf("admin users cannot be deleted")
		}

		// 🚫 Manager cannot delete another manager
		if role == models.RoleManager && requesterRole != models.RoleAdmin {
			return fmt.Errorf("only admin can delete a manager")
		}
	}

	// ✅ Soft delete instead of hard delete
	res, err := tx.Exec(ctx,
		`UPDATE users SET is_active = false WHERE id = ANY($1)`,
		userIDs,
	)
	if err != nil {
		log.Printf("failed to deactivate users: %v", err)
		return fmt.Errorf("failed to delete user")
	}

	if res.RowsAffected() == 0 {
		return fmt.Errorf("no users were deleted")
	}

	if err := tx.Commit(ctx); err != nil {
		log.Printf("failed to commit transaction: %v", err)
		return fmt.Errorf("failed to delete user")
	}

	return nil
}

func (r *userRepo) CreateNewUser(
	ctx context.Context,
	user *models.CreateUserType,
) error {

	failedMessage := "failed to create user"

	hashPassword, err := lib.HashPassword("password123")
	if err != nil || hashPassword == "" {
		return errors.New(failedMessage)
	}

	query := `
		INSERT INTO users (
			name,
			email,
			image,
			password,
			role,
			gender,
			phone,
			salary,
			is_active,
			last_password_reset_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8,$9, $10)
	`

	_, err = r.pool.Exec(
		ctx,
		query,
		user.Name,
		user.Email,
		user.Image,
		hashPassword,
		user.Role,
		user.Gender,
		user.Phone,
		user.Salary,
		false,
		time.Now().UnixMilli(),
	)

	if err != nil {
		// 🔥 pgx postgres error
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return errors.New("user already exists")
			}
		}
		return errors.New(failedMessage)
	}

	return nil
}

func (r *userRepo) GetUserStats(ctx context.Context) (*UserStats, error) {
	statsQuery := `
		SELECT 
			COUNT(*) as total_users,
			COUNT(CASE WHEN is_active THEN 1 END) as active_users,
			
			COUNT(CASE WHEN gender = 'male' THEN 1 END) as male_count,
			COUNT(CASE WHEN gender = 'female' THEN 1 END) as female_count,
			COUNT(CASE WHEN gender = 'other' THEN 1 END) as other_count,
			
			COUNT(CASE WHEN role = 'admin' THEN 1 END) as admin_count,
			COUNT(CASE WHEN role = 'chef' THEN 1 END) as chef_count,
			COUNT(CASE WHEN role = 'waiter' THEN 1 END) as waiter_count,
			COUNT(CASE WHEN role = 'cashier' THEN 1 END) as cashier_count,
			COUNT(CASE WHEN role = 'delivery_staff' THEN 1 END) as delivery_staff_count,
			COUNT(CASE WHEN role = 'manager' THEN 1 END) as manager_count,
			COUNT(CASE WHEN role = 'customer' THEN 1 END) as customer_count,
			
			COUNT(CASE WHEN created_at >= NOW() - INTERVAL '7 days' THEN 1 END) as recent_users_weekly
		FROM users`

	var stats UserStats
	err := r.pool.QueryRow(ctx, statsQuery).Scan(
		&stats.TotalUsers,
		&stats.ActiveUsers,
		&stats.MaleCount,
		&stats.FemaleCount,
		&stats.OtherCount,
		&stats.AdminCount,
		&stats.ChefCount,
		&stats.WaiterCount,
		&stats.CashierCount,
		&stats.DeliveryStaffCount,
		&stats.ManagerCount,
		&stats.CustomerCount,
		&stats.RecentUsersWeekly,
	)

	if err != nil {
		return nil, err
	}

	return &stats, nil
}

func (r *userRepo) GetAllUsers(ctx context.Context, search string, limit, page int, oldFirst bool) (*UserListResponse, error) {
	var (
		users []models.SafeUserType
		total int
	)

	failedMessage := "failed to get users"

	offset := page * limit
	orderBy := "DESC"
	if oldFirst {
		orderBy = "ASC"
	}

	// Base query
	query := `
        SELECT id, email, gender, image, is_active, role, name, phone, salary, created_at, updated_at
        FROM users
    `

	// Add search if provided (safe because we use parameterized query)
	var args []any
	if search != "" {
		query += " WHERE name ILIKE $1 OR email ILIKE $1 OR phone ILIKE $1"
		searchTerm := "%" + search + "%"
		args = append(args, searchTerm)
	}

	// Add ordering, limit, offset (safe to inject integers)
	query += fmt.Sprintf(" ORDER BY created_at %s LIMIT %d OFFSET %d", orderBy, limit, offset)

	// Execute query
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, errors.New(failedMessage)
	}
	defer rows.Close()

	for rows.Next() {
		var u models.SafeUserType
		if err := rows.Scan(
			&u.Id, &u.Email, &u.Gender, &u.Image, &u.IsActive,
			&u.Role, &u.Name, &u.Phone, &u.Salary,
			&u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	if rows.Err() != nil {
		return nil, errors.New(failedMessage)
	}

	// Count total users (with same search filter)
	countQuery := "SELECT COUNT(*) FROM users"
	if search != "" {
		countQuery += " WHERE name ILIKE $1 OR email ILIKE $1 OR phone ILIKE $1"
	}

	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, errors.New(failedMessage)
	}

	hasMore := (page+1)*limit < total
	nextPage := page + 1

	stats, err := r.GetUserStats(ctx)
	if err != nil {
		// Optionally: return users even if stats fail, or handle error
		// For now, we'll return partial data with error
		return nil, errors.New("failed to get user statistics")
	}
	return &UserListResponse{
		Users:      users,
		Total:      total,
		UserData:   *stats,
		HasMore:    hasMore,
		NextOffset: nextPage,
	}, nil
}

func (r *userRepo) EnsureAdminUserExists(ctx context.Context) error {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE role='admin')`).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return nil // admin already exists
	}

	// Get initial password from ENV
	initialPassword := config.AppConfig.AdminPassword
	hashedPassword, _ := lib.HashPassword(initialPassword)

	_, err = r.pool.Exec(ctx, `
		INSERT INTO users (name, email, password, role, gender, phone, is_active, last_password_reset_at )
		VALUES ($1, $2, $3, 'admin', 'male', $4, $5, $6)
	`, "System Admin", config.AppConfig.AdminEmail, hashedPassword, config.AppConfig.AdminPhone, true, time.Now().UnixMilli())
	return err
}

func (r *userRepo) GetUserInfoByEmail(email string, ctx context.Context) (*models.UserType, error) {
	query := `
		SELECT id, email, gender, image, is_active, last_password_reset_at,
		       role, name, phone, password, salary, created_at, updated_at
		FROM users
		WHERE email=$1
		LIMIT 1
	`

	var user models.UserType

	conn, err := r.pool.Acquire(ctx)
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

func (r *userRepo) GetUserById(id string, ctx context.Context) (*models.UserType, error) {
	query := `
		SELECT id, email, gender, image, is_active, last_password_reset_at,
		       role, name, phone, password, salary, created_at, updated_at
		FROM users
		WHERE id=$1
		LIMIT 1
	`

	var user models.UserType

	conn, err := r.pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire DB connection: %w", err)
	}
	defer conn.Release()

	err = conn.QueryRow(ctx, query, id).Scan(
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

func (r *userRepo) LoginUser(email, password string, ctx context.Context) (*models.UserType, error) {
	failedMessage := "incorrect credentials"
	if r.pool == nil {
		return nil, fmt.Errorf("failed to connect with server")
	}

	if email == "" || password == "" {
		log.Println("invalid login credentials")
		return nil, errors.New(failedMessage)
	}

	dbUser, err := r.GetUserInfoByEmail(email, ctx)
	if err != nil {
		log.Println("failed to get user info: %w", err)
		return nil, errors.New(failedMessage)
	}
	if dbUser == nil {
		log.Println("user not found")
		return nil, errors.New(failedMessage)
	}

	fmt.Println("this isthe db user : ", dbUser)

	status, err := lib.CheckPasswordHash(password, dbUser.Password)
	if err != nil {
		log.Println("failed to check password hash: %w", err)
		return nil, errors.New(failedMessage)
	}
	if !status {
		log.Println("failed to hash the user passsword : %w", err)
		return nil, errors.New(failedMessage)
	}

	return dbUser, nil
}

func NewUserRepository() UserRepo {
	pool, err := database.GetPostgresPool()

	if err != nil {
		return nil
	}

	return &userRepo{
		pool: pool,
	}
}
