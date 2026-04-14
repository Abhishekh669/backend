package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Abhishekh669/backend/internals/lib"
	"github.com/Abhishekh669/backend/internals/models"
	"github.com/Abhishekh669/backend/internals/rbac"
	"github.com/Abhishekh669/backend/internals/repository"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
)

type UserService interface {
	CreateForgetPasswordSession(ctx *gin.Context, email string) (string, error)
	CheckForgetPasswordPin(ctx *gin.Context, pin, token, email string, newPassword string) (bool, error)
	GetForgetPasswordSession(ctx *gin.Context, email, token string) (*models.PasswordResetRequest, error)
	UpdateUserPassword(ctx *gin.Context, newPassword, oldPassword string) error
	GetUserByNameService(c *gin.Context, userName *string) ([]models.UserTypeForAttendance, error)
	UpdateUserService(c *gin.Context, user *models.UpdateUserType) error
	DeleteUserService(c *gin.Context, userIds []string) error
	CreateNewUserService(c *gin.Context, user *models.CreateUserType) error
	GetUsersListService(c *gin.Context, search string, limit, offset int, oldFirstBool bool) (*repository.UserListResponse, error)
	GetUserByEmail(ctx context.Context, email string) (*models.UserType, error)
	LoginUserService(email, password string, ctx context.Context) (string, *models.UserType, error)
}

type userService struct {
	repo repository.UserRepo
}

func (s *userService) CheckForgetPasswordPin(ctx *gin.Context, pin, token, email, newPassword string) (bool, error) {
	user, err := s.repo.GetUserInfoByEmail(email, ctx.Request.Context())
	if err != nil {
		return false, errors.New("user lookup failed")
	}

	session, err := s.repo.GetForgetPasswordSession(ctx, email, token)
	if err != nil {
		return false, errors.New("session not found")
	}
	if session.IsUsed || session.UpdatedAt.Add(15*time.Minute).Before(time.Now()) {
		return false, errors.New("session already used or expired")
	}
	if session.PinCode != pin {
		return false, errors.New("incorrect pin")
	}

	if user == nil {
		return false, errors.New("user not found")
	}
	userUUId, err := uuid.FromString(user.Id)
	if err != nil {
		return false, errors.New("failed to check pin")
	}

	err = s.repo.MarkForgetPasswordSessionUsed(ctx.Request.Context(), session.ID)
	if err != nil {
		return false, errors.New("failed to mark session used")
	}
	hashedPassword, err := lib.HashPassword(newPassword)
	if err != nil {
		return false, errors.New("failed to hash new password")
	}
	err = s.repo.UpdateUserPassword(ctx.Request.Context(), userUUId, hashedPassword)
	if err != nil {
		return false, errors.New("failed to update password")
	}
	return true, nil
}

func (s *userService) GetForgetPasswordSession(ctx *gin.Context, token, email string) (*models.PasswordResetRequest, error) {

	session, err := s.repo.GetForgetPasswordSession(ctx, email, token)
	if err != nil || session == nil {
		return nil, errors.New("session not found")
	}
	if session.IsUsed || session.UpdatedAt.Add(15*time.Minute).Before(time.Now()) {
		return nil, errors.New("session already used")
	}
	return session, nil
}

func (s *userService) CreateForgetPasswordSession(ctx *gin.Context, email string) (string, error) {

	user, err := s.repo.GetUserInfoByEmail(email, ctx.Request.Context())
	if err != nil || user == nil {
		return "", errors.New("user with given email not found")
	}

	sessisionData, err := s.repo.GetForgetPasswordSessionByEmail(ctx, email)
	fmt.Println("this is the e exisitng sesiosn data : ", sessisionData)
	if sessisionData != nil &&
		!sessisionData.IsUsed {
		token, err := lib.GenerateToken()
		if err != nil {
			return "", errors.New("failed to generate token")
		}

		pin, err := lib.Generate6DigitCode()
		if err != nil {
			return "", errors.New("failed to generate pin")
		}

		err = s.repo.UpdateExistingPasswordSession(ctx.Request.Context(), sessisionData.ID, token, pin)
		if err != nil {
			return "", errors.New("failed to update password reset session")
		}

		err = lib.SendForgetPasswordTokenToEmail(email, pin)
		if err != nil {
			log.Println("failed to send email with reset token", err)
		}
		return token, nil

	}

	token, err := lib.GenerateToken()
	if err != nil {
		return "", errors.New("failed to generate token")
	}

	pin, err := lib.Generate6DigitCode()
	if err != nil {
		return "", errors.New("failed to generate pin")
	}

	err = s.repo.CreateForgetPasswordSession(ctx.Request.Context(), email, token, pin)
	if err != nil {
		return "", errors.New("failed to create password reset session")
	}

	err = lib.SendForgetPasswordTokenToEmail(email, pin)
	if err != nil {
		log.Println("failed to send email with reset token ", err)
	}

	return token, nil
}

func (s *userService) UpdateUserPassword(ctx *gin.Context, newPassword, oldPassword string) error {
	userId, err := GetUserIDFromContext(ctx)
	if err != nil {
		return errors.New("user not authenticated")
	}
	strUserId := userId.String()
	userData, err := s.repo.GetUserById(strUserId, ctx.Request.Context())
	if err != nil {
		return errors.New("user not found")
	}
	fmt.Println("this is the user data : ", userData)
	fmt.Println("thisis new passowrd : ", newPassword, " and this is old password : ", oldPassword)
	correct_pass, err := lib.CheckPasswordHash(oldPassword, userData.Password)
	if !correct_pass || err != nil {
		return errors.New("incorrect old password")
	}
	hashNewPassword, err := lib.HashPassword(newPassword)
	if err != nil {
		return errors.New("failed to hash new password")
	}
	return s.repo.UpdateUserPassword(ctx.Request.Context(), userId, hashNewPassword)

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
) (string, *models.UserType, error) {

	user, err := s.repo.LoginUser(email, password, ctx)
	if err != nil || user == nil {
		return "", nil, errors.New("incorrect credentials")
	}

	jwtData := lib.JwtDataType{
		UserId:              user.Id,
		Email:               user.Email,
		LastPasswordResetAt: user.LastPasswordResetAt,
	}

	jwtToken, err := lib.GenerateJWTToken(&jwtData)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate token")
	}

	return jwtToken, user, nil
}

func NewUserService(repo repository.UserRepo) UserService {
	return &userService{
		repo: repo,
	}
}
