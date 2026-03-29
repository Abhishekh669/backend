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
	Discount   float64 `json:"discount" db:"discount"`
}

type UserToken struct {
	ID uuid.UUID `json:"id" db:"id"`

	PhoneNumber string `json:"phone_number" db:"phone_number"`

	TotalTokens float64 `json:"total_tokens" db:"total_tokens"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
