package user

import (
	"context"
	"net/http"

	"github.com/Skorpsrgvch/reviewer-service/internal/model"
	"github.com/gin-gonic/gin"
)

type Service interface {
	SetUserActive(ctx context.Context, userID string, isActive bool) (*model.User, error)
	GetReviewPRs(ctx context.Context, userID string) ([]model.PullRequest, error)
}

type Handler struct {
	service Service
}

func NewHandler(s Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) SetIsActive(c *gin.Context) {
	if c.GetHeader("Authorization") != "Bearer admin" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{"code": "UNAUTHORIZED", "message": "admin token required"},
		})
		return
	}

	var req SetIsActiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "INVALID_JSON", "message": "invalid JSON"},
		})
		return
	}

	user, err := h.service.SetUserActive(c.Request.Context(), req.UserID, req.IsActive)
	if err != nil {
		if err.Error() == "user not found" {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"error": gin.H{"code": "NOT_FOUND", "message": "user not found"},
			})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL", "message": err.Error()},
		})
		return
	}

	response := SetIsActiveResponse{
		User: UserDTO{
			UserID:   user.ID,
			Username: user.Username,
			TeamName: user.TeamName,
			IsActive: user.IsActive,
		},
	}
	c.JSON(http.StatusOK, response)
}

func (h *Handler) GetReview(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "INVALID_PARAM", "message": "user_id is required"},
		})
		return
	}

	prs, err := h.service.GetReviewPRs(c.Request.Context(), userID)
	if err != nil {
		if err.Error() == "user not found" {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"error": gin.H{"code": "NOT_FOUND", "message": "user not found"},
			})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL", "message": err.Error()},
		})
		return
	}

	shortPRs := make([]PullRequestShort, len(prs))
	for i, pr := range prs {
		shortPRs[i] = PullRequestShort{
			PullRequestID:   pr.ID,
			PullRequestName: pr.Name,
			AuthorID:        pr.AuthorID,
			Status:          pr.Status,
		}
	}

	response := GetReviewResponse{
		UserID:       userID,
		PullRequests: shortPRs,
	}
	c.JSON(http.StatusOK, response)
}
