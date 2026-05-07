package repository

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"facilitador-de-doacoes/internal/model"
)

type campaignRepository struct {
	db *gorm.DB
}

func NewCampaignRepository(db *gorm.DB) CampaignRepository {
	return &campaignRepository{db: db}
}

func (r *campaignRepository) Create(campaign *model.Campaign) error {
	return r.db.Create(campaign).Error
}

func (r *campaignRepository) FindByID(id uuid.UUID) (*model.Campaign, error) {
	var campaign model.Campaign
	err := r.db.First(&campaign, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	return &campaign, nil
}

func (r *campaignRepository) FindAll(filters CampaignFilters) ([]*model.Campaign, error) {
	var campaigns []*model.Campaign
	q := r.db.Model(&model.Campaign{})

	if filters.Keyword != "" {
		q = q.Where("? = ANY(keywords)", filters.Keyword)
	}
	if filters.InstitutionID != nil {
		q = q.Where("institution_id = ?", *filters.InstitutionID)
	}
	if filters.IsUrgent != nil {
		q = q.Where("is_urgent = ?", *filters.IsUrgent)
	}

	if err := q.Find(&campaigns).Error; err != nil {
		return nil, err
	}
	return campaigns, nil
}

func (r *campaignRepository) FindByInstitutionID(institutionID uuid.UUID) ([]*model.Campaign, error) {
	var campaigns []*model.Campaign
	if err := r.db.Find(&campaigns, "institution_id = ?", institutionID).Error; err != nil {
		return nil, err
	}
	return campaigns, nil
}

func (r *campaignRepository) Update(campaign *model.Campaign) error {
	return r.db.Save(campaign).Error
}

func (r *campaignRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.Campaign{}, "id = ?", id).Error
}

func (r *campaignRepository) IncrementTotalRaised(id uuid.UUID, amount int64) error {
	return r.db.Model(&model.Campaign{}).
		Where("id = ?", id).
		UpdateColumn("total_raised", gorm.Expr("total_raised + ?", amount)).
		Error
}
