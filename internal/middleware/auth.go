package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"facilitador-de-doacoes/internal/model"
)

// JWTValidator validates a JWT token and returns the Auth0 subject (sub) claim.
type JWTValidator interface {
	ValidateToken(ctx context.Context, tokenString string) (string, error)
}

// Auth0UserLookup is the subset of usecase.UserUseCase needed by RequireUser.
// Using a minimal interface avoids a circular import and keeps the middleware testable.
type Auth0UserLookup interface {
	GetByAuth0ID(auth0ID string) (*model.User, error)
}

// AuthMiddleware validates the Bearer JWT and sets "auth0_sub" in the Gin context.
func AuthMiddleware(v JWTValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearerToken(c.Request)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization token"})
			return
		}

		sub, err := v.ValidateToken(c.Request.Context(), token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		c.Set("auth0_sub", sub)
		c.Next()
	}
}

// RequireUser looks up the local user by the Auth0 sub set in context and sets "userID".
// Must be chained after AuthMiddleware.
func RequireUser(lookup Auth0UserLookup) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawSub, ok := c.Get("auth0_sub")
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing auth0 subject"})
			return
		}

		user, err := lookup.GetByAuth0ID(rawSub.(string))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user not linked to an auth0 account"})
			return
		}

		c.Set("userID", user.ID)
		c.Next()
	}
}

// Auth0InstitutionLookup is the subset of usecase.InstitutionUseCase needed by RequireInstitution.
type Auth0InstitutionLookup interface {
	GetByAuth0ID(auth0ID string) (*model.Institution, error)
}

// RequireInstitution looks up the local institution by the Auth0 sub set in context and sets "institutionID".
// Must be chained after AuthMiddleware.
func RequireInstitution(lookup Auth0InstitutionLookup) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawSub, ok := c.Get("auth0_sub")
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing auth0 subject"})
			return
		}
		institution, err := lookup.GetByAuth0ID(rawSub.(string))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "institution not found"})
			return
		}
		c.Set("institutionID", institution.ID)
		c.Next()
	}
}

// auth0Validator wraps the Auth0 JWT validator to implement JWTValidator.
type auth0Validator struct {
	v *validator.Validator
}

// NewAuth0Validator creates a JWTValidator backed by Auth0's JWKS endpoint.
func NewAuth0Validator(domain, audience string) (JWTValidator, error) {
	issuerURL, err := url.Parse("https://" + domain + "/")
	if err != nil {
		return nil, fmt.Errorf("parse auth0 issuer url: %w", err)
	}

	provider := jwks.NewCachingProvider(issuerURL, 5*time.Minute)

	v, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		issuerURL.String(),
		[]string{audience},
	)
	if err != nil {
		return nil, fmt.Errorf("create jwt validator: %w", err)
	}

	return &auth0Validator{v: v}, nil
}

func (a *auth0Validator) ValidateToken(ctx context.Context, tokenString string) (string, error) {
	claims, err := a.v.ValidateToken(ctx, tokenString)
	if err != nil {
		return "", err
	}

	validated, ok := claims.(*validator.ValidatedClaims)
	if !ok {
		return "", errors.New("unexpected claims type")
	}

	return validated.RegisteredClaims.Subject, nil
}

func extractBearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if len(auth) > 7 && auth[:7] == "Bearer " {
		return auth[7:]
	}
	return ""
}

// userIDFromContext is a helper used by handlers to read the userID set by RequireUser.
func userIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	val, ok := c.Get("userID")
	if !ok {
		return uuid.Nil, false
	}
	id, ok := val.(uuid.UUID)
	return id, ok
}
