package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Abhishekh669/backend/internals/database"
	"github.com/Abhishekh669/backend/internals/models"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// OrderRepo interface defines all order-related operations
type OrderRepo interface {
	DeleteTablesSessionById(ctx context.Context, tableSessionId *uuid.UUID) error
	GetTableValidationByTableAndPhone(ctx context.Context, tableNumber int, phoneNumber string) (*models.TableValidation, error)
	GetTableValidationByID(ctx context.Context, id uuid.UUID) (*models.TableValidation, error)
	GetUnassignedTables(ctx context.Context) ([]models.TableValidation, error)
	DeleteTableApprovalByID(ctx context.Context, id uuid.UUID) error
	ApproveTableByWaiter(ctx context.Context, req *models.WaiterApprovalRequest) error
	CreateNewApprovalRequest(ctx context.Context, req *models.CustomerApprovalRequest) (*models.TableValidation, error)
	NewGetAllOrderForStatus(ctx context.Context) ([]models.CustomerOrderRequest, error)
	NewGetAllOrderRequest(ctx context.Context) ([]models.CustomerOrderRequest, error)
	NewGetTableSessionByTableAndPhone(ctx context.Context, tableNumber int, customerPhone string) (*models.CustomerOrderRequest, error)
	NewGetTableSessionByID(ctx context.Context, tableSessionID uuid.UUID) (*models.CustomerOrderRequest, error)
	NewApproveCustomerRequest(ctx context.Context, approveOrder *models.ApproveOrderType) (err error)
	NewCreateCustomerOrder(ctx context.Context, customerOrder *models.CreateCustomerOrderRequest) error
	GetTableSessionByTableAndPhone(ctx context.Context, tableNumber int, customerPhone string) (*models.CustomerOrderRequest, error)
	GetTableSessionByID(ctx context.Context, tableSessionID uuid.UUID) (*models.CustomerOrderRequest, error)
	GetAllOrderRequest(ctx context.Context) ([]models.CustomerOrderRequest, error)
	ApproveCustomerRequest(ctx context.Context, approveOrder *models.ApproveOrderType) error
	GetTableStatus(ctx context.Context, tableNumber int) (*models.TableSession, error)
	CreateCustomerOrder(ctx context.Context, customerOrder *models.CreateCustomerOrderRequest) error
}

// orderRepo implements OrderRepo interface
type orderRepo struct {
	pool *pgxpool.Pool
}

func (r *orderRepo) DeleteTablesSessionById(ctx context.Context, tableSessionId *uuid.UUID) error {

	query := `
		DELETE FROM table_session
		WHERE id = $1
		RETURNING id
	`

	var deletedID uuid.UUID
	err := r.pool.QueryRow(ctx, query, tableSessionId).Scan(&deletedID)
	if err != nil {
		return err
	}

	return nil
}

func (r *orderRepo) GetTableValidationByTableAndPhone(ctx context.Context, tableNumber int, phoneNumber string) (*models.TableValidation, error) {

	var table models.TableValidation

	query := `
		SELECT id, table_number, phone_number, waiter_id, created_at, updated_at
		FROM table_validation
		WHERE table_number = $1 AND phone_number = $2
		ORDER BY created_at DESC
		LIMIT 1
	`

	err := r.pool.QueryRow(ctx, query, tableNumber, phoneNumber).Scan(
		&table.ID,
		&table.TableNumber,
		&table.PhoneNumber,
		&table.WaiterID,
		&table.CreatedAt,
		&table.UpdatedAt,
	)

	fmt.Println("error in getting value table -val  : ", err)

	if err != nil {
		return nil, errors.New("table validation not found")
	}

	return &table, nil
}

// GetTableValidationByID fetches a table_validation record by its ID
func (r *orderRepo) GetTableValidationByID(ctx context.Context, id uuid.UUID) (*models.TableValidation, error) {
	var table models.TableValidation

	query := `
		SELECT id, table_number, phone_number, waiter_id, created_at, updated_at
		FROM table_validation
		WHERE id = $1
	`

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&table.ID,
		&table.TableNumber,
		&table.PhoneNumber,
		&table.WaiterID,
		&table.CreatedAt,
		&table.UpdatedAt,
	)

	if err != nil {
		return nil, errors.New("table validation not found")
	}

	return &table, nil
}

func (r *orderRepo) GetUnassignedTables(ctx context.Context) ([]models.TableValidation, error) {
	query := `
		SELECT 
			id,
			table_number,
			phone_number,
			waiter_id,
			created_at,
			updated_at
		FROM table_validation
		WHERE waiter_id IS NULL
		ORDER BY created_at ASC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tables := []models.TableValidation{}

	for rows.Next() {
		var t models.TableValidation
		err := rows.Scan(
			&t.ID,
			&t.TableNumber,
			&t.PhoneNumber,
			&t.WaiterID,
			&t.CreatedAt,
			&t.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		tables = append(tables, t)
	}

	return tables, nil
}

func (r *orderRepo) DeleteTableApprovalByID(ctx context.Context, id uuid.UUID) error {
	query := `
		DELETE FROM table_validation
		WHERE id = $1
		RETURNING id
	`

	var deletedID uuid.UUID
	err := r.pool.QueryRow(ctx, query, id).Scan(&deletedID)
	if err != nil {
		return err
	}

	return nil
}

// TODO : run a go routein for dleign the table sesiosn in which teh waiter is not assigned or create time ois mroe than 10 miutes
// ApproveTableByWaiter assigns a waiter to a table validation request
func (r *orderRepo) ApproveTableByWaiter(ctx context.Context, req *models.WaiterApprovalRequest) error {
	query := `
		UPDATE table_validation
		SET 
			waiter_id = $1,
			table_number = $2,
			phone_number = $3,
			updated_at = NOW()
		WHERE id = $4
		RETURNING id
	`

	var id uuid.UUID
	err := r.pool.QueryRow(ctx, query, req.WaiterId, req.TableNumber, req.Phone, req.Id).Scan(&id)
	if err != nil {
		return err
	}

	return nil
}

// CreateNewApprovalRequest inserts a new table validation record and returns the created row
func (r *orderRepo) CreateNewApprovalRequest(ctx context.Context, req *models.CustomerApprovalRequest) (*models.TableValidation, error) {
	fmt.Println("this is request in repo: ", req)

	// Start a transaction
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Step 1: Check if an active table session already exists for this table number
	tableSession := &models.TableSession{}
	checkSessionQuery := `
		SELECT id, table_number, open_time, close_time, status, created_at, updated_at
		FROM table_session
		WHERE open_time IS NOT NULL
		AND close_time IS NULL
		AND table_number = $1
	`
	err = tx.QueryRow(ctx, checkSessionQuery, req.TableNumber).Scan(
		&tableSession.ID,
		&tableSession.TableNumber,
		&tableSession.OpenTime,
		&tableSession.CloseTime,
		&tableSession.Status,
		&tableSession.CreatedAt,
		&tableSession.UpdatedAt,
	)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	// If a session exists, table is already booked
	if err == nil {
		return nil, errors.New("table is already booked")
	}

	// Step 2: No active session found — create a new table session
	createSessionQuery := `
		INSERT INTO table_session (table_number, open_time, status)
		VALUES ($1, NOW(), 'occupied')
		RETURNING id, table_number, open_time, close_time, status, created_at, updated_at
	`
	err = tx.QueryRow(ctx, createSessionQuery, req.TableNumber).Scan(
		&tableSession.ID,
		&tableSession.TableNumber,
		&tableSession.OpenTime,
		&tableSession.CloseTime,
		&tableSession.Status,
		&tableSession.CreatedAt,
		&tableSession.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Step 3: Update table_status to 'occupied'
	updateTableStatusQuery := `
		UPDATE table_status
		SET status = $1
		WHERE table_number = $2
	`
	_, err = tx.Exec(ctx, updateTableStatusQuery, models.TableOccupied, req.TableNumber)
	if err != nil {
		return nil, err
	}

	// Step 4: Create the approval request in table_validation
	insertValidationQuery := `
		INSERT INTO table_validation (table_number, phone_number)
		VALUES ($1, $2)
		RETURNING id, table_number, phone_number, waiter_id, created_at, updated_at
	`
	var tv models.TableValidation
	err = tx.QueryRow(ctx, insertValidationQuery, req.TableNumber, req.Phone).Scan(
		&tv.ID,
		&tv.TableNumber,
		&tv.PhoneNumber,
		&tv.WaiterID,
		&tv.CreatedAt,
		&tv.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Commit the transaction
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &tv, nil
}

func (r *orderRepo) NewGetAllOrderForStatus(ctx context.Context) ([]models.CustomerOrderRequest, error) {
	query := `
        WITH filtered_orders AS (
            -- First, get all orders that have at least one not-approved item
            SELECT DISTINCT o.id
            FROM orders o
            INNER JOIN order_items oi ON oi.order_id = o.id
            WHERE oi.status NOT IN ('not-approved', 'completed')
        ),
        session_data AS (
            SELECT 
                ts.id,
                ts.table_number,
                ts.open_time,
                ts.close_time,
                ts.status,
                ts.created_at,
                ts.updated_at
            FROM table_session ts
            WHERE ts.status = 'occupied' 
                AND ts.open_time IS NOT NULL
                AND ts.close_time IS NULL
        )
        SELECT 
            json_agg(
                json_build_object(
                    'id', o.id,
                    'status', o.status,
                    'customer_name', o.customer_name,
                    'customer_phone', o.customer_phone,
                    'note', o.note,
                    'table_session', json_build_object(
                        'id', ts.id,
                        'table_number', ts.table_number,
                        'open_time', ts.open_time,
                        'close_time', ts.close_time,
                        'status', ts.status,
                        'created_at', ts.created_at,
                        'updated_at', ts.updated_at
                    ),
                    'order_items', (
                        SELECT json_agg(
                            json_build_object(
                                'id', oi.id,
                                'price', oi.price,
                                'quantity', oi.quantity,
                                'order_id', oi.order_id,
                                'menu_id', oi.menu_item_id,
                                'status', oi.status,
                                'menu_name', mi.name,
                                'menu_image', mi.image_url,
                                'created_at', oi.created_at
                            )
                            ORDER BY oi.created_at DESC
                        )
                        FROM order_items oi
                        LEFT JOIN menu_items mi ON mi.id = oi.menu_item_id
                        WHERE oi.order_id = o.id AND oi.status != 'not-approved'
                    )
                )
                ORDER BY ts.table_number, o.created_at DESC
            ) as result
        FROM session_data ts
        INNER JOIN orders o ON o.table_session_id = ts.id
        INNER JOIN filtered_orders fo ON fo.id = o.id
        WHERE o.status NOT IN ('not-approved','completed')
    `

	var resultJSON []byte
	err := r.pool.QueryRow(ctx, query).Scan(&resultJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return []models.CustomerOrderRequest{}, nil
		}
		return nil, fmt.Errorf("failed to query order requests: %w", err)
	}

	var result []models.CustomerOrderRequest
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return result, nil
}
func (r *orderRepo) NewGetTableSessionByTableAndPhone(ctx context.Context, tableNumber int, customerPhone string) (*models.CustomerOrderRequest, error) {
	// First, find the active session for this table
	sessionQuery := `
        SELECT 
            id,
            table_number,
            open_time,
            close_time,
            status,
            created_at,
            updated_at
        FROM table_session
        WHERE table_number = $1
            AND status = 'occupied'
            AND open_time IS NOT NULL
            AND close_time IS NULL
        ORDER BY created_at DESC
        LIMIT 1
    `

	var session models.TableSession

	err := r.pool.QueryRow(ctx, sessionQuery, tableNumber).Scan(
		&session.ID,
		&session.TableNumber,
		&session.OpenTime,
		&session.CloseTime,
		&session.Status,
		&session.CreatedAt,
		&session.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			// No active session found, return empty data
			return &models.CustomerOrderRequest{
				OrderId:       uuid.Nil,
				Status:        "",
				Table:         models.TableSession{},
				CustomerName:  nil,
				CustomerPhone: nil,
				Note:          nil,
				OrderItems:    []models.OrderItemType{},
			}, nil
		}
		return nil, fmt.Errorf("failed to query session: %w", err)
	}

	// Search for orders with the given phone number and this table session
	ordersQuery := `
        WITH session_orders AS (
            SELECT 
                o.id,
                o.customer_name,
                o.customer_phone,
                o.note,
                o.status as order_status,
                o.created_at as order_created_at,
                o.waiter_id
            FROM orders o
            WHERE o.table_session_id = $1
                AND o.customer_phone = $2
            ORDER BY o.created_at DESC
        )
        SELECT 
            so.id as order_id,
            so.customer_name,
            so.customer_phone,
            so.note,
            so.order_status,
            so.order_created_at,
            COALESCE(
                (
                    SELECT json_agg(
                        json_build_object(
                            'id', oi.id,
                            'price', oi.price,
                            'quantity', oi.quantity,
                            'order_id', oi.order_id,
                            'menu_id', oi.menu_item_id,
                            'menu_image', mi.image_url,
                            'menu_name', mi.name,
                            'status', oi.status,
                            'created_at', oi.created_at
                        )
                        ORDER BY oi.created_at
                    )
                    FROM order_items oi
                    LEFT JOIN menu_items mi ON mi.id = oi.menu_item_id
                    WHERE oi.order_id = so.id
                ),
                '[]'::json
            ) as items,
            COALESCE(
                (
                    SELECT SUM(oi.quantity * oi.price)
                    FROM order_items oi
                    WHERE oi.order_id = so.id
                ),
                0
            ) as order_total
        FROM session_orders so
        ORDER BY so.order_created_at DESC
    `

	orderRows, err := r.pool.Query(ctx, ordersQuery, session.ID, customerPhone)
	if err != nil {
		return nil, fmt.Errorf("failed to query orders: %w", err)
	}
	defer orderRows.Close()

	// Initialize variables to track all orders
	var allOrderItems []models.OrderItemType
	var customerName *string
	var note *string
	var firstOrderID uuid.UUID
	var orderStatus models.OrderStatus
	var orderCount int

	for orderRows.Next() {
		var (
			orderID          uuid.UUID
			ordCustomerName  *string
			ordCustomerPhone *string
			ordNote          *string
			ordStatus        string
			orderCreated     time.Time
			itemsJSON        []byte
			orderTotal       float64
		)

		err := orderRows.Scan(
			&orderID,
			&ordCustomerName,
			&ordCustomerPhone,
			&ordNote,
			&ordStatus,
			&orderCreated,
			&itemsJSON,
			&orderTotal,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}

		// Parse order items from JSON
		var orderItems []models.OrderItemType
		if err := json.Unmarshal(itemsJSON, &orderItems); err != nil {
			return nil, fmt.Errorf("failed to unmarshal items: %w", err)
		}

		// Set the first order ID and status
		if orderCount == 0 {
			firstOrderID = orderID
			orderStatus = models.OrderStatus(ordStatus)
		}

		// Use the first non-null customer info
		if customerName == nil && ordCustomerName != nil {
			customerName = ordCustomerName
		}
		if note == nil && ordNote != nil {
			note = ordNote
		}

		// Append items to the combined list
		allOrderItems = append(allOrderItems, orderItems...)
		orderCount++
	}

	if err = orderRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating order rows: %w", err)
	}

	// If no orders found with the given phone number, return session with empty order items
	if orderCount == 0 {
		return &models.CustomerOrderRequest{
			OrderId:       uuid.Nil,
			Status:        "",
			Table:         session,
			CustomerName:  nil,
			CustomerPhone: &customerPhone,
			Note:          nil,
			OrderItems:    []models.OrderItemType{},
		}, nil
	}

	// Return complete data with orders found
	return &models.CustomerOrderRequest{
		OrderId:       firstOrderID,
		Status:        orderStatus,
		Table:         session,
		CustomerName:  customerName,
		CustomerPhone: &customerPhone,
		Note:          note,
		OrderItems:    allOrderItems,
	}, nil
}

func (r *orderRepo) GetTableSessionByTableAndPhone(ctx context.Context, tableNumber int, customerPhone string) (*models.CustomerOrderRequest, error) {
	// First, find the active session for this table with matching customer phone
	query := `
        WITH active_session AS (
            -- Find the active table session
            SELECT 
                ts.id,
                ts.table_number,
                ts.open_time,
                ts.close_time,
                ts.status,
                ts.created_at,
                ts.updated_at
            FROM table_session ts
            WHERE ts.table_number = $1
                AND ts.status = 'occupied'
				AND ts.open_time IS NOT NULL
                AND ts.close_time IS NULL
            ORDER BY ts.created_at DESC
            LIMIT 1
        ),
        session_orders AS (
            -- Get all orders for this session with matching phone
            -- and status other than 'not-approved'
            SELECT 
                o.id as order_id,
                o.customer_name,
                o.customer_phone,
                o.note,
                o.status as order_status,
                o.created_at as order_created_at,
                o.waiter_id,
                COALESCE(
                    (
                        SELECT json_agg(
                            json_build_object(
                                'id', oi.id,
                                'price', oi.price,
                                'quantity', oi.quantity,
                                'order_id', oi.order_id,
                                'menu_id', oi.menu_item_id,
                                'menu_image', mi.image_url,
                                'menu_name', mi.name,
                                'created_at', oi.created_at
                            )
                            ORDER BY oi.created_at
                        )
                        FROM order_items oi
                        LEFT JOIN menu_items mi ON mi.id = oi.menu_item_id
                        WHERE oi.order_id = o.id
                    ),
                    '[]'::json
                ) as items,
                -- Calculate order total
                COALESCE(
                    (
                        SELECT SUM(oi.quantity * oi.price)
                        FROM order_items oi
                        WHERE oi.order_id = o.id
                    ),
                    0
                ) as order_total
            FROM orders o
            INNER JOIN active_session as_ ON as_.id = o.table_session_id
            WHERE o.customer_phone = $2
                AND o.status != 'not-approved'  
            ORDER BY o.created_at DESC
        )
        SELECT 
            -- Session Info
            as_.id,
            as_.table_number,
            as_.open_time,
            as_.close_time,
            as_.status,
            as_.created_at,
            as_.updated_at,
            -- Aggregate all orders
            COALESCE(
                (
                    SELECT json_agg(
                        json_build_object(
                            'order_id', so.order_id,
                            'customer_name', so.customer_name,
                            'customer_phone', so.customer_phone,
                            'note', so.note,
                            'order_status', so.order_status,
                            'order_created_at', so.order_created_at,
                            'order_total', so.order_total,
                            'items', so.items::json
                        )
                    )
                    FROM session_orders so
                ),
                '[]'::json
            ) as orders,
            -- Get distinct customer info (assuming all orders have same customer)
            (
                SELECT so.customer_name
                FROM session_orders so
                LIMIT 1
            ) as primary_customer_name,
            (
                SELECT so.note
                FROM session_orders so
                LIMIT 1
            ) as primary_note
        FROM active_session as_
    `

	var (
		sessionID           uuid.UUID
		tableNum            int
		openTime            time.Time
		closeTime           *time.Time
		status              string
		createdAt           time.Time
		updatedAt           time.Time
		ordersJSON          []byte
		primaryCustomerName *string
		primaryNote         *string
	)

	err := r.pool.QueryRow(ctx, query, tableNumber, customerPhone).Scan(
		&sessionID,
		&tableNum,
		&openTime,
		&closeTime,
		&status,
		&createdAt,
		&updatedAt,
		&ordersJSON,
		&primaryCustomerName,
		&primaryNote,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("no active session found for table %d with phone %s", tableNumber, customerPhone)
		}
		return nil, fmt.Errorf("failed to query session: %w", err)
	}

	// Parse orders JSON
	var orders []struct {
		OrderID       uuid.UUID              `json:"order_id"`
		CustomerName  *string                `json:"customer_name"`
		CustomerPhone *string                `json:"customer_phone"`
		Note          *string                `json:"note"`
		OrderStatus   string                 `json:"order_status"`
		OrderCreated  time.Time              `json:"order_created_at"`
		OrderTotal    float64                `json:"order_total"`
		Items         []models.OrderItemType `json:"items"`
	}

	if err := json.Unmarshal(ordersJSON, &orders); err != nil {
		return nil, fmt.Errorf("failed to unmarshal orders: %w", err)
	}

	// If no orders found with status other than 'not-approved', return appropriate error
	if len(orders) == 0 {
		return nil, fmt.Errorf("no approved/pending orders found for table %d with phone %s", tableNumber, customerPhone)
	}

	// Combine all order items from all orders
	var allOrderItems []models.OrderItemType
	for _, order := range orders {
		allOrderItems = append(allOrderItems, order.Items...)
	}

	session := models.TableSession{
		ID:          sessionID,
		TableNumber: tableNum,
		OpenTime:    openTime,
		CloseTime:   closeTime,
		Status:      models.TableState(status),
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}

	return &models.CustomerOrderRequest{
		Table:         session,
		CustomerName:  primaryCustomerName,
		CustomerPhone: &customerPhone,
		Note:          primaryNote,
		OrderItems:    allOrderItems,
	}, nil
}

func (r *orderRepo) NewGetTableSessionByID(ctx context.Context, tableSessionID uuid.UUID) (*models.CustomerOrderRequest, error) {
	// First, get the specific table session
	sessionQuery := `
        SELECT 
            id,
            table_number,
            open_time,
            close_time,
            status,
            created_at,
            updated_at
        FROM table_session
        WHERE id = $1
            AND status = 'occupied' 
            AND close_time IS NULL
    `

	var session models.TableSession
	err := r.pool.QueryRow(ctx, sessionQuery, tableSessionID).Scan(
		&session.ID,
		&session.TableNumber,
		&session.OpenTime,
		&session.CloseTime,
		&session.Status,
		&session.CreatedAt,
		&session.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("table session not found or is not active: %w", err)
		}
		return nil, fmt.Errorf("failed to query session: %w", err)
	}

	// Get all orders for this session with their items
	ordersQuery := `
        WITH session_orders AS (
            SELECT 
                o.id,
                o.customer_name,
                o.customer_phone,
                o.note,
                o.status as order_status,
                o.created_at as order_created_at,
                o.waiter_id
            FROM orders o
            WHERE o.table_session_id = $1
            ORDER BY o.created_at DESC
        )
        SELECT 
            so.id as order_id,
            so.customer_name,
            so.customer_phone,
            so.note,
            so.order_status,
            so.order_created_at,
            COALESCE(
                (
                    SELECT json_agg(
                        json_build_object(
                            'id', oi.id,
                            'price', oi.price,
                            'quantity', oi.quantity,
                            'order_id', oi.order_id,
                            'menu_id', oi.menu_item_id,
                            'menu_image', mi.image_url,
                            'menu_name', mi.name,
                            'status', oi.status,
                            'created_at', oi.created_at
                        )
                        ORDER BY oi.created_at
                    )
                    FROM order_items oi
                    LEFT JOIN menu_items mi ON mi.id = oi.menu_item_id
                    WHERE oi.order_id = so.id
					AND oi.status = 'not-approved'
                ),
                '[]'::json
            ) as items,
            -- Calculate order total
            COALESCE(
                (
                    SELECT SUM(oi.quantity * oi.price)
                    FROM order_items oi
                    WHERE oi.order_id = so.id
                ),
                0
            ) as order_total
        FROM session_orders so
        ORDER BY so.order_created_at DESC
    `

	orderRows, err := r.pool.Query(ctx, ordersQuery, tableSessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to query orders: %w", err)
	}
	defer orderRows.Close()

	// Initialize variables to track all orders
	var allOrderItems []models.OrderItemType
	var customerName *string
	var customerPhone *string
	var note *string
	var orderCount int
	var firstOrderID uuid.UUID
	var orderStatus models.OrderStatus

	for orderRows.Next() {
		var (
			orderID          uuid.UUID
			ordCustomerName  *string
			ordCustomerPhone *string
			ordNote          *string
			ordStatus        string
			orderCreated     time.Time
			itemsJSON        []byte
			orderTotal       float64
		)

		err := orderRows.Scan(
			&orderID,
			&ordCustomerName,
			&ordCustomerPhone,
			&ordNote,
			&ordStatus,
			&orderCreated,
			&itemsJSON,
			&orderTotal,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}

		// Parse order items from JSON
		var orderItems []models.OrderItemType
		if err := json.Unmarshal(itemsJSON, &orderItems); err != nil {
			return nil, fmt.Errorf("failed to unmarshal items: %w", err)
		}

		// Set the first order ID and status
		if orderCount == 0 {
			firstOrderID = orderID
			orderStatus = models.OrderStatus(ordStatus)
		}

		// Use the first non-null customer info (assuming all orders in session have same customer)
		if customerName == nil && ordCustomerName != nil {
			customerName = ordCustomerName
		}
		if customerPhone == nil && ordCustomerPhone != nil {
			customerPhone = ordCustomerPhone
		}
		if note == nil && ordNote != nil {
			note = ordNote
		}

		// Append items to the combined list
		allOrderItems = append(allOrderItems, orderItems...)
		orderCount++
	}

	if err = orderRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating order rows: %w", err)
	}

	// If no orders found, return session with empty order items
	if len(allOrderItems) == 0 {
		return &models.CustomerOrderRequest{
			OrderId:       uuid.Nil, // or some default value
			Status:        "",       // or some default status
			Table:         session,
			CustomerName:  customerName,
			CustomerPhone: customerPhone,
			Note:          note,
			OrderItems:    []models.OrderItemType{},
		}, nil
	}

	return &models.CustomerOrderRequest{
		OrderId:       firstOrderID,
		Status:        orderStatus,
		Table:         session,
		CustomerName:  customerName,
		CustomerPhone: customerPhone,
		Note:          note,
		OrderItems:    allOrderItems,
	}, nil
}

func (r *orderRepo) GetTableSessionByID(ctx context.Context, tableSessionID uuid.UUID) (*models.CustomerOrderRequest, error) {
	// First, get the specific table session
	sessionQuery := `
        SELECT 
            id,
            table_number,
            open_time,
            close_time,
            status,
            created_at,
            updated_at
        FROM table_session
        WHERE id = $1
            AND status = 'occupied' 
            AND close_time IS NULL
    `

	var session models.TableSession
	err := r.pool.QueryRow(ctx, sessionQuery, tableSessionID).Scan(
		&session.ID,
		&session.TableNumber,
		&session.OpenTime,
		&session.CloseTime,
		&session.Status,
		&session.CreatedAt,
		&session.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("table session not found or is not active: %w", err)
		}
		return nil, fmt.Errorf("failed to query session: %w", err)
	}

	// Get all orders for this session with their items
	ordersQuery := `
        WITH session_orders AS (
            SELECT 
                o.id,
                o.customer_name,
                o.customer_phone,
                o.note,
                o.status as order_status,
                o.created_at as order_created_at,
                o.waiter_id
            FROM orders o
            WHERE o.table_session_id = $1
			 AND (o.status = 'not-approved' OR o.status = 'approved')
            ORDER BY o.created_at DESC
        )
        SELECT 
            so.id as order_id,
            so.customer_name,
            so.customer_phone,
            so.note,
            so.order_status,
            so.order_created_at,
            COALESCE(
                (
                    SELECT json_agg(
                        json_build_object(
                            'id', oi.id,
                            'price', oi.price,
                            'quantity', oi.quantity,
                            'order_id', oi.order_id,
                            'menu_id', oi.menu_item_id,
                            'menu_image', mi.image_url,
                            'menu_name', mi.name,
                            'created_at', oi.created_at
                        )
                        ORDER BY oi.created_at
                    )
                    FROM order_items oi
                    LEFT JOIN menu_items mi ON mi.id = oi.menu_item_id
                    WHERE oi.order_id = so.id
                ),
                '[]'::json
            ) as items,
            -- Calculate order total
            COALESCE(
                (
                    SELECT SUM(oi.quantity * oi.price)
                    FROM order_items oi
                    WHERE oi.order_id = so.id
                ),
                0
            ) as order_total
        FROM session_orders so
        ORDER BY so.order_created_at DESC
    `

	orderRows, err := r.pool.Query(ctx, ordersQuery, tableSessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to query orders: %w", err)
	}
	defer orderRows.Close()

	// Since we want all orders for this session in a single CustomerOrderRequest,
	// we'll combine them
	var allOrderItems []models.OrderItemType
	var customerName *string
	var customerPhone *string
	var note *string
	var orderCount int

	for orderRows.Next() {
		var (
			orderID          uuid.UUID
			ordCustomerName  *string
			ordCustomerPhone *string
			ordNote          *string
			orderStatus      string
			orderCreated     time.Time
			itemsJSON        []byte
			orderTotal       float64
		)

		err := orderRows.Scan(
			&orderID,
			&ordCustomerName,
			&ordCustomerPhone,
			&ordNote,
			&orderStatus,
			&orderCreated,
			&itemsJSON,
			&orderTotal,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}

		// Parse order items from JSON
		var orderItems []models.OrderItemType
		if err := json.Unmarshal(itemsJSON, &orderItems); err != nil {
			return nil, fmt.Errorf("failed to unmarshal items: %w", err)
		}

		// Use the first non-null customer info (assuming all orders in session have same customer)
		if customerName == nil && ordCustomerName != nil {
			customerName = ordCustomerName
		}
		if customerPhone == nil && ordCustomerPhone != nil {
			customerPhone = ordCustomerPhone
		}
		if note == nil && ordNote != nil {
			note = ordNote
		}

		// Append items to the combined list
		allOrderItems = append(allOrderItems, orderItems...)
		orderCount++
	}

	if err = orderRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating order rows: %w", err)
	}

	// If no orders found, return session with empty order items
	if len(allOrderItems) == 0 {
		return &models.CustomerOrderRequest{
			Table:         session,
			CustomerName:  customerName,
			CustomerPhone: customerPhone,
			Note:          note,
			OrderItems:    []models.OrderItemType{},
		}, nil
	}

	return &models.CustomerOrderRequest{
		Table:         session,
		CustomerName:  customerName,
		CustomerPhone: customerPhone,
		Note:          note,
		OrderItems:    allOrderItems,
	}, nil
}

// TODO: create a backgorund clearner of order or table sesison where the order are not approved since 10 minutes
func (r *orderRepo) GetAllOrderRequest(ctx context.Context) ([]models.CustomerOrderRequest, error) {
	// Get all occupied sessions that have at least one not-approved order with items
	query := `
        WITH sessions_with_items AS (
            SELECT DISTINCT 
                ts.id,
                ts.table_number,
                ts.open_time,
                ts.close_time,
                ts.status,
                ts.created_at,
                ts.updated_at
            FROM table_session ts
            INNER JOIN orders o ON o.table_session_id = ts.id
            INNER JOIN order_items oi ON oi.order_id = o.id
            WHERE ts.status = 'occupied' 
                AND ts.open_time IS NOT NULL
                AND ts.close_time IS NULL
                AND (o.status = 'not-approved' OR o.status = 'approved')
        )
        SELECT 
            s.id,
            s.table_number,
            s.open_time,
            s.close_time,
            s.status,
            s.created_at,
            s.updated_at,
            COALESCE(
                (
                    SELECT json_agg(
                        json_build_object(
                            'id', o.id,
                            'customer_name', o.customer_name,
                            'customer_phone', o.customer_phone,
                            'note', o.note,
                            'status', o.status,
                            'created_at', o.created_at,
                            'items', (
                                SELECT json_agg(
                                    json_build_object(
                                        'id', oi.id,
                                        'price', oi.price,
                                        'quantity', oi.quantity,
                                        'order_id', oi.order_id,
                                        'menu_id', oi.menu_item_id,
                                        'menu_image', mi.image_url,
                                        'menu_name', mi.name,
                                        'created_at', oi.created_at
                                    )
                                )
                                FROM order_items oi
                                LEFT JOIN menu_items mi ON mi.id = oi.menu_item_id
                                WHERE oi.order_id = o.id
                            )
                        )
                        ORDER BY o.created_at DESC
                    )
                    FROM orders o
                    WHERE o.table_session_id = s.id
                        AND o.status = 'not-approved'
                        AND EXISTS(
                            SELECT 1 FROM order_items oi WHERE oi.order_id = o.id
                        )
                ),
                '[]'::json
            ) as orders
        FROM sessions_with_items s
        ORDER BY s.table_number
    `

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query sessions with orders: %w", err)
	}
	defer rows.Close()

	var result []models.CustomerOrderRequest

	for rows.Next() {
		var session models.TableSession
		var ordersJSON []byte

		err := rows.Scan(
			&session.ID,
			&session.TableNumber,
			&session.OpenTime,
			&session.CloseTime,
			&session.Status,
			&session.CreatedAt,
			&session.UpdatedAt,
			&ordersJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}

		// Parse orders JSON
		var orders []struct {
			ID            uuid.UUID              `json:"id"`
			CustomerName  *string                `json:"customer_name"`
			CustomerPhone *string                `json:"customer_phone"`
			Note          *string                `json:"note"`
			Status        string                 `json:"status"`
			CreatedAt     time.Time              `json:"created_at"`
			Items         []models.OrderItemType `json:"items"`
		}

		if err := json.Unmarshal(ordersJSON, &orders); err != nil {
			return nil, fmt.Errorf("failed to unmarshal orders: %w", err)
		}

		// Combine all items and customer info from all orders
		var allItems []models.OrderItemType
		var customerName, customerPhone, note *string

		for _, order := range orders {
			// Take first non-null customer info
			if customerName == nil && order.CustomerName != nil {
				customerName = order.CustomerName
			}
			if customerPhone == nil && order.CustomerPhone != nil {
				customerPhone = order.CustomerPhone
			}
			if note == nil && order.Note != nil {
				note = order.Note
			}

			allItems = append(allItems, order.Items...)
		}

		result = append(result, models.CustomerOrderRequest{
			Table:         session,
			CustomerName:  customerName,
			CustomerPhone: customerPhone,
			Note:          note,
			OrderItems:    allItems,
		})
	}

	return result, nil
}

func (r *orderRepo) NewGetAllOrderRequest(ctx context.Context) ([]models.CustomerOrderRequest, error) {
	query := `
        WITH filtered_orders AS (
            -- First, get all orders that have at least one not-approved item
            SELECT DISTINCT o.id
            FROM orders o
            INNER JOIN order_items oi ON oi.order_id = o.id
            WHERE oi.status = 'not-approved'
        ),
        session_data AS (
            SELECT 
                ts.id,
                ts.table_number,
                ts.open_time,
                ts.close_time,
                ts.status,
                ts.created_at,
                ts.updated_at
            FROM table_session ts
            WHERE ts.status = 'occupied' 
                AND ts.open_time IS NOT NULL
                AND ts.close_time IS NULL
        )
        SELECT 
            json_agg(
                json_build_object(
                    'id', o.id,
                    'status', o.status,
                    'customer_name', o.customer_name,
                    'customer_phone', o.customer_phone,
                    'note', o.note,
                    'table_session', json_build_object(
                        'id', ts.id,
                        'table_number', ts.table_number,
                        'open_time', ts.open_time,
                        'close_time', ts.close_time,
                        'status', ts.status,
                        'created_at', ts.created_at,
                        'updated_at', ts.updated_at
                    ),
                    'order_items', (
                        SELECT json_agg(
                            json_build_object(
                                'id', oi.id,
                                'price', oi.price,
                                'quantity', oi.quantity,
                                'order_id', oi.order_id,
                                'menu_id', oi.menu_item_id,
                                'status', oi.status,
                                'menu_name', mi.name,
                                'menu_image', mi.image_url,
                                'created_at', oi.created_at
                            )
                            ORDER BY oi.created_at DESC
                        )
                        FROM order_items oi
                        LEFT JOIN menu_items mi ON mi.id = oi.menu_item_id
                        WHERE oi.order_id = o.id AND oi.status = 'not-approved'
                    )
                )
                ORDER BY ts.table_number, o.created_at DESC
            ) as result
        FROM session_data ts
        INNER JOIN orders o ON o.table_session_id = ts.id
        INNER JOIN filtered_orders fo ON fo.id = o.id
        WHERE o.status IN ('not-approved', 'approved')
    `

	var resultJSON []byte
	err := r.pool.QueryRow(ctx, query).Scan(&resultJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return []models.CustomerOrderRequest{}, nil
		}
		return nil, fmt.Errorf("failed to query order requests: %w", err)
	}

	var result []models.CustomerOrderRequest
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return result, nil
}

func (r *orderRepo) NewApproveCustomerRequest(ctx context.Context, approveOrder *models.ApproveOrderType) (err error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	// Optimize 1: Single query to get table session with lock to prevent race conditions
	var tableSession models.TableSession
	query := `
        SELECT id, table_number, open_time, close_time, status, created_at, updated_at
        FROM table_session
        WHERE id = $1
        FOR UPDATE  -- Lock the row to prevent concurrent modifications
    `

	err = tx.QueryRow(ctx, query, approveOrder.TableSessionID).Scan(
		&tableSession.ID,
		&tableSession.TableNumber,
		&tableSession.OpenTime,
		&tableSession.CloseTime,
		&tableSession.Status,
		&tableSession.CreatedAt,
		&tableSession.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("no such table session found")
		}
		return fmt.Errorf("failed to fetch table session: %w", err)
	}

	// Optimize 2: Batch table status updates if needed
	if tableSession.TableNumber != approveOrder.TableNumber {
		// Use a single batch for multiple operations
		batch := &pgx.Batch{}

		// Update session table number
		batch.Queue(`
            UPDATE table_session 
            SET table_number = $1, updated_at = NOW() 
            WHERE id = $2
        `, approveOrder.TableNumber, approveOrder.TableSessionID)

		// Update old table status
		batch.Queue(`
            UPDATE table_status SET status = 'empty' WHERE table_number = $1
        `, tableSession.TableNumber)

		// Update new table status
		batch.Queue(`
            UPDATE table_status SET status = 'occupied' WHERE table_number = $1
        `, approveOrder.TableNumber)

		// Execute batch
		br := tx.SendBatch(ctx, batch)
		defer br.Close()

		// Check results
		for i := 0; i < 3; i++ {
			_, err = br.Exec()
			if err != nil {
				return fmt.Errorf("failed batch operation %d: %w", i, err)
			}
		}
	}

	// Optimize 3: Update order with RETURNING to verify existence
	var updatedID uuid.UUID
	updateOrder := `
        UPDATE orders SET
            customer_name = COALESCE($1, customer_name),
            customer_phone = COALESCE($2, customer_phone),
            waiter_id = $3,
            note = COALESCE($4, note),
            status = $5::order_status_enum
        WHERE id = $6
        RETURNING id
    `

	err = tx.QueryRow(ctx, updateOrder,
		approveOrder.CustomerName,
		approveOrder.CustomerPhone,
		approveOrder.WaiterId,
		approveOrder.Note,
		models.OrderStatusApproved,
		approveOrder.ID,
	).Scan(&updatedID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("order not found")
		}
		return fmt.Errorf("failed to update order: %w", err)
	}

	// Optimize 4: Handle order items with complete sync logic
	if len(approveOrder.OrderMenuItems) > 0 {
		// Build bulk update/insert/delete using CTE with HasChanged flag
		itemsJSON, err := json.Marshal(approveOrder.OrderMenuItems)
		if err != nil {
			return fmt.Errorf("failed to marshal items: %w", err)
		}

		bulkQuery := `
            WITH 
            -- Lock existing items to prevent concurrent modifications
            existing_items AS (
                SELECT id, menu_item_id, quantity, price
                FROM order_items
                WHERE order_id = $2
                FOR UPDATE
            ),
            -- Parse incoming items from JSON
            incoming_items AS (
                SELECT 
                    (item->>'id')::uuid as id,
                    (item->>'menu_item_id')::uuid as menu_item_id,
                    (item->>'quantity')::float as quantity,
                    (item->>'price')::float as price,
                    (item->>'has_changed')::boolean as has_changed
                FROM json_array_elements($1::json) as item
            ),
            -- Identify items to DELETE (exist in DB but not in incoming items)
            items_to_delete AS (
                DELETE FROM order_items oi
                USING existing_items ei
                WHERE oi.id = ei.id
                    AND ei.id NOT IN (SELECT id FROM incoming_items WHERE id IS NOT NULL)
                RETURNING oi.id as deleted_id
            ),
            -- Update existing items that have changed (has_changed = true) or always update status to approved
            updated_items AS (
                UPDATE order_items oi
                SET 
                    quantity = ii.quantity,
                    price = ii.price,
                    status = $3::order_status_enum  -- Always set status to approved
                FROM incoming_items ii
                INNER JOIN existing_items ei ON ei.id = ii.id
                WHERE oi.id = ei.id
                    AND (ii.has_changed = true OR oi.status != $3::order_status_enum)  -- Update if changed OR not already approved
                    AND ii.quantity > 0  -- Only update if quantity > 0
                RETURNING oi.id as updated_id
            ),
            -- Delete items that have changed and have quantity <= 0
            deleted_changed_items AS (
                DELETE FROM order_items oi
                USING incoming_items ii
                INNER JOIN existing_items ei ON ei.id = ii.id
                WHERE oi.id = ei.id 
                    AND ii.has_changed = true 
                    AND ii.quantity <= 0
                RETURNING oi.id as deleted_changed_id
            ),
            -- Insert new items (those without IDs)
            inserted_items AS (
                INSERT INTO order_items (order_id, menu_item_id, quantity, price, status, created_at)
                SELECT 
                    $2,
                    ii.menu_item_id,
                    ii.quantity,
                    ii.price,
                    $3::order_status_enum,  -- Set status to approved
                    NOW()
                FROM incoming_items ii
                WHERE ii.quantity > 0
                    AND ii.id IS NULL  -- Only insert items without IDs (new items)
                RETURNING id as inserted_id
            ),
            -- Update existing items that haven't changed but need status update to approved
            status_update_items AS (
                UPDATE order_items oi
                SET status = $3::order_status_enum
                FROM incoming_items ii
                INNER JOIN existing_items ei ON ei.id = ii.id
                WHERE oi.id = ei.id
                    AND ii.has_changed = false
                    AND oi.status != $3::order_status_enum  -- Only update if not already approved
                    AND ii.quantity > 0
                RETURNING oi.id as status_updated_id
            )
            -- Return summary of all operations
            SELECT 
                (SELECT COUNT(*) FROM items_to_delete) as deleted_count,
                (SELECT COUNT(*) FROM updated_items) as updated_count,
                (SELECT COUNT(*) FROM deleted_changed_items) as deleted_changed_count,
                (SELECT COUNT(*) FROM inserted_items) as inserted_count,
                (SELECT COUNT(*) FROM status_update_items) as status_updated_count
        `

		var deletedCount, updatedCount, deletedChangedCount, insertedCount, statusUpdatedCount int
		err = tx.QueryRow(ctx, bulkQuery, itemsJSON, approveOrder.ID, models.OrderStatusApproved).Scan(
			&deletedCount, &updatedCount, &deletedChangedCount, &insertedCount, &statusUpdatedCount,
		)
		if err != nil {
			return fmt.Errorf("failed bulk item operation: %w", err)
		}
	} else {
		// If no items in approveOrder, delete ALL existing items (customer removed everything)
		_, err = tx.Exec(ctx, `DELETE FROM order_items WHERE order_id = $1`, approveOrder.ID)
		if err != nil {
			return fmt.Errorf("failed to delete all items: %w", err)
		}
	}

	// Optimize 5: Update session status
	_, err = tx.Exec(ctx,
		`UPDATE table_session 
         SET status='occupied', updated_at=NOW()
         WHERE id=$1`,
		approveOrder.TableSessionID,
	)

	if err != nil {
		return fmt.Errorf("failed to update session status: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("transaction commit failed: %w", err)
	}

	return nil
}

func (r *orderRepo) ApproveCustomerRequest(ctx context.Context, approveOrder *models.ApproveOrderType) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// 1. Get the current table session
	var tableSession struct {
		ID          uuid.UUID         `db:"id"`
		TableNumber int               `db:"table_number"`
		Status      models.TableState `db:"status"`
	}

	queryGetSession := `SELECT id, table_number, status FROM table_session WHERE id = $1`
	err = tx.QueryRow(ctx, queryGetSession, approveOrder.TableSessionID).Scan(
		&tableSession.ID,
		&tableSession.TableNumber,
		&tableSession.Status,
	)
	if err != nil {
		return fmt.Errorf("failed to get table session: %w", err)
	}

	// 2. Check if table number needs to be updated (if waiter changed the table number)
	if tableSession.TableNumber != approveOrder.TableNumber {
		// Update table_session table with new table number
		queryUpdateSession := `UPDATE table_session SET table_number = $1, updated_at = NOW() WHERE id = $2`
		_, err = tx.Exec(ctx, queryUpdateSession, approveOrder.TableNumber, approveOrder.TableSessionID)
		if err != nil {
			return fmt.Errorf("failed to update session table number: %w", err)
		}

		// Update table_status table for the old table to empty
		queryUpdateOldTableStatus := `UPDATE table_status SET status = 'empty' WHERE table_number = $1`
		_, err = tx.Exec(ctx, queryUpdateOldTableStatus, tableSession.TableNumber)
		if err != nil {
			return fmt.Errorf("failed to update old table status: %w", err)
		}

		// Update table_status table for the new table to occupied
		queryUpdateNewTableStatus := `UPDATE table_status SET status = 'occupied' WHERE table_number = $1`
		_, err = tx.Exec(ctx, queryUpdateNewTableStatus, approveOrder.TableNumber)
		if err != nil {
			return fmt.Errorf("failed to update new table status: %w", err)
		}
	}

	// 3. Get the order ID for this table session (the one with 'not-approved' status)
	var orderID uuid.UUID
	queryGetOrderID := `SELECT id FROM orders WHERE table_session_id = $1 AND status = 'not-approved'`
	err = tx.QueryRow(ctx, queryGetOrderID, approveOrder.TableSessionID).Scan(&orderID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("no pending order found for this table session")
		}
		return fmt.Errorf("failed to get order ID: %w", err)
	}

	// 4. Update the existing order
	queryUpdateOrder := `
        UPDATE orders SET
            customer_name = COALESCE($1, customer_name),
            customer_phone = COALESCE($2, customer_phone),
            waiter_id = $3,
            note = COALESCE($4, note),
            status = $5
        WHERE id = $6 AND status = 'not-approved'
    `
	cmd, err := tx.Exec(ctx, queryUpdateOrder,
		approveOrder.CustomerName,
		approveOrder.CustomerPhone,
		approveOrder.WaiterId,
		approveOrder.Note,
		models.OrderStatusApproved, // Set status to approved
		orderID,
	)
	if err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("no pending approval order found for this table session")
	}

	// 5. Get all existing order items for this order
	existingItemsQuery := `
		SELECT 
			oi.id,
			oi.menu_item_id,
			oi.quantity,
			oi.price
		FROM order_items oi
		WHERE oi.order_id = $1
	`
	existingRows, err := tx.Query(ctx, existingItemsQuery, orderID)
	if err != nil {
		return fmt.Errorf("failed to query existing items: %w", err)
	}
	defer existingRows.Close()

	// Create a map of existing items keyed by menu_item_id
	type existingItem struct {
		ID         uuid.UUID
		MenuItemID uuid.UUID
		Quantity   float64
		Price      float64
	}
	existingItems := make(map[uuid.UUID]existingItem)
	for existingRows.Next() {
		var item existingItem
		if err := existingRows.Scan(
			&item.ID,
			&item.MenuItemID,
			&item.Quantity,
			&item.Price,
		); err != nil {
			return fmt.Errorf("failed to scan existing item: %w", err)
		}
		existingItems[item.MenuItemID] = item
	}

	// 6. Process incoming items from the approval request
	incomingItems := make(map[uuid.UUID]models.ApproveOrderItem)
	for _, item := range approveOrder.OrderMenuItems {
		incomingItems[item.MenuItemID] = item
	}

	// 7. Handle items to delete (exist in DB but not in incoming request)
	for menuItemID, existingItem := range existingItems {
		if _, exists := incomingItems[menuItemID]; !exists {
			// This item exists in DB but not in the approval request - delete it
			deleteQuery := `DELETE FROM order_items WHERE id = $1`
			_, err = tx.Exec(ctx, deleteQuery, existingItem.ID)
			if err != nil {
				return fmt.Errorf("failed to delete removed order item (menu_item_id: %s): %w", menuItemID, err)
			}
			fmt.Printf("Deleted item %s (menu_item_id: %s) as it's not in incoming request\n", existingItem.ID, menuItemID)
		}
	}

	// 8. Process incoming items (update existing or insert new)
	for menuItemID, incomingItem := range incomingItems {
		if existingItem, exists := existingItems[menuItemID]; exists {
			// Item exists in database - check if it needs update
			if existingItem.Quantity != incomingItem.Quantity ||
				existingItem.Price != incomingItem.Price ||
				incomingItem.HasChanged {

				if incomingItem.Quantity > 0 {
					// Update existing item with new quantity/price
					updateQuery := `
						UPDATE order_items SET
							quantity = $1,
							price = $2
						WHERE id = $3
					`
					_, err = tx.Exec(ctx, updateQuery,
						incomingItem.Quantity,
						incomingItem.Price,
						existingItem.ID,
					)
					if err != nil {
						return fmt.Errorf("failed to update order item (menu_item_id: %s): %w", menuItemID, err)
					}
					fmt.Printf("Updated item %s (menu_item_id: %s) quantity: %f, price: %f\n",
						existingItem.ID, menuItemID, incomingItem.Quantity, incomingItem.Price)
				} else {
					// Quantity is 0, delete the item
					deleteQuery := `DELETE FROM order_items WHERE id = $1`
					_, err = tx.Exec(ctx, deleteQuery, existingItem.ID)
					if err != nil {
						return fmt.Errorf("failed to delete zero-quantity order item (menu_item_id: %s): %w", menuItemID, err)
					}
					fmt.Printf("Deleted item %s (menu_item_id: %s) as quantity is 0\n", existingItem.ID, menuItemID)
				}
			} else {
				fmt.Printf("Item %s (menu_item_id: %s) unchanged, skipping\n", existingItem.ID, menuItemID)
			}
		} else {
			// Item doesn't exist in database - insert new one
			if incomingItem.Quantity > 0 {
				newItemID, err := uuid.NewV4()
				if err != nil {
					return fmt.Errorf("failed to generate UUID: %w", err)
				}

				insertQuery := `
					INSERT INTO order_items (
						id, order_id, menu_item_id, quantity, price, created_at
					) VALUES (
						$1, $2, $3, $4, $5, $6
					)
				`
				_, err = tx.Exec(ctx, insertQuery,
					newItemID,
					orderID,
					incomingItem.MenuItemID,
					incomingItem.Quantity,
					incomingItem.Price,
					time.Now(),
				)
				if err != nil {
					return fmt.Errorf("failed to insert new order item (menu_item_id: %s): %w", menuItemID, err)
				}
				fmt.Printf("Inserted new item %s (menu_item_id: %s) quantity: %f, price: %f\n",
					newItemID, menuItemID, incomingItem.Quantity, incomingItem.Price)
			}
		}
	}

	// 9. Update the table session status to occupied (if not already)
	queryUpdateSessionStatus := `UPDATE table_session SET status = 'occupied', updated_at = NOW() WHERE id = $1`
	_, err = tx.Exec(ctx, queryUpdateSessionStatus, approveOrder.TableSessionID)
	if err != nil {
		return fmt.Errorf("failed to update session status: %w", err)
	}

	// Commit the transaction
	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	fmt.Printf("Order %s approved successfully with %d items\n", orderID, len(approveOrder.OrderMenuItems))
	return nil
}

func (r *orderRepo) GetTableStatus(ctx context.Context, tableNumber int) (*models.TableSession, error) {
	// First check if the table exists in table_status
	var tableStatus models.TableState
	var tableCapacity int

	checkTableQuery := `
        SELECT status, capacity 
        FROM table_status 
        WHERE table_number = $1
    `

	err := r.pool.QueryRow(ctx, checkTableQuery, tableNumber).Scan(&tableStatus, &tableCapacity)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("table %d does not exist", tableNumber)
		}
		return nil, fmt.Errorf("failed to check table existence: %w", err)
	}

	// Now check for active session
	query := `
        SELECT id, table_number, open_time, close_time, status
        FROM table_session
        WHERE table_number = $1 AND close_time IS NULL
        ORDER BY open_time DESC
        LIMIT 1
    `

	var session models.TableSession
	err = r.pool.QueryRow(ctx, query, tableNumber).Scan(
		&session.ID,
		&session.TableNumber,
		&session.OpenTime,
		&session.CloseTime,
		&session.Status,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// No active session found, return table status info
			return nil, &TableNotActiveError{
				TableNumber: tableNumber,
				Status:      tableStatus,
				Message:     fmt.Sprintf("no active session for table %d (current status: %s)", tableNumber, tableStatus),
			}
		}
		return nil, fmt.Errorf("failed to get table status: %w", err)
	}

	return &session, nil
}

// Custom error type for table not active
type TableNotActiveError struct {
	TableNumber int
	Status      models.TableState
	Message     string
}

func (e *TableNotActiveError) Error() string {
	return e.Message
}

func (r *orderRepo) CreateCustomerOrder(ctx context.Context, customerOrder *models.CreateCustomerOrderRequest) error {
	// Start a transaction
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Ensure rollback if anything fails
	defer tx.Rollback(ctx)

	var tableSessionID uuid.UUID

	// Check if there's an active table session
	tableSession, err := r.GetTableStatus(ctx, customerOrder.TableNumber)
	if err != nil {
		// Check if it's a TableNotActiveError
		var notActiveErr *TableNotActiveError
		if errors.As(err, &notActiveErr) {
			// Table exists but no active session
			// Check if table is empty before creating new session
			if notActiveErr.Status != models.TableEmpty {
				return fmt.Errorf("cannot create order: table %d is %s (must be empty)",
					customerOrder.TableNumber, notActiveErr.Status)
			}

			// Create new table session
			newSession := &models.TableSession{
				ID:          uuid.Must(uuid.NewV4()),
				TableNumber: customerOrder.TableNumber,
				OpenTime:    time.Now(),
				Status:      models.TableOccupied,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}

			// Insert new table session
			query := `
                INSERT INTO table_session (id, table_number, open_time, status, created_at, updated_at)
                VALUES ($1, $2, $3, $4, $5, $6)
            `

			_, err = tx.Exec(ctx, query,
				newSession.ID,
				newSession.TableNumber,
				newSession.OpenTime,
				newSession.Status,
				newSession.CreatedAt,
				newSession.UpdatedAt,
			)

			if err != nil {
				return fmt.Errorf("failed to create table session: %w", err)
			}

			tableSessionID = newSession.ID

			// Update table_status to occupied
			updateTableStatusQuery := `
                UPDATE table_status 
                SET status = $1 
                WHERE table_number = $2
            `
			_, err = tx.Exec(ctx, updateTableStatusQuery, models.TableOccupied, customerOrder.TableNumber)
			if err != nil {
				return fmt.Errorf("failed to update table status: %w", err)
			}
		} else {
			// Some other error (like table doesn't exist)
			return fmt.Errorf("failed to get table status: %w", err)
		}
	} else {
		// Active session exists
		tableSessionID = tableSession.ID
	}

	// Check if there's an existing order for this session with status 'not-approved'
	var existingOrderID uuid.UUID
	var orderExists bool

	checkOrderQuery := `
        SELECT id FROM orders 
        WHERE table_session_id = $1 
        LIMIT 1
    `

	err = tx.QueryRow(ctx, checkOrderQuery, tableSessionID).Scan(&existingOrderID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			orderExists = false
		} else {
			return fmt.Errorf("failed to check existing order: %w", err)
		}
	} else {
		orderExists = true
	}

	var orderID uuid.UUID

	if orderExists {
		// Use existing order
		orderID = existingOrderID

		// Update order details (only if provided)
		// Only update if the fields are provided (non-nil)
		if customerOrder.CustomerName != nil || customerOrder.CustomerPhone != nil || customerOrder.Note != nil {
			updateOrderQuery := `
				UPDATE orders 
				SET customer_name = COALESCE($1, customer_name),
					customer_phone = COALESCE($2, customer_phone),
					note = COALESCE($3, note)
				WHERE id = $4
			`

			_, err = tx.Exec(ctx, updateOrderQuery,
				customerOrder.CustomerName,
				customerOrder.CustomerPhone,
				customerOrder.Note,
				orderID,
			)

			if err != nil {
				return fmt.Errorf("failed to update existing order: %w", err)
			}
		}

	} else {
		// Create new order
		orderID = uuid.Must(uuid.NewV4())

		// Insert order
		orderQuery := `
            INSERT INTO orders (id, table_session_id, customer_name, customer_phone, note, status, created_at)
            VALUES ($1, $2, $3, $4, $5, $6, $7)
        `

		_, err = tx.Exec(ctx, orderQuery,
			orderID,
			tableSessionID,
			customerOrder.CustomerName,
			customerOrder.CustomerPhone,
			customerOrder.Note,
			models.OrderStatusNotApproved,
			time.Now(),
		)

		if err != nil {
			return fmt.Errorf("failed to create order: %w", err)
		}
	}

	// Validate order has at least one menu item
	if len(customerOrder.OrderMenuItems) == 0 {
		return fmt.Errorf("order must have at least one menu item")
	}

	// Check for duplicate items to avoid inserting the same menu item twice?
	// For now, we'll just insert all items (they could be duplicates if user adds same item multiple times)

	// Prepare batch for order items - APPEND new items without deleting existing ones
	batch := &pgx.Batch{}

	for _, item := range customerOrder.OrderMenuItems {
		orderItemID := uuid.Must(uuid.NewV4())

		itemQuery := `
            INSERT INTO order_items (id, order_id, menu_item_id, quantity, price, created_at)
            VALUES ($1, $2, $3, $4, $5, $6)
        `

		batch.Queue(itemQuery,
			orderItemID,
			orderID,
			item.MenuItemID,
			item.Quantity,
			item.Price,
			time.Now(),
		)
	}

	// Execute batch insert for order items
	br := tx.SendBatch(ctx, batch)

	// Process batch results
	for range customerOrder.OrderMenuItems {
		_, err = br.Exec()
		if err != nil {
			br.Close()
			return fmt.Errorf("failed to insert order item: %w", err)
		}
	}
	br.Close()

	// Commit the transaction
	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *orderRepo) NewCreateCustomerOrder(ctx context.Context, customerOrder *models.CreateCustomerOrderRequest) error {

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	// Fetch existing active table session (must exist — created during approval)
	tableSession := &models.TableSession{}
	query := `
		SELECT id, table_number, open_time, close_time, status, created_at, updated_at
		FROM table_session
		WHERE open_time IS NOT NULL
		AND close_time IS NULL
		AND table_number = $1
	`
	err = tx.QueryRow(ctx, query, customerOrder.TableNumber).Scan(
		&tableSession.ID,
		&tableSession.TableNumber,
		&tableSession.OpenTime,
		&tableSession.CloseTime,
		&tableSession.Status,
		&tableSession.CreatedAt,
		&tableSession.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("no active session found for table %d", customerOrder.TableNumber)
		}
		return fmt.Errorf("failed to fetch table session: %w", err)
	}

	// Check existing order
	var existingOrderID uuid.UUID
	var orderExists bool

	checkOrderQuery := `
		SELECT id FROM orders
		WHERE table_session_id = $1
		LIMIT 1
	`

	err = tx.QueryRow(ctx, checkOrderQuery, tableSession.ID).Scan(&existingOrderID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			orderExists = false
		} else {
			return fmt.Errorf("failed to check existing order: %w", err)
		}
	} else {
		orderExists = true
	}

	var orderID uuid.UUID

	if orderExists {

		orderID = existingOrderID

		if customerOrder.CustomerName != nil ||
			customerOrder.CustomerPhone != nil ||
			customerOrder.Note != nil {

			updateOrderQuery := `
				UPDATE orders
				SET customer_name = COALESCE($1, customer_name),
					customer_phone = COALESCE($2, customer_phone),
					note = COALESCE($3, note)
				WHERE id = $4
			`

			_, err = tx.Exec(ctx, updateOrderQuery,
				customerOrder.CustomerName,
				customerOrder.CustomerPhone,
				customerOrder.Note,
				orderID,
			)

			if err != nil {
				return fmt.Errorf("failed to update existing order: %w", err)
			}
		}

	} else {

		orderQuery := `
		INSERT INTO orders (
			table_session_id,
			customer_name,
			customer_phone,
			note,
			status,
			created_at
		)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING id
	`

		err = tx.QueryRow(ctx, orderQuery,
			tableSession.ID,
			customerOrder.CustomerName,
			customerOrder.CustomerPhone,
			customerOrder.Note,
			models.OrderStatusNotApproved,
			time.Now(),
		).Scan(&orderID)

		if err != nil {
			return fmt.Errorf("failed to create order: %w", err)
		}
	}

	if len(customerOrder.OrderMenuItems) == 0 {
		return fmt.Errorf("order must have at least one menu item")
	}

	for _, item := range customerOrder.OrderMenuItems {

		itemQuery := `
			INSERT INTO order_items (
				order_id,
				menu_item_id,
				status,
				quantity,
				price,
				created_at
			)
			VALUES ($1,$2,$3,$4,$5,$6)
		`

		_, err = tx.Exec(ctx, itemQuery,
			orderID,
			item.MenuItemID,
			models.OrderStatusNotApproved,
			item.Quantity,
			item.Price,
			time.Now(),
		)

		if err != nil {
			return fmt.Errorf("failed to insert order item: %w", err)
		}
	}

	// Commit transaction
	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("transaction commit failed: %w", err)
	}

	return nil
}

// NewOrderRepository creates a new order repository instance
func NewOrderRepository() OrderRepo {
	pool, err := database.GetPostgresPool()
	if err != nil {
		log.Printf("Failed to get postgres pool: %v", err)
		return nil
	}
	return &orderRepo{
		pool: pool,
	}
}
