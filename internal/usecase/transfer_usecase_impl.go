package usecase

import (
	"context"
	"fmt"

	"facilitador-de-doacoes/pkg/asaas"
)

// transferGateway abstracts the ASAAS transfer operation for testing.
type transferGateway interface {
	Transfer(ctx context.Context, req asaas.TransferRequest) (*asaas.TransferResult, error)
}

type transferUseCase struct {
	gateway transferGateway
}

func NewTransferUseCase(gateway transferGateway) TransferUseCase {
	return &transferUseCase{gateway: gateway}
}

func (uc *transferUseCase) Create(ctx context.Context, input CreateTransferInput) (*TransferOutput, error) {
	if input.PixKey == "" && input.BankCode == "" {
		return nil, fmt.Errorf("either pix_key or bank account details (bank_code) are required")
	}

	result, err := uc.gateway.Transfer(ctx, asaas.TransferRequest{
		Amount:       input.Amount,
		Description:  input.Description,
		PixKey:       input.PixKey,
		PixKeyType:   input.PixKeyType,
		BankCode:     input.BankCode,
		AccountName:  input.AccountName,
		OwnerName:    input.OwnerName,
		CPFCnpj:      input.CPFCnpj,
		Agency:       input.Agency,
		Account:      input.Account,
		AccountDigit: input.AccountDigit,
		AccountType:  input.AccountType,
	})
	if err != nil {
		return nil, fmt.Errorf("transfer: %w", err)
	}

	return &TransferOutput{
		ID:     result.ID,
		Status: result.Status,
	}, nil
}
