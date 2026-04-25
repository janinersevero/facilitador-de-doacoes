package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"facilitador-de-doacoes/internal/model"
	"facilitador-de-doacoes/internal/repository"
	"facilitador-de-doacoes/pkg/abacatepay"
)

type donationUseCase struct {
	repo     repository.DonationRepository
	userRepo repository.UserRepository
	client   *abacatepay.Client
}

func NewDonationUseCase(repo repository.DonationRepository, userRepo repository.UserRepository, client *abacatepay.Client) DonationUseCase {
	return &donationUseCase{repo: repo, userRepo: userRepo, client: client}
}

func (uc *donationUseCase) Create(input CreateDonationInput) (*model.Donation, error) {
	user, err := uc.userRepo.FindByID(input.UserID)
	if err != nil {
		return nil, fmt.Errorf("usuário não encontrado: %w", err)
	}

	pixReq := abacatepay.CreatePixRequest{
		Amount:      input.Amount,
		Description: "Doação",
		ExternalID:  uuid.New().String(),
	}
	if user.CPF != "" {
		pixReq.Customer = &abacatepay.Customer{
			Name:      user.Name,
			Email:     user.Email,
			Cellphone: user.Phone,
			TaxID:     user.CPF,
		}
	}

	pix, err := uc.client.CreatePix(context.Background(), pixReq)
	if err != nil {
		return nil, fmt.Errorf("criar pix: %w", err)
	}

	donation := &model.Donation{
		ID:        uuid.New(),
		UserID:    input.UserID,
		PixID:     pix.ID,
		BrCode:    pix.BrCode,
		QRCodeURL: pix.QRCodeURL,
		Amount:    input.Amount,
		Status:    "PENDING",
	}

	if err := uc.repo.Create(donation); err != nil {
		return nil, err
	}

	return donation, nil
}

func (uc *donationUseCase) GetByID(id uuid.UUID) (*model.Donation, error) {
	return uc.repo.FindByID(id)
}

func (uc *donationUseCase) GetAll() ([]*model.Donation, error) {
	return uc.repo.FindAll()
}

func (uc *donationUseCase) HandleWebhook(pixID, status string) error {
	donation, err := uc.repo.FindByPixID(pixID)
	if err != nil {
		return err
	}
	return uc.repo.UpdateStatus(donation.ID, status)
}
