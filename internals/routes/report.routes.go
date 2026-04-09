package routes

import (
	"github.com/Abhishekh669/backend/internals/app"
	"github.com/Abhishekh669/backend/internals/middlewares"
	"github.com/gin-gonic/gin"
)

func ReportServiceRouter(router *gin.RouterGroup, app *app.App) {
	reportRoute := router.Group("/report-service")

	reportRoute.GET("/default", middlewares.UserMiddleware(), app.ReportHandler.GetDefaultRevenueReport)
	reportRoute.GET("/custom", middlewares.UserMiddleware(), app.ReportHandler.GetCustomRangeRevenueReport)

	// Custom range report (no cache, direct DB)
	reportRoute.GET("/sales-default", middlewares.UserMiddleware(), app.ReportHandler.GetDefaultSalesReport)
	reportRoute.GET("/sales-custom", middlewares.UserMiddleware(), app.ReportHandler.GetCustomRangeSalesReport)

	reportRoute.GET("/customer-default", middlewares.UserMiddleware(), app.ReportHandler.GetDefaultCustomerReport)
	reportRoute.GET("/customer-custom", middlewares.UserMiddleware(), app.ReportHandler.GetCustomRangeCustomerReport)

	reportRoute.GET("/table-default", middlewares.UserMiddleware(), app.ReportHandler.GetDefaultTableReport)
	reportRoute.GET("/table-custom", middlewares.UserMiddleware(), app.ReportHandler.GetCustomRangeTablesReport)

	// Cache management (admin only)
	reportRoute.POST("/default/refresh", middlewares.UserMiddleware(), app.ReportHandler.RefreshDefaultReportCache)
	reportRoute.GET("/default/status", middlewares.UserMiddleware(), app.ReportHandler.GetDefaultReportCacheStatus)
}
