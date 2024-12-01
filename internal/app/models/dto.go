package models

import "errors"

var (
	ErrInvalidURL                  = errors.New("provided string is not valid url")
	ErrOriginalURLUniqueViolation  = errors.New("original url unique violation")
	ErrReachedMaxGenerationRetries = errors.New("reached max generation retries")
)

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	Result string `json:"result"`
}

type ShortenBatchRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type ShortenBatchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}
