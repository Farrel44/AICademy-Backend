package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

type RedisClient struct {
	*redis.Client
}

func NewRedisClient() *RedisClient {
	host := getenv("REDIS_HOST", "redis")
	port := getenv("REDIS_PORT", "6379")
	addr := fmt.Sprintf("%s:%s", host, port)

	pwd := os.Getenv("REDIS_PASSWORD")
	dbn := getenv("REDIS_DB", "0")
	db, _ := strconv.Atoi(dbn)

	// Log Redis configuration
	log.Printf("Connecting to Redis at %s (DB: %d)", addr, db)
	if pwd != "" {
		log.Printf("Redis password is set")
	} else {
		log.Printf("Redis password is not set")
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pwd,
		DB:       db,
	})

	// Test connection
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Printf("Redis connection error: %v", err)
		log.Printf("Redis features will be disabled")
	} else {
		log.Printf("Redis connected successfully")
	}

	return &RedisClient{Client: rdb}
}

// Set string value dengan TTL
func (r *RedisClient) Set(key string, value interface{}, expiration time.Duration) error {
	err := r.Client.Set(ctx, key, value, expiration).Err()
	if err != nil {
		log.Printf("Redis SET failed for key '%s': %v", key, err)
	} else {
		log.Printf("Redis SET: '%s' (TTL: %v)", key, expiration)
	}
	return err
}

// Get string value
func (r *RedisClient) Get(key string) (string, error) {
	val, err := r.Client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			log.Printf("Redis GET: '%s' -> CACHE MISS", key)
		} else {
			log.Printf("Redis GET failed for key '%s': %v", key, err)
		}
	} else {
		log.Printf("🎯 Redis GET: '%s' -> CACHE HIT", key)
	}
	return val, err
}

// SetJSON - simpan object/struct sebagai JSON dengan TTL
func (r *RedisClient) SetJSON(key string, value interface{}, expiration time.Duration) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		log.Printf("Redis SetJSON marshal failed for key '%s': %v", key, err)
		return err
	}

	err = r.Client.Set(ctx, key, jsonData, expiration).Err()
	if err != nil {
		log.Printf("Redis SetJSON failed for key '%s': %v", key, err)
	} else {
		log.Printf("Redis SetJSON: '%s' (Size: %d bytes, TTL: %v)", key, len(jsonData), expiration)
	}
	return err
}

// GetJSON - ambil dan unmarshal JSON value
func (r *RedisClient) GetJSON(key string, dest interface{}) error {
	val, err := r.Client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			log.Printf("Redis GetJSON: '%s' -> CACHE MISS", key)
		} else {
			log.Printf("Redis GetJSON failed for key '%s': %v", key, err)
		}
		return err
	}

	err = json.Unmarshal([]byte(val), dest)
	if err != nil {
		log.Printf("Redis GetJSON unmarshal failed for key '%s': %v", key, err)
		return err
	}

	log.Printf("Redis GetJSON: '%s' -> CACHE HIT (Size: %d bytes)", key, len(val))
	return nil
}

// Delete key
func (r *RedisClient) Delete(key string) error {
	deleted, err := r.Client.Del(ctx, key).Result()
	if err != nil {
		log.Printf("Redis DELETE failed for key '%s': %v", key, err)
	} else {
		if deleted > 0 {
			log.Printf("Redis DELETE: '%s' -> DELETED", key)
		} else {
			log.Printf("Redis DELETE: '%s' -> KEY NOT FOUND", key)
		}
	}
	return err
}

// Exists check if key exists
func (r *RedisClient) Exists(key string) (bool, error) {
	result, err := r.Client.Exists(ctx, key).Result()
	exists := result > 0
	if err != nil {
		log.Printf("Redis EXISTS failed for key '%s': %v", key, err)
	} else {
		log.Printf("Redis EXISTS: '%s' -> %v", key, exists)
	}
	return exists, err
}

// Incr - increment counter (untuk rate limiting)
func (r *RedisClient) Incr(key string) (int64, error) {
	count, err := r.Client.Incr(ctx, key).Result()
	if err != nil {
		log.Printf("Redis INCR failed for key '%s': %v", key, err)
	} else {
		log.Printf("Redis INCR: '%s' -> %d", key, count)
	}
	return count, err
}

// SetNX - set only if key doesn't exist (atomic)
func (r *RedisClient) SetNX(key string, value interface{}, expiration time.Duration) (bool, error) {
	success, err := r.Client.SetNX(ctx, key, value, expiration).Result()
	if err != nil {
		log.Printf("Redis SETNX failed for key '%s': %v", key, err)
	} else {
		if success {
			log.Printf("Redis SETNX: '%s' -> LOCK ACQUIRED (TTL: %v)", key, expiration)
		} else {
			log.Printf("Redis SETNX: '%s' -> LOCK ALREADY EXISTS", key)
		}
	}
	return success, err
}

// GetTTL - get remaining time to live
func (r *RedisClient) GetTTL(key string) (time.Duration, error) {
	ttl, err := r.Client.TTL(ctx, key).Result()
	if err != nil {
		log.Printf("Redis TTL failed for key '%s': %v", key, err)
	} else {
		if ttl == -1 {
			log.Printf("Redis TTL: '%s' -> PERMANENT", key)
		} else if ttl == -2 {
			log.Printf("Redis TTL: '%s' -> KEY NOT EXISTS", key)
		} else {
			log.Printf("Redis TTL: '%s' -> %v remaining", key, ttl)
		}
	}
	return ttl, err
}

// SetExpire - set TTL for existing key
func (r *RedisClient) SetExpire(key string, expiration time.Duration) error {
	success, err := r.Client.Expire(ctx, key, expiration).Result()
	if err != nil {
		log.Printf("Redis EXPIRE failed for key '%s': %v", key, err)
	} else {
		if success {
			log.Printf("Redis EXPIRE: '%s' -> TTL set to %v", key, expiration)
		} else {
			log.Printf("Redis EXPIRE: '%s' -> KEY NOT FOUND", key)
		}
	}
	return err
}

// Additional utility methods with logging

// FlushDB - clear all keys in current database (untuk testing)
func (r *RedisClient) FlushDB() error {
	err := r.Client.FlushDB(ctx).Err()
	if err != nil {
		log.Printf("Redis FLUSHDB failed: %v", err)
	} else {
		log.Printf("Redis FLUSHDB: All keys cleared")
	}
	return err
}

// GetInfo - get Redis server info
func (r *RedisClient) GetInfo() (string, error) {
	info, err := r.Client.Info(ctx).Result()
	if err != nil {
		log.Printf("Redis INFO failed: %v", err)
	} else {
		log.Printf("Redis INFO retrieved")
	}
	return info, err
}

// Keys - get all keys matching pattern (untuk debugging - jangan dipakai di production)
func (r *RedisClient) Keys(pattern string) ([]string, error) {
	keys, err := r.Client.Keys(ctx, pattern).Result()
	if err != nil {
		log.Printf("Redis KEYS failed for pattern '%s': %v", pattern, err)
	} else {
		log.Printf("Redis KEYS: pattern '%s' -> %d keys found", pattern, len(keys))
	}
	return keys, err
}

// Close connection
func (r *RedisClient) Close() error {
	err := r.Client.Close()
	if err != nil {
		log.Printf("Redis close failed: %v", err)
	} else {
		log.Printf("Redis connection closed")
	}
	return err
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
