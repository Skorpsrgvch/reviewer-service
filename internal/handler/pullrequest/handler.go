package pullrequest

import (
	"context"
	"net/http"
	"time"

	"github.com/Skorpsrgvch/reviewer-service/internal/model"
	"github.com/gin-gonic/gin"
)

type Service interface {
	CreatePR(ctx context.Context, prID, prName, authorID string) (*model.PullRequest, error)
	MergePR(ctx context.Context, prID string) (*model.PullRequest, error)
	ReassignReviewer(ctx context.Context, prID, oldReviewerID string) (*model.PullRequest, string, error)
}

type Handler struct {
	service Service
}

func NewHandler(s Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) Create(c *gin.Context) {
	if c.GetHeader("Authorization") != "Bearer admin" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{"code": "UNAUTHORIZED", "message": "admin token required"},
		})
		return
	}

	var req CreatePRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "INVALID_JSON", "message": "invalid JSON"},
		})
		return
	}

	pr, err := h.service.CreatePR(c.Request.Context(), req.PullRequestID, req.PullRequestName, req.AuthorID)
	if err != nil {
		switch {
		case err.Error() == "PR already exists":
			c.AbortWithStatusJSON(http.StatusConflict, gin.H{
				"error": gin.H{"code": "PR_EXISTS", "message": "PR id already exists"},
			})
		case err.Error() == "author not found":
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"error": gin.H{"code": "NOT_FOUND", "message": "author or team not found"},
			})
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{"code": "INTERNAL", "message": err.Error()},
			})
		}
		return
	}

	c.JSON(http.StatusCreated, CreatePRResponse{PR: toPullRequestDTO(pr)})
}

func (h *Handler) Merge(c *gin.Context) {
	if c.GetHeader("Authorization") != "Bearer admin" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{"code": "UNAUTHORIZED", "message": "admin token required"},
		})
		return
	}

	var req MergePRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "INVALID_JSON", "message": "invalid JSON"},
		})
		return
	}

	pr, err := h.service.MergePR(c.Request.Context(), req.PullRequestID)
	if err != nil {
		if err.Error() == "PR not found" {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"error": gin.H{"code": "NOT_FOUND", "message": "PR not found"},
			})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL", "message": err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, MergePRResponse{PR: toPullRequestDTO(pr)})
}

func (h *Handler) Reassign(c *gin.Context) {
	if c.GetHeader("Authorization") != "Bearer admin" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{"code": "UNAUTHORIZED", "message": "admin token required"},
		})
		return
	}

	var req ReassignPRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "INVALID_JSON", "message": "invalid JSON"},
		})
		return
	}

	pr, newReviewer, err := h.service.ReassignReviewer(c.Request.Context(), req.PullRequestID, req.OldReviewerID)
	if err != nil {
		switch {
		case err.Error() == "PR is already merged":
			c.AbortWithStatusJSON(http.StatusConflict, gin.H{
				"error": gin.H{"code": "PR_MERGED", "message": "cannot reassign on merged PR"},
			})
		case err.Error() == "reviewer is not assigned to this PR":
			c.AbortWithStatusJSON(http.StatusConflict, gin.H{
				"error": gin.H{"code": "NOT_ASSIGNED", "message": "reviewer is not assigned to this PR"},
			})
		case err.Error() == "no active replacement candidate in team":
			c.AbortWithStatusJSON(http.StatusConflict, gin.H{
				"error": gin.H{"code": "NO_CANDIDATE", "message": "no active replacement candidate in team"},
			})
		case err.Error() == "PR not found" || err.Error() == "user not found":
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"error": gin.H{"code": "NOT_FOUND", "message": err.Error()},
			})
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{"code": "INTERNAL", "message": err.Error()},
			})
		}
		return
	}

	c.JSON(http.StatusOK, ReassignPRResponse{
		PR:         toPullRequestDTO(pr),
		ReplacedBy: newReviewer,
	})
}

func toPullRequestDTO(pr *model.PullRequest) PullRequestDTO {
	createdAt := pr.CreatedAt.Format(time.RFC3339)
	var mergedAt *string
	if pr.MergedAt != nil {
		s := pr.MergedAt.Format(time.RFC3339)
		mergedAt = &s
	}
	return PullRequestDTO{
		PullRequestID:     pr.ID,
		PullRequestName:   pr.Name,
		AuthorID:          pr.AuthorID,
		Status:            pr.Status,
		AssignedReviewers: pr.AssignedReviewers,
		CreatedAt:         createdAt,
		MergedAt:          mergedAt,
	}
}
