// Package config handles reading application configuration from
// environment variables (the .env file or OS environment).
package config

import (
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

// DatabaseConfig holds PostgreSQL connection parameters.
// Each field is mapped from an environment variable via the `env` tag.
type DatabaseConfig struct {
	Host     string `env:"DB_HOST"`     // PostgreSQL server host/address
	Port     int    `env:"DB_PORT"`     // PostgreSQL port (usually 5432)
	User     string `env:"DB_USER"`     // Database user
	Password string `env:"DB_PASSWORD"` // Database user password
	Name     string `env:"DB_NAME"`     // Database name to use
	SSLMode  string `env:"DB_SSLMODE"`  // SSL mode: disable / require, etc.
}

// RedisConfig holds Redis connection parameters.
type RedisConfig struct {
	Host     string `env:"REDIS_HOST"`      // Redis server host/address
	Port     int    `env:"REDIS_PORT"`      // Redis port (usually 6379)
	Password string `env:"REDIS_PASSWORD"` // Redis password
	DB       int    `env:"REDIS_DB"`        // Redis database number
}

// JWTConfig holds JWT configuration parameters.
type JWTConfig struct {
	Secret string `env:"JWT_SECRET"` // JWT signing secret key
}

// Config is the root application configuration.
type Config struct {
	Database DatabaseConfig `env:"DB"`
	Redis    RedisConfig    `env:"REDIS"`
	JWT      JWTConfig      `env:"JWT"`
	HTTPPort int            `env:"HTTP_PORT" envDefault:"8080"`
}

// LoadConfig loads the .env file (if present) and parses all env vars
// into a Config struct.
//
// Note: godotenv.Load() intentionally ignores its error — the .env file is
// optional. When `go test` runs from a package directory (not the repo root),
// the .env file may be missing and the app must still work with OS env vars.
// The test helper (testutil) looks up .env manually when needed.
func LoadConfig() (*Config, error) {
	godotenv.Load()

	var cfg Config

	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
