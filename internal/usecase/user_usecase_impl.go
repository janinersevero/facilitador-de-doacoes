package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"facilitador-de-doacoes/internal/model"
	"facilitador-de-doacoes/internal/repository"
)

// FileUploader abstracts file-upload operations
// (implemented by supabase.Client) for test purposes ;)
type FileUploader interface {
	UploadFile(ctx context.Context, fileName string, data []byte, contentType string) (string, error)
}

type userUseCase struct {
	repo       repository.UserRepository
	uploader   FileUploader
	roleSetter RoleSetter
}

func NewUserUseCase(repo repository.UserRepository, uploader FileUploader, roleSetter RoleSetter) UserUseCase {
	return &userUseCase{repo: repo, uploader: uploader, roleSetter: roleSetter}
}

func (uc *userUseCase) Create(auth0ID string, input CreateUserInput) (*model.User, error) {
	// Se já existe pelo auth0_id, tenta setar o role novamente (retry após falha anterior)
	if existing, err := uc.repo.FindByAuth0ID(auth0ID); err == nil {
		if roleErr := uc.roleSetter.SetUserRole(context.Background(), auth0ID, existing.Role); roleErr != nil {
			log.Printf("warn: set auth0 role for existing user %s: %v", auth0ID, roleErr)
		}
		return existing, nil
	}

	_, err := uc.repo.FindByEmail(input.Email)
	if err == nil {
		return nil, model.ErrEmailAlreadyInUse
	}
	if !errors.Is(err, model.ErrNotFound) {
		return nil, err
	}

	role := input.Role
	if role == "" {
		role = "donor"
	}

	user := &model.User{
		Auth0ID:   auth0ID,
		Name:      input.Name,
		Email:     input.Email,
		Role:      role,
		CPF:       input.CPF,
		Birthdate: input.Birthdate,
		Phone:     input.Phone,
	}

	if err := uc.repo.Create(user); err != nil {
		return nil, err
	}

	if err := uc.roleSetter.SetUserRole(context.Background(), auth0ID, role); err != nil {
		log.Printf("warn: set auth0 role for new user %s: %v", auth0ID, err)
	}

	return user, nil
}

func (uc *userUseCase) GetByID(id uuid.UUID) (*model.User, error) {
	return uc.repo.FindByID(id)
}

func (uc *userUseCase) GetByAuth0ID(auth0ID string) (*model.User, error) {
	return uc.repo.FindByAuth0ID(auth0ID)
}

func (uc *userUseCase) GetAll() ([]*model.User, error) {
	return uc.repo.FindAll()
}

func (uc *userUseCase) Update(id uuid.UUID, input UpdateUserInput) (*model.User, error) {
	user, err := uc.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if input.Name != "" {
		user.Name = input.Name
	}
	if input.Email != "" {
		user.Email = input.Email
	}
	if input.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		user.Password = string(hashed)
	}
	if input.CPF != "" {
		user.CPF = input.CPF
	}
	if input.Birthdate != "" {
		user.Birthdate = input.Birthdate
	}
	if input.Phone != "" {
		user.Phone = input.Phone
	}
	if input.Role != "" {
		user.Role = input.Role
	}

	if err := uc.repo.Update(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (uc *userUseCase) Delete(id uuid.UUID) error {
	if _, err := uc.repo.FindByID(id); err != nil {
		return err
	}
	return uc.repo.Delete(id)
}

func (uc *userUseCase) UploadAvatar(id uuid.UUID, fileBytes []byte, contentType string) (*model.User, error) {
	user, err := uc.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	ext := "jpg"
	switch contentType {
	case "image/png":
		ext = "png"
	case "image/webp":
		ext = "webp"
	}

	fileName := fmt.Sprintf("avatars/%s-%s.%s", id.String(), uuid.New().String(), ext)

	publicURL, err := uc.uploader.UploadFile(context.Background(), fileName, fileBytes, contentType)
	if err != nil {
		return nil, fmt.Errorf("upload avatar: %w", err)
	}

	if err := uc.repo.UpdateAvatarURL(id, publicURL); err != nil {
		return nil, err
	}

	user.AvatarURL = publicURL
	return user, nil
}
