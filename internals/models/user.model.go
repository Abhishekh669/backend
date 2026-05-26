package models

import (
	"errors"
	"time"

	"github.com/gofrs/uuid"
)

type Gender string

const (
	GenderMale   Gender = "male"
	GenderFemale Gender = "female"
	GenderOther  Gender = "other"
)

type Role string

const (
	RoleAdmin         Role = "admin"
	RoleChef          Role = "chef"
	RoleWaiter        Role = "waiter"
	RoleCashier       Role = "cashier"
	RoleDeliveryStaff Role = "delivery_staff"
	RoleManager       Role = "manager"
	RoleCustomer      Role = "customer"
)

type CustomerFeedback struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Phone     *string   `json:"phone,omitempty" db:"phone"`
	Name      string    `json:"name" db:"name"`
	Text      string    `json:"text" db:"text"`
	Rating    int       `json:"rating" db:"rating"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type CreateCustomerFeedback struct {
	Phone  *string `json:"phone,omitempty" db:"phone"`
	Name   string  `json:"name" db:"name"`
	Text   string  `json:"text" db:"text"`
	Rating int     `json:"rating" db:"rating"`
}

type PasswordResetRequest struct {
	ID           uuid.UUID `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	SessionToken string    `json:"session_token" db:"session_token"`
	PinCode      string    `json:"pin_code" db:"pin_code"`

	IsUsed    bool      `json:"is_used" db:"is_used"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type UserType struct {
	Id                  string    `json:"id" db:"id"`
	Email               string    `json:"email" db:"email"`
	Gender              Gender    `json:"gender" db:"gender"`
	Image               *string   `json:"image,omitempty" db:"image"`
	IsActive            bool      `json:"is_active" db:"is_active"`
	LastPasswordResetAt int64     `json:"last_password_reset_at" db:"last_password_reset_at"`
	Role                Role      `json:"role" db:"role"`
	Name                string    `json:"name" db:"name"`
	Phone               string    `json:"phone" db:"phone"`
	Password            string    `json:"-" db:"password"`
	Salary              float64   `json:"salary" db:"salary"`
	CreatedAt           time.Time `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time `json:"updated_at" db:"updated_at"`
}

type SafeUserType struct {
	Id        string    `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	Gender    Gender    `json:"gender" db:"gender"`
	Image     *string   `json:"image,omitempty" db:"image"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	Role      Role      `json:"role" db:"role"`
	Name      string    `json:"name" db:"name"`
	Phone     string    `json:"phone" db:"phone"`
	Salary    float64   `json:"salary" db:"salary"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type UserLogin struct {
	Email    string `json:"email" db:"email"`
	Password string `json:"password" db:"password"`
}

type CreateUserType struct {
	Name   string  `json:"name"`
	Email  string  `json:"email"`
	Phone  string  `json:"phone"`
	Gender Gender  `json:"gender"`
	Role   Role    `json:"role"`
	Salary float64 `json:"salary"`
	Image  *string `json:"image,omitempty"`
}

type UpdateUserType struct {
	Id       string  `json:"id"`
	Name     string  `json:"name"`
	Email    string  `json:"email"`
	Phone    string  `json:"phone"`
	Gender   Gender  `json:"gender"`
	Role     Role    `json:"role"`
	Salary   float64 `json:"salary"`
	Image    *string `json:"image,omitempty"`
	IsActive bool    `json:"is_active"`
}

type DeleteUserPayload struct {
	UserIds []string `json:"userIds"`
}

func ParseRole(roleStr string) (Role, error) {
	switch roleStr {
	case string(RoleAdmin),
		string(RoleManager),
		string(RoleCashier),
		string(RoleChef),
		string(RoleWaiter),
		string(RoleCustomer),
		string(RoleDeliveryStaff):
		return Role(roleStr), nil
	default:
		return "", errors.New("invalid role")
	}
}

type UserTypeForAttendance struct {
	Id       string  `json:"id"`
	Name     string  `json:"name"`
	Email    string  `json:"email"`
	Phone    string  `json:"phone"`
	IsActive bool    `json:"is_active"`
	Image    *string `json:"image,omitempty"`
}
