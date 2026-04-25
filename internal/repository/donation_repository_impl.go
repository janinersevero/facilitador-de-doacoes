package repository

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"facilitador-de-doacoes/internal/model"
)

type donationRepository struct {
	db *gorm.DB
}

func NewDonationRepository(db *gorm.DB) DonationRepository {
	return &donationRepository{db: db}
}

func (r *donationRepository) Create(donation *model.Donation) error {
	return r.db.Create(donation).Error
}

func (r *donationRepository) FindByID(id uuid.UUID) (*model.Donation, error) {
	var d model.Donation
	err := r.db.First(&d, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	return &d, nil
}

func (r *donationRepository) FindByPixID(pixID string) (*model.Donation, error) {
	var d model.Donation
	err := r.db.First(&d, "pix_id = ?", pixID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	return &d, nil
}

func (r *donationRepository) FindAll() ([]*model.Donation, error) {
	var donations []*model.Donation
	if err := r.db.Find(&donations).Error; err != nil {
		return nil, err
	}
	return donations, nil
}

func (r *donationRepository) UpdateStatus(id uuid.UUID, status string) error {
	return r.db.Model(&model.Donation{}).
		Where("id = ?", id).
		Update("status", status).
		Error
}
