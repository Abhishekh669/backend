package handlers

import (
	"net/http"

	"github.com/Abhishekh669/backend/internals/models"
	"github.com/Abhishekh669/backend/internals/services"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
)

type SettingHandler struct {
	settingService services.SettingService
}

func (h *SettingHandler) GetRestaurantInformationHandler(c *gin.Context) {

	info, err := h.settingService.GetRestaurantInformation(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to get restaurant information",
			"success": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"info":    info,
		"success": true,
	})
}

func (h *SettingHandler) UpdateRestaurantInformationHandler(c *gin.Context) {

	var req models.UpdateRestaurantSettings

	// Bind JSON body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"success": false,
		})
		return
	}

	// Validate ID (IMPORTANT)
	if req.ID == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "id is required",
			"success": false,
		})
		return
	}

	// Name validation (optional but recommended)
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "restaurant name is required",
			"success": false,
		})
		return
	}

	// Call service
	err := h.settingService.UpdateRestaurantInformation(c, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to update restaurant information",
			"success": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "restaurant information updated successfully",
		"success": true,
	})
}

func NewSettingHandler(settingService services.SettingService) *SettingHandler {
	return &SettingHandler{settingService: settingService}
}
