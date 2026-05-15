package app

import (
	"net/http"

	chi "github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/masmuss/gokit-starter/docs"
	"github.com/masmuss/gokit-starter/internal/config"
	"github.com/masmuss/gokit-starter/internal/delivery/handler"
	deliverymiddleware "github.com/masmuss/gokit-starter/internal/delivery/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// NewRouter builds the HTTP routes and middleware stack.
func NewRouter(
	cfg *config.Config,
	authHandler *handler.AuthHandler,
	healthHandler *handler.HealthHandler,
	authMiddleware *deliverymiddleware.AuthMiddleware,
) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)

	r.Get("/health", healthHandler.Handle)
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
		r.With(authMiddleware.Require).Get("/profile", authHandler.Profile)
	})
	r.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs/index.html", http.StatusMovedPermanently)
	})
	r.Get("/docs/*", httpSwagger.WrapHandler)

	docs.SwaggerInfo.Title = cfg.App.Name
	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Version = "0.1.0"
	docs.SwaggerInfo.Description = "Boilerplate API starter with Chi, Ent, and JWT auth."
	docs.SwaggerInfo.Host = ""
	docs.SwaggerInfo.Schemes = []string{"http"}

	return r
}
