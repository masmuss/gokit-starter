package response

import (
	"errors"
	"fmt"
	"net/http"
)

// Kind defines the category of an error.
type Kind uint8

// Standard error kinds.
const (
	KindUnknown Kind = iota
	KindInternal
	KindNotFound
	KindUnauthorized
	KindForbidden
	KindBadRequest
	KindConflict
	KindValidation
)

// AppError represents a standardized application error.
type AppError struct {
	Kind    Kind
	Code    string
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the underlying error.
func (e *AppError) Unwrap() error {
	return e.Err
}

// New creates a new AppError.
func New(kind Kind, code, message string) *AppError {
	return &AppError{
		Kind:    kind,
		Code:    code,
		Message: message,
	}
}

// Wrap wraps an existing error into an AppError.
func Wrap(err error, kind Kind, code, message string) *AppError {
	return &AppError{
		Kind:    kind,
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// HTTPStatus returns the HTTP status code associated with the error kind.
func HTTPStatus(err error) int {
	var appErr *AppError
	if !errors.As(err, &appErr) {
		return http.StatusInternalServerError
	}

	switch appErr.Kind {
	case KindNotFound:
		return http.StatusNotFound
	case KindUnauthorized:
		return http.StatusUnauthorized
	case KindForbidden:
		return http.StatusForbidden
	case KindBadRequest:
		return http.StatusBadRequest
	case KindConflict:
		return http.StatusConflict
	case KindValidation:
		return http.StatusUnprocessableEntity
	case KindInternal:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// ErrorCode returns the machine-readable error code.
func ErrorCode(err error) string {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return "internal_error"
}

// Internal creates a KindInternal error.
func Internal(code, message string) *AppError {
	return New(KindInternal, code, message)
}

// NotFound creates a KindNotFound error.
func NotFound(code, message string) *AppError {
	return New(KindNotFound, code, message)
}

// Unauthorized creates a KindUnauthorized error.
func Unauthorized(code, message string) *AppError {
	return New(KindUnauthorized, code, message)
}

// Forbidden creates a KindForbidden error.
func Forbidden(code, message string) *AppError {
	return New(KindForbidden, code, message)
}

// BadRequest creates a KindBadRequest error.
func BadRequest(code, message string) *AppError {
	return New(KindBadRequest, code, message)
}

// Conflict creates a KindConflict error.
func Conflict(code, message string) *AppError {
	return New(KindConflict, code, message)
}

// Validation creates a KindValidation error.
func Validation(code, message string) *AppError {
	return New(KindValidation, code, message)
}
