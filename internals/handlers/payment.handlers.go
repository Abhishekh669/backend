package handlers

import (
	"fmt"
	"net/http"

	"github.com/Abhishekh669/backend/internals/models"
	"github.com/Abhishekh669/backend/internals/services"
	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	paymentService services.PaymentService
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
