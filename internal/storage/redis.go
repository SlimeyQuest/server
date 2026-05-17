package storage

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"

	"github.com/slimeyquest/server/internal/config"
)

// Redis wraps a Redis client.
type Redis struct {
	client *redis.Client
}

// NewRedis connects to Redis and verifies connectivity with Ping.
func NewRedis(ctx context.Context, cfg *config.Config) (*Redis, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return &Redis{client: client}, nil
}

// Client returns the underlying Redis client for future use.
func (r *Redis) Client() *redis.Client {
	return r.client
}

// Close closes the Redis client.
func (r *Redis) Close() error {
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}
