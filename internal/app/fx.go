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

	"go.uber.org/fx"

	"github.com/masmuss/gokit-starter/internal/config"
	"github.com/masmuss/gokit-starter/internal/delivery"
	"github.com/masmuss/gokit-starter/internal/delivery/handler"
	delivery_middleware "github.com/masmuss/gokit-starter/internal/delivery/middleware"
	"github.com/masmuss/gokit-starter/internal/infra/auth"
	"github.com/masmuss/gokit-starter/internal/infra/cache"
	"github.com/masmuss/gokit-starter/internal/infra/database"
	authmodule "github.com/masmuss/gokit-starter/internal/modules/auth"
	"github.com/masmuss/gokit-starter/internal/pkg/doc"
	"github.com/masmuss/gokit-starter/internal/pkg/eventbus"
	"github.com/masmuss/gokit-starter/internal/pkg/logger"
	"github.com/masmuss/gokit-starter/internal/pkg/validate"
)

// Module is the main application module for Fx.
var Module = fx.Module("app",
	fx.Provide(
		context.Background,
		config.LoadConfig,
		provideLogger,
		database.New,
		cache.NewRedisClientOptional,
		cache.NewCache,
		eventbus.NewInternalBus,
		fx.Annotate(
			func(b *eventbus.InternalBus) eventbus.Bus { return b },
			fx.As(new(eventbus.Bus)),
		),
		validate.New,
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
		provideAuthExpiresIn,
		fx.Annotate(
			provideServiceName,
			fx.ResultTags(`name:"serviceName"`),
		),
		fx.Annotate(
			provideAppVersion,
			fx.ResultTags(`name:"appVersion"`),
		),
		fx.Annotate(
			handler.NewHealthHandler,
			fx.ParamTags(`name:"serviceName"`, `name:"appVersion"`, ``),
			fx.As(new(delivery.RouteRegistrar)),
			fx.ResultTags(`group:"routes"`),
		),
		delivery_middleware.NewAuthMiddleware,
		provideDocBuilder,
		doc.NewHandler,
		fx.Annotate(
			NewRouter,
			fx.ParamTags(``, ``, `group:"routes"`),
		),
	),
	fx.Invoke(RunServer),
	authmodule.Module,
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

func provideAppVersion(cfg *config.Config) string {
	return cfg.App.Version
}

func provideDocBuilder(cfg *config.Config) *doc.Builder {
	return doc.NewBuilder(cfg.App.Name, cfg.App.Version, "Boilerplate API starter with Chi, Ent, and JWT auth.")
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
