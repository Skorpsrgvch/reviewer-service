package postgres

import (
	"context"
	"database/sql"

	_ "github.com/lib/pq"
)

type DBRepo struct {
	db *sql.DB
}

func NewDBRepo(db *sql.DB) *DBRepo {
	return &DBRepo{db: db}
}

func (r *DBRepo) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}
