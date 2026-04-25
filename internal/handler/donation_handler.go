package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"facilitador-de-doacoes/internal/model"
	"facilitador-de-doacoes/internal/usecase"
	"facilitador-de-doacoes/pkg/abacatepay"
)

type DonationHandler struct {
	uc usecase.DonationUseCase
}

func NewDonationHandler(uc usecase.DonationUseCase) *DonationHandler {
	return &DonationHandler{uc: uc}
}

func (h *DonationHandler) RegisterRoutes(r *gin.RouterGroup) {
	donations := r.Group("/donations")
	donations.POST("", h.Create)
	donations.GET("", h.GetAll)
	donations.GET("/:id", h.GetByID)
	donations.POST("/webhook", h.Webhook)
}

func (h *DonationHandler) Create(c *gin.Context) {
	var input usecase.CreateDonationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	donation, err := h.uc.Create(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, donation)
}

func (h *DonationHandler) GetAll(c *gin.Context) {
	donations, err := h.uc.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, donations)
}

func (h *DonationHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}

	donation, err := h.uc.GetByID(id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "doação não encontrada"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, donation)
}

func (h *DonationHandler) Webhook(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	secret := os.Getenv("ABACATEPAY_WEBHOOK_SECRET")
	if secret != "" {
		sig := c.GetHeader("X-Webhook-Signature")
		if !abacatepay.ValidateSignature(secret, sig, body) {
			c.Status(http.StatusUnauthorized)
			return
		}
	}

	var event abacatepay.WebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	if event.EventType == "billing.paid" {
		b := event.Data.Billing
		if err := h.uc.HandleWebhook(b.ID, b.Status); err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
	}

	c.Status(http.StatusOK)
}
