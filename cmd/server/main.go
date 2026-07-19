// Package main is the entry point for the server application.
package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	chi "github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/unrolled/secure"

	"github.com/masmuss/gokit-starter/internal/config"
	"github.com/masmuss/gokit-starter/internal/delivery"
	"github.com/masmuss/gokit-starter/internal/delivery/handler"
	deliverymiddleware "github.com/masmuss/gokit-starter/internal/delivery/middleware"
	"github.com/masmuss/gokit-starter/internal/infra/auth"
	"github.com/masmuss/gokit-starter/internal/infra/cache"
	"github.com/masmuss/gokit-starter/internal/infra/database"
	modapp "github.com/masmuss/gokit-starter/internal/modules/auth/app"
	modhandler "github.com/masmuss/gokit-starter/internal/modules/auth/handler"
	modinfra "github.com/masmuss/gokit-starter/internal/modules/auth/infra"
	"github.com/masmuss/gokit-starter/internal/pkg/audit"
	"github.com/masmuss/gokit-starter/internal/pkg/doc"
	"github.com/masmuss/gokit-starter/internal/pkg/eventbus"
	"github.com/masmuss/gokit-starter/internal/pkg/logger"
	"github.com/masmuss/gokit-starter/internal/pkg/validate"
)

// version is injected at build time via -ldflags.
var version = "dev"

func main() {
	os.Setenv("APP_VERSION", version)

	// Config
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Logger
	log := logger.New(cfg.App.Debug, nil)
	auditLog := audit.New(log)

	// Database
	ctx := context.Background()
	db, err := database.New(ctx, cfg)
	if err != nil {
		log.ErrorContext(ctx, "failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Redis
	redisClient := cache.NewRedisClientOptional(cfg, log)
	cacheStore := cache.NewCache(redisClient, cfg)

	// Event Bus
	bus := eventbus.NewInternalBus()
	_ = bus // ready for subscribers

	// Validator
	v := validate.New()

	// Auth Infrastructure
	hasher := auth.NewBcryptHasherFromConfig(cfg)
	var passwordHasher auth.PasswordHasher = hasher

	jwtMgr := auth.NewJWTManagerFromConfig(cfg)
	var tokenIssuer auth.TokenIssuer = jwtMgr
	var refreshIssuer auth.RefreshTokenIssuer = jwtMgr
	var tokenVerifier auth.TokenVerifier = jwtMgr

	blacklist := auth.NewTokenBlacklist(cacheStore)

	accessTTL := cfg.Auth.JWTTTL * 60
	refreshTTL := cfg.Auth.JWTRefreshTTL * 60

	// Auth Module
	repo := modinfra.NewRepositoryFromDB(db)
	var authRepo modapp.Repository = repo

	authSvc := modapp.New(
		authRepo, passwordHasher, tokenIssuer, refreshIssuer,
		tokenVerifier, blacklist, accessTTL, refreshTTL,
	)
	var authService modhandler.AuthService = authSvc

	authHandler := modhandler.NewAuthHandler(
		authService, log.With("module", "auth"), auditLog, v, tokenVerifier,
	)

	// Middleware
	authMiddleware := deliverymiddleware.NewAuthMiddleware(
		tokenVerifier, blacklist, log.With("module", "auth"), auditLog,
	)

	// Route Registrars
	routeRegistrars := []delivery.RouteRegistrar{
		handler.NewHealthHandler(cfg.App.Name, cfg.App.Version, log),
		delivery.RouteRegistrarFunc(func(r chi.Router) {
			authHandler.RegisterRoutes(r, authMiddleware)
		}),
	}

	// Doc
	docRegistrars := []doc.OperationRegistrar{
		handler.NewHealthDocRegistrar(),
		modhandler.NewAuthDocRegistrar(),
	}
	docBuilder := doc.NewBuilder(
		cfg.App.Name, cfg.App.Version,
		"Boilerplate API starter with Chi, Ent, and JWT auth.",
		docRegistrars,
	)
	docHandler := doc.NewHandler(docBuilder, log)

	// Router
	router := buildRouter(cfg, docHandler, routeRegistrars)

	// Server
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.App.Port),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		ln, listenErr := net.Listen("tcp", srv.Addr)
		if listenErr != nil {
			log.ErrorContext(ctx, "failed to listen", "error", listenErr)
			os.Exit(1)
		}

		log.InfoContext(ctx, "server started", "addr", srv.Addr)

		if serveErr := srv.Serve(ln); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			log.ErrorContext(context.Background(), "server error", "error", serveErr)
			os.Exit(1)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.InfoContext(ctx, "server stopping")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if shutdownErr := srv.Shutdown(shutdownCtx); shutdownErr != nil {
		log.ErrorContext(ctx, "server shutdown error", "error", shutdownErr)
	}

	if redisClient != nil {
		if closeErr := redisClient.Close(); closeErr != nil {
			log.ErrorContext(ctx, "redis close error", "error", closeErr)
		}
	}

	log.InfoContext(ctx, "server stopped")
}

func buildRouter(
	cfg *config.Config,
	docHandler *doc.Handler,
	registrars []delivery.RouteRegistrar,
) http.Handler {
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
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	for _, registrar := range registrars {
		registrar.RegisterRoutes(r)
	}

	r.Get("/docs", docHandler.ServeHTTP)
	r.Get("/docs/*", docHandler.ServeHTTP)

	return r
}
