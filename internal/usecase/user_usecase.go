package usecase

import (
	"context"

	"github.com/google/uuid"

	"facilitador-de-doacoes/internal/model"
)

// RoleSetter persists the Auth0 app_metadata role so the Login Action can inject
// it as a custom claim on the next token issuance.
type RoleSetter interface {
	SetUserRole(ctx context.Context, auth0UserID, role string) error
}

type CreateUserInput struct {
	Name      string `json:"name"  binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Role      string `json:"role"`
	CPF       string `json:"cpf"`
	Birthdate string `json:"birthdate"`
	Phone     string `json:"phone"`
}

type UpdateUserInput struct {
	Name      string `json:"name"`
	Email     string `json:"email"    binding:"omitempty,email"`
	Password  string `json:"password" binding:"omitempty,min=6"`
	Role      string `json:"role"`
	CPF       string `json:"cpf"`
	Birthdate string `json:"birthdate"`
	Phone     string `json:"phone"`
}

type UserUseCase interface {
	Create(auth0ID string, input CreateUserInput) (*model.User, error)
	GetByID(id uuid.UUID) (*model.User, error)
	GetByAuth0ID(auth0ID string) (*model.User, error)
	GetAll() ([]*model.User, error)
	Update(id uuid.UUID, input UpdateUserInput) (*model.User, error)
	Delete(id uuid.UUID) error
	UploadAvatar(id uuid.UUID, fileBytes []byte, contentType string) (*model.User, error)
}
