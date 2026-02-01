package handlers

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Abhishekh669/backend/internals/models"
	"github.com/Abhishekh669/backend/internals/services"
	"github.com/gin-gonic/gin"
)

type RawMaterialsHandler struct {
	rawMaterials services.RawMaterialService
}

type CreateJsonType struct {
	RawMaterials []models.CreateRawMaterialType `json:"raw_materials"`
}

func (h *RawMaterialsHandler) UpdateRawMaterialsHandler(c *gin.Context) {
	var data models.UpdateRawMaterials
	if err := c.ShouldBindJSON(&data); err != nil {
		fmt.Println("error in binding", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}

	err := h.rawMaterials.UpdateRawMaterialService(c, &data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "raw materials updated successfully",
		"success": true,
	})

}

func (h *RawMaterialsHandler) DeleteRawMaterialsHandler(c *gin.Context) {
	var data models.DeleteRawMaterialsPayload

	if err := c.ShouldBindJSON(&data); err != nil {
		fmt.Println("error in binding", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}

	if len(data.RawMaterialIds) < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no user selected", "success": false})
		return
	}

	err := h.rawMaterials.DeleteRawMaterails(c, data.RawMaterialIds)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}
	var message string
	if len(data.RawMaterialIds) > 0 {
		message = fmt.Sprintf(" %d raw materials deleted successfully", len(data.RawMaterialIds))
	} else {
		message = "raw material deleted successfully"
	}
	c.JSON(http.StatusOK, gin.H{"message": message, "success": true})

}

func (h *RawMaterialsHandler) GetRawMaterialsHandlers(c *gin.Context) {
	const (
		maxLimit     = 100
		defaultLimit = 20
		defaultPage  = 0
	)

	/* -------------------- Query params -------------------- */
	search := c.DefaultQuery("search", "")
	limitStr := c.DefaultQuery("limit", strconv.Itoa(defaultLimit))
	pageStr := c.DefaultQuery("page", strconv.Itoa(defaultPage))
	oldestFirstStr := c.DefaultQuery("oldFirst", "false")

	startingPriceStr := c.DefaultQuery("startingPrice", "0")
	endingPriceStr := c.DefaultQuery("endingPrice", "0")

	fromDateStr := c.DefaultQuery("fromDate", "")
	toDateStr := c.DefaultQuery("toDate", "")

	/* -------------------- Parse limit -------------------- */
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	/* -------------------- Parse page -------------------- */
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 0 {
		page = defaultPage
	}

	/* -------------------- Parse oldestFirst -------------------- */
	oldestFirst := strings.ToLower(oldestFirstStr) == "true"

	/* -------------------- Parse price -------------------- */
	startingPrice, err := strconv.Atoi(startingPriceStr)
	if err != nil || startingPrice < 0 {
		startingPrice = 0
	}

	endingPrice, err := strconv.Atoi(endingPriceStr)
	if err != nil || endingPrice <= 0 {
		endingPrice = math.MaxInt32
	}

	if endingPrice < startingPrice {
		endingPrice = startingPrice
	}

	/* -------------------- Parse dates (OPTIONAL) -------------------- */
	var fromDate *time.Time
	var toDate *time.Time
	const layout = "2006-01-02"

	if fromDateStr != "" {
		fd, err := time.Parse(layout, fromDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "invalid fromDate format (YYYY-MM-DD)",
			})
			return
		}
		fd = time.Date(fd.Year(), fd.Month(), fd.Day(), 0, 0, 0, 0, time.UTC)
		fromDate = &fd
	}

	if toDateStr != "" {
		td, err := time.Parse(layout, toDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "invalid toDate format (YYYY-MM-DD)",
			})
			return
		}
		td = time.Date(td.Year(), td.Month(), td.Day(), 23, 59, 59, 999999999, time.UTC)
		toDate = &td
	}

	/* -------------------- Service call -------------------- */
	result, err := h.rawMaterials.GetRawMaterialService(
		c,
		limit,
		page,
		fromDate,
		toDate,
		int32(startingPrice),
		int32(endingPrice),
		search,
		oldestFirst,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	/* -------------------- Response -------------------- */
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
		"message": "raw materials fetched successfully",
	})
}

func (h *RawMaterialsHandler) CreateRawMaterialHandlers(c *gin.Context) {
	var data CreateJsonType
	if err := c.ShouldBindJSON(&data); err != nil {
		fmt.Println("error in binding", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}

	fmt.Println("this is rawm amterials : ", data)

	err := h.rawMaterials.CreateRawMaterialService(c, data.RawMaterials)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "raw materials created successfully",
		"success": true,
	})

}

func NewRawMaterialHandler(rawMaterialService services.RawMaterialService) *RawMaterialsHandler {
	return &RawMaterialsHandler{
		rawMaterials: rawMaterialService,
	}
}
