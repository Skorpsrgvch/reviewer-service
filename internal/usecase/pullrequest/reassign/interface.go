package reassign

import (
	"context"

	"github.com/Skorpsrgvch/reviewer-service/internal/domain"
)

type PullRequestRepository interface {
	GetByID(ctx context.Context, id string) (*domain.PullRequest, error)
	UpdateReviewers(ctx context.Context, id string, reviewers []string) error
}

type TeamRepository interface {
	GetUsersInTeam(ctx context.Context, teamName string, onlyActive bool) ([]domain.User, error)
}

type UserRepository interface {
	GetTeamByUser(ctx context.Context, userID string) (string, error)
}
