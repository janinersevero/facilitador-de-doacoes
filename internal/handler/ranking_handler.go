package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"facilitador-de-doacoes/internal/usecase"
)

type RankingHandler struct {
	uc usecase.RankingUseCase
}

func NewRankingHandler(uc usecase.RankingUseCase) *RankingHandler {
	return &RankingHandler{uc: uc}
}

func (h *RankingHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/ranking", h.GetRanking)
}

func (h *RankingHandler) GetRanking(c *gin.Context) {
	limit := 10
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	ranking, err := h.uc.GetRanking(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ranking)
}
