package repository

import (
	"github.com/google/uuid"

	"facilitador-de-doacoes/internal/model"
)

type InstitutionRepository interface {
	Create(institution *model.Institution) error
	FindByID(id uuid.UUID) (*model.Institution, error)
	FindByCNPJ(cnpj string) (*model.Institution, error)
	FindByUserID(userID uuid.UUID) ([]*model.Institution, error)
	FindAll() ([]*model.Institution, error)
	Update(institution *model.Institution) error
	Delete(id uuid.UUID) error
}
