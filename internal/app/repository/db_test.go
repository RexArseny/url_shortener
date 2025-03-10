package repository

import (
	"context"
	"testing"

	"github.com/RexArseny/url_shortener/internal/app/logger"
	"github.com/RexArseny/url_shortener/internal/app/models"
	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
)

func TestDBRepositoryGetOriginalURL(t *testing.T) {
	testLogger, err := logger.InitLogger()
	assert.NoError(t, err)

	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := &DBRepository{
		logger: testLogger.Named("repository"),
		pool:   mock,
	}

	shortLink := "abc123"
	originalURL := "http://example.com"
	deleted := false

	mock.ExpectQuery("SELECT original_url, deleted FROM urls WHERE short_url=").
		WithArgs(shortLink).
		WillReturnRows(pgxmock.NewRows([]string{"original_url", "deleted"}).
			AddRow(originalURL, deleted))

	result, err := repo.GetOriginalURL(context.Background(), shortLink)
	assert.NoError(t, err)
	assert.Equal(t, originalURL, *result)
}

func TestDBRepositorySetLink(t *testing.T) {
	testLogger, err := logger.InitLogger()
	assert.NoError(t, err)

	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := &DBRepository{
		logger: testLogger.Named("repository"),
		pool:   mock,
	}

	originalURL := "http://example.com"
	shortURL := "abc123"
	userID := uuid.New()

	mock.ExpectQuery("INSERT INTO urls").
		WithArgs(shortURL, originalURL, userID).
		WillReturnRows(pgxmock.NewRows([]string{"short_url"}).AddRow(shortURL))

	result, err := repo.SetLink(context.Background(), originalURL, []string{shortURL}, userID)
	assert.NoError(t, err)
	assert.Equal(t, shortURL, *result)
}

func TestDBRepositorySetLinks(t *testing.T) {
	testLogger, err := logger.InitLogger()
	assert.NoError(t, err)

	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := &DBRepository{
		logger: testLogger.Named("repository"),
		pool:   mock,
	}

	batch := []models.ShortenBatchRequest{
		{OriginalURL: "http://example.com"},
	}
	shortURLs := [][]string{{"abc123"}}
	userID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT short_url, original_url FROM urls WHERE original_url = ANY").
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{"short_url", "original_url"}))
	mock.ExpectBatch().ExpectExec("INSERT INTO urls").
		WithArgs(shortURLs[0][0], batch[0].OriginalURL, userID).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectCommit()

	result, err := repo.SetLinks(context.Background(), batch, shortURLs, userID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestDBRepositoryGetShortLinksOfUser(t *testing.T) {
	testLogger, err := logger.InitLogger()
	assert.NoError(t, err)

	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := &DBRepository{
		logger: testLogger.Named("repository"),
		pool:   mock,
	}

	userID := uuid.New()

	mock.ExpectQuery("SELECT short_url, original_url FROM urls WHERE user_id =").
		WithArgs(userID).
		WillReturnRows(pgxmock.NewRows([]string{"short_url", "original_url"}).
			AddRow("abc123", "http://example.com"))

	result, err := repo.GetShortLinksOfUser(context.Background(), userID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestDBRepositoryDeleteURLs(t *testing.T) {
	testLogger, err := logger.InitLogger()
	assert.NoError(t, err)

	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := &DBRepository{
		logger: testLogger.Named("repository"),
		pool:   mock,
	}

	urls := []string{"abc123"}
	userID := uuid.New()

	mock.ExpectExec("INSERT INTO urls_for_delete").
		WithArgs(urls, userID).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = repo.DeleteURLs(context.Background(), urls, userID)
	assert.NoError(t, err)
}

func TestDBRepositoryDeleteURLsInDB(t *testing.T) {
	testLogger, err := logger.InitLogger()
	assert.NoError(t, err)

	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := &DBRepository{
		logger: testLogger.Named("repository"),
		pool:   mock,
	}

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id, urls, user_id FROM urls_for_delete LIMIT 1").
		WillReturnRows(pgxmock.NewRows([]string{"id", "urls", "user_id"}).
			AddRow(1, []string{"abc123"}, uuid.New()))
	mock.ExpectExec("UPDATE urls SET deleted = true").
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectExec("DELETE FROM urls_for_delete").
		WithArgs(1).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))
	mock.ExpectCommit()

	err = repo.DeleteURLsInDB(context.Background())
	assert.NoError(t, err)
}

func TestDBRepositoryPing(t *testing.T) {
	testLogger, err := logger.InitLogger()
	assert.NoError(t, err)

	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := &DBRepository{
		logger: testLogger.Named("repository"),
		pool:   mock,
	}

	mock.ExpectPing()

	err = repo.Ping(context.Background())
	assert.NoError(t, err)
}

func TestDBRepositoryClose(t *testing.T) {
	testLogger, err := logger.InitLogger()
	assert.NoError(t, err)

	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := &DBRepository{
		logger: testLogger.Named("repository"),
		pool:   mock,
	}

	mock.ExpectClose()

	repo.Close()
}
