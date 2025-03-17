package repository

import (
	"context"
	"errors"
	"net/url"
	"sync"

	"github.com/RexArseny/url_shortener/internal/app/models"
	"github.com/google/uuid"
)

// Links is a repository which stores data in memory.
type Links struct {
	m            *sync.Mutex
	shortLinks   map[string]string
	originalURLs map[string]ShortlURLInfo
}

// ShortlURLInfo is a model of URLs which stored in memory.
type ShortlURLInfo struct {
	originalURL string
	userID      uuid.UUID
	deleted     bool
}

// NewLinks create new Links.
func NewLinks() *Links {
	return &Links{
		m:            &sync.Mutex{},
		shortLinks:   make(map[string]string),
		originalURLs: make(map[string]ShortlURLInfo),
	}
}

// GetOriginalURL return original URL by short URL.
func (l *Links) GetOriginalURL(_ context.Context, shortLink string) (*string, error) {
	l.m.Lock()
	defer l.m.Unlock()
	originalURL := l.originalURLs[shortLink]
	if originalURL.originalURL == "" {
		return nil, errors.New("no original url by provided short url")
	}
	if originalURL.deleted {
		return nil, ErrURLIsDeleted
	}
	return &originalURL.originalURL, nil
}

// SetLink add short URL if such does not exist already.
func (l *Links) SetLink(
	_ context.Context,
	originalURL string,
	shortURLs []string,
	userID uuid.UUID,
) (*string, error) {
	l.m.Lock()
	defer l.m.Unlock()
	if shortLink, ok := l.shortLinks[originalURL]; ok {
		return &shortLink, ErrOriginalURLUniqueViolation
	}

	for _, shortURL := range shortURLs {
		if _, ok := l.originalURLs[shortURL]; ok {
			continue
		}
		l.shortLinks[originalURL] = shortURL
		l.originalURLs[shortURL] = ShortlURLInfo{
			originalURL: originalURL,
			userID:      userID,
			deleted:     false,
		}

		return &shortURL, nil
	}
	return nil, ErrReachedMaxGenerationRetries
}

// SetLinks add short URLs if such do not exist already.
func (l *Links) SetLinks(
	_ context.Context,
	batch []models.ShortenBatchRequest,
	shortURLs [][]string,
	userID uuid.UUID,
) ([]string, error) {
	result := make([]string, 0, len(batch))
	l.m.Lock()
	defer l.m.Unlock()

	var originalURLUniqueViolation bool
	for i := range batch {
		_, err := url.ParseRequestURI(batch[i].OriginalURL)
		if err != nil {
			return nil, ErrInvalidURL
		}

		if shortLink, ok := l.shortLinks[batch[i].OriginalURL]; ok {
			originalURLUniqueViolation = true
			result = append(result, shortLink)
			continue
		}

		var generated bool
		var shortURL string
		for _, shortURL = range shortURLs[i] {
			if _, ok := l.originalURLs[shortURL]; ok {
				continue
			}
			l.shortLinks[batch[i].OriginalURL] = shortURL
			l.originalURLs[shortURL] = ShortlURLInfo{
				originalURL: batch[i].OriginalURL,
				userID:      userID,
				deleted:     false,
			}

			generated = true
			break
		}

		if !generated {
			return nil, ErrReachedMaxGenerationRetries
		}
		result = append(result, shortURL)
	}

	if originalURLUniqueViolation {
		return result, ErrOriginalURLUniqueViolation
	}

	return result, nil
}

// GetShortLinksOfUser return URLs of user if such exist.
func (l *Links) GetShortLinksOfUser(_ context.Context, userID uuid.UUID) ([]models.ShortenOfUserResponse, error) {
	l.m.Lock()
	defer l.m.Unlock()

	var urls []models.ShortenOfUserResponse
	for shortURL, originalURLInfo := range l.originalURLs {
		if originalURLInfo.userID == userID {
			urls = append(urls, models.ShortenOfUserResponse{
				ShortURL:    shortURL,
				OriginalURL: originalURLInfo.originalURL,
			})
		}
	}

	return urls, nil
}

// DeleteURLs delete URLs.
func (l *Links) DeleteURLs(_ context.Context, urls []string, userID uuid.UUID) error {
	l.m.Lock()
	defer l.m.Unlock()

	for _, shortURL := range urls {
		if shortlURLInfo, ok := l.originalURLs[shortURL]; ok {
			if shortlURLInfo.userID == userID {
				shortlURLInfo.deleted = true
				l.originalURLs[shortURL] = shortlURLInfo
			}
		}
	}

	return nil
}

// Ping return info about connection.
func (l *Links) Ping(_ context.Context) error {
	return nil
}
