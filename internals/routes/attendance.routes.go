package routes

import (
	"github.com/Abhishekh669/backend/internals/app"
	"github.com/Abhishekh669/backend/internals/middlewares"
	"github.com/gin-gonic/gin"
)

func AttendanceServiceRouter(router *gin.RouterGroup, app *app.App) {
	attendanceServiceRoute := router.Group("/attendance-service")
	attendanceServiceRoute.POST("/check-in", middlewares.UserMiddleware(), app.AttendanceHandler.CheckInHandler)
	attendanceServiceRoute.PUT("/check-out", middlewares.UserMiddleware(), app.AttendanceHandler.CheckOutHandler)
	attendanceServiceRoute.GET("/current", middlewares.UserMiddleware(), app.AttendanceHandler.GetCurrentAttendanceHandler)
}
