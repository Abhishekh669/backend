package handlers

import (
	"fmt"
	"net/http"

	"github.com/Abhishekh669/backend/internals/models"
	"github.com/Abhishekh669/backend/internals/services"
	"github.com/gin-gonic/gin"
)

type AttendanceHandler struct {
	attendanceHanlder services.AttendanceService
}

func (h *AttendanceHandler) GetCurrentAttendanceHandler(c *gin.Context) {
	attendance, err := h.attendanceHanlder.GetCurrentAttendanceService(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    attendance,
		"message": "current attendance fetched successfully",
		"success": true,
	})
}

func (h *AttendanceHandler) CheckInHandler(c *gin.Context) {
	var data models.CheckInOutAttendanceType
	if err := c.ShouldBindJSON(&data); err != nil {
		fmt.Println("error in binding", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}

	err := h.attendanceHanlder.CheckInAttendanceService(c, &data)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "attendance checked in successfully",
		"success": true,
	})
}

func (h *AttendanceHandler) CheckOutHandler(c *gin.Context) {
	var data models.CheckInOutAttendanceType
	if err := c.ShouldBindJSON(&data); err != nil {
		fmt.Println("error in binding", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}

	err := h.attendanceHanlder.CheckOutAttendanceService(c, &data)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "attendance checked out successfully",
		"success": true,
	})
}

func NewAttendanceHandler(attendanceService services.AttendanceService) *AttendanceHandler {
	return &AttendanceHandler{
		attendanceHanlder: attendanceService,
	}
}
