package utils

import (
	"context"
	"os"
	"strconv"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient() *redis.Client {
	addr := getenv("REDIS_ADDR", "redis:6379")
	pwd := os.Getenv("REDIS_PASSWORD")
	dbn := getenv("REDIS_DB", "0")
	db, _ := strconv.Atoi(dbn)

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pwd,
		DB:       db,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		println(err)
	}
	return rdb
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
