package setActive

import (
	"context"
	"errors"

	"github.com/Skorpsrgvch/reviewer-service/internal/domain"
)

type Input struct {
	UserID   string
	IsActive bool
}

type Output struct {
	User     domain.User
	TeamName string
}

type Usecase struct {
	userFinder  UserFinder
	userUpdater UserUpdater
}

func NewUsecase(userFinder UserFinder, userUpdater UserUpdater) (*Usecase, error) {
	if userFinder == nil || userUpdater == nil {
		return nil, errors.New("userFinder and userUpdater are required")
	}
	return &Usecase{userFinder: userFinder, userUpdater: userUpdater}, nil
}

func (u *Usecase) Execute(ctx context.Context, input Input) (*Output, error) {
	user, err := u.userFinder.GetUserByID(ctx, input.UserID)
	if err != nil {
		return nil, err
	}

	user.SetActive(input.IsActive)
	if err := u.userUpdater.UpdateUser(ctx, user); err != nil {
		return nil, err
	}

	teamName, err := u.userFinder.GetTeamByUser(ctx, input.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			teamName = ""
		} else {
			return nil, err
		}
	}

	return &Output{
		User:     *user,
		TeamName: teamName,
	}, nil
}
