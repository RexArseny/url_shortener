package repository

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sync"

	"github.com/RexArseny/url_shortener/internal/app/models"
)

type Links struct {
	m            *sync.Mutex
	shortLinks   map[string]string
	originalURLs map[string]string
}

func NewLinks() *Links {
	return &Links{
		m:            &sync.Mutex{},
		shortLinks:   make(map[string]string),
		originalURLs: make(map[string]string),
	}
}

func (l *Links) GetOriginalURL(_ context.Context, shortLink string) (*string, error) {
	l.m.Lock()
	defer l.m.Unlock()
	originalURL := l.originalURLs[shortLink]
	if originalURL == "" {
		return nil, errors.New("no original url by provided short url")
	}
	return &originalURL, nil
}

func (l *Links) SetLink(_ context.Context, originalURL string) (*string, error) {
	l.m.Lock()
	defer l.m.Unlock()
	if shortLink, ok := l.shortLinks[originalURL]; ok {
		return &shortLink, models.ErrOriginalURLUniqueViolation
	}

	var retry int
	for retry < linkGenerationRetries {
		shortLink := generatePath()

		if _, ok := l.originalURLs[shortLink]; ok {
			retry++
			continue
		}
		l.shortLinks[originalURL] = shortLink
		l.originalURLs[shortLink] = originalURL

		return &shortLink, nil
	}
	return nil, errors.New("reached max generation retries")
}

func (l *Links) SetLinks(_ context.Context, batch []models.ShortenBatchRequest) ([]string, error) {
	result := make([]string, 0, len(batch))
	l.m.Lock()
	defer l.m.Unlock()

	var originalURLUniqueViolation bool
	for i := range batch {
		_, err := url.ParseRequestURI(batch[i].OriginalURL)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", models.ErrInvalidURL, err)
		}

		if shortLink, ok := l.shortLinks[batch[i].OriginalURL]; ok {
			originalURLUniqueViolation = true
			result = append(result, shortLink)
			continue
		}

		var retry int
		var generated bool
		var shortLink string
		for retry < linkGenerationRetries {
			shortLink = generatePath()

			if _, ok := l.originalURLs[shortLink]; ok {
				retry++
				continue
			}
			l.shortLinks[batch[i].OriginalURL] = shortLink
			l.originalURLs[shortLink] = batch[i].OriginalURL

			generated = true
			break
		}

		if !generated {
			return nil, errors.New("reached max generation retries")
		}
		result = append(result, shortLink)
	}

	if originalURLUniqueViolation {
		return result, models.ErrOriginalURLUniqueViolation
	}

	return result, nil
}

func (l *Links) Ping(_ context.Context) error {
	return errors.New("service in memory storage mode")
}
