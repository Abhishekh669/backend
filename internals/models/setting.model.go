package models

import (
	"time"

	"github.com/gofrs/uuid"
)

type RestaurantSettings struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Slogan    *string   `json:"slogan" db:"slogan"`
	LogoURL   *string   `json:"logo_url" db:"logo_url"`
	Phone     *string   `json:"phone" db:"phone"`
	Email     *string   `json:"email" db:"email"`
	Address   *string   `json:"address" db:"address"`
	Country   *string   `json:"country" db:"country"`
	State     *string   `json:"state" db:"state"`
	City      *string   `json:"city" db:"city"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type CreateRestaurantSettings struct {
	Name    string  `json:"name" db:"name"`
	Slogan  *string `json:"slogan" db:"slogan"`
	LogoURL *string `json:"logo_url" db:"logo_url"`
	Phone   *string `json:"phone" db:"phone"`
	Email   *string `json:"email" db:"email"`
	Address *string `json:"address" db:"address"`
	Country *string `json:"country" db:"country"`
	State   *string `json:"state" db:"state"`
	City    *string `json:"city" db:"city"`
}

type UpdateRestaurantSettings struct {
	ID      uuid.UUID `json:"id" db:"id"`
	Name    string    `json:"name" db:"name"`
	Slogan  *string   `json:"slogan" db:"slogan"`
	LogoURL *string   `json:"logo_url" db:"logo_url"`
	Phone   *string   `json:"phone" db:"phone"`
	Email   *string   `json:"email" db:"email"`
	Address *string   `json:"address" db:"address"`
	Country *string   `json:"country" db:"country"`
	State   *string   `json:"state" db:"state"`
	City    *string   `json:"city" db:"city"`
}
