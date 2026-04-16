package structs

// ──────────────────────────────────────────────────────────────────────────────
// User Order Configuration
// ──────────────────────────────────────────────────────────────────────────────

// UserOrderMapping maps API field names to database columns for users.
var UserOrderMapping = OrderMapping{
	"id":         "id",
	"name":       "name",
	"email":      "email",
	"created_at": "created_at",
	"updated_at": "updated_at",
}

// UserDefaultOrders defines the default sort order for user listings.
var UserDefaultOrders = []OrderConfig{
	{Column: "created_at", Direction: "ASC"},
}

// ──────────────────────────────────────────────────────────────────────────────
// Role Order Configuration
// ──────────────────────────────────────────────────────────────────────────────

// RoleOrderMapping maps API field names to database columns for roles.
var RoleOrderMapping = OrderMapping{
	"id":         "id",
	"role_name":  "role_name",
	"role_code":  "role_code",
	"created_at": "created_at",
	"updated_at": "updated_at",
}

// RoleDefaultOrders defines the default sort order for role listings.
var RoleDefaultOrders = []OrderConfig{
	{Column: "created_at", Direction: "ASC"},
}
