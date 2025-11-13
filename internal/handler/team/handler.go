package team

import (
	"context"
	"net/http"

	"github.com/Skorpsrgvch/reviewer-service/internal/model"
	"github.com/gin-gonic/gin"
)

type Service interface {
	CreateTeam(ctx context.Context, teamName string, members []model.User) (*model.Team, error)
	GetTeamByName(ctx context.Context, teamName string) (*model.Team, error)
}

type Handler struct {
	service Service
}

func NewHandler(s Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) AddTeam(c *gin.Context) {
	if c.GetHeader("Authorization") != "Bearer admin" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{"code": "UNAUTHORIZED", "message": "admin token required"},
		})
		return
	}

	var req CreateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "INVALID_JSON", "message": "invalid JSON"},
		})
		return
	}

	var members []model.User
	for _, m := range req.Members {
		members = append(members, model.User{
			ID:       m.UserID,
			Username: m.Username,
			// team_name будет установлен внутри сервиса
			IsActive: m.IsActive,
		})
	}

	team, err := h.service.CreateTeam(c.Request.Context(), req.TeamName, members)
	if err != nil {
		switch {
		case err.Error() == "team already exists":
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": gin.H{"code": "TEAM_EXISTS", "message": "team_name already exists"},
			})
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{"code": "INTERNAL", "message": err.Error()},
			})
		}
		return
	}

	response := CreateTeamResponse{
		Team: TeamDTO{
			TeamName: team.Name,
			Members:  toTeamMemberDTOs(team.Members),
		},
	}
	c.JSON(http.StatusCreated, response)
}

func (h *Handler) GetTeam(c *gin.Context) {
	teamName := c.Query("team_name")
	if teamName == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "INVALID_PARAM", "message": "team_name is required"},
		})
		return
	}

	team, err := h.service.GetTeamByName(c.Request.Context(), teamName)
	if err != nil {
		if err.Error() == "team not found" {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"error": gin.H{"code": "NOT_FOUND", "message": "team not found"},
			})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL", "message": err.Error()},
		})
		return
	}

	response := GetTeamResponse{
		TeamName: team.Name,
		Members:  toTeamMemberDTOs(team.Members),
	}
	c.JSON(http.StatusOK, response)
}

func toTeamMemberDTOs(users []model.User) []TeamMember {
	dtos := make([]TeamMember, len(users))
	for i, u := range users {
		dtos[i] = TeamMember{
			UserID:   u.ID,
			Username: u.Username,
			IsActive: u.IsActive,
		}
	}
	return dtos
}
