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

func reverseMessages(messages []map[string]interface{}) {
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
}
func populateMessageDetails(message model.Message) (bson.M, error) {
	// Retrieve the sender and receiver IDs

	// Fetch sender and receiver details (assuming you have a function to get user details)
	sender, err := GetUserById(message.Sender)
	if err != nil {
		return nil, fmt.Errorf("failed to get sender details: %v", err)
	}

	receiver, err := GetUserById(message.Receiver)
	if err != nil {
		return nil, fmt.Errorf("failed to get receiver details: %v", err)
	}

	// Handle the seenBy field (if available)

	var seenByUsers []map[string]interface{}
	for _, userID := range message.SeenBy {
		user, err := GetUserById(userID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch seen_by user: %v", err)
		}
		data := map[string]interface{}{
			"_id":      user.ID,
			"name":     user.Name,
			"email":    user.Email,
			"address":  user.Address,
			"field":    user.Field,
			"image":    user.Image,
			"codeName": user.CodeName,
		}
		seenByUsers = append(seenByUsers, data)
	}

	// Populated message structure
	populatedMessage := bson.M{
		"_id":            message.ID,
		"conversationId": message.ConversationId,
		"body":           message.Body,
		"sender": map[string]interface{}{
			"_id":      sender.ID,
			"name":     sender.Name,
			"email":    sender.Email,
			"address":  sender.Address,
			"field":    sender.Field,
			"image":    sender.Image,
			"codeName": sender.CodeName,
		},
		"receiver": map[string]interface{}{
			"_id":      receiver.ID,
			"name":     receiver.Name,
			"email":    receiver.Email,
			"address":  receiver.Address,
			"field":    receiver.Field,
			"image":    receiver.Image,
			"codeName": receiver.CodeName,
		},
		"seenBy":    seenByUsers,
		"createdAt": message.CreatedAt, // Here, it's an array of strings instead of ObjectIDs
	}

	return populatedMessage, nil
}

func populatedMessagesDetails(message bson.M) (bson.M, error) {
	senderID, ok := message["sender"].(primitive.ObjectID)
	if !ok {
		return nil, fmt.Errorf("invalid sender ID")
	}

	receiverID, ok := message["receiver"].(primitive.ObjectID)
	if !ok {
		return nil, fmt.Errorf(("invlaid reciever ID"))
	}

	sender, err := GetUserById(senderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sender: %v", err)
	}

	receiver, err := GetUserById(receiverID)
	if err != nil {
		return nil, fmt.Errorf("failed to get receiver: %v", err)
	}

	seenByIDs, ok := message["seenBy"].(bson.A)
	if !ok {
		return nil, fmt.Errorf("invalid seenBy field")
	}

	var seenByUsers []map[string]interface{}
	for _, userID := range seenByIDs {
		user, err := GetUserById(userID.(primitive.ObjectID))
		if err != nil {
			return nil, fmt.Errorf("failed to fetch seen_by user: %v", err)
		}
		data := map[string]interface{}{
			"_id":      user.ID,
			"name":     user.Name,
			"email":    user.Email,
			"address":  user.Address,
			"field":    user.Field,
			"image":    user.Image,
			"codeName": user.CodeName,
		}
		seenByUsers = append(seenByUsers, data)
	}

	populatedMessage := bson.M{
		"_id":            message["_id"],
		"conversationId": message["conversationId"],
		"body":           message["body"],
		"sender": map[string]interface{}{
			"_id":      sender.ID,
			"name":     sender.Name,
			"email":    sender.Email,
			"address":  sender.Address,
			"field":    sender.Field,
			"image":    sender.Image,
			"codeName": sender.CodeName,
		},
		"receiver": map[string]interface{}{
			"_id":      receiver.ID,
			"name":     receiver.Name,
			"email":    receiver.Email,
			"address":  receiver.Address,
			"field":    receiver.Field,
			"image":    receiver.Image,
			"codeName": receiver.CodeName,
		},

		"seenBy": seenByUsers,
	}
	return populatedMessage, nil
}

func GetAllConversationMessages(id primitive.ObjectID, lastID primitive.ObjectID, limit int) (map[string]interface{}, error) {

	message_collection, err := configuration.GetCollection("messages")
	if err != nil {
		return map[string]interface{}{}, fmt.Errorf("could not get the message collection: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var messages []map[string]interface{}

	filter := bson.M{
		"conversationId": id,
	}

	count, err := message_collection.CountDocuments(ctx, filter)
	if err != nil {
		return map[string]interface{}{}, fmt.Errorf("failed to count messages : %v", err)
	}

	if !lastID.IsZero() {
		filter["_id"] = bson.M{"$lt": lastID} // Fetch messages with _id less than lastID
	}

	options := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}).SetLimit(int64(limit))
	cursor, err := message_collection.Find(ctx, filter, options)
	if err != nil {
		return map[string]interface{}{}, fmt.Errorf("failed to get messages: %v", err)
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var message bson.M
		if err := cursor.Decode(&message); err != nil {
			return map[string]interface{}{}, fmt.Errorf("failed to decode message: %v", err)
		}

		populatedMessage, err := populatedMessagesDetails(message)
		if err != nil {
			return map[string]interface{}{}, fmt.Errorf("failed to populate message details: %v", err)
		}

		messages = append(messages, populatedMessage)
	}

	if err := cursor.Err(); err != nil {
		return map[string]interface{}{}, fmt.Errorf("cursor error: %v", err)
	}

	reverseMessages(messages)

	response := map[string]interface{}{
		"messages":      messages,
		"totalMessages": count,
	}

	return response, nil
}

func CreateMesssage(message model.Message) (map[string]interface{}, error) {
	message_collection, err := configuration.GetCollection("messages")

	if err != nil {
		return map[string]interface{}{}, fmt.Errorf("could not get the message collection: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	result, err := message_collection.InsertOne(ctx, message)

	if err != nil {
		return map[string]interface{}{}, fmt.Errorf("failed to create message: %v", err)
	}

	message.ID = result.InsertedID.(primitive.ObjectID)

	populatedMessage, err := populateMessageDetails(message)
	if err != nil {
		return map[string]interface{}{}, fmt.Errorf("failed to populate message details: %v", err)
	}
	return populatedMessage, nil
}

func GetMessageById(id primitive.ObjectID) (model.Message, error) {

	var message model.Message
	filter := bson.D{{Key: "_id", Value: id}}

	message_collection, err := configuration.GetCollection("messages")

	if err != nil {
		return model.Message{}, fmt.Errorf("could not get the message collection: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = message_collection.FindOne(ctx, filter).Decode(&message)
	fmt.Println("THis ishte reuslt after search the message ", message)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return model.Message{}, fmt.Errorf("no message found : %v", err)
		} else {
			return model.Message{}, fmt.Errorf("failed to find the message")
		}
	}
	fmt.Println("THis is check user ", message)
	return message, nil
}
