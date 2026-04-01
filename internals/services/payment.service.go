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

// PaymentService defines all payment-related operations
type PaymentService interface {
	// CreatePaymentSer
	GetAllOrderDetailsForCAshierByOrderId(c *gin.Context, orderId uuid.UUID) (*models.PaymentDetailsForCashierWithDiscount, error)
	GetAllApprovedOrdersForCashierService(c *gin.Context) ([]models.GetOrderDetailsForCashier, error)
	CreatePaymentService(c *gin.Context, req *models.CreatePayment) (*models.Payment, error)
	DeleteOrderByCashier(c *gin.Context, orderId uuid.UUID) error
}

type paymentService struct {
	repo repository.PaymentRepo
}

func (s *paymentService) DeleteOrderByCashier(c *gin.Context, orderId uuid.UUID) error {
	_, err := lib.HasPermissionCheck(c, rbac.DeleteOrder)
	if err != nil {
		return errors.New("user not authorized")
	}

	return s.repo.DeleteOrderByCashier(c.Request.Context(), orderId)
}

func (s *paymentService) GetAllOrderDetailsForCAshierByOrderId(c *gin.Context, orderId uuid.UUID) (*models.PaymentDetailsForCashierWithDiscount, error) {
	_, err := lib.HasPermissionCheck(c, rbac.ViewPayments)
	if err != nil {
		return nil, errors.New("user not authorized")
	}

	return s.repo.GetAllOrderDetailsForCashierByOrderId(c.Request.Context(), orderId)
}

func (s *paymentService) GetAllApprovedOrdersForCashierService(c *gin.Context) ([]models.GetOrderDetailsForCashier, error) {
	_, err := lib.HasPermissionCheck(c, rbac.ViewPayments)
	if err != nil {
		return nil, errors.New("user not authorized")
	}

	return s.repo.GetAllApprovedOrdersForCashier(c.Request.Context())
}

// CreatePaymentService handles creating a new payment
func (s *paymentService) CreatePaymentService(c *gin.Context, req *models.CreatePayment) (*models.Payment, error) {
	_, err := lib.HasPermissionCheck(c, rbac.CreatePayments)
	if err != nil {
		return nil, errors.New("user not authorized")
	}

	return s.repo.CreatePayment(c.Request.Context(), req)
}

// NewPaymentService creates a new PaymentService instance
func NewPaymentService(repo repository.PaymentRepo) PaymentService {
	return &paymentService{
		repo: repo,
	}
}
