package models

import (
	"time"

	"github.com/gofrs/uuid"
)

type OrderStatus string

const (
	OrderStatusNotApproved OrderStatus = "not-approved"
	OrderStatusPending     OrderStatus = "pending"
	OrderStatusProgress    OrderStatus = "progress"
	OrderStatusCompleted   OrderStatus = "completed"
	OrderStatusCancelled   OrderStatus = "cancelled"
)

type TableSession struct {
	ID          uuid.UUID  `db:"id" json:"id"`
	TableNumber int        `db:"table_number" json:"table_number"`
	OpenTime    time.Time  `db:"open_time" json:"open_time"`
	CloseTime   *time.Time `db:"close_time" json:"close_time,omitempty"`
	Status      TableState `db:"status" json:"status"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

type Order struct {
	ID             uuid.UUID   `db:"id" json:"id"`
	TableSessionID uuid.UUID   `db:"table_session_id" json:"table_session_id"`
	CustomerName   *string     `db:"customer_name" json:"customer_name,omitempty"`
	CustomerPhone  *string     `db:"customer_phone" json:"customer_phone,omitempty"`
	WaiterId       *uuid.UUID  `db:"waiter_id" json:"waiter_id,omitempty"` // Changed to pointer for nullable
	Note           *string     `db:"note" json:"note,omitempty"`
	Status         OrderStatus `db:"status" json:"status"`
	CreatedAt      time.Time   `db:"created_at" json:"created_at"`
}

type OrderItem struct {
	ID         uuid.UUID `db:"id" json:"id"`
	OrderID    uuid.UUID `db:"order_id" json:"order_id"`
	MenuItemID uuid.UUID `db:"menu_item_id" json:"menu_item_id"`
	Quantity   float64   `db:"quantity" json:"quantity"`
	Price      float64   `db:"price" json:"price"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

type CreateOrderMenuItems struct {
	MenuItemID uuid.UUID `db:"menu_item_id" json:"menu_item_id"`
	Quantity   float64   `db:"quantity" json:"quantity"`
	Price      float64   `db:"price" json:"price"`
}

type CreateCustomerOrderRequest struct {
	TableNumber    int                    `json:"table_number"`
	CustomerName   *string                `db:"customer_name" json:"customer_name,omitempty"`
	CustomerPhone  *string                `db:"customer_phone" json:"customer_phone,omitempty"`
	Note           *string                `db:"note" json:"note,omitempty"`
	OrderMenuItems []CreateOrderMenuItems `db:"order_menu_items" json:"order_menu_items"`
}

type ApproveOrderItem struct {
	ID         uuid.UUID `db:"id" json:"id"`
	OrderID    uuid.UUID `db:"order_id" json:"order_id"`
	MenuItemID uuid.UUID `db:"menu_item_id" json:"menu_item_id"`
	Quantity   float64   `db:"quantity" json:"quantity"`
	Price      float64   `db:"price" json:"price"`
	HasChanged bool      `db:"has_changed" json:"has_changed"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

type ApproveOrderType struct {
	ID             uuid.UUID          `db:"id" json:"id"`
	TableSessionID uuid.UUID          `db:"table_session_id" json:"table_session_id"`
	CustomerName   *string            `db:"customer_name" json:"customer_name,omitempty"`
	CustomerPhone  *string            `db:"customer_phone" json:"customer_phone,omitempty"`
	WaiterId       *uuid.UUID         `db:"waiter_id" json:"waiter_id,omitempty"` // Changed to pointer for nullable
	Note           *string            `db:"note" json:"note,omitempty"`
	TableNumber    int                `db:"table_number" json:"table_number"`
	OrderMenuItems []ApproveOrderItem `db:"order_menu_items" json:"order_menu_items"`
}

type OrderItemType struct {
	Id        uuid.UUID `json:"id"`
	Price     float64   `json:"price"`
	Quantity  float64   `json:"quantity"`
	OrderId   uuid.UUID `json:"order_id"`
	MenuId    uuid.UUID `json:"menu_id"`
	MenuImage *string   `json:"menu_image"`
	MenuName  string    `json:"menu_name"`
}

type CustomerOrderRequest struct {
	Table         TableSession    `json:"table_session"`
	CustomerName  *string         `json:"customer_name"`
	CustomerPhone *string         `json:"customer_phone"`
	Note          *string         `json:"note"`
	OrderItems    []OrderItemType `json:"order_items"`
}
