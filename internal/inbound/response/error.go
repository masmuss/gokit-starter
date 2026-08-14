package response

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/masmuss/gokit-starter/internal/pkg/apperr"
	"github.com/masmuss/gokit-starter/internal/pkg/validate"
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
	// Handle JSON parsing errors (400 Bad Request).
	if errors.Is(err, validate.ErrInvalidJSON) {
		return WriteError(w, http.StatusBadRequest, "invalid_json", "invalid json payload", nil)
	}

	// Handle struct validation errors (422 Unprocessable Entity).
	if validationErr, ok := errors.AsType[validate.Error](err); ok {
		return WriteError(
			w,
			http.StatusUnprocessableEntity,
			"validation_failed",
			validationErr.Message,
			validationErr.Fields,
		)
	}

	// Handle application errors.
	if appErr, ok := errors.AsType[*apperr.Error](err); ok {
		return WriteError(w, apperr.HTTPStatus(err), appErr.Code, appErr.Message, appErr.Meta)
	}

	// Unknown errors: return generic 500.
	return WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error", nil)
}
