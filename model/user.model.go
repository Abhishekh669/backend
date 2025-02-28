package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// database related works

type User struct {
	ID            primitive.ObjectID   `json:"_id,omitempty" bson:"_id,omitempty"`
	Name          string               `json:"name" bson:"name"`
	Email         string               `json:"email" bson:"email"`
	Image         string               `json:"image,omitempty" bson:"image,omitempty"`
	CreatedAt     time.Time            `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
	Conversations []primitive.ObjectID `json:"conversations,omitempty" bson:"conversations,omitempty"`
	IsAdmin       bool                 `json:"isAdmin" bson:"isAdmin"`
	AllowCreate   bool                 `json:"allowCreate" bson:"allowCreate"`
	IsVerified    bool                 `json:"isVerified" bson:"isVerified"`
	IsOnBoarded   bool                 `json:"isOnBoarded" bson:"isOnBoarded"`

	//fields after onbaording
	CodeName       string `json:"codeName,omitempty" bson:"codeName,omitempty"`
	PhoneNumber    string `json:"phoneNumber,omitempty" bson:"phoneNumber,omitempty"`
	Address        string `json:"address,omitempty" bson:"address,omitempty"`
	Age            int    `json:"age" bson:"age"`
	Qualification  string `json:"qualification,omitempty" bson:"qualification,omitempty"`
	MainField      string `json:"mainField,omitempty" bson:"mainField,omitempty"`
	Field          string `json:"field,omitempty" bson:"field,omitempty"`
	ChatPublicKey  string `json:"chatPublicKey,omitempty" bson:"chatPublicKey,omitempty"`
	ChatPrivateKey string `json:"chatPrivateKey,omitempty" bson:"chatPrivateKey,omitempty"`
}

type OnboardingRequest struct {
	CodeName      string `json:"codeName,omitempty" bson:"codeName,omitempty"`
	PhoneNumber   string `json:"phoneNumber,omitempty" bson:"phoneNumber,omitempty"`
	Address       string `json:"address,omitempty" bson:"address,omitempty"`
	Age           int    `json:"age" bson:"age"`
	Qualification string `json:"qualification,omitempty" bson:"qualification,omitempty"`
	Field         string `json:"field,omitempty" bson:"field,omitempty"`
	MainField     string `json:"mainField,omitempty" bson:"mainField,omitempty"`
}
