package pullrequest

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/Skorpsrgvch/reviewer-service/internal/model"
)

type UserRepository interface {
	GetUserByID(ctx context.Context, id string) (*model.User, error)
	GetActiveUsersByTeam(ctx context.Context, teamName string) ([]model.User, error)
}

type PullRequestRepository interface {
	CreatePR(ctx context.Context, pr *model.PullRequest) error
	GetPRByID(ctx context.Context, id string) (*model.PullRequest, error)
	UpdatePRReviewers(ctx context.Context, id string, reviewers []string) error
	UpdatePRStatusToMerged(ctx context.Context, id string, mergedAt time.Time) error
	PRExists(ctx context.Context, id string) (bool, error)
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

func (s *Service) CreatePR(
	ctx context.Context,
	pullRequestID string,
	pullRequestName string,
	authorID string,
) (*model.PullRequest, error) {

	exists, err := s.prRepo.PRExists(ctx, pullRequestID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, model.ErrPRExists
	}

	// Получаем автора
	author, err := s.userRepo.GetUserByID(ctx, authorID)
	if err != nil {
		if errors.Is(err, model.ErrUserNotFound) {
			return nil, model.ErrAuthorNotFound
		}
		return nil, err
	}

	// Получаем активных участников команды (кроме автора)
	activeMembers, err := s.userRepo.GetActiveUsersByTeam(ctx, author.TeamName)
	if err != nil {
		return nil, err
	}

	var candidates []model.User
	for _, u := range activeMembers {
		if u.ID != authorID {
			candidates = append(candidates, u)
		}
	}

	// Перемешиваем и выбираем до 2
	rand.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})

	var assigned []string
	for i := 0; i < len(candidates) && i < 2; i++ {
		assigned = append(assigned, candidates[i].ID)
	}

	pr := &model.PullRequest{
		ID:                pullRequestID,
		Name:              pullRequestName,
		AuthorID:          authorID,
		Status:            "OPEN",
		AssignedReviewers: assigned,
		CreatedAt:         time.Now().UTC(),
		MergedAt:          nil,
	}

	if err := s.prRepo.CreatePR(ctx, pr); err != nil {
		return nil, err
	}

	return pr, nil
}

func (s *Service) MergePR(ctx context.Context, prID string) (*model.PullRequest, error) {
	pr, err := s.prRepo.GetPRByID(ctx, prID)
	if err != nil {
		if errors.Is(err, model.ErrPRNotFound) {
			return nil, model.ErrPRNotFound
		}
		return nil, err
	}

	if pr.Status == "MERGED" {
		return pr, nil // идемпотентность
	}

	mergedAt := time.Now().UTC()
	if err := s.prRepo.UpdatePRStatusToMerged(ctx, prID, mergedAt); err != nil {
		return nil, err
	}

	pr.Status = "MERGED"
	pr.MergedAt = &mergedAt
	return pr, nil
}

func (s *Service) ReassignReviewer(ctx context.Context, prID, oldReviewerID string) (*model.PullRequest, string, error) {
	pr, err := s.prRepo.GetPRByID(ctx, prID)
	if err != nil {
		if errors.Is(err, model.ErrPRNotFound) {
			return nil, "", model.ErrPRNotFound
		}
		return nil, "", err
	}

	if pr.Status == "MERGED" {
		return nil, "", model.ErrPRMerged
	}

	found := false
	for _, r := range pr.AssignedReviewers {
		if r == oldReviewerID {
			found = true
			break
		}
	}
	if !found {
		return nil, "", model.ErrNotAssigned
	}

	reviewer, err := s.userRepo.GetUserByID(ctx, oldReviewerID)
	if err != nil {
		if errors.Is(err, model.ErrUserNotFound) {
			return nil, "", model.ErrPRNotFound
		}
		return nil, "", err
	}

	activeMembers, err := s.userRepo.GetActiveUsersByTeam(ctx, reviewer.TeamName)
	if err != nil {
		return nil, "", err
	}

	var candidates []string
	for _, u := range activeMembers {
		if u.ID != pr.AuthorID &&
			u.ID != oldReviewerID &&
			!isInSlice(u.ID, pr.AssignedReviewers) &&
			u.IsActive {
			candidates = append(candidates, u.ID)
		}
	}

	if len(candidates) == 0 {
		return nil, "", model.ErrNoCandidate
	}

	newReviewer := candidates[rand.Intn(len(candidates))]
	newReviewers := replaceInSlice(pr.AssignedReviewers, oldReviewerID, newReviewer)

	if err := s.prRepo.UpdatePRReviewers(ctx, prID, newReviewers); err != nil {
		return nil, "", err
	}

	pr.AssignedReviewers = newReviewers
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
