package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Donation struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null"   json:"user_id"`
	PixID     string    `gorm:"uniqueIndex"          json:"pix_id"`
	BrCode    string    `json:"br_code"`
	QRCodeURL string    `json:"qr_code_url"`
	Amount    int       `gorm:"not null"             json:"amount"`
	Status    string    `gorm:"default:'PENDING'"    json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (d *Donation) BeforeCreate(_ *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}
