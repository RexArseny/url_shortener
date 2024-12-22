package usecases

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/url"
	"time"

	"github.com/RexArseny/url_shortener/internal/app/models"
	"github.com/RexArseny/url_shortener/internal/app/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	shortLinkPathLength   = 8
	linkGenerationRetries = 5
	urlsDeleteTimer       = 1
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

type Interactor struct {
	urlRepository repository.Repository
	logger        *zap.Logger
	basicPath     string
}

func NewInteractor(
	ctx context.Context,
	logger *zap.Logger,
	basicPath string,
	urlRepository repository.Repository,
) Interactor {
	interactor := Interactor{
		logger:        logger,
		urlRepository: urlRepository,
		basicPath:     basicPath,
	}

	go interactor.run(ctx)

	return interactor
}

func (i *Interactor) run(ctx context.Context) {
	ticker := time.NewTicker(urlsDeleteTimer * time.Second)
	for range ticker.C {
		err := i.urlRepository.DeleteURLs(ctx)
		if err != nil {
			i.logger.Error("Can not delete urls", zap.Error(err))
			return
		}
	}
}

func (i *Interactor) CreateShortLink(
	ctx context.Context,
	originalURL string,
	userID uuid.UUID,
) (*string, error) {
	_, err := url.ParseRequestURI(originalURL)
	if err != nil {
		return nil, repository.ErrInvalidURL
	}

	shortURLs := make([]string, 0, linkGenerationRetries)
	for range linkGenerationRetries {
		shortURLs = append(shortURLs, i.generatePath())
	}

	shortURL, err := i.urlRepository.SetLink(ctx, originalURL, shortURLs, userID)
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
	userID uuid.UUID,
) ([]models.ShortenBatchResponse, error) {
	shortURLs := make([][]string, 0, len(batch))
	for range len(batch) {
		urls := make([]string, 0, linkGenerationRetries)
		for range linkGenerationRetries {
			urls = append(urls, i.generatePath())
		}
		shortURLs = append(shortURLs, urls)
	}

	result, err := i.urlRepository.SetLinks(ctx, batch, shortURLs, userID)
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

func (i *Interactor) GetShortLinksOfUser(
	ctx context.Context,
	userID uuid.UUID,
) ([]models.ShortenOfUserResponse, error) {
	urls, err := i.urlRepository.GetShortLinksOfUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("can not get urls of user: %w", err)
	}

	response := make([]models.ShortenOfUserResponse, 0, len(urls))
	for j := range urls {
		response = append(response, models.ShortenOfUserResponse{
			ShortURL:    i.formatURL(urls[j].ShortURL),
			OriginalURL: urls[j].OriginalURL,
		})
	}

	return response, nil
}

func (i *Interactor) DeleteURLs(ctx context.Context, urls []string, userID uuid.UUID) error {
	go func() {
		err := i.urlRepository.AddURLsForDelete(ctx, urls, userID)
		if err != nil {
			i.logger.Error("can not get add urls for delete", zap.Error(err))
		}
	}()

	return nil
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
