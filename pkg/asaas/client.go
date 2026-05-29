package asaas

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"time"
)

var nonDigit = regexp.MustCompile(`\D`)

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

const (
	productionBaseURL = "https://api.asaas.com/v3"
	sandboxBaseURL    = "https://sandbox.asaas.com/api/v3"
)

// Client is the ASAAS payment gateway client.
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new ASAAS client. Set sandbox=true for the sandbox environment.
func NewClient(apiKey string, sandbox bool) *Client {
	baseURL := productionBaseURL
	if sandbox {
		baseURL = sandboxBaseURL
	}
	return &Client{
		apiKey:     apiKey,
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) do(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("access_token", c.apiKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.httpClient.Do(req)
}

// ---- Customer ----

type createCustomerRequest struct {
	Name    string `json:"name"`
	CpfCnpj string `json:"cpfCnpj,omitempty"`
	Email   string `json:"email,omitempty"`
	Phone   string `json:"phone,omitempty"`
}

type customerResponse struct {
	ID      string       `json:"id"`
	Name    string       `json:"name"`
	CpfCnpj string       `json:"cpfCnpj"`
	Email   string       `json:"email"`
	Errors  []asaasError `json:"errors"`
}

type customerListResponse struct {
	Data []customerResponse `json:"data"`
}

type asaasError struct {
	Description string `json:"description"`
}

func (c *Client) createCustomer(ctx context.Context, req createCustomerRequest) (*customerResponse, error) {
	resp, err := c.do(ctx, http.MethodPost, "/customers", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result customerResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if len(result.Errors) > 0 {
		return nil, fmt.Errorf("asaas create customer: %s", result.Errors[0].Description)
	}
	return &result, nil
}

func (c *Client) findCustomerByCPF(ctx context.Context, cpf string) (*customerResponse, error) {
	path := "/customers?cpfCnpj=" + url.QueryEscape(cpf)
	resp, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result customerListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if len(result.Data) == 0 {
		return nil, nil
	}
	return &result.Data[0], nil
}

func (c *Client) getOrCreateCustomer(ctx context.Context, data CustomerData) (string, error) {
	cpf := nonDigit.ReplaceAllString(data.CPF, "")

	if cpf != "" {
		existing, err := c.findCustomerByCPF(ctx, cpf)
		if err != nil {
			return "", fmt.Errorf("find customer: %w", err)
		}
		if existing != nil {
			return existing.ID, nil
		}
	}

	created, err := c.createCustomer(ctx, createCustomerRequest{
		Name:    data.Name,
		CpfCnpj: cpf,
		Email:   data.Email,
		Phone:   data.Phone,
	})
	if err != nil {
		return "", fmt.Errorf("create customer: %w", err)
	}
	return created.ID, nil
}

// ---- Payment ----

type createPaymentRequest struct {
	Customer             string                `json:"customer"`
	BillingType          string                `json:"billingType"`
	Value                float64               `json:"value"`
	DueDate              string                `json:"dueDate"`
	Description          string                `json:"description,omitempty"`
	ExternalReference    string                `json:"externalReference,omitempty"`
	CreditCard           *creditCardRequest    `json:"creditCard,omitempty"`
	CreditCardHolderInfo *creditCardHolderInfo `json:"creditCardHolderInfo,omitempty"`
}

type creditCardRequest struct {
	HolderName  string `json:"holderName"`
	Number      string `json:"number"`
	ExpiryMonth string `json:"expiryMonth"`
	ExpiryYear  string `json:"expiryYear"`
	Ccv         string `json:"ccv"`
}

type creditCardHolderInfo struct {
	Name          string `json:"name"`
	Email         string `json:"email"`
	CpfCnpj       string `json:"cpfCnpj"`
	PostalCode    string `json:"postalCode"`
	AddressNumber string `json:"addressNumber"`
	Phone         string `json:"phone"`
}

type paymentAPIResponse struct {
	ID          string       `json:"id"`
	Status      string       `json:"status"`
	BillingType string       `json:"billingType"`
	Value       float64      `json:"value"`
	Errors      []asaasError `json:"errors"`
}

type pixQRCodeAPIResponse struct {
	EncodedImage   string       `json:"encodedImage"`
	Payload        string       `json:"payload"`
	ExpirationDate string       `json:"expirationDate"`
	Errors         []asaasError `json:"errors"`
}

func (c *Client) createPayment(ctx context.Context, req createPaymentRequest) (*paymentAPIResponse, error) {
	resp, err := c.do(ctx, http.MethodPost, "/payments", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result paymentAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if len(result.Errors) > 0 {
		return nil, fmt.Errorf("asaas create payment: %s", result.Errors[0].Description)
	}
	return &result, nil
}

func (c *Client) getPixQRCode(ctx context.Context, paymentID string) (*pixQRCodeAPIResponse, error) {
	resp, err := c.do(ctx, http.MethodGet, "/payments/"+paymentID+"/pixQrCode", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result pixQRCodeAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if len(result.Errors) > 0 {
		return nil, fmt.Errorf("asaas get pix qr code: %s", result.Errors[0].Description)
	}
	return &result, nil
}

// CreatePixPayment creates a PIX payment, handling customer lookup/creation and QR code retrieval.
func (c *Client) CreatePixPayment(ctx context.Context, req PixPaymentRequest) (*PixPaymentResult, error) {
	var customerID string
	if req.Customer != nil {
		id, err := c.getOrCreateCustomer(ctx, *req.Customer)
		if err != nil {
			return nil, err
		}
		customerID = id
	}

	dueDate := time.Now().AddDate(0, 0, 7).Format("2006-01-02")
	payment, err := c.createPayment(ctx, createPaymentRequest{
		Customer:          customerID,
		BillingType:       "PIX",
		Value:             float64(req.Amount) / 100.0,
		DueDate:           dueDate,
		Description:       req.Description,
		ExternalReference: req.ExternalRef,
	})
	if err != nil {
		return nil, err
	}

	qr, err := c.getPixQRCode(ctx, payment.ID)
	if err != nil {
		return nil, err
	}

	return &PixPaymentResult{
		PaymentID: payment.ID,
		BrCode:    qr.Payload,
		QRCodeURL: qr.EncodedImage,
		Status:    payment.Status,
	}, nil
}

// CreateCreditCardPayment creates a credit card payment through ASAAS.
func (c *Client) CreateCreditCardPayment(ctx context.Context, req CreditCardPaymentRequest) (*CreditCardPaymentResult, error) {
	customerID, err := c.getOrCreateCustomer(ctx, req.Customer)
	if err != nil {
		return nil, err
	}

	dueDate := time.Now().Format("2006-01-02")
	payment, err := c.createPayment(ctx, createPaymentRequest{
		Customer:          customerID,
		BillingType:       "CREDIT_CARD",
		Value:             float64(req.Amount) / 100.0,
		DueDate:           dueDate,
		Description:       req.Description,
		ExternalReference: req.ExternalRef,
		CreditCard: &creditCardRequest{
			HolderName:  req.Card.HolderName,
			Number:      req.Card.Number,
			ExpiryMonth: req.Card.ExpiryMonth,
			ExpiryYear:  req.Card.ExpiryYear,
			Ccv:         req.Card.CCV,
		},
		CreditCardHolderInfo: &creditCardHolderInfo{
			Name:          req.Customer.Name,
			Email:         req.Customer.Email,
			CpfCnpj:       nonDigit.ReplaceAllString(req.Customer.CPF, ""),
			Phone:         firstNonEmpty(req.Customer.Phone, req.Card.Phone),
			PostalCode:    nonDigit.ReplaceAllString(req.Card.PostalCode, ""),
			AddressNumber: req.Card.AddressNumber,
		},
	})
	if err != nil {
		return nil, err
	}

	return &CreditCardPaymentResult{
		PaymentID: payment.ID,
		Status:    payment.Status,
	}, nil
}

// ---- Transfer ----

type createTransferRequest struct {
	Value             float64      `json:"value"`
	BankAccount       *bankAccount `json:"bankAccount,omitempty"`
	PixAddressKey     string       `json:"pixAddressKey,omitempty"`
	PixAddressKeyType string       `json:"pixAddressKeyType,omitempty"`
	Description       string       `json:"description,omitempty"`
}

type bankAccount struct {
	Bank            bankCode `json:"bank"`
	AccountName     string   `json:"accountName"`
	OwnerName       string   `json:"ownerName"`
	CpfCnpj         string   `json:"cpfCnpj"`
	Agency          string   `json:"agency"`
	Account         string   `json:"account"`
	AccountDigit    string   `json:"accountDigit"`
	BankAccountType string   `json:"bankAccountType"`
}

type bankCode struct {
	Code string `json:"code"`
}

type transferAPIResponse struct {
	ID     string       `json:"id"`
	Value  float64      `json:"value"`
	Status string       `json:"status"`
	Type   string       `json:"type"`
	Errors []asaasError `json:"errors"`
}

// Transfer creates a bank or PIX transfer using ASAAS.
func (c *Client) Transfer(ctx context.Context, req TransferRequest) (*TransferResult, error) {
	payload := createTransferRequest{
		Value:       float64(req.Amount) / 100.0,
		Description: req.Description,
	}

	if req.PixKey != "" {
		payload.PixAddressKey = req.PixKey
		payload.PixAddressKeyType = req.PixKeyType
	} else {
		payload.BankAccount = &bankAccount{
			Bank:            bankCode{Code: req.BankCode},
			AccountName:     req.AccountName,
			OwnerName:       req.OwnerName,
			CpfCnpj:         req.CPFCnpj,
			Agency:          req.Agency,
			Account:         req.Account,
			AccountDigit:    req.AccountDigit,
			BankAccountType: req.AccountType,
		}
	}

	resp, err := c.do(ctx, http.MethodPost, "/transfers", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result transferAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if len(result.Errors) > 0 {
		return nil, fmt.Errorf("asaas transfer: %s", result.Errors[0].Description)
	}

	return &TransferResult{
		ID:     result.ID,
		Status: result.Status,
	}, nil
}

// ValidateWebhookToken validates the ASAAS webhook token sent in the asaas-access-token header.
func ValidateWebhookToken(token, expectedToken string) bool {
	return token != "" && token == expectedToken
}
