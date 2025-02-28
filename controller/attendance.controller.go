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

//TODO : handle the get attendance by date

func GetTodayAttendanceHandler(w http.ResponseWriter, r *http.Request) {
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

	if err != nil || user.ID.IsZero() {
		http.Error(w, "failed to get user data", http.StatusNotFound)
		return
	}

	fmt.Println("I am here for the today attendance : ")

	TodayAttendance, err := services.CheckTodayAttendanceNoFilter(time.Now())

	if err != nil || TodayAttendance.ID.IsZero() {
		http.Error(w, "failed to get  today attendance", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(TodayAttendance)

}

func UpdateAttendanceTrackerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Allow-Control-Allow-Methods", "POST")

	var AttendanceTrackerData model.UpdateAttendanceTracker
	if err := json.NewDecoder(r.Body).Decode(&AttendanceTrackerData); err != nil {
		http.Error(w, "failed to parse data ", http.StatusBadRequest)
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

	user, err := services.GetUserById(ObjectUserId)

	if err != nil || user.ID.IsZero() {
		http.Error(w, "failed to get user data", http.StatusNotFound)
		return
	}

	parsedDate, err := time.Parse(time.RFC3339, AttendanceTrackerData.Date)
	if err != nil {
		http.Error(w, "Invalid date format", http.StatusBadRequest)
		return
	}

	check_attendance, err := services.CheckTodayAttendance(parsedDate, AttendanceTrackerData.Field)
	if err != nil || check_attendance.ID.IsZero() {
		http.Error(w, "no attendance found", http.StatusBadRequest)
		return
	}

	updatedStatus, err := services.UpdateAttendace(parsedDate, AttendanceTrackerData)
	if err != nil || !updatedStatus {
		http.Error(w, "failed to update the attendance", http.StatusBadRequest)
		return

	}

	json.NewEncoder(w).Encode(updatedStatus)

}

func CreateAttendanceTrackerHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Allow-Control-Allow-Methods", "POST")

	var AttendanceTrackerData model.AttendanceTrackerRequest

	if err := json.NewDecoder(r.Body).Decode(&AttendanceTrackerData); err != nil {
		http.Error(w, "failed to parse data ", http.StatusBadRequest)
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

	user, err := services.GetUserById(ObjectUserId)

	if err != nil || user.ID.IsZero() {
		http.Error(w, "failed to get user data", http.StatusNotFound)
		return
	}

	newAttendanceTracker := model.AttendanceTrackerRequest{
		Attendance:  AttendanceTrackerData.Attendance,
		Date:        time.Now(),                  // Set the current date
		Field:       AttendanceTrackerData.Field, // Assume Field comes from the request
		SubmittedBy: user.Name,                   // Set SubmittedBy to the name of the user (or other property)
		CreatedAt:   time.Now(),                  // Set CreatedAt to the current time
	}

	success, err := services.CreateAttendanceTracker(newAttendanceTracker)

	if err != nil || !success {
		http.Error(w, "failed to create attendacne tracker", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(success)

}
