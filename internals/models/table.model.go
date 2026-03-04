package models

import (
	"time"

	"github.com/gofrs/uuid"
)

type TableState string

const (
	TableOccupied TableState = "occupied"
	TableEmpty    TableState = "empty"
	TableBooked   TableState = "booked"
)

type TableStatus struct {
	ID          uuid.UUID  `db:"id" json:"id"`                     // UUID primary key
	TableNumber int        `db:"table_number" json:"table_number"` // Table number
	Status      TableState `db:"status" json:"status"`             // Enum: occupied, empty, booked
	Capacity    int        `db:"capacity" json:"capacity"`         // Number of seats
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`     // Timestamp of creation
}

type CreateTableType struct {
	TableNumber int        `db:"table_number" json:"table_number"` // Table number
	Status      TableState `db:"status" json:"status"`             // Enum: occupied, empty, booked
	Capacity    int        `db:"capacity" json:"capacity"`         // Number of seats
}

type UpdateTableStatus struct {
	ID          uuid.UUID  `db:"id" json:"id"`                     // UUID primary key
	TableNumber int        `db:"table_number" json:"table_number"` // Table number
	Status      TableState `db:"status" json:"status"`             // Enum: occupied, empty, booked
	Capacity    int        `db:"capacity" json:"capacity"`         // Number of seats
}
