package repository

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"

	"github.com/RexArseny/url_shortener/internal/app/models"
	"github.com/google/uuid"
)

const fileMode = 0o600

type URL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	ID          int    `json:"id"`
	UserID      string `json:"user_id"`
}

type LinksWithFile struct {
	*Links
	file      *os.File
	currentID int
}

func NewLinksWithFile(fileStoragePath string) (*LinksWithFile, error) {
	file, err := os.OpenFile(fileStoragePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, fileMode)
	if err != nil {
		return nil, fmt.Errorf("can not open file: %w", err)
	}

	linksWithFile := &LinksWithFile{
		Links:     NewLinks(),
		file:      file,
		currentID: 0,
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var data URL
		err = json.Unmarshal(scanner.Bytes(), &data)
		if err != nil {
			return nil, fmt.Errorf("can not unmarshal data from file: %w", err)
		}

		if _, ok := linksWithFile.Links.shortLinks[data.OriginalURL]; ok {
			return nil, errors.New("duplicate original url in file")
		}
		if _, ok := linksWithFile.Links.originalURLs[data.ShortURL]; ok {
			return nil, errors.New("duplicate short url in file")
		}
		userID, err := uuid.Parse(data.UserID)
		if err != nil {
			userID = uuid.UUID{}
		}
		linksWithFile.Links.shortLinks[data.OriginalURL] = data.ShortURL
		linksWithFile.Links.originalURLs[data.ShortURL] = ShortlURLInfo{
			originalURL: data.OriginalURL,
			userID:      userID,
		}
		linksWithFile.currentID++
	}

	return linksWithFile, nil
}

func (l *LinksWithFile) SetLink(
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
		l.currentID++

		data, err := json.Marshal(URL{
			ID:          l.currentID,
			ShortURL:    shortURL,
			OriginalURL: originalURL,
		})
		if err != nil {
			return nil, fmt.Errorf("can not marshal data: %w", err)
		}
		_, err = fmt.Fprintf(l.file, "%s\n", data)
		if err != nil {
			return nil, fmt.Errorf("can not write data to file: %w", err)
		}

		return &shortURL, nil
	}
	return nil, ErrReachedMaxGenerationRetries
}

func (l *LinksWithFile) SetLinks(
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
			l.currentID++

			data, err := json.Marshal(URL{
				ID:          l.currentID,
				ShortURL:    shortURL,
				OriginalURL: batch[i].OriginalURL,
			})
			if err != nil {
				return nil, fmt.Errorf("can not marshal data: %w", err)
			}
			_, err = fmt.Fprintf(l.file, "%s\n", data)
			if err != nil {
				return nil, fmt.Errorf("can not write data to file: %w", err)
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

func (l *LinksWithFile) Close() error {
	err := l.file.Close()
	if err != nil {
		return fmt.Errorf("can not close file: %w", err)
	}
	return nil
}
