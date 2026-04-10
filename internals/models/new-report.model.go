package models

import "time"

// ─── Raw Material Overview Card ───────────────────────────────────────────────
// NOTE: TotalRevenue, GrossMargin, MarginPercent have been removed.
// The old query computed them via a CROSS JOIN that inflated values by N materials.

type NewRawMaterialOverviewCard struct {
	TotalMaterialUsed        float64 `json:"total_material_used"`
	TotalInvestment          float64 `json:"total_investment"`
	TotalOrders              int     `json:"total_orders"`
	HighestCostMaterialValue float64 `json:"highest_cost_material_value"`
	HighestCostMaterialName  string  `json:"highest_cost_material_name"`
	MostUsedMaterialQuantity float64 `json:"most_used_material_quantity"`
	MostUsedMaterialName     string  `json:"most_used_material_name"`
}

// ─── Raw Material Stats Card (All Time) ──────────────────────────────────────

type NewRawMaterialStatsCard struct {
	TotalMaterials            int     `json:"total_materials"`
	TotalCurrentStock         float64 `json:"total_current_stock"`
	TotalInventoryValue       float64 `json:"total_inventory_value"`
	TotalMaterialUsedAllTime  float64 `json:"total_material_used_all_time"`
	TotalInvestmentAllTime    float64 `json:"total_investment_all_time"`
	MaxUsedQuantity           float64 `json:"max_used_quantity"`
	MostUsedMaterialName      string  `json:"most_used_material_name"`
	MostUsedMaterialQuantity  float64 `json:"most_used_material_quantity"`
	MostExpensiveUnitCost     float64 `json:"most_expensive_unit_cost"`
	MostExpensiveMaterialName string  `json:"most_expensive_material_name"`
	AvgMaterialValue          float64 `json:"avg_material_value"`
}

// ─── Trend Point ─────────────────────────────────────────────────────────────

type NewRawMaterialTrendPoint struct {
	Period       string  `json:"period"`
	MaterialUsed float64 `json:"material_used"`
	TotalCost    float64 `json:"total_cost"`
	OrdersCount  int     `json:"orders_count"`
}

// ─── Paginated Trend ─────────────────────────────────────────────────────────

type NewRawMaterialPaginatedTrendPoints struct {
	Data       []NewRawMaterialTrendPoint `json:"data"`
	Pagination NewPaginationInfo          `json:"pagination"`
}

// ─── Top Used Raw Material ───────────────────────────────────────────────────

type NewTopUsedRawMaterial struct {
	MaterialID        string  `json:"material_id"`
	MaterialName      string  `json:"material_name"`
	Unit              string  `json:"unit"`
	UnitCost          float64 `json:"unit_cost"`
	TotalQuantityUsed float64 `json:"total_quantity_used"`
	TotalCost         float64 `json:"total_cost"`
	AffectedOrders    int     `json:"affected_orders"`
}

// ─── Usage Breakdown ─────────────────────────────────────────────────────────

type NewRawMaterialUsageBreakdown struct {
	MaterialID   string  `json:"material_id"`
	MaterialName string  `json:"material_name"`
	Unit         string  `json:"unit"`
	UnitCost     float64 `json:"unit_cost"`
	CurrentStock float64 `json:"current_stock"`
	PeriodUsage  float64 `json:"period_usage"`
	PeriodCost   float64 `json:"period_cost"`
	UsagePercent float64 `json:"usage_percent"`
	OrdersCount  int     `json:"orders_count"`
}

// ─── Peak Hour ───────────────────────────────────────────────────────────────

type NewRawMaterialPeakHour struct {
	Hour              int     `json:"hour"`
	TotalMaterialUsed float64 `json:"total_material_used"`
	TotalCost         float64 `json:"total_cost"`
	OrdersCount       int     `json:"orders_count"`
	UniqueItemsUsed   int     `json:"unique_items_used"`
}

// ─── Daily Usage Summary ─────────────────────────────────────────────────────

type NewDailyRawMaterialUsage struct {
	UsageDate           string  `json:"usage_date"`
	TotalMaterialUsed   float64 `json:"total_material_used"`
	TotalCost           float64 `json:"total_cost"`
	OrdersCount         int     `json:"orders_count"`
	UniqueMaterialsUsed int     `json:"unique_materials_used"`
}

// ─── Default Response ────────────────────────────────────────────────────────

type NewDefaultRawMaterialResponse struct {
	Overview               NewRawMaterialOverviewCard     `json:"overview"`
	StatsCard              NewRawMaterialStatsCard        `json:"stats_card"`
	DailyTrend             []NewRawMaterialTrendPoint     `json:"daily_trend"`
	WeeklyTrend            []NewRawMaterialTrendPoint     `json:"weekly_trend"`
	MonthlyTrend           []NewRawMaterialTrendPoint     `json:"monthly_trend"`
	YearlyTrend            []NewRawMaterialTrendPoint     `json:"yearly_trend"`
	TopUsedMaterials       []NewTopUsedRawMaterial        `json:"top_used_materials"`
	MaterialUsageBreakdown []NewRawMaterialUsageBreakdown `json:"material_usage_breakdown"`
	PeakUsageHours         []NewRawMaterialPeakHour       `json:"peak_usage_hours"`
	DailyUsageSummary      []NewDailyRawMaterialUsage     `json:"daily_usage_summary"`
}

// ─── Custom Range Response ───────────────────────────────────────────────────

type NewCustomRangeRawMaterialResponse struct {
	Overview               NewRawMaterialOverviewCard          `json:"overview"`
	StatsCard              NewRawMaterialStatsCard             `json:"stats_card"`
	DailyTrend             *NewRawMaterialPaginatedTrendPoints `json:"daily_trend"`
	WeeklyTrend            *NewRawMaterialPaginatedTrendPoints `json:"weekly_trend"`
	MonthlyTrend           *NewRawMaterialPaginatedTrendPoints `json:"monthly_trend"`
	YearlyTrend            *NewRawMaterialPaginatedTrendPoints `json:"yearly_trend"`
	TopUsedMaterials       []NewTopUsedRawMaterial             `json:"top_used_materials"`
	MaterialUsageBreakdown []NewRawMaterialUsageBreakdown      `json:"material_usage_breakdown"`
	PeakUsageHours         []NewRawMaterialPeakHour            `json:"peak_usage_hours"`
	DailyUsageSummary      []NewDailyRawMaterialUsage          `json:"daily_usage_summary"`
}

// ─── Request ─────────────────────────────────────────────────────────────────

type NewRawMaterialCustomRangeReportRequest struct {
	From  time.Time `json:"from"`
	To    time.Time `json:"to"`
	Limit int       `json:"limit"`
	Page  int       `json:"page"`
}

// ────────────────────────────────────────────────────────────────────────────
// NEW STAFF REPORT MODELS
// Mirrors the NewTable* pattern. Every type is prefixed with "New".
// ────────────────────────────────────────────────────────────────────────────

// ─── Pagination (reused from table report pattern) ───────────────────────────

// NewStaffPaginatedTrendPoints wraps a page of trend data.
type NewStaffPaginatedTrendPoints struct {
	Data       []NewStaffTrendPoint `json:"data"`
	Pagination NewPaginationInfo    `json:"pagination"`
}

// ─── Overview Card ────────────────────────────────────────────────────────────

// NewStaffOverviewCard provides the at-a-glance numbers for the chosen period.
type NewStaffOverviewCard struct {
	// Headcount
	TotalEmployees  int `json:"total_employees"`
	ActiveEmployees int `json:"active_employees"` // is_active = true

	// Attendance counters for the period
	TotalPresentDays int `json:"total_present_days"`
	TotalAbsentDays  int `json:"total_absent_days"`
	TotalLateDays    int `json:"total_late_days"`
	TotalHalfDays    int `json:"total_half_days"`
	TotalLeaveDays   int `json:"total_leave_days"`

	// Derived rates
	OverallAttendanceRate float64 `json:"overall_attendance_rate"` // present / (present+absent) * 100
	LateRate              float64 `json:"late_rate"`               // late / total_records * 100
	LeaveApprovalRate     float64 `json:"leave_approval_rate"`     // approved / total_leaves * 100

	// Work-hours
	TotalWorkHours float64 `json:"total_work_hours"`
	AvgWorkHours   float64 `json:"avg_work_hours_per_employee"` // per employee per present day

	// Peak metrics
	BusiestDay     string `json:"busiest_day"`      // day-of-week with most check-ins
	PeakAttendHour int    `json:"peak_attend_hour"` // most common check-in hour (0-23)
}

// ─── Stats Card (all-time / snapshot) ────────────────────────────────────────

// NewStaffStatsCard holds all-time aggregate figures (no date filter).
type NewStaffStatsCard struct {
	TotalEmployees             int     `json:"total_employees"`
	ActiveEmployees            int     `json:"active_employees"`
	TotalAttendanceRecords     int     `json:"total_attendance_records"`
	AllTimeWorkHours           float64 `json:"all_time_work_hours"`
	AvgSessionHours            float64 `json:"avg_session_hours"` // avg hours per present record
	MostPresentEmployeeID      string  `json:"most_present_employee_id"`
	MostPresentEmployeeName    string  `json:"most_present_employee_name"`
	MostPresentDays            int     `json:"most_present_days"`
	MostAbsentEmployeeID       string  `json:"most_absent_employee_id"`
	MostAbsentEmployeeName     string  `json:"most_absent_employee_name"`
	MostAbsentDays             int     `json:"most_absent_days"`
	LongestAvgServiceHours     float64 `json:"longest_avg_service_hours"` // employee with highest avg shift
	LongestServiceEmployeeID   string  `json:"longest_service_employee_id"`
	LongestServiceEmployeeName string  `json:"longest_service_employee_name"`
	TotalPendingLeaves         int     `json:"total_pending_leaves"`
	TotalApprovedLeaves        int     `json:"total_approved_leaves"`
}

// ─── Trend Point ─────────────────────────────────────────────────────────────

// NewStaffTrendPoint is one point on any trend chart (daily/weekly/monthly/yearly).
type NewStaffTrendPoint struct {
	Period         string  `json:"period"` // "2024-01-15" / "2024-W03" / "2024-01" / "2024"
	Present        int     `json:"present"`
	Absent         int     `json:"absent"`
	Late           int     `json:"late"`
	HalfDay        int     `json:"half_day"`
	OnLeave        int     `json:"on_leave"`
	TotalWorkHours float64 `json:"total_work_hours"`
	AttendanceRate float64 `json:"attendance_rate"` // present/(present+absent)*100
}

// ─── Individual Attendance Record ────────────────────────────────────────────

// NewEmployeeAttendanceRecord is one raw row from the attendance table,
// enriched with the employee's user details.
type NewEmployeeAttendanceRecord struct {
	// attendance table fields
	AttendanceID string  `json:"attendance_id"`
	WorkDate     string  `json:"work_date"`      // "YYYY-MM-DD"
	Status       string  `json:"status"`         // present | absent | late | half_day | leave
	CheckInTime  *string `json:"check_in_time"`  // nullable TIMESTAMPTZ → ISO string
	CheckOutTime *string `json:"check_out_time"` // nullable
	WorkHours    float64 `json:"work_hours"`     // derived: (check_out - check_in) / 3600, floored at 0
	NeedReview   bool    `json:"need_review"`

	// denormalised user fields (so the frontend never needs a second call)
	EmployeeID   string  `json:"employee_id"`
	EmployeeName string  `json:"employee_name"`
	Email        string  `json:"email"`
	Phone        string  `json:"phone"`
	Image        *string `json:"image"`
	Role         string  `json:"role"`
	Gender       string  `json:"gender"`
}

// ─── Per-Employee Attendance Row ──────────────────────────────────────────────

// NewEmployeeAttendanceSummary holds per-employee aggregate stats for the period
// plus every individual attendance record so the frontend can render both the
// summary card and the detail table from a single response.
type NewEmployeeAttendanceSummary struct {
	// user identity
	EmployeeID   string  `json:"employee_id"`
	EmployeeName string  `json:"employee_name"`
	Email        string  `json:"email"`
	Phone        string  `json:"phone"`
	Image        *string `json:"image"`
	Role         string  `json:"role"`
	Gender       string  `json:"gender"`

	// aggregated counters
	PresentDays     int     `json:"present_days"`
	AbsentDays      int     `json:"absent_days"`
	LateDays        int     `json:"late_days"`
	HalfDays        int     `json:"half_days"`
	LeaveDays       int     `json:"leave_days"`
	TotalWorkHours  float64 `json:"total_work_hours"`
	AvgWorkHours    float64 `json:"avg_work_hours"`  // per attended shift
	AttendanceRate  float64 `json:"attendance_rate"` // present/(present+absent)*100
	NeedReviewCount int     `json:"need_review_count"`

	// every individual record in the period — ordered by work_date ASC
	Records []NewEmployeeAttendanceRecord `json:"records"`
}

// ─── Most Present / Absent ────────────────────────────────────────────────────

// NewMostPresentEmployee is the top-N most present employees in the period.
type NewMostPresentEmployee struct {
	EmployeeID     string  `json:"employee_id"`
	EmployeeName   string  `json:"employee_name"`
	Email          string  `json:"email"`
	Phone          string  `json:"phone"`
	Image          *string `json:"image"`
	Role           string  `json:"role"`
	Gender         string  `json:"gender"`
	PresentDays    int     `json:"present_days"`
	TotalWorkHours float64 `json:"total_work_hours"`
	AttendanceRate float64 `json:"attendance_rate"`
}

// NewMostAbsentEmployee is the top-N most absent employees in the period.
type NewMostAbsentEmployee struct {
	EmployeeID     string  `json:"employee_id"`
	EmployeeName   string  `json:"employee_name"`
	Email          string  `json:"email"`
	Phone          string  `json:"phone"`
	Image          *string `json:"image"`
	Role           string  `json:"role"`
	Gender         string  `json:"gender"`
	AbsentDays     int     `json:"absent_days"`
	TotalWorkHours float64 `json:"total_work_hours"`
	AttendanceRate float64 `json:"attendance_rate"`
}

// ─── Longest Service (highest avg shift hours) ───────────────────────────────

// NewLongestServiceEmployee is the top-N employees by average hours per shift.
type NewLongestServiceEmployee struct {
	EmployeeID     string  `json:"employee_id"`
	EmployeeName   string  `json:"employee_name"`
	Email          string  `json:"email"`
	Phone          string  `json:"phone"`
	Image          *string `json:"image"`
	Role           string  `json:"role"`
	Gender         string  `json:"gender"`
	TotalWorkHours float64 `json:"total_work_hours"`
	AvgShiftHours  float64 `json:"avg_shift_hours"`
	PresentDays    int     `json:"present_days"`
	AttendanceRate float64 `json:"attendance_rate"`
}

// ─── Role Breakdown ───────────────────────────────────────────────────────────

// NewStaffRoleBreakdown shows headcount, hours and attendance per role.
type NewStaffRoleBreakdown struct {
	Role           string  `json:"role"`
	EmployeeCount  int     `json:"employee_count"`
	TotalWorkHours float64 `json:"total_work_hours"`
	AvgWorkHours   float64 `json:"avg_work_hours"`
	AttendanceRate float64 `json:"attendance_rate"`
	Salary         float64 `json:"total_salary"`
	AvgSalary      float64 `json:"avg_salary"`
	Percent        float64 `json:"percent"` // % of total headcount
}

// ─── Leave Analysis ───────────────────────────────────────────────────────────

// NewLeaveAnalysis aggregates leave request data for the period.
type NewLeaveAnalysis struct {
	TotalRequests     int                   `json:"total_requests"`
	PendingCount      int                   `json:"pending_count"`
	ApprovedCount     int                   `json:"approved_count"`
	RejectedCount     int                   `json:"rejected_count"`
	ApprovalRate      float64               `json:"approval_rate"`  // approved/total*100
	AvgLeaveDays      float64               `json:"avg_leave_days"` // avg duration of approved leaves
	TopLeaveEmployees []NewTopLeaveEmployee `json:"top_leave_employees"`
}

// NewTopLeaveEmployee is one row in the most-leaves list.
type NewTopLeaveEmployee struct {
	EmployeeID   string  `json:"employee_id"`
	EmployeeName string  `json:"employee_name"`
	Role         string  `json:"role"`
	LeaveCount   int     `json:"leave_count"`
	TotalDays    float64 `json:"total_days"`
}

// ─── Daily Summary ────────────────────────────────────────────────────────────

// NewDailyAttendanceSummary is one row in the daily breakdown table.
type NewDailyAttendanceSummary struct {
	WorkDate       string  `json:"work_date"`
	Present        int     `json:"present"`
	Absent         int     `json:"absent"`
	Late           int     `json:"late"`
	HalfDay        int     `json:"half_day"`
	OnLeave        int     `json:"on_leave"`
	TotalWorkHours float64 `json:"total_work_hours"`
	AttendanceRate float64 `json:"attendance_rate"`
}

// ─── Peak Hours ───────────────────────────────────────────────────────────────

// NewStaffPeakHour shows check-in traffic per hour of the day.
type NewStaffPeakHour struct {
	Hour         int     `json:"hour"` // 0–23
	CheckIns     int     `json:"check_ins"`
	CheckOuts    int     `json:"check_outs"`
	ActiveStaff  int     `json:"active_staff"` // distinct employees present in that hour
	AvgWorkHours float64 `json:"avg_work_hours"`
}

// ─── Payroll Summary ──────────────────────────────────────────────────────────

// NewPayrollSummary provides salary totals and per-role breakdown.
type NewPayrollSummary struct {
	TotalMonthlySalary float64                 `json:"total_monthly_salary"`
	TotalEmployees     int                     `json:"total_employees"`
	AvgSalary          float64                 `json:"avg_salary"`
	ByRole             []NewStaffRoleBreakdown `json:"by_role"`
}

// ─── Aggregated Response Types ────────────────────────────────────────────────

// NewDefaultStaffResponse is returned by the default (last-30-days) endpoint.
type NewDefaultStaffResponse struct {
	Overview                NewStaffOverviewCard           `json:"overview"`
	StatsCard               NewStaffStatsCard              `json:"stats_card"`
	DailyTrend              []NewStaffTrendPoint           `json:"daily_trend"`
	WeeklyTrend             []NewStaffTrendPoint           `json:"weekly_trend"`
	MonthlyTrend            []NewStaffTrendPoint           `json:"monthly_trend"`
	YearlyTrend             []NewStaffTrendPoint           `json:"yearly_trend"`
	DailySummary            []NewDailyAttendanceSummary    `json:"daily_summary"`
	EmployeeAttendance      []NewEmployeeAttendanceSummary `json:"employee_attendance"`
	MostPresentEmployees    []NewMostPresentEmployee       `json:"most_present_employees"`
	MostAbsentEmployees     []NewMostAbsentEmployee        `json:"most_absent_employees"`
	LongestServiceEmployees []NewLongestServiceEmployee    `json:"longest_service_employees"`
	RoleBreakdown           []NewStaffRoleBreakdown        `json:"role_breakdown"`
	LeaveAnalysis           NewLeaveAnalysis               `json:"leave_analysis"`
	PeakHours               []NewStaffPeakHour             `json:"peak_hours"`
	PayrollSummary          NewPayrollSummary              `json:"payroll_summary"`
}

// NewCustomRangeStaffResponse is returned by the custom-date-range endpoint.
// Trend fields are paginated; everything else matches the default response.
type NewCustomRangeStaffResponse struct {
	Overview                NewStaffOverviewCard           `json:"overview"`
	StatsCard               NewStaffStatsCard              `json:"stats_card"`
	DailyTrend              *NewStaffPaginatedTrendPoints  `json:"daily_trend"`
	WeeklyTrend             *NewStaffPaginatedTrendPoints  `json:"weekly_trend"`
	MonthlyTrend            *NewStaffPaginatedTrendPoints  `json:"monthly_trend"`
	YearlyTrend             *NewStaffPaginatedTrendPoints  `json:"yearly_trend"`
	DailySummary            []NewDailyAttendanceSummary    `json:"daily_summary"`
	EmployeeAttendance      []NewEmployeeAttendanceSummary `json:"employee_attendance"`
	MostPresentEmployees    []NewMostPresentEmployee       `json:"most_present_employees"`
	MostAbsentEmployees     []NewMostAbsentEmployee        `json:"most_absent_employees"`
	LongestServiceEmployees []NewLongestServiceEmployee    `json:"longest_service_employees"`
	RoleBreakdown           []NewStaffRoleBreakdown        `json:"role_breakdown"`
	LeaveAnalysis           NewLeaveAnalysis               `json:"leave_analysis"`
	PeakHours               []NewStaffPeakHour             `json:"peak_hours"`
	PayrollSummary          NewPayrollSummary              `json:"payroll_summary"`
}

// NewStaffCustomRangeReportRequest is the request payload for custom-range reports.
type NewStaffCustomRangeReportRequest struct {
	From  time.Time `json:"from"`
	To    time.Time `json:"to"`
	Limit int       `json:"limit"`
	Page  int       `json:"page"`
}

// ─── Shared Table Trend Point ────────────────────────────────────────────────

type NewTableTrendPoint struct {
	Period        string  `json:"period"`
	TotalSessions int     `json:"total_sessions"`
	AvgOccupancy  float64 `json:"avg_occupancy"`
	TotalRevenue  float64 `json:"total_revenue"`
}

// ─── Pagination ───────────────────────────────────────────────────────────────

type NewTablePaginatedTrendPoints struct {
	Data       []NewTableTrendPoint `json:"data"`
	Pagination NewPaginationInfo    `json:"pagination"`
}

// ─── Default Table Report Response ───────────────────────────────────────────

type NewDefaultTableResponse struct {
	Overview            NewTableOverviewCard     `json:"overview"`              // Last 30 days
	StatsCard           NewTableStatsCard        `json:"stats_card"`            // All time
	DailyTrend          []NewTableTrendPoint     `json:"daily_trend"`           // Last 7 days
	WeeklyTrend         []NewTableTrendPoint     `json:"weekly_trend"`          // Last 7 weeks
	MonthlyTrend        []NewTableTrendPoint     `json:"monthly_trend"`         // Last 7 months
	YearlyTrend         []NewTableTrendPoint     `json:"yearly_trend"`          // Last 7 years
	TopTables           []NewTopTable            `json:"top_tables"`            // Most used tables
	TableUsageBreakdown []NewTableUsageBreakdown `json:"table_usage_breakdown"` // Usage by table
	PeakHours           []NewTablePeakHour       `json:"peak_hours"`            // Peak hours for tables
	OccupancyRate       []NewOccupancyRate       `json:"occupancy_rate"`        // Occupancy by hour
	AvgSessionDuration  float64                  `json:"avg_session_duration"`  // Average session time
}

// ─── Custom Range Table Report Response ──────────────────────────────────────

type NewCustomRangeTableResponse struct {
	Overview            NewTableOverviewCard          `json:"overview"`              // For date range
	StatsCard           NewTableStatsCard             `json:"stats_card"`            // All time
	DailyTrend          *NewTablePaginatedTrendPoints `json:"daily_trend"`           // Paginated daily data
	WeeklyTrend         *NewTablePaginatedTrendPoints `json:"weekly_trend"`          // Paginated weekly data
	MonthlyTrend        *NewTablePaginatedTrendPoints `json:"monthly_trend"`         // Paginated monthly data
	YearlyTrend         *NewTablePaginatedTrendPoints `json:"yearly_trend"`          // Paginated yearly data
	TopTables           []NewTopTable                 `json:"top_tables"`            // Most used tables
	TableUsageBreakdown []NewTableUsageBreakdown      `json:"table_usage_breakdown"` // Usage by table
	PeakHours           []NewTablePeakHour            `json:"peak_hours"`            // Peak hours
	OccupancyRate       []NewOccupancyRate            `json:"occupancy_rate"`        // Occupancy by hour
	AvgSessionDuration  float64                       `json:"avg_session_duration"`  // Average session time
}

// ─── Table Overview Card ─────────────────────────────────────────────────────

type NewTableOverviewCard struct {
	TotalTables        int     `json:"total_tables"`
	ActiveTables       int     `json:"active_tables"`        // Currently occupied
	TotalSessions      int     `json:"total_sessions"`       // In period
	AvgOccupancyRate   float64 `json:"avg_occupancy_rate"`   // Percentage
	TotalTableRevenue  float64 `json:"total_table_revenue"`  // Revenue from table orders
	AvgSessionDuration float64 `json:"avg_session_duration"` // In minutes
	PeakOccupancyHour  int     `json:"peak_occupancy_hour"`  // Hour with highest occupancy
	PeakOccupancyRate  float64 `json:"peak_occupancy_rate"`  // Highest occupancy rate
}

// ─── Table Stats Card (All Time) ─────────────────────────────────────────────

type NewTableStatsCard struct {
	TotalTables          int     `json:"total_tables"`
	TotalCapacity        int     `json:"total_capacity"`
	TotalSessionsAllTime int     `json:"total_sessions_all_time"`
	TotalTableRevenue    float64 `json:"total_table_revenue"`
	AvgSessionDuration   float64 `json:"avg_session_duration"`
	MostUsedTable        int     `json:"most_used_table"`
	MostUsedTableCount   int     `json:"most_used_table_count"`
	BusiestDay           string  `json:"busiest_day"` // Day of week
}

// ─── Top Tables ──────────────────────────────────────────────────────────────

type NewTopTable struct {
	TableNumber    int     `json:"table_number"`
	Capacity       int     `json:"capacity"`
	TotalSessions  int     `json:"total_sessions"`
	TotalRevenue   float64 `json:"total_revenue"`
	AvgSessionTime float64 `json:"avg_session_time"` // In minutes
	OccupancyRate  float64 `json:"occupancy_rate"`   // Percentage
	TotalCustomers int     `json:"total_customers"`
}

// ─── Table Usage Breakdown ───────────────────────────────────────────────────

type NewTableUsageBreakdown struct {
	TableNumber    int     `json:"table_number"`
	Capacity       int     `json:"capacity"`
	TotalSessions  int     `json:"total_sessions"`
	TotalHoursUsed float64 `json:"total_hours_used"`
	TotalRevenue   float64 `json:"total_revenue"`
	UsagePercent   float64 `json:"usage_percent"`   // % of total sessions
	RevenuePercent float64 `json:"revenue_percent"` // % of total revenue
	AvgOrderValue  float64 `json:"avg_order_value"`
}

// ─── Table Peak Hour ─────────────────────────────────────────────────────────

type NewTablePeakHour struct {
	Hour          int     `json:"hour"`
	ActiveTables  int     `json:"active_tables"`
	OccupancyRate float64 `json:"occupancy_rate"`
	TotalRevenue  float64 `json:"total_revenue"`
	SessionsCount int     `json:"sessions_count"`
}

// ─── Occupancy Rate ──────────────────────────────────────────────────────────

type NewOccupancyRate struct {
	Hour          int     `json:"hour"`
	OccupiedCount int     `json:"occupied_count"`
	TotalCapacity int     `json:"total_capacity"`
	Rate          float64 `json:"rate"`
}

// ─── Request Types ───────────────────────────────────────────────────────────

type NewTableDefaultReportRequest struct {
	// No parameters needed
}

type NewTableCustomRangeReportRequest struct {
	From  time.Time `json:"from"`
	To    time.Time `json:"to"`
	Page  int       `json:"page"`
	Limit int       `json:"limit"`
}

type NewCustomerTrendPoint struct {
	Period     string `json:"period"`
	NewUsers   int    `json:"new_users"`
	TotalUsers int    `json:"total_users"`
}

// ─── Pagination ───────────────────────────────────────────────────────────────

type NewCustomerPaginatedTrendPoints struct {
	Data       []NewCustomerTrendPoint `json:"data"`
	Pagination NewPaginationInfo       `json:"pagination"`
}

// ─── Default Customer Report Response ─────────────────────────────────────────

type NewDefaultCustomerResponse struct {
	Overview          NewCustomerOverviewCard `json:"overview"`           // Last 30 days
	StatsCard         NewCustomerStatsCard    `json:"stats_card"`         // All time
	DailyTrend        []NewCustomerTrendPoint `json:"daily_trend"`        // Last 7 days
	WeeklyTrend       []NewCustomerTrendPoint `json:"weekly_trend"`       // Last 7 weeks
	MonthlyTrend      []NewCustomerTrendPoint `json:"monthly_trend"`      // Last 7 months
	YearlyTrend       []NewCustomerTrendPoint `json:"yearly_trend"`       // Last 7 years
	TopCustomers      []NewTopCustomer        `json:"top_customers"`      // Last 30 days
	FrequentCustomers []NewFrequentCustomer   `json:"frequent_customers"` // Last 30 days
	RetentionMetrics  NewRetentionMetrics     `json:"retention_metrics"`  // Last 30 days
	CustomerSegments  []NewCustomerSegment    `json:"customer_segments"`  // Based on spend
	StreakAnalytics   NewStreakAnalytics      `json:"streak_analytics"`   // Token streak data
	TokenAnalytics    NewTokenAnalytics       `json:"token_analytics"`    // Token economy
}

// ─── Custom Range Customer Report Response ───────────────────────────────────

type NewCustomRangeCustomerResponse struct {
	Overview          NewCustomerOverviewCard          `json:"overview"`           // For date range
	StatsCard         NewCustomerStatsCard             `json:"stats_card"`         // All time
	DailyTrend        *NewCustomerPaginatedTrendPoints `json:"daily_trend"`        // Paginated daily data
	WeeklyTrend       *NewCustomerPaginatedTrendPoints `json:"weekly_trend"`       // Paginated weekly data
	MonthlyTrend      *NewCustomerPaginatedTrendPoints `json:"monthly_trend"`      // Paginated monthly data
	YearlyTrend       *NewCustomerPaginatedTrendPoints `json:"yearly_trend"`       // Paginated yearly data
	TopCustomers      []NewTopCustomer                 `json:"top_customers"`      // For date range
	FrequentCustomers []NewFrequentCustomer            `json:"frequent_customers"` // For date range
	RetentionMetrics  NewRetentionMetrics              `json:"retention_metrics"`  // For date range
	CustomerSegments  []NewCustomerSegment             `json:"customer_segments"`  // For date range
	StreakAnalytics   NewStreakAnalytics               `json:"streak_analytics"`   // For date range
	TokenAnalytics    NewTokenAnalytics                `json:"token_analytics"`    // For date range
}

// ─── Customer Overview Card ──────────────────────────────────────────────────

type NewCustomerOverviewCard struct {
	TotalCustomers       int     `json:"total_customers"`
	NewCustomers         int     `json:"new_customers"`
	ActiveCustomers      int     `json:"active_customers"`    // Customers with orders in period
	ReturningCustomers   int     `json:"returning_customers"` // Customers with >1 order
	AvgOrdersPerCustomer float64 `json:"avg_orders_per_customer"`
	AvgSpendPerCustomer  float64 `json:"avg_spend_per_customer"`
	GrowthPercent        float64 `json:"growth_percent"` // vs previous period
}

// ─── Customer Stats Card (All Time) ──────────────────────────────────────────

type NewCustomerStatsCard struct {
	TotalCustomers        int     `json:"total_customers"`
	TotalOrders           int     `json:"total_orders"`
	TotalRevenue          float64 `json:"total_revenue"`
	AvgLifetimeValue      float64 `json:"avg_lifetime_value"`
	TotalTokensIssued     float64 `json:"total_tokens_issued"`
	TotalTokensRedeemed   float64 `json:"total_tokens_redeemed"`
	ActiveStreakCustomers int     `json:"active_streak_customers"`
}

// ─── Top Customers ───────────────────────────────────────────────────────────

type NewTopCustomer struct {
	CustomerID    string  `json:"customer_id"`
	CustomerName  string  `json:"customer_name"`
	PhoneNumber   string  `json:"phone_number"`
	TotalOrders   int     `json:"total_orders"`
	TotalSpent    float64 `json:"total_spent"`
	AvgOrderValue float64 `json:"avg_order_value"`
	LastOrderDate string  `json:"last_order_date"`
}

// ─── Frequent Customers ──────────────────────────────────────────────────────

type NewFrequentCustomer struct {
	CustomerID         string  `json:"customer_id"`
	CustomerName       string  `json:"customer_name"`
	PhoneNumber        string  `json:"phone_number"`
	VisitFrequency     float64 `json:"visit_frequency"` // Visits per month
	DaysSinceLastVisit int     `json:"days_since_last_visit"`
	TotalOrders        int     `json:"total_orders"`
	FavoriteCategory   string  `json:"favorite_category"`
}

// ─── Retention Metrics ───────────────────────────────────────────────────────

type NewRetentionMetrics struct {
	RetentionRate30Days  float64 `json:"retention_rate_30_days"`
	RetentionRate90Days  float64 `json:"retention_rate_90_days"`
	ChurnRate            float64 `json:"churn_rate"`
	RepeatPurchaseRate   float64 `json:"repeat_purchase_rate"`
	AvgDaysBetweenOrders float64 `json:"avg_days_between_orders"`
}

// ─── Customer Segment ────────────────────────────────────────────────────────

type NewCustomerSegment struct {
	Segment      string  `json:"segment"` // "High Spender", "Regular", "Occasional", "New"
	Count        int     `json:"count"`
	Percent      float64 `json:"percent"`
	AvgSpend     float64 `json:"avg_spend"`
	MinSpend     float64 `json:"min_spend"`
	MaxSpend     float64 `json:"max_spend"`
	TotalRevenue float64 `json:"total_revenue"`
}

// ─── Streak Analytics ────────────────────────────────────────────────────────

type NewStreakAnalytics struct {
	TotalStreakCustomers   int                     `json:"total_streak_customers"`
	AvgStreakLength        float64                 `json:"avg_streak_length"`
	MaxStreakLength        int                     `json:"max_streak_length"`
	StreakDistribution     []NewStreakDistribution `json:"streak_distribution"`
	MonthlyActiveStreakers int                     `json:"monthly_active_streakers"`
}

type NewStreakDistribution struct {
	StreakRange string  `json:"streak_range"` // "1-3 days", "4-7 days", "8-14 days", "15+ days"
	Count       int     `json:"count"`
	Percent     float64 `json:"percent"`
}

// ─── Token Analytics ─────────────────────────────────────────────────────────

type NewTokenAnalytics struct {
	TotalTokensEarned    float64               `json:"total_tokens_earned"`
	TotalTokensSpent     float64               `json:"total_tokens_spent"`
	ActiveTokenBalance   float64               `json:"active_token_balance"`
	AvgTokensPerCustomer float64               `json:"avg_tokens_per_customer"`
	TokenRedemptionRate  float64               `json:"token_redemption_rate"` // % of earned tokens that are spent
	TopTokenEarners      []NewTopTokenCustomer `json:"top_token_earners"`
}

type NewTopTokenCustomer struct {
	CustomerName string  `json:"customer_name"`
	PhoneNumber  string  `json:"phone_number"`
	TokensEarned float64 `json:"tokens_earned"`
	TokensSpent  float64 `json:"tokens_spent"`
	TokenBalance float64 `json:"token_balance"`
}

// ─── Request Types ───────────────────────────────────────────────────────────

type NewCustomerDefaultReportRequest struct {
	// No parameters needed for default report
}

type NewCustomerCustomRangeReportRequest struct {
	From  time.Time `json:"from"`
	To    time.Time `json:"to"`
	Page  int       `json:"page"`  // For paginated trends
	Limit int       `json:"limit"` // For paginated trends (default 10, max 50)
}

// ─── Shared Sales Trend Point ────────────────────────────────────────────────

type NewSalesTrendPoint struct {
	Period   string  `json:"period"`
	Orders   int     `json:"orders"`
	Revenue  float64 `json:"revenue"`
	Discount float64 `json:"discount"`
}

// ─── Pagination ───────────────────────────────────────────────────────────────

type NewSalesPaginatedTrendPoints struct {
	Data       []NewSalesTrendPoint `json:"data"`
	Pagination NewPaginationInfo    `json:"pagination"`
}

// ─── Default Sales Report Response ───────────────────────────────────────────

type NewDefaultSalesResponse struct {
	Overview             NewSalesOverviewCard      `json:"overview"`               // Last 30 days
	StatsCard            NewSalesStatsCard         `json:"stats_card"`             // All time
	DailyTrend           []NewSalesTrendPoint      `json:"daily_trend"`            // Last 7 days
	WeeklyTrend          []NewSalesTrendPoint      `json:"weekly_trend"`           // Last 7 weeks
	MonthlyTrend         []NewSalesTrendPoint      `json:"monthly_trend"`          // Last 7 months
	YearlyTrend          []NewSalesTrendPoint      `json:"yearly_trend"`           // Last 7 years
	TopSellingItems      []NewTopSellingItem       `json:"top_selling_items"`      // Last 30 days
	TopCategories        []NewTopCategory          `json:"top_categories"`         // Last 30 days
	OrderStatusBreakdown []NewOrderStatusBreakdown `json:"order_status_breakdown"` // Last 30 days
	TablePerformance     []NewTablePerformance     `json:"table_performance"`      // Last 30 days
	StaffPerformance     []NewStaffPerformance     `json:"staff_performance"`      // Last 30 days
	HourlySales          []NewHourlySalesPoint     `json:"hourly_sales"`           // Last 30 days
	DailySales           []NewDailySalesPoint      `json:"daily_sales"`            // Last 30 days
	MenuItemsOrderStats  []NewMenuItemOrderStat    `json:"menu_items_order_stats"` // Last 30 days
}

// ─── Custom Range Sales Report Response ──────────────────────────────────────

type NewCustomRangeSalesResponse struct {
	Overview             NewSalesOverviewCard          `json:"overview"`               // For the date range
	StatsCard            NewSalesStatsCard             `json:"stats_card"`             // All time
	DailyTrend           *NewSalesPaginatedTrendPoints `json:"daily_trend"`            // Paginated daily data
	WeeklyTrend          *NewSalesPaginatedTrendPoints `json:"weekly_trend"`           // Paginated weekly data
	MonthlyTrend         *NewSalesPaginatedTrendPoints `json:"monthly_trend"`          // Paginated monthly data
	YearlyTrend          *NewSalesPaginatedTrendPoints `json:"yearly_trend"`           // Paginated yearly data
	TopSellingItems      []NewTopSellingItem           `json:"top_selling_items"`      // For the date range
	TopCategories        []NewTopCategory              `json:"top_categories"`         // For the date range
	OrderStatusBreakdown []NewOrderStatusBreakdown     `json:"order_status_breakdown"` // For the date range
	TablePerformance     []NewTablePerformance         `json:"table_performance"`      // For the date range
	StaffPerformance     []NewStaffPerformance         `json:"staff_performance"`      // For the date range
	HourlySales          []NewHourlySalesPoint         `json:"hourly_sales"`           // For the date range
	DailySales           []NewDailySalesPoint          `json:"daily_sales"`            // For the date range
	MenuItemsOrderStats  []NewMenuItemOrderStat        `json:"menu_items_order_stats"` // Last 30 days

}

// ─── Sales Overview Card ─────────────────────────────────────────────────────

type NewSalesOverviewCard struct {
	TotalOrders       int     `json:"total_orders"`
	TotalRevenue      float64 `json:"total_revenue"`
	TotalDiscounts    float64 `json:"total_discounts"`
	AverageOrderValue float64 `json:"average_order_value"`
	ItemsPerOrder     float64 `json:"items_per_order"`
	CompletionRate    float64 `json:"completion_rate"`
	GrowthPercent     float64 `json:"growth_percent"`
}

// ─── Sales Stats Card (All Time) ─────────────────────────────────────────────

type NewSalesStatsCard struct {
	TotalOrders           int     `json:"total_orders"`
	CompletedOrders       int     `json:"completed_orders"`
	CancelledOrders       int     `json:"cancelled_orders"`
	TotalRevenue          float64 `json:"total_revenue"`
	TotalDiscounts        float64 `json:"total_discounts"`
	AverageOrderValue     float64 `json:"average_order_value"`
	UniqueCustomers       int     `json:"unique_customers"`
	CompletionRatePercent float64 `json:"completion_rate_percent"`
}

// ─── Top Selling Items ───────────────────────────────────────────────────────

type NewTopSellingItem struct {
	ItemID       string  `json:"item_id"`
	ItemName     string  `json:"item_name"`
	CategoryName string  `json:"category_name"`
	Quantity     float64 `json:"quantity"` // ← was int, now float64
	Revenue      float64 `json:"revenue"`
	OrderCount   int     `json:"order_count"`
}

// ─── Top Categories ──────────────────────────────────────────────────────────

type NewTopCategory struct {
	CategoryID   string  `json:"category_id"`
	CategoryName string  `json:"category_name"`
	Orders       int     `json:"orders"`
	Revenue      float64 `json:"revenue"`
	ItemsCount   float64 `json:"items_count"` // ← was int, now float64
}

// ─── Menu Item Order Stats ────────────────────────────────────────────────────

type NewMenuItemOrderStat struct {
	ItemID        string  `json:"item_id"`
	ItemName      string  `json:"item_name"`
	CategoryName  string  `json:"category_name"`
	ImageURL      string  `json:"image_url"`
	Price         float64 `json:"price"`
	TotalOrders   int     `json:"total_orders"`
	TotalQuantity float64 `json:"total_quantity"`
	TotalRevenue  float64 `json:"total_revenue"`
}

// ─── Order Status Breakdown ──────────────────────────────────────────────────

type NewOrderStatusBreakdown struct {
	Status  string  `json:"status"`
	Count   int     `json:"count"`
	Revenue float64 `json:"revenue"`
	Percent float64 `json:"percent"`
}

// ─── Table Performance ───────────────────────────────────────────────────────

type NewTablePerformance struct {
	TableNumber       int     `json:"table_number"`
	TotalOrders       int     `json:"total_orders"`
	TotalRevenue      float64 `json:"total_revenue"`
	AverageOrderValue float64 `json:"average_order_value"`
	TotalCustomers    int     `json:"total_customers"`
}

// ─── Staff Performance ───────────────────────────────────────────────────────

type NewStaffPerformance struct {
	StaffID           string  `json:"staff_id"`
	StaffName         string  `json:"staff_name"`
	Role              string  `json:"role"`
	OrdersServed      int     `json:"orders_served"`
	TotalRevenue      float64 `json:"total_revenue"`
	AverageOrderValue float64 `json:"average_order_value"`
}

// ─── Hourly Sales ────────────────────────────────────────────────────────────

type NewHourlySalesPoint struct {
	Hour    int     `json:"hour"`
	Orders  int     `json:"orders"`
	Revenue float64 `json:"revenue"`
}

// ─── Daily Sales ─────────────────────────────────────────────────────────────

type NewDailySalesPoint struct {
	DayOfWeek string  `json:"day_of_week"`
	Orders    int     `json:"orders"`
	Revenue   float64 `json:"revenue"`
}

// ─── Request Types ───────────────────────────────────────────────────────────

type NewSalesDefaultReportRequest struct {
	// No parameters needed for default report
}

type NewSalesCustomRangeReportRequest struct {
	From  time.Time `json:"from"`
	To    time.Time `json:"to"`
	Page  int       `json:"page"`  // For paginated trends
	Limit int       `json:"limit"` // For paginated trends (default 10, max 50)
}

// ─── Shared ───────────────────────────────────────────────────────────────────

type NewTrendPoint struct {
	Period  string  `json:"period"`
	Revenue float64 `json:"revenue"`
	Orders  int     `json:"orders"`
}

// ─── Pagination ───────────────────────────────────────────────────────────────

type NewPaginationInfo struct {
	Total    int  `json:"total"`
	HasMore  bool `json:"has_more"`
	NextPage int  `json:"next_page"`
	Limit    int  `json:"limit"`
	Page     int  `json:"page"`
}

type NewPaginatedTrendPoints struct {
	Data       []NewTrendPoint   `json:"data"`
	Pagination NewPaginationInfo `json:"pagination"`
}

// ─── Default Report Response (7 periods each) ─────────────────────────────────

type NewDefaultRevenueResponse struct {
	Overview       NewRevenueOverviewCard      `json:"overview"`        // Last 30 days
	StatsCard      NewRevenueStatsCard         `json:"stats_card"`      // All time
	DailyTrend     []NewTrendPoint             `json:"daily_trend"`     // Last 7 days
	WeeklyTrend    []NewTrendPoint             `json:"weekly_trend"`    // Last 7 weeks
	MonthlyTrend   []NewTrendPoint             `json:"monthly_trend"`   // Last 7 months
	YearlyTrend    []NewTrendPoint             `json:"yearly_trend"`    // Last 7 years
	PaymentMethods []NewPaymentMethodBreakdown `json:"payment_methods"` // Last 30 days
	Gateways       []NewGatewayBreakdown       `json:"gateways"`        // Last 30 days
	Discounts      NewDiscountAnalysis         `json:"discounts"`       // Last 30 days
	PeakHours      []NewPeakHourPoint          `json:"peak_hours"`      // Last 30 days
	PeakDays       []NewPeakDayPoint           `json:"peak_days"`       // Last 30 days
}

// ─── Custom Range Report Response (with pagination) ──────────────────────────

type NewCustomRangeRevenueResponse struct {
	Overview       NewRevenueOverviewCard      `json:"overview"`        // For the date range
	StatsCard      NewRevenueStatsCard         `json:"stats_card"`      // All time
	DailyTrend     *NewPaginatedTrendPoints    `json:"daily_trend"`     // Paginated daily data
	WeeklyTrend    *NewPaginatedTrendPoints    `json:"weekly_trend"`    // Paginated weekly data
	MonthlyTrend   *NewPaginatedTrendPoints    `json:"monthly_trend"`   // Paginated monthly data
	YearlyTrend    *NewPaginatedTrendPoints    `json:"yearly_trend"`    // Paginated yearly data
	PaymentMethods []NewPaymentMethodBreakdown `json:"payment_methods"` // For the date range
	Gateways       []NewGatewayBreakdown       `json:"gateways"`        // For the date range
	Discounts      NewDiscountAnalysis         `json:"discounts"`       // For the date range
	PeakHours      []NewPeakHourPoint          `json:"peak_hours"`      // For the date range
	PeakDays       []NewPeakDayPoint           `json:"peak_days"`       // For the date range
}

// ─── Revenue Overview Card ───────────────────────────────────────────────────

type NewRevenueOverviewCard struct {
	GrossRevenue      float64 `json:"gross_revenue"`
	NetRevenue        float64 `json:"net_revenue"`
	TotalDiscounts    float64 `json:"total_discounts"`
	TotalOrders       int     `json:"total_orders"`
	AverageOrderValue float64 `json:"average_order_value"`
	GrowthPercent     float64 `json:"growth_percent"`
}

// ─── Breakdown Types ─────────────────────────────────────────────────────────

type NewPaymentMethodBreakdown struct {
	Method  string  `json:"method"`
	Revenue float64 `json:"revenue"`
	Orders  int     `json:"orders"`
	Percent float64 `json:"percent"`
}

type NewGatewayBreakdown struct {
	Gateway string  `json:"gateway"`
	Revenue float64 `json:"revenue"`
	Orders  int     `json:"orders"`
	Percent float64 `json:"percent"`
}

type NewDiscountAnalysis struct {
	TotalDiscountsGiven float64 `json:"total_discounts_given"`
	GrossRevenue        float64 `json:"gross_revenue"`
	NetRevenue          float64 `json:"net_revenue"`
	DiscountRatePercent float64 `json:"discount_rate_percent"`
	OrdersWithDiscount  int     `json:"orders_with_discount"`
	TotalOrders         int     `json:"total_orders"`
}

type NewPeakHourPoint struct {
	Hour    int     `json:"hour"`
	Revenue float64 `json:"revenue"`
	Orders  int     `json:"orders"`
}

type NewPeakDayPoint struct {
	DayOfWeek string  `json:"day_of_week"`
	Revenue   float64 `json:"revenue"`
	Orders    int     `json:"orders"`
}

// ─── Stats Card (All Time) ───────────────────────────────────────────────────

type NewRevenueStatsCard struct {
	TotalGrossRevenue   float64 `json:"total_gross_revenue"`
	TotalNetRevenue     float64 `json:"total_net_revenue"`
	TotalOrders         int     `json:"total_orders"`
	TotalDiscounts      float64 `json:"total_discounts"`
	AverageOrderValue   float64 `json:"average_order_value"`
	TotalCustomers      int     `json:"total_customers"`
	DiscountRatePercent float64 `json:"discount_rate_percent"`
}

// ─── Request Types ───────────────────────────────────────────────────────────

type NewDefaultReportRequest struct {
	// No parameters needed for default report
}

type NewCustomRangeReportRequest struct {
	From  time.Time `json:"from"`
	To    time.Time `json:"to"`
	Page  int       `json:"page"`  // For paginated trends
	Limit int       `json:"limit"` // For paginated trends (default 10, max 50)
}
