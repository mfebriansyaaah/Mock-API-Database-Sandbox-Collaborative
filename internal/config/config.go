// Package config loads environment variables (optionally from a .env file)
// and exposes them as a typed struct. Defaults are applied for every field
// so callers can rely on the struct being fully populated.
package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config is the application's runtime configuration, populated from env vars.
type Config struct {
	Port      string
	ProjectID string
	AppEnv    string // "development" or "production"

	// Database configuration
	Database DatabaseConfig

	// Logger
	Logger LoggerConfig
}

// IsDevelopment returns true when AppEnv is "development" (the default).
func (c *Config) IsDevelopment() bool { return c.AppEnv == "development" }

// IsProduction returns true when AppEnv is "production".
func (c *Config) IsProduction() bool { return c.AppEnv == "production" }

// DatabaseConfig holds configuration for the database
type DatabaseConfig struct {
	// Type of database: "firestore", "postgresql", or "mysql"
	Type string

	// Firestore specific configuration
	FirestoreProjectID       string
	FirestoreCredentialsFile string

	// PostgreSQL specific configuration
	PostgreSQLHost     string
	PostgreSQLPort     string
	PostgreSQLUser     string
	PostgreSQLPassword string
	PostgreSQLDatabase string
	PostgreSQLSSLMode  string

	// MySQL specific configuration
	MySQLHost     string
	MySQLPort     string
	MySQLUser     string
	MySQLPassword string
	MySQLDatabase string
}

// LoggerConfig is the subset of Config consumed by the logger package.
type LoggerConfig struct {
	ChannelBuffer     int
	NumWorkers        int
	CleanupInterval   time.Duration
	MaxLogsPerProject int
}

// Env keys (kept exported for tests and tooling).
const (
	EnvPort                 = "PORT"
	EnvProjectID            = "GOOGLE_CLOUD_PROJECT"
	EnvCredentialsFile      = "GOOGLE_APPLICATION_CREDENTIALS"
	EnvAppEnv               = "APP_ENV"
	EnvLogChannelBuffer     = "LOG_CHANNEL_BUFFER"
	EnvLogNumWorkers        = "LOG_NUM_WORKERS"
	EnvLogCleanupInterval   = "LOG_CLEANUP_INTERVAL"
	EnvLogMaxLogsPerProject = "LOG_MAX_LOGS_PER_PROJECT"
)

// Database configuration environment variable keys
const (
	EnvDatabaseType             = "DATABASE_TYPE"
	EnvFirestoreProjectID       = "FIRESTORE_PROJECT_ID"
	EnvFirestoreCredentialsFile = "FIRESTORE_CREDENTIALS_FILE"
	EnvPostgreSQLHost           = "POSTGRES_HOST"
	EnvPostgreSQLPort           = "POSTGRES_PORT"
	EnvPostgreSQLUser           = "POSTGRES_USER"
	EnvPostgreSQLPassword       = "POSTGRES_PASSWORD"
	EnvPostgreSQLDatabase       = "POSTGRES_DB"
	EnvPostgreSQLSSLMode        = "POSTGRES_SSL_MODE"
	EnvMySQLHost                = "MYSQL_HOST"
	EnvMySQLPort                = "MYSQL_PORT"
	EnvMySQLUser                = "MYSQL_USER"
	EnvMySQLPassword            = "MYSQL_PASSWORD"
	EnvMySQLDatabase            = "MYSQL_DB"
)

// Default values applied when env vars are missing or invalid.
const (
	defaultPort                 = "8080"
	defaultLogChannelBuffer     = 1000
	defaultLogNumWorkers        = 10
	defaultLogCleanupInterval   = 5 * time.Minute
	defaultLogMaxLogsPerProject = 100
	defaultDatabaseType         = "firestore"
	defaultAppEnv               = "development"
)

// Load reads `.env` (if present) into the process env, then constructs a
// fully-populated Config. Missing .env files are not fatal; OS env wins
// (godotenv.Load does not overwrite existing env vars by default).
func Load() Config {
	if err := godotenv.Load(); err != nil {
		log.Printf("config: .env not loaded: %v", err)
	}

	cfg := Config{
		Port:      getString(EnvPort, defaultPort),
		ProjectID: getString(EnvProjectID, "mockapi-sandbox-dev"),
		AppEnv:    getString(EnvAppEnv, defaultAppEnv),
		Database: DatabaseConfig{
			Type:                     getString(EnvDatabaseType, defaultDatabaseType),
			FirestoreProjectID:       getString(EnvFirestoreProjectID, ""),
			FirestoreCredentialsFile: getString(EnvFirestoreCredentialsFile, ""),
			PostgreSQLHost:           getString(EnvPostgreSQLHost, ""),
			PostgreSQLPort:           getString(EnvPostgreSQLPort, "5432"),
			PostgreSQLUser:           getString(EnvPostgreSQLUser, ""),
			PostgreSQLPassword:       getString(EnvPostgreSQLPassword, ""),
			PostgreSQLDatabase:       getString(EnvPostgreSQLDatabase, ""),
			PostgreSQLSSLMode:        getString(EnvPostgreSQLSSLMode, "disable"),
			MySQLHost:                getString(EnvMySQLHost, ""),
			MySQLPort:                getString(EnvMySQLPort, "3306"),
			MySQLUser:                getString(EnvMySQLUser, ""),
			MySQLPassword:            getString(EnvMySQLPassword, ""),
			MySQLDatabase:            getString(EnvMySQLDatabase, ""),
		},
		Logger: LoggerConfig{
			ChannelBuffer:     getInt(EnvLogChannelBuffer, defaultLogChannelBuffer),
			NumWorkers:        getInt(EnvLogNumWorkers, defaultLogNumWorkers),
			CleanupInterval:   getDuration(EnvLogCleanupInterval, defaultLogCleanupInterval),
			MaxLogsPerProject: getInt(EnvLogMaxLogsPerProject, defaultLogMaxLogsPerProject),
		},
	}
	return cfg
}

func getString(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
		log.Printf("config: invalid int for %s=%q, using default %d", key, v, def)
	}
	return def
}

func getDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
		log.Printf("config: invalid duration for %s=%q, using default %s", key, v, def)
	}
	return def
}
