package get

import (
	"context"
	"errors"

	"github.com/Skorpsrgvch/reviewer-service/internal/domain"
)

// Input — входные данные юзкейса.
type Input struct {
	TeamName string
}

// Usecase реализует получение команды.
type Usecase struct {
	teamFinder TeamFinder
}

// NewUsecase создаёт новый юзкейс.
func NewUsecase(teamFinder TeamFinder) (*Usecase, error) {
	if teamFinder == nil {
		return nil, errors.New("teamFinder is required")
	}
	return &Usecase{teamFinder: teamFinder}, nil
}

// Execute выполняет получение команды по имени.
func (u *Usecase) Execute(ctx context.Context, input Input) (*domain.Team, error) {
	if input.TeamName == "" {
		return nil, errors.New("team name is required")
	}

	return u.teamFinder.FindTeamByName(ctx, input.TeamName)
}
