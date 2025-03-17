package repository

import (
	"context"
	"sync"
	"testing"

	"github.com/RexArseny/url_shortener/internal/app/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewLinks(t *testing.T) {
	t.Run("successful initialization", func(t *testing.T) {
		links := NewLinks()

		assert.NotNil(t, links)
		assert.NotNil(t, links.m)
		assert.NotNil(t, links.shortLinks)
		assert.NotNil(t, links.originalURLs)

		assert.IsType(t, &sync.Mutex{}, links.m)

		assert.Empty(t, links.shortLinks)
		assert.Empty(t, links.originalURLs)
	})

	t.Run("unsuccessful initialization", func(t *testing.T) {
		links1 := NewLinks()
		links2 := NewLinks()

		links1.shortLinks["https://example.com"] = "abc123"
		links1.originalURLs["abc123"] = ShortlURLInfo{
			originalURL: "https://example.com",
			userID:      uuid.New(),
			deleted:     false,
		}

		assert.Empty(t, links2.shortLinks)
		assert.Empty(t, links2.originalURLs)
	})
}

func TestLinksGetOriginalURL(t *testing.T) {
	links := NewLinks()

	shortURL := "abc123"
	originalURL := "http://example.com"
	links.originalURLs[shortURL] = ShortlURLInfo{
		originalURL: originalURL,
		deleted:     false,
	}

	result, err := links.GetOriginalURL(context.Background(), shortURL)
	assert.NoError(t, err)
	assert.Equal(t, originalURL, *result)

	_, err = links.GetOriginalURL(context.Background(), "nonexistent")
	assert.Error(t, err)
	assert.Equal(t, "no original url by provided short url", err.Error())

	links.originalURLs[shortURL] = ShortlURLInfo{
		originalURL: originalURL,
		deleted:     true,
	}
	_, err = links.GetOriginalURL(context.Background(), shortURL)
	assert.Error(t, err)
	assert.Equal(t, ErrURLIsDeleted, err)
}

func TestLinksSetLink(t *testing.T) {
	links := NewLinks()

	originalURL := "http://example.com"
	shortURLs := []string{"abc123", "def456"}
	userID := uuid.New()

	result, err := links.SetLink(context.Background(), originalURL, shortURLs, userID)
	assert.NoError(t, err)
	assert.Equal(t, shortURLs[0], *result)

	_, err = links.SetLink(context.Background(), originalURL, shortURLs, userID)
	assert.Error(t, err)
	assert.Equal(t, ErrOriginalURLUniqueViolation, err)

	links.originalURLs[shortURLs[0]] = ShortlURLInfo{originalURL: "http://another.com", userID: userID, deleted: false}
	links.originalURLs[shortURLs[1]] = ShortlURLInfo{originalURL: "http://yetanother.com", userID: userID, deleted: false}
	_, err = links.SetLink(context.Background(), "http://new.com", shortURLs, userID)
	assert.Error(t, err)
	assert.Equal(t, ErrReachedMaxGenerationRetries, err)
}

func TestLinksSetLinks(t *testing.T) {
	links := NewLinks()

	batch := []models.ShortenBatchRequest{
		{OriginalURL: "http://example.com"},
		{OriginalURL: "http://another.com"},
	}
	shortURLs := [][]string{{"abc123"}, {"def456"}}
	userID := uuid.New()

	result, err := links.SetLinks(context.Background(), batch, shortURLs, userID)
	assert.NoError(t, err)
	assert.Equal(t, []string{"abc123", "def456"}, result)

	_, err = links.SetLinks(context.Background(), batch, shortURLs, userID)
	assert.Error(t, err)
	assert.Equal(t, ErrOriginalURLUniqueViolation, err)

	invalidBatch := []models.ShortenBatchRequest{
		{OriginalURL: "invalid-url"},
	}
	_, err = links.SetLinks(context.Background(), invalidBatch, [][]string{{"ghi789"}}, userID)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidURL, err)
}

func TestLinksGetShortLinksOfUser(t *testing.T) {
	links := NewLinks()

	userID := uuid.New()
	shortURL1 := "abc123"
	shortURL2 := "def456"
	links.originalURLs[shortURL1] = ShortlURLInfo{
		originalURL: "http://example.com",
		userID:      userID,
		deleted:     false,
	}
	links.originalURLs[shortURL2] = ShortlURLInfo{
		originalURL: "http://another.com",
		userID:      userID,
		deleted:     false,
	}

	result, err := links.GetShortLinksOfUser(context.Background(), userID)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))
	assert.Contains(t, result, models.ShortenOfUserResponse{ShortURL: shortURL1, OriginalURL: "http://example.com"})
	assert.Contains(t, result, models.ShortenOfUserResponse{ShortURL: shortURL2, OriginalURL: "http://another.com"})

	emptyUserID := uuid.New()
	result, err = links.GetShortLinksOfUser(context.Background(), emptyUserID)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(result))
}

func TestLinksDeleteURLs(t *testing.T) {
	links := NewLinks()

	userID := uuid.New()
	shortURL := "abc123"
	links.originalURLs[shortURL] = ShortlURLInfo{
		originalURL: "http://example.com",
		userID:      userID,
		deleted:     false,
	}

	err := links.DeleteURLs(context.Background(), []string{shortURL}, userID)
	assert.NoError(t, err)
	assert.True(t, links.originalURLs[shortURL].deleted)

	anotherUserID := uuid.New()
	err = links.DeleteURLs(context.Background(), []string{shortURL}, anotherUserID)
	assert.NoError(t, err)
	assert.True(t, links.originalURLs[shortURL].deleted)
}

func TestLinksPing(t *testing.T) {
	links := NewLinks()

	err := links.Ping(context.Background())
	assert.Error(t, err)
	assert.Equal(t, "service in memory storage mode", err.Error())
}
