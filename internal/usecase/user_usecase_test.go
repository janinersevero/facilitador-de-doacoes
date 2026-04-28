package usecase_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	"facilitador-de-doacoes/internal/model"
	"facilitador-de-doacoes/internal/usecase"
	"facilitador-de-doacoes/internal/usecase/mocks"
)

// --------------- Create ---------------

func TestCreateUser_Success(t *testing.T) {
	repo := new(mocks.MockUserRepository)
	repo.On("FindByEmail", "alice@example.com").Return(nil, model.ErrNotFound)
	repo.On("Create", mock.AnythingOfType("*model.User")).Return(nil)

	uc := usecase.NewUserUseCase(repo, nil)

	user, err := uc.Create(usecase.CreateUserInput{
		Name:     "Alice",
		Email:    "alice@example.com",
		Password: "secret123",
	})

	assert.NoError(t, err)
	assert.Equal(t, "Alice", user.Name)
	assert.Equal(t, "alice@example.com", user.Email)
	assert.Equal(t, "donor", user.Role, "default role should be donor")
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("secret123")))
	repo.AssertExpectations(t)
}

func TestCreateUser_WithCustomRole(t *testing.T) {
	repo := new(mocks.MockUserRepository)
	repo.On("FindByEmail", "bob@example.com").Return(nil, model.ErrNotFound)
	repo.On("Create", mock.AnythingOfType("*model.User")).Return(nil)

	uc := usecase.NewUserUseCase(repo, nil)

	user, err := uc.Create(usecase.CreateUserInput{
		Name:     "Bob",
		Email:    "bob@example.com",
		Password: "secret123",
		Role:     "admin",
	})

	assert.NoError(t, err)
	assert.Equal(t, "admin", user.Role)
	repo.AssertExpectations(t)
}

func TestCreateUser_EmailAlreadyInUse(t *testing.T) {
	repo := new(mocks.MockUserRepository)
	existing := &model.User{Email: "alice@example.com"}
	repo.On("FindByEmail", "alice@example.com").Return(existing, nil)

	uc := usecase.NewUserUseCase(repo, nil)

	_, err := uc.Create(usecase.CreateUserInput{
		Name:     "Alice",
		Email:    "alice@example.com",
		Password: "secret123",
	})

	assert.ErrorIs(t, err, model.ErrEmailAlreadyInUse)
	repo.AssertExpectations(t)
}

func TestCreateUser_RepoFindError(t *testing.T) {
	repo := new(mocks.MockUserRepository)
	repo.On("FindByEmail", "alice@example.com").Return(nil, errors.New("db connection lost"))

	uc := usecase.NewUserUseCase(repo, nil)

	_, err := uc.Create(usecase.CreateUserInput{
		Name:     "Alice",
		Email:    "alice@example.com",
		Password: "secret123",
	})

	assert.Error(t, err)
	assert.Equal(t, "db connection lost", err.Error())
}

func TestCreateUser_RepoCreateError(t *testing.T) {
	repo := new(mocks.MockUserRepository)
	repo.On("FindByEmail", "alice@example.com").Return(nil, model.ErrNotFound)
	repo.On("Create", mock.AnythingOfType("*model.User")).Return(errors.New("insert failed"))

	uc := usecase.NewUserUseCase(repo, nil)

	_, err := uc.Create(usecase.CreateUserInput{
		Name:     "Alice",
		Email:    "alice@example.com",
		Password: "secret123",
	})

	assert.EqualError(t, err, "insert failed")
}

// --------------- GetByID ---------------

func TestGetByID_Success(t *testing.T) {
	repo := new(mocks.MockUserRepository)
	id := uuid.New()
	expected := &model.User{ID: id, Name: "Alice"}
	repo.On("FindByID", id).Return(expected, nil)

	uc := usecase.NewUserUseCase(repo, nil)

	user, err := uc.GetByID(id)
	assert.NoError(t, err)
	assert.Equal(t, expected, user)
}

func TestGetByID_NotFound(t *testing.T) {
	repo := new(mocks.MockUserRepository)
	id := uuid.New()
	repo.On("FindByID", id).Return(nil, model.ErrNotFound)

	uc := usecase.NewUserUseCase(repo, nil)

	_, err := uc.GetByID(id)
	assert.ErrorIs(t, err, model.ErrNotFound)
}

// --------------- GetAll ---------------

func TestGetAll_Success(t *testing.T) {
	repo := new(mocks.MockUserRepository)
	users := []*model.User{{Name: "Alice"}, {Name: "Bob"}}
	repo.On("FindAll").Return(users, nil)

	uc := usecase.NewUserUseCase(repo, nil)

	result, err := uc.GetAll()
	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

// --------------- Update ---------------

func TestUpdate_Success(t *testing.T) {
	repo := new(mocks.MockUserRepository)
	id := uuid.New()
	existing := &model.User{ID: id, Name: "Alice", Email: "alice@example.com", Role: "donor"}
	repo.On("FindByID", id).Return(existing, nil)
	repo.On("Update", mock.AnythingOfType("*model.User")).Return(nil)

	uc := usecase.NewUserUseCase(repo, nil)

	user, err := uc.Update(id, usecase.UpdateUserInput{
		Name:  "Alice Updated",
		Phone: "123456789",
	})

	assert.NoError(t, err)
	assert.Equal(t, "Alice Updated", user.Name)
	assert.Equal(t, "123456789", user.Phone)
	assert.Equal(t, "alice@example.com", user.Email, "unchanged fields should be preserved")
	repo.AssertExpectations(t)
}

func TestUpdate_PasswordIsHashed(t *testing.T) {
	repo := new(mocks.MockUserRepository)
	id := uuid.New()
	existing := &model.User{ID: id, Name: "Alice", Password: "old-hash"}
	repo.On("FindByID", id).Return(existing, nil)
	repo.On("Update", mock.AnythingOfType("*model.User")).Return(nil)

	uc := usecase.NewUserUseCase(repo, nil)

	user, err := uc.Update(id, usecase.UpdateUserInput{Password: "newpass123"})

	assert.NoError(t, err)
	assert.NotEqual(t, "newpass123", user.Password, "password should be hashed")
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("newpass123")))
}

func TestUpdate_NotFound(t *testing.T) {
	repo := new(mocks.MockUserRepository)
	id := uuid.New()
	repo.On("FindByID", id).Return(nil, model.ErrNotFound)

	uc := usecase.NewUserUseCase(repo, nil)

	_, err := uc.Update(id, usecase.UpdateUserInput{Name: "X"})
	assert.ErrorIs(t, err, model.ErrNotFound)
}

// --------------- Delete ---------------

func TestDelete_Success(t *testing.T) {
	repo := new(mocks.MockUserRepository)
	id := uuid.New()
	repo.On("FindByID", id).Return(&model.User{ID: id}, nil)
	repo.On("Delete", id).Return(nil)

	uc := usecase.NewUserUseCase(repo, nil)

	err := uc.Delete(id)
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestDelete_NotFound(t *testing.T) {
	repo := new(mocks.MockUserRepository)
	id := uuid.New()
	repo.On("FindByID", id).Return(nil, model.ErrNotFound)

	uc := usecase.NewUserUseCase(repo, nil)

	err := uc.Delete(id)
	assert.ErrorIs(t, err, model.ErrNotFound)
}

// --------------- UploadAvatar ---------------

func TestUploadAvatar_Success(t *testing.T) {
	repo := new(mocks.MockUserRepository)
	uploader := new(mocks.MockFileUploader)
	id := uuid.New()

	existing := &model.User{ID: id, Name: "Alice"}
	repo.On("FindByID", id).Return(existing, nil)
	uploader.On("UploadFile", mock.Anything, mock.MatchedBy(func(name string) bool {
		return len(name) > 0
	}), []byte("img-data"), "image/png").Return("https://cdn.example.com/avatar.png", nil)
	repo.On("UpdateAvatarURL", id, "https://cdn.example.com/avatar.png").Return(nil)

	uc := usecase.NewUserUseCase(repo, uploader)

	user, err := uc.UploadAvatar(id, []byte("img-data"), "image/png")

	assert.NoError(t, err)
	assert.Equal(t, "https://cdn.example.com/avatar.png", user.AvatarURL)
	repo.AssertExpectations(t)
	uploader.AssertExpectations(t)
}

func TestUploadAvatar_UserNotFound(t *testing.T) {
	repo := new(mocks.MockUserRepository)
	id := uuid.New()
	repo.On("FindByID", id).Return(nil, model.ErrNotFound)

	uc := usecase.NewUserUseCase(repo, nil)

	_, err := uc.UploadAvatar(id, []byte("img"), "image/png")
	assert.ErrorIs(t, err, model.ErrNotFound)
}

func TestUploadAvatar_UploadError(t *testing.T) {
	repo := new(mocks.MockUserRepository)
	uploader := new(mocks.MockFileUploader)
	id := uuid.New()

	repo.On("FindByID", id).Return(&model.User{ID: id}, nil)
	uploader.On("UploadFile", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return("", errors.New("storage unavailable"))

	uc := usecase.NewUserUseCase(repo, uploader)

	_, err := uc.UploadAvatar(id, []byte("img"), "image/jpeg")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "upload avatar")
}
