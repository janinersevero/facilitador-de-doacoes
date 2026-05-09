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
	"facilitador-de-doacoes/internal/repository"
	"facilitador-de-doacoes/internal/usecase"
)

type mockCampaignUseCase struct {
	mock.Mock
}

func (m *mockCampaignUseCase) Create(institutionID uuid.UUID, input usecase.CreateCampaignInput) (*model.Campaign, error) {
	args := m.Called(institutionID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Campaign), args.Error(1)
}

func (m *mockCampaignUseCase) GetByID(id uuid.UUID) (*model.Campaign, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Campaign), args.Error(1)
}

func (m *mockCampaignUseCase) GetAll(filters repository.CampaignFilters) ([]*model.Campaign, error) {
	args := m.Called(filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Campaign), args.Error(1)
}

func (m *mockCampaignUseCase) GetByInstitutionID(institutionID uuid.UUID) ([]*model.Campaign, error) {
	args := m.Called(institutionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Campaign), args.Error(1)
}

func (m *mockCampaignUseCase) Update(id uuid.UUID, institutionID uuid.UUID, input usecase.UpdateCampaignInput) (*model.Campaign, error) {
	args := m.Called(id, institutionID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Campaign), args.Error(1)
}

func (m *mockCampaignUseCase) Delete(id uuid.UUID, institutionID uuid.UUID) error {
	return m.Called(id, institutionID).Error(0)
}

func (m *mockCampaignUseCase) UpdateStatus(id uuid.UUID, institutionID uuid.UUID, status model.CampaignStatus) (*model.Campaign, error) {
	args := m.Called(id, institutionID, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Campaign), args.Error(1)
}

func setupCampaignRouter(uc usecase.CampaignUseCase, instID uuid.UUID) *gin.Engine {
	r := gin.New()
	h := handler.NewCampaignHandler(uc)
	api := r.Group("/api/v1")
	h.RegisterRoutes(api, fakeInstitutionAuth(instID))
	return r
}

func TestCampaignHandlerCreate_Success(t *testing.T) {
	uc := &mockCampaignUseCase{}
	instID := uuid.New()
	r := setupCampaignRouter(uc, instID)

	input := usecase.CreateCampaignInput{
		Title:       "Campanha Saúde",
		Description: "Ajuda médica",
		GoalAmount:  100000,
		Keywords:    []string{"saúde"},
	}
	expected := &model.Campaign{ID: uuid.New(), Title: input.Title, InstitutionID: instID}

	uc.On("Create", instID, input).Return(expected, nil)

	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/institutions/"+instID.String()+"/campaigns", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	uc.AssertExpectations(t)
}

func TestCampaignHandlerCreate_BadRequest(t *testing.T) {
	uc := &mockCampaignUseCase{}
	instID := uuid.New()
	r := setupCampaignRouter(uc, instID)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/institutions/"+instID.String()+"/campaigns", bytes.NewBuffer([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCampaignHandlerCreate_Forbidden(t *testing.T) {
	uc := &mockCampaignUseCase{}
	instID := uuid.New()
	otherInstID := uuid.New()
	r := setupCampaignRouter(uc, instID) // authenticated as instID

	input := usecase.CreateCampaignInput{Title: "X", Description: "Y", GoalAmount: 1000}
	body, _ := json.Marshal(input)
	// path has otherInstID but context has instID → mismatch → 403
	req := httptest.NewRequest(http.MethodPost, "/api/v1/institutions/"+otherInstID.String()+"/campaigns", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCampaignHandlerGetAll_NoFilter(t *testing.T) {
	uc := &mockCampaignUseCase{}
	r := setupCampaignRouter(uc, uuid.New())

	campaigns := []*model.Campaign{{ID: uuid.New(), Title: "A"}, {ID: uuid.New(), Title: "B"}}
	uc.On("GetAll", repository.CampaignFilters{}).Return(campaigns, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/campaigns", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp []*model.Campaign
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Len(t, resp, 2)
	uc.AssertExpectations(t)
}

func TestCampaignHandlerGetAll_WithKeyword(t *testing.T) {
	uc := &mockCampaignUseCase{}
	r := setupCampaignRouter(uc, uuid.New())

	campaigns := []*model.Campaign{{ID: uuid.New(), Title: "Saúde"}}
	uc.On("GetAll", repository.CampaignFilters{Keyword: "saúde"}).Return(campaigns, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/campaigns?keyword=saúde", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertExpectations(t)
}

func TestCampaignHandlerGetByID_Success(t *testing.T) {
	uc := &mockCampaignUseCase{}
	r := setupCampaignRouter(uc, uuid.New())

	id := uuid.New()
	expected := &model.Campaign{ID: id, Title: "Campanha"}
	uc.On("GetByID", id).Return(expected, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/campaigns/"+id.String(), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertExpectations(t)
}

func TestCampaignHandlerGetByID_NotFound(t *testing.T) {
	uc := &mockCampaignUseCase{}
	r := setupCampaignRouter(uc, uuid.New())

	id := uuid.New()
	uc.On("GetByID", id).Return(nil, model.ErrNotFound)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/campaigns/"+id.String(), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCampaignHandlerUpdate_Success(t *testing.T) {
	uc := &mockCampaignUseCase{}
	instID := uuid.New()
	r := setupCampaignRouter(uc, instID)

	id := uuid.New()
	input := usecase.UpdateCampaignInput{Title: "Novo Título"}
	expected := &model.Campaign{ID: id, Title: "Novo Título"}

	uc.On("Update", id, instID, input).Return(expected, nil)

	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/campaigns/"+id.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertExpectations(t)
}

func TestCampaignHandlerUpdate_Forbidden(t *testing.T) {
	uc := &mockCampaignUseCase{}
	instID := uuid.New()
	r := setupCampaignRouter(uc, instID)

	id := uuid.New()
	input := usecase.UpdateCampaignInput{Title: "X"}
	uc.On("Update", id, instID, input).Return(nil, model.ErrUnauthorized)

	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/campaigns/"+id.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCampaignHandlerDelete_Success(t *testing.T) {
	uc := &mockCampaignUseCase{}
	instID := uuid.New()
	r := setupCampaignRouter(uc, instID)

	id := uuid.New()
	uc.On("Delete", id, instID).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/campaigns/"+id.String(), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	uc.AssertExpectations(t)
}

func TestCampaignHandlerUpdateStatus_Success(t *testing.T) {
	uc := &mockCampaignUseCase{}
	instID := uuid.New()
	r := setupCampaignRouter(uc, instID)

	id := uuid.New()
	expected := &model.Campaign{ID: id, Status: model.CampaignStatusPaused}

	type statusBody struct {
		Status model.CampaignStatus `json:"status"`
	}
	uc.On("UpdateStatus", id, instID, model.CampaignStatusPaused).Return(expected, nil)

	body, _ := json.Marshal(statusBody{Status: model.CampaignStatusPaused})
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/campaigns/"+id.String()+"/status", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertExpectations(t)
}
