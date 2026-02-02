package services

import (
	"github.com/Abhishekh669/backend/internals/lib"
	"github.com/Abhishekh669/backend/internals/models"
	"github.com/Abhishekh669/backend/internals/rbac"
	"github.com/Abhishekh669/backend/internals/repository"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
)

type FoodCategoryService interface {
	GetFoodCategories(c *gin.Context) ([]models.Category, error)
	CreateCategoryService(c *gin.Context, categoryName string, parentId *uuid.UUID) error
}
type foodCategoryService struct {
	repo repository.FoodCategoryRepo
}

func (s *foodCategoryService) GetFoodCategories(c *gin.Context) ([]models.Category, error) {
	_, err := lib.HasPermissionCheck(c, rbac.ViewFoodCategory)
	if err != nil {
		return nil, err
	}
	return s.repo.GetFoodCategory(c.Request.Context())
}

func (s *foodCategoryService) CreateCategoryService(c *gin.Context, categoryName string, parentId *uuid.UUID) error {
	_, err := lib.HasPermissionCheck(c, rbac.CreateFoodCategory)
	if err != nil {
		return err
	}
	return s.repo.CreateCategory(c.Request.Context(), categoryName, parentId)
}

func NewFoodCategoryService(repo repository.FoodCategoryRepo) FoodCategoryService {
	return &foodCategoryService{
		repo: repo,
	}
}
