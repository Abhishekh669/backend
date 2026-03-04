package routes

import (
	"github.com/Abhishekh669/backend/internals/app"
	"github.com/Abhishekh669/backend/internals/middlewares"
	"github.com/gin-gonic/gin"
)

func TableRouter(router *gin.RouterGroup, app *app.App) {
	tableRouterService := router.Group("/table-service")

	tableRouterService.GET("/get-tables", app.TableHandler.GetTablesHandler)
	tableRouterService.POST("/create-tables", middlewares.UserMiddleware(), app.TableHandler.CreateTablesHandler)
	tableRouterService.PUT("/update-table", middlewares.UserMiddleware(), app.TableHandler.UpdateTableHandler)
	tableRouterService.POST("/delete-tables", middlewares.UserMiddleware(), app.TableHandler.DeleteTablesHandler)
}
