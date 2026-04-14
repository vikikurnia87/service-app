package database

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"

	"service-app/config"
)

// NewPostgresDB creates a new Bun DB instance connected to PostgreSQL.
func NewPostgresDB(cfg config.DatabaseConfig, env string, logger *slog.Logger) (*bun.DB, error) {
	// Build DSN using net/url
	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.User, cfg.Password),
		Host:   fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Path:   "/" + cfg.Name,
	}
	q := u.Query()
	q.Set("sslmode", cfg.SSLMode)
	u.RawQuery = q.Encode()
	dsn := u.String()

	// Initialize the PostgreSQL driver
	pgConnector := pgdriver.NewConnector(
		pgdriver.WithDSN(dsn),
		pgdriver.WithConnParams(map[string]any{
			"timezone": "UTC",
		}),
		pgdriver.WithTimeout(cfg.Timeout),
		pgdriver.WithDialTimeout(cfg.DialTimeout),
		pgdriver.WithReadTimeout(cfg.ReadTimeout),
		pgdriver.WithWriteTimeout(cfg.WriteTimeout),
	)

	// Open the database connection
	sqldb := sql.OpenDB(pgConnector)

	// Set pooling configuration
	sqldb.SetMaxOpenConns(cfg.MaxOpenConns)
	sqldb.SetMaxIdleConns(cfg.MaxIdleConns)
	sqldb.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	sqldb.SetConnMaxIdleTime(5 * time.Minute)

	// Bun DB instance
	db := bun.NewDB(sqldb, pgdialect.New())

	// Add debug query hook in non-production environments
	if env != "production" {
		db.AddQueryHook(bundebug.NewQueryHook(
			bundebug.WithVerbose(true),
		))
	}

	// Test connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to DB: %w", err)
	}

	logger.InfoContext(ctx, "✅ Postgres connected",
		slog.String("host", cfg.Host),
		slog.String("port", cfg.Port),
		slog.Int("max_open_conns", cfg.MaxOpenConns),
		slog.Int("max_idle_conns", cfg.MaxIdleConns),
	)

	return db, nil
}

// InitPostgresDatabase initializes the database and panics on failure.
func InitPostgresDatabase(cfg config.DatabaseConfig, env string, logger *slog.Logger) *bun.DB {
	db, err := NewPostgresDB(cfg, env, logger)
	if err != nil {
		panic(err)
	}
	return db
}

// Close gracefully closes the database connection.
func Close(db *bun.DB) error {
	if db != nil {
		return db.Close()
	}
	return nil
}
