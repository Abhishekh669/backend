package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Absent struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	UserID    primitive.ObjectID `bson:"userId" json:"userId"`
	Date      time.Time          `bson:"date" json:"date"`
	Reason    string             `bson:"reason,omitempty" json:"reason,omitempty"`
	Name      string             `bson:"name,omitempty" json:"name,omitempty"`
	Status    string             `bson:"status" json:"status"`
	CodeName  string             `bson:"codeName,omitempty" json:"codeName,omitempty"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt,omitempty" json:"updatedAt,omitempty"`
}

type AbsentRequest struct {
	CodeName string    `bson:"codeName,omitempty" json:"codeName,omitempty"`
	Date     time.Time `bson:"date" json:"date"`
	Name     string    `bson:"name,omitempty" json:"name,omitempty"`
	Reason   string    `bson:"reason,omitempty" json:"reason,omitempty"`
}
