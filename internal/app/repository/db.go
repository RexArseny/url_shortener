package repository

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/RexArseny/url_shortener/internal/app/models"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type DBRepository struct {
	logger *zap.Logger
	pool   *pgxpool.Pool
}

func NewDBRepository(ctx context.Context, logger *zap.Logger, connString string) (*DBRepository, error) {
	m, err := migrate.New("file://./internal/app/repository/migrations", connString)
	if err != nil {
		return nil, fmt.Errorf("can not create migration instance: %w", err)
	}
	err = m.Up()
	if err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return nil, fmt.Errorf("can not migrate up: %w", err)
		}
	}

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("can not create new pool for PostgreSQL server: %w", err)
	}
	err = pool.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("can not ping PostgreSQL server: %w", err)
	}

	return &DBRepository{
		logger: logger,
		pool:   pool,
	}, nil
}

func (d *DBRepository) GetOriginalURL(ctx context.Context, shortLink string) (*string, error) {
	var originalURL string
	err := d.pool.QueryRow(ctx, "SELECT original_url FROM urls WHERE short_url=$1", shortLink).Scan(&originalURL)
	if err != nil {
		return nil, fmt.Errorf("can not get original url: %w", err)
	}
	return &originalURL, nil
}

func (d *DBRepository) SetLink(ctx context.Context, originalURL string, shortURLs []string) (*string, error) {
	for _, shortURL := range shortURLs {
		var link string
		err := d.pool.QueryRow(ctx, `INSERT INTO urls (short_url, original_url) 
									VALUES ($1, $2) 
									ON CONFLICT (original_url) 
									DO UPDATE SET original_url=EXCLUDED.original_url 
									RETURNING short_url`, shortURL, originalURL).Scan(&link)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) &&
				pgErr.Code == pgerrcode.UniqueViolation &&
				pgErr.ConstraintName == "short_url_constraint" {
				continue
			}
			return nil, fmt.Errorf("can not set link: %w", err)
		}
		if link != shortURL {
			return &link, ErrOriginalURLUniqueViolation
		}

		return &shortURL, nil
	}
	return nil, ErrReachedMaxGenerationRetries
}

func (d *DBRepository) SetLinks(
	ctx context.Context,
	batch []models.ShortenBatchRequest,
	shortURLs [][]string,
) ([]string, error) {
	urls := make(map[string]string)
	originalURLs := make([]string, 0, len(batch))
	for _, originalURL := range batch {
		originalURLs = append(originalURLs, originalURL.OriginalURL)
	}

	tx, err := d.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("can not start transaction: %w", err)
	}
	defer func() {
		err = tx.Rollback(ctx)
		if err != nil && !strings.Contains(err.Error(), "tx is closed") {
			d.logger.Error("Can not rollback transaction", zap.Error(err))
		}
	}()

	var originalURLUniqueViolation bool
	rows, err := tx.Query(ctx, "SELECT short_url, original_url FROM urls WHERE original_url = ANY ($1)", originalURLs)
	if err != nil {
		return nil, fmt.Errorf("can not get original urls: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		originalURLUniqueViolation = true

		var shortURL string
		var originalURL string
		err = rows.Scan(
			&shortURL,
			&originalURL,
		)
		if err != nil {
			return nil, fmt.Errorf("can not read row: %w", err)
		}

		urls[originalURL] = shortURL
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("can not read rows: %w", err)
	}

	for i := len(originalURLs) - 1; i >= 0; i-- {
		if _, ok := urls[originalURLs[i]]; ok {
			originalURLs = append(originalURLs[:i], originalURLs[i+1:]...)
		}
	}

	for i := range originalURLs {
		_, err := url.ParseRequestURI(originalURLs[i])
		if err != nil {
			return nil, ErrInvalidURL
		}

		var generated bool
		var shortURL string
		for _, shortURL = range shortURLs[i] {
			commandTag, err := tx.Exec(ctx, `INSERT INTO urls (short_url, original_url) 
											VALUES ($1, $2) 
											ON CONFLICT (short_url) 
											DO NOTHING`, shortURL, originalURLs[i])
			if err != nil {
				return nil, fmt.Errorf("can not set link: %w", err)
			}
			if commandTag.RowsAffected() == 0 {
				continue
			}

			generated = true
			break
		}

		if !generated {
			return nil, ErrReachedMaxGenerationRetries
		}
		urls[originalURLs[i]] = shortURL
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("can not commit transaction: %w", err)
	}

	result := make([]string, 0, len(batch))
	for i := range batch {
		result = append(result, urls[batch[i].OriginalURL])
	}

	if originalURLUniqueViolation {
		return result, ErrOriginalURLUniqueViolation
	}

	return result, nil
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
