package routes

import (
	"github.com/Abhishekh669/backend/internals/app"
	"github.com/Abhishekh669/backend/internals/middlewares"
	"github.com/gin-gonic/gin"
)

func FoodCategoryServiceRouter(router *gin.RouterGroup, app *app.App) {
	foodCategoryServiceRoute := router.Group("/food-category-service")
	foodCategoryServiceRoute.POST("/create-category", middlewares.UserMiddleware(), app.FoodCategoryHandler.CreateFoodCategoryHandler)
	foodCategoryServiceRoute.GET("/get-all-categories", middlewares.UserMiddleware(), app.FoodCategoryHandler.GetFoodCategoriesHandlers)

}
