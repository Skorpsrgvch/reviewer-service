package create

import (
	"context"
	"errors"

	"github.com/Skorpsrgvch/reviewer-service/internal/domain"
)

type Input struct {
	TeamName string
	Members  []domain.User
}

type Usecase struct {
	teamSaver TeamSaver
}

func NewUsecase(teamSaver TeamSaver) (*Usecase, error) {
	if teamSaver == nil {
		return nil, errors.New("teamSaver is required")
	}
	return &Usecase{teamSaver: teamSaver}, nil
}

func (u *Usecase) Execute(ctx context.Context, input Input) error {
	team, err := domain.NewTeam(input.TeamName, input.Members)
	if err != nil {
		return err
	}
	return u.teamSaver.SaveTeam(ctx, team)
}
