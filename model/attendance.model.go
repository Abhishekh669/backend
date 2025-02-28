package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AttendanceRecord struct {
	UserID     string    `bson:"userId" json:"userId"`
	IsPresent  bool      `bson:"isPresent" json:"isPresent"`
	Reason     string    `bson:"reason,omitempty" json:"reason,omitempty"`
	RecordedAt time.Time `bson:"recordedAt" json:"recordedAt"`
}

type AttendanceTracker struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	Attendance  []AttendanceRecord `bson:"attendance" json:"attendance"`
	Date        time.Time          `bson:"date" json:"date"`
	Field       string             `bson:"field" json:"field"`
	SubmittedBy string             `bson:"submittedBy" json:"submittedBy"`
	CreatedAt   time.Time          `bson:"createdAt,omitempty" json:"createdAt"`
}

type AttendanceTrackerRequest struct {
	Attendance  []AttendanceRecord `bson:"attendance" json:"attendance"`
	Date        time.Time          `bson:"date" json:"date"`
	Field       string             `bson:"field" json:"field"`
	SubmittedBy string             `bson:"submittedBy" json:"submittedBy"`
	CreatedAt   time.Time          `bson:"createdAt,omitempty" json:"createdAt"`
}

type UpdateAttendanceTracker struct {
	ID          string             `bson:"_id,omitempty" json:"_id"`
	Attendance  []AttendanceRecord `bson:"attendance" json:"attendance"`
	Date        string             `bson:"date" json:"date"`
	Field       string             `bson:"field" json:"field"`
	SubmittedBy string             `bson:"submittedBy" json:"submittedBy"`
	CreatedAt   time.Time          `bson:"createdAt,omitempty" json:"createdAt"`
}

//we will get all the student and
//first create  attendance  when  the admi submit the attendance
