package entity

import (
	"time"

	"github.com/google/uuid"
)

type Role struct {
	ID          uuid.UUID  `db:"id" json:"id"`
	RoleGroupID uuid.UUID  `db:"role_group_id" json:"role_group_id"`
	Code        string     `db:"code" json:"code"`
	Name        string     `db:"name" json:"name"`
	Description string     `db:"description" json:"description"`
	Active      bool       `db:"is_active" json:"is_active"`
	CreatedAt   *time.Time `db:"created_at" json:"created_at"`
	CreatedBy   *uuid.UUID `db:"created_by" json:"created_by"`
	UpdatedAt   *time.Time `db:"updated_at" json:"updated_at"`
	UpdatedBy   *uuid.UUID `db:"updated_by" json:"updated_by"`
}
type RoleWithRelations struct {
	Role
	RoleGroupName string `json:"role_group_name" db:"role_group_name"`
	CreatedByName string `json:"created_by_name,omitempty" db:"created_by_name"`
	UpdatedByName string `json:"updated_by_name,omitempty" db:"updated_by_name"`
}
