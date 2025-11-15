package common

import (
	"errors"
	"net/http"

	"github.com/Skorpsrgvch/reviewer-service/internal/domain"
	"github.com/gin-gonic/gin"
)

type httpError struct {
	msg    string
	status int
}

func (e httpError) Error() string { return e.msg }
func (e httpError) Status() int   { return e.status }

func HttpError(msg string, status int) error {
	return httpError{msg: msg, status: status}
}

// mapDomainErrorToCode сопоставляет ошибку domain с кодом API
func mapDomainErrorToCode(err error) (code string, httpStatus int, message string) {
	switch {
	case errors.Is(err, domain.ErrPRExists):
		return "PR_EXISTS", http.StatusConflict, "PR id already exists"
	case errors.Is(err, domain.ErrPRNotFound):
		return "NOT_FOUND", http.StatusNotFound, "pull request not found"
	case errors.Is(err, domain.ErrPRAlreadyMerged):
		return "PR_MERGED", http.StatusConflict, "cannot reassign on merged PR"
	case errors.Is(err, domain.ErrReviewerNotAssigned):
		return "NOT_ASSIGNED", http.StatusConflict, "reviewer is not assigned to this PR"
	case errors.Is(err, domain.ErrNoActiveReviewers):
		return "NO_CANDIDATE", http.StatusConflict, "no active replacement candidate in team"
	case errors.Is(err, domain.ErrAuthorNotFound), errors.Is(err, domain.ErrUserNotFound), errors.Is(err, domain.ErrTeamNotFound):
		return "NOT_FOUND", http.StatusNotFound, "author, user, team or PR not found"
	case errors.Is(err, domain.ErrTeamExists):
		return "TEAM_EXISTS", http.StatusConflict, "team_name already exists"
	default:
		return "INTERNAL", http.StatusInternalServerError, "internal server error"
	}
}

func HandleError(c *gin.Context, err error) {
	// Сначала проверяем, не является ли ошибка кастомной HTTP-ошибкой
	if httpErr, ok := err.(interface{ Status() int }); ok {
		c.AbortWithStatusJSON(httpErr.Status(), gin.H{
			"error": gin.H{
				"code":    "INVALID_PARAM", // или динамически из err
				"message": err.Error(),
			},
		})
		return
	}

	// Иначе — обрабатываем как domain-ошибку
	code, status, message := mapDomainErrorToCode(err)

	if code == "INTERNAL" {
		// TODO: логирование
	}

	c.AbortWithStatusJSON(status, gin.H{
		"error": gin.H{
			"code":    code,
			"message": message,
		},
	})
}
