package getReview

import (
	"context"

	"github.com/Skorpsrgvch/reviewer-service/internal/domain"
)

type UserFinder interface {
	GetUserByID(ctx context.Context, id string) (*domain.User, error)
}

type PullRequestRepository interface {
	GetByReviewer(ctx context.Context, reviewerID string) ([]domain.PullRequest, error)
}
