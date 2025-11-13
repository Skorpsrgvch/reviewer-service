package model

import "errors"

var (
	ErrUserNotFound   = errors.New("user not found")
	ErrTeamNotFound   = errors.New("team not found")
	ErrPRExists       = errors.New("PR already exists")
	ErrPRNotFound     = errors.New("PR not found")
	ErrPRMerged       = errors.New("PR is already merged")
	ErrNotAssigned    = errors.New("reviewer is not assigned to this PR")
	ErrNoCandidate    = errors.New("no active replacement candidate in team")
	ErrAuthorNotFound = errors.New("author not found")
	ErrTeamExists     = errors.New("team already exists")
)
