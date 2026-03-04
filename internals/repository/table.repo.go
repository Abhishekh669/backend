package repository

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/Abhishekh669/backend/internals/database"
	"github.com/Abhishekh669/backend/internals/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TableRepo interface {
	UpdateTables(ctx context.Context, table *models.UpdateTableStatus) error
	DeleteTables(ctx context.Context, tableIds []string) error
	CreateTables(ctx context.Context, tables []models.CreateTableType) error
	ViewTables(ctx context.Context) ([]models.TableStatus, error)
}

type tableRepo struct {
	pool *pgxpool.Pool
}

func (r *tableRepo) UpdateTables(ctx context.Context, table *models.UpdateTableStatus) error {
	if table == nil {
		return fmt.Errorf("table data is required")
	}

	// Begin transaction
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		log.Printf("failed to begin transaction: %v", err)
		return fmt.Errorf("failed to update table")
	}

	// Ensure rollback (safe even after commit)
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	// Update query
	query := `
		UPDATE table_status
		SET table_number = $1,
		    status = $2,
		    capacity = $3
		WHERE id = $4
	`

	cmdTag, err := tx.Exec(
		ctx,
		query,
		table.TableNumber,
		table.Status,
		table.Capacity,
		table.ID,
	)

	if err != nil {
		log.Printf("failed to update table: %v", err)

		// Check if it's a PostgreSQL error
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505": // unique_violation
				return fmt.Errorf("table number %d already exists", table.TableNumber)
			case "23503": // foreign_key_violation
				return fmt.Errorf("referenced record not found")
			case "23514": // check_violation
				return fmt.Errorf("invalid table data")
			default:
				log.Printf("PostgreSQL error: %s - %s", pgErr.Code, pgErr.Message)
				return fmt.Errorf("database error occurred")
			}
		}

		return fmt.Errorf("failed to update table")
	}

	// Check if any row was actually updated
	if cmdTag.RowsAffected() == 0 {
		log.Println("no table updated, invalid table id")
		return fmt.Errorf("table not found")
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		log.Printf("failed to commit transaction: %v", err)
		return fmt.Errorf("failed to update table")
	}

	return nil
}
func (r *tableRepo) DeleteTables(ctx context.Context, tableIds []string) error {
	if len(tableIds) < 1 {
		return fmt.Errorf("no tables selected")
	}

	// Begin transaction
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		log.Printf("failed to begin transaction: %v", err)
		return fmt.Errorf("failed to delete tables")
	}

	// Ensure rollback (safe even after commit)
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	// Bulk delete
	query := `DELETE FROM table_status WHERE id = ANY($1)`
	res, err := tx.Exec(ctx, query, tableIds)
	if err != nil {
		log.Printf("failed to delete tables: %v", err)
		return fmt.Errorf("failed to delete tables")
	}

	// Validate affected rows
	if res.RowsAffected() == 0 {
		log.Println("no tables deleted, maybe IDs are invalid")
		return fmt.Errorf("failed to delete tables")
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		log.Printf("failed to commit transaction: %v", err)
		return fmt.Errorf("failed to delete tables")
	}

	return nil
}

func (r *tableRepo) CreateTables(ctx context.Context, tables []models.CreateTableType) error {
	// Validate input
	if len(tables) < 1 {
		return fmt.Errorf("no tables provided")
	}

	if len(tables) > 50 {
		return fmt.Errorf("cannot create more than 50 tables at once")
	}

	// Begin a transaction
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Ensure rollback if something goes wrong
	defer func() {
		_ = tx.Rollback(ctx) // safe to call, ignored if already committed
	}()

	// Insert statement (id is auto-generated)
	query := `
		INSERT INTO table_status (table_number, status, capacity, created_at)
		VALUES ($1, $2, $3, NOW())
	`

	for i, table := range tables {
		// Validate each table
		if table.TableNumber <= 0 {
			return fmt.Errorf("invalid table number %d at index %d", table.TableNumber, i)
		}
		if table.Capacity <= 0 || table.Capacity > 20 {
			return fmt.Errorf("invalid capacity %d at index %d (must be 1-20)", table.Capacity, i)
		}
		if !isValidStatus(table.Status) {
			return fmt.Errorf("invalid status %s at index %d", table.Status, i)
		}

		_, err := tx.Exec(ctx, query,
			table.TableNumber,
			table.Status,
			table.Capacity,
		)

		if err != nil {
			log.Printf("failed to insert table_number %d: %v", table.TableNumber, err)

			// Check for PostgreSQL errors
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				switch pgErr.Code {
				case "23505": // unique_violation
					return fmt.Errorf("table number %d already exists", table.TableNumber)
				case "23514": // check_violation
					return fmt.Errorf("invalid data for table %d: %s", table.TableNumber, pgErr.Message)
				case "22001": // string_data_right_truncation
					return fmt.Errorf("data too long for table %d", table.TableNumber)
				default:
					return fmt.Errorf("database error for table %d: %s (code: %s)",
						table.TableNumber, pgErr.Message, pgErr.Code)
				}
			}

			return fmt.Errorf("failed to insert table_number %d: %w", table.TableNumber, err)
		}
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Helper function to validate status
func isValidStatus(status models.TableState) bool {
	switch status {
	case models.TableEmpty, models.TableOccupied, models.TableBooked:
		return true
	default:
		return false
	}
}

func (r *tableRepo) ViewTables(ctx context.Context) ([]models.TableStatus, error) {
	query := `
        SELECT 
            id, 
            table_number, 
            status, 
            capacity, 
            created_at 
        FROM table_status 
        ORDER BY table_number ASC
    `
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	var tables []models.TableStatus

	for rows.Next() {
		var table models.TableStatus
		var status string

		err := rows.Scan(
			&table.ID,
			&table.TableNumber,
			&status,
			&table.Capacity,
			&table.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan table row: %w", err)
		}

		table.Status = models.TableState(status)

		if !isValidTableState(table.Status) {
			table.Status = models.TableEmpty
		}

		tables = append(tables, table)
	}

	// Check for errors AFTER iterating through all rows
	if err = rows.Err(); err != nil {

		// Log the error for debugging
		fmt.Printf("Error iterating table rows: %v\n", err)

		// Return the error and any tables collected so far?
		// Usually better to return nil and the error to avoid partial data
		return nil, fmt.Errorf("error iterating table rows: %w", err)
	}

	// Ensure we return empty slice, not nil
	if tables == nil {
		tables = make([]models.TableStatus, 0)
	}

	return tables, nil
}

// Helper function to validate table state
func isValidTableState(state models.TableState) bool {
	switch state {
	case models.TableOccupied, models.TableEmpty, models.TableBooked:
		return true
	default:
		return false
	}
}

func NewTableRepository() TableRepo {
	pool, err := database.GetPostgresPool()

	if err != nil {
		return nil
	}

	return &tableRepo{
		pool: pool,
	}
}
