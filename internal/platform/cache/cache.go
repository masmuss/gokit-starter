// Package cache provides caching functionality.
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/masmuss/gokit-starter/internal/config"
	"github.com/redis/go-redis/v9"
)

// Cache defines the interface for caching operations.
type Cache interface {
	Get(ctx context.Context, key string, dest any) error
	Set(ctx context.Context, key string, value any, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
}

// RedisCache implements the Cache interface using Redis.
type RedisCache struct {
	client *redis.Client
	prefix string
}

// NewRedisClientOptional creates a new Redis client with retry logic, returning nil if the connection fails.
func NewRedisClientOptional(cfg *config.Config, log *slog.Logger) *redis.Client {
	client, err := tryConnectRedis(cfg)
	if err != nil {
		log.Warn("redis unavailable, cache disabled", "error", err)
		return nil
	}
	return client
}

func tryConnectRedis(cfg *config.Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       0, // use default DB
	})

	var err error
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		err = client.Ping(ctx).Err()
		cancel()

		if err == nil {
			return client, nil
		}

		time.Sleep(1 * time.Second)
	}

	return nil, fmt.Errorf("failed to connect to redis after %d attempts: %w", maxRetries, err)
}

// NewRedisCache creates a new RedisCache instance.
func NewRedisCache(client *redis.Client, cfg *config.Config) *RedisCache {
	return &RedisCache{
		client: client,
		prefix: cfg.Cache.Prefix,
	}
}

// Get retrieves a value from the cache.
func (c *RedisCache) Get(ctx context.Context, key string, dest any) error {
	fullKey := c.formatKey(key)
	if c.client == nil {
		return fmt.Errorf("redis client is nil")
	}
	val, err := c.client.Get(ctx, fullKey).Bytes()
	if err != nil {
		return err
	}

	return json.Unmarshal(val, dest)
}

// Set stores a value in the cache with expiration.
func (c *RedisCache) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	fullKey := c.formatKey(key)
	if c.client == nil {
		return fmt.Errorf("redis client is nil")
	}
	val, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, fullKey, val, expiration).Err()
}

// Delete removes a value from the cache.
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	fullKey := c.formatKey(key)
	if c.client == nil {
		return fmt.Errorf("redis client is nil")
	}
	return c.client.Del(ctx, fullKey).Err()
}

func (c *RedisCache) formatKey(key string) string {
	if c.prefix == "" {
		return key
	}
	return fmt.Sprintf("%s:%s", c.prefix, key)
}
