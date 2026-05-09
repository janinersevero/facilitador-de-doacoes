package usecase

import (
	"time"

	"github.com/google/uuid"

	"facilitador-de-doacoes/internal/model"
	"facilitador-de-doacoes/internal/repository"
)

type CreateCampaignInput struct {
	Title         string     `json:"title"          binding:"required"`
	Description   string     `json:"description"    binding:"required"`
	GoalAmount    int64      `json:"goal_amount"    binding:"required,min=100"`
	IsUrgent      bool       `json:"is_urgent"`
	Keywords      []string   `json:"keywords"`
	CoverImageURL string     `json:"cover_image_url"`
	EndsAt        *time.Time `json:"ends_at"`
}

type UpdateCampaignInput struct {
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	GoalAmount    int64      `json:"goal_amount"    binding:"omitempty,min=100"`
	IsUrgent      *bool      `json:"is_urgent"`
	Keywords      []string   `json:"keywords"`
	CoverImageURL string     `json:"cover_image_url"`
	EndsAt        *time.Time `json:"ends_at"`
}

type CampaignUseCase interface {
	Create(institutionID uuid.UUID, input CreateCampaignInput) (*model.Campaign, error)
	GetByID(id uuid.UUID) (*model.Campaign, error)
	GetAll(filters repository.CampaignFilters) ([]*model.Campaign, error)
	GetByInstitutionID(institutionID uuid.UUID) ([]*model.Campaign, error)
	Update(id uuid.UUID, institutionID uuid.UUID, input UpdateCampaignInput) (*model.Campaign, error)
	Delete(id uuid.UUID, institutionID uuid.UUID) error
	UpdateStatus(id uuid.UUID, institutionID uuid.UUID, status model.CampaignStatus) (*model.Campaign, error)
}
