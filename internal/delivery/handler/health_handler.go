// Package handler contains HTTP delivery handlers.
package handler

import (
	"log/slog"
	"net/http"

	"github.com/masmuss/gokit-starter/internal/config"
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
	log         *slog.Logger
}

// NewHealthHandler creates a new health handler.
func NewHealthHandler(serviceName string, log *slog.Logger) *HealthHandler {
	return &HealthHandler{
		serviceName: serviceName,
		log:         log,
	}
}

// NewHealthHandlerFromConfig creates a HealthHandler from config.
func NewHealthHandlerFromConfig(cfg *config.Config, log *slog.Logger) *HealthHandler {
	return NewHealthHandler(cfg.App.Name, log)
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
