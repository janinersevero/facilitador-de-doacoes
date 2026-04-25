package usecase

import (
	"github.com/google/uuid"

	"facilitador-de-doacoes/internal/model"
)

type CreateDonationInput struct {
	UserID uuid.UUID `json:"user_id" binding:"required"`
	Amount int       `json:"amount"  binding:"required,min=100"`
}

type DonationUseCase interface {
	Create(input CreateDonationInput) (*model.Donation, error)
	GetByID(id uuid.UUID) (*model.Donation, error)
	GetAll() ([]*model.Donation, error)
	HandleWebhook(pixID, status string) error
}
