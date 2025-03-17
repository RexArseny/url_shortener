package repository

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/RexArseny/url_shortener/internal/app/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewLinksWithFile(t *testing.T) {
	t.Run("successful creation with empty file", func(t *testing.T) {
		file, err := os.CreateTemp("", "testfile")
		assert.NoError(t, err)

		linksWithFile, err := NewLinksWithFile(file.Name())
		assert.NoError(t, err)
		assert.NotNil(t, linksWithFile)
		assert.Equal(t, 0, linksWithFile.currentID)

		err = linksWithFile.Close()
		assert.NoError(t, err)
		err = file.Close()
		assert.NoError(t, err)
		err = os.Remove(file.Name())
		assert.NoError(t, err)
	})

	t.Run("successful creation with valid data in file", func(t *testing.T) {
		file, err := os.CreateTemp("", "testfile")
		assert.NoError(t, err)

		data := URL{
			OriginalURL: "https://example.com",
			ShortURL:    "abc123",
			UserID:      uuid.New().String(),
			Deleted:     false,
		}
		jsonData, err := json.Marshal(data)
		assert.NoError(t, err)

		_, err = file.WriteString(string(jsonData) + "\n")
		assert.NoError(t, err)
		err = file.Close()
		assert.NoError(t, err)

		linksWithFile, err := NewLinksWithFile(file.Name())
		assert.NoError(t, err)
		assert.NotNil(t, linksWithFile)
		assert.Equal(t, 1, linksWithFile.currentID)

		assert.Equal(t, data.ShortURL, linksWithFile.Links.shortLinks[data.OriginalURL])
		assert.Equal(t, data.OriginalURL, linksWithFile.Links.originalURLs[data.ShortURL].originalURL)

		err = linksWithFile.Close()
		assert.NoError(t, err)
		err = os.Remove(file.Name())
		assert.NoError(t, err)
	})

	t.Run("file open error", func(t *testing.T) {
		linksWithFile, err := NewLinksWithFile("/invalid/path")
		assert.Error(t, err)
		assert.Nil(t, linksWithFile)
		assert.Contains(t, err.Error(), "can not open file")
	})

	t.Run("invalid JSON data in file", func(t *testing.T) {
		file, err := os.CreateTemp("", "testfile")
		assert.NoError(t, err)

		_, err = file.WriteString("invalid json data\n")
		assert.NoError(t, err)
		err = file.Close()
		assert.NoError(t, err)

		linksWithFile, err := NewLinksWithFile(file.Name())
		assert.Error(t, err)
		assert.Nil(t, linksWithFile)
		assert.Contains(t, err.Error(), "can not unmarshal data from file")
	})

	t.Run("duplicate original URL in file", func(t *testing.T) {
		file, err := os.CreateTemp("", "testfile")
		assert.NoError(t, err)

		data1 := URL{
			OriginalURL: "https://example.com",
			ShortURL:    "abc123",
			UserID:      uuid.New().String(),
			Deleted:     false,
		}
		data2 := URL{
			OriginalURL: "https://example.com",
			ShortURL:    "def456",
			UserID:      uuid.New().String(),
			Deleted:     false,
		}
		jsonData1, err := json.Marshal(data1)
		assert.NoError(t, err)
		jsonData2, err := json.Marshal(data2)
		assert.NoError(t, err)

		_, err = file.WriteString(string(jsonData1) + "\n" + string(jsonData2) + "\n")
		assert.NoError(t, err)
		err = file.Close()
		assert.NoError(t, err)

		linksWithFile, err := NewLinksWithFile(file.Name())
		assert.Error(t, err)
		assert.Nil(t, linksWithFile)
		assert.Equal(t, "duplicate original url in file", err.Error())
	})

	t.Run("duplicate short URL in file", func(t *testing.T) {
		file, err := os.CreateTemp("", "testfile")
		assert.NoError(t, err)

		data1 := URL{
			OriginalURL: "https://example.com",
			ShortURL:    "abc123",
			UserID:      uuid.New().String(),
			Deleted:     false,
		}
		data2 := URL{
			OriginalURL: "https://another.com",
			ShortURL:    "abc123",
			UserID:      uuid.New().String(),
			Deleted:     false,
		}
		jsonData1, err := json.Marshal(data1)
		assert.NoError(t, err)
		jsonData2, err := json.Marshal(data2)
		assert.NoError(t, err)

		_, err = file.WriteString(string(jsonData1) + "\n" + string(jsonData2) + "\n")
		assert.NoError(t, err)
		err = file.Close()
		assert.NoError(t, err)

		linksWithFile, err := NewLinksWithFile(file.Name())
		assert.Error(t, err)
		assert.Nil(t, linksWithFile)
		assert.Equal(t, "duplicate short url in file", err.Error())
	})

	t.Run("invalid user ID in file", func(t *testing.T) {
		file, err := os.CreateTemp("", "testfile")
		assert.NoError(t, err)

		data := URL{
			OriginalURL: "https://example.com",
			ShortURL:    "abc123",
			UserID:      "invalid-uuid",
			Deleted:     false,
		}
		jsonData, err := json.Marshal(data)
		assert.NoError(t, err)

		_, err = file.WriteString(string(jsonData) + "\n")
		assert.NoError(t, err)
		err = file.Close()
		assert.NoError(t, err)

		linksWithFile, err := NewLinksWithFile(file.Name())
		assert.NoError(t, err)
		assert.NotNil(t, linksWithFile)

		assert.Equal(t, uuid.UUID{}, linksWithFile.Links.originalURLs[data.ShortURL].userID)
	})
}

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
