// Package doc provides programmatic OpenAPI spec building and Scalar UI serving.
package doc

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/swaggest/openapi-go"
	"github.com/swaggest/openapi-go/openapi31"

	"github.com/masmuss/gokit-starter/internal/delivery/response"
	"github.com/masmuss/gokit-starter/internal/modules/auth/handler"
)

// Builder builds and serves the OpenAPI specification.
type Builder struct {
	once      sync.Once
	initErr   error
	reflector *openapi31.Reflector
}

// NewBuilder creates a new spec builder.
func NewBuilder(title, version, description string) *Builder {
	r := openapi31.NewReflector()
	spec := r.SpecEns()
	spec.Info.WithTitle(title)
	spec.Info.WithVersion(version)
	spec.Info.WithDescription(description)

	spec.ComponentsEns().WithSecuritySchemesItem("Bearer", openapi31.SecuritySchemeOrReference{
		SecurityScheme: &openapi31.SecurityScheme{
			HTTPBearer: &openapi31.SecuritySchemeHTTPBearer{
				Scheme: "bearer",
			},
		},
	})
	spec.WithSecurity(map[string][]string{"Bearer": {}})

	b := &Builder{reflector: r}
	return b
}

func (b *Builder) init() {
	b.once.Do(func() {
		b.initErr = b.buildOperations()
	})
}

// SpecJSON returns the OpenAPI specification as JSON.
func (b *Builder) SpecJSON() ([]byte, error) {
	b.init()
	if b.initErr != nil {
		return nil, b.initErr
	}
	return json.MarshalIndent(b.reflector.Spec, "", "  ")
}

func (b *Builder) buildOperations() error {
	ops := []struct {
		method      string
		path        string
		summary     string
		description string
		tags        []string
		req         any
		resps       []respSpec
		secured     bool
	}{
		{
			method: http.MethodPost, path: "/auth/register",
			summary:     "Register a new user",
			description: "Register a new user and create an organization",
			tags:        []string{"auth"}, req: handler.RegisterRequest{},
			resps: []respSpec{
				{status: http.StatusCreated, structType: response.Envelope{}},
				{status: http.StatusBadRequest, structType: response.ErrorEnvelope{}},
				{status: http.StatusConflict, structType: response.ErrorEnvelope{}},
			},
		},
		{
			method: http.MethodPost, path: "/auth/login",
			summary:     "User login",
			description: "Authenticate user and return JWT token",
			tags:        []string{"auth"}, req: handler.LoginRequest{},
			resps: []respSpec{
				{status: http.StatusOK, structType: response.Envelope{}},
				{status: http.StatusUnauthorized, structType: response.ErrorEnvelope{}},
			},
		},
		{
			method: http.MethodGet, path: "/auth/profile",
			summary:     "Get current user profile",
			description: "Retrieve profile details for the authenticated user",
			tags:        []string{"auth"}, secured: true,
			resps: []respSpec{
				{status: http.StatusOK, structType: response.Envelope{}},
				{status: http.StatusUnauthorized, structType: response.ErrorEnvelope{}},
			},
		},
		{
			method: http.MethodGet, path: "/health",
			summary:     "Health check",
			description: "Returns the current service status.",
			tags:        []string{"health"},
			resps: []respSpec{
				{status: http.StatusOK, structType: response.Envelope{}},
			},
		},
		{
			method: http.MethodGet, path: "/version",
			summary:     "Application version",
			description: "Returns the build version of the application.",
			tags:        []string{"health"},
			resps: []respSpec{
				{status: http.StatusOK, structType: response.Envelope{}},
			},
		},
	}

	for _, op := range ops {
		oc, err := b.reflector.NewOperationContext(op.method, op.path)
		if err != nil {
			return err
		}

		oc.SetSummary(op.summary)
		oc.SetDescription(op.description)
		oc.SetTags(op.tags...)

		if op.req != nil {
			oc.AddReqStructure(op.req)
		}

		for _, r := range op.resps {
			oc.AddRespStructure(r.structType, openapi.WithHTTPStatus(r.status))
		}

		if op.secured {
			oc.AddSecurity("Bearer")
		}

		if addErr := b.reflector.AddOperation(oc); addErr != nil {
			return addErr
		}
	}

	return nil
}

type respSpec struct {
	status     int
	structType any
}
