package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"facilitador-de-doacoes/internal/usecase"
	"facilitador-de-doacoes/internal/usecase/mocks"
	"facilitador-de-doacoes/pkg/asaas"
)

// --------------- Create ---------------

func TestCreateTransfer_Success_PixKey(t *testing.T) {
	gw := new(mocks.MockTransferGateway)
	gw.On("Transfer", mock.Anything, mock.MatchedBy(func(req asaas.TransferRequest) bool {
		return req.Amount == 5000 && req.PixKey == "chave@pix.com" && req.PixKeyType == "EMAIL"
	})).Return(&asaas.TransferResult{ID: "tra-1", Status: "PENDING"}, nil)

	uc := usecase.NewTransferUseCase(gw)
	result, err := uc.Create(context.Background(), usecase.CreateTransferInput{
		Amount:     5000,
		PixKey:     "chave@pix.com",
		PixKeyType: "EMAIL",
	})

	assert.NoError(t, err)
	assert.Equal(t, "tra-1", result.ID)
	assert.Equal(t, "PENDING", result.Status)
	gw.AssertExpectations(t)
}

func TestCreateTransfer_Success_BankAccount(t *testing.T) {
	gw := new(mocks.MockTransferGateway)
	gw.On("Transfer", mock.Anything, mock.MatchedBy(func(req asaas.TransferRequest) bool {
		return req.Amount == 10000 && req.BankCode == "341"
	})).Return(&asaas.TransferResult{ID: "tra-2", Status: "PENDING"}, nil)

	uc := usecase.NewTransferUseCase(gw)
	result, err := uc.Create(context.Background(), usecase.CreateTransferInput{
		Amount:      10000,
		BankCode:    "341",
		AccountName: "Conta Teste",
		OwnerName:   "Test Owner",
	})

	assert.NoError(t, err)
	assert.Equal(t, "tra-2", result.ID)
}

func TestCreateTransfer_MissingDestination(t *testing.T) {
	uc := usecase.NewTransferUseCase(nil)
	_, err := uc.Create(context.Background(), usecase.CreateTransferInput{Amount: 1000})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "pix_key or bank account")
}

func TestCreateTransfer_GatewayError(t *testing.T) {
	gw := new(mocks.MockTransferGateway)
	gw.On("Transfer", mock.Anything, mock.Anything).Return(nil, errors.New("gateway unavailable"))

	uc := usecase.NewTransferUseCase(gw)
	_, err := uc.Create(context.Background(), usecase.CreateTransferInput{
		Amount: 1000,
		PixKey: "any@pix.com",
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "transfer: gateway unavailable")
}
