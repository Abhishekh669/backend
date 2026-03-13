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
