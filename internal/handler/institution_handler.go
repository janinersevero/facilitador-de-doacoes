package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"facilitador-de-doacoes/internal/model"
	"facilitador-de-doacoes/internal/usecase"
)

const userIDKey = "userID"
const institutionIDKey = "institutionID"
const auth0SubKey = "auth0_sub"

type InstitutionHandler struct {
	uc usecase.InstitutionUseCase
}

func NewInstitutionHandler(uc usecase.InstitutionUseCase) *InstitutionHandler {
	return &InstitutionHandler{uc: uc}
}

// RegisterRoutes registers public and protected routes.
// authOnly middlewares are applied to POST and PATCH /status.
// institutionAuth middlewares are applied to PUT and DELETE (requires institution identity).
func (h *InstitutionHandler) RegisterRoutes(r *gin.RouterGroup, authOnly []gin.HandlerFunc, institutionAuth []gin.HandlerFunc) {
	g := r.Group("/institutions")

	g.GET("", h.GetAll)
	g.GET("/:id", h.GetByID)

	authGroup := g.Group("", authOnly...)
	authGroup.POST("", h.Create)
	authGroup.PATCH("/:id/status", h.UpdateStatus)

	instGroup := g.Group("", institutionAuth...)
	instGroup.GET("/me", h.GetMe)
	instGroup.PUT("/:id", h.Update)
	instGroup.DELETE("/:id", h.Delete)
}

func (h *InstitutionHandler) Create(c *gin.Context) {
	auth0Sub, ok := c.Get(auth0SubKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var input usecase.CreateInstitutionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	institution, err := h.uc.Create(auth0Sub.(string), input)
	if err != nil {
		if errors.Is(err, model.ErrCNPJAlreadyInUse) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, institution)
}

func (h *InstitutionHandler) GetAll(c *gin.Context) {
	institutions, err := h.uc.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, institutions)
}

func (h *InstitutionHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	institution, err := h.uc.GetByID(id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "institution not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, institution)
}

func (h *InstitutionHandler) GetMe(c *gin.Context) {
	institutionID, ok := c.Get(institutionIDKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	institution, err := h.uc.GetByID(institutionID.(uuid.UUID))
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "institution not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, institution)
}

func (h *InstitutionHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	institutionID, ok := c.Get(institutionIDKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var input usecase.UpdateInstitutionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	institution, err := h.uc.Update(id, institutionID.(uuid.UUID), input)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "institution not found"})
			return
		}
		if errors.Is(err, model.ErrUnauthorized) {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, institution)
}

func (h *InstitutionHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	institutionID, ok := c.Get(institutionIDKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if err := h.uc.Delete(id, institutionID.(uuid.UUID)); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "institution not found"})
			return
		}
		if errors.Is(err, model.ErrUnauthorized) {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *InstitutionHandler) UpdateStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var input usecase.UpdateInstitutionStatusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	institution, err := h.uc.UpdateStatus(id, input)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "institution not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, institution)
}
