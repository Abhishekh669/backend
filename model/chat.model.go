// package model

// import (
// 	"time"

// 	"go.mongodb.org/mongo-driver/bson/primitive"
// )

// type Chat struct {
// 	ID            primitive.ObjectID   `json:"_id,omitempty" bson:"_id,omitempty"`
// 	Members       []primitive.ObjectID `json:"members" bson:"members"`
// 	IsGroup       bool                 `json:"isGroup,omitempty" bson:"isGroup,omitempty"`
// 	Name          string               `json:"image,omitempty" bson:"image,omitempty"`
// 	GroupImage    string               `json:"groupImage,omitempty" bson:"groupImage,omitempty"`
// 	Messages      []primitive.ObjectID `json:"messages,omitempty" bson:"messages,omitempty"`
// 	CreatedAt     time.Time            `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
// 	LastMessageAt time.Time            `json:"lastMessageAt,omitempty" bson:"lastMessageAT,omitempty"`
// }

// type ChatRequest struct {
// 	Members    []string `json:"members" bson:"members"`
// 	Name       string   `json:"name,omitempty" bson:"name,omitempty"`
// 	GroupImage string   `json:"groupImage,omitempty" bson:"groupImage,omitempty"`
// }

package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Chat struct {
	ID            primitive.ObjectID   `json:"_id,omitempty" bson:"_id,omitempty"`
	Members       []primitive.ObjectID `json:"members" bson:"members"`
	CreatedAt     time.Time            `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
	LastMessageAt time.Time            `json:"lastMessageAt,omitempty" bson:"lastMessageAT,omitempty"`
}

type ChatRequest struct {
	Members []string `json:"members" bson:"members"`
}
