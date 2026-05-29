package asaas_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"facilitador-de-doacoes/pkg/asaas"
)

// newTestClient creates a client pointed at a test server with the given API key.
func newTestClient(t *testing.T, server *httptest.Server) *asaas.Client {
	t.Helper()
	c := asaas.NewClient("test-api-key", false)
	asaas.SetBaseURL(c, server.URL)
	return c
}

func respondJSON(w http.ResponseWriter, status int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}

// --------------- CreatePixPayment ---------------

func TestCreatePixPayment_Success_WithNewCustomer(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/customers", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			respondJSON(w, 200, map[string]interface{}{"data": []interface{}{}})
			return
		}
		respondJSON(w, 200, map[string]interface{}{"id": "cus-new", "name": "Alice"})
	})

	mux.HandleFunc("/payments", func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, 200, map[string]interface{}{
			"id": "pay-pix-1", "status": "PENDING", "billingType": "PIX",
		})
	})

	mux.HandleFunc("/payments/pay-pix-1/pixQrCode", func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, 200, map[string]interface{}{
			"encodedImage": "base64img==",
			"payload":      "00020126...",
		})
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()
	client := newTestClient(t, srv)

	result, err := client.CreatePixPayment(context.Background(), asaas.PixPaymentRequest{
		Amount:      1000,
		Description: "Doação",
		ExternalRef: "ext-123",
		Customer: &asaas.CustomerData{
			Name:  "Alice",
			Email: "alice@example.com",
			CPF:   "12345678900",
		},
	})

	require.NoError(t, err)
	assert.Equal(t, "pay-pix-1", result.PaymentID)
	assert.Equal(t, "00020126...", result.BrCode)
	assert.Equal(t, "base64img==", result.QRCodeURL)
	assert.Equal(t, "PENDING", result.Status)
}

func TestCreatePixPayment_Success_WithExistingCustomer(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/customers", func(w http.ResponseWriter, r *http.Request) {
		// GET returns existing customer
		respondJSON(w, 200, map[string]interface{}{
			"data": []interface{}{
				map[string]interface{}{"id": "cus-existing", "name": "Bob"},
			},
		})
	})

	mux.HandleFunc("/payments", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "cus-existing", body["customer"])
		respondJSON(w, 200, map[string]interface{}{"id": "pay-pix-2", "status": "PENDING"})
	})

	mux.HandleFunc("/payments/pay-pix-2/pixQrCode", func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, 200, map[string]interface{}{
			"encodedImage": "img2",
			"payload":      "pix-payload",
		})
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()
	client := newTestClient(t, srv)

	result, err := client.CreatePixPayment(context.Background(), asaas.PixPaymentRequest{
		Amount:   500,
		Customer: &asaas.CustomerData{CPF: "99988877766"},
	})

	require.NoError(t, err)
	assert.Equal(t, "pay-pix-2", result.PaymentID)
}

func TestCreatePixPayment_WithoutCustomer(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/payments", func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, 200, map[string]interface{}{"id": "pay-anon", "status": "PENDING"})
	})

	mux.HandleFunc("/payments/pay-anon/pixQrCode", func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, 200, map[string]interface{}{"payload": "brcode", "encodedImage": "img"})
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()
	client := newTestClient(t, srv)

	result, err := client.CreatePixPayment(context.Background(), asaas.PixPaymentRequest{
		Amount:   200,
		Customer: nil,
	})

	require.NoError(t, err)
	assert.Equal(t, "pay-anon", result.PaymentID)
}

func TestCreatePixPayment_PaymentError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/payments", func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, 400, map[string]interface{}{
			"errors": []interface{}{
				map[string]interface{}{"description": "invalid value"},
			},
		})
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()
	client := newTestClient(t, srv)

	_, err := client.CreatePixPayment(context.Background(), asaas.PixPaymentRequest{Amount: 50})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid value")
}

func TestCreatePixPayment_QRCodeError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/payments", func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, 200, map[string]interface{}{"id": "pay-err", "status": "PENDING"})
	})
	mux.HandleFunc("/payments/pay-err/pixQrCode", func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, 400, map[string]interface{}{
			"errors": []interface{}{
				map[string]interface{}{"description": "payment not found"},
			},
		})
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()
	client := newTestClient(t, srv)

	_, err := client.CreatePixPayment(context.Background(), asaas.PixPaymentRequest{Amount: 100})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "payment not found")
}

// --------------- CreateCreditCardPayment ---------------

func TestCreateCreditCardPayment_Success(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/customers", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			respondJSON(w, 200, map[string]interface{}{"data": []interface{}{}})
			return
		}
		respondJSON(w, 200, map[string]interface{}{"id": "cus-cc", "name": "Carol"})
	})

	mux.HandleFunc("/payments", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "CREDIT_CARD", body["billingType"])
		respondJSON(w, 200, map[string]interface{}{"id": "pay-cc-1", "status": "CONFIRMED"})
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()
	client := newTestClient(t, srv)

	result, err := client.CreateCreditCardPayment(context.Background(), asaas.CreditCardPaymentRequest{
		Amount:      5000,
		Description: "Doação cartão",
		ExternalRef: "ext-cc-1",
		Customer: asaas.CustomerData{
			Name:  "Carol",
			Email: "carol@example.com",
			CPF:   "11122233344",
		},
		Card: asaas.CreditCardData{
			HolderName:    "CAROL SILVA",
			Number:        "4111111111111111",
			ExpiryMonth:   "12",
			ExpiryYear:    "2030",
			CCV:           "123",
			PostalCode:    "01310100",
			AddressNumber: "100",
		},
	})

	require.NoError(t, err)
	assert.Equal(t, "pay-cc-1", result.PaymentID)
	assert.Equal(t, "CONFIRMED", result.Status)
}

func TestCreateCreditCardPayment_Error(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/customers", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			respondJSON(w, 200, map[string]interface{}{"data": []interface{}{}})
			return
		}
		respondJSON(w, 200, map[string]interface{}{"id": "cus-x"})
	})
	mux.HandleFunc("/payments", func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, 400, map[string]interface{}{
			"errors": []interface{}{
				map[string]interface{}{"description": "invalid card number"},
			},
		})
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()
	client := newTestClient(t, srv)

	_, err := client.CreateCreditCardPayment(context.Background(), asaas.CreditCardPaymentRequest{
		Amount:   1000,
		Customer: asaas.CustomerData{Name: "Test", CPF: "00000000000"},
		Card:     asaas.CreditCardData{HolderName: "Test", Number: "bad"},
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid card number")
}

// --------------- Transfer ---------------

func TestTransfer_Success_PixKey(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/transfers", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "chave@pix.com", body["pixAddressKey"])
		assert.Equal(t, "EMAIL", body["pixAddressKeyType"])
		assert.Equal(t, 100.0, body["value"])
		respondJSON(w, 200, map[string]interface{}{"id": "tra-1", "status": "PENDING"})
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()
	client := newTestClient(t, srv)

	result, err := client.Transfer(context.Background(), asaas.TransferRequest{
		Amount:     10000,
		PixKey:     "chave@pix.com",
		PixKeyType: "EMAIL",
	})

	require.NoError(t, err)
	assert.Equal(t, "tra-1", result.ID)
	assert.Equal(t, "PENDING", result.Status)
}

func TestTransfer_Success_BankAccount(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/transfers", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		assert.NotNil(t, body["bankAccount"])
		respondJSON(w, 200, map[string]interface{}{"id": "tra-2", "status": "PENDING"})
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()
	client := newTestClient(t, srv)

	result, err := client.Transfer(context.Background(), asaas.TransferRequest{
		Amount:       5000,
		BankCode:     "341",
		AccountName:  "Conta Principal",
		OwnerName:    "João Silva",
		CPFCnpj:      "12345678900",
		Agency:       "0001",
		Account:      "12345",
		AccountDigit: "6",
		AccountType:  "CONTA_CORRENTE",
	})

	require.NoError(t, err)
	assert.Equal(t, "tra-2", result.ID)
}

func TestTransfer_Error(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/transfers", func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, 400, map[string]interface{}{
			"errors": []interface{}{
				map[string]interface{}{"description": "insufficient balance"},
			},
		})
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()
	client := newTestClient(t, srv)

	_, err := client.Transfer(context.Background(), asaas.TransferRequest{
		Amount: 999999999,
		PixKey: "test@example.com",
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient balance")
}

// --------------- ValidateWebhookToken ---------------

func TestValidateWebhookToken_Valid(t *testing.T) {
	assert.True(t, asaas.ValidateWebhookToken("my-secret-token", "my-secret-token"))
}

func TestValidateWebhookToken_Invalid(t *testing.T) {
	assert.False(t, asaas.ValidateWebhookToken("wrong-token", "my-secret-token"))
}

func TestValidateWebhookToken_EmptyToken(t *testing.T) {
	assert.False(t, asaas.ValidateWebhookToken("", "my-secret-token"))
}
