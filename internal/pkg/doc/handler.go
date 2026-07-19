package doc

import (
	"log/slog"
	"net/http"
	"sync"
)

const scalarHTML = `<!doctype html>
<html>
<head>
    <title>API Reference</title>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
</head>
<body>
    <script id="api-reference" data-url="/docs/openapi.json"></script>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
</body>
</html>`

// Handler serves the OpenAPI specification and Scalar UI.
type Handler struct {
	builder  *Builder
	log      *slog.Logger
	once     sync.Once
	specJSON []byte
	specErr  error
}

// NewHandler creates a new doc handler.
func NewHandler(builder *Builder, log *slog.Logger) *Handler {
	return &Handler{
		builder: builder,
		log:     log,
	}
}

// ServeHTTP handles the /docs routes.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/docs/openapi.json":
		h.serveSpec(w, r)
	default:
		h.serveUI(w, r)
	}
}

func (h *Handler) serveUI(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(scalarHTML))
}

func (h *Handler) serveSpec(w http.ResponseWriter, r *http.Request) {
	h.once.Do(func() {
		h.specJSON, h.specErr = h.builder.SpecJSON()
	})

	if h.specErr != nil {
		if h.log != nil {
			h.log.ErrorContext(r.Context(), "failed to build OpenAPI spec", "error", h.specErr)
		}
		http.Error(w, "failed to build spec", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(h.specJSON)
}
