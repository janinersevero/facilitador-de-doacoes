package usecase

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"facilitador-de-doacoes/internal/model"
	"facilitador-de-doacoes/internal/repository"
	"facilitador-de-doacoes/pkg/asaas"
	"facilitador-de-doacoes/pkg/queue"
)

type donationUseCase struct {
	repo         repository.DonationRepository
	userRepo     repository.UserRepository
	campaignRepo repository.CampaignRepository
	client       asaas.PaymentClient
	queueClient  *asynq.Client // nil disables enqueueing (e.g. in tests)
}

func NewDonationUseCase(
	repo repository.DonationRepository,
	userRepo repository.UserRepository,
	campaignRepo repository.CampaignRepository,
	client asaas.PaymentClient,
	queueClient *asynq.Client,
) DonationUseCase {
	return &donationUseCase{
		repo:         repo,
		userRepo:     userRepo,
		campaignRepo: campaignRepo,
		client:       client,
		queueClient:  queueClient,
	}
}

func (uc *donationUseCase) Create(input CreateDonationInput) (*model.Donation, error) {
	if (input.InstitutionID == nil) == (input.CampaignID == nil) {
		return nil, model.ErrInvalidDonationTarget
	}

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

	method := input.PaymentMethod
	if method == "" {
		method = "PIX"
	}

	var customer *asaas.CustomerData
	if user.CPF != "" {
		customer = &asaas.CustomerData{
			Name:  user.Name,
			Email: user.Email,
			Phone: user.Phone,
			CPF:   user.CPF,
		}
	}

	externalRef := uuid.New().String()
	var paymentID, brCode, qrCodeURL, status string

	switch method {
	case "PIX":
		result, err := uc.client.CreatePixPayment(context.Background(), asaas.PixPaymentRequest{
			Amount:      input.Amount,
			Description: "Doação",
			ExternalRef: externalRef,
			Customer:    customer,
		})
		if err != nil {
			return nil, fmt.Errorf("create pix: %w", err)
		}
		paymentID = result.PaymentID
		brCode = result.BrCode
		qrCodeURL = result.QRCodeURL
		status = result.Status

	case "CREDIT_CARD":
		if input.CreditCard == nil {
			return nil, fmt.Errorf("credit_card data required for CREDIT_CARD payment method")
		}
		if customer == nil {
			return nil, fmt.Errorf("user CPF is required for credit card payment")
		}
		result, err := uc.client.CreateCreditCardPayment(context.Background(), asaas.CreditCardPaymentRequest{
			Amount:      input.Amount,
			Description: "Doação",
			ExternalRef: externalRef,
			Customer:    *customer,
			Card: asaas.CreditCardData{
				HolderName:    input.CreditCard.HolderName,
				Number:        input.CreditCard.Number,
				ExpiryMonth:   input.CreditCard.ExpiryMonth,
				ExpiryYear:    input.CreditCard.ExpiryYear,
				CCV:           input.CreditCard.CCV,
				PostalCode:    input.CreditCard.PostalCode,
				AddressNumber: input.CreditCard.AddressNumber,
				Phone:         input.CreditCard.Phone,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("create credit card payment: %w", err)
		}
		paymentID = result.PaymentID
		status = result.Status

	default:
		return nil, fmt.Errorf("unsupported payment method: %s", method)
	}

	donationStatus := "PENDING"
	if status == "CONFIRMED" || status == "RECEIVED" {
		donationStatus = "PAID"
	}

	donation := &model.Donation{
		ID:            uuid.New(),
		UserID:        input.UserID,
		InstitutionID: input.InstitutionID,
		CampaignID:    input.CampaignID,
		PaymentMethod: method,
		PaymentID:     paymentID,
		BrCode:        brCode,
		QRCodeURL:     qrCodeURL,
		Amount:        input.Amount,
		Status:        donationStatus,
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

func (uc *donationUseCase) HandleWebhook(paymentID, status string) error {
	donation, err := uc.repo.FindByPaymentID(paymentID)
	if err != nil {
		return err
	}

	if err := uc.repo.UpdateStatus(donation.ID, status); err != nil {
		return err
	}

	if status == "PAID" {
		if donation.CampaignID != nil && uc.campaignRepo != nil {
			if err := uc.campaignRepo.IncrementTotalRaised(*donation.CampaignID, int64(donation.Amount)); err != nil {
				return fmt.Errorf("increment campaign total: %w", err)
			}
		}

		uc.enqueueDonationPaid(donation)
	}

	return nil
}

func (uc *donationUseCase) enqueueDonationPaid(donation *model.Donation) {
	if uc.queueClient == nil {
		return
	}

	campaignID := ""
	institutionID := ""

	if donation.InstitutionID != nil {
		institutionID = donation.InstitutionID.String()
	} else if donation.CampaignID != nil {
		// Donation to a campaign — resolve institution from the campaign
		campaign, err := uc.campaignRepo.FindByID(*donation.CampaignID)
		if err != nil {
			log.Printf("[queue] erro ao buscar campanha donation_id=%s: %v", donation.ID, err)
			return
		}
		campaignID = donation.CampaignID.String()
		institutionID = campaign.InstitutionID.String()
	}

	payload := queue.DonationPaidPayload{
		DonationID:    donation.ID.String(),
		UserID:        donation.UserID.String(),
		Amount:        donation.Amount,
		InstitutionID: institutionID,
		CampaignID:    campaignID,
	}

	if err := queue.EnqueueDonationPaid(uc.queueClient, context.Background(), payload); err != nil {
		log.Printf("[queue] erro ao enfileirar notificação donation_id=%s: %v", donation.ID, err)
	}
}
