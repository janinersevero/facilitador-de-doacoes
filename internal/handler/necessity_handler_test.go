package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"facilitador-de-doacoes/internal/handler"
	"facilitador-de-doacoes/internal/model"
	"facilitador-de-doacoes/internal/usecase"
)

type mockNecessityUseCase struct {
	mock.Mock
}

func (m *mockNecessityUseCase) Create(institutionID uuid.UUID, input usecase.CreateNecessityInput) (*model.Necessity, error) {
	args := m.Called(institutionID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Necessity), args.Error(1)
}

func (m *mockNecessityUseCase) GetByID(id uuid.UUID) (*model.Necessity, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Necessity), args.Error(1)
}

func (m *mockNecessityUseCase) GetByInstitutionID(institutionID uuid.UUID) ([]*model.Necessity, error) {
	args := m.Called(institutionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Necessity), args.Error(1)
}

func (m *mockNecessityUseCase) Update(id uuid.UUID, institutionID uuid.UUID, input usecase.UpdateNecessityInput) (*model.Necessity, error) {
	args := m.Called(id, institutionID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Necessity), args.Error(1)
}

func (m *mockNecessityUseCase) Delete(id uuid.UUID, institutionID uuid.UUID) error {
	return m.Called(id, institutionID).Error(0)
}

func (m *mockNecessityUseCase) UpdateStatus(id uuid.UUID, institutionID uuid.UUID, status model.NecessityStatus) (*model.Necessity, error) {
	args := m.Called(id, institutionID, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Necessity), args.Error(1)
}

func setupNecessityRouter(uc usecase.NecessityUseCase, instID uuid.UUID) *gin.Engine {
	r := gin.New()
	h := handler.NewNecessityHandler(uc)
	api := r.Group("/api/v1")
	h.RegisterRoutes(api, fakeInstitutionAuth(instID))
	return r
}

func TestNecessityHandlerCreate_Success(t *testing.T) {
	uc := &mockNecessityUseCase{}
	instID := uuid.New()
	r := setupNecessityRouter(uc, instID)

	input := usecase.CreateNecessityInput{
		Description: "Fraldas para bebês",
		Category:    "higiene",
		IsUrgent:    true,
	}
	expected := &model.Necessity{
		ID:            uuid.New(),
		InstitutionID: instID,
		Description:   input.Description,
		Category:      input.Category,
		IsUrgent:      true,
		Status:        model.NecessityStatusOpen,
	}

	uc.On("Create", instID, input).Return(expected, nil)

	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/institutions/"+instID.String()+"/necessities", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp model.Necessity
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, expected.Description, resp.Description)
	uc.AssertExpectations(t)
}

func TestNecessityHandlerCreate_BadRequest(t *testing.T) {
	uc := &mockNecessityUseCase{}
	instID := uuid.New()
	r := setupNecessityRouter(uc, instID)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/institutions/"+instID.String()+"/necessities", bytes.NewBuffer([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestNecessityHandlerCreate_Forbidden(t *testing.T) {
	uc := &mockNecessityUseCase{}
	instID := uuid.New()
	otherInstID := uuid.New()
	r := setupNecessityRouter(uc, instID) // authenticated as instID

	input := usecase.CreateNecessityInput{Description: "X", Category: "Y"}
	body, _ := json.Marshal(input)
	// path has otherInstID but context has instID → mismatch → 403
	req := httptest.NewRequest(http.MethodPost, "/api/v1/institutions/"+otherInstID.String()+"/necessities", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestNecessityHandlerGetByInstitution_Success(t *testing.T) {
	uc := &mockNecessityUseCase{}
	r := setupNecessityRouter(uc, uuid.New())

	instID := uuid.New()
	necessities := []*model.Necessity{
		{ID: uuid.New(), Description: "A"},
		{ID: uuid.New(), Description: "B"},
	}
	uc.On("GetByInstitutionID", instID).Return(necessities, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/institutions/"+instID.String()+"/necessities", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp []*model.Necessity
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Len(t, resp, 2)
	uc.AssertExpectations(t)
}

func TestNecessityHandlerGetByID_Success(t *testing.T) {
	uc := &mockNecessityUseCase{}
	r := setupNecessityRouter(uc, uuid.New())

	id := uuid.New()
	expected := &model.Necessity{ID: id, Description: "Remédios"}
	uc.On("GetByID", id).Return(expected, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/necessities/"+id.String(), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertExpectations(t)
}

func TestNecessityHandlerGetByID_NotFound(t *testing.T) {
	uc := &mockNecessityUseCase{}
	r := setupNecessityRouter(uc, uuid.New())

	id := uuid.New()
	uc.On("GetByID", id).Return(nil, model.ErrNotFound)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/necessities/"+id.String(), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestNecessityHandlerUpdate_Success(t *testing.T) {
	uc := &mockNecessityUseCase{}
	instID := uuid.New()
	r := setupNecessityRouter(uc, instID)

	id := uuid.New()
	input := usecase.UpdateNecessityInput{Description: "Novo"}
	expected := &model.Necessity{ID: id, Description: "Novo"}

	uc.On("Update", id, instID, input).Return(expected, nil)

	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/necessities/"+id.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertExpectations(t)
}

func TestNecessityHandlerUpdate_Forbidden(t *testing.T) {
	uc := &mockNecessityUseCase{}
	instID := uuid.New()
	r := setupNecessityRouter(uc, instID)

	id := uuid.New()
	input := usecase.UpdateNecessityInput{Description: "X"}
	uc.On("Update", id, instID, input).Return(nil, model.ErrUnauthorized)

	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/necessities/"+id.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestNecessityHandlerDelete_Success(t *testing.T) {
	uc := &mockNecessityUseCase{}
	instID := uuid.New()
	r := setupNecessityRouter(uc, instID)

	id := uuid.New()
	uc.On("Delete", id, instID).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/necessities/"+id.String(), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	uc.AssertExpectations(t)
}

func TestNecessityHandlerDelete_NotFound(t *testing.T) {
	uc := &mockNecessityUseCase{}
	instID := uuid.New()
	r := setupNecessityRouter(uc, instID)

	id := uuid.New()
	uc.On("Delete", id, instID).Return(model.ErrNotFound)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/necessities/"+id.String(), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestNecessityHandlerUpdateStatus_Success(t *testing.T) {
	uc := &mockNecessityUseCase{}
	instID := uuid.New()
	r := setupNecessityRouter(uc, instID)

	id := uuid.New()
	expected := &model.Necessity{ID: id, Status: model.NecessityStatusAttended}

	type statusBody struct {
		Status model.NecessityStatus `json:"status"`
	}
	uc.On("UpdateStatus", id, instID, model.NecessityStatusAttended).Return(expected, nil)

	body, _ := json.Marshal(statusBody{Status: model.NecessityStatusAttended})
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/necessities/"+id.String()+"/status", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertExpectations(t)
}

func TestNecessityHandlerUpdateStatus_Forbidden(t *testing.T) {
	uc := &mockNecessityUseCase{}
	instID := uuid.New()
	r := setupNecessityRouter(uc, instID)

	id := uuid.New()
	type statusBody struct {
		Status model.NecessityStatus `json:"status"`
	}
	uc.On("UpdateStatus", id, instID, model.NecessityStatusAttended).Return(nil, model.ErrUnauthorized)

	body, _ := json.Marshal(statusBody{Status: model.NecessityStatusAttended})
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/necessities/"+id.String()+"/status", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}
