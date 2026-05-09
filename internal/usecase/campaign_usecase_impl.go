package usecase

import (
	"github.com/google/uuid"
	"github.com/lib/pq"

	"facilitador-de-doacoes/internal/model"
	"facilitador-de-doacoes/internal/repository"
)

type campaignUseCase struct {
	repo repository.CampaignRepository
}

func NewCampaignUseCase(repo repository.CampaignRepository) CampaignUseCase {
	return &campaignUseCase{repo: repo}
}

func (uc *campaignUseCase) Create(institutionID uuid.UUID, input CreateCampaignInput) (*model.Campaign, error) {
	campaign := &model.Campaign{
		InstitutionID: institutionID,
		Title:         input.Title,
		Description:   input.Description,
		GoalAmount:    input.GoalAmount,
		IsUrgent:      input.IsUrgent,
		Keywords:      pq.StringArray(input.Keywords),
		CoverImageURL: input.CoverImageURL,
		EndsAt:        input.EndsAt,
		Status:        model.CampaignStatusActive,
	}

	if err := uc.repo.Create(campaign); err != nil {
		return nil, err
	}
	return campaign, nil
}

func (uc *campaignUseCase) GetByID(id uuid.UUID) (*model.Campaign, error) {
	return uc.repo.FindByID(id)
}

func (uc *campaignUseCase) GetAll(filters repository.CampaignFilters) ([]*model.Campaign, error) {
	return uc.repo.FindAll(filters)
}

func (uc *campaignUseCase) GetByInstitutionID(institutionID uuid.UUID) ([]*model.Campaign, error) {
	return uc.repo.FindByInstitutionID(institutionID)
}

func (uc *campaignUseCase) Update(id uuid.UUID, institutionID uuid.UUID, input UpdateCampaignInput) (*model.Campaign, error) {
	campaign, err := uc.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if campaign.InstitutionID != institutionID {
		return nil, model.ErrUnauthorized
	}

	if input.Title != "" {
		campaign.Title = input.Title
	}
	if input.Description != "" {
		campaign.Description = input.Description
	}
	if input.GoalAmount > 0 {
		campaign.GoalAmount = input.GoalAmount
	}
	if input.IsUrgent != nil {
		campaign.IsUrgent = *input.IsUrgent
	}
	if input.Keywords != nil {
		campaign.Keywords = pq.StringArray(input.Keywords)
	}
	if input.CoverImageURL != "" {
		campaign.CoverImageURL = input.CoverImageURL
	}
	if input.EndsAt != nil {
		campaign.EndsAt = input.EndsAt
	}

	if err := uc.repo.Update(campaign); err != nil {
		return nil, err
	}
	return campaign, nil
}

func (uc *campaignUseCase) Delete(id uuid.UUID, institutionID uuid.UUID) error {
	campaign, err := uc.repo.FindByID(id)
	if err != nil {
		return err
	}

	if campaign.InstitutionID != institutionID {
		return model.ErrUnauthorized
	}

	return uc.repo.Delete(id)
}

func (uc *campaignUseCase) UpdateStatus(id uuid.UUID, institutionID uuid.UUID, status model.CampaignStatus) (*model.Campaign, error) {
	campaign, err := uc.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if campaign.InstitutionID != institutionID {
		return nil, model.ErrUnauthorized
	}

	campaign.Status = status

	if err := uc.repo.Update(campaign); err != nil {
		return nil, err
	}
	return campaign, nil
}
