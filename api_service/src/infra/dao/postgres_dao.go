package dao

import (
	"context"
	"database/sql"
)

type PostgresDAO struct {
	db *sql.DB
}

func NewPostgresDAO(db *sql.DB) *PostgresDAO {
	return &PostgresDAO{db: db}
}

func (d *PostgresDAO) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	return d.db.QueryRowContext(ctx, query, args...)
}

func (d *PostgresDAO) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return d.db.ExecContext(ctx, query, args...)
}
