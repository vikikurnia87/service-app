package dto

// CreateRoleRequest is the request body for creating a role.
type CreateRoleRequest struct {
	RoleName string `json:"role_name" validate:"required"`
	RoleDesc string `json:"role_desc" validate:"required"`
	RoleCode string `json:"role_code" validate:"required"`
}

// UpdateRoleRequest is the request body for updating a role.
type UpdateRoleRequest struct {
	RoleName string `json:"role_name,omitempty"`
	RoleDesc string `json:"role_desc,omitempty"`
	Status   *int16 `json:"status,omitempty"`
}

// RoleResponse is the response body returned to the client.
type RoleResponse struct {
	ID       int64  `json:"id"`
	RoleName string `json:"role_name"`
	RoleDesc string `json:"role_desc"`
	RoleCode string `json:"role_code"`
	Status   int16  `json:"status"`
}
