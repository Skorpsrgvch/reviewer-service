package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Skorpsrgvch/reviewer-service/internal/domain"
	"github.com/lib/pq"
)

// PullRequestRepo реализует работу с PullRequest в PostgreSQL.
// Следует принципу: «один репозиторий — один агрегат».
type PullRequestRepo struct {
	db *sql.DB
}

func NewPullRequestRepo(db *sql.DB) *PullRequestRepo {
	return &PullRequestRepo{db: db}
}

// Save сохраняет новый PullRequest (только в статусе OPEN).
func (r *PullRequestRepo) Save(ctx context.Context, pr *domain.PullRequest) error {
	// Валидация: можно сохранять только OPEN-запросы
	if pr.Status() != domain.PROpen {
		return errors.New("only OPEN pull requests can be saved")
	}

	query := `
		INSERT INTO pull_requests (id, name, author_id, status, assigned_reviewers, created_at, merged_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.ExecContext(ctx, query,
		pr.ID(),
		pr.Name(),
		pr.AuthorID(),
		string(pr.Status()),
		pq.Array(pr.AssignedReviewers()),
		pr.CreatedAt(),
		nil, // merged_at = NULL для OPEN
	)
	return err
}

// GetByID возвращает PullRequest по ID.
func (r *PullRequestRepo) GetByID(ctx context.Context, id string) (*domain.PullRequest, error) {
	query := `
		SELECT id, name, author_id, status, assigned_reviewers, created_at, merged_at
		FROM pull_requests
		WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, id)

	var (
		idStr, name, authorID, statusStr string
		reviewers                        pq.StringArray
		createdAt                        time.Time
		mergedAt                         *time.Time
	)

	if err := row.Scan(&idStr, &name, &authorID, &statusStr, &reviewers, &createdAt, &mergedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrPRNotFound
		}
		return nil, err
	}

	status := domain.PRStatus(statusStr)
	assignedReviewers := []string(reviewers)

	return domain.RestorePullRequest(idStr, name, authorID, status, assignedReviewers, createdAt, mergedAt)
}

// PRExists проверяет существование PR по ID.
func (r *PullRequestRepo) PRExists(ctx context.Context, id string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM pull_requests WHERE id = $1)", id).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// UpdateReviewers обновляет список ревьюеров у существующего PR.
func (r *PullRequestRepo) UpdateReviewers(ctx context.Context, id string, reviewers []string) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE pull_requests SET assigned_reviewers = $1 WHERE id = $2",
		pq.Array(reviewers), id,
	)
	return err
}

// Merge переводит PR в статус MERGED с указанным временем.
// Идемпотентен: если уже MERGED — не ошибка.
func (r *PullRequestRepo) Merge(ctx context.Context, id string, mergedAt time.Time) error {
	res, err := r.db.ExecContext(ctx,
		"UPDATE pull_requests SET status = 'MERGED', merged_at = $1 WHERE id = $2 AND status = 'OPEN'",
		mergedAt, id,
	)
	if err != nil {
		return err
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		// Не обновилось → либо не существует, либо уже MERGED
		exists, err := r.PRExists(ctx, id)
		if err != nil {
			return err
		}
		if !exists {
			return domain.ErrPRNotFound
		}
		// Если существует — значит, уже MERGED → OK
	}
	return nil
}

// GetByReviewer возвращает все PR, где reviewerID в assigned_reviewers.
func (r *PullRequestRepo) GetByReviewer(ctx context.Context, reviewerID string) ([]domain.PullRequest, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, author_id, status, assigned_reviewers, created_at, merged_at
		 FROM pull_requests
		 WHERE $1 = ANY(assigned_reviewers)`,
		reviewerID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prs []domain.PullRequest
	for rows.Next() {
		var id, name, authorID, statusStr string
		var reviewers pq.StringArray
		var createdAt time.Time
		var mergedAt *time.Time

		if err := rows.Scan(&id, &name, &authorID, &statusStr, &reviewers, &createdAt, &mergedAt); err != nil {
			return nil, err
		}

		status := domain.PRStatus(statusStr)
		assigned := []string(reviewers)

		pr, err := domain.RestorePullRequest(id, name, authorID, status, assigned, createdAt, mergedAt)
		if err != nil {
			return nil, err
		}
		prs = append(prs, *pr)
	}
	return prs, nil
}

// GetReviewerStats возвращает количество OPEN PR на каждого ревьюера.
func (r *PullRequestRepo) GetReviewerStats(ctx context.Context) (map[string]int, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT reviewer, COUNT(*)
		FROM (
			SELECT unnest(assigned_reviewers) AS reviewer
			FROM pull_requests
			WHERE status = 'OPEN'
		) AS active_reviewers
		GROUP BY reviewer
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make(map[string]int)
	for rows.Next() {
		var reviewer string
		var count int
		if err := rows.Scan(&reviewer, &count); err != nil {
			return nil, err
		}
		stats[reviewer] = count
	}
	return stats, nil
}
