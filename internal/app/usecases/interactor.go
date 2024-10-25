package usecases

import (
	"fmt"
	"math/rand"
	"net/url"

	"github.com/RexArseny/url_shortener/internal/app/args"
	"github.com/RexArseny/url_shortener/internal/app/models"
)

const shortLinkPathLength = 8

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

type Interactor struct {
	links *models.Links
}

func NewInteractor() Interactor {
	return Interactor{
		links: models.NewLinks(),
	}
}

func (i *Interactor) CreateShortLink(originalURL string) (*string, error) {
	_, err := url.ParseRequestURI(originalURL)
	if err != nil {
		return nil, fmt.Errorf("provided string is not valid url: %s", err)
	}

	i.links.M.RLock()
	shortLink, ok := i.links.Links[originalURL]
	i.links.M.RUnlock()
	if ok {
		path := fmt.Sprintf("http://%s:%d/%s", args.DefaultDomain, args.DefaultPort, shortLink)
		return &path, nil
	}

	shortLink = i.generateShortLink()

	i.links.M.Lock()
	i.links.Links[originalURL] = shortLink
	i.links.M.Unlock()

	path := fmt.Sprintf("http://%s:%d/%s", args.DefaultDomain, args.DefaultPort, shortLink)

	return &path, nil
}

func (i *Interactor) generateShortLink() string {
	path := make([]rune, shortLinkPathLength)
	for i := range path {
		path[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(path)
}

func (i *Interactor) GetShortLink(shortLinkRequest string) (*string, error) {
	i.links.M.RLock()
	links := i.links.Links
	i.links.M.RUnlock()

	for originalURL, shortLink := range links {
		if shortLinkRequest == shortLink {
			return &originalURL, nil
		}
	}

	return nil, nil
}
