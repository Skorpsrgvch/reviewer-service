package pullrequest

import (
	"net/http"
	"time"

	common "github.com/Skorpsrgvch/reviewer-service/internal/adapter/http/common"
	prReassign "github.com/Skorpsrgvch/reviewer-service/internal/usecase/pullrequest/reassign"
	"github.com/gin-gonic/gin"
)

type reassignPRRequest struct {
	PullRequestID string `json:"pull_request_id" binding:"required"`
	OldReviewerID string `json:"old_reviewer_id" binding:"required"`
}

type reassignPRResponse struct {
	PR         pullRequestDTO `json:"pr"`
	ReplacedBy string         `json:"replaced_by"`
}

type ReassignHandler struct {
	usecase *prReassign.Usecase
}

func NewReassignHandler(usecase *prReassign.Usecase) *ReassignHandler {
	return &ReassignHandler{usecase: usecase}
}

func (h *ReassignHandler) Handle(c *gin.Context) {
	var req reassignPRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.HandleError(c, err)
		return
	}

	input := prReassign.Input{
		PullRequestID: req.PullRequestID,
		OldReviewerID: req.OldReviewerID,
	}

	pr, newReviewer, err := h.usecase.Execute(c.Request.Context(), input)
	if err != nil {
		common.HandleError(c, err)
		return
	}

	// Сериализуем полный PR
	createdAt := pr.CreatedAt().Format(time.RFC3339)
	var mergedAt *string
	if pr.MergedAt() != nil {
		s := pr.MergedAt().Format(time.RFC3339)
		mergedAt = &s
	}

	resp := reassignPRResponse{
		PR: pullRequestDTO{
			PullRequestID:     pr.ID(),
			PullRequestName:   pr.Name(),
			AuthorID:          pr.AuthorID(),
			Status:            string(pr.Status()),
			AssignedReviewers: pr.AssignedReviewers(),
			CreatedAt:         createdAt,
			MergedAt:          mergedAt,
		},
		ReplacedBy: newReviewer,
	}

	c.JSON(http.StatusOK, resp)
}
