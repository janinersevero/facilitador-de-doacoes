package usecase

import (
	"errors"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"facilitador-de-doacoes/internal/model"
	"facilitador-de-doacoes/internal/repository"
)

type userUseCase struct {
	repo repository.UserRepository
}

func NewUserUseCase(repo repository.UserRepository) UserUseCase {
	return &userUseCase{repo: repo}
}

func (uc *userUseCase) Create(input CreateUserInput) (*model.User, error) {
	_, err := uc.repo.FindByEmail(input.Email)
	if err == nil {
		return nil, model.ErrEmailAlreadyInUse
	}
	if !errors.Is(err, model.ErrNotFound) {
		return nil, err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	role := input.Role
	if role == "" {
		role = "donor"
	}

	user := &model.User{
		Name:      input.Name,
		Email:     input.Email,
		Password:  string(hashed),
		Role:      role,
		CPF:       input.CPF,
		Birthdate: input.Birthdate,
		Phone:     input.Phone,
	}

	if err := uc.repo.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (uc *userUseCase) GetByID(id uuid.UUID) (*model.User, error) {
	return uc.repo.FindByID(id)
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
