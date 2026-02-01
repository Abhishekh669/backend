package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Abhishekh669/backend/internals/database"
	"github.com/Abhishekh669/backend/internals/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RawMaterialsRepo interface {
	DeleteRawMaterials(ctx context.Context, rawMaterialsIds []string) error
	UpdateRawMaterial(ctx context.Context, updatedData *models.UpdateRawMaterials) error
	GetRawMaterials(ctx context.Context, limit, page int, fromDate *time.Time, toDate *time.Time, startingPrice, endingPrice int32, search string, oldFirst bool) (*RawMaterialsResponse, error)
	GetRawMaterialStats(ctx context.Context) (*RawMaterialStats, error)
	CreateRawMaterials(ctx context.Context, rawMaterials []models.CreateRawMaterialType) error
}

type rawMaterialRepo struct {
	pool *pgxpool.Pool
}

type RawMaterialStats struct {
	TotalMaterials int32   `json:"total_materials"`
	TotalQuantity  float64 `json:"total_quantity"`
	TotalPrice     float64 `json:"total_price"`
	RecentPrice    float64 `json:"recent_price"`
}

type RawMaterialsResponse struct {
	RawMaterials    []models.RawMaterials `json:"raw_materials"`
	Total           int                   `json:"total"`
	HasMore         bool                  `json:"has_more"`
	NextOffset      int                   `json:"next_offset"`
	RawMaterialData *RawMaterialStats     `json:"raw_materials_data"`
}

func (r *rawMaterialRepo) DeleteRawMaterials(ctx context.Context, rawMaterialsIds []string) error {
	if len(rawMaterialsIds) < 1 {
		return fmt.Errorf("no raw materials selected")
	}

	// Begin a transaction
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		log.Println("failed to begin transaction: %w", err)
		return fmt.Errorf("failed to deletet raw materials")
	}

	// Defer rollback in case of panic or early return
	defer func() {
		_ = tx.Rollback(ctx) // safe to call even if committed
	}()

	// Perform bulk delete
	query := `DELETE FROM raw_material WHERE id = ANY($1)`
	res, err := tx.Exec(ctx, query, rawMaterialsIds)
	if err != nil {
		log.Println("failed to delete raw materials: %w", err)
		return fmt.Errorf("failed to delete raw materials")
	}

	// Check how many rows were actually deleted
	if res.RowsAffected() == 0 {
		log.Println("no raw materials deleted, maybe IDs are invalid")
		return fmt.Errorf("failed to delete raw materials")
	}

	// Commit transaction
	if commitErr := tx.Commit(ctx); commitErr != nil {
		log.Println("failed to commit transaction: %w", commitErr)
		return fmt.Errorf("failed to delete raw materials")
	}

	return nil
}

func (r *rawMaterialRepo) UpdateRawMaterial(ctx context.Context, updatedData *models.UpdateRawMaterials) error {
	query := `
		UPDATE raw_material
		SET
			name       = $1,
			price      = $2,
			quantity      = $3,
			unit       = $4,
			updated_at = $5
		WHERE id = $6
	`
	res, err := r.pool.Exec(ctx, query,
		updatedData.Name,
		updatedData.Price,
		updatedData.Quantity,
		updatedData.Unit,
		time.Now(),
		updatedData.Id,
	)
	if err != nil {
		fmt.Println("this is the error ; ", err)
		return fmt.Errorf("failed to update raw materials")
	}

	if res.RowsAffected() == 0 {
		return fmt.Errorf("no raw materials updated, maybe ID is invalid")
	}

	return nil

}

func (r *rawMaterialRepo) GetRawMaterialStats(ctx context.Context) (*RawMaterialStats, error) {
	const query = `
		SELECT 
			COUNT(*)::INT AS total_materials,
			COALESCE(SUM(quantity), 0) AS total_quantity,
			COALESCE(SUM(price * quantity), 0) AS total_price,
			COALESCE(
				SUM(
					CASE
						WHEN created_at >= NOW() - INTERVAL '7 days'
						THEN price * quantity
						ELSE 0
					END
				),
			0) AS recent_price
		FROM raw_material;
	`

	var stats RawMaterialStats

	err := r.pool.
		QueryRow(ctx, query).
		Scan(
			&stats.TotalMaterials,
			&stats.TotalQuantity,
			&stats.TotalPrice,
			&stats.RecentPrice,
		)

	if err != nil {
		log.Println("error in getting raw materials stats : %w", err)
		return nil, errors.New("faield to get stats")
	}

	return &stats, nil
}

func (r *rawMaterialRepo) GetRawMaterials(ctx context.Context, limit, page int, fromDate *time.Time, toDate *time.Time, startingPrice, endingPrice int32, search string, oldFirst bool) (*RawMaterialsResponse, error) {
	fmt.Println("thisis hte queruy : ", limit, page, fromDate, toDate, search, startingPrice, endingPrice)
	errMessage := "failed to get raw materials"
	offset := page * limit
	orderBy := "DESC"

	if oldFirst {
		orderBy = "ASC"
	}

	listQuery := fmt.Sprintf(`
	SELECT
		id, name, price, quantity, unit, created_at, updated_at
	FROM raw_material
	WHERE
		($1::timestamptz IS NULL OR created_at >= $1::timestamptz)
		AND ($2::timestamptz IS NULL OR created_at <= $2::timestamptz)
		AND price BETWEEN $3 AND $4
		AND ($5 = '' OR name ILIKE '%%' || $5 || '%%')
	ORDER BY created_at %s
	LIMIT $6 OFFSET $7;
`, orderBy)

	rows, err := r.pool.Query(
		ctx,
		listQuery,
		fromDate,
		toDate,
		startingPrice,
		endingPrice,
		search,
		limit,
		offset,
	)

	if err != nil {
		log.Printf("error in getting raw materials : %v", err)
		return nil, errors.New(errMessage)
	}
	defer rows.Close()

	rawMaterials := make([]models.RawMaterials, 0, limit)
	for rows.Next() {
		var rm models.RawMaterials
		if err := rows.Scan(
			&rm.Id,
			&rm.Name,
			&rm.Price,
			&rm.Quantity,
			&rm.Unit,
			&rm.CreatedAt,
			&rm.UpdatedAt,
		); err != nil {
			return nil, err
		}
		rawMaterials = append(rawMaterials, rm)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.New(errMessage)
	}

	const countQuery = `
	SELECT COUNT(*)
	FROM raw_material
	WHERE
		($1::timestamptz IS NULL OR created_at >= $1::timestamptz)
		AND ($2::timestamptz IS NULL OR created_at <= $2::timestamptz)
		AND price BETWEEN $3 AND $4
		AND ($5 = '' OR name ILIKE '%' || $5 || '%');
`

	var total int
	err = r.pool.QueryRow(
		ctx,
		countQuery,
		fromDate,
		toDate,
		startingPrice,
		endingPrice,
		search,
	).Scan(&total)
	if err != nil {
		return nil, errors.New(errMessage)
	}

	stats, err := r.GetRawMaterialStats(ctx)
	if err != nil {
		return nil, errors.New(errMessage)
	}

	// ---------- 6. Pagination Info ----------
	hasMore := (page+1)*limit < total
	nextPage := page + 1

	// ---------- 7. Final Response ----------
	response := &RawMaterialsResponse{
		RawMaterials:    rawMaterials,
		Total:           total,
		HasMore:         hasMore,
		NextOffset:      nextPage,
		RawMaterialData: stats,
	}

	return response, nil

}

func (r *rawMaterialRepo) CreateRawMaterials(ctx context.Context, rawMaterials []models.CreateRawMaterialType) error {

	if len(rawMaterials) == 0 {
		return nil
	}

	fmt.Println("this is raw mateirals : ", rawMaterials)

	const failedMessage = "failed to create raw materials"

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
		INSERT INTO raw_material (
			name,
			price,
			quantity,
			unit,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
	`

	for _, rm := range rawMaterials {
		fmt.Println("thisis quentityt : ", rm.Quantity)
		_, err = tx.Exec(
			ctx,
			query,
			rm.Name,
			rm.Price,
			float64(rm.Quantity),
			rm.Unit,
		)
		if err != nil {
			log.Printf("insert raw_material failed: %+v", err)
			return errors.New(failedMessage)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		log.Printf("commit tx error: %v", err)
		return errors.New(failedMessage)
	}

	return nil
}

func NewRawMaterialsRepository() RawMaterialsRepo {
	pool, err := database.GetPostgresPool()

	if err != nil {
		return nil
	}

	return &rawMaterialRepo{
		pool: pool,
	}
}
