package handlers

import (
	"fmt"
	"net/http"

	"github.com/Abhishekh669/backend/internals/models"
	"github.com/Abhishekh669/backend/internals/services"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
)

type PaymentHandler struct {
	paymentService services.PaymentService
}

func (h *PaymentHandler) DeleteOrderByCashierHandler(c *gin.Context) {
	orderIdStr := c.Param("orderId")
	orderId, err := uuid.FromString(orderIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID", "success": false})
		return
	}

	err = h.paymentService.DeleteOrderByCashier(c, orderId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete order", "success": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "order deleted successfully", "success": true})
}

func (h *PaymentHandler) GetAllOrderDetailsForCashierByOrderIdHandler(c *gin.Context) {
	orderIdStr := c.Param("orderId")
	fmt.Println("thisis the payment order details : ", orderIdStr)
	orderId, err := uuid.FromString(orderIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID", "success": false})
		return
	}

	details, err := h.paymentService.GetAllOrderDetailsForCAshierByOrderId(c, orderId)
	fmt.Println("error in getting payment order details : ", details, err)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch order details", "success": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{"order": details, "success": true})
}

func (h *PaymentHandler) GetAllApprovedOrdersForCashierHandler(c *gin.Context) {
	orders, err := h.paymentService.GetAllApprovedOrdersForCashierService(c)
	fmt.Println("thisi shte orders  : ", orders, err)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch approved orders", "success": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{"orders": orders, "success": true})
}

func (h *PaymentHandler) CreatePaymentHandler(c *gin.Context) {
	var req models.CreatePayment
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}

	payment, err := h.paymentService.CreatePaymentService(c, &req)
	fmt.Println("this is the payment create section: ", payment, err)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create payment", "success": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{"payment": payment, "success": true})

}

func NewPaymentHandler(paymentService services.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
	}
}
