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
	Revenue     *models.NewDefaultRevenueResponse
	Sales       *models.NewDefaultSalesResponse
	Customers   *models.NewDefaultCustomerResponse
	Tables      *models.NewDefaultTableResponse
	LastUpdated time.Time
	IsReady     bool
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

func (c *DefaultRevenueCache) ReloadFromDB() {
	c.mu.Lock()
	c.isRefreshing = true
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		c.isRefreshing = false
		c.mu.Unlock()
	}()

	log.Println("📊 [DefaultRevenueCache] Starting cache reload...")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	data, err := c.repo.NewGetDefaultRevenueReport(ctx)
	if err != nil {
		log.Printf("❌ [DefaultRevenueCache] Failed to reload: %v", err)
		return
	}
	log.Printf("✅ [DefaultRevenueCache] Reload complete at %s", time.Now().Format("2006-01-02 15:04:05"))

	sales, err := c.repo.NewGetDefaultSalesReport(ctx)
	if err != nil {
		log.Printf("❌ [DefaultRevenueCache] Failed to reload sales data: %v", err)
		return
	}
	log.Printf("✅ [DefaultRevenueCache] Reload complete at %s", time.Now().Format("2006-01-02 15:04:05"))

	customers, err := c.repo.NewGetDefaultCustomerReport(ctx)
	if err != nil {
		log.Printf("❌ [DefaultRevenueCache] Failed to reload customer data: %v", err)
		return
	}
	log.Printf("✅ [DefaultRevenueCache] Reload complete at %s", time.Now().Format("2006-01-02 15:04:05"))

	tables, err := c.repo.NewGetDefaultTableReport(ctx)
	if err != nil {
		log.Printf("❌ [DefaultRevenueCache] Failed to reload table data: %v", err)
		return
	}
	log.Printf("✅ [DefaultRevenueCache] Reload complete at %s", time.Now().Format("2006-01-02 15:04:05"))

	c.mu.Lock()
	c.data = CachedDefaultRevenue{
		Revenue:     data,
		Sales:       sales,
		Customers:   customers,
		Tables:      tables,
		LastUpdated: time.Now(),
		IsReady:     true,
	}
	c.mu.Unlock()

}

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
