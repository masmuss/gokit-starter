// Package validation provides request validation helpers.
package validation

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// ErrInvalidJSON indicates the request body is not valid JSON.
var ErrInvalidJSON = errors.New("invalid json payload")

// FieldError describes a single validation failure.
type FieldError struct {
	Field string `json:"field"`
	Rule  string `json:"rule"`
}

// Error represents a validation failure payload.
type Error struct {
	Message string       `json:"message"`
	Fields  []FieldError `json:"fields"`
}

// Error returns the validation summary.
func (e Error) Error() string {
	return e.Message
}

// New returns a validator configured for JSON payloads.
func New() *validator.Validate {
	v := validator.New(validator.WithRequiredStructEnabled())
	v.RegisterTagNameFunc(jsonTagName)

	return v
}

// BindJSON decodes a JSON request body into dst.
func BindJSON(r *http.Request, dst any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return errors.Join(ErrInvalidJSON, err)
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.Join(ErrInvalidJSON, errors.New("unexpected trailing data"))
	}

	return nil
}

// ValidateStruct validates a struct with the provided validator.
func ValidateStruct(v *validator.Validate, value any) error {
	if err := v.Struct(value); err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			fields := make([]FieldError, 0, len(validationErrors))
			for _, validationError := range validationErrors {
				fields = append(fields, FieldError{
					Field: validationError.Field(),
					Rule:  validationError.Tag(),
				})
			}

			return Error{
				Message: "validation failed",
				Fields:  fields,
			}
		}

		return err
	}

	return nil
}

func jsonTagName(field reflect.StructField) string {
	name := strings.Split(field.Tag.Get("json"), ",")[0]
	if name == "-" {
		return ""
	}

	return name
}
