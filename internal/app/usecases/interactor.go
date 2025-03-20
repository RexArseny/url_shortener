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

// Variables of short URLs.
const (
	shortLinkPathLength   = 8
	linkGenerationRetries = 5
	urlsDeleteTimer       = 100
)

// Letters which are used for new short URL generation.
var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

// Interactor is responsible for managing the logic of the service.
type Interactor struct {
	urlRepository repository.Repository
	logger        *zap.Logger
	basicPath     string
}

// NewInteractor create new Interactor.
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

	go interactor.runDeleteFromDB(ctx)

	return interactor
}

// runDeleteFromDB is a runner that execute URLs deletyeion from URL deletion queue.
func (i *Interactor) runDeleteFromDB(ctx context.Context) {
	db, ok := i.urlRepository.(*repository.DBRepository)
	if !ok {
		return
	}

	ticker := time.NewTicker(urlsDeleteTimer * time.Millisecond)
	for range ticker.C {
		err := db.DeleteURLsInDB(ctx)
		if err != nil {
			i.logger.Error("Can not delete urls", zap.Error(err))
			return
		}
	}
}

// CreateShortLink create new short URL from original URL.
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

// CreateShortLinks create new short URLs from original URLs.
func (i *Interactor) CreateShortLinks(
	ctx context.Context,
	batch []models.ShortenBatchRequest,
	userID uuid.UUID,
) ([]models.ShortenBatchResponse, error) {
	shortURLs := make([][]string, 0, len(batch))
	for range batch {
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

// GetShortLink return original URL from short URL.
func (i *Interactor) GetShortLink(ctx context.Context, shortLink string) (*string, error) {
	originalURL, err := i.urlRepository.GetOriginalURL(ctx, shortLink)
	if err != nil {
		return nil, fmt.Errorf("can not get original url: %w", err)
	}

	return originalURL, nil
}

// GetShortLinksOfUser return all short and original URLs of user if such exist.
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

// DeleteURLs delete short URLs of user if such exist.
func (i *Interactor) DeleteURLs(ctx context.Context, urls []string, userID uuid.UUID) error {
	go func() {
		err := i.urlRepository.DeleteURLs(ctx, urls, userID)
		if err != nil {
			i.logger.Error("can not get add urls for delete", zap.Error(err))
		}
	}()

	return nil
}

// PingDB check the connection with database.
func (i *Interactor) PingDB(ctx context.Context) error {
	err := i.urlRepository.Ping(ctx)
	if err != nil {
		return fmt.Errorf("can not ping db: %w", err)
	}

	return nil
}

// generatePath create new short URL.
func (i *Interactor) generatePath() string {
	path := make([]rune, shortLinkPathLength)
	for j := range path {
		path[j] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(path)
}

// formatURL format short URL into short path.
func (i *Interactor) formatURL(shortURL string) string {
	return fmt.Sprintf("%s/%s", i.basicPath, shortURL)
}
