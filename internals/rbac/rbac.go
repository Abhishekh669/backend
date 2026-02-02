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
)

var RolePermissions = map[models.Role][]Permission{

	models.RoleAdmin: {
		AllPermissions,
	},

	models.RoleManager: {
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
	},

	models.RoleCashier: {
		// ❌ no user management
	},

	models.RoleChef: {
		ViewFoodCategory,
		UpdateFoodCategory,
		CreateFoodCategory,
		DeleteFoodCategory,
		ViewRawMaterials,
	},

	models.RoleCustomer: {},
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
