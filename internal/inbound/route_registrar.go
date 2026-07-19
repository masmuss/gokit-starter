// Package delivery provides shared types for HTTP delivery layer.
package delivery

import "github.com/go-chi/chi/v5"

// RouteRegistrar allows modules to register their routes on a chi router.
type RouteRegistrar interface {
	RegisterRoutes(r chi.Router)
}

// RouteRegistrarFunc is an adapter to use a function as a RouteRegistrar.
type RouteRegistrarFunc func(r chi.Router)

// RegisterRoutes implements RouteRegistrar.
func (f RouteRegistrarFunc) RegisterRoutes(r chi.Router) {
	f(r)
}
