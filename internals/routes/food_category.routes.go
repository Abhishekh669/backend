package routes

import (
	"github.com/Abhishekh669/backend/internals/app"
	"github.com/Abhishekh669/backend/internals/middlewares"
	"github.com/gin-gonic/gin"
)

func FoodCategoryServiceRouter(router *gin.RouterGroup, app *app.App) {
	foodCategoryServiceRoute := router.Group("/food-category-service")
	foodCategoryServiceRoute.POST("/create-category", middlewares.UserMiddleware(), app.FoodCategoryHandler.CreateFoodCategoryHandler)
	foodCategoryServiceRoute.POST("/create-menu-items", middlewares.UserMiddleware(), app.FoodCategoryHandler.CreateMenuItemsHandlers)
	foodCategoryServiceRoute.GET("/get-all-categories", middlewares.UserMiddleware(), app.FoodCategoryHandler.GetFoodCategoriesHandlers)
	foodCategoryServiceRoute.GET("/get-food-by-slug/", middlewares.UserMiddleware(), app.FoodCategoryHandler.GetFoodCategoriesBySlug)
	foodCategoryServiceRoute.PUT("/update-category", middlewares.UserMiddleware(), app.FoodCategoryHandler.UpdateCategoryHandler)
	foodCategoryServiceRoute.PUT("/update-menu-items", middlewares.UserMiddleware(), app.FoodCategoryHandler.UpdateMenuItemHandler)
	foodCategoryServiceRoute.POST("/delete-menu-items", middlewares.UserMiddleware(), app.FoodCategoryHandler.DeleteMenuItemsHandler)
	foodCategoryServiceRoute.POST("/delete-categories", middlewares.UserMiddleware(), app.FoodCategoryHandler.DeleteCategoriesHandler)
}
