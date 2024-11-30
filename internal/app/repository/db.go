package repository

import (
	"context"
	"fmt"

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
	return &DBRepository{
		pool: pool,
	}, nil
}

func (d *DBRepository) GetShortLink(ctx context.Context, originalURL string) (string, bool, error) {
	return "", false, nil
}

func (d *DBRepository) GetOriginalURL(ctx context.Context, shortLink string) (string, bool, error) {
	return "", false, nil
}

func (d *DBRepository) SetLink(ctx context.Context, originalURL string, shortLink string) (bool, error) {
	return false, nil
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
