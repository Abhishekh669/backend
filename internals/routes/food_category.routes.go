package routes

import (
	"github.com/Abhishekh669/backend/internals/algorithm"
	"github.com/Abhishekh669/backend/internals/app"
	"github.com/Abhishekh669/backend/internals/middlewares"
	"github.com/gin-gonic/gin"
)

func FoodCategoryServiceRouter(router *gin.RouterGroup, app *app.App, cache *algorithm.MenuCache) {
	foodCategoryServiceRoute := router.Group("/food-category-service")
	foodCategoryServiceRoute.POST("/create-category", middlewares.UserMiddleware(), app.FoodCategoryHandler.CreateFoodCategoryHandler)
	foodCategoryServiceRoute.POST("/create-menu-items", middlewares.UserMiddleware(), app.FoodCategoryHandler.CreateMenuItemsHandler(cache))
	foodCategoryServiceRoute.GET("/get-all-categories", middlewares.UserMiddleware(), app.FoodCategoryHandler.GetFoodCategoriesHandlers)
	foodCategoryServiceRoute.GET("/get-all-menu-items", app.FoodCategoryHandler.GetAllMenuItemsGroupedHander)
	foodCategoryServiceRoute.GET("/get-food-by-slug/", middlewares.UserMiddleware(), app.FoodCategoryHandler.GetMenuItemsBySlug)
	foodCategoryServiceRoute.PUT("/update-category", middlewares.UserMiddleware(), app.FoodCategoryHandler.UpdateCategoryHandler(cache))
	foodCategoryServiceRoute.PUT("/update-menu-item", middlewares.UserMiddleware(), app.FoodCategoryHandler.UpdateMenuItemHandler(cache))
	foodCategoryServiceRoute.POST("/delete-menu-items", middlewares.UserMiddleware(), app.FoodCategoryHandler.DeleteMenuItemsHandler(cache))
	foodCategoryServiceRoute.POST("/delete-categories", middlewares.UserMiddleware(), app.FoodCategoryHandler.DeleteCategoriesHandler(cache))
}
