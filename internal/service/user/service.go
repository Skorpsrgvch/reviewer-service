package user

import (
	"context"
	"errors"

	"github.com/Skorpsrgvch/reviewer-service/internal/model"
)

type UserRepository interface {
	GetUserByID(ctx context.Context, id string) (*model.User, error)
	UpdateUser(ctx context.Context, u *model.User) error
}

type PullRequestRepository interface {
	GetPRsByReviewer(ctx context.Context, reviewerID string) ([]model.PullRequest, error)
}

type Service struct {
	userRepo UserRepository
	prRepo   PullRequestRepository
}

func NewService(userRepo UserRepository, prRepo PullRequestRepository) *Service {
	return &Service{
		userRepo: userRepo,
		prRepo:   prRepo,
	}
}

func (s *Service) SetUserActive(ctx context.Context, userID string, isActive bool) (*model.User, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, model.ErrUserNotFound) {
			return nil, model.ErrUserNotFound
		}
		return nil, err
	}

	user.IsActive = isActive
	if err := s.userRepo.UpdateUser(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Service) GetReviewPRs(ctx context.Context, userID string) ([]model.PullRequest, error) {
	// Проверяем существование пользователя (даже неактивного)
	_, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, model.ErrUserNotFound) {
			return nil, model.ErrUserNotFound
		}
		return nil, err
	}

	return s.prRepo.GetPRsByReviewer(ctx, userID)
}
