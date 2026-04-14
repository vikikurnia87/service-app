package redis

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"

	"service-app/config"
)

var (
	// Client is the global Redis client instance.
	Client *redis.Client
	// Enabled indicates that Redis configuration is present.
	Enabled = false
	// Ready indicates that Redis is connected and accepting commands.
	Ready = false
	// Namespace is the Redis key namespace for this application.
	Namespace string
	// Expiration is the default cache expiration duration.
	Expiration time.Duration
)

// InitRedisDatabase initializes the Redis client if REDIS_HOST is configured.
// If REDIS_HOST is empty, Redis is skipped and the application continues without cache.
func InitRedisDatabase(ctx context.Context, cfg config.RedisConfig, logger *slog.Logger) {
	if !cfg.IsEnabled() {
		logger.InfoContext(ctx, "Redis disabled: REDIS_HOST not set.")
		return
	}

	Namespace = cfg.Namespace
	Expiration = cfg.Expiration
	addr := cfg.Addr()

	Client = redis.NewClient(&redis.Options{
		Addr:            addr,
		Password:        cfg.Password,
		DB:              cfg.DB,
		PoolSize:        cfg.PoolSize,
		MinIdleConns:    cfg.MinIdleConns,
		MaxRetries:      cfg.MaxRetries,
		MinRetryBackoff: cfg.MinRetryBackoff,
		MaxRetryBackoff: cfg.MaxRetryBackoff,
		DialTimeout:     cfg.DialTimeout,
		ReadTimeout:     cfg.ReadTimeout,
		WriteTimeout:    cfg.WriteTimeout,
		PoolTimeout:     cfg.PoolTimeout,
	})

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := Client.Ping(pingCtx).Err(); err != nil {
		logger.ErrorContext(ctx, "Failed to connect to Redis",
			slog.String("error", err.Error()),
			slog.String("addr", addr),
		)
		Client = nil
		return
	}

	Enabled = true
	Ready = true

	logger.InfoContext(ctx, "✅ Redis connected",
		slog.String("address", addr),
		slog.String("namespace", Namespace),
		slog.Int("pool_size", cfg.PoolSize),
		slog.Duration("expiration", Expiration),
	)
}

// RedisClose gracefully closes the Redis client connection.
func RedisClose(ctx context.Context, logger *slog.Logger) {
	if Client != nil {
		if err := Client.Close(); err != nil {
			logger.ErrorContext(ctx, "Failed to close Redis client",
				slog.String("error", err.Error()),
			)
		} else {
			logger.InfoContext(ctx, "Redis connection closed")
		}
	}
}

// GetClient returns the Redis client. May be nil if Redis is not ready.
func GetClient() *redis.Client {
	return Client
}

// IsReady returns true if Redis is connected and accepting commands.
func IsReady() bool {
	return Ready && Client != nil
}

// NewRedisClient creates a standalone Redis client (for testing or custom use).
// Unlike InitRedisDatabase, this does NOT set global state.
func NewRedisClient(cfg config.RedisConfig) (*redis.Client, error) {
	if !cfg.IsEnabled() {
		return nil, fmt.Errorf("redis is not configured (REDIS_HOST is empty)")
	}

	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return client, nil
}
