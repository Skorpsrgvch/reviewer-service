package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Skorpsrgvch/reviewer-service/internal/domain"
)

type TeamRepo struct {
	db *sql.DB
}

func NewTeamRepo(db *sql.DB) *TeamRepo {
	return &TeamRepo{db: db}
}

func (r *TeamRepo) CreateTeam(ctx context.Context, teamName string, members []domain.User) error {
	// 1. Создаём команду
	_, err := r.db.ExecContext(ctx, "INSERT INTO teams (name) VALUES ($1) ON CONFLICT (name) DO NOTHING", teamName)
	if err != nil {
		return err
	}

	// 2. Обрабатываем пользователей
	for _, u := range members {
		// Проверяем существование
		_, err := r.GetUserByID(ctx, u.ID())
		if err == nil {
			// Обновляем
			if err := r.UpdateUser(ctx, &u); err != nil {
				return err
			}
		} else if errors.Is(err, domain.ErrUserNotFound) {
			// Создаём
			if err := r.CreateUser(ctx, &u); err != nil {
				return err
			}
		} else {
			return err
		}

		// 3. Переназначаем команду
		_, err = r.db.ExecContext(ctx, "DELETE FROM team_members WHERE user_id = $1", u.ID())
		if err != nil {
			return err
		}
		_, err = r.db.ExecContext(ctx, "INSERT INTO team_members (team_name, user_id) VALUES ($1, $2)", teamName, u.ID())
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *TeamRepo) SaveTeam(ctx context.Context, team *domain.Team) error {
	teamName := team.Name()
	members := team.Members()

	// Вставляем команду
	_, err := r.db.ExecContext(ctx, "INSERT INTO teams (name) VALUES ($1) ON CONFLICT (name) DO NOTHING", teamName)
	if err != nil {
		return err
	}

	// Обрабатываем пользователей
	for _, u := range members {
		// Проверяем существование
		_, err := r.GetUserByID(ctx, u.ID())
		if err == nil {
			if err := r.UpdateUser(ctx, &u); err != nil {
				return err
			}
		} else if errors.Is(err, domain.ErrUserNotFound) {
			if err := r.CreateUser(ctx, &u); err != nil {
				return err
			}
		} else {
			return err
		}

		// Обновляем team_members
		_, err = r.db.ExecContext(ctx, "DELETE FROM team_members WHERE user_id = $1", u.ID())
		if err != nil {
			return err
		}
		_, err = r.db.ExecContext(ctx, "INSERT INTO team_members (team_name, user_id) VALUES ($1, $2)", teamName, u.ID())
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *TeamRepo) FindTeamByName(ctx context.Context, teamName string) (*domain.Team, error) {
	query := `
		SELECT u.id, u.username, u.is_active
		FROM users u
		JOIN team_members tm ON u.id = tm.user_id
		WHERE tm.team_name = $1
	`
	rows, err := r.db.QueryContext(ctx, query, teamName)
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

	return domain.NewTeam(teamName, users)
}

func (r *TeamRepo) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	// дублирование — временно, пока нет общего репозитория
	query := `SELECT id, username, is_active FROM users WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)
	var idStr, username string
	var isActive bool
	if err := row.Scan(&idStr, &username, &isActive); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return domain.NewUser(idStr, username, isActive)
}

func (r *TeamRepo) CreateUser(ctx context.Context, u *domain.User) error {
	query := `INSERT INTO users (id, username, is_active) VALUES ($1, $2, $3)`
	_, err := r.db.ExecContext(ctx, query, u.ID(), u.Username(), u.IsActive())
	return err
}

func (r *TeamRepo) UpdateUser(ctx context.Context, u *domain.User) error {
	query := `UPDATE users SET username = $1, is_active = $2 WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, u.Username(), u.IsActive(), u.ID())
	return err
}

func (r *TeamRepo) GetTeamByName(ctx context.Context, teamName string) (*domain.Team, error) {
	users, err := r.GetUsersInTeam(ctx, teamName, false)
	if err != nil {
		return nil, err
	}

	team, err := domain.NewTeam(teamName, users)
	if err != nil {
		return nil, err
	}
	return team, nil
}

func (r *TeamRepo) GetUsersInTeam(ctx context.Context, teamName string, onlyActive bool) ([]domain.User, error) {
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
