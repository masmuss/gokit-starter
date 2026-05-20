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

// Fail returns a failed response envelope.
func Fail(code, message string, meta any) ErrorEnvelope {
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

	return json.NewEncoder(w).Encode(Fail(code, message, meta))
}

// WriteAppError sends a JSON error response based on an apperr.Error.
func WriteAppError(w http.ResponseWriter, err error) error {
	status := apperr.HTTPStatus(err)
	code := apperr.ErrorCode(err)
	message := err.Error()

	return WriteError(w, status, code, message, nil)
}
