package services

import (
	"errors"
	"fmt"
	"log"

	"github.com/Abhishekh669/backend/internals/config"
	"github.com/Abhishekh669/backend/internals/lib"
	"github.com/Abhishekh669/backend/internals/models"
	"github.com/Abhishekh669/backend/internals/rbac"
	"github.com/Abhishekh669/backend/internals/repository"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
)

type AttendanceService interface {
	GetAllAttendanceRequestLeaveHistoryService(c *gin.Context, query *models.AttendanceLeaveHistory) (*repository.AttendanceLeaveByUserResponse, error)
	GetAllAttendanceRequestLeaveByUserIdService(c *gin.Context, query *models.AttendanceLeaveHistory) (*repository.AttendanceLeaveByUserResponse, error)
	CancelLeaveAttendanceByAdmin(c *gin.Context, leaveId *uuid.UUID) error
	AcceptLeaveAttendanceByAdmin(c *gin.Context, leaveId *uuid.UUID) error
	GetAllAttendanceRequestLeaveService(c *gin.Context) ([]models.AttendanceLeaveResponse, error)
	UpdateUserAttendanceService(c *gin.Context, req *models.UserUpdateAttendanceLeave) error
	GetTodayAttendanceService(c *gin.Context) (*models.AttendanceLeaveResponse, error)
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

	userIDStr, ok := value.(string)
	if !ok {
		return uuid.Nil, errors.New("invalid user_id type in context")
	}

	userID, err := uuid.FromString(userIDStr)
	if err != nil {
		return uuid.Nil, errors.New("invalid uuid format")
	}

	return userID, nil
}

func (s *attendanceService) GetAllAttendanceRequestLeaveHistoryService(c *gin.Context, query *models.AttendanceLeaveHistory) (*repository.AttendanceLeaveByUserResponse, error) {
	_, err := lib.HasPermissionCheck(c, rbac.ViewAttendance)
	if err != nil {
		return nil, errors.New("user not authorized")
	}

	return s.repo.GetAllAttendanceLeaveRequestsHistory(c.Request.Context(), query.Limit, query.Page, query.FromDate, query.ToDate, (*models.LeaveStatus)(&query.Status))
}

func (s *attendanceService) GetAllAttendanceRequestLeaveByUserIdService(c *gin.Context, query *models.AttendanceLeaveHistory) (*repository.AttendanceLeaveByUserResponse, error) {
	_, err := lib.HasPermissionCheck(c, rbac.ViewAttendance)
	if err != nil {
		fmt.Println("userare not authorized ")
		return nil, errors.New("user not authorized")
	}

	userId, err := GetUserIDFromContext(c)
	if err != nil {
		fmt.Println("user id is not present ")
		return nil, errors.New("failed to get employee id")
	}

	return s.repo.GetAttendanceLeaveRequestByUserId(c.Request.Context(), userId, query.Limit, query.Page, query.FromDate, query.ToDate, (*models.LeaveStatus)(&query.Status))
}

func (s *attendanceService) GetAllAttendanceRequestLeaveService(c *gin.Context) ([]models.AttendanceLeaveResponse, error) {
	_, err := lib.HasPermissionCheck(c, rbac.ViewAttendance)
	if err != nil {
		return nil, errors.New("user not authorized")
	}
	return s.repo.GetAllAttendanceLeaveRequest(c.Request.Context())
}

func (s *attendanceService) AcceptLeaveAttendanceByAdmin(c *gin.Context, leaveId *uuid.UUID) error {
	_, err := lib.HasPermissionCheck(c, rbac.UpdateAttendance)
	if err != nil {
		return errors.New("user not authorized")
	}

	userId, err := GetUserIDFromContext(c)
	if err != nil {
		return errors.New("failed to get employee id")
	}

	res, err := s.repo.AcceptLeaveRequestByAdmin(c.Request.Context(), *leaveId, userId)
	if err != nil {
		return err
	}

	// Send email asynchronously to not block the response
	go func() {
		emailData := lib.LeaveEmailData{
			EmployeeName:      res.EmployeeName,
			EmployeeEmail:     res.EmployeeEmail,
			Status:            models.LeaveApproved,
			StartDate:         res.StartDate,
			EndDate:           res.EndDate,
			Message:           res.Message,
			SupervisorMessage: res.SupervisorMessage,
		}

		mailService := lib.NewMailService(lib.EmailConfig{
			SMTPHost:       config.AppConfig.SMTPHost,
			SMTPPort:       config.AppConfig.SMTPPort,
			SenderEmail:    config.AppConfig.SMTPEmail,
			SenderPassword: config.AppConfig.SMTPPassword,
		})

		if err := mailService.SendLeaveStatusEmail(res.EmployeeEmail, emailData); err != nil {
			log.Printf("ERROR: failed to send leave approval email to %s: %v", res.EmployeeEmail, err)
		} else {
			log.Printf("SUCCESS: leave approval email sent to %s", res.EmployeeEmail)
		}
	}()

	return nil
}

func (s *attendanceService) CancelLeaveAttendanceByAdmin(c *gin.Context, leaveId *uuid.UUID) error {
	_, err := lib.HasPermissionCheck(c, rbac.UpdateAttendance)
	if err != nil {
		return errors.New("user not authorized")
	}

	userId, err := GetUserIDFromContext(c)
	if err != nil {
		return errors.New("failed to get employee id")
	}

	res, err := s.repo.CancelLeaveRequestByAdmin(c.Request.Context(), leaveId, userId)
	if err != nil {
		return err
	}

	go func() {
		emailData := lib.LeaveEmailData{
			EmployeeName:      res.EmployeeName,
			EmployeeEmail:     res.EmployeeEmail,
			Status:            models.LeaveRejected,
			StartDate:         res.StartDate,
			EndDate:           res.EndDate,
			Message:           res.Message,
			SupervisorMessage: res.SupervisorMessage,
		}

		mailService := lib.NewMailService(lib.EmailConfig{
			SMTPHost:       config.AppConfig.SMTPHost,
			SMTPPort:       config.AppConfig.SMTPPort,
			SenderEmail:    config.AppConfig.SMTPEmail,
			SenderPassword: config.AppConfig.SMTPPassword,
		})

		if err := mailService.SendLeaveStatusEmail(res.EmployeeEmail, emailData); err != nil {
			log.Printf("failed to send leave approval email to %s: %v", res.EmployeeEmail, err)
		}
	}()

	return nil
}

func (s *attendanceService) UpdateUserAttendanceService(c *gin.Context, req *models.UserUpdateAttendanceLeave) error {
	_, err := lib.HasPermissionCheck(c, rbac.ViewLeaveRequest)
	if err != nil {
		return errors.New("user not authorized")
	}
	return s.repo.UpdateCustomerLeave(c.Request.Context(), req)
}

func (s *attendanceService) GetTodayAttendanceService(c *gin.Context) (*models.AttendanceLeaveResponse, error) {

	userId, err := GetUserIDFromContext(c)
	if err != nil {
		return nil, errors.New("failed to get employee id")
	}
	return s.repo.GetTodayAttendanceLeave(c.Request.Context(), userId)
}

func (s *attendanceService) GetAttendanceRequestService(c *gin.Context) ([]models.AttendanceLeaveResponse, error) {
	_, err := lib.HasPermissionCheck(c, rbac.ViewLeaveRequest)
	if err != nil {
		return nil, errors.New("user not authorized")
	}
	return s.repo.GetAttendanceRequest(c.Request.Context())

}

func (s *attendanceService) CancelLeaveRequest(
	c *gin.Context,
	leaveId *uuid.UUID,
) error {

	userId, err := GetUserIDFromContext(c)
	if err != nil {
		return errors.New("failed to cancel request")
	}

	// TODO  : later add sending mail and deleting the attendance leave

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
		fmt.Println("thisis userid : ", userId)
		return errors.New("failed to create employ leave request")
	}

	fmt.Println("iam next one")

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
