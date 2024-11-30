package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DBRepository struct {
	pool *pgxpool.Pool
}

func NewDBRepository(ctx context.Context, connString string) (*DBRepository, error) {
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("can not create new pool for PostgreSQL server: %w", err)
	}
	err = pool.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("can not ping PostgreSQL server: %w", err)
	}
	_, err = pool.Exec(ctx, "CREATE TABLE IF NOT EXISTS urls (id SERIAL PRIMARY KEY, short_url text NOT NULL, original_url text NOT NULL, UNIQUE(short_url, original_url))")
	if err != nil {
		return nil, fmt.Errorf("can not create table: %w", err)
	}
	return &DBRepository{
		pool: pool,
	}, nil
}

func (d *DBRepository) GetShortLink(ctx context.Context, originalURL string) (string, bool, error) {
	var shortLink string
	err := d.pool.QueryRow(ctx, "SELECT short_url FROM urls WHERE original_url=$1", originalURL).Scan(&shortLink)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", false, nil
		}
		return "", false, fmt.Errorf("can not get short link: %w", err)
	}
	return shortLink, true, nil
}

func (d *DBRepository) GetOriginalURL(ctx context.Context, shortLink string) (string, bool, error) {
	var originalURL string
	err := d.pool.QueryRow(ctx, "SELECT original_url FROM urls WHERE short_url=$1", shortLink).Scan(&originalURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", false, nil
		}
		return "", false, fmt.Errorf("can not get original url: %w", err)
	}
	return originalURL, true, nil
}

func (d *DBRepository) SetLink(ctx context.Context, originalURL string, shortLink string) (bool, error) {
	_, err := d.pool.Exec(ctx, "INSERT INTO urls (short_url, original_url) VALUES ($1, $2) ON CONFLICT (short_url, original_url) DO NOTHING", shortLink, originalURL)
	if err != nil {
		return false, fmt.Errorf("can not set link: %w", err)
	}
	return true, nil
}

func (d *DBRepository) Ping(ctx context.Context) error {
	err := d.pool.Ping(ctx)
	if err != nil {
		return fmt.Errorf("can not ping PostgreSQL server: %w", err)
	}
	return nil
}

func (d *DBRepository) Close() {
	d.pool.Close()
}
