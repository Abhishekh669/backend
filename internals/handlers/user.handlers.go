package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Abhishekh669/backend/internals/lib"
	"github.com/Abhishekh669/backend/internals/models"
	"github.com/Abhishekh669/backend/internals/services"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService services.UserService
}

var (
	maxLimit      = 20
	defaultOffset = 0
)

type UserPassUpdateType struct {
	NewPassword string `json:"new_password" binding:"required"`
	OldPassword string `json:"old_password" binding:"required"`
}

type CreateSessionType struct {
	Email string `json:"email" binding:"required,email"`
}

type CheckPinType struct {
	Email       string `json:"email" binding:"required,email"`
	Token       string `json:"token" binding:"required"`
	Pin         string `json:"pin" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

func (h *UserHandler) CreateFeedBackHandler(c *gin.Context) {

	var data models.CreateCustomerFeedback
	if err := c.ShouldBindJSON(&data); err != nil {
		fmt.Println("error in binding", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}

	err := h.userService.CreateCustomerFeedBack(c, &data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to create feedback", "success": false})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "successfully submitted"})
}

func (h *UserHandler) GetCustomerFeedBacksHandler(c *gin.Context) {

	session, err := h.userService.GetCustomerService(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to get session", "success": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{"feedbacks": session, "success": true})
}

func (h *UserHandler) GetForgetPasswordSessionHandler(c *gin.Context) {
	token := c.Query("token")
	email := c.Query("email")

	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email is required", "success": false})
		return
	}
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token is required", "success": false})
		return
	}

	session, err := h.userService.GetForgetPasswordSession(c, token, email)
	fmt.Println("thisis forget passwor4d sesiosn : ", session, err)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{"session": session, "success": true})
}

func (h *UserHandler) CreateForgetPasswordSessionHandler(c *gin.Context) {

	var data CreateSessionType
	if err := c.ShouldBindJSON(&data); err != nil {
		fmt.Println("error in binding", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}

	if data.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email is required", "success": false})
		return
	}

	token, err := h.userService.CreateForgetPasswordSession(c, data.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token, "success": true})
}

func (h *UserHandler) CheckForgetPasswordPinHandler(c *gin.Context) {
	var data CheckPinType
	if err := c.ShouldBindJSON(&data); err != nil {
		fmt.Println("error in binding", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}

	if data.Email == "" || data.Token == "" || data.Pin == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email, token and pin are required", "success": false})
		return
	}

	valid, err := h.userService.CheckForgetPasswordPin(c, data.Pin, data.Token, data.Email, data.NewPassword)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}

	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid token or pin", "success": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "pin verified successfully", "success": true})
}

func (h *UserHandler) UpdateUserPasswordHandler(c *gin.Context) {
	var data UserPassUpdateType
	if err := c.ShouldBindJSON(&data); err != nil {
		fmt.Println("error in binding", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}

	err := h.userService.UpdateUserPassword(c, data.NewPassword, data.OldPassword)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "password updated successfully",
		"success": true,
	})
}

func (h *UserHandler) GetUserByNameHandler(c *gin.Context) {
	userName := c.Query("userName")
	if userName == "" {
		c.JSON(http.StatusOK, gin.H{"success": true, "users": nil})
		return
	}
	users, err := h.userService.GetUserByNameService(c, &userName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users":   users,
		"success": true,
	})
}

func (h *UserHandler) UpdateUserHandler(c *gin.Context) {
	var data models.UpdateUserType
	if err := c.ShouldBindJSON(&data); err != nil {
		fmt.Println("error in binding", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}

	err := h.userService.UpdateUserService(c, &data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "user updated successfully",
		"success": true,
	})

}

func (h *UserHandler) DeleteUserHandler(c *gin.Context) {
	var data models.DeleteUserPayload

	if err := c.ShouldBindJSON(&data); err != nil {
		fmt.Println("error in binding", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}

	if len(data.UserIds) < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no user selected", "success": false})
		return
	}

	err := h.userService.DeleteUserService(c, data.UserIds)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}
	var message string
	if len(data.UserIds) > 0 {
		message = fmt.Sprintf(" %d users deleted successfully", len(data.UserIds))
	} else {
		message = "user deleted successfully"
	}
	c.JSON(http.StatusOK, gin.H{"message": message, "success": true})

}

func (h *UserHandler) CreateNewUser(c *gin.Context) {
	var user models.CreateUserType

	if err := c.ShouldBindJSON(&user); err != nil {
		fmt.Println("error in binding", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}
	if user.Image != nil && *user.Image == "" {
		user.Image = nil
	}

	fmt.Println("this is user to create : ", user)

	err := h.userService.CreateNewUserService(c, &user)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "success": false})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": " User created successfully", "success": true})
}

func (h *UserHandler) GetUsersListHandler(c *gin.Context) {
	search := c.Query("search")
	limit := c.DefaultQuery("limit", "20")
	offset := c.DefaultQuery("offset", "0")
	oldestFirst := c.DefaultQuery("oldestFirst", "false")
	oldestFirstBool := false
	if strings.ToLower(oldestFirst) == "true" {
		oldestFirstBool = true
	}
	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		limitInt = maxLimit
	}

	offsetInt, err := strconv.Atoi(offset)
	if err != nil {
		offsetInt = defaultOffset
	}

	if limitInt > maxLimit {
		limitInt = maxLimit
	}

	paginatedData, err := h.userService.GetUsersListService(c, strings.TrimSpace(search), limitInt, offsetInt, oldestFirstBool)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "success": false})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": paginatedData, "success": true})
}

func ConvertUserTypeToSafeUserType(user *models.UserType) *models.SafeUserType {
	if user == nil {
		return nil
	}
	return &models.SafeUserType{
		Id:        user.Id,
		Email:     user.Email,
		Gender:    user.Gender,
		Image:     user.Image,
		IsActive:  user.IsActive,
		Role:      user.Role,
		Name:      user.Name,
		Phone:     user.Phone,
		Salary:    user.Salary,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

func (h *UserHandler) GetUserFromTokenHandler(c *gin.Context) {
	token, err := lib.ExtractTokenFromHeader(c)
	if err != nil || token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized user",
			"success": false,
		})
		return
	}

	claims, err := lib.ParseJwtToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized user",
			"success": false,
		})
		return
	}

	user, err := h.userService.GetUserByEmail(c.Request.Context(), claims.Email)
	if err != nil || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized user",
			"success": false,
		})
		return
	}

	safeUser := ConvertUserTypeToSafeUserType(user)

	if safeUser == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal server error",
			"success": false,
		})
		return
	}

	if user.LastPasswordResetAt != claims.LastPasswordResetAt {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "invalid session",
			"success": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "user verified successfully",
		"success": true,
		"user":    safeUser,
	})

}

func (h *UserHandler) LoginUserHandler(c *gin.Context) {
	var loginData models.UserLogin

	if err := c.ShouldBindJSON(&loginData); err != nil {
		fmt.Println("Error binding JSON in registration :", err)
		c.JSON(400, gin.H{"error": "Invalid request payload", "success": false})
		return
	}

	if loginData.Email == "" || loginData.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "incorrect credentials",
			"success": false,
		})
		return
	}

	token, user, err := h.userService.LoginUserService(loginData.Email, loginData.Password, c.Request.Context())
	if err != nil || user == nil {
		fmt.Println("Error in login service:", err)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	if token == "" {
		fmt.Println("Error in login service:", err)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "incorrect credentials",
			"success": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":   token,
		"success": true,
		"user":    ConvertUserTypeToSafeUserType(user),
		"message": "login successfully",
	})
}

func NewUserHandler(userService services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}
