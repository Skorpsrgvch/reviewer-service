package domain

import (
	"errors"
	"fmt"
	"time"
)

type PRStatus string

const (
	PROpen   PRStatus = "OPEN"
	PRMerged PRStatus = "MERGED"
)

type PullRequest struct {
	id                string
	name              string
	authorID          string
	status            PRStatus
	assignedReviewers []string
	createdAt         time.Time
	mergedAt          *time.Time
}

// NewPullRequest создаёт новый PR в статусе OPEN
func NewPullRequest(id, name, authorID string, reviewers []string) (*PullRequest, error) {
	if id == "" {
		return nil, fmt.Errorf("PR ID is required")
	}
	if name == "" {
		return nil, fmt.Errorf("PR name is required")
	}
	if authorID == "" {
		return nil, fmt.Errorf("author ID is required")
	}
	if len(reviewers) == 0 {
		return nil, fmt.Errorf("at least one reviewer is required")
	}

	return &PullRequest{
		id:                id,
		name:              name,
		authorID:          authorID,
		status:            PROpen,
		assignedReviewers: reviewers,
		createdAt:         time.Now().UTC(),
	}, nil
}

// ID возвращает идентификатор PR
func (pr *PullRequest) ID() string {
	return pr.id
}

// Name возвращает название PR
func (pr *PullRequest) Name() string {
	return pr.name
}

// AuthorID возвращает ID автора
func (pr *PullRequest) AuthorID() string {
	return pr.authorID
}

// Status возвращает статус
func (pr *PullRequest) Status() PRStatus {
	return pr.status
}

// AssignedReviewers возвращает список ревьюеров
func (pr *PullRequest) AssignedReviewers() []string {
	reviewers := make([]string, len(pr.assignedReviewers))
	copy(reviewers, pr.assignedReviewers)
	return reviewers
}

// CreatedAt возвращает время создания
func (pr *PullRequest) CreatedAt() time.Time {
	return pr.createdAt
}

// MergedAt возвращает время мержа (может быть nil)
func (pr *PullRequest) MergedAt() *time.Time {
	if pr.mergedAt == nil {
		return nil
	}
	t := *pr.mergedAt
	return &t
}

// CanBeMerged проверяет, можно ли мержить
func (pr *PullRequest) CanBeMerged() error {
	if pr.status == PRMerged {
		return ErrPRAlreadyMerged
	}
	return nil
}

// Merge переводит PR в статус MERGED
func (pr *PullRequest) Merge() error {
	if err := pr.CanBeMerged(); err != nil {
		return err
	}
	pr.status = PRMerged
	now := time.Now().UTC()
	pr.mergedAt = &now
	return nil
}

// IsReviewerAssigned проверяет, назначен ли ревьюер
func (pr *PullRequest) IsReviewerAssigned(reviewerID string) bool {
	for _, r := range pr.assignedReviewers {
		if r == reviewerID {
			return true
		}
	}
	return false
}

// ReplaceReviewer заменяет старого ревьюера на нового
func (pr *PullRequest) ReplaceReviewer(oldReviewerID, newReviewerID string) error {
	found := false
	for i, r := range pr.assignedReviewers {
		if r == oldReviewerID {
			pr.assignedReviewers[i] = newReviewerID
			found = true
			break
		}
	}
	if !found {
		return ErrReviewerNotAssigned
	}
	return nil
}

// RestorePullRequest создаёт PR из данных БД (используется только адаптером)
func RestorePullRequest(
	id, name, authorID string,
	status PRStatus,
	assignedReviewers []string,
	createdAt time.Time,
	mergedAt *time.Time,
) (*PullRequest, error) {
	if id == "" || name == "" || authorID == "" {
		return nil, errors.New("invalid PR data")
	}
	return &PullRequest{
		id:                id,
		name:              name,
		authorID:          authorID,
		status:            status,
		assignedReviewers: assignedReviewers,
		createdAt:         createdAt,
		mergedAt:          mergedAt,
	}, nil
}
