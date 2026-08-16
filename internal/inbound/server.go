package delivery

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
	"github.com/unrolled/secure"

	"github.com/masmuss/gokit-starter/internal/config"
	"github.com/masmuss/gokit-starter/internal/pkg/doc"
)

// ServerOptions configures the HTTP server routes and docs.
type ServerOptions struct {
	CoreRegistrars []RouteRegistrar
	APIRegistrars  []RouteRegistrar
	DocRegistrars  []doc.OperationRegistrar
}

// Server wraps the HTTP server and its dependencies.
type Server struct {
	srv *http.Server
	log *slog.Logger
}

// NewServer creates a new configured HTTP server.
func NewServer(cfg *config.Config, log *slog.Logger, opts ServerOptions) *Server {
	router := buildRouter(cfg, log, opts)

	return &Server{
		srv: &http.Server{
			Addr:              fmt.Sprintf(":%d", cfg.App.Port),
			Handler:           router,
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       10 * time.Second,
			WriteTimeout:      35 * time.Second,
			IdleTimeout:       60 * time.Second,
		},
		log: log,
	}
}

// Run starts the server and blocks until graceful shutdown is complete.
func (s *Server) Run(ctx context.Context) {
	serverErr := make(chan error, 1)
	go func() {
		ln, listenErr := net.Listen("tcp", s.srv.Addr)
		if listenErr != nil {
			serverErr <- fmt.Errorf("failed to listen: %w", listenErr)
			return
		}

		s.log.InfoContext(ctx, "server started", "addr", s.srv.Addr)

		if serveErr := s.srv.Serve(ln); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			serverErr <- fmt.Errorf("server error: %w", serveErr)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		s.log.InfoContext(ctx, "received signal", "signal", sig)
	case err := <-serverErr:
		s.log.ErrorContext(ctx, "server startup failed", "error", err)
	}

	s.log.InfoContext(ctx, "server stopping")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if shutdownErr := s.srv.Shutdown(shutdownCtx); shutdownErr != nil {
		s.log.ErrorContext(ctx, "server shutdown error", "error", shutdownErr)
	}

	s.log.InfoContext(ctx, "server stopped")
}

func buildRouter(cfg *config.Config, log *slog.Logger, opts ServerOptions) http.Handler {
	docBuilder := doc.NewBuilder(
		cfg.App.Name, cfg.App.Version,
		"Boilerplate API starter with Chi, Ent, and JWT auth.",
		opts.DocRegistrars,
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
	corsOrigins := []string{"*"}
	if cfg.App.Env != "local" {
		corsOrigins = []string{cfg.App.URL}
	}
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: corsOrigins,
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders: []string{"Link"},
		MaxAge:         300,
	}))

	for _, registrar := range opts.CoreRegistrars {
		registrar.RegisterRoutes(r)
	}

	r.Route("/api/v1", func(r chi.Router) {
		for _, registrar := range opts.APIRegistrars {
			registrar.RegisterRoutes(r)
		}
	})

	r.Get("/docs", docHandler.ServeHTTP)
	r.Get("/docs/*", docHandler.ServeHTTP)

	return r
}
