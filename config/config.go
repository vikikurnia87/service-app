package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration loaded from .env / environment variables.
type Config struct {
	AppName     string
	ServiceLang string
	Server      ServerConfig
	DB          DatabaseConfig
	Redis       RedisConfig
	Asynq       AsynqConfig
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Env          string
	Name         string
	URL          string
	Port         int
	ReadTimeout  int
	WriteTimeout int
	IdleTimeout  int
}

// Address returns the formatted listen address.
func (s ServerConfig) Address() string {
	return fmt.Sprintf(":%d", s.Port)
}

// DatabaseConfig holds PostgreSQL connection settings.
type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	NameTest        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	Timeout         time.Duration
	DialTimeout     time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
}

// RedisConfig holds Redis connection settings.
type RedisConfig struct {
	Host            string
	Port            string
	Password        string
	DB              int
	PoolSize        int
	MinIdleConns    int
	MaxRetries      int
	MinRetryBackoff time.Duration
	MaxRetryBackoff time.Duration
	DialTimeout     time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	PoolTimeout     time.Duration
	Namespace       string
	Expiration      time.Duration
}

// Addr returns the formatted Redis address (host:port).
func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%s", r.Host, r.Port)
}

// IsEnabled returns true if REDIS_HOST is configured.
func (r RedisConfig) IsEnabled() bool {
	return r.Host != ""
}

// AsynqConfig holds background job worker settings.
type AsynqConfig struct {
	Concurrency int
}

// Load reads configuration from .env file and environment variables.
// Environment variables always take precedence.
func Load() (*Config, error) {
	// Walk up directories to find .env file (works from any subdirectory)
	if envFile := findEnvFile(); envFile != "" {
		_ = godotenv.Load(envFile)
	}

	cfg := &Config{
		AppName:     getEnv("APP_NAME", "App-Dev"),
		ServiceLang: getEnv("SERVICE_LANG", "en"),
		Server: ServerConfig{
			Env:          getEnv("SERVER_ENV", "development"),
			Name:         getEnv("SERVER_NAME", "service-app-api"),
			URL:          getEnv("SERVER_URL", "127.0.0.1:6001"),
			Port:         getEnvInt("SERVER_PORT", 6001),
			ReadTimeout:  getEnvInt("SERVER_READ_TIMEOUT", 10),
			WriteTimeout: getEnvInt("SERVER_WRITE_TIMEOUT", 10),
			IdleTimeout:  getEnvInt("SERVER_IDLE_TIMEOUT", 15),
		},
		DB: DatabaseConfig{
			Host:            getEnv("DB_HOST", "127.0.0.1"),
			Port:            getEnv("DB_PORT", "5432"),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", "password"),
			Name:            getEnv("DB_NAME", "db_app"),
			NameTest:        getEnv("DB_NAME_TEST", "db_app_test"),
			SSLMode:         getEnv("DB_SSL_MODE", "disable"),
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: getEnvDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
			Timeout:         getEnvDuration("DB_TIMEOUT", 10*time.Second),
			DialTimeout:     getEnvDuration("DB_DIAL_TIMEOUT", 10*time.Second),
			ReadTimeout:     getEnvDuration("DB_READ_TIMEOUT", 10*time.Second),
			WriteTimeout:    getEnvDuration("DB_WRITE_TIMEOUT", 10*time.Second),
		},
		Redis: RedisConfig{
			Host:            getEnv("REDIS_HOST", ""),
			Port:            getEnv("REDIS_PORT", "6379"),
			Password:        getEnv("REDIS_PASSWORD", ""),
			DB:              getEnvInt("REDIS_DB", 0),
			PoolSize:        getEnvInt("REDIS_POOL_SIZE", 10),
			MinIdleConns:    getEnvInt("REDIS_MIN_IDLE_CONNS", 5),
			MaxRetries:      getEnvInt("REDIS_MAX_RETRIES", 3),
			MinRetryBackoff: getEnvDuration("REDIS_MIN_RETRY_BACKOFF", 8*time.Millisecond),
			MaxRetryBackoff: getEnvDuration("REDIS_MAX_RETRY_BACKOFF", 512*time.Millisecond),
			DialTimeout:     getEnvDuration("REDIS_DIAL_TIMEOUT", 5*time.Second),
			ReadTimeout:     getEnvDuration("REDIS_READ_TIMEOUT", 3*time.Second),
			WriteTimeout:    getEnvDuration("REDIS_WRITE_TIMEOUT", 3*time.Second),
			PoolTimeout:     getEnvDuration("REDIS_POOL_TIMEOUT", 4*time.Second),
			Namespace:       getEnv("REDIS_NAMESPACE", "service-app"),
			Expiration:      getEnvDuration("REDIS_EXPIRATION", 10*time.Minute),
		},
		Asynq: AsynqConfig{
			Concurrency: getEnvInt("ASYNQ_CONCURRENCY", 10),
		},
	}

	return cfg, nil
}

// findEnvFile walks up directories from CWD to find a .env file.
func findEnvFile() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	for {
		envPath := filepath.Join(dir, ".env")
		if _, err := os.Stat(envPath); err == nil {
			return envPath
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	}
	return defaultVal
}
