package repository

import (
	"github.com/google/uuid"

	"facilitador-de-doacoes/internal/model"
)

type NecessityRepository interface {
	Create(necessity *model.Necessity) error
	FindByID(id uuid.UUID) (*model.Necessity, error)
	FindByInstitutionID(institutionID uuid.UUID) ([]*model.Necessity, error)
	Update(necessity *model.Necessity) error
	Delete(id uuid.UUID) error
}
