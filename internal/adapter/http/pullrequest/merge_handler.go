package pullrequest

import (
	"net/http"
	"time"

	common "github.com/Skorpsrgvch/reviewer-service/internal/adapter/http/common"
	prMerge "github.com/Skorpsrgvch/reviewer-service/internal/usecase/pullrequest/merge"
	"github.com/gin-gonic/gin"
)

type mergePRRequest struct {
	PullRequestID string `json:"pull_request_id" binding:"required"`
}

type mergePRResponse struct {
	PR pullRequestDTO `json:"pr"`
}

type MergeHandler struct {
	usecase *prMerge.Usecase
}

func NewMergeHandler(usecase *prMerge.Usecase) *MergeHandler {
	return &MergeHandler{usecase: usecase}
}

func (h *MergeHandler) Handle(c *gin.Context) {
	var req mergePRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.HandleError(c, err)
		return
	}

	input := prMerge.Input{
		PullRequestID: req.PullRequestID,
	}

	pr, err := h.usecase.Execute(c.Request.Context(), input)
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

	resp := mergePRResponse{
		PR: pullRequestDTO{
			PullRequestID:     pr.ID(),
			PullRequestName:   pr.Name(),
			AuthorID:          pr.AuthorID(),
			Status:            string(pr.Status()),
			AssignedReviewers: pr.AssignedReviewers(),
			CreatedAt:         createdAt,
			MergedAt:          mergedAt,
		},
	}

	c.JSON(http.StatusOK, resp)
}
