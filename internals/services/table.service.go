package services

import (
	"github.com/Abhishekh669/backend/internals/lib"
	"github.com/Abhishekh669/backend/internals/models"
	"github.com/Abhishekh669/backend/internals/rbac"
	"github.com/Abhishekh669/backend/internals/repository"
	"github.com/gin-gonic/gin"
)

type TableService interface {
	DeleteTableService(c *gin.Context, tableIds []string) error
	UpdateTableService(c *gin.Context, table *models.UpdateTableStatus) error
	ViewTableService(c *gin.Context) ([]models.TableStatus, error)
	CreateNewTableService(c *gin.Context, tables []models.CreateTableType) error
}

type tableService struct {
	repo repository.TableRepo
}

func (s *tableService) UpdateTableService(c *gin.Context, table *models.UpdateTableStatus) error {
	_, err := lib.HasPermissionCheck(c, rbac.UpdateTable)
	if err != nil {
		return err
	}

	return s.repo.UpdateTables(c.Request.Context(), table)
}

func (s *tableService) DeleteTableService(c *gin.Context, tableIds []string) error {
	_, err := lib.HasPermissionCheck(c, rbac.DeleteTable)
	if err != nil {
		return err
	}

	return s.repo.DeleteTables(c.Request.Context(), tableIds)
}

func (s *tableService) ViewTableService(c *gin.Context) ([]models.TableStatus, error) {

	return s.repo.ViewTables(c.Request.Context())
}

func (s *tableService) CreateNewTableService(c *gin.Context, tables []models.CreateTableType) error {
	_, err := lib.HasPermissionCheck(c, rbac.CreateTable)
	if err != nil {
		return err
	}
	return s.repo.CreateTables(c.Request.Context(), tables)
}

func NewTableService(repo repository.TableRepo) TableService {
	return &tableService{
		repo: repo,
	}
}
