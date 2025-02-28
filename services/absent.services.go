package services

import (
	"context"
	"fmt"
	"time"

	"github.com/Abhishekh669/backend/configuration"
	"github.com/Abhishekh669/backend/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UpdateAbsentRequestType struct {
	ID     string `json:"_id" bson:"_id"`
	Reason string `bson:"reason,omitempty" json:"reason,omitempty"`
	Date   string `bson:"date,omitempty" json:"date,omitempty"`
}

func UpdateAbsentRequest(id primitive.ObjectID, reason string, date time.Time) (bool, error) {
	absent_collection, err := configuration.GetCollection("absent")
	if err != nil {
		return false, fmt.Errorf("failed to get the absent collection")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	updateData, err := absent_collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{
			"$set": bson.M{
				"reason":    reason,
				"updatedAt": time.Now(),
				"date":      date,
			},
		},
	)

	if err != nil {
		return false, fmt.Errorf("failed to update absent request: %v", err)
	}

	// Check if a document was actually modified
	if updateData.MatchedCount == 0 {
		return false, fmt.Errorf("no absent request found with the given ID")
	}

	return true, nil

}

func GetAbsentRequestById(id primitive.ObjectID, userId primitive.ObjectID) (model.Absent, error) {
	absent_collection, err := configuration.GetCollection("absent")
	if err != nil {
		return model.Absent{}, fmt.Errorf("failed to get the absent collection")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"_id":    id,
		"userId": userId,
	}

	var absentRequest model.Absent

	err = absent_collection.FindOne(ctx, filter).Decode(&absentRequest)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return model.Absent{}, fmt.Errorf("no absent request found : %v", err)
		} else {
			return model.Absent{}, fmt.Errorf("failed to find the absent request")
		}
	}

	return absentRequest, nil

}

func GetUserAllAbsentResults(userId primitive.ObjectID) ([]model.Absent, error) {
	absent_collection, err := configuration.GetCollection("absent")
	if err != nil {
		return []model.Absent{}, fmt.Errorf("failed to get the absetn collection")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var absentResults []model.Absent
	filter := bson.M{
		"userId": userId,
	}
	options := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})

	cursor, err := absent_collection.Find(ctx, filter, options)
	if err != nil {
		return []model.Absent{}, fmt.Errorf("faild to get absent data")
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var absentResult model.Absent
		if err := cursor.Decode(&absentResult); err != nil {
			return []model.Absent{}, fmt.Errorf("faield to decode workspace")
		}
		absentResults = append(absentResults, absentResult)
	}

	if err := cursor.Err(); err != nil {
		return []model.Absent{}, fmt.Errorf("cursor error")
	}

	return absentResults, nil

}

func CreateAbsent(absentData model.Absent) (bool, error) {
	absent_collection, err := configuration.GetCollection("absent")
	if err != nil {
		return false, fmt.Errorf("failed to get the absetn collection")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	insertedID, err := absent_collection.InsertOne(ctx, absentData)

	if err != nil || insertedID.InsertedID.(primitive.ObjectID).IsZero() {
		return false, fmt.Errorf("failed to create attendacne: %v", err)
	}

	return true, nil
}

func CheckTodayAbsentRequest(userId primitive.ObjectID) (bool, error) {
	if userId.IsZero() {
		return false, fmt.Errorf("fialed to get the userid")
	}
	absent_collection, err := configuration.GetCollection("absent")
	if err != nil {
		return false, fmt.Errorf("failed to get the absent collection")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	today := time.Now().Truncate(24 * time.Hour) // Get today's date at midnight

	filter := bson.M{
		"userId": userId,
		"date": bson.M{
			"$gte": today,
			"$lt":  today.Add(24 * time.Hour),
		},
	}

	count, err := absent_collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
