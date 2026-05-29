package asaas

import "context"

// CustomerData holds customer information for payment requests.
type CustomerData struct {
	Name  string
	Email string
	Phone string
	CPF   string
}

// CreditCardData holds credit card details for payment.
type CreditCardData struct {
	HolderName    string
	Number        string
	ExpiryMonth   string
	ExpiryYear    string
	CCV           string
	PostalCode    string
	AddressNumber string
	Phone         string // fallback when customer has no phone registered
}

// PixPaymentRequest is the input for creating a PIX payment.
type PixPaymentRequest struct {
	Amount      int // in cents
	Description string
	ExternalRef string
	Customer    *CustomerData
}

// PixPaymentResult is returned after a PIX payment is created.
type PixPaymentResult struct {
	PaymentID string
	BrCode    string // PIX copy-paste code (payload)
	QRCodeURL string // base64-encoded QR image from ASAAS
	Status    string
}

// CreditCardPaymentRequest is the input for creating a credit card payment.
type CreditCardPaymentRequest struct {
	Amount      int // in cents
	Description string
	ExternalRef string
	Customer    CustomerData
	Card        CreditCardData
}

// CreditCardPaymentResult is returned after a credit card payment is created.
type CreditCardPaymentResult struct {
	PaymentID string
	Status    string
}

// PaymentClient abstracts PIX and credit card payment operations.
type PaymentClient interface {
	CreatePixPayment(ctx context.Context, req PixPaymentRequest) (*PixPaymentResult, error)
	CreateCreditCardPayment(ctx context.Context, req CreditCardPaymentRequest) (*CreditCardPaymentResult, error)
}

// TransferRequest is the input for creating a transfer.
type TransferRequest struct {
	Amount      int // in cents
	Description string
	// PIX transfer fields
	PixKey     string
	PixKeyType string // CPF, EMAIL, PHONE, EVP
	// Bank transfer (TED) fields
	BankCode     string
	AccountName  string
	OwnerName    string
	CPFCnpj      string
	Agency       string
	Account      string
	AccountDigit string
	AccountType  string // CONTA_CORRENTE or CONTA_POUPANCA
}

// TransferResult is returned after a transfer is initiated.
type TransferResult struct {
	ID     string
	Status string
}

// WebhookEvent is the ASAAS webhook event payload.
type WebhookEvent struct {
	Event   string         `json:"event"`
	Payment WebhookPayment `json:"payment"`
}

// WebhookPayment holds the payment data inside a webhook event.
type WebhookPayment struct {
	ID          string  `json:"id"`
	Status      string  `json:"status"`
	BillingType string  `json:"billingType"`
	Value       float64 `json:"value"`
}
