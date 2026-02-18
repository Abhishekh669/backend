package routes

import (
	"github.com/Abhishekh669/backend/internals/app"
	"github.com/gin-gonic/gin"
)

func SetUpRoutes(app *gin.Engine, appConfig *app.App) {
	apiGroup := app.Group("/api/v1")
	UserServiceRouter(apiGroup, appConfig)
	RawMaterialServiceRouter(apiGroup, appConfig)
	FoodCategoryServiceRouter(apiGroup, appConfig)
	AttendanceServiceRouter(apiGroup, appConfig)
}
