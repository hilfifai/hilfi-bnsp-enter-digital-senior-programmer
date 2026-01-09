// internal/service/assessment.service.go
package service

import (
	"api-jellyfish/internal/entity"
	"api-jellyfish/internal/pkg/errors"
	"api-jellyfish/internal/repository"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type IAssessmentService interface {
	CreateAssessment(ctx context.Context, assessment *entity.Assessment, userID uuid.UUID) error
	GetAllAssessments(ctx context.Context) ([]entity.Assessment, error)
	GetActiveAssessments(ctx context.Context) ([]entity.Assessment, error)
	GetAssessmentByID(ctx context.Context, id uuid.UUID) (*entity.Assessment, error)
	GetAssessmentWithItems(ctx context.Context, id uuid.UUID) (*entity.AssessmentWithItems, error)
	GetAssessmentsByModuleID(ctx context.Context, moduleID uuid.UUID) ([]entity.Assessment, error)
	UpdateAssessment(ctx context.Context, id uuid.UUID, assessment *entity.Assessment, userID uuid.UUID) (*entity.Assessment, error)
	DeleteAssessment(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

type AssessmentService struct {
	assessmentRepo     repository.IAssessmentRepository
	assessmentItemRepo repository.IAssessmentItemRepository
}

func NewAssessmentService(
	assessmentRepo repository.IAssessmentRepository,
	assessmentItemRepo repository.IAssessmentItemRepository,
) IAssessmentService {
	return &AssessmentService{
		assessmentRepo:     assessmentRepo,
		assessmentItemRepo: assessmentItemRepo,
	}
}

func (s *AssessmentService) validateUniqueCode(ctx context.Context, code string, excludeID uuid.UUID) error {

	exists, err := s.assessmentRepo.IsCodeExists(ctx, code, excludeID)
	if err != nil {
		return fmt.Errorf("%w: %v", errors.ErrAssessmentValidation, err)
	}
	if exists {
		return errors.ErrAssessmentCodeExists
	}
	return nil
}
func (s *AssessmentService) CreateAssessment(ctx context.Context, assessment *entity.Assessment, userID uuid.UUID) error {

	if err := s.validateUniqueCode(ctx, assessment.Code, uuid.Nil); err != nil {
		return err
	}
	assessment.ID = uuid.New()
	assessment.CreatedBy = userID
	assessment.UpdatedBy = userID
	assessment.CreatedAt = time.Now()
	assessment.UpdatedAt = time.Now()

	if err := s.assessmentRepo.Create(ctx, assessment); err != nil {
		return fmt.Errorf("%w: %v", errors.ErrCreateAssessment, err)
	}

	return nil
}

func (s *AssessmentService) GetAllAssessments(ctx context.Context) ([]entity.Assessment, error) {
	assessments, err := s.assessmentRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrGetAssessment, err)
	}
	return assessments, nil
}

func (s *AssessmentService) GetActiveAssessments(ctx context.Context) ([]entity.Assessment, error) {
	assessments, err := s.assessmentRepo.FindActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrGetAssessment, err)
	}
	return assessments, nil
}

func (s *AssessmentService) GetAssessmentByID(ctx context.Context, id uuid.UUID) (*entity.Assessment, error) {
	assessment, err := s.assessmentRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrAssessmentNotFound, err)
	}
	return assessment, nil
}

func (s *AssessmentService) GetAssessmentWithItems(ctx context.Context, id uuid.UUID) (*entity.AssessmentWithItems, error) {
	assessment, err := s.assessmentRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrAssessmentNotFound, err)
	}

	items, err := s.assessmentItemRepo.FindByAssessmentID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrGetAssessmentItem, err)
	}

	assessmentWithItems := &entity.AssessmentWithItems{
		Assessment: *assessment,
		Items:      items,
	}

	return assessmentWithItems, nil
}

func (s *AssessmentService) GetAssessmentsByModuleID(ctx context.Context, moduleID uuid.UUID) ([]entity.Assessment, error) {
	assessments, err := s.assessmentRepo.FindByModuleID(ctx, moduleID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrGetAssessment, err)
	}
	return assessments, nil
}

func (s *AssessmentService) UpdateAssessment(ctx context.Context, id uuid.UUID, assessment *entity.Assessment, userID uuid.UUID) (*entity.Assessment, error) {
	existing, err := s.assessmentRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrAssessmentNotFound, err)
	}
	if err := s.validateUniqueCode(ctx, assessment.Code, id); err != nil {
		return nil, err
	}
	assessment.ID = id
	assessment.UpdatedBy = userID
	assessment.UpdatedAt = time.Now()
	assessment.CreatedBy = existing.CreatedBy
	assessment.CreatedAt = existing.CreatedAt

	if err := s.assessmentRepo.Update(ctx, id, assessment); err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrUpdateAssessment, err)
	}

	updatedAssessment, err := s.assessmentRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrGetAssessment, err)
	}

	return updatedAssessment, nil
}

func (s *AssessmentService) DeleteAssessment(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	_, err := s.assessmentRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("%w: %v", errors.ErrAssessmentNotFound, err)
	}

	if err := s.assessmentRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("%w: %v", errors.ErrDeleteAssessment, err)
	}

	return nil
}
