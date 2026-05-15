package app

import (
	"log/slog"
	"net/http"

	chi "github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/masmuss/gokit-starter/docs"
	"github.com/masmuss/gokit-starter/internal/delivery/handler"
	httpSwagger "github.com/swaggo/http-swagger/v2"
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
	r.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs/index.html", http.StatusMovedPermanently)
	})
	r.Get("/docs/*", httpSwagger.WrapHandler)

	docs.SwaggerInfo.Title = serviceName
	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Version = "0.1.0"

	return r
}
