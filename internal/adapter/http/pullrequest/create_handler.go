package pullrequest

import (
	"net/http"
	"time"

	common "github.com/Skorpsrgvch/reviewer-service/internal/adapter/http/common"
	prCreate "github.com/Skorpsrgvch/reviewer-service/internal/usecase/pullrequest/create"
	"github.com/gin-gonic/gin"
)

type createPRRequest struct {
	PullRequestID   string `json:"pull_request_id" binding:"required"`
	PullRequestName string `json:"pull_request_name" binding:"required"`
	AuthorID        string `json:"author_id" binding:"required"`
}

type createPRResponse struct {
	PR pullRequestDTO `json:"pr"`
}

type pullRequestDTO struct {
	PullRequestID     string   `json:"pull_request_id"`
	PullRequestName   string   `json:"pull_request_name"`
	AuthorID          string   `json:"author_id"`
	Status            string   `json:"status"`
	AssignedReviewers []string `json:"assigned_reviewers"`
	CreatedAt         string   `json:"created_at"`
	MergedAt          *string  `json:"mergedAt,omitempty"` // nullable → *string
}

type CreateHandler struct {
	usecase *prCreate.Usecase
}

func NewCreateHandler(usecase *prCreate.Usecase) *CreateHandler {
	return &CreateHandler{usecase: usecase}
}

func (h *CreateHandler) Handle(c *gin.Context) {
	var req createPRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.HandleError(c, err)
		return
	}

	input := prCreate.Input{
		PullRequestID:   req.PullRequestID,
		PullRequestName: req.PullRequestName,
		AuthorID:        req.AuthorID,
	}

	createdPR, err := h.usecase.Execute(c.Request.Context(), input)
	if err != nil {
		common.HandleError(c, err)
		return
	}

	resp := createPRResponse{
		PR: pullRequestDTO{
			PullRequestID:     createdPR.ID(),
			PullRequestName:   createdPR.Name(),
			AuthorID:          createdPR.AuthorID(),
			Status:            string(createdPR.Status()),
			AssignedReviewers: createdPR.AssignedReviewers(),
			CreatedAt:         createdPR.CreatedAt().Format(time.RFC3339),
		},
	}

	c.JSON(http.StatusCreated, resp)
}
