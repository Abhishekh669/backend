package models

import (
	"time"

	"github.com/gofrs/uuid"
)

type OrderStatus string

const (
	OrderStatusNotApproved OrderStatus = "not-approved"
	OrderStatusApproved    OrderStatus = "approved"
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
	ID         uuid.UUID   `db:"id" json:"id"`
	OrderID    uuid.UUID   `db:"order_id" json:"order_id"`
	MenuItemID uuid.UUID   `db:"menu_item_id" json:"menu_item_id"`
	Quantity   float64     `db:"quantity" json:"quantity"`
	Price      float64     `db:"price" json:"price"`
	Status     OrderStatus `db:"status" json:"status"`
	CreatedAt  time.Time   `db:"created_at" json:"created_at"`
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
	Id        uuid.UUID   `json:"id"`
	Price     float64     `json:"price"`
	Quantity  float64     `json:"quantity"`
	OrderId   uuid.UUID   `json:"order_id"`
	MenuId    uuid.UUID   `json:"menu_id"`
	MenuImage *string     `json:"menu_image"`
	Status    OrderStatus `json:"status"`
	MenuName  string      `json:"menu_name"`
	CreatedAt time.Time   `json:"created_at"`
}

type CustomerOrderRequest struct {
	OrderId       uuid.UUID       `json:"id"`
	Status        OrderStatus     `json:"status"`
	Table         TableSession    `json:"table_session"`
	CustomerName  *string         `json:"customer_name"`
	CustomerPhone *string         `json:"customer_phone"`
	Note          *string         `json:"note"`
	OrderItems    []OrderItemType `json:"order_items"`
}

type CustomerApprovalRequest struct {
	TableNumber int    `json:"table_number"`
	Phone       string `db:"phone" json:"phone"`
}

type WaiterApprovalRequest struct {
	Id          uuid.UUID `json:"id"`
	WaiterId    uuid.UUID `json:"waiter_id"`
	TableNumber int       `json:"table_number"`
	Phone       string    `db:"phone" json:"phone"`
}

type TableValidation struct {
	ID          uuid.UUID  `db:"id" json:"id"`                     // UUID primary key
	TableNumber int        `db:"table_number" json:"table_number"` // INT NOT NULL
	PhoneNumber string     `db:"phone_number" json:"phone_number"` // TEXT NOT NULL
	WaiterID    *uuid.UUID `db:"waiter_id" json:"waiter_id"`       // Nullable UUID reference
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`     // TIMESTAMPTZ NOT NULL
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`     // TIMESTAMPTZ NOT NULL
}

type UpdateOrderItem struct {
	Status      OrderStatus `json:"status"`
	OrderItemId string      `json:"order_item_id"`
	OrderId     string      `json:"order_id"`
}
