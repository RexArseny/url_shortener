package repository

import "context"

type Repository interface {
	GetShortLink(ctx context.Context, originalURL string) (string, bool, error)
	GetOriginalURL(ctx context.Context, shortLink string) (string, bool, error)
	SetLink(ctx context.Context, originalURL string, shortLink string) (bool, error)
	SetLinks(ctx context.Context, originalURLs []Batch) error
	Ping(ctx context.Context) error
}

type Batch struct {
	OriginalURL string
	ShortURL    string
}
