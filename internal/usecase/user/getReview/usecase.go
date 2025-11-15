package getReview

import (
	"context"
	"errors"

	"github.com/Skorpsrgvch/reviewer-service/internal/domain"
)

type Input struct {
	UserID string
}

type Output struct {
	UserID       string
	PullRequests []domain.PullRequest
}

type Usecase struct {
	prRepo     PullRequestRepository
	userFinder UserFinder
}

func NewUsecase(prRepo PullRequestRepository, userFinder UserFinder) (*Usecase, error) {
	if prRepo == nil || userFinder == nil {
		return nil, errors.New("dependencies required")
	}
	return &Usecase{prRepo: prRepo, userFinder: userFinder}, nil
}

func (u *Usecase) Execute(ctx context.Context, input Input) (*Output, error) {
	_, err := u.userFinder.GetUserByID(ctx, input.UserID)
	if err != nil {
		return nil, err
	}

	prs, err := u.prRepo.GetByReviewer(ctx, input.UserID)
	if err != nil {
		return nil, err
	}

	return &Output{
		UserID:       input.UserID,
		PullRequests: prs,
	}, nil
}
