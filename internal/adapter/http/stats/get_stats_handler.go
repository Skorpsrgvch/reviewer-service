package stats

import (
	"net/http"

	common "github.com/Skorpsrgvch/reviewer-service/internal/adapter/http/common"
	statsUC "github.com/Skorpsrgvch/reviewer-service/internal/usecase/stats/get"
	"github.com/gin-gonic/gin"
)

type getStatsResponse struct {
	Assignments map[string]int `json:"assignments"`
}

type GetHandler struct {
	usecase *statsUC.Usecase
}

func NewGetHandler(usecase *statsUC.Usecase) *GetHandler {
	return &GetHandler{usecase: usecase}
}

func (h *GetHandler) Handle(c *gin.Context) {
	stats, err := h.usecase.Execute(c.Request.Context())
	if err != nil {
		common.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, getStatsResponse{Assignments: stats})
}
