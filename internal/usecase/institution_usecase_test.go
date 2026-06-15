package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"facilitador-de-doacoes/internal/model"
	"facilitador-de-doacoes/internal/usecase"
)

type mockSupabaseClient struct{ mock.Mock }

func (m *mockSupabaseClient) UploadFile(ctx context.Context, fileName string, data []byte, contentType string) (string, error) {
	args := m.Called(ctx, fileName, data, contentType)
	return args.String(0), args.Error(1)
}

type mockRoleSetter struct{ mock.Mock }

func (m *mockRoleSetter) SetUserRole(ctx context.Context, auth0UserID, role string) error {
	return m.Called(ctx, auth0UserID, role).Error(0)
}

type mockInstitutionRepo struct {
	mock.Mock
}

func (m *mockInstitutionRepo) Create(institution *model.Institution) error {
	return m.Called(institution).Error(0)
}

func (m *mockInstitutionRepo) FindByID(id uuid.UUID) (*model.Institution, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Institution), args.Error(1)
}

func (m *mockInstitutionRepo) FindByCNPJ(cnpj string) (*model.Institution, error) {
	args := m.Called(cnpj)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Institution), args.Error(1)
}

func (m *mockInstitutionRepo) FindByAuth0ID(auth0ID string) (*model.Institution, error) {
	args := m.Called(auth0ID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Institution), args.Error(1)
}

func (m *mockInstitutionRepo) FindAll() ([]*model.Institution, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Institution), args.Error(1)
}

func (m *mockInstitutionRepo) Update(institution *model.Institution) error {
	return m.Called(institution).Error(0)
}

func (m *mockInstitutionRepo) Delete(id uuid.UUID) error {
	return m.Called(id).Error(0)
}

func newInstitutionUseCase(repo *mockInstitutionRepo) usecase.InstitutionUseCase {
	return usecase.NewInstitutionUseCase(repo, nil, nil)
}

func TestInstitutionCreate_Success(t *testing.T) {
	repo := &mockInstitutionRepo{}
	roleSetter := new(mockRoleSetter)
	uc := usecase.NewInstitutionUseCase(repo, roleSetter, nil)

	auth0ID := "auth0|abc123"
	input := usecase.CreateInstitutionInput{
		Name:      "ONG Esperança",
		LegalName: "ONG Esperança LTDA",
		CNPJ:      "12.345.678/0001-90",
		Category:  "social",
	}

	repo.On("FindByAuth0ID", auth0ID).Return(nil, model.ErrNotFound)
	repo.On("FindByCNPJ", input.CNPJ).Return(nil, model.ErrNotFound)
	repo.On("Create", mock.MatchedBy(func(i *model.Institution) bool {
		return i.Name == input.Name && i.CNPJ == input.CNPJ && i.Auth0ID == auth0ID
	})).Return(nil)
	roleSetter.On("SetUserRole", mock.Anything, auth0ID, "institution").Return(nil)

	institution, err := uc.Create(auth0ID, input)

	assert.NoError(t, err)
	assert.Equal(t, input.Name, institution.Name)
	assert.Equal(t, input.LegalName, institution.LegalName)
	assert.Equal(t, input.CNPJ, institution.CNPJ)
	assert.Equal(t, auth0ID, institution.Auth0ID)
	assert.Equal(t, model.InstitutionStatusPending, institution.Status)
	repo.AssertExpectations(t)
	roleSetter.AssertExpectations(t)
}

func TestInstitutionCreate_CNPJAlreadyInUse(t *testing.T) {
	repo := &mockInstitutionRepo{}
	uc := newInstitutionUseCase(repo)

	input := usecase.CreateInstitutionInput{
		Name:      "ONG Esperança",
		LegalName: "ONG Esperança LTDA",
		CNPJ:      "12.345.678/0001-90",
	}
	existing := &model.Institution{ID: uuid.New(), CNPJ: input.CNPJ}

	repo.On("FindByAuth0ID", "auth0|other").Return(nil, model.ErrNotFound)
	repo.On("FindByCNPJ", input.CNPJ).Return(existing, nil)

	_, err := uc.Create("auth0|other", input)

	assert.ErrorIs(t, err, model.ErrCNPJAlreadyInUse)
	repo.AssertExpectations(t)
}

func TestInstitutionGetByID_Success(t *testing.T) {
	repo := &mockInstitutionRepo{}
	uc := newInstitutionUseCase(repo)

	id := uuid.New()
	expected := &model.Institution{ID: id, Name: "ONG Esperança"}

	repo.On("FindByID", id).Return(expected, nil)

	institution, err := uc.GetByID(id)

	assert.NoError(t, err)
	assert.Equal(t, expected, institution)
	repo.AssertExpectations(t)
}

func TestInstitutionGetByID_NotFound(t *testing.T) {
	repo := &mockInstitutionRepo{}
	uc := newInstitutionUseCase(repo)

	id := uuid.New()
	repo.On("FindByID", id).Return(nil, model.ErrNotFound)

	_, err := uc.GetByID(id)

	assert.ErrorIs(t, err, model.ErrNotFound)
	repo.AssertExpectations(t)
}

func TestInstitutionGetAll_Success(t *testing.T) {
	repo := &mockInstitutionRepo{}
	uc := newInstitutionUseCase(repo)

	institutions := []*model.Institution{
		{ID: uuid.New(), Name: "ONG A"},
		{ID: uuid.New(), Name: "ONG B"},
	}

	repo.On("FindAll").Return(institutions, nil)

	result, err := uc.GetAll()

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	repo.AssertExpectations(t)
}

func TestInstitutionGetByAuth0ID_Success(t *testing.T) {
	repo := &mockInstitutionRepo{}
	uc := newInstitutionUseCase(repo)

	auth0ID := "auth0|abc123"
	expected := &model.Institution{ID: uuid.New(), Auth0ID: auth0ID, Name: "ONG A"}

	repo.On("FindByAuth0ID", auth0ID).Return(expected, nil)

	result, err := uc.GetByAuth0ID(auth0ID)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	repo.AssertExpectations(t)
}

func TestInstitutionUpdate_Success(t *testing.T) {
	repo := &mockInstitutionRepo{}
	uc := newInstitutionUseCase(repo)

	id := uuid.New()
	existing := &model.Institution{ID: id, Name: "ONG Velha"}
	input := usecase.UpdateInstitutionInput{Name: "ONG Nova"}

	repo.On("FindByID", id).Return(existing, nil)
	repo.On("Update", mock.MatchedBy(func(i *model.Institution) bool {
		return i.Name == "ONG Nova"
	})).Return(nil)

	institution, err := uc.Update(id, id, input) // id == institutionID (updating itself)

	assert.NoError(t, err)
	assert.Equal(t, "ONG Nova", institution.Name)
	repo.AssertExpectations(t)
}

func TestInstitutionUpdate_NotFound(t *testing.T) {
	repo := &mockInstitutionRepo{}
	uc := newInstitutionUseCase(repo)

	id := uuid.New()
	repo.On("FindByID", id).Return(nil, model.ErrNotFound)

	_, err := uc.Update(id, id, usecase.UpdateInstitutionInput{Name: "Nova"})

	assert.ErrorIs(t, err, model.ErrNotFound)
	repo.AssertExpectations(t)
}

func TestInstitutionUpdate_Unauthorized(t *testing.T) {
	repo := &mockInstitutionRepo{}
	uc := newInstitutionUseCase(repo)

	id := uuid.New()
	otherInstID := uuid.New()

	// id != otherInstID → ErrUnauthorized without hitting repo
	_, err := uc.Update(id, otherInstID, usecase.UpdateInstitutionInput{Name: "Nova"})

	assert.ErrorIs(t, err, model.ErrUnauthorized)
	repo.AssertExpectations(t) // no repo calls expected
}

func TestInstitutionDelete_Success(t *testing.T) {
	repo := &mockInstitutionRepo{}
	uc := newInstitutionUseCase(repo)

	id := uuid.New()
	repo.On("FindByID", id).Return(&model.Institution{ID: id}, nil)
	repo.On("Delete", id).Return(nil)

	err := uc.Delete(id, id)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestInstitutionDelete_NotFound(t *testing.T) {
	repo := &mockInstitutionRepo{}
	uc := newInstitutionUseCase(repo)

	id := uuid.New()
	repo.On("FindByID", id).Return(nil, model.ErrNotFound)

	err := uc.Delete(id, id)

	assert.ErrorIs(t, err, model.ErrNotFound)
	repo.AssertExpectations(t)
}

func TestInstitutionDelete_Unauthorized(t *testing.T) {
	repo := &mockInstitutionRepo{}
	uc := newInstitutionUseCase(repo)

	id := uuid.New()
	otherInstID := uuid.New()

	// id != otherInstID → ErrUnauthorized without hitting repo
	err := uc.Delete(id, otherInstID)

	assert.ErrorIs(t, err, model.ErrUnauthorized)
	repo.AssertExpectations(t) // no repo calls expected
}

func TestInstitutionUpdateStatus_Approve(t *testing.T) {
	repo := &mockInstitutionRepo{}
	uc := newInstitutionUseCase(repo)

	id := uuid.New()
	existing := &model.Institution{ID: id, Status: model.InstitutionStatusPending}

	repo.On("FindByID", id).Return(existing, nil)
	repo.On("Update", mock.MatchedBy(func(i *model.Institution) bool {
		return i.Status == model.InstitutionStatusApproved && i.ApprovedAt != nil
	})).Return(nil)

	institution, err := uc.UpdateStatus(id, usecase.UpdateInstitutionStatusInput{
		Status: model.InstitutionStatusApproved,
	})

	assert.NoError(t, err)
	assert.Equal(t, model.InstitutionStatusApproved, institution.Status)
	assert.NotNil(t, institution.ApprovedAt)
	repo.AssertExpectations(t)
}

func TestInstitutionUpdateStatus_Reject(t *testing.T) {
	repo := &mockInstitutionRepo{}
	uc := newInstitutionUseCase(repo)

	id := uuid.New()
	approvedAt := time.Now()
	existing := &model.Institution{
		ID:         id,
		Status:     model.InstitutionStatusApproved,
		ApprovedAt: &approvedAt,
	}

	repo.On("FindByID", id).Return(existing, nil)
	repo.On("Update", mock.MatchedBy(func(i *model.Institution) bool {
		return i.Status == model.InstitutionStatusRejected &&
			i.RejectionReason == "documentação inválida" &&
			i.ApprovedAt == nil
	})).Return(nil)

	institution, err := uc.UpdateStatus(id, usecase.UpdateInstitutionStatusInput{
		Status:          model.InstitutionStatusRejected,
		RejectionReason: "documentação inválida",
	})

	assert.NoError(t, err)
	assert.Equal(t, model.InstitutionStatusRejected, institution.Status)
	assert.Equal(t, "documentação inválida", institution.RejectionReason)
	assert.Nil(t, institution.ApprovedAt)
	repo.AssertExpectations(t)
}

func TestInstitutionUploadImage_Logo_Success(t *testing.T) {
	repo := &mockInstitutionRepo{}
	supabase := &mockSupabaseClient{}
	id := uuid.New()
	uc := usecase.NewInstitutionUseCase(repo, nil, supabase)

	existing := &model.Institution{ID: id, Name: "ONG A"}
	data := []byte("fake-image")
	contentType := "image/png"
	expectedURL := "https://cdn.example.com/logo.png"

	repo.On("FindByID", id).Return(existing, nil)
	supabase.On("UploadFile", mock.Anything, mock.AnythingOfType("string"), data, contentType).Return(expectedURL, nil)
	repo.On("Update", mock.MatchedBy(func(i *model.Institution) bool {
		return i.LogoURL == expectedURL
	})).Return(nil)

	institution, err := uc.UploadImage(id, id, "logo", data, contentType)

	assert.NoError(t, err)
	assert.Equal(t, expectedURL, institution.LogoURL)
	repo.AssertExpectations(t)
	supabase.AssertExpectations(t)
}

func TestInstitutionUploadImage_Cover_Success(t *testing.T) {
	repo := &mockInstitutionRepo{}
	supabase := &mockSupabaseClient{}
	id := uuid.New()
	uc := usecase.NewInstitutionUseCase(repo, nil, supabase)

	existing := &model.Institution{ID: id, Name: "ONG A"}
	data := []byte("fake-image")
	contentType := "image/jpeg"
	expectedURL := "https://cdn.example.com/cover.jpg"

	repo.On("FindByID", id).Return(existing, nil)
	supabase.On("UploadFile", mock.Anything, mock.AnythingOfType("string"), data, contentType).Return(expectedURL, nil)
	repo.On("Update", mock.MatchedBy(func(i *model.Institution) bool {
		return i.CoverImageURL == expectedURL
	})).Return(nil)

	institution, err := uc.UploadImage(id, id, "cover", data, contentType)

	assert.NoError(t, err)
	assert.Equal(t, expectedURL, institution.CoverImageURL)
	repo.AssertExpectations(t)
	supabase.AssertExpectations(t)
}

func TestInstitutionUploadImage_Unauthorized(t *testing.T) {
	repo := &mockInstitutionRepo{}
	uc := usecase.NewInstitutionUseCase(repo, nil, nil)

	id := uuid.New()
	otherID := uuid.New()

	_, err := uc.UploadImage(id, otherID, "logo", []byte{}, "image/png")

	assert.ErrorIs(t, err, model.ErrUnauthorized)
	repo.AssertExpectations(t)
}

func TestInstitutionUploadImage_NotFound(t *testing.T) {
	repo := &mockInstitutionRepo{}
	supabase := &mockSupabaseClient{}
	id := uuid.New()
	uc := usecase.NewInstitutionUseCase(repo, nil, supabase)

	repo.On("FindByID", id).Return(nil, model.ErrNotFound)

	_, err := uc.UploadImage(id, id, "logo", []byte{}, "image/png")

	assert.ErrorIs(t, err, model.ErrNotFound)
	repo.AssertExpectations(t)
}
