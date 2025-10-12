package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

const (
	MinSearchLength    = 2
	MaxSearchLength    = 100
	SearchCacheTTL     = 5 * time.Minute
	CountCacheTTL      = 10 * time.Minute
	SearchRateLimit    = 10
	SearchRateLimitTTL = time.Minute
)

// SearchValidation validates search parameters
type SearchValidation struct {
	Query string
	Page  int
	Limit int
}

// ValidateSearchParams validates and sanitizes search parameters
func ValidateSearchParams(query string, page, limit int) (*SearchValidation, error) {
	// Sanitize query
	query = strings.TrimSpace(query)

	// Validate search length
	if len(query) > 0 && len(query) < MinSearchLength {
		return nil, fmt.Errorf("search query must be at least %d characters", MinSearchLength)
	}

	if len(query) > MaxSearchLength {
		return nil, fmt.Errorf("search query cannot exceed %d characters", MaxSearchLength)
	}

	// Validate pagination
	if page < 1 {
		page = 1
	}

	if limit < 1 {
		limit = 10
	}

	if limit > 100 {
		limit = 100
	}

	return &SearchValidation{
		Query: query,
		Page:  page,
		Limit: limit,
	}, nil
}

// GenerateSearchCacheKey generates a consistent cache key for search results
func GenerateSearchCacheKey(prefix, query string, page, limit int) string {
	searchStr := fmt.Sprintf("%s:q=%s:p=%d:l=%d", prefix, strings.ToLower(query), page, limit)
	hash := md5.Sum([]byte(searchStr))
	return fmt.Sprintf("search:%s:%s", prefix, hex.EncodeToString(hash[:]))
}

// GenerateCountCacheKey generates a cache key for count queries
func GenerateCountCacheKey(prefix, query string) string {
	searchStr := fmt.Sprintf("%s:count:q=%s", prefix, strings.ToLower(query))
	hash := md5.Sum([]byte(searchStr))
	return fmt.Sprintf("count:%s:%s", prefix, hex.EncodeToString(hash[:]))
}

// GenerateRateLimitKey generates a rate limit key for user searches
func GenerateRateLimitKey(userID, endpoint string) string {
	return fmt.Sprintf("rate_limit:search:%s:%s", endpoint, userID)
}

// CheckSearchRateLimit checks if user has exceeded search rate limit
func CheckSearchRateLimit(redis *RedisClient, userID, endpoint string) (bool, error) {
	if redis.Client == nil {
		return true, nil // Allow if Redis is disabled
	}

	key := GenerateRateLimitKey(userID, endpoint)
	count, err := redis.Incr(key)
	if err != nil {
		return true, nil // Allow on Redis error
	}

	if count == 1 {
		redis.SetExpire(key, SearchRateLimitTTL)
	}

	return count <= SearchRateLimit, nil
}

// SearchResult represents cached search results with metadata
type SearchResult struct {
	Data       interface{} `json:"data"`
	TotalCount int64       `json:"total_count"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	CachedAt   time.Time   `json:"cached_at"`
}

// CacheSearchResult caches search results with metadata
func CacheSearchResult(redis *RedisClient, cacheKey string, data interface{}, totalCount int64, page, limit int) error {
	if redis.Client == nil {
		return nil // Skip if Redis is disabled
	}

	result := SearchResult{
		Data:       data,
		TotalCount: totalCount,
		Page:       page,
		Limit:      limit,
		CachedAt:   time.Now(),
	}

	return redis.SetJSON(cacheKey, result, SearchCacheTTL)
}

// GetCachedSearchResult retrieves cached search results
func GetCachedSearchResult(redis *RedisClient, cacheKey string) (*SearchResult, error) {
	if redis.Client == nil {
		return nil, fmt.Errorf("cache disabled")
	}

	var result SearchResult
	err := redis.GetJSON(cacheKey, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// CacheCount caches count results for search queries
func CacheCount(redis *RedisClient, cacheKey string, count int64) error {
	if redis.Client == nil {
		return nil // Skip if Redis is disabled
	}

	return redis.Set(cacheKey, count, CountCacheTTL)
}

// GetCachedCount retrieves cached count
func GetCachedCount(redis *RedisClient, cacheKey string) (int64, error) {
	if redis.Client == nil {
		return 0, fmt.Errorf("cache disabled")
	}

	countStr, err := redis.Get(cacheKey)
	if err != nil {
		return 0, err
	}

	var count int64
	_, err = fmt.Sscanf(countStr, "%d", &count)
	return count, err
}

// InvalidateSearchCache invalidates related search caches
func InvalidateSearchCache(redis *RedisClient, prefixes ...string) error {
	if redis.Client == nil {
		return nil // Skip if Redis is disabled
	}

	for _, prefix := range prefixes {
		// Get all search cache keys for this prefix
		searchPattern := fmt.Sprintf("search:%s:*", prefix)
		countPattern := fmt.Sprintf("count:%s:*", prefix)

		searchKeys, _ := redis.Keys(searchPattern)
		countKeys, _ := redis.Keys(countPattern)

		// Delete search cache keys
		for _, key := range searchKeys {
			redis.Delete(key)
		}

		// Delete count cache keys
		for _, key := range countKeys {
			redis.Delete(key)
		}
	}

	return nil
}

// BuildSearchQuery builds optimized search query conditions
func BuildSearchQuery(query string) string {
	if query == "" {
		return ""
	}

	// Sanitize and prepare for ILIKE
	sanitized := strings.ToLower(strings.TrimSpace(query))
	sanitized = strings.ReplaceAll(sanitized, "%", "\\%")
	sanitized = strings.ReplaceAll(sanitized, "_", "\\_")

	return fmt.Sprintf("%%%s%%", sanitized)
}

// IsValidSearchTerm checks if search term is valid for performance
func IsValidSearchTerm(query string) bool {
	if len(query) == 0 {
		return true // Empty search allowed
	}

	return len(query) >= MinSearchLength && len(query) <= MaxSearchLength
}
