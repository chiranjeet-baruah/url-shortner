package driven

import (
	"context"
	"errors"

	"url-shortener/internal/v1/shortener/constant"

	"github.com/redis/go-redis/v9"
)

// ErrCacheMiss is returned when a short code is not found in the cache.
var ErrCacheMiss = errors.New("cache miss")

// RedisCache implements domain.URLCache using Redis as the caching backend.
// Redis is an in-memory data store, making lookups much faster than a DB query.
type RedisCache struct {
	client *redis.Client // The Redis client connection
}

// NewRedisCache creates a new RedisCache with the given Redis client.
func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{client: client}
}

// Get retrieves the original URL for a short code from Redis.
// Returns ErrCacheMiss if the key doesn't exist (i.e. cache miss).
func (c *RedisCache) Get(ctx context.Context, code string) (string, error) {
	val, err := c.client.Get(ctx, code).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrCacheMiss
	}
	return val, err
}

// Set stores a short-code → original-URL mapping in Redis with a time-to-live (TTL).
// After the TTL expires, the entry is automatically removed by Redis.
func (c *RedisCache) Set(ctx context.Context, code string, originalURL string) error {
	return c.client.Set(ctx, code, originalURL, constant.CacheTTL).Err()
}
