package create

import (
	"context"

	"github.com/Skorpsrgvch/reviewer-service/internal/domain"
)

type PullRequestSaver interface {
	Save(ctx context.Context, pr *domain.PullRequest) error
	PRExists(ctx context.Context, id string) (bool, error)
}

type TeamFinder interface {
	GetUsersInTeam(ctx context.Context, teamName string, onlyActive bool) ([]domain.User, error)
}

type UserFinder interface {
	GetTeamByUser(ctx context.Context, userID string) (string, error)
}
