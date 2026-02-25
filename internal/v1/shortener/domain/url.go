// Package domain contains the core business entities of the URL shortener.
// In hexagonal architecture, domain models live at the center and have no
// dependencies on external packages (databases, HTTP frameworks, etc.).
package domain

import "time"

// URL represents the core domain entity — a shortened URL.
// Each URL maps a short code (e.g. "aB3kZ9x") to the full original URL.
type URL struct {
	ID          int64     // Auto-generated database primary key
	ShortCode   string    // The unique short code used in the shortened link
	OriginalURL string    // The full destination URL that the short code points to
	ClickCount  int64     // How many times this short URL has been visited
	CreatedAt   time.Time // Timestamp of when this URL was created
}
