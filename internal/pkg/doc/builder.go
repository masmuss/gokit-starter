// Package doc provides programmatic OpenAPI spec building and Scalar UI serving.
package doc

import (
	"encoding/json"
	"sync"

	"github.com/swaggest/openapi-go"
	"github.com/swaggest/openapi-go/openapi31"
)

// OperationRegistrar allows modules to register their OpenAPI operations.
type OperationRegistrar interface {
	RegisterOperations(r *openapi31.Reflector) error
}

// Builder builds and serves the OpenAPI specification.
type Builder struct {
	once       sync.Once
	initErr    error
	reflector  *openapi31.Reflector
	registrars []OperationRegistrar
}

// NewBuilder creates a new spec builder that collects operations from registrars.
func NewBuilder(title, version, description string, registrars []OperationRegistrar) *Builder {
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

	return &Builder{reflector: r, registrars: registrars}
}

func (b *Builder) init() {
	b.once.Do(func() {
		for _, registrar := range b.registrars {
			if err := registrar.RegisterOperations(b.reflector); err != nil {
				b.initErr = err
				return
			}
		}
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

// AddOperation is a helper to register a single operation on a reflector.
func AddOperation(
	r *openapi31.Reflector,
	method, path, summary, description string,
	tags []string,
	req any,
	resps []RespSpec,
	secured bool,
) error {
	oc, err := r.NewOperationContext(method, path)
	if err != nil {
		return err
	}

	oc.SetSummary(summary)
	oc.SetDescription(description)
	oc.SetTags(tags...)

	if req != nil {
		oc.AddReqStructure(req)
	}

	for _, resp := range resps {
		oc.AddRespStructure(resp.StructType, openapi.WithHTTPStatus(resp.Status))
	}

	if secured {
		oc.AddSecurity("Bearer")
	}

	return r.AddOperation(oc)
}

// RespSpec describes an API response for the OpenAPI spec.
type RespSpec struct {
	Status     int
	StructType any
}
