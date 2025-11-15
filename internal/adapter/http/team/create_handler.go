package team

import (
	"net/http"

	common "github.com/Skorpsrgvch/reviewer-service/internal/adapter/http/common"
	"github.com/Skorpsrgvch/reviewer-service/internal/domain"
	teamCreate "github.com/Skorpsrgvch/reviewer-service/internal/usecase/team/create"
	"github.com/gin-gonic/gin"
)

// createTeamRequest — DTO для входящего JSON
type createTeamRequest struct {
	TeamName string    `json:"team_name" binding:"required"`
	Members  []userDTO `json:"members" binding:"required,dive"`
}

type userDTO struct {
	UserID   string `json:"user_id" binding:"required"`
	Username string `json:"username" binding:"required"`
	IsActive bool   `json:"is_active"`
}

// createTeamResponse — DTO для ответа
type createTeamResponse struct {
	Team teamDTO `json:"team"`
}

type teamDTO struct {
	TeamName string    `json:"team_name"`
	Members  []userDTO `json:"members"`
}

type CreateHandler struct {
	usecase *teamCreate.Usecase
}

func NewCreateHandler(usecase *teamCreate.Usecase) *CreateHandler {
	return &CreateHandler{usecase: usecase}
}

func (h *CreateHandler) Handle(c *gin.Context) {
	var req createTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.HandleError(c, err)
		return
	}

	// Преобразуем в domain.User
	var members []domain.User
	for _, m := range req.Members {
		user, err := domain.NewUser(m.UserID, m.Username, m.IsActive)
		if err != nil {
			common.HandleError(c, err)
			return
		}
		members = append(members, *user)
	}

	input := teamCreate.Input{
		TeamName: req.TeamName,
		Members:  members,
	}

	if err := h.usecase.Execute(c.Request.Context(), input); err != nil {
		common.HandleError(c, err)
		return
	}

	// Формируем ответ
	resp := createTeamResponse{
		Team: teamDTO{
			TeamName: req.TeamName,
			Members:  req.Members,
		},
	}

	c.JSON(http.StatusCreated, resp)
}
