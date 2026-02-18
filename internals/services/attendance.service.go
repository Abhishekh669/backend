package services

import (
	"github.com/Abhishekh669/backend/internals/lib"
	"github.com/Abhishekh669/backend/internals/models"
	"github.com/Abhishekh669/backend/internals/rbac"
	"github.com/Abhishekh669/backend/internals/repository"
	"github.com/gin-gonic/gin"
)

type AttendanceService interface {
	GetCurrentAttendanceService(c *gin.Context) (*models.CurrentAttendance, error)
	CheckInAttendanceService(c *gin.Context, req *models.CheckInOutAttendanceType) error
	CheckOutAttendanceService(c *gin.Context, req *models.CheckInOutAttendanceType) error
}
type attendanceService struct {
	repo repository.AttendanceRepo
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
