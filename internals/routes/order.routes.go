package routes

import (
	"github.com/Abhishekh669/backend/internals/app"
	"github.com/Abhishekh669/backend/internals/middlewares"
	"github.com/gin-gonic/gin"
)

func OrderServiceRouter(router *gin.RouterGroup, app *app.App) {
	orderServiceRoute := router.Group("/order-service")
	orderServiceRoute.POST("/create-order", app.Orderhandler.CreateCustomerHandler)
	orderServiceRoute.POST("/approve-order", middlewares.UserMiddleware())
	orderServiceRoute.GET("/get-order-requests", app.Orderhandler.GetAllOrderRequestHandler)
	orderServiceRoute.GET("/get-request-by-table-num-n-phone", app.Orderhandler.GetOrderRequestByTableNumberNPhone)
	orderServiceRoute.GET("/get-request-by-table-session-id/:table-session-id", middlewares.UserMiddleware(), app.Orderhandler.GetOrderRequestByTableSessionIdHandler)
}
