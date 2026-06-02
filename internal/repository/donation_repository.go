package repository

import (
	"github.com/google/uuid"

	"facilitador-de-doacoes/internal/model"
)

type DonationRepository interface {
	Create(donation *model.Donation) error
	FindByID(id uuid.UUID) (*model.Donation, error)
	FindByPaymentID(paymentID string) (*model.Donation, error)
	FindAll() ([]*model.Donation, error)
	UpdateStatus(id uuid.UUID, status string) error
	GetRanking(limit int) ([]*model.RankingEntry, error)
}
