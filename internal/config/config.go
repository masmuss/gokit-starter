// Package config provides application configuration management.
package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all application configuration.
type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Log      LogConfig
	Bcrypt   BcryptConfig
	Session  SessionConfig
	Cache    CacheConfig
	Redis    RedisConfig
}

// AppConfig holds application-level configuration.
type AppConfig struct {
	Name  string
	Env   string
	Debug bool
	URL   string
	Port  int
}

// DatabaseConfig holds database connection configuration.
type DatabaseConfig struct {
	Host     string
	Port     int
	Database string
	Username string
	Password string
}

// LogConfig holds logging configuration.
type LogConfig struct {
	Channel string
	Stack   string
	Level   string
}

// BcryptConfig holds bcrypt hashing configuration.
type BcryptConfig struct {
	Rounds int
}

// SessionConfig holds session configuration.
type SessionConfig struct {
	Driver   string
	Lifetime int
	Encrypt  bool
	Path     string
	Domain   string
}

// CacheConfig holds cache configuration.
type CacheConfig struct {
	Store  string
	Prefix string
}

// RedisConfig holds Redis connection configuration.
type RedisConfig struct {
	Client   string
	Host     string
	Password string
	Port     int
}

// LoadConfig loads configuration from environment variables and .env file.
func LoadConfig() (*Config, error) {
	v := viper.New()

	v.SetConfigName(".env")
	v.SetConfigType("env")
	v.AddConfigPath(".")

	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		var configFileNotFound viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFound) {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("app.name", "gokit-starter")
	v.SetDefault("app.env", "local")
	v.SetDefault("app.debug", true)
	v.SetDefault("app.url", "http://localhost")
	v.SetDefault("app.port", 8080)

	v.SetDefault("database.host", "127.0.0.1")
	v.SetDefault("database.port", 3306)
	v.SetDefault("database.database", "gokit_starter")
	v.SetDefault("database.username", "gokit_starter")
	v.SetDefault("database.password", "secret")

	v.SetDefault("log.channel", "stack")
	v.SetDefault("log.stack", "single")
	v.SetDefault("log.level", "debug")

	v.SetDefault("bcrypt.rounds", 12)

	v.SetDefault("session.driver", "database")
	v.SetDefault("session.lifetime", 120)
	v.SetDefault("session.encrypt", false)
	v.SetDefault("session.path", "/")
	v.SetDefault("session.domain", "")

	v.SetDefault("cache.store", "database")
	v.SetDefault("cache.prefix", "")

	v.SetDefault("redis.client", "phpredis")
	v.SetDefault("redis.host", "127.0.0.1")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.password", "")
}
