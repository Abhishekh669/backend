package algorithm

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/Abhishekh669/backend/internals/models"
	"github.com/Abhishekh669/backend/internals/repository"
)

type CachedDefaultRevenue struct {
	Revenue      *models.NewDefaultRevenueResponse
	Sales        *models.NewDefaultSalesResponse
	Customers    *models.NewDefaultCustomerResponse
	Tables       *models.NewDefaultTableResponse
	Staffs       *models.NewDefaultStaffResponse
	RawMaterials *models.NewDefaultRawMaterialResponse
	LastUpdated  time.Time
	IsReady      bool
}

type DefaultRevenueCache struct {
	mu           sync.RWMutex
	data         CachedDefaultRevenue
	repo         repository.ReportRepo
	isRefreshing bool
}

func NewDefaultRevenueCache(repo repository.ReportRepo) *DefaultRevenueCache {
	return &DefaultRevenueCache{
		repo: repo,
	}
}

type reportResult struct {
	reportType string
	data       interface{}
	err        error
}

func (c *DefaultRevenueCache) ReloadFromDB() {
	c.mu.Lock()
	c.isRefreshing = true
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		c.isRefreshing = false
		c.mu.Unlock()
	}()

	log.Println("🚀 [DefaultRevenueCache] Starting concurrent cache reload...")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	startTime := time.Now()

	// Channel to collect results
	resultChan := make(chan reportResult, 6)
	var wg sync.WaitGroup

	// Launch concurrent goroutines for each report type
	wg.Add(6)

	// 1. Revenue Report
	go func() {
		defer wg.Done()
		start := time.Now()
		log.Println("📊 [Revenue] Fetching revenue report...")
		data, err := c.repo.NewGetDefaultRevenueReport(ctx)
		if err != nil {
			log.Printf("❌ [Revenue] Failed to fetch: %v", err)
			resultChan <- reportResult{reportType: "revenue", data: nil, err: err}
			return
		}
		log.Printf("✅ [Revenue] Fetched successfully in %v", time.Since(start))
		resultChan <- reportResult{reportType: "revenue", data: data, err: nil}
	}()

	// 2. Sales Report
	go func() {
		defer wg.Done()
		start := time.Now()
		log.Println("📈 [Sales] Fetching sales report...")
		data, err := c.repo.NewGetDefaultSalesReport(ctx)
		if err != nil {
			log.Printf("❌ [Sales] Failed to fetch: %v", err)
			resultChan <- reportResult{reportType: "sales", data: nil, err: err}
			return
		}
		log.Printf("✅ [Sales] Fetched successfully in %v", time.Since(start))
		resultChan <- reportResult{reportType: "sales", data: data, err: nil}
	}()

	// 3. Customer Report
	go func() {
		defer wg.Done()
		start := time.Now()
		log.Println("👥 [Customers] Fetching customer report...")
		data, err := c.repo.NewGetDefaultCustomerReport(ctx)
		if err != nil {
			log.Printf("❌ [Customers] Failed to fetch: %v", err)
			resultChan <- reportResult{reportType: "customers", data: nil, err: err}
			return
		}
		log.Printf("✅ [Customers] Fetched successfully in %v", time.Since(start))
		resultChan <- reportResult{reportType: "customers", data: data, err: nil}
	}()

	// 4. Table Report
	go func() {
		defer wg.Done()
		start := time.Now()
		log.Println("🪑 [Tables] Fetching table report...")
		data, err := c.repo.NewGetDefaultTableReport(ctx)
		if err != nil {
			log.Printf("❌ [Tables] Failed to fetch: %v", err)
			resultChan <- reportResult{reportType: "tables", data: nil, err: err}
			return
		}
		log.Printf("✅ [Tables] Fetched successfully in %v", time.Since(start))
		resultChan <- reportResult{reportType: "tables", data: data, err: nil}
	}()

	// 5. Staff Report
	go func() {
		defer wg.Done()
		start := time.Now()
		log.Println("👨‍💼 [Staff] Fetching staff report...")
		data, err := c.repo.NewGetDefaultStaffReport(ctx)
		if err != nil {
			log.Printf("❌ [Staff] Failed to fetch: %v", err)
			resultChan <- reportResult{reportType: "staff", data: nil, err: err}
			return
		}
		log.Printf("✅ [Staff] Fetched successfully in %v", time.Since(start))
		resultChan <- reportResult{reportType: "staff", data: data, err: nil}
	}()

	// 6. Raw Material Report
	go func() {
		defer wg.Done()
		start := time.Now()
		log.Println("📦 [RawMaterials] Fetching raw materials report...")
		data, err := c.repo.NewGetDefaultRawMaterialReport(ctx)
		if err != nil {
			log.Printf("❌ [RawMaterials] Failed to fetch: %v", err)
			resultChan <- reportResult{reportType: "raw_materials", data: nil, err: err}
			return
		}
		log.Printf("✅ [RawMaterials] Fetched successfully in %v", time.Since(start))
		resultChan <- reportResult{reportType: "raw_materials", data: data, err: nil}
	}()

	// Close result channel when all goroutines are done
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	var revenueData *models.NewDefaultRevenueResponse
	var salesData *models.NewDefaultSalesResponse
	var customersData *models.NewDefaultCustomerResponse
	var tablesData *models.NewDefaultTableResponse
	var staffsData *models.NewDefaultStaffResponse
	var rawMaterialsData *models.NewDefaultRawMaterialResponse

	var hasError bool
	var errorCount int

	for result := range resultChan {
		if result.err != nil {
			hasError = true
			errorCount++
			log.Printf("⚠️ [%s] Error collected: %v", result.reportType, result.err)
			continue
		}

		switch result.reportType {
		case "revenue":
			revenueData = result.data.(*models.NewDefaultRevenueResponse)
			log.Println("📊 [Revenue] Data stored in cache")
		case "sales":
			salesData = result.data.(*models.NewDefaultSalesResponse)
			log.Println("📈 [Sales] Data stored in cache")
		case "customers":
			customersData = result.data.(*models.NewDefaultCustomerResponse)
			log.Println("👥 [Customers] Data stored in cache")
		case "tables":
			tablesData = result.data.(*models.NewDefaultTableResponse)
			log.Println("🪑 [Tables] Data stored in cache")
		case "staff":
			staffsData = result.data.(*models.NewDefaultStaffResponse)
			log.Println("👨‍💼 [Staff] Data stored in cache")
		case "raw_materials":
			rawMaterialsData = result.data.(*models.NewDefaultRawMaterialResponse)
			log.Println("📦 [RawMaterials] Data stored in cache")
		}
	}

	totalDuration := time.Since(startTime)

	if hasError {
		log.Printf("⚠️ [DefaultRevenueCache] Cache reload completed with %d error(s) in %v", errorCount, totalDuration)
	} else {
		log.Printf("✅ [DefaultRevenueCache] Cache reload completed successfully in %v", totalDuration)
	}

	// Update cache with collected data (even if partial errors)
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = CachedDefaultRevenue{
		Revenue:      revenueData,
		Sales:        salesData,
		Customers:    customersData,
		Tables:       tablesData,
		Staffs:       staffsData,
		RawMaterials: rawMaterialsData,
		LastUpdated:  time.Now(),
		IsReady:      true, // Cache is ready even with partial data
	}

	log.Printf("📊 [DefaultRevenueCache] Cache updated at %s", time.Now().Format("2006-01-02 15:04:05"))

	// Log summary of what was loaded
	loadedCount := 0
	if revenueData != nil {
		loadedCount++
	}
	if salesData != nil {
		loadedCount++
	}
	if customersData != nil {
		loadedCount++
	}
	if tablesData != nil {
		loadedCount++
	}
	if staffsData != nil {
		loadedCount++
	}
	if rawMaterialsData != nil {
		loadedCount++
	}
	log.Printf("📊 [DefaultRevenueCache] Successfully loaded %d/6 report types", loadedCount)
}

// Getter methods remain the same
func (c *DefaultRevenueCache) GetSalesReport() (*models.NewDefaultSalesResponse, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.data.Sales, c.data.IsReady
}

func (c *DefaultRevenueCache) GetRevenueReport() (*models.NewDefaultRevenueResponse, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.data.Revenue, c.data.IsReady
}

func (c *DefaultRevenueCache) GetCustomerReport() (*models.NewDefaultCustomerResponse, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.data.Customers, c.data.IsReady
}

func (c *DefaultRevenueCache) GetTableReport() (*models.NewDefaultTableResponse, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.data.Tables, c.data.IsReady
}

func (c *DefaultRevenueCache) GetStaffReport() (*models.NewDefaultStaffResponse, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.data.Staffs, c.data.IsReady
}

func (c *DefaultRevenueCache) GetRawMaterialReport() (*models.NewDefaultRawMaterialResponse, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.data.RawMaterials, c.data.IsReady
}

func (c *DefaultRevenueCache) GetLastUpdated() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.data.LastUpdated
}

func (c *DefaultRevenueCache) IsReady() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.data.IsReady
}

func (c *DefaultRevenueCache) IsRefreshing() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isRefreshing
}
