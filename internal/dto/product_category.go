package dto

import (
	"github.com/google/uuid"
)

type CreateProductCategoryRequest struct {
	Code             string     `json:"code" validate:"required,min=2,max=50"`
	Name             string     `json:"name" validate:"required,min=2,max=200"`
	Description      *string    `json:"description,omitempty"`
	ParentCategoryID *uuid.UUID `json:"parent_category_id,omitempty"`
	IsActive         bool       `json:"is_active"`
}

type UpdateProductCategoryRequest struct {
	Name             string     `json:"name" validate:"required,min=2,max=200"`
	Description      *string    `json:"description,omitempty"`
	ParentCategoryID *uuid.UUID `json:"parent_category_id,omitempty"`
	IsActive         bool       `json:"is_active"`
}

type ProductCategoryResponse struct {
	ID               uuid.UUID  `json:"id"`
	Code             string     `json:"code"`
	Name             string     `json:"name"`
	Description      *string    `json:"description,omitempty"`
	ParentCategoryID *uuid.UUID `json:"parent_category_id,omitempty"`
	ParentCategory   *string    `json:"parent_category_name,omitempty"`
	IsActive         bool       `json:"is_active"`
	CreatedAt        string     `json:"created_at"`
	UpdatedAt        string     `json:"updated_at"`
}

type ProductCategoryFilterDTO struct {
	ParentID  *uuid.UUID `form:"parent_id,omitempty"`
	IsActive  *bool      `form:"is_active,omitempty"`
	Page      int        `form:"page" validate:"min=1"`
	Limit     int        `form:"limit" validate:"min=1,max=100"`
	SortBy    string     `form:"sort_by"`
	SortOrder string     `form:"sort_order" validate:"oneof=asc desc"`
}
