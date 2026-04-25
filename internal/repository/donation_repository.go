package repository

import (
	"github.com/google/uuid"

	"facilitador-de-doacoes/internal/model"
)

type DonationRepository interface {
	Create(donation *model.Donation) error
	FindByID(id uuid.UUID) (*model.Donation, error)
	FindByPixID(pixID string) (*model.Donation, error)
	FindAll() ([]*model.Donation, error)
	UpdateStatus(id uuid.UUID, status string) error
}
