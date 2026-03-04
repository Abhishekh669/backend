package rbac

import "github.com/Abhishekh669/backend/internals/models"

type Permission string

// User Management
const (
	ViewUsers      Permission = "view:users"
	CreateUsers    Permission = "create:users"
	UpdateUsers    Permission = "update:users"
	DeleteUsers    Permission = "delete:users"
	AllPermissions Permission = "*"

	//table
	CreateTable Permission = "create:table"
	DeleteTable Permission = "delete:table"
	UpdateTable Permission = "update:table"
	ViewTable   Permission = "view:table"

	//raw-materials
	CreateRawMaterials Permission = "create:raw_materials"
	DeleteRawMaterials Permission = "delete:raw_materials"
	UpdateRawMaterials Permission = "update:raw_materials"
	ViewRawMaterials   Permission = "view:raw_materials"

	//food category
	ViewFoodCategory   Permission = "view:food_category"
	CreateFoodCategory Permission = "create:food_category"
	UpdateFoodCategory Permission = "update:food_category"
	DeleteFoodCategory Permission = "delete:food_category"

	//sub cateogry
	ViewFoodSubCategory Permission = "view:food_subcategory"

	// SubCategory – Write
	CreateFoodSubCategory Permission = "create:food_subcategory"
	UpdateFoodSubCategory Permission = "update:food_subcategory"
	DeleteFoodSubCategory Permission = "delete:food_subcategory"

	// attendance
	CheckInAttendance  Permission = "checkin:attendance"
	CheckOutAttendance Permission = "checkout:attendance"
	ViewAttendance     Permission = "view:attendance"
	UpdateAttendance   Permission = "update:attendance"
	DeleteAttendance   Permission = "delete:attendance"

	ViewLeaveRequest   Permission = "view:leave_request"
	DeleteLeaveRequest Permission = "delete:leave_reqeust"
	UpdateLeaveRequest Permission = "update:leave_request"
	CancelLeaveRequest Permission = "update:cancel_leave_request"
)

var RolePermissions = map[models.Role][]Permission{

	models.RoleAdmin: {
		AllPermissions,
	},

	models.RoleManager: {
		CancelLeaveRequest,
		ViewLeaveRequest,
		UpdateLeaveRequest,
		DeleteLeaveRequest,
		CheckInAttendance,
		CheckOutAttendance,
		ViewAttendance,
		ViewUsers,
		CreateUsers,
		UpdateUsers,
		ViewRawMaterials,
		CreateRawMaterials,
		DeleteRawMaterials,
		UpdateRawMaterials,
		DeleteRawMaterials,
		// ❌ no DeleteUsers
		ViewFoodCategory,
		CreateFoodCategory,
		UpdateFoodCategory,
		DeleteFoodCategory,
		ViewFoodSubCategory,
		CreateFoodSubCategory,
		UpdateFoodSubCategory,
		DeleteFoodSubCategory,
		ViewTable,
		DeleteTable,
		UpdateTable,
		CreateTable,
	},

	models.RoleCashier: {
		CancelLeaveRequest,
		ViewAttendance,
		ViewTable,
		UpdateTable,

		// ❌ no user management
	},

	models.RoleChef: {
		CancelLeaveRequest,
		ViewAttendance,
		ViewFoodCategory,
		UpdateFoodCategory,
		CreateFoodCategory,
		DeleteFoodCategory,
		ViewRawMaterials,
		ViewTable,
	},

	models.RoleDeliveryStaff: {
		CancelLeaveRequest,
		ViewAttendance,
	},

	models.RoleWaiter: {
		CancelLeaveRequest,
		ViewAttendance,
		ViewTable,
		UpdateTable,
	},

	models.RoleCustomer: {
		ViewTable,
	},
}

func HasPermission(role *models.Role, permission Permission) bool {
	if role == nil {
		return false
	}

	permissions, ok := RolePermissions[*role]
	if !ok {
		return false
	}

	for _, p := range permissions {
		if p == AllPermissions || p == permission {
			return true
		}
	}

	return false
}
