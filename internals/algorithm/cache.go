package algorithm

import (
	"context"
	"fmt"
	"sync"

	"github.com/Abhishekh669/backend/internals/models"
	"github.com/Abhishekh669/backend/internals/repository"
)

type MenuCache struct {
	mu   sync.RWMutex
	data map[string]repository.CategoryMenuGroup
	repo repository.FoodCategoryRepo
}

func NewMenuCache(repo repository.FoodCategoryRepo) *MenuCache {
	return &MenuCache{
		data: make(map[string]repository.CategoryMenuGroup),
		repo: repo,
	}
}

func (c *MenuCache) ReloadFromDB() {
	menu, err := c.repo.GetAllMenuItemsGrouped(context.Background())
	if err != nil {
		fmt.Println("failed to load menu:", err)
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = menu // atomic swap
}

func (c *MenuCache) GetAll() map[string]repository.CategoryMenuGroup {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]repository.CategoryMenuGroup, len(c.data))

	for k, v := range c.data {
		itemsCopy := make([]models.MenuItemsResponse, len(v.MenuItems))
		copy(itemsCopy, v.MenuItems)

		result[k] = repository.CategoryMenuGroup{
			CategoryName: v.CategoryName,
			CategorySlug: v.CategorySlug,
			MenuItems:    itemsCopy,
		}
	}

	return result
}
