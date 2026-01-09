// internal/service/role.service.go
package service

import (
	"api-jellyfish/internal/entity"
	"api-jellyfish/internal/pkg/errors"
	"api-jellyfish/internal/repository"
	"context"
	"fmt"

	"github.com/google/uuid"
)

type IRoleService interface {
	GetRoleByUser(ctx context.Context, userID uuid.UUID) ([]entity.Role, error)
	GetRoleByID(ctx context.Context, id uuid.UUID) (*entity.Role, error)
	CreateRole(ctx context.Context, role *entity.Role, userID uuid.UUID) error
	UpdateRole(ctx context.Context, role *entity.Role, userID uuid.UUID) error
	GetAllRoles(ctx context.Context) ([]entity.Role, error)
}

type RoleService struct {
	roleRepo repository.IRoleRepository
}

func NewRoleService(roleRepo repository.IRoleRepository) IRoleService {
	return &RoleService{
		roleRepo: roleRepo,
	}
}

func (s *RoleService) GetRoleByUser(ctx context.Context, userID uuid.UUID) ([]entity.Role, error) {
	roles := s.roleRepo.GetRoleByUser(ctx, userID)
	if len(roles) == 0 {
		return nil, errors.ErrRoleNotFound
	}
	return roles, nil
}

func (s *RoleService) GetRoleByID(ctx context.Context, id uuid.UUID) (*entity.Role, error) {
	role, err := s.roleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrRoleNotFound, err)
	}
	return role, nil
}

func (s *RoleService) CreateRole(ctx context.Context, role *entity.Role, userID uuid.UUID) error {
	role.ID = uuid.New()
	role.CreatedBy = &userID

	if err := s.roleRepo.Create(ctx, role); err != nil {
		return fmt.Errorf("%w: %v", errors.ErrCreateRole, err)
	}

	return nil
}

func (s *RoleService) UpdateRole(ctx context.Context, role *entity.Role, userID uuid.UUID) error {
	// Check if role exists
	existing, err := s.roleRepo.GetByID(ctx, role.ID)
	if err != nil {
		return fmt.Errorf("%w: %v", errors.ErrRoleNotFound, err)
	}

	role.CreatedBy = existing.CreatedBy
	role.CreatedAt = existing.CreatedAt
	role.UpdatedBy = &userID

	if err := s.roleRepo.Update(ctx, role); err != nil {
		return fmt.Errorf("%w: %v", errors.ErrUpdateRole, err)
	}

	return nil
}

func (s *RoleService) GetAllRoles(ctx context.Context) ([]entity.Role, error) {
	roles, err := s.roleRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrGetRole, err)
	}
	return roles, nil
}
