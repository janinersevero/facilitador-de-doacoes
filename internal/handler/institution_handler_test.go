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

func init() {
	gin.SetMode(gin.TestMode)
}

// mockInstitutionUseCase implements usecase.InstitutionUseCase for testing.
type mockInstitutionUseCase struct {
	mock.Mock
}

func (m *mockInstitutionUseCase) Create(userID uuid.UUID, input usecase.CreateInstitutionInput) (*model.Institution, error) {
	args := m.Called(userID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Institution), args.Error(1)
}

func (m *mockInstitutionUseCase) GetByID(id uuid.UUID) (*model.Institution, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Institution), args.Error(1)
}

func (m *mockInstitutionUseCase) GetAll() ([]*model.Institution, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Institution), args.Error(1)
}

func (m *mockInstitutionUseCase) GetByUserID(userID uuid.UUID) ([]*model.Institution, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Institution), args.Error(1)
}

func (m *mockInstitutionUseCase) Update(id uuid.UUID, userID uuid.UUID, input usecase.UpdateInstitutionInput) (*model.Institution, error) {
	args := m.Called(id, userID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Institution), args.Error(1)
}

func (m *mockInstitutionUseCase) Delete(id uuid.UUID, userID uuid.UUID) error {
	args := m.Called(id, userID)
	return args.Error(0)
}

func (m *mockInstitutionUseCase) UpdateStatus(id uuid.UUID, input usecase.UpdateInstitutionStatusInput) (*model.Institution, error) {
	args := m.Called(id, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Institution), args.Error(1)
}

// fakeAuth injects a userID into the Gin context, simulating auth middleware.
func fakeAuth(userID uuid.UUID) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("userID", userID)
		c.Next()
	}
}

func setupRouter(uc usecase.InstitutionUseCase, userID uuid.UUID) *gin.Engine {
	r := gin.New()
	h := handler.NewInstitutionHandler(uc)
	api := r.Group("/api/v1")
	h.RegisterRoutes(api, fakeAuth(userID))
	return r
}

func TestHandlerCreate_Success(t *testing.T) {
	uc := &mockInstitutionUseCase{}
	userID := uuid.New()
	r := setupRouter(uc, userID)

	input := usecase.CreateInstitutionInput{
		Name:      "ONG Esperança",
		LegalName: "ONG Esperança LTDA",
		CNPJ:      "12.345.678/0001-90",
	}
	expected := &model.Institution{
		ID:     uuid.New(),
		UserID: userID,
		Name:   input.Name,
		CNPJ:   input.CNPJ,
		Status: model.InstitutionStatusPending,
	}

	uc.On("Create", userID, input).Return(expected, nil)

	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/institutions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp model.Institution
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, expected.Name, resp.Name)
	uc.AssertExpectations(t)
}

func TestHandlerCreate_BadRequest(t *testing.T) {
	uc := &mockInstitutionUseCase{}
	r := setupRouter(uc, uuid.New())

	// Missing required fields
	body := []byte(`{}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/institutions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandlerCreate_CNPJConflict(t *testing.T) {
	uc := &mockInstitutionUseCase{}
	userID := uuid.New()
	r := setupRouter(uc, userID)

	input := usecase.CreateInstitutionInput{
		Name:      "ONG Esperança",
		LegalName: "ONG Esperança LTDA",
		CNPJ:      "12.345.678/0001-90",
	}

	uc.On("Create", userID, input).Return(nil, model.ErrCNPJAlreadyInUse)

	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/institutions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	uc.AssertExpectations(t)
}

func TestHandlerGetAll_Success(t *testing.T) {
	uc := &mockInstitutionUseCase{}
	r := setupRouter(uc, uuid.New())

	institutions := []*model.Institution{
		{ID: uuid.New(), Name: "ONG A"},
		{ID: uuid.New(), Name: "ONG B"},
	}
	uc.On("GetAll").Return(institutions, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/institutions", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp []*model.Institution
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Len(t, resp, 2)
	uc.AssertExpectations(t)
}

func TestHandlerGetByID_Success(t *testing.T) {
	uc := &mockInstitutionUseCase{}
	r := setupRouter(uc, uuid.New())

	id := uuid.New()
	expected := &model.Institution{ID: id, Name: "ONG Esperança"}
	uc.On("GetByID", id).Return(expected, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/institutions/"+id.String(), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp model.Institution
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, expected.Name, resp.Name)
	uc.AssertExpectations(t)
}

func TestHandlerGetByID_NotFound(t *testing.T) {
	uc := &mockInstitutionUseCase{}
	r := setupRouter(uc, uuid.New())

	id := uuid.New()
	uc.On("GetByID", id).Return(nil, model.ErrNotFound)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/institutions/"+id.String(), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	uc.AssertExpectations(t)
}

func TestHandlerGetByID_InvalidUUID(t *testing.T) {
	uc := &mockInstitutionUseCase{}
	r := setupRouter(uc, uuid.New())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/institutions/not-a-uuid", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandlerUpdate_Success(t *testing.T) {
	uc := &mockInstitutionUseCase{}
	userID := uuid.New()
	r := setupRouter(uc, userID)

	id := uuid.New()
	input := usecase.UpdateInstitutionInput{Name: "ONG Atualizada"}
	expected := &model.Institution{ID: id, Name: "ONG Atualizada"}

	uc.On("Update", id, userID, input).Return(expected, nil)

	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/institutions/"+id.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertExpectations(t)
}

func TestHandlerUpdate_NotFound(t *testing.T) {
	uc := &mockInstitutionUseCase{}
	userID := uuid.New()
	r := setupRouter(uc, userID)

	id := uuid.New()
	input := usecase.UpdateInstitutionInput{Name: "Nova"}

	uc.On("Update", id, userID, input).Return(nil, model.ErrNotFound)

	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/institutions/"+id.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	uc.AssertExpectations(t)
}

func TestHandlerUpdate_Forbidden(t *testing.T) {
	uc := &mockInstitutionUseCase{}
	userID := uuid.New()
	r := setupRouter(uc, userID)

	id := uuid.New()
	input := usecase.UpdateInstitutionInput{Name: "Nova"}

	uc.On("Update", id, userID, input).Return(nil, model.ErrUnauthorized)

	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/institutions/"+id.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	uc.AssertExpectations(t)
}

func TestHandlerDelete_Success(t *testing.T) {
	uc := &mockInstitutionUseCase{}
	userID := uuid.New()
	r := setupRouter(uc, userID)

	id := uuid.New()
	uc.On("Delete", id, userID).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/institutions/"+id.String(), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	uc.AssertExpectations(t)
}

func TestHandlerDelete_NotFound(t *testing.T) {
	uc := &mockInstitutionUseCase{}
	userID := uuid.New()
	r := setupRouter(uc, userID)

	id := uuid.New()
	uc.On("Delete", id, userID).Return(model.ErrNotFound)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/institutions/"+id.String(), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	uc.AssertExpectations(t)
}

func TestHandlerDelete_Forbidden(t *testing.T) {
	uc := &mockInstitutionUseCase{}
	userID := uuid.New()
	r := setupRouter(uc, userID)

	id := uuid.New()
	uc.On("Delete", id, userID).Return(model.ErrUnauthorized)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/institutions/"+id.String(), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	uc.AssertExpectations(t)
}

func TestHandlerUpdateStatus_Success(t *testing.T) {
	uc := &mockInstitutionUseCase{}
	r := setupRouter(uc, uuid.New())

	id := uuid.New()
	input := usecase.UpdateInstitutionStatusInput{
		Status: model.InstitutionStatusApproved,
	}
	expected := &model.Institution{ID: id, Status: model.InstitutionStatusApproved}

	uc.On("UpdateStatus", id, input).Return(expected, nil)

	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/institutions/"+id.String()+"/status", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertExpectations(t)
}

func TestHandlerUpdateStatus_NotFound(t *testing.T) {
	uc := &mockInstitutionUseCase{}
	r := setupRouter(uc, uuid.New())

	id := uuid.New()
	input := usecase.UpdateInstitutionStatusInput{Status: model.InstitutionStatusApproved}

	uc.On("UpdateStatus", id, input).Return(nil, model.ErrNotFound)

	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/institutions/"+id.String()+"/status", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	uc.AssertExpectations(t)
}
