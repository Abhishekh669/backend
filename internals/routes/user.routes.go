package routes

import (
	"github.com/Abhishekh669/backend/internals/app"
	"github.com/Abhishekh669/backend/internals/middlewares"
	"github.com/gin-gonic/gin"
)

func UserServiceRouter(router *gin.RouterGroup, app *app.App) {
	userServiceRoute := router.Group("/user-service")
	userServiceRoute.POST("/login-user", app.UserHandler.LoginUserHandler)
	userServiceRoute.POST("/create-new-user", middlewares.UserMiddleware(), app.UserHandler.CreateNewUser)
	userServiceRoute.POST("/delete-user", middlewares.UserMiddleware(), app.UserHandler.DeleteUserHandler)
	userServiceRoute.GET("/get-user-from-token", app.UserHandler.GetUserFromTokenHandler)
	userServiceRoute.GET("/get-all-users", middlewares.UserMiddleware(), app.UserHandler.GetUsersListHandler)
	userServiceRoute.PUT("/update-user", middlewares.UserMiddleware(), app.UserHandler.UpdateUserHandler)
	userServiceRoute.GET("/get-users-by-name", middlewares.UserMiddleware(), app.UserHandler.GetUserByNameHandler)
	userServiceRoute.GET("/get-feedbacks", app.UserHandler.GetCustomerFeedBacksHandler)

	userServiceRoute.PUT("/update-user-password", middlewares.UserMiddleware(), app.UserHandler.UpdateUserPasswordHandler)
	userServiceRoute.GET("/get-forget-password-session", app.UserHandler.GetForgetPasswordSessionHandler)
	userServiceRoute.POST("/create-forget-password-session", app.UserHandler.CreateForgetPasswordSessionHandler)
	userServiceRoute.POST("/check-forget-password-pin", app.UserHandler.CheckForgetPasswordPinHandler)
	userServiceRoute.POST("/create-feedback", app.UserHandler.CreateFeedBackHandler)

}
