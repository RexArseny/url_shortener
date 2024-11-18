package usecases

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/url"
	"os"

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
	links     *models.Links
	file      *os.File
	basicPath string
}

func NewInteractor(basicPath string, fileStoragePath string) (*Interactor, error) {
	file, err := os.OpenFile(fileStoragePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("can not open file: %w", err)
	}
	links := models.NewLinks()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var data models.URL
		err = json.Unmarshal(scanner.Bytes(), &data)
		if err != nil {
			return nil, fmt.Errorf("can not unmarshal data from file: %w", err)
		}
		links.SetLink(data.OriginalURL, data.ShortURL)
	}
	return &Interactor{
		links:     links,
		basicPath: basicPath,
		file:      file,
	}, nil
}

func (i *Interactor) CloseFile() error {
	err := i.file.Close()
	if err != nil {
		return fmt.Errorf("can not close file: %w", err)
	}
	return nil
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
		if id, ok := i.links.SetLink(originalURL, string(path)); ok {
			shortLink := string(path)

			data, err := json.Marshal(models.URL{
				ID:          id,
				ShortURL:    shortLink,
				OriginalURL: originalURL,
			})
			if err != nil {
				return nil, fmt.Errorf("can not marshal data: %w", err)
			}
			_, err = fmt.Fprintf(i.file, "%s\n", data)
			if err != nil {
				return nil, fmt.Errorf("can not write data to file: %w", err)
			}

			return &shortLink, nil
		}
		retry++
	}
	return nil, ErrMaxGenerationRetries
}

func (i *Interactor) GetShortLink(shortLink string) (*string, error) {
	originalURL, ok := i.links.GetOriginalURL(shortLink)
	if !ok {
		return nil, errors.New("no url by short link")
	}

	return &originalURL, nil
}
