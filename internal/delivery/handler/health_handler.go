// Package handler contains HTTP delivery handlers.
package handler

import (
	"log/slog"
	"net/http"

	"github.com/masmuss/gokit-starter/internal/delivery/response"
)

type healthResponse struct {
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

// Handle returns the current service health status.
func (h *HealthHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if h.log != nil {
		h.log.DebugContext(r.Context(), "health check", "path", r.URL.Path)
	}

	_ = response.WriteJSON(w, http.StatusOK, response.OK(healthResponse{
		Service: h.serviceName,
		Status:  "ok",
	}, "ok"))
}
