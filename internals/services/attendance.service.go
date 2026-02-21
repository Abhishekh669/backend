package services

import (
	"errors"

	"github.com/Abhishekh669/backend/internals/lib"
	"github.com/Abhishekh669/backend/internals/models"
	"github.com/Abhishekh669/backend/internals/rbac"
	"github.com/Abhishekh669/backend/internals/repository"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
)

type AttendanceService interface {
	CancelLeaveRequest(c *gin.Context, leaveId *uuid.UUID) error
	DeleteLeaveRequest(c *gin.Context, leaveId *uuid.UUID) error
	UpdateLeaveRequest(c *gin.Context, req *models.UpdateAttendanceLeave) error
	CreateEmployeeRequest(c *gin.Context, req *models.CreateAttendanceLeave) error
	GetAttendanceHistory(c *gin.Context, query *models.AttendanceHistoryQuery) (*models.AttendanceHistoryResponse, error)
	DeleteAttendanceByIdService(c *gin.Context, attendanceID string) error
	UpdateAttendanceService(c *gin.Context, req *models.AttendanceUpdate) error
	GetCurrentAttendanceService(c *gin.Context) (*models.CurrentAttendance, error)
	CheckInAttendanceService(c *gin.Context, req *models.CheckInOutAttendanceType) error
	CheckOutAttendanceService(c *gin.Context, req *models.CheckInOutAttendanceType) error
}
type attendanceService struct {
	repo repository.AttendanceRepo
}

func GetUserIDFromContext(c *gin.Context) (uuid.UUID, error) {
	value, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, errors.New("user_id not found in context")
	}

	userID, ok := value.(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.New("invalid user_id type in context")
	}

	return userID, nil
}

func (s *attendanceService) CancelLeaveRequest(
	c *gin.Context,
	leaveId *uuid.UUID,
) error {

	_, err := lib.HasPermissionCheck(c, rbac.CancelLeaveRequest)
	if err != nil {
		return errors.New("user not authorized")
	}

	userId, err := GetUserIDFromContext(c)
	if err != nil {
		return errors.New("failed to cancel request")
	}

	return s.repo.CancelLeaveRequest(
		c.Request.Context(),
		leaveId,
		&userId,
	)
}

func (s *attendanceService) DeleteLeaveRequest(c *gin.Context, leaveId *uuid.UUID) error {

	_, err := lib.HasPermissionCheck(c, rbac.DeleteLeaveRequest)
	if err != nil {
		return errors.New("user not authorized")
	}

	return s.repo.DeleteLeaveRequest(
		c.Request.Context(),
		leaveId,
	)
}

func (s *attendanceService) UpdateLeaveRequest(c *gin.Context, req *models.UpdateAttendanceLeave) error {

	_, err := lib.HasPermissionCheck(c, rbac.UpdateLeaveRequest)
	if err != nil {
		return errors.New("user not authorized")
	}

	return s.repo.UpdateLeaveRequest(
		c.Request.Context(),
		req,
	)
}
func (s *attendanceService) CreateEmployeeRequest(c *gin.Context, req *models.CreateAttendanceLeave) error {

	_, err := lib.HasPermissionCheck(c, rbac.ViewLeaveRequest)
	if err != nil {
		return errors.New("user not authorized")
	}

	userId, err := GetUserIDFromContext(c)
	if err != nil {
		return errors.New("failed to cancel request")
	}

	// Ensure employee cannot spoof employee_id
	req.EmployeeID = userId

	return s.repo.CreateEmployeeRequest(
		c.Request.Context(),
		req,
	)
}

func (s *attendanceService) GetAttendanceHistory(c *gin.Context, query *models.AttendanceHistoryQuery) (*models.AttendanceHistoryResponse, error) {
	_, err := lib.HasPermissionCheck(c, rbac.ViewAttendance)
	if err != nil {
		return nil, errors.New("user not unauthorized")
	}
	return s.repo.GetAttendanceHistory(c.Request.Context(), query.Limit, query.Page, query.FromDate, query.ToDate, query.EmployeeId)
}

func (s *attendanceService) DeleteAttendanceByIdService(c *gin.Context, attendanceID string) error {
	_, err := lib.HasPermissionCheck(c, rbac.DeleteAttendance)
	if err != nil {
		return err
	}
	return s.repo.DeleteAttendanceById(c.Request.Context(), attendanceID)
}

func (s *attendanceService) UpdateAttendanceService(c *gin.Context, req *models.AttendanceUpdate) error {
	_, err := lib.HasPermissionCheck(c, rbac.UpdateAttendance)
	if err != nil {
		return err
	}
	return s.repo.UpdateAttendance(c.Request.Context(), req)
}

func (s *attendanceService) GetCurrentAttendanceService(c *gin.Context) (*models.CurrentAttendance, error) {
	_, err := lib.HasPermissionCheck(c, rbac.ViewAttendance)
	if err != nil {
		return nil, err
	}
	return s.repo.GetCurrentAttendance(c.Request.Context())
}

func (s *attendanceService) CheckInAttendanceService(c *gin.Context, req *models.CheckInOutAttendanceType) error {
	_, err := lib.HasPermissionCheck(c, rbac.CheckInAttendance)

	if err != nil {
		return err
	}
	return s.repo.CheckInAttendance(c.Request.Context(), req)
}

func (s *attendanceService) CheckOutAttendanceService(c *gin.Context, req *models.CheckInOutAttendanceType) error {
	_, err := lib.HasPermissionCheck(c, rbac.CheckOutAttendance)

	if err != nil {
		return err
	}
	return s.repo.CheckOutAttendance(c.Request.Context(), req)
}

func NewAttendanceService(repo repository.AttendanceRepo) AttendanceService {
	return &attendanceService{
		repo: repo,
	}
}
