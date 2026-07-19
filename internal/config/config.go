// Package config provides application configuration management.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config holds all application configuration.
type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Database DatabaseConfig `mapstructure:"database"`
	Log      LogConfig      `mapstructure:"log"`
	Bcrypt   BcryptConfig   `mapstructure:"bcrypt"`
	Auth     AuthConfig     `mapstructure:"auth"`
	Session  SessionConfig  `mapstructure:"session"`
	Cache    CacheConfig    `mapstructure:"cache"`
	Redis    RedisConfig    `mapstructure:"redis"`
}

// AppConfig holds application-level configuration.
type AppConfig struct {
	Name    string `mapstructure:"name"    validate:"required"`
	Env     string `mapstructure:"env"     validate:"required,oneof=local development staging production"`
	Debug   bool   `mapstructure:"debug"`
	URL     string `mapstructure:"url"     validate:"required,url"`
	Port    int    `mapstructure:"port"    validate:"required,gt=0"`
	Version string `mapstructure:"version"`
}

// DatabaseConfig holds database connection configuration.
type DatabaseConfig struct {
	Host     string `mapstructure:"host"     validate:"required"`
	Port     int    `mapstructure:"port"     validate:"required,gt=0"`
	Database string `mapstructure:"database" validate:"required"`
	Username string `mapstructure:"username" validate:"required"`
	Password string `mapstructure:"password" validate:"required"`
}

// LogConfig holds logging configuration.
type LogConfig struct {
	Channel string `mapstructure:"channel"`
	Stack   string `mapstructure:"stack"`
	Level   string `mapstructure:"level"`
}

// BcryptConfig holds bcrypt hashing configuration.
type BcryptConfig struct {
	Rounds int `mapstructure:"rounds" validate:"required,gte=10,lte=31"`
}

// AuthConfig holds authentication configuration.
type AuthConfig struct {
	JWTSecret     string `mapstructure:"jwt_secret"      validate:"required,min=32"`
	JWTIssuer     string `mapstructure:"jwt_issuer"      validate:"required"`
	JWTTTL        int    `mapstructure:"jwt_ttl"         validate:"required,gt=0"`
	JWTRefreshTTL int    `mapstructure:"jwt_refresh_ttl" validate:"required,gt=0"`
}

// SessionConfig holds session configuration.
type SessionConfig struct {
	Driver   string `mapstructure:"driver"`
	Lifetime int    `mapstructure:"lifetime"`
	Encrypt  bool   `mapstructure:"encrypt"`
	Path     string `mapstructure:"path"`
	Domain   string `mapstructure:"domain"`
}

// CacheConfig holds cache configuration.
type CacheConfig struct {
	Store  string `mapstructure:"store"  validate:"required"`
	Prefix string `mapstructure:"prefix"`
}

// RedisConfig holds Redis connection configuration.
type RedisConfig struct {
	Client   string `mapstructure:"client"`
	Host     string `mapstructure:"host"     validate:"required"`
	Password string `mapstructure:"password"`
	Port     int    `mapstructure:"port"     validate:"required,gt=0"`
}

// LoadConfig loads configuration from environment variables and .env file.
func LoadConfig() (*Config, error) {
	loadDotEnv()

	v := viper.New()

	setDefaults(v)
	bindEnv(v)

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	validate := validator.New()
	if err := validate.Struct(&cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("app.name", "gokit-starter")
	v.SetDefault("app.env", "local")
	v.SetDefault("app.debug", true)
	v.SetDefault("app.url", "http://localhost")
	v.SetDefault("app.port", 8080)
	v.SetDefault("app.version", "dev")

	v.SetDefault("database.host", "127.0.0.1")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.database", "gokit_starter")
	v.SetDefault("database.username", "gokit_starter")
	v.SetDefault("database.password", "secret")

	v.SetDefault("log.channel", "stack")
	v.SetDefault("log.stack", "single")
	v.SetDefault("log.level", "debug")

	v.SetDefault("bcrypt.rounds", 12)

	v.SetDefault("auth.jwt_secret", "change-me-at-least-32-chars-long-!!!")
	v.SetDefault("auth.jwt_issuer", "gokit-starter")
	v.SetDefault("auth.jwt_ttl", 60)
	v.SetDefault("auth.jwt_refresh_ttl", 10080) // 7 days in minutes

	v.SetDefault("session.driver", "database")
	v.SetDefault("session.lifetime", 120)
	v.SetDefault("session.encrypt", false)
	v.SetDefault("session.path", "/")
	v.SetDefault("session.domain", "")

	v.SetDefault("cache.store", "redis")
	v.SetDefault("cache.prefix", "gokit")

	v.SetDefault("redis.client", "go-redis")
	v.SetDefault("redis.host", "127.0.0.1")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.password", "")
}

func bindEnv(v *viper.Viper) {
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	_ = v.BindEnv("app.name", "APP_NAME")
	_ = v.BindEnv("app.env", "APP_ENV")
	_ = v.BindEnv("app.debug", "APP_DEBUG")
	_ = v.BindEnv("app.url", "APP_URL")
	_ = v.BindEnv("app.port", "APP_PORT")
	_ = v.BindEnv("app.version", "APP_VERSION")

	_ = v.BindEnv("database.host", "DB_HOST")
	_ = v.BindEnv("database.port", "DB_PORT")
	_ = v.BindEnv("database.database", "DB_DATABASE")
	_ = v.BindEnv("database.username", "DB_USERNAME")
	_ = v.BindEnv("database.password", "DB_PASSWORD")

	_ = v.BindEnv("log.channel", "LOG_CHANNEL")
	_ = v.BindEnv("log.stack", "LOG_STACK")
	_ = v.BindEnv("log.level", "LOG_LEVEL")

	_ = v.BindEnv("bcrypt.rounds", "BCRYPT_ROUNDS")

	_ = v.BindEnv("auth.jwt_secret", "AUTH_JWT_SECRET")
	_ = v.BindEnv("auth.jwt_issuer", "AUTH_JWT_ISSUER")
	_ = v.BindEnv("auth.jwt_ttl", "AUTH_JWT_TTL")
	_ = v.BindEnv("auth.jwt_refresh_ttl", "AUTH_JWT_REFRESH_TTL")

	_ = v.BindEnv("session.driver", "SESSION_DRIVER")
	_ = v.BindEnv("session.lifetime", "SESSION_LIFETIME")
	_ = v.BindEnv("session.encrypt", "SESSION_ENCRYPT")
	_ = v.BindEnv("session.path", "SESSION_PATH")
	_ = v.BindEnv("session.domain", "SESSION_DOMAIN")

	_ = v.BindEnv("cache.store", "CACHE_STORE")
	_ = v.BindEnv("cache.prefix", "CACHE_PREFIX")

	_ = v.BindEnv("redis.client", "REDIS_CLIENT")
	_ = v.BindEnv("redis.host", "REDIS_HOST")
	_ = v.BindEnv("redis.password", "REDIS_PASSWORD")
	_ = v.BindEnv("redis.port", "REDIS_PORT")
}

func loadDotEnv() {
	if envFile := findDotEnv(); envFile != "" {
		_ = godotenv.Load(envFile)
	}
}

func findDotEnv() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	for {
		candidate := filepath.Join(dir, ".env")
		if info, statErr := os.Stat(candidate); statErr == nil && !info.IsDir() {
			return candidate
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}

		dir = parent
	}
}
