package usecases

import (
	"errors"
	"fmt"
	"math/rand"
	"net/url"

	"github.com/RexArseny/url_shortener/internal/app/models"
)

const (
	shortLinkPathLength   = 8
	linkGenerationRetries = 5
)

var (
	letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	ErrMaxGenerationRetries = errors.New("reached max generation retries")
)

type Interactor struct {
	repository models.Repository
	basicPath  string
}

func NewInteractor(basicPath string, repository models.Repository) Interactor {
	return Interactor{
		repository: repository,
		basicPath:  basicPath,
	}
}

func (i *Interactor) CreateShortLink(originalURL string) (*string, error) {
	_, err := url.ParseRequestURI(originalURL)
	if err != nil {
		return nil, fmt.Errorf("provided string is not valid url: %w", err)
	}

	shortLink, ok := i.repository.GetShortLink(originalURL)
	if ok {
		path := fmt.Sprintf("%s/%s", i.basicPath, shortLink)
		return &path, nil
	}

	link, err := i.generateShortLink(originalURL)
	if err != nil {
		return nil, fmt.Errorf("can not generate short link: %w", err)
	}

	path := fmt.Sprintf("%s/%s", i.basicPath, *link)

	return &path, nil
}

func (i *Interactor) generateShortLink(originalURL string) (*string, error) {
	var retry int
	for retry < linkGenerationRetries {
		path := make([]rune, shortLinkPathLength)
		for i := range path {
			path[i] = letterRunes[rand.Intn(len(letterRunes))]
		}
		ok, err := i.repository.SetLink(originalURL, string(path))
		if err != nil {
			return nil, fmt.Errorf("can not set link: %w", err)
		}
		if ok {
			shortLink := string(path)

			return &shortLink, nil
		}
		retry++
	}
	return nil, ErrMaxGenerationRetries
}

func (i *Interactor) GetShortLink(shortLink string) (*string, error) {
	originalURL, ok := i.repository.GetOriginalURL(shortLink)
	if !ok {
		return nil, errors.New("no url by short link")
	}

	return &originalURL, nil
}
