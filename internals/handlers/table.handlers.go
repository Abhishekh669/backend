package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Abhishekh669/backend/internals/models"
	"github.com/Abhishekh669/backend/internals/services"
	"github.com/gin-gonic/gin"
)

type TableHandler struct {
	tableService services.TableService
}

func (h *TableHandler) DeleteTablesHandler(c *gin.Context) {
	var req struct {
		TableIDs []string `json:"table_ids"`
	}

	if err := c.ShouldBindJSON(&req); err != nil || len(req.TableIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "table_ids are required",
		})
		return
	}

	if err := h.tableService.DeleteTableService(c, req.TableIDs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "tables deleted successfully",
	})
}

func (h *TableHandler) UpdateTableHandler(c *gin.Context) {
	var table models.UpdateTableStatus

	if err := c.ShouldBindJSON(&table); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid request body",
		})
		return
	}

	if err := h.tableService.UpdateTableService(c, &table); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "table updated successfully",
	})
}

func (h *TableHandler) GetTablesHandler(c *gin.Context) {
	tables, err := h.tableService.ViewTableService(c)

	fmt.Println("this is tables : ", tables)
	fmt.Println("this is error : ", err)

	if err != nil {
		// Check if this is a permission error
		if err.Error() == "error not found" || strings.Contains(err.Error(), "permission") {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "You don't have permission to view tables",
				"tables":  []interface{}{},
			})
			return
		}

		// For other types of errors
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch tables",
			"tables":  []interface{}{},
		})
		return
	}

	// Ensure tables is never nil
	if tables == nil {
		tables = make([]models.TableStatus, 0)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Tables retrieved successfully",
		"tables":  tables,
	})
}

func (h *TableHandler) CreateTablesHandler(c *gin.Context) {
	var tables []models.CreateTableType
	if err := c.ShouldBindJSON(&tables); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	err := h.tableService.CreateNewTableService(c, tables)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "created successfully"})
}

func NewTableHandler(tableService services.TableService) *TableHandler {
	return &TableHandler{
		tableService: tableService,
	}
}
