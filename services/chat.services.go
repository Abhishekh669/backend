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
)

func CreateChat(chat model.Chat) (model.Chat, error) {
	chat_collection, err := configuration.GetCollection("chats")

	if err != nil {
		return model.Chat{}, fmt.Errorf("could not get the chat collection: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	if chat.CreatedAt.IsZero() {
		chat.CreatedAt = time.Now()
	}

	result, err := chat_collection.InsertOne(ctx, chat)

	if err != nil {
		return model.Chat{}, fmt.Errorf("failed to create chat: %v", err)
	}

	chat.ID = result.InsertedID.(primitive.ObjectID)
	fmt.Println("This is user after creting user : ", chat)
	return chat, nil

}

func GetUserInConversation(ObjectConversationId primitive.ObjectID, ObjectUserId primitive.ObjectID) (bool, error) {

	filter := bson.M{
		"_id":     ObjectConversationId,
		"members": bson.M{"$in": []primitive.ObjectID{ObjectUserId}},
	}

	conversation_collection, err := configuration.GetCollection("chats")
	if err != nil {
		return false, fmt.Errorf("failed to get conversation: %v", err)
	}

	var chat bson.M
	err = conversation_collection.FindOne(context.Background(), filter).Decode(&chat)

	fmt.Println("this is conversation check")
	if err == mongo.ErrNoDocuments {
		// No conversation found
		return false, fmt.Errorf("failed to get conversation")
	} else if err != nil {
		// An error occurred during the query
		log.Printf("Error fetching conversation: %v", err)
		return false, fmt.Errorf("failed to get conversation: %v", err)
	}

	return true, nil

}

func GetConversationById(chatId primitive.ObjectID) (model.Chat, error) {

	if chatId.IsZero() {
		return model.Chat{}, fmt.Errorf("id is empty")

	}
	chat_collection, err := configuration.GetCollection("chats")
	if err != nil {
		return model.Chat{}, fmt.Errorf("could not get the chat collection: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	var Chat model.Chat

	filter := bson.D{{Key: "_id", Value: chatId}}

	err = chat_collection.FindOne(ctx, filter).Decode(&Chat)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return model.Chat{}, fmt.Errorf("no user found : %v", err)
		} else {
			return model.Chat{}, fmt.Errorf("failed to find the user")
		}
	}
	return Chat, nil

}
