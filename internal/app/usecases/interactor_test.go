package usecases

import (
	"net/url"
	"path"
	"testing"

	"github.com/RexArseny/url_shortener/internal/app/config"
	"github.com/RexArseny/url_shortener/internal/app/models"
	"github.com/stretchr/testify/assert"
)

func TestCreateShortLink(t *testing.T) {
	interactor := NewInteractor(config.DefaultBasicPath, models.NewLinks())

	result1, err := interactor.CreateShortLink("")
	assert.Error(t, err)
	assert.Nil(t, result1)

	result2, err := interactor.CreateShortLink("abc")
	assert.Error(t, err)
	assert.Nil(t, result2)

	result3, err := interactor.CreateShortLink("https://ya.ru")
	assert.NoError(t, err)
	assert.NotEmpty(t, result3)
	parsedURL, err := url.ParseRequestURI(*result3)
	assert.NoError(t, err)
	assert.NotNil(t, parsedURL)

	result4, err := interactor.CreateShortLink("https://ya.ru")
	assert.NoError(t, err)
	assert.NotEmpty(t, result4)
	assert.Equal(t, result3, result4)
}

func TestGetShortLink(t *testing.T) {
	interactor := NewInteractor(config.DefaultBasicPath, models.NewLinks())

	link, err := interactor.CreateShortLink("https://ya.ru")
	assert.NoError(t, err)
	assert.NotEmpty(t, link)

	parsedURL, err := url.ParseRequestURI(*link)
	assert.NoError(t, err)
	assert.NotEmpty(t, parsedURL)

	result1, err := interactor.GetShortLink("")
	assert.Error(t, err)
	assert.Nil(t, result1)

	result2, err := interactor.GetShortLink("abc")
	assert.Error(t, err)
	assert.Nil(t, result2)

	result3, err := interactor.GetShortLink(path.Base(parsedURL.Path))
	assert.NoError(t, err)
	assert.NotEmpty(t, result3)
	assert.Equal(t, "https://ya.ru", *result3)
}
