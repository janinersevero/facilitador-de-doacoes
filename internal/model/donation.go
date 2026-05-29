package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Donation struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey"                     json:"id"`
	UserID        uuid.UUID  `gorm:"type:uuid;not null;index"                 json:"user_id"`
	InstitutionID *uuid.UUID `gorm:"type:uuid;index"                          json:"institution_id,omitempty"`
	CampaignID    *uuid.UUID `gorm:"type:uuid;index"                          json:"campaign_id,omitempty"`
	PaymentMethod string     `gorm:"not null;default:'PIX'"                   json:"payment_method"`
	PaymentID     string     `gorm:"column:payment_id;uniqueIndex"            json:"payment_id"`
	BrCode        string     `                                                json:"br_code,omitempty"`
	QRCodeURL     string     `                                                json:"qr_code_url,omitempty"`
	Amount        int        `gorm:"not null"                                 json:"amount"`
	Status        string     `gorm:"default:'PENDING'"                        json:"status"`
	CreatedAt     time.Time  `                                                json:"created_at"`
	UpdatedAt     time.Time  `                                                json:"updated_at"`
}

func (d *Donation) BeforeCreate(_ *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}
