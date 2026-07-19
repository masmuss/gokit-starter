package handler

import (
	"net/http"

	"github.com/swaggest/openapi-go/openapi31"

	"github.com/masmuss/gokit-starter/internal/inbound/response"
	"github.com/masmuss/gokit-starter/internal/pkg/doc"
)

// HealthDocRegistrar registers health and version endpoints in the OpenAPI spec.
type HealthDocRegistrar struct{}

// NewHealthDocRegistrar creates a new health doc registrar.
func NewHealthDocRegistrar() *HealthDocRegistrar {
	return &HealthDocRegistrar{}
}

// RegisterOperations implements doc.OperationRegistrar.
func (h *HealthDocRegistrar) RegisterOperations(r *openapi31.Reflector) error {
	ops := []struct {
		method, path, summary, description string
		tags                               []string
		req                                any
		resps                              []doc.RespSpec
		secured                            bool
	}{
		{
			method: http.MethodGet, path: "/health",
			summary: "Health check", description: "Returns the current service status.",
			tags: []string{"health"},
			resps: []doc.RespSpec{
				{Status: http.StatusOK, StructType: response.Envelope{}},
			},
		},
		{
			method: http.MethodGet, path: "/version",
			summary: "Application version", description: "Returns the build version of the application.",
			tags: []string{"health"},
			resps: []doc.RespSpec{
				{Status: http.StatusOK, StructType: response.Envelope{}},
			},
		},
	}

	for _, op := range ops {
		if err := doc.AddOperation(
			r, op.method, op.path, op.summary, op.description,
			op.tags, op.req, op.resps, op.secured,
		); err != nil {
			return err
		}
	}

	return nil
}
