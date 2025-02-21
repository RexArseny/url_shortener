package models

// ShortenRequest is a model for URL shortening request.
type ShortenRequest struct {
	URL string `json:"url"`
}

// ShortenResponse is a model for URL shortening response.
type ShortenResponse struct {
	Result string `json:"result"`
}

// ShortenBatchRequest is a model for URLs shortening request.
type ShortenBatchRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// ShortenBatchResponse is a model for URLs shortening response.
type ShortenBatchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// ShortenOfUserResponse is a model for URL of user response.
type ShortenOfUserResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
