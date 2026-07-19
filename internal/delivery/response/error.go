package response

import (
	"encoding/json"
	"net/http"

	"github.com/masmuss/gokit-starter/internal/pkg/apperr"
)

// ErrorEnvelope defines the standard API error response shape.
type ErrorEnvelope struct {
	Message string `json:"message"`
	Error   string `json:"error"`
	Meta    any    `json:"meta,omitempty"`
}

// fail returns a failed response envelope.
func fail(code, message string, meta any) ErrorEnvelope {
	return ErrorEnvelope{
		Message: message,
		Error:   code,
		Meta:    meta,
	}
}

// WriteError sends a JSON error response with the provided status code.
func WriteError(w http.ResponseWriter, status int, code, message string, meta any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(fail(code, message, meta))
}

// WriteAppError sends a JSON error response based on an apperr.Error.
// Non-apperr errors are returned as generic internal errors to avoid
// leaking details to the client.
func WriteAppError(w http.ResponseWriter, err error) error {
	status := apperr.HTTPStatus(err)
	code := apperr.ErrorCode(err)
	message := err.Error()

	if code == "internal_error" {
		message = "internal server error"
	}

	return WriteError(w, status, code, message, nil)
}
