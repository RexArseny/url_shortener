package models

import "sync"

type URL struct {
	ID          int    `json:"id"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type Links struct {
	m            *sync.Mutex
	shortLinks   map[string]string
	originalURLs map[string]string
	currentID    int
}

func NewLinks() *Links {
	return &Links{
		m:            &sync.Mutex{},
		shortLinks:   make(map[string]string),
		originalURLs: make(map[string]string),
	}
}

func (l *Links) GetShortLink(originalURL string) (string, bool) {
	l.m.Lock()
	defer l.m.Unlock()
	shortLink, ok := l.shortLinks[originalURL]
	return shortLink, ok
}

func (l *Links) GetOriginalURL(shortLink string) (string, bool) {
	l.m.Lock()
	defer l.m.Unlock()
	originalURL, ok := l.originalURLs[shortLink]
	return originalURL, ok
}

func (l *Links) SetLink(originalURL string, shortLink string) (int, bool) {
	l.m.Lock()
	defer l.m.Unlock()
	if _, ok := l.shortLinks[originalURL]; ok {
		return 0, false
	}
	if _, ok := l.originalURLs[shortLink]; ok {
		return 0, false
	}
	l.shortLinks[originalURL] = shortLink
	l.originalURLs[shortLink] = originalURL
	l.currentID++
	return l.currentID, true
}
