package usecase

import (
	"github.com/google/uuid"

	"facilitador-de-doacoes/internal/model"
	"facilitador-de-doacoes/internal/repository"
)

type necessityUseCase struct {
	repo repository.NecessityRepository
}

func NewNecessityUseCase(repo repository.NecessityRepository) NecessityUseCase {
	return &necessityUseCase{repo: repo}
}

func (uc *necessityUseCase) Create(institutionID uuid.UUID, input CreateNecessityInput) (*model.Necessity, error) {
	necessity := &model.Necessity{
		InstitutionID: institutionID,
		Description:   input.Description,
		Category:      input.Category,
		IsUrgent:      input.IsUrgent,
		Status:        model.NecessityStatusOpen,
	}

	if err := uc.repo.Create(necessity); err != nil {
		return nil, err
	}
	return necessity, nil
}

func (uc *necessityUseCase) GetByID(id uuid.UUID) (*model.Necessity, error) {
	return uc.repo.FindByID(id)
}

func (uc *necessityUseCase) GetByInstitutionID(institutionID uuid.UUID) ([]*model.Necessity, error) {
	return uc.repo.FindByInstitutionID(institutionID)
}

func (uc *necessityUseCase) Update(id uuid.UUID, institutionID uuid.UUID, input UpdateNecessityInput) (*model.Necessity, error) {
	necessity, err := uc.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if necessity.InstitutionID != institutionID {
		return nil, model.ErrUnauthorized
	}

	if input.Description != "" {
		necessity.Description = input.Description
	}
	if input.Category != "" {
		necessity.Category = input.Category
	}
	if input.IsUrgent != nil {
		necessity.IsUrgent = *input.IsUrgent
	}

	if err := uc.repo.Update(necessity); err != nil {
		return nil, err
	}
	return necessity, nil
}

func (uc *necessityUseCase) Delete(id uuid.UUID, institutionID uuid.UUID) error {
	necessity, err := uc.repo.FindByID(id)
	if err != nil {
		return err
	}

	if necessity.InstitutionID != institutionID {
		return model.ErrUnauthorized
	}

	return uc.repo.Delete(id)
}

func (uc *necessityUseCase) UpdateStatus(id uuid.UUID, institutionID uuid.UUID, status model.NecessityStatus) (*model.Necessity, error) {
	necessity, err := uc.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if necessity.InstitutionID != institutionID {
		return nil, model.ErrUnauthorized
	}

	necessity.Status = status

	if err := uc.repo.Update(necessity); err != nil {
		return nil, err
	}
	return necessity, nil
}
