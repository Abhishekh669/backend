package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Abhishekh669/backend/internals/algorithm"
	"github.com/Abhishekh669/backend/internals/lib"
	"github.com/Abhishekh669/backend/internals/models"
	"github.com/Abhishekh669/backend/internals/rbac"
	"github.com/Abhishekh669/backend/internals/repository"
	"github.com/gin-gonic/gin"
)

type ReportHandler struct {
	reportRepo         repository.ReportRepo
	defaultReportCache *algorithm.DefaultRevenueCache
}

func NewReportHandler(repo repository.ReportRepo, cache *algorithm.DefaultRevenueCache) *ReportHandler {
	return &ReportHandler{
		reportRepo:         repo,
		defaultReportCache: cache,
	}
}

func (h *ReportHandler) GetCustomRangeRawMaterialsReport(c *gin.Context) {
	// Parse query parameters
	_, err := lib.HasPermissionCheck(c, rbac.ViewReports)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "user not authorized to refresh reports",
			"success": false,
		})
		return
	}
	fromStr := c.Query("from")
	toStr := c.Query("to")

	if fromStr == "" || toStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "from and to parameters are required",
			"success": false,
		})
		return
	}

	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid from date format (use YYYY-MM-DD)",
			"success": false,
		})
		return
	}

	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid to date format (use YYYY-MM-DD)",
			"success": false,
		})
		return
	}

	// Parse pagination with defaults
	limit := c.DefaultQuery("limit", "10")
	page := c.DefaultQuery("page", "0")

	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		limitInt = 10
	}

	pageInt, err := strconv.Atoi(page)
	if err != nil {
		pageInt = 0
	}

	req := &models.NewRawMaterialCustomRangeReportRequest{
		From:  from,
		To:    to,
		Limit: limitInt,
		Page:  pageInt,
	}

	report, err := h.reportRepo.NewGetCustomRangeRawMaterialReport(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	c.Header("X-Cache", "BYPASS")
	c.JSON(http.StatusOK, gin.H{
		"report":  report,
		"success": true,
	})
}

// GET /api/v2/reports/revenue/default
func (h *ReportHandler) GetDefaultRawMaterialsReport(c *gin.Context) {
	_, err := lib.HasPermissionCheck(c, rbac.ViewReports)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "user not authorized to refresh reports",
			"success": false,
		})
		return
	}
	// Try to get from cache
	cachedReport, isReady := h.defaultReportCache.GetRawMaterialReport()

	if isReady && cachedReport != nil {
		// Add cache headers
		c.Header("X-Cache", "HIT")
		c.Header("X-Cache-Last-Updated", h.defaultReportCache.GetLastUpdated().Format(time.RFC3339))
		c.Header("X-Cache-Is-Refreshing", boolToString(h.defaultReportCache.IsRefreshing()))
		c.JSON(http.StatusOK, gin.H{
			"report":  cachedReport,
			"success": true,
		})
		return
	}

	// Cache miss - fetch from DB (should not happen after initial load)
	c.Header("X-Cache", "MISS")
	report, err := h.reportRepo.NewGetDefaultRawMaterialReport(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"report":  report,
		"success": true,
	})
}

func (h *ReportHandler) GetCustomRangeStaffsReport(c *gin.Context) {
	_, err := lib.HasPermissionCheck(c, rbac.ViewReports)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "user not authorized to refresh reports",
			"success": false,
		})
		return
	}
	// Parse query parameters
	fromStr := c.Query("from")
	toStr := c.Query("to")

	if fromStr == "" || toStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "from and to parameters are required",
			"success": false,
		})
		return
	}

	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid from date format (use YYYY-MM-DD)",
			"success": false,
		})
		return
	}

	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid to date format (use YYYY-MM-DD)",
			"success": false,
		})
		return
	}

	// Parse pagination with defaults
	limit := c.DefaultQuery("limit", "10")
	page := c.DefaultQuery("page", "0")

	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		limitInt = 10
	}

	pageInt, err := strconv.Atoi(page)
	if err != nil {
		pageInt = 0
	}

	req := &models.NewStaffCustomRangeReportRequest{
		From:  from,
		To:    to,
		Limit: limitInt,
		Page:  pageInt,
	}

	report, err := h.reportRepo.NewGetCustomRangeStaffReport(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	c.Header("X-Cache", "BYPASS")
	c.JSON(http.StatusOK, gin.H{
		"report":  report,
		"success": true,
	})
}

// GET /api/v2/reports/revenue/default
func (h *ReportHandler) GetDefaultStaffReport(c *gin.Context) {
	_, err := lib.HasPermissionCheck(c, rbac.ViewReports)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "user not authorized to refresh reports",
			"success": false,
		})
		return
	}
	// Try to get from cache
	cachedReport, isReady := h.defaultReportCache.GetStaffReport()

	if isReady && cachedReport != nil {
		// Add cache headers
		c.Header("X-Cache", "HIT")
		c.Header("X-Cache-Last-Updated", h.defaultReportCache.GetLastUpdated().Format(time.RFC3339))
		c.Header("X-Cache-Is-Refreshing", boolToString(h.defaultReportCache.IsRefreshing()))
		c.JSON(http.StatusOK, gin.H{
			"report":  cachedReport,
			"success": true,
		})
		return
	}

	// Cache miss - fetch from DB (should not happen after initial load)
	c.Header("X-Cache", "MISS")
	report, err := h.reportRepo.NewGetDefaultStaffReport(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"report":  report,
		"success": true,
	})
}

func (h *ReportHandler) GetCustomRangeTablesReport(c *gin.Context) {
	_, err := lib.HasPermissionCheck(c, rbac.ViewReports)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "user not authorized to refresh reports",
			"success": false,
		})
		return
	}
	// Parse query parameters
	fromStr := c.Query("from")
	toStr := c.Query("to")

	if fromStr == "" || toStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "from and to parameters are required",
			"success": false,
		})
		return
	}

	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid from date format (use YYYY-MM-DD)",
			"success": false,
		})
		return
	}

	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid to date format (use YYYY-MM-DD)",
			"success": false,
		})
		return
	}

	// Parse pagination with defaults
	limit := c.DefaultQuery("limit", "10")
	page := c.DefaultQuery("page", "0")

	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		limitInt = 10
	}

	pageInt, err := strconv.Atoi(page)
	if err != nil {
		pageInt = 0
	}

	req := &models.NewTableCustomRangeReportRequest{
		From:  from,
		To:    to,
		Limit: limitInt,
		Page:  pageInt,
	}

	report, err := h.reportRepo.NewGetCustomRangeTableReport(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	c.Header("X-Cache", "BYPASS")
	c.JSON(http.StatusOK, gin.H{
		"report":  report,
		"success": true,
	})
}

// GET /api/v2/reports/revenue/default
func (h *ReportHandler) GetDefaultTableReport(c *gin.Context) {
	_, err := lib.HasPermissionCheck(c, rbac.ViewReports)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "user not authorized to refresh reports",
			"success": false,
		})
		return
	}
	// Try to get from cache
	cachedReport, isReady := h.defaultReportCache.GetTableReport()

	if isReady && cachedReport != nil {
		// Add cache headers
		c.Header("X-Cache", "HIT")
		c.Header("X-Cache-Last-Updated", h.defaultReportCache.GetLastUpdated().Format(time.RFC3339))
		c.Header("X-Cache-Is-Refreshing", boolToString(h.defaultReportCache.IsRefreshing()))
		c.JSON(http.StatusOK, gin.H{
			"report":  cachedReport,
			"success": true,
		})
		return
	}

	// Cache miss - fetch from DB (should not happen after initial load)
	c.Header("X-Cache", "MISS")
	report, err := h.reportRepo.NewGetDefaultTableReport(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"report":  report,
		"success": true,
	})
}

func (h *ReportHandler) GetCustomRangeCustomerReport(c *gin.Context) {
	_, err := lib.HasPermissionCheck(c, rbac.ViewReports)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "user not authorized to refresh reports",
			"success": false,
		})
		return
	}
	// Parse query parameters
	fromStr := c.Query("from")
	toStr := c.Query("to")

	if fromStr == "" || toStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "from and to parameters are required",
			"success": false,
		})
		return
	}

	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid from date format (use YYYY-MM-DD)",
			"success": false,
		})
		return
	}

	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid to date format (use YYYY-MM-DD)",
			"success": false,
		})
		return
	}

	// Parse pagination with defaults
	limit := c.DefaultQuery("limit", "10")
	page := c.DefaultQuery("page", "0")

	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		limitInt = 10
	}

	pageInt, err := strconv.Atoi(page)
	if err != nil {
		pageInt = 0
	}

	req := &models.NewCustomerCustomRangeReportRequest{
		From:  from,
		To:    to,
		Limit: limitInt,
		Page:  pageInt,
	}

	// This always hits the database directly (no cache)
	report, err := h.reportRepo.NewGetCustomRangeCustomerReport(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	c.Header("X-Cache", "BYPASS")
	c.JSON(http.StatusOK, gin.H{
		"report":  report,
		"success": true,
	})
}

// GET /api/v2/reports/revenue/default
func (h *ReportHandler) GetDefaultCustomerReport(c *gin.Context) {
	_, err := lib.HasPermissionCheck(c, rbac.ViewReports)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "user not authorized to refresh reports",
			"success": false,
		})
		return
	}
	// Try to get from cache
	cachedReport, isReady := h.defaultReportCache.GetCustomerReport()

	if isReady && cachedReport != nil {
		// Add cache headers
		c.Header("X-Cache", "HIT")
		c.Header("X-Cache-Last-Updated", h.defaultReportCache.GetLastUpdated().Format(time.RFC3339))
		c.Header("X-Cache-Is-Refreshing", boolToString(h.defaultReportCache.IsRefreshing()))
		c.JSON(http.StatusOK, gin.H{
			"report":  cachedReport,
			"success": true,
		})
		return
	}

	// Cache miss - fetch from DB (should not happen after initial load)
	c.Header("X-Cache", "MISS")
	report, err := h.reportRepo.NewGetDefaultCustomerReport(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"report":  report,
		"success": true,
	})
}

func (h *ReportHandler) GetCustomRangeSalesReport(c *gin.Context) {
	_, err := lib.HasPermissionCheck(c, rbac.ViewReports)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "user not authorized to refresh reports",
			"success": false,
		})
		return
	}
	// Parse query parameters
	fromStr := c.Query("from")
	toStr := c.Query("to")

	if fromStr == "" || toStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "from and to parameters are required",
			"success": false,
		})
		return
	}

	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid from date format (use YYYY-MM-DD)",
			"success": false,
		})
		return
	}

	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid to date format (use YYYY-MM-DD)",
			"success": false,
		})
		return
	}

	// Parse pagination with defaults
	limit := c.DefaultQuery("limit", "10")
	page := c.DefaultQuery("page", "0")

	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		limitInt = 10
	}

	pageInt, err := strconv.Atoi(page)
	if err != nil {
		pageInt = 0
	}

	req := &models.NewSalesCustomRangeReportRequest{
		From:  from,
		To:    to,
		Limit: limitInt,
		Page:  pageInt,
	}

	// This always hits the database directly (no cache)
	report, err := h.reportRepo.NewGetCustomRangeSalesReport(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	c.Header("X-Cache", "BYPASS")
	c.JSON(http.StatusOK, gin.H{
		"report":  report,
		"success": true,
	})
}

// GET /api/v2/reports/revenue/default
func (h *ReportHandler) GetDefaultSalesReport(c *gin.Context) {
	_, err := lib.HasPermissionCheck(c, rbac.ViewReports)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "user not authorized to refresh reports",
			"success": false,
		})
		return
	}
	// Try to get from cache
	cachedReport, isReady := h.defaultReportCache.GetSalesReport()

	if isReady && cachedReport != nil {
		// Add cache headers
		c.Header("X-Cache", "HIT")
		c.Header("X-Cache-Last-Updated", h.defaultReportCache.GetLastUpdated().Format(time.RFC3339))
		c.Header("X-Cache-Is-Refreshing", boolToString(h.defaultReportCache.IsRefreshing()))
		c.JSON(http.StatusOK, gin.H{
			"report":  cachedReport,
			"success": true,
		})
		return
	}

	// Cache miss - fetch from DB (should not happen after initial load)
	c.Header("X-Cache", "MISS")
	report, err := h.reportRepo.NewGetDefaultSalesReport(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"report":  report,
		"success": true,
	})
}

// GetDefaultRevenueReport returns default report (7 days, 7 weeks, 7 months, 7 years) - CACHED
// GET /api/v2/reports/revenue/default
func (h *ReportHandler) GetDefaultRevenueReport(c *gin.Context) {
	_, err := lib.HasPermissionCheck(c, rbac.ViewReports)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "user not authorized to refresh reports",
			"success": false,
		})
		return
	}
	// Try to get from cache
	cachedReport, isReady := h.defaultReportCache.GetRevenueReport()

	if isReady && cachedReport != nil {
		// Add cache headers
		c.Header("X-Cache", "HIT")
		c.Header("X-Cache-Last-Updated", h.defaultReportCache.GetLastUpdated().Format(time.RFC3339))
		c.Header("X-Cache-Is-Refreshing", boolToString(h.defaultReportCache.IsRefreshing()))
		c.JSON(http.StatusOK, gin.H{
			"report":  cachedReport,
			"success": true,
		})
		return
	}

	// Cache miss - fetch from DB (should not happen after initial load)
	c.Header("X-Cache", "MISS")
	report, err := h.reportRepo.NewGetDefaultRevenueReport(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"report":  report,
		"success": true,
	})
}

// GetCustomRangeRevenueReport returns custom range report with pagination - DIRECT DB (NO CACHE)
// GET /api/v2/reports/revenue/custom?from=2024-01-01&to=2024-12-31&page=0&limit=10
func (h *ReportHandler) GetCustomRangeRevenueReport(c *gin.Context) {
	_, err := lib.HasPermissionCheck(c, rbac.ViewReports)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "user not authorized to refresh reports",
			"success": false,
		})
		return
	}
	// Parse query parameters
	fromStr := c.Query("from")
	toStr := c.Query("to")

	if fromStr == "" || toStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "from and to parameters are required",
			"success": false,
		})
		return
	}

	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid from date format (use YYYY-MM-DD)",
			"success": false,
		})
		return
	}

	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid to date format (use YYYY-MM-DD)",
			"success": false,
		})
		return
	}

	// Parse pagination with defaults
	limit := c.DefaultQuery("limit", "10")
	page := c.DefaultQuery("page", "0")

	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		limitInt = 10
	}

	pageInt, err := strconv.Atoi(page)
	if err != nil {
		pageInt = 0
	}

	req := &models.NewCustomRangeReportRequest{
		From:  from,
		To:    to,
		Limit: limitInt,
		Page:  pageInt,
	}

	// This always hits the database directly (no cache)
	report, err := h.reportRepo.NewGetCustomRangeRevenueReport(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	c.Header("X-Cache", "BYPASS")
	c.JSON(http.StatusOK, gin.H{
		"report":  report,
		"success": true,
	})
}

func (h *ReportHandler) RefreshDefaultReportCache(c *gin.Context) {

	_, err := lib.HasPermissionCheck(c, rbac.ViewReports)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "user not authorized to refresh reports",
			"success": false,
		})
		return
	}
	twoHoursAgo := time.Now().Add(-1 * time.Hour)
	if h.defaultReportCache.GetLastUpdated().After(twoHoursAgo) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "cache was refreshed less than 1 hours ago, please wait before refreshing again",
			"success": false,
		})
		return
	}
	err = h.defaultReportCache.ReloadFromDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to refresh cache: " + err.Error(),
			"success": false,
		})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message": "successfully refreshed",
		"success": true,
	})
}

// GetDefaultReportCacheStatus returns cache status
// GET /api/v2/reports/revenue/default/status
func (h *ReportHandler) GetDefaultReportCacheStatus(c *gin.Context) {
	status := gin.H{
		"is_ready":      h.defaultReportCache.IsReady(),
		"is_refreshing": h.defaultReportCache.IsRefreshing(),
		"last_updated":  h.defaultReportCache.GetLastUpdated(),
	}

	c.JSON(http.StatusOK, status)
}

// boolToString helper function
func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
