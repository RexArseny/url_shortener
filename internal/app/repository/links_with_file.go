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
)

const fileMode = 0o600

type URL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	ID          int    `json:"id"`
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
		linksWithFile.Links.shortLinks[data.OriginalURL] = data.ShortURL
		linksWithFile.Links.originalURLs[data.ShortURL] = data.OriginalURL
		linksWithFile.currentID++
	}

	return linksWithFile, nil
}

func (l *LinksWithFile) SetLink(_ context.Context, originalURL string) (*string, error) {
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

		data, err := json.Marshal(URL{
			ID:          l.currentID,
			ShortURL:    shortLink,
			OriginalURL: originalURL,
		})
		if err != nil {
			return nil, fmt.Errorf("can not marshal data: %w", err)
		}
		_, err = fmt.Fprintf(l.file, "%s\n", data)
		if err != nil {
			return nil, fmt.Errorf("can not write data to file: %w", err)
		}

		return &shortLink, nil
	}
	return nil, errors.New("reached max generation retries")
}

func (l *LinksWithFile) SetLinks(_ context.Context, batch []models.ShortenBatchRequest) ([]string, error) {
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
			l.currentID++

			data, err := json.Marshal(URL{
				ID:          l.currentID,
				ShortURL:    shortLink,
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
			return nil, errors.New("reached max generation retries")
		}
		result = append(result, shortLink)
	}

	if originalURLUniqueViolation {
		return result, models.ErrOriginalURLUniqueViolation
	}

	return result, nil
}

func (l *LinksWithFile) Ping(_ context.Context) error {
	return errors.New("service in file storage mode")
}

func (l *LinksWithFile) Close() error {
	err := l.file.Close()
	if err != nil {
		return fmt.Errorf("can not close file: %w", err)
	}
	return nil
}
