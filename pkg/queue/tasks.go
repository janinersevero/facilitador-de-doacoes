package queue

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"
)

// Task type names — used to register handlers in the worker
const (
	TypeDonationPaid = "notification:donation_paid"
)

// DonationPaidPayload holds the minimal data enqueued when a donation is confirmed.
// The worker fetches the full data from the database using these IDs.
type DonationPaidPayload struct {
	DonationID    string
	UserID        string
	Amount        int
	InstitutionID string
	CampaignID    string // empty when the donation is made directly to an institution
}

// NewDonationPaidTask creates an asynq task to notify the chatbot.
// MaxRetry: 3 attempts with a 15-minute interval between each.
func NewDonationPaidTask(payload DonationPaidPayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(
		TypeDonationPaid,
		data,
		asynq.MaxRetry(3),
		asynq.Retention(72*60*60), // retain in DLQ for 72h
	), nil
}

// EnqueueDonationPaid enqueues a notification task in the default queue.
func EnqueueDonationPaid(client *asynq.Client, ctx context.Context, payload DonationPaidPayload) error {
	task, err := NewDonationPaidTask(payload)
	if err != nil {
		return err
	}
	_, err = client.EnqueueContext(ctx, task)
	return err
}
