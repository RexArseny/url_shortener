package repository

import (
	"context"
	"errors"
	"sync"
)

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

func (l *Links) GetShortLink(_ context.Context, originalURL string) (string, bool, error) {
	l.m.Lock()
	defer l.m.Unlock()
	shortLink, ok := l.shortLinks[originalURL]
	return shortLink, ok, nil
}

func (l *Links) GetOriginalURL(_ context.Context, shortLink string) (string, bool, error) {
	l.m.Lock()
	defer l.m.Unlock()
	originalURL, ok := l.originalURLs[shortLink]
	return originalURL, ok, nil
}

func (l *Links) SetLink(_ context.Context, originalURL string, shortLink string) (bool, error) {
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
	return true, nil
}

func (l *Links) Ping(_ context.Context) error {
	return errors.New("service in memory storage mode")
}