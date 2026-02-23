package algorithm

import (
	"github.com/gofrs/uuid"
)

type MenuCache struct {
	Categories       []Category
	CategoryChildren map[uuid.UUID][]uuid.UUID
	ItemsByCategory  map[uuid.UUID][]MenuItem
}

type Category struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	Name         string     `json:"name" db:"name"`
	Slug         string     `json:"slug" db:"slug"`
	ParentID     *uuid.UUID `json:"parent_id,omitempty" db:"parent_id"`
	Level        int        `json:"level" db:"level"`
	IsActive     bool       `json:"is_active" db:"is_active"`
	DisplayOrder int        `json:"display_order" db:"display_order"`
}

type MenuItem struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description *string   `json:"description,omitempty" db:"description"`
	Price       float64   `json:"price" db:"price"`
	CategoryID  uuid.UUID `json:"category_id" db:"category_id"`
	IsAvailable bool      `json:"is_available" db:"is_available"`
	ImageURL    *string   `json:"image_url,omitempty" db:"image_url"`
}
