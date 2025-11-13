package stats

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Service interface {
	GetReviewerStats(ctx context.Context) (map[string]int, error)
}

type Handler struct {
	service Service
}

func NewHandler(s Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) GetStats(c *gin.Context) {
	stats, err := h.service.GetReviewerStats(c.Request.Context())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to get stats"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"assignments": stats})
}
