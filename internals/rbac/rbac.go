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
	},

	models.RoleCashier: {
		// ❌ no user management
	},

	models.RoleChef: {
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
