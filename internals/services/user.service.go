package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/Abhishekh669/backend/internals/lib"
	"github.com/Abhishekh669/backend/internals/models"
	"github.com/Abhishekh669/backend/internals/rbac"
	"github.com/Abhishekh669/backend/internals/repository"
	"github.com/gin-gonic/gin"
)

type UserService interface {
	GetUserByNameService(c *gin.Context, userName *string) ([]models.UserTypeForAttendance, error)
	UpdateUserService(c *gin.Context, user *models.UpdateUserType) error
	DeleteUserService(c *gin.Context, userIds []string) error
	CreateNewUserService(c *gin.Context, user *models.CreateUserType) error
	GetUsersListService(c *gin.Context, search string, limit, offset int, oldFirstBool bool) (*repository.UserListResponse, error)
	GetUserByEmail(ctx context.Context, email string) (*models.UserType, error)
	LoginUserService(email, password string, ctx context.Context) (string, error)
}

type userService struct {
	repo repository.UserRepo
}

func (s *userService) GetUserByNameService(c *gin.Context, userName *string) ([]models.UserTypeForAttendance, error) {
	_, err := lib.HasPermissionCheck(c, rbac.ViewUsers)
	if err != nil {
		return nil, errors.New("user not authorized")
	}

	return s.repo.GetUserDataByName(c.Request.Context(), *userName)
}

func (s *userService) UpdateUserService(c *gin.Context, user *models.UpdateUserType) error {
	requesterRole, err := lib.HasPermissionCheck(c, rbac.UpdateUsers)
	if err != nil {
		return err
	}
	existingUser, err := s.repo.GetUserById(user.Id, c.Request.Context())
	if err != nil {
		return errors.New("user not found")
	}

	currentTargetRole := existingUser.Role
	newTargetRole := user.Role

	// 🚫 Only Admin can modify Admin accounts
	if currentTargetRole == models.RoleAdmin && *requesterRole != models.RoleAdmin {
		return errors.New("not authorized to modify admin")
	}

	// 🚫 Only Admin can assign Admin role
	if newTargetRole == models.RoleAdmin && *requesterRole != models.RoleAdmin {
		return errors.New("not authorized to promote user to admin")
	}

	// 🚫 Manager cannot modify another Manager
	if newTargetRole != currentTargetRole && *requesterRole != models.RoleAdmin {
		return errors.New("not authorized to change user role")
	}

	return s.repo.UpdateUser(c.Request.Context(), user)
}

func (s *userService) DeleteUserService(c *gin.Context, userIds []string) error {
	role, err := lib.HasPermissionCheck(c, rbac.DeleteUsers)
	if err != nil {
		return err
	}
	return s.repo.DeleteUser(c.Request.Context(), userIds, *role)
}

func (s *userService) CreateNewUserService(c *gin.Context, user *models.CreateUserType) error {
	_, err := lib.HasPermissionCheck(c, rbac.CreateUsers)
	if err != nil {
		return err
	}
	return s.repo.CreateNewUser(c.Request.Context(), user)
}

func (s *userService) GetUsersListService(c *gin.Context, search string, limit, offset int, oldFirstBool bool) (*repository.UserListResponse, error) {
	_, err := lib.HasPermissionCheck(c, rbac.ViewUsers)
	if err != nil {
		return nil, err
	}

	return s.repo.GetAllUsers(c.Request.Context(), search, limit, offset, oldFirstBool)
}

func (s *userService) GetUserByEmail(ctx context.Context, email string) (*models.UserType, error) {
	return s.repo.GetUserInfoByEmail(email, ctx)
}

func (s *userService) LoginUserService(
	email, password string,
	ctx context.Context,
) (string, error) {

	user, err := s.repo.LoginUser(email, password, ctx)
	if err != nil || user == nil {
		return "", errors.New("incorrect credentials")
	}

	jwtData := lib.JwtDataType{
		UserId:              user.Id,
		Email:               user.Email,
		LastPasswordResetAt: user.LastPasswordResetAt,
	}

	jwtToken, err := lib.GenerateJWTToken(&jwtData)
	if err != nil {
		return "", fmt.Errorf("failed to generate token")
	}

	return jwtToken, nil
}

func NewUserService(repo repository.UserRepo) UserService {
	return &userService{
		repo: repo,
	}
}
