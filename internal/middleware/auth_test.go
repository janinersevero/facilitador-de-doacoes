package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"facilitador-de-doacoes/internal/middleware"
	"facilitador-de-doacoes/internal/model"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// mockJWTValidator implements middleware.JWTValidator for testing.
type mockJWTValidator struct {
	mock.Mock
}

func (m *mockJWTValidator) ValidateToken(ctx context.Context, tokenString string) (string, error) {
	args := m.Called(ctx, tokenString)
	return args.String(0), args.Error(1)
}

// mockUserUseCase implements the subset of usecase.UserUseCase needed by RequireUser.
type mockUserUseCase struct {
	mock.Mock
}

func (m *mockUserUseCase) GetByAuth0ID(auth0ID string) (*model.User, error) {
	args := m.Called(auth0ID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	v := &mockJWTValidator{}
	sub := "auth0|12345"

	v.On("ValidateToken", mock.Anything, "valid-token").Return(sub, nil)

	r := gin.New()
	r.GET("/test", middleware.AuthMiddleware(v), func(c *gin.Context) {
		got, _ := c.Get("auth0_sub")
		c.JSON(http.StatusOK, gin.H{"sub": got})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), sub)
	v.AssertExpectations(t)
}

func TestAuthMiddleware_MissingToken(t *testing.T) {
	v := &mockJWTValidator{}

	r := gin.New()
	r.GET("/test", middleware.AuthMiddleware(v), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	v.AssertNotCalled(t, "ValidateToken")
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	v := &mockJWTValidator{}

	v.On("ValidateToken", mock.Anything, "bad-token").Return("", errors.New("signature invalid"))

	r := gin.New()
	r.GET("/test", middleware.AuthMiddleware(v), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer bad-token")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	v.AssertExpectations(t)
}

func TestRequireUser_UserFound(t *testing.T) {
	userUC := &mockUserUseCase{}
	userID := uuid.New()
	sub := "auth0|12345"
	user := &model.User{ID: userID, Auth0ID: sub}

	userUC.On("GetByAuth0ID", sub).Return(user, nil)

	r := gin.New()
	r.GET("/test",
		func(c *gin.Context) { c.Set("auth0_sub", sub); c.Next() },
		middleware.RequireUser(userUC),
		func(c *gin.Context) {
			got, _ := c.Get("userID")
			c.JSON(http.StatusOK, gin.H{"user_id": got.(uuid.UUID).String()})
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), userID.String())
	userUC.AssertExpectations(t)
}

func TestRequireUser_UserNotFound(t *testing.T) {
	userUC := &mockUserUseCase{}
	sub := "auth0|ghost"

	userUC.On("GetByAuth0ID", sub).Return(nil, model.ErrNotFound)

	r := gin.New()
	r.GET("/test",
		func(c *gin.Context) { c.Set("auth0_sub", sub); c.Next() },
		middleware.RequireUser(userUC),
		func(c *gin.Context) { c.Status(http.StatusOK) },
	)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	userUC.AssertExpectations(t)
}

func TestRequireUser_MissingAuth0Sub(t *testing.T) {
	userUC := &mockUserUseCase{}

	r := gin.New()
	r.GET("/test",
		// No middleware that sets auth0_sub
		middleware.RequireUser(userUC),
		func(c *gin.Context) { c.Status(http.StatusOK) },
	)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
