package handler_test

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
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

type mockInstitutionUseCase struct {
	mock.Mock
}

func (m *mockInstitutionUseCase) Create(auth0ID string, input usecase.CreateInstitutionInput) (*model.Institution, error) {
	args := m.Called(auth0ID, input)
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

func (m *mockInstitutionUseCase) GetByAuth0ID(auth0ID string) (*model.Institution, error) {
	args := m.Called(auth0ID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Institution), args.Error(1)
}

func (m *mockInstitutionUseCase) Update(id uuid.UUID, institutionID uuid.UUID, input usecase.UpdateInstitutionInput) (*model.Institution, error) {
	args := m.Called(id, institutionID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Institution), args.Error(1)
}

func (m *mockInstitutionUseCase) Delete(id uuid.UUID, institutionID uuid.UUID) error {
	return m.Called(id, institutionID).Error(0)
}

func (m *mockInstitutionUseCase) UpdateStatus(id uuid.UUID, input usecase.UpdateInstitutionStatusInput) (*model.Institution, error) {
	args := m.Called(id, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Institution), args.Error(1)
}

func (m *mockInstitutionUseCase) UploadImage(id uuid.UUID, institutionID uuid.UUID, imageType string, data []byte, contentType string) (*model.Institution, error) {
	args := m.Called(id, institutionID, imageType, data, contentType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Institution), args.Error(1)
}

// fakeAuth injects userID into Gin context (for user-authenticated routes).
func fakeAuth(userID uuid.UUID) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("userID", userID)
		c.Next()
	}
}

// fakeAuthSub injects auth0_sub into Gin context (simulates AuthMiddleware).
func fakeAuthSub(sub string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("auth0_sub", sub)
		c.Next()
	}
}

// fakeInstitutionAuth injects institutionID into Gin context (simulates RequireInstitution).
func fakeInstitutionAuth(institutionID uuid.UUID) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("institutionID", institutionID)
		c.Next()
	}
}

func setupRouter(uc usecase.InstitutionUseCase, auth0Sub string, instID uuid.UUID) *gin.Engine {
	r := gin.New()
	h := handler.NewInstitutionHandler(uc)
	api := r.Group("/api/v1")
	h.RegisterRoutes(api,
		[]gin.HandlerFunc{fakeAuthSub(auth0Sub)},
		[]gin.HandlerFunc{fakeInstitutionAuth(instID)},
	)
	return r
}

func TestHandlerCreate_Success(t *testing.T) {
	uc := &mockInstitutionUseCase{}
	auth0Sub := "auth0|abc123"
	instID := uuid.New()
	r := setupRouter(uc, auth0Sub, instID)

	input := usecase.CreateInstitutionInput{
		Name:      "ONG Esperança",
		LegalName: "ONG Esperança LTDA",
		CNPJ:      "12.345.678/0001-90",
	}
	expected := &model.Institution{
		ID:      instID,
		Auth0ID: auth0Sub,
		Name:    input.Name,
		CNPJ:    input.CNPJ,
		Status:  model.InstitutionStatusPending,
	}

	uc.On("Create", auth0Sub, input).Return(expected, nil)

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
	r := setupRouter(uc, "auth0|x", uuid.New())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/institutions", bytes.NewBuffer([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandlerCreate_CNPJConflict(t *testing.T) {
	uc := &mockInstitutionUseCase{}
	auth0Sub := "auth0|abc123"
	r := setupRouter(uc, auth0Sub, uuid.New())

	input := usecase.CreateInstitutionInput{
		Name:      "ONG Esperança",
		LegalName: "ONG Esperança LTDA",
		CNPJ:      "12.345.678/0001-90",
	}
	uc.On("Create", auth0Sub, input).Return(nil, model.ErrCNPJAlreadyInUse)

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
	r := setupRouter(uc, "auth0|x", uuid.New())

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
	r := setupRouter(uc, "auth0|x", uuid.New())

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
	r := setupRouter(uc, "auth0|x", uuid.New())

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
	r := setupRouter(uc, "auth0|x", uuid.New())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/institutions/not-a-uuid", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandlerUpdate_Success(t *testing.T) {
	uc := &mockInstitutionUseCase{}
	instID := uuid.New()
	r := setupRouter(uc, "auth0|x", instID)

	input := usecase.UpdateInstitutionInput{Name: "ONG Atualizada"}
	expected := &model.Institution{ID: instID, Name: "ONG Atualizada"}

	uc.On("Update", instID, instID, input).Return(expected, nil)

	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/institutions/"+instID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertExpectations(t)
}

func TestHandlerUpdate_NotFound(t *testing.T) {
	uc := &mockInstitutionUseCase{}
	instID := uuid.New()
	r := setupRouter(uc, "auth0|x", instID)

	input := usecase.UpdateInstitutionInput{Name: "Nova"}
	uc.On("Update", instID, instID, input).Return(nil, model.ErrNotFound)

	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/institutions/"+instID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	uc.AssertExpectations(t)
}

func TestHandlerUpdate_Forbidden(t *testing.T) {
	uc := &mockInstitutionUseCase{}
	instID := uuid.New()
	otherInstID := uuid.New()
	r := setupRouter(uc, "auth0|x", instID) // authenticated as instID

	input := usecase.UpdateInstitutionInput{Name: "Nova"}
	uc.On("Update", otherInstID, instID, input).Return(nil, model.ErrUnauthorized)

	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/institutions/"+otherInstID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	uc.AssertExpectations(t)
}

func TestHandlerDelete_Success(t *testing.T) {
	uc := &mockInstitutionUseCase{}
	instID := uuid.New()
	r := setupRouter(uc, "auth0|x", instID)

	uc.On("Delete", instID, instID).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/institutions/"+instID.String(), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	uc.AssertExpectations(t)
}

func TestHandlerDelete_NotFound(t *testing.T) {
	uc := &mockInstitutionUseCase{}
	instID := uuid.New()
	r := setupRouter(uc, "auth0|x", instID)

	uc.On("Delete", instID, instID).Return(model.ErrNotFound)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/institutions/"+instID.String(), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	uc.AssertExpectations(t)
}

func TestHandlerDelete_Forbidden(t *testing.T) {
	uc := &mockInstitutionUseCase{}
	instID := uuid.New()
	otherInstID := uuid.New()
	r := setupRouter(uc, "auth0|x", instID)

	uc.On("Delete", otherInstID, instID).Return(model.ErrUnauthorized)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/institutions/"+otherInstID.String(), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	uc.AssertExpectations(t)
}

func TestHandlerUpdateStatus_Success(t *testing.T) {
	uc := &mockInstitutionUseCase{}
	r := setupRouter(uc, "auth0|x", uuid.New())

	id := uuid.New()
	input := usecase.UpdateInstitutionStatusInput{Status: model.InstitutionStatusApproved}
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
	r := setupRouter(uc, "auth0|x", uuid.New())

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

func newMultipartRequest(t *testing.T, method, url string, fileContent []byte) (*http.Request, string) {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("image", "test.png")
	assert.NoError(t, err)
	_, err = part.Write(fileContent)
	assert.NoError(t, err)
	writer.Close()
	req := httptest.NewRequest(method, url, body)
	return req, writer.FormDataContentType()
}

func TestHandlerUploadImage_Logo_Success(t *testing.T) {
	uc := &mockInstitutionUseCase{}
	instID := uuid.New()
	r := setupRouter(uc, "auth0|x", instID)

	imgData := []byte("fake-png-data")
	expected := &model.Institution{ID: instID, LogoURL: "https://cdn.example.com/logo.png"}

	uc.On("UploadImage", instID, instID, "logo", mock.AnythingOfType("[]uint8"), mock.AnythingOfType("string")).
		Return(expected, nil)

	req, ct := newMultipartRequest(t, http.MethodPatch, "/api/v1/institutions/"+instID.String()+"/images?type=logo", imgData)
	req.Header.Set("Content-Type", ct)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertExpectations(t)
}

func TestHandlerUploadImage_InvalidType(t *testing.T) {
	uc := &mockInstitutionUseCase{}
	instID := uuid.New()
	r := setupRouter(uc, "auth0|x", instID)

	req, ct := newMultipartRequest(t, http.MethodPatch, "/api/v1/institutions/"+instID.String()+"/images?type=banner", []byte("data"))
	req.Header.Set("Content-Type", ct)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandlerUploadImage_Forbidden(t *testing.T) {
	uc := &mockInstitutionUseCase{}
	instID := uuid.New()
	otherInstID := uuid.New()
	r := setupRouter(uc, "auth0|x", instID)

	uc.On("UploadImage", otherInstID, instID, "logo", mock.AnythingOfType("[]uint8"), mock.AnythingOfType("string")).
		Return(nil, model.ErrUnauthorized)

	req, ct := newMultipartRequest(t, http.MethodPatch, "/api/v1/institutions/"+otherInstID.String()+"/images?type=logo", []byte("data"))
	req.Header.Set("Content-Type", ct)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	uc.AssertExpectations(t)
}
