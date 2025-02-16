package usecases

import (
	"context"
	"net/url"
	"path"
	"testing"

	"github.com/RexArseny/url_shortener/internal/app/config"
	"github.com/RexArseny/url_shortener/internal/app/logger"
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

	for i := 0; i < b.N; i++ {
		link, err := interactor.CreateShortLink(context.Background(), "", userID)
		assert.NoError(b, err)
		assert.NotEmpty(b, link)
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

	for i := 0; i < b.N; i++ {
		result1, err := interactor.GetShortLink(context.Background(), "")
		assert.Error(b, err)
		assert.Nil(b, result1)
	}
}
