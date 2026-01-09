package repository

import (
	"api-jellyfish/internal/entity"
	"context"

	"github.com/google/uuid"
)

type IPermissionRepository interface {
	GetPermissionByRole(ctx context.Context, roleIDs []uuid.UUID) []entity.Permission
}

type PermissionRepository struct{}

func NewPermissionRepository() IPermissionRepository {
	return &PermissionRepository{}
}

func (r *PermissionRepository) GetPermissionByRole(ctx context.Context, roleIDs []uuid.UUID) []entity.Permission {
	// Stub implementation - returns empty slice
	return []entity.Permission{}
}
