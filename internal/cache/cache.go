package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache defines the interface for cache operations.
type Cache interface {
	Get(ctx context.Context, key string, dest any) error
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
}

// ──────────────────────────────────────────────────────────────────────────────
// Redis Cache Implementation
// ──────────────────────────────────────────────────────────────────────────────

// redisCache implements Cache using Redis.
type redisCache struct {
	client *redis.Client
}

// NewRedisCache creates a new Cache backed by Redis.
func NewRedisCache(client *redis.Client) Cache {
	return &redisCache{client: client}
}

func (c *redisCache) Get(ctx context.Context, key string, dest any) error {
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("cache get %q: %w", key, err)
	}

	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return fmt.Errorf("cache unmarshal %q: %w", key, err)
	}

	return nil
}

func (c *redisCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("cache marshal %q: %w", key, err)
	}

	if err := c.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("cache set %q: %w", key, err)
	}

	return nil
}

func (c *redisCache) Delete(ctx context.Context, key string) error {
	if err := c.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("cache delete %q: %w", key, err)
	}
	return nil
}

func (c *redisCache) Exists(ctx context.Context, key string) (bool, error) {
	n, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("cache exists %q: %w", key, err)
	}
	return n > 0, nil
}

// ──────────────────────────────────────────────────────────────────────────────
// No-op Cache Implementation (used when Redis is not available)
// ──────────────────────────────────────────────────────────────────────────────

// ErrCacheNotAvailable is returned by noopCache.Get to indicate cache is disabled.
var ErrCacheNotAvailable = errors.New("cache not available")

// noopCache is a no-operation cache that does nothing.
// Used as a fallback when Redis is not configured.
type noopCache struct{}

// NewNoopCache creates a cache that always misses.
// Set/Delete operations are silently ignored.
func NewNoopCache() Cache {
	return &noopCache{}
}

func (n *noopCache) Get(_ context.Context, _ string, _ any) error {
	return ErrCacheNotAvailable
}

func (n *noopCache) Set(_ context.Context, _ string, _ any, _ time.Duration) error {
	return nil
}

func (n *noopCache) Delete(_ context.Context, _ string) error {
	return nil
}

func (n *noopCache) Exists(_ context.Context, _ string) (bool, error) {
	return false, nil
}
