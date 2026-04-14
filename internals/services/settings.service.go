package services

import (
	"errors"

	"github.com/Abhishekh669/backend/internals/lib"
	"github.com/Abhishekh669/backend/internals/models"
	"github.com/Abhishekh669/backend/internals/rbac"
	"github.com/Abhishekh669/backend/internals/repository"
	"github.com/gin-gonic/gin"
)

type SettingService interface {
	GetRestaurantInformation(ctx *gin.Context) (*models.RestaurantSettings, error)
	UpdateRestaurantInformation(c *gin.Context, info *models.UpdateRestaurantSettings) error
}
type settingService struct {
	repo repository.SettingRepo
}

func (s *settingService) GetRestaurantInformation(ctx *gin.Context) (*models.RestaurantSettings, error) {
	_, err := lib.HasPermissionCheck(ctx, rbac.ViewRestaurantInformation)
	if err != nil {
		return nil, errors.New("user not authorized")
	}
	return s.repo.GetRestaurantInformation(ctx)
}

func (s *settingService) UpdateRestaurantInformation(c *gin.Context, info *models.UpdateRestaurantSettings) error {
	_, err := lib.HasPermissionCheck(c, rbac.UpdateRestaurantInformation)
	if err != nil {
		return errors.New("user not authorized")
	}
	return s.repo.UpdateRestaurantInformation(c.Request.Context(), info)
}

func NewSettingService(repo repository.SettingRepo) SettingService {
	return &settingService{repo: repo}
}
