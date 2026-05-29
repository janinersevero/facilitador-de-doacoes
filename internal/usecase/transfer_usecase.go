package usecase

import "context"

// CreateTransferInput is the input for initiating a transfer via ASAAS.
// Provide either (PixKey + PixKeyType) for PIX transfers
// or bank account fields (BankCode, Agency, Account, etc.) for TED transfers.
type CreateTransferInput struct {
	Amount      int    `json:"amount"       binding:"required,min=100"`
	Description string `json:"description"`
	// PIX transfer
	PixKey     string `json:"pix_key"`
	PixKeyType string `json:"pix_key_type"` // CPF, EMAIL, PHONE, EVP
	// Bank (TED) transfer
	BankCode     string `json:"bank_code"`
	AccountName  string `json:"account_name"`
	OwnerName    string `json:"owner_name"`
	CPFCnpj      string `json:"cpf_cnpj"`
	Agency       string `json:"agency"`
	Account      string `json:"account"`
	AccountDigit string `json:"account_digit"`
	AccountType  string `json:"account_type"` // CONTA_CORRENTE or CONTA_POUPANCA
}

// TransferOutput holds the result of a transfer operation.
type TransferOutput struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type TransferUseCase interface {
	Create(ctx context.Context, input CreateTransferInput) (*TransferOutput, error)
}
