package team

// CreateTeamRequest — тело запроса для /team/add
type CreateTeamRequest struct {
	TeamName string       `json:"team_name"`
	Members  []TeamMember `json:"members"`
}

type TeamMember struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

// CreateTeamResponse — ответ на успешное создание команды
type CreateTeamResponse struct {
	Team TeamDTO `json:"team"`
}

type TeamDTO struct {
	TeamName string       `json:"team_name"`
	Members  []TeamMember `json:"members"`
}

// GetTeamResponse — ответ для /team/get
type GetTeamResponse struct {
	TeamName string       `json:"team_name"`
	Members  []TeamMember `json:"members"`
}
