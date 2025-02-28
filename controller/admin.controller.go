package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/Abhishekh669/backend/configuration"
	"github.com/Abhishekh669/backend/model"
	"github.com/Abhishekh669/backend/services"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type LoginType struct {
	Email    string `json:"email" bson:"email"`
	Password string `json:"password" bson:"password"`
	UserId   string `json:"userId" bson:"userId"`
}

type VerifyUserType struct {
	UserId string `json:"userId,omitempty" bson:"userId,omitempty"`
}

type SetOrRemoveAdminType struct {
	UserId string `json:"userId,omitempty" bson:"userId,omitempty"`
}

type AcceptOrRejectType struct {
	ID     string `json:"id"`
	UserId string `json:"userId"`
}

func AcceptAbsentRequestHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Allow-Control-Allow-Methods", "POST")

	var Data AcceptOrRejectType

	if err := json.NewDecoder(r.Body).Decode(&Data); err != nil {
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	adminSession, err := configuration.GetAdminSessionData(r)

	if err != nil {
		http.Error(w, "invalid session", http.StatusBadRequest)
		return

	}

	ObjectUserId, err := primitive.ObjectIDFromHex(adminSession.UserId)
	if err != nil {
		http.Error(w, "failed to parse user id ", http.StatusForbidden)
		return
	}

	currentUser, err := services.GetUserById(ObjectUserId)

	if err != nil || currentUser.ID.IsZero() || !currentUser.IsAdmin {
		http.Error(w, "Not authenticated", http.StatusInternalServerError)
		return
	}

	ObjectAbsentId, err := primitive.ObjectIDFromHex(Data.ID)
	if err != nil {
		http.Error(w, "failed to parse given id ", http.StatusForbidden)
		return
	}

	ObjectAbsentUserId, err := primitive.ObjectIDFromHex(Data.UserId)
	if err != nil {
		http.Error(w, "failed to parse user id ", http.StatusForbidden)
		return
	}

	currentUserCheck, err := services.GetUserById(ObjectAbsentUserId)

	if err != nil || currentUserCheck.ID.IsZero() || !currentUser.IsAdmin {
		http.Error(w, "Not authenticated", http.StatusInternalServerError)
		return
	}

	fmt.Println("checkign hte absetn id : ", ObjectAbsentId)

	checkSearchId, err := services.GetAbsentRequestById(ObjectAbsentId, ObjectAbsentUserId)

	fmt.Println("i am her hoi new", checkSearchId)
	if err != nil || checkSearchId.ID.IsZero() {
		http.Error(w, "Request not found", http.StatusInternalServerError)
		return
	}

	acceptUser, err := services.AcceptAbsentRequest(ObjectAbsentId)

	if err != nil {
		http.Error(w, "Request not found", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(acceptUser)

}

func RejectAbsentRequestHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Allow-Control-Allow-Methods", "POST")

	var Data AcceptOrRejectType

	if err := json.NewDecoder(r.Body).Decode(&Data); err != nil {
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	adminSession, err := configuration.GetAdminSessionData(r)

	if err != nil {
		http.Error(w, "invalid session", http.StatusBadRequest)
		return

	}

	ObjectUserId, err := primitive.ObjectIDFromHex(adminSession.UserId)
	if err != nil {
		http.Error(w, "failed to parse user id ", http.StatusForbidden)
		return
	}

	currentUser, err := services.GetUserById(ObjectUserId)

	if err != nil || currentUser.ID.IsZero() || !currentUser.IsAdmin {
		http.Error(w, "Not authenticated", http.StatusInternalServerError)
		return
	}

	ObjectAbsentId, err := primitive.ObjectIDFromHex(Data.ID)
	if err != nil {
		http.Error(w, "failed to parse given id ", http.StatusForbidden)
		return
	}

	ObjectAbsentUserId, err := primitive.ObjectIDFromHex(Data.UserId)
	if err != nil {
		http.Error(w, "failed to parse user id ", http.StatusForbidden)
		return
	}

	currentUserCheck, err := services.GetUserById(ObjectAbsentUserId)

	if err != nil || currentUserCheck.ID.IsZero() || !currentUser.IsAdmin {
		http.Error(w, "Not authenticated", http.StatusInternalServerError)
		return
	}

	fmt.Println("checkign hte absetn id : ", ObjectAbsentId)

	checkSearchId, err := services.GetAbsentRequestById(ObjectAbsentId, ObjectAbsentUserId)

	fmt.Println("i am her hoi new", checkSearchId)
	if err != nil || checkSearchId.ID.IsZero() {
		http.Error(w, "Request not found", http.StatusInternalServerError)
		return
	}

	rejectUser, err := services.RejectAbsentRequest(ObjectAbsentId)

	if err != nil {
		http.Error(w, "Request not found", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(rejectUser)

}

func GetUserAbsentByIdHandler(w http.ResponseWriter, r *http.Request) {
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

	currentUser, err := services.GetUserById(ObjectUserId)

	if err != nil || currentUser.ID.IsZero() {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	params := mux.Vars(r)

	userId := params["id"]

	ObjectSearchId, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		http.Error(w, "failed to parse user id ", http.StatusForbidden)
		return
	}

	checkSearchId, err := services.GetUserById(ObjectSearchId)

	if err != nil || checkSearchId.ID.IsZero() {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	searchIdData, err := services.GetUserAbsentRequestById(ObjectSearchId)

	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(searchIdData)

}

func GetAllAbsentResultsHandler(w http.ResponseWriter, r *http.Request) {
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

	currentUser, err := services.GetUserById(ObjectUserId)

	if err != nil || currentUser.ID.IsZero() {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	absentResults, err := services.GetAllAbsentRequest()
	if err != nil {
		http.Error(w, "failed to get absent results", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(absentResults)

}

func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Allow-Control-Allow-Methods", "DELETE")

	params := mux.Vars(r)

	userId := params["id"]

	ObjectUserId, err := primitive.ObjectIDFromHex(userId)
	if err != nil || ObjectUserId.IsZero() {
		http.Error(w, "failed to parse user id ", http.StatusForbidden)
		return
	}

	user, err := services.GetUserById(ObjectUserId)

	if err != nil || user.ID.IsZero() {
		http.Error(w, "no user found", http.StatusNotFound)
		return
	}

	delete_user, err := services.DeleteUser(ObjectUserId)
	if err != nil {
		http.Error(w, "failed to delete user", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(delete_user)

}

func SetOrRemoveAdminHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Allow-Control-Allow-Methods", "POST")

	var UserIdReq SetOrRemoveAdminType
	fmt.Println(" i am for post okie remove or set admin")

	if err := json.NewDecoder(r.Body).Decode(&UserIdReq); err != nil {
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	fmt.Println("this ishte userid reques in set or remoe admin : ", UserIdReq)

	ObjectUserId, err := primitive.ObjectIDFromHex(UserIdReq.UserId)
	if err != nil || ObjectUserId.IsZero() {
		http.Error(w, "failed to parse user id ", http.StatusForbidden)
		return
	}

	fmt.Println("i am here in reject user handler")

	user, err := services.GetUserById(ObjectUserId)

	if err != nil || user.ID.IsZero() {
		http.Error(w, "no user found", http.StatusNotFound)
		return
	}

	if user.IsAdmin {
		remove_admin, err := services.RemoveAdmin(ObjectUserId)
		if err != nil {
			http.Error(w, "failed to remove as admin ", http.StatusForbidden)
			return
		}

		fmt.Println("after remvoing adin : ", remove_admin)

		json.NewEncoder(w).Encode(remove_admin)
	} else {
		set_admin, err := services.SetAdmin(ObjectUserId)
		if err != nil {
			http.Error(w, "failed to set as admin ", http.StatusForbidden)
			return
		}

		fmt.Println("aftering setting admin : ", set_admin)

		json.NewEncoder(w).Encode(set_admin)

	}

}

func RejectUserHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Allow-Control-Allow-Methods", "GET")

	params := mux.Vars(r)

	userId := params["id"]

	if userId == "" {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	sessionData, err := configuration.GetSessionData(r)

	if err != nil {
		http.Error(w, "invalid session", http.StatusBadRequest)
		return

	}
	adminSessionData, err := configuration.GetAdminSessionData(r)
	if err != nil {
		http.Error(w, "invalid admin session", http.StatusBadRequest)
		return

	}

	if adminSessionData.UserId != sessionData.UserId {
		http.Error(w, "Not allowed", http.StatusForbidden)
		return
	}

	ObjectUserId, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		http.Error(w, "failed to parse user id ", http.StatusForbidden)
		return
	}

	fmt.Println("i am here in reject user handler")

	user, err := services.GetUserById(ObjectUserId)

	if err != nil || user.ID.IsZero() {
		http.Error(w, "no user found", http.StatusNotFound)
		return
	}

	rejectUser, err := services.RejectUser(user.ID)

	if err != nil || !rejectUser {
		http.Error(w, "failed to reject user", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(rejectUser)

}

func GetUnVerifiedUserHandler(w http.ResponseWriter, r *http.Request) {
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

	user, err := services.GetUserById(ObjectUserId)

	if err != nil || user.ID.IsZero() || !user.IsAdmin {
		http.Error(w, "no user found", http.StatusForbidden)
		return
	}

	users, err := services.GetUnverifiedUser()

	if err != nil {
		http.Error(w, "failed to get users", http.StatusForbidden)
		return
	}

	json.NewEncoder(w).Encode(users)

}

func VerifyUserHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("i am fine hoi tw ")
	w.Header().Set("Content-Type", "application/json")
	fmt.Println("2nd step")
	w.Header().Set("Allow-Control-Allow-Methods", "GET")
	fmt.Println("3r step")

	params := mux.Vars(r)

	userId := params["id"]

	fmt.Println("i am first", userId)

	if userId == "" {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}
	fmt.Println("i am her")

	sessionData, err := configuration.GetSessionData(r)

	if err != nil {
		http.Error(w, "invalid session", http.StatusBadRequest)
		return

	}
	adminSessionData, err := configuration.GetAdminSessionData(r)
	if err != nil {
		http.Error(w, "invalid admin session", http.StatusBadRequest)
		return

	}

	if adminSessionData.UserId != sessionData.UserId {
		http.Error(w, "Not allowed", http.StatusForbidden)
		return
	}

	fmt.Println("i am in verify user handler")

	ObjectUserId, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		http.Error(w, "failed to parse user id ", http.StatusForbidden)
		return
	}
	fmt.Println("fine till now")
	user, err := services.GetUserById(ObjectUserId)
	fmt.Println("i am in user find : ", user)

	if err != nil || user.ID.IsZero() {
		http.Error(w, "no user found", http.StatusNotFound)
		return
	}

	verifyUser, err := services.VerifyUser(user.ID)

	if err != nil || !verifyUser {
		http.Error(w, "failed to verify user", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(verifyUser)

}

func AdminCreateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Allow-Control-Allow-Methods", "POST")
	var AdminReq model.AdminRequest

	if err := json.NewDecoder(r.Body).Decode(&AdminReq); err != nil {
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	AdminSessionData, err := configuration.GetAdminSessionData(r)

	if err != nil {
		http.Error(w, "invalid admin session", http.StatusBadRequest)
		return

	}

	if AdminSessionData.UserId != AdminReq.UserId {
		http.Error(w, "User not authenticated", http.StatusBadRequest)
		return
	}

	ObjectUserId, err := primitive.ObjectIDFromHex(AdminReq.UserId)
	if err != nil {
		http.Error(w, "failed to parse user id ", http.StatusForbidden)
		return
	}

	user, err := services.GetUserById(ObjectUserId)
	fmt.Println("This is the user i found in the admin : ", user)

	if err != nil || user.ID.IsZero() || !user.IsAdmin || !user.AllowCreate {
		http.Error(w, "Unauthorized  ", http.StatusForbidden)
		return
	}

	exist_admin, err := services.CheckAdmin(ObjectUserId, AdminReq.Email)

	fmt.Println("this is the exist admin : ", exist_admin)

	if err == nil {
		json.NewEncoder(w).Encode(exist_admin)
		return

	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(AdminReq.Password), bcrypt.DefaultCost)

	if err != nil {
		http.Error(w, "failed to create pass", http.StatusForbidden)
		return
	}

	newAdmin := model.Admin{
		Email:    AdminReq.Email,
		Password: string(hashedPassword),
		UserId:   ObjectUserId,
	}

	created_admin, err := services.CreateAdmin(newAdmin)

	if err != nil {
		http.Error(w, "failed to create admin", http.StatusForbidden)
		return
	}

	// w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(created_admin)

}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Allow-Control-Allow-Methods", "POST")
	fmt.Println("I am for login")
	var LoginData LoginType
	if err := json.NewDecoder(r.Body).Decode(&LoginData); err != nil {
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	ObjectUserId, err := primitive.ObjectIDFromHex(LoginData.UserId)
	if err != nil {
		http.Error(w, "failed to parse user id ", http.StatusForbidden)
		return
	}

	user, err := services.GetUserById(ObjectUserId)
	fmt.Println("This is the user i found in the admin : ", user)

	if err != nil || user.ID.IsZero() || !user.IsAdmin {
		http.Error(w, "Unauthorized  ", http.StatusForbidden)
		return
	}

	admin_data, err := services.CheckAdmin(ObjectUserId, LoginData.Email)

	fmt.Println("this is the exist admin : ", admin_data)

	if err != nil {
		http.Error(w, "no admin found ", http.StatusForbidden)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(admin_data.Password), []byte(LoginData.Password))

	if err != nil {
		http.Error(w, "invalid email or password", http.StatusForbidden)
		return
	}

	var (
		sessionKey      = []byte(os.Getenv("ADMIN_SECRET"))
		sessionStore    = sessions.NewCookieStore(sessionKey)
		sessionName     = "admin_token"
		sessionLifeTime = 7 * 24 * time.Hour // Session lasts for 7 days
	)

	fmt.Println("this is hte created admin new", sessionKey)
	//create admin token

	session, err := sessionStore.Get(r, sessionName)
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}
	fmt.Println("I am in session")
	session.Values["adminId"] = admin_data.ID.Hex()
	session.Values["email"] = admin_data.Email
	session.Values["userId"] = admin_data.UserId.Hex()
	session.Values["authenticated"] = true

	session.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   int(sessionLifeTime.Seconds()),
		HttpOnly: true,
		Secure:   false, //TODO : true in production only
		SameSite: http.SameSiteNoneMode,
	}

	err = session.Save(r, w)

	fmt.Println("I am for agian sessio ncheck", session.Values)

	if err != nil {
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}
	fmt.Println("I am mother fuckign hell")
	// w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Login successful"})

}

func GetAllAdminHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Allow-Control-Allow-Methods", "GET")
	admin_len, err := services.GetAllAdmin()

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed  to get admin : %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	responseData := map[string]interface{}{
		"message":      "got admin length successfully",
		"admin_length": admin_len,
	}
	err = json.NewEncoder(w).Encode(responseData)
	if err != nil {
		// If encoding the response fails, return an error response
		http.Error(w, fmt.Sprintf("Failed to encode created user: %v", err), http.StatusInternalServerError)
	}
}

func GetAdminFromTokenHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Allow-Control-Allow-Methods", "GET")

	AdminSessionData, err := configuration.GetAdminSessionData(r)

	if err != nil {
		http.Error(w, "invalid admin session", http.StatusBadRequest)
		return

	}
	if !AdminSessionData.Authenticated {
		http.Error(w, fmt.Sprintf("Not authenticated : %v", err), http.StatusInternalServerError)
		return
	}

	ObjectUserId, err := primitive.ObjectIDFromHex(AdminSessionData.UserId)
	if err != nil || ObjectUserId.IsZero() {
		http.Error(w, fmt.Sprintf("Failed  to get user id : %v", err), http.StatusInternalServerError)
		return
	}

	ObjectAdminId, err := primitive.ObjectIDFromHex(AdminSessionData.AdminId)
	if err != nil || ObjectAdminId.IsZero() {
		http.Error(w, fmt.Sprintf("Failed  to get admin id : %v", err), http.StatusInternalServerError)
		return
	}

	admin_collection, err := configuration.GetCollection("admin")

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed  to get admin collection : %v", err), http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var admin model.Admin

	err = admin_collection.FindOne(ctx, bson.M{"_id": ObjectAdminId, "userId": ObjectUserId}).Decode(&admin)

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed  to get admin : %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

	responseData := map[string]interface{}{
		"data": map[string]interface{}{
			"_id":    admin.ID,
			"email":  admin.Email,
			"userId": admin.UserId,
		},
	}

	err = json.NewEncoder(w).Encode(responseData)
	if err != nil {
		// If encoding the response fails, return an error response
		http.Error(w, fmt.Sprintf("Failed to get admin: %v", err), http.StatusInternalServerError)
	}

}
