package models

type Repository interface {
	GetShortLink(originalURL string) (string, bool)
	GetOriginalURL(shortLink string) (string, bool)
	SetLink(originalURL string, shortLink string) (bool, error)
}
