package usecases

import (
	"context"
	"net/url"
	"path"
	"testing"
	"time"

	"github.com/RexArseny/url_shortener/internal/app/config"
	"github.com/RexArseny/url_shortener/internal/app/logger"
	"github.com/RexArseny/url_shortener/internal/app/models"
	"github.com/RexArseny/url_shortener/internal/app/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCreateShortLink(t *testing.T) {
	ctx := context.Background()

	testLogger, err := logger.InitLogger()
	assert.NoError(t, err)

	userID := uuid.New()
	interactor := NewInteractor(
		ctx,
		testLogger.Named("interactor"),
		config.DefaultBasicPath,
		repository.NewLinks(),
	)

	result1, err := interactor.CreateShortLink(context.Background(), "", userID)
	assert.Error(t, err)
	assert.Nil(t, result1)

	result2, err := interactor.CreateShortLink(context.Background(), "abc", userID)
	assert.Error(t, err)
	assert.Nil(t, result2)

	result3, err := interactor.CreateShortLink(context.Background(), "https://ya.ru", userID)
	assert.NoError(t, err)
	assert.NotEmpty(t, result3)
	parsedURL, err := url.ParseRequestURI(*result3)
	assert.NoError(t, err)
	assert.NotNil(t, parsedURL)

	result4, err := interactor.CreateShortLink(context.Background(), "https://ya.ru", userID)
	assert.Error(t, err)
	assert.NotNil(t, result4)
}

func TestCreateShortLinks(t *testing.T) {
	ctx := context.Background()

	testLogger, err := logger.InitLogger()
	assert.NoError(t, err)

	userID := uuid.New()
	interactor := NewInteractor(
		ctx,
		testLogger.Named("interactor"),
		config.DefaultBasicPath,
		repository.NewLinks(),
	)

	result1, err := interactor.CreateShortLinks(context.Background(), nil, userID)
	assert.NoError(t, err)
	assert.Empty(t, result1)

	result2, err := interactor.CreateShortLinks(context.Background(), []models.ShortenBatchRequest{}, userID)
	assert.NoError(t, err)
	assert.Empty(t, result2)

	result3, err := interactor.CreateShortLinks(context.Background(), []models.ShortenBatchRequest{{
		CorrelationID: "1",
		OriginalURL:   "abc",
	}}, userID)
	assert.Error(t, err)
	assert.Nil(t, result3)

	result4, err := interactor.CreateShortLinks(context.Background(), []models.ShortenBatchRequest{{
		CorrelationID: "2",
		OriginalURL:   "https://ya.ru",
	}}, userID)
	assert.NoError(t, err)
	assert.NotEmpty(t, result4)
	parsedURL, err := url.ParseRequestURI(result4[0].ShortURL)
	assert.NoError(t, err)
	assert.NotNil(t, parsedURL)

	result5, err := interactor.CreateShortLinks(context.Background(), []models.ShortenBatchRequest{{
		CorrelationID: "2",
		OriginalURL:   "https://ya.ru",
	}}, userID)
	assert.Error(t, err)
	assert.NotNil(t, result5)
}

func TestGetShortLink(t *testing.T) {
	ctx := context.Background()

	testLogger, err := logger.InitLogger()
	assert.NoError(t, err)

	userID := uuid.New()
	interactor := NewInteractor(
		ctx,
		testLogger.Named("interactor"),
		config.DefaultBasicPath,
		repository.NewLinks(),
	)

	link, err := interactor.CreateShortLink(context.Background(), "https://ya.ru", userID)
	assert.NoError(t, err)
	assert.NotEmpty(t, link)

	parsedURL, err := url.ParseRequestURI(*link)
	assert.NoError(t, err)
	assert.NotEmpty(t, parsedURL)

	result1, err := interactor.GetShortLink(context.Background(), "")
	assert.Error(t, err)
	assert.Nil(t, result1)

	result2, err := interactor.GetShortLink(context.Background(), "abc")
	assert.Error(t, err)
	assert.Nil(t, result2)

	result3, err := interactor.GetShortLink(context.Background(), path.Base(parsedURL.Path))
	assert.NoError(t, err)
	assert.NotEmpty(t, result3)
	assert.Equal(t, "https://ya.ru", *result3)
}

func TestGetShortLinksOfUser(t *testing.T) {
	ctx := context.Background()

	testLogger, err := logger.InitLogger()
	assert.NoError(t, err)

	userID := uuid.New()
	interactor := NewInteractor(
		ctx,
		testLogger.Named("interactor"),
		config.DefaultBasicPath,
		repository.NewLinks(),
	)

	link, err := interactor.CreateShortLink(context.Background(), "https://ya.ru", userID)
	assert.NoError(t, err)
	assert.NotEmpty(t, link)

	parsedURL, err := url.ParseRequestURI(*link)
	assert.NoError(t, err)
	assert.NotEmpty(t, parsedURL)

	result1, err := interactor.GetShortLinksOfUser(context.Background(), userID)
	assert.NoError(t, err)
	assert.NotEmpty(t, result1)
	assert.Equal(t, "https://ya.ru", result1[0].OriginalURL)
}

func TestDeleteURLs(t *testing.T) {
	ctx := context.Background()

	testLogger, err := logger.InitLogger()
	assert.NoError(t, err)

	userID := uuid.New()
	interactor := NewInteractor(
		ctx,
		testLogger.Named("interactor"),
		config.DefaultBasicPath,
		repository.NewLinks(),
	)

	link, err := interactor.CreateShortLink(context.Background(), "https://ya.ru", userID)
	assert.NoError(t, err)
	assert.NotEmpty(t, link)

	parsedURL, err := url.ParseRequestURI(*link)
	assert.NoError(t, err)
	assert.NotEmpty(t, parsedURL)

	err = interactor.DeleteURLs(context.Background(), []string{path.Base(parsedURL.Path)}, userID)
	assert.NoError(t, err)

	time.Sleep(time.Second)

	result1, err := interactor.GetShortLink(context.Background(), path.Base(parsedURL.Path))
	assert.Error(t, err)
	assert.Nil(t, result1)
}

func TestPingDB(t *testing.T) {
	ctx := context.Background()

	testLogger, err := logger.InitLogger()
	assert.NoError(t, err)

	interactor := NewInteractor(
		ctx,
		testLogger.Named("interactor"),
		config.DefaultBasicPath,
		repository.NewLinks(),
	)

	err = interactor.PingDB(context.Background())
	assert.NoError(t, err)
}

func BenchmarkCreateShortLink(b *testing.B) {
	ctx := context.Background()

	testLogger, err := logger.InitLogger()
	assert.NoError(b, err)

	userID := uuid.New()
	interactor := NewInteractor(
		ctx,
		testLogger.Named("interactor"),
		config.DefaultBasicPath,
		repository.NewLinks(),
	)

	for range b.N {
		result1, err := interactor.CreateShortLink(context.Background(), "", userID)
		assert.Error(b, err)
		assert.Nil(b, result1)
	}
}

func BenchmarkGetShortLink(b *testing.B) {
	ctx := context.Background()

	testLogger, err := logger.InitLogger()
	assert.NoError(b, err)

	userID := uuid.New()
	interactor := NewInteractor(
		ctx,
		testLogger.Named("interactor"),
		config.DefaultBasicPath,
		repository.NewLinks(),
	)

	link, err := interactor.CreateShortLink(context.Background(), "https://ya.ru", userID)
	assert.NoError(b, err)
	assert.NotEmpty(b, link)

	parsedURL, err := url.ParseRequestURI(*link)
	assert.NoError(b, err)
	assert.NotEmpty(b, parsedURL)

	for range b.N {
		result1, err := interactor.GetShortLink(context.Background(), "")
		assert.Error(b, err)
		assert.Nil(b, result1)
	}
}
