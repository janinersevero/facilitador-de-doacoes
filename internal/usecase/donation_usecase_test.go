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
	"facilitador-de-doacoes/pkg/abacatepay"
)

func instID() *uuid.UUID { v := uuid.New(); return &v }
func campID() *uuid.UUID { v := uuid.New(); return &v }

// --------------- Create ---------------

func TestCreateDonation_Success_ToInstitution(t *testing.T) {
	donationRepo := new(mocks.MockDonationRepository)
	userRepo := new(mocks.MockUserRepository)
	campaignRepo := new(mocks.MockCampaignRepository)
	pixClient := new(mocks.MockPixClient)

	userID := uuid.New()
	iID := instID()
	user := &model.User{ID: userID, Name: "Alice", Email: "alice@example.com", CPF: "12345678900", Phone: "11999999999"}
	userRepo.On("FindByID", userID).Return(user, nil)

	pixClient.On("CreatePix", mock.Anything, mock.MatchedBy(func(req abacatepay.CreatePixRequest) bool {
		return req.Amount == 1000 && req.Customer != nil && req.Customer.TaxID == "12345678900"
	})).Return(&abacatepay.PixData{
		ID:        "pix-123",
		BrCode:    "00020126...",
		QRCodeURL: "https://qr.example.com/pix-123",
		Status:    "PENDING",
	}, nil)

	donationRepo.On("Create", mock.AnythingOfType("*model.Donation")).Return(nil)

	uc := usecase.NewDonationUseCase(donationRepo, userRepo, campaignRepo, pixClient)

	donation, err := uc.Create(usecase.CreateDonationInput{
		UserID:        userID,
		Amount:        1000,
		InstitutionID: iID,
	})

	assert.NoError(t, err)
	assert.Equal(t, userID, donation.UserID)
	assert.Equal(t, iID, donation.InstitutionID)
	assert.Nil(t, donation.CampaignID)
	assert.Equal(t, "pix-123", donation.PixID)
	assert.Equal(t, "PENDING", donation.Status)
}

func TestCreateDonation_Success_ToCampaign(t *testing.T) {
	donationRepo := new(mocks.MockDonationRepository)
	userRepo := new(mocks.MockUserRepository)
	campaignRepo := new(mocks.MockCampaignRepository)
	pixClient := new(mocks.MockPixClient)

	userID := uuid.New()
	cID := campID()
	user := &model.User{ID: userID, Name: "Bob", Email: "bob@example.com"}
	campaign := &model.Campaign{ID: *cID, Status: model.CampaignStatusActive}

	campaignRepo.On("FindByID", *cID).Return(campaign, nil)
	userRepo.On("FindByID", userID).Return(user, nil)
	pixClient.On("CreatePix", mock.Anything, mock.Anything).Return(&abacatepay.PixData{ID: "pix-456"}, nil)
	donationRepo.On("Create", mock.AnythingOfType("*model.Donation")).Return(nil)

	uc := usecase.NewDonationUseCase(donationRepo, userRepo, campaignRepo, pixClient)

	donation, err := uc.Create(usecase.CreateDonationInput{
		UserID:     userID,
		Amount:     500,
		CampaignID: cID,
	})

	assert.NoError(t, err)
	assert.Equal(t, cID, donation.CampaignID)
	assert.Nil(t, donation.InstitutionID)
	assert.Equal(t, "pix-456", donation.PixID)
}

func TestCreateDonation_InvalidTarget_BothNil(t *testing.T) {
	uc := usecase.NewDonationUseCase(nil, nil, nil, nil)

	_, err := uc.Create(usecase.CreateDonationInput{
		UserID: uuid.New(),
		Amount: 1000,
		// neither InstitutionID nor CampaignID set
	})

	assert.ErrorIs(t, err, model.ErrInvalidDonationTarget)
}

func TestCreateDonation_InvalidTarget_BothSet(t *testing.T) {
	uc := usecase.NewDonationUseCase(nil, nil, nil, nil)

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

	uc := usecase.NewDonationUseCase(nil, nil, campaignRepo, nil)

	_, err := uc.Create(usecase.CreateDonationInput{
		UserID:     uuid.New(),
		Amount:     500,
		CampaignID: cID,
	})

	assert.ErrorIs(t, err, model.ErrCampaignNotActive)
}

func TestCreateDonation_WithoutCPF(t *testing.T) {
	donationRepo := new(mocks.MockDonationRepository)
	userRepo := new(mocks.MockUserRepository)
	campaignRepo := new(mocks.MockCampaignRepository)
	pixClient := new(mocks.MockPixClient)

	userID := uuid.New()
	iID := instID()
	user := &model.User{ID: userID, Name: "Bob", Email: "bob@example.com"}
	userRepo.On("FindByID", userID).Return(user, nil)

	pixClient.On("CreatePix", mock.Anything, mock.MatchedBy(func(req abacatepay.CreatePixRequest) bool {
		return req.Customer == nil
	})).Return(&abacatepay.PixData{ID: "pix-456", BrCode: "brcode"}, nil)

	donationRepo.On("Create", mock.AnythingOfType("*model.Donation")).Return(nil)

	uc := usecase.NewDonationUseCase(donationRepo, userRepo, campaignRepo, pixClient)

	donation, err := uc.Create(usecase.CreateDonationInput{
		UserID:        userID,
		Amount:        500,
		InstitutionID: iID,
	})

	assert.NoError(t, err)
	assert.Equal(t, "pix-456", donation.PixID)
}

func TestCreateDonation_UserNotFound(t *testing.T) {
	donationRepo := new(mocks.MockDonationRepository)
	userRepo := new(mocks.MockUserRepository)
	campaignRepo := new(mocks.MockCampaignRepository)

	userID := uuid.New()
	iID := instID()
	userRepo.On("FindByID", userID).Return(nil, model.ErrNotFound)

	uc := usecase.NewDonationUseCase(donationRepo, userRepo, campaignRepo, nil)

	_, err := uc.Create(usecase.CreateDonationInput{
		UserID:        userID,
		Amount:        1000,
		InstitutionID: iID,
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}

func TestCreateDonation_PixClientError(t *testing.T) {
	donationRepo := new(mocks.MockDonationRepository)
	userRepo := new(mocks.MockUserRepository)
	campaignRepo := new(mocks.MockCampaignRepository)
	pixClient := new(mocks.MockPixClient)

	userID := uuid.New()
	iID := instID()
	userRepo.On("FindByID", userID).Return(&model.User{ID: userID}, nil)
	pixClient.On("CreatePix", mock.Anything, mock.Anything).Return(nil, errors.New("payment gateway down"))

	uc := usecase.NewDonationUseCase(donationRepo, userRepo, campaignRepo, pixClient)

	_, err := uc.Create(usecase.CreateDonationInput{
		UserID:        userID,
		Amount:        1000,
		InstitutionID: iID,
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "create pix")
}

func TestCreateDonation_RepoCreateError(t *testing.T) {
	donationRepo := new(mocks.MockDonationRepository)
	userRepo := new(mocks.MockUserRepository)
	campaignRepo := new(mocks.MockCampaignRepository)
	pixClient := new(mocks.MockPixClient)

	userID := uuid.New()
	iID := instID()
	userRepo.On("FindByID", userID).Return(&model.User{ID: userID}, nil)
	pixClient.On("CreatePix", mock.Anything, mock.Anything).Return(&abacatepay.PixData{ID: "pix-x"}, nil)
	donationRepo.On("Create", mock.Anything).Return(errors.New("db write error"))

	uc := usecase.NewDonationUseCase(donationRepo, userRepo, campaignRepo, pixClient)

	_, err := uc.Create(usecase.CreateDonationInput{
		UserID:        userID,
		Amount:        1000,
		InstitutionID: iID,
	})

	assert.EqualError(t, err, "db write error")
}

// --------------- GetByID ---------------

func TestGetDonationByID_Success(t *testing.T) {
	donationRepo := new(mocks.MockDonationRepository)
	id := uuid.New()
	expected := &model.Donation{ID: id, Amount: 500}
	donationRepo.On("FindByID", id).Return(expected, nil)

	uc := usecase.NewDonationUseCase(donationRepo, nil, nil, nil)

	donation, err := uc.GetByID(id)
	assert.NoError(t, err)
	assert.Equal(t, expected, donation)
}

func TestGetDonationByID_NotFound(t *testing.T) {
	donationRepo := new(mocks.MockDonationRepository)
	id := uuid.New()
	donationRepo.On("FindByID", id).Return(nil, model.ErrNotFound)

	uc := usecase.NewDonationUseCase(donationRepo, nil, nil, nil)

	_, err := uc.GetByID(id)
	assert.ErrorIs(t, err, model.ErrNotFound)
}

// --------------- GetAll ---------------

func TestGetAllDonations_Success(t *testing.T) {
	donationRepo := new(mocks.MockDonationRepository)
	donations := []*model.Donation{{Amount: 100}, {Amount: 200}}
	donationRepo.On("FindAll").Return(donations, nil)

	uc := usecase.NewDonationUseCase(donationRepo, nil, nil, nil)

	result, err := uc.GetAll()
	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

// --------------- HandleWebhook ---------------

func TestHandleWebhook_Success_NoCampaign(t *testing.T) {
	donationRepo := new(mocks.MockDonationRepository)
	id := uuid.New()
	donationRepo.On("FindByPixID", "pix-abc").Return(&model.Donation{ID: id, PixID: "pix-abc"}, nil)
	donationRepo.On("UpdateStatus", id, "PAID").Return(nil)

	uc := usecase.NewDonationUseCase(donationRepo, nil, nil, nil)

	err := uc.HandleWebhook("pix-abc", "PAID")
	assert.NoError(t, err)
	donationRepo.AssertExpectations(t)
}

func TestHandleWebhook_Success_WithCampaign(t *testing.T) {
	donationRepo := new(mocks.MockDonationRepository)
	campaignRepo := new(mocks.MockCampaignRepository)

	cID := uuid.New()
	id := uuid.New()
	donation := &model.Donation{ID: id, PixID: "pix-camp", Amount: 1000, CampaignID: &cID}

	donationRepo.On("FindByPixID", "pix-camp").Return(donation, nil)
	donationRepo.On("UpdateStatus", id, "PAID").Return(nil)
	campaignRepo.On("IncrementTotalRaised", cID, int64(1000)).Return(nil)

	uc := usecase.NewDonationUseCase(donationRepo, nil, campaignRepo, nil)

	err := uc.HandleWebhook("pix-camp", "PAID")
	assert.NoError(t, err)
	donationRepo.AssertExpectations(t)
	campaignRepo.AssertExpectations(t)
}

func TestHandleWebhook_NoCampaignIncrement_WhenNotPaid(t *testing.T) {
	donationRepo := new(mocks.MockDonationRepository)
	campaignRepo := new(mocks.MockCampaignRepository)

	cID := uuid.New()
	id := uuid.New()
	donation := &model.Donation{ID: id, PixID: "pix-x", Amount: 500, CampaignID: &cID}

	donationRepo.On("FindByPixID", "pix-x").Return(donation, nil)
	donationRepo.On("UpdateStatus", id, "CANCELLED").Return(nil)
	// IncrementTotalRaised must NOT be called for non-PAID events

	uc := usecase.NewDonationUseCase(donationRepo, nil, campaignRepo, nil)

	err := uc.HandleWebhook("pix-x", "CANCELLED")
	assert.NoError(t, err)
	campaignRepo.AssertNotCalled(t, "IncrementTotalRaised")
}

func TestHandleWebhook_DonationNotFound(t *testing.T) {
	donationRepo := new(mocks.MockDonationRepository)
	donationRepo.On("FindByPixID", "pix-unknown").Return(nil, model.ErrNotFound)

	uc := usecase.NewDonationUseCase(donationRepo, nil, nil, nil)

	err := uc.HandleWebhook("pix-unknown", "PAID")
	assert.ErrorIs(t, err, model.ErrNotFound)
}

func TestHandleWebhook_UpdateStatusError(t *testing.T) {
	donationRepo := new(mocks.MockDonationRepository)
	id := uuid.New()
	donationRepo.On("FindByPixID", "pix-abc").Return(&model.Donation{ID: id}, nil)
	donationRepo.On("UpdateStatus", id, "PAID").Return(errors.New("update failed"))

	uc := usecase.NewDonationUseCase(donationRepo, nil, nil, nil)

	err := uc.HandleWebhook("pix-abc", "PAID")
	assert.EqualError(t, err, "update failed")
}
