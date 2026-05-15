// Package cache provides caching functionality.
package cache

import (
	"context"
	"encoding/json"
	"fmt"
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

// NewRedisClient creates a new Redis client.
func NewRedisClient(cfg *config.Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       0, // use default DB
	})

	// Check connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return client, nil
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
	val, err := c.client.Get(ctx, fullKey).Bytes()
	if err != nil {
		return err
	}

	return json.Unmarshal(val, dest)
}

// Set stores a value in the cache with expiration.
func (c *RedisCache) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	fullKey := c.formatKey(key)
	val, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, fullKey, val, expiration).Err()
}

// Delete removes a value from the cache.
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	fullKey := c.formatKey(key)
	return c.client.Del(ctx, fullKey).Err()
}

func (c *RedisCache) formatKey(key string) string {
	if c.prefix == "" {
		return key
	}
	return fmt.Sprintf("%s:%s", c.prefix, key)
}
