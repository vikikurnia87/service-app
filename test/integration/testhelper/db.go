package testhelper

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/url"
	"testing"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"

	"service-app/config"
	"service-app/internal/model"
)

// SetupTestDB creates a Bun DB connection to the test database and ensures
// the required tables exist. Uses DB.NameTest from config (e.g. db_app_test).
func SetupTestDB(t *testing.T) *bun.DB {
	t.Helper()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	dbName := cfg.DB.NameTest
	if dbName == "" {
		dbName = cfg.DB.Name + "_test"
	}

	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.DB.User, cfg.DB.Password),
		Host:   fmt.Sprintf("%s:%s", cfg.DB.Host, cfg.DB.Port),
		Path:   "/" + dbName,
	}
	q := u.Query()
	q.Set("sslmode", cfg.DB.SSLMode)
	u.RawQuery = q.Encode()

	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(u.String())))
	sqldb.SetMaxOpenConns(5)
	sqldb.SetMaxIdleConns(2)

	db := bun.NewDB(sqldb, pgdialect.New())

	if err := sqldb.Ping(); err != nil {
		t.Fatalf("failed to ping test database: %v (DB: %s)", err, dbName)
	}

	slog.Info("test database connected", "db", dbName)

	ctx := context.Background()
	models := []any{
		(*model.Role)(nil),
		(*model.User)(nil),
	}

	for _, m := range models {
		if _, err := db.NewCreateTable().Model(m).IfNotExists().Exec(ctx); err != nil {
			t.Fatalf("failed to create table: %v", err)
		}
	}

	return db
}

// TeardownTestDB closes the database connection.
func TeardownTestDB(t *testing.T, db *bun.DB) {
	t.Helper()
	if err := db.Close(); err != nil {
		t.Errorf("failed to close test database: %v", err)
	}
}

// CleanTable truncates the given table, removing all rows and resetting sequences.
func CleanTable(t *testing.T, db *bun.DB, tableName string) {
	t.Helper()
	ctx := context.Background()
	query := fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", tableName)
	if _, err := db.ExecContext(ctx, query); err != nil {
		t.Fatalf("failed to clean table %s: %v", tableName, err)
	}
}

// CleanAllTables truncates all test tables.
func CleanAllTables(t *testing.T, db *bun.DB) {
	t.Helper()
	CleanTable(t, db, "t_user")
	CleanTable(t, db, "t_role")
}

// LoadRedisConfig loads the Redis config from .env for integration tests.
// Returns the RedisConfig (may have empty Host if Redis is not configured).
func LoadRedisConfig(t *testing.T) config.RedisConfig {
	t.Helper()
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	return cfg.Redis
}
