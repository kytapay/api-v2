package config

import (
	"context"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

// InitRedis initializes and returns a Redis client
func InitRedis() *redis.Client {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	redisPassword := os.Getenv("REDIS_PASSWORD")
	if redisPassword == "" {
		redisPassword = "" // No password
	}

	client := redis.NewClient(&redis.Options{
		Addr:         redisAddr,
		Password:     redisPassword,
		DB:           0,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		MinIdleConns: 5,
	})

	// Test the connection with retry
	var err error
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		_, err = client.Ping(ctx).Result()
		if err == nil {
			break
		}
		if i < maxRetries-1 {
			time.Sleep(2 * time.Second)
		}
	}
	if err != nil {
		panic("Failed to connect to Redis: " + err.Error())
	}

	return client
}

