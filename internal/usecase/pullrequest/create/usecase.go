package create

import (
	"context"
	"errors"
	"math/rand"

	"github.com/Skorpsrgvch/reviewer-service/internal/domain"
)

type Input struct {
	PullRequestID   string
	PullRequestName string
	AuthorID        string
}

type Usecase struct {
	prSaver    PullRequestSaver
	userFinder UserFinder
	teamFinder TeamFinder
}

func NewUsecase(prSaver PullRequestSaver, userFinder UserFinder, teamFinder TeamFinder) (*Usecase, error) {
	if prSaver == nil || userFinder == nil || teamFinder == nil {
		return nil, errors.New("all dependencies are required")
	}
	return &Usecase{
		prSaver:    prSaver,
		userFinder: userFinder,
		teamFinder: teamFinder,
	}, nil
}

func (u *Usecase) Execute(ctx context.Context, input Input) (*domain.PullRequest, error) {
	// Проверка существования PR
	exists, err := u.prSaver.PRExists(ctx, input.PullRequestID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.ErrPRExists
	}

	// Получаем команду автора
	teamName, err := u.userFinder.GetTeamByUser(ctx, input.AuthorID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrAuthorNotFound
		}
		return nil, err
	}

	// Получаем активных участников команды (без автора)
	activeMembers, err := u.teamFinder.GetUsersInTeam(ctx, teamName, true)
	if err != nil {
		return nil, err
	}

	var candidates []string
	for _, user := range activeMembers {
		if user.ID() != input.AuthorID {
			candidates = append(candidates, user.ID())
		}
	}

	if len(candidates) == 0 {
		return nil, domain.ErrNoActiveReviewers
	}

	// Выбираем до 2 случайных
	rand.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})
	var assigned []string
	for i := 0; i < len(candidates) && i < 2; i++ {
		assigned = append(assigned, candidates[i])
	}

	pr, err := domain.NewPullRequest(input.PullRequestID, input.PullRequestName, input.AuthorID, assigned)
	if err != nil {
		return nil, err
	}

	if err := u.prSaver.Save(ctx, pr); err != nil {
		return nil, err
	}

	return pr, nil
}
