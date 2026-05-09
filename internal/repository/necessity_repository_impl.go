package repository

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"facilitador-de-doacoes/internal/model"
)

type necessityRepository struct {
	db *gorm.DB
}

func NewNecessityRepository(db *gorm.DB) NecessityRepository {
	return &necessityRepository{db: db}
}

func (r *necessityRepository) Create(necessity *model.Necessity) error {
	return r.db.Create(necessity).Error
}

func (r *necessityRepository) FindByID(id uuid.UUID) (*model.Necessity, error) {
	var necessity model.Necessity
	err := r.db.First(&necessity, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	return &necessity, nil
}

func (r *necessityRepository) FindByInstitutionID(institutionID uuid.UUID) ([]*model.Necessity, error) {
	var necessities []*model.Necessity
	if err := r.db.Find(&necessities, "institution_id = ?", institutionID).Error; err != nil {
		return nil, err
	}
	return necessities, nil
}

func (r *necessityRepository) Update(necessity *model.Necessity) error {
	return r.db.Save(necessity).Error
}

func (r *necessityRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.Necessity{}, "id = ?", id).Error
}
