// Package app contains the application entry and server setup.
package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/masmuss/gokit-starter/internal/config"
	"github.com/masmuss/gokit-starter/internal/platform/database"
	"github.com/masmuss/gokit-starter/internal/platform/logger"
)

const shutdownTimeout = 5 * time.Second

// App is the main application container.
type App struct {
	Config *config.Config
	Logger *slog.Logger
	DB     *database.DB
	Router http.Handler
}

// Bootstrap loads configuration, logging, and database dependencies.
func Bootstrap(ctx context.Context, logWriter io.Writer) (*config.Config, *slog.Logger, *database.DB, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("load config: %w", err)
	}

	log := logger.New(cfg.App.Debug, logWriter)

	db, err := database.New(ctx, cfg)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("connect database: %w", err)
	}

	return cfg, log, db, nil
}

// New wires the HTTP router and runtime dependencies into an App instance.
func New(cfg *config.Config, log *slog.Logger, db *database.DB) *App {
	return &App{
		Config: cfg,
		Logger: log,
		DB:     db,
		Router: NewRouter(cfg, db, log),
	}
}

// Start runs the HTTP server and waits for shutdown.
func (a *App) Start(ctx context.Context) (err error) {
	defer func() {
		if closeErr := a.DB.Close(); closeErr != nil {
			if err == nil {
				err = fmt.Errorf("close database: %w", closeErr)
				return
			}

			err = errors.Join(err, fmt.Errorf("close database: %w", closeErr))
		}
	}()

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", a.Config.App.Port),
		Handler:           a.Router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)

	go func() {
		a.Logger.InfoContext(ctx, "server started", "addr", server.Addr)

		if serveErr := server.ListenAndServe(); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			errCh <- serveErr
			return
		}

		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		if shutdownErr := server.Shutdown(shutdownCtx); shutdownErr != nil {
			if closeErr := server.Close(); closeErr != nil {
				return errors.Join(
					fmt.Errorf("shutdown server: %w", shutdownErr),
					fmt.Errorf("close server: %w", closeErr),
				)
			}

			return fmt.Errorf("shutdown server: %w", shutdownErr)
		}

		a.Logger.InfoContext(ctx, "server stopped")
		return nil
	case serveErr := <-errCh:
		if serveErr == nil {
			return nil
		}

		return fmt.Errorf("server error: %w", serveErr)
	}
}
