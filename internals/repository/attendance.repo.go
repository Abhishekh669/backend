package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Abhishekh669/backend/internals/database"
	"github.com/Abhishekh669/backend/internals/models"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AttendanceRepo interface {
	GetCurrentAttendance(c context.Context) (*models.CurrentAttendance, error)
	CheckInAttendance(c context.Context, req *models.CheckInOutAttendanceType) error
	CheckOutAttendance(c context.Context, req *models.CheckInOutAttendanceType) error
}

type attendanceRepo struct {
	pool *pgxpool.Pool
}

func (r *attendanceRepo) GetCurrentAttendance(ctx context.Context) (*models.CurrentAttendance, error) {
	query := `
		SELECT 
			e.id,
			e.name,
			e.email,
			e.role,
			e.image,
			e.phone,

			a.id,
			a.employee_id,
			a.work_date,
			a.check_in_time,
			a.check_out_time,
			a.need_review,
			a.status,
			a.created_at,
			a.updated_at

		FROM users e
		LEFT JOIN attendance a
			ON e.id = a.employee_id
			AND a.work_date = CURRENT_DATE
		WHERE e.is_active = true
		ORDER BY e.name;
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query current attendance: %w", err)
	}
	defer rows.Close()

	var employees []models.AttendanceData
	stats := models.CurrentAttendanceStats{}

	for rows.Next() {
		var data models.AttendanceData

		// Employee fields
		var employeeImage *string

		// Attendance nullable fields
		var attendanceID *uuid.UUID
		var attEmployeeID *uuid.UUID
		var workDate *time.Time
		var checkIn *time.Time
		var checkOut *time.Time
		var needReview *bool
		var status *models.AttendanceStatus
		var createdAt *time.Time
		var updatedAt *time.Time

		err := rows.Scan(
			&data.EmployeeId,
			&data.EmployeeName,
			&data.EmployeeEmail,
			&data.EmployeeRole,
			&employeeImage,
			&data.EmployeePhone,

			&attendanceID,
			&attEmployeeID,
			&workDate,
			&checkIn,
			&checkOut,
			&needReview,
			&status,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		data.EmployeeImage = employeeImage
		stats.TotalEmployees++

		// If attendance exists
		if attendanceID != nil {
			// Validate required fields are not nil
			if attEmployeeID == nil || workDate == nil || needReview == nil || status == nil || createdAt == nil || updatedAt == nil {
				log.Printf("Incomplete attendance data for employee %s, skipping row", data.EmployeeId)
				data.Attendance = nil
				stats.AbsentEmployees++
				employees = append(employees, data)
				continue
			}

			data.Attendance = &models.Attendance{
				ID:           *attendanceID,
				EmployeeID:   *attEmployeeID,
				WorkDate:     *workDate,
				CheckInTime:  checkIn,  // Can be nil
				CheckOutTime: checkOut, // Can be nil
				NeedReview:   *needReview,
				Status:       *status,
				CreatedAt:    *createdAt,
				UpdatedAt:    *updatedAt,
			}

			// Stats calculation
			switch *status {
			case models.StatusPresent, models.StatusLate, models.StatusHalfDay:
				stats.PresentEmployees++
			case models.StatusLeave:
				stats.LeaveEmployees++
			default:
				// For any other status (like Absent, etc), count as absent
				stats.AbsentEmployees++
			}
		} else {
			// No attendance record = Absent
			data.Attendance = nil
			stats.AbsentEmployees++
		}

		employees = append(employees, data)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	// Verify stats add up
	if stats.TotalEmployees != (stats.PresentEmployees + stats.AbsentEmployees + stats.LeaveEmployees) {
		// Log warning but don't fail
		log.Printf("Warning: Stats mismatch - Total: %d, Present: %d, Absent: %d, Leave: %d",
			stats.TotalEmployees, stats.PresentEmployees, stats.AbsentEmployees, stats.LeaveEmployees)
	}

	return &models.CurrentAttendance{
		Stats:     stats,
		Employees: employees,
	}, nil
}

func (r *attendanceRepo) CheckInAttendance(c context.Context, req *models.CheckInOutAttendanceType) error {
	if req == nil {
		return errors.New("request cannot be nil")
	}

	employeeUUID, err := uuid.FromString(req.EmployeeID)
	if err != nil {
		return errors.New("invalid employee_id")
	}

	query := `
		INSERT INTO attendance (
			employee_id,
			work_date,
			check_in_time,
			status
		)
		VALUES ($1, CURRENT_DATE, NOW(), 'present')
		RETURNING id
	`

	var id uuid.UUID
	err = r.pool.QueryRow(c, query, employeeUUID).Scan(&id)
	if err != nil {

		// PostgreSQL specific error handling
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {

			// 23505 = unique_violation
			if pgErr.Code == "23505" {
				return errors.New("attendance already marked for today")
			}

			// 23503 = foreign_key_violation
			if pgErr.Code == "23503" {
				return errors.New("employee does not exist")
			}
		}

		return err
	}

	return nil
}

func (r *attendanceRepo) CheckOutAttendance(c context.Context, req *models.CheckInOutAttendanceType) error {
	if req == nil {
		return errors.New("request cannot be nil")
	}

	employeeUUID, err := uuid.FromString(req.EmployeeID)
	if err != nil {
		return errors.New("invalid employee_id")
	}

	query := `
		UPDATE attendance
		SET check_out_time = NOW()
		WHERE employee_id = $1
		AND work_date = CURRENT_DATE
		AND check_out_time IS NULL
		RETURNING id
	`

	var id uuid.UUID
	err = r.pool.QueryRow(c, query, employeeUUID).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("no active check-in found or already checked out")
		}
		return err
	}

	return nil
}

func NewAttendanceRepository() AttendanceRepo {
	pool, err := database.GetPostgresPool()

	if err != nil {
		return nil
	}

	return &attendanceRepo{
		pool: pool,
	}
}
