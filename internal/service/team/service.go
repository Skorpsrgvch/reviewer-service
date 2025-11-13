package team

import (
	"context"
	"errors"

	"github.com/Skorpsrgvch/reviewer-service/internal/model"
)

type UserRepository interface {
	GetUserByID(ctx context.Context, id string) (*model.User, error)
	CreateUser(ctx context.Context, u *model.User) error
	UpdateUser(ctx context.Context, u *model.User) error
	GetUsersByTeam(ctx context.Context, teamName string) ([]model.User, error)
}

type Service struct {
	userRepo UserRepository
}

func NewService(userRepo UserRepository) *Service {
	return &Service{userRepo: userRepo}
}

// CreateTeam принимает имя команды и список пользователей (model.User)
func (s *Service) CreateTeam(ctx context.Context, teamName string, members []model.User) (*model.Team, error) {
	// Проверяем, существует ли команда (через наличие пользователей)
	existingUsers, err := s.userRepo.GetUsersByTeam(ctx, teamName)
	if err == nil && len(existingUsers) > 0 {
		return nil, model.ErrTeamExists
	}
	// Если ошибка НЕ "not found" — пробрасываем
	if err != nil && !errors.Is(err, model.ErrTeamNotFound) {
		return nil, err
	}

	// Создаём/обновляем пользователей
	for _, user := range members {
		// Убеждаемся, что у пользователя правильное team_name
		userWithTeam := user
		userWithTeam.TeamName = teamName

		_, getUserErr := s.userRepo.GetUserByID(ctx, user.ID)
		if getUserErr == nil {
			// Пользователь существует — обновляем
			if err := s.userRepo.UpdateUser(ctx, &userWithTeam); err != nil {
				return nil, err
			}
		} else if errors.Is(getUserErr, model.ErrUserNotFound) {
			// Создаём нового
			if err := s.userRepo.CreateUser(ctx, &userWithTeam); err != nil {
				return nil, err
			}
		} else {
			return nil, getUserErr
		}
	}

	return &model.Team{
		Name:    teamName,
		Members: members,
	}, nil
}

func (s *Service) GetTeamByName(ctx context.Context, teamName string) (*model.Team, error) {
	users, err := s.userRepo.GetUsersByTeam(ctx, teamName)
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, model.ErrTeamNotFound
	}
	return &model.Team{
		Name:    teamName,
		Members: users,
	}, nil
}
