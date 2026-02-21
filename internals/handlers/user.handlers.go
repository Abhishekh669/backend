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

	token, err := h.userService.LoginUserService(loginData.Email, loginData.Password, c.Request.Context())
	if err != nil {
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
		"message": "login successfully",
	})
}

func NewUserHandler(userService services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}
