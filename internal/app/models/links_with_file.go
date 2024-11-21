package models

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

const fileMode = 0o600

type URL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	ID          int    `json:"id"`
}

type LinksWithFile struct {
	*Links
	file      *os.File
	currentID int
}

func NewLinksWithFile(fileStoragePath string) (*LinksWithFile, error) {
	file, err := os.OpenFile(fileStoragePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, fileMode)
	if err != nil {
		return nil, fmt.Errorf("can not open file: %w", err)
	}

	linksWithFile := &LinksWithFile{
		Links:     NewLinks(),
		file:      file,
		currentID: 0,
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var data URL
		err = json.Unmarshal(scanner.Bytes(), &data)
		if err != nil {
			return nil, fmt.Errorf("can not unmarshal data from file: %w", err)
		}

		if _, ok := linksWithFile.Links.shortLinks[data.OriginalURL]; ok {
			return nil, errors.New("duplicate original url in file")
		}
		if _, ok := linksWithFile.Links.originalURLs[data.ShortURL]; ok {
			return nil, errors.New("duplicate short url in file")
		}
		linksWithFile.Links.shortLinks[data.OriginalURL] = data.ShortURL
		linksWithFile.Links.originalURLs[data.ShortURL] = data.OriginalURL
		linksWithFile.currentID++
	}

	return linksWithFile, nil
}

func (l *LinksWithFile) GetShortLink(originalURL string) (string, bool) {
	return l.Links.GetShortLink(originalURL)
}

func (l *LinksWithFile) GetOriginalURL(shortLink string) (string, bool) {
	return l.Links.GetOriginalURL(shortLink)
}

func (l *LinksWithFile) SetLink(originalURL string, shortLink string) (bool, error) {
	l.m.Lock()
	defer l.m.Unlock()
	if _, ok := l.shortLinks[originalURL]; ok {
		return false, nil
	}
	if _, ok := l.originalURLs[shortLink]; ok {
		return false, nil
	}
	l.shortLinks[originalURL] = shortLink
	l.originalURLs[shortLink] = originalURL
	l.currentID++

	data, err := json.Marshal(URL{
		ID:          l.currentID,
		ShortURL:    shortLink,
		OriginalURL: originalURL,
	})
	if err != nil {
		return false, fmt.Errorf("can not marshal data: %w", err)
	}
	_, err = fmt.Fprintf(l.file, "%s\n", data)
	if err != nil {
		return false, fmt.Errorf("can not write data to file: %w", err)
	}

	return true, nil
}

func (l *LinksWithFile) Close() error {
	err := l.file.Close()
	if err != nil {
		return fmt.Errorf("can not close file: %w", err)
	}
	return nil
}
