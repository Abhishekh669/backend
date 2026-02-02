package handlers

import (
	"fmt"
	"net/http"

	"github.com/Abhishekh669/backend/internals/services"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
)

type FoodCategoryHandler struct {
	foodCategoryService services.FoodCategoryService
}

type CreateFoodCategoryType struct {
	CategoryName string `json:"category_name"`
	ParentId     string `json:"parent_id"`
}

func (h *FoodCategoryHandler) GetFoodCategoriesHandlers(c *gin.Context) {
	foodCategories, err := h.foodCategoryService.GetFoodCategories(c)

	if err != nil {
		fmt.Println("error in binding", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "successfully got data", "success": true, "categories": foodCategories})

}

func (h *FoodCategoryHandler) CreateFoodCategoryHandler(c *gin.Context) {
	var data CreateFoodCategoryType
	if err := c.ShouldBindJSON(&data); err != nil {
		fmt.Println("error in binding", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}
	var parentUUID *uuid.UUID

	if data.ParentId != "" {
		parsed, err := uuid.FromString(data.ParentId)
		if err != nil {
			fmt.Println("failed to covert to parent uuid : ", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "failed to create food category",
				"success": false,
			})
			return
		}
		parentUUID = &parsed
	}

	err := h.foodCategoryService.CreateCategoryService(c, data.CategoryName, parentUUID)
	if err != nil {
		fmt.Println("this is hte error in creating food cateogyr : ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "food category created successfully",
		"success": true,
	})

}

func NewFoodCategoryHandler(foodCategory services.FoodCategoryService) *FoodCategoryHandler {
	return &FoodCategoryHandler{
		foodCategoryService: foodCategory,
	}
}
