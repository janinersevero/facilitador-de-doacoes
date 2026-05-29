package usecase

import (
	"github.com/google/uuid"

	"facilitador-de-doacoes/internal/model"
)

// CreditCardInput holds credit card data for payment (never persisted).
type CreditCardInput struct {
	HolderName    string `json:"holder_name"     binding:"required"`
	Number        string `json:"number"          binding:"required"`
	ExpiryMonth   string `json:"expiry_month"    binding:"required"`
	ExpiryYear    string `json:"expiry_year"     binding:"required"`
	CCV           string `json:"ccv"             binding:"required"`
	PostalCode    string `json:"postal_code"     binding:"required"`
	AddressNumber string `json:"address_number"  binding:"required"`
	Phone         string `json:"phone"` // opcional: usado quando o usuário não tem telefone cadastrado
}

type CreateDonationInput struct {
	UserID        uuid.UUID        `json:"user_id"         binding:"required"`
	Amount        int              `json:"amount"          binding:"required,min=100"`
	InstitutionID *uuid.UUID       `json:"institution_id"`
	CampaignID    *uuid.UUID       `json:"campaign_id"`
	PaymentMethod string           `json:"payment_method"` // "PIX" (default) or "CREDIT_CARD"
	CreditCard    *CreditCardInput `json:"credit_card"`    // required when PaymentMethod = "CREDIT_CARD"
}

type DonationUseCase interface {
	Create(input CreateDonationInput) (*model.Donation, error)
	GetByID(id uuid.UUID) (*model.Donation, error)
	GetAll() ([]*model.Donation, error)
	HandleWebhook(paymentID, status string) error
}
