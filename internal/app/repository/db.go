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

// DBRepository is a repository which stores data in database.
type DBRepository struct {
	logger *zap.Logger
	pool   IPool
}

// NewDBRepository create new DBRepository.
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

// GetOriginalURL return original URL by short URL.
func (d *DBRepository) GetOriginalURL(ctx context.Context, shortLink string) (*string, error) {
	var originalURL string
	var deleted bool
	err := d.pool.QueryRow(ctx, `SELECT original_url, deleted 
								FROM urls WHERE short_url=$1`, shortLink).Scan(&originalURL, &deleted)
	if err != nil {
		return nil, fmt.Errorf("can not get original url: %w", err)
	}
	if deleted {
		return nil, ErrURLIsDeleted
	}
	return &originalURL, nil
}

// SetLink add short URL if such does not exist already.
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
									RETURNING short_url`, shortURL, originalURL, userID).Scan(&link)
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

// SetLinks add short URLs if such do not exist already.
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
			_, err = url.ParseRequestURI(originalURLs[j])
			if err != nil {
				return nil, ErrInvalidURL
			}

			b.Queue(`INSERT INTO urls (short_url, original_url, user_id) 
			VALUES ($1, $2, $3) 
			ON CONFLICT (short_url) 
			DO NOTHING`, shortURLs[j][i], originalURLs[j], userID)
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

		if j == 0 {
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

// GetShortLinksOfUser return URLs of user if such exist.
func (d *DBRepository) GetShortLinksOfUser(
	ctx context.Context,
	userID uuid.UUID,
) ([]models.ShortenOfUserResponse, error) {
	rows, err := d.pool.Query(ctx, "SELECT short_url, original_url FROM urls WHERE user_id = $1", userID)
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

// DeleteURLs add URLs to deletion queue.
func (d *DBRepository) DeleteURLs(ctx context.Context, urls []string, userID uuid.UUID) error {
	_, err := d.pool.Exec(ctx, "INSERT INTO urls_for_delete (urls, user_id) VALUES ($1, $2)", urls, userID)
	if err != nil {
		return fmt.Errorf("can not add urls for delete: %w", err)
	}

	return nil
}

// DeleteURLsInDB get and delete URLs from deletion queue.
func (d *DBRepository) DeleteURLsInDB(ctx context.Context) error {
	tx, err := d.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("can not start transaction: %w", err)
	}
	defer func() {
		err = tx.Rollback(ctx)
		if err != nil && !strings.Contains(err.Error(), "tx is closed") {
			d.logger.Error("Can not rollback transaction", zap.Error(err))
		}
	}()

	var id int
	var urls []string
	var userID uuid.UUID
	err = tx.QueryRow(ctx, "SELECT id, urls, user_id FROM urls_for_delete LIMIT 1").Scan(&id, &urls, &userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("can not get urls for delete: %w", err)
	}

	_, err = tx.Exec(ctx, `UPDATE urls SET deleted = true 
								WHERE user_id = $1 AND short_url = ANY ($2)`, userID, urls)
	if err != nil {
		return fmt.Errorf("can not delete urls: %w", err)
	}

	_, err = tx.Exec(ctx, "DELETE FROM urls_for_delete WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("can not clear urls for delete: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("can not commit transaction: %w", err)
	}

	return nil
}

// Ping check connection with database.
func (d *DBRepository) Ping(ctx context.Context) error {
	err := d.pool.Ping(ctx)
	if err != nil {
		return fmt.Errorf("can not ping PostgreSQL server: %w", err)
	}
	return nil
}

// Stats return statistic of shortened urls and users in service.
func (d *DBRepository) Stats(ctx context.Context) (*models.Stats, error) {
	var urls int
	var users int
	err := d.pool.QueryRow(ctx, "SELECT COUNT(DISTINCT short_url), COUNT(DISTINCT user_id) FROM urls").Scan(&urls, &users)
	if err != nil {
		return nil, fmt.Errorf("can not get stats: %w", err)
	}

	return &models.Stats{
		URLs:  urls,
		Users: users,
	}, nil
}

// Close all connections with database.
func (d *DBRepository) Close() {
	d.pool.Close()
}
