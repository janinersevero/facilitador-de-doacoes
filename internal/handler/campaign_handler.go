package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"facilitador-de-doacoes/internal/model"
	"facilitador-de-doacoes/internal/repository"
	"facilitador-de-doacoes/internal/usecase"
)

type CampaignHandler struct {
	uc usecase.CampaignUseCase
}

func NewCampaignHandler(uc usecase.CampaignUseCase) *CampaignHandler {
	return &CampaignHandler{uc: uc}
}

// RegisterRoutes sets up campaign routes.
// Public: GET /campaigns, GET /campaigns/:id, GET /institutions/:institution_id/campaigns
// Protected: POST /institutions/:institution_id/campaigns, PUT/DELETE/PATCH /campaigns/:id
func (h *CampaignHandler) RegisterRoutes(r *gin.RouterGroup, authMiddlewares ...gin.HandlerFunc) {
	// Public read-only routes
	r.GET("/campaigns", h.GetAll)
	r.GET("/campaigns/:id", h.GetByID)
	r.GET("/institutions/:institution_id/campaigns", h.GetByInstitution)

	// Protected write routes
	protected := r.Group("", authMiddlewares...)
	protected.POST("/institutions/:institution_id/campaigns", h.Create)
	protected.PUT("/campaigns/:id", h.Update)
	protected.DELETE("/campaigns/:id", h.Delete)
	protected.PATCH("/campaigns/:id/status", h.UpdateStatus)
}

func (h *CampaignHandler) Create(c *gin.Context) {
	institutionID, err := uuid.Parse(c.Param("institution_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid institution_id"})
		return
	}

	userID, ok := c.Get(userIDKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var input usecase.CreateCampaignInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	campaign, err := h.uc.Create(userID.(uuid.UUID), institutionID, input)
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

	c.JSON(http.StatusCreated, campaign)
}

func (h *CampaignHandler) GetAll(c *gin.Context) {
	filters := repository.CampaignFilters{
		Keyword: c.Query("keyword"),
	}

	if urgentStr := c.Query("is_urgent"); urgentStr == "true" {
		v := true
		filters.IsUrgent = &v
	} else if urgentStr == "false" {
		v := false
		filters.IsUrgent = &v
	}

	campaigns, err := h.uc.GetAll(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, campaigns)
}

func (h *CampaignHandler) GetByInstitution(c *gin.Context) {
	institutionID, err := uuid.Parse(c.Param("institution_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid institution_id"})
		return
	}

	campaigns, err := h.uc.GetByInstitutionID(institutionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, campaigns)
}

func (h *CampaignHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	campaign, err := h.uc.GetByID(id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "campaign not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, campaign)
}

func (h *CampaignHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	userID, ok := c.Get(userIDKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var input usecase.UpdateCampaignInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	campaign, err := h.uc.Update(id, userID.(uuid.UUID), input)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "campaign not found"})
			return
		}
		if errors.Is(err, model.ErrUnauthorized) {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, campaign)
}

func (h *CampaignHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	userID, ok := c.Get(userIDKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if err := h.uc.Delete(id, userID.(uuid.UUID)); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "campaign not found"})
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

func (h *CampaignHandler) UpdateStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	userID, ok := c.Get(userIDKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var body struct {
		Status model.CampaignStatus `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	campaign, err := h.uc.UpdateStatus(id, userID.(uuid.UUID), body.Status)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "campaign not found"})
			return
		}
		if errors.Is(err, model.ErrUnauthorized) {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, campaign)
}
