package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/Abhishekh669/backend/internals/database"
	"github.com/Abhishekh669/backend/internals/models"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FoodCategoryRepo interface {
	GetFoodCategory(ctx context.Context) ([]models.Category, error)
	CreateCategory(c context.Context, name string, parentID *uuid.UUID) error
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

func (r *foodCategoryRepo) CreateCategory(
	ctx context.Context,
	name string,
	parentID *uuid.UUID, // nil for root category
) error {

	slug := slugify(name)

	fmt.Println("this is hte new slug : ", slug)

	level := 1
	if parentID != nil {
		// fetch parent level
		err := r.pool.QueryRow(
			ctx,
			`SELECT level FROM categories WHERE id = $1`,
			parentID,
		).Scan(&level)

		if err != nil {

			log.Println("error in creating category : ", err)
			return fmt.Errorf("failed to fetch parent category: %w", err)
		}

		level = level + 1
	}

	_, err := r.pool.Exec(
		ctx,
		`
		INSERT INTO categories (
			name,
			slug,
			parent_id,
			level
		)
		VALUES ($1, $2, $3, $4)
		`,
		name,
		slug,
		parentID,
		level,
	)

	if err != nil {
		log.Println("error in creating category : ", err)

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {

			// 23505 = unique_violation
			if pgErr.Code == "23505" {

				switch pgErr.ConstraintName {
				case "uq_root_category_slug":
					return fmt.Errorf(" category with this name already exists")

				case "categories_slug_key":
					return fmt.Errorf("a category with this slug already exists")

				default:
					return fmt.Errorf("category is already created")
				}
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
