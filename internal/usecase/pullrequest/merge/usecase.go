package merge

import (
	"context"
	"errors"
	"time"

	"github.com/Skorpsrgvch/reviewer-service/internal/domain"
)

type Input struct {
	PullRequestID string
}

type Usecase struct {
	prFinder PullRequestFinder
	prMerger PullRequestMerger
}

func NewUsecase(prFinder, prMerger interface{}) (*Usecase, error) {
	finder, ok1 := prFinder.(PullRequestFinder)
	merger, ok2 := prMerger.(PullRequestMerger)
	if !ok1 || !ok2 {
		return nil, errors.New("invalid dependencies")
	}
	return &Usecase{prFinder: finder, prMerger: merger}, nil
}

func (u *Usecase) Execute(ctx context.Context, input Input) (*domain.PullRequest, error) {
	pr, err := u.prFinder.GetByID(ctx, input.PullRequestID)
	if err != nil {
		return nil, err
	}

	if pr.Status() == domain.PRMerged {
		// Идемпотентность: возвращаем существующий PR
		return pr, nil
	}

	// Выполняем мерж в БД
	if err := u.prMerger.Merge(ctx, input.PullRequestID, time.Now().UTC()); err != nil {
		return nil, err
	}

	// Обновляем статус локально для возврата
	_ = pr.Merge() // безопасно, потому что только что смержили
	return pr, nil
}
