package services

import (
	"errors"
	"log"
	"strings"

	"github.com/Abhishekh669/backend/internals/lib"
	"github.com/Abhishekh669/backend/internals/models"

	"github.com/Abhishekh669/backend/internals/rbac"
	"github.com/Abhishekh669/backend/internals/repository"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
)

type FoodCategoryService interface {
	UpdateMenuItemService(c *gin.Context, menuItem *models.UpdateMenuItemType) error
	UpdateCategoryService(c *gin.Context, category *models.UpdateCategoryType) error
	DeleteMenuItemsService(c *gin.Context, menuItemIds []string) error
	DeleteCategoryService(c *gin.Context, categoryIds []string) error
	CreateMenuItemsService(c *gin.Context, menuItems []models.CreateMenuItemType, categoryId *uuid.UUID) error
	GetFoodCategoriesBySlug(c *gin.Context, slug string) (*repository.GetCategoriesBySlug, error)
	GetFoodCategories(c *gin.Context) ([]models.Category, error)
	CreateCategoryService(c *gin.Context, categoryName string, slugPath []string) error
}
type foodCategoryService struct {
	repo repository.FoodCategoryRepo
}

func parseSlugPath(raw string) []string {
	raw = strings.TrimPrefix(raw, "/")
	if raw == "" {
		return []string{}
	}
	return strings.Split(raw, "/")
}

func (s *foodCategoryService) UpdateMenuItemService(c *gin.Context, menuItem *models.UpdateMenuItemType) error {
	_, err := lib.HasPermissionCheck(c, rbac.UpdateFoodSubCategory)
	if err != nil {
		return err
	}
	return s.repo.UpdateMenuItems(c.Request.Context(), menuItem)
}

func (s *foodCategoryService) UpdateCategoryService(c *gin.Context, category *models.UpdateCategoryType) error {
	_, err := lib.HasPermissionCheck(c, rbac.UpdateFoodCategory)
	if err != nil {
		return err
	}
	return s.repo.UpdateCategory(c.Request.Context(), category)
}

func (s *foodCategoryService) DeleteMenuItemsService(c *gin.Context, menuItemIds []string) error {
	_, err := lib.HasPermissionCheck(c, rbac.DeleteFoodCategory)
	if err != nil {
		return err
	}
	return s.repo.DeleteMenuItems(c.Request.Context(), menuItemIds)
}

func (s *foodCategoryService) DeleteCategoryService(c *gin.Context, categoryIds []string) error {
	_, err := lib.HasPermissionCheck(c, rbac.DeleteFoodCategory)
	if err != nil {
		return err
	}
	return s.repo.DeleteCategories(c.Request.Context(), categoryIds)
}

func (s *foodCategoryService) CreateMenuItemsService(c *gin.Context, menuItems []models.CreateMenuItemType, categoryId *uuid.UUID) error {
	_, err := lib.HasPermissionCheck(c, rbac.CreateFoodSubCategory)
	if err != nil {
		return err
	}
	return s.repo.CreateMenuItems(c.Request.Context(), menuItems, categoryId)
}

func (s *foodCategoryService) GetFoodCategoriesBySlug(c *gin.Context, slug string) (*repository.GetCategoriesBySlug, error) {
	rawSlug := parseSlugPath(slug)
	_, err := lib.HasPermissionCheck(c, rbac.ViewFoodCategory)
	if err != nil {
		log.Println("error in food service in get food by slug : ", err)
		return nil, errors.New("failed to get food categories")
	}
	return s.repo.GetFoodCategoriesFromSlug(c.Request.Context(), rawSlug)
}

func (s *foodCategoryService) GetFoodCategories(c *gin.Context) ([]models.Category, error) {
	_, err := lib.HasPermissionCheck(c, rbac.ViewFoodCategory)
	if err != nil {
		return nil, err
	}
	return s.repo.GetFoodCategory(c.Request.Context())
}

func (s *foodCategoryService) CreateCategoryService(c *gin.Context, categoryName string, slugPath []string) error {
	_, err := lib.HasPermissionCheck(c, rbac.CreateFoodCategory)
	if err != nil {
		return err
	}
	return s.repo.CreateCateogry(c.Request.Context(), slugPath, categoryName)
}

func NewFoodCategoryService(repo repository.FoodCategoryRepo) FoodCategoryService {
	return &foodCategoryService{
		repo: repo,
	}
}
