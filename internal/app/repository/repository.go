package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/RexArseny/url_shortener/internal/app/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

var (
	ErrInvalidURL                  = errors.New("provided string is not valid url")
	ErrOriginalURLUniqueViolation  = errors.New("original url unique violation")
	ErrReachedMaxGenerationRetries = errors.New("reached max generation retries")
	ErrURLIsDeleted                = errors.New("url is deleted")
)

type Repository interface {
	GetOriginalURL(
		ctx context.Context,
		shortLink string,
	) (*string, error)
	GetShortLinksOfUser(
		ctx context.Context,
		userID uuid.UUID,
	) ([]models.ShortenOfUserResponse, error)
	SetLink(
		ctx context.Context,
		originalURL string,
		shortURLs []string,
		userID uuid.UUID,
	) (*string, error)
	SetLinks(
		ctx context.Context,
		batch []models.ShortenBatchRequest,
		shortURLs [][]string,
		userID uuid.UUID,
	) ([]string, error)
	AddURLsForDelete(
		ctx context.Context,
		urls []string,
		userID uuid.UUID,
	) error
	DeleteURLs(ctx context.Context) error
	Ping(ctx context.Context) error
}

type Batch struct {
	OriginalURL string
	ShortURL    string
}

func NewRepository(
	ctx context.Context,
	logger *zap.Logger,
	fileStoragePath string,
	databaseDSN string,
) (Repository, func() error, error) {
	switch {
	case databaseDSN != "":
		dbRepository, err := NewDBRepository(ctx, logger, databaseDSN)
		if err != nil {
			return nil, nil, fmt.Errorf("can not init db repository: %w", err)
		}
		return dbRepository,
			func() error {
				dbRepository.Close()
				return nil
			},
			nil
	case fileStoragePath != "":
		linksWithFile, err := NewLinksWithFile(fileStoragePath)
		if err != nil {
			return nil, nil, fmt.Errorf("can not init file repository: %w", err)
		}
		return linksWithFile, linksWithFile.Close, nil
	default:
		links := NewLinks()
		return links, nil, nil
	}
}
