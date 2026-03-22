package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Abhishekh669/backend/internals/models"
	"github.com/Abhishekh669/backend/internals/services"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
)

type OrderHandler struct {
	orderService services.OrderService
}

func (h *OrderHandler) DeleteTableSessionByIdHandler(c *gin.Context) {
	idParam := c.Param("id")
	tableID, err := uuid.FromString(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid table ID", "success": false})
		return
	}

	phoneNumber := c.Query("phoneNumber")
	if len(phoneNumber) < 9 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid phone number", "success": false})
		return
	}

	tableNumberString := c.Query("tableNumber")
	tableNumber, err := strconv.Atoi(tableNumberString)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid table_number",
			"success": false,
		})
		return
	}

	if err := h.orderService.DeleteTableSessionByIdService(c, &tableID, tableNumber, phoneNumber); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "success": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Table service deleted successfully"})
}

func (h *OrderHandler) GetTableValidationFromTokenHandler(c *gin.Context) {
	tableData, exists := c.Get("table_data")

	if !exists {
		c.JSON(500, gin.H{
			"success": false,
			"error":   "table data not found in context",
		})
		return
	}

	c.JSON(200, gin.H{
		"success":          true,
		"table_validation": tableData,
	})
}

func (h *OrderHandler) GetTableValidationByPhoneAndTableHandler(c *gin.Context) {
	phone := c.Query("phone")
	tableNumberStr := c.Query("table_number")

	if phone == "" || tableNumberStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "phone and table_number are required",
			"success": false,
		})
		return
	}

	tableNumber, err := strconv.Atoi(tableNumberStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid table_number",
			"success": false,
		})
		return
	}

	token, reqStatus, err := h.orderService.GetTableValidationByPhoneNTable(c, phone, tableNumber)
	fmt.Println("thisi shte error : ", err)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to get status",
			"success": false,
		})
		return
	}

	switch reqStatus {

	case services.OrderNotFound:
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"status":  services.OrderNotFound,
			"message": "No request found",
		})

	case services.OrderNotApproved:
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"status":  services.OrderNotApproved,
			"message": "Request is pending approval",
		})

	case services.OrderApproved:
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"status":  services.OrderApproved,
			"token":   token,
		})

	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "unknown status",
			"success": false,
		})
	}
}

func (h *OrderHandler) GetTableValidationByIDHandler(c *gin.Context) {
	idParam := c.Param("id")

	id, err := uuid.FromString(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid UUID",
			"success": false,
		})
		return
	}

	result, err := h.orderService.GetTableValidationById(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"table":   result,
	})
}

// ── Create New Table Approval Request ─────────────────────
func (h *OrderHandler) CreateNewApprovalRequestHandler(c *gin.Context) {
	var req models.CustomerApprovalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}

	if len(req.Phone) < 9 && len(req.Phone) > 10 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid phone number", "success": false})
		return
	}

	fmt.Println("this is new order approv req : ", req)

	tableValidation, err := h.orderService.CreateNewApprovalRequestService(c, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "success": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"table":   tableValidation,
	})
}

// ── Approve Table By Waiter ──────────────────────────────
func (h *OrderHandler) ApproveTableByWaiterHandler(c *gin.Context) {
	var req models.WaiterApprovalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}

	token, err := h.orderService.ApproveTableByWaiterService(c, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "success": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Table approved successfully", "token": token})
}

// ── Delete Table Validation by ID ────────────────────────
func (h *OrderHandler) DeleteTableValidationHandler(c *gin.Context) {
	idParam := c.Param("id")
	tableID, err := uuid.FromString(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid table ID", "success": false})
		return
	}

	if err := h.orderService.DeleteTableValidationService(c, tableID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "success": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Table validation deleted successfully"})
}

// ── Get All Unassigned Tables ───────────────────────────
func (h *OrderHandler) GetUnassignedTablesHandler(c *gin.Context) {
	tables, err := h.orderService.GetUnassignedTablesService(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "success": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"tables":  tables,
	})
}

func (h *OrderHandler) GetAllOrderStatusHandler(c *gin.Context) {
	orderRequests, err := h.orderService.GetAllOrderStatusService(c)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error(), "success": false})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"order_requests": orderRequests,
	})
}

func (h *OrderHandler) GetOrderRequestByTableNumberNPhone(c *gin.Context) {
	// Get query parameters
	phoneNumber := c.Query("phone")
	tableNumberStr := c.Query("table_number")

	// Validate required parameters
	if phoneNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "phone number is required",
			"success": false,
		})
		return
	}

	if tableNumberStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "table number is required",
			"success": false,
		})
		return
	}

	// Parse table number to int
	tableNumber, err := strconv.Atoi(tableNumberStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid table number format",
			"success": false,
		})
		return
	}

	// Validate phone number (basic validation)
	if len(phoneNumber) < 10 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid phone number format",
			"success": false,
		})
		return
	}

	// Call service to get order request
	orderRequestByPhone, err := h.orderService.GetOrderReqeustFromTableNumberAndPhoneNumber(c, tableNumber, phoneNumber)

	fmt.Println("this is order request by phone  and table  : ", orderRequestByPhone)

	if err != nil {
		// Handle different error cases
		if strings.Contains(err.Error(), "no active session found") {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "No active order found for this table and phone number",
				"success": false,
			})
			return
		}

		if strings.Contains(err.Error(), "no orders found") {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "No orders found for this session",
				"success": false,
			})
			return
		}

		// Log the error for debugging
		fmt.Printf("Error fetching order request: %v\n", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch order request",
			"success": false,
		})
		return
	}

	// Return successful response
	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"order_request": orderRequestByPhone,
		"message":       "Order request fetched successfully",
	})
}
func (h *OrderHandler) GetOrderRequestByTableSessionIdHandler(c *gin.Context) {

	// Get the session ID from path parameter
	sessionIdStr := c.Param("table-session-id")

	// Check if parameter is provided
	if sessionIdStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "table session id is required",
			"success": false,
		})
		return
	}

	// Parse the UUID
	sessionId, err := uuid.FromString(sessionIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid table session id format",
			"success": false,
		})
		return
	}

	// Pass the UUID directly (not a pointer)
	orderRequestByTableId, err := h.orderService.GetOrderRequestFromTableSession(c, sessionId)

	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"order_request": orderRequestByTableId,
	})
}

func (h *OrderHandler) GetAllOrderRequestHandler(c *gin.Context) {
	orderRequests, err := h.orderService.GetAllOrderRequests(c)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error(), "success": false})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"order_requests": orderRequests,
	})
}

func (h *OrderHandler) ApproveCustomerOrderHandler(c *gin.Context) {
	var orderData *models.ApproveOrderType
	if err := c.ShouldBindJSON(&orderData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}

	err := h.orderService.ApproveCustomerOrder(c, orderData)
	fmt.Println("this isthe error in order approval : ", err)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error(), "success": false})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "created order successfully",
		"success": true,
	})
}

func (h *OrderHandler) CreateCustomerHandler(c *gin.Context) {
	var orderData *models.CreateCustomerOrderRequest
	if err := c.ShouldBindJSON(&orderData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}

	err := h.orderService.CreateCustomerService(c, orderData)
	fmt.Println("this is order service  err : ", err)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error(), "success": false})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "requested order successfully",
		"success": true,
	})
}

func NewOrderHandler(orderService services.OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}
