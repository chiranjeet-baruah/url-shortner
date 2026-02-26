// Package driven contains "driven" adapters — implementations of the domain
// ports that are *driven* (called) by the application core.
// This file implements the URLRepository port using PostgreSQL.
package driven

import (
	"context"
	"database/sql"

	"url-shortener/internal/v1/shortener/domain"
)

// PostgresRepo implements domain.URLRepository using a PostgreSQL database.
// It uses Go's standard database/sql package with parameterized queries
// to prevent SQL injection.
type PostgresRepo struct {
	db *sql.DB // The database connection pool managed by database/sql
}

// NewPostgresRepo creates a new PostgresRepo with the given database connection.
func NewPostgresRepo(db *sql.DB) *PostgresRepo {
	return &PostgresRepo{db: db}
}

// AutoMigrate creates the "urls" table and index if they don't already exist.
// This is a simple migration strategy suitable for small projects.
func (r *PostgresRepo) AutoMigrate(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS urls (
			id           SERIAL PRIMARY KEY,
			short_code   VARCHAR(10) UNIQUE NOT NULL,
			original_url TEXT NOT NULL,
			click_count  BIGINT DEFAULT 0,
			created_at   TIMESTAMPTZ DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_urls_short_code ON urls(short_code);
	`
	_, err := r.db.ExecContext(ctx, query)
	return err
}

// Save inserts a new URL record into the database.
// It uses PostgreSQL's RETURNING clause to populate the ID and CreatedAt
// fields on the domain object after insertion.
func (r *PostgresRepo) Save(ctx context.Context, u *domain.URL) error {
	query := `INSERT INTO urls (short_code, original_url) VALUES ($1, $2) RETURNING id, created_at`
	return r.db.QueryRowContext(ctx, query, u.ShortCode, u.OriginalURL).Scan(&u.ID, &u.CreatedAt)
}

// FindByCode retrieves a URL record by its short code.
// Returns an error if the code doesn't exist (sql.ErrNoRows).
func (r *PostgresRepo) FindByCode(ctx context.Context, code string) (*domain.URL, error) {
	u := &domain.URL{}
	query := `SELECT id, short_code, original_url, click_count, created_at FROM urls WHERE short_code = $1`
	err := r.db.QueryRowContext(ctx, query, code).Scan(&u.ID, &u.ShortCode, &u.OriginalURL, &u.ClickCount, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// IncrementClicks atomically adds 1 to the click counter for the given short code.
// This is called from the background click tracker goroutine.
func (r *PostgresRepo) IncrementClicks(ctx context.Context, code string) error {
	query := `UPDATE urls SET click_count = click_count + 1 WHERE short_code = $1`
	_, err := r.db.ExecContext(ctx, query, code)
	return err
}
