package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"facilitador-de-doacoes/internal/model"
	"facilitador-de-doacoes/internal/repository"
	"facilitador-de-doacoes/pkg/abacatepay"
)

// PixClient abstracts PIX payment operations
// (implemented by abacatepay.Client) for test purposes ;)
type PixClient interface {
	CreatePix(ctx context.Context, req abacatepay.CreatePixRequest) (*abacatepay.PixData, error)
}

type donationUseCase struct {
	repo         repository.DonationRepository
	userRepo     repository.UserRepository
	campaignRepo repository.CampaignRepository
	client       PixClient
}

func NewDonationUseCase(
	repo repository.DonationRepository,
	userRepo repository.UserRepository,
	campaignRepo repository.CampaignRepository,
	client PixClient,
) DonationUseCase {
	return &donationUseCase{
		repo:         repo,
		userRepo:     userRepo,
		campaignRepo: campaignRepo,
		client:       client,
	}
}

func (uc *donationUseCase) Create(input CreateDonationInput) (*model.Donation, error) {
	// Exactly one target must be provided
	if (input.InstitutionID == nil) == (input.CampaignID == nil) {
		return nil, model.ErrInvalidDonationTarget
	}

	// If donating to a campaign, validate it exists and is active
	if input.CampaignID != nil {
		campaign, err := uc.campaignRepo.FindByID(*input.CampaignID)
		if err != nil {
			return nil, fmt.Errorf("campaign not found: %w", err)
		}
		if campaign.Status != model.CampaignStatusActive {
			return nil, model.ErrCampaignNotActive
		}
	}

	user, err := uc.userRepo.FindByID(input.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
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
		return nil, fmt.Errorf("create pix: %w", err)
	}

	donation := &model.Donation{
		ID:            uuid.New(),
		UserID:        input.UserID,
		InstitutionID: input.InstitutionID,
		CampaignID:    input.CampaignID,
		PixID:         pix.ID,
		BrCode:        pix.BrCode,
		QRCodeURL:     pix.QRCodeURL,
		Amount:        input.Amount,
		Status:        "PENDING",
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

	if err := uc.repo.UpdateStatus(donation.ID, status); err != nil {
		return err
	}

	// Atomically increment campaign total when payment is confirmed
	if status == "PAID" && donation.CampaignID != nil && uc.campaignRepo != nil {
		if err := uc.campaignRepo.IncrementTotalRaised(*donation.CampaignID, int64(donation.Amount)); err != nil {
			return fmt.Errorf("increment campaign total: %w", err)
		}
	}

	return nil
}
