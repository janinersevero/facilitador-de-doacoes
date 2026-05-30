package queue

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/hibiken/asynq"

	"facilitador-de-doacoes/internal/repository"
)

// NotificationPayload is the request body sent to the chatbot /notify endpoint.
type NotificationPayload struct {
	UserID          string `json:"user_id"`
	UserName        string `json:"user_name"`
	UserPhone       string `json:"user_phone"`
	Amount          int    `json:"amount"`
	InstitutionName string `json:"institution_name"`
	CampaignName    string `json:"campaign_name,omitempty"`
	DonationID      string `json:"donation_id"`
}

// NotificationHandler processes tasks of type TypeDonationPaid.
type NotificationHandler struct {
	userRepo        repository.UserRepository
	institutionRepo repository.InstitutionRepository
	campaignRepo    repository.CampaignRepository
	notifyURL       string
	httpClient      *http.Client
}

func NewNotificationHandler(
	userRepo repository.UserRepository,
	institutionRepo repository.InstitutionRepository,
	campaignRepo repository.CampaignRepository,
	notifyURL string,
) *NotificationHandler {
	return &NotificationHandler{
		userRepo:        userRepo,
		institutionRepo: institutionRepo,
		campaignRepo:    campaignRepo,
		notifyURL:       notifyURL,
		httpClient:      &http.Client{Timeout: 10 * time.Second},
	}
}

func (h *NotificationHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var p DonationPaidPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	// Fetch full data from the database
	userID, err := parseUUID(p.UserID)
	if err != nil {
		return fmt.Errorf("invalid user_id: %w", err)
	}
	user, err := h.userRepo.FindByID(userID)
	if err != nil {
		return fmt.Errorf("find user: %w", err)
	}

	institutionID, err := parseUUID(p.InstitutionID)
	if err != nil {
		return fmt.Errorf("invalid institution_id: %w", err)
	}
	institution, err := h.institutionRepo.FindByID(institutionID)
	if err != nil {
		return fmt.Errorf("find institution: %w", err)
	}

	notification := NotificationPayload{
		UserID:          p.UserID,
		UserName:        user.Name,
		UserPhone:       user.Phone,
		Amount:          p.Amount,
		InstitutionName: institution.Name,
		DonationID:      p.DonationID,
	}

	// Fetch campaign name if present
	if p.CampaignID != "" {
		campaignID, err := parseUUID(p.CampaignID)
		if err == nil {
			campaign, err := h.campaignRepo.FindByID(campaignID)
			if err == nil {
				notification.CampaignName = campaign.Title
			}
		}
	}

	if err := h.sendNotification(ctx, notification); err != nil {
		return err
	}

	log.Printf("[worker] notificação enviada: donation_id=%s user=%s", p.DonationID, user.Name)
	return nil
}

func (h *NotificationHandler) sendNotification(ctx context.Context, payload NotificationPayload) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.notifyURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("notify request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("notify endpoint returned %d", resp.StatusCode)
	}

	return nil
}
