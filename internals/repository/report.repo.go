package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Abhishekh669/backend/internals/database"
	"github.com/Abhishekh669/backend/internals/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ─── Interface ────────────────────────────────────────────────────────────────

type ReportRepo interface {
	NewGetCustomRangeRawMaterialReport(ctx context.Context, req *models.NewRawMaterialCustomRangeReportRequest) (*models.NewCustomRangeRawMaterialResponse, error)
	NewGetDefaultRawMaterialReport(ctx context.Context) (*models.NewDefaultRawMaterialResponse, error)

	NewGetDefaultStaffReport(ctx context.Context) (*models.NewDefaultStaffResponse, error)
	NewGetCustomRangeStaffReport(ctx context.Context, req *models.NewStaffCustomRangeReportRequest) (*models.NewCustomRangeStaffResponse, error)

	NewGetDefaultTableReport(ctx context.Context) (*models.NewDefaultTableResponse, error)
	NewGetCustomRangeTableReport(ctx context.Context, req *models.NewTableCustomRangeReportRequest) (*models.NewCustomRangeTableResponse, error)

	NewGetDefaultCustomerReport(ctx context.Context) (*models.NewDefaultCustomerResponse, error)
	NewGetCustomRangeCustomerReport(ctx context.Context, req *models.NewCustomerCustomRangeReportRequest) (*models.NewCustomRangeCustomerResponse, error)

	NewGetDefaultSalesReport(ctx context.Context) (*models.NewDefaultSalesResponse, error)
	NewGetCustomRangeSalesReport(ctx context.Context, req *models.NewSalesCustomRangeReportRequest) (*models.NewCustomRangeSalesResponse, error)

	NewGetCustomRangeRevenueReport(ctx context.Context, req *models.NewCustomRangeReportRequest) (*models.NewCustomRangeRevenueResponse, error)
	NewGetDefaultRevenueReport(ctx context.Context) (*models.NewDefaultRevenueResponse, error)
	GetRevenueReport(ctx context.Context, from, to time.Time) (*models.RevenueReportResponse, error)
	GetSalesReport(ctx context.Context, from, to time.Time) (*models.SalesReportResponse, error)
	GetCustomerReport(ctx context.Context, from, to time.Time) (*models.CustomerReportResponse, error)
	GetTableReport(ctx context.Context, from, to time.Time) (*models.TableReportResponse, error)
	GetStaffReport(ctx context.Context, from, to time.Time) (*models.ExtendedStaffReportResponse, error)
	GetFinancialSummary(ctx context.Context) (*models.FinancialSummaryResponse, error)
	GetRawMaterialReport(ctx context.Context) (*models.RawMaterialReportResponse, error)
}

type reportRepo struct {
	pool *pgxpool.Pool
}

func NewReportRepo() ReportRepo {

	pool, err := database.GetPostgresPool()

	if err != nil {
		return nil
	}

	return &reportRepo{pool: pool}
}

// ─── Main Entry ───────────────────────────────────────────────────────────────
// revenue report analysis
func (r *reportRepo) GetRevenueReport(ctx context.Context, from, to time.Time) (*models.RevenueReportResponse, error) {
	overview, err := r.getRevenueOverview(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("revenue overview: %w", err)
	}

	dailyTrend, err := r.getDailyTrend(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("daily trend: %w", err)
	}

	weeklyTrend, err := r.getWeeklyTrend(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("weekly trend: %w", err)
	}

	monthlyTrend, err := r.getMonthlyTrend(ctx)
	if err != nil {
		return nil, fmt.Errorf("monthly trend: %w", err)
	}

	yearlyTrend, err := r.getYearlyTrend(ctx)
	if err != nil {
		return nil, fmt.Errorf("yearly trend: %w", err)
	}

	paymentMethods, err := r.getPaymentMethodBreakdown(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("payment methods: %w", err)
	}

	gateways, err := r.getGatewayBreakdown(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("gateways: %w", err)
	}

	discounts, err := r.getDiscountAnalysis(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("discounts: %w", err)
	}

	peakHours, err := r.getPeakHours(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("peak hours: %w", err)
	}

	peakDays, err := r.getPeakDays(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("peak days: %w", err)
	}

	return &models.RevenueReportResponse{
		Overview:       overview,
		DailyTrend:     dailyTrend,
		WeeklyTrend:    weeklyTrend,
		MonthlyTrend:   monthlyTrend,
		YearlyTrend:    yearlyTrend,
		PaymentMethods: paymentMethods,
		Gateways:       gateways,
		Discounts:      discounts,
		PeakHours:      peakHours,
		PeakDays:       peakDays,
	}, nil
}

// ─── Overview Card ────────────────────────────────────────────────────────────

func (r *reportRepo) getRevenueOverview(ctx context.Context, from, to time.Time) (models.RevenueOverviewCard, error) {
	query := `
		SELECT
			COALESCE(SUM(p.paid_amount + p.discount), 0) AS gross_revenue,
			COALESCE(SUM(p.paid_amount), 0)              AS net_revenue,
			COALESCE(SUM(p.discount), 0)                 AS total_discounts,
			COUNT(p.id)                                  AS total_orders,
			COALESCE(AVG(p.paid_amount), 0)              AS avg_order_value
		FROM payments p
		WHERE p.created_at BETWEEN $1 AND $2
	`

	var gross, net, discounts, avg float64
	var totalOrders int
	if err := r.pool.QueryRow(ctx, query, from, to).Scan(
		&gross, &net, &discounts, &totalOrders, &avg,
	); err != nil {
		return models.RevenueOverviewCard{}, fmt.Errorf("failed to scan revenue overview: %w", err)
	}

	// Growth vs previous same-length window
	diff := to.Sub(from)
	var prevNet float64
	if err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(paid_amount), 0)
		FROM payments
		WHERE created_at BETWEEN $1 AND $2
	`, from.Add(-diff), from).Scan(&prevNet); err != nil {
		return models.RevenueOverviewCard{}, fmt.Errorf("failed to scan previous period: %w", err)
	}

	var growth float64
	if prevNet > 0 {
		growth = ((net - prevNet) / prevNet) * 100
	}

	return models.RevenueOverviewCard{
		GrossRevenue:      gross,
		NetRevenue:        net,
		TotalDiscounts:    discounts,
		TotalOrders:       totalOrders,
		AverageOrderValue: avg,
		GrowthPercent:     growth,
	}, nil
}

// ─── Daily Trend ──────────────────────────────────────────────────────────────

func (r *reportRepo) getDailyTrend(ctx context.Context, from, to time.Time) ([]models.TrendPoint, error) {
	query := `
		SELECT COALESCE(
			json_agg(
				json_build_object('period', period, 'revenue', revenue, 'orders', orders)
				ORDER BY period ASC
			), '[]'::json
		)
		FROM (
			SELECT
				TO_CHAR(p.created_at AT TIME ZONE 'Asia/Kathmandu', 'YYYY-MM-DD') AS period,
				COALESCE(SUM(p.paid_amount), 0)                                    AS revenue,
				COUNT(p.id)                                                        AS orders
			FROM payments p
			WHERE p.created_at BETWEEN $1 AND $2
			GROUP BY TO_CHAR(p.created_at AT TIME ZONE 'Asia/Kathmandu', 'YYYY-MM-DD')
		) sub
	`
	return r.scanTrendPoints(ctx, query, from, to)
}

// ─── Weekly Trend ─────────────────────────────────────────────────────────────

func (r *reportRepo) getWeeklyTrend(ctx context.Context, from, to time.Time) ([]models.TrendPoint, error) {
	query := `
		SELECT COALESCE(
			json_agg(
				json_build_object('period', period, 'revenue', revenue, 'orders', orders)
				ORDER BY period ASC
			), '[]'::json
		)
		FROM (
			SELECT
				TO_CHAR(DATE_TRUNC('week', p.created_at AT TIME ZONE 'Asia/Kathmandu'), 'IYYY-"W"IW') AS period,
				COALESCE(SUM(p.paid_amount), 0)                                                        AS revenue,
				COUNT(p.id)                                                                            AS orders
			FROM payments p
			WHERE p.created_at BETWEEN $1 AND $2
			GROUP BY DATE_TRUNC('week', p.created_at AT TIME ZONE 'Asia/Kathmandu')
		) sub
	`
	return r.scanTrendPoints(ctx, query, from, to)
}

// ─── Monthly Trend ────────────────────────────────────────────────────────────

func (r *reportRepo) getMonthlyTrend(ctx context.Context) ([]models.TrendPoint, error) {
	query := `
		SELECT COALESCE(
			json_agg(
				json_build_object('period', period, 'revenue', revenue, 'orders', orders)
				ORDER BY period ASC
			), '[]'::json
		)
		FROM (
			SELECT
				TO_CHAR(DATE_TRUNC('month', p.created_at AT TIME ZONE 'Asia/Kathmandu'), 'YYYY-MM') AS period,
				COALESCE(SUM(p.paid_amount), 0)                                                      AS revenue,
				COUNT(p.id)                                                                          AS orders
			FROM payments p
			WHERE p.created_at >= NOW() - INTERVAL '24 months'
			GROUP BY DATE_TRUNC('month', p.created_at AT TIME ZONE 'Asia/Kathmandu')
		) sub
	`

	var resultJSON []byte
	if err := r.pool.QueryRow(ctx, query).Scan(&resultJSON); err != nil {
		return nil, fmt.Errorf("failed to query monthly trend: %w", err)
	}
	var result []models.TrendPoint
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal monthly trend: %w", err)
	}
	return result, nil
}

// ─── Yearly Trend ─────────────────────────────────────────────────────────────

func (r *reportRepo) getYearlyTrend(ctx context.Context) ([]models.TrendPoint, error) {
	query := `
		SELECT COALESCE(
			json_agg(
				json_build_object('period', period, 'revenue', revenue, 'orders', orders)
				ORDER BY period ASC
			), '[]'::json
		)
		FROM (
			SELECT
				TO_CHAR(DATE_TRUNC('year', p.created_at AT TIME ZONE 'Asia/Kathmandu'), 'YYYY') AS period,
				COALESCE(SUM(p.paid_amount), 0)                                                  AS revenue,
				COUNT(p.id)                                                                      AS orders
			FROM payments p
			GROUP BY DATE_TRUNC('year', p.created_at AT TIME ZONE 'Asia/Kathmandu')
		) sub
	`

	var resultJSON []byte
	if err := r.pool.QueryRow(ctx, query).Scan(&resultJSON); err != nil {
		return nil, fmt.Errorf("failed to query yearly trend: %w", err)
	}
	var result []models.TrendPoint
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal yearly trend: %w", err)
	}
	return result, nil
}

// ─── Payment Method Breakdown ─────────────────────────────────────────────────

func (r *reportRepo) getPaymentMethodBreakdown(ctx context.Context, from, to time.Time) ([]models.PaymentMethodBreakdown, error) {
	query := `
		SELECT COALESCE(
			json_agg(
				json_build_object(
					'method',  method,
					'revenue', revenue,
					'orders',  orders,
					'percent', ROUND((revenue / NULLIF(total_rev, 0) * 100)::numeric, 2)
				)
			), '[]'::json
		)
		FROM (
			SELECT
				payment_method::text                  AS method,
				COALESCE(SUM(paid_amount), 0)         AS revenue,
				COUNT(id)                             AS orders,
				SUM(SUM(paid_amount)) OVER ()         AS total_rev
			FROM payments
			WHERE created_at BETWEEN $1 AND $2
			GROUP BY payment_method
		) sub
	`

	var resultJSON []byte
	if err := r.pool.QueryRow(ctx, query, from, to).Scan(&resultJSON); err != nil {
		return nil, fmt.Errorf("failed to query payment method breakdown: %w", err)
	}
	var result []models.PaymentMethodBreakdown
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payment method breakdown: %w", err)
	}
	return result, nil
}

// ─── Gateway Breakdown ────────────────────────────────────────────────────────

func (r *reportRepo) getGatewayBreakdown(ctx context.Context, from, to time.Time) ([]models.GatewayBreakdown, error) {
	query := `
		SELECT COALESCE(
			json_agg(
				json_build_object(
					'gateway', gateway,
					'revenue', revenue,
					'orders',  orders,
					'percent', ROUND((revenue / NULLIF(total_rev, 0) * 100)::numeric, 2)
				)
			), '[]'::json
		)
		FROM (
			SELECT
				online_gateway::text                  AS gateway,
				COALESCE(SUM(paid_amount), 0)         AS revenue,
				COUNT(id)                             AS orders,
				SUM(SUM(paid_amount)) OVER ()         AS total_rev
			FROM payments
			WHERE created_at BETWEEN $1 AND $2
			  AND payment_method = 'online'
			  AND online_gateway IS NOT NULL
			GROUP BY online_gateway
		) sub
	`

	var resultJSON []byte
	if err := r.pool.QueryRow(ctx, query, from, to).Scan(&resultJSON); err != nil {
		return nil, fmt.Errorf("failed to query gateway breakdown: %w", err)
	}
	var result []models.GatewayBreakdown
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal gateway breakdown: %w", err)
	}
	return result, nil
}

// ─── Discount Analysis ────────────────────────────────────────────────────────

func (r *reportRepo) getDiscountAnalysis(ctx context.Context, from, to time.Time) (models.DiscountAnalysis, error) {
	query := `
		SELECT
			COALESCE(SUM(discount), 0)                                                            AS total_discounts,
			COALESCE(SUM(paid_amount + discount), 0)                                             AS gross_revenue,
			COALESCE(SUM(paid_amount), 0)                                                        AS net_revenue,
			COALESCE(ROUND((SUM(discount) / NULLIF(SUM(paid_amount + discount), 0) * 100)::numeric, 2), 0) AS discount_rate,
			COUNT(*) FILTER (WHERE discount > 0)                                                 AS orders_with_discount,
			COUNT(*)                                                                             AS total_orders
		FROM payments
		WHERE created_at BETWEEN $1 AND $2
	`

	var d models.DiscountAnalysis
	if err := r.pool.QueryRow(ctx, query, from, to).Scan(
		&d.TotalDiscountsGiven,
		&d.GrossRevenue,
		&d.NetRevenue,
		&d.DiscountRatePercent,
		&d.OrdersWithDiscount,
		&d.TotalOrders,
	); err != nil {
		return models.DiscountAnalysis{}, fmt.Errorf("failed to scan discount analysis: %w", err)
	}
	return d, nil
}

// ─── Peak Hours ───────────────────────────────────────────────────────────────

func (r *reportRepo) getPeakHours(ctx context.Context, from, to time.Time) ([]models.PeakHourPoint, error) {
	query := `
		SELECT COALESCE(
			json_agg(
				json_build_object('hour', hour, 'revenue', revenue, 'orders', orders)
				ORDER BY hour ASC
			), '[]'::json
		)
		FROM (
			SELECT
				EXTRACT(HOUR FROM created_at AT TIME ZONE 'Asia/Kathmandu')::int AS hour,
				COALESCE(SUM(paid_amount), 0)                                    AS revenue,
				COUNT(id)                                                        AS orders
			FROM payments
			WHERE created_at BETWEEN $1 AND $2
			GROUP BY EXTRACT(HOUR FROM created_at AT TIME ZONE 'Asia/Kathmandu')
		) sub
	`

	var resultJSON []byte
	if err := r.pool.QueryRow(ctx, query, from, to).Scan(&resultJSON); err != nil {
		return nil, fmt.Errorf("failed to query peak hours: %w", err)
	}
	var result []models.PeakHourPoint
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal peak hours: %w", err)
	}
	return result, nil
}

// ─── Peak Days ────────────────────────────────────────────────────────────────

func (r *reportRepo) getPeakDays(ctx context.Context, from, to time.Time) ([]models.PeakDayPoint, error) {
	query := `
		SELECT COALESCE(
			json_agg(
				json_build_object('day_of_week', day_of_week, 'revenue', revenue, 'orders', orders)
				ORDER BY dow_num ASC
			), '[]'::json
		)
		FROM (
			SELECT
				TRIM(TO_CHAR(created_at AT TIME ZONE 'Asia/Kathmandu', 'Day'))   AS day_of_week,
				EXTRACT(DOW FROM created_at AT TIME ZONE 'Asia/Kathmandu')::int  AS dow_num,
				COALESCE(SUM(paid_amount), 0)                                    AS revenue,
				COUNT(id)                                                        AS orders
			FROM payments
			WHERE created_at BETWEEN $1 AND $2
			GROUP BY
				TRIM(TO_CHAR(created_at AT TIME ZONE 'Asia/Kathmandu', 'Day')),
				EXTRACT(DOW FROM created_at AT TIME ZONE 'Asia/Kathmandu')
		) sub
	`

	var resultJSON []byte
	if err := r.pool.QueryRow(ctx, query, from, to).Scan(&resultJSON); err != nil {
		return nil, fmt.Errorf("failed to query peak days: %w", err)
	}
	var result []models.PeakDayPoint
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal peak days: %w", err)
	}
	return result, nil
}

// ─── shared scan helper ───────────────────────────────────────────────────────

func (r *reportRepo) scanTrendPoints(ctx context.Context, query string, args ...interface{}) ([]models.TrendPoint, error) {
	var resultJSON []byte
	if err := r.pool.QueryRow(ctx, query, args...).Scan(&resultJSON); err != nil {
		return nil, fmt.Errorf("failed to query trend: %w", err)
	}
	var result []models.TrendPoint
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal trend: %w", err)
	}
	return result, nil
}

//sales analysis

// ─── Main Entry ───────────────────────────────────────────────────────────────

func (r *reportRepo) GetSalesReport(ctx context.Context, from, to time.Time) (*models.SalesReportResponse, error) {
	overview, err := r.getSalesOverview(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("sales overview: %w", err)
	}

	bestByQty, err := r.getBestSellingByQty(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("best by qty: %w", err)
	}

	bestByRevenue, err := r.getBestSellingByRevenue(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("best by revenue: %w", err)
	}

	bestCategories, err := r.getBestSellingCategories(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("best categories: %w", err)
	}

	slowestMoving, err := r.getSlowestMovingItems(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("slowest moving: %w", err)
	}

	frequentPairs, err := r.getFrequentPairs(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("frequent pairs: %w", err)
	}

	dailyTrend, err := r.getSalesDailyTrend(ctx)
	if err != nil {
		return nil, fmt.Errorf("daily trend: %w", err)
	}

	// handle err
	weeklyTrend, err := r.getSalesWeeklyTrend(ctx)
	if err != nil {
		return nil, fmt.Errorf("weekly trend: %w", err)
	}
	// handle err
	monthlyTrend, err := r.getSalesMonthlyTrend(ctx)
	if err != nil {
		return nil, fmt.Errorf("weekly trend: %w", err)
	}
	// handle err
	yearlyTrend, err := r.getSalesYearlyTrend(ctx)
	if err != nil {
		return nil, fmt.Errorf("weekly trend: %w", err)
	}
	// handle err

	return &models.SalesReportResponse{
		Overview:       overview,
		BestByQty:      bestByQty,
		BestByRevenue:  bestByRevenue,
		BestCategories: bestCategories,
		SlowestMoving:  slowestMoving,
		FrequentPairs:  frequentPairs,
		DailyTrend:     dailyTrend,
		WeeklyTrend:    weeklyTrend,
		MonthlyTrend:   monthlyTrend,
		YearlyTrend:    yearlyTrend,
	}, nil
}

// ─── Overview Card ────────────────────────────────────────────────────────────

func (r *reportRepo) getSalesOverview(ctx context.Context, from, to time.Time) (models.SalesOverviewCard, error) {
	query := `
		SELECT
			COALESCE(SUM(oi.quantity), 0)     AS total_items_sold,
			COUNT(DISTINCT oi.menu_item_id)   AS unique_menu_items,
			COUNT(DISTINCT oi.order_id)       AS total_orders_placed
		FROM order_items oi
		JOIN orders o ON o.id = oi.order_id
		WHERE o.status = 'completed'
		  AND oi.created_at BETWEEN $1 AND $2
	`

	var s models.SalesOverviewCard
	if err := r.pool.QueryRow(ctx, query, from, to).Scan(
		&s.TotalItemsSold,
		&s.UniqueMenuItems,
		&s.TotalOrdersPlaced,
	); err != nil {
		return models.SalesOverviewCard{}, fmt.Errorf("failed to scan sales overview: %w", err)
	}

	// top selling item name
	err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(mi.name, '')
		FROM order_items oi
		JOIN orders o  ON o.id  = oi.order_id
		JOIN menu_items mi ON mi.id = oi.menu_item_id
		WHERE o.status = 'completed'
		  AND oi.created_at BETWEEN $1 AND $2
		GROUP BY mi.name
		ORDER BY SUM(oi.quantity) DESC
		LIMIT 1
	`, from, to).Scan(&s.TopSellingItem)
	if err != nil {
		s.TopSellingItem = ""
	}

	// top category name
	err = r.pool.QueryRow(ctx, `
		SELECT COALESCE(c.name, '')
		FROM order_items oi
		JOIN orders o      ON o.id  = oi.order_id
		JOIN menu_items mi ON mi.id = oi.menu_item_id
		JOIN categories c  ON c.id  = mi.category_id
		WHERE o.status = 'completed'
		  AND oi.created_at BETWEEN $1 AND $2
		GROUP BY c.name
		ORDER BY SUM(oi.quantity) DESC
		LIMIT 1
	`, from, to).Scan(&s.TopCategory)
	if err != nil {
		s.TopCategory = ""
	}

	return s, nil
}

// ─── Best Selling by Qty ──────────────────────────────────────────────────────

func (r *reportRepo) getBestSellingByQty(ctx context.Context, from, to time.Time) ([]models.BestSellingItem, error) {
	query := `
		SELECT COALESCE(
			json_agg(
				json_build_object(
					'menu_item_id',   menu_item_id,
					'menu_name',      menu_name,
					'category_name',  category_name,
					'total_qty',      total_qty,
					'total_revenue',  total_revenue,
					'order_count',    order_count
				) ORDER BY total_qty DESC
			), '[]'::json
		)
		FROM (
			SELECT
				oi.menu_item_id::text               AS menu_item_id,
				COALESCE(mi.name, '')               AS menu_name,
				COALESCE(c.name, '')                AS category_name,
				COALESCE(SUM(oi.quantity), 0)       AS total_qty,
				COALESCE(SUM(oi.quantity * oi.price), 0) AS total_revenue,
				COUNT(DISTINCT oi.order_id)         AS order_count
			FROM order_items oi
			JOIN orders o      ON o.id  = oi.order_id
			JOIN menu_items mi ON mi.id = oi.menu_item_id
			JOIN categories c  ON c.id  = mi.category_id
			WHERE o.status = 'completed'
			  AND oi.created_at BETWEEN $1 AND $2
			GROUP BY oi.menu_item_id, mi.name, c.name
			ORDER BY total_qty DESC
			LIMIT 20
		) sub
	`

	var resultJSON []byte
	if err := r.pool.QueryRow(ctx, query, from, to).Scan(&resultJSON); err != nil {
		return nil, fmt.Errorf("failed to query best by qty: %w", err)
	}
	var result []models.BestSellingItem
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal best by qty: %w", err)
	}
	return result, nil
}

// ─── Best Selling by Revenue ──────────────────────────────────────────────────

func (r *reportRepo) getBestSellingByRevenue(ctx context.Context, from, to time.Time) ([]models.BestSellingItem, error) {
	query := `
		SELECT COALESCE(
			json_agg(
				json_build_object(
					'menu_item_id',   menu_item_id,
					'menu_name',      menu_name,
					'category_name',  category_name,
					'total_qty',      total_qty,
					'total_revenue',  total_revenue,
					'order_count',    order_count
				) ORDER BY total_revenue DESC
			), '[]'::json
		)
		FROM (
			SELECT
				oi.menu_item_id::text                    AS menu_item_id,
				COALESCE(mi.name, '')                    AS menu_name,
				COALESCE(c.name, '')                     AS category_name,
				COALESCE(SUM(oi.quantity), 0)            AS total_qty,
				COALESCE(SUM(oi.quantity * oi.price), 0) AS total_revenue,
				COUNT(DISTINCT oi.order_id)              AS order_count
			FROM order_items oi
			JOIN orders o      ON o.id  = oi.order_id
			JOIN menu_items mi ON mi.id = oi.menu_item_id
			JOIN categories c  ON c.id  = mi.category_id
			WHERE o.status = 'completed'
			  AND oi.created_at BETWEEN $1 AND $2
			GROUP BY oi.menu_item_id, mi.name, c.name
			ORDER BY total_revenue DESC
			LIMIT 20
		) sub
	`

	var resultJSON []byte
	if err := r.pool.QueryRow(ctx, query, from, to).Scan(&resultJSON); err != nil {
		return nil, fmt.Errorf("failed to query best by revenue: %w", err)
	}
	var result []models.BestSellingItem
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal best by revenue: %w", err)
	}
	return result, nil
}

// ─── Best Selling Categories ──────────────────────────────────────────────────

func (r *reportRepo) getBestSellingCategories(ctx context.Context, from, to time.Time) ([]models.BestSellingCategory, error) {
	query := `
		SELECT COALESCE(
			json_agg(
				json_build_object(
					'category_id',    category_id,
					'category_name',  category_name,
					'total_qty',      total_qty,
					'total_revenue',  total_revenue,
					'item_count',     item_count
				) ORDER BY total_revenue DESC
			), '[]'::json
		)
		FROM (
			SELECT
				c.id::text                               AS category_id,
				COALESCE(c.name, '')                     AS category_name,
				COALESCE(SUM(oi.quantity), 0)            AS total_qty,
				COALESCE(SUM(oi.quantity * oi.price), 0) AS total_revenue,
				COUNT(DISTINCT oi.menu_item_id)          AS item_count
			FROM order_items oi
			JOIN orders o      ON o.id  = oi.order_id
			JOIN menu_items mi ON mi.id = oi.menu_item_id
			JOIN categories c  ON c.id  = mi.category_id
			WHERE o.status = 'completed'
			  AND oi.created_at BETWEEN $1 AND $2
			GROUP BY c.id, c.name
		) sub
	`

	var resultJSON []byte
	if err := r.pool.QueryRow(ctx, query, from, to).Scan(&resultJSON); err != nil {
		return nil, fmt.Errorf("failed to query best categories: %w", err)
	}
	var result []models.BestSellingCategory
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal best categories: %w", err)
	}
	return result, nil
}

// ─── Slowest Moving Items ─────────────────────────────────────────────────────

func (r *reportRepo) getSlowestMovingItems(ctx context.Context, from, to time.Time) ([]models.SlowestMovingItem, error) {
	query := `
		SELECT COALESCE(
			json_agg(
				json_build_object(
					'menu_item_id',  menu_item_id,
					'menu_name',     menu_name,
					'category_name', category_name,
					'total_qty',     total_qty,
					'total_revenue', total_revenue,
					'last_ordered',  last_ordered
				) ORDER BY total_qty ASC
			), '[]'::json
		)
		FROM (
			SELECT
				mi.id::text                              AS menu_item_id,
				COALESCE(mi.name, '')                    AS menu_name,
				COALESCE(c.name, '')                     AS category_name,
				COALESCE(SUM(oi.quantity), 0)            AS total_qty,
				COALESCE(SUM(oi.quantity * oi.price), 0) AS total_revenue,
				TO_CHAR(MAX(oi.created_at), 'YYYY-MM-DD HH24:MI') AS last_ordered
			FROM menu_items mi
			LEFT JOIN order_items oi ON oi.menu_item_id = mi.id
				AND oi.created_at BETWEEN $1 AND $2
			LEFT JOIN orders o ON o.id = oi.order_id AND o.status = 'completed'
			LEFT JOIN categories c ON c.id = mi.category_id
			WHERE mi.is_available = TRUE
			GROUP BY mi.id, mi.name, c.name
			ORDER BY total_qty ASC
			LIMIT 20
		) sub
	`

	var resultJSON []byte
	if err := r.pool.QueryRow(ctx, query, from, to).Scan(&resultJSON); err != nil {
		return nil, fmt.Errorf("failed to query slowest items: %w", err)
	}
	var result []models.SlowestMovingItem
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal slowest items: %w", err)
	}
	return result, nil
}

// ─── Frequently Ordered Together (Basket Analysis) ───────────────────────────

func (r *reportRepo) getFrequentPairs(ctx context.Context, from, to time.Time) ([]models.FrequentPairItem, error) {
	query := `
		SELECT COALESCE(
			json_agg(
				json_build_object(
					'item_a',     item_a,
					'item_b',     item_b,
					'pair_count', pair_count
				) ORDER BY pair_count DESC
			), '[]'::json
		)
		FROM (
			SELECT
				COALESCE(mi_a.name, '') AS item_a,
				COALESCE(mi_b.name, '') AS item_b,
				COUNT(*)               AS pair_count
			FROM order_items a
			JOIN order_items b       ON a.order_id = b.order_id AND a.menu_item_id < b.menu_item_id
			JOIN menu_items mi_a     ON mi_a.id = a.menu_item_id
			JOIN menu_items mi_b     ON mi_b.id = b.menu_item_id
			JOIN orders o            ON o.id = a.order_id
			WHERE o.status = 'completed'
			  AND a.created_at BETWEEN $1 AND $2
			GROUP BY mi_a.name, mi_b.name
			ORDER BY pair_count DESC
			LIMIT 20
		) sub
	`

	var resultJSON []byte
	if err := r.pool.QueryRow(ctx, query, from, to).Scan(&resultJSON); err != nil {
		return nil, fmt.Errorf("failed to query frequent pairs: %w", err)
	}
	var result []models.FrequentPairItem
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal frequent pairs: %w", err)
	}
	return result, nil
}

func (r *reportRepo) GetStaffReport(ctx context.Context, from, to time.Time) (*models.ExtendedStaffReportResponse, error) {
	attendanceOverview, err := r.getAttendanceOverview(ctx)
	if err != nil {
		return nil, fmt.Errorf("attendance overview: %w", err)
	}

	dailySummary, err := r.getDailyAttendanceSummary(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("daily attendance summary: %w", err)
	}

	employeeAttendance, err := r.getEmployeeAttendance(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("employee attendance: %w", err)
	}

	leaveManagement, err := r.getLeaveManagement(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("leave management: %w", err)
	}

	payroll, err := r.getPayrollReport(ctx)
	if err != nil {
		return nil, fmt.Errorf("payroll: %w", err)
	}

	// NEW: Get most present employees
	mostPresent, err := r.getMostPresentEmployees(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("most present employees: %w", err)
	}

	// NEW: Get most absent employees
	mostAbsent, err := r.getMostAbsentEmployees(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("most absent employees: %w", err)
	}

	return &models.ExtendedStaffReportResponse{
		AttendanceOverview:   attendanceOverview,
		DailySummary:         dailySummary,
		EmployeeAttendance:   employeeAttendance,
		LeaveManagement:      leaveManagement,
		Payroll:              payroll,
		MostPresentEmployees: mostPresent,
		MostAbsentEmployees:  mostAbsent,
	}, nil
}

// ─── Most Present Employees (Top 10) ─────────────────────────────────────────

func (r *reportRepo) getMostPresentEmployees(ctx context.Context, from, to time.Time) ([]models.MostPresentEmployee, error) {
	query := `
		SELECT COALESCE(
			json_agg(
				json_build_object(
					'employee_id',      employee_id,
					'employee_name',    employee_name,
					'email',            email,
					'phone',            phone,
					'image',            image,
					'role',             role,
					'gender',           gender,
					'present_days',     present_days,
					'attendance_rate',  attendance_rate
				) ORDER BY present_days DESC
			), '[]'::json
		)
		FROM (
			SELECT
				u.id::text                              AS employee_id,
				COALESCE(u.name, '')                    AS employee_name,
				COALESCE(u.email, '')                   AS email,
				COALESCE(u.phone, '')                   AS phone,
				u.image                                 AS image,
				u.role::text                            AS role,
				u.gender::text                          AS gender,
				COUNT(*) FILTER (WHERE a.status = 'present' OR a.status = 'late') AS present_days,
				ROUND(
					COUNT(*) FILTER (WHERE a.status IN ('present', 'late'))::numeric
					/ NULLIF(COUNT(a.id), 0) * 100,
					2
				)                                       AS attendance_rate
			FROM users u
			LEFT JOIN attendance a ON a.employee_id = u.id
				AND a.work_date BETWEEN $1 AND $2
			WHERE u.role != 'customer'
			  AND u.is_active = TRUE
			GROUP BY u.id, u.name, u.email, u.phone, u.image, u.role, u.gender
			HAVING COUNT(a.id) > 0  -- Only employees with attendance records
			ORDER BY present_days DESC
			LIMIT 10
		) sub
	`

	var resultJSON []byte
	if err := r.pool.QueryRow(ctx, query, from, to).Scan(&resultJSON); err != nil {
		return nil, fmt.Errorf("failed to query most present employees: %w", err)
	}
	var result []models.MostPresentEmployee
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal most present employees: %w", err)
	}
	return result, nil
}

// ─── Most Absent Employees (Top 10) ──────────────────────────────────────────

func (r *reportRepo) getMostAbsentEmployees(ctx context.Context, from, to time.Time) ([]models.MostAbsentEmployee, error) {
	query := `
		SELECT COALESCE(
			json_agg(
				json_build_object(
					'employee_id',      employee_id,
					'employee_name',    employee_name,
					'email',            email,
					'phone',            phone,
					'image',            image,
					'role',             role,
					'gender',           gender,
					'absent_days',      absent_days,
					'attendance_rate',  attendance_rate
				) ORDER BY absent_days DESC
			), '[]'::json
		)
		FROM (
			SELECT
				u.id::text                              AS employee_id,
				COALESCE(u.name, '')                    AS employee_name,
				COALESCE(u.email, '')                   AS email,
				COALESCE(u.phone, '')                   AS phone,
				u.image                                 AS image,
				u.role::text                            AS role,
				u.gender::text                          AS gender,
				COUNT(*) FILTER (WHERE a.status = 'absent') AS absent_days,
				ROUND(
					COUNT(*) FILTER (WHERE a.status IN ('present', 'late'))::numeric
					/ NULLIF(COUNT(a.id), 0) * 100,
					2
				)                                       AS attendance_rate
			FROM users u
			LEFT JOIN attendance a ON a.employee_id = u.id
				AND a.work_date BETWEEN $1 AND $2
			WHERE u.role != 'customer'
			  AND u.is_active = TRUE
			GROUP BY u.id, u.name, u.email, u.phone, u.image, u.role, u.gender
			HAVING COUNT(a.id) > 0  -- Only employees with attendance records
			ORDER BY absent_days DESC
			LIMIT 10
		) sub
	`

	var resultJSON []byte
	if err := r.pool.QueryRow(ctx, query, from, to).Scan(&resultJSON); err != nil {
		return nil, fmt.Errorf("failed to query most absent employees: %w", err)
	}
	var result []models.MostAbsentEmployee
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal most absent employees: %w", err)
	}
	return result, nil
}

// ─── Attendance Overview Card ─────────────────────────────────────────────────

func (r *reportRepo) getAttendanceOverview(ctx context.Context) (models.AttendanceOverviewCard, error) {
	query := `
		SELECT
			COUNT(DISTINCT u.id)                                                      AS total_employees,
			COUNT(*) FILTER (WHERE a.status = 'present'  AND a.work_date = CURRENT_DATE) AS present_today,
			COUNT(*) FILTER (WHERE a.status = 'absent'   AND a.work_date = CURRENT_DATE) AS absent_today,
			COUNT(*) FILTER (WHERE a.status = 'late'     AND a.work_date = CURRENT_DATE) AS late_today,
			COUNT(*) FILTER (WHERE a.status = 'leave'    AND a.work_date = CURRENT_DATE) AS on_leave_today,
			COUNT(*) FILTER (WHERE a.need_review = TRUE  AND a.work_date = CURRENT_DATE) AS need_review
		FROM users u
		LEFT JOIN attendance a ON a.employee_id = u.id
		WHERE u.role != 'customer'
		  AND u.is_active = TRUE
	`

	var c models.AttendanceOverviewCard
	if err := r.pool.QueryRow(ctx, query).Scan(
		&c.TotalEmployees,
		&c.PresentToday,
		&c.AbsentToday,
		&c.LateToday,
		&c.OnLeaveToday,
		&c.NeedReviewCount,
	); err != nil {
		return models.AttendanceOverviewCard{}, fmt.Errorf("failed to scan attendance overview: %w", err)
	}

	if c.TotalEmployees > 0 {
		c.AttendanceRate = float64(c.PresentToday+c.LateToday) / float64(c.TotalEmployees) * 100
	}

	return c, nil
}

// ─── Daily Attendance Summary ─────────────────────────────────────────────────

func (r *reportRepo) getDailyAttendanceSummary(ctx context.Context, from, to time.Time) ([]models.DailyAttendanceSummary, error) {
	query := `
		SELECT COALESCE(
			json_agg(
				json_build_object(
					'work_date', work_date,
					'present',   present,
					'absent',    absent,
					'late',      late,
					'half_day',  half_day,
					'on_leave',  on_leave
				) ORDER BY work_date ASC
			), '[]'::json
		)
		FROM (
			SELECT
				work_date::text                                      AS work_date,
				COUNT(*) FILTER (WHERE status = 'present')          AS present,
				COUNT(*) FILTER (WHERE status = 'absent')           AS absent,
				COUNT(*) FILTER (WHERE status = 'late')             AS late,
				COUNT(*) FILTER (WHERE status = 'half_day')         AS half_day,
				COUNT(*) FILTER (WHERE status = 'leave')            AS on_leave
			FROM attendance
			WHERE work_date BETWEEN $1 AND $2
			GROUP BY work_date
		) sub
	`

	var resultJSON []byte
	if err := r.pool.QueryRow(ctx, query, from, to).Scan(&resultJSON); err != nil {
		return nil, fmt.Errorf("failed to query daily attendance: %w", err)
	}
	var result []models.DailyAttendanceSummary
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal daily attendance: %w", err)
	}
	return result, nil
}

// ─── Per-Employee Attendance ──────────────────────────────────────────────────

func (r *reportRepo) getEmployeeAttendance(ctx context.Context, from, to time.Time) ([]models.EmployeeAttendanceRow, error) {
	query := `
		SELECT COALESCE(
			json_agg(
				json_build_object(
					'employee_id',      employee_id,
					'employee_name',    employee_name,
					'role',             role,
					'present_days',     present_days,
					'absent_days',      absent_days,
					'late_days',        late_days,
					'attendance_rate',  attendance_rate,
					'avg_work_hours',   avg_work_hours,
					'need_review',      need_review
				) ORDER BY attendance_rate ASC
			), '[]'::json
		)
		FROM (
			SELECT
				u.id::text                                                              AS employee_id,
				COALESCE(u.name, '')                                                   AS employee_name,
				u.role::text                                                           AS role,
				COUNT(*) FILTER (WHERE a.status = 'present')                          AS present_days,
				COUNT(*) FILTER (WHERE a.status = 'absent')                           AS absent_days,
				COUNT(*) FILTER (WHERE a.status = 'late')                             AS late_days,
				ROUND(
					COUNT(*) FILTER (WHERE a.status IN ('present', 'late'))::numeric
					/ NULLIF(COUNT(a.id), 0) * 100,
					2
				)                                                                      AS attendance_rate,
				COALESCE(
					AVG(
						EXTRACT(EPOCH FROM (a.check_out_time - a.check_in_time)) / 3600
					) FILTER (WHERE a.check_in_time IS NOT NULL AND a.check_out_time IS NOT NULL),
					0
				)                                                                      AS avg_work_hours,
				BOOL_OR(a.need_review)                                                AS need_review
			FROM users u
			LEFT JOIN attendance a ON a.employee_id = u.id
				AND a.work_date BETWEEN $1 AND $2
			WHERE u.role != 'customer'
			  AND u.is_active = TRUE
			GROUP BY u.id, u.name, u.role
		) sub
	`

	var resultJSON []byte
	if err := r.pool.QueryRow(ctx, query, from, to).Scan(&resultJSON); err != nil {
		return nil, fmt.Errorf("failed to query employee attendance: %w", err)
	}
	var result []models.EmployeeAttendanceRow
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal employee attendance: %w", err)
	}
	return result, nil
}

// ─── Leave Management ─────────────────────────────────────────────────────────

func (r *reportRepo) getLeaveManagement(ctx context.Context, from, to time.Time) (models.LeaveManagementReport, error) {
	var l models.LeaveManagementReport

	if err := r.pool.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE status = 'pending')  AS pending,
			COUNT(*) FILTER (WHERE status = 'approved') AS approved,
			COUNT(*) FILTER (WHERE status = 'rejected') AS rejected
		FROM attendance_leave
		WHERE created_at BETWEEN $1 AND $2
	`, from, to).Scan(
		&l.PendingCount,
		&l.ApprovedCount,
		&l.RejectedCount,
	); err != nil {
		return models.LeaveManagementReport{}, fmt.Errorf("failed to scan leave totals: %w", err)
	}

	total := l.ApprovedCount + l.RejectedCount
	if total > 0 {
		l.ApprovalRate = float64(l.ApprovedCount) / float64(total) * 100
	}

	// Top employees by leave requests
	tleQuery := `
		SELECT COALESCE(
			json_agg(
				json_build_object(
					'employee_id',    employee_id,
					'employee_name',  employee_name,
					'leave_count',    leave_count
				) ORDER BY leave_count DESC
			), '[]'::json
		)
		FROM (
			SELECT
				al.employee_id::text          AS employee_id,
				COALESCE(u.name, '')          AS employee_name,
				COUNT(al.id)                  AS leave_count
			FROM attendance_leave al
			JOIN users u ON u.id = al.employee_id
			WHERE al.created_at BETWEEN $1 AND $2
			GROUP BY al.employee_id, u.name
			ORDER BY leave_count DESC
			LIMIT 10
		) sub
	`

	var resultJSON []byte
	if err := r.pool.QueryRow(ctx, tleQuery, from, to).Scan(&resultJSON); err != nil {
		return models.LeaveManagementReport{}, fmt.Errorf("failed to query top leave employees: %w", err)
	}
	if err := json.Unmarshal(resultJSON, &l.TopLeaveEmployees); err != nil {
		return models.LeaveManagementReport{}, fmt.Errorf("failed to unmarshal top leave employees: %w", err)
	}

	return l, nil
}

// ─── Payroll Report ───────────────────────────────────────────────────────────

func (r *reportRepo) getPayrollReport(ctx context.Context) (models.PayrollReport, error) {
	var p models.PayrollReport

	if err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(salary), 0)
		FROM users
		WHERE role != 'customer'
		  AND is_active = TRUE
	`).Scan(&p.TotalMonthlySalary); err != nil {
		return models.PayrollReport{}, fmt.Errorf("failed to scan total salary: %w", err)
	}

	// Salary by role
	roleQuery := `
		SELECT COALESCE(
			json_agg(
				json_build_object(
					'role',            role,
					'employee_count',  employee_count,
					'total_salary',    total_salary,
					'average_salary',  average_salary,
					'percent',         ROUND((total_salary / NULLIF(grand_total, 0) * 100)::numeric, 2)
				) ORDER BY total_salary DESC
			), '[]'::json
		)
		FROM (
			SELECT
				role::text                      AS role,
				COUNT(id)                       AS employee_count,
				COALESCE(SUM(salary), 0)        AS total_salary,
				COALESCE(AVG(salary), 0)        AS average_salary,
				SUM(SUM(salary)) OVER ()        AS grand_total
			FROM users
			WHERE role != 'customer'
			  AND is_active = TRUE
			GROUP BY role
		) sub
	`

	var resultJSON []byte
	if err := r.pool.QueryRow(ctx, roleQuery).Scan(&resultJSON); err != nil {
		return models.PayrollReport{}, fmt.Errorf("failed to query salary by role: %w", err)
	}
	if err := json.Unmarshal(resultJSON, &p.SalaryByRole); err != nil {
		return models.PayrollReport{}, fmt.Errorf("failed to unmarshal salary by role: %w", err)
	}

	// Labor cost % vs revenue this month
	var monthRevenue float64
	if err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(paid_amount), 0)
		FROM payments
		WHERE created_at >= DATE_TRUNC('month', NOW())
	`).Scan(&monthRevenue); err == nil && monthRevenue > 0 {
		p.LaborCostPercent = (p.TotalMonthlySalary / monthRevenue) * 100
	}

	return p, nil
}

// ─── Interface ────────────────────────────────────────────────────────────────

// ─── Main Entry ───────────────────────────────────────────────────────────────

func (r *reportRepo) GetTableReport(ctx context.Context, from, to time.Time) (*models.TableReportResponse, error) {
	overview, err := r.getTableOverview(ctx)
	if err != nil {
		return nil, fmt.Errorf("table overview: %w", err)
	}

	utilization, err := r.getTableUtilization(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("table utilization: %w", err)
	}

	peakHours, err := r.getPeakOccupancyHours(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("peak occupancy hours: %w", err)
	}

	turnover, err := r.getTableTurnover(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("table turnover: %w", err)
	}

	longestIdle, err := r.getLongestIdleTables(ctx)
	if err != nil {
		return nil, fmt.Errorf("longest idle tables: %w", err)
	}

	return &models.TableReportResponse{
		Overview:    overview,
		Utilization: utilization,
		PeakHours:   peakHours,
		Turnover:    turnover,
		LongestIdle: longestIdle,
	}, nil
}

// ─── Overview Card ────────────────────────────────────────────────────────────

func (r *reportRepo) getTableOverview(ctx context.Context) (models.TableOverviewCard, error) {
	query := `
		SELECT
			COUNT(*)                                          AS total_tables,
			COUNT(*) FILTER (WHERE status = 'occupied')      AS currently_occupied,
			ROUND(
				COUNT(*) FILTER (WHERE status = 'occupied')::numeric
				/ NULLIF(COUNT(*), 0) * 100,
				2
			)                                                AS utilization_percent
		FROM table_status
	`

	var c models.TableOverviewCard
	if err := r.pool.QueryRow(ctx, query).Scan(
		&c.TotalTables,
		&c.CurrentlyOccupied,
		&c.UtilizationPercent,
	); err != nil {
		return models.TableOverviewCard{}, fmt.Errorf("failed to scan table overview: %w", err)
	}

	// avg session duration + total sessions today
	err := r.pool.QueryRow(ctx, `
		SELECT
			COALESCE(AVG(EXTRACT(EPOCH FROM (COALESCE(close_time, NOW()) - open_time)) / 60), 0) AS avg_minutes,
			COUNT(*) FILTER (WHERE open_time::date = CURRENT_DATE)                              AS sessions_today
		FROM table_session
		WHERE open_time IS NOT NULL
	`).Scan(&c.AvgSessionMinutes, &c.TotalSessionsToday)
	if err != nil {
		return models.TableOverviewCard{}, fmt.Errorf("failed to scan session stats: %w", err)
	}

	return c, nil
}

// ─── Table Utilization ────────────────────────────────────────────────────────

func (r *reportRepo) getTableUtilization(ctx context.Context, from, to time.Time) ([]models.TableUtilizationRow, error) {
	query := `
		SELECT COALESCE(
			json_agg(
				json_build_object(
					'table_number',         table_number,
					'total_sessions',       total_sessions,
					'avg_session_minutes',  avg_session_minutes,
					'utilization_percent',  utilization_percent,
					'total_revenue',        total_revenue
				) ORDER BY total_revenue DESC
			), '[]'::json
		)
		FROM (
			SELECT
				ts.table_number,
				COUNT(ts.id)                                                   AS total_sessions,
				COALESCE(
					AVG(EXTRACT(EPOCH FROM (COALESCE(ts.close_time, NOW()) - ts.open_time)) / 60),
					0
				)                                                              AS avg_session_minutes,
				ROUND(
					COUNT(ts.id)::numeric
					/ NULLIF(
						EXTRACT(EPOCH FROM ($2::timestamptz - $1::timestamptz)) / 3600 / 24,
						0
					) * 100,
					2
				)                                                              AS utilization_percent,
				COALESCE(SUM(p.paid_amount), 0)                               AS total_revenue
			FROM table_session ts
			LEFT JOIN orders o  ON o.table_session_id = ts.id
			LEFT JOIN payments p ON p.order_id = o.id
			WHERE ts.open_time BETWEEN $1 AND $2
			GROUP BY ts.table_number
		) sub
	`

	var resultJSON []byte
	if err := r.pool.QueryRow(ctx, query, from, to).Scan(&resultJSON); err != nil {
		return nil, fmt.Errorf("failed to query table utilization: %w", err)
	}
	var result []models.TableUtilizationRow
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal table utilization: %w", err)
	}
	return result, nil
}

// ─── Peak Occupancy Hours ─────────────────────────────────────────────────────

func (r *reportRepo) getPeakOccupancyHours(ctx context.Context, from, to time.Time) ([]models.PeakOccupancyHour, error) {
	query := `
		SELECT COALESCE(
			json_agg(
				json_build_object(
					'hour',           hour,
					'avg_occupied_tables', avg_occupied,
					'total_sessions', total_sessions
				) ORDER BY hour ASC
			), '[]'::json
		)
		FROM (
			SELECT
				EXTRACT(HOUR FROM open_time AT TIME ZONE 'Asia/Kathmandu')::int AS hour,
				ROUND(AVG(table_number)::numeric, 2)                            AS avg_occupied,
				COUNT(id)                                                       AS total_sessions
			FROM table_session
			WHERE open_time BETWEEN $1 AND $2
			GROUP BY EXTRACT(HOUR FROM open_time AT TIME ZONE 'Asia/Kathmandu')
		) sub
	`

	var resultJSON []byte
	if err := r.pool.QueryRow(ctx, query, from, to).Scan(&resultJSON); err != nil {
		return nil, fmt.Errorf("failed to query peak occupancy hours: %w", err)
	}
	var result []models.PeakOccupancyHour
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal peak occupancy hours: %w", err)
	}
	return result, nil
}

// ─── Table Turnover ───────────────────────────────────────────────────────────

func (r *reportRepo) getTableTurnover(ctx context.Context, from, to time.Time) ([]models.TableTurnoverRow, error) {
	query := `
		SELECT COALESCE(
			json_agg(
				json_build_object(
					'table_number',     table_number,
					'sessions_per_day', sessions_per_day,
					'total_sessions',   total_sessions
				) ORDER BY sessions_per_day DESC
			), '[]'::json
		)
		FROM (
			SELECT
				table_number,
				COUNT(id) AS total_sessions,
				ROUND(
					COUNT(id)::numeric
					/ NULLIF(
						EXTRACT(EPOCH FROM ($2::timestamptz - $1::timestamptz)) / 86400,
						0
					),
					2
				) AS sessions_per_day
			FROM table_session
			WHERE open_time BETWEEN $1 AND $2
			GROUP BY table_number
		) sub
	`

	var resultJSON []byte
	if err := r.pool.QueryRow(ctx, query, from, to).Scan(&resultJSON); err != nil {
		return nil, fmt.Errorf("failed to query table turnover: %w", err)
	}
	var result []models.TableTurnoverRow
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal table turnover: %w", err)
	}
	return result, nil
}

// ─── Longest Idle Tables ──────────────────────────────────────────────────────

func (r *reportRepo) getLongestIdleTables(ctx context.Context) ([]models.LongestIdleTable, error) {
	query := `
		SELECT COALESCE(
			json_agg(
				json_build_object(
					'table_number',   table_number,
					'last_closed_at', last_closed_at,
					'idle_hours',     idle_hours,
					'current_status', current_status
				) ORDER BY idle_hours DESC
			), '[]'::json
		)
		FROM (
			SELECT
				ts_stat.table_number,
				TO_CHAR(last_session.close_time, 'YYYY-MM-DD HH24:MI') AS last_closed_at,
				ROUND(
					EXTRACT(EPOCH FROM (NOW() - COALESCE(last_session.close_time, NOW() - INTERVAL '999 hours'))) / 3600,
					2
				)                                                        AS idle_hours,
				ts_stat.status::text                                     AS current_status
			FROM table_status ts_stat
			LEFT JOIN LATERAL (
				SELECT close_time
				FROM table_session
				WHERE table_number = ts_stat.table_number
				  AND close_time IS NOT NULL
				ORDER BY close_time DESC
				LIMIT 1
			) last_session ON TRUE
			WHERE ts_stat.status = 'empty'
		) sub
	`

	var resultJSON []byte
	if err := r.pool.QueryRow(ctx, query).Scan(&resultJSON); err != nil {
		return nil, fmt.Errorf("failed to query longest idle tables: %w", err)
	}
	var result []models.LongestIdleTable
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal longest idle tables: %w", err)
	}
	return result, nil
}

// ─── Main Entry ───────────────────────────────────────────────────────────────

func (r *reportRepo) GetCustomerReport(ctx context.Context, from, to time.Time) (*models.CustomerReportResponse, error) {
	overview, err := r.getCustomerOverview(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("customer overview: %w", err)
	}

	topCustomers, err := r.getTopCustomers(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("top customers: %w", err)
	}

	visitFrequency, err := r.getVisitFrequency(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("visit frequency: %w", err)
	}

	tokenEconomy, err := r.getTokenEconomy(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("token economy: %w", err)
	}

	streakLeaderboard, err := r.getStreakLeaderboard(ctx)
	if err != nil {
		return nil, fmt.Errorf("streak leaderboard: %w", err)
	}

	return &models.CustomerReportResponse{
		Overview:          overview,
		TopCustomers:      topCustomers,
		VisitFrequency:    visitFrequency,
		TokenEconomy:      tokenEconomy,
		StreakLeaderboard: streakLeaderboard,
	}, nil
}

// ─── Overview Card ────────────────────────────────────────────────────────────

func (r *reportRepo) getCustomerOverview(ctx context.Context, from, to time.Time) (models.CustomerOverviewCard, error) {
	// unique customers, new vs returning
	query := `
		WITH all_customers AS (
			SELECT
				customer_phone,
				MIN(created_at) AS first_order
			FROM orders
			WHERE customer_phone IS NOT NULL
			GROUP BY customer_phone
		),
		in_range AS (
			SELECT DISTINCT customer_phone
			FROM orders
			WHERE customer_phone IS NOT NULL
			  AND created_at BETWEEN $1 AND $2
		)
		SELECT
			COUNT(ir.customer_phone)                                              AS total_unique,
			COUNT(*) FILTER (WHERE ac.first_order BETWEEN $1 AND $2)            AS new_customers,
			COUNT(*) FILTER (WHERE ac.first_order < $1)                         AS returning_customers
		FROM in_range ir
		JOIN all_customers ac ON ac.customer_phone = ir.customer_phone
	`

	var c models.CustomerOverviewCard
	if err := r.pool.QueryRow(ctx, query, from, to).Scan(
		&c.TotalUniqueCustomers,
		&c.NewCustomers,
		&c.ReturningCustomers,
	); err != nil {
		return models.CustomerOverviewCard{}, fmt.Errorf("failed to scan customer overview: %w", err)
	}

	if c.TotalUniqueCustomers > 0 {
		c.RetentionPercent = float64(c.ReturningCustomers) / float64(c.TotalUniqueCustomers) * 100
	}

	// Token circulation
	if err := r.pool.QueryRow(ctx, `
		SELECT
			COALESCE(SUM(total_tokens), 0)  AS total_in_circulation
		FROM user_tokens
	`).Scan(&c.TotalTokensInCirculation); err != nil {
		return models.CustomerOverviewCard{}, fmt.Errorf("failed to scan token circulation: %w", err)
	}

	// Total spent + redemption rate
	if err := r.pool.QueryRow(ctx, `
		SELECT
			COALESCE(SUM(amount), 0) AS total_spent
		FROM token_transactions
		WHERE type = 'SPEND'
	`).Scan(&c.TotalTokensSpent); err != nil {
		return models.CustomerOverviewCard{}, fmt.Errorf("failed to scan tokens spent: %w", err)
	}

	var totalEarned float64
	if err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount), 0)
		FROM token_transactions
		WHERE type = 'EARN'
	`).Scan(&totalEarned); err == nil && totalEarned > 0 {
		c.TokenRedemptionRate = (c.TotalTokensSpent / totalEarned) * 100
	}

	return c, nil
}

// ─── Top Customers ────────────────────────────────────────────────────────────

func (r *reportRepo) getTopCustomers(ctx context.Context, from, to time.Time) ([]models.TopCustomerRow, error) {
	query := `
		SELECT COALESCE(
			json_agg(
				json_build_object(
					'phone',          phone,
					'customer_name',  customer_name,
					'total_spend',    total_spend,
					'visit_count',    visit_count,
					'total_tokens',   total_tokens,
					'current_streak', current_streak
				) ORDER BY total_spend DESC
			), '[]'::json
		)
		FROM (
			SELECT
				o.customer_phone                         AS phone,
				COALESCE(MAX(o.customer_name), '')       AS customer_name,
				COALESCE(SUM(p.paid_amount), 0)          AS total_spend,
				COUNT(DISTINCT o.id)                     AS visit_count,
				COALESCE(ut.total_tokens, 0)             AS total_tokens,
				COALESCE(cs.current_streak, 0)           AS current_streak
			FROM orders o
			JOIN payments p     ON p.order_id = o.id
			LEFT JOIN user_tokens ut ON ut.phone_number = o.customer_phone
			LEFT JOIN customer_streaks cs ON cs.phone_number = o.customer_phone
			WHERE o.customer_phone IS NOT NULL
			  AND o.created_at BETWEEN $1 AND $2
			  AND o.status = 'completed'
			GROUP BY o.customer_phone, ut.total_tokens, cs.current_streak
			ORDER BY total_spend DESC
			LIMIT 20
		) sub
	`

	var resultJSON []byte
	if err := r.pool.QueryRow(ctx, query, from, to).Scan(&resultJSON); err != nil {
		return nil, fmt.Errorf("failed to query top customers: %w", err)
	}
	var result []models.TopCustomerRow
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal top customers: %w", err)
	}
	return result, nil
}

// ─── Visit Frequency Distribution ────────────────────────────────────────────

func (r *reportRepo) getVisitFrequency(ctx context.Context, from, to time.Time) ([]models.CustomerVisitFrequency, error) {
	query := `
		SELECT COALESCE(
			json_agg(
				json_build_object(
					'visit_bucket',    visit_bucket,
					'customer_count',  customer_count,
					'percent',         ROUND((customer_count::numeric / NULLIF(total, 0) * 100), 2)
				) ORDER BY min_visits ASC
			), '[]'::json
		)
		FROM (
			SELECT
				visit_bucket,
				min_visits,
				COUNT(customer_phone) AS customer_count,
				SUM(COUNT(customer_phone)) OVER () AS total
			FROM (
				SELECT
					customer_phone,
					COUNT(DISTINCT id) AS visits,
					CASE
						WHEN COUNT(DISTINCT id) = 1          THEN '1 visit'
						WHEN COUNT(DISTINCT id) BETWEEN 2 AND 5 THEN '2–5 visits'
						WHEN COUNT(DISTINCT id) BETWEEN 6 AND 10 THEN '6–10 visits'
						ELSE '10+ visits'
					END AS visit_bucket,
					CASE
						WHEN COUNT(DISTINCT id) = 1          THEN 1
						WHEN COUNT(DISTINCT id) BETWEEN 2 AND 5 THEN 2
						WHEN COUNT(DISTINCT id) BETWEEN 6 AND 10 THEN 6
						ELSE 11
					END AS min_visits
				FROM orders
				WHERE customer_phone IS NOT NULL
				  AND created_at BETWEEN $1 AND $2
				GROUP BY customer_phone
			) bucketed
			GROUP BY visit_bucket, min_visits
		) sub
	`

	var resultJSON []byte
	if err := r.pool.QueryRow(ctx, query, from, to).Scan(&resultJSON); err != nil {
		return nil, fmt.Errorf("failed to query visit frequency: %w", err)
	}
	var result []models.CustomerVisitFrequency
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal visit frequency: %w", err)
	}
	return result, nil
}

// ─── Token Economy ────────────────────────────────────────────────────────────

func (r *reportRepo) getTokenEconomy(ctx context.Context, from, to time.Time) (models.TokenEconomyReport, error) {
	// Overall totals
	var t models.TokenEconomyReport

	if err := r.pool.QueryRow(ctx, `
		SELECT
			COALESCE(SUM(amount) FILTER (WHERE type = 'EARN'),   0) AS earned,
			COALESCE(SUM(amount) FILTER (WHERE type = 'SPEND'),  0) AS spent,
			COALESCE(SUM(amount) FILTER (WHERE type = 'STREAK'), 0) AS streak
		FROM token_transactions
		WHERE created_at BETWEEN $1 AND $2
	`, from, to).Scan(&t.TotalEarned, &t.TotalSpent, &t.TotalStreakBonuses); err != nil {
		return models.TokenEconomyReport{}, fmt.Errorf("failed to scan token economy totals: %w", err)
	}

	t.NetCirculation = t.TotalEarned - t.TotalSpent
	if t.TotalEarned > 0 {
		t.RedemptionRate = (t.TotalSpent / t.TotalEarned) * 100
	}

	// Monthly earn/spend/streak trend
	trendQuery := `
		SELECT COALESCE(
			json_agg(
				json_build_object(
					'period',       period,
					'earned',       earned,
					'spent',        spent,
					'streak_bonus', streak_bonus
				) ORDER BY period ASC
			), '[]'::json
		)
		FROM (
			SELECT
				TO_CHAR(DATE_TRUNC('month', created_at AT TIME ZONE 'Asia/Kathmandu'), 'YYYY-MM') AS period,
				COALESCE(SUM(amount) FILTER (WHERE type = 'EARN'),   0) AS earned,
				COALESCE(SUM(amount) FILTER (WHERE type = 'SPEND'),  0) AS spent,
				COALESCE(SUM(amount) FILTER (WHERE type = 'STREAK'), 0) AS streak_bonus
			FROM token_transactions
			WHERE created_at BETWEEN $1 AND $2
			GROUP BY DATE_TRUNC('month', created_at AT TIME ZONE 'Asia/Kathmandu')
		) sub
	`

	var resultJSON []byte
	if err := r.pool.QueryRow(ctx, trendQuery, from, to).Scan(&resultJSON); err != nil {
		return models.TokenEconomyReport{}, fmt.Errorf("failed to query token trend: %w", err)
	}
	if err := json.Unmarshal(resultJSON, &t.EarnSpendTrend); err != nil {
		return models.TokenEconomyReport{}, fmt.Errorf("failed to unmarshal token trend: %w", err)
	}

	return t, nil
}

// ─── Streak Leaderboard ───────────────────────────────────────────────────────

func (r *reportRepo) getStreakLeaderboard(ctx context.Context) ([]models.StreakLeaderboardRow, error) {
	query := `
		SELECT COALESCE(
			json_agg(
				json_build_object(
					'phone',          phone_number,
					'current_streak', current_streak,
					'monthly_days',   monthly_days,
					'last_visit',     TO_CHAR(last_visit, 'YYYY-MM-DD')
				) ORDER BY current_streak DESC, monthly_days DESC
			), '[]'::json
		)
		FROM (
			SELECT phone_number, current_streak, monthly_days, last_visit
			FROM customer_streaks
			WHERE current_streak > 0
			ORDER BY current_streak DESC
			LIMIT 20
		) sub
	`

	var resultJSON []byte
	if err := r.pool.QueryRow(ctx, query).Scan(&resultJSON); err != nil {
		return nil, fmt.Errorf("failed to query streak leaderboard: %w", err)
	}
	var result []models.StreakLeaderboardRow
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal streak leaderboard: %w", err)
	}
	return result, nil
}

func (r *reportRepo) GetFinancialSummary(ctx context.Context) (*models.FinancialSummaryResponse, error) {
	var invested float64
	if err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(price * quantity), 0)
		FROM raw_material
	`).Scan(&invested); err != nil {
		return nil, fmt.Errorf("failed to scan total invested: %w", err)
	}

	var earned float64
	if err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(paid_amount), 0)
		FROM payments
	`).Scan(&earned); err != nil {
		return nil, fmt.Errorf("failed to scan total earned: %w", err)
	}

	grossProfit := earned - invested
	var profitPercent float64
	if invested > 0 {
		profitPercent = (grossProfit / invested) * 100
	}

	return &models.FinancialSummaryResponse{
		TotalInvested: invested,
		TotalEarned:   earned,
		GrossProfit:   grossProfit,
		ProfitPercent: profitPercent,
	}, nil
}

func (r *reportRepo) GetRawMaterialReport(ctx context.Context) (*models.RawMaterialReportResponse, error) {
	var overview models.RawMaterialOverviewCard
	if err := r.pool.QueryRow(ctx, `
		SELECT
			COUNT(*)                                              AS total_materials,
			COALESCE(SUM(price * quantity), 0)                   AS total_inventory_value,
			COUNT(*) FILTER (WHERE quantity <= 5)                AS low_stock_count,
			COALESCE(SUM(price * quantity), 0)                   AS total_invested
		FROM raw_material
	`).Scan(
		&overview.TotalMaterials,
		&overview.TotalInventoryValue,
		&overview.LowStockCount,
		&overview.TotalInvested,
	); err != nil {
		return nil, fmt.Errorf("failed to scan raw material overview: %w", err)
	}

	var resultJSON []byte
	if err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(
			json_agg(
				json_build_object(
					'id',            id,
					'name',          name,
					'quantity',      quantity,
					'unit',          unit,
					'price',         price,
					'total_value',   total_value,
					'value_percent', ROUND((total_value / NULLIF(grand_total, 0) * 100)::numeric, 2),
					'is_low_stock',  is_low_stock,
					'last_updated',  last_updated
				) ORDER BY total_value DESC
			), '[]'::json
		)
		FROM (
			SELECT
				id::text                             AS id,
				COALESCE(name, '')                   AS name,
				COALESCE(quantity, 0)                AS quantity,
				COALESCE(unit, 'kg')                 AS unit,
				COALESCE(price, 0)                   AS price,
				COALESCE(price * quantity, 0)        AS total_value,
				SUM(COALESCE(price * quantity, 0)) OVER () AS grand_total,
				quantity <= 5                        AS is_low_stock,
				updated_at                           AS last_updated
			FROM raw_material
		) sub
	`).Scan(&resultJSON); err != nil {
		return nil, fmt.Errorf("failed to query raw materials: %w", err)
	}

	var materials []models.RawMaterialRow
	if err := json.Unmarshal(resultJSON, &materials); err != nil {
		return nil, fmt.Errorf("failed to unmarshal raw materials: %w", err)
	}

	return &models.RawMaterialReportResponse{
		Overview:  overview,
		Materials: materials,
	}, nil
}

func (r *reportRepo) getSalesDailyTrend(ctx context.Context) ([]models.SalesTrendPoint, error) {
	query := `
		SELECT COALESCE(
			json_agg(
				json_build_object(
					'period',          period,
					'total_items_sold', total_items_sold,
					'total_orders',    total_orders,
					'total_revenue',   total_revenue
				) ORDER BY period ASC
			), '[]'::json
		)
		FROM (
			SELECT
				TO_CHAR(oi.created_at AT TIME ZONE 'Asia/Kathmandu', 'YYYY-MM-DD') AS period,
				COALESCE(SUM(oi.quantity), 0)            AS total_items_sold,
				COUNT(DISTINCT oi.order_id)              AS total_orders,
				COALESCE(SUM(oi.quantity * oi.price), 0) AS total_revenue
			FROM order_items oi
			JOIN orders o ON o.id = oi.order_id
			WHERE o.status = 'completed'
			  AND oi.created_at >= NOW() - INTERVAL '90 days'
			GROUP BY TO_CHAR(oi.created_at AT TIME ZONE 'Asia/Kathmandu', 'YYYY-MM-DD')
		) sub
	`
	return r.scanSalesTrendPoints(ctx, query)
}

func (r *reportRepo) getSalesWeeklyTrend(ctx context.Context) ([]models.SalesTrendPoint, error) {
	query := `
		SELECT COALESCE(
			json_agg(
				json_build_object(
					'period',          period,
					'total_items_sold', total_items_sold,
					'total_orders',    total_orders,
					'total_revenue',   total_revenue
				) ORDER BY period ASC
			), '[]'::json
		)
		FROM (
			SELECT
				TO_CHAR(DATE_TRUNC('week', oi.created_at AT TIME ZONE 'Asia/Kathmandu'), 'IYYY-"W"IW') AS period,
				COALESCE(SUM(oi.quantity), 0)            AS total_items_sold,
				COUNT(DISTINCT oi.order_id)              AS total_orders,
				COALESCE(SUM(oi.quantity * oi.price), 0) AS total_revenue
			FROM order_items oi
			JOIN orders o ON o.id = oi.order_id
			WHERE o.status = 'completed'
			  AND oi.created_at >= NOW() - INTERVAL '52 weeks'
			GROUP BY DATE_TRUNC('week', oi.created_at AT TIME ZONE 'Asia/Kathmandu')
		) sub
	`
	return r.scanSalesTrendPoints(ctx, query)
}

func (r *reportRepo) getSalesMonthlyTrend(ctx context.Context) ([]models.SalesTrendPoint, error) {
	query := `
		SELECT COALESCE(
			json_agg(
				json_build_object(
					'period',          period,
					'total_items_sold', total_items_sold,
					'total_orders',    total_orders,
					'total_revenue',   total_revenue
				) ORDER BY period ASC
			), '[]'::json
		)
		FROM (
			SELECT
				TO_CHAR(DATE_TRUNC('month', oi.created_at AT TIME ZONE 'Asia/Kathmandu'), 'YYYY-MM') AS period,
				COALESCE(SUM(oi.quantity), 0)            AS total_items_sold,
				COUNT(DISTINCT oi.order_id)              AS total_orders,
				COALESCE(SUM(oi.quantity * oi.price), 0) AS total_revenue
			FROM order_items oi
			JOIN orders o ON o.id = oi.order_id
			WHERE o.status = 'completed'
			  AND oi.created_at >= NOW() - INTERVAL '24 months'
			GROUP BY DATE_TRUNC('month', oi.created_at AT TIME ZONE 'Asia/Kathmandu')
		) sub
	`
	return r.scanSalesTrendPoints(ctx, query)
}

func (r *reportRepo) getSalesYearlyTrend(ctx context.Context) ([]models.SalesTrendPoint, error) {
	query := `
		SELECT COALESCE(
			json_agg(
				json_build_object(
					'period',          period,
					'total_items_sold', total_items_sold,
					'total_orders',    total_orders,
					'total_revenue',   total_revenue
				) ORDER BY period ASC
			), '[]'::json
		)
		FROM (
			SELECT
				TO_CHAR(DATE_TRUNC('year', oi.created_at AT TIME ZONE 'Asia/Kathmandu'), 'YYYY') AS period,
				COALESCE(SUM(oi.quantity), 0)            AS total_items_sold,
				COUNT(DISTINCT oi.order_id)              AS total_orders,
				COALESCE(SUM(oi.quantity * oi.price), 0) AS total_revenue
			FROM order_items oi
			JOIN orders o ON o.id = oi.order_id
			WHERE o.status = 'completed'
			GROUP BY DATE_TRUNC('year', oi.created_at AT TIME ZONE 'Asia/Kathmandu')
		) sub
	`
	return r.scanSalesTrendPoints(ctx, query)
}

func (r *reportRepo) scanSalesTrendPoints(ctx context.Context, query string) ([]models.SalesTrendPoint, error) {
	var resultJSON []byte
	if err := r.pool.QueryRow(ctx, query).Scan(&resultJSON); err != nil {
		return nil, fmt.Errorf("failed to query sales trend: %w", err)
	}
	var result []models.SalesTrendPoint
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal sales trend: %w", err)
	}
	return result, nil
}

// ─── Customer Trends ──────────────────────────────────────────────────────────

func (r *reportRepo) getCustomerDailyTrend(ctx context.Context) ([]models.CustomerTrendPoint, error) {
	query := `
		SELECT COALESCE(
			json_agg(
				json_build_object(
					'period',              period,
					'new_customers',       new_customers,
					'returning_customers', returning_customers,
					'total_orders',        total_orders,
					'total_revenue',       total_revenue
				) ORDER BY period ASC
			), '[]'::json
		)
		FROM (
			SELECT
				TO_CHAR(o.created_at AT TIME ZONE 'Asia/Kathmandu', 'YYYY-MM-DD') AS period,
				COUNT(DISTINCT o.customer_phone) FILTER (
					WHERE o.customer_phone NOT IN (
						SELECT customer_phone FROM orders
						WHERE created_at < DATE_TRUNC('day', o.created_at AT TIME ZONE 'Asia/Kathmandu')
						  AND customer_phone IS NOT NULL
					)
				) AS new_customers,
				COUNT(DISTINCT o.customer_phone) FILTER (
					WHERE o.customer_phone IN (
						SELECT customer_phone FROM orders
						WHERE created_at < DATE_TRUNC('day', o.created_at AT TIME ZONE 'Asia/Kathmandu')
						  AND customer_phone IS NOT NULL
					)
				) AS returning_customers,
				COUNT(DISTINCT o.id)             AS total_orders,
				COALESCE(SUM(p.paid_amount), 0)  AS total_revenue
			FROM orders o
			LEFT JOIN payments p ON p.order_id = o.id
			WHERE o.customer_phone IS NOT NULL
			  AND o.status = 'completed'
			  AND o.created_at >= NOW() - INTERVAL '90 days'
			GROUP BY TO_CHAR(o.created_at AT TIME ZONE 'Asia/Kathmandu', 'YYYY-MM-DD')
		) sub
	`
	return r.scanCustomerTrendPoints(ctx, query)
}

func (r *reportRepo) getCustomerWeeklyTrend(ctx context.Context) ([]models.CustomerTrendPoint, error) {
	query := `
		SELECT COALESCE(
			json_agg(
				json_build_object(
					'period',              period,
					'new_customers',       new_customers,
					'returning_customers', returning_customers,
					'total_orders',        total_orders,
					'total_revenue',       total_revenue
				) ORDER BY period ASC
			), '[]'::json
		)
		FROM (
			SELECT
				TO_CHAR(DATE_TRUNC('week', o.created_at AT TIME ZONE 'Asia/Kathmandu'), 'IYYY-"W"IW') AS period,
				COUNT(DISTINCT o.customer_phone) FILTER (
					WHERE o.customer_phone NOT IN (
						SELECT customer_phone FROM orders
						WHERE created_at < DATE_TRUNC('week', o.created_at AT TIME ZONE 'Asia/Kathmandu')
						  AND customer_phone IS NOT NULL
					)
				) AS new_customers,
				COUNT(DISTINCT o.customer_phone) FILTER (
					WHERE o.customer_phone IN (
						SELECT customer_phone FROM orders
						WHERE created_at < DATE_TRUNC('week', o.created_at AT TIME ZONE 'Asia/Kathmandu')
						  AND customer_phone IS NOT NULL
					)
				) AS returning_customers,
				COUNT(DISTINCT o.id)             AS total_orders,
				COALESCE(SUM(p.paid_amount), 0)  AS total_revenue
			FROM orders o
			LEFT JOIN payments p ON p.order_id = o.id
			WHERE o.customer_phone IS NOT NULL
			  AND o.status = 'completed'
			  AND o.created_at >= NOW() - INTERVAL '52 weeks'
			GROUP BY DATE_TRUNC('week', o.created_at AT TIME ZONE 'Asia/Kathmandu')
		) sub
	`
	return r.scanCustomerTrendPoints(ctx, query)
}

func (r *reportRepo) getCustomerMonthlyTrend(ctx context.Context) ([]models.CustomerTrendPoint, error) {
	query := `
		SELECT COALESCE(
			json_agg(
				json_build_object(
					'period',              period,
					'new_customers',       new_customers,
					'returning_customers', returning_customers,
					'total_orders',        total_orders,
					'total_revenue',       total_revenue
				) ORDER BY period ASC
			), '[]'::json
		)
		FROM (
			SELECT
				TO_CHAR(DATE_TRUNC('month', o.created_at AT TIME ZONE 'Asia/Kathmandu'), 'YYYY-MM') AS period,
				COUNT(DISTINCT o.customer_phone) FILTER (
					WHERE o.customer_phone NOT IN (
						SELECT customer_phone FROM orders
						WHERE created_at < DATE_TRUNC('month', o.created_at AT TIME ZONE 'Asia/Kathmandu')
						  AND customer_phone IS NOT NULL
					)
				) AS new_customers,
				COUNT(DISTINCT o.customer_phone) FILTER (
					WHERE o.customer_phone IN (
						SELECT customer_phone FROM orders
						WHERE created_at < DATE_TRUNC('month', o.created_at AT TIME ZONE 'Asia/Kathmandu')
						  AND customer_phone IS NOT NULL
					)
				) AS returning_customers,
				COUNT(DISTINCT o.id)             AS total_orders,
				COALESCE(SUM(p.paid_amount), 0)  AS total_revenue
			FROM orders o
			LEFT JOIN payments p ON p.order_id = o.id
			WHERE o.customer_phone IS NOT NULL
			  AND o.status = 'completed'
			  AND o.created_at >= NOW() - INTERVAL '24 months'
			GROUP BY DATE_TRUNC('month', o.created_at AT TIME ZONE 'Asia/Kathmandu')
		) sub
	`
	return r.scanCustomerTrendPoints(ctx, query)
}

func (r *reportRepo) getCustomerYearlyTrend(ctx context.Context) ([]models.CustomerTrendPoint, error) {
	query := `
		SELECT COALESCE(
			json_agg(
				json_build_object(
					'period',              period,
					'new_customers',       new_customers,
					'returning_customers', returning_customers,
					'total_orders',        total_orders,
					'total_revenue',       total_revenue
				) ORDER BY period ASC
			), '[]'::json
		)
		FROM (
			SELECT
				TO_CHAR(DATE_TRUNC('year', o.created_at AT TIME ZONE 'Asia/Kathmandu'), 'YYYY') AS period,
				COUNT(DISTINCT o.customer_phone) FILTER (
					WHERE o.customer_phone NOT IN (
						SELECT customer_phone FROM orders
						WHERE created_at < DATE_TRUNC('year', o.created_at AT TIME ZONE 'Asia/Kathmandu')
						  AND customer_phone IS NOT NULL
					)
				) AS new_customers,
				COUNT(DISTINCT o.customer_phone) FILTER (
					WHERE o.customer_phone IN (
						SELECT customer_phone FROM orders
						WHERE created_at < DATE_TRUNC('year', o.created_at AT TIME ZONE 'Asia/Kathmandu')
						  AND customer_phone IS NOT NULL
					)
				) AS returning_customers,
				COUNT(DISTINCT o.id)             AS total_orders,
				COALESCE(SUM(p.paid_amount), 0)  AS total_revenue
			FROM orders o
			LEFT JOIN payments p ON p.order_id = o.id
			WHERE o.customer_phone IS NOT NULL
			  AND o.status = 'completed'
			GROUP BY DATE_TRUNC('year', o.created_at AT TIME ZONE 'Asia/Kathmandu')
		) sub
	`
	return r.scanCustomerTrendPoints(ctx, query)
}

func (r *reportRepo) scanCustomerTrendPoints(ctx context.Context, query string) ([]models.CustomerTrendPoint, error) {
	var resultJSON []byte
	if err := r.pool.QueryRow(ctx, query).Scan(&resultJSON); err != nil {
		return nil, fmt.Errorf("failed to query customer trend: %w", err)
	}
	var result []models.CustomerTrendPoint
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal customer trend: %w", err)
	}
	return result, nil
}

// ─── Table Trends ─────────────────────────────────────────────────────────────

func (r *reportRepo) getTableDailyTrend(ctx context.Context) ([]models.TableTrendPoint, error) {
	query := `
		SELECT COALESCE(
			json_agg(
				json_build_object(
					'period',         period,
					'total_sessions', total_sessions,
					'total_revenue',  total_revenue,
					'avg_session_minutes', avg_session_minutes
				) ORDER BY period ASC
			), '[]'::json
		)
		FROM (
			SELECT
				TO_CHAR(ts.open_time AT TIME ZONE 'Asia/Kathmandu', 'YYYY-MM-DD') AS period,
				COUNT(ts.id)                                                        AS total_sessions,
				COALESCE(SUM(p.paid_amount), 0)                                    AS total_revenue,
				COALESCE(AVG(
					EXTRACT(EPOCH FROM (COALESCE(ts.close_time, NOW()) - ts.open_time)) / 60
				), 0)                                                               AS avg_session_minutes
			FROM table_session ts
			LEFT JOIN orders o   ON o.table_session_id = ts.id
			LEFT JOIN payments p ON p.order_id = o.id
			WHERE ts.open_time >= NOW() - INTERVAL '90 days'
			GROUP BY TO_CHAR(ts.open_time AT TIME ZONE 'Asia/Kathmandu', 'YYYY-MM-DD')
		) sub
	`
	return r.scanTableTrendPoints(ctx, query)
}

func (r *reportRepo) getTableWeeklyTrend(ctx context.Context) ([]models.TableTrendPoint, error) {
	query := `
		SELECT COALESCE(
			json_agg(
				json_build_object(
					'period',         period,
					'total_sessions', total_sessions,
					'total_revenue',  total_revenue,
					'avg_session_minutes', avg_session_minutes
				) ORDER BY period ASC
			), '[]'::json
		)
		FROM (
			SELECT
				TO_CHAR(DATE_TRUNC('week', ts.open_time AT TIME ZONE 'Asia/Kathmandu'), 'IYYY-"W"IW') AS period,
				COUNT(ts.id)                                                        AS total_sessions,
				COALESCE(SUM(p.paid_amount), 0)                                    AS total_revenue,
				COALESCE(AVG(
					EXTRACT(EPOCH FROM (COALESCE(ts.close_time, NOW()) - ts.open_time)) / 60
				), 0)                                                               AS avg_session_minutes
			FROM table_session ts
			LEFT JOIN orders o   ON o.table_session_id = ts.id
			LEFT JOIN payments p ON p.order_id = o.id
			WHERE ts.open_time >= NOW() - INTERVAL '52 weeks'
			GROUP BY DATE_TRUNC('week', ts.open_time AT TIME ZONE 'Asia/Kathmandu')
		) sub
	`
	return r.scanTableTrendPoints(ctx, query)
}

func (r *reportRepo) getTableMonthlyTrend(ctx context.Context) ([]models.TableTrendPoint, error) {
	query := `
		SELECT COALESCE(
			json_agg(
				json_build_object(
					'period',         period,
					'total_sessions', total_sessions,
					'total_revenue',  total_revenue,
					'avg_session_minutes', avg_session_minutes
				) ORDER BY period ASC
			), '[]'::json
		)
		FROM (
			SELECT
				TO_CHAR(DATE_TRUNC('month', ts.open_time AT TIME ZONE 'Asia/Kathmandu'), 'YYYY-MM') AS period,
				COUNT(ts.id)                                                        AS total_sessions,
				COALESCE(SUM(p.paid_amount), 0)                                    AS total_revenue,
				COALESCE(AVG(
					EXTRACT(EPOCH FROM (COALESCE(ts.close_time, NOW()) - ts.open_time)) / 60
				), 0)                                                               AS avg_session_minutes
			FROM table_session ts
			LEFT JOIN orders o   ON o.table_session_id = ts.id
			LEFT JOIN payments p ON p.order_id = o.id
			WHERE ts.open_time >= NOW() - INTERVAL '24 months'
			GROUP BY DATE_TRUNC('month', ts.open_time AT TIME ZONE 'Asia/Kathmandu')
		) sub
	`
	return r.scanTableTrendPoints(ctx, query)
}

func (r *reportRepo) getTableYearlyTrend(ctx context.Context) ([]models.TableTrendPoint, error) {
	query := `
		SELECT COALESCE(
			json_agg(
				json_build_object(
					'period',         period,
					'total_sessions', total_sessions,
					'total_revenue',  total_revenue,
					'avg_session_minutes', avg_session_minutes
				) ORDER BY period ASC
			), '[]'::json
		)
		FROM (
			SELECT
				TO_CHAR(DATE_TRUNC('year', ts.open_time AT TIME ZONE 'Asia/Kathmandu'), 'YYYY') AS period,
				COUNT(ts.id)                                                        AS total_sessions,
				COALESCE(SUM(p.paid_amount), 0)                                    AS total_revenue,
				COALESCE(AVG(
					EXTRACT(EPOCH FROM (COALESCE(ts.close_time, NOW()) - ts.open_time)) / 60
				), 0)                                                               AS avg_session_minutes
			FROM table_session ts
			LEFT JOIN orders o   ON o.table_session_id = ts.id
			LEFT JOIN payments p ON p.order_id = o.id
			GROUP BY DATE_TRUNC('year', ts.open_time AT TIME ZONE 'Asia/Kathmandu')
		) sub
	`
	return r.scanTableTrendPoints(ctx, query)
}

func (r *reportRepo) scanTableTrendPoints(ctx context.Context, query string) ([]models.TableTrendPoint, error) {
	var resultJSON []byte
	if err := r.pool.QueryRow(ctx, query).Scan(&resultJSON); err != nil {
		return nil, fmt.Errorf("failed to query table trend: %w", err)
	}
	var result []models.TableTrendPoint
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal table trend: %w", err)
	}
	return result, nil
}

// ─── Staff Trends ─────────────────────────────────────────────────────────────
// Note: daily staff trend == existing DailySummary, so we add weekly/monthly/yearly only

func (r *reportRepo) getStaffWeeklyTrend(ctx context.Context) ([]models.StaffTrendPoint, error) {
	query := `
		SELECT COALESCE(
			json_agg(
				json_build_object(
					'period',          period,
					'present',         present,
					'absent',          absent,
					'late',            late,
					'on_leave',        on_leave,
					'attendance_rate', attendance_rate
				) ORDER BY period ASC
			), '[]'::json
		)
		FROM (
			SELECT
				TO_CHAR(DATE_TRUNC('week', work_date AT TIME ZONE 'Asia/Kathmandu'), 'IYYY-"W"IW') AS period,
				COUNT(*) FILTER (WHERE status = 'present')  AS present,
				COUNT(*) FILTER (WHERE status = 'absent')   AS absent,
				COUNT(*) FILTER (WHERE status = 'late')     AS late,
				COUNT(*) FILTER (WHERE status = 'leave')    AS on_leave,
				ROUND(
					COUNT(*) FILTER (WHERE status IN ('present','late'))::numeric
					/ NULLIF(COUNT(*), 0) * 100, 2
				) AS attendance_rate
			FROM attendance
			WHERE work_date >= NOW() - INTERVAL '52 weeks'
			GROUP BY DATE_TRUNC('week', work_date AT TIME ZONE 'Asia/Kathmandu')
		) sub
	`
	return r.scanStaffTrendPoints(ctx, query)
}

func (r *reportRepo) getStaffMonthlyTrend(ctx context.Context) ([]models.StaffTrendPoint, error) {
	query := `
		SELECT COALESCE(
			json_agg(
				json_build_object(
					'period',          period,
					'present',         present,
					'absent',          absent,
					'late',            late,
					'on_leave',        on_leave,
					'attendance_rate', attendance_rate
				) ORDER BY period ASC
			), '[]'::json
		)
		FROM (
			SELECT
				TO_CHAR(DATE_TRUNC('month', work_date AT TIME ZONE 'Asia/Kathmandu'), 'YYYY-MM') AS period,
				COUNT(*) FILTER (WHERE status = 'present')  AS present,
				COUNT(*) FILTER (WHERE status = 'absent')   AS absent,
				COUNT(*) FILTER (WHERE status = 'late')     AS late,
				COUNT(*) FILTER (WHERE status = 'leave')    AS on_leave,
				ROUND(
					COUNT(*) FILTER (WHERE status IN ('present','late'))::numeric
					/ NULLIF(COUNT(*), 0) * 100, 2
				) AS attendance_rate
			FROM attendance
			WHERE work_date >= NOW() - INTERVAL '24 months'
			GROUP BY DATE_TRUNC('month', work_date AT TIME ZONE 'Asia/Kathmandu')
		) sub
	`
	return r.scanStaffTrendPoints(ctx, query)
}

func (r *reportRepo) getStaffYearlyTrend(ctx context.Context) ([]models.StaffTrendPoint, error) {
	query := `
		SELECT COALESCE(
			json_agg(
				json_build_object(
					'period',          period,
					'present',         present,
					'absent',          absent,
					'late',            late,
					'on_leave',        on_leave,
					'attendance_rate', attendance_rate
				) ORDER BY period ASC
			), '[]'::json
		)
		FROM (
			SELECT
				TO_CHAR(DATE_TRUNC('year', work_date AT TIME ZONE 'Asia/Kathmandu'), 'YYYY') AS period,
				COUNT(*) FILTER (WHERE status = 'present')  AS present,
				COUNT(*) FILTER (WHERE status = 'absent')   AS absent,
				COUNT(*) FILTER (WHERE status = 'late')     AS late,
				COUNT(*) FILTER (WHERE status = 'leave')    AS on_leave,
				ROUND(
					COUNT(*) FILTER (WHERE status IN ('present','late'))::numeric
					/ NULLIF(COUNT(*), 0) * 100, 2
				) AS attendance_rate
			FROM attendance
			GROUP BY DATE_TRUNC('year', work_date AT TIME ZONE 'Asia/Kathmandu')
		) sub
	`
	return r.scanStaffTrendPoints(ctx, query)
}

func (r *reportRepo) scanStaffTrendPoints(ctx context.Context, query string) ([]models.StaffTrendPoint, error) {
	var resultJSON []byte
	if err := r.pool.QueryRow(ctx, query).Scan(&resultJSON); err != nil {
		return nil, fmt.Errorf("failed to query staff trend: %w", err)
	}
	var result []models.StaffTrendPoint
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal staff trend: %w", err)
	}
	return result, nil
}

// ─── Financial Trends ─────────────────────────────────────────────────────────

func (r *reportRepo) getFinancialMonthlyTrend(ctx context.Context) ([]models.FinancialTrendPoint, error) {
	query := `
		SELECT COALESCE(
			json_agg(
				json_build_object(
					'period',       period,
					'revenue',      revenue,
					'material_cost', material_cost,
					'gross_profit', gross_profit
				) ORDER BY period ASC
			), '[]'::json
		)
		FROM (
			SELECT
				TO_CHAR(DATE_TRUNC('month', created_at AT TIME ZONE 'Asia/Kathmandu'), 'YYYY-MM') AS period,
				COALESCE(SUM(paid_amount), 0) AS revenue,
				0::float8                     AS material_cost,  -- no purchase_date on raw_material, so 0
				COALESCE(SUM(paid_amount), 0) AS gross_profit
			FROM payments
			WHERE created_at >= NOW() - INTERVAL '24 months'
			GROUP BY DATE_TRUNC('month', created_at AT TIME ZONE 'Asia/Kathmandu')
		) sub
	`
	return r.scanFinancialTrendPoints(ctx, query)
}

func (r *reportRepo) getFinancialYearlyTrend(ctx context.Context) ([]models.FinancialTrendPoint, error) {
	query := `
		SELECT COALESCE(
			json_agg(
				json_build_object(
					'period',        period,
					'revenue',       revenue,
					'material_cost', material_cost,
					'gross_profit',  gross_profit
				) ORDER BY period ASC
			), '[]'::json
		)
		FROM (
			SELECT
				TO_CHAR(DATE_TRUNC('year', created_at AT TIME ZONE 'Asia/Kathmandu'), 'YYYY') AS period,
				COALESCE(SUM(paid_amount), 0) AS revenue,
				0::float8                     AS material_cost,
				COALESCE(SUM(paid_amount), 0) AS gross_profit
			FROM payments
			GROUP BY DATE_TRUNC('year', created_at AT TIME ZONE 'Asia/Kathmandu')
		) sub
	`
	return r.scanFinancialTrendPoints(ctx, query)
}

func (r *reportRepo) scanFinancialTrendPoints(ctx context.Context, query string) ([]models.FinancialTrendPoint, error) {
	var resultJSON []byte
	if err := r.pool.QueryRow(ctx, query).Scan(&resultJSON); err != nil {
		return nil, fmt.Errorf("failed to query financial trend: %w", err)
	}
	var result []models.FinancialTrendPoint
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal financial trend: %w", err)
	}
	return result, nil
}
