package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Admin struct representing the Admin model in Go
type Admin struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Email     string             `json:"email" bson:"email"`
	Password  string             `json:"password" bson:"password"`
	UserId    primitive.ObjectID `json:"userId" bson:"userId"`                           // Foreign key reference to User model
	CreatedAt time.Time          `json:"createdAt,omitempty" bson:"createdAt,omitempty"` // Default to current time
}

type AdminRequest struct {
	Email    string `json:"email,omitempty" bson:"email,omitempty"`
	Password string `json:"password,omitempty" bson:"password,omitempty"`
	UserId   string `json:"userId,omitempty" bson:"userId,omitempty"` // Foreign key reference to User model
}
