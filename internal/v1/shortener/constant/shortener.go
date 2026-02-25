// Package constant holds application-wide configuration values.
package constant

import "time"

const (
	// CacheTTL is how long a URL mapping stays in Redis before expiring.
	CacheTTL = 1 * time.Hour

	// ChannelBufferSize is the capacity of the click tracker's buffered channel.
	// If more than 1000 clicks arrive before they can be written to the DB,
	// new clicks will be dropped (logged, not lost silently).
	ChannelBufferSize = 1000

	// CodeLength is the number of characters in a generated short code.
	// 7 base-62 characters give ~3.5 trillion unique codes.
	CodeLength = 7
)
