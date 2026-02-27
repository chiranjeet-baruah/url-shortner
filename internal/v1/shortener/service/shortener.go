// Package service contains the core business logic of the URL shortener.
// It sits between the HTTP handlers (driver adapter) and the database/cache
// (driven adapters), orchestrating the flow of data through the application.
package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/url"
	"time"

	"url-shortener/internal/v1/shortener/constant"
	"url-shortener/internal/v1/shortener/domain"
	"url-shortener/internal/v1/shortener/dto"
)

// Service implements the URL shortening business logic.
// It depends only on domain interfaces (ports), not on concrete implementations,
// which makes it easy to test and swap out infrastructure.
type Service struct {
	repo    domain.URLRepository // Persistent storage (e.g. PostgreSQL)
	cache   domain.URLCache      // Fast lookup cache (e.g. Redis)
	tracker domain.ClickTracker  // Asynchronous click counter
	baseURL string               // Public base URL used to build short links
}

// New creates a new Service with the given dependencies.
// This is a form of "dependency injection" — the caller decides which
// concrete implementations to use.
func New(repo domain.URLRepository, cache domain.URLCache, tracker domain.ClickTracker, baseURL string) *Service {
	return &Service{
		repo:    repo,
		cache:   cache,
		tracker: tracker,
		baseURL: baseURL,
	}
}

// Shorten validates the incoming URL, generates a unique short code,
// persists the mapping to the database, warms the cache, and returns
// the shortened URL.
func (s *Service) Shorten(ctx context.Context, req dto.ShortenRequest) (dto.ShortenResponse, error) {
	// Step 1: Validate that the provided string is a proper URL.
	if _, err := url.ParseRequestURI(req.URL); err != nil {
		return dto.ShortenResponse{}, errors.New("invalid URL")
	}

	// Step 2: Generate a short code from the URL + current timestamp.
	code := generateCode(req.URL)

	// Step 3: Save the mapping to the database.
	u := &domain.URL{
		ShortCode:   code,
		OriginalURL: req.URL,
	}
	if err := s.repo.Save(ctx, u); err != nil {
		return dto.ShortenResponse{}, fmt.Errorf("saving URL: %w", err)
	}

	// Step 4: Pre-populate the cache so the first redirect is fast.
	// Cache errors are non-fatal — the redirect will still work via DB lookup.
	if err := s.cache.Set(ctx, code, req.URL); err != nil {
		log.Printf("cache set error: %v", err)
	}

	return dto.ShortenResponse{
		ShortURL: s.baseURL + "/r/" + code,
	}, nil
}

// Resolve looks up the original URL for a given short code.
// It first checks the cache (fast path), then falls back to the database
// (slow path). Every successful resolve also tracks a click asynchronously.
func (s *Service) Resolve(ctx context.Context, code string) (string, error) {
	// Fast path: try the Redis cache first.
	originalURL, err := s.cache.Get(ctx, code)
	if err == nil {
		s.tracker.Track(code)
		return originalURL, nil
	}

	// Slow path: cache miss — look up in PostgreSQL.
	u, err := s.repo.FindByCode(ctx, code)
	if err != nil {
		return "", fmt.Errorf("finding URL: %w", err)
	}

	// Backfill the cache so the next request for this code is fast.
	if err := s.cache.Set(ctx, code, u.OriginalURL); err != nil {
		log.Printf("cache set error: %v", err)
	}

	s.tracker.Track(code)
	return u.OriginalURL, nil
}

// Stats returns analytics for a short code — the original URL and total click count.
func (s *Service) Stats(ctx context.Context, code string) (dto.StatsResponse, error) {
	u, err := s.repo.FindByCode(ctx, code)
	if err != nil {
		return dto.StatsResponse{}, fmt.Errorf("finding URL: %w", err)
	}

	return dto.StatsResponse{
		ShortCode:   u.ShortCode,
		OriginalURL: u.OriginalURL,
		ClickCount:  u.ClickCount,
	}, nil
}

// base62Chars is the character set used to encode short codes.
// Base-62 uses digits + uppercase + lowercase letters (no special characters).
const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// generateCode creates a unique short code for a URL.
// It hashes the URL combined with the current timestamp to avoid collisions,
// then picks characters from the base-62 set to build a human-friendly code.
func generateCode(rawURL string) string {
	// Combine URL + nanosecond timestamp to ensure uniqueness.
	data := fmt.Sprintf("%s%d", rawURL, time.Now().UnixNano())
	// SHA-256 produces a deterministic, uniformly distributed hash.
	hash := sha256.Sum256([]byte(data))
	hexStr := hex.EncodeToString(hash[:])

	// Pick CodeLength characters from the hash, mapped into base-62.
	var code []byte
	for i := 0; i < constant.CodeLength; i++ {
		idx := hexStr[i] % byte(len(base62Chars))
		code = append(code, base62Chars[idx])
	}
	return string(code)
}
