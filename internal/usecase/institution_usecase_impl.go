package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"facilitador-de-doacoes/internal/model"
	"facilitador-de-doacoes/internal/repository"
)

type institutionUseCase struct {
	repo       repository.InstitutionRepository
	roleSetter RoleSetter
	supabaseClient SupabaseClient
}

func NewInstitutionUseCase(repo repository.InstitutionRepository, roleSetter RoleSetter, supabaseClient SupabaseClient) InstitutionUseCase {
	return &institutionUseCase{repo: repo, roleSetter: roleSetter, supabaseClient: supabaseClient}
}

func (uc *institutionUseCase) Create(auth0ID string, input CreateInstitutionInput) (*model.Institution, error) {
	// Se já existe pelo auth0_id, tenta setar o role novamente (retry após falha anterior)
	if existing, err := uc.repo.FindByAuth0ID(auth0ID); err == nil {
		if roleErr := uc.roleSetter.SetUserRole(context.Background(), auth0ID, "institution"); roleErr != nil {
			log.Printf("warn: set auth0 role for existing institution %s: %v", auth0ID, roleErr)
		}
		return existing, nil
	}

	_, err := uc.repo.FindByCNPJ(input.CNPJ)
	if err == nil {
		return nil, model.ErrCNPJAlreadyInUse
	}
	if !errors.Is(err, model.ErrNotFound) {
		return nil, err
	}

	institution := &model.Institution{
		Auth0ID:       auth0ID,
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
		WebsiteURL:    input.WebsiteURL,
		Status:        model.InstitutionStatusPending,
	}

	if err := uc.repo.Create(institution); err != nil {
		return nil, err
	}

	if err := uc.roleSetter.SetUserRole(context.Background(), auth0ID, "institution"); err != nil {
		log.Printf("warn: set auth0 role for new institution %s: %v", auth0ID, err)
	}

	return institution, nil
}

func (uc *institutionUseCase) GetByID(id uuid.UUID) (*model.Institution, error) {
	return uc.repo.FindByID(id)
}

func (uc *institutionUseCase) GetAll() ([]*model.Institution, error) {
	return uc.repo.FindAll()
}

func (uc *institutionUseCase) GetByAuth0ID(auth0ID string) (*model.Institution, error) {
	return uc.repo.FindByAuth0ID(auth0ID)
}

func (uc *institutionUseCase) Update(id uuid.UUID, institutionID uuid.UUID, input UpdateInstitutionInput) (*model.Institution, error) {
	if id != institutionID {
		return nil, model.ErrUnauthorized
	}

	institution, err := uc.repo.FindByID(id)
	if err != nil {
		return nil, err
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
	if input.WebsiteURL != "" {
		institution.WebsiteURL = input.WebsiteURL
	}

	if err := uc.repo.Update(institution); err != nil {
		return nil, err
	}
	return institution, nil
}

func (uc *institutionUseCase) Delete(id uuid.UUID, institutionID uuid.UUID) error {
	if id != institutionID {
		return model.ErrUnauthorized
	}

	_, err := uc.repo.FindByID(id)
	if err != nil {
		return err
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

func(uc *institutionUseCase) UploadImage(id uuid.UUID, institutionID uuid.UUID, imageType string, data []byte, contentType string) (*model.Institution, error) {
	if id != institutionID {
		return nil, model.ErrUnauthorized
	}

	institution, err := uc.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	fileName := fmt.Sprintf("instituitons/%s/%s", id.String(), imageType)
	url, err := uc.supabaseClient.UploadFile(context.Background(), fileName, data, contentType)
	if err != nil {
		return nil, err
	}

	if imageType == "logo"{
		institution.LogoURL = url
	} else if imageType == "cover"{
		institution.CoverImageURL = url
	} else {
		return nil, fmt.Errorf("Invalid image type: %s", imageType)
	}

	if err := uc.repo.Update(institution); err != nil {
		return nil, err
	}

	return institution, nil
}
