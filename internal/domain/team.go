package domain

import "fmt"

type Team struct {
	name    string
	members []User
}

// NewTeam создаёт новую команду
func NewTeam(name string, members []User) (*Team, error) {
	if name == "" {
		return nil, fmt.Errorf("team name is required")
	}
	if len(members) == 0 {
		return nil, fmt.Errorf("team must have at least one member")
	}
	return &Team{name: name, members: members}, nil
}

func (t *Team) Name() string {
	return t.name
}

func (t *Team) Members() []User {
	// Возвращаем копию, чтобы избежать мутаций извне
	members := make([]User, len(t.members))
	copy(members, t.members)
	return members
}

// FindActiveMembers возвращает активных участников команды, исключая указанных по ID
func (t *Team) FindActiveMembers(excludeIDs []string) []string {
	exclude := make(map[string]bool)
	for _, id := range excludeIDs {
		exclude[id] = true
	}

	var candidates []string
	for _, u := range t.members {
		if u.IsActive() && !exclude[u.ID()] {
			candidates = append(candidates, u.ID())
		}
	}
	return candidates
}
