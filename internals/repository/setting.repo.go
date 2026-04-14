package repository

import (
	"context"
	"fmt"

	"github.com/Abhishekh669/backend/internals/database"
	"github.com/Abhishekh669/backend/internals/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PaymentRepo interface
type SettingRepo interface {
	GetRestaurantInformation(ctx context.Context) (*models.RestaurantSettings, error)
	UpdateRestaurantInformation(ctx context.Context, info *models.UpdateRestaurantSettings) error
}

// settingRepo struct
type settingRepo struct {
	pool *pgxpool.Pool
}

func (r *settingRepo) GetRestaurantInformation(ctx context.Context) (*models.RestaurantSettings, error) {
	query := `
		SELECT 
			id,
			name,
			slogan,
			logo_url,
			phone,
			email,
			address,
			country,
			state,
			city,
			created_at,
			updated_at
		FROM restaurant_information
		LIMIT 1
	`

	var info models.RestaurantSettings

	err := r.pool.QueryRow(ctx, query).Scan(
		&info.ID,
		&info.Name,
		&info.Slogan,
		&info.LogoURL,
		&info.Phone,
		&info.Email,
		&info.Address,
		&info.Country,
		&info.State,
		&info.City,
		&info.CreatedAt,
		&info.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get restaurant information: %w", err)
	}

	return &info, nil
}

func (r *settingRepo) UpdateRestaurantInformation(
	ctx context.Context,
	info *models.UpdateRestaurantSettings,
) error {

	query := `
		UPDATE restaurant_information
		SET 
			name = $1,
			slogan = $2,
			logo_url = $3,
			phone = $4,
			email = $5,
			address = $6,
			country = $7,
			state = $8,
			city = $9,
			updated_at = NOW()
		WHERE singleton_key = TRUE
	`

	result, err := r.pool.Exec(
		ctx,
		query,
		info.Name,
		info.Slogan,
		info.LogoURL,
		info.Phone,
		info.Email,
		info.Address,
		info.Country,
		info.State,
		info.City,
	)

	if err != nil {
		return fmt.Errorf("failed to update restaurant information: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("restaurant information not found")
	}

	return nil
}

func NewSettingRepository() SettingRepo {
	pool, err := database.GetPostgresPool()
	if err != nil {
		return nil
	}
	return &settingRepo{pool: pool}
}
