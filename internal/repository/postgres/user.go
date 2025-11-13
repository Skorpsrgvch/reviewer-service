package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Skorpsrgvch/reviewer-service/internal/model"
)

// Реализует:
// - team.UserRepository
// - user.UserRepository
// - pullrequest.UserRepository

func (r *DBRepo) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	query := `
		SELECT id, username, team_name, is_active
		FROM users
		WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, id)

	var u model.User
	err := row.Scan(&u.ID, &u.Username, &u.TeamName, &u.IsActive)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *DBRepo) CreateUser(ctx context.Context, u *model.User) error {
	query := `
		INSERT INTO users (id, username, team_name, is_active)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.db.ExecContext(ctx, query, u.ID, u.Username, u.TeamName, u.IsActive)
	return err
}

func (r *DBRepo) UpdateUser(ctx context.Context, u *model.User) error {
	query := `
		UPDATE users
		SET username = $1, team_name = $2, is_active = $3
		WHERE id = $4
	`
	_, err := r.db.ExecContext(ctx, query, u.Username, u.TeamName, u.IsActive, u.ID)
	return err
}

func (r *DBRepo) GetUsersByTeam(ctx context.Context, teamName string) ([]model.User, error) {
	query := `
		SELECT id, username, team_name, is_active
		FROM users
		WHERE team_name = $1
	`
	rows, err := r.db.QueryContext(ctx, query, teamName)
	if err != nil {
		// Если нет строк — это "команда не найдена"
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrTeamNotFound
		}
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Username, &u.TeamName, &u.IsActive); err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	if len(users) == 0 {
		return nil, model.ErrTeamNotFound
	}
	return users, nil
}

func (r *DBRepo) GetActiveUsersByTeam(ctx context.Context, teamName string) ([]model.User, error) {
	query := `
		SELECT id, username, team_name, is_active
		FROM users
		WHERE team_name = $1 AND is_active = true
	`
	rows, err := r.db.QueryContext(ctx, query, teamName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Username, &u.TeamName, &u.IsActive); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}
