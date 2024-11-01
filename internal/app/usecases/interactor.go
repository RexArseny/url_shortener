package usecases

import (
	"errors"
	"fmt"
	"math/rand"
	"net/url"

	"github.com/RexArseny/url_shortener/internal/app/models"
)

const shortLinkPathLength = 8

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

type Interactor struct {
	links     *models.Links
	basicPath string
}

func NewInteractor(basicPath string) Interactor {
	return Interactor{
		links:     models.NewLinks(),
		basicPath: basicPath,
	}
}

func (i *Interactor) CreateShortLink(originalURL string) (*string, error) {
	_, err := url.ParseRequestURI(originalURL)
	if err != nil {
		return nil, fmt.Errorf("provided string is not valid url: %w", err)
	}

	shortLink, ok := i.links.GetShortLink(originalURL)
	if ok {
		path := fmt.Sprintf("%s/%s", i.basicPath, shortLink)
		return &path, nil
	}

	shortLink = i.generateShortLink()

	i.links.SetLink(originalURL, shortLink)

	path := fmt.Sprintf("%s/%s", i.basicPath, shortLink)

	return &path, nil
}

func (i *Interactor) generateShortLink() string {
	for {
		path := make([]rune, shortLinkPathLength)
		for i := range path {
			path[i] = letterRunes[rand.Intn(len(letterRunes))]
		}
		if _, ok := i.links.GetShortLink(string(path)); !ok {
			return string(path)
		}
	}
}

func (i *Interactor) GetShortLink(shortLink string) (*string, error) {
	originalURL, ok := i.links.GetOriginalURL(shortLink)
	if !ok {
		return nil, errors.New("no url by short link")
	}

	return &originalURL, nil
}
