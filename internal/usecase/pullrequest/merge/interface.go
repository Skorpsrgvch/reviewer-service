package merge

import (
	"context"
	"time"

	"github.com/Skorpsrgvch/reviewer-service/internal/domain"
)

type PullRequestFinder interface {
	GetByID(ctx context.Context, id string) (*domain.PullRequest, error)
}

type PullRequestMerger interface {
	Merge(ctx context.Context, id string, mergedAt time.Time) error
}
