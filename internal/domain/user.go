package domain

import "fmt"

type User struct {
	id       string
	username string
	isActive bool
}

// NewUser создаёт нового пользователя
func NewUser(id, username string, isActive bool) (*User, error) {
	if id == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	if username == "" {
		return nil, fmt.Errorf("username is required")
	}
	return &User{id: id, username: username, isActive: isActive}, nil
}

// ID возвращает идентификатор пользователя
func (u *User) ID() string {
	return u.id
}

// Username возвращает имя пользователя
func (u *User) Username() string {
	return u.username
}

// IsActive возвращает статус активности
func (u *User) IsActive() bool {
	return u.isActive
}

// SetActive обновляет статус активности
func (u *User) SetActive(active bool) {
	u.isActive = active
}
