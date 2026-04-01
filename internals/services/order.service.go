package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Abhishekh669/backend/internals/lib"
	"github.com/Abhishekh669/backend/internals/models"
	"github.com/Abhishekh669/backend/internals/rbac"
	"github.com/Abhishekh669/backend/internals/repository"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
)

type OrderService interface {
	GetAllOrderHistoryForAdmin(c *gin.Context, limit, page int, fromDate, toDate *time.Time) (*repository.OrderHistoryResponse, error)
	GetAllApprovalRequestService(c *gin.Context) ([]models.TableValidation, error)
	UpdateOrderItemService(c *gin.Context, updateData *models.UpdateOrderItem) error
	DeleteTableSessionByIdService(c *gin.Context, id *uuid.UUID, tableNumber int, phoneNumber string) error
	GetTableValidationByPhoneNTable(ctx context.Context, phone string, tableNumber int) (string, ReqStatus, error)
	GetTableValidationById(ctx context.Context, id uuid.UUID) (*models.TableValidation, error)
	CreateNewApprovalRequestService(c *gin.Context, req *models.CustomerApprovalRequest) (*models.TableValidation, error)
	ApproveTableByWaiterService(c *gin.Context, req *models.WaiterApprovalRequest) error
	DeleteTableValidationService(c *gin.Context, id uuid.UUID) error
	GetUnassignedTablesService(c *gin.Context) ([]models.TableValidation, error)
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

type ReqStatus string

const (
	OrderNotFound    ReqStatus = "not_found"
	OrderNotApproved ReqStatus = "not_approved"
	OrderApproved    ReqStatus = "approved"
)

func (s *orderService) GetAllOrderHistoryForAdmin(c *gin.Context, limit, page int, fromDate, toDate *time.Time) (*repository.OrderHistoryResponse, error) {
	_, err := lib.HasPermissionCheck(c, rbac.ViewOrder)
	if err != nil {
		return nil, errors.New("user not authorized")
	}
	return s.repo.GetAllOrderHistoryForAdmin(c.Request.Context(), limit, page, fromDate, toDate)
}

func (s *orderService) GetAllApprovalRequestService(c *gin.Context) ([]models.TableValidation, error) {
	_, err := lib.HasPermissionCheck(c, rbac.ViewOrder)
	if err != nil {
		return nil, errors.New("user not authorized")
	}
	return s.repo.GetAllOrderApprovalRequest(c.Request.Context())
}

func (s *orderService) UpdateOrderItemService(c *gin.Context, updateData *models.UpdateOrderItem) error {
	_, err := lib.HasPermissionCheck(c, rbac.UpdateOrder)
	if err != nil {
		return errors.New("user not authorized")
	}
	return s.repo.UpdateOrderItemStatus(c.Request.Context(), &updateData.Status, updateData.OrderItemId, updateData.OrderId)
}

func (s *orderService) DeleteTableSessionByIdService(c *gin.Context, id *uuid.UUID, tableNumber int, phoneNumber string) error {
	_, err := lib.HasPermissionCheck(c, rbac.DeleteOrder)
	if err != nil {
		return errors.New("user not authorized")
	}

	return s.repo.DeleteTablesSessionById(c.Request.Context(), id, tableNumber, phoneNumber)
}

func (s *orderService) GetTableValidationByPhoneNTable(ctx context.Context, phone string, tableNumber int) (string, ReqStatus, error) {
	table, err := s.repo.GetTableValidationByTableAndPhone(ctx, tableNumber, phone)
	fmt.Println("this is table in valget : ", table)
	if err != nil {
		return "", OrderNotFound, err
	}

	if table.WaiterID == nil {
		return "", OrderNotApproved, nil
	}

	sessionJwtData := lib.OrderApprovalDataType{
		Id:          table.ID.String(),
		PhoneNumber: table.PhoneNumber,
		TableNumber: tableNumber,
	}

	sessionToken, err := lib.GenerateOrderApprovalToken(&sessionJwtData)
	if err != nil {
		return "", OrderNotApproved, err
	}

	return sessionToken, OrderApproved, nil

}

func (s *orderService) GetTableValidationById(ctx context.Context, id uuid.UUID) (*models.TableValidation, error) {
	return s.repo.GetTableValidationByID(ctx, id)
}

// ── Create New Approval Request ──────────────────────────────
func (s *orderService) CreateNewApprovalRequestService(c *gin.Context, req *models.CustomerApprovalRequest) (*models.TableValidation, error) {
	return s.repo.CreateNewApprovalRequest(c.Request.Context(), req)
}

// ── Approve Table By Waiter ────────────────────────────────
func (s *orderService) ApproveTableByWaiterService(c *gin.Context, req *models.WaiterApprovalRequest) error {
	_, err := lib.HasPermissionCheck(c, rbac.CreateOrder)
	if err != nil {
		return errors.New("user not authorized")
	}

	err = s.repo.ApproveTableByWaiter(c.Request.Context(), req)

	if err != nil {
		return errors.New("failed to approve table")
	}

	table, err := s.repo.GetTableValidationByTableAndPhone(c.Request.Context(), req.TableNumber, req.Phone)
	if err != nil {
		return err
	}

	if table.WaiterID == nil {
		return errors.New("Request is not approved")
	}

	return nil
}

// ── Delete Table Validation By ID ─────────────────────────
func (s *orderService) DeleteTableValidationService(c *gin.Context, id uuid.UUID) error {
	_, err := lib.HasPermissionCheck(c, rbac.DeleteOrder)
	if err != nil {
		return errors.New("user not authorized")
	}

	return s.repo.DeleteTableApprovalByID(c.Request.Context(), id)
}

// ── Get All Unassigned Tables ─────────────────────────────
func (s *orderService) GetUnassignedTablesService(c *gin.Context) ([]models.TableValidation, error) {
	_, err := lib.HasPermissionCheck(c, rbac.ViewOrder)
	if err != nil {
		return nil, errors.New("user not authorized")
	}

	return s.repo.GetUnassignedTables(c.Request.Context())
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
	fmt.Println("thisi si beforeh te saving user id : ", approveOrder)
	_, err := lib.HasPermissionCheck(c, rbac.CreateOrder)
	if err != nil {
		return errors.New("user not authorized")
	}

	userId, err := GetUserIDFromContext(c)
	if err != nil {
		return errors.New("unable to fetch user ID from context")
	}
	approveOrder.WaiterId = userId
	fmt.Println("iamnowapproving the order : ", approveOrder)
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
