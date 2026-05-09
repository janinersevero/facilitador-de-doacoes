package usecase

import (
	"github.com/google/uuid"

	"facilitador-de-doacoes/internal/model"
)

type CreateNecessityInput struct {
	Description string `json:"description" binding:"required"`
	Category    string `json:"category"    binding:"required"`
	IsUrgent    bool   `json:"is_urgent"`
}

type UpdateNecessityInput struct {
	Description string `json:"description"`
	Category    string `json:"category"`
	IsUrgent    *bool  `json:"is_urgent"`
}

type NecessityUseCase interface {
	Create(institutionID uuid.UUID, input CreateNecessityInput) (*model.Necessity, error)
	GetByID(id uuid.UUID) (*model.Necessity, error)
	GetByInstitutionID(institutionID uuid.UUID) ([]*model.Necessity, error)
	Update(id uuid.UUID, institutionID uuid.UUID, input UpdateNecessityInput) (*model.Necessity, error)
	Delete(id uuid.UUID, institutionID uuid.UUID) error
	UpdateStatus(id uuid.UUID, institutionID uuid.UUID, status model.NecessityStatus) (*model.Necessity, error)
}
