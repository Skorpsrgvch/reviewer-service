package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Skorpsrgvch/reviewer-service/internal/model"
	"github.com/lib/pq"
)

// Реализует:
// - pullrequest.PullRequestRepository
// - user.PullRequestRepository

func (r *DBRepo) PRExists(ctx context.Context, id string) (bool, error) {
	query := `SELECT 1 FROM pull_requests WHERE id = $1`
	var exists int
	err := r.db.QueryRowContext(ctx, query, id).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *DBRepo) CreatePR(ctx context.Context, pr *model.PullRequest) error {
	query := `
		INSERT INTO pull_requests (id, name, author_id, status, assigned_reviewers, created_at, merged_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.ExecContext(ctx, query,
		pr.ID,
		pr.Name,
		pr.AuthorID,
		pr.Status,
		pq.Array(pr.AssignedReviewers), // ← Используем pq.Array
		pr.CreatedAt,
		pr.MergedAt,
	)
	return err
}

func (r *DBRepo) GetPRByID(ctx context.Context, id string) (*model.PullRequest, error) {
	query := `
		SELECT id, name, author_id, status, assigned_reviewers, created_at, merged_at
		FROM pull_requests
		WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, id)

	var pr model.PullRequest
	var reviewers pq.StringArray // ← Тип для массива
	var mergedAt *time.Time

	err := row.Scan(
		&pr.ID,
		&pr.Name,
		&pr.AuthorID,
		&pr.Status,
		&reviewers,
		&pr.CreatedAt,
		&mergedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrPRNotFound
		}
		return nil, err
	}

	// Преобразуем pq.StringArray → []string
	pr.AssignedReviewers = []string(reviewers)
	pr.MergedAt = mergedAt

	return &pr, nil
}

func (r *DBRepo) UpdatePRReviewers(ctx context.Context, id string, reviewers []string) error {
	query := `
		UPDATE pull_requests
		SET assigned_reviewers = $1
		WHERE id = $2
	`
	_, err := r.db.ExecContext(ctx, query, pq.Array(reviewers), id) // ← pq.Array
	return err
}

func (r *DBRepo) UpdatePRStatusToMerged(ctx context.Context, id string, mergedAt time.Time) error {
	query := `
		UPDATE pull_requests
		SET status = 'MERGED', merged_at = $1
		WHERE id = $2 AND status = 'OPEN'
	`
	res, err := r.db.ExecContext(ctx, query, mergedAt, id)
	if err != nil {
		return err
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		exists, _ := r.PRExists(ctx, id)
		if !exists {
			return model.ErrPRNotFound
		}
		// Идемпотентность: если уже MERGED — OK
	}
	return nil
}

func (r *DBRepo) GetPRsByReviewer(ctx context.Context, reviewerID string) ([]model.PullRequest, error) {
	// Используем оператор ANY для поиска по массиву
	query := `
		SELECT id, name, author_id, status, assigned_reviewers, created_at, merged_at
		FROM pull_requests
		WHERE $1 = ANY(assigned_reviewers)
	`
	rows, err := r.db.QueryContext(ctx, query, reviewerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prs []model.PullRequest
	for rows.Next() {
		var pr model.PullRequest
		var reviewers pq.StringArray
		var mergedAt *time.Time

		err := rows.Scan(
			&pr.ID,
			&pr.Name,
			&pr.AuthorID,
			&pr.Status,
			&reviewers,
			&pr.CreatedAt,
			&mergedAt,
		)
		if err != nil {
			return nil, err
		}

		pr.AssignedReviewers = []string(reviewers)
		pr.MergedAt = mergedAt
		prs = append(prs, pr)
	}
	return prs, nil
}

func (r *DBRepo) GetReviewerStats(ctx context.Context) (map[string]int, error) {
	query := `
		SELECT reviewer, COUNT(*) 
		FROM (
			SELECT unnest(assigned_reviewers) AS reviewer
			FROM pull_requests
			WHERE status = 'OPEN'
		) AS reviewers
		GROUP BY reviewer
	`
	rows, err := r.db.QueryContext(ctx, query)
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
