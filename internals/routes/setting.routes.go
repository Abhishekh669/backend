package routes

import (
	"github.com/Abhishekh669/backend/internals/app"
	"github.com/Abhishekh669/backend/internals/middlewares"
	"github.com/gin-gonic/gin"
)

func SettingServiceRouter(router *gin.RouterGroup, app *app.App) {
	// Create restaurant information
	settingRoute := router.Group("/setting-service")

	// Update restaurant information
	settingRoute.PUT("/restaurant-information", middlewares.UserMiddleware(), app.SettingHandler.UpdateRestaurantInformationHandler)
	// Get restaurant information
	settingRoute.GET("/restaurant-information", middlewares.UserMiddleware(), app.SettingHandler.GetRestaurantInformationHandler)
}
