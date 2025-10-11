package utils

import (
	"context"
	"strconv"
	"strings"
	"time"
)

type CacheMetrics struct {
	HitRate     float64 `json:"hit_rate"`
	MissRate    float64 `json:"miss_rate"`
	MemoryUsage int64   `json:"memory_usage"`
	KeyCount    int64   `json:"key_count"`
	Uptime      int64   `json:"uptime"`
}

func (cm *CacheManager) GetCacheMetrics() (*CacheMetrics, error) {
	ctx := context.Background()

	// Get Redis INFO
	info, err := cm.redis.Client.Info(ctx, "stats", "memory", "server").Result()
	if err != nil {
		return nil, err
	}

	metrics := &CacheMetrics{}
	lines := strings.Split(info, "\r\n")

	for _, line := range lines {
		if strings.Contains(line, ":") {
			parts := strings.Split(line, ":")
			if len(parts) != 2 {
				continue
			}

			key := parts[0]
			value := parts[1]

			switch key {
			case "keyspace_hits":
				if hits, err := strconv.ParseInt(value, 10, 64); err == nil {
					// Calculate hit rate when we have both hits and misses
					if metrics.MissRate > 0 {
						total := float64(hits) + metrics.MissRate
						metrics.HitRate = float64(hits) / total * 100
						metrics.MissRate = metrics.MissRate / total * 100
					}
				}
			case "keyspace_misses":
				if misses, err := strconv.ParseInt(value, 10, 64); err == nil {
					metrics.MissRate = float64(misses)
				}
			case "used_memory":
				if mem, err := strconv.ParseInt(value, 10, 64); err == nil {
					metrics.MemoryUsage = mem
				}
			case "uptime_in_seconds":
				if uptime, err := strconv.ParseInt(value, 10, 64); err == nil {
					metrics.Uptime = uptime
				}
			}
		}
	}

	// Get key count
	keyCount, err := cm.redis.Client.DBSize(ctx).Result()
	if err == nil {
		metrics.KeyCount = keyCount
	}

	return metrics, nil
}

// Clean up expired keys periodically
func (cm *CacheManager) StartCacheCleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	go func() {
		for range ticker.C {
			cm.cleanupExpiredKeys()
			cm.monitorMemoryUsage()
		}
	}()
}

func (cm *CacheManager) cleanupExpiredKeys() {
	ctx := context.Background()

	// Clean up expired keys (Redis does this automatically, but we can help)
	script := `
        local keys = redis.call('SCAN', 0, 'MATCH', '*', 'COUNT', 100)
        local expired = 0
        for i=1,#keys[2] do
            local ttl = redis.call('TTL', keys[2][i])
            if ttl == -1 then
                redis.call('DEL', keys[2][i])
                expired = expired + 1
            end
        end
        return expired
    `

	cm.redis.Client.Eval(ctx, script, []string{})
}

func (cm *CacheManager) monitorMemoryUsage() {
	metrics, err := cm.GetCacheMetrics()
	if err != nil {
		return
	}

	maxMemory := int64(512 * 1024 * 1024) // 512MB limit
	if metrics.MemoryUsage > maxMemory {
		// Implement memory cleanup strategy
		cm.cleanupLowFrequencyKeys()
	}
}

func (cm *CacheManager) cleanupLowFrequencyKeys() {
	// Clean up keys dengan frequency rendah
	pattern := "freq:*"
	cm.InvalidateByPattern(pattern)
}
