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

func (r *donationRepository) FindByPaymentID(paymentID string) (*model.Donation, error) {
	var d model.Donation
	err := r.db.First(&d, "payment_id = ?", paymentID).Error
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

func (r *donationRepository) GetRanking(limit int) ([]*model.RankingEntry, error) {
	var entries []*model.RankingEntry

	err := r.db.Raw(`
		SELECT
			u.id        AS user_id,
			u.name      AS user_name,
			u.avatar_url AS avatar_url,
			SUM(d.amount) / 100 AS points,
			SUM(d.amount)       AS total_donated,
			COUNT(d.id)         AS donation_count
		FROM donations d
		JOIN users u ON u.id = d.user_id
		WHERE d.status = 'PAID'
		GROUP BY u.id, u.name, u.avatar_url
		ORDER BY points DESC
		LIMIT ?
	`, limit).Scan(&entries).Error

	if err != nil {
		return nil, err
	}

	for i := range entries {
		entries[i].Position = i + 1
	}

	return entries, nil
}
