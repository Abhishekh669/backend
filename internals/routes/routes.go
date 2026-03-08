package routes

import (
	"github.com/Abhishekh669/backend/internals/algorithm"
	"github.com/Abhishekh669/backend/internals/app"
	"github.com/gin-gonic/gin"
)

func SetUpRoutes(app *gin.Engine, appConfig *app.App, newCache *algorithm.MenuCache) {
	apiGroup := app.Group("/api/v1")
	UserServiceRouter(apiGroup, appConfig)
	RawMaterialServiceRouter(apiGroup, appConfig)
	FoodCategoryServiceRouter(apiGroup, appConfig, newCache)
	AttendanceServiceRouter(apiGroup, appConfig)
	TableRouter(apiGroup, appConfig)
	OrderServiceRouter(apiGroup, appConfig)
}
