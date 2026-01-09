package role

// CreateRoleRequest represents a request to create a new role
type CreateRoleRequest struct {
	RoleGroupID string `json:"roleGroupId" validate:"required"`
	Code        string `json:"code" validate:"required"`
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	Active      bool   `json:"active"`
}

// UpdateRoleRequest represents a request to update an existing role
type UpdateRoleRequest struct {
	RoleGroupID string `json:"roleGroupId"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Active      bool   `json:"active"`
}
