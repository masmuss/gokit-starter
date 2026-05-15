package app

import (
	"log/slog"

	chi "github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/masmuss/gokit-starter/internal/delivery/handler"
)

// NewRouter builds the HTTP routes and middleware stack.
func NewRouter(serviceName string, log *slog.Logger) *chi.Mux {
	r := chi.NewRouter()
	healthHandler := handler.NewHealthHandler(serviceName, log)

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)

	r.Get("/health", healthHandler.Handle)

	return r
}
