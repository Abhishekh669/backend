package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TimerWorkspace struct {
	ID          primitive.ObjectID   `json:"_id,omitempty" bson:"_id,omitempty"`
	UserId      primitive.ObjectID   `json:"userId,omitempty" bson:"userId,omitempty"`
	Name        string               `json:"name,omitempty" bson:"name,omitempty"`
	Description string               `json:"description,omitempty" bson:"description,omitempty"`
	TimerIds    []primitive.ObjectID `json:"timerIds,omitempty" bson:"timerIds,omitempty"`
	CreatedAt   time.Time            `json:"createdAt" bson:"createdAt"`
}

type TimerWorkspaceRequest struct {
	UserId      string `json:"userId,omitempty" bson:"userId,omitempty"`
	Name        string `json:"name,omitempty" bson:"name,omitempty"`
	Description string `json:"description,omitempty" bson:"description,omitempty"`
}

type Timer struct {
	ID               primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	UserId           primitive.ObjectID `json:"userId,omitempty" bson:"userId,omitempty"`
	TimerWorkspaceId primitive.ObjectID `json:"timerWorkspaceId,omitempty" bson:"timerWorkspaceId,omitempty"`
	Name             string             `json:"name,omitempty" bson:"name,omitempty"`
	EndDate          time.Time          `json:"endDate,omitempty" bson:"endDate,omitempty"`
	Type             string             `json:"type,omitempty" bson:"type,omitempty"`
	CreatedAt        time.Time          `json:"createdAt" bson:"createdAt"`
}

type TimerRequest struct {
	TimerWorkspaceId string             `json:"timerWorkspaceId,omitempty" bson:"timerWorkspaceId,omitempty"`
	UserId           primitive.ObjectID `json:"userId,omitempty" bson:"userId,omitempty"`
	Name             string             `json:"name,omitempty" bson:"name,omitempty"`
	EndDate          time.Time          `json:"endDate,omitempty" bson:"endDate,omitempty"`
	Type             string             `json:"type,omitempty" bson:"type,omitempty"`
}
