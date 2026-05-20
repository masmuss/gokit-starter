package app

import (
	"net/http"
	"time"

	chi "github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"github.com/unrolled/secure"

	"github.com/masmuss/gokit-starter/docs"
	"github.com/masmuss/gokit-starter/internal/config"
	"github.com/masmuss/gokit-starter/internal/delivery"
)

// NewRouter builds the HTTP routes and middleware stack.
func NewRouter(
	cfg *config.Config,
	registrars []delivery.RouteRegistrar,
) http.Handler {
	r := chi.NewRouter()

	// Security Headers
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

	// Rate limiting: 100 requests per minute per IP
	r.Use(httprate.LimitByIP(100, 1*time.Minute))

	// CORS configuration
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

	r.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs/index.html", http.StatusMovedPermanently)
	})
	r.Get("/docs/*", httpSwagger.WrapHandler)

	docs.SwaggerInfo.Title = cfg.App.Name
	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Version = cfg.App.Version
	docs.SwaggerInfo.Description = "Boilerplate API starter with Chi, Ent, and JWT auth."
	docs.SwaggerInfo.Host = ""
	docs.SwaggerInfo.Schemes = []string{"http"}

	return r
}
