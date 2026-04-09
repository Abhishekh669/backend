package algorithm

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/Abhishekh669/backend/internals/models"
	"github.com/Abhishekh669/backend/internals/repository"
)

// ─── Cached Report Data ───────────────────────────────────────────────────────

type CachedReports struct {
	Revenue     *models.RevenueReportResponse
	Sales       *models.SalesReportResponse
	Customer    *models.CustomerReportResponse
	Table       *models.TableReportResponse
	Staff       *models.ExtendedStaffReportResponse
	Financial   *models.FinancialSummaryResponse
	RawMaterial *models.RawMaterialReportResponse

	LastUpdated time.Time
	IsReady     bool
}

// ─── Report Cache ─────────────────────────────────────────────────────────────

type ReportCache struct {
	mu           sync.RWMutex
	data         CachedReports
	repo         repository.ReportRepo
	isRefreshing bool // true while a manual/nightly reload is running
	defaultFrom  func() time.Time
	defaultTo    func() time.Time
}

func NewReportCache(repo repository.ReportRepo) *ReportCache {
	return &ReportCache{
		repo: repo,
		defaultFrom: func() time.Time {
			return time.Now().AddDate(0, 0, -30)
		},
		defaultTo: func() time.Time {
			return time.Now()
		},
	}
}

// ─── ReloadFromDB (startup + nightly job) ────────────────────────────────────
// Always reads from DB and replaces the cache.
// Tab GET endpoints are never blocked — they keep serving stale cache
// while this runs, and serve fresh data the moment it finishes.

func (c *ReportCache) ReloadFromDB() {
	c.mu.Lock()
	c.isRefreshing = true
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		c.isRefreshing = false
		c.mu.Unlock()
	}()

	log.Println("📊 [ReportCache] Starting report cache reload...")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	from := c.defaultFrom()
	to := c.defaultTo()

	revenue, err := c.repo.GetRevenueReport(ctx, from, to)
	if err != nil {
		log.Printf("❌ [ReportCache] revenue: %v", err)
		return
	}

	sales, err := c.repo.GetSalesReport(ctx, from, to)
	if err != nil {
		log.Printf("❌ [ReportCache] sales: %v", err)
		return
	}

	customer, err := c.repo.GetCustomerReport(ctx, from, to)
	if err != nil {
		log.Printf("❌ [ReportCache] customer: %v", err)
		return
	}

	table, err := c.repo.GetTableReport(ctx, from, to)
	if err != nil {
		log.Printf("❌ [ReportCache] table: %v", err)
		return
	}

	staff, err := c.repo.GetStaffReport(ctx, from, to)
	if err != nil {
		log.Printf("❌ [ReportCache] staff: %v", err)
		return
	}

	financial, err := c.repo.GetFinancialSummary(ctx)
	if err != nil {
		log.Printf("❌ [ReportCache] financial: %v", err)
		return
	}

	rawMaterial, err := c.repo.GetRawMaterialReport(ctx)
	if err != nil {
		log.Printf("❌ [ReportCache] raw material: %v", err)
		return
	}

	// Atomic swap — readers are never blocked for more than this tiny lock window
	c.mu.Lock()
	c.data = CachedReports{
		Revenue:     revenue,
		Sales:       sales,
		Customer:    customer,
		Table:       table,
		Staff:       staff,
		Financial:   financial,
		RawMaterial: rawMaterial,
		LastUpdated: time.Now(),
		IsReady:     true,
	}
	c.mu.Unlock()

	log.Printf("✅ [ReportCache] Reload complete at %s", time.Now().Format("2006-01-02 15:04:05"))
}

// ─── ReloadWithRange (manual refresh triggered by admin) ─────────────────────
// Same as ReloadFromDB but uses the date range the admin chose.
// Old cached data stays live and is served to any concurrent tab requests
// until the new data is ready and atomically swapped in.

func (c *ReportCache) ReloadWithRange(ctx context.Context, from, to time.Time) error {
	c.mu.Lock()
	c.isRefreshing = true
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		c.isRefreshing = false
		c.mu.Unlock()
	}()

	log.Printf("🔄 [ReportCache] Manual refresh: %s → %s", from.Format("2006-01-02"), to.Format("2006-01-02"))

	revenue, err := c.repo.GetRevenueReport(ctx, from, to)
	if err != nil {
		return err
	}

	sales, err := c.repo.GetSalesReport(ctx, from, to)
	if err != nil {
		return err
	}

	customer, err := c.repo.GetCustomerReport(ctx, from, to)
	if err != nil {
		return err
	}

	table, err := c.repo.GetTableReport(ctx, from, to)
	if err != nil {
		return err
	}

	staff, err := c.repo.GetStaffReport(ctx, from, to)
	if err != nil {
		return err
	}

	financial, err := c.repo.GetFinancialSummary(ctx)
	if err != nil {
		return err
	}

	rawMaterial, err := c.repo.GetRawMaterialReport(ctx)
	if err != nil {
		return err
	}

	// Atomic swap — readers keep serving old data until this exact line
	c.mu.Lock()
	c.data = CachedReports{
		Revenue:     revenue,
		Sales:       sales,
		Customer:    customer,
		Table:       table,
		Staff:       staff,
		Financial:   financial,
		RawMaterial: rawMaterial,
		LastUpdated: time.Now(),
		IsReady:     true,
	}
	c.mu.Unlock()

	log.Printf("✅ [ReportCache] Manual refresh complete at %s", time.Now().Format("2006-01-02 15:04:05"))
	return nil
}

// ─── Getters ─────────────────────────────────────────────────────────────────
// All reads use RLock so multiple tab fetches run fully in parallel.
// Writers (ReloadFromDB / ReloadWithRange) only block readers for the
// single atomic swap at the end — never during the DB queries.

func (c *ReportCache) GetRevenue() (*models.RevenueReportResponse, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.data.Revenue, c.data.IsReady
}

func (c *ReportCache) GetSales() (*models.SalesReportResponse, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.data.Sales, c.data.IsReady
}

func (c *ReportCache) GetCustomer() (*models.CustomerReportResponse, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.data.Customer, c.data.IsReady
}

func (c *ReportCache) GetTable() (*models.TableReportResponse, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.data.Table, c.data.IsReady
}

func (c *ReportCache) GetStaff() (*models.ExtendedStaffReportResponse, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.data.Staff, c.data.IsReady
}

func (c *ReportCache) GetFinancial() (*models.FinancialSummaryResponse, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.data.Financial, c.data.IsReady
}

func (c *ReportCache) GetRawMaterial() (*models.RawMaterialReportResponse, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.data.RawMaterial, c.data.IsReady
}

func (c *ReportCache) GetLastUpdated() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.data.LastUpdated
}

func (c *ReportCache) IsReady() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.data.IsReady
}

// IsRefreshing returns true while a reload is in progress.
// Frontend uses this to show a spinner on the Refresh button
// without blocking tab navigation — tabs keep serving from cache.
func (c *ReportCache) IsRefreshing() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isRefreshing
}
