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
	"facilitador-de-doacoes/pkg/asaas"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	donation, err := h.uc.GetByID(id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "donation not found"})
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

	token := os.Getenv("ASAAS_WEBHOOK_TOKEN")
	if token != "" {
		if !asaas.ValidateWebhookToken(c.GetHeader("asaas-access-token"), token) {
			c.Status(http.StatusUnauthorized)
			return
		}
	}

	var event asaas.WebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	switch event.Event {
	case "PAYMENT_RECEIVED", "PAYMENT_CONFIRMED":
		status := mapAsaasStatus(event.Payment.Status)
		if err := h.uc.HandleWebhook(event.Payment.ID, status); err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
	}

	c.Status(http.StatusOK)
}

// mapAsaasStatus normalizes ASAAS payment statuses to internal statuses.
func mapAsaasStatus(asaasStatus string) string {
	switch asaasStatus {
	case "RECEIVED", "CONFIRMED":
		return "PAID"
	case "OVERDUE":
		return "OVERDUE"
	case "REFUNDED", "REFUND_REQUESTED":
		return "REFUNDED"
	default:
		return asaasStatus
	}
}
