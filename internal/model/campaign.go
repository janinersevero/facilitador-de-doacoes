package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type CampaignStatus string

const (
	CampaignStatusActive    CampaignStatus = "active"
	CampaignStatusPaused    CampaignStatus = "paused"
	CampaignStatusCompleted CampaignStatus = "completed"
	CampaignStatusCancelled CampaignStatus = "cancelled"
)

type Campaign struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey"        json:"id"`
	InstitutionID uuid.UUID      `gorm:"type:uuid;not null;index"    json:"institution_id"`
	Title         string         `gorm:"not null"                    json:"title"`
	Description   string         `gorm:"not null"                    json:"description"`
	GoalAmount    int64          `gorm:"not null"                    json:"goal_amount"`
	TotalRaised   int64          `gorm:"default:0"                   json:"total_raised"`
	IsUrgent      bool           `gorm:"default:false"               json:"is_urgent"`
	Keywords      pq.StringArray `gorm:"type:text[]"                 json:"keywords"`
	Status        CampaignStatus `gorm:"default:'active'"            json:"status"`
	CoverImageURL string         `                                   json:"cover_image_url,omitempty"`
	EndsAt        *time.Time     `                                   json:"ends_at,omitempty"`
	CreatedAt     time.Time      `                                   json:"created_at"`
	UpdatedAt     time.Time      `                                   json:"updated_at"`
}

func (c *Campaign) BeforeCreate(_ *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}
