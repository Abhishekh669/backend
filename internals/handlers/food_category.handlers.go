package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Abhishekh669/backend/internals/algorithm"
	"github.com/Abhishekh669/backend/internals/models"
	"github.com/Abhishekh669/backend/internals/services"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
)

type FoodCategoryHandler struct {
	foodCategoryService services.FoodCategoryService
}

type CreateFoodCategoryType struct {
	CategoryName string   `json:"category_name"`
	SlugPath     []string `json:"slug_path"`
}

type CreateMenuItemsType struct {
	CategoryId *string                     `json:"category_id"`
	MenuItems  []models.CreateMenuItemType `json:"menu_items"`
}

type DeleteCategoriesPayload struct {
	CategoriesId []string `json:"category_ids"`
}

type DeleteMenuItemsPayload struct {
	MenuItemsId []string `json:"menu_items_ids"`
}

func StringToUUIDPtr(idStr string) (*uuid.UUID, error) {
	parsedUUID, err := uuid.FromString(idStr)
	if err != nil {
		return nil, err
	}

	return &parsedUUID, nil
}

func (h *FoodCategoryHandler) UpdateMenuItemHandler(newCache *algorithm.MenuCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		var data models.UpdateMenuItemType
		if err := c.ShouldBindJSON(&data); err != nil {
			fmt.Println("error in binding", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
			return
		}

		err := h.foodCategoryService.UpdateMenuItemService(c, &data)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
			return
		}

		go newCache.ReloadFromDB()

		c.JSON(http.StatusOK, gin.H{
			"message": "menu item updated successfully",
			"success": true,
		})

	}
}
func (h *FoodCategoryHandler) UpdateCategoryHandler(newCache *algorithm.MenuCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		var data models.UpdateCategoryType
		if err := c.ShouldBindJSON(&data); err != nil {
			fmt.Println("error in binding", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
			return
		}

		err := h.foodCategoryService.UpdateCategoryService(c, &data)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
			return
		}

		go newCache.ReloadFromDB()

		c.JSON(http.StatusOK, gin.H{
			"message": "category updated successfully",
			"success": true,
		})

	}
}

func (h *FoodCategoryHandler) DeleteMenuItemsHandler(newCache *algorithm.MenuCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		var deleteIds DeleteMenuItemsPayload
		if err := c.ShouldBindJSON(&deleteIds); err != nil {
			fmt.Println("error in binding", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
			return
		}
		if len(deleteIds.MenuItemsId) < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no menu items selected", "success": false})
			return
		}

		err := h.foodCategoryService.DeleteMenuItemsService(c, deleteIds.MenuItemsId)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
			return
		}
		var message string
		if len(deleteIds.MenuItemsId) > 0 {
			message = fmt.Sprintf(" %d menu items deleted successfully", len(deleteIds.MenuItemsId))
		} else {
			message = "menu item deleted successfully"
		}
		go newCache.ReloadFromDB()
		c.JSON(http.StatusOK, gin.H{"message": message, "success": true})

	}
}
func (h *FoodCategoryHandler) DeleteCategoriesHandler(newCache *algorithm.MenuCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		var data DeleteCategoriesPayload
		if err := c.ShouldBindJSON(&data); err != nil {
			fmt.Println("error in binding", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
			return
		}

		if len(data.CategoriesId) < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no categories selected", "success": false})
			return
		}

		err := h.foodCategoryService.DeleteCategoryService(c, data.CategoriesId)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
			return
		}
		var message string
		if len(data.CategoriesId) > 0 {
			message = fmt.Sprintf(" %d categories deleted successfully", len(data.CategoriesId))
		} else {
			message = "category deleted successfully"
		}
		go newCache.ReloadFromDB()
		c.JSON(http.StatusOK, gin.H{"message": message, "success": true})

	}
}

func (h *FoodCategoryHandler) CreateMenuItemsHandler(newCache *algorithm.MenuCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		var data CreateMenuItemsType
		if err := c.ShouldBindJSON(&data); err != nil {
			fmt.Println("error in binding", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
			return
		}

		if data.CategoryId == nil {
			fmt.Println("not parent found")
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid data",
				"success": false,
			})
			return
		}

		category_uuid, err := StringToUUIDPtr(*data.CategoryId)
		if err != nil {
			fmt.Println("error in parsing category id : ", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid category id",
				"success": false,
			})
			return
		}

		fmt.Println("this is menu items : ", data)

		err = h.foodCategoryService.CreateMenuItemsService(c, data.MenuItems, category_uuid)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
			return
		}

		go newCache.ReloadFromDB()

		c.JSON(http.StatusOK, gin.H{
			"message": "menu items created successfully",
			"success": true,
		})
	}
}

func (h *FoodCategoryHandler) GetFoodCategoriesBySlug(c *gin.Context) {
	rawSlug := c.Query("slug")
	if rawSlug == "" {
		log.Println("not enough slug")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid slug", "success": false})
		return
	}
	res, err := h.foodCategoryService.GetFoodCategoriesBySlug(c, rawSlug)

	if err != nil {
		log.Println("error in getting data by slug : ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "got categories successfully",
		"data":    res,
		"success": true,
	})

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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload", "success": false})
		return
	}

	err := h.foodCategoryService.CreateCategoryService(c, data.CategoryName, data.SlugPath)
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
