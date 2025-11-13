package stats

import (
	"context"
)

type PullRequestRepository interface {
	GetReviewerStats(ctx context.Context) (map[string]int, error)
}

type Service struct {
	prRepo PullRequestRepository
}

func NewService(prRepo PullRequestRepository) *Service {
	return &Service{prRepo: prRepo}
}

func (s *Service) GetReviewerStats(ctx context.Context) (map[string]int, error) {
	return s.prRepo.GetReviewerStats(ctx)
}
