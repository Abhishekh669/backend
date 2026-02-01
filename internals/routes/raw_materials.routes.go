package routes

import (
	"github.com/Abhishekh669/backend/internals/app"
	"github.com/Abhishekh669/backend/internals/middlewares"
	"github.com/gin-gonic/gin"
)

func RawMaterialServiceRouter(router *gin.RouterGroup, app *app.App) {
	rawMaterialServiceRoute := router.Group("/raw-material-service")
	rawMaterialServiceRoute.POST("/create-raw-materials", middlewares.UserMiddleware(), app.RawMaterialHandler.CreateRawMaterialHandlers)
	rawMaterialServiceRoute.POST("/delete-raw-materials", middlewares.UserMiddleware(), app.RawMaterialHandler.DeleteRawMaterialsHandler)
	rawMaterialServiceRoute.GET("/get-raw-materials", middlewares.UserMiddleware(), app.RawMaterialHandler.GetRawMaterialsHandlers)
	rawMaterialServiceRoute.PUT("/update-raw-material", middlewares.UserMiddleware(), app.RawMaterialHandler.UpdateRawMaterialsHandler)

}
