package repository

import (
	"context"
	"errors"
	"net/url"
	"sync"

	"github.com/RexArseny/url_shortener/internal/app/models"
	"github.com/google/uuid"
)

type Links struct {
	m            *sync.Mutex
	shortLinks   map[string]string
	originalURLs map[string]ShortlURLInfo
}

type ShortlURLInfo struct {
	originalURL string
	userID      uuid.UUID
}

func NewLinks() *Links {
	return &Links{
		m:            &sync.Mutex{},
		shortLinks:   make(map[string]string),
		originalURLs: make(map[string]ShortlURLInfo),
	}
}

func (l *Links) GetOriginalURL(_ context.Context, shortLink string) (*string, error) {
	l.m.Lock()
	defer l.m.Unlock()
	originalURL := l.originalURLs[shortLink]
	if originalURL.originalURL == "" {
		return nil, errors.New("no original url by provided short url")
	}
	return &originalURL.originalURL, nil
}

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
		}

		return &shortURL, nil
	}
	return nil, ErrReachedMaxGenerationRetries
}

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

func (l *Links) GetShortLinksOfUser(ctx context.Context, userID uuid.UUID) ([]models.ShortenOfUserResponse, error) {
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

func (l *Links) Ping(_ context.Context) error {
	return errors.New("service in memory storage mode")
}
