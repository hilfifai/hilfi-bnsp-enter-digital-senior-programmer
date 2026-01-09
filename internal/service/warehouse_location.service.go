// internal/service/warehouse_location.service.go
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

type IWarehouseLocationService interface {
	CreateLocation(ctx context.Context, req dto.CreateWarehouseLocationRequest, userID uuid.UUID) (*entity.WarehouseLocation, error)
	GetLocationByID(ctx context.Context, id uuid.UUID) (*entity.WarehouseLocation, error)
	GetLocationsByWarehouse(ctx context.Context, warehouseID uuid.UUID) ([]entity.WarehouseLocation, error)
	UpdateLocation(ctx context.Context, id uuid.UUID, req dto.UpdateWarehouseLocationRequest, userID uuid.UUID) (*entity.WarehouseLocation, error)
	DeleteLocation(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	GetAvailableLocations(ctx context.Context, warehouseID uuid.UUID, requiredCapacity int) ([]entity.WarehouseLocation, error)
}

type WarehouseLocationService struct {
	warehouseRepo repository.IWarehouseRepository
	locationRepo  repository.IWarehouseLocationRepository
}

func NewWarehouseLocationService(
	warehouseRepo repository.IWarehouseRepository,
	locationRepo repository.IWarehouseLocationRepository,
) IWarehouseLocationService {
	return &WarehouseLocationService{
		warehouseRepo: warehouseRepo,
		locationRepo:  locationRepo,
	}
}

func (s *WarehouseLocationService) CreateLocation(ctx context.Context, req dto.CreateWarehouseLocationRequest, userID uuid.UUID) (*entity.WarehouseLocation, error) {
	// Validate warehouse exists
	warehouse, err := s.warehouseRepo.FindByID(ctx, req.WarehouseID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrWarehouseNotFound, err)
	}

	// Check location code uniqueness within warehouse
	exists, err := s.locationRepo.IsCodeExists(ctx, req.Code, req.WarehouseID, uuid.Nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrLocationValidation, err)
	}
	if exists {
		return nil, errors.ErrLocationCodeExists
	}

	location := &entity.WarehouseLocation{
		WarehouseID:  req.WarehouseID,
		Code:         req.Code,
		Name:         req.Name,
		Description:  req.Description,
		Zone:         req.Zone,
		Aisle:        req.Aisle,
		Rack:         req.Rack,
		Shelf:        req.Shelf,
		Bin:          req.Bin,
		Capacity:     req.Capacity,
		CurrentUsage: 0,
		IsActive:     req.IsActive,
		CreatedBy:    userID,
		UpdatedBy:    userID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	location.ID = uuid.New()

	if err := s.locationRepo.Create(ctx, location); err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrCreateLocation, err)
	}

	// Add warehouse info to response
	location.Warehouse = warehouse

	return location, nil
}

func (s *WarehouseLocationService) GetLocationByID(ctx context.Context, id uuid.UUID) (*entity.WarehouseLocation, error) {
	location, err := s.locationRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrLocationNotFound, err)
	}

	// Get warehouse info if warehouse_id is set
	if location.WarehouseID != uuid.Nil {
		warehouse, err := s.warehouseRepo.FindByID(ctx, location.WarehouseID)
		if err == nil {
			location.Warehouse = warehouse
		}
	}

	return location, nil
}

func (s *WarehouseLocationService) GetAvailableLocations(ctx context.Context, warehouseID uuid.UUID, requiredCapacity int) ([]entity.WarehouseLocation, error) {
	// Validate warehouse exists
	_, err := s.warehouseRepo.FindByID(ctx, warehouseID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrWarehouseNotFound, err)
	}

	locations, err := s.locationRepo.GetAvailableLocations(ctx, warehouseID, requiredCapacity)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrGetLocation, err)
	}

	return locations, nil
}

func (s *WarehouseLocationService) GetLocationsByWarehouse(ctx context.Context, warehouseID uuid.UUID) ([]entity.WarehouseLocation, error) {
	// Only validate warehouse if warehouseID is not nil
	if warehouseID != uuid.Nil {
		_, err := s.warehouseRepo.FindByID(ctx, warehouseID)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", errors.ErrWarehouseNotFound, err)
		}
	}

	return s.locationRepo.FindByWarehouseID(ctx, warehouseID)
}

func (s *WarehouseLocationService) UpdateLocation(ctx context.Context, id uuid.UUID, req dto.UpdateWarehouseLocationRequest, userID uuid.UUID) (*entity.WarehouseLocation, error) {
	location, err := s.locationRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrLocationNotFound, err)
	}

	// Check location code uniqueness
	if req.Code != nil {
		exists, err := s.locationRepo.IsCodeExists(ctx, *req.Code, location.WarehouseID, id)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", errors.ErrLocationValidation, err)
		}
		if exists {
			return nil, errors.ErrLocationCodeExists
		}
		location.Code = *req.Code
	}

	if req.Name != nil {
		location.Name = *req.Name
	}
	if req.Description != nil {
		location.Description = req.Description
	}
	if req.Zone != nil {
		location.Zone = req.Zone
	}
	if req.Aisle != nil {
		location.Aisle = req.Aisle
	}
	if req.Rack != nil {
		location.Rack = req.Rack
	}
	if req.Shelf != nil {
		location.Shelf = req.Shelf
	}
	if req.Bin != nil {
		location.Bin = req.Bin
	}
	if req.Capacity != nil {
		location.Capacity = req.Capacity
	}
	if req.IsActive != nil {
		location.IsActive = *req.IsActive
	}

	location.UpdatedBy = userID
	location.UpdatedAt = time.Now()

	if err := s.locationRepo.Update(ctx, id, location); err != nil {
		return nil, fmt.Errorf("failed to update location: %w", err)
	}

	return location, nil
}

func (s *WarehouseLocationService) DeleteLocation(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	_, err := s.locationRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("%w: %v", errors.ErrLocationNotFound, err)
	}

	if err := s.locationRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete location: %w", err)
	}

	return nil
}
