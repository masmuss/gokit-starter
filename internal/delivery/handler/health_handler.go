package handler

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/masmuss/gokit-starter/internal/delivery/response"
)

// HealthResponse represents the health endpoint payload.
type HealthResponse struct {
	Service string `json:"service"`
	Status  string `json:"status"`
}

// HealthHandler serves the service health endpoint.
type HealthHandler struct {
	serviceName string
	version     string
	log         *slog.Logger
}

// NewHealthHandler creates a new health handler.
func NewHealthHandler(serviceName, version string, log *slog.Logger) *HealthHandler {
	return &HealthHandler{
		serviceName: serviceName,
		version:     version,
		log:         log,
	}
}

// Handle returns the current service health status.
// @Summary Health check
// @Description Returns the current service status.
// @Tags health
// @Produce json
// @Success 200 {object} response.Envelope
// @Router /health [get]
func (h *HealthHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if h.log != nil {
		h.log.DebugContext(r.Context(), "health check", "path", r.URL.Path)
	}

	_ = response.WriteJSON(w, http.StatusOK, response.OK(HealthResponse{
		Service: h.serviceName,
		Status:  "ok",
	}, "ok"))
}

// Version returns the build version info.
// @Summary Application version
// @Description Returns the build version of the application.
// @Tags health
// @Produce json
// @Success 200 {object} response.Envelope
// @Router /version [get]
func (h *HealthHandler) Version(w http.ResponseWriter, _ *http.Request) {
	_ = response.WriteJSON(w, http.StatusOK, response.OK(map[string]string{
		"service": h.serviceName,
		"version": h.version,
	}, "ok"))
}

// RegisterRoutes registers health routes on the given router.
func (h *HealthHandler) RegisterRoutes(r chi.Router) {
	r.Get("/health", h.Handle)
	r.Get("/version", h.Version)
}
