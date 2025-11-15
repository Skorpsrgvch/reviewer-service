package user

import (
	"net/http"

	common "github.com/Skorpsrgvch/reviewer-service/internal/adapter/http/common"
	userSetActive "github.com/Skorpsrgvch/reviewer-service/internal/usecase/user/setActive"
	"github.com/gin-gonic/gin"
)

type setIsActiveRequest struct {
	UserID   string `json:"user_id" binding:"required"`
	IsActive bool   `json:"is_active"`
}

type setIsActiveResponse struct {
	User userDTO `json:"user"`
}

type userDTO struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
	TeamName string `json:"team_name"`
}

type SetActiveHandler struct {
	usecase *userSetActive.Usecase
}

func NewSetActiveHandler(usecase *userSetActive.Usecase) *SetActiveHandler {
	return &SetActiveHandler{usecase: usecase}
}

func (h *SetActiveHandler) Handle(c *gin.Context) {
	var req setIsActiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.HandleError(c, err)
		return
	}

	input := userSetActive.Input{
		UserID:   req.UserID,
		IsActive: req.IsActive,
	}

	output, err := h.usecase.Execute(c.Request.Context(), input)
	if err != nil {
		common.HandleError(c, err)
		return
	}

	userDomain := output.User
	resp := setIsActiveResponse{
		User: userDTO{
			UserID:   userDomain.ID(),
			Username: userDomain.Username(),
			IsActive: userDomain.IsActive(),
			TeamName: output.TeamName,
		},
	}

	c.JSON(http.StatusOK, resp)
}
