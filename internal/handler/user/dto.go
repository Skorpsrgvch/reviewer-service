package user

// SetIsActiveRequest — тело запроса для /users/setIsActive
type SetIsActiveRequest struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

// SetIsActiveResponse — ответ на обновление активности
type SetIsActiveResponse struct {
	User UserDTO `json:"user"`
}

// UserDTO — полная информация о пользователе
type UserDTO struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}

// GetReviewResponse — ответ для /users/getReview
type GetReviewResponse struct {
	UserID       string             `json:"user_id"`
	PullRequests []PullRequestShort `json:"pull_requests"`
}

// PullRequestShort — краткая информация о PR (без назначенных ревьюеров и времени)
type PullRequestShort struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
	Status          string `json:"status"`
}
