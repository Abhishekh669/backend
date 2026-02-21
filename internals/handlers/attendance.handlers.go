package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Abhishekh669/backend/internals/models"
	"github.com/Abhishekh669/backend/internals/services"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
)

type AttendanceHandler struct {
	attendanceService services.AttendanceService
}

func (h *AttendanceHandler) CancelLeaveRequest(c *gin.Context) {
	leaveIDStr := c.Param("id")

	leaveID, err := uuid.FromString(leaveIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid leave id"})
		return
	}

	if err := h.attendanceService.CancelLeaveRequest(c, &leaveID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "leave request cancelled successfully",
	})
}

func (h *AttendanceHandler) DeleteLeaveRequest(c *gin.Context) {
	leaveIDStr := c.Param("id")

	leaveID, err := uuid.FromString(leaveIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid leave id"})
		return
	}

	if err := h.attendanceService.DeleteLeaveRequest(c, &leaveID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "leave request deleted successfully",
	})
}

func (h *AttendanceHandler) UpdateLeaveRequest(c *gin.Context) {
	var req models.UpdateAttendanceLeave

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.attendanceService.UpdateLeaveRequest(c, &req); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "leave request updated successfully",
	})
}

func (h *AttendanceHandler) CreateEmployeeRequest(c *gin.Context) {
	var req models.CreateAttendanceLeave

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.attendanceService.CreateEmployeeRequest(c, &req); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "leave request created successfully",
	})
}

func (h *AttendanceHandler) GetAttendanceHistory(c *gin.Context) {
	// Parse employee_id (optional)
	var employeeID *uuid.UUID
	if empIDStr := c.Query("employee_id"); empIDStr != "" {
		parsedID, err := uuid.FromString(empIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid employee_id format",
				"success": false,
			})
			return
		}
		employeeID = &parsedID
	}

	// Parse fromDate (optional)
	var fromDate *time.Time
	if fromDateStr := c.Query("startingDate"); fromDateStr != "" {
		parsedDate, err := time.Parse("2006-01-02", fromDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid fromDate format. Use YYYY-MM-DD",
				"success": false,
			})
			return
		}
		fromDate = &parsedDate
	}

	// Parse toDate (optional)
	var toDate *time.Time
	if toDateStr := c.Query("endingDate"); toDateStr != "" {
		parsedDate, err := time.Parse("2006-01-02", toDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid toDate format. Use YYYY-MM-DD",
				"success": false,
			})
			return
		}
		toDate = &parsedDate
	}

	// Parse limit (default to 10)
	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil || parsedLimit <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid limit parameter",
				"success": false,
			})
			return
		}
		limit = parsedLimit
	}

	// Parse page (default to 0)
	page := 0
	if pageStr := c.Query("page"); pageStr != "" {
		parsedPage, err := strconv.Atoi(pageStr)
		if err != nil || parsedPage < 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid page parameter",
				"success": false,
			})
			return
		}
		page = parsedPage
	}

	// Validate date range if both are provided
	if fromDate != nil && toDate != nil && toDate.Before(*fromDate) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "toDate cannot be before fromDate",
			"success": false,
		})
		return
	}

	// Build the query struct
	query := &models.AttendanceHistoryQuery{
		EmployeeId: employeeID,
		FromDate:   fromDate,
		ToDate:     toDate,
		Limit:      limit,
		Page:       page,
	}

	// Call service with the query struct
	response, err := h.attendanceService.GetAttendanceHistory(c, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	// Return successful response
	c.JSON(http.StatusOK, gin.H{
		"history": response,
		"success": true,
	})
}
func (h *AttendanceHandler) DeleteAttendanceByIdHandler(c *gin.Context) {
	attendanceID := c.Param("id")
	if attendanceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "attendance_id is required", "success": false})
		return
	}
	err := h.attendanceService.DeleteAttendanceByIdService(c, attendanceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "attendance deleted successfully",
		"success": true,
	})
}

func (h *AttendanceHandler) UpdateAttendanceHandler(c *gin.Context) {
	var req models.AttendanceUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}

	fmt.Println("this is attendance update : ',", req)

	err := h.attendanceService.UpdateAttendanceService(c, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "attendance updated successfully",
		"success": true,
	})
}

func (h *AttendanceHandler) GetCurrentAttendanceHandler(c *gin.Context) {
	attendance, err := h.attendanceService.GetCurrentAttendanceService(c)
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

	err := h.attendanceService.CheckInAttendanceService(c, &data)

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

	err := h.attendanceService.CheckOutAttendanceService(c, &data)

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
		attendanceService: attendanceService,
	}
}
