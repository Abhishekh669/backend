package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Abhishekh669/backend/configuration"
	"github.com/Abhishekh669/backend/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DeleteWorkspaceTimerType struct {
	UserId      primitive.ObjectID `json:"userId,omitempty" bson:"userId,omitempty"`
	TimerId     primitive.ObjectID `json:"timerId,omitempty" bson:"timerId,omitempty"`
	WorkspaceId primitive.ObjectID `json:"workspaceId,omitempty" bson:"workspaceId,omitempty"`
}

type DeleteWorkspaceTimerRequestType struct {
	TimerId     string `json:"timerId,omitempty" bson:"timerId,omitempty"`
	WorkspaceId string `json:"workspaceId,omitempty" bson:"workspaceId,omitempty"`
}

func DeleteWorkspace(workspaceId primitive.ObjectID) (bool, error) {
	// Get the timerWorkspace collection
	timerWorkspaceCollection, err := configuration.GetCollection("timerWorkspace")
	if err != nil {
		return false, fmt.Errorf("failed to get the timer workspace collection: %v", err)
	}

	// Get the timer collection
	timerCollection, err := configuration.GetCollection("timer")
	if err != nil {
		return false, fmt.Errorf("failed to get the timer collection: %v", err)
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Step 1: Delete the workspace from the timerWorkspace collection
	workspaceFilter := bson.M{"_id": workspaceId}
	workspaceDeleteResult, err := timerWorkspaceCollection.DeleteOne(ctx, workspaceFilter)
	if err != nil {
		return false, fmt.Errorf("failed to delete workspace: %v", err)
	}

	// Check if the workspace was found and deleted
	if workspaceDeleteResult.DeletedCount == 0 {
		return false, fmt.Errorf("no workspace found with the given ID: %v", workspaceId)
	}

	// Step 2: Delete all timers associated with the workspace
	timerFilter := bson.M{"timerWorkspaceId": workspaceId}
	timerDeleteResult, err := timerCollection.DeleteMany(ctx, timerFilter)
	if err != nil {
		return false, fmt.Errorf("failed to delete timers: %v", err)
	}

	// Log the number of timers deleted (optional)
	log.Printf("Deleted %v timers associated with workspace ID: %v", timerDeleteResult.DeletedCount, workspaceId)

	return true, nil
}

func DeleteWorkspaceTimer(data DeleteWorkspaceTimerType) (bool, error) {
	timerCollection, err := configuration.GetCollection("timer")
	if err != nil {
		return false, fmt.Errorf("failed to get the timer collection: %v", err)
	}

	// Get the timerWorkspace collection
	timerWorkspaceCollection, err := configuration.GetCollection("timerWorkspace")
	if err != nil {
		return false, fmt.Errorf("failed to get the timer workspace collection: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Step 1: Delete the timer from the timer collection
	timerFilter := bson.M{"_id": data.TimerId}
	timerDeleteResult, err := timerCollection.DeleteOne(ctx, timerFilter)
	if err != nil {
		return false, fmt.Errorf("failed to delete timer: %v", err)
	}
	if timerDeleteResult.DeletedCount == 0 {
		return false, fmt.Errorf("no timer found with the given ID: %v", data.TimerId)
	}

	// Step 2: Remove the timer ID from the timerIds array in the timerWorkspace collection
	timerWorkspaceFilter := bson.M{"_id": data.WorkspaceId}
	timerWorkspaceUpdate := bson.M{
		"$pull": bson.M{
			"timerIds": data.TimerId,
		},
	}

	// Perform the update operation
	timerWorkspaceUpdateResult, err := timerWorkspaceCollection.UpdateOne(ctx, timerWorkspaceFilter, timerWorkspaceUpdate)
	if err != nil {
		return false, fmt.Errorf("failed to update timer workspace: %v", err)
	}

	// Check if the timerWorkspace document was found and updated
	if timerWorkspaceUpdateResult.MatchedCount == 0 {
		return false, fmt.Errorf("no timer workspace found with the given ID: %v", data.WorkspaceId)
	}

	return true, nil

}

func EditTimerWorkspaceData(timerData model.TimerWorkspace) (bool, error) {
	if timerData.ID.IsZero() {
		return false, fmt.Errorf("no user id provided")
	}

	timerWorkspace_collection, err := configuration.GetCollection("timerWorkspace")
	if err != nil {
		return false, fmt.Errorf("could not get the timer workspace  collection: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"name":        timerData.Name,
			"description": timerData.Description,
		},
	}

	fmt.Println("this ish te namae in edit workspace : ", timerData.Name)
	fmt.Println("this ish te descritpion  workspace: ", timerData.Description)

	filter := bson.M{"_id": timerData.ID}

	result, err := timerWorkspace_collection.UpdateOne(ctx, filter, update)

	if err != nil {
		return false, fmt.Errorf("failed to update timer workspace")
	}

	if result.MatchedCount == 0 {
		return false, fmt.Errorf("timer workspace not found")
	}
	fmt.Println("i am fine hoi tw saboi updated")
	return true, nil

}

func EditTimerData(timerData model.Timer) (bool, error) {
	if timerData.ID.IsZero() {
		return false, fmt.Errorf("no user id provided")
	}

	timer_collection, err := configuration.GetCollection("timer")
	if err != nil {
		return false, fmt.Errorf("could not get the user collection: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"name":    timerData.Name,
			"endDate": timerData.EndDate,
			"type":    timerData.Type,
		},
	}

	filter := bson.M{"_id": timerData.ID}

	result, err := timer_collection.UpdateOne(ctx, filter, update)

	if err != nil {
		return false, fmt.Errorf("failed to update timer")
	}

	if result.MatchedCount == 0 {
		return false, fmt.Errorf("timer not found")
	}
	return true, nil

}

func GetTimeWorkspaceById(id primitive.ObjectID, userId primitive.ObjectID) (model.TimerWorkspace, error) {
	filter := bson.D{{Key: "_id", Value: id}, {Key: "userId", Value: userId}}

	var timerWorkspace model.TimerWorkspace

	timerWorkspace_collection, err := configuration.GetCollection("timerWorkspace")

	if err != nil {
		return model.TimerWorkspace{}, fmt.Errorf("could not get the timer workspace collection: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = timerWorkspace_collection.FindOne(ctx, filter).Decode(&timerWorkspace)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return model.TimerWorkspace{}, fmt.Errorf("no timer found : %v", err)
		} else {
			return model.TimerWorkspace{}, fmt.Errorf("failed to find the timer")
		}
	}
	return timerWorkspace, nil

}

func GetUserTimerWorkspacesOkie(userId primitive.ObjectID) ([]model.TimerWorkspace, error) {
	if userId.IsZero() {
		return []model.TimerWorkspace{}, fmt.Errorf("user id  not  proviced")
	}

	timerWorkspace_collection, err := configuration.GetCollection("timerWorkspace")

	if err != nil {
		return []model.TimerWorkspace{}, fmt.Errorf("failed to get timer workspace collection")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	var timerWorkspaces []model.TimerWorkspace
	filter := bson.M{
		"userId": userId,
	}

	cursor, err := timerWorkspace_collection.Find(ctx, filter)
	if err != nil {
		return []model.TimerWorkspace{}, fmt.Errorf("faild to get workspace")
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var timerWorkspace model.TimerWorkspace
		if err := cursor.Decode(&timerWorkspace); err != nil {
			return []model.TimerWorkspace{}, fmt.Errorf("faield to decode workspace")
		}
		timerWorkspaces = append(timerWorkspaces, timerWorkspace)
	}

	if err := cursor.Err(); err != nil {
		return []model.TimerWorkspace{}, fmt.Errorf("cursor error")
	}

	return timerWorkspaces, nil

}

func CreateTimerWorkspace(timerWorkspace model.TimerWorkspace) (model.TimerWorkspace, error) {
	timerWorkspace_collection, err := configuration.GetCollection("timerWorkspace")

	if err != nil {
		return model.TimerWorkspace{}, fmt.Errorf("failed to get timer  workspace collection")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	result, err := timerWorkspace_collection.InsertOne(ctx, timerWorkspace)

	if err != nil {
		return model.TimerWorkspace{}, fmt.Errorf("failed to createa a timer")
	}

	timerWorkspace.ID = result.InsertedID.(primitive.ObjectID)
	return timerWorkspace, nil

}

func GetUserTimerWorkspaces(userId primitive.ObjectID) ([]model.TimerWorkspace, error) {
	if userId.IsZero() {
		return []model.TimerWorkspace{}, fmt.Errorf("user id  not  proviced")
	}

	timerWorkspace_collection, err := configuration.GetCollection("timerWorkspace")

	if err != nil {
		return []model.TimerWorkspace{}, fmt.Errorf("failed to get timer workspace collection")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	var timerWorkspaces []model.TimerWorkspace
	filter := bson.M{
		"userId": userId,
	}
	options := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})

	cursor, err := timerWorkspace_collection.Find(ctx, filter, options)
	if err != nil {
		return []model.TimerWorkspace{}, fmt.Errorf("faild to get workspace")
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var timerWorkspace model.TimerWorkspace
		if err := cursor.Decode(&timerWorkspace); err != nil {
			return []model.TimerWorkspace{}, fmt.Errorf("faield to decode workspace")
		}
		timerWorkspaces = append(timerWorkspaces, timerWorkspace)
	}

	if err := cursor.Err(); err != nil {
		return []model.TimerWorkspace{}, fmt.Errorf("cursor error")
	}

	return timerWorkspaces, nil

}

func CreateTimer(timer model.Timer) (model.Timer, error) {
	timer_collection, err := configuration.GetCollection("timer")
	if err != nil {
		return model.Timer{}, fmt.Errorf("failed to get timer collection")
	}

	timerWorkspace_collection, err := configuration.GetCollection("timerWorkspace")
	if err != nil {
		return model.Timer{}, fmt.Errorf("failed to get workspace timer collection")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := timer_collection.InsertOne(ctx, timer)
	if err != nil {
		return model.Timer{}, fmt.Errorf("failed to create a timer")
	}

	timer.ID = result.InsertedID.(primitive.ObjectID)

	filter := bson.D{{Key: "_id", Value: timer.TimerWorkspaceId}} // Assuming the workspace ID is part of the timer model
	update := bson.D{
		{Key: "$push", Value: bson.D{
			{Key: "timerIds", Value: timer.ID},
		}},
	}

	_, err = timerWorkspace_collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return model.Timer{}, fmt.Errorf("failed to update TimerWorkspace with the new timer ID")
	}

	return timer, nil
}

func GetUserWorkspaceTimers(userId primitive.ObjectID, workspaceId primitive.ObjectID) ([]model.Timer, error) {
	if userId.IsZero() || workspaceId.IsZero() {
		return []model.Timer{}, fmt.Errorf("no user id proviced")
	}

	timer_collection, err := configuration.GetCollection("timer")

	if err != nil {
		return []model.Timer{}, fmt.Errorf("failed to get timer collection")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	var timers []model.Timer
	filter := bson.M{
		"userId":           userId,
		"timerWorkspaceId": workspaceId,
	}

	options := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})

	cursor, err := timer_collection.Find(ctx, filter, options)
	if err != nil {
		return []model.Timer{}, fmt.Errorf("faild to get users")
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var timer model.Timer
		if err := cursor.Decode(&timer); err != nil {
			return []model.Timer{}, fmt.Errorf("faield to decode timer")
		}
		timers = append(timers, timer)
	}

	if err := cursor.Err(); err != nil {
		return []model.Timer{}, fmt.Errorf("cursor error")
	}

	return timers, nil

}

func GetTimerById(id primitive.ObjectID, userId primitive.ObjectID, workspaceId primitive.ObjectID) (model.Timer, error) {

	filter := bson.D{{Key: "_id", Value: id}, {Key: "userId", Value: userId}, {Key: "timerWorkspaceId", Value: workspaceId}}

	var timer model.Timer

	timer_collection, err := configuration.GetCollection("timer")

	if err != nil {
		return model.Timer{}, fmt.Errorf("could not get the timer collection: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = timer_collection.FindOne(ctx, filter).Decode(&timer)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return model.Timer{}, fmt.Errorf("no timer found : %v", err)
		} else {
			return model.Timer{}, fmt.Errorf("failed to find the timer")
		}
	}
	return timer, nil

}
