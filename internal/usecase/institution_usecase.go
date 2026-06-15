package usecase

import (
	"context"

	"github.com/google/uuid"

	"facilitador-de-doacoes/internal/model"
)

type SupabaseClient interface {
	UploadFile(ctx context.Context, fileName string, data []byte, contentType string) (string, error)
}

type CreateInstitutionInput struct {
	Name          string `json:"name"           binding:"required"`
	LegalName     string `json:"legal_name"     binding:"required"`
	CNPJ          string `json:"cnpj"           binding:"required"`
	Description   string `json:"description"`
	Email         string `json:"email"          binding:"omitempty,email"`
	Phone         string `json:"phone"`
	Address       string `json:"address"`
	ZipCode       string `json:"zip_code"`
	Category      string `json:"category"`
	LogoURL       string `json:"logo_url"`
	CoverImageURL string `json:"cover_image_url"`
	WebsiteURL    string `json:"website_url"`
}

type UpdateInstitutionInput struct {
	Name          string `json:"name"`
	LegalName     string `json:"legal_name"`
	Description   string `json:"description"`
	Email         string `json:"email"          binding:"omitempty,email"`
	Phone         string `json:"phone"`
	Address       string `json:"address"`
	ZipCode       string `json:"zip_code"`
	Category      string `json:"category"`
	LogoURL       string `json:"logo_url"`
	CoverImageURL string `json:"cover_image_url"`
	WebsiteURL    string `json:"website_url"`
}

type UpdateInstitutionStatusInput struct {
	Status          model.InstitutionStatus `json:"status"            binding:"required"`
	RejectionReason string                  `json:"rejection_reason"`
}

type InstitutionUseCase interface {
	Create(auth0ID string, input CreateInstitutionInput) (*model.Institution, error)
	GetByID(id uuid.UUID) (*model.Institution, error)
	GetAll() ([]*model.Institution, error)
	GetByAuth0ID(auth0ID string) (*model.Institution, error)
	Update(id uuid.UUID, institutionID uuid.UUID, input UpdateInstitutionInput) (*model.Institution, error)
	Delete(id uuid.UUID, institutionID uuid.UUID) error
	UpdateStatus(id uuid.UUID, input UpdateInstitutionStatusInput) (*model.Institution, error)
	UploadImage(id uuid.UUID, institutionID uuid.UUID, imageType string, data []byte, contentType string) (*model.Institution, error)
}
