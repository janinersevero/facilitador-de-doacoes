package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type InstitutionStatus string

const (
	InstitutionStatusPending  InstitutionStatus = "pending"
	InstitutionStatusApproved InstitutionStatus = "approved"
	InstitutionStatusRejected InstitutionStatus = "rejected"
)

type Institution struct {
	ID              uuid.UUID         `gorm:"type:uuid;primaryKey"        json:"id"`
	UserID          uuid.UUID         `gorm:"type:uuid;not null;index"    json:"user_id"`
	Name            string            `gorm:"not null"                    json:"name"`
	LegalName       string            `gorm:"not null"                    json:"legal_name"`
	CNPJ            string            `gorm:"uniqueIndex;not null"        json:"cnpj"`
	Description     string            `                                   json:"description,omitempty"`
	Email           string            `                                   json:"email,omitempty"`
	Phone           string            `                                   json:"phone,omitempty"`
	Address         string            `                                   json:"address,omitempty"`
	ZipCode         string            `                                   json:"zip_code,omitempty"`
	Category        string            `                                   json:"category,omitempty"`
	LogoURL         string            `                                   json:"logo_url,omitempty"`
	CoverImageURL   string            `                                   json:"cover_image_url,omitempty"`
	VideoURL        string            `                                   json:"video_url,omitempty"`
	WebsiteURL      string            `                                   json:"website_url,omitempty"`
	Status          InstitutionStatus `gorm:"default:'pending'"           json:"status"`
	RejectionReason string            `                                   json:"rejection_reason,omitempty"`
	ApprovedAt      *time.Time        `                                   json:"approved_at,omitempty"`
	CreatedAt       time.Time         `                                   json:"created_at"`
	UpdatedAt       time.Time         `                                   json:"updated_at"`
}

func (i *Institution) BeforeCreate(_ *gorm.DB) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	return nil
}
