// internal/service/supplier.service.go
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

type ISupplierService interface {
	CreateSupplier(ctx context.Context, req dto.CreateSupplierRequest, userID uuid.UUID) (*entity.Supplier, error)
	GetSupplierByID(ctx context.Context, id uuid.UUID) (*entity.Supplier, error)
	GetSuppliers(ctx context.Context, activeOnly bool) ([]entity.Supplier, error)
	UpdateSupplier(ctx context.Context, id uuid.UUID, req dto.UpdateSupplierRequest, userID uuid.UUID) (*entity.Supplier, error)
	DeleteSupplier(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

type SupplierService struct {
	supplierRepo repository.ISupplierRepository
}

func NewSupplierService(supplierRepo repository.ISupplierRepository) ISupplierService {
	return &SupplierService{
		supplierRepo: supplierRepo,
	}
}

func (s *SupplierService) CreateSupplier(ctx context.Context, req dto.CreateSupplierRequest, userID uuid.UUID) (*entity.Supplier, error) {
	// Check if supplier code already exists
	exists, err := s.supplierRepo.IsCodeExists(ctx, req.Code, uuid.Nil)
	if err != nil {
		return nil, fmt.Errorf("failed to check supplier code: %w", err)
	}
	if exists {
		return nil, errors.ErrSupplierCodeExists
	}

	supplier := &entity.Supplier{
		ID:            uuid.New(),
		Code:          req.Code,
		Name:          req.Name,
		ContactPerson: req.ContactPerson,
		Email:         req.Email,
		Phone:         req.Phone,
		Address:       req.Address,
		TaxNumber:     req.TaxNumber,
		PaymentTerms:  req.PaymentTerms,
		IsActive:      true,
		CreatedBy:     userID,
		UpdatedBy:     userID,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.supplierRepo.Create(ctx, supplier); err != nil {
		return nil, fmt.Errorf("failed to create supplier: %w", err)
	}

	return supplier, nil
}

func (s *SupplierService) GetSupplierByID(ctx context.Context, id uuid.UUID) (*entity.Supplier, error) {
	supplier, err := s.supplierRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("supplier not found: %w", err)
	}
	return supplier, nil
}

func (s *SupplierService) GetSuppliers(ctx context.Context, activeOnly bool) ([]entity.Supplier, error) {
	return s.supplierRepo.FindAll(ctx, activeOnly)
}

func (s *SupplierService) UpdateSupplier(ctx context.Context, id uuid.UUID, req dto.UpdateSupplierRequest, userID uuid.UUID) (*entity.Supplier, error) {
	supplier, err := s.supplierRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("supplier not found: %w", err)
	}

	// Check if code is being changed and if new code already exists
	if req.Code != supplier.Code {
		exists, err := s.supplierRepo.IsCodeExists(ctx, req.Code, id)
		if err != nil {
			return nil, fmt.Errorf("failed to check supplier code: %w", err)
		}
		if exists {
			return nil, errors.ErrSupplierCodeExists
		}
	}

	supplier.Code = req.Code
	supplier.Name = req.Name
	supplier.ContactPerson = req.ContactPerson
	supplier.Email = req.Email
	supplier.Phone = req.Phone
	supplier.Address = req.Address
	supplier.TaxNumber = req.TaxNumber
	supplier.PaymentTerms = req.PaymentTerms
	supplier.IsActive = req.IsActive
	supplier.UpdatedBy = userID
	supplier.UpdatedAt = time.Now()

	if err := s.supplierRepo.Update(ctx, id, supplier); err != nil {
		return nil, fmt.Errorf("failed to update supplier: %w", err)
	}

	return supplier, nil
}

func (s *SupplierService) DeleteSupplier(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	_, err := s.supplierRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("supplier not found: %w", err)
	}

	if err := s.supplierRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete supplier: %w", err)
	}

	return nil
}
