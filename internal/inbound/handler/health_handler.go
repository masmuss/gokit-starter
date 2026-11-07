// Package handler provides HTTP handlers for cross-cutting concerns.
package handler

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/masmuss/gokit-starter/internal/inbound/response"
)

// HealthChecker verifies a dependency is healthy.
type HealthChecker interface {
	Ping(ctx context.Context) error
}

// HealthResponse represents the health endpoint payload.
type HealthResponse struct {
	Service string `json:"service"`
	Status  string `json:"status"`
}

// ReadinessResponse represents the readiness endpoint payload.
type ReadinessResponse struct {
	Service string            `json:"service"`
	Status  string            `json:"status"`
	Checks  map[string]string `json:"checks"`
}

// HealthHandler serves the service health endpoint.
type HealthHandler struct {
	serviceName string
	version     string
	log         *slog.Logger
	checkers    map[string]HealthChecker
}

// NewHealthHandler creates a new health handler.
func NewHealthHandler(
	serviceName, version string,
	log *slog.Logger,
	checkers ...func() (string, HealthChecker),
) *HealthHandler {
	m := make(map[string]HealthChecker)
	for _, fn := range checkers {
		name, checker := fn()
		m[name] = checker
	}
	return &HealthHandler{
		serviceName: serviceName,
		version:     version,
		log:         log,
		checkers:    m,
	}
}

// Handle returns the current service health status (liveness).
func (h *HealthHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if h.log != nil {
		h.log.DebugContext(r.Context(), "health check", "path", r.URL.Path)
	}

	_ = response.WriteJSON(w, http.StatusOK, response.OK(HealthResponse{
		Service: h.serviceName,
		Status:  "ok",
	}, "ok"))
}

// Readiness checks all dependencies and reports readiness.
func (h *HealthHandler) Readiness(w http.ResponseWriter, r *http.Request) {
	checks := make(map[string]string, len(h.checkers))
	allOK := true

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	for name, checker := range h.checkers {
		if err := checker.Ping(ctx); err != nil {
			checks[name] = err.Error()
			allOK = false
		} else {
			checks[name] = "ok"
		}
	}

	status := "ok"
	httpStatus := http.StatusOK
	if !allOK {
		status = "degraded"
		httpStatus = http.StatusServiceUnavailable
	}

	_ = response.WriteJSON(w, httpStatus, response.OK(ReadinessResponse{
		Service: h.serviceName,
		Status:  status,
		Checks:  checks,
	}, status))
}

// Version returns the build version info.
func (h *HealthHandler) Version(w http.ResponseWriter, _ *http.Request) {
	_ = response.WriteJSON(w, http.StatusOK, response.OK(map[string]string{
		"service": h.serviceName,
		"version": h.version,
	}, "ok"))
}

// RegisterRoutes registers health routes on the given router.
func (h *HealthHandler) RegisterRoutes(r chi.Router) {
	r.Get("/health", h.Handle)
	r.Get("/readyz", h.Readiness)
	r.Get("/version", h.Version)
}
