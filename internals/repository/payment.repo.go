package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/Abhishekh669/backend/internals/database"
	"github.com/Abhishekh669/backend/internals/models"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PaymentRepo interface {
	GetAllOrderDetailsForCashierByOrderId(ctx context.Context, orderId uuid.UUID) (*models.PaymentDetailsForCashierWithDiscount, error)
	GetAllApprovedOrdersForCashier(ctx context.Context) ([]models.GetOrderDetailsForCashier, error)
	CreatePayment(ctx context.Context, paymentData *models.CreatePayment) (*models.Payment, error)
}
type paymentRepo struct {
	pool *pgxpool.Pool
}

func (r *paymentRepo) GetAllOrderDetailsForCashierByOrderId(ctx context.Context, orderId uuid.UUID) (*models.PaymentDetailsForCashierWithDiscount, error) {
	query := `
        SELECT
            o.id              AS order_id,
            o.status,
            o.customer_name,
            o.customer_phone,
            o.waiter_id,
            ts.table_number,

            -- waiter info from users table
            u.name            AS waiter_name,
            u.image           AS waiter_image,

            -- order items
            oi.id             AS item_id,
            oi.price          AS item_price,
            oi.quantity       AS item_quantity,
            oi.status         AS item_status,
            oi.created_at     AS item_created_at,
            oi.menu_item_id,

            -- menu info
            mi.name           AS menu_name,
            mi.image_url      AS menu_image,

            -- token info (nullable)
            ut.id             AS token_id,
            ut.phone_number   AS token_phone,
            ut.total_tokens,
            ut.created_at     AS token_created_at,
            ut.updated_at     AS token_updated_at

        FROM orders o
        JOIN table_session ts  ON ts.id = o.table_session_id
        LEFT JOIN users u      ON u.id  = o.waiter_id
        JOIN order_items oi    ON oi.order_id = o.id
        JOIN menu_items mi     ON mi.id = oi.menu_item_id
        LEFT JOIN user_tokens ut ON ut.phone_number = o.customer_phone

        WHERE o.id = $1
    `

	rows, err := r.pool.Query(ctx, query, orderId)
	if err != nil {
		return nil, fmt.Errorf("GetAllOrderDetailsForCashierByOrderId query: %w", err)
	}
	defer rows.Close()

	var result *models.PaymentDetailsForCashierWithDiscount

	for rows.Next() {
		var (
			// order-level (same on every row)
			oID           uuid.UUID
			oStatus       models.OrderStatus
			customerName  *string
			customerPhone *string
			tableNumber   int

			// waiter (all nullable due to LEFT JOIN)
			waiterID    *uuid.UUID
			waiterName  *string
			waiterImage *string

			// order item
			itemID        uuid.UUID
			itemPrice     float64
			itemQty       float64
			itemStatus    models.OrderStatus
			itemCreatedAt time.Time
			menuItemID    uuid.UUID
			menuName      string
			menuImage     *string

			// token (all nullable)
			tokenID        *uuid.UUID
			tokenPhone     *string
			totalTokens    *float64
			tokenCreatedAt *time.Time
			tokenUpdatedAt *time.Time
		)

		if err := rows.Scan(
			&oID, &oStatus, &customerName, &customerPhone, &waiterID, &tableNumber,
			&waiterName, &waiterImage,
			&itemID, &itemPrice, &itemQty, &itemStatus, &itemCreatedAt, &menuItemID,
			&menuName, &menuImage,
			&tokenID, &tokenPhone, &totalTokens, &tokenCreatedAt, &tokenUpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("GetAllOrderDetailsForCashierByOrderId scan: %w", err)
		}

		// resolve waiter fields with dummy fallbacks
		resolvedWaiterID := uuid.Nil
		if waiterID != nil {
			resolvedWaiterID = *waiterID
		}

		resolvedWaiterName := "N/A"
		if waiterName != nil {
			resolvedWaiterName = *waiterName
		}

		var resolvedWaiterImage *string
		if waiterImage != nil {
			resolvedWaiterImage = waiterImage
		}

		// initialise the result struct once using the first row
		if result == nil {
			result = &models.PaymentDetailsForCashierWithDiscount{
				OrderId:       oID,
				Status:        oStatus,
				TableNumber:   tableNumber,
				CustomerName:  customerName,
				CustomerPhone: customerPhone,
				WaiterId:      resolvedWaiterID,
				WaiterName:    resolvedWaiterName,
				WaiterImage:   resolvedWaiterImage,
			}

			// attach token only when a matching row exists
			if tokenID != nil {
				result.Token = &models.UserToken{
					ID:          *tokenID,
					PhoneNumber: *tokenPhone,
					TotalTokens: *totalTokens,
					CreatedAt:   *tokenCreatedAt,
					UpdatedAt:   *tokenUpdatedAt,
				}
			}
		}

		result.OrderMenuItems = append(result.OrderMenuItems, models.OrderItemType{
			Id:        itemID,
			Price:     itemPrice,
			Quantity:  itemQty,
			OrderId:   oID,
			MenuId:    menuItemID,
			MenuImage: menuImage,
			Status:    itemStatus,
			MenuName:  menuName,
			CreatedAt: itemCreatedAt,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetAllOrderDetailsForCashierByOrderId rows: %w", err)
	}

	if result == nil {
		return nil, fmt.Errorf("order not found: %s", orderId)
	}

	return result, nil
}

func (r *paymentRepo) GetAllApprovedOrdersForCashier(ctx context.Context) ([]models.GetOrderDetailsForCashier, error) {

	query := `
		SELECT 
			o.id,
			o.status,
			ts.table_number,
			o.customer_name,
			o.customer_phone,
			o.waiter_id,
			COALESCE(u.name, ''),
			u.image,
			o.created_at,

			oi.id,
			oi.order_id,
			oi.menu_item_id,
			oi.quantity,
			oi.price,
			oi.status,
			COALESCE(mi.name, ''),
			mi.image_url

		FROM orders o
		JOIN table_session ts 
			ON o.table_session_id = ts.id

		LEFT JOIN users u 
			ON o.waiter_id = u.id

		LEFT JOIN order_items oi 
			ON oi.order_id = o.id

		LEFT JOIN menu_items mi
			ON mi.id = oi.menu_item_id

		WHERE 
			o.status = 'approved'
			AND DATE(o.created_at) = CURRENT_DATE

		ORDER BY o.created_at DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch orders: %w", err)
	}
	defer rows.Close()

	orderMap := make(map[uuid.UUID]*models.GetOrderDetailsForCashier)
	// preserve insertion order
	orderKeys := make([]uuid.UUID, 0)

	for rows.Next() {
		var (
			orderID       uuid.UUID
			orderStatus   models.OrderStatus
			tableNumber   int
			customerName  *string
			customerPhone *string
			waiterID      *uuid.UUID // nullable in DB
			waiterName    string
			waiterImage   *string
			createdAt     time.Time

			itemID      *uuid.UUID
			itemOrderID *uuid.UUID
			menuItemID  *uuid.UUID
			quantity    *float64
			price       *float64
			itemStatus  *models.OrderStatus
			menuName    string
			menuImage   *string
		)

		err := rows.Scan(
			&orderID,
			&orderStatus,
			&tableNumber,
			&customerName,
			&customerPhone,
			&waiterID,
			&waiterName,
			&waiterImage,
			&createdAt,

			&itemID,
			&itemOrderID,
			&menuItemID,
			&quantity,
			&price,
			&itemStatus,
			&menuName,
			&menuImage,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		if _, exists := orderMap[orderID]; !exists {
			order := &models.GetOrderDetailsForCashier{
				OrderId:        orderID,
				Status:         orderStatus,
				TableNumber:    tableNumber,
				CustomerName:   customerName,
				CustomerPhone:  customerPhone,
				WaiterName:     waiterName,
				WaiterImage:    waiterImage,
				CreatedAt:      createdAt,
				OrderMenuItems: []models.OrderItemType{},
			}
			// safely dereference nullable waiter_id
			if waiterID != nil {
				order.WaiterId = *waiterID
			}

			orderMap[orderID] = order
			orderKeys = append(orderKeys, orderID)
		}

		if itemID != nil && menuItemID != nil && quantity != nil && price != nil && itemStatus != nil && itemOrderID != nil {
			orderMap[orderID].OrderMenuItems = append(
				orderMap[orderID].OrderMenuItems,
				models.OrderItemType{
					Id:        *itemID,
					OrderId:   *itemOrderID,
					MenuId:    *menuItemID,
					Quantity:  *quantity,
					Price:     *price,
					Status:    *itemStatus,
					MenuName:  menuName,
					MenuImage: menuImage,
				},
			)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	// preserve DESC order from SQL (map iteration is random)
	result := make([]models.GetOrderDetailsForCashier, 0, len(orderMap))
	for _, id := range orderKeys {
		result = append(result, *orderMap[id])
	}

	return result, nil
}

// TODO: implement the deleting the expiry token with in 30 days
func (r *paymentRepo) CreatePayment(ctx context.Context, paymentData *models.CreatePayment) (*models.Payment, error) {

	if paymentData == nil {
		return nil, fmt.Errorf("payment data cannot be nil")
	}

	// 🔒 Basic validations
	if paymentData.OrderID == uuid.Nil {
		return nil, fmt.Errorf("order_id is required")
	}
	if paymentData.PaidAmount <= 0 {
		return nil, fmt.Errorf("paid_amount must be greater than 0")
	}
	if paymentData.PaymentMethod == "" {
		return nil, fmt.Errorf("payment_method is required")
	}
	if paymentData.PaymentMethod == models.PaymentMethodOnline && paymentData.OnlineGateway == nil {
		return nil, fmt.Errorf("online_gateway is required for online payments")
	}

	// Start transaction
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// 🔍 Step 1: Check order exists
	var exists bool
	err = tx.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM orders WHERE id = $1
		)
	`, paymentData.OrderID).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check order: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("order not found")
	}

	// 🔍 Step 2: Get table session info for this order
	var tableNumber int
	var phoneNumber string
	var tableSessionID uuid.UUID

	err = tx.QueryRow(ctx, `
		SELECT ts.id, ts.table_number, tv.phone_number
		FROM table_session ts
		JOIN table_validation tv
			ON ts.table_number = tv.table_number
		WHERE tv.table_number = ts.table_number
		LIMIT 1
	`).Scan(&tableSessionID, &tableNumber, &phoneNumber)
	if err != nil {
		// If no table session, continue without table cleanup
		tableSessionID = uuid.Nil
	}

	// 💾 Step 3: Insert payment
	var payment models.Payment
	query := `
		INSERT INTO payments (
			order_id,
			payment_method,
			online_gateway,
			paid_amount,
			discount,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING 
			id, order_id, payment_method, online_gateway, paid_amount, discount, created_at, updated_at
	`
	err = tx.QueryRow(ctx, query,
		paymentData.OrderID,
		paymentData.PaymentMethod,
		paymentData.OnlineGateway,
		paymentData.PaidAmount,
		paymentData.Discount,
	).Scan(
		&payment.ID,
		&payment.OrderID,
		&payment.PaymentMethod,
		&payment.OnlineGateway,
		&payment.PaidAmount,
		&payment.Discount,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	// 🔥 Step 4: Update order status
	_, err = tx.Exec(ctx, `
		UPDATE orders
		SET status = 'completed',
		    updated_at = NOW()
		WHERE id = $1
	`, paymentData.OrderID)
	if err != nil {
		return nil, fmt.Errorf("failed to update order payment status: %w", err)
	}

	// 🔄 Step 5: Table/session cleanup
	if tableSessionID != uuid.Nil {
		// 5a: Delete table_validation entry
		_, err = tx.Exec(ctx, `
			DELETE FROM table_validation
			WHERE table_number = $1 AND phone_number = $2
		`, tableNumber, phoneNumber)
		if err != nil {
			return nil, fmt.Errorf("failed to delete table validation: %w", err)
		}

		// 5b: Update table_session to close session
		_, err = tx.Exec(ctx, `
			UPDATE table_session
			SET close_time = NOW(),
			    updated_at = NOW()
			WHERE id = $1
		`, tableSessionID)
		if err != nil {
			return nil, fmt.Errorf("failed to update table session: %w", err)
		}

		// 5c: Update table_status to empty
		_, err = tx.Exec(ctx, `
			UPDATE table_status
			SET status = 'empty',
			    updated_at = NOW()
			WHERE table_number = $1
		`, tableNumber)
		if err != nil {
			return nil, fmt.Errorf("failed to update table status: %w", err)
		}
	}

	// 🚀 Step 6: Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &payment, nil
}

func NewPaymentRepository() PaymentRepo {
	pool, err := database.GetPostgresPool()
	if err != nil {
		return nil
	}

	return &paymentRepo{
		pool: pool,
	}
}
