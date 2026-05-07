package repository

import (
	"github.com/google/uuid"

	"facilitador-de-doacoes/internal/model"
)

type CampaignFilters struct {
	Keyword       string
	InstitutionID *uuid.UUID
	IsUrgent      *bool
}

type CampaignRepository interface {
	Create(campaign *model.Campaign) error
	FindByID(id uuid.UUID) (*model.Campaign, error)
	FindAll(filters CampaignFilters) ([]*model.Campaign, error)
	FindByInstitutionID(institutionID uuid.UUID) ([]*model.Campaign, error)
	Update(campaign *model.Campaign) error
	Delete(id uuid.UUID) error
	IncrementTotalRaised(id uuid.UUID, amount int64) error
}
