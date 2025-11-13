package model

import "time"

type PullRequest struct {
	ID                string
	Name              string
	AuthorID          string
	Status            string // "OPEN" | "MERGED"
	AssignedReviewers []string
	CreatedAt         time.Time
	MergedAt          *time.Time
}
