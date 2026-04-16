package dto

import "service-app/internal/structs"

// CreateUserRequest is the request body for creating a user.
type CreateUserRequest struct {
	Name   string `json:"name" validate:"required"`
	Email  string `json:"email" validate:"required,email"`
	RoleID *int64 `json:"role_id,omitempty"`
}

// UpdateUserRequest is the request body for updating a user.
type UpdateUserRequest struct {
	Name   string `json:"name,omitempty"`
	Email  string `json:"email,omitempty"`
	RoleID *int64 `json:"role_id,omitempty"`
}

// UserResponse is the response body returned to the client.
type UserResponse struct {
	ID     int64         `json:"id"`
	Name   string        `json:"name"`
	Email  string        `json:"email"`
	RoleID *int64        `json:"role_id,omitempty"`
	Role   *RoleResponse `json:"role,omitempty"`
}

// PaginatedResponse wraps paginated data with metadata.
type PaginatedResponse struct {
	Data any          `json:"data"`
	Meta structs.Meta `json:"meta"`
}
