package get

import (
	"context"
	"errors"
)

type Usecase struct {
	statsReader StatsReader
}

func NewUsecase(statsReader StatsReader) (*Usecase, error) {
	if statsReader == nil {
		return nil, errors.New("statsReader is required")
	}
	return &Usecase{statsReader: statsReader}, nil
}

func (u *Usecase) Execute(ctx context.Context) (map[string]int, error) {
	return u.statsReader.GetReviewerStats(ctx)
}
