// Package main is the entry point for the server application.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/redis/go-redis/v9"

	"github.com/masmuss/gokit-starter/internal/config"
	delivery "github.com/masmuss/gokit-starter/internal/inbound"
	"github.com/masmuss/gokit-starter/internal/inbound/handler"
	authmodule "github.com/masmuss/gokit-starter/internal/modules/auth"
	"github.com/masmuss/gokit-starter/internal/outbound/authtoken"
	"github.com/masmuss/gokit-starter/internal/outbound/cache"
	"github.com/masmuss/gokit-starter/internal/outbound/database"
	"github.com/masmuss/gokit-starter/internal/pkg/audit"
	"github.com/masmuss/gokit-starter/internal/pkg/doc"
	"github.com/masmuss/gokit-starter/internal/pkg/logger"
	"github.com/masmuss/gokit-starter/internal/pkg/validate"
)

// version is injected at build time via -ldflags.
var version = "dev"

func main() {
	os.Setenv("APP_VERSION", version)

	cfg := loadConfig()
	log := logger.New(cfg.App.Debug, nil)
	auditLog := audit.New(log)

	ctx := context.Background()

	db := openDatabase(ctx, cfg, log)
	defer db.Close()

	redisClient, cacheStore := openCache(cfg, log)
	defer func() {
		if redisClient != nil {
			if err := redisClient.Close(); err != nil {
				log.ErrorContext(ctx, "failed to close redis client", "error", err)
			}
		}
	}()

	jwtMgr := authtoken.NewJWTManagerFromConfig(cfg)
	hasher := authtoken.NewBcryptHasherFromConfig(cfg)

	authMod := authmodule.Wire(authmodule.Dependencies{
		DB:             db,
		CacheStore:     cacheStore,
		PasswordHasher: authtoken.PasswordHasher(hasher),
		JWTManager:     jwtMgr,
		Log:            log,
		Audit:          auditLog,
		Validator:      validate.New(),
		AccessTTL:      cfg.Auth.JWTTTL * 60,
		RefreshTTL:     cfg.Auth.JWTRefreshTTL * 60,
	})

	srv := delivery.NewServer(cfg, log, delivery.ServerOptions{
		CoreRegistrars: []delivery.RouteRegistrar{
			handler.NewHealthHandler(cfg.App.Name, cfg.App.Version, log),
		},
		APIRegistrars: []delivery.RouteRegistrar{
			authMod.Registrar,
		},
		DocRegistrars: []doc.OperationRegistrar{
			handler.NewHealthDocRegistrar(),
			authMod.DocRegistrar,
		},
	})

	srv.Run(ctx)
}

func loadConfig() *config.Config {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}
	return cfg
}

func openDatabase(ctx context.Context, cfg *config.Config, log *slog.Logger) *database.DB {
	db, err := database.New(ctx, cfg)
	if err != nil {
		log.ErrorContext(ctx, "failed to connect to database", "error", err)
		os.Exit(1)
	}
	return db
}

func openCache(cfg *config.Config, log *slog.Logger) (*redis.Client, cache.Cache) {
	client := cache.NewRedisClientOptional(cfg, log)
	store := cache.NewCache(client, cfg)
	return client, store
}
