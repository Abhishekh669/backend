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
	attendanceServiceRoute.GET("/history", middlewares.UserMiddleware(), app.AttendanceHandler.GetAttendanceHistory)
	attendanceServiceRoute.PUT("/update", middlewares.UserMiddleware(), app.AttendanceHandler.UpdateAttendanceHandler)
	attendanceServiceRoute.DELETE("/delete/:id", middlewares.UserMiddleware(), app.AttendanceHandler.DeleteAttendanceByIdHandler)

	attendanceServiceRoute.POST("/leave", middlewares.UserMiddleware(), app.AttendanceHandler.CreateEmployeeRequest)

	attendanceServiceRoute.PUT("/leave", middlewares.UserMiddleware(), app.AttendanceHandler.UpdateLeaveRequest)

	attendanceServiceRoute.DELETE("/leave/:id", middlewares.UserMiddleware(), app.AttendanceHandler.DeleteLeaveRequest)

	attendanceServiceRoute.PUT("/leave/:id/cancel", middlewares.UserMiddleware(), app.AttendanceHandler.CancelLeaveRequest)

	attendanceServiceRoute.GET("/get-today-attendance", middlewares.UserMiddleware(), app.AttendanceHandler.GetTodayAttendanceHandler)

	attendanceServiceRoute.PUT("/update-user-leave", middlewares.UserMiddleware(), app.AttendanceHandler.UpdateUserLeaveRequest)
}
