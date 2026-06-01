package usecase_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"facilitador-de-doacoes/internal/model"
	"facilitador-de-doacoes/internal/usecase"
	"facilitador-de-doacoes/internal/usecase/mocks"
	"facilitador-de-doacoes/pkg/asaas"
)

func instID() *uuid.UUID { v := uuid.New(); return &v }
func campID() *uuid.UUID { v := uuid.New(); return &v }

// --------------- Create — PIX ---------------

func TestCreateDonation_Success_ToInstitution(t *testing.T) {
	donationRepo := new(mocks.MockDonationRepository)
	userRepo := new(mocks.MockUserRepository)
	campaignRepo := new(mocks.MockCampaignRepository)
	paymentClient := new(mocks.MockPaymentClient)

	userID := uuid.New()
	iID := instID()
	user := &model.User{ID: userID, Name: "Alice", Email: "alice@example.com", CPF: "12345678900", Phone: "11999999999"}
	userRepo.On("FindByID", userID).Return(user, nil)

	paymentClient.On("CreatePixPayment", mock.Anything, mock.MatchedBy(func(req asaas.PixPaymentRequest) bool {
		return req.Amount == 1000 && req.Customer != nil && req.Customer.CPF == "12345678900"
	})).Return(&asaas.PixPaymentResult{
		PaymentID: "pay-123",
		BrCode:    "00020126...",
		QRCodeURL: "base64img==",
		Status:    "PENDING",
	}, nil)

	donationRepo.On("Create", mock.AnythingOfType("*model.Donation")).Return(nil)

	uc := usecase.NewDonationUseCase(donationRepo, userRepo, campaignRepo, paymentClient, nil)

	donation, err := uc.Create(usecase.CreateDonationInput{
		UserID:        userID,
		Amount:        1000,
		InstitutionID: iID,
	})

	assert.NoError(t, err)
	assert.Equal(t, userID, donation.UserID)
	assert.Equal(t, iID, donation.InstitutionID)
	assert.Nil(t, donation.CampaignID)
	assert.Equal(t, "pay-123", donation.PaymentID)
	assert.Equal(t, "PIX", donation.PaymentMethod)
	assert.Equal(t, "PENDING", donation.Status)
}

func TestCreateDonation_Success_ToCampaign(t *testing.T) {
	donationRepo := new(mocks.MockDonationRepository)
	userRepo := new(mocks.MockUserRepository)
	campaignRepo := new(mocks.MockCampaignRepository)
	paymentClient := new(mocks.MockPaymentClient)

	userID := uuid.New()
	cID := campID()
	user := &model.User{ID: userID, Name: "Bob", Email: "bob@example.com"}
	campaign := &model.Campaign{ID: *cID, Status: model.CampaignStatusActive}

	campaignRepo.On("FindByID", *cID).Return(campaign, nil)
	userRepo.On("FindByID", userID).Return(user, nil)
	paymentClient.On("CreatePixPayment", mock.Anything, mock.Anything).Return(&asaas.PixPaymentResult{PaymentID: "pay-456"}, nil)
	donationRepo.On("Create", mock.AnythingOfType("*model.Donation")).Return(nil)

	uc := usecase.NewDonationUseCase(donationRepo, userRepo, campaignRepo, paymentClient, nil)

	donation, err := uc.Create(usecase.CreateDonationInput{
		UserID:     userID,
		Amount:     500,
		CampaignID: cID,
	})

	assert.NoError(t, err)
	assert.Equal(t, cID, donation.CampaignID)
	assert.Nil(t, donation.InstitutionID)
	assert.Equal(t, "pay-456", donation.PaymentID)
}

func TestCreateDonation_InvalidTarget_BothNil(t *testing.T) {
	uc := usecase.NewDonationUseCase(nil, nil, nil, nil, nil)

	_, err := uc.Create(usecase.CreateDonationInput{UserID: uuid.New(), Amount: 1000})

	assert.ErrorIs(t, err, model.ErrInvalidDonationTarget)
}

func TestCreateDonation_InvalidTarget_BothSet(t *testing.T) {
	uc := usecase.NewDonationUseCase(nil, nil, nil, nil, nil)

	_, err := uc.Create(usecase.CreateDonationInput{
		UserID:        uuid.New(),
		Amount:        1000,
		InstitutionID: instID(),
		CampaignID:    campID(),
	})

	assert.ErrorIs(t, err, model.ErrInvalidDonationTarget)
}

func TestCreateDonation_CampaignNotActive(t *testing.T) {
	campaignRepo := new(mocks.MockCampaignRepository)
	cID := campID()
	campaign := &model.Campaign{ID: *cID, Status: model.CampaignStatusCompleted}
	campaignRepo.On("FindByID", *cID).Return(campaign, nil)

	uc := usecase.NewDonationUseCase(nil, nil, campaignRepo, nil, nil)

	_, err := uc.Create(usecase.CreateDonationInput{UserID: uuid.New(), Amount: 500, CampaignID: cID})

	assert.ErrorIs(t, err, model.ErrCampaignNotActive)
}

func TestCreateDonation_WithoutCPF(t *testing.T) {
	donationRepo := new(mocks.MockDonationRepository)
	userRepo := new(mocks.MockUserRepository)
	campaignRepo := new(mocks.MockCampaignRepository)
	paymentClient := new(mocks.MockPaymentClient)

	userID := uuid.New()
	iID := instID()
	user := &model.User{ID: userID, Name: "Bob", Email: "bob@example.com"}
	userRepo.On("FindByID", userID).Return(user, nil)

	paymentClient.On("CreatePixPayment", mock.Anything, mock.MatchedBy(func(req asaas.PixPaymentRequest) bool {
		return req.Customer == nil
	})).Return(&asaas.PixPaymentResult{PaymentID: "pay-456", BrCode: "brcode"}, nil)

	donationRepo.On("Create", mock.AnythingOfType("*model.Donation")).Return(nil)

	uc := usecase.NewDonationUseCase(donationRepo, userRepo, campaignRepo, paymentClient, nil)

	donation, err := uc.Create(usecase.CreateDonationInput{UserID: userID, Amount: 500, InstitutionID: iID})

	assert.NoError(t, err)
	assert.Equal(t, "pay-456", donation.PaymentID)
}

func TestCreateDonation_UserNotFound(t *testing.T) {
	userRepo := new(mocks.MockUserRepository)
	iID := instID()
	userID := uuid.New()
	userRepo.On("FindByID", userID).Return(nil, model.ErrNotFound)

	uc := usecase.NewDonationUseCase(nil, userRepo, nil, nil, nil)

	_, err := uc.Create(usecase.CreateDonationInput{UserID: userID, Amount: 1000, InstitutionID: iID})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}

func TestCreateDonation_PixClientError(t *testing.T) {
	donationRepo := new(mocks.MockDonationRepository)
	userRepo := new(mocks.MockUserRepository)
	paymentClient := new(mocks.MockPaymentClient)
	iID := instID()
	userID := uuid.New()
	userRepo.On("FindByID", userID).Return(&model.User{ID: userID}, nil)
	paymentClient.On("CreatePixPayment", mock.Anything, mock.Anything).Return(nil, errors.New("payment gateway down"))

	uc := usecase.NewDonationUseCase(donationRepo, userRepo, nil, paymentClient, nil)

	_, err := uc.Create(usecase.CreateDonationInput{UserID: userID, Amount: 1000, InstitutionID: iID})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "create pix")
}

func TestCreateDonation_RepoCreateError(t *testing.T) {
	donationRepo := new(mocks.MockDonationRepository)
	userRepo := new(mocks.MockUserRepository)
	paymentClient := new(mocks.MockPaymentClient)
	iID := instID()
	userID := uuid.New()
	userRepo.On("FindByID", userID).Return(&model.User{ID: userID}, nil)
	paymentClient.On("CreatePixPayment", mock.Anything, mock.Anything).Return(&asaas.PixPaymentResult{PaymentID: "pay-x"}, nil)
	donationRepo.On("Create", mock.Anything).Return(errors.New("db write error"))

	uc := usecase.NewDonationUseCase(donationRepo, userRepo, nil, paymentClient, nil)

	_, err := uc.Create(usecase.CreateDonationInput{UserID: userID, Amount: 1000, InstitutionID: iID})

	assert.EqualError(t, err, "db write error")
}

// --------------- Create — Credit Card ---------------

func TestCreateDonation_CreditCard_Success(t *testing.T) {
	donationRepo := new(mocks.MockDonationRepository)
	userRepo := new(mocks.MockUserRepository)
	paymentClient := new(mocks.MockPaymentClient)

	userID := uuid.New()
	iID := instID()
	user := &model.User{ID: userID, Name: "Carol", Email: "carol@example.com", CPF: "11122233344"}
	userRepo.On("FindByID", userID).Return(user, nil)

	paymentClient.On("CreateCreditCardPayment", mock.Anything, mock.MatchedBy(func(req asaas.CreditCardPaymentRequest) bool {
		return req.Amount == 2000 && req.Customer.CPF == "11122233344" && req.Card.Number == "4111111111111111"
	})).Return(&asaas.CreditCardPaymentResult{PaymentID: "pay-cc-1", Status: "CONFIRMED"}, nil)

	donationRepo.On("Create", mock.AnythingOfType("*model.Donation")).Return(nil)

	uc := usecase.NewDonationUseCase(donationRepo, userRepo, nil, paymentClient, nil)

	donation, err := uc.Create(usecase.CreateDonationInput{
		UserID:        userID,
		Amount:        2000,
		InstitutionID: iID,
		PaymentMethod: "CREDIT_CARD",
		CreditCard: &usecase.CreditCardInput{
			HolderName:    "CAROL SILVA",
			Number:        "4111111111111111",
			ExpiryMonth:   "12",
			ExpiryYear:    "2030",
			CCV:           "123",
			PostalCode:    "01310100",
			AddressNumber: "100",
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, "pay-cc-1", donation.PaymentID)
	assert.Equal(t, "CREDIT_CARD", donation.PaymentMethod)
	assert.Equal(t, "PAID", donation.Status) // CONFIRMED maps to PAID
}

func TestCreateDonation_CreditCard_MissingCardData(t *testing.T) {
	userRepo := new(mocks.MockUserRepository)
	userID := uuid.New()
	iID := instID()
	user := &model.User{ID: userID, CPF: "12345678900"}
	userRepo.On("FindByID", userID).Return(user, nil)

	uc := usecase.NewDonationUseCase(nil, userRepo, nil, new(mocks.MockPaymentClient), nil)

	_, err := uc.Create(usecase.CreateDonationInput{
		UserID:        userID,
		Amount:        1000,
		InstitutionID: iID,
		PaymentMethod: "CREDIT_CARD",
		// CreditCard is nil
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "credit_card data required")
}

func TestCreateDonation_CreditCard_MissingCPF(t *testing.T) {
	userRepo := new(mocks.MockUserRepository)
	userID := uuid.New()
	iID := instID()
	user := &model.User{ID: userID, Name: "No CPF"} // no CPF
	userRepo.On("FindByID", userID).Return(user, nil)

	uc := usecase.NewDonationUseCase(nil, userRepo, nil, new(mocks.MockPaymentClient), nil)

	_, err := uc.Create(usecase.CreateDonationInput{
		UserID:        userID,
		Amount:        1000,
		InstitutionID: iID,
		PaymentMethod: "CREDIT_CARD",
		CreditCard:    &usecase.CreditCardInput{HolderName: "Test", Number: "4111", ExpiryMonth: "01", ExpiryYear: "2030", CCV: "123", PostalCode: "00000-000", AddressNumber: "1"},
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "CPF is required")
}

func TestCreateDonation_UnsupportedPaymentMethod(t *testing.T) {
	userRepo := new(mocks.MockUserRepository)
	userID := uuid.New()
	iID := instID()
	user := &model.User{ID: userID}
	userRepo.On("FindByID", userID).Return(user, nil)

	uc := usecase.NewDonationUseCase(nil, userRepo, nil, new(mocks.MockPaymentClient), nil)

	_, err := uc.Create(usecase.CreateDonationInput{
		UserID:        userID,
		Amount:        1000,
		InstitutionID: iID,
		PaymentMethod: "BOLETO",
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported payment method")
}

// --------------- GetByID ---------------

func TestGetDonationByID_Success(t *testing.T) {
	donationRepo := new(mocks.MockDonationRepository)
	id := uuid.New()
	expected := &model.Donation{ID: id, Amount: 500}
	donationRepo.On("FindByID", id).Return(expected, nil)

	uc := usecase.NewDonationUseCase(donationRepo, nil, nil, nil, nil)

	donation, err := uc.GetByID(id)
	assert.NoError(t, err)
	assert.Equal(t, expected, donation)
}

func TestGetDonationByID_NotFound(t *testing.T) {
	donationRepo := new(mocks.MockDonationRepository)
	id := uuid.New()
	donationRepo.On("FindByID", id).Return(nil, model.ErrNotFound)

	uc := usecase.NewDonationUseCase(donationRepo, nil, nil, nil, nil)

	_, err := uc.GetByID(id)
	assert.ErrorIs(t, err, model.ErrNotFound)
}

// --------------- GetAll ---------------

func TestGetAllDonations_Success(t *testing.T) {
	donationRepo := new(mocks.MockDonationRepository)
	donations := []*model.Donation{{Amount: 100}, {Amount: 200}}
	donationRepo.On("FindAll").Return(donations, nil)

	uc := usecase.NewDonationUseCase(donationRepo, nil, nil, nil, nil)

	result, err := uc.GetAll()
	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

// --------------- HandleWebhook ---------------

func TestHandleWebhook_Success_NoCampaign(t *testing.T) {
	donationRepo := new(mocks.MockDonationRepository)
	id := uuid.New()
	donationRepo.On("FindByPaymentID", "pay-abc").Return(&model.Donation{ID: id, PaymentID: "pay-abc"}, nil)
	donationRepo.On("UpdateStatus", id, "PAID").Return(nil)

	uc := usecase.NewDonationUseCase(donationRepo, nil, nil, nil, nil)

	err := uc.HandleWebhook("pay-abc", "PAID")
	assert.NoError(t, err)
	donationRepo.AssertExpectations(t)
}

func TestHandleWebhook_Success_WithCampaign(t *testing.T) {
	donationRepo := new(mocks.MockDonationRepository)
	campaignRepo := new(mocks.MockCampaignRepository)

	cID := uuid.New()
	id := uuid.New()
	donation := &model.Donation{ID: id, PaymentID: "pay-camp", Amount: 1000, CampaignID: &cID}

	donationRepo.On("FindByPaymentID", "pay-camp").Return(donation, nil)
	donationRepo.On("UpdateStatus", id, "PAID").Return(nil)
	campaignRepo.On("IncrementTotalRaised", cID, int64(1000)).Return(nil)

	uc := usecase.NewDonationUseCase(donationRepo, nil, campaignRepo, nil, nil)

	err := uc.HandleWebhook("pay-camp", "PAID")
	assert.NoError(t, err)
	donationRepo.AssertExpectations(t)
	campaignRepo.AssertExpectations(t)
}

func TestHandleWebhook_NoCampaignIncrement_WhenNotPaid(t *testing.T) {
	donationRepo := new(mocks.MockDonationRepository)
	campaignRepo := new(mocks.MockCampaignRepository)

	cID := uuid.New()
	id := uuid.New()
	donation := &model.Donation{ID: id, PaymentID: "pay-x", Amount: 500, CampaignID: &cID}

	donationRepo.On("FindByPaymentID", "pay-x").Return(donation, nil)
	donationRepo.On("UpdateStatus", id, "OVERDUE").Return(nil)

	uc := usecase.NewDonationUseCase(donationRepo, nil, campaignRepo, nil, nil)

	err := uc.HandleWebhook("pay-x", "OVERDUE")
	assert.NoError(t, err)
	campaignRepo.AssertNotCalled(t, "IncrementTotalRaised")
}

func TestHandleWebhook_DonationNotFound(t *testing.T) {
	donationRepo := new(mocks.MockDonationRepository)
	donationRepo.On("FindByPaymentID", "pay-unknown").Return(nil, model.ErrNotFound)

	uc := usecase.NewDonationUseCase(donationRepo, nil, nil, nil, nil)

	err := uc.HandleWebhook("pay-unknown", "PAID")
	assert.ErrorIs(t, err, model.ErrNotFound)
}

func TestHandleWebhook_UpdateStatusError(t *testing.T) {
	donationRepo := new(mocks.MockDonationRepository)
	id := uuid.New()
	donationRepo.On("FindByPaymentID", "pay-abc").Return(&model.Donation{ID: id}, nil)
	donationRepo.On("UpdateStatus", id, "PAID").Return(errors.New("update failed"))

	uc := usecase.NewDonationUseCase(donationRepo, nil, nil, nil, nil)

	err := uc.HandleWebhook("pay-abc", "PAID")
	assert.EqualError(t, err, "update failed")
}
