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

func CheckTodayAttendanceNoFilter(date time.Time) (model.AttendanceTracker, error) {
	// Get the attendance collection
	attendance_tracker_collection, err := configuration.GetCollection("attendanceTracker")
	if err != nil {
		return model.AttendanceTracker{}, fmt.Errorf("failed to get the attendance collection: %v", err)
	}

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Format the input date to UTC and extract only the date part (ignoring time)
	formattedDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)

	// Define the filter to check for today's attendance
	filter := bson.M{
		"date": bson.M{
			"$gte": formattedDate,                     // Greater than or equal to the start of the day
			"$lt":  formattedDate.Add(24 * time.Hour), // Less than the start of the next day
		},
	}

	// Query the database
	var attendance model.AttendanceTracker
	err = attendance_tracker_collection.FindOne(ctx, filter).Decode(&attendance)

	// Handle the result
	if err == mongo.ErrNoDocuments {
		return model.AttendanceTracker{}, nil // Attendance not found
	} else if err != nil {
		return model.AttendanceTracker{}, fmt.Errorf("error querying attendance: %v", err)
	}

	// Attendance found, return the document and true
	return attendance, nil
}

func CheckTodayAttendance(date time.Time, field string) (model.AttendanceTracker, error) {
	// Get the attendance collection
	attendance_tracker_collection, err := configuration.GetCollection("attendanceTracker")
	if err != nil {
		return model.AttendanceTracker{}, fmt.Errorf("failed to get the attendance collection: %v", err)
	}

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Format the input date to UTC and extract only the date part (ignoring time)
	formattedDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)

	// Define the filter to check for today's attendance
	filter := bson.M{
		"date": bson.M{
			"$gte": formattedDate,                     // Greater than or equal to the start of the day
			"$lt":  formattedDate.Add(24 * time.Hour), // Less than the start of the next day
		},
		"field": field, // Additional filter for the specific field
	}

	// Query the database
	var attendance model.AttendanceTracker
	err = attendance_tracker_collection.FindOne(ctx, filter).Decode(&attendance)

	// Handle the result
	if err == mongo.ErrNoDocuments {
		return model.AttendanceTracker{}, nil // Attendance not found
	} else if err != nil {
		return model.AttendanceTracker{}, fmt.Errorf("error querying attendance: %v", err)
	}

	// Attendance found, return the document and true
	return attendance, nil
}

func UpdateAttendace(date time.Time, attendanceData model.UpdateAttendanceTracker) (bool, error) {
	attendance_tracker_collection, err := configuration.GetCollection("attendanceTracker")
	if err != nil {
		return false, fmt.Errorf("failed to get the attendance collection")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ObjectAttendanceID, err := primitive.ObjectIDFromHex(attendanceData.ID)
	if err != nil {
		return false, fmt.Errorf("failed to get id")
	}

	filter := bson.M{
		"_id":   ObjectAttendanceID,
		"field": attendanceData.Field,
	}

	update := bson.M{
		"$set": bson.M{
			"date":        date,
			"attendance":  attendanceData.Attendance,
			"submittedBy": attendanceData.SubmittedBy,
			"createdAt":   time.Now(),
		},
	}

	updatedResult, err := attendance_tracker_collection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))

	if err != nil {
		return false, fmt.Errorf("failed to update reuslt")
	}

	if updatedResult.ModifiedCount == 0 {
		// Record updated successfully

		return false, fmt.Errorf("failed to update result")
	}
	return true, nil
}

func CreateAttendanceTracker(attendance model.AttendanceTrackerRequest) (bool, error) {
	attendance_tracker_collection, err := configuration.GetCollection("attendanceTracker")
	if err != nil {
		return false, fmt.Errorf("failed to get the attendance collection")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	insertedID, err := attendance_tracker_collection.InsertOne(ctx, attendance)

	if err != nil || insertedID.InsertedID.(primitive.ObjectID).IsZero() {
		return false, fmt.Errorf("failed to create attendacne: %v", err)
	}

	return true, nil
}

func GetAttendanceByDate(date time.Time) (model.AttendanceTracker, error) {
	// Get the attendance collection
	attendance_tracker_collection, err := configuration.GetCollection("attendanceTracker")
	if err != nil {
		return model.AttendanceTracker{}, fmt.Errorf("failed to get the attendance collection: %v", err)
	}

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Format the input date to UTC and extract only the date part (ignoring time)
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.Add(24 * time.Hour) // End of the day

	// Define the filter to match documents where the date falls within the given day
	filter := bson.M{
		"date": bson.M{
			"$gte": startOfDay, // Greater than or equal to the start of the day
			"$lt":  endOfDay,   // Less than the start of the next day
		},
	}

	// Query the database
	var attendance model.AttendanceTracker
	err = attendance_tracker_collection.FindOne(ctx, filter).Decode(&attendance)

	// Handle the result
	if err == mongo.ErrNoDocuments {
		return model.AttendanceTracker{}, fmt.Errorf("no attendance record found for the given date")
	} else if err != nil {
		return model.AttendanceTracker{}, fmt.Errorf("error retrieving attendance: %w", err)
	}

	// Return the attendance document
	return attendance, nil
}
