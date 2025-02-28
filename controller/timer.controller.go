package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Abhishekh669/backend/configuration"
	"github.com/Abhishekh669/backend/model"
	"github.com/Abhishekh669/backend/services"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EditTimerRequest struct {
	TimerId string    `json:"timerId,omitempty" bson:"timerId,omitempty"`
	UserId  string    `json:"userId,omitempty" bson:"userId,omitempty"`
	Name    string    `json:"name,omitempty" bson:"name,omitempty"`
	EndDate time.Time `json:"endDate,omitempty" bson:"endDate,omitempty"`
	Type    string    `json:"type,omitempty" bson:"type,omitempty"`
}

type DeleteTimerRequest struct {
	WorkspaceId string `json:"workspaceId,omitempty" bson:"workspaceId,omitempty"`
}

func DeleteTimerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Allow-Control-Allow-Methods", "POST")

	params := mux.Vars(r)

	timerId := params["id"]

	var DeleteTimeRequest DeleteTimerRequest
	err := json.NewDecoder(r.Body).Decode(&DeleteTimeRequest)
	fmt.Println("this hte update data in server : ", DeleteTimeRequest)

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse user data: %v", err), http.StatusBadRequest)
		return
	}

	fmt.Println("this is hte workspace Id in delete tiem : ", DeleteTimeRequest.WorkspaceId)
	fmt.Println("this is hte timer Id in delete tiem : ", timerId)

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

	if currentUser.ID == primitive.NilObjectID || currentUser.Email == "" || !currentUser.IsOnBoarded {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	ObjectWorkspaceId, err := primitive.ObjectIDFromHex(DeleteTimeRequest.WorkspaceId)
	if err != nil {
		http.Error(w, "failed to parse workspace id ", http.StatusForbidden)
		return
	}

	check_workspace, err := services.GetTimeWorkspaceById(ObjectWorkspaceId, ObjectUserId)

	if err != nil || check_workspace.ID.IsZero() || check_workspace.UserId.IsZero() {
		http.Error(w, "failed to get workspace  ", http.StatusForbidden)
		return
	}

	ObjectTimerId, err := primitive.ObjectIDFromHex(timerId)

	if err != nil {
		http.Error(w, "Failed to parse userid ", http.StatusInternalServerError)
		return
	}

	check_timer, err := services.GetTimerById(ObjectTimerId, ObjectUserId, ObjectWorkspaceId)

	if err != nil || check_timer.ID.IsZero() || check_timer.UserId.IsZero() || check_timer.TimerWorkspaceId.IsZero() {
		http.Error(w, "no timer exists ", http.StatusForbidden)
		return
	}

	var delete_data = services.DeleteWorkspaceTimerType{
		UserId:      ObjectUserId,
		TimerId:     ObjectTimerId,
		WorkspaceId: ObjectWorkspaceId,
	}

	deleteTImer, err := services.DeleteWorkspaceTimer(delete_data)

	if err != nil {
		http.Error(w, "failed to delete timer ", http.StatusForbidden)
		return
	}

	json.NewEncoder(w).Encode(deleteTImer)

}

func DeleteWorkspaceHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Allow-Control-Allow-Methods", "DELETE")

	params := mux.Vars(r)

	workspaceId := params["id"]
	fmt.Println("this is workspaceid in the delete : ", workspaceId)

	if workspaceId == "" {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

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

	if currentUser.ID == primitive.NilObjectID || currentUser.Email == "" || !currentUser.IsOnBoarded {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	ObjectWorkspaceId, err := primitive.ObjectIDFromHex(workspaceId)
	if err != nil {
		http.Error(w, "failed to parse workspace id ", http.StatusForbidden)
		return
	}

	check_workspace, err := services.GetTimeWorkspaceById(ObjectWorkspaceId, ObjectUserId)

	if err != nil || check_workspace.ID.IsZero() || check_workspace.UserId.IsZero() {
		http.Error(w, "failed to get workspace  ", http.StatusForbidden)
		return
	}

	deleteWorkspace, err := services.DeleteWorkspace(ObjectWorkspaceId)
	if err != nil || check_workspace.ID.IsZero() || check_workspace.UserId.IsZero() {
		http.Error(w, "failed to delete workspace  ", http.StatusForbidden)
		return
	}

	json.NewEncoder(w).Encode(deleteWorkspace)

}

func EditTimerDataHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Allow-Control-Allow-Methods", "POST")

	var timerData EditTimerRequest

	err := json.NewDecoder(r.Body).Decode(&timerData)
	fmt.Println("this hte update data in server : ", timerData)

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse user data: %v", err), http.StatusBadRequest)
		return
	}

	sessionData, err := configuration.GetSessionData(r)

	if err != nil {
		http.Error(w, "invalid session", http.StatusBadRequest)
		return

	}

	if sessionData.UserId != timerData.UserId {
		http.Error(w, "not authenticated", http.StatusBadRequest)
		return
	}

	ObjectUserId, err := primitive.ObjectIDFromHex(sessionData.UserId)

	if err != nil {
		http.Error(w, "Failed to parse userid ", http.StatusInternalServerError)
		return
	}

	ObjectTimerId, err := primitive.ObjectIDFromHex(timerData.TimerId)

	if err != nil {
		http.Error(w, "Failed to parse userid ", http.StatusInternalServerError)
		return
	}

	currentUser, err := services.GetUserById(ObjectUserId)

	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	if currentUser.ID == primitive.NilObjectID || currentUser.Email == "" || !currentUser.IsOnBoarded {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	params := mux.Vars(r)

	workspaceId := params["id"]

	ObjectWorkspaceId, err := primitive.ObjectIDFromHex(workspaceId)
	if err != nil {
		http.Error(w, "failed to parse workspace id ", http.StatusForbidden)
		return
	}

	check_timer, err := services.GetTimerById(ObjectTimerId, ObjectUserId, ObjectWorkspaceId)

	if err != nil || check_timer.ID.IsZero() || check_timer.UserId.IsZero() || check_timer.TimerWorkspaceId.IsZero() {
		http.Error(w, "no timer exists ", http.StatusForbidden)
		return
	}

	fmt.Println("i have passed man test")

	new_timer_data := model.Timer{
		ID:      ObjectTimerId,
		Name:    timerData.Name,
		EndDate: timerData.EndDate,
		Type:    timerData.Type,
	}

	edit_timer, err := services.EditTimerData(new_timer_data)
	if err != nil {
		http.Error(w, "failed to update timer ", http.StatusForbidden)
		return
	}
	fmt.Println("this is hte edit timer : ", edit_timer)

	json.NewEncoder(w).Encode(edit_timer)

}

func EditTimerWorkspaceHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Allow-Control-Allow-Methods", "POST")

	var timerData model.TimerWorkspaceRequest
	err := json.NewDecoder(r.Body).Decode(&timerData)

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse user data: %v", err), http.StatusBadRequest)
		return
	}

	fmt.Println("this is the update workspace : ", timerData)

	sessionData, err := configuration.GetSessionData(r)

	if err != nil {
		http.Error(w, "invalid session", http.StatusBadRequest)
		return

	}

	if sessionData.UserId != timerData.UserId {
		http.Error(w, "not authenticated", http.StatusBadRequest)
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

	if currentUser.ID == primitive.NilObjectID || currentUser.Email == "" || !currentUser.IsOnBoarded {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	params := mux.Vars(r)

	workspaceId := params["id"]

	ObjectWorkspaceId, err := primitive.ObjectIDFromHex(workspaceId)
	if err != nil {
		http.Error(w, "failed to parse workspace id ", http.StatusForbidden)
		return
	}

	check_workspace, err := services.GetTimeWorkspaceById(ObjectWorkspaceId, ObjectUserId)

	if err != nil || check_workspace.ID.IsZero() || check_workspace.UserId.IsZero() {
		http.Error(w, "failed to parse workspace id ", http.StatusForbidden)
		return
	}

	var workspaceData = model.TimerWorkspace{
		Name:        timerData.Name,
		Description: timerData.Description,
		ID:          check_workspace.ID,
	}

	fmt.Println("this hsite workspaceDAta : ", workspaceData)

	edit_workspace, err := services.EditTimerWorkspaceData(workspaceData)

	if err != nil {
		http.Error(w, "failed to update workspace ", http.StatusForbidden)
		return
	}

	json.NewEncoder(w).Encode(edit_workspace)

}

func GetUserTimerWorkspacesHandler(w http.ResponseWriter, r *http.Request) {
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

	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	if currentUser.ID == primitive.NilObjectID || currentUser.Email == "" || !currentUser.IsOnBoarded {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	timerWorkspaces, err := services.GetUserTimerWorkspaces(ObjectUserId)

	if err != nil {
		http.Error(w, "failed to get workspaces", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(timerWorkspaces)

}

func CreateTimerWorkspaceHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Allow-Control-Allow-Methods", "GET")

	var timerWorkspace model.TimerWorkspaceRequest

	err := json.NewDecoder(r.Body).Decode(&timerWorkspace)
	fmt.Println("this is the timer data for now : ", timerWorkspace)

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse user data: %v", err), http.StatusBadRequest)
		return
	}

	sessionData, err := configuration.GetSessionData(r)

	if err != nil {
		http.Error(w, "invalid session", http.StatusBadRequest)
		return

	}

	if sessionData.UserId != timerWorkspace.UserId {
		http.Error(w, "user not authenticated", http.StatusForbidden)
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

	var new_workspace = model.TimerWorkspace{
		UserId:      ObjectUserId,
		Name:        timerWorkspace.Name,
		Description: timerWorkspace.Description,
		CreatedAt:   time.Now(),
	}

	created_workspace, err := services.CreateTimerWorkspace(new_workspace)

	if err != nil {
		http.Error(w, "failed to create timer workspace", http.StatusInternalServerError)
		return
	}

	if created_workspace.ID.IsZero() || created_workspace.UserId.IsZero() {
		http.Error(w, "failed to create timer workspace", http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(created_workspace)
	if err != nil {
		fmt.Println("i am here in error ")
		// If encoding the response fails, return an error response
		http.Error(w, fmt.Sprintf("Failed to encode  timer data: %v", err), http.StatusInternalServerError)
	}

}

// func GetTimerByIdHandler(w http.ResponseWriter, r *http.Request) {

// 	w.Header().Set("Content-Type", "application/json")
// 	w.Header().Set("Allow-Control-Allow-Methods", "GET")

// 	params := mux.Vars(r)

// 	timerId := params["id"]

// 	sessionData, err := configuration.GetSessionData(r)

// 	if err != nil {
// 		http.Error(w, "invalid session", http.StatusBadRequest)
// 		return

// 	}

// 	if sessionData.UserId != "" {
// 		http.Error(w, "User not authenticated", http.StatusBadRequest)
// 		return
// 	}

// 	userId := sessionData.UserId
// 	ObjectUserId, err := primitive.ObjectIDFromHex(userId)
// 	if err != nil || ObjectUserId.IsZero() {
// 		http.Error(w, "failed to parse timer id ", http.StatusForbidden)
// 		return
// 	}

// 	ObjectTimerId, err := primitive.ObjectIDFromHex(timerId)
// 	if err != nil {
// 		http.Error(w, "failed to parse timer id ", http.StatusForbidden)
// 		return
// 	}

// 	timer, err := services.GetTimerById(ObjectTimerId, ObjectUserId)

// 	if err != nil {
// 		http.Error(w, "failed to get timer data", http.StatusNotFound)
// 		return
// 	}

// 	fmt.Println("this is the timer data okie in server : ", timer)

// 	err = json.NewEncoder(w).Encode(timer)
// 	if err != nil {
// 		fmt.Println("i am here in error ")
// 		// If encoding the response fails, return an error response
// 		http.Error(w, fmt.Sprintf("Failed to encode  timer data: %v", err), http.StatusInternalServerError)
// 	}

// }

func GetUserWorkspaceTimersController(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Allow-Control-Allow-Methods", "GET")

	params := mux.Vars(r)

	workspaceId := params["id"]

	fmt.Println("workspace Id : ", workspaceId)
	if workspaceId == "" {
		http.Error(w, "no user id provided", http.StatusBadRequest)
		return
	}

	fmt.Println("this ishte workspace id ")

	ObjectTimerWorkspaceId, err := primitive.ObjectIDFromHex(workspaceId)
	if err != nil {
		http.Error(w, "failed to parse timer workspace id ", http.StatusForbidden)
		return
	}

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

	check_timer_workspace, err := services.GetTimeWorkspaceById(ObjectTimerWorkspaceId, ObjectUserId)
	if err != nil || check_timer_workspace.ID.IsZero() || check_timer_workspace.UserId.IsZero() {
		http.Error(w, "no workspace present", http.StatusForbidden)
		return
	}

	result, err := services.GetUserWorkspaceTimers(ObjectUserId, ObjectTimerWorkspaceId)
	if err != nil {
		http.Error(w, "failed to get workspace timers", http.StatusForbidden)
		return
	}

	json.NewEncoder(w).Encode(result)

}

func CreateTimerController(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Allow-Control-Allow-Methods", "POST")

	var timer model.TimerRequest

	err := json.NewDecoder(r.Body).Decode(&timer)
	fmt.Println("this is the timer data for now : ", timer)

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse user data: %v", err), http.StatusBadRequest)
		return
	}

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

	ObjectTimerWorkspaceId, err := primitive.ObjectIDFromHex(timer.TimerWorkspaceId)

	if err != nil {
		http.Error(w, "Failed to parse userid ", http.StatusInternalServerError)
		return
	}

	if currentUser.ID == primitive.NilObjectID || currentUser.Email == "" {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	var timer_data = model.Timer{
		Name:             timer.Name,
		Type:             timer.Type,
		UserId:           ObjectUserId,
		TimerWorkspaceId: ObjectTimerWorkspaceId,
		EndDate:          timer.EndDate,
		CreatedAt:        time.Now(),
	}

	new_timer, err := services.CreateTimer(timer_data)

	if err != nil || new_timer.ID.IsZero() {
		http.Error(w, "failed to create new timer", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(new_timer)
}
