package repository

import (
	"context"
	"os"
	"testing"

	"github.com/RexArseny/url_shortener/internal/app/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestLinksWithFileSetLink(t *testing.T) {
	tmpFile, err := os.CreateTemp("./", "*.test")
	assert.NoError(t, err)

	linksWithFile, err := NewLinksWithFile(tmpFile.Name())
	assert.NoError(t, err)

	defer func() {
		err = linksWithFile.Close()
		assert.NoError(t, err)
		err = tmpFile.Close()
		assert.NoError(t, err)
		err = os.Remove(tmpFile.Name())
		assert.NoError(t, err)
	}()

	originalURL := "http://example.com"
	shortURLs := []string{"abc123"}
	userID := uuid.New()

	result, err := linksWithFile.SetLink(context.Background(), originalURL, shortURLs, userID)
	assert.NoError(t, err)
	assert.Equal(t, shortURLs[0], *result)

	fileContent, err := os.ReadFile(tmpFile.Name())
	assert.NoError(t, err)
	assert.Contains(t, string(fileContent), originalURL)
	assert.Contains(t, string(fileContent), shortURLs[0])
}

func TestSetLinks(t *testing.T) {
	tmpFile, err := os.CreateTemp("./", "*.test")
	assert.NoError(t, err)

	linksWithFile, err := NewLinksWithFile(tmpFile.Name())
	assert.NoError(t, err)

	defer func() {
		err = linksWithFile.Close()
		assert.NoError(t, err)
		err = tmpFile.Close()
		assert.NoError(t, err)
		err = os.Remove(tmpFile.Name())
		assert.NoError(t, err)
	}()

	batch := []models.ShortenBatchRequest{
		{OriginalURL: "http://example.com"},
		{OriginalURL: "http://another.com"},
	}
	shortURLs := [][]string{{"abc123"}, {"def456"}}
	userID := uuid.New()

	result, err := linksWithFile.SetLinks(context.Background(), batch, shortURLs, userID)
	assert.NoError(t, err)
	assert.Equal(t, []string{"abc123", "def456"}, result)

	fileContent, err := os.ReadFile(tmpFile.Name())
	assert.NoError(t, err)
	assert.Contains(t, string(fileContent), batch[0].OriginalURL)
	assert.Contains(t, string(fileContent), batch[1].OriginalURL)
}

func TestDeleteURLs(t *testing.T) {
	tmpFile, err := os.CreateTemp("./", "*.test")
	assert.NoError(t, err)

	linksWithFile, err := NewLinksWithFile(tmpFile.Name())
	assert.NoError(t, err)

	defer func() {
		err = linksWithFile.Close()
		assert.NoError(t, err)
		err = tmpFile.Close()
		assert.NoError(t, err)
		err = os.Remove(tmpFile.Name())
		assert.NoError(t, err)
	}()

	originalURL := "http://example.com"
	shortURL := "abc123"
	userID := uuid.New()
	_, err = linksWithFile.SetLink(context.Background(), originalURL, []string{shortURL}, userID)
	assert.NoError(t, err)

	err = linksWithFile.DeleteURLs(context.Background(), []string{shortURL}, userID)
	assert.NoError(t, err)

	assert.True(t, linksWithFile.Links.originalURLs[shortURL].deleted)

	fileContent, err := os.ReadFile(tmpFile.Name())
	assert.NoError(t, err)
	assert.Contains(t, string(fileContent), `"deleted":true`)
}

func TestClose(t *testing.T) {
	tmpFile, err := os.CreateTemp("./", "*.test")
	assert.NoError(t, err)

	linksWithFile, err := NewLinksWithFile(tmpFile.Name())
	assert.NoError(t, err)

	defer func() {
		err = tmpFile.Close()
		assert.NoError(t, err)
		err = os.Remove(tmpFile.Name())
		assert.NoError(t, err)
	}()

	err = linksWithFile.Close()
	assert.NoError(t, err)

	_, err = linksWithFile.file.WriteString("test")
	assert.Error(t, err)
}
