package service

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/teris-io/shortid"
	"github.com/william1nguyen/shortygo/internal/cache"
	"github.com/william1nguyen/shortygo/internal/config"
)

type URLService struct {
	cache *cache.RedisCache
}

type ShortenRequest struct {
	URL string `json:"url" binding:"required"`
	TTL int    `json:"ttl,omitempty"`
}

type ShortenResponse struct {
	ShortURL    string `json:"short_url"`
	ShortID     string `json:"short_id"`
	OriginalURL string `json:"origin_url"`
	ExpiresAt   int64  `json:"expires_at"`
	CreatedAt   int64  `json:"created_at"`
}

type URLStats struct {
	ShortID     string `json:"short_id"`
	OriginalURL string `json:"origin_url"`
	ExpiresAt   int64  `json:"expires_at"`
	CreatedAt   int64  `json:"created_at"`
}

const (
	DefaultTTL = 24 * time.Hour
	MaxTTL     = 365 * 24 * time.Hour
	MinTTL     = 1 * time.Minute
	MaxRetres  = 3
)

func NewURLService(cache *cache.RedisCache) *URLService {
	return &URLService{cache: cache}
}

func (s *URLService) ShortenURL(ctx context.Context, req *ShortenRequest) (*ShortenResponse, error) {
	normalizeURL, err := s.normalizeURL(req.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	if err := s.validateURL(normalizeURL); err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	ttl := s.determineTTL(req.TTL)

	shortID, err := s.generateUniqueShortID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate short ID: %w", err)
	}

	if err := s.cache.Set(ctx, shortID, normalizeURL, ttl); err != nil {
		return nil, fmt.Errorf("failed to store URL: %w", err)
	}

	now := time.Now()
	expiresAt := now.Add(ttl)

	return &ShortenResponse{
		ShortURL:    fmt.Sprintf("%s/%s", config.Load().BaseURL, shortID),
		ShortID:     shortID,
		OriginalURL: normalizeURL,
		ExpiresAt:   expiresAt.Unix(),
		CreatedAt:   now.Unix(),
	}, nil
}

func (s *URLService) GetOriginalURL(ctx context.Context, shortID string) (string, error) {
	if shortID == "" {
		return "", fmt.Errorf("short ID cannot be empty")
	}

	if err := s.validateShortID(shortID); err != nil {
		return "", fmt.Errorf("invalid short ID: %w", err)
	}

	url, err := s.cache.Get(ctx, shortID)
	if err != nil {
		return "", fmt.Errorf("URL not found or expired")
	}

	return url, nil
}

func (s *URLService) GetCacheMetrics() *cache.CacheMetrics {
	return s.cache.GetMetrics()
}

func (s *URLService) normalizeURL(URL string) (string, error) {
	if URL == "" {
		return "", fmt.Errorf("URL cannot be empty")
	}

	if !strings.HasPrefix(URL, "http://") && !strings.HasPrefix(URL, "https://") {
		URL = "https://" + URL
	}

	parsedURL, err := url.Parse(URL)
	if err != nil {
		return "", fmt.Errorf("malformed URL: %w", err)
	}

	return parsedURL.String(), nil
}

func (s *URLService) validateURL(URL string) error {
	parsedURL, err := url.Parse(URL)
	if err != nil {
		return fmt.Errorf("malformed URL: %w", err)
	}

	if parsedURL.Host == "" {
		return fmt.Errorf("URL must have a valid host")
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("URL must use HTTP or HTTPS protocol")
	}

	return nil
}

func (s *URLService) determineTTL(requestTTL int) time.Duration {
	if requestTTL <= 0 {
		return DefaultTTL
	}

	ttl := time.Duration(requestTTL) * time.Second

	ttl = max(ttl, MinTTL)
	ttl = min(ttl, MaxTTL)

	return ttl
}

func (s *URLService) generateUniqueShortID(ctx context.Context) (string, error) {
	for i := 0; i < MaxRetres; i++ {
		shortID, err := shortid.Generate()
		if err != nil {
			continue
		}

		exists, err := s.cache.Exists(ctx, shortID)
		if err != nil {
			continue
		}

		if !exists {
			return shortID, nil
		}
	}

	return "", fmt.Errorf("failed to generate unique short ID after %d retries", MaxRetres)
}

func (s *URLService) validateShortID(shortID string) error {
	if len(shortID) == 0 {
		return fmt.Errorf("short ID cannot be empty")
	}

	if len(shortID) > 50 {
		return fmt.Errorf("short ID too long")
	}

	return nil
}
