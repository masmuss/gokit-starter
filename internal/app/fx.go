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

	"github.com/redis/go-redis/v9"

	"github.com/masmuss/gokit-starter/internal/config"
	"github.com/masmuss/gokit-starter/internal/delivery"
	"github.com/masmuss/gokit-starter/internal/delivery/handler"
	deliverymiddleware "github.com/masmuss/gokit-starter/internal/delivery/middleware"
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
		fx.Annotate(
			func(h *auth.BcryptHasher) auth.PasswordHasher { return h },
		),
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
		provideAppInfo,
		provideHealthHandler,
		provideAuthMiddleware,
		provideDocBuilder,
		doc.NewHandler,
		provideRouter,
	),
	fx.Invoke(RunServer, registerDBHooks, registerRedisHooks),
	authmodule.Module,
)

func provideLogger(cfg *config.Config) *slog.Logger {
	return logger.New(cfg.App.Debug, nil)
}

func provideAuthExpiresIn(cfg *config.Config) int {
	return cfg.Auth.JWTTTL * 60
}

type appInfoOut struct {
	fx.Out
	ServiceName string `name:"serviceName"`
	AppVersion  string `name:"appVersion"`
}

func provideAppInfo(cfg *config.Config) appInfoOut {
	return appInfoOut{
		ServiceName: cfg.App.Name,
		AppVersion:  cfg.App.Version,
	}
}

type healthHandlerDeps struct {
	fx.In
	ServiceName string `name:"serviceName"`
	AppVersion  string `name:"appVersion"`
	Log         *slog.Logger
}

type healthHandlerOut struct {
	fx.Out
	Registrar delivery.RouteRegistrar `group:"routes"`
}

func provideHealthHandler(deps healthHandlerDeps) healthHandlerOut {
	return healthHandlerOut{
		Registrar: handler.NewHealthHandler(deps.ServiceName, deps.AppVersion, deps.Log),
	}
}

type routerDeps struct {
	fx.In
	Config     *config.Config
	DocHandler *doc.Handler
	Registrars []delivery.RouteRegistrar `group:"routes"`
}

func provideRouter(deps routerDeps) http.Handler {
	return NewRouter(deps.Config, deps.DocHandler, deps.Registrars)
}

func provideAuthMiddleware(verifier auth.TokenVerifier, log *slog.Logger) *deliverymiddleware.AuthMiddleware {
	return deliverymiddleware.NewAuthMiddleware(verifier, log.With("module", "auth"))
}

func provideDocBuilder(cfg *config.Config) *doc.Builder {
	return doc.NewBuilder(cfg.App.Name, cfg.App.Version, "Boilerplate API starter with Chi, Ent, and JWT auth.")
}

func registerDBHooks(lc fx.Lifecycle, db *database.DB) {
	lc.Append(fx.Hook{
		OnStop: func(_ context.Context) error {
			return db.Close()
		},
	})
}

func registerRedisHooks(lc fx.Lifecycle, client *redis.Client) {
	lc.Append(fx.Hook{
		OnStop: func(_ context.Context) error {
			if client != nil {
				return client.Close()
			}
			return nil
		},
	})
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
