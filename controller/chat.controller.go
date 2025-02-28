package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Abhishekh669/backend/configuration"
	"github.com/Abhishekh669/backend/model"
	"github.com/Abhishekh669/backend/services"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PaginationQuery struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// Define default values for limit and offset

func CreateChatHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("i am here")
	w.Header().Set("Content-Type", "application/json")
	fmt.Println("Fialed ot get json")
	w.Header().Set("Allow-Control-Allow-Methods", "POST")
	var chatReq model.ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&chatReq); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	fmt.Println("Chat is ", chatReq)

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

	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}
	if currentUser.ID == primitive.NilObjectID || currentUser.Email == "" {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	members := append(chatReq.Members, sessionData.UserId)

	var memberIDs []primitive.ObjectID

	for _, member := range members {
		memberID, err := primitive.ObjectIDFromHex(member)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid member ID format: %s", err), http.StatusBadRequest)
			return
		}

		memberIDs = append(memberIDs, memberID)
	}

	// Query to check if the conversation already exists

	query := bson.M{"members": bson.M{"$all": memberIDs, "$size": 2}}

	//check if conversation already exist in the db

	chat_collection, err := configuration.GetCollection("chats")
	if err != nil {
		http.Error(w, "chat collection not found", http.StatusInternalServerError)
		return
	}

	var existingChat model.Chat

	err = chat_collection.FindOne(context.Background(), query).Decode(&existingChat)

	if err == nil {
		aggegratePipeline := []bson.M{
			{
				"$match": query,
			},
			{
				"$lookup": bson.M{
					"from":         "users",
					"localField":   "members",
					"foreignField": "_id",
					"as":           "user_details",
				},
			},
		}
		cursor, err := chat_collection.Aggregate(context.Background(), aggegratePipeline)

		if err != nil {
			http.Error(w, "Error populating user details", http.StatusInternalServerError)
			return
		}

		var populatedChat model.Chat

		if cursor.Next(context.Background()) {
			if err := cursor.Decode(&populatedChat); err != nil {
				http.Error(w, "Failed to decode populated conversation", http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(populatedChat)
		return

	}

	newChat := model.Chat{
		Members: memberIDs,
	}

	created_chat, err := services.CreateChat(newChat)

	if err != nil {
		http.Error(w, "Failed to create conversation", http.StatusInternalServerError)
		return
	}

	for _, memberId := range memberIDs {
		users, err := configuration.GetCollection("users")
		users.UpdateOne(
			context.Background(),
			bson.M{"_id": memberId},
			bson.M{"$addToSet": bson.M{"conversations": created_chat.ID}},
		)
		if err != nil {
			http.Error(w, "Failed to update user with new conversation", http.StatusInternalServerError)
			return
		}

	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(created_chat)

}

// do improvement
func GetConversationByIdHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Allow-Control-Allow-Methods", "GET")

	params := mux.Vars(r)
	conversationId := params["id"]

	ObjectConversationId, err := primitive.ObjectIDFromHex(conversationId)
	if err != nil {
		http.Error(w, "failed to parse conversation id ", http.StatusForbidden)
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

	//fetch members
	var members []model.User

	for _, memberID := range conversation.Members {
		user, err := services.GetUserById(memberID)
		if err != nil {
			http.Error(w, "failed to fetch user", http.StatusInternalServerError)
			return
		}
		members = append(members, user)
	}

	//paginate messages

	// var query PaginationQuery

	// query.Limit = 4
	// query.Offset = 0

	// // Fetch limit and offset from query parameters
	// limitStr := r.URL.Query().Get("limit")
	// offsetStr := r.URL.Query().Get("offset")

	// // If limit is provided, use it, but make sure it's within bounds (e.g., max limit of 3)
	// if limitStr != "" {
	// 	limit, err := strconv.Atoi(limitStr)
	// 	if err != nil {
	// 		http.Error(w, "Invalid limit", http.StatusBadRequest)
	// 		return
	// 	}
	// 	query.Limit = limit
	// }

	// // If offset is provided, use it
	// if offsetStr != "" {
	// 	offset, err := strconv.Atoi(offsetStr)
	// 	if err != nil {
	// 		http.Error(w, "Invalid offset", http.StatusBadRequest)
	// 		return
	// 	}
	// 	query.Offset = offset
	// }

	// if query.Limit > 4 {
	// 	query.Limit = 4
	// }

	// //fetch messages

	// var populatedMessages []map[string]interface{}

	// if len(conversation.Messages) > 0 {
	// 	paginatedMessages := conversation.Messages[query.Offset:min(query.Offset+query.Limit, len(conversation.Messages))]

	// 	fmt.Println("Paginated Message for now : ", paginatedMessages)

	// 	for _, messageID := range paginatedMessages {
	// 		message, err := services.GetMessageById(messageID)
	// 		fmt.Println(":i am here ")

	// 		if err != nil {
	// 			http.Error(w, "failed to fetch message", http.StatusInternalServerError)
	// 			return
	// 		}
	// 		fmt.Println("I am for sender")
	// 		sender, err := services.GetUserById(message.Sender)

	// 		if err != nil {
	// 			http.Error(w, "failed to fetch sender", http.StatusInternalServerError)
	// 			return
	// 		}

	// 		var seenByUsers []model.User

	// 		for _, seenByID := range message.SeenBy {
	// 			seenByUser, err := services.GetUserById(seenByID)

	// 			if err != nil {
	// 				http.Error(w, "failed to fetch seen_by user", http.StatusInternalServerError)
	// 				return
	// 			}

	// 			seenByUsers = append(seenByUsers, seenByUser)

	// 		}

	// 		messageJson := map[string]interface{}{
	// 			"_id":            message.ID,
	// 			"conversationId": message.ConversationId,
	// 			"body":           message.Body,
	// 			"sender": map[string]interface{}{
	// 				"_id":   sender.ID,
	// 				"name":  sender.Name,
	// 				"image": sender.Image,
	// 				"email": sender.Email,
	// 			},
	// 			"seenBy": seenByUsers,
	// 		}

	// 		populatedMessages = append(populatedMessages, messageJson)

	// 	}

	// }

	response := map[string]interface{}{
		"_id":           conversation.ID,
		"members":       members,
		"createdAt":     conversation.CreatedAt,
		"lastMessageAt": conversation.LastMessageAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}
