package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Abhishekh669/backend/configuration"
	"github.com/Abhishekh669/backend/model"
	"github.com/Abhishekh669/backend/services"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func UpdateAbsentRequestHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Allow-Control-Allow-Methods", "POST")

	var UpdateData services.UpdateAbsentRequestType
	if err := json.NewDecoder(r.Body).Decode(&UpdateData); err != nil {
		fmt.Println("this is hte meian error : ", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
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

	currentUser, err := services.GetUserById(ObjectUserId)

	if err != nil || currentUser.ID.IsZero() {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	ObjectAbsentRequestId, err := primitive.ObjectIDFromHex(UpdateData.ID)
	if err != nil {
		http.Error(w, "failed to parse absent user id ", http.StatusForbidden)
		return
	}

	checkAbsentRecord, err := services.GetAbsentRequestById(ObjectAbsentRequestId, ObjectUserId)

	if err != nil || checkAbsentRecord.ID.IsZero() {

		if err != nil {
			http.Error(w, "failed to get absent request  ", http.StatusForbidden)
			return
		}
	}

	if checkAbsentRecord.Status == "rejected" || checkAbsentRecord.Status == "accepted" {
		http.Error(w, "failed to  update the absent record ", http.StatusForbidden)
		return

	}

	parsedDate, err := time.Parse(time.RFC3339, UpdateData.Date)
	if err != nil {
		http.Error(w, "Invalid date format", http.StatusBadRequest)
		return
	}

	updateAbsent, err := services.UpdateAbsentRequest(ObjectAbsentRequestId, UpdateData.Reason, parsedDate)
	if err != nil || !updateAbsent {
		http.Error(w, "failed to parse update the absent record ", http.StatusForbidden)
		return
	}
	fmt.Println("passed update ")

	json.NewEncoder(w).Encode(updateAbsent)

}

func CreateStudentAbsentHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Allow-Control-Allow-Methods", "POST")

	fmt.Println("i am for creating oki")

	var AbsentData model.AbsentRequest

	if err := json.NewDecoder(r.Body).Decode(&AbsentData); err != nil {
		fmt.Println("this is hte meian error : ", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
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

	currentUser, err := services.GetUserById(ObjectUserId)

	if err != nil || currentUser.ID.IsZero() {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	fmt.Println("i am fine till here")

	checkTodayAbsent, err := services.CheckTodayAbsentRequest(ObjectUserId)
	if err != nil {
		http.Error(w, "Failed to check absent", http.StatusBadRequest)
		return
	}

	if checkTodayAbsent {
		http.Error(w, "Already absent exist of today", http.StatusBadRequest)
		return
	}

	fmt.Println("just before creating datea")

	absentData := model.Absent{
		UserID:    ObjectUserId,
		Date:      AbsentData.Date,
		Reason:    AbsentData.Reason,
		Name:      AbsentData.Name,
		Status:    "pending",
		CodeName:  AbsentData.CodeName,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	absent, err := services.CreateAbsent(absentData)

	if err != nil || !absent {
		http.Error(w, "failed to create absetn", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(absent)

}

func GetUserAllAbsentResultsHandler(w http.ResponseWriter, r *http.Request) {
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

	absentResults, err := services.GetUserAllAbsentResults(ObjectUserId)
	if err != nil {
		http.Error(w, "failed to get absent results", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(absentResults)

}
