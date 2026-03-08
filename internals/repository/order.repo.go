package repository

import (
	"context"
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
                AND ts.close_time IS NULL
            ORDER BY ts.created_at DESC
            LIMIT 1
        ),
        session_orders AS (
            -- Get all orders for this session with matching phone
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
                                'menu_name', mi.name
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
                AND o.status NOT IN ('completed', 'cancelled')
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
                AND o.status NOT IN ('completed', 'cancelled')
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
                            'menu_name', mi.name
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

//TODO: create a backgorund clearner of order or table sesison where the order are not approved since 10 minutes

func (r *orderRepo) GetAllOrderRequest(ctx context.Context) ([]models.CustomerOrderRequest, error) {
	// First, get all active sessions
	sessionsQuery := `
        SELECT 
            id,
            table_number,
            open_time,
            close_time,
            status,
            created_at,
            updated_at
        FROM table_session
        WHERE status = 'occupied' 
            AND close_time IS NULL
        ORDER BY table_number
    `

	sessionRows, err := r.pool.Query(ctx, sessionsQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query sessions: %w", err)
	}
	defer sessionRows.Close()

	var result []models.CustomerOrderRequest

	for sessionRows.Next() {
		var session models.TableSession
		err := sessionRows.Scan(
			&session.ID,
			&session.TableNumber,
			&session.OpenTime,
			&session.CloseTime,
			&session.Status,
			&session.CreatedAt,
			&session.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}

		// Get orders for this session
		ordersQuery := `
            SELECT 
                o.id,
                o.customer_name,
                o.customer_phone,
                o.note,
                o.status,
                o.created_at,
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
                                'menu_name', mi.name
                            )
                        )
                        FROM order_items oi
                        LEFT JOIN menu_items mi ON mi.id = oi.menu_item_id
                        WHERE oi.order_id = o.id
                    ),
                    '[]'::json
                ) as items
            FROM orders o
            WHERE o.table_session_id = $1
                AND o.status NOT IN ('completed', 'cancelled')
            ORDER BY o.created_at DESC
        `

		orderRows, err := r.pool.Query(ctx, ordersQuery, session.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to query orders: %w", err)
		}

		for orderRows.Next() {
			var (
				orderID       uuid.UUID
				customerName  *string
				customerPhone *string
				note          *string
				orderStatus   string
				orderCreated  time.Time
				itemsJSON     []byte
			)

			err := orderRows.Scan(
				&orderID,
				&customerName,
				&customerPhone,
				&note,
				&orderStatus,
				&orderCreated,
				&itemsJSON,
			)
			if err != nil {
				orderRows.Close()
				return nil, fmt.Errorf("failed to scan order: %w", err)
			}

			var orderItems []models.OrderItemType
			if err := json.Unmarshal(itemsJSON, &orderItems); err != nil {
				orderRows.Close()
				return nil, fmt.Errorf("failed to unmarshal items: %w", err)
			}

			result = append(result, models.CustomerOrderRequest{
				Table:         session,
				CustomerName:  customerName,
				CustomerPhone: customerPhone,
				Note:          note,
				OrderItems:    orderItems,
			})
		}
		orderRows.Close()
	}

	return result, nil
}

func (r *orderRepo) ApproveCustomerRequest(ctx context.Context, approveOrder *models.ApproveOrderType) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// 1. Get the current table session
	var tableSession struct {
		ID          uuid.UUID `db:"id"`
		TableNumber int       `db:"table_number"`
		Status      string    `db:"status"`
	}

	queryGetSession := `SELECT id, table_number, status FROM table_session WHERE id = $1`
	err = tx.QueryRow(ctx, queryGetSession, approveOrder.TableSessionID).Scan(
		&tableSession.ID,
		&tableSession.TableNumber,
		&tableSession.Status,
	)
	if err != nil {
		return err
	}

	// 2. Check if table number needs to be updated (if waiter changed the table number)
	if tableSession.TableNumber != approveOrder.TableNumber {
		// Update table_session table with new table number
		queryUpdateSession := `UPDATE table_session SET table_number = $1, updated_at = NOW() WHERE id = $2`
		_, err = tx.Exec(ctx, queryUpdateSession, approveOrder.TableNumber, approveOrder.TableSessionID)
		if err != nil {
			return err
		}

		// Update table_status table for the old table to empty
		queryUpdateOldTableStatus := `UPDATE table_status SET status = 'empty' WHERE table_number = $1`
		_, err = tx.Exec(ctx, queryUpdateOldTableStatus, tableSession.TableNumber)
		if err != nil {
			return err
		}

		// Update table_status table for the new table to occupied
		queryUpdateNewTableStatus := `UPDATE table_status SET status = 'occupied' WHERE table_number = $1`
		_, err = tx.Exec(ctx, queryUpdateNewTableStatus, approveOrder.TableNumber)
		if err != nil {
			return err
		}
	}

	// 3. Update the existing order
	queryUpdateOrder := `
        UPDATE orders SET
            customer_name = $1,
            customer_phone = $2,
            waiter_id = $3,
            note = $4,
            status = $5
        WHERE table_session_id = $6 AND status = 'not-approved'
    `
	cmd, err := tx.Exec(ctx, queryUpdateOrder,
		approveOrder.CustomerName,
		approveOrder.CustomerPhone,
		approveOrder.WaiterId,
		approveOrder.Note,
		models.OrderStatusPending, // Set status to pending
		approveOrder.TableSessionID,
	)
	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("no pending approval order found for this table session")
	}

	// 4. Get the order ID for this table session
	var orderID uuid.UUID
	queryGetOrderID := `SELECT id FROM orders WHERE table_session_id = $1 AND status = 'pending'`
	err = tx.QueryRow(ctx, queryGetOrderID, approveOrder.TableSessionID).Scan(&orderID)
	if err != nil {
		return err
	}

	// 5. Update only the order items that have changed
	for _, item := range approveOrder.OrderMenuItems {
		if item.HasChanged {
			// Check if the item already exists
			var existingID uuid.UUID
			queryCheckExisting := `SELECT id FROM order_items WHERE order_id = $1 AND menu_item_id = $2`
			err = tx.QueryRow(ctx, queryCheckExisting, orderID, item.MenuItemID).Scan(&existingID)

			if err == nil {
				// Item exists, update it
				queryUpdateOrderItem := `
					UPDATE order_items SET
						quantity = $1,
						price = $2,
						has_changed = $3,
						updated_at = $4
					WHERE order_id = $5 AND menu_item_id = $6
				`
				_, err = tx.Exec(ctx, queryUpdateOrderItem,
					item.Quantity,
					item.Price,
					item.HasChanged,
					time.Now(),
					orderID,
					item.MenuItemID,
				)
			} else {
				// Item doesn't exist, insert new one
				queryInsertOrderItem := `
					INSERT INTO order_items (
						id, order_id, menu_item_id, quantity, price, has_changed, created_at, updated_at
					) VALUES (
						$1, $2, $3, $4, $5, $6, $7, $8
					)
				`
				_, err = tx.Exec(ctx, queryInsertOrderItem,
					uuid.Must(uuid.NewV4()), // Generate new ID using gofrs/uuid
					orderID,
					item.MenuItemID,
					item.Quantity,
					item.Price,
					item.HasChanged,
					time.Now(),
					time.Now(),
				)
			}

			if err != nil {
				return err
			}
		}
		// If HasChanged is false, do nothing for this item
	}

	// 6. Update the table session status to occupied
	queryUpdateSessionStatus := `UPDATE table_session SET status = 'occupied', updated_at = NOW() WHERE id = $1`
	_, err = tx.Exec(ctx, queryUpdateSessionStatus, approveOrder.TableSessionID)
	if err != nil {
		return err
	}

	// Commit the transaction
	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

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
				CloseTime:   nil,
				Status:      models.TableOccupied,
			}

			// Insert new table session
			query := `
                INSERT INTO table_session (id, table_number, open_time, status)
                VALUES ($1, $2, $3, $4)
            `

			_, err = tx.Exec(ctx, query,
				newSession.ID,
				newSession.TableNumber,
				newSession.OpenTime,
				newSession.Status,
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

	// Create the order
	orderID := uuid.Must(uuid.NewV4())
	order := &models.Order{
		ID:             orderID,
		TableSessionID: tableSessionID,
		CustomerName:   customerOrder.CustomerName,
		CustomerPhone:  customerOrder.CustomerPhone,
		Note:           customerOrder.Note,
		Status:         models.OrderStatusNotApproved,
		CreatedAt:      time.Now(),
	}

	// Insert order
	orderQuery := `
        INSERT INTO orders (id, table_session_id, customer_name, customer_phone, note, status, created_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `

	_, err = tx.Exec(ctx, orderQuery,
		order.ID,
		order.TableSessionID,
		order.CustomerName,
		order.CustomerPhone,
		order.Note,
		order.Status,
		order.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}

	// Create order items
	if len(customerOrder.OrderMenuItems) == 0 {
		return fmt.Errorf("order must have at least one menu item")
	}

	// Prepare batch for order items
	batch := &pgx.Batch{}

	for _, item := range customerOrder.OrderMenuItems {
		orderItem := &models.OrderItem{
			ID:         uuid.Must(uuid.NewV4()),
			OrderID:    orderID,
			MenuItemID: item.MenuItemID,
			Quantity:   item.Quantity,
			Price:      item.Price,
			CreatedAt:  time.Now(),
		}

		itemQuery := `
            INSERT INTO order_items (id, order_id, menu_item_id, quantity, price, created_at)
            VALUES ($1, $2, $3, $4, $5, $6)
        `

		batch.Queue(itemQuery,
			orderItem.ID,
			orderItem.OrderID,
			orderItem.MenuItemID,
			orderItem.Quantity,
			orderItem.Price,
			orderItem.CreatedAt,
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
