// internal/service/product_category.service.go
package service

import (
	"api-jellyfish/internal/dto"
	"api-jellyfish/internal/entity"
	"api-jellyfish/internal/repository"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type IProductCategoryService interface {
	CreateCategory(ctx context.Context, req dto.CreateProductCategoryRequest, userID uuid.UUID) (*entity.ProductCategory, error)
	GetCategoryByID(ctx context.Context, id uuid.UUID) (*entity.ProductCategory, error)
	GetCategoryByCode(ctx context.Context, code string) (*entity.ProductCategory, error)
	GetCategories(ctx context.Context, activeOnly bool) ([]entity.ProductCategory, error)
	GetCategoriesWithPagination(ctx context.Context, filter dto.ProductCategoryFilterDTO) ([]entity.ProductCategory, int, error)
	UpdateCategory(ctx context.Context, id uuid.UUID, req dto.UpdateProductCategoryRequest, userID uuid.UUID) (*entity.ProductCategory, error)
	DeleteCategory(ctx context.Context, id uuid.UUID) error
}

type ProductCategoryService struct {
	categoryRepo repository.IProductCategoryRepository
}

func NewProductCategoryService(categoryRepo repository.IProductCategoryRepository) IProductCategoryService {
	return &ProductCategoryService{
		categoryRepo: categoryRepo,
	}
}

func (s *ProductCategoryService) CreateCategory(ctx context.Context, req dto.CreateProductCategoryRequest, userID uuid.UUID) (*entity.ProductCategory, error) {
	// Check if code already exists
	exists, err := s.categoryRepo.IsCodeExists(ctx, req.Code, uuid.Nil)
	if err != nil {
		return nil, fmt.Errorf("failed to check category code: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("category code already exists")
	}

	// Validate parent category if provided
	if req.ParentCategoryID != nil {
		parent, err := s.categoryRepo.FindByID(ctx, *req.ParentCategoryID)
		if err != nil || parent == nil {
			return nil, fmt.Errorf("parent category not found")
		}
	}

	category := &entity.ProductCategory{
		ID:               uuid.New(),
		Code:             req.Code,
		Name:             req.Name,
		Description:      req.Description,
		ParentCategoryID: req.ParentCategoryID,
		IsActive:         req.IsActive,
		CreatedBy:        userID,
		UpdatedBy:        userID,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := s.categoryRepo.Create(ctx, category); err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	return category, nil
}

func (s *ProductCategoryService) GetCategoryByID(ctx context.Context, id uuid.UUID) (*entity.ProductCategory, error) {
	category, err := s.categoryRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("category not found: %w", err)
	}
	return category, nil
}

func (s *ProductCategoryService) GetCategoryByCode(ctx context.Context, code string) (*entity.ProductCategory, error) {
	category, err := s.categoryRepo.FindByCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("category not found: %w", err)
	}
	return category, nil
}

func (s *ProductCategoryService) GetCategories(ctx context.Context, activeOnly bool) ([]entity.ProductCategory, error) {
	return s.categoryRepo.FindAll(ctx, activeOnly)
}

func (s *ProductCategoryService) GetCategoriesWithPagination(ctx context.Context, filter dto.ProductCategoryFilterDTO) ([]entity.ProductCategory, int, error) {
	// Set defaults
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 {
		filter.Limit = 10
	}
	if filter.SortBy == "" {
		filter.SortBy = "name"
	}
	if filter.SortOrder == "" {
		filter.SortOrder = "asc"
	}

	categories, err := s.categoryRepo.FindAll(ctx, filter.IsActive != nil && *filter.IsActive)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get categories: %w", err)
	}

	// Simple pagination - in real implementation, you'd have a Count method in repository
	total := len(categories)

	return categories, total, nil
}

func (s *ProductCategoryService) UpdateCategory(ctx context.Context, id uuid.UUID, req dto.UpdateProductCategoryRequest, userID uuid.UUID) (*entity.ProductCategory, error) {
	// Check if category exists
	existing, err := s.categoryRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("category not found: %w", err)
	}

	// Validate parent category if provided
	if req.ParentCategoryID != nil && *req.ParentCategoryID != id {
		parent, err := s.categoryRepo.FindByID(ctx, *req.ParentCategoryID)
		if err != nil || parent == nil {
			return nil, fmt.Errorf("parent category not found")
		}
	}

	existing.Name = req.Name
	existing.Description = req.Description
	existing.ParentCategoryID = req.ParentCategoryID
	existing.IsActive = req.IsActive
	existing.UpdatedBy = userID
	existing.UpdatedAt = time.Now()

	if err := s.categoryRepo.Update(ctx, id, existing); err != nil {
		return nil, fmt.Errorf("failed to update category: %w", err)
	}

	return existing, nil
}

func (s *ProductCategoryService) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	// Check if category exists
	_, err := s.categoryRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("category not found: %w", err)
	}

	if err := s.categoryRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	return nil
}
