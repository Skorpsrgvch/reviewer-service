package get

import "context"

type StatsReader interface {
	GetReviewerStats(ctx context.Context) (map[string]int, error)
}
