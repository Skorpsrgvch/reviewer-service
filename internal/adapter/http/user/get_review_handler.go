package user

import (
	"net/http"

	common "github.com/Skorpsrgvch/reviewer-service/internal/adapter/http/common"
	userGetReview "github.com/Skorpsrgvch/reviewer-service/internal/usecase/user/getReview"
	"github.com/gin-gonic/gin"
)

type getReviewResponse struct {
	UserID       string       `json:"user_id"`
	PullRequests []prShortDTO `json:"pull_requests"`
}

type prShortDTO struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
	Status          string `json:"status"`
}

type GetReviewHandler struct {
	usecase *userGetReview.Usecase
}

func NewGetReviewHandler(usecase *userGetReview.Usecase) *GetReviewHandler {
	return &GetReviewHandler{usecase: usecase}
}

func (h *GetReviewHandler) Handle(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		common.HandleError(c, common.HttpError("user_id is required", http.StatusBadRequest))
		return
	}

	input := userGetReview.Input{UserID: userID}
	output, err := h.usecase.Execute(c.Request.Context(), input)
	if err != nil {
		common.HandleError(c, err)
		return
	}

	var prs []prShortDTO
	for _, pr := range output.PullRequests {
		prs = append(prs, prShortDTO{
			PullRequestID:   pr.ID(),
			PullRequestName: pr.Name(),
			AuthorID:        pr.AuthorID(),
			Status:          string(pr.Status()),
		})
	}

	resp := getReviewResponse{
		UserID:       output.UserID,
		PullRequests: prs,
	}

	c.JSON(http.StatusOK, resp)
}
