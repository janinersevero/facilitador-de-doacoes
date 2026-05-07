package usecase_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"facilitador-de-doacoes/internal/model"
	"facilitador-de-doacoes/internal/usecase"
)

// mockInstitutionRepo implements repository.InstitutionRepository for testing.
type mockInstitutionRepo struct {
	mock.Mock
}

func (m *mockInstitutionRepo) Create(institution *model.Institution) error {
	args := m.Called(institution)
	return args.Error(0)
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

func (m *mockInstitutionRepo) FindByUserID(userID uuid.UUID) ([]*model.Institution, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Institution), args.Error(1)
}

func (m *mockInstitutionRepo) FindAll() ([]*model.Institution, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Institution), args.Error(1)
}

func (m *mockInstitutionRepo) Update(institution *model.Institution) error {
	args := m.Called(institution)
	return args.Error(0)
}

func (m *mockInstitutionRepo) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func newInstitutionUseCase(repo *mockInstitutionRepo) usecase.InstitutionUseCase {
	return usecase.NewInstitutionUseCase(repo)
}

func TestInstitutionCreate_Success(t *testing.T) {
	repo := &mockInstitutionRepo{}
	uc := newInstitutionUseCase(repo)

	userID := uuid.New()
	input := usecase.CreateInstitutionInput{
		Name:      "ONG Esperança",
		LegalName: "ONG Esperança LTDA",
		CNPJ:      "12.345.678/0001-90",
		Category:  "social",
	}

	repo.On("FindByCNPJ", input.CNPJ).Return(nil, model.ErrNotFound)
	repo.On("Create", mock.MatchedBy(func(i *model.Institution) bool {
		return i.Name == input.Name && i.CNPJ == input.CNPJ && i.UserID == userID
	})).Return(nil)

	institution, err := uc.Create(userID, input)

	assert.NoError(t, err)
	assert.Equal(t, input.Name, institution.Name)
	assert.Equal(t, input.LegalName, institution.LegalName)
	assert.Equal(t, input.CNPJ, institution.CNPJ)
	assert.Equal(t, userID, institution.UserID)
	assert.Equal(t, model.InstitutionStatusPending, institution.Status)
	repo.AssertExpectations(t)
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

	repo.On("FindByCNPJ", input.CNPJ).Return(existing, nil)

	_, err := uc.Create(uuid.New(), input)

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

func TestInstitutionGetByUserID_Success(t *testing.T) {
	repo := &mockInstitutionRepo{}
	uc := newInstitutionUseCase(repo)

	userID := uuid.New()
	institutions := []*model.Institution{
		{ID: uuid.New(), UserID: userID, Name: "ONG A"},
	}

	repo.On("FindByUserID", userID).Return(institutions, nil)

	result, err := uc.GetByUserID(userID)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	repo.AssertExpectations(t)
}

func TestInstitutionUpdate_Success(t *testing.T) {
	repo := &mockInstitutionRepo{}
	uc := newInstitutionUseCase(repo)

	userID := uuid.New()
	id := uuid.New()
	existing := &model.Institution{
		ID:     id,
		UserID: userID,
		Name:   "ONG Velha",
	}
	input := usecase.UpdateInstitutionInput{
		Name: "ONG Nova",
	}

	repo.On("FindByID", id).Return(existing, nil)
	repo.On("Update", mock.MatchedBy(func(i *model.Institution) bool {
		return i.Name == "ONG Nova"
	})).Return(nil)

	institution, err := uc.Update(id, userID, input)

	assert.NoError(t, err)
	assert.Equal(t, "ONG Nova", institution.Name)
	repo.AssertExpectations(t)
}

func TestInstitutionUpdate_NotFound(t *testing.T) {
	repo := &mockInstitutionRepo{}
	uc := newInstitutionUseCase(repo)

	id := uuid.New()
	repo.On("FindByID", id).Return(nil, model.ErrNotFound)

	_, err := uc.Update(id, uuid.New(), usecase.UpdateInstitutionInput{Name: "Nova"})

	assert.ErrorIs(t, err, model.ErrNotFound)
	repo.AssertExpectations(t)
}

func TestInstitutionUpdate_Unauthorized(t *testing.T) {
	repo := &mockInstitutionRepo{}
	uc := newInstitutionUseCase(repo)

	ownerID := uuid.New()
	otherUserID := uuid.New()
	id := uuid.New()
	existing := &model.Institution{ID: id, UserID: ownerID, Name: "ONG"}

	repo.On("FindByID", id).Return(existing, nil)

	_, err := uc.Update(id, otherUserID, usecase.UpdateInstitutionInput{Name: "Nova"})

	assert.ErrorIs(t, err, model.ErrUnauthorized)
	repo.AssertExpectations(t)
}

func TestInstitutionDelete_Success(t *testing.T) {
	repo := &mockInstitutionRepo{}
	uc := newInstitutionUseCase(repo)

	userID := uuid.New()
	id := uuid.New()
	existing := &model.Institution{ID: id, UserID: userID}

	repo.On("FindByID", id).Return(existing, nil)
	repo.On("Delete", id).Return(nil)

	err := uc.Delete(id, userID)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestInstitutionDelete_NotFound(t *testing.T) {
	repo := &mockInstitutionRepo{}
	uc := newInstitutionUseCase(repo)

	id := uuid.New()
	repo.On("FindByID", id).Return(nil, model.ErrNotFound)

	err := uc.Delete(id, uuid.New())

	assert.ErrorIs(t, err, model.ErrNotFound)
	repo.AssertExpectations(t)
}

func TestInstitutionDelete_Unauthorized(t *testing.T) {
	repo := &mockInstitutionRepo{}
	uc := newInstitutionUseCase(repo)

	ownerID := uuid.New()
	otherUserID := uuid.New()
	id := uuid.New()
	existing := &model.Institution{ID: id, UserID: ownerID}

	repo.On("FindByID", id).Return(existing, nil)

	err := uc.Delete(id, otherUserID)

	assert.ErrorIs(t, err, model.ErrUnauthorized)
	repo.AssertExpectations(t)
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
