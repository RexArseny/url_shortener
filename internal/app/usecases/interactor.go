package usecases

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/url"

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

func (i *Interactor) generateShortLink(ctx context.Context, originalURL string) (*string, error) {
	var retry int
	for retry < linkGenerationRetries {
		path := make([]rune, shortLinkPathLength)
		for i := range path {
			path[i] = letterRunes[rand.Intn(len(letterRunes))]
		}
		ok, err := i.urlRepository.SetLink(ctx, originalURL, string(path))
		if err != nil {
			return nil, fmt.Errorf("can not set link: %w", err)
		}
		if ok {
			shortLink := string(path)

			return &shortLink, nil
		}
		retry++
	}
	return nil, errors.New("reached max generation retries")
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
