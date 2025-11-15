package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Skorpsrgvch/reviewer-service/internal/domain"
)

// Реализует интерфейс из usecase (будет объявлен позже)
type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	query := `SELECT id, username, is_active FROM users WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)

	var idStr, username string
	var isActive bool
	err := row.Scan(&idStr, &username, &isActive)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	// domain.User создаётся через конструктор
	user, err := domain.NewUser(idStr, username, isActive)
	if err != nil {
		return nil, err // не должно происходить, но на всякий случай
	}
	return user, nil
}

func (r *UserRepo) CreateUser(ctx context.Context, u *domain.User) error {
	query := `INSERT INTO users (id, username, is_active) VALUES ($1, $2, $3)`
	_, err := r.db.ExecContext(ctx, query, u.ID(), u.Username(), u.IsActive())
	return err
}

func (r *UserRepo) UpdateUser(ctx context.Context, u *domain.User) error {
	query := `UPDATE users SET username = $1, is_active = $2 WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, u.Username(), u.IsActive(), u.ID())
	return err
}

// GetUsersInTeam возвращает domain.User
func (r *UserRepo) GetUsersInTeam(ctx context.Context, teamName string, onlyActive bool) ([]domain.User, error) {
	query := `
		SELECT u.id, u.username, u.is_active
		FROM users u
		JOIN team_members tm ON u.id = tm.user_id
		WHERE tm.team_name = $1
	`
	args := []interface{}{teamName}
	if onlyActive {
		query += " AND u.is_active = true"
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var id, username string
		var isActive bool
		if err := rows.Scan(&id, &username, &isActive); err != nil {
			return nil, err
		}
		user, err := domain.NewUser(id, username, isActive)
		if err != nil {
			return nil, err
		}
		users = append(users, *user)
	}

	if len(users) == 0 {
		return nil, domain.ErrTeamNotFound
	}
	return users, nil
}

func (r *UserRepo) GetTeamByUser(ctx context.Context, userID string) (string, error) {
	var teamName string
	err := r.db.QueryRowContext(ctx,
		"SELECT team_name FROM team_members WHERE user_id = $1",
		userID,
	).Scan(&teamName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", domain.ErrUserNotFound
		}
		return "", err
	}
	return teamName, nil
}
