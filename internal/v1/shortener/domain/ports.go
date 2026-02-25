package domain

import "context"

// URLRepository defines the interface for persisting URLs.
// In hexagonal architecture these interfaces are called "ports" — they
// describe WHAT the application needs without specifying HOW it's done.
// The actual implementation (e.g. PostgreSQL) lives in the adapter layer.
type URLRepository interface {
	// Save stores a new shortened URL in the database.
	Save(ctx context.Context, u *URL) error
	// FindByCode looks up a URL by its short code (e.g. "aB3kZ9x").
	FindByCode(ctx context.Context, code string) (*URL, error)
	// IncrementClicks increases the click counter for a short code by 1.
	IncrementClicks(ctx context.Context, code string) error
}

// URLCache defines the interface for caching URL lookups.
// Caching avoids hitting the database on every redirect, improving performance.
type URLCache interface {
	// Get retrieves the original URL for a short code from the cache.
	// Returns an error on cache miss.
	Get(ctx context.Context, code string) (string, error)
	// Set stores a short-code-to-URL mapping in the cache.
	Set(ctx context.Context, code string, originalURL string) error
}

// ClickTracker defines the interface for asynchronously recording URL clicks.
// Clicks are tracked in the background so they don't slow down redirects.
type ClickTracker interface {
	// Track queues a click event for the given short code.
	Track(code string)
	// Start begins the background goroutine that processes click events.
	Start(ctx context.Context)
	// Stop gracefully shuts down the tracker, waiting for pending clicks to be processed.
	Stop()
}
