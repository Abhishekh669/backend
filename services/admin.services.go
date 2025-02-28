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

func RejectAbsentRequest(id primitive.ObjectID) (bool, error) {
	absent_collection, err := configuration.GetCollection("absent")
	if err != nil {
		return false, fmt.Errorf("failed to get the absetn collection")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"status": "rejected"}}

	result, err := absent_collection.UpdateOne(ctx, filter, update)

	if err != nil {
		return false, fmt.Errorf("failed to update absent request: %v", err)
	}

	if result.MatchedCount == 0 {
		return false, fmt.Errorf("no document found with the given ID")
	}

	return true, nil

}

func AcceptAbsentRequest(id primitive.ObjectID) (bool, error) {
	absent_collection, err := configuration.GetCollection("absent")
	if err != nil {
		return false, fmt.Errorf("failed to get the absetn collection")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"status": "accepted"}}

	result, err := absent_collection.UpdateOne(ctx, filter, update)

	if err != nil {
		return false, fmt.Errorf("failed to update absent request: %v", err)
	}

	if result.MatchedCount == 0 {
		return false, fmt.Errorf("no document found with the given ID")
	}

	return true, nil

}

func GetUserAbsentRequestById(id primitive.ObjectID) ([]model.Absent, error) {
	absent_collection, err := configuration.GetCollection("absent")
	if err != nil {
		return []model.Absent{}, fmt.Errorf("failed to get the absetn collection")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"userId": id}

	var absentRequests []model.Absent
	cursor, err := absent_collection.Find(ctx, filter)
	if err != nil {
		return []model.Absent{}, fmt.Errorf("faild to get absent results")
	}

	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var absentRequest model.Absent
		if err := cursor.Decode(&absentRequest); err != nil {
			return []model.Absent{}, fmt.Errorf("failed to get absent")
		}
		absentRequests = append(absentRequests, absentRequest)
	}

	if err := cursor.Err(); err != nil {
		return []model.Absent{}, fmt.Errorf("cursor error : %v", err)
	}

	return absentRequests, nil

}

func GetAllAbsentRequest() ([]model.Absent, error) {
	absent_collection, err := configuration.GetCollection("absent")
	if err != nil {
		return []model.Absent{}, fmt.Errorf("failed to get the absetn collection")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var absentResults []model.Absent

	options := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})

	cursor, err := absent_collection.Find(ctx, bson.M{}, options)
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

	fmt.Println("this is hte all absetn data : ", absentResults)

	return absentResults, nil

}

// func AcceptUser(userId string) (bool, error) {
// 	user_collection, err := configuration.GetCollection("users")
// 	if err != nil {
// 		return []model.User{}, fmt.Errorf("could not get the user collection: %v", err)
// 	}

// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()

// 	update := bson.M{
// 		"isVerified": true,
// 	}

// }
func DeleteUser(userId primitive.ObjectID) (bool, error) {
	user_collection, err := configuration.GetCollection("users")
	if err != nil {
		return false, fmt.Errorf("failed to get the user collection: %v", err)
	}

	timerWorkspace_collection, err := configuration.GetCollection("timerWorkspace")
	if err != nil {
		return false, fmt.Errorf("failed to get the timer workspace collection: %v", err)
	}

	timer_collection, err := configuration.GetCollection("timer")
	if err != nil {
		return false, fmt.Errorf("failed to get the timer collection: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Filter to delete user by userId
	userFilter := bson.M{
		"_id": userId,
	}

	// Delete user
	result, err := user_collection.DeleteOne(ctx, userFilter)
	if err != nil {
		return false, fmt.Errorf("failed to delete user: %v", err)
	}

	if result.DeletedCount == 0 {
		return false, fmt.Errorf("no user found with the given ID")
	}

	// Delete related documents in timerWorkspace collection where userId matches
	workspaceFilter := bson.M{
		"userId": userId, // assuming the field is named "userId"
	}
	_, err = timerWorkspace_collection.DeleteMany(ctx, workspaceFilter)
	if err != nil {
		return false, fmt.Errorf("failed to delete related timer workspace documents: %v", err)
	}

	// Delete related documents in timer collection where userId matches
	timerFilter := bson.M{
		"userId": userId, // assuming the field is named "userId"
	}
	_, err = timer_collection.DeleteMany(ctx, timerFilter)
	if err != nil {
		return false, fmt.Errorf("failed to delete related timer documents: %v", err)
	}

	return true, nil
}

func SetAdmin(userId primitive.ObjectID) (bool, error) {
	user_collection, err := configuration.GetCollection("users")
	if err != nil {
		return false, fmt.Errorf("failed to get the user collection: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"isAdmin": true, // Set isVerified to true
		},
	}

	// Define an options struct to return the updated document
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var updatedUser model.User

	err = user_collection.FindOneAndUpdate(ctx, bson.M{"_id": userId}, update, opts).Decode(&updatedUser)
	if err != nil {
		return false, fmt.Errorf("could not update user: %v", err)
	}
	if updatedUser.IsVerified && updatedUser.IsOnBoarded {
		return true, nil
	} else {
		return false, fmt.Errorf("failed to verify user")
	}

}

func RemoveAdmin(userId primitive.ObjectID) (bool, error) {
	user_collection, err := configuration.GetCollection("users")
	if err != nil {
		return false, fmt.Errorf("failed to get the user collection: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"isAdmin": false, // Set isVerified to true
		},
	}

	// Define an options struct to return the updated document
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var updatedUser model.User

	err = user_collection.FindOneAndUpdate(ctx, bson.M{"_id": userId}, update, opts).Decode(&updatedUser)
	if err != nil {
		return false, fmt.Errorf("could not update user: %v", err)
	}
	if updatedUser.IsVerified && updatedUser.IsOnBoarded {
		return true, nil
	} else {
		return false, fmt.Errorf("failed to verify user")
	}

}

func RejectUser(userId primitive.ObjectID) (bool, error) {

	user_collection, err := configuration.GetCollection("users")
	if err != nil {
		return false, fmt.Errorf("failed to get the user collection: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"_id": userId,
	}

	result, err := user_collection.DeleteOne(ctx, filter)

	if err != nil {
		return false, fmt.Errorf("failed to delete user: %v", err)
	}

	if result.DeletedCount == 0 {
		return false, fmt.Errorf("no user found with the given ID")
	}

	return true, nil

}

func GetUnverifiedUser() ([]model.User, error) {
	user_collection, err := configuration.GetCollection("users")
	if err != nil {
		return []model.User{}, fmt.Errorf("could not get the user collection: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"isOnBoarded": true,
		"isVerified":  false,
		"isAdmin":     false,
	}

	var users []model.User

	cursor, err := user_collection.Find(ctx, filter)

	if err != nil {
		return nil, fmt.Errorf("error fetching unverified users: %v", err)
	}

	for cursor.Next(ctx) {
		var user model.User
		if err := cursor.Decode(&user); err != nil {
			return []model.User{}, fmt.Errorf("failed to get users: %v", err)
		}
		users = append(users, user)

	}

	if err := cursor.Err(); err != nil {
		return []model.User{}, fmt.Errorf("cursor error : %v", err)
	}

	return users, nil

}

func VerifyUser(userId primitive.ObjectID) (bool, error) {
	user_collection, err := configuration.GetCollection("users")
	if err != nil {
		return false, fmt.Errorf("could not get the user collection: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	update := bson.M{
		"$set": bson.M{
			"isVerified": true, // Set isVerified to true
		},
	}

	// Define an options struct to return the updated document
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var updatedUser model.User

	err = user_collection.FindOneAndUpdate(ctx, bson.M{"_id": userId}, update, opts).Decode(&updatedUser)
	if err != nil {
		return false, fmt.Errorf("could not update user: %v", err)
	}
	if updatedUser.IsVerified && updatedUser.IsOnBoarded {
		return true, nil
	} else {
		return false, fmt.Errorf("failed to verify user")
	}

}

func CreateAdmin(admin model.Admin) (model.Admin, error) {
	admin_collection, err := configuration.GetCollection("admin")

	if err != nil {
		return model.Admin{}, fmt.Errorf("could not get the admin collection: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if admin.CreatedAt.IsZero() {
		admin.CreatedAt = time.Now()
	}

	result, err := admin_collection.InsertOne(ctx, admin)

	if err != nil {
		return model.Admin{}, fmt.Errorf("failed to create admin: %v", err)
	}

	admin.ID = result.InsertedID.(primitive.ObjectID)
	fmt.Println("This is user after creting admin : ", admin)
	return admin, nil

}

func GetAdminById(id primitive.ObjectID) (model.Admin, error) {

	filter := bson.D{{Key: "_id", Value: id}}

	var admin model.Admin

	admin_collection, err := configuration.GetCollection("admin")

	if err != nil {
		return model.Admin{}, fmt.Errorf("could not get the admin collection: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = admin_collection.FindOne(ctx, filter).Decode(&admin)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return model.Admin{}, fmt.Errorf("no user found : %v", err)
		} else {
			return model.Admin{}, fmt.Errorf("failed to find the user")
		}
	}
	return admin, nil

}

func CheckAdmin(userId primitive.ObjectID, email string) (model.Admin, error) {

	checkAdminLength, err := GetAllAdmin()

	if err != nil {
		return model.Admin{}, fmt.Errorf("something went wrong")
	}

	if checkAdminLength >= 2 {
		return model.Admin{}, fmt.Errorf("max admin user reached")
	}

	if userId.IsZero() || email == "" {
		return model.Admin{}, fmt.Errorf("userid or email not got ")
	}

	var admin model.Admin

	admin_collection, err := configuration.GetCollection("admin")

	if err != nil {
		return model.Admin{}, fmt.Errorf("could not get the admin collection: %v", err)
	}

	filter := bson.M{"userId": userId, "email": email}
	err = admin_collection.FindOne(context.Background(), filter).Decode(&admin)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return model.Admin{}, fmt.Errorf("admin not found with the provided userId and email")
		}
		return model.Admin{}, fmt.Errorf("could not find admin: %v", err)
	}

	fmt.Println("This is the admin : ", admin)

	// Return the found admin
	return admin, nil

}

func GetAllAdmin() (int, error) {
	admin_collection, err := configuration.GetCollection("admin")
	if err != nil {
		return 0, fmt.Errorf("could not get the admin collection: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var result []model.Admin
	cursor, err := admin_collection.Find(ctx, bson.M{})
	if err != nil {
		return 0, fmt.Errorf("faild to get admin")
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var admin model.Admin
		if err := cursor.Decode(&admin); err != nil {
			return 0, fmt.Errorf("failed to decode admin: %v", err)
		}
		result = append(result, admin)

	}

	if err := cursor.Err(); err != nil {
		return 0, fmt.Errorf("cursor error : %v", err)
	}

	return len(result), nil

}
