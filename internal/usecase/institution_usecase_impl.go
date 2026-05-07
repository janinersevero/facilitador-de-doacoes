package usecase

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"facilitador-de-doacoes/internal/model"
	"facilitador-de-doacoes/internal/repository"
)

type institutionUseCase struct {
	repo repository.InstitutionRepository
}

func NewInstitutionUseCase(repo repository.InstitutionRepository) InstitutionUseCase {
	return &institutionUseCase{repo: repo}
}

func (uc *institutionUseCase) Create(userID uuid.UUID, input CreateInstitutionInput) (*model.Institution, error) {
	_, err := uc.repo.FindByCNPJ(input.CNPJ)
	if err == nil {
		return nil, model.ErrCNPJAlreadyInUse
	}
	if !errors.Is(err, model.ErrNotFound) {
		return nil, err
	}

	institution := &model.Institution{
		UserID:        userID,
		Name:          input.Name,
		LegalName:     input.LegalName,
		CNPJ:          input.CNPJ,
		Description:   input.Description,
		Email:         input.Email,
		Phone:         input.Phone,
		Address:       input.Address,
		ZipCode:       input.ZipCode,
		Category:      input.Category,
		LogoURL:       input.LogoURL,
		CoverImageURL: input.CoverImageURL,
		VideoURL:      input.VideoURL,
		WebsiteURL:    input.WebsiteURL,
		Status:        model.InstitutionStatusPending,
	}

	if err := uc.repo.Create(institution); err != nil {
		return nil, err
	}
	return institution, nil
}

func (uc *institutionUseCase) GetByID(id uuid.UUID) (*model.Institution, error) {
	return uc.repo.FindByID(id)
}

func (uc *institutionUseCase) GetAll() ([]*model.Institution, error) {
	return uc.repo.FindAll()
}

func (uc *institutionUseCase) GetByUserID(userID uuid.UUID) ([]*model.Institution, error) {
	return uc.repo.FindByUserID(userID)
}

func (uc *institutionUseCase) Update(id uuid.UUID, userID uuid.UUID, input UpdateInstitutionInput) (*model.Institution, error) {
	institution, err := uc.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if institution.UserID != userID {
		return nil, model.ErrUnauthorized
	}

	if input.Name != "" {
		institution.Name = input.Name
	}
	if input.LegalName != "" {
		institution.LegalName = input.LegalName
	}
	if input.Description != "" {
		institution.Description = input.Description
	}
	if input.Email != "" {
		institution.Email = input.Email
	}
	if input.Phone != "" {
		institution.Phone = input.Phone
	}
	if input.Address != "" {
		institution.Address = input.Address
	}
	if input.ZipCode != "" {
		institution.ZipCode = input.ZipCode
	}
	if input.Category != "" {
		institution.Category = input.Category
	}
	if input.LogoURL != "" {
		institution.LogoURL = input.LogoURL
	}
	if input.CoverImageURL != "" {
		institution.CoverImageURL = input.CoverImageURL
	}
	if input.VideoURL != "" {
		institution.VideoURL = input.VideoURL
	}
	if input.WebsiteURL != "" {
		institution.WebsiteURL = input.WebsiteURL
	}

	if err := uc.repo.Update(institution); err != nil {
		return nil, err
	}
	return institution, nil
}

func (uc *institutionUseCase) Delete(id uuid.UUID, userID uuid.UUID) error {
	institution, err := uc.repo.FindByID(id)
	if err != nil {
		return err
	}

	if institution.UserID != userID {
		return model.ErrUnauthorized
	}

	return uc.repo.Delete(id)
}

func (uc *institutionUseCase) UpdateStatus(id uuid.UUID, input UpdateInstitutionStatusInput) (*model.Institution, error) {
	institution, err := uc.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	institution.Status = input.Status

	switch input.Status {
	case model.InstitutionStatusApproved:
		now := time.Now()
		institution.ApprovedAt = &now
		institution.RejectionReason = ""
	case model.InstitutionStatusRejected:
		institution.RejectionReason = input.RejectionReason
		institution.ApprovedAt = nil
	}

	if err := uc.repo.Update(institution); err != nil {
		return nil, err
	}
	return institution, nil
}
