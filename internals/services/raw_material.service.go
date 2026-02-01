package services

import (
	"fmt"
	"time"

	"github.com/Abhishekh669/backend/internals/lib"
	"github.com/Abhishekh669/backend/internals/models"
	"github.com/Abhishekh669/backend/internals/rbac"
	"github.com/Abhishekh669/backend/internals/repository"
	"github.com/gin-gonic/gin"
)

type RawMaterialService interface {
	UpdateRawMaterialService(c *gin.Context, data *models.UpdateRawMaterials) error
	DeleteRawMaterails(c *gin.Context, rawMaterials []string) error
	GetRawMaterialService(c *gin.Context, limit, page int, fromDate *time.Time, toDate *time.Time, startingPrice, endingPrice int32, search string, oldFirst bool) (*repository.RawMaterialsResponse, error)
	CreateRawMaterialService(c *gin.Context, rawMaterials []models.CreateRawMaterialType) error
}
type rawMaterialService struct {
	repo repository.RawMaterialsRepo
}

func (s *rawMaterialService) UpdateRawMaterialService(c *gin.Context, data *models.UpdateRawMaterials) error {
	if data.Unit == "" || data.Id == "" || data.Name == "" || data.Price < 0 || data.Quantity < 0 {
		return fmt.Errorf("failed to update raw materials")
	}
	_, err := lib.HasPermissionCheck(c, rbac.UpdateRawMaterials)
	if err != nil {
		return err
	}
	return s.repo.UpdateRawMaterial(c.Request.Context(), data)
}

func (s *rawMaterialService) DeleteRawMaterails(c *gin.Context, rawMaterials []string) error {
	_, err := lib.HasPermissionCheck(c, rbac.DeleteRawMaterials)
	if err != nil {
		return err
	}
	return s.repo.DeleteRawMaterials(c.Request.Context(), rawMaterials)
}

func (s *rawMaterialService) GetRawMaterialService(c *gin.Context, limit, page int, fromDate *time.Time, toDate *time.Time, startingPrice, endingPrice int32, search string, oldFirst bool) (*repository.RawMaterialsResponse, error) {
	_, err := lib.HasPermissionCheck(c, rbac.ViewRawMaterials)
	if err != nil {
		return nil, err
	}

	return s.repo.GetRawMaterials(c.Request.Context(), limit, page, fromDate, toDate, startingPrice, endingPrice, search, oldFirst)
}

func (s *rawMaterialService) CreateRawMaterialService(c *gin.Context, rawMaterials []models.CreateRawMaterialType) error {
	_, err := lib.HasPermissionCheck(c, rbac.CreateRawMaterials)
	if err != nil {
		return err
	}
	return s.repo.CreateRawMaterials(c.Request.Context(), rawMaterials)
}

func NewRawMaterialService(repo repository.RawMaterialsRepo) RawMaterialService {
	return &rawMaterialService{
		repo: repo,
	}
}
