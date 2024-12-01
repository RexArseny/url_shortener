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

var (
	letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	ErrInvalidURL = errors.New("provided string is not valid url")
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
		return nil, fmt.Errorf("%w: %w", ErrInvalidURL, err)
	}

	shortLink, ok, err := i.urlRepository.GetShortLink(ctx, originalURL)
	if err != nil {
		return nil, fmt.Errorf("can not get short link: %w", err)
	}
	if ok {
		path := fmt.Sprintf("%s/%s", i.basicPath, shortLink)
		return &path, nil
	}

	link, err := i.generateShortLink(ctx, originalURL)
	if err != nil {
		return nil, fmt.Errorf("can not generate short link: %w", err)
	}

	path := fmt.Sprintf("%s/%s", i.basicPath, *link)

	return &path, nil
}

func (i *Interactor) CreateShortLinks(
	ctx context.Context,
	originalURLs []models.ShortenBatchRequest,
) ([]models.ShortenBatchResponse, error) {
	batch := make([]repository.Batch, 0, len(originalURLs))
	for _, originalURL := range originalURLs {
		_, err := url.ParseRequestURI(originalURL.OriginalURL)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrInvalidURL, err)
		}

		batch = append(batch, repository.Batch{
			OriginalURL: originalURL.OriginalURL,
			ShortURL:    i.generatePath(),
		})
	}

	err := i.urlRepository.SetLinks(ctx, batch)
	if err != nil {
		return nil, fmt.Errorf("can not create short links: %w", err)
	}

	response := make([]models.ShortenBatchResponse, 0, len(originalURLs))
	for j, shortLink := range batch {
		response = append(response, models.ShortenBatchResponse{
			CorrelationID: originalURLs[j].CorrelationID,
			ShortURL:      fmt.Sprintf("%s/%s", i.basicPath, shortLink.ShortURL),
		})
	}

	return response, nil
}

func (i *Interactor) GetShortLink(ctx context.Context, shortLink string) (*string, error) {
	originalURL, ok, err := i.urlRepository.GetOriginalURL(ctx, shortLink)
	if err != nil {
		return nil, fmt.Errorf("can not get original url: %w", err)
	}
	if !ok {
		return nil, errors.New("no url by short link")
	}

	return &originalURL, nil
}

func (i *Interactor) PingDB(ctx context.Context) error {
	err := i.urlRepository.Ping(ctx)
	if err != nil {
		return fmt.Errorf("can not ping db: %w", err)
	}

	return nil
}

func (i *Interactor) generateShortLink(ctx context.Context, originalURL string) (*string, error) {
	var retry int
	for retry < linkGenerationRetries {
		path := i.generatePath()
		ok, err := i.urlRepository.SetLink(ctx, originalURL, path)
		if err != nil {
			return nil, fmt.Errorf("can not set link: %w", err)
		}
		if ok {
			return &path, nil
		}
		retry++
	}
	return nil, errors.New("reached max generation retries")
}

func (i *Interactor) generatePath() string {
	path := make([]rune, shortLinkPathLength)
	for i := range path {
		path[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(path)
}
