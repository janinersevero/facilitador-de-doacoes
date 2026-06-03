package handler_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"facilitador-de-doacoes/internal/handler"
	"facilitador-de-doacoes/internal/model"
)

type mockRankingUseCase struct {
	mock.Mock
}

func (m *mockRankingUseCase) GetRanking(limit int) ([]*model.RankingEntry, error) {
	args := m.Called(limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.RankingEntry), args.Error(1)
}

func setupRankingRouter(uc *mockRankingUseCase) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := handler.NewRankingHandler(uc)
	h.RegisterRoutes(r.Group("/api/v1"))
	return r
}

func TestGetRanking_Handler_Success(t *testing.T) {
	uc := new(mockRankingUseCase)

	expected := []*model.RankingEntry{
		{Position: 1, UserID: uuid.New(), UserName: "Carlos", TotalDonated: 420000, DonationCount: 5},
		{Position: 2, UserID: uuid.New(), UserName: "Fernanda", TotalDonated: 380000, DonationCount: 3},
	}

	uc.On("GetRanking", 10).Return(expected, nil)

	router := setupRankingRouter(uc)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/ranking", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result []*model.RankingEntry
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "Carlos", result[0].UserName)
	assert.Equal(t, int64(420000), result[0].TotalDonated)
	uc.AssertExpectations(t)
}

func TestGetRanking_Handler_WithCustomLimit(t *testing.T) {
	uc := new(mockRankingUseCase)

	uc.On("GetRanking", 5).Return([]*model.RankingEntry{}, nil)

	router := setupRankingRouter(uc)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/ranking?limit=5", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertExpectations(t)
}

func TestGetRanking_Handler_InvalidLimit_UsesDefault(t *testing.T) {
	uc := new(mockRankingUseCase)

	uc.On("GetRanking", 10).Return([]*model.RankingEntry{}, nil)

	router := setupRankingRouter(uc)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/ranking?limit=abc", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertExpectations(t)
}

func TestGetRanking_Handler_Error(t *testing.T) {
	uc := new(mockRankingUseCase)

	uc.On("GetRanking", 10).Return(nil, errors.New("database error"))

	router := setupRankingRouter(uc)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/ranking", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var body map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(t, err)
	assert.Equal(t, "database error", body["error"])
	uc.AssertExpectations(t)
}
