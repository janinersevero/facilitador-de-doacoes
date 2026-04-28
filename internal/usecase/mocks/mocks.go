package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"facilitador-de-doacoes/internal/model"
	"facilitador-de-doacoes/pkg/abacatepay"
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

func (m *MockDonationRepository) FindByPixID(pixID string) (*model.Donation, error) {
	args := m.Called(pixID)
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

// ---------- FileUploader ----------

type MockFileUploader struct {
	mock.Mock
}

func (m *MockFileUploader) UploadFile(ctx context.Context, fileName string, data []byte, contentType string) (string, error) {
	args := m.Called(ctx, fileName, data, contentType)
	return args.String(0), args.Error(1)
}

// ---------- PixClient ----------

type MockPixClient struct {
	mock.Mock
}

func (m *MockPixClient) CreatePix(ctx context.Context, req abacatepay.CreatePixRequest) (*abacatepay.PixData, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*abacatepay.PixData), args.Error(1)
}
