package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"facilitador-de-doacoes/internal/usecase"
)

type TransferHandler struct {
	uc usecase.TransferUseCase
}

func NewTransferHandler(uc usecase.TransferUseCase) *TransferHandler {
	return &TransferHandler{uc: uc}
}

func (h *TransferHandler) RegisterRoutes(r *gin.RouterGroup, middleware ...gin.HandlerFunc) {
	transfers := r.Group("/transfers", middleware...)
	transfers.POST("", h.Create)
}

func (h *TransferHandler) Create(c *gin.Context) {
	var input usecase.CreateTransferInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.uc.Create(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, result)
}
