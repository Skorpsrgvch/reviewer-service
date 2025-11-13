package pullrequest

// CreatePRRequest — тело запроса для /pullRequest/create
type CreatePRRequest struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
}

// CreatePRResponse — ответ на создание PR
type CreatePRResponse struct {
	PR PullRequestDTO `json:"pr"`
}

// MergePRRequest — тело запроса для /pullRequest/merge
type MergePRRequest struct {
	PullRequestID string `json:"pull_request_id"`
}

// MergePRResponse — ответ на merge
type MergePRResponse struct {
	PR PullRequestDTO `json:"pr"`
}

// ReassignPRRequest — тело запроса для /pullRequest/reassign
type ReassignPRRequest struct {
	PullRequestID string `json:"pull_request_id"`
	OldReviewerID string `json:"old_reviewer_id"` // Обрати внимание: в OpenAPI — old_reviewer_id
}

// ReassignPRResponse — ответ на переназначение
type ReassignPRResponse struct {
	PR         PullRequestDTO `json:"pr"`
	ReplacedBy string         `json:"replaced_by"`
}

// PullRequestDTO — полная информация о Pull Request (в ответах)
type PullRequestDTO struct {
	PullRequestID     string   `json:"pull_request_id"`
	PullRequestName   string   `json:"pull_request_name"`
	AuthorID          string   `json:"author_id"`
	Status            string   `json:"status"`
	AssignedReviewers []string `json:"assigned_reviewers"`
	CreatedAt         string   `json:"createdAt,omitempty"` // RFC3339 строка
	MergedAt          *string  `json:"mergedAt,omitempty"`  // nullable → *string
}
