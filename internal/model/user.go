package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Auth0ID   string    `gorm:"uniqueIndex"          json:"auth0_id,omitempty"`
	Name      string    `gorm:"not null"             json:"name"`
	Email     string    `gorm:"uniqueIndex;not null" json:"email"`
	Password  string    `                            json:"-"`
	Role      string    `gorm:"default:'donor'"      json:"role"`
	CPF       string    `gorm:"uniqueIndex"          json:"cpf,omitempty"`
	Birthdate string    `json:"birthdate,omitempty"`
	Phone     string    `json:"phone,omitempty"`
	AvatarURL string    `json:"avatar_url,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (u *User) BeforeCreate(_ *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
