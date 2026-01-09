// internal/repository/assessment_item.repository.go
package repository

import (
	"api-jellyfish/internal/entity"
	database "api-jellyfish/internal/pkg/db"
	"api-jellyfish/internal/pkg/errors"
	"context"

	"github.com/google/uuid"
)

type IAssessmentItemRepository interface {
	Create(ctx context.Context, item *entity.AssessmentItem) error
	FindByAssessmentID(ctx context.Context, assessmentID uuid.UUID) ([]entity.AssessmentItem, error)
	FindByID(ctx context.Context, id uuid.UUID) (*entity.AssessmentItem, error)
	FindActiveByAssessmentID(ctx context.Context, assessmentID uuid.UUID) ([]entity.AssessmentItem, error)
	Update(ctx context.Context, id uuid.UUID, item *entity.AssessmentItem) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByAssessmentID(ctx context.Context, assessmentID uuid.UUID) error
	IsCodeExists(ctx context.Context, code string, AssesmentID uuid.UUID, excludeID uuid.UUID) (bool, error)
}

type AssessmentItemRepository struct {
	database *database.Database
}

func NewAssessmentItemRepository(database *database.Database) IAssessmentItemRepository {
	return &AssessmentItemRepository{database: database}
}

func (r *AssessmentItemRepository) IsCodeExists(ctx context.Context, code string, AssesmentID uuid.UUID, excludeID uuid.UUID) (bool, error) {
	var count int
	query := "SELECT COUNT(*) FROM m_assessment_item WHERE code = $1 and assessment_id = $2"
	args := []interface{}{code}
	args = append(args, AssesmentID)

	if excludeID != uuid.Nil {
		query += " AND id != $3"
		args = append(args, excludeID)
	}

	err := r.database.DB.Get(&count, query, args...)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
func (r *AssessmentItemRepository) Create(ctx context.Context, item *entity.AssessmentItem) error {
	query := `
		INSERT INTO m_assessment_item (
			id, assessment_id, code, name, file_id, filename, ord, is_active, 
			created_by, updated_by, created_at, updated_at
		) VALUES (
			:id, :assessment_id, :code, :name, :file_id, :filename, :ord, :is_active,
			:created_by, :updated_by, :created_at, :updated_at
		)
	`

	_, err := r.database.DB.NamedExecContext(ctx, query, item)
	return err
}

func (r *AssessmentItemRepository) FindByAssessmentID(ctx context.Context, assessmentID uuid.UUID) ([]entity.AssessmentItem, error) {
	var items []entity.AssessmentItem
	query := `SELECT 
			id, assessment_id, code, name, file_id, filename, ord, is_active,
			created_by, updated_by, created_at, updated_at
		FROM m_assessment_item 
		WHERE assessment_id = $1 
		ORDER BY ord, created_at ASC`

	err := r.database.DB.SelectContext(ctx, &items, query, assessmentID)
	return items, err
}

func (r *AssessmentItemRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.AssessmentItem, error) {
	var item entity.AssessmentItem
	query := `SELECT 
			id, assessment_id, code, name, file_id, filename, ord, is_active,
			created_by, updated_by, created_at, updated_at
		FROM m_assessment_item 
		WHERE id = $1 
		LIMIT 1`

	err := r.database.DB.GetContext(ctx, &item, query, id)
	if err != nil {
		return nil, errors.ErrAssessmentItemNotFound
	}
	return &item, nil
}

func (r *AssessmentItemRepository) FindActiveByAssessmentID(ctx context.Context, assessmentID uuid.UUID) ([]entity.AssessmentItem, error) {
	var items []entity.AssessmentItem
	query := `SELECT 
			id, assessment_id, code, name, file_id, filename, ord, is_active,
			created_by, updated_by, created_at, updated_at
		FROM m_assessment_item 
		WHERE assessment_id = $1 AND is_active = true 
		ORDER BY ord, created_at ASC`

	err := r.database.DB.SelectContext(ctx, &items, query, assessmentID)
	return items, err
}

func (r *AssessmentItemRepository) Update(ctx context.Context, id uuid.UUID, item *entity.AssessmentItem) error {
	query := `
		UPDATE m_assessment_item
		SET
			assessment_id = :assessment_id,
			code = :code,
			name = :name,
			file_id = :file_id,
			filename = :filename,
			ord = :ord,
			is_active = :is_active,
			updated_by = :updated_by,
			updated_at = :updated_at
		WHERE id = :id
	`

	item.ID = id
	_, err := r.database.DB.NamedExecContext(ctx, query, item)
	return err
}

func (r *AssessmentItemRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM m_assessment_item WHERE id = $1`
	_, err := r.database.DB.ExecContext(ctx, query, id)
	return err
}

func (r *AssessmentItemRepository) DeleteByAssessmentID(ctx context.Context, assessmentID uuid.UUID) error {
	query := `DELETE FROM m_assessment_item WHERE assessment_id = $1`
	_, err := r.database.DB.ExecContext(ctx, query, assessmentID)
	return err
}
