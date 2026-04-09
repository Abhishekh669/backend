package models

import "time"

type SalaryByRole struct {
	Role          string  `json:"role"`
	EmployeeCount int     `json:"employee_count"`
	TotalSalary   float64 `json:"total_salary"`
	AverageSalary float64 `json:"average_salary"`
	Percent       float64 `json:"percent"`
}

// ─── NEW: Most Present Employees (Full User Details) ─────────────────────────

type MostPresentEmployee struct {
	EmployeeID     string  `json:"employee_id"`
	EmployeeName   string  `json:"employee_name"`
	Email          string  `json:"email"`
	Phone          string  `json:"phone"`
	Image          *string `json:"image"` // Can be NULL
	Role           string  `json:"role"`
	Gender         string  `json:"gender"`
	PresentDays    int     `json:"present_days"`
	AttendanceRate float64 `json:"attendance_rate"`
}

// ─── NEW: Most Absent Employees (Full User Details) ─────────────────────────

type MostAbsentEmployee struct {
	EmployeeID     string  `json:"employee_id"`
	EmployeeName   string  `json:"employee_name"`
	Email          string  `json:"email"`
	Phone          string  `json:"phone"`
	Image          *string `json:"image"` // Can be NULL
	Role           string  `json:"role"`
	Gender         string  `json:"gender"`
	AbsentDays     int     `json:"absent_days"`
	AttendanceRate float64 `json:"attendance_rate"`
}

// ─── NEW: Extended Staff Report Response with Most Present/Absent ────────────

// ─── Gateway Breakdown ───────────────────────────────────────────────────────────

type GatewayBreakdown struct {
	Gateway string  `json:"gateway"`
	Revenue float64 `json:"revenue"`
	Orders  int     `json:"orders"`
	Percent float64 `json:"percent"`
}

// ─── Shared ───────────────────────────────────────────────────────────────────

type TrendPoint struct {
	Period  string  `json:"period"`
	Revenue float64 `json:"revenue"`
	Orders  int     `json:"orders"`
}

// ─── Revenue ──────────────────────────────────────────────────────────────────

type RevenueOverviewCard struct {
	GrossRevenue      float64 `json:"gross_revenue"`
	NetRevenue        float64 `json:"net_revenue"`
	TotalDiscounts    float64 `json:"total_discounts"`
	TotalOrders       int     `json:"total_orders"`
	AverageOrderValue float64 `json:"average_order_value"`
	GrowthPercent     float64 `json:"growth_percent"`
}

type PaymentMethodBreakdown struct {
	Method  string  `json:"method"`
	Revenue float64 `json:"revenue"`
	Orders  int     `json:"orders"`
	Percent float64 `json:"percent"`
}

type DiscountAnalysis struct {
	TotalDiscountsGiven float64 `json:"total_discounts_given"`
	GrossRevenue        float64 `json:"gross_revenue"`
	NetRevenue          float64 `json:"net_revenue"`
	DiscountRatePercent float64 `json:"discount_rate_percent"`
	OrdersWithDiscount  int     `json:"orders_with_discount"`
	TotalOrders         int     `json:"total_orders"`
}

type PeakHourPoint struct {
	Hour    int     `json:"hour"`
	Revenue float64 `json:"revenue"`
	Orders  int     `json:"orders"`
}

type PeakDayPoint struct {
	DayOfWeek string  `json:"day_of_week"`
	Revenue   float64 `json:"revenue"`
	Orders    int     `json:"orders"`
}

type RevenueReportResponse struct {
	Overview       RevenueOverviewCard      `json:"overview"`
	DailyTrend     []TrendPoint             `json:"daily_trend"`
	WeeklyTrend    []TrendPoint             `json:"weekly_trend"`
	MonthlyTrend   []TrendPoint             `json:"monthly_trend"`
	YearlyTrend    []TrendPoint             `json:"yearly_trend"`
	PaymentMethods []PaymentMethodBreakdown `json:"payment_methods"`
	Gateways       []GatewayBreakdown       `json:"gateways"`
	Discounts      DiscountAnalysis         `json:"discounts"`
	PeakHours      []PeakHourPoint          `json:"peak_hours"`
	PeakDays       []PeakDayPoint           `json:"peak_days"`
}

// ─── Sales ────────────────────────────────────────────────────────────────────

type SalesOverviewCard struct {
	TotalItemsSold    float64 `json:"total_items_sold"`
	UniqueMenuItems   int     `json:"unique_menu_items"`
	TopSellingItem    string  `json:"top_selling_item"`
	TopCategory       string  `json:"top_category"`
	TotalOrdersPlaced int     `json:"total_orders_placed"`
}

type BestSellingItem struct {
	MenuItemID   string  `json:"menu_item_id"`
	MenuName     string  `json:"menu_name"`
	CategoryName string  `json:"category_name"`
	TotalQty     float64 `json:"total_qty"`
	TotalRevenue float64 `json:"total_revenue"`
	OrderCount   int     `json:"order_count"`
}

type BestSellingCategory struct {
	CategoryID   string  `json:"category_id"`
	CategoryName string  `json:"category_name"`
	TotalQty     float64 `json:"total_qty"`
	TotalRevenue float64 `json:"total_revenue"`
	ItemCount    int     `json:"item_count"`
}

type SlowestMovingItem struct {
	MenuItemID   string  `json:"menu_item_id"`
	MenuName     string  `json:"menu_name"`
	CategoryName string  `json:"category_name"`
	TotalQty     float64 `json:"total_qty"`
	TotalRevenue float64 `json:"total_revenue"`
	LastOrdered  *string `json:"last_ordered"`
}

type FrequentPairItem struct {
	ItemA     string `json:"item_a"`
	ItemB     string `json:"item_b"`
	PairCount int    `json:"pair_count"`
}

// ─── Table & Session ──────────────────────────────────────────────────────────

type TableOverviewCard struct {
	TotalTables        int     `json:"total_tables"`
	CurrentlyOccupied  int     `json:"currently_occupied"`
	UtilizationPercent float64 `json:"utilization_percent"`
	AvgSessionMinutes  float64 `json:"avg_session_minutes"`
	TotalSessionsToday int     `json:"total_sessions_today"`
}

type TableUtilizationRow struct {
	TableNumber        int     `json:"table_number"`
	TotalSessions      int     `json:"total_sessions"`
	AvgSessionMinutes  float64 `json:"avg_session_minutes"`
	UtilizationPercent float64 `json:"utilization_percent"`
	TotalRevenue       float64 `json:"total_revenue"`
}

type PeakOccupancyHour struct {
	Hour          int     `json:"hour"`
	AvgOccupied   float64 `json:"avg_occupied_tables"`
	TotalSessions int     `json:"total_sessions"`
}

type TableTurnoverRow struct {
	TableNumber    int     `json:"table_number"`
	SessionsPerDay float64 `json:"sessions_per_day"`
	TotalSessions  int     `json:"total_sessions"`
}

type LongestIdleTable struct {
	TableNumber   int     `json:"table_number"`
	LastClosedAt  *string `json:"last_closed_at"`
	IdleHours     float64 `json:"idle_hours"`
	CurrentStatus string  `json:"current_status"`
}

// ─── Customer & Loyalty ───────────────────────────────────────────────────────

type CustomerOverviewCard struct {
	TotalUniqueCustomers     int     `json:"total_unique_customers"`
	NewCustomers             int     `json:"new_customers"`
	ReturningCustomers       int     `json:"returning_customers"`
	RetentionPercent         float64 `json:"retention_percent"`
	TotalTokensInCirculation float64 `json:"total_tokens_in_circulation"`
	TotalTokensSpent         float64 `json:"total_tokens_spent"`
	TokenRedemptionRate      float64 `json:"token_redemption_rate"`
}

type TopCustomerRow struct {
	Phone         string  `json:"phone"`
	CustomerName  string  `json:"customer_name"`
	TotalSpend    float64 `json:"total_spend"`
	VisitCount    int     `json:"visit_count"`
	TotalTokens   float64 `json:"total_tokens"`
	CurrentStreak int     `json:"current_streak"`
}

type CustomerVisitFrequency struct {
	VisitBucket   string  `json:"visit_bucket"` // "1 visit", "2-5 visits", etc.
	CustomerCount int     `json:"customer_count"`
	Percent       float64 `json:"percent"`
}

type TokenEconomyReport struct {
	TotalEarned        float64           `json:"total_earned"`
	TotalSpent         float64           `json:"total_spent"`
	TotalStreakBonuses float64           `json:"total_streak_bonuses"`
	NetCirculation     float64           `json:"net_circulation"`
	RedemptionRate     float64           `json:"redemption_rate_percent"`
	EarnSpendTrend     []TokenTrendPoint `json:"earn_spend_trend"`
}

type TokenTrendPoint struct {
	Period string  `json:"period"`
	Earned float64 `json:"earned"`
	Spent  float64 `json:"spent"`
	Streak float64 `json:"streak_bonus"`
}

type StreakLeaderboardRow struct {
	Phone         string `json:"phone"`
	CurrentStreak int    `json:"current_streak"`
	MonthlyDays   int    `json:"monthly_days"`
	LastVisit     string `json:"last_visit"`
}

// ─── Staff & HR ───────────────────────────────────────────────────────────────

type AttendanceOverviewCard struct {
	TotalEmployees  int     `json:"total_employees"`
	PresentToday    int     `json:"present_today"`
	AbsentToday     int     `json:"absent_today"`
	LateToday       int     `json:"late_today"`
	OnLeaveToday    int     `json:"on_leave_today"`
	NeedReviewCount int     `json:"need_review_count"`
	AttendanceRate  float64 `json:"attendance_rate_percent"`
}

type DailyAttendanceSummary struct {
	WorkDate string `json:"work_date"`
	Present  int    `json:"present"`
	Absent   int    `json:"absent"`
	Late     int    `json:"late"`
	HalfDay  int    `json:"half_day"`
	OnLeave  int    `json:"on_leave"`
}

type EmployeeAttendanceRow struct {
	EmployeeID     string  `json:"employee_id"`
	EmployeeName   string  `json:"employee_name"`
	Role           string  `json:"role"`
	PresentDays    int     `json:"present_days"`
	AbsentDays     int     `json:"absent_days"`
	LateDays       int     `json:"late_days"`
	AttendanceRate float64 `json:"attendance_rate_percent"`
	AvgWorkHours   float64 `json:"avg_work_hours"`
	NeedReview     bool    `json:"need_review"`
}

type LeaveManagementReport struct {
	PendingCount      int                `json:"pending_count"`
	ApprovedCount     int                `json:"approved_count"`
	RejectedCount     int                `json:"rejected_count"`
	ApprovalRate      float64            `json:"approval_rate_percent"`
	TopLeaveEmployees []TopLeaveEmployee `json:"top_leave_employees"`
}

type TopLeaveEmployee struct {
	EmployeeID   string `json:"employee_id"`
	EmployeeName string `json:"employee_name"`
	LeaveCount   int    `json:"leave_count"`
}

type PayrollReport struct {
	TotalMonthlySalary float64           `json:"total_monthly_salary"`
	SalaryByRole       []SalaryByRoleRow `json:"salary_by_role"`
	LaborCostPercent   float64           `json:"labor_cost_percent"`
}

type SalaryByRoleRow struct {
	Role          string  `json:"role"`
	EmployeeCount int     `json:"employee_count"`
	TotalSalary   float64 `json:"total_salary"`
	AverageSalary float64 `json:"average_salary"`
	Percent       float64 `json:"percent"`
}

type StaffReportResponse struct {
	AttendanceOverview AttendanceOverviewCard   `json:"attendance_overview"`
	DailySummary       []DailyAttendanceSummary `json:"daily_summary"`
	EmployeeAttendance []EmployeeAttendanceRow  `json:"employee_attendance"`
	LeaveManagement    LeaveManagementReport    `json:"leave_management"`
	Payroll            PayrollReport            `json:"payroll"`
}

// ─── Raw Materials ────────────────────────────────────────────────────────────

type RawMaterialOverviewCard struct {
	TotalMaterials      int     `json:"total_materials"`
	TotalInventoryValue float64 `json:"total_inventory_value"`
	LowStockCount       int     `json:"low_stock_count"`
	TotalInvested       float64 `json:"total_invested"`
}

type RawMaterialRow struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Quantity     float64   `json:"quantity"`
	Unit         string    `json:"unit"`
	Price        float64   `json:"price"`
	TotalValue   float64   `json:"total_value"`
	ValuePercent float64   `json:"value_percent"`
	IsLowStock   bool      `json:"is_low_stock"`
	LastUpdated  time.Time `json:"last_updated"`
}

type RawMaterialReportResponse struct {
	Overview  RawMaterialOverviewCard `json:"overview"`
	Materials []RawMaterialRow        `json:"materials"`
}

// from here
type SalesTrendPoint struct {
	Period         string  `json:"period"`
	TotalItemsSold float64 `json:"total_items_sold"`
	TotalOrders    int     `json:"total_orders"`
	TotalRevenue   float64 `json:"total_revenue"`
}

type SalesReportResponse struct {
	Overview       SalesOverviewCard     `json:"overview"`
	BestByQty      []BestSellingItem     `json:"best_by_qty"`
	BestByRevenue  []BestSellingItem     `json:"best_by_revenue"`
	BestCategories []BestSellingCategory `json:"best_categories"`
	SlowestMoving  []SlowestMovingItem   `json:"slowest_moving"`
	FrequentPairs  []FrequentPairItem    `json:"frequent_pairs"`
	// ─── NEW trend fields ───
	DailyTrend   []SalesTrendPoint `json:"daily_trend"`
	WeeklyTrend  []SalesTrendPoint `json:"weekly_trend"`
	MonthlyTrend []SalesTrendPoint `json:"monthly_trend"`
	YearlyTrend  []SalesTrendPoint `json:"yearly_trend"`
}

// ─── Customer ─────────────────────────────────────────────────────────────────

// Add this new type
type CustomerTrendPoint struct {
	Period             string  `json:"period"`
	NewCustomers       int     `json:"new_customers"`
	ReturningCustomers int     `json:"returning_customers"`
	TotalOrders        int     `json:"total_orders"`
	TotalRevenue       float64 `json:"total_revenue"`
}

type CustomerReportResponse struct {
	Overview          CustomerOverviewCard     `json:"overview"`
	TopCustomers      []TopCustomerRow         `json:"top_customers"`
	VisitFrequency    []CustomerVisitFrequency `json:"visit_frequency"`
	TokenEconomy      TokenEconomyReport       `json:"token_economy"`
	StreakLeaderboard []StreakLeaderboardRow   `json:"streak_leaderboard"`
	// ─── NEW trend fields ───
	DailyTrend   []CustomerTrendPoint `json:"daily_trend"`
	WeeklyTrend  []CustomerTrendPoint `json:"weekly_trend"`
	MonthlyTrend []CustomerTrendPoint `json:"monthly_trend"`
	YearlyTrend  []CustomerTrendPoint `json:"yearly_trend"`
}

// ─── Table ────────────────────────────────────────────────────────────────────

// Add this new type
type TableTrendPoint struct {
	Period        string  `json:"period"`
	TotalSessions int     `json:"total_sessions"`
	TotalRevenue  float64 `json:"total_revenue"`
	AvgSessionMin float64 `json:"avg_session_minutes"`
}

type TableReportResponse struct {
	Overview    TableOverviewCard     `json:"overview"`
	Utilization []TableUtilizationRow `json:"utilization"`
	PeakHours   []PeakOccupancyHour   `json:"peak_hours"`
	Turnover    []TableTurnoverRow    `json:"turnover"`
	LongestIdle []LongestIdleTable    `json:"longest_idle"`
	// ─── NEW trend fields ───
	DailyTrend   []TableTrendPoint `json:"daily_trend"`
	WeeklyTrend  []TableTrendPoint `json:"weekly_trend"`
	MonthlyTrend []TableTrendPoint `json:"monthly_trend"`
	YearlyTrend  []TableTrendPoint `json:"yearly_trend"`
}

// ─── Staff ────────────────────────────────────────────────────────────────────

// Add this new type
type StaffTrendPoint struct {
	Period         string  `json:"period"`
	Present        int     `json:"present"`
	Absent         int     `json:"absent"`
	Late           int     `json:"late"`
	OnLeave        int     `json:"on_leave"`
	AttendanceRate float64 `json:"attendance_rate"`
}

type ExtendedStaffReportResponse struct {
	AttendanceOverview   AttendanceOverviewCard   `json:"attendance_overview"`
	DailySummary         []DailyAttendanceSummary `json:"daily_summary"`
	EmployeeAttendance   []EmployeeAttendanceRow  `json:"employee_attendance"`
	LeaveManagement      LeaveManagementReport    `json:"leave_management"`
	Payroll              PayrollReport            `json:"payroll"`
	MostPresentEmployees []MostPresentEmployee    `json:"most_present_employees"`
	MostAbsentEmployees  []MostAbsentEmployee     `json:"most_absent_employees"`
	// ─── NEW trend fields ───
	WeeklyTrend  []StaffTrendPoint `json:"weekly_trend"`
	MonthlyTrend []StaffTrendPoint `json:"monthly_trend"`
	YearlyTrend  []StaffTrendPoint `json:"yearly_trend"`
}

// ─── Financial ────────────────────────────────────────────────────────────────

// Add this new type
type FinancialTrendPoint struct {
	Period       string  `json:"period"`
	Revenue      float64 `json:"revenue"`
	MaterialCost float64 `json:"material_cost"` // raw material purchases that period
	GrossProfit  float64 `json:"gross_profit"`
}

type FinancialSummaryResponse struct {
	TotalInvested float64 `json:"total_invested"`
	TotalEarned   float64 `json:"total_earned"`
	GrossProfit   float64 `json:"gross_profit"`
	ProfitPercent float64 `json:"profit_percent"`
	// ─── NEW trend fields ───
	MonthlyTrend []FinancialTrendPoint `json:"monthly_trend"`
	YearlyTrend  []FinancialTrendPoint `json:"yearly_trend"`
}
