// Package dto (Data Transfer Object) defines the request and response
// structures used to communicate between the HTTP layer and the service layer.
// These are separate from domain models so that the API shape can change
// independently of the internal data model.
package dto

// ShortenRequest is the JSON body sent by the client when creating a short URL.
// Example: {"url": "https://example.com/very/long/path"}
type ShortenRequest struct {
	URL string `json:"url"` // The long URL to shorten
}

// ShortenResponse is returned after successfully creating a short URL.
// Example: {"short_url": "http://localhost:8080/r/aB3kZ9x"}
type ShortenResponse struct {
	ShortURL string `json:"short_url"` // The full shortened URL ready to share
}

// StatsResponse contains analytics data for a specific short URL.
type StatsResponse struct {
	ShortCode   string `json:"short_code"`   // The short code identifier
	OriginalURL string `json:"original_url"` // The destination URL
	ClickCount  int64  `json:"click_count"`  // Total number of times this link was visited
}
