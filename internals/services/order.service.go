package services

import (
	"errors"

	"github.com/Abhishekh669/backend/internals/lib"
	"github.com/Abhishekh669/backend/internals/models"
	"github.com/Abhishekh669/backend/internals/rbac"
	"github.com/Abhishekh669/backend/internals/repository"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
)

type OrderService interface {
	GetAllOrderStatusService(c *gin.Context) ([]models.CustomerOrderRequest, error)
	GetOrderReqeustFromTableNumberAndPhoneNumber(c *gin.Context, tableNumber int, phoneNumber string) (*models.CustomerOrderRequest, error)
	GetOrderRequestFromTableSession(c *gin.Context, tableSessionId uuid.UUID) (*models.CustomerOrderRequest, error)
	GetAllOrderRequests(c *gin.Context) ([]models.CustomerOrderRequest, error)
	ApproveCustomerOrder(c *gin.Context, approveOrder *models.ApproveOrderType) error
	CreateCustomerService(c *gin.Context, cusotmerOrder *models.CreateCustomerOrderRequest) error
}
type orderService struct {
	repo repository.OrderRepo
}

func (s *orderService) GetAllOrderStatusService(c *gin.Context) ([]models.CustomerOrderRequest, error) {
	_, err := lib.HasPermissionCheck(c, rbac.ViewOrder)
	if err != nil {
		return nil, errors.New("user not authorized")
	}
	return s.repo.NewGetAllOrderForStatus(c.Request.Context())
}

func (s *orderService) GetOrderReqeustFromTableNumberAndPhoneNumber(c *gin.Context, tableNumber int, phoneNumber string) (*models.CustomerOrderRequest, error) {
	return s.repo.NewGetTableSessionByTableAndPhone(c.Request.Context(), tableNumber, phoneNumber)
}

func (s *orderService) GetOrderRequestFromTableSession(c *gin.Context, tableSessionId uuid.UUID) (*models.CustomerOrderRequest, error) {
	_, err := lib.HasPermissionCheck(c, rbac.ViewOrder)
	if err != nil {
		return nil, errors.New("user not authorized")
	}

	return s.repo.NewGetTableSessionByID(c.Request.Context(), tableSessionId)
}

func (s *orderService) GetAllOrderRequests(c *gin.Context) ([]models.CustomerOrderRequest, error) {
	return s.repo.NewGetAllOrderRequest(c.Request.Context())
}

func (s *orderService) ApproveCustomerOrder(c *gin.Context, approveOrder *models.ApproveOrderType) error {
	_, err := lib.HasPermissionCheck(c, rbac.CreateOrder)
	if err != nil {
		return errors.New("user not authorized")
	}
	return s.repo.NewApproveCustomerRequest(c.Request.Context(), approveOrder)
}

func (s *orderService) CreateCustomerService(c *gin.Context, cusotmerOrder *models.CreateCustomerOrderRequest) error {
	return s.repo.NewCreateCustomerOrder(c.Request.Context(), cusotmerOrder)
}

func NewOrderService(repo repository.OrderRepo) OrderService {
	return &orderService{
		repo: repo,
	}
}
