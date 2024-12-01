package repository

import (
	"context"
	"math/rand"

	"github.com/RexArseny/url_shortener/internal/app/models"
)

const (
	shortLinkPathLength   = 8
	linkGenerationRetries = 5
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

type Repository interface {
	GetOriginalURL(ctx context.Context, shortLink string) (*string, error)
	SetLink(ctx context.Context, originalURL string) (*string, error)
	SetLinks(ctx context.Context, batch []models.ShortenBatchRequest) ([]string, error)
	Ping(ctx context.Context) error
}

type Batch struct {
	OriginalURL string
	ShortURL    string
}

func generatePath() string {
	path := make([]rune, shortLinkPathLength)
	for i := range path {
		path[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(path)
}
