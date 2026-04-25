package abacatepay

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const baseURL = "https://api.abacatepay.com/v2"

type Client struct {
	apiKey     string
	httpClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

type Customer struct {
	Name      string `json:"name"`
	Cellphone string `json:"cellphone,omitempty"`
	Email     string `json:"email"`
	TaxID     string `json:"taxId,omitempty"`
}

type pixRequestData struct {
	Amount      int       `json:"amount"`
	Description string    `json:"description,omitempty"`
	ExternalID  string    `json:"externalId,omitempty"`
	Customer    *Customer `json:"customer,omitempty"`
}

type createPixBody struct {
	Method string         `json:"method"`
	Data   pixRequestData `json:"data"`
}

// CreatePixRequest cria um PIX transparente para qualquer valor.
type CreatePixRequest struct {
	Amount      int
	Description string
	ExternalID  string
	Customer    *Customer
}

type PixData struct {
	ID        string `json:"id"`
	BrCode    string `json:"brCode"`
	QRCodeURL string `json:"qrCodeUrl"`
	Status    string `json:"status"`
}

type pixResponse struct {
	Data  *PixData `json:"data"`
	Error *string  `json:"error"`
}

func (c *Client) CreatePix(ctx context.Context, req CreatePixRequest) (*PixData, error) {
	payload := createPixBody{
		Method: "PIX",
		Data: pixRequestData{
			Amount:      req.Amount,
			Description: req.Description,
			ExternalID:  req.ExternalID,
			Customer:    req.Customer,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/transparents/create", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result pixResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Error != nil && *result.Error != "" {
		return nil, fmt.Errorf("abacatepay: %s", *result.Error)
	}

	return result.Data, nil
}

// WebhookEvent é o payload enviado pelo Abacate Pay nos eventos de pagamento.
type WebhookEvent struct {
	EventType string      `json:"eventType"`
	DevMode   bool        `json:"devMode"`
	Data      WebhookData `json:"data"`
}

type WebhookData struct {
	Billing WebhookBilling `json:"billing"`
}

type WebhookBilling struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Method string `json:"method"`
}

// ValidateSignature verifica a assinatura HMAC-SHA256 do header X-Webhook-Signature.
func ValidateSignature(secret, signature string, body []byte) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}
