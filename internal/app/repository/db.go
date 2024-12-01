package repository

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/RexArseny/url_shortener/internal/app/models"
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
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("can not create new pool for PostgreSQL server: %w", err)
	}
	err = pool.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("can not ping PostgreSQL server: %w", err)
	}
	_, err = pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS urls 
							(id SERIAL PRIMARY KEY, 
							short_url text NOT NULL, 
							original_url text NOT NULL, 
							UNIQUE(short_url), 
							UNIQUE(original_url))`)
	if err != nil {
		return nil, fmt.Errorf("can not create table: %w", err)
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

func (d *DBRepository) SetLink(ctx context.Context, originalURL string) (*string, error) {
	var retry int
	for retry < linkGenerationRetries {
		shortLink := generatePath()

		var link string
		err := d.pool.QueryRow(ctx, `INSERT INTO urls (short_url, original_url) 
									VALUES ($1, $2) 
									ON CONFLICT (original_url) 
									DO UPDATE SET original_url=EXCLUDED.original_url 
									RETURNING short_url`, shortLink, originalURL).Scan(&link)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) &&
				pgErr.Code == pgerrcode.UniqueViolation &&
				pgErr.ConstraintName == "urls_short_url_key" {
				retry++
				continue
			}
			return nil, fmt.Errorf("can not set link: %w", err)
		}
		if link != shortLink {
			return &link, models.ErrOriginalURLUniqueViolation
		}

		return &shortLink, nil
	}
	return nil, models.ErrReachedMaxGenerationRetries
}

func (d *DBRepository) SetLinks(ctx context.Context, batch []models.ShortenBatchRequest) ([]string, error) {
	shortURLs := make(map[string]string)
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

		shortURLs[originalURL] = shortURL
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("can not read rows: %w", err)
	}

	for i := len(originalURLs) - 1; i >= 0; i-- {
		if _, ok := shortURLs[originalURLs[i]]; ok {
			originalURLs = append(originalURLs[:i], originalURLs[i+1:]...)
		}
	}

	for i := range originalURLs {
		_, err := url.ParseRequestURI(originalURLs[i])
		if err != nil {
			return nil, models.ErrInvalidURL
		}

		var retry int
		var generated bool
		var shortLink string
		for retry < linkGenerationRetries {
			shortLink = generatePath()

			commandTag, err := tx.Exec(ctx, `INSERT INTO urls (short_url, original_url) 
											VALUES ($1, $2) 
											ON CONFLICT (short_url) 
											DO NOTHING`, shortLink, originalURLs[i])
			if err != nil {
				return nil, fmt.Errorf("can not set link: %w", err)
			}
			if commandTag.RowsAffected() == 0 {
				retry++
				continue
			}

			generated = true
			break
		}

		if !generated {
			return nil, models.ErrReachedMaxGenerationRetries
		}
		shortURLs[originalURLs[i]] = shortLink
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("can not commit transaction: %w", err)
	}

	result := make([]string, 0, len(batch))
	for i := range batch {
		result = append(result, shortURLs[batch[i].OriginalURL])
	}

	if originalURLUniqueViolation {
		return result, models.ErrOriginalURLUniqueViolation
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
