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

func (l *Links) SetLink(originalURL string, shortLink string) bool {
	l.m.Lock()
	defer l.m.Unlock()
	if _, ok := l.shortLinks[originalURL]; ok {
		return false
	}
	if _, ok := l.originalURLs[shortLink]; ok {
		return false
	}
	l.shortLinks[originalURL] = shortLink
	l.originalURLs[shortLink] = originalURL
	return true
}
