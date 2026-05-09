package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NecessityStatus string

const (
	NecessityStatusOpen     NecessityStatus = "open"
	NecessityStatusAttended NecessityStatus = "attended"
)

type Necessity struct {
	ID            uuid.UUID       `gorm:"type:uuid;primaryKey"        json:"id"`
	InstitutionID uuid.UUID       `gorm:"type:uuid;not null;index"    json:"institution_id"`
	Description   string          `gorm:"not null"                    json:"description"`
	Category      string          `gorm:"not null"                    json:"category"`
	IsUrgent      bool            `gorm:"default:false"               json:"is_urgent"`
	Status        NecessityStatus `gorm:"default:'open'"              json:"status"`
	CreatedAt     time.Time       `                                   json:"created_at"`
	UpdatedAt     time.Time       `                                   json:"updated_at"`
}

func (n *Necessity) BeforeCreate(_ *gorm.DB) error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	return nil
}
