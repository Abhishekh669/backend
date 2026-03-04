package algorithm

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/Abhishekh669/backend/internals/models"
	"github.com/Abhishekh669/backend/internals/repository"
	"github.com/gofrs/uuid"
)

type MenuCache struct {
	mu               sync.RWMutex
	Categories       []models.CategoryCache
	CategoryChildren map[uuid.UUID][]uuid.UUID
	ItemsByCategory  map[uuid.UUID][]models.MenuItemCache
	repo             repository.FoodCategoryRepo
}

func NewMenuCache(repo repository.FoodCategoryRepo) *MenuCache {
	return &MenuCache{
		Categories:       []models.CategoryCache{},
		CategoryChildren: make(map[uuid.UUID][]uuid.UUID),
		ItemsByCategory:  make(map[uuid.UUID][]models.MenuItemCache),
		repo:             repo,
	}
}

/* =========================
   LOAD / RELOAD
========================= */

func (c *MenuCache) ReloadFromDB() {
	c.mu.Lock()
	defer c.mu.Unlock()
	cat, err := c.repo.GetAllCategoriesFromDB(context.Background())

	if err != nil {
		fmt.Println("failed to load categgory")
		return
	}

	menu, err := c.repo.GetAllMenuItemsFromDB(context.Background())
	if err != nil {
		fmt.Println("failed to load menu items")
		return
	}
	c.Categories = cat
	c.CategoryChildren = make(map[uuid.UUID][]uuid.UUID)
	c.ItemsByCategory = make(map[uuid.UUID][]models.MenuItemCache)

	for _, cat := range cat {
		if cat.ParentID != nil {
			c.CategoryChildren[*cat.ParentID] =
				append(c.CategoryChildren[*cat.ParentID], cat.ID)
		}
	}

	for _, item := range menu {
		c.ItemsByCategory[item.CategoryID] =
			append(c.ItemsByCategory[item.CategoryID], item)
	}
}

/* =========================
   READ METHODS
========================= */

// All menu items (no category filter)
func (c *MenuCache) GetAllMenuItems() []models.MenuItemCache {
	c.mu.RLock()
	defer c.mu.RUnlock()

	all := make([]models.MenuItemCache, 0)
	for _, items := range c.ItemsByCategory {
		all = append(all, items...)
	}
	return all
}

// Menu items by category + subcategories
func (c *MenuCache) GetMenuItemsByCategory(
	categoryID uuid.UUID,
) []models.MenuItemCache {
	c.mu.RLock()
	defer c.mu.RUnlock()

	descendants := make(map[uuid.UUID]struct{})
	c.collectDescendants(categoryID, descendants)

	items := make([]models.MenuItemCache, 0)
	for cid := range descendants {
		items = append(items, c.ItemsByCategory[cid]...)
	}
	return items
}

func (c *MenuCache) collectDescendants(
	id uuid.UUID,
	out map[uuid.UUID]struct{},
) {
	out[id] = struct{}{}
	for _, child := range c.CategoryChildren[id] {
		if _, exists := out[child]; !exists {
			c.collectDescendants(child, out)
		}
	}
}

/* =========================
   DELETE METHODS
========================= */

func (c *MenuCache) DeleteCategory(categoryID uuid.UUID) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	descendants := make(map[uuid.UUID]struct{})
	c.collectDescendants(categoryID, descendants)

	// Remove categories
	filtered := make([]models.CategoryCache, 0)
	for _, cat := range c.Categories {
		if _, del := descendants[cat.ID]; !del {
			filtered = append(filtered, cat)
		}
	}
	c.Categories = filtered

	// Remove children mappings
	for id := range descendants {
		delete(c.CategoryChildren, id)
		delete(c.ItemsByCategory, id)
	}

	for pid, children := range c.CategoryChildren {
		newChildren := []uuid.UUID{}
		for _, cid := range children {
			if _, del := descendants[cid]; !del {
				newChildren = append(newChildren, cid)
			}
		}
		c.CategoryChildren[pid] = newChildren
	}

	return nil
}

func (c *MenuCache) DeleteMenuItem(itemID uuid.UUID) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for cid, items := range c.ItemsByCategory {
		newItems := make([]models.MenuItemCache, 0)
		found := false

		for _, item := range items {
			if item.ID == itemID {
				found = true
				continue
			}
			newItems = append(newItems, item)
		}

		if found {
			c.ItemsByCategory[cid] = newItems
			return nil
		}
	}

	return errors.New("menu item not found")
}

/* =========================
   UPDATE METHODS
========================= */

func (c *MenuCache) UpdateCategory(updated models.CategoryCache) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i, cat := range c.Categories {
		if cat.ID == updated.ID {
			c.Categories[i] = updated
			c.rebuildCategoryChildren()
			return nil
		}
	}
	return errors.New("category not found")
}

func (c *MenuCache) UpdateMenuItem(updated models.MenuItemCache) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	items := c.ItemsByCategory[updated.CategoryID]
	for i, item := range items {
		if item.ID == updated.ID {
			items[i] = updated
			c.ItemsByCategory[updated.CategoryID] = items
			return nil
		}
	}
	return errors.New("menu item not found")
}

func (c *MenuCache) rebuildCategoryChildren() {
	c.CategoryChildren = make(map[uuid.UUID][]uuid.UUID)
	for _, cat := range c.Categories {
		if cat.ParentID != nil {
			c.CategoryChildren[*cat.ParentID] =
				append(c.CategoryChildren[*cat.ParentID], cat.ID)
		}
	}
}

func (c *MenuCache) AddCategory(category models.CategoryCache) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if category already exists
	for _, cat := range c.Categories {
		if cat.ID == category.ID {
			return errors.New("category already exists")
		}
	}

	// Add to categories slice
	c.Categories = append(c.Categories, category)

	// Update CategoryChildren map if it has a parent
	if category.ParentID != nil {
		c.CategoryChildren[*category.ParentID] = append(c.CategoryChildren[*category.ParentID], category.ID)
	}

	return nil
}

func (c *MenuCache) AddSubCategory(parentID uuid.UUID, subCategory models.CategoryCache) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Verify parent exists
	parentExists := false
	for _, cat := range c.Categories {
		if cat.ID == parentID {
			parentExists = true
			break
		}
	}
	if !parentExists {
		return errors.New("parent category not found")
	}

	// Check if subcategory already exists
	for _, cat := range c.Categories {
		if cat.ID == subCategory.ID {
			return errors.New("subcategory already exists")
		}
	}

	// Set the parent ID for the subcategory
	subCategory.ParentID = &parentID

	// Add to categories slice
	c.Categories = append(c.Categories, subCategory)

	// Add to CategoryChildren map
	c.CategoryChildren[parentID] = append(c.CategoryChildren[parentID], subCategory.ID)

	return nil
}

func (c *MenuCache) AddMenuItem(item models.MenuItemCache) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Verify category exists
	categoryExists := false
	for _, cat := range c.Categories {
		if cat.ID == item.CategoryID {
			categoryExists = true
			break
		}
	}
	if !categoryExists {
		return errors.New("category not found for menu item")
	}

	// Check if item already exists in the category
	for _, existingItem := range c.ItemsByCategory[item.CategoryID] {
		if existingItem.ID == item.ID {
			return errors.New("menu item already exists in this category")
		}
	}

	// Add to ItemsByCategory map
	c.ItemsByCategory[item.CategoryID] = append(c.ItemsByCategory[item.CategoryID], item)

	return nil
}

func (c *MenuCache) GetFullMenuSnapshot() (
	[]models.CategoryCache,
	map[uuid.UUID][]uuid.UUID,
	map[uuid.UUID][]models.MenuItemCache,
) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// shallow copies (safe for read-only response)
	categories := make([]models.CategoryCache, len(c.Categories))
	copy(categories, c.Categories)

	categoryChildren := make(map[uuid.UUID][]uuid.UUID)
	for k, v := range c.CategoryChildren {
		categoryChildren[k] = append([]uuid.UUID{}, v...)
	}

	itemsByCategory := make(map[uuid.UUID][]models.MenuItemCache)
	for k, v := range c.ItemsByCategory {
		itemsByCategory[k] = append([]models.MenuItemCache{}, v...)
	}

	return categories, categoryChildren, itemsByCategory
}
