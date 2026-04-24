// repository/report_repo.go
package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Abhishekh669/backend/internals/models"
	"golang.org/x/sync/errgroup"
)

const (
	DefaultTrendLimit = 10
	MaxTrendLimit     = 50
)

const (
	DefaultSalesTrendLimit = 10
	MaxSalesTrendLimit     = 50
)

// type NewSalesRepo interface {
// 	NewGetDefaultSalesReport(ctx context.Context) (*models.NewDefaultSalesResponse, error)
// 	NewGetCustomRangeSalesReport(ctx context.Context, req *models.NewSalesCustomRangeReportRequest) (*models.NewCustomRangeSalesResponse, error)

// 	// Trend functions
// 	NewGetLast7DaysSalesTrend(ctx context.Context) ([]models.NewSalesTrendPoint, error)
// 	NewGetLast7WeeksSalesTrend(ctx context.Context) ([]models.NewSalesTrendPoint, error)
// 	NewGetLast7MonthsSalesTrend(ctx context.Context) ([]models.NewSalesTrendPoint, error)
// 	NewGetLast7YearsSalesTrend(ctx context.Context) ([]models.NewSalesTrendPoint, error)

// 	// Custom trend functions
// 	NewGetCustomDailySalesTrend(ctx context.Context, from, to time.Time, limit, page int) (*models.NewSalesPaginatedTrendPoints, error)
// 	NewGetCustomWeeklySalesTrend(ctx context.Context, from, to time.Time, limit, page int) (*models.NewSalesPaginatedTrendPoints, error)
// 	NewGetCustomMonthlySalesTrend(ctx context.Context, from, to time.Time, limit, page int) (*models.NewSalesPaginatedTrendPoints, error)
// 	NewGetCustomYearlySalesTrend(ctx context.Context, from, to time.Time, limit, page int) (*models.NewSalesPaginatedTrendPoints, error)

// 	// Analytics functions
// 	NewGetSalesOverview(ctx context.Context, from, to time.Time) (models.NewSalesOverviewCard, error)
// 	NewGetSalesStatsCard(ctx context.Context) (models.NewSalesStatsCard, error)
// 	NewGetTopSellingItems(ctx context.Context, from, to time.Time, limit int) ([]models.NewTopSellingItem, error)
// 	NewGetTopCategories(ctx context.Context, from, to time.Time, limit int) ([]models.NewTopCategory, error)
// 	NewGetOrderStatusBreakdown(ctx context.Context, from, to time.Time) ([]models.NewOrderStatusBreakdown, error)
// 	NewGetTablePerformance(ctx context.Context, from, to time.Time) ([]models.NewTablePerformance, error)
// 	NewGetStaffPerformance(ctx context.Context, from, to time.Time) ([]models.NewStaffPerformance, error)
// 	NewGetHourlySales(ctx context.Context, from, to time.Time) ([]models.NewHourlySalesPoint, error)
// 	NewGetDailySales(ctx context.Context, from, to time.Time) ([]models.NewDailySalesPoint, error)
// }

const (
	DefaultCustomerTrendLimit = 10
	MaxCustomerTrendLimit     = 50
)

const (
	DefaultTableTrendLimit = 10
	MaxTableTrendLimit     = 50
)

// ─── Constants ────────────────────────────────────────────────────────────────

const (
	DefaultStaffTrendLimit = 10
	MaxStaffTrendLimit     = 100
	TopNStaff              = 10 // how many rows for "top N" lists
)

const (
	DefaultRawMaterialTrendLimit = 10
	MaxRawMaterialTrendLimit     = 100
	TopNRawMaterials             = 10
)

func (r *reportRepo) NewGetDashboardReport(ctx context.Context) {

}

// ────────────────────────────────────────────────────────────────────────────
// DEFAULT REPORT
// ────────────────────────────────────────────────────────────────────────────

func (r *reportRepo) NewGetDefaultRawMaterialReport(ctx context.Context) (*models.NewDefaultRawMaterialResponse, error) {
	to := time.Now()
	from := to.AddDate(0, 0, -30)

	var (
		overview               models.NewRawMaterialOverviewCard
		statsCard              models.NewRawMaterialStatsCard
		dailyTrend             []models.NewRawMaterialTrendPoint
		weeklyTrend            []models.NewRawMaterialTrendPoint
		monthlyTrend           []models.NewRawMaterialTrendPoint
		yearlyTrend            []models.NewRawMaterialTrendPoint
		topUsedMaterials       []models.NewTopUsedRawMaterial
		materialUsageBreakdown []models.NewRawMaterialUsageBreakdown
		peakUsageHours         []models.NewRawMaterialPeakHour
		dailyUsageSummary      []models.NewDailyRawMaterialUsage
	)

	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() error { var e error; overview, e = r.NewGetRawMaterialOverview(gCtx, from, to); return e })
	g.Go(func() error { var e error; statsCard, e = r.NewGetRawMaterialStatsCard(gCtx); return e })
	g.Go(func() error { var e error; dailyTrend, e = r.NewGetLast7DaysRawMaterialTrend(gCtx); return e })
	g.Go(func() error { var e error; weeklyTrend, e = r.NewGetLast7WeeksRawMaterialTrend(gCtx); return e })
	g.Go(func() error { var e error; monthlyTrend, e = r.NewGetLast7MonthsRawMaterialTrend(gCtx); return e })
	g.Go(func() error { var e error; yearlyTrend, e = r.NewGetLast7YearsRawMaterialTrend(gCtx); return e })
	g.Go(func() error {
		var e error
		topUsedMaterials, e = r.NewGetTopUsedRawMaterials(gCtx, from, to, TopNRawMaterials)
		return e
	})
	g.Go(func() error {
		var e error
		materialUsageBreakdown, e = r.NewGetRawMaterialUsageBreakdown(gCtx, from, to)
		return e
	})
	g.Go(func() error { var e error; peakUsageHours, e = r.NewGetRawMaterialPeakHours(gCtx, from, to); return e })
	g.Go(func() error {
		var e error
		dailyUsageSummary, e = r.NewGetDailyRawMaterialUsage(gCtx, from, to)
		return e
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return &models.NewDefaultRawMaterialResponse{
		Overview:               overview,
		StatsCard:              statsCard,
		DailyTrend:             dailyTrend,
		WeeklyTrend:            weeklyTrend,
		MonthlyTrend:           monthlyTrend,
		YearlyTrend:            yearlyTrend,
		TopUsedMaterials:       topUsedMaterials,
		MaterialUsageBreakdown: materialUsageBreakdown,
		PeakUsageHours:         peakUsageHours,
		DailyUsageSummary:      dailyUsageSummary,
	}, nil
}

// ────────────────────────────────────────────────────────────────────────────
// CUSTOM RANGE REPORT
// ────────────────────────────────────────────────────────────────────────────

func (r *reportRepo) NewGetCustomRangeRawMaterialReport(ctx context.Context, req *models.NewRawMaterialCustomRangeReportRequest) (*models.NewCustomRangeRawMaterialResponse, error) {
	from := req.From
	to := req.To.Truncate(24 * time.Hour).Add(24*time.Hour - time.Second)

	limit := req.Limit
	if limit <= 0 {
		limit = DefaultRawMaterialTrendLimit
	}
	if limit > MaxRawMaterialTrendLimit {
		limit = MaxRawMaterialTrendLimit
	}
	page := req.Page
	if page < 0 {
		page = 0
	}

	var (
		overview               models.NewRawMaterialOverviewCard
		statsCard              models.NewRawMaterialStatsCard
		dailyTrend             *models.NewRawMaterialPaginatedTrendPoints
		weeklyTrend            *models.NewRawMaterialPaginatedTrendPoints
		monthlyTrend           *models.NewRawMaterialPaginatedTrendPoints
		yearlyTrend            *models.NewRawMaterialPaginatedTrendPoints
		topUsedMaterials       []models.NewTopUsedRawMaterial
		materialUsageBreakdown []models.NewRawMaterialUsageBreakdown
		peakUsageHours         []models.NewRawMaterialPeakHour
		dailyUsageSummary      []models.NewDailyRawMaterialUsage
	)

	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() error { var e error; overview, e = r.NewGetRawMaterialOverview(gCtx, from, to); return e })
	g.Go(func() error { var e error; statsCard, e = r.NewGetRawMaterialStatsCard(gCtx); return e })
	g.Go(func() error {
		var e error
		dailyTrend, e = r.NewGetCustomDailyRawMaterialTrend(gCtx, from, to, limit, page)
		return e
	})
	g.Go(func() error {
		var e error
		weeklyTrend, e = r.NewGetCustomWeeklyRawMaterialTrend(gCtx, from, to, limit, page)
		return e
	})
	g.Go(func() error {
		var e error
		monthlyTrend, e = r.NewGetCustomMonthlyRawMaterialTrend(gCtx, from, to, limit, page)
		return e
	})
	g.Go(func() error {
		var e error
		yearlyTrend, e = r.NewGetCustomYearlyRawMaterialTrend(gCtx, from, to, limit, page)
		return e
	})
	g.Go(func() error {
		var e error
		topUsedMaterials, e = r.NewGetTopUsedRawMaterials(gCtx, from, to, TopNRawMaterials)
		return e
	})
	g.Go(func() error {
		var e error
		materialUsageBreakdown, e = r.NewGetRawMaterialUsageBreakdown(gCtx, from, to)
		return e
	})
	g.Go(func() error { var e error; peakUsageHours, e = r.NewGetRawMaterialPeakHours(gCtx, from, to); return e })
	g.Go(func() error {
		var e error
		dailyUsageSummary, e = r.NewGetDailyRawMaterialUsage(gCtx, from, to)
		return e
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return &models.NewCustomRangeRawMaterialResponse{
		Overview:               overview,
		StatsCard:              statsCard,
		DailyTrend:             dailyTrend,
		WeeklyTrend:            weeklyTrend,
		MonthlyTrend:           monthlyTrend,
		YearlyTrend:            yearlyTrend,
		TopUsedMaterials:       topUsedMaterials,
		MaterialUsageBreakdown: materialUsageBreakdown,
		PeakUsageHours:         peakUsageHours,
		DailyUsageSummary:      dailyUsageSummary,
	}, nil
}

// ────────────────────────────────────────────────────────────────────────────
// OVERVIEW
//
// The raw_material table has no usage-history column, so:
//   - TotalMaterialUsed  = sum of current quantity (all stock on hand)
//   - TotalInvestment    = sum of quantity * price
//   - TotalOrders        = completed/approved orders in the period
//   - HighestCost*       = material with the highest unit price
//   - MostUsed*          = material with the highest quantity on hand
//     (proxy: more stock ordered → more used)
// ────────────────────────────────────────────────────────────────────────────

func (r *reportRepo) NewGetRawMaterialOverview(ctx context.Context, from, to time.Time) (models.NewRawMaterialOverviewCard, error) {
	query := `
		SELECT
			COALESCE(SUM(quantity), 0)                                  AS total_material_used,
			COALESCE(SUM(quantity * price), 0)                          AS total_investment,
			(
				SELECT COUNT(DISTINCT o.id)
				FROM orders o
				WHERE o.created_at BETWEEN $1::timestamptz AND $2::timestamptz
				  AND o.status IN ('completed', 'approved')
			)                                                           AS total_orders,
			COALESCE(MAX(price), 0)                                     AS highest_cost_material_value,
			COALESCE((SELECT name  FROM raw_material ORDER BY price    DESC LIMIT 1), '') AS highest_cost_material_name,
			COALESCE((SELECT quantity FROM raw_material ORDER BY quantity DESC LIMIT 1), 0) AS most_used_material_quantity,
			COALESCE((SELECT name  FROM raw_material ORDER BY quantity DESC LIMIT 1), '') AS most_used_material_name
		FROM raw_material
	`

	var o models.NewRawMaterialOverviewCard
	if err := r.pool.QueryRow(ctx, query, from, to).Scan(
		&o.TotalMaterialUsed,
		&o.TotalInvestment,
		&o.TotalOrders,
		&o.HighestCostMaterialValue,
		&o.HighestCostMaterialName,
		&o.MostUsedMaterialQuantity,
		&o.MostUsedMaterialName,
	); err != nil {
		return models.NewRawMaterialOverviewCard{}, fmt.Errorf("failed to scan raw material overview: %w", err)
	}
	return o, nil
}

// ────────────────────────────────────────────────────────────────────────────
// STATS CARD (all-time, no date filter)
// ────────────────────────────────────────────────────────────────────────────

func (r *reportRepo) NewGetRawMaterialStatsCard(ctx context.Context) (models.NewRawMaterialStatsCard, error) {
	query := `
		WITH ms AS (
			SELECT
				id, name, price AS unit_cost,
				quantity AS current_stock,
				quantity * price AS stock_value
			FROM raw_material
		)
		SELECT
			COUNT(*)                                                        AS total_materials,
			COALESCE(SUM(current_stock), 0)                                 AS total_current_stock,
			COALESCE(SUM(stock_value), 0)                                   AS total_inventory_value,
			-- no usage history → use current stock as proxy
			COALESCE(SUM(current_stock), 0)                                 AS total_material_used_all_time,
			COALESCE(SUM(stock_value), 0)                                   AS total_investment_all_time,
			COALESCE(MAX(current_stock), 0)                                 AS max_used_quantity,
			COALESCE((SELECT name         FROM ms ORDER BY current_stock DESC LIMIT 1), '') AS most_used_name,
			COALESCE((SELECT current_stock FROM ms ORDER BY current_stock DESC LIMIT 1), 0) AS most_used_qty,
			COALESCE(MAX(unit_cost), 0)                                     AS most_expensive_unit_cost,
			COALESCE((SELECT name FROM ms ORDER BY unit_cost DESC LIMIT 1), '') AS most_expensive_name,
			COALESCE(AVG(stock_value), 0)                                   AS avg_material_value
		FROM ms
	`

	var s models.NewRawMaterialStatsCard
	if err := r.pool.QueryRow(ctx, query).Scan(
		&s.TotalMaterials,
		&s.TotalCurrentStock,
		&s.TotalInventoryValue,
		&s.TotalMaterialUsedAllTime,
		&s.TotalInvestmentAllTime,
		&s.MaxUsedQuantity,
		&s.MostUsedMaterialName,
		&s.MostUsedMaterialQuantity,
		&s.MostExpensiveUnitCost,
		&s.MostExpensiveMaterialName,
		&s.AvgMaterialValue,
	); err != nil {
		return models.NewRawMaterialStatsCard{}, fmt.Errorf("failed to scan raw material stats card: %w", err)
	}
	return s, nil
}

// ────────────────────────────────────────────────────────────────────────────
// FIXED-WINDOW TREND QUERIES
// Scan columns: period, material_used (count added that period), total_cost, orders_count
// ────────────────────────────────────────────────────────────────────────────

func (r *reportRepo) queryRawMaterialTrend(ctx context.Context, query string) ([]models.NewRawMaterialTrendPoint, error) {
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query raw material trend: %w", err)
	}
	defer rows.Close()

	var result []models.NewRawMaterialTrendPoint
	for rows.Next() {
		var pt models.NewRawMaterialTrendPoint
		if err := rows.Scan(&pt.Period, &pt.MaterialUsed, &pt.TotalCost, &pt.OrdersCount); err != nil {
			return nil, fmt.Errorf("failed to scan raw material trend: %w", err)
		}
		result = append(result, pt)
	}
	return result, nil
}

func (r *reportRepo) NewGetLast7DaysRawMaterialTrend(ctx context.Context) ([]models.NewRawMaterialTrendPoint, error) {
	return r.queryRawMaterialTrend(ctx, `
		WITH daily AS (
			SELECT
				DATE(created_at AT TIME ZONE 'Asia/Kathmandu') AS period_date,
				COUNT(*)                                        AS material_count,
				COALESCE(SUM(price * quantity), 0)              AS total_value
			FROM raw_material
			WHERE created_at >= CURRENT_DATE - INTERVAL '7 days'
			GROUP BY DATE(created_at AT TIME ZONE 'Asia/Kathmandu')
		)
		SELECT TO_CHAR(period_date, 'YYYY-MM-DD'), material_count, total_value, 0
		FROM daily ORDER BY period_date ASC`)
}

func (r *reportRepo) NewGetLast7WeeksRawMaterialTrend(ctx context.Context) ([]models.NewRawMaterialTrendPoint, error) {
	return r.queryRawMaterialTrend(ctx, `
		WITH weekly AS (
			SELECT
				DATE_TRUNC('week', created_at AT TIME ZONE 'Asia/Kathmandu') AS period_date,
				COUNT(*)                                                      AS material_count,
				COALESCE(SUM(price * quantity), 0)                            AS total_value
			FROM raw_material
			WHERE created_at >= CURRENT_DATE - INTERVAL '7 weeks'
			GROUP BY DATE_TRUNC('week', created_at AT TIME ZONE 'Asia/Kathmandu')
		)
		SELECT TO_CHAR(period_date, 'IYYY-"W"IW'), material_count, total_value, 0
		FROM weekly ORDER BY period_date ASC`)
}

func (r *reportRepo) NewGetLast7MonthsRawMaterialTrend(ctx context.Context) ([]models.NewRawMaterialTrendPoint, error) {
	return r.queryRawMaterialTrend(ctx, `
		WITH monthly AS (
			SELECT
				DATE_TRUNC('month', created_at AT TIME ZONE 'Asia/Kathmandu') AS period_date,
				COUNT(*)                                                       AS material_count,
				COALESCE(SUM(price * quantity), 0)                             AS total_value
			FROM raw_material
			WHERE created_at >= CURRENT_DATE - INTERVAL '7 months'
			GROUP BY DATE_TRUNC('month', created_at AT TIME ZONE 'Asia/Kathmandu')
		)
		SELECT TO_CHAR(period_date, 'YYYY-MM'), material_count, total_value, 0
		FROM monthly ORDER BY period_date ASC`)
}

func (r *reportRepo) NewGetLast7YearsRawMaterialTrend(ctx context.Context) ([]models.NewRawMaterialTrendPoint, error) {
	return r.queryRawMaterialTrend(ctx, `
		WITH yearly AS (
			SELECT
				DATE_TRUNC('year', created_at AT TIME ZONE 'Asia/Kathmandu') AS period_date,
				COUNT(*)                                                      AS material_count,
				COALESCE(SUM(price * quantity), 0)                            AS total_value
			FROM raw_material
			WHERE created_at >= CURRENT_DATE - INTERVAL '7 years'
			GROUP BY DATE_TRUNC('year', created_at AT TIME ZONE 'Asia/Kathmandu')
		)
		SELECT TO_CHAR(period_date, 'YYYY'), material_count, total_value, 0
		FROM yearly ORDER BY period_date ASC`)
}

// ────────────────────────────────────────────────────────────────────────────
// PAGINATED CUSTOM TREND HELPER
// ────────────────────────────────────────────────────────────────────────────

func (r *reportRepo) scanPaginatedRawMaterialTrend(
	ctx context.Context,
	query string,
	from, to time.Time,
	limit, page, offset, total int,
) (*models.NewRawMaterialPaginatedTrendPoints, error) {
	var raw []byte
	if err := r.pool.QueryRow(ctx, query, from, to, limit, offset).Scan(&raw); err != nil {
		return nil, fmt.Errorf("paginated raw material trend query: %w", err)
	}
	var result []models.NewRawMaterialTrendPoint
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("paginated raw material trend unmarshal: %w", err)
	}
	hasMore := (page+1)*limit < total
	next := page + 1
	if !hasMore {
		next = page
	}
	return &models.NewRawMaterialPaginatedTrendPoints{
		Data: result,
		Pagination: models.NewPaginationInfo{
			Total: total, HasMore: hasMore, NextPage: next, Limit: limit, Page: page,
		},
	}, nil
}

func (r *reportRepo) NewGetCustomDailyRawMaterialTrend(ctx context.Context, from, to time.Time, limit, page int) (*models.NewRawMaterialPaginatedTrendPoints, error) {
	offset := page * limit
	var total int
	if err := r.pool.QueryRow(ctx,
		`SELECT COUNT(DISTINCT DATE(created_at AT TIME ZONE 'Asia/Kathmandu'))
		 FROM raw_material WHERE created_at BETWEEN $1::timestamptz AND $2::timestamptz`,
		from, to,
	).Scan(&total); err != nil {
		return nil, fmt.Errorf("count daily raw material trend: %w", err)
	}
	q := `SELECT COALESCE(json_agg(json_build_object(
			'period',period,'material_used',material_used,'total_cost',total_cost,'orders_count',orders_count
		) ORDER BY period),'[]'::json) FROM (
		SELECT TO_CHAR(DATE(created_at AT TIME ZONE 'Asia/Kathmandu'),'YYYY-MM-DD') AS period,
			COUNT(*)::float AS material_used,
			COALESCE(SUM(price*quantity),0) AS total_cost,
			0 AS orders_count
		FROM raw_material
		WHERE created_at BETWEEN $1::timestamptz AND $2::timestamptz
		GROUP BY DATE(created_at AT TIME ZONE 'Asia/Kathmandu')
		ORDER BY 1 ASC LIMIT $3 OFFSET $4) sub`
	return r.scanPaginatedRawMaterialTrend(ctx, q, from, to, limit, page, offset, total)
}

func (r *reportRepo) NewGetCustomWeeklyRawMaterialTrend(ctx context.Context, from, to time.Time, limit, page int) (*models.NewRawMaterialPaginatedTrendPoints, error) {
	offset := page * limit
	var total int
	if err := r.pool.QueryRow(ctx,
		`SELECT COUNT(DISTINCT DATE_TRUNC('week', created_at AT TIME ZONE 'Asia/Kathmandu'))
		 FROM raw_material WHERE created_at BETWEEN $1::timestamptz AND $2::timestamptz`,
		from, to,
	).Scan(&total); err != nil {
		return nil, fmt.Errorf("count weekly raw material trend: %w", err)
	}
	q := `SELECT COALESCE(json_agg(json_build_object(
			'period',period,'material_used',material_used,'total_cost',total_cost,'orders_count',orders_count
		) ORDER BY week_start),'[]'::json) FROM (
		SELECT DATE_TRUNC('week',created_at AT TIME ZONE 'Asia/Kathmandu') AS week_start,
			TO_CHAR(DATE_TRUNC('week',created_at AT TIME ZONE 'Asia/Kathmandu'),'IYYY-"W"IW') AS period,
			COUNT(*)::float AS material_used,
			COALESCE(SUM(price*quantity),0) AS total_cost,
			0 AS orders_count
		FROM raw_material
		WHERE created_at BETWEEN $1::timestamptz AND $2::timestamptz
		GROUP BY DATE_TRUNC('week',created_at AT TIME ZONE 'Asia/Kathmandu')
		ORDER BY week_start ASC LIMIT $3 OFFSET $4) sub`
	return r.scanPaginatedRawMaterialTrend(ctx, q, from, to, limit, page, offset, total)
}

func (r *reportRepo) NewGetCustomMonthlyRawMaterialTrend(ctx context.Context, from, to time.Time, limit, page int) (*models.NewRawMaterialPaginatedTrendPoints, error) {
	offset := page * limit
	var total int
	if err := r.pool.QueryRow(ctx,
		`SELECT COUNT(DISTINCT DATE_TRUNC('month', created_at AT TIME ZONE 'Asia/Kathmandu'))
		 FROM raw_material WHERE created_at BETWEEN $1::timestamptz AND $2::timestamptz`,
		from, to,
	).Scan(&total); err != nil {
		return nil, fmt.Errorf("count monthly raw material trend: %w", err)
	}
	q := `SELECT COALESCE(json_agg(json_build_object(
			'period',period,'material_used',material_used,'total_cost',total_cost,'orders_count',orders_count
		) ORDER BY month_start),'[]'::json) FROM (
		SELECT DATE_TRUNC('month',created_at AT TIME ZONE 'Asia/Kathmandu') AS month_start,
			TO_CHAR(DATE_TRUNC('month',created_at AT TIME ZONE 'Asia/Kathmandu'),'YYYY-MM') AS period,
			COUNT(*)::float AS material_used,
			COALESCE(SUM(price*quantity),0) AS total_cost,
			0 AS orders_count
		FROM raw_material
		WHERE created_at BETWEEN $1::timestamptz AND $2::timestamptz
		GROUP BY DATE_TRUNC('month',created_at AT TIME ZONE 'Asia/Kathmandu')
		ORDER BY month_start ASC LIMIT $3 OFFSET $4) sub`
	return r.scanPaginatedRawMaterialTrend(ctx, q, from, to, limit, page, offset, total)
}

func (r *reportRepo) NewGetCustomYearlyRawMaterialTrend(ctx context.Context, from, to time.Time, limit, page int) (*models.NewRawMaterialPaginatedTrendPoints, error) {
	offset := page * limit
	var total int
	if err := r.pool.QueryRow(ctx,
		`SELECT COUNT(DISTINCT DATE_TRUNC('year', created_at AT TIME ZONE 'Asia/Kathmandu'))
		 FROM raw_material WHERE created_at BETWEEN $1::timestamptz AND $2::timestamptz`,
		from, to,
	).Scan(&total); err != nil {
		return nil, fmt.Errorf("count yearly raw material trend: %w", err)
	}
	q := `SELECT COALESCE(json_agg(json_build_object(
			'period',period,'material_used',material_used,'total_cost',total_cost,'orders_count',orders_count
		) ORDER BY year_start),'[]'::json) FROM (
		SELECT DATE_TRUNC('year',created_at AT TIME ZONE 'Asia/Kathmandu') AS year_start,
			TO_CHAR(DATE_TRUNC('year',created_at AT TIME ZONE 'Asia/Kathmandu'),'YYYY') AS period,
			COUNT(*)::float AS material_used,
			COALESCE(SUM(price*quantity),0) AS total_cost,
			0 AS orders_count
		FROM raw_material
		WHERE created_at BETWEEN $1::timestamptz AND $2::timestamptz
		GROUP BY DATE_TRUNC('year',created_at AT TIME ZONE 'Asia/Kathmandu')
		ORDER BY year_start ASC LIMIT $3 OFFSET $4) sub`
	return r.scanPaginatedRawMaterialTrend(ctx, q, from, to, limit, page, offset, total)
}

// ────────────────────────────────────────────────────────────────────────────
// TOP USED RAW MATERIALS
//
// FIX: the old query passed $1/$2 (from/to) but never used them in the SQL,
//      causing Postgres error 42P18 "could not determine data type of parameter".
//      The date params are now REMOVED from the function signature at the call
//      site — we don't filter raw_material by date (no usage history table).
//      The function still accepts from/to so the interface stays consistent;
//      they are simply not passed to the DB.
// ────────────────────────────────────────────────────────────────────────────

func (r *reportRepo) NewGetTopUsedRawMaterials(ctx context.Context, from, to time.Time, limit int) ([]models.NewTopUsedRawMaterial, error) {
	// No date parameters passed — raw_material has no usage history, so we
	// rank by current quantity (highest stock = most procured = proxy for usage).
	query := `
		SELECT
			id::text,
			name,
			unit,
			price                       AS unit_cost,
			quantity                    AS total_quantity_used,
			quantity * price            AS total_cost,
			0                           AS affected_orders
		FROM raw_material
		ORDER BY quantity DESC
		LIMIT $1`

	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("top used raw materials: %w", err)
	}
	defer rows.Close()

	var result []models.NewTopUsedRawMaterial
	for rows.Next() {
		var m models.NewTopUsedRawMaterial
		if err := rows.Scan(
			&m.MaterialID, &m.MaterialName, &m.Unit, &m.UnitCost,
			&m.TotalQuantityUsed, &m.TotalCost, &m.AffectedOrders,
		); err != nil {
			return nil, fmt.Errorf("scan top used material: %w", err)
		}
		result = append(result, m)
	}
	return result, nil
}

// ────────────────────────────────────────────────────────────────────────────
// RAW MATERIAL USAGE BREAKDOWN
//
// FIX: same 42P18 error — $1/$2 were passed but not used.
//      Removed date params from the DB call; compute usage_percent correctly
//      using a window function instead of the hardcoded 100.
// ────────────────────────────────────────────────────────────────────────────

func (r *reportRepo) NewGetRawMaterialUsageBreakdown(ctx context.Context, from, to time.Time) ([]models.NewRawMaterialUsageBreakdown, error) {
	// No date parameters sent to DB — raw_material has no time-series usage data.
	query := `
		WITH totals AS (
			SELECT
				NULLIF(SUM(quantity), 0) AS total_qty,
				NULLIF(SUM(quantity * price), 0) AS total_value
			FROM raw_material
		)
		SELECT
			rm.id::text,
			rm.name,
			rm.unit,
			rm.price                        AS unit_cost,
			rm.quantity                     AS current_stock,
			rm.quantity                     AS period_usage,
			rm.quantity * rm.price          AS period_cost,
			COALESCE(rm.quantity / t.total_qty * 100, 0)       AS usage_percent,
			0                               AS orders_count
		FROM raw_material rm
		CROSS JOIN totals t
		ORDER BY rm.quantity DESC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("raw material usage breakdown: %w", err)
	}
	defer rows.Close()

	var result []models.NewRawMaterialUsageBreakdown
	for rows.Next() {
		var m models.NewRawMaterialUsageBreakdown
		if err := rows.Scan(
			&m.MaterialID, &m.MaterialName, &m.Unit, &m.UnitCost,
			&m.CurrentStock, &m.PeriodUsage, &m.PeriodCost,
			&m.UsagePercent, &m.OrdersCount,
		); err != nil {
			return nil, fmt.Errorf("scan usage breakdown: %w", err)
		}
		result = append(result, m)
	}
	return result, nil
}

// ────────────────────────────────────────────────────────────────────────────
// PEAK USAGE HOURS  (based on order activity, proxied from order timestamps)
// ────────────────────────────────────────────────────────────────────────────

func (r *reportRepo) NewGetRawMaterialPeakHours(ctx context.Context, from, to time.Time) ([]models.NewRawMaterialPeakHour, error) {
	query := `
		SELECT
			EXTRACT(HOUR FROM o.created_at AT TIME ZONE 'Asia/Kathmandu')::int  AS hour,
			COUNT(DISTINCT oi.menu_item_id)::float                              AS total_material_used,
			COALESCE(SUM(p.paid_amount), 0)                                     AS total_cost,
			COUNT(DISTINCT o.id)                                                AS orders_count,
			COUNT(DISTINCT oi.menu_item_id)                                     AS unique_items_used
		FROM orders o
		LEFT JOIN order_items oi ON oi.order_id = o.id
		LEFT JOIN payments   p  ON p.order_id  = o.id
		WHERE o.created_at BETWEEN $1::timestamptz AND $2::timestamptz
		  AND o.status IN ('completed', 'approved')
		GROUP BY EXTRACT(HOUR FROM o.created_at AT TIME ZONE 'Asia/Kathmandu')
		ORDER BY hour ASC`

	rows, err := r.pool.Query(ctx, query, from, to)
	if err != nil {
		return nil, fmt.Errorf("raw material peak hours: %w", err)
	}
	defer rows.Close()

	var result []models.NewRawMaterialPeakHour
	for rows.Next() {
		var ph models.NewRawMaterialPeakHour
		if err := rows.Scan(
			&ph.Hour, &ph.TotalMaterialUsed, &ph.TotalCost,
			&ph.OrdersCount, &ph.UniqueItemsUsed,
		); err != nil {
			return nil, fmt.Errorf("scan peak hour: %w", err)
		}
		result = append(result, ph)
	}
	return result, nil
}

// ────────────────────────────────────────────────────────────────────────────
// DAILY RAW MATERIAL USAGE SUMMARY
// ────────────────────────────────────────────────────────────────────────────

func (r *reportRepo) NewGetDailyRawMaterialUsage(ctx context.Context, from, to time.Time) ([]models.NewDailyRawMaterialUsage, error) {
	query := `
		SELECT
			TO_CHAR(DATE(o.created_at AT TIME ZONE 'Asia/Kathmandu'), 'YYYY-MM-DD') AS usage_date,
			COUNT(DISTINCT oi.menu_item_id)::float                                  AS total_material_used,
			COALESCE(SUM(p.paid_amount), 0)                                         AS total_cost,
			COUNT(DISTINCT o.id)                                                    AS orders_count,
			COUNT(DISTINCT oi.menu_item_id)                                         AS unique_materials_used
		FROM orders o
		LEFT JOIN order_items oi ON oi.order_id = o.id
		LEFT JOIN payments   p  ON p.order_id  = o.id
		WHERE o.created_at BETWEEN $1::timestamptz AND $2::timestamptz
		  AND o.status IN ('completed', 'approved')
		GROUP BY DATE(o.created_at AT TIME ZONE 'Asia/Kathmandu')
		ORDER BY usage_date ASC`

	rows, err := r.pool.Query(ctx, query, from, to)
	if err != nil {
		return nil, fmt.Errorf("daily raw material usage: %w", err)
	}
	defer rows.Close()

	var result []models.NewDailyRawMaterialUsage
	for rows.Next() {
		var d models.NewDailyRawMaterialUsage
		if err := rows.Scan(
			&d.UsageDate, &d.TotalMaterialUsed, &d.TotalCost,
			&d.OrdersCount, &d.UniqueMaterialsUsed,
		); err != nil {
			return nil, fmt.Errorf("scan daily usage: %w", err)
		}
		result = append(result, d)
	}
	return result, nil
}

// ────────────────────────────────────────────────────────────────────────────
// DEFAULT REPORT
// ────────────────────────────────────────────────────────────────────────────

func (r *reportRepo) NewGetDefaultStaffReport(ctx context.Context) (*models.NewDefaultStaffResponse, error) {
	to := time.Now()
	from := to.AddDate(0, 0, -30)

	var (
		overview           models.NewStaffOverviewCard
		statsCard          models.NewStaffStatsCard
		dailyTrend         []models.NewStaffTrendPoint
		weeklyTrend        []models.NewStaffTrendPoint
		monthlyTrend       []models.NewStaffTrendPoint
		yearlyTrend        []models.NewStaffTrendPoint
		dailySummary       []models.NewDailyAttendanceSummary
		employeeAttendance []models.NewEmployeeAttendanceSummary
		mostPresent        []models.NewMostPresentEmployee
		mostAbsent         []models.NewMostAbsentEmployee
		longestService     []models.NewLongestServiceEmployee
		roleBreakdown      []models.NewStaffRoleBreakdown
		leaveAnalysis      models.NewLeaveAnalysis
		peakHours          []models.NewStaffPeakHour
		payrollSummary     models.NewPayrollSummary
	)

	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() error { var e error; overview, e = r.NewGetStaffOverview(gCtx, from, to); return e })
	g.Go(func() error { var e error; statsCard, e = r.NewGetStaffStatsCard(gCtx); return e })
	g.Go(func() error { var e error; dailyTrend, e = r.NewGetLast7DaysStaffTrend(gCtx); return e })
	g.Go(func() error { var e error; weeklyTrend, e = r.NewGetLast7WeeksStaffTrend(gCtx); return e })
	g.Go(func() error { var e error; monthlyTrend, e = r.NewGetLast7MonthsStaffTrend(gCtx); return e })
	g.Go(func() error { var e error; yearlyTrend, e = r.NewGetLast7YearsStaffTrend(gCtx); return e })
	g.Go(func() error { var e error; dailySummary, e = r.NewGetStaffDailySummary(gCtx, from, to); return e })
	g.Go(func() error {
		var e error
		employeeAttendance, e = r.NewGetEmployeeAttendanceSummary(gCtx, from, to)
		return e
	})
	g.Go(func() error {
		var e error
		mostPresent, e = r.NewGetMostPresentEmployees(gCtx, from, to, TopNStaff)
		return e
	})
	g.Go(func() error {
		var e error
		mostAbsent, e = r.NewGetMostAbsentEmployees(gCtx, from, to, TopNStaff)
		return e
	})
	g.Go(func() error {
		var e error
		longestService, e = r.NewGetLongestServiceEmployees(gCtx, from, to, TopNStaff)
		return e
	})
	g.Go(func() error { var e error; roleBreakdown, e = r.NewGetStaffRoleBreakdown(gCtx, from, to); return e })
	g.Go(func() error { var e error; leaveAnalysis, e = r.NewGetLeaveAnalysis(gCtx, from, to); return e })
	g.Go(func() error { var e error; peakHours, e = r.NewGetStaffPeakHours(gCtx, from, to); return e })
	g.Go(func() error { var e error; payrollSummary, e = r.NewGetPayrollSummary(gCtx); return e })

	if err := g.Wait(); err != nil {
		return nil, err
	}
	return &models.NewDefaultStaffResponse{
		Overview: overview, StatsCard: statsCard,
		DailyTrend: dailyTrend, WeeklyTrend: weeklyTrend,
		MonthlyTrend: monthlyTrend, YearlyTrend: yearlyTrend,
		DailySummary: dailySummary, EmployeeAttendance: employeeAttendance,
		MostPresentEmployees: mostPresent, MostAbsentEmployees: mostAbsent,
		LongestServiceEmployees: longestService, RoleBreakdown: roleBreakdown,
		LeaveAnalysis: leaveAnalysis, PeakHours: peakHours, PayrollSummary: payrollSummary,
	}, nil
}

// ────────────────────────────────────────────────────────────────────────────
// CUSTOM RANGE REPORT
// ────────────────────────────────────────────────────────────────────────────

func (r *reportRepo) NewGetCustomRangeStaffReport(ctx context.Context, req *models.NewStaffCustomRangeReportRequest) (*models.NewCustomRangeStaffResponse, error) {
	from := req.From
	to := req.To.Truncate(24 * time.Hour).Add(24*time.Hour - time.Second)
	limit := req.Limit
	if limit <= 0 {
		limit = DefaultStaffTrendLimit
	}
	if limit > MaxStaffTrendLimit {
		limit = MaxStaffTrendLimit
	}
	page := req.Page
	if page < 0 {
		page = 0
	}

	var (
		overview           models.NewStaffOverviewCard
		statsCard          models.NewStaffStatsCard
		dailyTrend         *models.NewStaffPaginatedTrendPoints
		weeklyTrend        *models.NewStaffPaginatedTrendPoints
		monthlyTrend       *models.NewStaffPaginatedTrendPoints
		yearlyTrend        *models.NewStaffPaginatedTrendPoints
		dailySummary       []models.NewDailyAttendanceSummary
		employeeAttendance []models.NewEmployeeAttendanceSummary
		mostPresent        []models.NewMostPresentEmployee
		mostAbsent         []models.NewMostAbsentEmployee
		longestService     []models.NewLongestServiceEmployee
		roleBreakdown      []models.NewStaffRoleBreakdown
		leaveAnalysis      models.NewLeaveAnalysis
		peakHours          []models.NewStaffPeakHour
		payrollSummary     models.NewPayrollSummary
	)

	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() error { var e error; overview, e = r.NewGetStaffOverview(gCtx, from, to); return e })
	g.Go(func() error { var e error; statsCard, e = r.NewGetStaffStatsCard(gCtx); return e })
	g.Go(func() error {
		var e error
		dailyTrend, e = r.NewGetCustomDailyStaffTrend(gCtx, from, to, limit, page)
		return e
	})
	g.Go(func() error {
		var e error
		weeklyTrend, e = r.NewGetCustomWeeklyStaffTrend(gCtx, from, to, limit, page)
		return e
	})
	g.Go(func() error {
		var e error
		monthlyTrend, e = r.NewGetCustomMonthlyStaffTrend(gCtx, from, to, limit, page)
		return e
	})
	g.Go(func() error {
		var e error
		yearlyTrend, e = r.NewGetCustomYearlyStaffTrend(gCtx, from, to, limit, page)
		return e
	})
	g.Go(func() error { var e error; dailySummary, e = r.NewGetStaffDailySummary(gCtx, from, to); return e })
	g.Go(func() error {
		var e error
		employeeAttendance, e = r.NewGetEmployeeAttendanceSummary(gCtx, from, to)
		return e
	})
	g.Go(func() error {
		var e error
		mostPresent, e = r.NewGetMostPresentEmployees(gCtx, from, to, TopNStaff)
		return e
	})
	g.Go(func() error {
		var e error
		mostAbsent, e = r.NewGetMostAbsentEmployees(gCtx, from, to, TopNStaff)
		return e
	})
	g.Go(func() error {
		var e error
		longestService, e = r.NewGetLongestServiceEmployees(gCtx, from, to, TopNStaff)
		return e
	})
	g.Go(func() error { var e error; roleBreakdown, e = r.NewGetStaffRoleBreakdown(gCtx, from, to); return e })
	g.Go(func() error { var e error; leaveAnalysis, e = r.NewGetLeaveAnalysis(gCtx, from, to); return e })
	g.Go(func() error { var e error; peakHours, e = r.NewGetStaffPeakHours(gCtx, from, to); return e })
	g.Go(func() error { var e error; payrollSummary, e = r.NewGetPayrollSummary(gCtx); return e })

	if err := g.Wait(); err != nil {
		return nil, err
	}
	return &models.NewCustomRangeStaffResponse{
		Overview: overview, StatsCard: statsCard,
		DailyTrend: dailyTrend, WeeklyTrend: weeklyTrend,
		MonthlyTrend: monthlyTrend, YearlyTrend: yearlyTrend,
		DailySummary: dailySummary, EmployeeAttendance: employeeAttendance,
		MostPresentEmployees: mostPresent, MostAbsentEmployees: mostAbsent,
		LongestServiceEmployees: longestService, RoleBreakdown: roleBreakdown,
		LeaveAnalysis: leaveAnalysis, PeakHours: peakHours, PayrollSummary: payrollSummary,
	}, nil
}

// ────────────────────────────────────────────────────────────────────────────
// OVERVIEW
//
// FIX 1 – work_hours: CASE WHEN check_in IS NULL THEN 0 guards absent/leave
//          rows that have no clock data.  The old COALESCE(check_out, NOW())
//          on a NULL check_in would produce NOW()-NULL = NULL (fine) BUT on
//          some Postgres builds the COALESCE returns NOW() and NULL-NOW()
//          produces a large negative interval.
// FIX 2 – attendance_rate: numerator = present+late+half_day (all showed up).
//          'leave' is excluded from both sides — it is approved absence.
// FIX 3 – late_rate denominator = all "showed up" records (present+late+half_day).
// FIX 4 – avg_work_hours: AVG filtered to check_in IS NOT NULL only.
// ────────────────────────────────────────────────────────────────────────────

func (r *reportRepo) NewGetStaffOverview(ctx context.Context, from, to time.Time) (models.NewStaffOverviewCard, error) {
	query := `
		WITH period_attendance AS (
			SELECT
				a.status,
				a.need_review,
				CASE
					WHEN a.check_in_time IS NULL THEN 0
					ELSE GREATEST(
						EXTRACT(EPOCH FROM (
							COALESCE(a.check_out_time, NOW()) - a.check_in_time
						)) / 3600, 0)
				END AS work_hours,
				a.check_in_time
			FROM attendance a
			WHERE a.work_date BETWEEN $1::date AND $2::date
		),
		counts AS (
			SELECT
				COUNT(*) FILTER (WHERE status = 'present')                        AS present_days,
				COUNT(*) FILTER (WHERE status = 'absent')                         AS absent_days,
				COUNT(*) FILTER (WHERE status = 'late')                           AS late_days,
				COUNT(*) FILTER (WHERE status = 'half_day')                       AS half_days,
				COUNT(*) FILTER (WHERE status = 'leave')                          AS leave_days,
				COUNT(*) FILTER (WHERE need_review = TRUE)                        AS need_review_count,
				COALESCE(SUM(work_hours), 0)                                      AS total_work_hours,
				COALESCE(AVG(work_hours) FILTER (WHERE check_in_time IS NOT NULL), 0) AS avg_work_hours
			FROM period_attendance
		),
		leave_stats AS (
			SELECT
				COUNT(*)                                           AS total_leaves,
				COUNT(*) FILTER (WHERE status = 'approved')       AS approved_leaves
			FROM attendance_leave
			WHERE start_date::date BETWEEN $1::date AND $2::date
		),
		busiest AS (
			SELECT TRIM(TO_CHAR(a.work_date, 'Day')) AS day
			FROM attendance a
			WHERE a.work_date BETWEEN $1::date AND $2::date
			  AND a.status IN ('present','late','half_day')
			GROUP BY TO_CHAR(a.work_date, 'Day')
			ORDER BY COUNT(*) DESC
			LIMIT 1
		),
		peak_hour AS (
			SELECT EXTRACT(HOUR FROM check_in_time AT TIME ZONE 'Asia/Kathmandu')::int AS hour
			FROM attendance
			WHERE work_date BETWEEN $1::date AND $2::date
			  AND check_in_time IS NOT NULL
			GROUP BY EXTRACT(HOUR FROM check_in_time AT TIME ZONE 'Asia/Kathmandu')
			ORDER BY COUNT(*) DESC
			LIMIT 1
		)
		SELECT
			(SELECT COUNT(*) FROM users WHERE role != 'customer')::int             AS total_employees,
			(SELECT COUNT(*) FROM users WHERE role != 'customer' AND is_active)::int AS active_employees,
			c.present_days,
			c.absent_days,
			c.late_days,
			c.half_days,
			c.leave_days,
			-- FIX 2: present+late+half_day counted as "attended"
			COALESCE(
				((c.present_days + c.late_days + c.half_days)::float
				 / NULLIF(c.present_days + c.late_days + c.half_days + c.absent_days, 0)) * 100,
			0) AS overall_attendance_rate,
			-- FIX 3: late_rate = late / all who showed up
			COALESCE(
				(c.late_days::float
				 / NULLIF(c.present_days + c.late_days + c.half_days, 0)) * 100,
			0) AS late_rate,
			COALESCE(
				(ls.approved_leaves::float / NULLIF(ls.total_leaves, 0)) * 100,
			0) AS leave_approval_rate,
			c.total_work_hours,
			c.avg_work_hours,
			COALESCE((SELECT day  FROM busiest),   'Unknown') AS busiest_day,
			COALESCE((SELECT hour FROM peak_hour),  0)        AS peak_attend_hour
		FROM counts c
		CROSS JOIN leave_stats ls
	`

	var o models.NewStaffOverviewCard
	if err := r.pool.QueryRow(ctx, query, from, to).Scan(
		&o.TotalEmployees, &o.ActiveEmployees,
		&o.TotalPresentDays, &o.TotalAbsentDays, &o.TotalLateDays,
		&o.TotalHalfDays, &o.TotalLeaveDays,
		&o.OverallAttendanceRate, &o.LateRate, &o.LeaveApprovalRate,
		&o.TotalWorkHours, &o.AvgWorkHours,
		&o.BusiestDay, &o.PeakAttendHour,
	); err != nil {
		return models.NewStaffOverviewCard{}, fmt.Errorf("failed to scan staff overview: %w", err)
	}
	return o, nil
}

// ────────────────────────────────────────────────────────────────────────────
// STATS CARD  (all-time)
//
// FIX A – The old per_employee CTE did:
//            LEFT JOIN work_hours wh ON wh.employee_id = a.employee_id
//          This is a many-to-many self-join on the same table alias: for an
//          employee with N attendance rows the join produces N² rows, making
//          SUM(hours) inflate by a factor of N.
//          Fixed by computing hours directly in the aggregate — no join.
// FIX B – CASE guard for NULL check_in (same as FIX 1).
// FIX C – longest_service filters WHERE present_days > 0 so employees with
//          zero clocked shifts don't appear with avg_hours = 0.
// ────────────────────────────────────────────────────────────────────────────

func (r *reportRepo) NewGetStaffStatsCard(ctx context.Context) (models.NewStaffStatsCard, error) {
	query := `
		WITH per_employee AS (
			-- FIX A: inline hour calculation, no separate join
			SELECT
				a.employee_id,
				COUNT(*) FILTER (WHERE a.status IN ('present','late','half_day')) AS present_days,
				COUNT(*) FILTER (WHERE a.status = 'absent')                       AS absent_days,
				COALESCE(SUM(
					CASE WHEN a.check_in_time IS NULL THEN 0
					ELSE GREATEST(EXTRACT(EPOCH FROM (COALESCE(a.check_out_time,NOW()) - a.check_in_time))/3600, 0)
					END
				), 0) AS total_hours,
				-- avg over clocked rows only (NULL excluded from AVG automatically)
				COALESCE(AVG(
					CASE WHEN a.check_in_time IS NULL THEN NULL
					ELSE GREATEST(EXTRACT(EPOCH FROM (COALESCE(a.check_out_time,NOW()) - a.check_in_time))/3600, 0)
					END
				), 0) AS avg_hours
			FROM attendance a
			GROUP BY a.employee_id
		),
		most_present AS (
			SELECT pe.employee_id, u.name, pe.present_days
			FROM per_employee pe JOIN users u ON u.id = pe.employee_id
			ORDER BY pe.present_days DESC LIMIT 1
		),
		most_absent AS (
			SELECT pe.employee_id, u.name, pe.absent_days
			FROM per_employee pe JOIN users u ON u.id = pe.employee_id
			ORDER BY pe.absent_days DESC LIMIT 1
		),
		longest_service AS (
			-- FIX C: only employees who actually clocked in at least once
			SELECT pe.employee_id, u.name, pe.avg_hours
			FROM per_employee pe JOIN users u ON u.id = pe.employee_id
			WHERE pe.present_days > 0
			ORDER BY pe.avg_hours DESC LIMIT 1
		),
		leave_counts AS (
			SELECT
				COUNT(*) FILTER (WHERE status = 'pending')  AS pending,
				COUNT(*) FILTER (WHERE status = 'approved') AS approved
			FROM attendance_leave
		),
		totals AS (
			SELECT
				COALESCE(SUM(
					CASE WHEN check_in_time IS NULL THEN 0
					ELSE GREATEST(EXTRACT(EPOCH FROM (COALESCE(check_out_time,NOW()) - check_in_time))/3600, 0)
					END
				), 0) AS all_hours,
				COALESCE(AVG(
					CASE WHEN check_in_time IS NULL THEN NULL
					ELSE GREATEST(EXTRACT(EPOCH FROM (COALESCE(check_out_time,NOW()) - check_in_time))/3600, 0)
					END
				), 0) AS avg_hours
			FROM attendance
		)
		SELECT
			(SELECT COUNT(*) FROM users WHERE role != 'customer')::int            AS total_employees,
			(SELECT COUNT(*) FROM users WHERE role != 'customer' AND is_active)::int AS active_employees,
			(SELECT COUNT(*) FROM attendance)::int                                AS total_records,
			(SELECT all_hours FROM totals)                                        AS all_time_work_hours,
			(SELECT avg_hours FROM totals)                                        AS avg_session_hours,
			COALESCE((SELECT employee_id::text FROM most_present),  '')           AS mp_id,
			COALESCE((SELECT name              FROM most_present),  '')           AS mp_name,
			COALESCE((SELECT present_days      FROM most_present),  0)            AS mp_days,
			COALESCE((SELECT employee_id::text FROM most_absent),   '')           AS ma_id,
			COALESCE((SELECT name              FROM most_absent),   '')           AS ma_name,
			COALESCE((SELECT absent_days       FROM most_absent),   0)            AS ma_days,
			COALESCE((SELECT avg_hours         FROM longest_service),0)           AS ls_avg_hours,
			COALESCE((SELECT employee_id::text FROM longest_service),'')          AS ls_id,
			COALESCE((SELECT name              FROM longest_service),'')          AS ls_name,
			COALESCE((SELECT pending  FROM leave_counts), 0)                      AS pending_leaves,
			COALESCE((SELECT approved FROM leave_counts), 0)                      AS approved_leaves
	`

	var s models.NewStaffStatsCard
	if err := r.pool.QueryRow(ctx, query).Scan(
		&s.TotalEmployees, &s.ActiveEmployees, &s.TotalAttendanceRecords,
		&s.AllTimeWorkHours, &s.AvgSessionHours,
		&s.MostPresentEmployeeID, &s.MostPresentEmployeeName, &s.MostPresentDays,
		&s.MostAbsentEmployeeID, &s.MostAbsentEmployeeName, &s.MostAbsentDays,
		&s.LongestAvgServiceHours, &s.LongestServiceEmployeeID, &s.LongestServiceEmployeeName,
		&s.TotalPendingLeaves, &s.TotalApprovedLeaves,
	); err != nil {
		return models.NewStaffStatsCard{}, fmt.Errorf("failed to scan staff stats card: %w", err)
	}
	return s, nil
}

// ────────────────────────────────────────────────────────────────────────────
// SHARED: fixed-window trend scanner
// attendance_rate = (present+late+half_day) / (present+late+half_day+absent) * 100
// 'leave' excluded from rate.
// work_hours guarded by CASE WHEN check_in IS NULL THEN 0.
// ────────────────────────────────────────────────────────────────────────────

func (r *reportRepo) queryStaffTrend(ctx context.Context, query string) ([]models.NewStaffTrendPoint, error) {
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query staff trend: %w", err)
	}
	defer rows.Close()
	var result []models.NewStaffTrendPoint
	for rows.Next() {
		var pt models.NewStaffTrendPoint
		if err := rows.Scan(&pt.Period, &pt.Present, &pt.Absent, &pt.Late, &pt.HalfDay, &pt.OnLeave, &pt.TotalWorkHours, &pt.AttendanceRate); err != nil {
			return nil, fmt.Errorf("failed to scan staff trend: %w", err)
		}
		result = append(result, pt)
	}
	return result, nil
}

func (r *reportRepo) NewGetLast7DaysStaffTrend(ctx context.Context) ([]models.NewStaffTrendPoint, error) {
	return r.queryStaffTrend(ctx, `
		WITH daily AS (
			SELECT
				a.work_date AS period_date,
				COUNT(*) FILTER (WHERE a.status = 'present')  AS present,
				COUNT(*) FILTER (WHERE a.status = 'absent')   AS absent,
				COUNT(*) FILTER (WHERE a.status = 'late')     AS late,
				COUNT(*) FILTER (WHERE a.status = 'half_day') AS half_day,
				COUNT(*) FILTER (WHERE a.status = 'leave')    AS on_leave,
				COALESCE(SUM(CASE WHEN a.check_in_time IS NULL THEN 0
					ELSE GREATEST(EXTRACT(EPOCH FROM (COALESCE(a.check_out_time,NOW())-a.check_in_time))/3600,0)
					END),0) AS total_work_hours
			FROM attendance a
			WHERE a.work_date >= CURRENT_DATE - INTERVAL '7 days'
			GROUP BY a.work_date
		)
		SELECT TO_CHAR(period_date,'YYYY-MM-DD'), present, absent, late, half_day, on_leave, total_work_hours,
			COALESCE(((present+late+half_day)::float/NULLIF(present+late+half_day+absent,0))*100,0)
		FROM daily ORDER BY period_date ASC`)
}

func (r *reportRepo) NewGetLast7WeeksStaffTrend(ctx context.Context) ([]models.NewStaffTrendPoint, error) {
	return r.queryStaffTrend(ctx, `
		WITH weekly AS (
			SELECT
				DATE_TRUNC('week',a.work_date) AS period_date,
				COUNT(*) FILTER (WHERE a.status = 'present')  AS present,
				COUNT(*) FILTER (WHERE a.status = 'absent')   AS absent,
				COUNT(*) FILTER (WHERE a.status = 'late')     AS late,
				COUNT(*) FILTER (WHERE a.status = 'half_day') AS half_day,
				COUNT(*) FILTER (WHERE a.status = 'leave')    AS on_leave,
				COALESCE(SUM(CASE WHEN a.check_in_time IS NULL THEN 0
					ELSE GREATEST(EXTRACT(EPOCH FROM (COALESCE(a.check_out_time,NOW())-a.check_in_time))/3600,0)
					END),0) AS total_work_hours
			FROM attendance a
			WHERE a.work_date >= CURRENT_DATE - INTERVAL '7 weeks'
			GROUP BY DATE_TRUNC('week',a.work_date)
		)
		SELECT TO_CHAR(period_date,'IYYY-"W"IW'), present, absent, late, half_day, on_leave, total_work_hours,
			COALESCE(((present+late+half_day)::float/NULLIF(present+late+half_day+absent,0))*100,0)
		FROM weekly ORDER BY period_date ASC`)
}

func (r *reportRepo) NewGetLast7MonthsStaffTrend(ctx context.Context) ([]models.NewStaffTrendPoint, error) {
	return r.queryStaffTrend(ctx, `
		WITH monthly AS (
			SELECT
				DATE_TRUNC('month',a.work_date) AS period_date,
				COUNT(*) FILTER (WHERE a.status = 'present')  AS present,
				COUNT(*) FILTER (WHERE a.status = 'absent')   AS absent,
				COUNT(*) FILTER (WHERE a.status = 'late')     AS late,
				COUNT(*) FILTER (WHERE a.status = 'half_day') AS half_day,
				COUNT(*) FILTER (WHERE a.status = 'leave')    AS on_leave,
				COALESCE(SUM(CASE WHEN a.check_in_time IS NULL THEN 0
					ELSE GREATEST(EXTRACT(EPOCH FROM (COALESCE(a.check_out_time,NOW())-a.check_in_time))/3600,0)
					END),0) AS total_work_hours
			FROM attendance a
			WHERE a.work_date >= CURRENT_DATE - INTERVAL '7 months'
			GROUP BY DATE_TRUNC('month',a.work_date)
		)
		SELECT TO_CHAR(period_date,'YYYY-MM'), present, absent, late, half_day, on_leave, total_work_hours,
			COALESCE(((present+late+half_day)::float/NULLIF(present+late+half_day+absent,0))*100,0)
		FROM monthly ORDER BY period_date ASC`)
}

func (r *reportRepo) NewGetLast7YearsStaffTrend(ctx context.Context) ([]models.NewStaffTrendPoint, error) {
	return r.queryStaffTrend(ctx, `
		WITH yearly AS (
			SELECT
				DATE_TRUNC('year',a.work_date) AS period_date,
				COUNT(*) FILTER (WHERE a.status = 'present')  AS present,
				COUNT(*) FILTER (WHERE a.status = 'absent')   AS absent,
				COUNT(*) FILTER (WHERE a.status = 'late')     AS late,
				COUNT(*) FILTER (WHERE a.status = 'half_day') AS half_day,
				COUNT(*) FILTER (WHERE a.status = 'leave')    AS on_leave,
				COALESCE(SUM(CASE WHEN a.check_in_time IS NULL THEN 0
					ELSE GREATEST(EXTRACT(EPOCH FROM (COALESCE(a.check_out_time,NOW())-a.check_in_time))/3600,0)
					END),0) AS total_work_hours
			FROM attendance a
			WHERE a.work_date >= CURRENT_DATE - INTERVAL '7 years'
			GROUP BY DATE_TRUNC('year',a.work_date)
		)
		SELECT TO_CHAR(period_date,'YYYY'), present, absent, late, half_day, on_leave, total_work_hours,
			COALESCE(((present+late+half_day)::float/NULLIF(present+late+half_day+absent,0))*100,0)
		FROM yearly ORDER BY period_date ASC`)
}

// ────────────────────────────────────────────────────────────────────────────
// CUSTOM TREND QUERIES (paginated)
// ────────────────────────────────────────────────────────────────────────────

func (r *reportRepo) scanPaginatedStaffTrend(ctx context.Context, query string, from, to time.Time, limit, page, offset, total int) (*models.NewStaffPaginatedTrendPoints, error) {
	var raw []byte
	if err := r.pool.QueryRow(ctx, query, from, to, limit, offset).Scan(&raw); err != nil {
		return nil, fmt.Errorf("paginated staff trend query: %w", err)
	}
	var result []models.NewStaffTrendPoint
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("paginated staff trend unmarshal: %w", err)
	}
	hasMore := (page+1)*limit < total
	next := page + 1
	if !hasMore {
		next = page
	}
	return &models.NewStaffPaginatedTrendPoints{
		Data:       result,
		Pagination: models.NewPaginationInfo{Total: total, HasMore: hasMore, NextPage: next, Limit: limit, Page: page},
	}, nil
}

// shared inline SQL fragments reused in all four custom trend queries
const trendHoursExpr = `COALESCE(SUM(CASE WHEN a.check_in_time IS NULL THEN 0 ELSE GREATEST(EXTRACT(EPOCH FROM (COALESCE(a.check_out_time,NOW())-a.check_in_time))/3600,0) END),0)`
const trendRateExpr = `COALESCE(((COUNT(*) FILTER(WHERE a.status IN('present','late','half_day')))::float / NULLIF(COUNT(*) FILTER(WHERE a.status IN('present','late','half_day','absent')),0))*100,0)`

func (r *reportRepo) NewGetCustomDailyStaffTrend(ctx context.Context, from, to time.Time, limit, page int) (*models.NewStaffPaginatedTrendPoints, error) {
	offset := page * limit
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(DISTINCT work_date) FROM attendance WHERE work_date BETWEEN $1::date AND $2::date`, from, to).Scan(&total); err != nil {
		return nil, fmt.Errorf("count daily staff trend: %w", err)
	}
	q := `SELECT COALESCE(json_agg(json_build_object('period',period,'present',present,'absent',absent,'late',late,'half_day',half_day,'on_leave',on_leave,'total_work_hours',total_work_hours,'attendance_rate',attendance_rate) ORDER BY period),'[]'::json) FROM (
		SELECT TO_CHAR(a.work_date,'YYYY-MM-DD') AS period,
			COUNT(*) FILTER(WHERE a.status='present') AS present, COUNT(*) FILTER(WHERE a.status='absent') AS absent,
			COUNT(*) FILTER(WHERE a.status='late') AS late, COUNT(*) FILTER(WHERE a.status='half_day') AS half_day,
			COUNT(*) FILTER(WHERE a.status='leave') AS on_leave,` + trendHoursExpr + ` AS total_work_hours,` + trendRateExpr + ` AS attendance_rate
		FROM attendance a WHERE a.work_date BETWEEN $1::date AND $2::date GROUP BY a.work_date ORDER BY a.work_date ASC LIMIT $3 OFFSET $4) sub`
	return r.scanPaginatedStaffTrend(ctx, q, from, to, limit, page, offset, total)
}

func (r *reportRepo) NewGetCustomWeeklyStaffTrend(ctx context.Context, from, to time.Time, limit, page int) (*models.NewStaffPaginatedTrendPoints, error) {
	offset := page * limit
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(DISTINCT DATE_TRUNC('week',work_date)) FROM attendance WHERE work_date BETWEEN $1::date AND $2::date`, from, to).Scan(&total); err != nil {
		return nil, fmt.Errorf("count weekly staff trend: %w", err)
	}
	q := `SELECT COALESCE(json_agg(json_build_object('period',period,'present',present,'absent',absent,'late',late,'half_day',half_day,'on_leave',on_leave,'total_work_hours',total_work_hours,'attendance_rate',attendance_rate) ORDER BY week_start),'[]'::json) FROM (
		SELECT DATE_TRUNC('week',a.work_date) AS week_start, TO_CHAR(DATE_TRUNC('week',a.work_date),'IYYY-"W"IW') AS period,
			COUNT(*) FILTER(WHERE a.status='present') AS present, COUNT(*) FILTER(WHERE a.status='absent') AS absent,
			COUNT(*) FILTER(WHERE a.status='late') AS late, COUNT(*) FILTER(WHERE a.status='half_day') AS half_day,
			COUNT(*) FILTER(WHERE a.status='leave') AS on_leave,` + trendHoursExpr + ` AS total_work_hours,` + trendRateExpr + ` AS attendance_rate
		FROM attendance a WHERE a.work_date BETWEEN $1::date AND $2::date GROUP BY DATE_TRUNC('week',a.work_date) ORDER BY week_start ASC LIMIT $3 OFFSET $4) sub`
	return r.scanPaginatedStaffTrend(ctx, q, from, to, limit, page, offset, total)
}

func (r *reportRepo) NewGetCustomMonthlyStaffTrend(ctx context.Context, from, to time.Time, limit, page int) (*models.NewStaffPaginatedTrendPoints, error) {
	offset := page * limit
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(DISTINCT DATE_TRUNC('month',work_date)) FROM attendance WHERE work_date BETWEEN $1::date AND $2::date`, from, to).Scan(&total); err != nil {
		return nil, fmt.Errorf("count monthly staff trend: %w", err)
	}
	q := `SELECT COALESCE(json_agg(json_build_object('period',period,'present',present,'absent',absent,'late',late,'half_day',half_day,'on_leave',on_leave,'total_work_hours',total_work_hours,'attendance_rate',attendance_rate) ORDER BY month_start),'[]'::json) FROM (
		SELECT DATE_TRUNC('month',a.work_date) AS month_start, TO_CHAR(DATE_TRUNC('month',a.work_date),'YYYY-MM') AS period,
			COUNT(*) FILTER(WHERE a.status='present') AS present, COUNT(*) FILTER(WHERE a.status='absent') AS absent,
			COUNT(*) FILTER(WHERE a.status='late') AS late, COUNT(*) FILTER(WHERE a.status='half_day') AS half_day,
			COUNT(*) FILTER(WHERE a.status='leave') AS on_leave,` + trendHoursExpr + ` AS total_work_hours,` + trendRateExpr + ` AS attendance_rate
		FROM attendance a WHERE a.work_date BETWEEN $1::date AND $2::date GROUP BY DATE_TRUNC('month',a.work_date) ORDER BY month_start ASC LIMIT $3 OFFSET $4) sub`
	return r.scanPaginatedStaffTrend(ctx, q, from, to, limit, page, offset, total)
}

func (r *reportRepo) NewGetCustomYearlyStaffTrend(ctx context.Context, from, to time.Time, limit, page int) (*models.NewStaffPaginatedTrendPoints, error) {
	offset := page * limit
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(DISTINCT DATE_TRUNC('year',work_date)) FROM attendance WHERE work_date BETWEEN $1::date AND $2::date`, from, to).Scan(&total); err != nil {
		return nil, fmt.Errorf("count yearly staff trend: %w", err)
	}
	q := `SELECT COALESCE(json_agg(json_build_object('period',period,'present',present,'absent',absent,'late',late,'half_day',half_day,'on_leave',on_leave,'total_work_hours',total_work_hours,'attendance_rate',attendance_rate) ORDER BY year_start),'[]'::json) FROM (
		SELECT DATE_TRUNC('year',a.work_date) AS year_start, TO_CHAR(DATE_TRUNC('year',a.work_date),'YYYY') AS period,
			COUNT(*) FILTER(WHERE a.status='present') AS present, COUNT(*) FILTER(WHERE a.status='absent') AS absent,
			COUNT(*) FILTER(WHERE a.status='late') AS late, COUNT(*) FILTER(WHERE a.status='half_day') AS half_day,
			COUNT(*) FILTER(WHERE a.status='leave') AS on_leave,` + trendHoursExpr + ` AS total_work_hours,` + trendRateExpr + ` AS attendance_rate
		FROM attendance a WHERE a.work_date BETWEEN $1::date AND $2::date GROUP BY DATE_TRUNC('year',a.work_date) ORDER BY year_start ASC LIMIT $3 OFFSET $4) sub`
	return r.scanPaginatedStaffTrend(ctx, q, from, to, limit, page, offset, total)
}

// ────────────────────────────────────────────────────────────────────────────
// DAILY SUMMARY
// ────────────────────────────────────────────────────────────────────────────

func (r *reportRepo) NewGetStaffDailySummary(ctx context.Context, from, to time.Time) ([]models.NewDailyAttendanceSummary, error) {
	query := `
		SELECT
			TO_CHAR(a.work_date,'YYYY-MM-DD'),
			COUNT(*) FILTER(WHERE a.status='present'),
			COUNT(*) FILTER(WHERE a.status='absent'),
			COUNT(*) FILTER(WHERE a.status='late'),
			COUNT(*) FILTER(WHERE a.status='half_day'),
			COUNT(*) FILTER(WHERE a.status='leave'),
			COALESCE(SUM(CASE WHEN a.check_in_time IS NULL THEN 0
				ELSE GREATEST(EXTRACT(EPOCH FROM (COALESCE(a.check_out_time,NOW())-a.check_in_time))/3600,0)
				END),0),
			COALESCE(((COUNT(*) FILTER(WHERE a.status IN('present','late','half_day')))::float
				/NULLIF(COUNT(*) FILTER(WHERE a.status IN('present','late','half_day','absent')),0))*100,0)
		FROM attendance a
		WHERE a.work_date BETWEEN $1::date AND $2::date
		GROUP BY a.work_date ORDER BY a.work_date ASC`

	rows, err := r.pool.Query(ctx, query, from, to)
	if err != nil {
		return nil, fmt.Errorf("daily summary: %w", err)
	}
	defer rows.Close()
	var result []models.NewDailyAttendanceSummary
	for rows.Next() {
		var d models.NewDailyAttendanceSummary
		if err := rows.Scan(&d.WorkDate, &d.Present, &d.Absent, &d.Late, &d.HalfDay, &d.OnLeave, &d.TotalWorkHours, &d.AttendanceRate); err != nil {
			return nil, fmt.Errorf("scan daily summary: %w", err)
		}
		result = append(result, d)
	}
	return result, nil
}

// ────────────────────────────────────────────────────────────────────────────
// PER-EMPLOYEE SUMMARY + RAW RECORDS
// ────────────────────────────────────────────────────────────────────────────

func (r *reportRepo) NewGetEmployeeAttendanceSummary(ctx context.Context, from, to time.Time) ([]models.NewEmployeeAttendanceSummary, error) {
	aggQuery := `
		SELECT
			u.id::text, u.name, u.email, u.phone, u.image, u.role::text, u.gender::text,
			COUNT(*) FILTER(WHERE a.status='present')  AS present_days,
			COUNT(*) FILTER(WHERE a.status='absent')   AS absent_days,
			COUNT(*) FILTER(WHERE a.status='late')     AS late_days,
			COUNT(*) FILTER(WHERE a.status='half_day') AS half_days,
			COUNT(*) FILTER(WHERE a.status='leave')    AS leave_days,
			COALESCE(SUM(CASE WHEN a.check_in_time IS NULL THEN 0
				ELSE GREATEST(EXTRACT(EPOCH FROM (COALESCE(a.check_out_time,NOW())-a.check_in_time))/3600,0)
				END),0) AS total_work_hours,
			COALESCE(AVG(CASE WHEN a.check_in_time IS NULL THEN NULL
				ELSE GREATEST(EXTRACT(EPOCH FROM (COALESCE(a.check_out_time,NOW())-a.check_in_time))/3600,0)
				END),0) AS avg_work_hours,
			COALESCE(((COUNT(*) FILTER(WHERE a.status IN('present','late','half_day')))::float
				/NULLIF(COUNT(*) FILTER(WHERE a.status IN('present','late','half_day','absent')),0))*100,0) AS attendance_rate,
			COUNT(*) FILTER(WHERE a.need_review=TRUE) AS need_review_count
		FROM users u
		LEFT JOIN attendance a ON a.employee_id=u.id AND a.work_date BETWEEN $1::date AND $2::date
		WHERE u.role != 'customer'
		GROUP BY u.id,u.name,u.email,u.phone,u.image,u.role,u.gender
		ORDER BY present_days DESC, u.name ASC`

	aggRows, err := r.pool.Query(ctx, aggQuery, from, to)
	if err != nil {
		return nil, fmt.Errorf("employee attendance summary: %w", err)
	}
	defer aggRows.Close()

	var result []models.NewEmployeeAttendanceSummary
	idxByID := make(map[string]int)
	for aggRows.Next() {
		var e models.NewEmployeeAttendanceSummary
		if err := aggRows.Scan(
			&e.EmployeeID, &e.EmployeeName, &e.Email, &e.Phone, &e.Image, &e.Role, &e.Gender,
			&e.PresentDays, &e.AbsentDays, &e.LateDays, &e.HalfDays, &e.LeaveDays,
			&e.TotalWorkHours, &e.AvgWorkHours, &e.AttendanceRate, &e.NeedReviewCount,
		); err != nil {
			return nil, fmt.Errorf("scan employee attendance summary: %w", err)
		}
		idxByID[e.EmployeeID] = len(result)
		result = append(result, e)
	}
	if err := aggRows.Err(); err != nil {
		return nil, fmt.Errorf("row error employee attendance summary: %w", err)
	}

	recQuery := `
		SELECT
			a.id::text,
			TO_CHAR(a.work_date,'YYYY-MM-DD'),
			a.status::text,
			TO_CHAR(a.check_in_time  AT TIME ZONE 'Asia/Kathmandu','YYYY-MM-DD"T"HH24:MI:SS'),
			TO_CHAR(a.check_out_time AT TIME ZONE 'Asia/Kathmandu','YYYY-MM-DD"T"HH24:MI:SS'),
			CASE WHEN a.check_in_time IS NULL THEN 0
				ELSE GREATEST(EXTRACT(EPOCH FROM (COALESCE(a.check_out_time,NOW())-a.check_in_time))/3600,0)
			END AS work_hours,
			a.need_review,
			u.id::text, u.name, u.email, u.phone, u.image, u.role::text, u.gender::text
		FROM attendance a
		JOIN users u ON u.id=a.employee_id
		WHERE a.work_date BETWEEN $1::date AND $2::date AND u.role != 'customer'
		ORDER BY a.work_date ASC, u.name ASC`

	recRows, err := r.pool.Query(ctx, recQuery, from, to)
	if err != nil {
		return nil, fmt.Errorf("attendance records: %w", err)
	}
	defer recRows.Close()
	for recRows.Next() {
		var rec models.NewEmployeeAttendanceRecord
		if err := recRows.Scan(
			&rec.AttendanceID, &rec.WorkDate, &rec.Status,
			&rec.CheckInTime, &rec.CheckOutTime, &rec.WorkHours, &rec.NeedReview,
			&rec.EmployeeID, &rec.EmployeeName, &rec.Email, &rec.Phone, &rec.Image, &rec.Role, &rec.Gender,
		); err != nil {
			return nil, fmt.Errorf("scan attendance record: %w", err)
		}
		if idx, ok := idxByID[rec.EmployeeID]; ok {
			result[idx].Records = append(result[idx].Records, rec)
		}
	}
	if err := recRows.Err(); err != nil {
		return nil, fmt.Errorf("row error attendance records: %w", err)
	}
	return result, nil
}

// ────────────────────────────────────────────────────────────────────────────
// MOST PRESENT / ABSENT / LONGEST SERVICE
// ────────────────────────────────────────────────────────────────────────────

func (r *reportRepo) NewGetMostPresentEmployees(ctx context.Context, from, to time.Time, limit int) ([]models.NewMostPresentEmployee, error) {
	query := `
		SELECT u.id::text,u.name,u.email,u.phone,u.image,u.role::text,u.gender::text,
			COUNT(*) FILTER(WHERE a.status IN('present','late','half_day')) AS present_days,
			COALESCE(SUM(CASE WHEN a.check_in_time IS NULL THEN 0
				ELSE GREATEST(EXTRACT(EPOCH FROM (COALESCE(a.check_out_time,NOW())-a.check_in_time))/3600,0)
				END),0) AS total_work_hours,
			COALESCE(((COUNT(*) FILTER(WHERE a.status IN('present','late','half_day')))::float
				/NULLIF(COUNT(*) FILTER(WHERE a.status IN('present','late','half_day','absent')),0))*100,0)
		FROM users u JOIN attendance a ON a.employee_id=u.id AND a.work_date BETWEEN $1::date AND $2::date
		WHERE u.role != 'customer'
		GROUP BY u.id,u.name,u.email,u.phone,u.image,u.role,u.gender
		ORDER BY present_days DESC LIMIT $3`
	rows, err := r.pool.Query(ctx, query, from, to, limit)
	if err != nil {
		return nil, fmt.Errorf("most present: %w", err)
	}
	defer rows.Close()
	var result []models.NewMostPresentEmployee
	for rows.Next() {
		var e models.NewMostPresentEmployee
		if err := rows.Scan(&e.EmployeeID, &e.EmployeeName, &e.Email, &e.Phone, &e.Image, &e.Role, &e.Gender, &e.PresentDays, &e.TotalWorkHours, &e.AttendanceRate); err != nil {
			return nil, fmt.Errorf("scan most present: %w", err)
		}
		result = append(result, e)
	}
	return result, nil
}

func (r *reportRepo) NewGetMostAbsentEmployees(ctx context.Context, from, to time.Time, limit int) ([]models.NewMostAbsentEmployee, error) {
	query := `
		SELECT u.id::text,u.name,u.email,u.phone,u.image,u.role::text,u.gender::text,
			COUNT(*) FILTER(WHERE a.status='absent') AS absent_days,
			COALESCE(SUM(CASE WHEN a.check_in_time IS NULL THEN 0
				ELSE GREATEST(EXTRACT(EPOCH FROM (COALESCE(a.check_out_time,NOW())-a.check_in_time))/3600,0)
				END),0) AS total_work_hours,
			COALESCE(((COUNT(*) FILTER(WHERE a.status IN('present','late','half_day')))::float
				/NULLIF(COUNT(*) FILTER(WHERE a.status IN('present','late','half_day','absent')),0))*100,0)
		FROM users u JOIN attendance a ON a.employee_id=u.id AND a.work_date BETWEEN $1::date AND $2::date
		WHERE u.role != 'customer'
		GROUP BY u.id,u.name,u.email,u.phone,u.image,u.role,u.gender
		ORDER BY absent_days DESC LIMIT $3`
	rows, err := r.pool.Query(ctx, query, from, to, limit)
	if err != nil {
		return nil, fmt.Errorf("most absent: %w", err)
	}
	defer rows.Close()
	var result []models.NewMostAbsentEmployee
	for rows.Next() {
		var e models.NewMostAbsentEmployee
		if err := rows.Scan(&e.EmployeeID, &e.EmployeeName, &e.Email, &e.Phone, &e.Image, &e.Role, &e.Gender, &e.AbsentDays, &e.TotalWorkHours, &e.AttendanceRate); err != nil {
			return nil, fmt.Errorf("scan most absent: %w", err)
		}
		result = append(result, e)
	}
	return result, nil
}

func (r *reportRepo) NewGetLongestServiceEmployees(ctx context.Context, from, to time.Time, limit int) ([]models.NewLongestServiceEmployee, error) {
	// avg_shift_hours = mean of (check_out - check_in) for rows that have clock data.
	// Ordered by avg_shift_hours DESC so the employee who stays longest per shift tops the list.
	// WHERE check_in IS NOT NULL ensures employees with only absent/leave records are excluded.
	query := `
		SELECT u.id::text,u.name,u.email,u.phone,u.image,u.role::text,u.gender::text,
			COALESCE(SUM(CASE WHEN a.check_in_time IS NULL THEN 0
				ELSE GREATEST(EXTRACT(EPOCH FROM (COALESCE(a.check_out_time,NOW())-a.check_in_time))/3600,0)
				END),0) AS total_work_hours,
			-- AVG with NULL for unchecked rows → only clocked rows contribute
			COALESCE(AVG(CASE WHEN a.check_in_time IS NULL THEN NULL
				ELSE GREATEST(EXTRACT(EPOCH FROM (COALESCE(a.check_out_time,NOW())-a.check_in_time))/3600,0)
				END),0) AS avg_shift_hours,
			COUNT(*) FILTER(WHERE a.status IN('present','late','half_day')) AS present_days,
			COALESCE(((COUNT(*) FILTER(WHERE a.status IN('present','late','half_day')))::float
				/NULLIF(COUNT(*) FILTER(WHERE a.status IN('present','late','half_day','absent')),0))*100,0)
		FROM users u
		JOIN attendance a ON a.employee_id=u.id AND a.work_date BETWEEN $1::date AND $2::date
		WHERE u.role != 'customer'
		  AND a.check_in_time IS NOT NULL
		GROUP BY u.id,u.name,u.email,u.phone,u.image,u.role,u.gender
		ORDER BY avg_shift_hours DESC LIMIT $3`
	rows, err := r.pool.Query(ctx, query, from, to, limit)
	if err != nil {
		return nil, fmt.Errorf("longest service: %w", err)
	}
	defer rows.Close()
	var result []models.NewLongestServiceEmployee
	for rows.Next() {
		var e models.NewLongestServiceEmployee
		if err := rows.Scan(&e.EmployeeID, &e.EmployeeName, &e.Email, &e.Phone, &e.Image, &e.Role, &e.Gender, &e.TotalWorkHours, &e.AvgShiftHours, &e.PresentDays, &e.AttendanceRate); err != nil {
			return nil, fmt.Errorf("scan longest service: %w", err)
		}
		result = append(result, e)
	}
	return result, nil
}

// ────────────────────────────────────────────────────────────────────────────
// ROLE BREAKDOWN
// ────────────────────────────────────────────────────────────────────────────

func (r *reportRepo) NewGetStaffRoleBreakdown(ctx context.Context, from, to time.Time) ([]models.NewStaffRoleBreakdown, error) {
	query := `
		WITH per_role AS (
			SELECT
				u.role::text AS role,
				COUNT(DISTINCT u.id) AS employee_count,
				COALESCE(SUM(u.salary),0) AS total_salary,
				COALESCE(AVG(u.salary),0) AS avg_salary,
				COALESCE(SUM(CASE WHEN a.check_in_time IS NULL THEN 0
					ELSE GREATEST(EXTRACT(EPOCH FROM (COALESCE(a.check_out_time,NOW())-a.check_in_time))/3600,0)
					END),0) AS total_work_hours,
				COALESCE(AVG(CASE WHEN a.check_in_time IS NULL THEN NULL
					ELSE GREATEST(EXTRACT(EPOCH FROM (COALESCE(a.check_out_time,NOW())-a.check_in_time))/3600,0)
					END),0) AS avg_work_hours,
				COALESCE(((COUNT(*) FILTER(WHERE a.status IN('present','late','half_day')))::float
					/NULLIF(COUNT(*) FILTER(WHERE a.status IN('present','late','half_day','absent')),0))*100,0) AS attendance_rate
			FROM users u
			LEFT JOIN attendance a ON a.employee_id=u.id AND a.work_date BETWEEN $1::date AND $2::date
			WHERE u.role != 'customer'
			GROUP BY u.role
		),
		total_emp AS (SELECT COUNT(*) AS cnt FROM users WHERE role != 'customer')
		SELECT pr.role,pr.employee_count,pr.total_work_hours,pr.avg_work_hours,pr.attendance_rate,
			pr.total_salary,pr.avg_salary,
			COALESCE((pr.employee_count::float/NULLIF(t.cnt,0))*100,0)
		FROM per_role pr, total_emp t ORDER BY pr.employee_count DESC`
	rows, err := r.pool.Query(ctx, query, from, to)
	if err != nil {
		return nil, fmt.Errorf("role breakdown: %w", err)
	}
	defer rows.Close()
	var result []models.NewStaffRoleBreakdown
	for rows.Next() {
		var rb models.NewStaffRoleBreakdown
		if err := rows.Scan(&rb.Role, &rb.EmployeeCount, &rb.TotalWorkHours, &rb.AvgWorkHours, &rb.AttendanceRate, &rb.Salary, &rb.AvgSalary, &rb.Percent); err != nil {
			return nil, fmt.Errorf("scan role breakdown: %w", err)
		}
		result = append(result, rb)
	}
	return result, nil
}

// ────────────────────────────────────────────────────────────────────────────
// LEAVE ANALYSIS
//
// FIX: avg_leave_days now counts inclusively: Mon–Fri = 5 days.
//      Old formula gave Mon–Fri = 4 (off by 1). Added +1 and GREATEST guard.
// ────────────────────────────────────────────────────────────────────────────

func (r *reportRepo) NewGetLeaveAnalysis(ctx context.Context, from, to time.Time) (models.NewLeaveAnalysis, error) {
	var la models.NewLeaveAnalysis
	if err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*),
			COUNT(*) FILTER(WHERE status='pending'),
			COUNT(*) FILTER(WHERE status='approved'),
			COUNT(*) FILTER(WHERE status='rejected'),
			COALESCE((COUNT(*) FILTER(WHERE status='approved')::float/NULLIF(COUNT(*),0))*100,0),
			COALESCE(AVG(GREATEST(EXTRACT(EPOCH FROM(end_date-start_date))/86400+1,0)) FILTER(WHERE status='approved'),0)
		FROM attendance_leave WHERE start_date::date BETWEEN $1::date AND $2::date`, from, to,
	).Scan(&la.TotalRequests, &la.PendingCount, &la.ApprovedCount, &la.RejectedCount, &la.ApprovalRate, &la.AvgLeaveDays); err != nil {
		return models.NewLeaveAnalysis{}, fmt.Errorf("leave analysis: %w", err)
	}

	rows, err := r.pool.Query(ctx, `
		SELECT u.id::text,u.name,u.role::text,COUNT(*) AS leave_count,
			COALESCE(SUM(GREATEST(EXTRACT(EPOCH FROM(al.end_date-al.start_date))/86400+1,0)),0)
		FROM attendance_leave al JOIN users u ON u.id=al.employee_id
		WHERE al.start_date::date BETWEEN $1::date AND $2::date
		GROUP BY u.id,u.name,u.role ORDER BY leave_count DESC LIMIT 10`, from, to)
	if err != nil {
		return models.NewLeaveAnalysis{}, fmt.Errorf("top leave employees: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var tl models.NewTopLeaveEmployee
		if err := rows.Scan(&tl.EmployeeID, &tl.EmployeeName, &tl.Role, &tl.LeaveCount, &tl.TotalDays); err != nil {
			return models.NewLeaveAnalysis{}, fmt.Errorf("scan top leave: %w", err)
		}
		la.TopLeaveEmployees = append(la.TopLeaveEmployees, tl)
	}
	return la, nil
}

// ────────────────────────────────────────────────────────────────────────────
// PEAK HOURS
// ────────────────────────────────────────────────────────────────────────────

func (r *reportRepo) NewGetStaffPeakHours(ctx context.Context, from, to time.Time) ([]models.NewStaffPeakHour, error) {
	query := `
		SELECT
			EXTRACT(HOUR FROM a.check_in_time AT TIME ZONE 'Asia/Kathmandu')::int AS hour,
			COUNT(*) AS check_ins,
			COUNT(*) FILTER(WHERE a.check_out_time IS NOT NULL) AS check_outs,
			COUNT(DISTINCT a.employee_id) AS active_staff,
			COALESCE(AVG(CASE WHEN a.check_in_time IS NULL THEN NULL
				ELSE GREATEST(EXTRACT(EPOCH FROM (COALESCE(a.check_out_time,NOW())-a.check_in_time))/3600,0)
				END),0) AS avg_work_hours
		FROM attendance a
		WHERE a.work_date BETWEEN $1::date AND $2::date AND a.check_in_time IS NOT NULL
		GROUP BY EXTRACT(HOUR FROM a.check_in_time AT TIME ZONE 'Asia/Kathmandu')
		ORDER BY hour ASC`
	rows, err := r.pool.Query(ctx, query, from, to)
	if err != nil {
		return nil, fmt.Errorf("staff peak hours: %w", err)
	}
	defer rows.Close()
	var result []models.NewStaffPeakHour
	for rows.Next() {
		var ph models.NewStaffPeakHour
		if err := rows.Scan(&ph.Hour, &ph.CheckIns, &ph.CheckOuts, &ph.ActiveStaff, &ph.AvgWorkHours); err != nil {
			return nil, fmt.Errorf("scan peak hour: %w", err)
		}
		result = append(result, ph)
	}
	return result, nil
}

// ────────────────────────────────────────────────────────────────────────────
// PAYROLL SUMMARY
// ────────────────────────────────────────────────────────────────────────────

func (r *reportRepo) NewGetPayrollSummary(ctx context.Context) (models.NewPayrollSummary, error) {
	var ps models.NewPayrollSummary
	if err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(salary),0),COUNT(*),COALESCE(AVG(salary),0) FROM users WHERE role!='customer'`,
	).Scan(&ps.TotalMonthlySalary, &ps.TotalEmployees, &ps.AvgSalary); err != nil {
		return models.NewPayrollSummary{}, fmt.Errorf("payroll summary: %w", err)
	}
	rows, err := r.pool.Query(ctx, `
		WITH per_role AS (
			SELECT role::text,COUNT(*) AS ec,COALESCE(SUM(salary),0) AS ts,COALESCE(AVG(salary),0) AS as_
			FROM users WHERE role!='customer' GROUP BY role
		), total_emp AS (SELECT COUNT(*) AS cnt FROM users WHERE role!='customer')
		SELECT pr.role,pr.ec,0::float,0::float,0::float,pr.ts,pr.as_,
			COALESCE((pr.ec::float/NULLIF(t.cnt,0))*100,0)
		FROM per_role pr,total_emp t ORDER BY pr.ts DESC`)
	if err != nil {
		return models.NewPayrollSummary{}, fmt.Errorf("payroll by role: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var rb models.NewStaffRoleBreakdown
		if err := rows.Scan(&rb.Role, &rb.EmployeeCount, &rb.TotalWorkHours, &rb.AvgWorkHours, &rb.AttendanceRate, &rb.Salary, &rb.AvgSalary, &rb.Percent); err != nil {
			return models.NewPayrollSummary{}, fmt.Errorf("scan payroll role: %w", err)
		}
		ps.ByRole = append(ps.ByRole, rb)
	}
	return ps, nil
}

// ────────────────────────────────────────────────────────────────────────────────
// DEFAULT TABLE REPORT
// ────────────────────────────────────────────────────────────────────────────────

func (r *reportRepo) NewGetDefaultTableReport(ctx context.Context) (*models.NewDefaultTableResponse, error) {
	to := time.Now()
	from := to.AddDate(0, 0, -30)

	var (
		overview            models.NewTableOverviewCard
		statsCard           models.NewTableStatsCard
		dailyTrend          []models.NewTableTrendPoint
		weeklyTrend         []models.NewTableTrendPoint
		monthlyTrend        []models.NewTableTrendPoint
		yearlyTrend         []models.NewTableTrendPoint
		topTables           []models.NewTopTable
		tableUsageBreakdown []models.NewTableUsageBreakdown
		peakHours           []models.NewTablePeakHour
		occupancyRate       []models.NewOccupancyRate
		avgSessionDuration  float64
	)

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		var err error
		overview, err = r.NewGetTableOverview(gCtx, from, to)
		if err != nil {
			return fmt.Errorf("table overview: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		var err error
		statsCard, err = r.NewGetTableStatsCard(gCtx)
		if err != nil {
			return fmt.Errorf("stats card: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		var err error
		dailyTrend, err = r.NewGetLast7DaysTableTrend(gCtx)
		if err != nil {
			return fmt.Errorf("daily trend: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		var err error
		weeklyTrend, err = r.NewGetLast7WeeksTableTrend(gCtx)
		if err != nil {
			return fmt.Errorf("weekly trend: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		var err error
		monthlyTrend, err = r.NewGetLast7MonthsTableTrend(gCtx)
		if err != nil {
			return fmt.Errorf("monthly trend: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		var err error
		yearlyTrend, err = r.NewGetLast7YearsTableTrend(gCtx)
		if err != nil {
			return fmt.Errorf("yearly trend: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		var err error
		topTables, err = r.NewGetTopTables(gCtx, from, to, 10)
		if err != nil {
			return fmt.Errorf("top tables: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		var err error
		tableUsageBreakdown, err = r.NewGetTableUsageBreakdown(gCtx, from, to)
		if err != nil {
			return fmt.Errorf("table usage breakdown: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		var err error
		peakHours, err = r.NewGetTablePeakHours(gCtx, from, to)
		if err != nil {
			return fmt.Errorf("peak hours: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		var err error
		occupancyRate, err = r.NewGetOccupancyRate(gCtx, from, to)
		if err != nil {
			return fmt.Errorf("occupancy rate: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		var err error
		avgSessionDuration, err = r.NewGetAvgSessionDuration(gCtx, from, to)
		if err != nil {
			return fmt.Errorf("avg session duration: %w", err)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return &models.NewDefaultTableResponse{
		Overview:            overview,
		StatsCard:           statsCard,
		DailyTrend:          dailyTrend,
		WeeklyTrend:         weeklyTrend,
		MonthlyTrend:        monthlyTrend,
		YearlyTrend:         yearlyTrend,
		TopTables:           topTables,
		TableUsageBreakdown: tableUsageBreakdown,
		PeakHours:           peakHours,
		OccupancyRate:       occupancyRate,
		AvgSessionDuration:  avgSessionDuration,
	}, nil
}

// ────────────────────────────────────────────────────────────────────────────
// CUSTOM RANGE TABLE REPORT
// ────────────────────────────────────────────────────────────────────────────

func (r *reportRepo) NewGetCustomRangeTableReport(ctx context.Context, req *models.NewTableCustomRangeReportRequest) (*models.NewCustomRangeTableResponse, error) {
	from := req.From
	to := req.To.Truncate(24 * time.Hour).Add(24*time.Hour - time.Second)

	limit := req.Limit
	if limit <= 0 {
		limit = DefaultTableTrendLimit
	}
	if limit > MaxTableTrendLimit {
		limit = MaxTableTrendLimit
	}
	page := req.Page
	if page < 0 {
		page = 0
	}

	var (
		overview            models.NewTableOverviewCard
		statsCard           models.NewTableStatsCard
		dailyTrend          *models.NewTablePaginatedTrendPoints
		weeklyTrend         *models.NewTablePaginatedTrendPoints
		monthlyTrend        *models.NewTablePaginatedTrendPoints
		yearlyTrend         *models.NewTablePaginatedTrendPoints
		topTables           []models.NewTopTable
		tableUsageBreakdown []models.NewTableUsageBreakdown
		peakHours           []models.NewTablePeakHour
		occupancyRate       []models.NewOccupancyRate
		avgSessionDuration  float64
	)

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		var err error
		overview, err = r.NewGetTableOverview(gCtx, from, to)
		return err
	})
	g.Go(func() error {
		var err error
		statsCard, err = r.NewGetTableStatsCard(gCtx)
		return err
	})
	g.Go(func() error {
		var err error
		dailyTrend, err = r.NewGetCustomDailyTableTrend(gCtx, from, to, limit, page)
		return err
	})
	g.Go(func() error {
		var err error
		weeklyTrend, err = r.NewGetCustomWeeklyTableTrend(gCtx, from, to, limit, page)
		return err
	})
	g.Go(func() error {
		var err error
		monthlyTrend, err = r.NewGetCustomMonthlyTableTrend(gCtx, from, to, limit, page)
		return err
	})
	g.Go(func() error {
		var err error
		yearlyTrend, err = r.NewGetCustomYearlyTableTrend(gCtx, from, to, limit, page)
		return err
	})
	g.Go(func() error {
		var err error
		topTables, err = r.NewGetTopTables(gCtx, from, to, 10)
		return err
	})
	g.Go(func() error {
		var err error
		tableUsageBreakdown, err = r.NewGetTableUsageBreakdown(gCtx, from, to)
		return err
	})
	g.Go(func() error {
		var err error
		peakHours, err = r.NewGetTablePeakHours(gCtx, from, to)
		return err
	})
	g.Go(func() error {
		var err error
		occupancyRate, err = r.NewGetOccupancyRate(gCtx, from, to)
		return err
	})
	g.Go(func() error {
		var err error
		avgSessionDuration, err = r.NewGetAvgSessionDuration(gCtx, from, to)
		return err
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return &models.NewCustomRangeTableResponse{
		Overview:            overview,
		StatsCard:           statsCard,
		DailyTrend:          dailyTrend,
		WeeklyTrend:         weeklyTrend,
		MonthlyTrend:        monthlyTrend,
		YearlyTrend:         yearlyTrend,
		TopTables:           topTables,
		TableUsageBreakdown: tableUsageBreakdown,
		PeakHours:           peakHours,
		OccupancyRate:       occupancyRate,
		AvgSessionDuration:  avgSessionDuration,
	}, nil
}

// ────────────────────────────────────────────────────────────────────────────
// TREND QUERIES — FIXED
// BUG 1 FIX: avg_occupancy was AVG(session_duration_minutes).
// Now it is: (COUNT(DISTINCT table_number) / total_tables) * 100
// giving a real 0–100 occupancy percentage per period.
// ────────────────────────────────────────────────────────────────────────────

func (r *reportRepo) NewGetLast7DaysTableTrend(ctx context.Context) ([]models.NewTableTrendPoint, error) {
	// FIX BUG 1: avg_occupancy is now real occupancy % not avg duration
	query := `
		WITH total_tables AS (
			SELECT COUNT(*) AS cnt FROM table_status
		),
		daily_stats AS (
			SELECT
				DATE(ts.open_time AT TIME ZONE 'Asia/Kathmandu') AS period,
				COUNT(DISTINCT ts.id)                             AS total_sessions,
				COALESCE(SUM(p.paid_amount), 0)                  AS total_revenue,
				COUNT(DISTINCT ts.table_number)                  AS active_tables
			FROM table_session ts
			LEFT JOIN orders   o ON ts.id = o.table_session_id
			LEFT JOIN payments p ON o.id  = p.order_id
			WHERE ts.open_time >= NOW() - INTERVAL '7 days'
			GROUP BY DATE(ts.open_time AT TIME ZONE 'Asia/Kathmandu')
		)
		SELECT
			TO_CHAR(d.period, 'YYYY-MM-DD')                                          AS period,
			d.total_sessions,
			LEAST(COALESCE((d.active_tables::float / NULLIF(t.cnt, 0)) * 100, 0), 100) AS avg_occupancy,
			d.total_revenue
		FROM daily_stats d, total_tables t
		ORDER BY d.period ASC
	`
	return r.queryTableTrend(ctx, query)
}

func (r *reportRepo) NewGetLast7WeeksTableTrend(ctx context.Context) ([]models.NewTableTrendPoint, error) {
	query := `
		WITH total_tables AS (
			SELECT COUNT(*) AS cnt FROM table_status
		),
		weekly_stats AS (
			SELECT
				DATE_TRUNC('week', ts.open_time AT TIME ZONE 'Asia/Kathmandu') AS period,
				COUNT(DISTINCT ts.id)                                           AS total_sessions,
				COALESCE(SUM(p.paid_amount), 0)                                AS total_revenue,
				COUNT(DISTINCT ts.table_number)                                AS active_tables
			FROM table_session ts
			LEFT JOIN orders   o ON ts.id = o.table_session_id
			LEFT JOIN payments p ON o.id  = p.order_id
			WHERE ts.open_time >= NOW() - INTERVAL '7 weeks'
			GROUP BY DATE_TRUNC('week', ts.open_time AT TIME ZONE 'Asia/Kathmandu')
		)
		SELECT
			TO_CHAR(w.period, 'IYYY-"W"IW')                                          AS period,
			w.total_sessions,
			LEAST(COALESCE((w.active_tables::float / NULLIF(t.cnt, 0)) * 100, 0), 100) AS avg_occupancy,
			w.total_revenue
		FROM weekly_stats w, total_tables t
		ORDER BY w.period ASC
	`
	return r.queryTableTrend(ctx, query)
}

func (r *reportRepo) NewGetLast7MonthsTableTrend(ctx context.Context) ([]models.NewTableTrendPoint, error) {
	query := `
		WITH total_tables AS (
			SELECT COUNT(*) AS cnt FROM table_status
		),
		monthly_stats AS (
			SELECT
				DATE_TRUNC('month', ts.open_time AT TIME ZONE 'Asia/Kathmandu') AS period,
				COUNT(DISTINCT ts.id)                                            AS total_sessions,
				COALESCE(SUM(p.paid_amount), 0)                                 AS total_revenue,
				COUNT(DISTINCT ts.table_number)                                 AS active_tables
			FROM table_session ts
			LEFT JOIN orders   o ON ts.id = o.table_session_id
			LEFT JOIN payments p ON o.id  = p.order_id
			WHERE ts.open_time >= NOW() - INTERVAL '7 months'
			GROUP BY DATE_TRUNC('month', ts.open_time AT TIME ZONE 'Asia/Kathmandu')
		)
		SELECT
			TO_CHAR(m.period, 'YYYY-MM')                                              AS period,
			m.total_sessions,
			LEAST(COALESCE((m.active_tables::float / NULLIF(t.cnt, 0)) * 100, 0), 100) AS avg_occupancy,
			m.total_revenue
		FROM monthly_stats m, total_tables t
		ORDER BY m.period ASC
	`
	return r.queryTableTrend(ctx, query)
}

func (r *reportRepo) NewGetLast7YearsTableTrend(ctx context.Context) ([]models.NewTableTrendPoint, error) {
	query := `
		WITH total_tables AS (
			SELECT COUNT(*) AS cnt FROM table_status
		),
		yearly_stats AS (
			SELECT
				DATE_TRUNC('year', ts.open_time AT TIME ZONE 'Asia/Kathmandu') AS period,
				COUNT(DISTINCT ts.id)                                           AS total_sessions,
				COALESCE(SUM(p.paid_amount), 0)                                AS total_revenue,
				COUNT(DISTINCT ts.table_number)                                AS active_tables
			FROM table_session ts
			LEFT JOIN orders   o ON ts.id = o.table_session_id
			LEFT JOIN payments p ON o.id  = p.order_id
			WHERE ts.open_time >= NOW() - INTERVAL '7 years'
			GROUP BY DATE_TRUNC('year', ts.open_time AT TIME ZONE 'Asia/Kathmandu')
		)
		SELECT
			TO_CHAR(y.period, 'YYYY')                                                 AS period,
			y.total_sessions,
			LEAST(COALESCE((y.active_tables::float / NULLIF(t.cnt, 0)) * 100, 0), 100) AS avg_occupancy,
			y.total_revenue
		FROM yearly_stats y, total_tables t
		ORDER BY y.period ASC
	`
	return r.queryTableTrend(ctx, query)
}

func (r *reportRepo) queryTableTrend(ctx context.Context, query string) ([]models.NewTableTrendPoint, error) {
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query table trend: %w", err)
	}
	defer rows.Close()

	var result []models.NewTableTrendPoint
	for rows.Next() {
		var point models.NewTableTrendPoint
		if err := rows.Scan(&point.Period, &point.TotalSessions, &point.AvgOccupancy, &point.TotalRevenue); err != nil {
			return nil, fmt.Errorf("failed to scan table trend: %w", err)
		}
		result = append(result, point)
	}
	return result, nil
}

// ────────────────────────────────────────────────────────────────────────────
// CUSTOM TREND QUERIES — FIXED (same BUG 1 fix applied)
// ────────────────────────────────────────────────────────────────────────────

func (r *reportRepo) NewGetCustomDailyTableTrend(ctx context.Context, from, to time.Time, limit, page int) (*models.NewTablePaginatedTrendPoints, error) {
	offset := page * limit

	var total int
	countQuery := `
		SELECT COUNT(DISTINCT DATE(open_time AT TIME ZONE 'Asia/Kathmandu'))
		FROM table_session
		WHERE open_time BETWEEN $1 AND $2
	`
	if err := r.pool.QueryRow(ctx, countQuery, from, to).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count custom daily table trend: %w", err)
	}

	// FIX BUG 1: avg_occupancy = real occupancy % per day
	query := `
		SELECT COALESCE(
			json_agg(
				json_build_object(
					'period',         period,
					'total_sessions', total_sessions,
					'avg_occupancy',  avg_occupancy,
					'total_revenue',  total_revenue
				)
				ORDER BY period ASC
			), '[]'::json
		)
		FROM (
			SELECT
				TO_CHAR(DATE(ts.open_time AT TIME ZONE 'Asia/Kathmandu'), 'YYYY-MM-DD') AS period,
				COUNT(DISTINCT ts.id)                                                     AS total_sessions,
				LEAST(
					COALESCE(
						(COUNT(DISTINCT ts.table_number)::float
						 / NULLIF((SELECT COUNT(*) FROM table_status), 0)) * 100,
					0), 100)                                                              AS avg_occupancy,
				COALESCE(SUM(p.paid_amount), 0)                                          AS total_revenue
			FROM table_session ts
			LEFT JOIN orders   o ON ts.id = o.table_session_id
			LEFT JOIN payments p ON o.id  = p.order_id
			WHERE ts.open_time BETWEEN $1 AND $2
			GROUP BY DATE(ts.open_time AT TIME ZONE 'Asia/Kathmandu')
			ORDER BY DATE(ts.open_time AT TIME ZONE 'Asia/Kathmandu') ASC
			LIMIT $3 OFFSET $4
		) sub
	`

	var resultJSON []byte
	if err := r.pool.QueryRow(ctx, query, from, to, limit, offset).Scan(&resultJSON); err != nil {
		return nil, fmt.Errorf("failed to query custom daily table trend: %w", err)
	}

	var result []models.NewTableTrendPoint
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal custom daily table trend: %w", err)
	}

	hasMore := (page+1)*limit < total
	nextPage := page + 1
	if !hasMore {
		nextPage = page
	}

	return &models.NewTablePaginatedTrendPoints{
		Data: result,
		Pagination: models.NewPaginationInfo{
			Total:    total,
			HasMore:  hasMore,
			NextPage: nextPage,
			Limit:    limit,
			Page:     page,
		},
	}, nil
}

func (r *reportRepo) NewGetCustomWeeklyTableTrend(ctx context.Context, from, to time.Time, limit, page int) (*models.NewTablePaginatedTrendPoints, error) {
	offset := page * limit

	var total int
	countQuery := `
		SELECT COUNT(DISTINCT DATE_TRUNC('week', open_time AT TIME ZONE 'Asia/Kathmandu'))
		FROM table_session
		WHERE open_time BETWEEN $1 AND $2
	`
	if err := r.pool.QueryRow(ctx, countQuery, from, to).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count custom weekly table trend: %w", err)
	}

	query := `
		SELECT COALESCE(
			json_agg(
				json_build_object(
					'period',         period,
					'total_sessions', total_sessions,
					'avg_occupancy',  avg_occupancy,
					'total_revenue',  total_revenue
				)
				ORDER BY week_start ASC
			), '[]'::json
		)
		FROM (
			SELECT
				TO_CHAR(DATE_TRUNC('week', ts.open_time AT TIME ZONE 'Asia/Kathmandu'), 'IYYY-"W"IW') AS period,
				DATE_TRUNC('week', ts.open_time AT TIME ZONE 'Asia/Kathmandu')                        AS week_start,
				COUNT(DISTINCT ts.id)                                                                  AS total_sessions,
				LEAST(
					COALESCE(
						(COUNT(DISTINCT ts.table_number)::float
						 / NULLIF((SELECT COUNT(*) FROM table_status), 0)) * 100,
					0), 100)                                                                            AS avg_occupancy,
				COALESCE(SUM(p.paid_amount), 0)                                                        AS total_revenue
			FROM table_session ts
			LEFT JOIN orders   o ON ts.id = o.table_session_id
			LEFT JOIN payments p ON o.id  = p.order_id
			WHERE ts.open_time BETWEEN $1 AND $2
			GROUP BY DATE_TRUNC('week', ts.open_time AT TIME ZONE 'Asia/Kathmandu')
			ORDER BY week_start ASC
			LIMIT $3 OFFSET $4
		) sub
	`

	var resultJSON []byte
	if err := r.pool.QueryRow(ctx, query, from, to, limit, offset).Scan(&resultJSON); err != nil {
		return nil, fmt.Errorf("failed to query custom weekly table trend: %w", err)
	}

	var result []models.NewTableTrendPoint
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal custom weekly table trend: %w", err)
	}

	hasMore := (page+1)*limit < total
	nextPage := page + 1
	if !hasMore {
		nextPage = page
	}

	return &models.NewTablePaginatedTrendPoints{
		Data: result,
		Pagination: models.NewPaginationInfo{
			Total:    total,
			HasMore:  hasMore,
			NextPage: nextPage,
			Limit:    limit,
			Page:     page,
		},
	}, nil
}

func (r *reportRepo) NewGetCustomMonthlyTableTrend(ctx context.Context, from, to time.Time, limit, page int) (*models.NewTablePaginatedTrendPoints, error) {
	offset := page * limit

	var total int
	countQuery := `
		SELECT COUNT(DISTINCT DATE_TRUNC('month', open_time AT TIME ZONE 'Asia/Kathmandu'))
		FROM table_session
		WHERE open_time BETWEEN $1 AND $2
	`
	if err := r.pool.QueryRow(ctx, countQuery, from, to).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count custom monthly table trend: %w", err)
	}

	query := `
		SELECT COALESCE(
			json_agg(
				json_build_object(
					'period',         period,
					'total_sessions', total_sessions,
					'avg_occupancy',  avg_occupancy,
					'total_revenue',  total_revenue
				)
				ORDER BY month_start ASC
			), '[]'::json
		)
		FROM (
			SELECT
				TO_CHAR(DATE_TRUNC('month', ts.open_time AT TIME ZONE 'Asia/Kathmandu'), 'YYYY-MM') AS period,
				DATE_TRUNC('month', ts.open_time AT TIME ZONE 'Asia/Kathmandu')                     AS month_start,
				COUNT(DISTINCT ts.id)                                                                AS total_sessions,
				LEAST(
					COALESCE(
						(COUNT(DISTINCT ts.table_number)::float
						 / NULLIF((SELECT COUNT(*) FROM table_status), 0)) * 100,
					0), 100)                                                                         AS avg_occupancy,
				COALESCE(SUM(p.paid_amount), 0)                                                     AS total_revenue
			FROM table_session ts
			LEFT JOIN orders   o ON ts.id = o.table_session_id
			LEFT JOIN payments p ON o.id  = p.order_id
			WHERE ts.open_time BETWEEN $1 AND $2
			GROUP BY DATE_TRUNC('month', ts.open_time AT TIME ZONE 'Asia/Kathmandu')
			ORDER BY month_start ASC
			LIMIT $3 OFFSET $4
		) sub
	`

	var resultJSON []byte
	if err := r.pool.QueryRow(ctx, query, from, to, limit, offset).Scan(&resultJSON); err != nil {
		return nil, fmt.Errorf("failed to query custom monthly table trend: %w", err)
	}

	var result []models.NewTableTrendPoint
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal custom monthly table trend: %w", err)
	}

	hasMore := (page+1)*limit < total
	nextPage := page + 1
	if !hasMore {
		nextPage = page
	}

	return &models.NewTablePaginatedTrendPoints{
		Data: result,
		Pagination: models.NewPaginationInfo{
			Total:    total,
			HasMore:  hasMore,
			NextPage: nextPage,
			Limit:    limit,
			Page:     page,
		},
	}, nil
}

func (r *reportRepo) NewGetCustomYearlyTableTrend(ctx context.Context, from, to time.Time, limit, page int) (*models.NewTablePaginatedTrendPoints, error) {
	offset := page * limit

	var total int
	countQuery := `
		SELECT COUNT(DISTINCT DATE_TRUNC('year', open_time AT TIME ZONE 'Asia/Kathmandu'))
		FROM table_session
		WHERE open_time BETWEEN $1 AND $2
	`
	if err := r.pool.QueryRow(ctx, countQuery, from, to).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count custom yearly table trend: %w", err)
	}

	query := `
		SELECT COALESCE(
			json_agg(
				json_build_object(
					'period',         period,
					'total_sessions', total_sessions,
					'avg_occupancy',  avg_occupancy,
					'total_revenue',  total_revenue
				)
				ORDER BY year_start ASC
			), '[]'::json
		)
		FROM (
			SELECT
				TO_CHAR(DATE_TRUNC('year', ts.open_time AT TIME ZONE 'Asia/Kathmandu'), 'YYYY') AS period,
				DATE_TRUNC('year', ts.open_time AT TIME ZONE 'Asia/Kathmandu')                  AS year_start,
				COUNT(DISTINCT ts.id)                                                            AS total_sessions,
				LEAST(
					COALESCE(
						(COUNT(DISTINCT ts.table_number)::float
						 / NULLIF((SELECT COUNT(*) FROM table_status), 0)) * 100,
					0), 100)                                                                     AS avg_occupancy,
				COALESCE(SUM(p.paid_amount), 0)                                                 AS total_revenue
			FROM table_session ts
			LEFT JOIN orders   o ON ts.id = o.table_session_id
			LEFT JOIN payments p ON o.id  = p.order_id
			WHERE ts.open_time BETWEEN $1 AND $2
			GROUP BY DATE_TRUNC('year', ts.open_time AT TIME ZONE 'Asia/Kathmandu')
			ORDER BY year_start ASC
			LIMIT $3 OFFSET $4
		) sub
	`

	var resultJSON []byte
	if err := r.pool.QueryRow(ctx, query, from, to, limit, offset).Scan(&resultJSON); err != nil {
		return nil, fmt.Errorf("failed to query custom yearly table trend: %w", err)
	}

	var result []models.NewTableTrendPoint
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal custom yearly table trend: %w", err)
	}

	hasMore := (page+1)*limit < total
	nextPage := page + 1
	if !hasMore {
		nextPage = page
	}

	return &models.NewTablePaginatedTrendPoints{
		Data: result,
		Pagination: models.NewPaginationInfo{
			Total:    total,
			HasMore:  hasMore,
			NextPage: nextPage,
			Limit:    limit,
			Page:     page,
		},
	}, nil
}

// ────────────────────────────────────────────────────────────────────────────
// OVERVIEW — FIXED
// BUG 2 FIX: avg_occupancy_rate was AVG(session_duration_minutes).
// Now it is: (COUNT(DISTINCT table_number) / total_tables) * 100
// ────────────────────────────────────────────────────────────────────────────

func (r *reportRepo) NewGetTableOverview(ctx context.Context, from, to time.Time) (models.NewTableOverviewCard, error) {
	query := `
		WITH total_tables AS (
			SELECT COUNT(*) AS cnt FROM table_status
		),
		table_stats AS (
			SELECT
				(SELECT COUNT(*) FROM table_status)                    AS total_tables,
				-- active_tables = currently occupied right now (not period-scoped)
				(SELECT COUNT(*) FROM table_status WHERE status = 'occupied') AS active_tables,
				COUNT(DISTINCT ts.id)                                  AS total_sessions,
				-- FIX BUG 2: real avg session duration in minutes (used only for AvgSessionDuration)
				COALESCE(AVG(
					GREATEST(
						EXTRACT(EPOCH FROM (
							COALESCE(ts.close_time, NOW()) - ts.open_time
						)) / 60,
					0)
				), 0) AS avg_session_duration,
				COALESCE(SUM(p.paid_amount), 0) AS total_revenue,
				-- FIX BUG 2: real occupancy % = distinct active tables / total tables * 100
				LEAST(
					COALESCE(
						(COUNT(DISTINCT ts.table_number)::float
						 / NULLIF((SELECT COUNT(*) FROM table_status), 0)) * 100,
					0), 100)                                           AS avg_occupancy_rate
			FROM table_session ts
			LEFT JOIN orders   o ON ts.id = o.table_session_id
			LEFT JOIN payments p ON o.id  = p.order_id
			WHERE ts.open_time BETWEEN $1 AND $2
		),
		peak_hour AS (
			SELECT
				EXTRACT(HOUR FROM ts.open_time AT TIME ZONE 'Asia/Kathmandu') AS hour,
				COUNT(DISTINCT ts.table_number)                               AS active_tables_in_hour
			FROM table_session ts
			WHERE ts.open_time BETWEEN $1 AND $2
			GROUP BY EXTRACT(HOUR FROM ts.open_time AT TIME ZONE 'Asia/Kathmandu')
			ORDER BY active_tables_in_hour DESC
			LIMIT 1
		)
		SELECT
			ts.total_tables,
			ts.active_tables,
			ts.total_sessions,
			ts.avg_occupancy_rate,
			ts.total_revenue,
			ts.avg_session_duration,
			COALESCE(ph.hour::int, 0)                                          AS peak_occupancy_hour,
			-- peak occupancy rate = active tables at peak hour / total tables * 100, capped at 100
			LEAST(
				COALESCE(
					(ph.active_tables_in_hour::float / NULLIF(t.cnt, 0)) * 100,
				0), 100)                                                        AS peak_occupancy_rate
		FROM table_stats ts
		CROSS JOIN total_tables t
		LEFT JOIN peak_hour ph ON TRUE
	`

	var overview models.NewTableOverviewCard
	if err := r.pool.QueryRow(ctx, query, from, to).Scan(
		&overview.TotalTables,
		&overview.ActiveTables,
		&overview.TotalSessions,
		&overview.AvgOccupancyRate,
		&overview.TotalTableRevenue,
		&overview.AvgSessionDuration,
		&overview.PeakOccupancyHour,
		&overview.PeakOccupancyRate,
	); err != nil {
		return models.NewTableOverviewCard{}, fmt.Errorf("failed to scan table overview: %w", err)
	}

	return overview, nil
}

// ────────────────────────────────────────────────────────────────────────────
// STATS CARD — unchanged, no bugs
// ────────────────────────────────────────────────────────────────────────────

func (r *reportRepo) NewGetTableStatsCard(ctx context.Context) (models.NewTableStatsCard, error) {
	query := `
		SELECT
			(SELECT COUNT(*) FROM table_status)                AS total_tables,
			(SELECT COALESCE(SUM(capacity), 0) FROM table_status) AS total_capacity,
			COUNT(DISTINCT ts.id)                              AS total_sessions_all_time,
			COALESCE(SUM(p.paid_amount), 0)                   AS total_table_revenue,
			-- Keep avg_session_duration here as actual minutes — this is for display, not a %
			COALESCE(AVG(
				GREATEST(
					EXTRACT(EPOCH FROM (ts.close_time - ts.open_time)) / 60,
				0)
			), 0)                                              AS avg_session_duration,
			COALESCE((
				SELECT ts2.table_number
				FROM table_session ts2
				GROUP BY ts2.table_number
				ORDER BY COUNT(*) DESC
				LIMIT 1
			), 0)                                              AS most_used_table,
			COALESCE((
				SELECT COUNT(*)
				FROM table_session ts2
				GROUP BY ts2.table_number
				ORDER BY COUNT(*) DESC
				LIMIT 1
			), 0)                                              AS most_used_table_count,
			COALESCE((
				SELECT TRIM(TO_CHAR(DATE(ts2.open_time), 'Day'))
				FROM table_session ts2
				GROUP BY DATE(ts2.open_time)
				ORDER BY COUNT(*) DESC
				LIMIT 1
			), 'Unknown')                                      AS busiest_day
		FROM table_session ts
		LEFT JOIN orders   o ON ts.id = o.table_session_id
		LEFT JOIN payments p ON o.id  = p.order_id
	`

	var stats models.NewTableStatsCard
	if err := r.pool.QueryRow(ctx, query).Scan(
		&stats.TotalTables,
		&stats.TotalCapacity,
		&stats.TotalSessionsAllTime,
		&stats.TotalTableRevenue,
		&stats.AvgSessionDuration,
		&stats.MostUsedTable,
		&stats.MostUsedTableCount,
		&stats.BusiestDay,
	); err != nil {
		return models.NewTableStatsCard{}, fmt.Errorf("failed to scan table stats card: %w", err)
	}

	return stats, nil
}

// ────────────────────────────────────────────────────────────────────────────
// TOP TABLES — unchanged, no bugs
// ────────────────────────────────────────────────────────────────────────────

func (r *reportRepo) NewGetTopTables(ctx context.Context, from, to time.Time, limit int) ([]models.NewTopTable, error) {
	query := `
		SELECT
			ts.table_number,
			COALESCE(tstat.capacity, 0)                                     AS capacity,
			COUNT(DISTINCT ts.id)                                           AS total_sessions,
			COALESCE(SUM(p.paid_amount), 0)                                 AS total_revenue,
			COALESCE(AVG(
				GREATEST(
					EXTRACT(EPOCH FROM (ts.close_time - ts.open_time)) / 60,
				0)
			), 0)                                                           AS avg_session_time,
			LEAST(
				COALESCE(
					(COUNT(DISTINCT ts.id)::float
					 / NULLIF((SELECT COUNT(*) FROM table_session WHERE open_time BETWEEN $1 AND $2), 0)) * 100,
				0), 100)                                                    AS occupancy_rate,
			COUNT(DISTINCT o.customer_phone)                                AS total_customers
		FROM table_session ts
		LEFT JOIN table_status tstat ON ts.table_number = tstat.table_number
		LEFT JOIN orders       o     ON ts.id = o.table_session_id
		LEFT JOIN payments     p     ON o.id  = p.order_id
		WHERE ts.open_time BETWEEN $1 AND $2
		GROUP BY ts.table_number, tstat.capacity
		ORDER BY total_sessions DESC
		LIMIT $3
	`

	rows, err := r.pool.Query(ctx, query, from, to, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query top tables: %w", err)
	}
	defer rows.Close()

	var result []models.NewTopTable
	for rows.Next() {
		var table models.NewTopTable
		if err := rows.Scan(
			&table.TableNumber,
			&table.Capacity,
			&table.TotalSessions,
			&table.TotalRevenue,
			&table.AvgSessionTime,
			&table.OccupancyRate,
			&table.TotalCustomers,
		); err != nil {
			return nil, fmt.Errorf("failed to scan top table: %w", err)
		}
		result = append(result, table)
	}

	return result, nil
}

// ────────────────────────────────────────────────────────────────────────────
// TABLE USAGE BREAKDOWN — FIXED
// BUG 3 FIX: total_hours_used was going negative due to timezone mismatch.
// GREATEST(..., 0) clamps the subtraction result so it can never be negative.
// ────────────────────────────────────────────────────────────────────────────

func (r *reportRepo) NewGetTableUsageBreakdown(ctx context.Context, from, to time.Time) ([]models.NewTableUsageBreakdown, error) {
	query := `
		WITH table_stats AS (
			SELECT
				ts.table_number,
				COALESCE(tstat.capacity, 0) AS capacity,
				COUNT(DISTINCT ts.id)       AS total_sessions,
				-- FIX BUG 3: GREATEST(..., 0) prevents negative hours from tz-offset bugs
				COALESCE(SUM(
					GREATEST(
						EXTRACT(EPOCH FROM (
							COALESCE(ts.close_time, NOW()) - ts.open_time
						)) / 3600,
					0)
				), 0)                       AS total_hours_used,
				COALESCE(SUM(p.paid_amount), 0)  AS total_revenue,
				COALESCE(AVG(p.paid_amount), 0)  AS avg_order_value
			FROM table_session ts
			LEFT JOIN table_status tstat ON ts.table_number = tstat.table_number
			LEFT JOIN orders       o     ON ts.id = o.table_session_id
			LEFT JOIN payments     p     ON o.id  = p.order_id
			WHERE ts.open_time BETWEEN $1 AND $2
			GROUP BY ts.table_number, tstat.capacity
		),
		totals AS (
			SELECT
				SUM(total_sessions) AS total_sessions_all,
				SUM(total_revenue)  AS total_revenue_all
			FROM table_stats
		)
		SELECT
			ts.table_number,
			ts.capacity,
			ts.total_sessions,
			ts.total_hours_used,
			ts.total_revenue,
			COALESCE((ts.total_sessions::float / NULLIF(t.total_sessions_all, 0)) * 100, 0) AS usage_percent,
			COALESCE((ts.total_revenue::float  / NULLIF(t.total_revenue_all,   0)) * 100, 0) AS revenue_percent,
			ts.avg_order_value
		FROM table_stats ts, totals t
		ORDER BY ts.total_sessions DESC
	`

	rows, err := r.pool.Query(ctx, query, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to query table usage breakdown: %w", err)
	}
	defer rows.Close()

	var result []models.NewTableUsageBreakdown
	for rows.Next() {
		var breakdown models.NewTableUsageBreakdown
		if err := rows.Scan(
			&breakdown.TableNumber,
			&breakdown.Capacity,
			&breakdown.TotalSessions,
			&breakdown.TotalHoursUsed,
			&breakdown.TotalRevenue,
			&breakdown.UsagePercent,
			&breakdown.RevenuePercent,
			&breakdown.AvgOrderValue,
		); err != nil {
			return nil, fmt.Errorf("failed to scan table usage breakdown: %w", err)
		}
		result = append(result, breakdown)
	}

	return result, nil
}

// ────────────────────────────────────────────────────────────────────────────
// PEAK HOURS — FIXED
// BUG 4 FIX: occupancy_rate could exceed 100% because it counted cumulative
// distinct tables across all days in the range, not per specific hour.
// Now we count distinct table_numbers only within each hour bucket, and cap
// the result at 100 with LEAST(..., 100).
// ────────────────────────────────────────────────────────────────────────────

func (r *reportRepo) NewGetTablePeakHours(ctx context.Context, from, to time.Time) ([]models.NewTablePeakHour, error) {
	query := `
		WITH hourly_stats AS (
			SELECT
				EXTRACT(HOUR FROM ts.open_time AT TIME ZONE 'Asia/Kathmandu') AS hour,
				-- FIX BUG 4: distinct table_number per hour, not cumulative
				COUNT(DISTINCT ts.table_number)                               AS active_tables,
				COUNT(DISTINCT ts.id)                                         AS sessions_count,
				COALESCE(SUM(p.paid_amount), 0)                               AS total_revenue,
				(SELECT COUNT(*) FROM table_status)                           AS total_tables
			FROM table_session ts
			LEFT JOIN orders   o ON ts.id = o.table_session_id
			LEFT JOIN payments p ON o.id  = p.order_id
			WHERE ts.open_time BETWEEN $1 AND $2
			GROUP BY EXTRACT(HOUR FROM ts.open_time AT TIME ZONE 'Asia/Kathmandu')
		)
		SELECT
			hour::int,
			active_tables,
			-- FIX BUG 4: cap at 100
			LEAST(
				COALESCE((active_tables::float / NULLIF(total_tables, 0)) * 100, 0),
			100) AS occupancy_rate,
			total_revenue,
			sessions_count
		FROM hourly_stats
		ORDER BY hour ASC
	`

	rows, err := r.pool.Query(ctx, query, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to query table peak hours: %w", err)
	}
	defer rows.Close()

	var result []models.NewTablePeakHour
	for rows.Next() {
		var peak models.NewTablePeakHour
		if err := rows.Scan(
			&peak.Hour,
			&peak.ActiveTables,
			&peak.OccupancyRate,
			&peak.TotalRevenue,
			&peak.SessionsCount,
		); err != nil {
			return nil, fmt.Errorf("failed to scan table peak hour: %w", err)
		}
		result = append(result, peak)
	}

	return result, nil
}

// ────────────────────────────────────────────────────────────────────────────
// OCCUPANCY RATE — IMPLEMENTED (was missing from original code)
// BUG 5: NewGetOccupancyRate was called by the orchestrator but never defined.
// Rate = (occupied_count / total_capacity) * 100, returned as 0–100.
// occupied_count = distinct tables that had a session open during that hour.
// total_capacity = SUM of capacity column from table_status.
// ────────────────────────────────────────────────────────────────────────────

func (r *reportRepo) NewGetOccupancyRate(ctx context.Context, from, to time.Time) ([]models.NewOccupancyRate, error) {
	query := `
		WITH capacity_info AS (
			SELECT
				COUNT(*)        AS total_tables,
				SUM(capacity)   AS total_capacity
			FROM table_status
		),
		hourly AS (
			SELECT
				EXTRACT(HOUR FROM ts.open_time AT TIME ZONE 'Asia/Kathmandu') AS hour,
				COUNT(DISTINCT ts.table_number)                               AS occupied_count
			FROM table_session ts
			WHERE ts.open_time BETWEEN $1 AND $2
			GROUP BY EXTRACT(HOUR FROM ts.open_time AT TIME ZONE 'Asia/Kathmandu')
		)
		SELECT
			h.hour::int,
			h.occupied_count,
			ci.total_capacity::int,
			-- Rate as 0–100 percentage of capacity seats that were occupied, capped at 100
			LEAST(
				COALESCE(
					(h.occupied_count::float / NULLIF(ci.total_capacity, 0)) * 100,
				0), 100) AS rate
		FROM hourly h
		CROSS JOIN capacity_info ci
		ORDER BY h.hour ASC
	`

	rows, err := r.pool.Query(ctx, query, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to query occupancy rate: %w", err)
	}
	defer rows.Close()

	var result []models.NewOccupancyRate
	for rows.Next() {
		var o models.NewOccupancyRate
		if err := rows.Scan(
			&o.Hour,
			&o.OccupiedCount,
			&o.TotalCapacity,
			&o.Rate,
		); err != nil {
			return nil, fmt.Errorf("failed to scan occupancy rate: %w", err)
		}
		result = append(result, o)
	}

	return result, nil
}

// ────────────────────────────────────────────────────────────────────────────
// AVG SESSION DURATION — standalone helper
// Returns average session duration in minutes for the given period.
// Uses GREATEST to guard against negative durations from timezone bugs.
// ────────────────────────────────────────────────────────────────────────────

func (r *reportRepo) NewGetAvgSessionDuration(ctx context.Context, from, to time.Time) (float64, error) {
	query := `
		SELECT COALESCE(AVG(
			GREATEST(
				EXTRACT(EPOCH FROM (
					COALESCE(ts.close_time, NOW()) - ts.open_time
				)) / 60,
			0)
		), 0)
		FROM table_session ts
		WHERE ts.open_time BETWEEN $1 AND $2
	`

	var avg float64
	if err := r.pool.QueryRow(ctx, query, from, to).Scan(&avg); err != nil {
		return 0, fmt.Errorf("failed to query avg session duration: %w", err)
	}

	return avg, nil
}

// ────────────────────────────────────────────────────────────────────────────────
// DEFAULT CUSTOMER REPORT
// ────────────────────────────────────────────────────────────────────────────────

func (r *reportRepo) NewGetDefaultCustomerReport(ctx context.Context) (*models.NewDefaultCustomerResponse, error) {
	to := time.Now()
	from := to.AddDate(0, 0, -30)

	var (
		overview          models.NewCustomerOverviewCard
		statsCard         models.NewCustomerStatsCard
		dailyTrend        []models.NewCustomerTrendPoint
		weeklyTrend       []models.NewCustomerTrendPoint
		monthlyTrend      []models.NewCustomerTrendPoint
		yearlyTrend       []models.NewCustomerTrendPoint
		topCustomers      []models.NewTopCustomer
		frequentCustomers []models.NewFrequentCustomer
		retentionMetrics  models.NewRetentionMetrics
		customerSegments  []models.NewCustomerSegment
		streakAnalytics   models.NewStreakAnalytics
		tokenAnalytics    models.NewTokenAnalytics
	)

	g, gCtx := errgroup.WithContext(ctx)

	// Get overview for last 30 days
	g.Go(func() error {
		var err error
		overview, err = r.NewGetCustomerOverview(gCtx, from, to)
		if err != nil {
			return fmt.Errorf("customer overview: %w", err)
		}
		return nil
	})

	// Get all-time stats card
	g.Go(func() error {
		var err error
		statsCard, err = r.NewGetCustomerStatsCard(gCtx)
		if err != nil {
			return fmt.Errorf("stats card: %w", err)
		}
		return nil
	})

	// Get trends - daily
	g.Go(func() error {
		var err error
		dailyTrend, err = r.NewGetLast7DaysCustomerTrend(gCtx)
		if err != nil {
			return fmt.Errorf("daily trend: %w", err)
		}
		return nil
	})

	// Get trends - weekly
	g.Go(func() error {
		var err error
		weeklyTrend, err = r.NewGetLast7WeeksCustomerTrend(gCtx)
		if err != nil {
			return fmt.Errorf("weekly trend: %w", err)
		}
		return nil
	})

	// Get trends - monthly
	g.Go(func() error {
		var err error
		monthlyTrend, err = r.NewGetLast7MonthsCustomerTrend(gCtx)
		if err != nil {
			return fmt.Errorf("monthly trend: %w", err)
		}
		return nil
	})

	// Get trends - yearly
	g.Go(func() error {
		var err error
		yearlyTrend, err = r.NewGetLast7YearsCustomerTrend(gCtx)
		if err != nil {
			return fmt.Errorf("yearly trend: %w", err)
		}
		return nil
	})

	// Get top customers
	g.Go(func() error {
		var err error
		topCustomers, err = r.NewGetTopCustomers(gCtx, from, to, 10)
		if err != nil {
			return fmt.Errorf("top customers: %w", err)
		}
		return nil
	})

	// Get frequent customers
	g.Go(func() error {
		var err error
		frequentCustomers, err = r.NewGetFrequentCustomers(gCtx, from, to, 10)
		if err != nil {
			return fmt.Errorf("frequent customers: %w", err)
		}
		return nil
	})

	// Get retention metrics
	g.Go(func() error {
		var err error
		retentionMetrics, err = r.NewGetRetentionMetrics(gCtx, from, to)
		if err != nil {
			return fmt.Errorf("retention metrics: %w", err)
		}
		return nil
	})

	// Get customer segments
	g.Go(func() error {
		var err error
		customerSegments, err = r.NewGetCustomerSegments(gCtx, from, to)
		if err != nil {
			return fmt.Errorf("customer segments: %w", err)
		}
		return nil
	})

	// Get streak analytics
	g.Go(func() error {
		var err error
		streakAnalytics, err = r.NewGetStreakAnalytics(gCtx, from, to)
		if err != nil {
			return fmt.Errorf("streak analytics: %w", err)
		}
		return nil
	})

	// Get token analytics
	g.Go(func() error {
		var err error
		tokenAnalytics, err = r.NewGetTokenAnalytics(gCtx, from, to)
		if err != nil {
			return fmt.Errorf("token analytics: %w", err)
		}
		return nil
	})

	// Wait for all goroutines to complete
	if err := g.Wait(); err != nil {
		return nil, err
	}

	return &models.NewDefaultCustomerResponse{
		Overview:          overview,
		StatsCard:         statsCard,
		DailyTrend:        dailyTrend,
		WeeklyTrend:       weeklyTrend,
		MonthlyTrend:      monthlyTrend,
		YearlyTrend:       yearlyTrend,
		TopCustomers:      topCustomers,
		FrequentCustomers: frequentCustomers,
		RetentionMetrics:  retentionMetrics,
		CustomerSegments:  customerSegments,
		StreakAnalytics:   streakAnalytics,
		TokenAnalytics:    tokenAnalytics,
	}, nil
}

// ────────────────────────────────────────────────────────────────────────────────
// CUSTOM RANGE CUSTOMER REPORT
// ────────────────────────────────────────────────────────────────────────────────

func (r *reportRepo) NewGetCustomRangeCustomerReport(ctx context.Context, req *models.NewCustomerCustomRangeReportRequest) (*models.NewCustomRangeCustomerResponse, error) {
	from := req.From
	to := req.To.Truncate(24 * time.Hour).Add(24*time.Hour - time.Second)

	// Set pagination defaults
	limit := req.Limit
	if limit <= 0 {
		limit = DefaultCustomerTrendLimit
	}
	if limit > MaxCustomerTrendLimit {
		limit = MaxCustomerTrendLimit
	}
	page := req.Page
	if page < 0 {
		page = 0
	}

	var (
		overview          models.NewCustomerOverviewCard
		statsCard         models.NewCustomerStatsCard
		dailyTrend        *models.NewCustomerPaginatedTrendPoints
		weeklyTrend       *models.NewCustomerPaginatedTrendPoints
		monthlyTrend      *models.NewCustomerPaginatedTrendPoints
		yearlyTrend       *models.NewCustomerPaginatedTrendPoints
		topCustomers      []models.NewTopCustomer
		frequentCustomers []models.NewFrequentCustomer
		retentionMetrics  models.NewRetentionMetrics
		customerSegments  []models.NewCustomerSegment
		streakAnalytics   models.NewStreakAnalytics
		tokenAnalytics    models.NewTokenAnalytics
	)

	g, gCtx := errgroup.WithContext(ctx)

	// Get overview
	g.Go(func() error {
		var err error
		overview, err = r.NewGetCustomerOverview(gCtx, from, to)
		if err != nil {
			return fmt.Errorf("customer overview: %w", err)
		}
		return nil
	})

	// Get stats card
	g.Go(func() error {
		var err error
		statsCard, err = r.NewGetCustomerStatsCard(gCtx)
		if err != nil {
			return fmt.Errorf("stats card: %w", err)
		}
		return nil
	})

	// Get custom range trends
	g.Go(func() error {
		var err error
		dailyTrend, err = r.NewGetCustomDailyCustomerTrend(gCtx, from, to, limit, page)
		if err != nil {
			return fmt.Errorf("custom daily trend: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		var err error
		weeklyTrend, err = r.NewGetCustomWeeklyCustomerTrend(gCtx, from, to, limit, page)
		if err != nil {
			return fmt.Errorf("custom weekly trend: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		var err error
		monthlyTrend, err = r.NewGetCustomMonthlyCustomerTrend(gCtx, from, to, limit, page)
		if err != nil {
			return fmt.Errorf("custom monthly trend: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		var err error
		yearlyTrend, err = r.NewGetCustomYearlyCustomerTrend(gCtx, from, to, limit, page)
		if err != nil {
			return fmt.Errorf("custom yearly trend: %w", err)
		}
		return nil
	})

	// Get analytics
	g.Go(func() error {
		var err error
		topCustomers, err = r.NewGetTopCustomers(gCtx, from, to, 10)
		if err != nil {
			return fmt.Errorf("top customers: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		var err error
		frequentCustomers, err = r.NewGetFrequentCustomers(gCtx, from, to, 10)
		if err != nil {
			return fmt.Errorf("frequent customers: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		var err error
		retentionMetrics, err = r.NewGetRetentionMetrics(gCtx, from, to)
		if err != nil {
			return fmt.Errorf("retention metrics: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		var err error
		customerSegments, err = r.NewGetCustomerSegments(gCtx, from, to)
		if err != nil {
			return fmt.Errorf("customer segments: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		var err error
		streakAnalytics, err = r.NewGetStreakAnalytics(gCtx, from, to)
		if err != nil {
			return fmt.Errorf("streak analytics: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		var err error
		tokenAnalytics, err = r.NewGetTokenAnalytics(gCtx, from, to)
		if err != nil {
			return fmt.Errorf("token analytics: %w", err)
		}
		return nil
	})

	// Wait for all goroutines to complete
	if err := g.Wait(); err != nil {
		return nil, err
	}

	return &models.NewCustomRangeCustomerResponse{
		Overview:          overview,
		StatsCard:         statsCard,
		DailyTrend:        dailyTrend,
		WeeklyTrend:       weeklyTrend,
		MonthlyTrend:      monthlyTrend,
		YearlyTrend:       yearlyTrend,
		TopCustomers:      topCustomers,
		FrequentCustomers: frequentCustomers,
		RetentionMetrics:  retentionMetrics,
		CustomerSegments:  customerSegments,
		StreakAnalytics:   streakAnalytics,
		TokenAnalytics:    tokenAnalytics,
	}, nil
}

// ────────────────────────────────────────────────────────────────────────────────
// TREND FUNCTIONS
// ────────────────────────────────────────────────────────────────────────────────

func (r *reportRepo) NewGetLast7DaysCustomerTrend(ctx context.Context) ([]models.NewCustomerTrendPoint, error) {
	query := `
		WITH daily_customers AS (
			SELECT 
				DATE(first_order) AS period,
				COUNT(DISTINCT customer_phone) AS new_customers
			FROM (
				SELECT 
					customer_phone,
					MIN(created_at) AS first_order
				FROM orders
				WHERE customer_phone IS NOT NULL AND customer_phone != ''
				GROUP BY customer_phone
			) customer_first_orders
			WHERE first_order >= NOW() - INTERVAL '7 days'
			GROUP BY DATE(first_order)
		)
		SELECT 
			TO_CHAR(period, 'YYYY-MM-DD') AS period,
			new_customers,
			SUM(new_customers) OVER (ORDER BY period) AS total_customers
		FROM daily_customers
		ORDER BY period ASC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query last 7 days customer trend: %w", err)
	}
	defer rows.Close()

	var result []models.NewCustomerTrendPoint
	for rows.Next() {
		var point models.NewCustomerTrendPoint
		if err := rows.Scan(&point.Period, &point.NewUsers, &point.TotalUsers); err != nil {
			return nil, fmt.Errorf("failed to scan customer trend: %w", err)
		}
		result = append(result, point)
	}

	return result, nil
}

func (r *reportRepo) NewGetLast7WeeksCustomerTrend(ctx context.Context) ([]models.NewCustomerTrendPoint, error) {
	query := `
		WITH weekly_customers AS (
			SELECT 
				DATE_TRUNC('week', first_order AT TIME ZONE 'Asia/Kathmandu') AS period,
				COUNT(DISTINCT customer_phone) AS new_customers
			FROM (
				SELECT 
					customer_phone,
					MIN(created_at) AS first_order
				FROM orders
				WHERE customer_phone IS NOT NULL AND customer_phone != ''
				GROUP BY customer_phone
			) customer_first_orders
			WHERE first_order >= NOW() - INTERVAL '7 weeks'
			GROUP BY DATE_TRUNC('week', first_order AT TIME ZONE 'Asia/Kathmandu')
		)
		SELECT 
			TO_CHAR(period, 'IYYY-"W"IW') AS period,
			new_customers,
			SUM(new_customers) OVER (ORDER BY period) AS total_customers
		FROM weekly_customers
		ORDER BY period ASC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query last 7 weeks customer trend: %w", err)
	}
	defer rows.Close()

	var result []models.NewCustomerTrendPoint
	for rows.Next() {
		var point models.NewCustomerTrendPoint
		if err := rows.Scan(&point.Period, &point.NewUsers, &point.TotalUsers); err != nil {
			return nil, fmt.Errorf("failed to scan customer trend: %w", err)
		}
		result = append(result, point)
	}

	return result, nil
}

func (r *reportRepo) NewGetLast7MonthsCustomerTrend(ctx context.Context) ([]models.NewCustomerTrendPoint, error) {
	query := `
		WITH monthly_customers AS (
			SELECT 
				DATE_TRUNC('month', first_order AT TIME ZONE 'Asia/Kathmandu') AS period,
				COUNT(DISTINCT customer_phone) AS new_customers
			FROM (
				SELECT 
					customer_phone,
					MIN(created_at) AS first_order
				FROM orders
				WHERE customer_phone IS NOT NULL AND customer_phone != ''
				GROUP BY customer_phone
			) customer_first_orders
			WHERE first_order >= NOW() - INTERVAL '7 months'
			GROUP BY DATE_TRUNC('month', first_order AT TIME ZONE 'Asia/Kathmandu')
		)
		SELECT 
			TO_CHAR(period, 'YYYY-MM') AS period,
			new_customers,
			SUM(new_customers) OVER (ORDER BY period) AS total_customers
		FROM monthly_customers
		ORDER BY period ASC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query last 7 months customer trend: %w", err)
	}
	defer rows.Close()

	var result []models.NewCustomerTrendPoint
	for rows.Next() {
		var point models.NewCustomerTrendPoint
		if err := rows.Scan(&point.Period, &point.NewUsers, &point.TotalUsers); err != nil {
			return nil, fmt.Errorf("failed to scan customer trend: %w", err)
		}
		result = append(result, point)
	}

	return result, nil
}

func (r *reportRepo) NewGetLast7YearsCustomerTrend(ctx context.Context) ([]models.NewCustomerTrendPoint, error) {
	query := `
		WITH yearly_customers AS (
			SELECT 
				DATE_TRUNC('year', first_order AT TIME ZONE 'Asia/Kathmandu') AS period,
				COUNT(DISTINCT customer_phone) AS new_customers
			FROM (
				SELECT 
					customer_phone,
					MIN(created_at) AS first_order
				FROM orders
				WHERE customer_phone IS NOT NULL AND customer_phone != ''
				GROUP BY customer_phone
			) customer_first_orders
			WHERE first_order >= NOW() - INTERVAL '7 years'
			GROUP BY DATE_TRUNC('year', first_order AT TIME ZONE 'Asia/Kathmandu')
		)
		SELECT 
			TO_CHAR(period, 'YYYY') AS period,
			new_customers,
			SUM(new_customers) OVER (ORDER BY period) AS total_customers
		FROM yearly_customers
		ORDER BY period ASC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query last 7 years customer trend: %w", err)
	}
	defer rows.Close()

	var result []models.NewCustomerTrendPoint
	for rows.Next() {
		var point models.NewCustomerTrendPoint
		if err := rows.Scan(&point.Period, &point.NewUsers, &point.TotalUsers); err != nil {
			return nil, fmt.Errorf("failed to scan customer trend: %w", err)
		}
		result = append(result, point)
	}

	return result, nil
}

// ────────────────────────────────────────────────────────────────────────────────
// CUSTOM TREND FUNCTIONS
// ────────────────────────────────────────────────────────────────────────────────

// ────────────────────────────────────────────────────────────────────────────────
// CUSTOM WEEKLY CUSTOMER TREND
// ────────────────────────────────────────────────────────────────────────────────

func (r *reportRepo) NewGetCustomWeeklyCustomerTrend(ctx context.Context, from, to time.Time, limit, page int) (*models.NewCustomerPaginatedTrendPoints, error) {
	offset := page * limit

	var total int
	countQuery := `
		SELECT COUNT(DISTINCT DATE_TRUNC('week', first_order AT TIME ZONE 'Asia/Kathmandu'))
		FROM (
			SELECT 
				customer_phone,
				MIN(created_at) AS first_order
			FROM orders
			WHERE customer_phone IS NOT NULL AND customer_phone != ''
			GROUP BY customer_phone
		) customer_first_orders
		WHERE first_order BETWEEN $1 AND $2
	`
	if err := r.pool.QueryRow(ctx, countQuery, from, to).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count custom weekly customer trend: %w", err)
	}

	query := `
		SELECT COALESCE(
			json_agg(
				json_build_object('period', period, 'new_users', new_customers, 'total_users', total_customers)
				ORDER BY week_start ASC
			), '[]'::json
		)
		FROM (
			SELECT 
				TO_CHAR(DATE_TRUNC('week', first_order AT TIME ZONE 'Asia/Kathmandu'), 'IYYY-"W"IW') AS period,
				DATE_TRUNC('week', first_order AT TIME ZONE 'Asia/Kathmandu') AS week_start,
				COUNT(DISTINCT customer_phone) AS new_customers,
				SUM(COUNT(DISTINCT customer_phone)) OVER (ORDER BY DATE_TRUNC('week', first_order AT TIME ZONE 'Asia/Kathmandu')) AS total_customers
			FROM (
				SELECT 
					customer_phone,
					MIN(created_at) AS first_order
				FROM orders
				WHERE customer_phone IS NOT NULL AND customer_phone != ''
				GROUP BY customer_phone
			) customer_first_orders
			WHERE first_order BETWEEN $1 AND $2
			GROUP BY DATE_TRUNC('week', first_order AT TIME ZONE 'Asia/Kathmandu')
			ORDER BY week_start ASC
			LIMIT $3 OFFSET $4
		) sub
	`

	var resultJSON []byte
	if err := r.pool.QueryRow(ctx, query, from, to, limit, offset).Scan(&resultJSON); err != nil {
		return nil, fmt.Errorf("failed to query custom weekly customer trend: %w", err)
	}

	var result []models.NewCustomerTrendPoint
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal custom weekly customer trend: %w", err)
	}

	hasMore := (page+1)*limit < total
	nextPage := page + 1
	if !hasMore {
		nextPage = page
	}

	return &models.NewCustomerPaginatedTrendPoints{
		Data: result,
		Pagination: models.NewPaginationInfo{
			Total:    total,
			HasMore:  hasMore,
			NextPage: nextPage,
			Limit:    limit,
			Page:     page,
		},
	}, nil
}

// ────────────────────────────────────────────────────────────────────────────────
// CUSTOM MONTHLY CUSTOMER TREND
// ────────────────────────────────────────────────────────────────────────────────

func (r *reportRepo) NewGetCustomMonthlyCustomerTrend(ctx context.Context, from, to time.Time, limit, page int) (*models.NewCustomerPaginatedTrendPoints, error) {
	offset := page * limit

	var total int
	countQuery := `
		SELECT COUNT(DISTINCT DATE_TRUNC('month', created_at AT TIME ZONE 'Asia/Kathmandu'))
		FROM users
		WHERE created_at BETWEEN $1 AND $2
	`
	if err := r.pool.QueryRow(ctx, countQuery, from, to).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count custom monthly customer trend: %w", err)
	}

	query := `
		SELECT COALESCE(
			json_agg(
				json_build_object('period', period, 'new_users', new_users, 'total_users', total_users)
				ORDER BY month_start ASC
			), '[]'::json
		)
		FROM (
			SELECT 
				TO_CHAR(DATE_TRUNC('month', created_at AT TIME ZONE 'Asia/Kathmandu'), 'YYYY-MM') AS period,
				DATE_TRUNC('month', created_at AT TIME ZONE 'Asia/Kathmandu') AS month_start,
				COUNT(*) AS new_users,
				SUM(COUNT(*)) OVER (ORDER BY DATE_TRUNC('month', created_at AT TIME ZONE 'Asia/Kathmandu')) AS total_users
			FROM users
			WHERE created_at BETWEEN $1 AND $2
			GROUP BY DATE_TRUNC('month', created_at AT TIME ZONE 'Asia/Kathmandu')
			ORDER BY month_start ASC
			LIMIT $3 OFFSET $4
		) sub
	`

	var resultJSON []byte
	if err := r.pool.QueryRow(ctx, query, from, to, limit, offset).Scan(&resultJSON); err != nil {
		return nil, fmt.Errorf("failed to query custom monthly customer trend: %w", err)
	}

	var result []models.NewCustomerTrendPoint
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal custom monthly customer trend: %w", err)
	}

	hasMore := (page+1)*limit < total
	nextPage := page + 1
	if !hasMore {
		nextPage = page
	}

	return &models.NewCustomerPaginatedTrendPoints{
		Data: result,
		Pagination: models.NewPaginationInfo{
			Total:    total,
			HasMore:  hasMore,
			NextPage: nextPage,
			Limit:    limit,
			Page:     page,
		},
	}, nil
}

// ────────────────────────────────────────────────────────────────────────────────
// CUSTOM YEARLY CUSTOMER TREND
// ────────────────────────────────────────────────────────────────────────────────

func (r *reportRepo) NewGetCustomYearlyCustomerTrend(ctx context.Context, from, to time.Time, limit, page int) (*models.NewCustomerPaginatedTrendPoints, error) {
	offset := page * limit

	var total int
	countQuery := `
		SELECT COUNT(DISTINCT DATE_TRUNC('year', created_at AT TIME ZONE 'Asia/Kathmandu'))
		FROM users
		WHERE created_at BETWEEN $1 AND $2
	`
	if err := r.pool.QueryRow(ctx, countQuery, from, to).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count custom yearly customer trend: %w", err)
	}

	query := `
		SELECT COALESCE(
			json_agg(
				json_build_object('period', period, 'new_users', new_users, 'total_users', total_users)
				ORDER BY year_start ASC
			), '[]'::json
		)
		FROM (
			SELECT 
				TO_CHAR(DATE_TRUNC('year', created_at AT TIME ZONE 'Asia/Kathmandu'), 'YYYY') AS period,
				DATE_TRUNC('year', created_at AT TIME ZONE 'Asia/Kathmandu') AS year_start,
				COUNT(*) AS new_users,
				SUM(COUNT(*)) OVER (ORDER BY DATE_TRUNC('year', created_at AT TIME ZONE 'Asia/Kathmandu')) AS total_users
			FROM users
			WHERE created_at BETWEEN $1 AND $2
			GROUP BY DATE_TRUNC('year', created_at AT TIME ZONE 'Asia/Kathmandu')
			ORDER BY year_start ASC
			LIMIT $3 OFFSET $4
		) sub
	`

	var resultJSON []byte
	if err := r.pool.QueryRow(ctx, query, from, to, limit, offset).Scan(&resultJSON); err != nil {
		return nil, fmt.Errorf("failed to query custom yearly customer trend: %w", err)
	}

	var result []models.NewCustomerTrendPoint
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal custom yearly customer trend: %w", err)
	}

	hasMore := (page+1)*limit < total
	nextPage := page + 1
	if !hasMore {
		nextPage = page
	}

	return &models.NewCustomerPaginatedTrendPoints{
		Data: result,
		Pagination: models.NewPaginationInfo{
			Total:    total,
			HasMore:  hasMore,
			NextPage: nextPage,
			Limit:    limit,
			Page:     page,
		},
	}, nil
}

func (r *reportRepo) NewGetCustomDailyCustomerTrend(ctx context.Context, from, to time.Time, limit, page int) (*models.NewCustomerPaginatedTrendPoints, error) {
	offset := page * limit

	var total int
	countQuery := `
		SELECT COUNT(DISTINCT DATE(first_order))
		FROM (
			SELECT 
				customer_phone,
				MIN(created_at) AS first_order
			FROM orders
			WHERE customer_phone IS NOT NULL AND customer_phone != ''
			GROUP BY customer_phone
		) customer_first_orders
		WHERE first_order BETWEEN $1 AND $2
	`
	if err := r.pool.QueryRow(ctx, countQuery, from, to).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count custom daily customer trend: %w", err)
	}

	query := `
		SELECT COALESCE(
			json_agg(
				json_build_object('period', period, 'new_users', new_customers, 'total_users', total_customers)
				ORDER BY period ASC
			), '[]'::json
		)
		FROM (
			SELECT 
				TO_CHAR(DATE(first_order), 'YYYY-MM-DD') AS period,
				DATE(first_order) AS order_date,
				COUNT(DISTINCT customer_phone) AS new_customers,
				SUM(COUNT(DISTINCT customer_phone)) OVER (ORDER BY DATE(first_order)) AS total_customers
			FROM (
				SELECT 
					customer_phone,
					MIN(created_at) AS first_order
				FROM orders
				WHERE customer_phone IS NOT NULL AND customer_phone != ''
				GROUP BY customer_phone
			) customer_first_orders
			WHERE first_order BETWEEN $1 AND $2
			GROUP BY DATE(first_order)
			ORDER BY order_date ASC
			LIMIT $3 OFFSET $4
		) sub
	`

	var resultJSON []byte
	if err := r.pool.QueryRow(ctx, query, from, to, limit, offset).Scan(&resultJSON); err != nil {
		return nil, fmt.Errorf("failed to query custom daily customer trend: %w", err)
	}

	var result []models.NewCustomerTrendPoint
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal custom daily customer trend: %w", err)
	}

	hasMore := (page+1)*limit < total
	nextPage := page + 1
	if !hasMore {
		nextPage = page
	}

	return &models.NewCustomerPaginatedTrendPoints{
		Data: result,
		Pagination: models.NewPaginationInfo{
			Total:    total,
			HasMore:  hasMore,
			NextPage: nextPage,
			Limit:    limit,
			Page:     page,
		},
	}, nil
}

// Similar implementations for Weekly, Monthly, Yearly custom trends...
// (I'll skip them for brevity but they follow the same pattern)

// ────────────────────────────────────────────────────────────────────────────────
// ANALYTICS FUNCTIONS
// ────────────────────────────────────────────────────────────────────────────────

func (r *reportRepo) NewGetCustomerOverview(ctx context.Context, from, to time.Time) (models.NewCustomerOverviewCard, error) {
	query := `
		WITH all_customers AS (
			SELECT DISTINCT customer_phone
			FROM orders
			WHERE customer_phone IS NOT NULL AND customer_phone != ''
		),
		new_customers AS (
			SELECT DISTINCT customer_phone
			FROM orders
			WHERE customer_phone IS NOT NULL AND customer_phone != ''
			GROUP BY customer_phone
			HAVING MIN(created_at) BETWEEN $1 AND $2
		),
		active_customers AS (
			SELECT DISTINCT customer_phone
			FROM orders
			WHERE created_at BETWEEN $1 AND $2
				AND customer_phone IS NOT NULL AND customer_phone != ''
		),
		customer_order_stats AS (
			SELECT 
				o.customer_phone,
				COUNT(DISTINCT o.id) AS order_count,
				COALESCE(SUM(p.paid_amount), 0) AS total_spent
			FROM orders o
			LEFT JOIN payments p ON o.id = p.order_id
			WHERE o.created_at BETWEEN $1 AND $2
				AND o.customer_phone IS NOT NULL AND o.customer_phone != ''
			GROUP BY o.customer_phone
		)
		SELECT 
			(SELECT COUNT(*) FROM all_customers) AS total_customers,
			(SELECT COUNT(*) FROM new_customers) AS new_customers,
			(SELECT COUNT(*) FROM active_customers) AS active_customers,
			(SELECT COUNT(*) FROM customer_order_stats WHERE order_count > 1) AS returning_customers,
			COALESCE(AVG(order_count), 0) AS avg_orders_per_customer,
			COALESCE(AVG(total_spent), 0) AS avg_spend_per_customer,
			0 AS growth_percent
		FROM customer_order_stats
	`

	var overview models.NewCustomerOverviewCard
	err := r.pool.QueryRow(ctx, query, from, to).Scan(
		&overview.TotalCustomers,
		&overview.NewCustomers,
		&overview.ActiveCustomers,
		&overview.ReturningCustomers,
		&overview.AvgOrdersPerCustomer,
		&overview.AvgSpendPerCustomer,
		&overview.GrowthPercent,
	)
	if err != nil && err.Error() != "sql: no rows in result set" {
		return models.NewCustomerOverviewCard{}, fmt.Errorf("failed to scan customer overview: %w", err)
	}

	// Calculate growth vs previous period
	diff := to.Sub(from)
	prevFrom := from.Add(-diff)

	var prevCustomers int
	if err := r.pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT customer_phone)
		FROM orders
		WHERE created_at BETWEEN $1 AND $2
			AND customer_phone IS NOT NULL AND customer_phone != ''
	`, prevFrom, from).Scan(&prevCustomers); err != nil && err.Error() != "sql: no rows in result set" {
		return overview, nil // Don't fail if no previous data
	}

	if prevCustomers > 0 {
		overview.GrowthPercent = (float64(overview.NewCustomers-prevCustomers) / float64(prevCustomers)) * 100
	}

	return overview, nil
}

func (r *reportRepo) NewGetCustomerStatsCard(ctx context.Context) (models.NewCustomerStatsCard, error) {
	query := `
		SELECT
			COUNT(DISTINCT o.customer_phone) AS total_customers,
			COUNT(DISTINCT o.id) AS total_orders,
			COALESCE(SUM(p.paid_amount), 0) AS total_revenue,
			CASE 
				WHEN COUNT(DISTINCT o.customer_phone) > 0 
				THEN COALESCE(SUM(p.paid_amount), 0) / COUNT(DISTINCT o.customer_phone)
				ELSE 0 
			END AS avg_lifetime_value,
			COALESCE(SUM(CASE WHEN tt.type = 'EARN' THEN tt.amount ELSE 0 END), 0) AS total_tokens_issued,
			COALESCE(SUM(CASE WHEN tt.type = 'SPEND' THEN tt.amount ELSE 0 END), 0) AS total_tokens_redeemed,
			COUNT(DISTINCT CASE WHEN cs.current_streak > 0 THEN cs.phone_number END) AS active_streak_customers
		FROM orders o
		LEFT JOIN payments p ON o.id = p.order_id
		LEFT JOIN token_transactions tt ON o.customer_phone = tt.phone_number
		LEFT JOIN customer_streaks cs ON o.customer_phone = cs.phone_number
		WHERE o.customer_phone IS NOT NULL AND o.customer_phone != ''
	`

	var stats models.NewCustomerStatsCard
	if err := r.pool.QueryRow(ctx, query).Scan(
		&stats.TotalCustomers,
		&stats.TotalOrders,
		&stats.TotalRevenue,
		&stats.AvgLifetimeValue,
		&stats.TotalTokensIssued,
		&stats.TotalTokensRedeemed,
		&stats.ActiveStreakCustomers,
	); err != nil {
		return models.NewCustomerStatsCard{}, fmt.Errorf("failed to scan customer stats card: %w", err)
	}

	return stats, nil
}

func (r *reportRepo) NewGetTopCustomers(ctx context.Context, from, to time.Time, limit int) ([]models.NewTopCustomer, error) {
	query := `
		SELECT 
			o.customer_phone AS customer_id,
			COALESCE(MAX(o.customer_name), 'Guest') AS customer_name,
			o.customer_phone AS phone_number,
			COUNT(DISTINCT o.id) AS total_orders,
			COALESCE(SUM(p.paid_amount), 0) AS total_spent,
			COALESCE(AVG(p.paid_amount), 0) AS avg_order_value,
			TO_CHAR(MAX(o.created_at), 'YYYY-MM-DD') AS last_order_date
		FROM orders o
		LEFT JOIN payments p ON o.id = p.order_id
		WHERE o.created_at BETWEEN $1 AND $2
			AND o.customer_phone IS NOT NULL AND o.customer_phone != ''
		GROUP BY o.customer_phone
		ORDER BY total_spent DESC
		LIMIT $3
	`

	rows, err := r.pool.Query(ctx, query, from, to, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query top customers: %w", err)
	}
	defer rows.Close()

	var result []models.NewTopCustomer
	for rows.Next() {
		var customer models.NewTopCustomer
		if err := rows.Scan(
			&customer.CustomerID,
			&customer.CustomerName,
			&customer.PhoneNumber,
			&customer.TotalOrders,
			&customer.TotalSpent,
			&customer.AvgOrderValue,
			&customer.LastOrderDate,
		); err != nil {
			return nil, fmt.Errorf("failed to scan top customer: %w", err)
		}
		result = append(result, customer)
	}

	return result, nil
}
func (r *reportRepo) NewGetFrequentCustomers(ctx context.Context, from, to time.Time, limit int) ([]models.NewFrequentCustomer, error) {
	query := `
		WITH customer_visits AS (
			SELECT 
				o.customer_phone AS phone_number,
				COALESCE(MAX(o.customer_name), 'Guest') AS customer_name,
				COUNT(DISTINCT o.id) AS total_orders,
				COUNT(DISTINCT DATE(o.created_at)) AS visit_days,
				EXTRACT(DAY FROM NOW() - MAX(o.created_at)) AS days_since_last_visit
			FROM orders o
			WHERE o.created_at BETWEEN $1 AND $2
				AND o.customer_phone IS NOT NULL AND o.customer_phone != ''
			GROUP BY o.customer_phone
			HAVING COUNT(DISTINCT o.id) >= 2
		)
		SELECT 
			phone_number AS customer_id,
			customer_name,
			phone_number,
			(visit_days::float / 30) AS visit_frequency,
			days_since_last_visit::int,
			total_orders,
			'Unknown' AS favorite_category
		FROM customer_visits
		ORDER BY visit_frequency DESC
		LIMIT $3
	`

	rows, err := r.pool.Query(ctx, query, from, to, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query frequent customers: %w", err)
	}
	defer rows.Close()

	var result []models.NewFrequentCustomer
	for rows.Next() {
		var customer models.NewFrequentCustomer
		if err := rows.Scan(
			&customer.CustomerID,
			&customer.CustomerName,
			&customer.PhoneNumber,
			&customer.VisitFrequency,
			&customer.DaysSinceLastVisit,
			&customer.TotalOrders,
			&customer.FavoriteCategory,
		); err != nil {
			return nil, fmt.Errorf("failed to scan frequent customer: %w", err)
		}
		result = append(result, customer)
	}

	return result, nil
}

func (r *reportRepo) NewGetRetentionMetrics(ctx context.Context, from, to time.Time) (models.NewRetentionMetrics, error) {
	query := `
		WITH customer_orders AS (
			SELECT 
				o.customer_phone,
				COUNT(DISTINCT o.id) AS order_count,
				MIN(o.created_at) AS first_order,
				MAX(o.created_at) AS last_order
			FROM orders o
			WHERE o.customer_phone IS NOT NULL AND o.customer_phone != ''
			GROUP BY o.customer_phone
		)
		SELECT 
			COALESCE(
				(COUNT(CASE WHEN order_count >= 2 AND last_order >= NOW() - INTERVAL '30 days' THEN 1 END)::float / 
				NULLIF(COUNT(CASE WHEN first_order <= NOW() - INTERVAL '30 days' THEN 1 END), 0)) * 100, 0
			) AS retention_rate_30,
			COALESCE(
				(COUNT(CASE WHEN order_count >= 2 AND last_order >= NOW() - INTERVAL '90 days' THEN 1 END)::float / 
				NULLIF(COUNT(CASE WHEN first_order <= NOW() - INTERVAL '90 days' THEN 1 END), 0)) * 100, 0
			) AS retention_rate_90,
			COALESCE(
				(COUNT(CASE WHEN last_order < NOW() - INTERVAL '90 days' THEN 1 END)::float / 
				NULLIF(COUNT(*), 0)) * 100, 0
			) AS churn_rate,
			COALESCE(
				(COUNT(CASE WHEN order_count >= 2 THEN 1 END)::float / 
				NULLIF(COUNT(CASE WHEN order_count >= 1 THEN 1 END), 0)) * 100, 0
			) AS repeat_purchase_rate,
			0 AS avg_days_between_orders
		FROM customer_orders
	`

	var metrics models.NewRetentionMetrics
	if err := r.pool.QueryRow(ctx, query).Scan(
		&metrics.RetentionRate30Days,
		&metrics.RetentionRate90Days,
		&metrics.ChurnRate,
		&metrics.RepeatPurchaseRate,
		&metrics.AvgDaysBetweenOrders,
	); err != nil {
		return models.NewRetentionMetrics{}, fmt.Errorf("failed to scan retention metrics: %w", err)
	}

	return metrics, nil
}

func (r *reportRepo) NewGetCustomerSegments(ctx context.Context, from, to time.Time) ([]models.NewCustomerSegment, error) {
	query := `
		WITH customer_spend AS (
			SELECT 
				o.customer_phone,
				COALESCE(SUM(p.paid_amount), 0) AS total_spent,
				COUNT(DISTINCT o.id) AS order_count
			FROM orders o
			LEFT JOIN payments p ON o.id = p.order_id
			WHERE o.created_at BETWEEN $1 AND $2
				AND o.customer_phone IS NOT NULL AND o.customer_phone != ''
			GROUP BY o.customer_phone
		),
		segments AS (
			SELECT 
				CASE 
					WHEN total_spent >= 10000 THEN 'High Spender'
					WHEN total_spent >= 5000 THEN 'Regular'
					WHEN total_spent > 0 THEN 'Occasional'
					ELSE 'New'
				END AS segment,
				COUNT(*) AS count,
				AVG(total_spent) AS avg_spend,
				MIN(total_spent) AS min_spend,
				MAX(total_spent) AS max_spend,
				SUM(total_spent) AS total_revenue
			FROM customer_spend
			GROUP BY segment
		),
		total_customers AS (
			SELECT SUM(count) AS total FROM segments
		)
		SELECT 
			s.segment,
			s.count,
			(s.count::float / tc.total) * 100 AS percent,
			s.avg_spend,
			s.min_spend,
			s.max_spend,
			s.total_revenue
		FROM segments s, total_customers tc
		ORDER BY s.avg_spend DESC
	`

	rows, err := r.pool.Query(ctx, query, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to query customer segments: %w", err)
	}
	defer rows.Close()

	var result []models.NewCustomerSegment
	for rows.Next() {
		var segment models.NewCustomerSegment
		if err := rows.Scan(&segment.Segment, &segment.Count, &segment.Percent, &segment.AvgSpend, &segment.MinSpend, &segment.MaxSpend, &segment.TotalRevenue); err != nil {
			return nil, fmt.Errorf("failed to scan customer segment: %w", err)
		}
		result = append(result, segment)
	}

	return result, nil
}
func (r *reportRepo) NewGetStreakAnalytics(ctx context.Context, from, to time.Time) (models.NewStreakAnalytics, error) {
	query := `
		WITH streak_stats AS (
			SELECT 
				current_streak,
				CASE 
					WHEN current_streak BETWEEN 1 AND 3 THEN '1-3 days'
					WHEN current_streak BETWEEN 4 AND 7 THEN '4-7 days'
					WHEN current_streak BETWEEN 8 AND 14 THEN '8-14 days'
					WHEN current_streak >= 15 THEN '15+ days'
					ELSE '0 days'
				END AS streak_range
			FROM customer_streaks
		),
		total_streakers AS (
			SELECT COUNT(*) AS total FROM customer_streaks WHERE current_streak > 0
		)
		SELECT 
			COALESCE((SELECT total FROM total_streakers), 0) AS total_streak_customers,
			COALESCE(AVG(current_streak), 0) AS avg_streak_length,
			COALESCE(MAX(current_streak), 0) AS max_streak_length,
			COUNT(DISTINCT CASE WHEN updated_at >= NOW() - INTERVAL '30 days' THEN phone_number END) AS monthly_active_streakers
		FROM customer_streaks
	`

	var analytics models.NewStreakAnalytics
	if err := r.pool.QueryRow(ctx, query).Scan(
		&analytics.TotalStreakCustomers,
		&analytics.AvgStreakLength,
		&analytics.MaxStreakLength,
		&analytics.MonthlyActiveStreakers,
	); err != nil {
		return models.NewStreakAnalytics{}, fmt.Errorf("failed to scan streak analytics: %w", err)
	}

	// Get streak distribution
	distQuery := `
		WITH streak_stats AS (
			SELECT 
				CASE 
					WHEN current_streak BETWEEN 1 AND 3 THEN '1-3 days'
					WHEN current_streak BETWEEN 4 AND 7 THEN '4-7 days'
					WHEN current_streak BETWEEN 8 AND 14 THEN '8-14 days'
					WHEN current_streak >= 15 THEN '15+ days'
					ELSE '0 days'
				END AS streak_range,
				COUNT(*) AS count
			FROM customer_streaks
			WHERE current_streak > 0
			GROUP BY streak_range
		),
		total AS (
			SELECT SUM(count) AS total FROM streak_stats
		)
		SELECT 
			ss.streak_range,
			ss.count,
			(ss.count::float / t.total) * 100 AS percent
		FROM streak_stats ss, total t
		ORDER BY 
			CASE ss.streak_range
				WHEN '1-3 days' THEN 1
				WHEN '4-7 days' THEN 2
				WHEN '8-14 days' THEN 3
				WHEN '15+ days' THEN 4
			END
	`

	rows, err := r.pool.Query(ctx, distQuery)
	if err != nil {
		return analytics, fmt.Errorf("failed to query streak distribution: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var dist models.NewStreakDistribution
		if err := rows.Scan(&dist.StreakRange, &dist.Count, &dist.Percent); err != nil {
			return analytics, fmt.Errorf("failed to scan streak distribution: %w", err)
		}
		analytics.StreakDistribution = append(analytics.StreakDistribution, dist)
	}

	return analytics, nil
}

func (r *reportRepo) NewGetTokenAnalytics(ctx context.Context, from, to time.Time) (models.NewTokenAnalytics, error) {
	query := `
		WITH token_summary AS (
			SELECT 
				COALESCE(SUM(CASE WHEN type = 'EARN' THEN amount ELSE 0 END), 0) AS total_earned,
				COALESCE(SUM(CASE WHEN type = 'SPEND' THEN amount ELSE 0 END), 0) AS total_spent,
				COUNT(DISTINCT phone_number) AS total_customers
			FROM token_transactions
			WHERE created_at BETWEEN $1 AND $2
		),
		current_balance AS (
			SELECT COALESCE(SUM(total_tokens), 0) AS active_balance
			FROM user_tokens
		)
		SELECT 
			ts.total_earned,
			ts.total_spent,
			cb.active_balance,
			CASE 
				WHEN ts.total_customers > 0 
				THEN ts.total_earned / ts.total_customers
				ELSE 0 
			END AS avg_tokens_per_customer,
			CASE 
				WHEN ts.total_earned > 0 
				THEN (ts.total_spent / ts.total_earned) * 100
				ELSE 0 
			END AS token_redemption_rate
		FROM token_summary ts, current_balance cb
	`

	var analytics models.NewTokenAnalytics
	if err := r.pool.QueryRow(ctx, query, from, to).Scan(
		&analytics.TotalTokensEarned,
		&analytics.TotalTokensSpent,
		&analytics.ActiveTokenBalance,
		&analytics.AvgTokensPerCustomer,
		&analytics.TokenRedemptionRate,
	); err != nil {
		return models.NewTokenAnalytics{}, fmt.Errorf("failed to scan token analytics: %w", err)
	}

	// Get top token earners
	earnersQuery := `
		SELECT 
			u.name AS customer_name,
			u.phone AS phone_number,
			COALESCE(SUM(CASE WHEN tt.type = 'EARN' THEN tt.amount ELSE 0 END), 0) AS tokens_earned,
			COALESCE(SUM(CASE WHEN tt.type = 'SPEND' THEN tt.amount ELSE 0 END), 0) AS tokens_spent,
			COALESCE(ut.total_tokens, 0) AS token_balance
		FROM users u
		LEFT JOIN token_transactions tt ON u.phone = tt.phone_number
		LEFT JOIN user_tokens ut ON u.phone = ut.phone_number
		WHERE tt.created_at BETWEEN $1 AND $2
		GROUP BY u.id, u.name, u.phone, ut.total_tokens
		ORDER BY tokens_earned DESC
		LIMIT 10
	`

	rows, err := r.pool.Query(ctx, earnersQuery, from, to)
	if err != nil {
		return analytics, fmt.Errorf("failed to query top token earners: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var earner models.NewTopTokenCustomer
		if err := rows.Scan(&earner.CustomerName, &earner.PhoneNumber, &earner.TokensEarned, &earner.TokensSpent, &earner.TokenBalance); err != nil {
			return analytics, fmt.Errorf("failed to scan top token earner: %w", err)
		}
		analytics.TopTokenEarners = append(analytics.TopTokenEarners, earner)
	}

	return analytics, nil
}

// Additional custom trend implementations would go here...
// (Weekly, Monthly, Yearly custom trends following the same pattern as daily)

// ────────────────────────────────────────────────────────────────────────────────
// DEFAULT SALES REPORT - Returns pre-defined data (last 7/30 days with caching)
// ────────────────────────────────────────────────────────────────────────────────

// ─── Default: Last 7 Days ─────────────────────────────────────────────────────

func (r *reportRepo) NewGetLast7DaysMenuItemOrderStats(ctx context.Context) ([]models.NewMenuItemOrderStat, error) {
	to := time.Now()
	from := to.AddDate(0, 0, -7)

	query := `
        SELECT
            mi.id::text AS item_id,
            mi.name AS item_name,
            COALESCE(c.name, 'Uncategorized') AS category_name,
            COALESCE(mi.image_url, '') AS image_url,
            mi.price,
            COALESCE(COUNT(DISTINCT oi.order_id), 0) AS total_orders,
            COALESCE(SUM(oi.quantity), 0) AS total_quantity,
            COALESCE(SUM(oi.price * oi.quantity), 0) AS total_revenue
        FROM menu_items mi
        LEFT JOIN categories c ON mi.category_id = c.id
        LEFT JOIN order_items oi ON mi.id = oi.menu_item_id
        LEFT JOIN orders o ON oi.order_id = o.id
            AND o.created_at BETWEEN $1 AND $2
        GROUP BY mi.id, mi.name, c.name, mi.image_url, mi.price
        ORDER BY total_orders DESC
    `

	rows, err := r.pool.Query(ctx, query, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to query last 7 days menu item order stats: %w", err)
	}
	defer rows.Close()

	var result []models.NewMenuItemOrderStat
	for rows.Next() {
		var item models.NewMenuItemOrderStat
		if err := rows.Scan(
			&item.ItemID,
			&item.ItemName,
			&item.CategoryName,
			&item.ImageURL,
			&item.Price,
			&item.TotalOrders,
			&item.TotalQuantity,
			&item.TotalRevenue,
		); err != nil {
			return nil, fmt.Errorf("failed to scan last 7 days menu item order stat: %w", err)
		}
		result = append(result, item)
	}

	if result == nil {
		result = []models.NewMenuItemOrderStat{}
	}

	return result, nil
}

// ─── Custom Date Range ────────────────────────────────────────────────────────

func (r *reportRepo) NewGetCustomRangeMenuItemOrderStats(ctx context.Context, from, to time.Time) ([]models.NewMenuItemOrderStat, error) {
	query := `
        SELECT
            mi.id::text AS item_id,
            mi.name AS item_name,
            COALESCE(c.name, 'Uncategorized') AS category_name,
            COALESCE(mi.image_url, '') AS image_url,
            mi.price,
            COALESCE(COUNT(DISTINCT oi.order_id), 0) AS total_orders,
            COALESCE(SUM(oi.quantity), 0) AS total_quantity,
            COALESCE(SUM(oi.price * oi.quantity), 0) AS total_revenue
        FROM menu_items mi
        LEFT JOIN categories c ON mi.category_id = c.id
        LEFT JOIN order_items oi ON mi.id = oi.menu_item_id
        LEFT JOIN orders o ON oi.order_id = o.id
            AND o.created_at BETWEEN $1 AND $2
        GROUP BY mi.id, mi.name, c.name, mi.image_url, mi.price
        ORDER BY total_orders DESC
    `

	rows, err := r.pool.Query(ctx, query, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to query custom range menu item order stats: %w", err)
	}
	defer rows.Close()

	var result []models.NewMenuItemOrderStat
	for rows.Next() {
		var item models.NewMenuItemOrderStat
		if err := rows.Scan(
			&item.ItemID,
			&item.ItemName,
			&item.CategoryName,
			&item.ImageURL,
			&item.Price,
			&item.TotalOrders,
			&item.TotalQuantity,
			&item.TotalRevenue,
		); err != nil {
			return nil, fmt.Errorf("failed to scan custom range menu item order stat: %w", err)
		}
		result = append(result, item)
	}

	if result == nil {
		result = []models.NewMenuItemOrderStat{}
	}

	return result, nil
}

func (r *reportRepo) NewGetDefaultSalesReport(ctx context.Context) (*models.NewDefaultSalesResponse, error) {
	to := time.Now()
	from := to.AddDate(0, 0, -30)

	var (
		overview             models.NewSalesOverviewCard
		statsCard            models.NewSalesStatsCard
		dailyTrend           []models.NewSalesTrendPoint
		weeklyTrend          []models.NewSalesTrendPoint
		monthlyTrend         []models.NewSalesTrendPoint
		yearlyTrend          []models.NewSalesTrendPoint
		topSellingItems      []models.NewTopSellingItem
		topCategories        []models.NewTopCategory
		orderStatusBreakdown []models.NewOrderStatusBreakdown
		tablePerformance     []models.NewTablePerformance
		staffPerformance     []models.NewStaffPerformance
		hourlySales          []models.NewHourlySalesPoint
		dailySales           []models.NewDailySalesPoint
		menuItemOrderStats   []models.NewMenuItemOrderStat
	)

	g, gCtx := errgroup.WithContext(ctx)

	// Get overview for last 30 days
	g.Go(func() error {
		var err error
		overview, err = r.NewGetSalesOverview(gCtx, from, to)
		if err != nil {
			return fmt.Errorf("sales overview: %w", err)
		}
		return nil
	})

	// Get all-time stats card
	g.Go(func() error {
		var err error
		statsCard, err = r.NewGetSalesStatsCard(gCtx)
		if err != nil {
			return fmt.Errorf("stats card: %w", err)
		}
		return nil
	})

	// Get last 7 days trend
	g.Go(func() error {
		var err error
		dailyTrend, err = r.NewGetLast7DaysSalesTrend(gCtx)
		if err != nil {
			return fmt.Errorf("daily trend: %w", err)
		}
		return nil
	})

	// Get last 7 weeks trend
	g.Go(func() error {
		var err error
		weeklyTrend, err = r.NewGetLast7WeeksSalesTrend(gCtx)
		if err != nil {
			return fmt.Errorf("weekly trend: %w", err)
		}
		return nil
	})

	// Get last 7 months trend
	g.Go(func() error {
		var err error
		monthlyTrend, err = r.NewGetLast7MonthsSalesTrend(gCtx)
		if err != nil {
			return fmt.Errorf("monthly trend: %w", err)
		}
		return nil
	})

	// Get last 7 years trend
	g.Go(func() error {
		var err error
		yearlyTrend, err = r.NewGetLast7YearsSalesTrend(gCtx)
		if err != nil {
			return fmt.Errorf("yearly trend: %w", err)
		}
		return nil
	})

	// Get top selling items
	g.Go(func() error {
		var err error
		topSellingItems, err = r.NewGetTopSellingItems(gCtx, from, to, 10)
		if err != nil {
			return fmt.Errorf("top selling items: %w", err)
		}
		return nil
	})

	// Get top categories
	g.Go(func() error {
		var err error
		topCategories, err = r.NewGetTopCategories(gCtx, from, to, 5)
		if err != nil {
			return fmt.Errorf("top categories: %w", err)
		}
		return nil
	})

	// Get order status breakdown
	g.Go(func() error {
		var err error
		orderStatusBreakdown, err = r.NewGetOrderStatusBreakdown(gCtx, from, to)
		if err != nil {
			return fmt.Errorf("order status breakdown: %w", err)
		}
		return nil
	})

	// Get table performance
	g.Go(func() error {
		var err error
		tablePerformance, err = r.NewGetTablePerformance(gCtx, from, to)
		if err != nil {
			return fmt.Errorf("table performance: %w", err)
		}
		return nil
	})

	// Get staff performance
	g.Go(func() error {
		var err error
		staffPerformance, err = r.NewGetStaffPerformance(gCtx, from, to)
		if err != nil {
			return fmt.Errorf("staff performance: %w", err)
		}
		return nil
	})

	// Get hourly sales
	g.Go(func() error {
		var err error
		hourlySales, err = r.NewGetHourlySales(gCtx, from, to)
		if err != nil {
			return fmt.Errorf("hourly sales: %w", err)
		}
		return nil
	})

	// Get daily sales
	g.Go(func() error {
		var err error
		dailySales, err = r.NewGetDailySales(gCtx, from, to)
		if err != nil {
			return fmt.Errorf("daily sales: %w", err)
		}
		return nil
	})

	// Get menu item order stats
	g.Go(func() error {
		var err error
		menuItemOrderStats, err = r.NewGetLast7DaysMenuItemOrderStats(gCtx)
		if err != nil {
			return fmt.Errorf("menu item order stats: %w", err)
		}
		return nil
	})

	// Wait for all goroutines to complete
	if err := g.Wait(); err != nil {
		return nil, err
	}

	return &models.NewDefaultSalesResponse{
		Overview:             overview,
		StatsCard:            statsCard,
		DailyTrend:           dailyTrend,
		WeeklyTrend:          weeklyTrend,
		MonthlyTrend:         monthlyTrend,
		YearlyTrend:          yearlyTrend,
		TopSellingItems:      topSellingItems,
		TopCategories:        topCategories,
		OrderStatusBreakdown: orderStatusBreakdown,
		TablePerformance:     tablePerformance,
		StaffPerformance:     staffPerformance,
		HourlySales:          hourlySales,
		DailySales:           dailySales,
		MenuItemsOrderStats:  menuItemOrderStats,
	}, nil
}

// ────────────────────────────────────────────────────────────────────────────────
// CUSTOM RANGE SALES REPORT - Returns paginated data for specific date range
// ────────────────────────────────────────────────────────────────────────────────

func (r *reportRepo) NewGetCustomRangeSalesReport(ctx context.Context, req *models.NewSalesCustomRangeReportRequest) (*models.NewCustomRangeSalesResponse, error) {
	from := req.From
	to := req.To.Truncate(24 * time.Hour).Add(24*time.Hour - time.Second)

	// Set pagination defaults
	limit := req.Limit
	if limit <= 0 {
		limit = DefaultSalesTrendLimit
	}
	if limit > MaxSalesTrendLimit {
		limit = MaxSalesTrendLimit
	}
	page := req.Page
	if page < 0 {
		page = 0
	}

	var (
		overview             models.NewSalesOverviewCard
		statsCard            models.NewSalesStatsCard
		dailyTrend           *models.NewSalesPaginatedTrendPoints
		weeklyTrend          *models.NewSalesPaginatedTrendPoints
		monthlyTrend         *models.NewSalesPaginatedTrendPoints
		yearlyTrend          *models.NewSalesPaginatedTrendPoints
		topSellingItems      []models.NewTopSellingItem
		topCategories        []models.NewTopCategory
		orderStatusBreakdown []models.NewOrderStatusBreakdown
		tablePerformance     []models.NewTablePerformance
		staffPerformance     []models.NewStaffPerformance
		hourlySales          []models.NewHourlySalesPoint
		dailySales           []models.NewDailySalesPoint
		menuItemOrderStats   []models.NewMenuItemOrderStat
	)

	g, gCtx := errgroup.WithContext(ctx)

	// Get overview for custom date range
	g.Go(func() error {
		var err error
		overview, err = r.NewGetSalesOverview(gCtx, from, to)
		if err != nil {
			return fmt.Errorf("sales overview: %w", err)
		}
		return nil
	})

	// Get all-time stats card
	g.Go(func() error {
		var err error
		statsCard, err = r.NewGetSalesStatsCard(gCtx)
		if err != nil {
			return fmt.Errorf("stats card: %w", err)
		}
		return nil
	})

	// Get custom daily trend
	g.Go(func() error {
		var err error
		dailyTrend, err = r.NewGetCustomDailySalesTrend(gCtx, from, to, limit, page)
		if err != nil {
			return fmt.Errorf("custom daily trend: %w", err)
		}
		return nil
	})

	// Get custom weekly trend
	g.Go(func() error {
		var err error
		weeklyTrend, err = r.NewGetCustomWeeklySalesTrend(gCtx, from, to, limit, page)
		if err != nil {
			return fmt.Errorf("custom weekly trend: %w", err)
		}
		return nil
	})

	// Get custom monthly trend
	g.Go(func() error {
		var err error
		monthlyTrend, err = r.NewGetCustomMonthlySalesTrend(gCtx, from, to, limit, page)
		if err != nil {
			return fmt.Errorf("custom monthly trend: %w", err)
		}
		return nil
	})

	// Get custom yearly trend
	g.Go(func() error {
		var err error
		yearlyTrend, err = r.NewGetCustomYearlySalesTrend(gCtx, from, to, limit, page)
		if err != nil {
			return fmt.Errorf("custom yearly trend: %w", err)
		}
		return nil
	})

	// Get top selling items
	g.Go(func() error {
		var err error
		topSellingItems, err = r.NewGetTopSellingItems(gCtx, from, to, 10)
		if err != nil {
			return fmt.Errorf("top selling items: %w", err)
		}
		return nil
	})

	// Get top categories
	g.Go(func() error {
		var err error
		topCategories, err = r.NewGetTopCategories(gCtx, from, to, 5)
		if err != nil {
			return fmt.Errorf("top categories: %w", err)
		}
		return nil
	})

	// Get order status breakdown
	g.Go(func() error {
		var err error
		orderStatusBreakdown, err = r.NewGetOrderStatusBreakdown(gCtx, from, to)
		if err != nil {
			return fmt.Errorf("order status breakdown: %w", err)
		}
		return nil
	})

	// Get table performance
	g.Go(func() error {
		var err error
		tablePerformance, err = r.NewGetTablePerformance(gCtx, from, to)
		if err != nil {
			return fmt.Errorf("table performance: %w", err)
		}
		return nil
	})

	// Get staff performance
	g.Go(func() error {
		var err error
		staffPerformance, err = r.NewGetStaffPerformance(gCtx, from, to)
		if err != nil {
			return fmt.Errorf("staff performance: %w", err)
		}
		return nil
	})

	// Get hourly sales
	g.Go(func() error {
		var err error
		hourlySales, err = r.NewGetHourlySales(gCtx, from, to)
		if err != nil {
			return fmt.Errorf("hourly sales: %w", err)
		}
		return nil
	})

	// Get daily sales
	g.Go(func() error {
		var err error
		dailySales, err = r.NewGetDailySales(gCtx, from, to)
		if err != nil {
			return fmt.Errorf("daily sales: %w", err)
		}
		return nil
	})

	// Get menu item order stats
	g.Go(func() error {
		var err error
		menuItemOrderStats, err = r.NewGetCustomRangeMenuItemOrderStats(gCtx, from, to)
		if err != nil {
			return fmt.Errorf("menu item order stats: %w", err)
		}
		return nil
	})

	// Wait for all goroutines to complete
	if err := g.Wait(); err != nil {
		return nil, err
	}

	return &models.NewCustomRangeSalesResponse{
		Overview:             overview,
		StatsCard:            statsCard,
		DailyTrend:           dailyTrend,
		WeeklyTrend:          weeklyTrend,
		MonthlyTrend:         monthlyTrend,
		YearlyTrend:          yearlyTrend,
		TopSellingItems:      topSellingItems,
		TopCategories:        topCategories,
		OrderStatusBreakdown: orderStatusBreakdown,
		TablePerformance:     tablePerformance,
		StaffPerformance:     staffPerformance,
		HourlySales:          hourlySales,
		DailySales:           dailySales,
		MenuItemsOrderStats:  menuItemOrderStats,
	}, nil
}

// ────────────────────────────────────────────────────────────────────────────────
// Last 7 Days, Weeks, Months, Years Sales Trend Functions
// ────────────────────────────────────────────────────────────────────────────────

func (r *reportRepo) NewGetLast7DaysSalesTrend(ctx context.Context) ([]models.NewSalesTrendPoint, error) {
	query := `
		SELECT 
			TO_CHAR(DATE(o.created_at AT TIME ZONE 'Asia/Kathmandu'), 'YYYY-MM-DD') AS period,
			COUNT(DISTINCT o.id) AS orders,
			COALESCE(SUM(p.paid_amount), 0) AS revenue,
			COALESCE(SUM(p.discount), 0) AS discount
		FROM orders o
		LEFT JOIN payments p ON o.id = p.order_id
		WHERE o.created_at >= NOW() - INTERVAL '7 days'
		GROUP BY DATE(o.created_at AT TIME ZONE 'Asia/Kathmandu')
		ORDER BY DATE(o.created_at AT TIME ZONE 'Asia/Kathmandu') ASC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query last 7 days sales trend: %w", err)
	}
	defer rows.Close()

	var result []models.NewSalesTrendPoint
	for rows.Next() {
		var item models.NewSalesTrendPoint
		if err := rows.Scan(&item.Period, &item.Orders, &item.Revenue, &item.Discount); err != nil {
			return nil, fmt.Errorf("failed to scan last 7 days sales trend: %w", err)
		}
		result = append(result, item)
	}

	return result, nil
}

func (r *reportRepo) NewGetLast7WeeksSalesTrend(ctx context.Context) ([]models.NewSalesTrendPoint, error) {
	query := `
		SELECT 
			TO_CHAR(DATE_TRUNC('week', o.created_at AT TIME ZONE 'Asia/Kathmandu'), 'IYYY-"W"IW') AS period,
			COUNT(DISTINCT o.id) AS orders,
			COALESCE(SUM(p.paid_amount), 0) AS revenue,
			COALESCE(SUM(p.discount), 0) AS discount
		FROM orders o
		LEFT JOIN payments p ON o.id = p.order_id
		WHERE o.created_at >= NOW() - INTERVAL '7 weeks'
		GROUP BY DATE_TRUNC('week', o.created_at AT TIME ZONE 'Asia/Kathmandu')
		ORDER BY DATE_TRUNC('week', o.created_at AT TIME ZONE 'Asia/Kathmandu') ASC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query last 7 weeks sales trend: %w", err)
	}
	defer rows.Close()

	var result []models.NewSalesTrendPoint
	for rows.Next() {
		var item models.NewSalesTrendPoint
		if err := rows.Scan(&item.Period, &item.Orders, &item.Revenue, &item.Discount); err != nil {
			return nil, fmt.Errorf("failed to scan last 7 weeks sales trend: %w", err)
		}
		result = append(result, item)
	}

	return result, nil
}

func (r *reportRepo) NewGetLast7MonthsSalesTrend(ctx context.Context) ([]models.NewSalesTrendPoint, error) {
	query := `
		SELECT 
			TO_CHAR(DATE_TRUNC('month', o.created_at AT TIME ZONE 'Asia/Kathmandu'), 'YYYY-MM') AS period,
			COUNT(DISTINCT o.id) AS orders,
			COALESCE(SUM(p.paid_amount), 0) AS revenue,
			COALESCE(SUM(p.discount), 0) AS discount
		FROM orders o
		LEFT JOIN payments p ON o.id = p.order_id
		WHERE o.created_at >= NOW() - INTERVAL '7 months'
		GROUP BY DATE_TRUNC('month', o.created_at AT TIME ZONE 'Asia/Kathmandu')
		ORDER BY DATE_TRUNC('month', o.created_at AT TIME ZONE 'Asia/Kathmandu') ASC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query last 7 months sales trend: %w", err)
	}
	defer rows.Close()

	var result []models.NewSalesTrendPoint
	for rows.Next() {
		var item models.NewSalesTrendPoint
		if err := rows.Scan(&item.Period, &item.Orders, &item.Revenue, &item.Discount); err != nil {
			return nil, fmt.Errorf("failed to scan last 7 months sales trend: %w", err)
		}
		result = append(result, item)
	}

	return result, nil
}

func (r *reportRepo) NewGetLast7YearsSalesTrend(ctx context.Context) ([]models.NewSalesTrendPoint, error) {
	query := `
		SELECT 
			TO_CHAR(DATE_TRUNC('year', o.created_at AT TIME ZONE 'Asia/Kathmandu'), 'YYYY') AS period,
			COUNT(DISTINCT o.id) AS orders,
			COALESCE(SUM(p.paid_amount), 0) AS revenue,
			COALESCE(SUM(p.discount), 0) AS discount
		FROM orders o
		LEFT JOIN payments p ON o.id = p.order_id
		WHERE o.created_at >= NOW() - INTERVAL '7 years'
		GROUP BY DATE_TRUNC('year', o.created_at AT TIME ZONE 'Asia/Kathmandu')
		ORDER BY DATE_TRUNC('year', o.created_at AT TIME ZONE 'Asia/Kathmandu') ASC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query last 7 years sales trend: %w", err)
	}
	defer rows.Close()

	var result []models.NewSalesTrendPoint
	for rows.Next() {
		var item models.NewSalesTrendPoint
		if err := rows.Scan(&item.Period, &item.Orders, &item.Revenue, &item.Discount); err != nil {
			return nil, fmt.Errorf("failed to scan last 7 years sales trend: %w", err)
		}
		result = append(result, item)
	}

	return result, nil
}

// ────────────────────────────────────────────────────────────────────────────────
// Sales Overview Card
// ────────────────────────────────────────────────────────────────────────────────

func (r *reportRepo) NewGetSalesOverview(ctx context.Context, from, to time.Time) (models.NewSalesOverviewCard, error) {
	query := `
		WITH order_stats AS (
			SELECT 
				COUNT(DISTINCT o.id) AS total_orders,
				COUNT(CASE WHEN o.status = 'completed' THEN 1 END) AS completed_orders,
				COALESCE(SUM(p.paid_amount), 0) AS total_revenue,
				COALESCE(SUM(p.discount), 0) AS total_discounts,
				COALESCE(AVG(p.paid_amount), 0) AS avg_order_value,
				COALESCE(AVG(oi.item_count), 0) AS items_per_order
			FROM orders o
			LEFT JOIN payments p ON o.id = p.order_id
			LEFT JOIN (
				SELECT order_id, COUNT(*) AS item_count
				FROM order_items
				GROUP BY order_id
			) oi ON o.id = oi.order_id
			WHERE o.created_at BETWEEN $1 AND $2
		)
		SELECT 
			total_orders,
			total_revenue,
			total_discounts,
			avg_order_value,
			items_per_order,
			CASE 
				WHEN total_orders > 0 
				THEN (completed_orders::float / total_orders) * 100
				ELSE 0 
			END AS completion_rate,
			0 AS growth_percent
		FROM order_stats
	`

	var overview models.NewSalesOverviewCard
	if err := r.pool.QueryRow(ctx, query, from, to).Scan(
		&overview.TotalOrders,
		&overview.TotalRevenue,
		&overview.TotalDiscounts,
		&overview.AverageOrderValue,
		&overview.ItemsPerOrder,
		&overview.CompletionRate,
		&overview.GrowthPercent,
	); err != nil {
		return models.NewSalesOverviewCard{}, fmt.Errorf("failed to scan sales overview: %w", err)
	}

	// Calculate growth vs previous period
	diff := to.Sub(from)
	var prevOrders int
	if err := r.pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT id)
		FROM orders
		WHERE created_at BETWEEN $1 AND $2
	`, from.Add(-diff), from).Scan(&prevOrders); err != nil {
		return models.NewSalesOverviewCard{}, fmt.Errorf("failed to scan previous period: %w", err)
	}

	if prevOrders > 0 {
		overview.GrowthPercent = (float64(overview.TotalOrders-prevOrders) / float64(prevOrders)) * 100
	}

	return overview, nil
}

// ────────────────────────────────────────────────────────────────────────────────
// Sales Stats Card (All Time)
// ────────────────────────────────────────────────────────────────────────────────

func (r *reportRepo) NewGetSalesStatsCard(ctx context.Context) (models.NewSalesStatsCard, error) {
	query := `
		SELECT
			COUNT(DISTINCT o.id) AS total_orders,
			COUNT(CASE WHEN o.status = 'completed' THEN 1 END) AS completed_orders,
			COUNT(CASE WHEN o.status = 'cancelled' THEN 1 END) AS cancelled_orders,
			COALESCE(SUM(p.paid_amount), 0) AS total_revenue,
			COALESCE(SUM(p.discount), 0) AS total_discounts,
			COALESCE(AVG(p.paid_amount), 0) AS average_order_value,
			COUNT(DISTINCT o.customer_phone) AS unique_customers,
			CASE 
				WHEN COUNT(DISTINCT o.id) > 0 
				THEN (COUNT(CASE WHEN o.status = 'completed' THEN 1 END) * 100.0) / COUNT(DISTINCT o.id)
				ELSE 0 
			END AS completion_rate_percent
		FROM orders o
		LEFT JOIN payments p ON o.id = p.order_id
	`

	var stats models.NewSalesStatsCard
	if err := r.pool.QueryRow(ctx, query).Scan(
		&stats.TotalOrders,
		&stats.CompletedOrders,
		&stats.CancelledOrders,
		&stats.TotalRevenue,
		&stats.TotalDiscounts,
		&stats.AverageOrderValue,
		&stats.UniqueCustomers,
		&stats.CompletionRatePercent,
	); err != nil {
		return models.NewSalesStatsCard{}, fmt.Errorf("failed to scan sales stats card: %w", err)
	}

	return stats, nil
}

// ────────────────────────────────────────────────────────────────────────────────
// Custom Range Sales Trend Functions (With Pagination)
// ────────────────────────────────────────────────────────────────────────────────

func (r *reportRepo) NewGetCustomDailySalesTrend(ctx context.Context, from, to time.Time, limit, page int) (*models.NewSalesPaginatedTrendPoints, error) {
	offset := page * limit

	// Get total count first
	var total int
	countQuery := `
		SELECT COUNT(DISTINCT TO_CHAR(o.created_at AT TIME ZONE 'Asia/Kathmandu', 'YYYY-MM-DD'))
		FROM orders o
		WHERE o.created_at BETWEEN $1 AND $2
	`
	if err := r.pool.QueryRow(ctx, countQuery, from, to).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count custom daily sales trend: %w", err)
	}

	// Get paginated data
	query := `
		SELECT COALESCE(
			json_agg(
				json_build_object('period', period, 'orders', orders, 'revenue', revenue, 'discount', discount)
				ORDER BY period ASC
			), '[]'::json
		)
		FROM (
			SELECT
				TO_CHAR(DATE(o.created_at AT TIME ZONE 'Asia/Kathmandu'), 'YYYY-MM-DD') AS period,
				COUNT(DISTINCT o.id) AS orders,
				COALESCE(SUM(p.paid_amount), 0) AS revenue,
				COALESCE(SUM(p.discount), 0) AS discount
			FROM orders o
			LEFT JOIN payments p ON o.id = p.order_id
			WHERE o.created_at BETWEEN $1 AND $2
			GROUP BY DATE(o.created_at AT TIME ZONE 'Asia/Kathmandu')
			ORDER BY DATE(o.created_at AT TIME ZONE 'Asia/Kathmandu') ASC
			LIMIT $3 OFFSET $4
		) sub
	`

	var resultJSON []byte
	if err := r.pool.QueryRow(ctx, query, from, to, limit, offset).Scan(&resultJSON); err != nil {
		return nil, fmt.Errorf("failed to query custom daily sales trend: %w", err)
	}

	var result []models.NewSalesTrendPoint
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal custom daily sales trend: %w", err)
	}

	hasMore := (page+1)*limit < total
	nextPage := page + 1
	if !hasMore {
		nextPage = page
	}

	return &models.NewSalesPaginatedTrendPoints{
		Data: result,
		Pagination: models.NewPaginationInfo{
			Total:    total,
			HasMore:  hasMore,
			NextPage: nextPage,
			Limit:    limit,
			Page:     page,
		},
	}, nil
}

func (r *reportRepo) NewGetCustomWeeklySalesTrend(ctx context.Context, from, to time.Time, limit, page int) (*models.NewSalesPaginatedTrendPoints, error) {
	offset := page * limit

	var total int
	countQuery := `
		SELECT COUNT(DISTINCT DATE_TRUNC('week', o.created_at AT TIME ZONE 'Asia/Kathmandu'))
		FROM orders o
		WHERE o.created_at BETWEEN $1 AND $2
	`
	if err := r.pool.QueryRow(ctx, countQuery, from, to).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count custom weekly sales trend: %w", err)
	}

	query := `
		SELECT COALESCE(
			json_agg(
				json_build_object('period', period, 'orders', orders, 'revenue', revenue, 'discount', discount)
				ORDER BY week_start ASC
			), '[]'::json
		)
		FROM (
			SELECT
				TO_CHAR(DATE_TRUNC('week', o.created_at AT TIME ZONE 'Asia/Kathmandu'), 'IYYY-"W"IW') AS period,
				DATE_TRUNC('week', o.created_at AT TIME ZONE 'Asia/Kathmandu') AS week_start,
				COUNT(DISTINCT o.id) AS orders,
				COALESCE(SUM(p.paid_amount), 0) AS revenue,
				COALESCE(SUM(p.discount), 0) AS discount
			FROM orders o
			LEFT JOIN payments p ON o.id = p.order_id
			WHERE o.created_at BETWEEN $1 AND $2
			GROUP BY DATE_TRUNC('week', o.created_at AT TIME ZONE 'Asia/Kathmandu')
			ORDER BY week_start ASC
			LIMIT $3 OFFSET $4
		) sub
	`

	var resultJSON []byte
	if err := r.pool.QueryRow(ctx, query, from, to, limit, offset).Scan(&resultJSON); err != nil {
		return nil, fmt.Errorf("failed to query custom weekly sales trend: %w", err)
	}

	var result []models.NewSalesTrendPoint
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal custom weekly sales trend: %w", err)
	}

	hasMore := (page+1)*limit < total
	nextPage := page + 1
	if !hasMore {
		nextPage = page
	}

	return &models.NewSalesPaginatedTrendPoints{
		Data: result,
		Pagination: models.NewPaginationInfo{
			Total:    total,
			HasMore:  hasMore,
			NextPage: nextPage,
			Limit:    limit,
			Page:     page,
		},
	}, nil
}

func (r *reportRepo) NewGetCustomMonthlySalesTrend(ctx context.Context, from, to time.Time, limit, page int) (*models.NewSalesPaginatedTrendPoints, error) {
	offset := page * limit

	var total int
	countQuery := `
		SELECT COUNT(DISTINCT DATE_TRUNC('month', created_at AT TIME ZONE 'Asia/Kathmandu'))
		FROM orders
		WHERE created_at BETWEEN $1 AND $2
	`
	if err := r.pool.QueryRow(ctx, countQuery, from, to).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count custom monthly sales trend: %w", err)
	}

	query := `
		SELECT COALESCE(
			json_agg(
				json_build_object('period', period, 'orders', orders, 'revenue', revenue, 'discount', discount)
				ORDER BY month_start ASC
			), '[]'::json
		)
		FROM (
			SELECT
				TO_CHAR(DATE_TRUNC('month', o.created_at AT TIME ZONE 'Asia/Kathmandu'), 'YYYY-MM') AS period,
				DATE_TRUNC('month', o.created_at AT TIME ZONE 'Asia/Kathmandu') AS month_start,
				COUNT(DISTINCT o.id) AS orders,
				COALESCE(SUM(p.paid_amount), 0) AS revenue,
				COALESCE(SUM(p.discount), 0) AS discount
			FROM orders o
			LEFT JOIN payments p ON o.id = p.order_id
			WHERE o.created_at BETWEEN $1 AND $2
			GROUP BY DATE_TRUNC('month', o.created_at AT TIME ZONE 'Asia/Kathmandu')
			ORDER BY month_start ASC
			LIMIT $3 OFFSET $4
		) sub
	`

	var resultJSON []byte
	if err := r.pool.QueryRow(ctx, query, from, to, limit, offset).Scan(&resultJSON); err != nil {
		return nil, fmt.Errorf("failed to query custom monthly sales trend: %w", err)
	}

	var result []models.NewSalesTrendPoint
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal custom monthly sales trend: %w", err)
	}

	hasMore := (page+1)*limit < total
	nextPage := page + 1
	if !hasMore {
		nextPage = page
	}

	return &models.NewSalesPaginatedTrendPoints{
		Data: result,
		Pagination: models.NewPaginationInfo{
			Total:    total,
			HasMore:  hasMore,
			NextPage: nextPage,
			Limit:    limit,
			Page:     page,
		},
	}, nil
}

func (r *reportRepo) NewGetCustomYearlySalesTrend(ctx context.Context, from, to time.Time, limit, page int) (*models.NewSalesPaginatedTrendPoints, error) {
	offset := page * limit

	var total int
	countQuery := `
		SELECT COUNT(DISTINCT DATE_TRUNC('year', created_at AT TIME ZONE 'Asia/Kathmandu'))
		FROM orders
		WHERE created_at BETWEEN $1 AND $2
	`
	if err := r.pool.QueryRow(ctx, countQuery, from, to).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count custom yearly sales trend: %w", err)
	}

	query := `
		SELECT COALESCE(
			json_agg(
				json_build_object('period', period, 'orders', orders, 'revenue', revenue, 'discount', discount)
				ORDER BY year_start ASC
			), '[]'::json
		)
		FROM (
			SELECT
				TO_CHAR(DATE_TRUNC('year', o.created_at AT TIME ZONE 'Asia/Kathmandu'), 'YYYY') AS period,
				DATE_TRUNC('year', o.created_at AT TIME ZONE 'Asia/Kathmandu') AS year_start,
				COUNT(DISTINCT o.id) AS orders,
				COALESCE(SUM(p.paid_amount), 0) AS revenue,
				COALESCE(SUM(p.discount), 0) AS discount
			FROM orders o
			LEFT JOIN payments p ON o.id = p.order_id
			WHERE o.created_at BETWEEN $1 AND $2
			GROUP BY DATE_TRUNC('year', o.created_at AT TIME ZONE 'Asia/Kathmandu')
			ORDER BY year_start ASC
			LIMIT $3 OFFSET $4
		) sub
	`

	var resultJSON []byte
	if err := r.pool.QueryRow(ctx, query, from, to, limit, offset).Scan(&resultJSON); err != nil {
		return nil, fmt.Errorf("failed to query custom yearly sales trend: %w", err)
	}

	var result []models.NewSalesTrendPoint
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal custom yearly sales trend: %w", err)
	}

	hasMore := (page+1)*limit < total
	nextPage := page + 1
	if !hasMore {
		nextPage = page
	}

	return &models.NewSalesPaginatedTrendPoints{
		Data: result,
		Pagination: models.NewPaginationInfo{
			Total:    total,
			HasMore:  hasMore,
			NextPage: nextPage,
			Limit:    limit,
			Page:     page,
		},
	}, nil
}

// ────────────────────────────────────────────────────────────────────────────────
// Analytics Functions
// ────────────────────────────────────────────────────────────────────────────────

func (r *reportRepo) NewGetTopSellingItems(ctx context.Context, from, to time.Time, limit int) ([]models.NewTopSellingItem, error) {
	query := `
		SELECT 
			mi.id::text AS item_id,
			mi.name AS item_name,
			c.name AS category_name,
			COALESCE(SUM(oi.quantity), 0) AS quantity,
			COALESCE(SUM(oi.price * oi.quantity), 0) AS revenue,
			COUNT(DISTINCT oi.order_id) AS order_count
		FROM order_items oi
		JOIN menu_items mi ON oi.menu_item_id = mi.id
		LEFT JOIN categories c ON mi.category_id = c.id
		JOIN orders o ON oi.order_id = o.id
		WHERE o.created_at BETWEEN $1 AND $2
		GROUP BY mi.id, mi.name, c.name
		ORDER BY quantity DESC
		LIMIT $3
	`

	rows, err := r.pool.Query(ctx, query, from, to, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query top selling items: %w", err)
	}
	defer rows.Close()

	var result []models.NewTopSellingItem
	for rows.Next() {
		var item models.NewTopSellingItem
		if err := rows.Scan(&item.ItemID, &item.ItemName, &item.CategoryName, &item.Quantity, &item.Revenue, &item.OrderCount); err != nil {
			return nil, fmt.Errorf("failed to scan top selling item: %w", err)
		}
		result = append(result, item)
	}

	return result, nil
}

func (r *reportRepo) NewGetTopCategories(ctx context.Context, from, to time.Time, limit int) ([]models.NewTopCategory, error) {
	query := `
		SELECT 
			c.id::text AS category_id,
			c.name AS category_name,
			COUNT(DISTINCT oi.order_id) AS orders,
			COALESCE(SUM(oi.price * oi.quantity), 0) AS revenue,
			COALESCE(SUM(oi.quantity), 0) AS items_count
		FROM categories c
		JOIN menu_items mi ON c.id = mi.category_id
		JOIN order_items oi ON mi.id = oi.menu_item_id
		JOIN orders o ON oi.order_id = o.id
		WHERE o.created_at BETWEEN $1 AND $2
		GROUP BY c.id, c.name
		ORDER BY revenue DESC
		LIMIT $3
	`

	rows, err := r.pool.Query(ctx, query, from, to, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query top categories: %w", err)
	}
	defer rows.Close()

	var result []models.NewTopCategory
	for rows.Next() {
		var item models.NewTopCategory
		if err := rows.Scan(&item.CategoryID, &item.CategoryName, &item.Orders, &item.Revenue, &item.ItemsCount); err != nil {
			return nil, fmt.Errorf("failed to scan top category: %w", err)
		}
		result = append(result, item)
	}

	return result, nil
}

func (r *reportRepo) NewGetOrderStatusBreakdown(ctx context.Context, from, to time.Time) ([]models.NewOrderStatusBreakdown, error) {
	query := `
		WITH totals AS (
			SELECT COUNT(*) AS total_orders
			FROM orders
			WHERE created_at BETWEEN $1 AND $2
		)
		SELECT
			o.status::text AS status,
			COUNT(o.id) AS count,
			COALESCE(SUM(p.paid_amount), 0) AS revenue,
			CASE 
				WHEN (SELECT total_orders FROM totals) > 0 
				THEN (COUNT(o.id) * 100.0) / (SELECT total_orders FROM totals)
				ELSE 0 
			END AS percent
		FROM orders o
		LEFT JOIN payments p ON o.id = p.order_id
		WHERE o.created_at BETWEEN $1 AND $2
		GROUP BY o.status
		ORDER BY count DESC
	`

	rows, err := r.pool.Query(ctx, query, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to query order status breakdown: %w", err)
	}
	defer rows.Close()

	var result []models.NewOrderStatusBreakdown
	for rows.Next() {
		var item models.NewOrderStatusBreakdown
		if err := rows.Scan(&item.Status, &item.Count, &item.Revenue, &item.Percent); err != nil {
			return nil, fmt.Errorf("failed to scan order status: %w", err)
		}
		result = append(result, item)
	}

	return result, nil
}

func (r *reportRepo) NewGetTablePerformance(ctx context.Context, from, to time.Time) ([]models.NewTablePerformance, error) {
	query := `
		SELECT
			COALESCE(ts.table_number, 0) AS table_number,
			COUNT(DISTINCT o.id) AS total_orders,
			COALESCE(SUM(p.paid_amount), 0) AS total_revenue,
			COALESCE(AVG(p.paid_amount), 0) AS average_order_value,
			COUNT(DISTINCT o.customer_phone) AS total_customers
		FROM table_session ts
		JOIN orders o ON ts.id = o.table_session_id
		LEFT JOIN payments p ON o.id = p.order_id
		WHERE o.created_at BETWEEN $1 AND $2
		GROUP BY ts.table_number
		ORDER BY total_revenue DESC
	`

	rows, err := r.pool.Query(ctx, query, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to query table performance: %w", err)
	}
	defer rows.Close()

	var result []models.NewTablePerformance
	for rows.Next() {
		var item models.NewTablePerformance
		if err := rows.Scan(&item.TableNumber, &item.TotalOrders, &item.TotalRevenue, &item.AverageOrderValue, &item.TotalCustomers); err != nil {
			return nil, fmt.Errorf("failed to scan table performance: %w", err)
		}
		result = append(result, item)
	}

	return result, nil
}

func (r *reportRepo) NewGetStaffPerformance(ctx context.Context, from, to time.Time) ([]models.NewStaffPerformance, error) {
	query := `
		SELECT
			u.id::text AS staff_id,
			u.name AS staff_name,
			u.role::text AS role,
			COUNT(DISTINCT o.id) AS orders_served,
			COALESCE(SUM(p.paid_amount), 0) AS total_revenue,
			COALESCE(AVG(p.paid_amount), 0) AS average_order_value
		FROM users u
		JOIN orders o ON u.id = o.waiter_id
		LEFT JOIN payments p ON o.id = p.order_id
		WHERE o.created_at BETWEEN $1 AND $2
			AND u.role IN ('waiter', 'cashier')
		GROUP BY u.id, u.name, u.role
		ORDER BY total_revenue DESC
	`

	rows, err := r.pool.Query(ctx, query, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to query staff performance: %w", err)
	}
	defer rows.Close()

	var result []models.NewStaffPerformance
	for rows.Next() {
		var item models.NewStaffPerformance
		if err := rows.Scan(&item.StaffID, &item.StaffName, &item.Role, &item.OrdersServed, &item.TotalRevenue, &item.AverageOrderValue); err != nil {
			return nil, fmt.Errorf("failed to scan staff performance: %w", err)
		}
		result = append(result, item)
	}

	return result, nil
}

func (r *reportRepo) NewGetHourlySales(ctx context.Context, from, to time.Time) ([]models.NewHourlySalesPoint, error) {
	query := `
		SELECT
			EXTRACT(HOUR FROM o.created_at AT TIME ZONE 'Asia/Kathmandu')::int AS hour,
			COUNT(DISTINCT o.id) AS orders,
			COALESCE(SUM(p.paid_amount), 0) AS revenue
		FROM orders o
		LEFT JOIN payments p ON o.id = p.order_id
		WHERE o.created_at BETWEEN $1 AND $2
		GROUP BY EXTRACT(HOUR FROM o.created_at AT TIME ZONE 'Asia/Kathmandu')
		ORDER BY hour ASC
	`

	rows, err := r.pool.Query(ctx, query, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to query hourly sales: %w", err)
	}
	defer rows.Close()

	var result []models.NewHourlySalesPoint
	for rows.Next() {
		var item models.NewHourlySalesPoint
		if err := rows.Scan(&item.Hour, &item.Orders, &item.Revenue); err != nil {
			return nil, fmt.Errorf("failed to scan hourly sales: %w", err)
		}
		result = append(result, item)
	}

	return result, nil
}

func (r *reportRepo) NewGetDailySales(ctx context.Context, from, to time.Time) ([]models.NewDailySalesPoint, error) {
	query := `
		SELECT
			TRIM(TO_CHAR(o.created_at AT TIME ZONE 'Asia/Kathmandu', 'Day')) AS day_of_week,
			COUNT(DISTINCT o.id) AS orders,
			COALESCE(SUM(p.paid_amount), 0) AS revenue
		FROM orders o
		LEFT JOIN payments p ON o.id = p.order_id
		WHERE o.created_at BETWEEN $1 AND $2
		GROUP BY TO_CHAR(o.created_at AT TIME ZONE 'Asia/Kathmandu', 'Day'), 
				 EXTRACT(DOW FROM o.created_at AT TIME ZONE 'Asia/Kathmandu')
		ORDER BY EXTRACT(DOW FROM o.created_at AT TIME ZONE 'Asia/Kathmandu') ASC
	`

	rows, err := r.pool.Query(ctx, query, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to query daily sales: %w", err)
	}
	defer rows.Close()

	var result []models.NewDailySalesPoint
	for rows.Next() {
		var item models.NewDailySalesPoint
		if err := rows.Scan(&item.DayOfWeek, &item.Orders, &item.Revenue); err != nil {
			return nil, fmt.Errorf("failed to scan daily sales: %w", err)
		}
		result = append(result, item)
	}

	return result, nil
}

// ────────────────────────────────────────────────────────────────────────────────
// DEFAULT REPORT API - Returns 7 periods of each trend (7 days, 7 weeks, 7 months, 7 years)
// This data will be cached and refreshed nightly
// ────────────────────────────────────────────────────────────────────────────────

func (r *reportRepo) NewGetDefaultRevenueReport(ctx context.Context) (*models.NewDefaultRevenueResponse, error) {
	// Overview is last 30 days
	to := time.Now()
	from := to.AddDate(0, 0, -30)

	var (
		overview       models.NewRevenueOverviewCard
		statsCard      models.NewRevenueStatsCard
		dailyTrend     []models.NewTrendPoint
		weeklyTrend    []models.NewTrendPoint
		monthlyTrend   []models.NewTrendPoint
		yearlyTrend    []models.NewTrendPoint
		paymentMethods []models.NewPaymentMethodBreakdown
		gateways       []models.NewGatewayBreakdown
		discounts      models.NewDiscountAnalysis
		peakHours      []models.NewPeakHourPoint
		peakDays       []models.NewPeakDayPoint
	)

	g, gCtx := errgroup.WithContext(ctx)

	// Get overview for last 30 days
	g.Go(func() error {
		var err error
		overview, err = r.NewGetRevenueOverview(gCtx, from, to)
		if err != nil {
			return fmt.Errorf("revenue overview: %w", err)
		}
		return nil
	})

	// Get all-time stats card
	g.Go(func() error {
		var err error
		statsCard, err = r.NewGetRevenueStatsCard(gCtx)
		if err != nil {
			return fmt.Errorf("stats card: %w", err)
		}
		return nil
	})

	// Get last 7 days trend
	g.Go(func() error {
		var err error
		dailyTrend, err = r.NewGetLast7DaysTrend(gCtx)
		if err != nil {
			return fmt.Errorf("daily trend: %w", err)
		}
		return nil
	})

	// Get last 7 weeks trend
	g.Go(func() error {
		var err error
		weeklyTrend, err = r.NewGetLast7WeeksTrend(gCtx)
		if err != nil {
			return fmt.Errorf("weekly trend: %w", err)
		}
		return nil
	})

	// Get last 7 months trend
	g.Go(func() error {
		var err error
		monthlyTrend, err = r.NewGetLast7MonthsTrend(gCtx)
		if err != nil {
			return fmt.Errorf("monthly trend: %w", err)
		}
		return nil
	})

	// Get last 7 years trend
	g.Go(func() error {
		var err error
		yearlyTrend, err = r.NewGetLast7YearsTrend(gCtx)
		if err != nil {
			return fmt.Errorf("yearly trend: %w", err)
		}
		return nil
	})

	// Get payment methods breakdown
	g.Go(func() error {
		var err error
		paymentMethods, err = r.NewGetPaymentMethodBreakdown(gCtx, from, to)
		if err != nil {
			return fmt.Errorf("payment methods: %w", err)
		}
		return nil
	})

	// Get gateways breakdown
	g.Go(func() error {
		var err error
		gateways, err = r.NewGetGatewayBreakdown(gCtx, from, to)
		if err != nil {
			return fmt.Errorf("gateways: %w", err)
		}
		return nil
	})

	// Get discounts analysis
	g.Go(func() error {
		var err error
		discounts, err = r.NewGetDiscountAnalysis(gCtx, from, to)
		if err != nil {
			return fmt.Errorf("discounts: %w", err)
		}
		return nil
	})

	// Get peak hours
	g.Go(func() error {
		var err error
		peakHours, err = r.NewGetPeakHours(gCtx, from, to)
		if err != nil {
			return fmt.Errorf("peak hours: %w", err)
		}
		return nil
	})

	// Get peak days
	g.Go(func() error {
		var err error
		peakDays, err = r.NewGetPeakDays(gCtx, from, to)
		if err != nil {
			return fmt.Errorf("peak days: %w", err)
		}
		return nil
	})

	// Wait for all goroutines to complete
	if err := g.Wait(); err != nil {
		return nil, err
	}

	return &models.NewDefaultRevenueResponse{
		Overview:       overview,
		StatsCard:      statsCard,
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

// ────────────────────────────────────────────────────────────────────────────────
// CUSTOM RANGE REPORT API - Returns paginated data for specific date range
// This always hits the database directly (no caching)
// ────────────────────────────────────────────────────────────────────────────────

func (r *reportRepo) NewGetCustomRangeRevenueReport(ctx context.Context, req *models.NewCustomRangeReportRequest) (*models.NewCustomRangeRevenueResponse, error) {
	from := req.From
	to := req.To.Truncate(24 * time.Hour).Add(24*time.Hour - time.Second)

	// Set pagination defaults
	limit := req.Limit
	if limit <= 0 {
		limit = DefaultTrendLimit
	}
	if limit > MaxTrendLimit {
		limit = MaxTrendLimit
	}
	page := req.Page
	if page < 0 {
		page = 0
	}

	var (
		overview       models.NewRevenueOverviewCard
		statsCard      models.NewRevenueStatsCard
		dailyTrend     *models.NewPaginatedTrendPoints
		weeklyTrend    *models.NewPaginatedTrendPoints
		monthlyTrend   *models.NewPaginatedTrendPoints
		yearlyTrend    *models.NewPaginatedTrendPoints
		paymentMethods []models.NewPaymentMethodBreakdown
		gateways       []models.NewGatewayBreakdown
		discounts      models.NewDiscountAnalysis
		peakHours      []models.NewPeakHourPoint
		peakDays       []models.NewPeakDayPoint
	)

	g, gCtx := errgroup.WithContext(ctx)

	// Get overview for custom date range
	g.Go(func() error {
		var err error
		overview, err = r.NewGetRevenueOverview(gCtx, from, to)
		if err != nil {
			return fmt.Errorf("revenue overview: %w", err)
		}
		return nil
	})

	// Get all-time stats card
	g.Go(func() error {
		var err error
		statsCard, err = r.NewGetRevenueStatsCard(gCtx)
		if err != nil {
			return fmt.Errorf("stats card: %w", err)
		}
		return nil
	})

	// Get custom daily trend
	g.Go(func() error {
		var err error
		dailyTrend, err = r.NewGetCustomDailyTrend(gCtx, from, to, limit, page)
		if err != nil {
			return fmt.Errorf("custom daily trend: %w", err)
		}
		return nil
	})

	// Get custom weekly trend
	g.Go(func() error {
		var err error
		weeklyTrend, err = r.NewGetCustomWeeklyTrend(gCtx, from, to, limit, page)
		if err != nil {
			return fmt.Errorf("custom weekly trend: %w", err)
		}
		return nil
	})

	// Get custom monthly trend
	g.Go(func() error {
		var err error
		monthlyTrend, err = r.NewGetCustomMonthlyTrend(gCtx, from, to, limit, page)
		if err != nil {
			return fmt.Errorf("custom monthly trend: %w", err)
		}
		return nil
	})

	// Get custom yearly trend
	g.Go(func() error {
		var err error
		yearlyTrend, err = r.NewGetCustomYearlyTrend(gCtx, from, to, limit, page)
		if err != nil {
			return fmt.Errorf("custom yearly trend: %w", err)
		}
		return nil
	})

	// Get payment methods breakdown
	g.Go(func() error {
		var err error
		paymentMethods, err = r.NewGetPaymentMethodBreakdown(gCtx, from, to)
		if err != nil {
			return fmt.Errorf("payment methods: %w", err)
		}
		return nil
	})

	// Get gateways breakdown
	g.Go(func() error {
		var err error
		gateways, err = r.NewGetGatewayBreakdown(gCtx, from, to)
		if err != nil {
			return fmt.Errorf("gateways: %w", err)
		}
		return nil
	})

	// Get discounts analysis
	g.Go(func() error {
		var err error
		discounts, err = r.NewGetDiscountAnalysis(gCtx, from, to)
		if err != nil {
			return fmt.Errorf("discounts: %w", err)
		}
		return nil
	})

	// Get peak hours
	g.Go(func() error {
		var err error
		peakHours, err = r.NewGetPeakHours(gCtx, from, to)
		if err != nil {
			return fmt.Errorf("peak hours: %w", err)
		}
		return nil
	})

	// Get peak days
	g.Go(func() error {
		var err error
		peakDays, err = r.NewGetPeakDays(gCtx, from, to)
		if err != nil {
			return fmt.Errorf("peak days: %w", err)
		}
		return nil
	})

	// Wait for all goroutines to complete
	if err := g.Wait(); err != nil {
		return nil, err
	}

	return &models.NewCustomRangeRevenueResponse{
		Overview:       overview,
		StatsCard:      statsCard,
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

// ────────────────────────────────────────────────────────────────────────────────
// NEW: Last 7 Days, Weeks, Months, Years Trend Functions (No Pagination)
// ────────────────────────────────────────────────────────────────────────────────
// ────────────────────────────────────────────────────────────────────────────────
// NEW: Last 7 Days, Weeks, Months, Years Trend Functions (No Pagination)
// ────────────────────────────────────────────────────────────────────────────────

func (r *reportRepo) NewGetLast7DaysTrend(ctx context.Context) ([]models.NewTrendPoint, error) {
	query := `
		SELECT 
			TO_CHAR(DATE(p.created_at AT TIME ZONE 'Asia/Kathmandu'), 'YYYY-MM-DD') AS period,
			COALESCE(SUM(p.paid_amount), 0) AS revenue,
			COUNT(p.id) AS orders
		FROM payments p
		WHERE p.created_at >= NOW() - INTERVAL '7 days'
		GROUP BY DATE(p.created_at AT TIME ZONE 'Asia/Kathmandu')
		ORDER BY DATE(p.created_at AT TIME ZONE 'Asia/Kathmandu') ASC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query last 7 days trend: %w", err)
	}
	defer rows.Close()

	var result []models.NewTrendPoint
	for rows.Next() {
		var item models.NewTrendPoint
		if err := rows.Scan(&item.Period, &item.Revenue, &item.Orders); err != nil {
			return nil, fmt.Errorf("failed to scan last 7 days trend: %w", err)
		}
		result = append(result, item)
	}

	return result, nil
}

func (r *reportRepo) NewGetLast7WeeksTrend(ctx context.Context) ([]models.NewTrendPoint, error) {
	query := `
		SELECT 
			TO_CHAR(DATE_TRUNC('week', p.created_at AT TIME ZONE 'Asia/Kathmandu'), 'IYYY-"W"IW') AS period,
			COALESCE(SUM(p.paid_amount), 0) AS revenue,
			COUNT(p.id) AS orders
		FROM payments p
		WHERE p.created_at >= NOW() - INTERVAL '7 weeks'
		GROUP BY DATE_TRUNC('week', p.created_at AT TIME ZONE 'Asia/Kathmandu')
		ORDER BY DATE_TRUNC('week', p.created_at AT TIME ZONE 'Asia/Kathmandu') ASC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query last 7 weeks trend: %w", err)
	}
	defer rows.Close()

	var result []models.NewTrendPoint
	for rows.Next() {
		var item models.NewTrendPoint
		if err := rows.Scan(&item.Period, &item.Revenue, &item.Orders); err != nil {
			return nil, fmt.Errorf("failed to scan last 7 weeks trend: %w", err)
		}
		result = append(result, item)
	}

	return result, nil
}

func (r *reportRepo) NewGetLast7MonthsTrend(ctx context.Context) ([]models.NewTrendPoint, error) {
	query := `
		SELECT 
			TO_CHAR(DATE_TRUNC('month', p.created_at AT TIME ZONE 'Asia/Kathmandu'), 'YYYY-MM') AS period,
			COALESCE(SUM(p.paid_amount), 0) AS revenue,
			COUNT(p.id) AS orders
		FROM payments p
		WHERE p.created_at >= NOW() - INTERVAL '7 months'
		GROUP BY DATE_TRUNC('month', p.created_at AT TIME ZONE 'Asia/Kathmandu')
		ORDER BY DATE_TRUNC('month', p.created_at AT TIME ZONE 'Asia/Kathmandu') ASC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query last 7 months trend: %w", err)
	}
	defer rows.Close()

	var result []models.NewTrendPoint
	for rows.Next() {
		var item models.NewTrendPoint
		if err := rows.Scan(&item.Period, &item.Revenue, &item.Orders); err != nil {
			return nil, fmt.Errorf("failed to scan last 7 months trend: %w", err)
		}
		result = append(result, item)
	}

	return result, nil
}

func (r *reportRepo) NewGetLast7YearsTrend(ctx context.Context) ([]models.NewTrendPoint, error) {
	query := `
		SELECT 
			TO_CHAR(DATE_TRUNC('year', p.created_at AT TIME ZONE 'Asia/Kathmandu'), 'YYYY') AS period,
			COALESCE(SUM(p.paid_amount), 0) AS revenue,
			COUNT(p.id) AS orders
		FROM payments p
		WHERE p.created_at >= NOW() - INTERVAL '7 years'
		GROUP BY DATE_TRUNC('year', p.created_at AT TIME ZONE 'Asia/Kathmandu')
		ORDER BY DATE_TRUNC('year', p.created_at AT TIME ZONE 'Asia/Kathmandu') ASC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query last 7 years trend: %w", err)
	}
	defer rows.Close()

	var result []models.NewTrendPoint
	for rows.Next() {
		var item models.NewTrendPoint
		if err := rows.Scan(&item.Period, &item.Revenue, &item.Orders); err != nil {
			return nil, fmt.Errorf("failed to scan last 7 years trend: %w", err)
		}
		result = append(result, item)
	}

	return result, nil
}

// ────────────────────────────────────────────────────────────────────────────────
// Stats Card (All Time)
// ────────────────────────────────────────────────────────────────────────────────

func (r *reportRepo) NewGetRevenueStatsCard(ctx context.Context) (models.NewRevenueStatsCard, error) {
	query := `
		SELECT
			COALESCE(SUM(p.paid_amount + p.discount), 0) AS total_gross_revenue,
			COALESCE(SUM(p.paid_amount), 0)              AS total_net_revenue,
			COUNT(p.id)                                  AS total_orders,
			COALESCE(SUM(p.discount), 0)                 AS total_discounts,
			COALESCE(AVG(p.paid_amount), 0)              AS avg_order_value,
			COUNT(DISTINCT o.customer_phone)             AS total_customers,
			CASE 
				WHEN SUM(p.paid_amount + p.discount) > 0 
				THEN (SUM(p.discount) * 100.0) / SUM(p.paid_amount + p.discount)
				ELSE 0 
			END AS discount_rate_percent
		FROM payments p
		LEFT JOIN orders o ON p.order_id = o.id
	`

	var stats models.NewRevenueStatsCard
	if err := r.pool.QueryRow(ctx, query).Scan(
		&stats.TotalGrossRevenue,
		&stats.TotalNetRevenue,
		&stats.TotalOrders,
		&stats.TotalDiscounts,
		&stats.AverageOrderValue,
		&stats.TotalCustomers,
		&stats.DiscountRatePercent,
	); err != nil {
		return models.NewRevenueStatsCard{}, fmt.Errorf("failed to scan stats card: %w", err)
	}

	return stats, nil
}

// ────────────────────────────────────────────────────────────────────────────────
// Overview Card
// ────────────────────────────────────────────────────────────────────────────────

func (r *reportRepo) NewGetRevenueOverview(ctx context.Context, from, to time.Time) (models.NewRevenueOverviewCard, error) {
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

	var overview models.NewRevenueOverviewCard
	if err := r.pool.QueryRow(ctx, query, from, to).Scan(
		&overview.GrossRevenue,
		&overview.NetRevenue,
		&overview.TotalDiscounts,
		&overview.TotalOrders,
		&overview.AverageOrderValue,
	); err != nil {
		return models.NewRevenueOverviewCard{}, fmt.Errorf("failed to scan revenue overview: %w", err)
	}

	// Calculate growth vs previous period
	diff := to.Sub(from)
	var prevNet float64
	if err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(paid_amount), 0)
		FROM payments
		WHERE created_at BETWEEN $1 AND $2
	`, from.Add(-diff), from).Scan(&prevNet); err != nil {
		return models.NewRevenueOverviewCard{}, fmt.Errorf("failed to scan previous period: %w", err)
	}

	if prevNet > 0 {
		overview.GrowthPercent = ((overview.NetRevenue - prevNet) / prevNet) * 100
	}

	return overview, nil
}

// ────────────────────────────────────────────────────────────────────────────────
// Default Trend Functions (Last N periods - No pagination, just returns N items)
// ────────────────────────────────────────────────────────────────────────────────

func (r *reportRepo) NewGetDefaultDailyTrend(ctx context.Context, limit int) (*models.NewPaginatedTrendPoints, error) {
	query := `
		WITH daily_data AS (
			SELECT
				TO_CHAR(p.created_at AT TIME ZONE 'Asia/Kathmandu', 'YYYY-MM-DD') AS period,
				COALESCE(SUM(p.paid_amount), 0)                                    AS revenue,
				COUNT(p.id)                                                        AS orders,
				MAX(p.created_at) AS sort_date
			FROM payments p
			WHERE p.created_at >= NOW() - INTERVAL '1 day' * $1
			GROUP BY TO_CHAR(p.created_at AT TIME ZONE 'Asia/Kathmandu', 'YYYY-MM-DD')
		)
		SELECT COALESCE(
			json_agg(
				json_build_object('period', period, 'revenue', revenue, 'orders', orders)
				ORDER BY sort_date ASC
			), '[]'::json
		)
		FROM (
			SELECT * FROM daily_data
			ORDER BY sort_date DESC
			LIMIT $2
		) sub
	`

	var resultJSON []byte
	if err := r.pool.QueryRow(ctx, query, limit*2, limit).Scan(&resultJSON); err != nil {
		return nil, fmt.Errorf("failed to query default daily trend: %w", err)
	}

	var result []models.NewTrendPoint
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal default daily trend: %w", err)
	}

	return &models.NewPaginatedTrendPoints{
		Data: result,
		Pagination: models.NewPaginationInfo{
			Total:    len(result),
			HasMore:  false, // Default trends don't have more pages
			NextPage: 0,
			Limit:    limit,
			Page:     0,
		},
	}, nil
}

func (r *reportRepo) NewGetDefaultWeeklyTrend(ctx context.Context, limit int) (*models.NewPaginatedTrendPoints, error) {
	query := `
		WITH weekly_data AS (
			SELECT
				TO_CHAR(DATE_TRUNC('week', p.created_at AT TIME ZONE 'Asia/Kathmandu'), 'IYYY-"W"IW') AS period,
				DATE_TRUNC('week', p.created_at AT TIME ZONE 'Asia/Kathmandu') AS week_start,
				COALESCE(SUM(p.paid_amount), 0) AS revenue,
				COUNT(p.id) AS orders
			FROM payments p
			WHERE p.created_at >= NOW() - INTERVAL '1 week' * $1
			GROUP BY DATE_TRUNC('week', p.created_at AT TIME ZONE 'Asia/Kathmandu')
		)
		SELECT COALESCE(
			json_agg(
				json_build_object('period', period, 'revenue', revenue, 'orders', orders)
				ORDER BY week_start ASC
			), '[]'::json
		)
		FROM (
			SELECT * FROM weekly_data
			ORDER BY week_start DESC
			LIMIT $2
		) sub
	`

	var resultJSON []byte
	if err := r.pool.QueryRow(ctx, query, limit*2, limit).Scan(&resultJSON); err != nil {
		return nil, fmt.Errorf("failed to query default weekly trend: %w", err)
	}

	var result []models.NewTrendPoint
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal default weekly trend: %w", err)
	}

	return &models.NewPaginatedTrendPoints{
		Data: result,
		Pagination: models.NewPaginationInfo{
			Total:    len(result),
			HasMore:  false,
			NextPage: 0,
			Limit:    limit,
			Page:     0,
		},
	}, nil
}

func (r *reportRepo) NewGetDefaultMonthlyTrend(ctx context.Context, limit int) (*models.NewPaginatedTrendPoints, error) {
	query := `
		WITH monthly_data AS (
			SELECT
				TO_CHAR(DATE_TRUNC('month', p.created_at AT TIME ZONE 'Asia/Kathmandu'), 'YYYY-MM') AS period,
				DATE_TRUNC('month', p.created_at AT TIME ZONE 'Asia/Kathmandu') AS month_start,
				COALESCE(SUM(p.paid_amount), 0) AS revenue,
				COUNT(p.id) AS orders
			FROM payments p
			GROUP BY DATE_TRUNC('month', p.created_at AT TIME ZONE 'Asia/Kathmandu')
		)
		SELECT COALESCE(
			json_agg(
				json_build_object('period', period, 'revenue', revenue, 'orders', orders)
				ORDER BY month_start ASC
			), '[]'::json
		)
		FROM (
			SELECT * FROM monthly_data
			ORDER BY month_start DESC
			LIMIT $1
		) sub
	`

	var resultJSON []byte
	if err := r.pool.QueryRow(ctx, query, limit).Scan(&resultJSON); err != nil {
		return nil, fmt.Errorf("failed to query default monthly trend: %w", err)
	}

	var result []models.NewTrendPoint
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal default monthly trend: %w", err)
	}

	return &models.NewPaginatedTrendPoints{
		Data: result,
		Pagination: models.NewPaginationInfo{
			Total:    len(result),
			HasMore:  false,
			NextPage: 0,
			Limit:    limit,
			Page:     0,
		},
	}, nil
}

func (r *reportRepo) NewGetDefaultYearlyTrend(ctx context.Context, limit int) (*models.NewPaginatedTrendPoints, error) {
	query := `
		WITH yearly_data AS (
			SELECT
				TO_CHAR(DATE_TRUNC('year', p.created_at AT TIME ZONE 'Asia/Kathmandu'), 'YYYY') AS period,
				DATE_TRUNC('year', p.created_at AT TIME ZONE 'Asia/Kathmandu') AS year_start,
				COALESCE(SUM(p.paid_amount), 0) AS revenue,
				COUNT(p.id) AS orders
			FROM payments p
			GROUP BY DATE_TRUNC('year', p.created_at AT TIME ZONE 'Asia/Kathmandu')
		)
		SELECT COALESCE(
			json_agg(
				json_build_object('period', period, 'revenue', revenue, 'orders', orders)
				ORDER BY year_start ASC
			), '[]'::json
		)
		FROM (
			SELECT * FROM yearly_data
			ORDER BY year_start DESC
			LIMIT $1
		) sub
	`

	var resultJSON []byte
	if err := r.pool.QueryRow(ctx, query, limit).Scan(&resultJSON); err != nil {
		return nil, fmt.Errorf("failed to query default yearly trend: %w", err)
	}

	var result []models.NewTrendPoint
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal default yearly trend: %w", err)
	}

	return &models.NewPaginatedTrendPoints{
		Data: result,
		Pagination: models.NewPaginationInfo{
			Total:    len(result),
			HasMore:  false,
			NextPage: 0,
			Limit:    limit,
			Page:     0,
		},
	}, nil
}

// ────────────────────────────────────────────────────────────────────────────────
// Custom Range Trend Functions (With Pagination)
// ────────────────────────────────────────────────────────────────────────────────

func (r *reportRepo) NewGetCustomDailyTrend(ctx context.Context, from, to time.Time, limit, page int) (*models.NewPaginatedTrendPoints, error) {
	offset := page * limit

	// Get total count first
	var total int
	countQuery := `
		SELECT COUNT(DISTINCT TO_CHAR(p.created_at AT TIME ZONE 'Asia/Kathmandu', 'YYYY-MM-DD'))
		FROM payments p
		WHERE p.created_at BETWEEN $1 AND $2
	`
	if err := r.pool.QueryRow(ctx, countQuery, from, to).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count custom daily trend: %w", err)
	}

	// Get paginated data
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
			ORDER BY period ASC
			LIMIT $3 OFFSET $4
		) sub
	`

	var resultJSON []byte
	if err := r.pool.QueryRow(ctx, query, from, to, limit, offset).Scan(&resultJSON); err != nil {
		return nil, fmt.Errorf("failed to query custom daily trend: %w", err)
	}

	var result []models.NewTrendPoint
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal custom daily trend: %w", err)
	}

	hasMore := (page+1)*limit < total
	nextPage := page + 1
	if !hasMore {
		nextPage = page
	}

	return &models.NewPaginatedTrendPoints{
		Data: result,
		Pagination: models.NewPaginationInfo{
			Total:    total,
			HasMore:  hasMore,
			NextPage: nextPage,
			Limit:    limit,
			Page:     page,
		},
	}, nil
}

func (r *reportRepo) NewGetCustomWeeklyTrend(ctx context.Context, from, to time.Time, limit, page int) (*models.NewPaginatedTrendPoints, error) {
	offset := page * limit

	// Get total count
	var total int
	countQuery := `
		SELECT COUNT(DISTINCT DATE_TRUNC('week', p.created_at AT TIME ZONE 'Asia/Kathmandu'))
		FROM payments p
		WHERE p.created_at BETWEEN $1 AND $2
	`
	if err := r.pool.QueryRow(ctx, countQuery, from, to).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count custom weekly trend: %w", err)
	}

	// Get paginated data
	query := `
		SELECT COALESCE(
			json_agg(
				json_build_object('period', period, 'revenue', revenue, 'orders', orders)
				ORDER BY week_start ASC
			), '[]'::json
		)
		FROM (
			SELECT
				TO_CHAR(DATE_TRUNC('week', p.created_at AT TIME ZONE 'Asia/Kathmandu'), 'IYYY-"W"IW') AS period,
				DATE_TRUNC('week', p.created_at AT TIME ZONE 'Asia/Kathmandu') AS week_start,
				COALESCE(SUM(p.paid_amount), 0) AS revenue,
				COUNT(p.id) AS orders
			FROM payments p
			WHERE p.created_at BETWEEN $1 AND $2
			GROUP BY DATE_TRUNC('week', p.created_at AT TIME ZONE 'Asia/Kathmandu')
			ORDER BY week_start ASC
			LIMIT $3 OFFSET $4
		) sub
	`

	var resultJSON []byte
	if err := r.pool.QueryRow(ctx, query, from, to, limit, offset).Scan(&resultJSON); err != nil {
		return nil, fmt.Errorf("failed to query custom weekly trend: %w", err)
	}

	var result []models.NewTrendPoint
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal custom weekly trend: %w", err)
	}

	hasMore := (page+1)*limit < total
	nextPage := page + 1
	if !hasMore {
		nextPage = page
	}

	return &models.NewPaginatedTrendPoints{
		Data: result,
		Pagination: models.NewPaginationInfo{
			Total:    total,
			HasMore:  hasMore,
			NextPage: nextPage,
			Limit:    limit,
			Page:     page,
		},
	}, nil
}

func (r *reportRepo) NewGetCustomMonthlyTrend(ctx context.Context, from, to time.Time, limit, page int) (*models.NewPaginatedTrendPoints, error) {
	offset := page * limit

	// Get total count
	var total int
	countQuery := `
		SELECT COUNT(DISTINCT DATE_TRUNC('month', created_at AT TIME ZONE 'Asia/Kathmandu'))
		FROM payments
		WHERE created_at BETWEEN $1 AND $2
	`
	if err := r.pool.QueryRow(ctx, countQuery, from, to).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count custom monthly trend: %w", err)
	}

	// Get paginated data
	query := `
		SELECT COALESCE(
			json_agg(
				json_build_object('period', period, 'revenue', revenue, 'orders', orders)
				ORDER BY month_start ASC
			), '[]'::json
		)
		FROM (
			SELECT
				TO_CHAR(DATE_TRUNC('month', p.created_at AT TIME ZONE 'Asia/Kathmandu'), 'YYYY-MM') AS period,
				DATE_TRUNC('month', p.created_at AT TIME ZONE 'Asia/Kathmandu') AS month_start,
				COALESCE(SUM(p.paid_amount), 0) AS revenue,
				COUNT(p.id) AS orders
			FROM payments p
			WHERE p.created_at BETWEEN $1 AND $2
			GROUP BY DATE_TRUNC('month', p.created_at AT TIME ZONE 'Asia/Kathmandu')
			ORDER BY month_start ASC
			LIMIT $3 OFFSET $4
		) sub
	`

	var resultJSON []byte
	if err := r.pool.QueryRow(ctx, query, from, to, limit, offset).Scan(&resultJSON); err != nil {
		return nil, fmt.Errorf("failed to query custom monthly trend: %w", err)
	}

	var result []models.NewTrendPoint
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal custom monthly trend: %w", err)
	}

	hasMore := (page+1)*limit < total
	nextPage := page + 1
	if !hasMore {
		nextPage = page
	}

	return &models.NewPaginatedTrendPoints{
		Data: result,
		Pagination: models.NewPaginationInfo{
			Total:    total,
			HasMore:  hasMore,
			NextPage: nextPage,
			Limit:    limit,
			Page:     page,
		},
	}, nil
}

func (r *reportRepo) NewGetCustomYearlyTrend(ctx context.Context, from, to time.Time, limit, page int) (*models.NewPaginatedTrendPoints, error) {
	offset := page * limit

	// Get total count
	var total int
	countQuery := `
		SELECT COUNT(DISTINCT DATE_TRUNC('year', created_at AT TIME ZONE 'Asia/Kathmandu'))
		FROM payments
		WHERE created_at BETWEEN $1 AND $2
	`
	if err := r.pool.QueryRow(ctx, countQuery, from, to).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count custom yearly trend: %w", err)
	}

	// Get paginated data
	query := `
		SELECT COALESCE(
			json_agg(
				json_build_object('period', period, 'revenue', revenue, 'orders', orders)
				ORDER BY year_start ASC
			), '[]'::json
		)
		FROM (
			SELECT
				TO_CHAR(DATE_TRUNC('year', p.created_at AT TIME ZONE 'Asia/Kathmandu'), 'YYYY') AS period,
				DATE_TRUNC('year', p.created_at AT TIME ZONE 'Asia/Kathmandu') AS year_start,
				COALESCE(SUM(p.paid_amount), 0) AS revenue,
				COUNT(p.id) AS orders
			FROM payments p
			WHERE p.created_at BETWEEN $1 AND $2
			GROUP BY DATE_TRUNC('year', p.created_at AT TIME ZONE 'Asia/Kathmandu')
			ORDER BY year_start ASC
			LIMIT $3 OFFSET $4
		) sub
	`

	var resultJSON []byte
	if err := r.pool.QueryRow(ctx, query, from, to, limit, offset).Scan(&resultJSON); err != nil {
		return nil, fmt.Errorf("failed to query custom yearly trend: %w", err)
	}

	var result []models.NewTrendPoint
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal custom yearly trend: %w", err)
	}

	hasMore := (page+1)*limit < total
	nextPage := page + 1
	if !hasMore {
		nextPage = page
	}

	return &models.NewPaginatedTrendPoints{
		Data: result,
		Pagination: models.NewPaginationInfo{
			Total:    total,
			HasMore:  hasMore,
			NextPage: nextPage,
			Limit:    limit,
			Page:     page,
		},
	}, nil
}

// ────────────────────────────────────────────────────────────────────────────────
// Breakdown Functions
// ────────────────────────────────────────────────────────────────────────────────

func (r *reportRepo) NewGetPaymentMethodBreakdown(ctx context.Context, from, to time.Time) ([]models.NewPaymentMethodBreakdown, error) {
	query := `
		WITH totals AS (
			SELECT SUM(paid_amount) AS total_revenue
			FROM payments
			WHERE created_at BETWEEN $1 AND $2
		)
		SELECT
			p.payment_method::text AS method,
			COALESCE(SUM(p.paid_amount), 0) AS revenue,
			COUNT(p.id) AS orders,
			CASE 
				WHEN (SELECT total_revenue FROM totals) > 0 
				THEN (SUM(p.paid_amount) * 100.0) / (SELECT total_revenue FROM totals)
				ELSE 0 
			END AS percent
		FROM payments p
		WHERE p.created_at BETWEEN $1 AND $2
		GROUP BY p.payment_method
		ORDER BY revenue DESC
	`

	rows, err := r.pool.Query(ctx, query, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to query payment methods: %w", err)
	}
	defer rows.Close()

	var result []models.NewPaymentMethodBreakdown
	for rows.Next() {
		var item models.NewPaymentMethodBreakdown
		if err := rows.Scan(&item.Method, &item.Revenue, &item.Orders, &item.Percent); err != nil {
			return nil, fmt.Errorf("failed to scan payment method: %w", err)
		}
		result = append(result, item)
	}

	return result, nil
}

func (r *reportRepo) NewGetGatewayBreakdown(ctx context.Context, from, to time.Time) ([]models.NewGatewayBreakdown, error) {
	query := `
		WITH totals AS (
			SELECT SUM(paid_amount) AS total_revenue
			FROM payments
			WHERE created_at BETWEEN $1 AND $2 AND payment_method = 'online'
		)
		SELECT
			COALESCE(p.online_gateway::text, 'unknown') AS gateway,
			COALESCE(SUM(p.paid_amount), 0) AS revenue,
			COUNT(p.id) AS orders,
			CASE 
				WHEN (SELECT total_revenue FROM totals) > 0 
				THEN (SUM(p.paid_amount) * 100.0) / (SELECT total_revenue FROM totals)
				ELSE 0 
			END AS percent
		FROM payments p
		WHERE p.created_at BETWEEN $1 AND $2 AND p.payment_method = 'online'
		GROUP BY p.online_gateway
		ORDER BY revenue DESC
	`

	rows, err := r.pool.Query(ctx, query, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to query gateways: %w", err)
	}
	defer rows.Close()

	var result []models.NewGatewayBreakdown
	for rows.Next() {
		var item models.NewGatewayBreakdown
		if err := rows.Scan(&item.Gateway, &item.Revenue, &item.Orders, &item.Percent); err != nil {
			return nil, fmt.Errorf("failed to scan gateway: %w", err)
		}
		result = append(result, item)
	}

	return result, nil
}

func (r *reportRepo) NewGetDiscountAnalysis(ctx context.Context, from, to time.Time) (models.NewDiscountAnalysis, error) {
	query := `
		SELECT
			COALESCE(SUM(p.discount), 0) AS total_discounts_given,
			COALESCE(SUM(p.paid_amount + p.discount), 0) AS gross_revenue,
			COALESCE(SUM(p.paid_amount), 0) AS net_revenue,
			CASE 
				WHEN SUM(p.paid_amount + p.discount) > 0 
				THEN (SUM(p.discount) * 100.0) / SUM(p.paid_amount + p.discount)
				ELSE 0 
			END AS discount_rate_percent,
			COUNT(CASE WHEN p.discount > 0 THEN 1 END) AS orders_with_discount,
			COUNT(p.id) AS total_orders
		FROM payments p
		WHERE p.created_at BETWEEN $1 AND $2
	`

	var result models.NewDiscountAnalysis
	if err := r.pool.QueryRow(ctx, query, from, to).Scan(
		&result.TotalDiscountsGiven,
		&result.GrossRevenue,
		&result.NetRevenue,
		&result.DiscountRatePercent,
		&result.OrdersWithDiscount,
		&result.TotalOrders,
	); err != nil {
		return models.NewDiscountAnalysis{}, fmt.Errorf("failed to scan discount analysis: %w", err)
	}

	return result, nil
}

func (r *reportRepo) NewGetPeakHours(ctx context.Context, from, to time.Time) ([]models.NewPeakHourPoint, error) {
	query := `
		SELECT
			EXTRACT(HOUR FROM p.created_at AT TIME ZONE 'Asia/Kathmandu')::int AS hour,
			COALESCE(SUM(p.paid_amount), 0) AS revenue,
			COUNT(p.id) AS orders
		FROM payments p
		WHERE p.created_at BETWEEN $1 AND $2
		GROUP BY EXTRACT(HOUR FROM p.created_at AT TIME ZONE 'Asia/Kathmandu')
		ORDER BY hour ASC
	`

	rows, err := r.pool.Query(ctx, query, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to query peak hours: %w", err)
	}
	defer rows.Close()

	var result []models.NewPeakHourPoint
	for rows.Next() {
		var item models.NewPeakHourPoint
		if err := rows.Scan(&item.Hour, &item.Revenue, &item.Orders); err != nil {
			return nil, fmt.Errorf("failed to scan peak hour: %w", err)
		}
		result = append(result, item)
	}

	return result, nil
}

func (r *reportRepo) NewGetPeakDays(ctx context.Context, from, to time.Time) ([]models.NewPeakDayPoint, error) {
	query := `
		SELECT
			TRIM(TO_CHAR(p.created_at AT TIME ZONE 'Asia/Kathmandu', 'Day')) AS day_of_week,
			COALESCE(SUM(p.paid_amount), 0) AS revenue,
			COUNT(p.id) AS orders
		FROM payments p
		WHERE p.created_at BETWEEN $1 AND $2
		GROUP BY TO_CHAR(p.created_at AT TIME ZONE 'Asia/Kathmandu', 'Day'), 
		         EXTRACT(DOW FROM p.created_at AT TIME ZONE 'Asia/Kathmandu')
		ORDER BY EXTRACT(DOW FROM p.created_at AT TIME ZONE 'Asia/Kathmandu') ASC
	`

	rows, err := r.pool.Query(ctx, query, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to query peak days: %w", err)
	}
	defer rows.Close()

	var result []models.NewPeakDayPoint
	for rows.Next() {
		var item models.NewPeakDayPoint
		if err := rows.Scan(&item.DayOfWeek, &item.Revenue, &item.Orders); err != nil {
			return nil, fmt.Errorf("failed to scan peak day: %w", err)
		}
		result = append(result, item)
	}

	return result, nil
}
