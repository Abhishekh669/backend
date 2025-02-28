package controller

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Abhishekh669/backend/configuration"
	"github.com/Abhishekh669/backend/model"
	"github.com/Abhishekh669/backend/services"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OnboardingType struct {
	CodeName      string `json:"codeName,omitempty" bson:"codeName,omitempty"`
	PhoneNumber   string `json:"phoneNumber,omitempty" bson:"phoneNumber,omitempty"`
	Address       string `json:"address,omitempty" bson:"address,omitempty"`
	Age           int    `json:"age" bson:"age"`
	Qualification string `json:"qualification,omitempty" bson:"qualification,omitempty"`
	Field         string `json:"field,omitempty" bson:"field,omitempty"`
	UserId        string `json:"userId,omitempty" bson:"userId,omitempty"`
	MainField     string `json:"mainField,omitempty" bson:"mainField,omitempty"`
}

func GetUserByIdHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Allow-Control-Allow-Methods", "GET")

	params := mux.Vars(r)

	userId := params["id"]

	ObjectUserId, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		http.Error(w, "failed to parse user id ", http.StatusForbidden)
		return
	}

	user, err := services.GetUserById(ObjectUserId)

	if err != nil {
		http.Error(w, "failed to get user data", http.StatusNotFound)
		return
	}

	err = json.NewEncoder(w).Encode(user)
	if err != nil {
		fmt.Println("i am here in error ")
		// If encoding the response fails, return an error response
		http.Error(w, fmt.Sprintf("Failed to encode created user: %v", err), http.StatusInternalServerError)
	}

}

func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Allow-Control-Allow-Methods", "POST")

	var user model.User

	err := json.NewDecoder(r.Body).Decode(&user)

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse user data: %v", err), http.StatusBadRequest)
		return
	}

	created_user, err := services.CreateUser(user)

	if err != nil {
		// If an error occurs while creating the user, return an error response
		http.Error(w, fmt.Sprintf("Failed to create user: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(created_user)

	if err != nil {
		// If encoding the response fails, return an error response
		http.Error(w, fmt.Sprintf("Failed to encode created user: %v", err), http.StatusInternalServerError)
	}

}

func GetAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Allow-Control-Allow-Methods", "GET")

	sessionData, err := configuration.GetSessionData(r)

	if err != nil {
		http.Error(w, "invalid session", http.StatusBadRequest)
		return

	}

	ObjectUserId, err := primitive.ObjectIDFromHex(sessionData.UserId)
	if err != nil {
		http.Error(w, "failed to parse user id ", http.StatusForbidden)
		return
	}

	check_user, err := services.GetUserById(ObjectUserId)

	if err != nil || check_user.ID.IsZero() {
		http.Error(w, "failed to get user data", http.StatusNotFound)
		return
	}

	users, err := services.GetAllUsers()

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed  to get users: %v", err), http.StatusInternalServerError)
		return
	}

	// w.WriteHeader(http.StatusFound)
	// w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(users)
	if err != nil {
		// If encoding the response fails, return an error response
		http.Error(w, fmt.Sprintf("Failed to encode created user: %v", err), http.StatusInternalServerError)
	}

}

func OnboardingUserHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Allow-Control-Allow-Methods", "POST")

	var OnboardingData OnboardingType

	if err := json.NewDecoder(r.Body).Decode(&OnboardingData); err != nil {
		http.Error(w, "invlaid payload", http.StatusBadRequest)
		return
	}

	if OnboardingData.UserId == "" {
		http.Error(w, "user id not passed", http.StatusBadRequest)
		return
	}

	sessionData, err := configuration.GetSessionData(r)

	if err != nil {
		http.Error(w, "invalid session", http.StatusBadRequest)
		return

	}

	if sessionData.UserId != OnboardingData.UserId {
		http.Error(w, "User not authenticated", http.StatusBadRequest)
		return
	}

	// user_collection, err := configuration.GetCollection("users")

	// if err != nil {
	// 	http.Error(w, "failed to get user collection", http.StatusInternalServerError)
	// 	return
	// }

	fmt.Println("I am in check user section in onbaording ")
	ObjectUserId, err := primitive.ObjectIDFromHex(OnboardingData.UserId)
	if err != nil {
		http.Error(w, "failed to parse user id ", http.StatusForbidden)
		return
	}

	//update the user

	var OnboardedUser = model.OnboardingRequest{
		CodeName:      OnboardingData.CodeName,
		PhoneNumber:   OnboardingData.PhoneNumber,
		Address:       OnboardingData.Address,
		Age:           OnboardingData.Age,
		Qualification: OnboardingData.Qualification,
		Field:         OnboardingData.Field,
		MainField:     OnboardingData.MainField,
	}

	updated_user, err := services.OnboardingUser(OnboardedUser, ObjectUserId)

	if err != nil {
		http.Error(w, "failed to update the user ", http.StatusForbidden)
		return
	}

	err = json.NewEncoder(w).Encode(updated_user)
	if err != nil {
		// If encoding the response fails, return an error response
		http.Error(w, fmt.Sprintf("Failed to encode updated user: %v", err), http.StatusInternalServerError)
	}

}
