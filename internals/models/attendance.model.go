package models

import (
	"time"

	"github.com/gofrs/uuid"
)

type AttendanceStatus string

const (
	StatusPresent AttendanceStatus = "present"
	StatusAbsent  AttendanceStatus = "absent"
	StatusLeave   AttendanceStatus = "leave"
	StatusLate    AttendanceStatus = "late"
	StatusHalfDay AttendanceStatus = "half_day"
)

type Attendance struct {
	ID           uuid.UUID        `json:"id" db:"id"`
	EmployeeID   uuid.UUID        `json:"employee_id" db:"employee_id"`
	WorkDate     time.Time        `json:"work_date" db:"work_date"`
	CheckInTime  *time.Time       `json:"check_in_time,omitempty" db:"check_in_time"`
	CheckOutTime *time.Time       `json:"check_out_time,omitempty" db:"check_out_time"`
	NeedReview   bool             `json:"need_review" db:"need_review"`
	Status       AttendanceStatus `json:"status" db:"status"`
	CreatedAt    time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at" db:"updated_at"`
}

type CheckInOutAttendanceType struct {
	EmployeeID string `json:"employee_id" db:"employee_id"`
}

type CurrentAttendanceStats struct {
	TotalEmployees   int `json:"total_employees"`
	PresentEmployees int `json:"present_employees"`
	AbsentEmployees  int `json:"absent_employees"`
	LeaveEmployees   int `json:"leave_employees"`
}

type AttendanceData struct {
	Attendance    *Attendance `json:"attendance"`
	EmployeeId    uuid.UUID   `json:"employee_id"`
	EmployeeName  string      `json:"employee_name"`
	EmployeeEmail string      `json:"employee_email"`
	EmployeeRole  Role        `json:"employee_role"`
	EmployeeImage *string     `json:"employee_image,omitempty"`
	EmployeePhone string      `json:"employee_phone"`
}

type CurrentAttendance struct {
	Stats     CurrentAttendanceStats `json:"stats"`
	Employees []AttendanceData       `json:"employees"`
}

type AttendanceUpdate struct {
	AttendanceID string            `json:"attendance_id"`
	CheckInTime  *string           `json:"check_in_time"`
	CheckOutTime *string           `json:"check_out_time"`
	NeedReview   *bool             `json:"need_review"`
	Status       *AttendanceStatus `json:"status"`
}

// AttendanceHistoryData represents a single attendance record with employee details
type AttendanceHistoryData struct {
	ID            uuid.UUID        `json:"id"`
	EmployeeID    uuid.UUID        `json:"employee_id"`
	EmployeeName  string           `json:"employee_name"`
	EmployeeEmail string           `json:"employee_email"`
	EmployeeRole  Role             `json:"employee_role"`
	EmployeeImage *string          `json:"employee_image,omitempty"`
	EmployeePhone string           `json:"employee_phone"`
	WorkDate      time.Time        `json:"work_date"`
	CheckInTime   *time.Time       `json:"check_in_time,omitempty"`
	CheckOutTime  *time.Time       `json:"check_out_time,omitempty"`
	NeedReview    bool             `json:"need_review"`
	Status        AttendanceStatus `json:"status"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
}

// AttendanceHistoryStats represents summary statistics for attendance history
type AttendanceHistoryStats struct {
	TotalRecords     int `json:"total_records"`
	PresentCount     int `json:"present_count"`
	AbsentCount      int `json:"absent_count"`
	LeaveCount       int `json:"leave_count"`
	LateCount        int `json:"late_count"`
	HalfDayCount     int `json:"half_day_count"`
	NeedsReviewCount int `json:"needs_review_count"`
}

// AttendanceHistoryResponse matches the pagination pattern
type AttendanceHistoryResponse struct {
	Data     []AttendanceHistoryData `json:"data"`
	Total    int                     `json:"total"`
	Page     int                     `json:"page"`
	Limit    int                     `json:"limit"`
	HasMore  bool                    `json:"hasMore"`
	NextPage int                     `json:"nextPage"`
	Stats    *AttendanceHistoryStats `json:"stats,omitempty"`
}

type AttendanceHistoryQuery struct {
	EmployeeId *uuid.UUID `json:"employee_id"`
	FromDate   *time.Time `json:"fromDate"`
	ToDate     *time.Time `json:"toDate"`
	Limit      int        `json:"limit"`
	Page       int        `json:"page"`
}

type LeaveStatus string

const (
	LeavePending  LeaveStatus = "pending"
	LeaveApproved LeaveStatus = "approved"
	LeaveRejected LeaveStatus = "rejected"
)

type AttendanceLeave struct {
	ID                uuid.UUID   `json:"id"`
	EmployeeID        uuid.UUID   `json:"employee_id"`
	CheckedBy         *uuid.UUID  `json:"checked_by,omitempty"`
	StartDate         time.Time   `json:"start_date"`
	EndDate           time.Time   `json:"end_date"`
	Message           string      `json:"message"`
	SupervisorMessage *string     `json:"supervisor_message,omitempty"`
	Status            LeaveStatus `json:"status"`
	CreatedAt         time.Time   `json:"created_at"`
	UpdatedAt         time.Time   `json:"updated_at"`
}

type CreateAttendanceLeave struct {
	EmployeeID uuid.UUID `json:"employee_id"`
	StartDate  time.Time `json:"start_date"`
	EndDate    time.Time `json:"end_date"`
	Message    string    `json:"message"`
}

type UserUpdateAttendanceLeave struct {
	Id         uuid.UUID `json:"id"`
	EmployeeID uuid.UUID `json:"employee_id"`
	StartDate  time.Time `json:"start_date"`
	EndDate    time.Time `json:"end_date"`
	Message    string    `json:"message"`
}

type UserCreateAttendanceLeave struct {
	EmployeeID string    `json:"employee_id"`
	StartDate  time.Time `json:"start_date"`
	EndDate    time.Time `json:"end_date"`
	Message    string    `json:"message"`
}

type UpdateAttendanceLeave struct {
	ID                uuid.UUID   `json:"id"`
	EmployeeID        uuid.UUID   `json:"employee_id"`
	CheckedBy         *uuid.UUID  `json:"checked_by,omitempty"`
	StartDate         time.Time   `json:"start_date"`
	EndDate           time.Time   `json:"end_date"`
	Message           string      `json:"message"`
	SupervisorMessage *string     `json:"supervisor_message,omitempty"`
	Status            LeaveStatus `json:"status"`
}

type AttendanceLeaveResponse struct {
	ID                uuid.UUID   `json:"id"`
	EmployeeID        uuid.UUID   `json:"employee_id"`
	EmployeeName      string      `json:"employee_name"`
	EmployeeEmail     string      `json:"employee_email"`
	EmployeeImage     *string     `json:"employee_image"`
	CheckedBy         *uuid.UUID  `json:"checked_by,omitempty"`
	StartDate         time.Time   `json:"start_date"`
	EndDate           time.Time   `json:"end_date"`
	Message           string      `json:"message"`
	SupervisorMessage *string     `json:"supervisor_message,omitempty"`
	Status            LeaveStatus `json:"status"`
	CreatedAt         time.Time   `json:"created_at"`
	UpdatedAt         time.Time   `json:"updated_at"`
}
