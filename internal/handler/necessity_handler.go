package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"facilitador-de-doacoes/internal/model"
	"facilitador-de-doacoes/internal/usecase"
)

type NecessityHandler struct {
	uc usecase.NecessityUseCase
}

func NewNecessityHandler(uc usecase.NecessityUseCase) *NecessityHandler {
	return &NecessityHandler{uc: uc}
}

func (h *NecessityHandler) RegisterRoutes(r *gin.RouterGroup, authMiddlewares ...gin.HandlerFunc) {
	r.GET("/institutions/:id/necessities", h.GetByInstitution)
	r.GET("/necessities/:id", h.GetByID)

	protected := r.Group("", authMiddlewares...)
	protected.POST("/institutions/:id/necessities", h.Create)
	protected.PUT("/necessities/:id", h.Update)
	protected.DELETE("/necessities/:id", h.Delete)
	protected.PATCH("/necessities/:id/status", h.UpdateStatus)
}

func (h *NecessityHandler) Create(c *gin.Context) {
	pathInstID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid institution_id"})
		return
	}

	ctxInstID, ok := c.Get(institutionIDKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	institutionID := ctxInstID.(uuid.UUID)
	if pathInstID != institutionID {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	var input usecase.CreateNecessityInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	necessity, err := h.uc.Create(institutionID, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, necessity)
}

func (h *NecessityHandler) GetByInstitution(c *gin.Context) {
	institutionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid institution_id"})
		return
	}

	necessities, err := h.uc.GetByInstitutionID(institutionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, necessities)
}

func (h *NecessityHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	necessity, err := h.uc.GetByID(id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "necessity not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, necessity)
}

func (h *NecessityHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	ctxInstID, ok := c.Get(institutionIDKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var input usecase.UpdateNecessityInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	necessity, err := h.uc.Update(id, ctxInstID.(uuid.UUID), input)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "necessity not found"})
			return
		}
		if errors.Is(err, model.ErrUnauthorized) {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, necessity)
}

func (h *NecessityHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	ctxInstID, ok := c.Get(institutionIDKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if err := h.uc.Delete(id, ctxInstID.(uuid.UUID)); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "necessity not found"})
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

func (h *NecessityHandler) UpdateStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	ctxInstID, ok := c.Get(institutionIDKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var body struct {
		Status model.NecessityStatus `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	necessity, err := h.uc.UpdateStatus(id, ctxInstID.(uuid.UUID), body.Status)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "necessity not found"})
			return
		}
		if errors.Is(err, model.ErrUnauthorized) {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, necessity)
}
