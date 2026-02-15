package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/Abhishekh669/backend/internals/database"
	"github.com/Abhishekh669/backend/internals/models"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GetCategoriesBySlug struct {
	Breadcrumb []models.Category `json:"breadcrumb"`
	Children   []models.Category `json:"children"`
	MenuItems  []models.MenuItem `json:"menu_items"`
}

type FoodCategoryRepo interface {
	UpdateMenuItems(ctx context.Context, menu_item *models.UpdateMenuItemType) error
	UpdateCategory(ctx context.Context, category *models.UpdateCategoryType) error
	DeleteMenuItems(ctx context.Context, menuItemIds []string) error
	DeleteCategories(ctx context.Context, categoryIds []string) error
	CreateMenuItems(ctx context.Context, menuItems []models.CreateMenuItemType, categoryId *uuid.UUID) error
	GetFoodCategoriesFromSlug(ctx context.Context, slugs []string) (*GetCategoriesBySlug, error)
	GetFoodCategory(ctx context.Context) ([]models.Category, error)
	CreateCateogry(ctx context.Context, slugPath []string, name string) error
}

type foodCategoryRepo struct {
	pool *pgxpool.Pool
}

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	re := regexp.MustCompile(`[^a-z0-9]+`)
	s = re.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

func (r *foodCategoryRepo) UpdateMenuItems(ctx context.Context, menu_item *models.UpdateMenuItemType) error {
	errorMessage := "failed to update menu item"
	if menu_item.Name == "" || menu_item.Price < 0 {
		return errors.New("menu item name cannot be empty")
	}

	query := `
		UPDATE menu_items
		SET
			name       = $1,
			description 	 = $2,
			price = $3,
			is_available = $4,
			image_url = $5,
			display_order =$6,
			updated_at = $7
		WHERE id = $9
	`
	res, err := r.pool.Exec(ctx, query,
		menu_item.Name,
		menu_item.Description,
		menu_item.Price,
		menu_item.IsAvailable,
		menu_item.ImageURL,
		menu_item.DisplayOrder,
		time.Now(),
		menu_item.ID,
	)
	if err != nil {
		log.Printf("failed to update menu item: %v", err)
		return errors.New(errorMessage)
	}
	if res.RowsAffected() == 0 {

		return errors.New("menu item not found")

	}
	return nil
}

func (r *foodCategoryRepo) UpdateCategory(ctx context.Context, category *models.UpdateCategoryType) error {
	slug := slugify(category.Name)
	errorMessage := "failed to update category"
	if category.Name == "" {
		return errors.New("category name cannot be empty")
	}

	query := `
		UPDATE categories
		SET
			name       = $1,
			slug 	 = $2,
			updated_at = $3
		WHERE id = $4
	`
	res, err := r.pool.Exec(ctx, query,
		category.Name,
		slug,
		time.Now(),
		category.ID,
	)
	if err != nil {
		log.Printf("failed to update category: %v", err)
		return errors.New(errorMessage)
	}
	if res.RowsAffected() == 0 {

		return errors.New("category not found")

	}
	return nil
}

func (r *foodCategoryRepo) DeleteMenuItems(ctx context.Context, menuItemIds []string) error {
	if len(menuItemIds) == 0 {
		return fmt.Errorf("no menu items selected")
	}

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		log.Printf("failed to begin transaction: %v", err)
		return fmt.Errorf("failed to delete menu items")
	}
	defer tx.Rollback(ctx)

	// ✅ Permanently delete the menu items
	res, err := tx.Exec(ctx, `
		DELETE FROM menu_items 
		WHERE id = ANY($1)
	`, menuItemIds)
	if err != nil {
		log.Printf("failed to delete menu items: %v", err)
		return fmt.Errorf("failed to delete menu items")
	}

	if res.RowsAffected() == 0 {
		return fmt.Errorf("no menu items were deleted")
	}

	// ✅ Commit the transaction
	if err := tx.Commit(ctx); err != nil {
		log.Printf("failed to commit transaction: %v", err)
		return fmt.Errorf("failed to delete menu items")
	}

	log.Printf("Successfully deleted %d menu items permanently", len(menuItemIds))
	return nil
}

func (r *foodCategoryRepo) DeleteCategories(ctx context.Context, categoryIds []string) error {
	if len(categoryIds) == 0 {
		return fmt.Errorf("no categories selected")
	}

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		log.Printf("failed to begin transaction: %v", err)
		return fmt.Errorf("failed to delete categories")
	}
	defer tx.Rollback(ctx)

	// ✅ Permanently delete the categories
	// This will automatically delete all child categories and menu items
	// due to ON DELETE CASCADE constraints
	res, err := tx.Exec(ctx, `
		DELETE FROM categories 
		WHERE id = ANY($1)
	`, categoryIds)
	if err != nil {
		log.Printf("failed to delete categories: %v", err)
		return fmt.Errorf("failed to delete categories")
	}

	if res.RowsAffected() == 0 {
		return fmt.Errorf("no categories were deleted")
	}

	// ✅ Commit the transaction
	if err := tx.Commit(ctx); err != nil {
		log.Printf("failed to commit transaction: %v", err)
		return fmt.Errorf("failed to delete categories")
	}

	log.Printf("Successfully deleted %d categories and all associated menu items permanently", len(categoryIds))
	return nil
}
func (r *foodCategoryRepo) CreateMenuItems(ctx context.Context, menuItems []models.CreateMenuItemType, categoryId *uuid.UUID) error {

	if len(menuItems) == 0 || categoryId == nil {
		return nil
	}

	fmt.Println("this is menu items : ", menuItems)

	const failedMessage = "failed to create menu items"

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		log.Printf("begin tx error: %v", err)
		return errors.New(failedMessage)
	}

	// rollback safety
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	query := `
		INSERT INTO menu_items (
			name,
			price,
			description,
			category_id,
			is_available,
			image_url,
			display_order,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7,	 NOW(), NOW())
	`

	for _, rm := range menuItems {
		_, err = tx.Exec(
			ctx,
			query,
			rm.Name,
			rm.Price,
			rm.Description,
			categoryId,
			rm.IsAvailable,
			rm.ImageURL,
			rm.DisplayOrder,
		)

		if err != nil {
			log.Printf("insert menu item failed: %+v", err)
			return errors.New(failedMessage)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		log.Printf("commit tx error: %v", err)
		return errors.New(failedMessage)
	}

	return nil
}

func (r *foodCategoryRepo) GetParentIDForSlug(ctx context.Context, slugs []string) (*uuid.UUID, error) {
	if len(slugs) == 0 {
		return nil, fmt.Errorf("slug cannot be empty")
	}

	var parentID *uuid.UUID = nil

	for _, slug := range slugs {
		var id uuid.UUID

		err := r.pool.QueryRow(ctx, `
			SELECT id FROM categories
			WHERE slug = $1
			AND parent_id IS NOT DISTINCT FROM $2
			AND is_active = true
		`, slug, parentID).Scan(&id)

		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				// This slug does not exist → parent is the last existing one
				return parentID, nil
			}
			return nil, err
		}

		parentID = &id
	}

	return parentID, nil
}

func (r *foodCategoryRepo) GetFoodCategoriesFromSlug(
	ctx context.Context,
	slugs []string,
) (*GetCategoriesBySlug, error) {

	if len(slugs) == 0 || len(slugs) > 5 {
		return nil, fmt.Errorf("invalid slug depth")
	}

	var parentID *uuid.UUID = nil
	var breadcrumb []models.Category

	// ---------------------------------------------------
	// 1️⃣ Resolve slug hierarchy
	// ---------------------------------------------------
	for _, slug := range slugs {

		var category models.Category

		err := r.pool.QueryRow(ctx, `
			SELECT id, name, slug, parent_id, level,
			       is_active, display_order,
			       created_at, updated_at
			FROM categories
			WHERE slug = $1
			AND parent_id IS NOT DISTINCT FROM $2
			AND is_active = true
		`, slug, parentID).Scan(
			&category.ID,
			&category.Name,
			&category.Slug,
			&category.ParentID,
			&category.Level,
			&category.IsActive,
			&category.DisplayOrder,
			&category.CreatedAt,
			&category.UpdatedAt,
		)

		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, fmt.Errorf("category not found")
			}
			return nil, err
		}

		breadcrumb = append(breadcrumb, category)
		parentID = &category.ID
	}

	// ---------------------------------------------------
	// 2️⃣ Fetch children of last category
	// ---------------------------------------------------
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, slug, parent_id, level,
		       is_active, display_order,
		       created_at, updated_at
		FROM categories
		WHERE parent_id = $1
		AND is_active = true
		ORDER BY display_order
	`, *parentID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var children []models.Category

	for rows.Next() {
		var child models.Category
		err := rows.Scan(
			&child.ID,
			&child.Name,
			&child.Slug,
			&child.ParentID,
			&child.Level,
			&child.IsActive,
			&child.DisplayOrder,
			&child.CreatedAt,
			&child.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		children = append(children, child)
	}

	fmt.Println("this is parent id : ", parentID)

	var menuItems []models.MenuItem

	menuRows, err := r.pool.Query(ctx, `
			SELECT id, name, description, price,
			       category_id, is_available,
			       image_url, display_order,
			       created_at, updated_at
			FROM menu_items
			WHERE category_id = $1
			ORDER BY display_order
		`, *parentID)

	if err != nil {
		return nil, err
	}
	defer menuRows.Close()

	for menuRows.Next() {
		var item models.MenuItem

		err := menuRows.Scan(
			&item.ID,
			&item.Name,
			&item.Description,
			&item.Price,
			&item.CategoryID,
			&item.IsAvailable,
			&item.ImageURL,
			&item.DisplayOrder,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		menuItems = append(menuItems, item)
	}

	fmt.Println("this is children : ", menuItems)
	return &GetCategoriesBySlug{
		Breadcrumb: breadcrumb,
		Children:   children,
		MenuItems:  menuItems,
	}, nil
}

func (r *foodCategoryRepo) GetFoodCategory(ctx context.Context) ([]models.Category, error) {
	query := `
		SELECT
			id,
			name,
			slug,
			parent_id,
			level,
			is_active,
			display_order,
			created_at,
			updated_at
		FROM categories
		WHERE is_active = TRUE
		AND parent_id IS NULL
		ORDER BY level ASC, display_order ASC, created_at ASC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := make([]models.Category, 0)

	for rows.Next() {
		var c models.Category

		err := rows.Scan(
			&c.ID,
			&c.Name,
			&c.Slug,
			&c.ParentID, // *uuid.UUID handles NULL
			&c.Level,
			&c.IsActive,
			&c.DisplayOrder,
			&c.CreatedAt,
			&c.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		categories = append(categories, c)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return categories, nil
}

// func (r *foodCategoryRepo) CreateCategory(
// 	ctx context.Context,
// 	name string,
// 	parentID *uuid.UUID, // nil for root category
// ) error {

// 	slug := slugify(name)

// 	fmt.Println("this is hte new slug : ", slug)

// 	level := 1
// 	if parentID != nil {
// 		// fetch parent level
// 		err := r.pool.QueryRow(
// 			ctx,
// 			`SELECT level FROM categories WHERE id = $1`,
// 			parentID,
// 		).Scan(&level)

// 		if err != nil {

// 			log.Println("error in creating category : ", err)
// 			return fmt.Errorf("failed to fetch parent category: %w", err)
// 		}

// 		level = level + 1
// 	}

// 	_, err := r.pool.Exec(
// 		ctx,
// 		`
// 		INSERT INTO categories (
// 			name,
// 			slug,
// 			parent_id,
// 			level
// 		)
// 		VALUES ($1, $2, $3, $4)
// 		`,
// 		name,
// 		slug,
// 		parentID,
// 		level,
// 	)

// 	if err != nil {
// 		log.Println("error in creating category : ", err)

// 		var pgErr *pgconn.PgError
// 		if errors.As(err, &pgErr) {

// 			// 23505 = unique_violation
// 			if pgErr.Code == "23505" {

// 				switch pgErr.ConstraintName {
// 				case "uq_root_category_slug":
// 					return fmt.Errorf(" category with this name already exists")

// 				case "categories_slug_key":
// 					return fmt.Errorf("a category with this slug already exists")

// 				default:
// 					return fmt.Errorf("category is already created")
// 				}
// 			}
// 		}

// 		return fmt.Errorf("failed to create category: %w", err)
// 	}

// 	return nil
// }

func (r *foodCategoryRepo) CreateCateogry(ctx context.Context, slugPath []string, name string) error {
	if len(slugPath) > 5 {
		return fmt.Errorf("cannot create category beyonad level 5")
	}
	fmt.Println("this is the naem : ", name)
	var parentId *uuid.UUID = nil
	level := 1

	if len(slugPath) > 0 {
		var err error
		parentId, err = r.GetParentIDForSlug(ctx, slugPath)
		if err != nil {
			return fmt.Errorf("failed to find parent category: %w", err)
		}

		if parentId == nil {
			return fmt.Errorf("parent category does not exist for slug path: %v", slugPath)
		}

		err = r.pool.QueryRow(ctx, `
				SELECT level FROM categories WHERE id = $1
			`, parentId).Scan(&level)
		if err != nil {
			return fmt.Errorf("failed to fetch parent level: %w", err)
		}
		level++

	}

	if level > 5 {
		return fmt.Errorf("cannot create category beyond level 5")
	}

	slug := slugify(name)
	_, err := r.pool.Exec(ctx, `
		INSERT INTO categories (
			name,
			slug,
			parent_id,
			level
		) 
		VALUES ($1, $2, $3, $4)
	`, name, slug, parentId, level)

	if err != nil {
		fmt.Println("error in creating category : ", err)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.ConstraintName {
			case "uq_root_category_slug":
				return fmt.Errorf("root category with this slug already exists")
			case "uq_child_category_slug":
				return fmt.Errorf("a category with this slug already exists under the parent")
			default:
				return fmt.Errorf("category already exists")
			}
		}
		return fmt.Errorf("failed to create category: %w", err)
	}
	return nil
}

func NewFoodCategoryRepository() FoodCategoryRepo {
	pool, err := database.GetPostgresPool()

	if err != nil {
		return nil
	}

	return &foodCategoryRepo{
		pool: pool,
	}
}
