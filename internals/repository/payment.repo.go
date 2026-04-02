package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Abhishekh669/backend/internals/database"
	"github.com/Abhishekh669/backend/internals/lib"
	"github.com/Abhishekh669/backend/internals/models"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const guestPhone = "9800000000"

// PaymentRepo interface
type PaymentRepo interface {
	DeleteOrderByCashier(ctx context.Context, orderId uuid.UUID) error
	GetAllOrderDetailsForCashierByOrderId(ctx context.Context, orderId uuid.UUID) (*models.PaymentDetailsForCashierWithDiscount, error)
	GetAllApprovedOrdersForCashier(ctx context.Context) ([]models.GetOrderDetailsForCashier, error)
	CreatePayment(ctx context.Context, paymentData *models.CreatePayment) (*models.Payment, error)
}

// paymentRepo struct
type paymentRepo struct {
	pool *pgxpool.Pool
}

func (r *paymentRepo) DeleteOrderByCashier(ctx context.Context, orderId uuid.UUID) error {
	// Implementation for deleting order by cashier
	query := `DELETE FROM orders WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, orderId)
	if err != nil {
		return fmt.Errorf("failed to delete order: %w", err)
	}
	return nil
}

// ─────────────────────────────────────────────
// GetAllOrderDetailsForCashierByOrderId
// Returns order details + token info + pre-calculated discount preview
// ─────────────────────────────────────────────
// ─────────────────────────────────────────────
// GetAllOrderDetailsForCashierByOrderId
// Returns order details + token info + pre-calculated discount preview
// ─────────────────────────────────────────────
func (r *paymentRepo) GetAllOrderDetailsForCashierByOrderId(ctx context.Context, orderId uuid.UUID) (*models.PaymentDetailsForCashierWithDiscount, error) {
	// 1. Fetch the order
	order, err := r.GetOrderById(ctx, orderId)
	if err != nil {
		// If order not found, return a proper error
		return nil, fmt.Errorf("order not found: %w", err)
	}

	// Ensure order is not nil
	if order == nil {
		return nil, fmt.Errorf("order is nil for ID: %v", orderId)
	}

	// 2 & 3. Skip token/streak fetch entirely for guest phone
	var token *models.UserToken
	var streak *models.CustomerStreak
	if order.CustomerPhone != nil && *order.CustomerPhone != guestPhone && *order.CustomerPhone != "" {
		token, _ = r.GetUserTokenByPhone(ctx, *order.CustomerPhone)
		streak, _ = r.GetCustomerStreakByPhone(ctx, *order.CustomerPhone)
	}

	// 4. Combine token + streak info
	var tokenDetails *models.TokenDetailsOfCustomer

	// Calculate discount first
	discount := 0.0
	if token != nil && token.TotalTokens > 100 {
		// Sum up total bill
		totalAmount := 0.0
		for _, item := range order.OrderMenuItems {
			totalAmount += item.Price * item.Quantity
		}

		rawDiscount := lib.CalculateDiscountFromTokens(token.TotalTokens) // 50% of tokens

		// If discount exceeds the bill, cap it at bill total
		if rawDiscount >= totalAmount {
			discount = totalAmount
		} else {
			discount = rawDiscount
		}
	}

	// Only create tokenDetails if we have token or streak data
	if token != nil || streak != nil {
		tokenDetails = &models.TokenDetailsOfCustomer{
			Token:         token,
			CurrentStreak: 0,
			LastVisit:     nil,
			MonthlyDays:   0,
			Discount:      discount,
		}
		if streak != nil {
			tokenDetails.CurrentStreak = streak.CurrentStreak
			tokenDetails.LastVisit = streak.LastVisit
			tokenDetails.MonthlyDays = streak.MonthlyDays
		}
	} else {
		// Even if no token/streak, we might still want to create tokenDetails for discount preview
		// Only create if discount > 0 or you want to show discount info
		if discount > 0 {
			tokenDetails = &models.TokenDetailsOfCustomer{
				Token:         nil,
				CurrentStreak: 0,
				LastVisit:     nil,
				MonthlyDays:   0,
				Discount:      discount,
			}
		}
	}

	// Ensure OrderMenuItems is never nil (initialize empty slice if needed)
	if order.OrderMenuItems == nil {
		order.OrderMenuItems = []models.OrderItemType{}
	}

	// Handle potential nil values with defaults
	customerName := order.CustomerName
	if customerName == nil {
		defaultName := "Guest"
		customerName = &defaultName
	}

	customerPhone := order.CustomerPhone
	if customerPhone == nil {
		defaultPhone := ""
		customerPhone = &defaultPhone
	}

	waiterName := order.WaiterName
	if waiterName == "" {
		waiterName = "N/A"
	}

	// Handle WaiterId (if it's zero UUID, use nil or zero)
	waiterId := order.WaiterId
	if waiterId == uuid.Nil {
		// If it's zero UUID, you might want to use a default or keep it
		// For now, keep it as is
	}

	// 6. Construct final result
	result := &models.PaymentDetailsForCashierWithDiscount{
		TokenDetails:   tokenDetails,
		OrderMenuItems: order.OrderMenuItems,
		OrderId:        order.OrderId,
		Status:         order.Status,
		TableNumber:    order.TableNumber,
		CustomerName:   customerName,
		CustomerPhone:  customerPhone,
		WaiterId:       waiterId,
		WaiterName:     waiterName,
		WaiterImage:    order.WaiterImage, // WaiterImage can be nil, that's fine
	}

	return result, nil
}

// ─────────────────────────────────────────────
// CreatePayment
// ─────────────────────────────────────────────
func (r *paymentRepo) CreatePayment(ctx context.Context, paymentData *models.CreatePayment) (*models.Payment, error) {

	if paymentData == nil {
		return nil, fmt.Errorf("payment data cannot be nil")
	}

	// Basic validations
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

	// ── Step 1: Fetch order + customer phone ──────────────────────────────
	// ── Step 1: Fetch order + customer phone + table session ───────────────
	var phone *string
	var orderExists bool
	var tableSessionID uuid.UUID

	err = tx.QueryRow(ctx, `
	SELECT 
		EXISTS(SELECT 1 FROM orders WHERE id = $1),
		customer_phone,
		table_session_id
	FROM orders 
	WHERE id = $1
`, paymentData.OrderID).Scan(&orderExists, &phone, &tableSessionID)

	if err != nil {
		return nil, fmt.Errorf("failed to check order: %w", err)
	}

	if tableSessionID == uuid.Nil {
		return nil, fmt.Errorf("invalid table session associated with order")
	}

	if !orderExists {
		return nil, fmt.Errorf("order not found")
	}

	// ── Step 2: Fetch total bill amount from order items ──────────────────
	var totalAmount float64
	err = tx.QueryRow(ctx, `
		SELECT COALESCE(SUM(price * quantity), 0)
		FROM order_items WHERE order_id = $1
	`, paymentData.OrderID).Scan(&totalAmount)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate total amount: %w", err)
	}
	var tableNumber int
	var tablePhoneNumber string
	// ── Step 3: Get table session info using order ─────────────────────────
	err = tx.QueryRow(ctx, `
		SELECT ts.table_number, tv.phone_number
		FROM table_session ts
		LEFT JOIN table_validation tv 
			ON ts.table_number = tv.table_number
		WHERE ts.id = $1
		LIMIT 1
	`, tableSessionID).Scan(&tableNumber, &tablePhoneNumber)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// No session found — safe fallback
			return nil, fmt.Errorf("no active table session found for order: %w", err)
		} else {
			// Real DB error
			return nil, fmt.Errorf("failed to fetch table session: %w", err)
		}
	}

	// ── Step 4: Discount + Token Logic (skip entirely for guest) ──────────
	discount := 0.0
	isGuest := phone == nil || *phone == guestPhone

	if !isGuest {

		var currentTokens float64
		err = tx.QueryRow(ctx, `
			SELECT total_tokens FROM user_tokens WHERE phone_number = $1
		`, *phone).Scan(&currentTokens)
		if err != nil {
			currentTokens = 0
		}

		if currentTokens > 100 {
			rawDiscount := lib.CalculateDiscountFromTokens(currentTokens)

			var tokensUsed float64

			if rawDiscount >= totalAmount {

				discount = totalAmount
				remainingDiscountValue := rawDiscount - totalAmount
				remainingTokens := remainingDiscountValue * 2
				tokensUsed = currentTokens - remainingTokens

				_, err = tx.Exec(ctx, `
					UPDATE user_tokens
					SET total_tokens = $1, updated_at = NOW()
					WHERE phone_number = $2
				`, remainingTokens, *phone)
				if err != nil {
					return nil, fmt.Errorf("failed to update remaining tokens: %w", err)
				}

			} else {

				discount = rawDiscount
				tokensUsed = currentTokens

				_, err = tx.Exec(ctx, `
					UPDATE user_tokens
					SET total_tokens = 0, updated_at = NOW()
					WHERE phone_number = $1
				`, *phone)
				if err != nil {
					return nil, fmt.Errorf("failed to reset user tokens: %w", err)
				}
			}

			_, err = tx.Exec(ctx, `
				INSERT INTO token_transactions (phone_number, amount, type, source, reference_id)
				VALUES ($1, $2, 'SPEND', 'DISCOUNT', $3)
			`, *phone, tokensUsed, paymentData.OrderID)
			if err != nil {
				return nil, fmt.Errorf("failed to insert token spend transaction: %w", err)
			}
		}

		amountAfterDiscount := totalAmount - discount
		earnedTokens := lib.CalculateOrderTokens(amountAfterDiscount)

		if earnedTokens > 0 {

			_, err = tx.Exec(ctx, `
				INSERT INTO user_tokens (phone_number, total_tokens)
				VALUES ($1, $2)
				ON CONFLICT (phone_number)
				DO UPDATE SET
					total_tokens = user_tokens.total_tokens + $2,
					updated_at = NOW()
			`, *phone, earnedTokens)

			if err != nil {
				return nil, fmt.Errorf("failed to update earned tokens: %w", err)
			}

			_, err = tx.Exec(ctx, `
				INSERT INTO token_transactions (phone_number, amount, type, source, reference_id)
				VALUES ($1, $2, 'EARN', 'ORDER', $3)
			`, *phone, earnedTokens, paymentData.OrderID)

			if err != nil {
				return nil, fmt.Errorf("failed to insert earn token transaction: %w", err)
			}
		}

		// ── Streak logic (unchanged) ───────────────────────────────────────
		today := time.Now().UTC().Truncate(24 * time.Hour)

		var lastVisit *time.Time
		var currentStreak, monthlyDays int

		err = tx.QueryRow(ctx, `
			SELECT last_visit, current_streak, monthly_days
			FROM customer_streaks WHERE phone_number = $1
		`, *phone).Scan(&lastVisit, &currentStreak, &monthlyDays)

		if err != nil {

			_, err = tx.Exec(ctx, `
				INSERT INTO customer_streaks (phone_number, current_streak, last_visit, monthly_days)
				VALUES ($1, 1, $2, 1)
			`, *phone, today)

			if err != nil {
				return nil, fmt.Errorf("failed to insert customer streak: %w", err)
			}

			streakTokens := lib.CalculateStreakTokens(1, 1)

			if streakTokens > 0 {

				_, err = tx.Exec(ctx, `
					INSERT INTO user_tokens (phone_number, total_tokens)
					VALUES ($1, $2)
					ON CONFLICT (phone_number)
					DO UPDATE SET
						total_tokens = user_tokens.total_tokens + $2,
						updated_at = NOW()
				`, *phone, streakTokens)

				if err != nil {
					return nil, fmt.Errorf("failed to update streak tokens: %w", err)
				}

				_, err = tx.Exec(ctx, `
					INSERT INTO token_transactions (phone_number, amount, type, source, reference_id)
					VALUES ($1, $2, 'STREAK', 'STREAK', $3)
				`, *phone, streakTokens, paymentData.OrderID)

				if err != nil {
					return nil, fmt.Errorf("failed to insert streak token transaction: %w", err)
				}
			}
		}

	}

	// ── Step 5: Insert payment ───────────────────────────────────────────
	var payment models.Payment

	err = tx.QueryRow(ctx, `
		INSERT INTO payments (
			order_id, payment_method, online_gateway,
			paid_amount, discount, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, order_id, payment_method, online_gateway, paid_amount, discount, created_at, updated_at
	`,
		paymentData.OrderID,
		paymentData.PaymentMethod,
		paymentData.OnlineGateway,
		paymentData.PaidAmount,
		discount,
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

	// ── Step 6: Mark order completed ─────────────────────────────────────
	_, err = tx.Exec(ctx, `
		UPDATE orders SET status = 'completed'
		WHERE id = $1
	`, paymentData.OrderID)

	if err != nil {
		return nil, fmt.Errorf("failed to update order status: %w", err)
	}

	// ── Step 7: Table/session cleanup ─────────────────────────────────────
	if tableSessionID != uuid.Nil {

		_, err = tx.Exec(ctx, `
			DELETE FROM table_validation
			WHERE table_number = $1 AND phone_number = $2
		`, tableNumber, tablePhoneNumber)

		if err != nil {
			return nil, fmt.Errorf("failed to delete table validation: %w", err)
		}

		_, err = tx.Exec(ctx, `
			UPDATE table_session 
			SET close_time = NOW(), updated_at = NOW()
			WHERE id = $1
		`, tableSessionID)

		if err != nil {
			return nil, fmt.Errorf("failed to update table session: %w", err)
		}

		_, err = tx.Exec(ctx, `
			UPDATE table_status 
			SET status = 'empty'
			WHERE table_number = $1
		`, tableNumber)

		if err != nil {
			return nil, fmt.Errorf("failed to update table status: %w", err)
		}
	}

	// ── Step 8: Commit ───────────────────────────────────────────────────
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &payment, nil
}

// ─────────────────────────────────────────────
// Helper Functions
// ─────────────────────────────────────────────

func (r *paymentRepo) GetOrderById(ctx context.Context, orderId uuid.UUID) (*models.GetOrderDetailsForCashier, error) {
	order := &models.GetOrderDetailsForCashier{
		OrderMenuItems: []models.OrderItemType{}, // Initialize empty slice
		WaiterName:     "N/A",                    // Default waiter name
	}

	query := `
		SELECT o.id, o.customer_name, o.customer_phone, o.status, ts.table_number,
		       COALESCE(u.id, '00000000-0000-0000-0000-000000000000') AS waiter_id, 
		       COALESCE(u.name, 'N/A') AS waiter_name, 
		       u.image AS waiter_image, 
		       o.created_at
		FROM orders o
		JOIN table_session ts ON ts.id = o.table_session_id
		LEFT JOIN users u ON u.id = o.waiter_id
		WHERE o.id = $1
	`

	err := r.pool.QueryRow(ctx, query, orderId).Scan(
		&order.OrderId,
		&order.CustomerName,
		&order.CustomerPhone,
		&order.Status,
		&order.TableNumber,
		&order.WaiterId,
		&order.WaiterName,
		&order.WaiterImage,
		&order.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	itemsQuery := `
		SELECT oi.id, oi.order_id, oi.menu_item_id, 
		       COALESCE(mi.name, 'Unknown Item') AS name, 
		       mi.image_url, 
		       oi.quantity, oi.price, oi.status, oi.created_at
		FROM order_items oi
		LEFT JOIN menu_items mi ON mi.id = oi.menu_item_id
		WHERE oi.order_id = $1
	`

	rows, err := r.pool.Query(ctx, itemsQuery, orderId)
	if err != nil {
		return nil, fmt.Errorf("failed to get order items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item models.OrderItemType
		if err := rows.Scan(
			&item.Id,
			&item.OrderId,
			&item.MenuId,
			&item.MenuName,
			&item.MenuImage,
			&item.Quantity,
			&item.Price,
			&item.Status,
			&item.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan order item: %w", err)
		}

		// Ensure MenuName has a default if empty
		if item.MenuName == "" {
			item.MenuName = "Unknown Item"
		}

		order.OrderMenuItems = append(order.OrderMenuItems, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return order, nil
}
func (r *paymentRepo) GetUserTokenByPhone(ctx context.Context, phone string) (*models.UserToken, error) {
	token := &models.UserToken{}
	query := `SELECT id, phone_number, total_tokens, created_at, updated_at FROM user_tokens WHERE phone_number = $1`
	err := r.pool.QueryRow(ctx, query, phone).Scan(
		&token.ID,
		&token.PhoneNumber,
		&token.TotalTokens,
		&token.CreatedAt,
		&token.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (r *paymentRepo) GetCustomerStreakByPhone(ctx context.Context, phone string) (*models.CustomerStreak, error) {
	streak := &models.CustomerStreak{}
	query := `SELECT phone_number, current_streak, last_visit, monthly_days, created_at, updated_at
			  FROM customer_streaks WHERE phone_number = $1`
	err := r.pool.QueryRow(ctx, query, phone).Scan(
		&streak.PhoneNumber,
		&streak.CurrentStreak,
		&streak.LastVisit,
		&streak.MonthlyDays,
		&streak.CreatedAt,
		&streak.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return streak, nil
}

func (r *paymentRepo) GetAllApprovedOrdersForCashier(ctx context.Context) ([]models.GetOrderDetailsForCashier, error) {
	query := `
		SELECT 
			o.id, o.status, ts.table_number,
			o.customer_name, o.customer_phone,
			o.waiter_id, COALESCE(u.name, ''), u.image, o.created_at,
			oi.id, oi.order_id, oi.menu_item_id, oi.quantity, oi.price, oi.status,
			COALESCE(mi.name, ''), mi.image_url
		FROM orders o
		JOIN table_session ts ON o.table_session_id = ts.id
		LEFT JOIN users u ON o.waiter_id = u.id
		LEFT JOIN order_items oi ON oi.order_id = o.id
		LEFT JOIN menu_items mi ON mi.id = oi.menu_item_id
		WHERE o.status = 'approved' AND DATE(o.created_at) = CURRENT_DATE
		ORDER BY o.created_at DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch orders: %w", err)
	}
	defer rows.Close()

	orderMap := make(map[uuid.UUID]*models.GetOrderDetailsForCashier)
	orderKeys := make([]uuid.UUID, 0)

	for rows.Next() {
		var (
			orderID                         uuid.UUID
			orderStatus                     models.OrderStatus
			tableNumber                     int
			customerName, customerPhone     *string
			waiterID                        *uuid.UUID
			waiterName                      string
			waiterImage                     *string
			createdAt                       time.Time
			itemID, itemOrderID, menuItemID *uuid.UUID
			quantity, price                 *float64
			itemStatus                      *models.OrderStatus
			menuName                        string
			menuImage                       *string
		)

		err := rows.Scan(
			&orderID, &orderStatus, &tableNumber,
			&customerName, &customerPhone,
			&waiterID, &waiterName, &waiterImage, &createdAt,
			&itemID, &itemOrderID, &menuItemID, &quantity, &price, &itemStatus,
			&menuName, &menuImage,
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

	result := make([]models.GetOrderDetailsForCashier, 0, len(orderMap))
	for _, id := range orderKeys {
		result = append(result, *orderMap[id])
	}

	return result, nil
}

func NewPaymentRepository() PaymentRepo {
	pool, err := database.GetPostgresPool()
	if err != nil {
		return nil
	}
	return &paymentRepo{pool: pool}
}

// package repository
// import (
// 	"context"
// 	"fmt"
// 	"time"

// 	"github.com/Abhishekh669/backend/internals/database"
// 	"github.com/Abhishekh669/backend/internals/models"
// 	"github.com/gofrs/uuid"
// 	"github.com/jackc/pgx/v5/pgxpool"
// )

// // PaymentRepo interface
// type PaymentRepo interface {
// 	GetAllOrderDetailsForCashierByOrderId(ctx context.Context, orderId uuid.UUID) (*models.PaymentDetailsForCashierWithDiscount, error)
// 	GetAllApprovedOrdersForCashier(ctx context.Context) ([]models.GetOrderDetailsForCashier, error)
// 	CreatePayment(ctx context.Context, paymentData *models.CreatePayment) (*models.Payment, error)
// }

// // paymentRepo struct
// type paymentRepo struct {
// 	pool *pgxpool.Pool
// }

// func (r *paymentRepo) GetAllOrderDetailsForCashierByOrderId(ctx context.Context, orderId uuid.UUID) (*models.PaymentDetailsForCashierWithDiscount, error) {
// 	// 1. Fetch the order
// 	order, err := r.GetOrderById(ctx, orderId)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// 2. Fetch the customer token details
// 	const guestPhone = "9800000000"

// 	var token *models.UserToken
// 	if order.CustomerPhone != nil && *order.CustomerPhone != guestPhone {
// 		token, _ = r.GetUserTokenByPhone(ctx, *order.CustomerPhone)
// 	}

// 	// 3. Fetch the customer streak details
// 	var streak *models.CustomerStreak
// 	if order.CustomerPhone != nil && *order.CustomerPhone != guestPhone {
// 		streak, _ = r.GetCustomerStreakByPhone(ctx, *order.CustomerPhone)
// 	}

// 	// 4. Combine token + streak info
// 	var tokenDetails *models.TokenDetailsOfCustomer
// 	if token != nil || streak != nil {
// 		tokenDetails = &models.TokenDetailsOfCustomer{
// 			Token:         token,
// 			CurrentStreak: 0,
// 			LastVisit:     nil,
// 			MonthlyDays:   0,
// 		}
// 		if streak != nil {
// 			tokenDetails.CurrentStreak = streak.CurrentStreak
// 			tokenDetails.LastVisit = streak.LastVisit
// 			tokenDetails.MonthlyDays = streak.MonthlyDays
// 		}
// 	}

// 	// 5. Construct final result
// 	result := &models.PaymentDetailsForCashierWithDiscount{
// 		TokenDetails:   tokenDetails,
// 		OrderMenuItems: order.OrderMenuItems,
// 		OrderId:        order.OrderId,
// 		Status:         order.Status,
// 		TableNumber:    order.TableNumber,
// 		CustomerName:   order.CustomerName,
// 		CustomerPhone:  order.CustomerPhone,
// 		WaiterId:       order.WaiterId,
// 		WaiterName:     order.WaiterName,
// 		WaiterImage:    order.WaiterImage,
// 	}

// 	return result, nil
// }

// // ---------------- Helper Functions ----------------

// // GetOrderById fetches order + items
// func (r *paymentRepo) GetOrderById(ctx context.Context, orderId uuid.UUID) (*models.GetOrderDetailsForCashier, error) {
// 	order := &models.GetOrderDetailsForCashier{}

// 	// Fetch order info
// 	query := `
// 		SELECT o.id, o.customer_name, o.customer_phone, o.status, ts.table_number,
// 		       u.id AS waiter_id, u.name AS waiter_name, u.image AS waiter_image, o.created_at
// 		FROM orders o
// 		JOIN table_session ts ON ts.id = o.table_session_id
// 		LEFT JOIN users u ON u.id = o.waiter_id
// 		WHERE o.id = $1
// 	`
// 	err := r.pool.QueryRow(ctx, query, orderId).Scan(
// 		&order.OrderId,
// 		&order.CustomerName,
// 		&order.CustomerPhone,
// 		&order.Status,
// 		&order.TableNumber,
// 		&order.WaiterId,
// 		&order.WaiterName,
// 		&order.WaiterImage,
// 		&order.CreatedAt,
// 	)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get order: %w", err)
// 	}

// 	// Fetch order items
// 	itemsQuery := `
// 		SELECT oi.id, oi.order_id, oi.menu_item_id, mi.name, mi.image_url, oi.quantity, oi.price, oi.status, oi.created_at
// 		FROM order_items oi
// 		JOIN menu_items mi ON mi.id = oi.menu_item_id
// 		WHERE oi.order_id = $1
// 	`
// 	rows, err := r.pool.Query(ctx, itemsQuery, orderId)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get order items: %w", err)
// 	}
// 	defer rows.Close()

// 	for rows.Next() {
// 		var item models.OrderItemType
// 		if err := rows.Scan(
// 			&item.Id,
// 			&item.OrderId,
// 			&item.MenuId,
// 			&item.MenuName,
// 			&item.MenuImage,
// 			&item.Quantity,
// 			&item.Price,
// 			&item.Status,
// 			&item.CreatedAt,
// 		); err != nil {
// 			return nil, fmt.Errorf("failed to scan order item: %w", err)
// 		}
// 		order.OrderMenuItems = append(order.OrderMenuItems, item)
// 	}

// 	return order, nil
// }

// // GetUserTokenByPhone fetches user token
// func (r *paymentRepo) GetUserTokenByPhone(ctx context.Context, phone string) (*models.UserToken, error) {
// 	token := &models.UserToken{}
// 	query := `SELECT id, phone_number, total_tokens, created_at, updated_at FROM user_tokens WHERE phone_number = $1`
// 	err := r.pool.QueryRow(ctx, query, phone).Scan(
// 		&token.ID,
// 		&token.PhoneNumber,
// 		&token.TotalTokens,
// 		&token.CreatedAt,
// 		&token.UpdatedAt,
// 	)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return token, nil
// }

// // GetCustomerStreakByPhone fetches streak info
// func (r *paymentRepo) GetCustomerStreakByPhone(ctx context.Context, phone string) (*models.CustomerStreak, error) {
// 	streak := &models.CustomerStreak{}
// 	query := `SELECT phone_number, current_streak, last_visit, monthly_days, created_at, updated_at
// 			  FROM customer_streaks
// 			  WHERE phone_number = $1`
// 	err := r.pool.QueryRow(ctx, query, phone).Scan(
// 		&streak.PhoneNumber,
// 		&streak.CurrentStreak,
// 		&streak.LastVisit,
// 		&streak.MonthlyDays,
// 		&streak.CreatedAt,
// 		&streak.UpdatedAt,
// 	)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return streak, nil
// }

// func (r *paymentRepo) GetAllApprovedOrdersForCashier(ctx context.Context) ([]models.GetOrderDetailsForCashier, error) {

// 	query := `
// 		SELECT
// 			o.id,
// 			o.status,
// 			ts.table_number,
// 			o.customer_name,
// 			o.customer_phone,
// 			o.waiter_id,
// 			COALESCE(u.name, ''),
// 			u.image,
// 			o.created_at,

// 			oi.id,
// 			oi.order_id,
// 			oi.menu_item_id,
// 			oi.quantity,
// 			oi.price,
// 			oi.status,
// 			COALESCE(mi.name, ''),
// 			mi.image_url

// 		FROM orders o
// 		JOIN table_session ts
// 			ON o.table_session_id = ts.id

// 		LEFT JOIN users u
// 			ON o.waiter_id = u.id

// 		LEFT JOIN order_items oi
// 			ON oi.order_id = o.id

// 		LEFT JOIN menu_items mi
// 			ON mi.id = oi.menu_item_id

// 		WHERE
// 			o.status = 'approved'
// 			AND DATE(o.created_at) = CURRENT_DATE

// 		ORDER BY o.created_at DESC
// 	`

// 	rows, err := r.pool.Query(ctx, query)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to fetch orders: %w", err)
// 	}
// 	defer rows.Close()

// 	orderMap := make(map[uuid.UUID]*models.GetOrderDetailsForCashier)
// 	// preserve insertion order
// 	orderKeys := make([]uuid.UUID, 0)

// 	for rows.Next() {
// 		var (
// 			orderID       uuid.UUID
// 			orderStatus   models.OrderStatus
// 			tableNumber   int
// 			customerName  *string
// 			customerPhone *string
// 			waiterID      *uuid.UUID // nullable in DB
// 			waiterName    string
// 			waiterImage   *string
// 			createdAt     time.Time

// 			itemID      *uuid.UUID
// 			itemOrderID *uuid.UUID
// 			menuItemID  *uuid.UUID
// 			quantity    *float64
// 			price       *float64
// 			itemStatus  *models.OrderStatus
// 			menuName    string
// 			menuImage   *string
// 		)

// 		err := rows.Scan(
// 			&orderID,
// 			&orderStatus,
// 			&tableNumber,
// 			&customerName,
// 			&customerPhone,
// 			&waiterID,
// 			&waiterName,
// 			&waiterImage,
// 			&createdAt,

// 			&itemID,
// 			&itemOrderID,
// 			&menuItemID,
// 			&quantity,
// 			&price,
// 			&itemStatus,
// 			&menuName,
// 			&menuImage,
// 		)
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to scan row: %w", err)
// 		}

// 		if _, exists := orderMap[orderID]; !exists {
// 			order := &models.GetOrderDetailsForCashier{
// 				OrderId:        orderID,
// 				Status:         orderStatus,
// 				TableNumber:    tableNumber,
// 				CustomerName:   customerName,
// 				CustomerPhone:  customerPhone,
// 				WaiterName:     waiterName,
// 				WaiterImage:    waiterImage,
// 				CreatedAt:      createdAt,
// 				OrderMenuItems: []models.OrderItemType{},
// 			}
// 			// safely dereference nullable waiter_id
// 			if waiterID != nil {
// 				order.WaiterId = *waiterID
// 			}

// 			orderMap[orderID] = order
// 			orderKeys = append(orderKeys, orderID)
// 		}

// 		if itemID != nil && menuItemID != nil && quantity != nil && price != nil && itemStatus != nil && itemOrderID != nil {
// 			orderMap[orderID].OrderMenuItems = append(
// 				orderMap[orderID].OrderMenuItems,
// 				models.OrderItemType{
// 					Id:        *itemID,
// 					OrderId:   *itemOrderID,
// 					MenuId:    *menuItemID,
// 					Quantity:  *quantity,
// 					Price:     *price,
// 					Status:    *itemStatus,
// 					MenuName:  menuName,
// 					MenuImage: menuImage,
// 				},
// 			)
// 		}
// 	}

// 	if err := rows.Err(); err != nil {
// 		return nil, fmt.Errorf("row iteration error: %w", err)
// 	}

// 	// preserve DESC order from SQL (map iteration is random)
// 	result := make([]models.GetOrderDetailsForCashier, 0, len(orderMap))
// 	for _, id := range orderKeys {
// 		result = append(result, *orderMap[id])
// 	}

// 	return result, nil
// }

// // TODO: implement the deleting the expiry token with in 30 days
// func (r *paymentRepo) CreatePayment(ctx context.Context, paymentData *models.CreatePayment) (*models.Payment, error) {

// 	if paymentData == nil {
// 		return nil, fmt.Errorf("payment data cannot be nil")
// 	}

// 	// 🔒 Basic validations
// 	if paymentData.OrderID == uuid.Nil {
// 		return nil, fmt.Errorf("order_id is required")
// 	}
// 	if paymentData.PaidAmount <= 0 {
// 		return nil, fmt.Errorf("paid_amount must be greater than 0")
// 	}
// 	if paymentData.PaymentMethod == "" {
// 		return nil, fmt.Errorf("payment_method is required")
// 	}
// 	if paymentData.PaymentMethod == models.PaymentMethodOnline && paymentData.OnlineGateway == nil {
// 		return nil, fmt.Errorf("online_gateway is required for online payments")
// 	}

// 	// Start transaction
// 	tx, err := r.pool.Begin(ctx)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to begin transaction: %w", err)
// 	}
// 	defer tx.Rollback(ctx)

// 	// 🔍 Step 1: Check order exists
// 	var exists bool
// 	err = tx.QueryRow(ctx, `
// 		SELECT EXISTS(
// 			SELECT 1 FROM orders WHERE id = $1
// 		)
// 	`, paymentData.OrderID).Scan(&exists)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to check order: %w", err)
// 	}
// 	if !exists {
// 		return nil, fmt.Errorf("order not found")
// 	}

// 	// 🔍 Step 2: Get table session info for this order
// 	var tableNumber int
// 	var phoneNumber string
// 	var tableSessionID uuid.UUID

// 	err = tx.QueryRow(ctx, `
// 		SELECT ts.id, ts.table_number, tv.phone_number
// 		FROM table_session ts
// 		JOIN table_validation tv
// 			ON ts.table_number = tv.table_number
// 		WHERE tv.table_number = ts.table_number
// 		LIMIT 1
// 	`).Scan(&tableSessionID, &tableNumber, &phoneNumber)
// 	if err != nil {
// 		// If no table session, continue without table cleanup
// 		tableSessionID = uuid.Nil
// 	}

// 	// 💾 Step 3: Insert payment
// 	var payment models.Payment
// 	query := `
// 		INSERT INTO payments (
// 			order_id,
// 			payment_method,
// 			online_gateway,
// 			paid_amount,
// 			discount,
// 			created_at,
// 			updated_at
// 		)
// 		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
// 		RETURNING
// 			id, order_id, payment_method, online_gateway, paid_amount, discount, created_at, updated_at
// 	`
// 	err = tx.QueryRow(ctx, query,
// 		paymentData.OrderID,
// 		paymentData.PaymentMethod,
// 		paymentData.OnlineGateway,
// 		paymentData.PaidAmount,
// 		paymentData.Discount,
// 	).Scan(
// 		&payment.ID,
// 		&payment.OrderID,
// 		&payment.PaymentMethod,
// 		&payment.OnlineGateway,
// 		&payment.PaidAmount,
// 		&payment.Discount,
// 		&payment.CreatedAt,
// 		&payment.UpdatedAt,
// 	)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create payment: %w", err)
// 	}

// 	// 🔥 Step 4: Update order status
// 	_, err = tx.Exec(ctx, `
// 		UPDATE orders
// 		SET status = 'completed',
// 		    updated_at = NOW()
// 		WHERE id = $1
// 	`, paymentData.OrderID)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to update order payment status: %w", err)
// 	}

// 	// 🔄 Step 5: Table/session cleanup
// 	if tableSessionID != uuid.Nil {
// 		// 5a: Delete table_validation entry
// 		_, err = tx.Exec(ctx, `
// 			DELETE FROM table_validation
// 			WHERE table_number = $1 AND phone_number = $2
// 		`, tableNumber, phoneNumber)
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to delete table validation: %w", err)
// 		}

// 		// 5b: Update table_session to close session
// 		_, err = tx.Exec(ctx, `
// 			UPDATE table_session
// 			SET close_time = NOW(),
// 			    updated_at = NOW()
// 			WHERE id = $1
// 		`, tableSessionID)
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to update table session: %w", err)
// 		}

// 		// 5c: Update table_status to empty
// 		_, err = tx.Exec(ctx, `
// 			UPDATE table_status
// 			SET status = 'empty',
// 			    updated_at = NOW()
// 			WHERE table_number = $1
// 		`, tableNumber)
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to update table status: %w", err)
// 		}
// 	}

// 	// 🚀 Step 6: Commit transaction
// 	if err := tx.Commit(ctx); err != nil {
// 		return nil, fmt.Errorf("failed to commit transaction: %w", err)
// 	}

// 	return &payment, nil
// }

// func NewPaymentRepository() PaymentRepo {
// 	pool, err := database.GetPostgresPool()
// 	if err != nil {
// 		return nil
// 	}

// 	return &paymentRepo{
// 		pool: pool,
// 	}
// }
