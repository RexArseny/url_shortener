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
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

type DBRepository struct {
	logger *zap.Logger
	pool   *Pool
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

	pool, err := NewPool(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("can not create new pool: %w", err)
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

func (d *DBRepository) SetLink(
	ctx context.Context,
	originalURL string,
	shortURLs []string,
	userID uuid.UUID,
) (*string, error) {
	for _, shortURL := range shortURLs {
		var link string
		err := d.pool.QueryRow(ctx, `INSERT INTO urls (short_url, original_url, user_id) 
									VALUES ($1, $2, $3) 
									ON CONFLICT (original_url) 
									DO UPDATE SET original_url=EXCLUDED.original_url 
									RETURNING short_url`, shortURL, originalURL, userID.String()).Scan(&link)
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
	userID uuid.UUID,
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

	for i := range shortURLs {
		b := &pgx.Batch{}

		for j := range originalURLs {
			_, err := url.ParseRequestURI(originalURLs[j])
			if err != nil {
				return nil, ErrInvalidURL
			}

			b.Queue(`INSERT INTO urls (short_url, original_url, user_id) 
			VALUES ($1, $2, $3) 
			ON CONFLICT (short_url) 
			DO NOTHING`, shortURLs[j][i], originalURLs[j], userID.String())
		}

		br := tx.SendBatch(ctx, b)

		var j int
		for k, originalURL := range originalURLs {
			commandTag, err := br.Exec()
			if err != nil {
				return nil, fmt.Errorf("can not set link: %w", err)
			}
			if commandTag.RowsAffected() == 0 {
				originalURLs[j] = originalURL
				j++
				continue
			}
			urls[originalURL] = shortURLs[k][i]
		}

		err = br.Close()
		if err != nil {
			return nil, fmt.Errorf("can not close batch: %w", err)
		}

		if len(originalURLs) == 0 {
			err = tx.Commit(ctx)
			if err != nil {
				return nil, fmt.Errorf("can not commit transaction: %w", err)
			}

			result := make([]string, 0, len(batch))
			for j := range batch {
				result = append(result, urls[batch[j].OriginalURL])
			}

			if originalURLUniqueViolation {
				return result, ErrOriginalURLUniqueViolation
			}

			return result, nil
		}

		originalURLs = originalURLs[:j]
	}

	return nil, ErrReachedMaxGenerationRetries
}

func (d *DBRepository) GetShortLinksOfUser(
	ctx context.Context,
	userID uuid.UUID,
) ([]models.ShortenOfUserResponse, error) {
	rows, err := d.pool.Query(ctx, "SELECT short_url, original_url FROM urls WHERE user_id = $1", userID.String())
	if err != nil {
		return nil, fmt.Errorf("can not get urls of user: %w", err)
	}
	defer rows.Close()

	var urls []models.ShortenOfUserResponse
	for rows.Next() {
		var shortURL string
		var originalURL string
		err = rows.Scan(
			&shortURL,
			&originalURL,
		)
		if err != nil {
			return nil, fmt.Errorf("can not read row: %w", err)
		}

		urls = append(urls, models.ShortenOfUserResponse{
			ShortURL:    shortURL,
			OriginalURL: originalURL,
		})
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("can not read rows: %w", err)
	}

	return urls, nil
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
