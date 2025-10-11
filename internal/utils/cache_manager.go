package utils

import (
	"context"
	"crypto/md5"
	"fmt"
	"time"
)

type CacheConfig struct {
	ShortTTL  time.Duration // 2-5 minutes untuk data yang sering berubah
	MediumTTL time.Duration // 15-30 minutes untuk data semi-static
	LongTTL   time.Duration // 1-6 hours untuk data yang jarang berubah
}

type CacheManager struct {
	redis  *RedisClient
	config *CacheConfig
}

func NewCacheConfig() *CacheConfig {
	return &CacheConfig{
		ShortTTL:  2 * time.Minute,
		MediumTTL: 15 * time.Minute,
		LongTTL:   1 * time.Hour,
	}
}

func NewCacheManager(redis *RedisClient) *CacheManager {
	return &CacheManager{
		redis:  redis,
		config: NewCacheConfig(),
	}
}

// Generate smart cache key dengan hash untuk mengurangi variasi
func (cm *CacheManager) GenerateCacheKey(prefix string, params ...interface{}) string {
	if len(params) == 0 {
		return prefix
	}

	hash := md5.Sum([]byte(fmt.Sprintf("%v", params)))
	return fmt.Sprintf("%s:%x", prefix, hash)
}

// Pattern-based cache invalidation
func (cm *CacheManager) InvalidateByPattern(pattern string) error {
	ctx := context.Background()

	// Use SCAN to find keys matching pattern
	iter := cm.redis.Client.Scan(ctx, 0, pattern, 0).Iterator()
	var keys []string

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return err
	}

	if len(keys) > 0 {
		return cm.redis.Client.Del(ctx, keys...).Err()
	}

	return nil
}

// Check if should cache based on data size and frequency
func (cm *CacheManager) ShouldCache(dataSize int, requestFrequency int) bool {
	maxCacheSize := 1024 * 1024 // 1MB
	minFrequency := 3           // minimal 3 request per jam

	return dataSize < maxCacheSize && requestFrequency >= minFrequency
}

// Smart cache with TTL selection
func (cm *CacheManager) SetWithSmartTTL(key string, value interface{}, cacheType string) error {
	var ttl time.Duration

	switch cacheType {
	case "short":
		ttl = cm.config.ShortTTL
	case "medium":
		ttl = cm.config.MediumTTL
	case "long":
		ttl = cm.config.LongTTL
	default:
		ttl = cm.config.MediumTTL
	}

	return cm.redis.SetJSON(key, value, ttl)
}

// Track request frequency
func (cm *CacheManager) TrackRequestFrequency(key string) (int64, error) {
	frequencyKey := fmt.Sprintf("freq:%s", key)
	count, err := cm.redis.Incr(frequencyKey)
	if err != nil {
		return 0, err
	}

	if count == 1 {
		cm.redis.SetExpire(frequencyKey, 1*time.Hour)
	}

	return count, nil
}
