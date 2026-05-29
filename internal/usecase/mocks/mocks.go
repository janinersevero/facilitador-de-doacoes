package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"facilitador-de-doacoes/internal/model"
	"facilitador-de-doacoes/internal/repository"
	"facilitador-de-doacoes/pkg/asaas"
)

// ---------- UserRepository ----------

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *model.User) error {
	return m.Called(user).Error(0)
}

func (m *MockUserRepository) FindByID(id uuid.UUID) (*model.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) FindByEmail(email string) (*model.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) FindByAuth0ID(auth0ID string) (*model.User, error) {
	args := m.Called(auth0ID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) FindAll() ([]*model.User, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.User), args.Error(1)
}

func (m *MockUserRepository) Update(user *model.User) error {
	return m.Called(user).Error(0)
}

func (m *MockUserRepository) UpdateAvatarURL(id uuid.UUID, url string) error {
	return m.Called(id, url).Error(0)
}

func (m *MockUserRepository) Delete(id uuid.UUID) error {
	return m.Called(id).Error(0)
}

// ---------- DonationRepository ----------

type MockDonationRepository struct {
	mock.Mock
}

func (m *MockDonationRepository) Create(donation *model.Donation) error {
	return m.Called(donation).Error(0)
}

func (m *MockDonationRepository) FindByID(id uuid.UUID) (*model.Donation, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Donation), args.Error(1)
}

func (m *MockDonationRepository) FindByPaymentID(paymentID string) (*model.Donation, error) {
	args := m.Called(paymentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Donation), args.Error(1)
}

func (m *MockDonationRepository) FindAll() ([]*model.Donation, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Donation), args.Error(1)
}

func (m *MockDonationRepository) UpdateStatus(id uuid.UUID, status string) error {
	return m.Called(id, status).Error(0)
}

// ---------- CampaignRepository ----------

type MockCampaignRepository struct {
	mock.Mock
}

func (m *MockCampaignRepository) Create(campaign *model.Campaign) error {
	return m.Called(campaign).Error(0)
}

func (m *MockCampaignRepository) FindByID(id uuid.UUID) (*model.Campaign, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Campaign), args.Error(1)
}

func (m *MockCampaignRepository) FindAll(filters repository.CampaignFilters) ([]*model.Campaign, error) {
	args := m.Called(filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Campaign), args.Error(1)
}

func (m *MockCampaignRepository) FindByInstitutionID(institutionID uuid.UUID) ([]*model.Campaign, error) {
	args := m.Called(institutionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Campaign), args.Error(1)
}

func (m *MockCampaignRepository) Update(campaign *model.Campaign) error {
	return m.Called(campaign).Error(0)
}

func (m *MockCampaignRepository) Delete(id uuid.UUID) error {
	return m.Called(id).Error(0)
}

func (m *MockCampaignRepository) IncrementTotalRaised(id uuid.UUID, amount int64) error {
	return m.Called(id, amount).Error(0)
}

// ---------- NecessityRepository ----------

type MockNecessityRepository struct {
	mock.Mock
}

func (m *MockNecessityRepository) Create(necessity *model.Necessity) error {
	return m.Called(necessity).Error(0)
}

func (m *MockNecessityRepository) FindByID(id uuid.UUID) (*model.Necessity, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Necessity), args.Error(1)
}

func (m *MockNecessityRepository) FindByInstitutionID(institutionID uuid.UUID) ([]*model.Necessity, error) {
	args := m.Called(institutionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Necessity), args.Error(1)
}

func (m *MockNecessityRepository) Update(necessity *model.Necessity) error {
	return m.Called(necessity).Error(0)
}

func (m *MockNecessityRepository) Delete(id uuid.UUID) error {
	return m.Called(id).Error(0)
}

// ---------- FileUploader ----------

type MockFileUploader struct {
	mock.Mock
}

func (m *MockFileUploader) UploadFile(ctx context.Context, fileName string, data []byte, contentType string) (string, error) {
	args := m.Called(ctx, fileName, data, contentType)
	return args.String(0), args.Error(1)
}

// ---------- RoleSetter ----------

type MockRoleSetter struct {
	mock.Mock
}

func (m *MockRoleSetter) SetUserRole(ctx context.Context, auth0UserID, role string) error {
	return m.Called(ctx, auth0UserID, role).Error(0)
}

// ---------- PaymentClient ----------

type MockPaymentClient struct {
	mock.Mock
}

func (m *MockPaymentClient) CreatePixPayment(ctx context.Context, req asaas.PixPaymentRequest) (*asaas.PixPaymentResult, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*asaas.PixPaymentResult), args.Error(1)
}

func (m *MockPaymentClient) CreateCreditCardPayment(ctx context.Context, req asaas.CreditCardPaymentRequest) (*asaas.CreditCardPaymentResult, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*asaas.CreditCardPaymentResult), args.Error(1)
}

// ---------- TransferGateway ----------

type MockTransferGateway struct {
	mock.Mock
}

func (m *MockTransferGateway) Transfer(ctx context.Context, req asaas.TransferRequest) (*asaas.TransferResult, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*asaas.TransferResult), args.Error(1)
}
