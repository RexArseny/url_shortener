package usecases

import (
	"context"
	"net/url"
	"path"
	"testing"

	"github.com/RexArseny/url_shortener/internal/app/config"
	"github.com/RexArseny/url_shortener/internal/app/repository"
	"github.com/stretchr/testify/assert"
)

func TestCreateShortLink(t *testing.T) {
	interactor := NewInteractor(config.DefaultBasicPath, repository.NewLinks())

	result1, err := interactor.CreateShortLink(context.Background(), "")
	assert.Error(t, err)
	assert.Nil(t, result1)

	result2, err := interactor.CreateShortLink(context.Background(), "abc")
	assert.Error(t, err)
	assert.Nil(t, result2)

	result3, err := interactor.CreateShortLink(context.Background(), "https://ya.ru")
	assert.NoError(t, err)
	assert.NotEmpty(t, result3)
	parsedURL, err := url.ParseRequestURI(*result3)
	assert.NoError(t, err)
	assert.NotNil(t, parsedURL)

	result4, err := interactor.CreateShortLink(context.Background(), "https://ya.ru")
	assert.NoError(t, err)
	assert.NotEmpty(t, result4)
	assert.Equal(t, result3, result4)
}

func TestGetShortLink(t *testing.T) {
	interactor := NewInteractor(config.DefaultBasicPath, repository.NewLinks())

	link, err := interactor.CreateShortLink(context.Background(), "context.Background(),https://ya.ru")
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
