// Package main is the entry point for the server application.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/redis/go-redis/v9"
	"github.com/unrolled/secure"

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

	router := buildRouter(cfg, []delivery.RouteRegistrar{
		handler.NewHealthHandler(cfg.App.Name, cfg.App.Version, log),
		authMod.Registrar,
	}, []doc.OperationRegistrar{
		handler.NewHealthDocRegistrar(),
		authMod.DocRegistrar,
	}, log)

	runServer(ctx, serverConfig{
		Router:      router,
		Port:        cfg.App.Port,
		RedisClient: redisClient,
	}, log)
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

func buildRouter(
	cfg *config.Config,
	registrars []delivery.RouteRegistrar,
	docRegistrars []doc.OperationRegistrar,
	log *slog.Logger,
) http.Handler {
	docBuilder := doc.NewBuilder(
		cfg.App.Name, cfg.App.Version,
		"Boilerplate API starter with Chi, Ent, and JWT auth.",
		docRegistrars,
	)
	docHandler := doc.NewHandler(docBuilder, log)

	r := chi.NewRouter()

	secureMiddleware := secure.New(secure.Options{
		FrameDeny:          true,
		ContentTypeNosniff: true,
		BrowserXssFilter:   true,
		IsDevelopment:      cfg.App.Env == "local",
	})

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.CleanPath)
	r.Use(secureMiddleware.Handler)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(httprate.LimitByIP(100, 1*time.Minute))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders: []string{"Link"},
		MaxAge:         300,
	}))

	for _, registrar := range registrars {
		registrar.RegisterRoutes(r)
	}

	r.Get("/docs", docHandler.ServeHTTP)
	r.Get("/docs/*", docHandler.ServeHTTP)

	return r
}

type serverConfig struct {
	Router      http.Handler
	Port        int
	RedisClient *redis.Client
}

func runServer(ctx context.Context, cfg serverConfig, log *slog.Logger) {
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           cfg.Router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		ln, listenErr := net.Listen("tcp", srv.Addr)
		if listenErr != nil {
			serverErr <- fmt.Errorf("failed to listen: %w", listenErr)
			return
		}

		log.InfoContext(ctx, "server started", "addr", srv.Addr)

		if serveErr := srv.Serve(ln); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			serverErr <- fmt.Errorf("server error: %w", serveErr)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		log.InfoContext(ctx, "received signal", "signal", sig)
	case err := <-serverErr:
		log.ErrorContext(ctx, "server startup failed", "error", err)
	}

	log.InfoContext(ctx, "server stopping")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if shutdownErr := srv.Shutdown(shutdownCtx); shutdownErr != nil {
		log.ErrorContext(ctx, "server shutdown error", "error", shutdownErr)
	}

	if cfg.RedisClient != nil {
		if closeErr := cfg.RedisClient.Close(); closeErr != nil {
			log.ErrorContext(ctx, "redis close error", "error", closeErr)
		}
	}

	log.InfoContext(ctx, "server stopped")
}
