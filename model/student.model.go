package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Student struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Image       string             `bson:"image,omitempty" json:"image,omitempty"`
	ImageURL    string             `bson:"image_url,omitempty" json:"image_url,omitempty"`
	CreatedAt   time.Time          `bson:"createdAt,omitempty" json:"createdAt,omitempty"`
	IsVerified  bool               `bson:"isVerified" json:"isVerified"`
	IsOnBoarded bool               `bson:"isOnBoarded" json:"isOnBoarded"`
	Name        string             `bson:"name" json:"name"`
	PhoneNumber string             `bson:"phoneNumber,omitempty" json:"phoneNumber,omitempty"`
	Email       string             `bson:"email,omitempty" json:"email,omitempty"`
	Address     string             `bson:"address,omitempty" json:"address,omitempty"`
	UserID      primitive.ObjectID `bson:"userId,omitempty" json:"userId,omitempty"`
}
