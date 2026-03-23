package routes

import (
	"github.com/Abhishekh669/backend/internals/app"
	"github.com/Abhishekh669/backend/internals/middlewares"
	"github.com/gin-gonic/gin"
)

func OrderServiceRouter(router *gin.RouterGroup, app *app.App) {
	orderServiceRoute := router.Group("/order-service")
	orderServiceRoute.GET("/get-orders-status", middlewares.UserMiddleware(), app.Orderhandler.GetAllOrderStatusHandler)
	orderServiceRoute.POST("/create-order", middlewares.CustomerMiddleware(app), app.Orderhandler.CreateCustomerHandler)
	orderServiceRoute.POST("/approve-order", middlewares.UserMiddleware(), app.Orderhandler.ApproveCustomerOrderHandler)
	orderServiceRoute.GET("/get-order-requests", app.Orderhandler.GetAllOrderRequestHandler)
	orderServiceRoute.GET("/get-request-by-table-num-n-phone", middlewares.CustomerMiddleware(app), app.Orderhandler.GetOrderRequestByTableNumberNPhone)
	orderServiceRoute.GET("/get-request-by-table-session-id/:table-session-id", middlewares.UserMiddleware(), app.Orderhandler.GetOrderRequestByTableSessionIdHandler)

	orderServiceRoute.POST("/table-approval", app.Orderhandler.CreateNewApprovalRequestHandler)
	orderServiceRoute.PUT("/table-approve-by-waiter", middlewares.UserMiddleware(), app.Orderhandler.ApproveTableByWaiterHandler)
	orderServiceRoute.DELETE("/table-delete/:id", middlewares.UserMiddleware(), app.Orderhandler.DeleteTableValidationHandler)
	orderServiceRoute.DELETE("/table-session-delete/:id", middlewares.UserMiddleware(), app.Orderhandler.DeleteTableSessionByIdHandler)
	orderServiceRoute.GET("/tables-unassigned", middlewares.UserMiddleware(), app.Orderhandler.GetUnassignedTablesHandler)
	orderServiceRoute.GET("/get-table-validation-by-id/:id", app.Orderhandler.GetTableValidationByIDHandler)
	orderServiceRoute.GET("/get-table-validation-by-phone-n-number", app.Orderhandler.GetTableValidationByPhoneAndTableHandler)
	orderServiceRoute.GET("/get-table-validation-from-token", middlewares.CustomerMiddleware(app), app.Orderhandler.GetTableValidationFromTokenHandler)

	//mobile seciton
	orderServiceRoute.PUT("/update-order-item", middlewares.UserMiddleware(), app.Orderhandler.UpdateOrderItemHandler)
	orderServiceRoute.GET("/get-all-approval-requests", middlewares.UserMiddleware(), app.Orderhandler.GetAllApprovalRequestHandler)

}
