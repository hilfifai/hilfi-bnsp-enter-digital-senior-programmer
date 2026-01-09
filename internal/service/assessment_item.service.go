// internal/service/assessment_item.service.go
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

type IAssessmentItemService interface {
	CreateAssessmentItem(ctx context.Context, item *entity.AssessmentItem, userID uuid.UUID) error
	GetItemsByAssessmentID(ctx context.Context, assessmentID uuid.UUID) ([]entity.AssessmentItem, error)
	GetActiveItemsByAssessmentID(ctx context.Context, assessmentID uuid.UUID) ([]entity.AssessmentItem, error)
	GetAssessmentItemByID(ctx context.Context, id uuid.UUID) (*entity.AssessmentItem, error)
	UpdateAssessmentItem(ctx context.Context, id uuid.UUID, item *entity.AssessmentItem, userID uuid.UUID) (*entity.AssessmentItem, error)
	DeleteAssessmentItem(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

type AssessmentItemService struct {
	assessmentItemRepo repository.IAssessmentItemRepository
}

func NewAssessmentItemService(
	assessmentItemRepo repository.IAssessmentItemRepository,
) IAssessmentItemService {
	return &AssessmentItemService{
		assessmentItemRepo: assessmentItemRepo,
	}
}

func (s *AssessmentItemService) CreateAssessmentItem(ctx context.Context, item *entity.AssessmentItem, userID uuid.UUID) error {
	if err := s.validateUniqueCode(ctx, item.Code, item.AssessmentID, uuid.Nil); err != nil {
		return err
	}

	item.ID = uuid.New()
	item.CreatedBy = userID
	item.UpdatedBy = userID
	item.CreatedAt = time.Now()
	item.UpdatedAt = time.Now()

	if err := s.assessmentItemRepo.Create(ctx, item); err != nil {
		return fmt.Errorf("%w: %v", errors.ErrCreateAssessmentItem, err)
	}

	return nil
}

func (s *AssessmentItemService) GetItemsByAssessmentID(ctx context.Context, assessmentID uuid.UUID) ([]entity.AssessmentItem, error) {
	items, err := s.assessmentItemRepo.FindByAssessmentID(ctx, assessmentID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrGetAssessmentItem, err)
	}
	return items, nil
}

func (s *AssessmentItemService) GetActiveItemsByAssessmentID(ctx context.Context, assessmentID uuid.UUID) ([]entity.AssessmentItem, error) {
	items, err := s.assessmentItemRepo.FindActiveByAssessmentID(ctx, assessmentID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrGetAssessmentItem, err)
	}
	return items, nil
}

func (s *AssessmentItemService) GetAssessmentItemByID(ctx context.Context, id uuid.UUID) (*entity.AssessmentItem, error) {
	item, err := s.assessmentItemRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrAssessmentItemNotFound, err)
	}
	return item, nil
}

func (s *AssessmentItemService) UpdateAssessmentItem(ctx context.Context, id uuid.UUID, item *entity.AssessmentItem, userID uuid.UUID) (*entity.AssessmentItem, error) {
	existing, err := s.assessmentItemRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrAssessmentItemNotFound, err)
	}
	if err := s.validateUniqueCode(ctx, item.Code, item.AssessmentID, id); err != nil {
		return nil, err
	}
	item.ID = id
	item.UpdatedBy = userID
	item.UpdatedAt = time.Now()
	item.CreatedBy = existing.CreatedBy
	item.CreatedAt = existing.CreatedAt

	if err := s.assessmentItemRepo.Update(ctx, id, item); err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrUpdateAssessmentItem, err)
	}

	updatedItem, err := s.assessmentItemRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrGetAssessmentItem, err)
	}

	return updatedItem, nil
}

func (s *AssessmentItemService) DeleteAssessmentItem(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	_, err := s.assessmentItemRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("%w: %v", errors.ErrAssessmentItemNotFound, err)
	}

	if err := s.assessmentItemRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("%w: %v", errors.ErrDeleteAssessmentItem, err)
	}

	return nil
}

func (s *AssessmentItemService) validateUniqueCode(ctx context.Context, code string, AssessmentID uuid.UUID, excludeID uuid.UUID) error {

	exists, err := s.assessmentItemRepo.IsCodeExists(ctx, code, AssessmentID, excludeID)
	if err != nil {
		return fmt.Errorf("%w: %v", errors.ErrAssessmentItemValidation, err)
	}
	if exists {
		return errors.ErrAssessmentItemCodeExists
	}
	return nil
}
