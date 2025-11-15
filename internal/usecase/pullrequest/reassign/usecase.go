package reassign

import (
	"context"
	"errors"
	"math/rand"

	"github.com/Skorpsrgvch/reviewer-service/internal/domain"
)

type Input struct {
	PullRequestID string
	OldReviewerID string
}

type Output struct {
	NewReviewerID string
}

type Usecase struct {
	prRepo   PullRequestRepository
	userRepo UserRepository
	teamRepo TeamRepository
}

func NewUsecase(prRepo PullRequestRepository, userRepo UserRepository, teamRepo TeamRepository) (*Usecase, error) {
	if prRepo == nil || userRepo == nil || teamRepo == nil {
		return nil, errors.New("all dependencies required")
	}
	return &Usecase{prRepo: prRepo, userRepo: userRepo, teamRepo: teamRepo}, nil
}

func (u *Usecase) Execute(ctx context.Context, input Input) (*domain.PullRequest, string, error) {
	pr, err := u.prRepo.GetByID(ctx, input.PullRequestID)
	if err != nil {
		return nil, "", err
	}

	if pr.Status() == domain.PRMerged {
		return nil, "", domain.ErrPRAlreadyMerged
	}

	if !pr.IsReviewerAssigned(input.OldReviewerID) {
		return nil, "", domain.ErrReviewerNotAssigned
	}

	// Получаем команду старого ревьюера
	reviewerTeam, err := u.userRepo.GetTeamByUser(ctx, input.OldReviewerID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, "", domain.ErrUserNotFound
		}
		return nil, "", err
	}

	// Получаем активных участников его команды
	activeMembers, err := u.teamRepo.GetUsersInTeam(ctx, reviewerTeam, true)
	if err != nil {
		return nil, "", err
	}

	var candidates []string
	for _, user := range activeMembers {
		if user.ID() != pr.AuthorID() &&
			user.ID() != input.OldReviewerID &&
			!isInSlice(user.ID(), pr.AssignedReviewers()) {
			candidates = append(candidates, user.ID())
		}
	}

	if len(candidates) == 0 {
		return nil, "", domain.ErrNoActiveReviewers
	}

	newReviewer := candidates[rand.Intn(len(candidates))]
	newReviewers := replaceInSlice(pr.AssignedReviewers(), input.OldReviewerID, newReviewer)

	if err := u.prRepo.UpdateReviewers(ctx, input.PullRequestID, newReviewers); err != nil {
		return nil, "", err
	}

	// Обновляем локальное состояние PR
	_ = pr.ReplaceReviewer(input.OldReviewerID, newReviewer)
	return pr, newReviewer, nil
}

func isInSlice(id string, slice []string) bool {
	for _, s := range slice {
		if s == id {
			return true
		}
	}
	return false
}

func replaceInSlice(slice []string, old, new string) []string {
	for i, s := range slice {
		if s == old {
			slice[i] = new
			break
		}
	}
	return slice
}
