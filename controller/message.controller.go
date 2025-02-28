package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Abhishekh669/backend/configuration"
	"github.com/Abhishekh669/backend/model"
	"github.com/Abhishekh669/backend/services"
	"github.com/Abhishekh669/backend/sockets"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	defaultLimit = 20 // Default number of messages to fetch
	maxLimit     = 20 // Maximum allowed limit
)

func GetAllConversationMessagesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Allow-Control-Allow-Methods", "GET")

	sessionData, err := configuration.GetSessionData(r)

	if err != nil {
		http.Error(w, "invalid session", http.StatusBadRequest)
		return

	}

	ObjectUserId, err := primitive.ObjectIDFromHex(sessionData.UserId)

	if err != nil {
		http.Error(w, "Failed to parse userid ", http.StatusInternalServerError)
		return
	}

	currentUser, err := services.GetUserById(ObjectUserId)

	if err != nil || currentUser.ID.IsZero() {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	params := mux.Vars(r)
	conversationId := params["id"]

	ObjectConversationId, err := primitive.ObjectIDFromHex(conversationId)

	if err != nil {
		http.Error(w, "Failed to parse conversation id ", http.StatusInternalServerError)
		return
	}

	conversation, err := services.GetConversationById(ObjectConversationId)

	if err != nil {
		http.Error(w, "failed to get  conversation  ", http.StatusForbidden)
		return
	}

	if conversation.ID.IsZero() || len(conversation.Members) == 0 {
		http.Error(w, "No conversation  found  ", http.StatusForbidden)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	lastIDStr := r.URL.Query().Get("lastID")

	query := struct {
		Limit  int
		LastID primitive.ObjectID
	}{
		Limit:  defaultLimit,          // Default limit
		LastID: primitive.NilObjectID, // Default lastID
	}

	// Parse and validate limit
	if limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 0 {
			http.Error(w, "Invalid limit", http.StatusBadRequest)
			return
		}
		// Ensure limit does not exceed the maximum allowed value
		if limit > maxLimit {
			limit = maxLimit
		}
		query.Limit = limit
	}

	// Parse and validate lastID
	if lastIDStr != "" {
		lastID, err := primitive.ObjectIDFromHex(lastIDStr)
		if err != nil {
			http.Error(w, "Invalid lastID", http.StatusBadRequest)
			return
		}
		query.LastID = lastID
	}

	messages, err := services.GetAllConversationMessages(ObjectConversationId, query.LastID, query.Limit)

	if err != nil {
		http.Error(w, "Error : failed to get messages", http.StatusInternalServerError)
	}

	json.NewEncoder(w).Encode(messages)

}

func CreateMessageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Allow-Control-Allow-Methods", "POST")

	var MessageReq model.MessageRequest
	fmt.Println("This is the new r.body : ", r.Body)

	if err := json.NewDecoder(r.Body).Decode(&MessageReq); err != nil {
		fmt.Println("this is hte meian error : ", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	sessionData, err := configuration.GetSessionData(r)

	if err != nil {
		http.Error(w, "invalid session", http.StatusBadRequest)
		return

	}

	if sessionData.UserId != MessageReq.Sender {
		http.Error(w, "User not authenticated", http.StatusBadRequest)
		return
	}

	ObjectConversationId, err := primitive.ObjectIDFromHex(MessageReq.ConversationId)
	if err != nil {
		http.Error(w, "failed to parse conversation id ", http.StatusForbidden)
		return
	}
	ObjectUserId, err := primitive.ObjectIDFromHex(sessionData.UserId)
	if err != nil {
		http.Error(w, "failed to parse user id ", http.StatusForbidden)
		return
	}

	ObjectReceiverID, err := primitive.ObjectIDFromHex(MessageReq.Receiver)
	if err != nil {
		http.Error(w, "failed to parse receiver user id ", http.StatusForbidden)
		return
	}

	currentUser, err := services.GetUserById(ObjectReceiverID)

	if err != nil || currentUser.ID.IsZero() {
		http.Error(w, "Receiver not found", http.StatusInternalServerError)
		return
	}

	if ObjectConversationId == primitive.NilObjectID || ObjectUserId == primitive.NilObjectID {
		http.Error(w, "Invalid object ID", http.StatusForbidden)
		return
	}

	user_in_conversation, err := services.GetUserInConversation(ObjectConversationId, ObjectUserId)

	if err != nil || !user_in_conversation {
		http.Error(w, "User not found  in the conversation ", http.StatusForbidden)
		return
	}

	message := model.Message{
		ConversationId: ObjectConversationId,
		Sender:         ObjectUserId,
		Receiver:       ObjectReceiverID,
		Body:           MessageReq.Body,
		Image:          MessageReq.Image,
		SeenBy:         []primitive.ObjectID{ObjectUserId},
		CreatedAt:      time.Now(),
	}

	new_message, err := services.CreateMesssage(message)

	if err != nil {
		http.Error(w, "Failed to create conversation ", http.StatusForbidden)
		return
	}

	chat_collection, err := configuration.GetCollection("chats")
	if err != nil {
		http.Error(w, "failed to get chat collection", http.StatusInternalServerError)
		return
	}
	_, err = chat_collection.UpdateOne(
		context.Background(),
		bson.M{"_id": ObjectConversationId},
		bson.M{
			"$push": bson.M{
				"messages": new_message["_id"],
			},
			"$set": bson.M{
				"lastMessageAt": message.CreatedAt.UTC().Format(time.RFC3339),
			},
		},
	)
	if err != nil {
		http.Error(w, "failed to update chat", http.StatusInternalServerError)
		return
	}

	//TODO : will add socket io
	manager := sockets.GetManager()

	data, err := json.Marshal(new_message)
	if err != nil {
		http.Error(w, "failed to marshal message", http.StatusInternalServerError)
		return
	}

	fmt.Println("passed in converting the data")

	// Create the WebSocket event
	event := sockets.Event{
		Type:    sockets.EventNewMessage,
		Payload: data,
	}

	// Check if the receiver is connected
	if client, ok := manager.Clients[MessageReq.Receiver]; ok {
		// Receiver is connected, send the message
		client.Egress <- event
		log.Printf("Message sent to receiver: %s", MessageReq.Receiver)
	} else {
		// Receiver is not connected
		log.Printf("Receiver %s is not connected. Message not sent.", MessageReq.Receiver)
	}

	json.NewEncoder(w).Encode(new_message)

}
