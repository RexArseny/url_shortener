package usecases

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/RexArseny/url_shortener/internal/app/models"
	"github.com/RexArseny/url_shortener/internal/app/repository"
)

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
		return nil, models.ErrInvalidURL
	}

	shortLink, err := i.urlRepository.SetLink(ctx, originalURL)
	if err != nil {
		if errors.Is(err, models.ErrOriginalURLUniqueViolation) && shortLink != nil {
			path := fmt.Sprintf("%s/%s", i.basicPath, *shortLink)

			return &path, models.ErrOriginalURLUniqueViolation
		}
		return nil, fmt.Errorf("can not set short link: %w", err)
	}

	path := fmt.Sprintf("%s/%s", i.basicPath, *shortLink)

	return &path, nil
}

func (i *Interactor) CreateShortLinks(
	ctx context.Context,
	batch []models.ShortenBatchRequest,
) ([]models.ShortenBatchResponse, error) {
	result, err := i.urlRepository.SetLinks(ctx, batch)
	if err != nil {
		if errors.Is(err, models.ErrOriginalURLUniqueViolation) && result != nil {
			response := make([]models.ShortenBatchResponse, 0, len(result))
			for j := range result {
				response = append(response, models.ShortenBatchResponse{
					CorrelationID: batch[j].CorrelationID,
					ShortURL:      fmt.Sprintf("%s/%s", i.basicPath, result[j]),
				})
			}

			return response, models.ErrOriginalURLUniqueViolation
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
			ShortURL:      fmt.Sprintf("%s/%s", i.basicPath, result[j]),
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
