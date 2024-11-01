package models

import "sync"

type Links struct {
	m            *sync.Mutex
	shortLinks   map[string]string
	originalURLs map[string]string
}

func NewLinks() *Links {
	return &Links{
		m:            &sync.Mutex{},
		shortLinks:   make(map[string]string),
		originalURLs: make(map[string]string),
	}
}

func (l *Links) GetShortLink(originalURL string) (string, bool) {
	shortLink, ok := l.shortLinks[originalURL]
	return shortLink, ok
}

func (l *Links) GetOriginalURL(shortLink string) (string, bool) {
	originalURL, ok := l.originalURLs[shortLink]
	return originalURL, ok
}

func (l *Links) SetLink(originalURL string, shortLink string) {
	l.m.Lock()
	defer l.m.Unlock()
	l.shortLinks[originalURL] = shortLink
	l.originalURLs[shortLink] = originalURL
}
