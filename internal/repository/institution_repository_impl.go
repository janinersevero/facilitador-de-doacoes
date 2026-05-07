package repository

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"facilitador-de-doacoes/internal/model"
)

type institutionRepository struct {
	db *gorm.DB
}

func NewInstitutionRepository(db *gorm.DB) InstitutionRepository {
	return &institutionRepository{db: db}
}

func (r *institutionRepository) Create(institution *model.Institution) error {
	return r.db.Create(institution).Error
}

func (r *institutionRepository) FindByID(id uuid.UUID) (*model.Institution, error) {
	var institution model.Institution
	err := r.db.First(&institution, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	return &institution, nil
}

func (r *institutionRepository) FindByCNPJ(cnpj string) (*model.Institution, error) {
	var institution model.Institution
	err := r.db.First(&institution, "cnpj = ?", cnpj).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	return &institution, nil
}

func (r *institutionRepository) FindByUserID(userID uuid.UUID) ([]*model.Institution, error) {
	var institutions []*model.Institution
	if err := r.db.Find(&institutions, "user_id = ?", userID).Error; err != nil {
		return nil, err
	}
	return institutions, nil
}

func (r *institutionRepository) FindAll() ([]*model.Institution, error) {
	var institutions []*model.Institution
	if err := r.db.Find(&institutions).Error; err != nil {
		return nil, err
	}
	return institutions, nil
}

func (r *institutionRepository) Update(institution *model.Institution) error {
	return r.db.Save(institution).Error
}

func (r *institutionRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.Institution{}, "id = ?", id).Error
}
