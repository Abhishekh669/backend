package routes

import (
	"github.com/Abhishekh669/backend/internals/app"
	"github.com/Abhishekh669/backend/internals/middlewares"
	"github.com/gin-gonic/gin"
)

func PaymentRouter(router *gin.RouterGroup, app *app.App) {
	paymentRoute := router.Group("/payment-service")
	paymentRoute.POST("/create", middlewares.UserMiddleware(), app.PaymentHandler.CreatePaymentHandler)
	paymentRoute.GET("/approved-orders", middlewares.UserMiddleware(), app.PaymentHandler.GetAllApprovedOrdersForCashierHandler)
}
