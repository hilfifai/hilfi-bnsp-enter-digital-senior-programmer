// internal/service/warehouse.service.go
package service

import (
	"api-jellyfish/internal/dto"
	"api-jellyfish/internal/entity"
	"api-jellyfish/internal/pkg/errors"
	"api-jellyfish/internal/repository"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type IWarehouseService interface {
	CreateWarehouse(ctx context.Context, req dto.CreateWarehouseRequest, userID uuid.UUID) (*entity.Warehouse, error)
	GetWarehouseByID(ctx context.Context, id uuid.UUID) (*entity.Warehouse, error)
	GetAllWarehouses(ctx context.Context) ([]entity.Warehouse, error)
	UpdateWarehouse(ctx context.Context, id uuid.UUID, req dto.UpdateWarehouseRequest, userID uuid.UUID) (*entity.Warehouse, error)
	DeleteWarehouse(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	GetWarehouseStats(ctx context.Context, id uuid.UUID) (*entity.WarehouseStats, error)
}

type WarehouseService struct {
	warehouseRepo repository.IWarehouseRepository
	inventoryRepo repository.IInventoryRepository
}

func NewWarehouseService(
	warehouseRepo repository.IWarehouseRepository,
	inventoryRepo repository.IInventoryRepository,
) IWarehouseService {
	return &WarehouseService{
		warehouseRepo: warehouseRepo,
		inventoryRepo: inventoryRepo,
	}
}

func (s *WarehouseService) validateWarehouseData(ctx context.Context, req dto.CreateWarehouseRequest, excludeID uuid.UUID) error {
	// Check code uniqueness
	exists, err := s.warehouseRepo.IsCodeExists(ctx, req.Code, excludeID)
	if err != nil {
		return fmt.Errorf("%w: %v", errors.ErrWarehouseValidation, err)
	}
	if exists {
		return errors.ErrWarehouseCodeExists
	}
	return nil
}

func (s *WarehouseService) CreateWarehouse(ctx context.Context, req dto.CreateWarehouseRequest, userID uuid.UUID) (*entity.Warehouse, error) {
	if err := s.validateWarehouseData(ctx, req, uuid.Nil); err != nil {
		return nil, err
	}

	warehouse := &entity.Warehouse{
		Code:          req.Code,
		Name:          req.Name,
		Address:       req.Address,
		ContactPerson: req.ContactPerson,
		Phone:         req.Phone,
		IsActive:      req.IsActive,
		CreatedBy:     userID,
		UpdatedBy:     userID,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	warehouse.ID = uuid.New()

	if err := s.warehouseRepo.Create(ctx, warehouse); err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrCreateWarehouse, err)
	}

	return warehouse, nil
}

func (s *WarehouseService) GetWarehouseByID(ctx context.Context, id uuid.UUID) (*entity.Warehouse, error) {
	warehouse, err := s.warehouseRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrWarehouseNotFound, err)
	}
	return warehouse, nil
}

func (s *WarehouseService) GetAllWarehouses(ctx context.Context) ([]entity.Warehouse, error) {
	return s.warehouseRepo.FindAll(ctx)
}

func (s *WarehouseService) UpdateWarehouse(ctx context.Context, id uuid.UUID, req dto.UpdateWarehouseRequest, userID uuid.UUID) (*entity.Warehouse, error) {
	warehouse, err := s.warehouseRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrWarehouseNotFound, err)
	}

	if err := s.validateWarehouseData(ctx, dto.CreateWarehouseRequest{
		Code:          req.Code,
		Name:          req.Name,
		Address:       req.Address,
		ContactPerson: req.ContactPerson,
		Phone:         req.Phone,
		IsActive:      req.IsActive,
	}, id); err != nil {
		return nil, err
	}

	warehouse.Code = req.Code
	warehouse.Name = req.Name
	warehouse.Address = req.Address
	warehouse.ContactPerson = req.ContactPerson
	warehouse.Phone = req.Phone
	warehouse.IsActive = req.IsActive
	warehouse.UpdatedBy = userID
	warehouse.UpdatedAt = time.Now()

	if err := s.warehouseRepo.Update(ctx, id, warehouse); err != nil {
		return nil, fmt.Errorf("failed to update warehouse: %w", err)
	}

	return warehouse, nil
}

func (s *WarehouseService) DeleteWarehouse(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	_, err := s.warehouseRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("%w: %v", errors.ErrWarehouseNotFound, err)
	}

	if err := s.warehouseRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete warehouse: %w", err)
	}

	return nil
}

func (s *WarehouseService) GetWarehouseStats(ctx context.Context, id uuid.UUID) (*entity.WarehouseStats, error) {
	warehouse, err := s.warehouseRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrWarehouseNotFound, err)
	}

	stats, err := s.warehouseRepo.GetWarehouseStats(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get warehouse stats: %w", err)
	}

	stats.ID = warehouse.ID
	stats.Name = warehouse.Name
	return stats, nil
}
