package models

import (
	"time"

	"github.com/gofrs/uuid"
)

type PaymentMethod string

const (
	PaymentMethodCash   PaymentMethod = "cash"
	PaymentMethodOnline PaymentMethod = "online"
)

type OnlineGateway string

const (
	GatewayEsewa   OnlineGateway = "esewa"
	GatewayKhalti  OnlineGateway = "khalti"
	GatewayFonepay OnlineGateway = "fonepay"
	GatewayBanking OnlineGateway = "banking"
	GatewayOther   OnlineGateway = "other"
)

type Payment struct {
	ID uuid.UUID `json:"id" db:"id"`

	OrderID uuid.UUID `json:"order_id" db:"order_id"`

	PaymentMethod PaymentMethod `json:"payment_method" db:"payment_method"`

	// Nullable field
	OnlineGateway *OnlineGateway `json:"online_gateway,omitempty" db:"online_gateway"`

	PaidAmount float64 `json:"paid_amount" db:"paid_amount"`
	Discount   float64 `json:"discount" db:"discount"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type CreatePayment struct {
	OrderID uuid.UUID `json:"order_id" db:"order_id"`

	PaymentMethod PaymentMethod `json:"payment_method" db:"payment_method"`

	// Nullable field
	OnlineGateway *OnlineGateway `json:"online_gateway,omitempty" db:"online_gateway"`

	PaidAmount float64 `json:"paid_amount" db:"paid_amount"`
}

type UserToken struct {
	ID uuid.UUID `json:"id" db:"id"`

	PhoneNumber string `json:"phone_number" db:"phone_number"`

	TotalTokens float64 `json:"total_tokens" db:"total_tokens"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type TokenTransactionType string

const (
	TokenEarn   TokenTransactionType = "EARN"
	TokenSpend  TokenTransactionType = "SPEND"
	TokenStreak TokenTransactionType = "STREAK"
)

type CustomerStreak struct {
	PhoneNumber   string     `json:"phone_number" db:"phone_number"`
	CurrentStreak int        `json:"current_streak" db:"current_streak"`
	LastVisit     *time.Time `json:"last_visit,omitempty" db:"last_visit"`
	MonthlyDays   int        `json:"monthly_days" db:"monthly_days"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

type TokenTransaction struct {
	ID          uuid.UUID            `json:"id" db:"id"`
	PhoneNumber string               `json:"phone_number" db:"phone_number"`
	Amount      float64              `json:"amount" db:"amount"`
	Type        TokenTransactionType `json:"type" db:"type"`
	Source      *string              `json:"source,omitempty" db:"source"`
	ReferenceID *uuid.UUID           `json:"reference_id,omitempty" db:"reference_id"`
	CreatedAt   time.Time            `json:"created_at" db:"created_at"`
}

type TokenDetailsOfCustomer struct {
	Token         *UserToken `json:"token_details,omitempty"`
	CurrentStreak int        `json:"current_streak"`

	LastVisit *time.Time `json:"last_visit,omitempty"`

	MonthlyDays int `json:"monthly_days"`

	Discount float64 `json:"discount"`
}

type PaymentDetailsForCashierWithDiscount struct {
	TokenDetails   *TokenDetailsOfCustomer `json:"token_details,omitempty"`
	OrderMenuItems []OrderItemType         `json:"order_menu_items"`
	OrderId        uuid.UUID               `json:"order_id"`
	Status         OrderStatus             `json:"status"`
	TableNumber    int                     `json:"table_number"`
	CustomerName   *string                 `json:"customer_name"`
	CustomerPhone  *string                 `json:"customer_phone"`
	WaiterId       uuid.UUID               `json:"waiter_id"`
	WaiterName     string                  `json:"waiter_name"`
	WaiterImage    *string                 `json:"waiter_image"`
}
