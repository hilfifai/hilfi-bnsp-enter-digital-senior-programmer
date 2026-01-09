// internal/repository/assessment.repository.go
package repository

import (
	"api-jellyfish/internal/entity"
	database "api-jellyfish/internal/pkg/db"
	"api-jellyfish/internal/pkg/errors"
	"context"

	"github.com/google/uuid"
)

type IAssessmentRepository interface {
	Create(ctx context.Context, assessment *entity.Assessment) error
	FindAll(ctx context.Context) ([]entity.Assessment, error)
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Assessment, error)
	FindActive(ctx context.Context) ([]entity.Assessment, error)
	FindByModuleID(ctx context.Context, moduleID uuid.UUID) ([]entity.Assessment, error)
	Update(ctx context.Context, id uuid.UUID, assessment *entity.Assessment) error
	Delete(ctx context.Context, id uuid.UUID) error
	IsCodeExists(ctx context.Context, code string, excludeID uuid.UUID) (bool, error)
}

type AssessmentRepository struct {
	database *database.Database
}

func NewAssessmentRepository(database *database.Database) IAssessmentRepository {
	return &AssessmentRepository{database: database}
}

func (r *AssessmentRepository) Create(ctx context.Context, assessment *entity.Assessment) error {
	query := `
		INSERT INTO m_assessment (
			id, module_id, name, ord, is_active, code, created_by, updated_by, 
			created_at, updated_at, assigned_role_ids
		) VALUES (
			:id, :module_id, :name, :ord, :is_active, :code, :created_by, :updated_by,
			:created_at, :updated_at, :assigned_role_ids
		)
	`

	_, err := r.database.DB.NamedExecContext(ctx, query, assessment)
	return err
}

func (r *AssessmentRepository) FindAll(ctx context.Context) ([]entity.Assessment, error) {
	var assessments []entity.Assessment
	query := `SELECT 
			id, module_id, name, ord, is_active, code, created_by, updated_by,
			created_at, updated_at, assigned_role_ids
		FROM m_assessment 
		ORDER BY ord, created_at DESC`

	err := r.database.DB.SelectContext(ctx, &assessments, query)
	return assessments, err
}

func (r *AssessmentRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Assessment, error) {
	var assessment entity.Assessment
	query := `SELECT 
			id, module_id, name, ord, is_active, code, created_by, updated_by,
			created_at, updated_at, assigned_role_ids
		FROM m_assessment 
		WHERE id = $1 
		LIMIT 1`

	err := r.database.DB.GetContext(ctx, &assessment, query, id)
	if err != nil {
		return nil, errors.ErrAssessmentNotFound
	}
	return &assessment, nil
}

func (r *AssessmentRepository) FindActive(ctx context.Context) ([]entity.Assessment, error) {
	var assessments []entity.Assessment
	query := `SELECT 
			id, module_id, name, ord, is_active, code, created_by, updated_by,
			created_at, updated_at, assigned_role_ids
		FROM m_assessment 
		WHERE is_active = true 
		ORDER BY ord, created_at ASC`

	err := r.database.DB.SelectContext(ctx, &assessments, query)
	return assessments, err
}

func (r *AssessmentRepository) FindByModuleID(ctx context.Context, moduleID uuid.UUID) ([]entity.Assessment, error) {
	var assessments []entity.Assessment
	query := `SELECT 
			id, module_id, name, ord, is_active, code, created_by, updated_by,
			created_at, updated_at, assigned_role_ids
		FROM m_assessment 
		WHERE module_id = $1 
		ORDER BY ord, created_at ASC`

	err := r.database.DB.SelectContext(ctx, &assessments, query, moduleID)
	return assessments, err
}

func (r *AssessmentRepository) Update(ctx context.Context, id uuid.UUID, assessment *entity.Assessment) error {
	query := `
		UPDATE m_assessment
		SET
			module_id = :module_id,
			name = :name,
			ord = :ord,
			is_active = :is_active,
			code = :code,
			updated_by = :updated_by,
			updated_at = :updated_at,
			assigned_role_ids = :assigned_role_ids
		WHERE id = :id
	`

	assessment.ID = id
	_, err := r.database.DB.NamedExecContext(ctx, query, assessment)
	return err
}

func (r *AssessmentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM m_assessment WHERE id = $1`
	_, err := r.database.DB.ExecContext(ctx, query, id)
	return err
}

func (r *AssessmentRepository) IsCodeExists(ctx context.Context, code string, excludeID uuid.UUID) (bool, error) {
	var count int
	query := "SELECT COUNT(*) FROM m_assessment WHERE code = $1 "
	args := []interface{}{code}

	if excludeID != uuid.Nil {
		query += " AND id != $2"
		args = append(args, excludeID)
	}

	err := r.database.DB.Get(&count, query, args...)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
