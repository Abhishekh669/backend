package services

import (
	"context"
	"errors"
	"log"
	"strings"

	"github.com/Abhishekh669/backend/internals/lib"
	"github.com/Abhishekh669/backend/internals/models"

	"github.com/Abhishekh669/backend/internals/rbac"
	"github.com/Abhishekh669/backend/internals/repository"
	"github.com/gin-gonic/gin"
)

type FoodCategoryService interface {
	GetAllMenuItemsGroupedService(ctx context.Context) (map[string]repository.CategoryMenuGroup, error)
	NewGetFoodCategories(c *gin.Context) ([]models.NewCategory, error)
	NewCreateCategoryService(c *gin.Context, categoryName string) error
	NewUpdateCategoryService(c *gin.Context, category *models.NewUpdateCategoryType) error
	NewGetAllMenuItemsService(c *gin.Context, slug string) ([]models.MenuItemsResponse, error)
	UpdateMenuItemService(c *gin.Context, menuItem *models.UpdateMenuItemType) error
	UpdateCategoryService(c *gin.Context, category *models.UpdateCategoryType) error
	DeleteMenuItemsService(c *gin.Context, menuItemIds []string) error
	DeleteCategoryService(c *gin.Context, categoryIds []string) error
	CreateMenuItemsService(c *gin.Context, menuItems []models.CreateMenuItemType, categorySlug string) error
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

func (s *foodCategoryService) GetAllMenuItemsGroupedService(ctx context.Context) (map[string]repository.CategoryMenuGroup, error) {
	return s.repo.GetAllMenuItemsGrouped(ctx)
}

func (s *foodCategoryService) NewGetFoodCategories(c *gin.Context) ([]models.NewCategory, error) {
	_, err := lib.HasPermissionCheck(c, rbac.ViewFoodCategory)
	if err != nil {
		log.Println("error in food service in get food by slug : ", err)
		return nil, errors.New("failed to get food categories")
	}
	return s.repo.NewGetFoodCategory(c.Request.Context())
}

func (s *foodCategoryService) NewCreateCategoryService(c *gin.Context, categoryName string) error {
	_, err := lib.HasPermissionCheck(c, rbac.CreateFoodCategory)
	if err != nil {
		return err
	}
	return s.repo.NewCreateCategory(c.Request.Context(), categoryName)
}

func (s *foodCategoryService) NewGetAllMenuItemsService(c *gin.Context, slug string) ([]models.MenuItemsResponse, error) {
	_, err := lib.HasPermissionCheck(c, rbac.ViewFoodCategory)
	if err != nil {
		return nil, err
	}
	return s.repo.NewGetAllTheMenuItemsFromSlug(c.Request.Context(), slug)
}

func (s *foodCategoryService) NewUpdateCategoryService(c *gin.Context, category *models.NewUpdateCategoryType) error {
	_, err := lib.HasPermissionCheck(c, rbac.UpdateFoodCategory)
	if err != nil {
		return err
	}
	return s.repo.NewUpdateCategory(c.Request.Context(), category)
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

func (s *foodCategoryService) CreateMenuItemsService(c *gin.Context, menuItems []models.CreateMenuItemType, categorySlug string) error {
	_, err := lib.HasPermissionCheck(c, rbac.CreateFoodSubCategory)
	if err != nil {
		return err
	}
	return s.repo.CreateMenuItems(c.Request.Context(), menuItems, categorySlug)
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
