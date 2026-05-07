package repository

import (
	"github.com/google/uuid"

	"facilitador-de-doacoes/internal/model"
)

type UserRepository interface {
	Create(user *model.User) error
	FindByID(id uuid.UUID) (*model.User, error)
	FindByEmail(email string) (*model.User, error)
	FindByAuth0ID(auth0ID string) (*model.User, error)
	FindAll() ([]*model.User, error)
	Update(user *model.User) error
	UpdateAvatarURL(id uuid.UUID, url string) error
	Delete(id uuid.UUID) error
}
