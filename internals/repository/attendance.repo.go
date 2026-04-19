package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Abhishekh669/backend/internals/database"
	"github.com/Abhishekh669/backend/internals/models"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AttendanceRepo interface {
	GetTodayUserAttendanceByUserId(c context.Context, empUUID uuid.UUID) (*models.Attendance, error)
	CreateDailyAbsentAttendance(ctx context.Context) error
	GetAllAttendanceLeaveRequestsHistory(c context.Context, limit, page int, fromDate *time.Time, toDate *time.Time, status *models.LeaveStatus) (*AttendanceLeaveByUserResponse, error)
	AcceptLeaveRequestByAdmin(c context.Context, id uuid.UUID, checkedBy uuid.UUID) (*models.AttendanceLeaveResponse, error)
	GetAllAttendanceLeaveRequest(c context.Context) ([]models.AttendanceLeaveResponse, error)
	UpdateCustomerLeave(c context.Context, req *models.UserUpdateAttendanceLeave) error
	GetTodayAttendanceLeave(c context.Context, empUUID uuid.UUID) (*models.AttendanceLeaveResponse, error)
	NewUpdateLeaveRequest(c context.Context, req *models.UpdateAttendanceLeave) (*models.AttendanceLeave, error)
	GetAttendanceLeaveRequestByUserId(c context.Context, empUUID uuid.UUID, limit, page int, fromDate *time.Time, toDate *time.Time, status *models.LeaveStatus) (*AttendanceLeaveByUserResponse, error)
	GetAttendanceRequest(c context.Context) ([]models.AttendanceLeaveResponse, error)
	CancelLeaveRequestByAdmin(c context.Context, leaveId *uuid.UUID, checkedBy uuid.UUID) (*models.AttendanceLeaveResponse, error)
	CancelLeaveRequest(c context.Context, leaveId *uuid.UUID, requesterId *uuid.UUID) error
	DeleteLeaveRequest(c context.Context, leaveId *uuid.UUID) error
	UpdateLeaveRequest(c context.Context, req *models.UpdateAttendanceLeave) error
	CreateEmployeeRequest(c context.Context, req *models.CreateAttendanceLeave) error
	DeleteInactiveLeaveRequestAttendance(ctx context.Context) error
	AutoReviewIncompleteAttendance(ctx context.Context) error
	GetAttendanceHistory(ctx context.Context, limit, page int, fromDate, toDate *time.Time, employeeID *uuid.UUID) (*models.AttendanceHistoryResponse, error)
	DeleteAttendanceById(ctx context.Context, attendanceID string) error
	UpdateAttendance(ctx context.Context, req *models.AttendanceUpdate) error
	GetCurrentAttendance(c context.Context) (*models.CurrentAttendance, error)
	CheckInAttendance(c context.Context, req *models.CheckInOutAttendanceType) error
	CheckOutAttendance(c context.Context, req *models.CheckInOutAttendanceType) error
}

type attendanceRepo struct {
	pool *pgxpool.Pool
}

type AttendanceLeaveByUserStats struct {
	TotalRequests    int `json:"total_requests"`
	PendingRequests  int `json:"pending_requests"`
	ApprovedRequests int `json:"approved_requests"`
	RejectedRequests int `json:"rejected_requests"`
}

type AttendanceLeaveByUserResponse struct {
	Requests   []models.AttendanceLeaveResponse `json:"requests"`
	Total      int                              `json:"total"`
	HasMore    bool                             `json:"has_more"`
	NextOffset int                              `json:"next_offset"`
	Stats      *AttendanceLeaveByUserStats      `json:"stats"`
}

func (r *attendanceRepo) CreateDailyAbsentAttendance(ctx context.Context) error {

	query := `
		INSERT INTO attendance (
			employee_id,
			work_date,
			check_in_time,
			check_out_time,
			status
		)
		SELECT 
			u.id,
			(NOW() AT TIME ZONE 'Asia/Kathmandu')::date,
			NULL,
			NULL,
			'absent'
		FROM users u
		WHERE 
			u.is_active = TRUE
			AND u.role != 'customer'

			-- ❗ Skip employees who are on approved leave today
			AND NOT EXISTS (
				SELECT 1 
				FROM attendance_leave al
				WHERE 
					al.employee_id = u.id
					AND al.status = 'approved'
					AND (NOW() AT TIME ZONE 'Asia/Kathmandu') BETWEEN al.start_date AND al.end_date
			)

		ON CONFLICT (employee_id, work_date) DO NOTHING;
	`

	_, err := r.pool.Exec(ctx, query)
	return err
}
func (r *attendanceRepo) GetTodayUserAttendanceByUserId(
	c context.Context,
	empUUID uuid.UUID,
) (*models.Attendance, error) {

	query := `
		SELECT 
			id,
			employee_id,
			work_date,
			check_in_time,
			check_out_time,
			need_review,
			status,
			created_at,
			updated_at
		FROM attendance
		WHERE employee_id = $1
		AND work_date = CURRENT_DATE
		LIMIT 1;
	`

	row := r.pool.QueryRow(c, query, empUUID)

	var att models.Attendance

	err := row.Scan(
		&att.ID,
		&att.EmployeeID,
		&att.WorkDate,
		&att.CheckInTime,
		&att.CheckOutTime,
		&att.NeedReview,
		&att.Status,
		&att.CreatedAt,
		&att.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {

			loc, _ := time.LoadLocation("Asia/Kathmandu")
			now := time.Now().In(loc)

			// 🔥 Return dummy attendance
			return &models.Attendance{
				EmployeeID:   empUUID,
				WorkDate:     time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc),
				CheckInTime:  nil,
				CheckOutTime: nil,
				NeedReview:   false,
				Status:       models.StatusAbsent, // ✅ FIXED (not present)
				CreatedAt:    now,
				UpdatedAt:    now,
			}, nil
		}
		return nil, err
	}

	return &att, nil
}
func (r *attendanceRepo) GetAllAttendanceLeaveRequestsHistory(c context.Context, limit, page int, fromDate *time.Time, toDate *time.Time, status *models.LeaveStatus) (*AttendanceLeaveByUserResponse, error) {
	errMessage := "failed to get attendance leave requests"
	offset := page * limit

	// ── Normalize optional filters to interface{} so pgx sends true NULL ──
	var fromVal, toVal, statusVal interface{}
	if fromDate != nil {
		fromVal = *fromDate
	}
	if toDate != nil {
		toVal = *toDate
	}
	if status != nil {
		statusVal = *status
	}

	// ── List query ────────────────────────────────────────────────────────
	rows, err := r.pool.Query(c, `
        SELECT
            al.id, al.employee_id, u.name, u.email, u.image,
            al.checked_by, al.start_date, al.end_date,
            al.message, al.supervisor_message, al.status,
            al.created_at, al.updated_at
        FROM attendance_leave al
        INNER JOIN users u ON u.id = al.employee_id
        WHERE
            ($1::timestamptz IS NULL OR al.created_at >= $1::timestamptz)
            AND ($2::timestamptz IS NULL OR al.created_at <= $2::timestamptz)
            AND ($3::leave_status IS NULL OR al.status    = $3::leave_status)
        ORDER BY al.created_at DESC
        LIMIT $4 OFFSET $5
    `, fromVal, toVal, statusVal, limit, offset)
	if err != nil {
		log.Printf("error getting leave requests: %v", err)
		return nil, errors.New(errMessage)
	}
	defer rows.Close()

	requests := make([]models.AttendanceLeaveResponse, 0, limit)
	for rows.Next() {
		var leave models.AttendanceLeaveResponse
		if err := rows.Scan(
			&leave.ID,
			&leave.EmployeeID,
			&leave.EmployeeName,
			&leave.EmployeeEmail,
			&leave.EmployeeImage,
			&leave.CheckedBy,
			&leave.StartDate,
			&leave.EndDate,
			&leave.Message,
			&leave.SupervisorMessage,
			&leave.Status,
			&leave.CreatedAt,
			&leave.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("%s: %w", errMessage, err)
		}
		requests = append(requests, leave)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.New(errMessage)
	}

	// ── Count query ───────────────────────────────────────────────────────
	var total int
	err = r.pool.QueryRow(c, `
        SELECT COUNT(*)
        FROM attendance_leave
        WHERE
            ($1::timestamptz IS NULL OR created_at >= $1::timestamptz)
            AND ($2::timestamptz IS NULL OR created_at <= $2::timestamptz)
            AND ($3::leave_status IS NULL OR status    = $3::leave_status)
    `, fromVal, toVal, statusVal).Scan(&total)
	if err != nil {
		return nil, errors.New(errMessage)
	}

	// ── Stats query (date filtered, all statuses) ─────────────────────────
	var stats AttendanceLeaveByUserStats
	err = r.pool.QueryRow(c, `
        SELECT
            COUNT(*)                                         AS total_requests,
            COUNT(*) FILTER (WHERE status = 'pending')      AS pending_requests,
            COUNT(*) FILTER (WHERE status = 'approved')     AS approved_requests,
            COUNT(*) FILTER (WHERE status = 'rejected')     AS rejected_requests
        FROM attendance_leave
        WHERE
            ($1::timestamptz IS NULL OR created_at >= $1::timestamptz)
            AND ($2::timestamptz IS NULL OR created_at <= $2::timestamptz)
    `, fromVal, toVal).Scan(
		&stats.TotalRequests,
		&stats.PendingRequests,
		&stats.ApprovedRequests,
		&stats.RejectedRequests,
	)
	if err != nil {
		return nil, errors.New(errMessage)
	}

	hasMore := (page+1)*limit < total
	nextOffset := page + 1

	return &AttendanceLeaveByUserResponse{
		Requests:   requests,
		Total:      total,
		HasMore:    hasMore,
		NextOffset: nextOffset,
		Stats:      &stats,
	}, nil
}

func (r *attendanceRepo) GetAllAttendanceLeaveRequest(c context.Context) ([]models.AttendanceLeaveResponse, error) {
	query := `
		SELECT 
			al.id,
			al.employee_id,
			u.name AS employee_name,
			u.email AS employee_email,
			u.image AS employee_image,
			al.checked_by,
			al.start_date,
			al.end_date,
			al.message,
			al.supervisor_message,
			al.status,
			al.created_at,
			al.updated_at
		FROM attendance_leave al
		INNER JOIN users u ON u.id = al.employee_id
		WHERE al.status = 'pending'
		ORDER BY al.created_at DESC
	`

	rows, err := r.pool.Query(c, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query attendance leave requests: %w", err)
	}
	defer rows.Close()

	var leaves []models.AttendanceLeaveResponse

	for rows.Next() {
		var leave models.AttendanceLeaveResponse
		err := rows.Scan(
			&leave.ID,
			&leave.EmployeeID,
			&leave.EmployeeName,
			&leave.EmployeeEmail,
			&leave.EmployeeImage,
			&leave.CheckedBy,
			&leave.StartDate,
			&leave.EndDate,
			&leave.Message,
			&leave.SupervisorMessage,
			&leave.Status,
			&leave.CreatedAt,
			&leave.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan attendance leave row: %w", err)
		}
		leaves = append(leaves, leave)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return leaves, nil
}

func (r *attendanceRepo) UpdateCustomerLeave(c context.Context, req *models.UserUpdateAttendanceLeave) error {
	query := `
		UPDATE attendance_leave
		SET
			start_date = $1,
			end_date   = $2,
			message    = $3,
			updated_at = NOW()
		WHERE id = $4
		  AND employee_id = $5
		  AND status = 'pending'
	`

	result, err := r.pool.Exec(c, query,
		req.StartDate,
		req.EndDate,
		req.Message,
		req.Id,
		req.EmployeeID,
	)
	if err != nil {
		return fmt.Errorf("failed to update attendance leave: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("leave request not found, already processed, or does not belong to employee")
	}

	return nil
}

func (r *attendanceRepo) GetTodayAttendanceLeave(c context.Context, empUUID uuid.UUID) (*models.AttendanceLeaveResponse, error) {
	errMessage := "failed to get today's attendance leave"
	var leave models.AttendanceLeaveResponse
	err := r.pool.QueryRow(c, `
        SELECT
            al.id,
            al.employee_id,
            u.name,
            u.email,
            u.image,
            al.checked_by,
            al.start_date,
            al.end_date,
            al.message,
            al.supervisor_message,
            al.status,
            al.created_at,
            al.updated_at
        FROM attendance_leave al
        INNER JOIN users u ON u.id = al.employee_id
        WHERE
            al.employee_id = $1
            AND (
                -- approved: return if today is within the leave range
                (
                    al.status = 'approved'
                    AND (al.start_date AT TIME ZONE 'Asia/Kathmandu')::DATE <= (NOW() AT TIME ZONE 'Asia/Kathmandu')::DATE
                    AND (al.end_date   AT TIME ZONE 'Asia/Kathmandu')::DATE >= (NOW() AT TIME ZONE 'Asia/Kathmandu')::DATE
                )
                OR
                -- pending: return if today is within the leave range
                (
                    al.status = 'pending'
                    AND (al.start_date AT TIME ZONE 'Asia/Kathmandu')::DATE <= (NOW() AT TIME ZONE 'Asia/Kathmandu')::DATE
                    AND (al.end_date   AT TIME ZONE 'Asia/Kathmandu')::DATE >= (NOW() AT TIME ZONE 'Asia/Kathmandu')::DATE
                )
                OR
                -- rejected: only return if BOTH updated today AND start_date is today
                -- so next day they are free to request even if range was multi-day
                (
                    al.status = 'rejected'
                    AND (al.updated_at  AT TIME ZONE 'Asia/Kathmandu')::DATE = (NOW() AT TIME ZONE 'Asia/Kathmandu')::DATE
                    AND (al.start_date  AT TIME ZONE 'Asia/Kathmandu')::DATE = (NOW() AT TIME ZONE 'Asia/Kathmandu')::DATE
                )
            )
        ORDER BY al.created_at DESC
        LIMIT 1
    `, empUUID).Scan(
		&leave.ID,
		&leave.EmployeeID,
		&leave.EmployeeName,
		&leave.EmployeeEmail,
		&leave.EmployeeImage,
		&leave.CheckedBy,
		&leave.StartDate,
		&leave.EndDate,
		&leave.Message,
		&leave.SupervisorMessage,
		&leave.Status,
		&leave.CreatedAt,
		&leave.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		log.Printf("error getting today's attendance leave: %v", err)
		return nil, errors.New(errMessage)
	}
	return &leave, nil
}

func (r *attendanceRepo) GetAttendanceLeaveRequestByUserId(c context.Context, empUUID uuid.UUID, limit, page int, fromDate, toDate *time.Time, status *models.LeaveStatus) (*AttendanceLeaveByUserResponse, error) {
	const errMessage = "failed to get attendance leave requests"

	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if page < 0 {
		page = 0
	}

	offset := page * limit

	// ── Treat empty string status as nil ──────────────────────────────────
	if status != nil && string(*status) == "" {
		status = nil
	}

	// ── Check employee exists ─────────────────────────────────────────────
	var exists bool
	err := r.pool.QueryRow(c, `
		SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)
	`, empUUID).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to verify employee: %w", errMessage, err)
	}
	if !exists {
		return nil, fmt.Errorf("employee not found with ID: %s", empUUID.String())
	}

	// ── List query ────────────────────────────────────────────────────────
	query := `
		SELECT
			al.id,
			al.employee_id,
			u.name,
			u.email,
			u.image,
			al.checked_by,
			al.start_date,
			al.end_date,
			al.message,
			al.supervisor_message,
			al.status,
			al.created_at,
			al.updated_at
		FROM attendance_leave al
		INNER JOIN users u ON u.id = al.employee_id
		WHERE al.employee_id = $1
	`
	args := []interface{}{empUUID}
	argPos := 2

	if fromDate != nil {
		query += fmt.Sprintf(" AND al.created_at >= $%d::timestamptz", argPos)
		args = append(args, *fromDate)
		argPos++
	}
	if toDate != nil {
		query += fmt.Sprintf(" AND al.created_at <= $%d::timestamptz", argPos)
		args = append(args, *toDate)
		argPos++
	}
	if status != nil {
		query += fmt.Sprintf(" AND al.status = $%d::text::leave_status", argPos)
		args = append(args, string(*status))
		argPos++
	}

	query += fmt.Sprintf(" ORDER BY al.created_at DESC LIMIT $%d OFFSET $%d", argPos, argPos+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(c, query, args...)
	if err != nil {
		log.Printf("[ERROR] Failed to execute leave requests query: %v", err)
		return nil, fmt.Errorf("%s: %w", errMessage, err)
	}
	defer rows.Close()

	requests := make([]models.AttendanceLeaveResponse, 0, limit)
	for rows.Next() {
		var leave models.AttendanceLeaveResponse
		if err := rows.Scan(
			&leave.ID,
			&leave.EmployeeID,
			&leave.EmployeeName,
			&leave.EmployeeEmail,
			&leave.EmployeeImage,
			&leave.CheckedBy,
			&leave.StartDate,
			&leave.EndDate,
			&leave.Message,
			&leave.SupervisorMessage,
			&leave.Status,
			&leave.CreatedAt,
			&leave.UpdatedAt,
		); err != nil {
			log.Printf("[ERROR] Failed to scan row: %v", err)
			return nil, fmt.Errorf("%s: failed to scan row: %w", errMessage, err)
		}
		requests = append(requests, leave)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", errMessage, err)
	}

	// ── Count query ───────────────────────────────────────────────────────
	countQuery := `
		SELECT COUNT(*)
		FROM attendance_leave
		WHERE employee_id = $1
	`
	countArgs := []interface{}{empUUID}
	countArgPos := 2

	if fromDate != nil {
		countQuery += fmt.Sprintf(" AND created_at >= $%d::timestamptz", countArgPos)
		countArgs = append(countArgs, *fromDate)
		countArgPos++
	}
	if toDate != nil {
		countQuery += fmt.Sprintf(" AND created_at <= $%d::timestamptz", countArgPos)
		countArgs = append(countArgs, *toDate)
		countArgPos++
	}
	if status != nil {
		countQuery += fmt.Sprintf(" AND status = $%d::text::leave_status", countArgPos)
		countArgs = append(countArgs, string(*status))
		countArgPos++
	}

	var total int
	err = r.pool.QueryRow(c, countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get total count: %w", errMessage, err)
	}

	// ── Stats query (date filtered, all statuses) ─────────────────────────
	statsQuery := `
		SELECT
			COUNT(*)                                         AS total_requests,
			COUNT(*) FILTER (WHERE status = 'pending')      AS pending_requests,
			COUNT(*) FILTER (WHERE status = 'approved')     AS approved_requests,
			COUNT(*) FILTER (WHERE status = 'rejected')     AS rejected_requests
		FROM attendance_leave
		WHERE employee_id = $1
	`
	statsArgs := []interface{}{empUUID}
	statsArgPos := 2

	if fromDate != nil {
		statsQuery += fmt.Sprintf(" AND created_at >= $%d::timestamptz", statsArgPos)
		statsArgs = append(statsArgs, *fromDate)
		statsArgPos++
	}
	if toDate != nil {
		statsQuery += fmt.Sprintf(" AND created_at <= $%d::timestamptz", statsArgPos)
		statsArgs = append(statsArgs, *toDate)
		statsArgPos++
	}

	var stats AttendanceLeaveByUserStats
	err = r.pool.QueryRow(c, statsQuery, statsArgs...).Scan(
		&stats.TotalRequests,
		&stats.PendingRequests,
		&stats.ApprovedRequests,
		&stats.RejectedRequests,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get stats: %w", errMessage, err)
	}

	hasMore := offset+limit < total
	nextOffset := page + 1
	if !hasMore {
		nextOffset = -1
	}

	return &AttendanceLeaveByUserResponse{
		Requests:   requests,
		Total:      total,
		HasMore:    hasMore,
		NextOffset: nextOffset,
		Stats:      &stats,
	}, nil
}
func (r *attendanceRepo) GetAttendanceRequest(c context.Context) ([]models.AttendanceLeaveResponse, error) {
	var responses []models.AttendanceLeaveResponse

	query := `
        SELECT 
            al.id,
            al.employee_id,
            u.name as employee_name,
            u.email as employee_email,
            u.image as employee_image,
            al.checked_by,
            al.start_date,
            al.end_date,
            al.message,
            al.supervisor_message,
            al.status,
            al.created_at,
            al.updated_at
        FROM attendance_leave al
        LEFT JOIN users u ON al.employee_id = u.id
        ORDER BY al.created_at DESC
    `

	rows, err := r.pool.Query(c, query)
	if err != nil {
		return nil, fmt.Errorf("error querying attendance requests: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var resp models.AttendanceLeaveResponse
		var employeeImage *string
		var checkedBy *uuid.UUID
		var supervisorMessage *string

		err := rows.Scan(
			&resp.ID,
			&resp.EmployeeID,
			&resp.EmployeeName,
			&resp.EmployeeEmail,
			&employeeImage,
			&checkedBy,
			&resp.StartDate,
			&resp.EndDate,
			&resp.Message,
			&supervisorMessage,
			&resp.Status,
			&resp.CreatedAt,
			&resp.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning attendance request: %w", err)
		}

		// Assign nullable fields directly (pgx/v5 handles NULLs by setting pointers to nil)
		resp.EmployeeImage = employeeImage
		resp.CheckedBy = checkedBy
		resp.SupervisorMessage = supervisorMessage

		responses = append(responses, resp)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating attendance requests: %w", err)
	}

	return responses, nil
}

func (r *attendanceRepo) AcceptLeaveRequestByAdmin(c context.Context, id uuid.UUID, checkedBy uuid.UUID) (*models.AttendanceLeaveResponse, error) {
	// ── Begin transaction ──
	tx, err := r.pool.Begin(c)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(c) // no-op if already committed

	// ── Step 1: Accept the leave request ──
	acceptQuery := `
        UPDATE attendance_leave al
        SET
            status     = 'approved',
            checked_by = $2,
            updated_at = NOW()
        FROM users u
        WHERE al.id     = $1
          AND al.status = 'pending'
          AND u.id      = al.employee_id
        RETURNING
            al.id,
            al.employee_id,
            u.name  AS employee_name,
            u.email AS employee_email,
            u.image AS employee_image,
            al.checked_by,
            al.start_date,
            al.end_date,
            al.message,
            al.supervisor_message,
            al.status,
            al.created_at,
            al.updated_at
    `
	var res models.AttendanceLeaveResponse
	err = tx.QueryRow(c, acceptQuery, id, checkedBy).Scan(
		&res.ID,
		&res.EmployeeID,
		&res.EmployeeName,
		&res.EmployeeEmail,
		&res.EmployeeImage,
		&res.CheckedBy,
		&res.StartDate,
		&res.EndDate,
		&res.Message,
		&res.SupervisorMessage,
		&res.Status,
		&res.CreatedAt,
		&res.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("leave request not found or already processed")
		}
		return nil, fmt.Errorf("failed to accept leave request: %w", err)
	}

	// ── Step 2: Create attendance records for each day in the leave range ──
	// Normalize to date only (strip time component)
	startDate := res.StartDate.UTC().Truncate(24 * time.Hour)
	endDate := res.EndDate.UTC().Truncate(24 * time.Hour)

	attendanceQuery := `
        INSERT INTO attendance (
            employee_id,
            work_date,
            status,
            created_at,
            updated_at
        ) VALUES ($1, $2, 'leave', NOW(), NOW())
        ON CONFLICT (employee_id, work_date)
            DO UPDATE SET
                status     = 'leave',
                updated_at = NOW()
    `

	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		_, err = tx.Exec(c, attendanceQuery, res.EmployeeID, d)
		if err != nil {
			return nil, fmt.Errorf("failed to create attendance record for %s: %w", d.Format("2006-01-02"), err)
		}
	}

	// ── Step 3: Commit ──
	if err = tx.Commit(c); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &res, nil
}

func (r *attendanceRepo) CancelLeaveRequestByAdmin(c context.Context, leaveId *uuid.UUID, checkedBy uuid.UUID) (*models.AttendanceLeaveResponse, error) {
	if leaveId == nil {
		return nil, fmt.Errorf("leave ID is required")
	}
	query := `
        UPDATE attendance_leave al
        SET
            status     = 'rejected',
            checked_by = $2,
            updated_at = NOW()
        FROM users u
        WHERE al.id     = $1
          AND al.status = 'pending'
          AND u.id      = al.employee_id
        RETURNING
            al.id,
            al.employee_id,
            u.name  AS employee_name,
            u.email AS employee_email,
            u.image AS employee_image,
            al.checked_by,
            al.start_date,
            al.end_date,
            al.message,
            al.supervisor_message,
            al.status,
            al.created_at,
            al.updated_at
    `
	var res models.AttendanceLeaveResponse
	err := r.pool.QueryRow(c, query, leaveId, checkedBy).Scan(
		&res.ID,
		&res.EmployeeID,
		&res.EmployeeName,
		&res.EmployeeEmail,
		&res.EmployeeImage,
		&res.CheckedBy,
		&res.StartDate,
		&res.EndDate,
		&res.Message,
		&res.SupervisorMessage,
		&res.Status,
		&res.CreatedAt,
		&res.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("leave request not found or already processed")
		}
		return nil, fmt.Errorf("failed to cancel leave request: %w", err)
	}
	return &res, nil
}
func (r *attendanceRepo) CancelLeaveRequest(c context.Context, leaveId *uuid.UUID, requesterId *uuid.UUID) error {
	if leaveId == nil {
		return fmt.Errorf("leave ID is required")
	}
	if requesterId == nil {
		return fmt.Errorf("requester ID is required")
	}

	// Start transaction
	tx, err := r.pool.Begin(c)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(c)

	// First, get the leave request details
	var leave struct {
		EmployeeID uuid.UUID
		Status     models.LeaveStatus
		StartDate  time.Time
		EndDate    time.Time
	}

	err = tx.QueryRow(c, `
		SELECT employee_id, status, start_date, end_date 
		FROM attendance_leave 
		WHERE id = $1
	`, leaveId).Scan(
		&leave.EmployeeID,
		&leave.Status,
		&leave.StartDate,
		&leave.EndDate,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("leave request not found")
		}
		return fmt.Errorf("failed to fetch leave request: %w", err)
	}

	// Check if the requester is the employee who created the leave
	if leave.EmployeeID != *requesterId {
		return fmt.Errorf("unauthorized: you can only cancel your own leave requests")
	}

	// Check if the leave request is still pending
	if leave.Status != models.LeavePending {
		return fmt.Errorf("cannot cancel leave request with status: %s (only pending requests can be cancelled)", leave.Status)
	}

	// Delete the leave request (cancelling = deleting since it's pending)
	_, err = tx.Exec(c, `
		DELETE FROM attendance_leave 
		WHERE id = $1 AND status = $2
	`, leaveId, models.LeavePending)

	if err != nil {
		return fmt.Errorf("failed to cancel leave request: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(c); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *attendanceRepo) DeleteLeaveRequest(c context.Context, leaveId *uuid.UUID) error {
	if leaveId == nil {
		return fmt.Errorf("leave ID is required")
	}

	// Start transaction
	tx, err := r.pool.Begin(c)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(c)

	// First, get the leave request details to check if it was approved
	var leave struct {
		Status     models.LeaveStatus
		EmployeeID uuid.UUID
		StartDate  time.Time
		EndDate    time.Time
	}

	err = tx.QueryRow(c, `
		SELECT status, employee_id, start_date, end_date 
		FROM attendance_leave 
		WHERE id = $1
	`, leaveId).Scan(
		&leave.Status,
		&leave.EmployeeID,
		&leave.StartDate,
		&leave.EndDate,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("leave request not found")
		}
		return fmt.Errorf("failed to fetch leave request: %w", err)
	}

	// Delete the leave request
	result, err := tx.Exec(c, `
		DELETE FROM attendance_leave 
		WHERE id = $1
	`, leaveId)

	if err != nil {
		return fmt.Errorf("failed to delete leave request: %w", err)
	}

	// Check if any row was actually deleted
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("leave request not found")
	}

	// Commit transaction
	if err = tx.Commit(c); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
func (r *attendanceRepo) NewUpdateLeaveRequest(c context.Context, req *models.UpdateAttendanceLeave) (*models.AttendanceLeave, error) {
	if req.StartDate.After(req.EndDate) {
		return nil, fmt.Errorf("start date cannot be after end date")
	}

	tx, err := r.pool.Begin(c)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(c)

	// Check if leave request exists and fetch current state
	var currentLeave models.AttendanceLeave
	err = tx.QueryRow(c, `
		SELECT id, employee_id, checked_by, start_date, end_date, message, 
		       supervisor_message, status, created_at, updated_at
		FROM attendance_leave
		WHERE id = $1
	`, req.ID).Scan(
		&currentLeave.ID,
		&currentLeave.EmployeeID,
		&currentLeave.CheckedBy,
		&currentLeave.StartDate,
		&currentLeave.EndDate,
		&currentLeave.Message,
		&currentLeave.SupervisorMessage,
		&currentLeave.Status,
		&currentLeave.CreatedAt,
		&currentLeave.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("leave request not found")
		}
		return nil, fmt.Errorf("failed to fetch leave request: %w", err)
	}

	// Prevent updating an already finalized leave
	if currentLeave.Status == models.LeaveApproved || currentLeave.Status == models.LeaveRejected {
		return nil, fmt.Errorf("cannot update a leave request that has already been %s", currentLeave.Status)
	}

	// Check if employee exists
	var employeeExists bool
	err = tx.QueryRow(c, `
		SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)
	`, req.EmployeeID).Scan(&employeeExists)
	if err != nil {
		return nil, fmt.Errorf("failed to check employee existence: %w", err)
	}
	if !employeeExists {
		return nil, fmt.Errorf("employee not found")
	}

	// Check if checked_by user exists (if provided)
	if req.CheckedBy != nil {
		var checkerExists bool
		err = tx.QueryRow(c, `
			SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)
		`, req.CheckedBy).Scan(&checkerExists)
		if err != nil {
			return nil, fmt.Errorf("failed to check reviewer existence: %w", err)
		}
		if !checkerExists {
			return nil, fmt.Errorf("reviewer (checked_by) user not found")
		}
	}

	// Update the leave request
	var updated models.AttendanceLeave
	err = tx.QueryRow(c, `
		UPDATE attendance_leave
		SET
			checked_by         = $1,
			start_date         = $2,
			end_date           = $3,
			message            = $4,
			supervisor_message = $5,
			status             = $6,
			updated_at         = NOW()
		WHERE id = $7
		RETURNING id, employee_id, checked_by, start_date, end_date, message,
		          supervisor_message, status, created_at, updated_at
	`,
		req.CheckedBy,
		req.StartDate,
		req.EndDate,
		req.Message,
		req.SupervisorMessage,
		req.Status,
		req.ID,
	).Scan(
		&updated.ID,
		&updated.EmployeeID,
		&updated.CheckedBy,
		&updated.StartDate,
		&updated.EndDate,
		&updated.Message,
		&updated.SupervisorMessage,
		&updated.Status,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update leave request: %w", err)
	}

	// Only apply attendance changes when the leave is being approved
	if req.Status == models.LeaveApproved {
		today := time.Now().UTC().Truncate(24 * time.Hour)
		leaveStart := req.StartDate.UTC().Truncate(24 * time.Hour)
		leaveEnd := req.EndDate.UTC().Truncate(24 * time.Hour)

		// Check if today falls within the leave range
		todayInRange := !today.Before(leaveStart) && !today.After(leaveEnd)

		if todayInRange {
			// Check if today's attendance record already exists
			var todayAttendanceExists bool
			err = tx.QueryRow(c, `
				SELECT EXISTS(
					SELECT 1 FROM attendance
					WHERE employee_id = $1 AND work_date = $2
				)
			`, req.EmployeeID, today).Scan(&todayAttendanceExists)
			if err != nil {
				return nil, fmt.Errorf("failed to check today's attendance: %w", err)
			}

			if todayAttendanceExists {
				// Today already has a record — update it to half_day
				_, err = tx.Exec(c, `
					UPDATE attendance
					SET status     = $1,
					    need_review = FALSE,
					    updated_at = NOW()
					WHERE employee_id = $2 AND work_date = $3
				`, models.StatusHalfDay, req.EmployeeID, today)
				if err != nil {
					return nil, fmt.Errorf("failed to update today's attendance to half_day: %w", err)
				}
			} else {
				// No record yet for today — insert as half_day
				_, err = tx.Exec(c, `
					INSERT INTO attendance (employee_id, work_date, status, need_review, created_at, updated_at)
					VALUES ($1, $2, $3, FALSE, NOW(), NOW())
					ON CONFLICT (employee_id, work_date) DO UPDATE
					    SET status     = EXCLUDED.status,
					        updated_at = NOW()
				`, req.EmployeeID, today, models.StatusLeave)
				if err != nil {
					return nil, fmt.Errorf("failed to insert today's half_day attendance: %w", err)
				}
			}

			// Handle all OTHER days in the range (excluding today) as leave
			_, err = tx.Exec(c, `
				INSERT INTO attendance (employee_id, work_date, status, need_review, created_at, updated_at)
				SELECT 
					$1,
					generate_series::DATE,
					$2,
					FALSE,
					NOW(),
					NOW()
				FROM generate_series($3::DATE, $4::DATE, '1 day'::INTERVAL) generate_series
				WHERE generate_series::DATE != $5::DATE
				ON CONFLICT (employee_id, work_date) DO UPDATE
				    SET status     = EXCLUDED.status,
				        updated_at = NOW()
			`, req.EmployeeID, models.StatusLeave, leaveStart, leaveEnd, today)
			if err != nil {
				return nil, fmt.Errorf("failed to upsert leave attendance records: %w", err)
			}

		} else {
			// Today is NOT in the leave range — mark all days as leave
			_, err = tx.Exec(c, `
				INSERT INTO attendance (employee_id, work_date, status, need_review, created_at, updated_at)
				SELECT 
					$1,
					generate_series::DATE,
					$2,
					FALSE,
					NOW(),
					NOW()
				FROM generate_series($3::DATE, $4::DATE, '1 day'::INTERVAL) generate_series
				ON CONFLICT (employee_id, work_date) DO UPDATE
				    SET status     = EXCLUDED.status,
				        updated_at = NOW()
			`, req.EmployeeID, models.StatusLeave, leaveStart, leaveEnd)
			if err != nil {
				return nil, fmt.Errorf("failed to upsert leave attendance records: %w", err)
			}
		}
	}

	if err = tx.Commit(c); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &updated, nil
}

func (r *attendanceRepo) UpdateLeaveRequest(c context.Context, req *models.UpdateAttendanceLeave) error {
	// Basic validation
	if req.StartDate.After(req.EndDate) {
		return fmt.Errorf("start date cannot be after end date")
	}

	// Start transaction
	tx, err := r.pool.Begin(c)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(c)

	// Check if the leave request exists and get current details
	var currentLeave struct {
		EmployeeID uuid.UUID
		StartDate  time.Time
		EndDate    time.Time
		Status     models.LeaveStatus
	}

	err = tx.QueryRow(c, `
        SELECT employee_id, start_date, end_date, status 
        FROM attendance_leave 
        WHERE id = $1
    `, req.ID).Scan(
		&currentLeave.EmployeeID,
		&currentLeave.StartDate,
		&currentLeave.EndDate,
		&currentLeave.Status,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("leave request not found")
		}
		return fmt.Errorf("failed to fetch leave request: %w", err)
	}

	// Check if employee exists (if EmployeeID is being changed)
	if req.EmployeeID != currentLeave.EmployeeID {
		var exists bool
		err = tx.QueryRow(c, `
            SELECT EXISTS(
                SELECT 1 FROM users 
                WHERE id = $1 
            )
        `, req.EmployeeID).Scan(&exists)

		if err != nil {
			return fmt.Errorf("failed to check employee existence: %w", err)
		}

		if !exists {
			return fmt.Errorf("employee not found")
		}
	}

	// Check if checked_by user exists (if provided)
	if req.CheckedBy != nil {
		var checkerExists bool
		err = tx.QueryRow(c, `
            SELECT EXISTS(
                SELECT 1 FROM users 
                WHERE id = $1 
            )
        `, req.CheckedBy).Scan(&checkerExists)

		if err != nil {
			return fmt.Errorf("failed to check approver existence: %w", err)
		}

		if !checkerExists {
			return fmt.Errorf("approver not found")
		}
	}

	// Update the leave request
	_, err = tx.Exec(c, `
        UPDATE attendance_leave 
        SET 
            employee_id = $1,
            checked_by = $2,
            start_date = $3,
            end_date = $4,
            message = $5,
            supervisor_message = $6,
            status = $7,
            updated_at = NOW()
        WHERE id = $8
    `,
		req.EmployeeID,
		req.CheckedBy,
		req.StartDate,
		req.EndDate,
		req.Message,
		req.SupervisorMessage,
		req.Status,
		req.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update leave request: %w", err)
	}

	// Handle attendance records based on status changes
	if req.Status == models.LeaveApproved {
		// If approved, create/update attendance records for the date range

		// First, delete any existing attendance records for this date range
		// to avoid conflicts (only if dates or employee changed)
		if req.EmployeeID != currentLeave.EmployeeID ||
			!req.StartDate.Equal(currentLeave.StartDate) ||
			!req.EndDate.Equal(currentLeave.EndDate) {

			_, err = tx.Exec(c, `
				DELETE FROM attendance 
				WHERE employee_id = $1 
				AND work_date BETWEEN $2 AND $3
			`, currentLeave.EmployeeID, currentLeave.StartDate, currentLeave.EndDate)

			if err != nil {
				return fmt.Errorf("failed to delete old attendance records: %w", err)
			}
		}

		// Generate all dates between start_date and end_date
		currentDate := req.StartDate
		for !currentDate.After(req.EndDate) {
			// Insert attendance record for each date with 'leave' status
			_, err = tx.Exec(c, `
				INSERT INTO attendance (
					employee_id,
					work_date,
					status,
					need_review,
					created_at,
					updated_at
				) VALUES ($1, $2, $3, $4, NOW(), NOW())
				ON CONFLICT (employee_id, work_date) 
				DO UPDATE SET 
					status = EXCLUDED.status,
					need_review = EXCLUDED.need_review,
					updated_at = NOW()
			`,
				req.EmployeeID,
				currentDate.Format("2006-01-02"), // Format as DATE
				models.StatusLeave,
				false, // need_review = false for approved leaves
			)

			if err != nil {
				return fmt.Errorf("failed to create attendance record for %s: %w",
					currentDate.Format("2006-01-02"), err)
			}

			// Move to next day
			currentDate = currentDate.AddDate(0, 0, 1)
		}
	} else if req.Status == models.LeaveRejected && currentLeave.Status == models.LeaveApproved {
		// If changing from approved to rejected, remove the attendance records
		_, err = tx.Exec(c, `
			DELETE FROM attendance 
			WHERE employee_id = $1 
			AND work_date BETWEEN $2 AND $3
		`, req.EmployeeID, req.StartDate, req.EndDate)

		if err != nil {
			return fmt.Errorf("failed to delete attendance records for rejected leave: %w", err)
		}
	}

	// Commit transaction
	if err = tx.Commit(c); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *attendanceRepo) CreateEmployeeRequest(c context.Context, req *models.CreateAttendanceLeave) error {
	if req.StartDate.After(req.EndDate) {
		return fmt.Errorf("start date cannot be after end date")
	}

	tx, err := r.pool.Begin(c)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(c)

	// ── Block check ──────────────────────────────────────────────────────────
	var blockReason string
	err = tx.QueryRow(c, `
        SELECT
            CASE
                -- pending + today in range → already has active pending request
                WHEN al.status = 'pending'
                     AND (al.start_date AT TIME ZONE 'Asia/Kathmandu')::DATE <= (NOW() AT TIME ZONE 'Asia/Kathmandu')::DATE
                     AND (al.end_date   AT TIME ZONE 'Asia/Kathmandu')::DATE >= (NOW() AT TIME ZONE 'Asia/Kathmandu')::DATE
                THEN 'you already have a pending leave request for today'

                -- approved + today in range → active leave, cannot request
                WHEN al.status = 'approved'
                     AND (al.start_date AT TIME ZONE 'Asia/Kathmandu')::DATE <= (NOW() AT TIME ZONE 'Asia/Kathmandu')::DATE
                     AND (al.end_date   AT TIME ZONE 'Asia/Kathmandu')::DATE >= (NOW() AT TIME ZONE 'Asia/Kathmandu')::DATE
                THEN 'you have an approved leave active until ' || (al.end_date AT TIME ZONE 'Asia/Kathmandu')::DATE::TEXT

                -- rejected + updated today + start_date was today → blocked same day
                -- if start_date was tomorrow or beyond, allow (falls into ELSE)
                WHEN al.status = 'rejected'
                     AND (al.updated_at  AT TIME ZONE 'Asia/Kathmandu')::DATE = (NOW() AT TIME ZONE 'Asia/Kathmandu')::DATE
                     AND (al.start_date  AT TIME ZONE 'Asia/Kathmandu')::DATE = (NOW() AT TIME ZONE 'Asia/Kathmandu')::DATE
                THEN 'your leave request was rejected today, you can request again tomorrow'

                ELSE ''
            END
        FROM attendance_leave al
        WHERE al.employee_id = $1
        ORDER BY al.created_at DESC
        LIMIT 1
    `, req.EmployeeID).Scan(&blockReason)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("failed to check existing leave request: %w", err)
	}
	if blockReason != "" {
		return errors.New(blockReason)
	}

	// ── Insert ───────────────────────────────────────────────────────────────
	_, err = tx.Exec(c, `
        INSERT INTO attendance_leave (
            employee_id,
            start_date,
            end_date,
            message,
            status,
            created_at,
            updated_at
        ) VALUES (
            $1,
            $2::TIMESTAMPTZ AT TIME ZONE 'Asia/Kathmandu',
            $3::TIMESTAMPTZ AT TIME ZONE 'Asia/Kathmandu',
            $4,
            $5,
            NOW(),
            NOW()
        )
    `,
		req.EmployeeID,
		req.StartDate,
		req.EndDate,
		req.Message,
		models.LeavePending,
	)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return fmt.Errorf("a leave request already exists for this date range")
		}
		return fmt.Errorf("failed to create leave request: %w", err)
	}

	if err = tx.Commit(c); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (r *attendanceRepo) DeleteInactiveLeaveRequestAttendance(ctx context.Context) error {
	query := `
		DELETE FROM attendance_leave
		WHERE status = 'pending'
		  AND checked_by IS NULL
		  AND created_at < NOW() - INTERVAL '2 days'
	`

	_, err := r.pool.Exec(ctx, query)
	if err != nil {
		return err
	}

	return nil
}

func (r *attendanceRepo) AutoReviewIncompleteAttendance(ctx context.Context) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// 1️⃣ Update incomplete attendance
	updateQuery := `
		UPDATE attendance
		SET 
			need_review = TRUE,
			check_out_time = NOW(),
			updated_at = NOW()
		WHERE check_in_time IS NOT NULL
		AND check_out_time IS NULL
		AND work_date < CURRENT_DATE
		RETURNING employee_id;
	`

	rows, err := tx.Query(ctx, updateQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Track affected employees
	employeeCount := make(map[string]int)

	for rows.Next() {
		var empID string
		if err := rows.Scan(&empID); err != nil {
			return err
		}
		employeeCount[empID]++
	}

	// 2️⃣ Check total incomplete history per employee
	for empID := range employeeCount {
		var count int

		err := tx.QueryRow(ctx, `
			SELECT COUNT(*) 
			FROM attendance
			WHERE employee_id = $1
			AND need_review = TRUE;
		`, empID).Scan(&count)

		if err != nil {
			return err
		}

		if count > 10 {
			log.Printf("⚠ ALERT: Employee %s has %d incomplete attendance records\n", empID, count)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	log.Println("✅ Daily auto review completed")
	return nil
}

func (r *attendanceRepo) GetAttendanceHistory(ctx context.Context, limit, page int, fromDate, toDate *time.Time, employeeID *uuid.UUID) (*models.AttendanceHistoryResponse, error) {
	fmt.Println("this is query : ", limit, page, fromDate, toDate, employeeID)
	errMessage := "failed to get attendance history"
	offset := page * limit
	orderBy := "DESC" // Most recent first

	// Main list query with employee filter only
	listQuery := fmt.Sprintf(`
	SELECT 
		a.id,
		a.employee_id,
		a.work_date,
		a.check_in_time,
		a.check_out_time,
		a.need_review,
		a.status,
		a.created_at,
		a.updated_at,
		
		u.name,
		u.email,
		u.role,
		u.image,
		u.phone
		
	FROM attendance a
	INNER JOIN users u ON a.employee_id = u.id
	WHERE
		($1::uuid IS NULL OR a.employee_id = $1::uuid)
		AND ($2::timestamptz IS NULL OR a.work_date >= $2::timestamptz)
		AND ($3::timestamptz IS NULL OR a.work_date <= $3::timestamptz)
	ORDER BY a.work_date %s, u.name ASC
	LIMIT $4 OFFSET $5;
`, orderBy)

	rows, err := r.pool.Query(
		ctx,
		listQuery,
		employeeID,
		fromDate,
		toDate,
		limit,
		offset,
	)

	if err != nil {
		log.Printf("error in getting attendance history: %v", err)
		return nil, errors.New(errMessage)
	}
	defer rows.Close()

	attendanceRecords := make([]models.AttendanceHistoryData, 0, limit)
	for rows.Next() {
		var record models.AttendanceHistoryData

		// Nullable fields
		var checkInTime *time.Time
		var checkOutTime *time.Time
		var userImage *string

		if err := rows.Scan(
			// Attendance fields
			&record.ID,
			&record.EmployeeID,
			&record.WorkDate,
			&checkInTime,
			&checkOutTime,
			&record.NeedReview,
			&record.Status,
			&record.CreatedAt,
			&record.UpdatedAt,

			// User fields
			&record.EmployeeName,
			&record.EmployeeEmail,
			&record.EmployeeRole,
			&userImage,
			&record.EmployeePhone,
		); err != nil {
			log.Printf("error scanning attendance history row: %v", err)
			return nil, errors.New(errMessage)
		}

		record.CheckInTime = checkInTime
		record.CheckOutTime = checkOutTime
		record.EmployeeImage = userImage

		attendanceRecords = append(attendanceRecords, record)
	}

	if err := rows.Err(); err != nil {
		log.Printf("error iterating attendance history rows: %v", err)
		return nil, errors.New(errMessage)
	}

	// Count query for pagination
	countQuery := `
	SELECT COUNT(*)
	FROM attendance a
	INNER JOIN users u ON a.employee_id = u.id
	WHERE
		($1::uuid IS NULL OR a.employee_id = $1::uuid)
		AND ($2::timestamptz IS NULL OR a.work_date >= $2::timestamptz)
		AND ($3::timestamptz IS NULL OR a.work_date <= $3::timestamptz);
	`

	var total int
	err = r.pool.QueryRow(
		ctx,
		countQuery,
		employeeID,
		fromDate,
		toDate,
	).Scan(&total)
	if err != nil {
		log.Printf("error counting attendance history: %v", err)
		return nil, errors.New(errMessage)
	}

	// Get statistics based on filters
	stats, err := r.GetAttendanceHistoryStats(ctx, employeeID, fromDate, toDate)
	if err != nil {
		log.Printf("error getting attendance stats: %v", err)
		// Don't return error, just log it - we can still return the data without stats
	}

	// Pagination info
	hasMore := (page+1)*limit < total
	nextPage := page + 1

	// Final response
	response := &models.AttendanceHistoryResponse{
		Data:     attendanceRecords,
		Total:    total,
		Page:     page,
		Limit:    limit,
		HasMore:  hasMore,
		NextPage: nextPage,
		Stats:    stats,
	}

	return response, nil
}

// Updated GetAttendanceHistoryStats without search parameter
func (r *attendanceRepo) GetAttendanceHistoryStats(ctx context.Context, employeeID *uuid.UUID, fromDate, toDate *time.Time) (*models.AttendanceHistoryStats, error) {
	statsQuery := `
	SELECT 
		COUNT(*) as total_records,
		COUNT(CASE WHEN status = 'present' THEN 1 END) as present_count,
		COUNT(CASE WHEN status = 'absent' THEN 1 END) as absent_count,
		COUNT(CASE WHEN status = 'leave' THEN 1 END) as leave_count,
		COUNT(CASE WHEN status = 'late' THEN 1 END) as late_count,
		COUNT(CASE WHEN status = 'half_day' THEN 1 END) as half_day_count,
		COUNT(CASE WHEN need_review = true THEN 1 END) as needs_review_count
	FROM attendance a
	INNER JOIN users u ON a.employee_id = u.id
	WHERE
		($1::uuid IS NULL OR a.employee_id = $1::uuid)
		AND ($2::timestamptz IS NULL OR a.work_date >= $2::timestamptz)
		AND ($3::timestamptz IS NULL OR a.work_date <= $3::timestamptz);
	`

	var stats models.AttendanceHistoryStats
	err := r.pool.QueryRow(ctx, statsQuery, employeeID, fromDate, toDate).Scan(
		&stats.TotalRecords,
		&stats.PresentCount,
		&stats.AbsentCount,
		&stats.LeaveCount,
		&stats.LateCount,
		&stats.HalfDayCount,
		&stats.NeedsReviewCount,
	)
	if err != nil {
		return nil, err
	}

	return &stats, nil
}

func (r *attendanceRepo) DeleteAttendanceById(
	ctx context.Context,
	attendanceID string,
) error {

	if attendanceID == "" {
		return errors.New("attendance_id is required")
	}

	id, err := uuid.FromString(attendanceID)
	if err != nil {
		return errors.New("invalid attendance_id")
	}

	query := `DELETE FROM attendance WHERE id = $1`

	cmdTag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return errors.New("attendance not found")
	}

	return nil
}

func (r *attendanceRepo) UpdateAttendance(
	ctx context.Context,
	req *models.AttendanceUpdate,
) error {
	if req == nil {
		return errors.New("request cannot be nil")
	}

	if req.AttendanceID == "" {
		return errors.New("attendance_id is required")
	}

	id, err := uuid.FromString(req.AttendanceID)
	if err != nil {
		return errors.New("invalid attendance_id")
	}

	setClauses := []string{}
	args := []interface{}{}
	argPos := 1

	// -------------------------
	// CHECK IN TIME
	// -------------------------
	if req.CheckInTime != nil {
		setClauses = append(setClauses, fmt.Sprintf("check_in_time = $%d", argPos))
		args = append(args, *req.CheckInTime)
		argPos++
	} else {
		setClauses = append(setClauses, fmt.Sprintf("check_in_time = $%d", argPos))
		args = append(args, nil) // set NULL
		argPos++
	}

	// -------------------------
	// CHECK OUT TIME
	// -------------------------
	if req.CheckOutTime != nil {
		setClauses = append(setClauses, fmt.Sprintf("check_out_time = $%d", argPos))
		args = append(args, *req.CheckOutTime)
		argPos++
	} else {
		setClauses = append(setClauses, fmt.Sprintf("check_out_time = $%d", argPos))
		args = append(args, nil) // set NULL
		argPos++
	}

	// -------------------------
	// NEED REVIEW
	// -------------------------
	setClauses = append(setClauses, fmt.Sprintf("need_review = $%d", argPos))
	args = append(args, req.NeedReview)
	argPos++

	// -------------------------
	// STATUS
	// -------------------------
	setClauses = append(setClauses, fmt.Sprintf("status = $%d", argPos))
	args = append(args, req.Status)
	argPos++

	// Always update updated_at
	setClauses = append(setClauses, "updated_at = NOW()")

	// Final query
	query := fmt.Sprintf(`
		UPDATE attendance
		SET %s
		WHERE id = $%d
	`, strings.Join(setClauses, ", "), argPos)

	args = append(args, id)

	// Execute
	cmdTag, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return errors.New("attendance not found")
	}

	return nil
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

	// Start transaction
	tx, err := r.pool.Begin(c)
	if err != nil {
		return err
	}
	defer tx.Rollback(c) // safe rollback

	// 1️⃣ Insert attendance
	insertQuery := `
		INSERT INTO attendance (
			employee_id,
			work_date,
			check_in_time,
			status
		)
		VALUES ($1, CURRENT_DATE, NOW(), 'present')
		RETURNING id
	`

	var attendanceID uuid.UUID
	err = tx.QueryRow(c, insertQuery, employeeUUID).Scan(&attendanceID)
	if err != nil {

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {

			if pgErr.Code == "23505" {
				return errors.New("attendance already marked for today")
			}

			if pgErr.Code == "23503" {
				return errors.New("employee does not exist")
			}
		}

		return err
	}

	// 2️⃣ Update user active = true
	updateUserQuery := `
		UPDATE users
		SET is_active = $1
		WHERE id = $2
	`

	res, err := tx.Exec(c, updateUserQuery, true, employeeUUID)
	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return errors.New("no user updated")
	}

	// Commit
	if err := tx.Commit(c); err != nil {
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

	tx, err := r.pool.Begin(c)
	if err != nil {
		return err
	}
	defer tx.Rollback(c)

	// 1️⃣ Update attendance (check out)
	updateAttendanceQuery := `
		UPDATE attendance
		SET check_out_time = NOW()
		WHERE employee_id = $1
		AND work_date = CURRENT_DATE
		AND check_out_time IS NULL
		RETURNING id
	`

	var attendanceID uuid.UUID
	err = tx.QueryRow(c, updateAttendanceQuery, employeeUUID).Scan(&attendanceID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("no active check-in found or already checked out")
		}
		return err
	}

	// 2️⃣ Make user inactive
	updateUserQuery := `
		UPDATE users
		SET is_active = $1
		WHERE id = $2
	`

	res, err := tx.Exec(c, updateUserQuery, false, employeeUUID)
	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return errors.New("no user updated")
	}

	// Commit transaction
	if err := tx.Commit(c); err != nil {
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
