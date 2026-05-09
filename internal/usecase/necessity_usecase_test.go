package usecase_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"facilitador-de-doacoes/internal/model"
	"facilitador-de-doacoes/internal/usecase"
	"facilitador-de-doacoes/internal/usecase/mocks"
)

func TestNecessityCreate_Success(t *testing.T) {
	nr := new(mocks.MockNecessityRepository)
	uc := usecase.NewNecessityUseCase(nr)

	instID := uuid.New()
	input := usecase.CreateNecessityInput{
		Description: "Fraldas descartáveis para bebês",
		Category:    "higiene",
		IsUrgent:    true,
	}

	nr.On("Create", mock.MatchedBy(func(n *model.Necessity) bool {
		return n.InstitutionID == instID &&
			n.Description == input.Description &&
			n.Category == input.Category &&
			n.IsUrgent == true &&
			n.Status == model.NecessityStatusOpen
	})).Return(nil)

	necessity, err := uc.Create(instID, input)

	assert.NoError(t, err)
	assert.Equal(t, input.Description, necessity.Description)
	assert.Equal(t, input.Category, necessity.Category)
	assert.True(t, necessity.IsUrgent)
	assert.Equal(t, model.NecessityStatusOpen, necessity.Status)
	assert.Equal(t, instID, necessity.InstitutionID)
	nr.AssertExpectations(t)
}

func TestNecessityGetByID_Success(t *testing.T) {
	nr := new(mocks.MockNecessityRepository)
	uc := usecase.NewNecessityUseCase(nr)

	id := uuid.New()
	expected := &model.Necessity{ID: id, Description: "Remédios"}
	nr.On("FindByID", id).Return(expected, nil)

	necessity, err := uc.GetByID(id)

	assert.NoError(t, err)
	assert.Equal(t, expected, necessity)
	nr.AssertExpectations(t)
}

func TestNecessityGetByID_NotFound(t *testing.T) {
	nr := new(mocks.MockNecessityRepository)
	uc := usecase.NewNecessityUseCase(nr)

	id := uuid.New()
	nr.On("FindByID", id).Return(nil, model.ErrNotFound)

	_, err := uc.GetByID(id)

	assert.ErrorIs(t, err, model.ErrNotFound)
}

func TestNecessityGetByInstitutionID_Success(t *testing.T) {
	nr := new(mocks.MockNecessityRepository)
	uc := usecase.NewNecessityUseCase(nr)

	instID := uuid.New()
	necessities := []*model.Necessity{
		{ID: uuid.New(), InstitutionID: instID, Description: "A"},
		{ID: uuid.New(), InstitutionID: instID, Description: "B"},
	}
	nr.On("FindByInstitutionID", instID).Return(necessities, nil)

	result, err := uc.GetByInstitutionID(instID)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	nr.AssertExpectations(t)
}

func TestNecessityUpdate_Success(t *testing.T) {
	nr := new(mocks.MockNecessityRepository)
	uc := usecase.NewNecessityUseCase(nr)

	instID := uuid.New()
	id := uuid.New()
	existing := &model.Necessity{ID: id, InstitutionID: instID, Description: "Antigo", Category: "saúde"}
	input := usecase.UpdateNecessityInput{Description: "Novo"}

	nr.On("FindByID", id).Return(existing, nil)
	nr.On("Update", mock.MatchedBy(func(n *model.Necessity) bool {
		return n.Description == "Novo"
	})).Return(nil)

	necessity, err := uc.Update(id, instID, input)

	assert.NoError(t, err)
	assert.Equal(t, "Novo", necessity.Description)
	assert.Equal(t, "saúde", necessity.Category)
	nr.AssertExpectations(t)
}

func TestNecessityUpdate_NotFound(t *testing.T) {
	nr := new(mocks.MockNecessityRepository)
	uc := usecase.NewNecessityUseCase(nr)

	id := uuid.New()
	nr.On("FindByID", id).Return(nil, model.ErrNotFound)

	_, err := uc.Update(id, uuid.New(), usecase.UpdateNecessityInput{Description: "X"})

	assert.ErrorIs(t, err, model.ErrNotFound)
}

func TestNecessityUpdate_Unauthorized(t *testing.T) {
	nr := new(mocks.MockNecessityRepository)
	uc := usecase.NewNecessityUseCase(nr)

	instID := uuid.New()
	otherInstID := uuid.New()
	id := uuid.New()

	nr.On("FindByID", id).Return(&model.Necessity{ID: id, InstitutionID: instID}, nil)

	_, err := uc.Update(id, otherInstID, usecase.UpdateNecessityInput{Description: "X"})

	assert.ErrorIs(t, err, model.ErrUnauthorized)
}

func TestNecessityDelete_Success(t *testing.T) {
	nr := new(mocks.MockNecessityRepository)
	uc := usecase.NewNecessityUseCase(nr)

	instID := uuid.New()
	id := uuid.New()

	nr.On("FindByID", id).Return(&model.Necessity{ID: id, InstitutionID: instID}, nil)
	nr.On("Delete", id).Return(nil)

	err := uc.Delete(id, instID)

	assert.NoError(t, err)
	nr.AssertExpectations(t)
}

func TestNecessityDelete_Unauthorized(t *testing.T) {
	nr := new(mocks.MockNecessityRepository)
	uc := usecase.NewNecessityUseCase(nr)

	instID := uuid.New()
	otherInstID := uuid.New()
	id := uuid.New()

	nr.On("FindByID", id).Return(&model.Necessity{ID: id, InstitutionID: instID}, nil)

	err := uc.Delete(id, otherInstID)

	assert.ErrorIs(t, err, model.ErrUnauthorized)
}

func TestNecessityUpdateStatus_MarkAttended(t *testing.T) {
	nr := new(mocks.MockNecessityRepository)
	uc := usecase.NewNecessityUseCase(nr)

	instID := uuid.New()
	id := uuid.New()

	nr.On("FindByID", id).Return(&model.Necessity{ID: id, InstitutionID: instID, Status: model.NecessityStatusOpen}, nil)
	nr.On("Update", mock.MatchedBy(func(n *model.Necessity) bool {
		return n.Status == model.NecessityStatusAttended
	})).Return(nil)

	necessity, err := uc.UpdateStatus(id, instID, model.NecessityStatusAttended)

	assert.NoError(t, err)
	assert.Equal(t, model.NecessityStatusAttended, necessity.Status)
	nr.AssertExpectations(t)
}

func TestNecessityUpdateStatus_Unauthorized(t *testing.T) {
	nr := new(mocks.MockNecessityRepository)
	uc := usecase.NewNecessityUseCase(nr)

	instID := uuid.New()
	otherInstID := uuid.New()
	id := uuid.New()

	nr.On("FindByID", id).Return(&model.Necessity{ID: id, InstitutionID: instID}, nil)

	_, err := uc.UpdateStatus(id, otherInstID, model.NecessityStatusAttended)

	assert.ErrorIs(t, err, model.ErrUnauthorized)
}
