package domain

import "errors"

var (
	ErrPRExists            = errors.New("pull request already exists")
	ErrPRNotFound          = errors.New("pull request not found")
	ErrPRAlreadyMerged     = errors.New("pull request is already merged")
	ErrReviewerNotAssigned = errors.New("reviewer is not assigned to this pull request")
	ErrNoActiveReviewers   = errors.New("no active reviewers available for reassignment")
	ErrAuthorNotFound      = errors.New("author not found")
	ErrUserNotFound        = errors.New("user not found")
	ErrTeamExists          = errors.New("team already exists")
	ErrTeamNotFound        = errors.New("team not found")
)
