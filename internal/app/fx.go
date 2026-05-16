// Package app handles application-wide logic and DI.
package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/masmuss/gokit-starter/internal/config"
	"github.com/masmuss/gokit-starter/internal/delivery/handler"
	deliverymiddleware "github.com/masmuss/gokit-starter/internal/delivery/middleware"
	authapp "github.com/masmuss/gokit-starter/internal/modules/auth/app"
	authinfra "github.com/masmuss/gokit-starter/internal/modules/auth/infra"
	"github.com/masmuss/gokit-starter/internal/platform/auth"
	"github.com/masmuss/gokit-starter/internal/platform/cache"
	"github.com/masmuss/gokit-starter/internal/platform/database"
	"github.com/masmuss/gokit-starter/internal/platform/eventbus"
	"github.com/masmuss/gokit-starter/internal/platform/logger"
	"github.com/masmuss/gokit-starter/internal/platform/validation"
	"go.uber.org/fx"
)

// Module is the main application module for Fx.
var Module = fx.Module("app",
	fx.Provide(
		context.Background,
		config.LoadConfig,
		provideLogger,
		database.New,
		cache.NewRedisClientOptional,
		fx.Annotate(
			cache.NewRedisCache,
			fx.As(new(cache.Cache)),
		),
		eventbus.NewInternalBus,
		fx.Annotate(
			func(b *eventbus.InternalBus) eventbus.Bus { return b },
			fx.As(new(eventbus.Bus)),
		),
		validation.New,
		auth.NewBcryptHasherFromConfig,
		auth.NewJWTManagerFromConfig,
		fx.Annotate(
			func(m *auth.JWTManager) auth.TokenIssuer { return m },
			fx.As(new(auth.TokenIssuer)),
		),
		fx.Annotate(
			func(m *auth.JWTManager) auth.TokenVerifier { return m },
			fx.As(new(auth.TokenVerifier)),
		),
		authinfra.NewRepositoryFromDB,
		provideAuthExpiresIn,
		authapp.New,
		provideServiceName,
		handler.NewHealthHandler,
		handler.NewAuthHandler,
		deliverymiddleware.NewAuthMiddleware,
		NewRouter,
	),
	fx.Invoke(RunServer),
)

func provideLogger(cfg *config.Config) *slog.Logger {
	return logger.New(cfg.App.Debug, nil)
}

func provideAuthExpiresIn(cfg *config.Config) int {
	return cfg.Auth.JWTTTL * 60
}

func provideServiceName(cfg *config.Config) string {
	return cfg.App.Name
}

// RunServer starts the HTTP server using Fx Lifecycle.
func RunServer(lc fx.Lifecycle, cfg *config.Config, log *slog.Logger, router http.Handler) {
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.App.Port),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ln, err := net.Listen("tcp", srv.Addr)
			if err != nil {
				return err
			}

			log.InfoContext(ctx, "server started", "addr", srv.Addr)

			go func() {
				if serveErr := srv.Serve(ln); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
					log.ErrorContext(context.Background(), "server error", "error", serveErr)
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.InfoContext(ctx, "server stopping")
			return srv.Shutdown(ctx)
		},
	})
}
