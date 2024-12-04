package usecases

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/url"

	"github.com/RexArseny/url_shortener/internal/app/models"
	"github.com/RexArseny/url_shortener/internal/app/repository"
)

const (
	shortLinkPathLength   = 8
	linkGenerationRetries = 5
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

type Interactor struct {
	urlRepository repository.Repository
	basicPath     string
}

func NewInteractor(basicPath string, urlRepository repository.Repository) Interactor {
	return Interactor{
		urlRepository: urlRepository,
		basicPath:     basicPath,
	}
}

func (i *Interactor) CreateShortLink(ctx context.Context, originalURL string) (*string, error) {
	_, err := url.ParseRequestURI(originalURL)
	if err != nil {
		return nil, repository.ErrInvalidURL
	}

	shortURLs := make([]string, 0, linkGenerationRetries)
	for range linkGenerationRetries {
		shortURLs = append(shortURLs, i.generatePath())
	}

	shortURL, err := i.urlRepository.SetLink(ctx, originalURL, shortURLs)
	if err != nil {
		if errors.Is(err, repository.ErrOriginalURLUniqueViolation) && shortURL != nil {
			path := i.formatURL(*shortURL)

			return &path, repository.ErrOriginalURLUniqueViolation
		}
		return nil, fmt.Errorf("can not set short link: %w", err)
	}

	path := i.formatURL(*shortURL)

	return &path, nil
}

func (i *Interactor) CreateShortLinks(
	ctx context.Context,
	batch []models.ShortenBatchRequest,
) ([]models.ShortenBatchResponse, error) {
	shortURLs := make([][]string, 0, len(batch))
	for range len(batch) {
		urls := make([]string, 0, linkGenerationRetries)
		for range linkGenerationRetries {
			urls = append(urls, i.generatePath())
		}
		shortURLs = append(shortURLs, urls)
	}

	result, err := i.urlRepository.SetLinks(ctx, batch, shortURLs)
	if err != nil {
		if errors.Is(err, repository.ErrOriginalURLUniqueViolation) && result != nil {
			response := make([]models.ShortenBatchResponse, 0, len(result))
			for j := range result {
				response = append(response, models.ShortenBatchResponse{
					CorrelationID: batch[j].CorrelationID,
					ShortURL:      i.formatURL(result[j]),
				})
			}

			return response, repository.ErrOriginalURLUniqueViolation
		}
		return nil, fmt.Errorf("can not create short links: %w", err)
	}
	if len(result) != len(batch) {
		return nil, errors.New("amount of results does not match amount of requests")
	}

	response := make([]models.ShortenBatchResponse, 0, len(result))
	for j := range result {
		response = append(response, models.ShortenBatchResponse{
			CorrelationID: batch[j].CorrelationID,
			ShortURL:      i.formatURL(result[j]),
		})
	}

	return response, nil
}

func (i *Interactor) GetShortLink(ctx context.Context, shortLink string) (*string, error) {
	originalURL, err := i.urlRepository.GetOriginalURL(ctx, shortLink)
	if err != nil {
		return nil, fmt.Errorf("can not get original url: %w", err)
	}

	return originalURL, nil
}

func (i *Interactor) PingDB(ctx context.Context) error {
	err := i.urlRepository.Ping(ctx)
	if err != nil {
		return fmt.Errorf("can not ping db: %w", err)
	}

	return nil
}

func (i *Interactor) generatePath() string {
	path := make([]rune, shortLinkPathLength)
	for j := range path {
		path[j] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(path)
}

func (i *Interactor) formatURL(shortURL string) string {
	return fmt.Sprintf("%s/%s", i.basicPath, shortURL)
}
