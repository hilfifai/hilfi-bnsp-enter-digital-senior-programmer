// internal/service/customer.service.go
package service

import (
	"api-jellyfish/internal/dto"
	"api-jellyfish/internal/entity"
	errpkg "api-jellyfish/internal/pkg/errors"
	"api-jellyfish/internal/repository"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ICustomerService interface {
	CreateCustomer(ctx context.Context, req dto.CreateCustomerRequest, userID uuid.UUID) (*entity.Customer, error)
	GetCustomerByID(ctx context.Context, id uuid.UUID) (*entity.Customer, error)
	GetCustomers(ctx context.Context, activeOnly bool) ([]entity.Customer, error)
	UpdateCustomer(ctx context.Context, id uuid.UUID, req dto.UpdateCustomerRequest, userID uuid.UUID) (*entity.Customer, error)
	DeleteCustomer(ctx context.Context, id uuid.UUID) error
}

type CustomerService struct {
	customerRepo repository.ICustomerRepository
}

func NewCustomerService(customerRepo repository.ICustomerRepository) ICustomerService {
	return &CustomerService{
		customerRepo: customerRepo,
	}
}

func (s *CustomerService) CreateCustomer(ctx context.Context, req dto.CreateCustomerRequest, userID uuid.UUID) (*entity.Customer, error) {
	// Check if customer code already exists
	exists, err := s.customerRepo.IsCodeExists(ctx, req.Code, uuid.Nil)
	if err != nil {
		return nil, fmt.Errorf("failed to check customer code: %w", err)
	}
	if exists {
		return nil, errpkg.ErrCustomerCodeExists
	}

	customer := &entity.Customer{
		ID:            uuid.New(),
		Code:          req.Code,
		Name:          req.Name,
		ContactPerson: "",
		Phone:         "",
		Email:         "",
		Address:       "",
		TaxNumber:     "",
		CustomerType:  "REGULAR",
		CreditLimit:   0,
		PaymentTerms:  0,
		IsActive:      true,
		Status:        "ACTIVE",
		CreatedBy:     userID,
		UpdatedBy:     userID,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	if req.ContactPerson != nil {
		customer.ContactPerson = *req.ContactPerson
	}
	if req.Phone != nil {
		customer.Phone = *req.Phone
	}
	if req.Email != nil {
		customer.Email = *req.Email
	}
	if req.Address != nil {
		customer.Address = *req.Address
	}
	if req.TaxNumber != nil {
		customer.TaxNumber = *req.TaxNumber
	}

	if err := s.customerRepo.Create(ctx, customer); err != nil {
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}
	return customer, nil
}

func (s *CustomerService) GetCustomerByID(ctx context.Context, id uuid.UUID) (*entity.Customer, error) {
	customer, err := s.customerRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("customer not found: %w", err)
	}
	return customer, nil
}

func (s *CustomerService) GetCustomers(ctx context.Context, activeOnly bool) ([]entity.Customer, error) {
	return s.customerRepo.FindAll(ctx, activeOnly)
}

func (s *CustomerService) UpdateCustomer(ctx context.Context, id uuid.UUID, req dto.UpdateCustomerRequest, userID uuid.UUID) (*entity.Customer, error) {
	customer, err := s.customerRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("customer not found: %w", err)
	}

	// Check if the new code already exists (excluding current customer)
	if req.Code != customer.Code {
		exists, err := s.customerRepo.IsCodeExists(ctx, req.Code, id)
		if err != nil {
			return nil, fmt.Errorf("failed to check customer code: %w", err)
		}
		if exists {
			return nil, errpkg.ErrCustomerCodeExists
		}
	}

	customer.Code = req.Code
	customer.Name = req.Name
	customer.UpdatedBy = userID
	customer.UpdatedAt = time.Now()

	if req.ContactPerson != nil {
		customer.ContactPerson = *req.ContactPerson
	}
	if req.Email != nil {
		customer.Email = *req.Email
	}
	if req.Phone != nil {
		customer.Phone = *req.Phone
	}
	if req.Address != nil {
		customer.Address = *req.Address
	}
	if req.TaxNumber != nil {
		customer.TaxNumber = *req.TaxNumber
	}

	if err := s.customerRepo.Update(ctx, id, customer); err != nil {
		return nil, fmt.Errorf("failed to update customer: %w", err)
	}

	return customer, nil
}

func (s *CustomerService) DeleteCustomer(ctx context.Context, id uuid.UUID) error {
	// Check if customer exists
	_, err := s.customerRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("customer not found: %w", err)
	}

	if err := s.customerRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete customer: %w", err)
	}
	return nil
}
