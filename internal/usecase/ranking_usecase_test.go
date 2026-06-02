package usecase_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"facilitador-de-doacoes/internal/model"
	"facilitador-de-doacoes/internal/usecase"
	"facilitador-de-doacoes/internal/usecase/mocks"
)

func TestGetRanking_Success(t *testing.T) {
	donationRepo := new(mocks.MockDonationRepository)

	expected := []*model.RankingEntry{
		{Position: 1, UserID: uuid.New(), UserName: "Carlos", TotalDonated: 420000, DonationCount: 5},
		{Position: 2, UserID: uuid.New(), UserName: "Fernanda", TotalDonated: 380000, DonationCount: 3},
	}

	donationRepo.On("GetRanking", 10).Return(expected, nil)

	uc := usecase.NewRankingUseCase(donationRepo)
	result, err := uc.GetRanking(10)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	donationRepo.AssertExpectations(t)
}

func TestGetRanking_DefaultLimit_WhenZero(t *testing.T) {
	donationRepo := new(mocks.MockDonationRepository)

	donationRepo.On("GetRanking", 10).Return([]*model.RankingEntry{}, nil)

	uc := usecase.NewRankingUseCase(donationRepo)
	result, err := uc.GetRanking(0)

	assert.NoError(t, err)
	assert.Empty(t, result)
	donationRepo.AssertExpectations(t)
}

func TestGetRanking_DefaultLimit_WhenNegative(t *testing.T) {
	donationRepo := new(mocks.MockDonationRepository)

	donationRepo.On("GetRanking", 10).Return([]*model.RankingEntry{}, nil)

	uc := usecase.NewRankingUseCase(donationRepo)
	result, err := uc.GetRanking(-5)

	assert.NoError(t, err)
	assert.Empty(t, result)
	donationRepo.AssertExpectations(t)
}

func TestGetRanking_DefaultLimit_WhenOverMax(t *testing.T) {
	donationRepo := new(mocks.MockDonationRepository)

	donationRepo.On("GetRanking", 10).Return([]*model.RankingEntry{}, nil)

	uc := usecase.NewRankingUseCase(donationRepo)
	result, err := uc.GetRanking(200)

	assert.NoError(t, err)
	assert.Empty(t, result)
	donationRepo.AssertExpectations(t)
}

func TestGetRanking_RepositoryError(t *testing.T) {
	donationRepo := new(mocks.MockDonationRepository)

	donationRepo.On("GetRanking", 10).Return(nil, errors.New("db connection failed"))

	uc := usecase.NewRankingUseCase(donationRepo)
	result, err := uc.GetRanking(10)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "db connection failed", err.Error())
	donationRepo.AssertExpectations(t)
}

func TestGetRanking_EmptyResult(t *testing.T) {
	donationRepo := new(mocks.MockDonationRepository)

	donationRepo.On("GetRanking", 5).Return([]*model.RankingEntry{}, nil)

	uc := usecase.NewRankingUseCase(donationRepo)
	result, err := uc.GetRanking(5)

	assert.NoError(t, err)
	assert.Empty(t, result)
	donationRepo.AssertExpectations(t)
}
