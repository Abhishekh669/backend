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
)

func CreateUser(user model.User) (model.User, error) {
	user_collection, err := configuration.GetCollection("users")

	if err != nil {
		return model.User{}, fmt.Errorf("could not get the user collection: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}
	fmt.Println("This is the user for creation :", user)

	result, err := user_collection.InsertOne(ctx, user)

	if err != nil {
		return model.User{}, fmt.Errorf("failed to create user: %v", err)
	}

	user.ID = result.InsertedID.(primitive.ObjectID)
	fmt.Println("This is user after creting user : ", user)
	return user, nil

}

func GetUserById(id primitive.ObjectID) (model.User, error) {

	filter := bson.D{{Key: "_id", Value: id}}

	var user model.User

	user_collection, err := configuration.GetCollection("users")

	if err != nil {
		return model.User{}, fmt.Errorf("could not get the user collection: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = user_collection.FindOne(ctx, filter).Decode(&user)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return model.User{}, fmt.Errorf("no user found : %v", err)
		} else {
			return model.User{}, fmt.Errorf("failed to find the user")
		}
	}
	return user, nil

}

func CheckUser(email string) (model.User, error) {
	if email == "" {
		return model.User{}, fmt.Errorf("email or google id missing")
	}

	user_collection, err := configuration.GetCollection("users")

	if err != nil {
		return model.User{}, fmt.Errorf("could not get the user collection: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fmt.Println("this is the context : ", ctx)

	var result model.User
	err = user_collection.FindOne(ctx, bson.M{"email": email}).Decode(&result)
	fmt.Println("THis ishte reuslt after search the user ", result)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return model.User{}, fmt.Errorf("no user found : %v", err)
		} else {
			return model.User{}, fmt.Errorf("failed to find the user")
		}
	}
	fmt.Println("THis is check user ", result)
	return result, nil

}

func GetAllUsers() ([]model.User, error) {
	user_collection, err := configuration.GetCollection("users")
	if err != nil {
		return []model.User{}, fmt.Errorf("could not get the user collection: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var result []model.User
	cursor, err := user_collection.Find(ctx, bson.M{})
	if err != nil {
		return []model.User{}, fmt.Errorf("faild to get users")
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var user model.User
		if err := cursor.Decode(&user); err != nil {
			return []model.User{}, fmt.Errorf("failed to decode user: %v", err)
		}
		result = append(result, user)

	}

	if err := cursor.Err(); err != nil {
		return []model.User{}, fmt.Errorf("cursor error : %v", err)
	}

	return result, nil

}

func OnboardingUser(onboarding_data model.OnboardingRequest, userId primitive.ObjectID) (model.User, error) {
	if userId.IsZero() {
		return model.User{}, fmt.Errorf("no user id provided")
	}

	user_collection, err := configuration.GetCollection("users")
	if err != nil {
		return model.User{}, fmt.Errorf("could not get the user collection: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"codeName":      onboarding_data.CodeName,
			"phoneNumber":   onboarding_data.PhoneNumber,
			"address":       onboarding_data.Address,
			"age":           onboarding_data.Age,
			"qualification": onboarding_data.Qualification,
			"field":         onboarding_data.Field,
			"isOnBoarded":   true,
			"mainField":     onboarding_data.MainField,
		},
	}

	filter := bson.M{"_id": userId}

	result, err := user_collection.UpdateOne(ctx, filter, update)

	if err != nil {
		return model.User{}, fmt.Errorf("failed to update user")
	}

	if result.MatchedCount == 0 {
		return model.User{}, fmt.Errorf("user not found")
	}

	var updatedUser model.User
	err = user_collection.FindOne(ctx, filter).Decode(&updatedUser)

	if err != nil {
		return model.User{}, fmt.Errorf("failed to fetch updated user: %v", err)
	}

	return updatedUser, nil

}
