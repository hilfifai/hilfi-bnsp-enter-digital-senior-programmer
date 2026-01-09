package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateWarehouseRequest struct {
	Code          string  `json:"code" validate:"required,min=3,max=50"`
	Name          string  `json:"name" validate:"required,min=3,max=200"`
	Address       *string `json:"address,omitempty"`
	ContactPerson *string `json:"contact_person,omitempty"`
	Phone         *string `json:"phone,omitempty"`
	IsActive      bool    `json:"is_active"`
}

type UpdateWarehouseRequest struct {
	Code          string  `json:"code" validate:"required,min=3,max=50"`
	Name          string  `json:"name" validate:"required,min=3,max=200"`
	Address       *string `json:"address,omitempty"`
	ContactPerson *string `json:"contact_person,omitempty"`
	Phone         *string `json:"phone,omitempty"`
	IsActive      bool    `json:"is_active"`
}

type CreateWarehouseLocationRequest struct {
	WarehouseID uuid.UUID `json:"warehouse_id" validate:"required"`
	Code        string    `json:"code" validate:"required,min=2,max=50"`
	Name        string    `json:"name" validate:"required,min=3,max=100"`
	Description *string   `json:"description,omitempty"`
	Zone        *string   `json:"zone,omitempty"`
	Aisle       *string   `json:"aisle,omitempty"`
	Rack        *string   `json:"rack,omitempty"`
	Shelf       *string   `json:"shelf,omitempty"`
	Bin         *string   `json:"bin,omitempty"`
	Capacity    *int      `json:"capacity,omitempty"`
	IsActive    bool      `json:"is_active"`
}

type WarehouseResponse struct {
	ID            uuid.UUID `json:"id"`
	Code          string    `json:"code"`
	Name          string    `json:"name"`
	Address       *string   `json:"address,omitempty"`
	ContactPerson *string   `json:"contact_person,omitempty"`
	Phone         *string   `json:"phone,omitempty"`
	IsActive      bool      `json:"is_active"`
	LocationCount int       `json:"location_count"`
	TotalCapacity int       `json:"total_capacity"`
	UsedCapacity  int       `json:"used_capacity"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type UpdateWarehouseLocationRequest struct {
	Code        *string `json:"code,omitempty" validate:"omitempty,min=2,max=50"`
	Name        *string `json:"name,omitempty" validate:"omitempty,min=3,max=100"`
	Description *string `json:"description,omitempty"`
	Zone        *string `json:"zone,omitempty"`
	Aisle       *string `json:"aisle,omitempty"`
	Rack        *string `json:"rack,omitempty"`
	Shelf       *string `json:"shelf,omitempty"`
	Bin         *string `json:"bin,omitempty"`
	Capacity    *int    `json:"capacity,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty"`
}
