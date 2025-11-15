package team

import (
	"net/http"

	common "github.com/Skorpsrgvch/reviewer-service/internal/adapter/http/common"
	teamGet "github.com/Skorpsrgvch/reviewer-service/internal/usecase/team/get"
	"github.com/gin-gonic/gin"
)

type getTeamResponse struct {
	Team teamDTO `json:"team"`
}

type GetHandler struct {
	usecase *teamGet.Usecase
}

func NewGetHandler(usecase *teamGet.Usecase) *GetHandler {
	return &GetHandler{usecase: usecase}
}

func (h *GetHandler) Handle(c *gin.Context) {
	teamName := c.Query("team_name")
	if teamName == "" {
		common.HandleError(c, httpError("team_name is required", http.StatusBadRequest))
		return
	}

	input := teamGet.Input{TeamName: teamName}
	team, err := h.usecase.Execute(c.Request.Context(), input)
	if err != nil {
		common.HandleError(c, err)
		return
	}

	// Преобразуем domain.Team → DTO
	var members []userDTO
	for _, u := range team.Members() {
		members = append(members, userDTO{
			UserID:   u.ID(),
			Username: u.Username(),
			IsActive: u.IsActive(),
		})
	}

	resp := getTeamResponse{
		Team: teamDTO{
			TeamName: team.Name(),
			Members:  members,
		},
	}

	c.JSON(http.StatusOK, resp)
}

// Вспомогательная функция для кастомных ошибок
func httpError(msg string, status int) error {
	return &customHTTPError{msg: msg, status: status}
}

type customHTTPError struct {
	msg    string
	status int
}

func (e *customHTTPError) Error() string { return e.msg }
func (e *customHTTPError) Status() int   { return e.status }
