// Package response defines standard HTTP response envelopes.
package response

import (
	"encoding/json"
	"net/http"
)

// Envelope defines the standard API response shape.
type Envelope[T any] struct {
	Message string `json:"message,omitempty"`
	Data    T      `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
	Meta    any    `json:"meta,omitempty"`
}

// OK returns a successful response envelope.
func OK[T any](data T, message string) Envelope[T] {
	return Envelope[T]{
		Message: message,
		Data:    data,
	}
}

// Fail returns a failed response envelope.
func Fail(message string) Envelope[struct{}] {
	return Envelope[struct{}]{
		Message: message,
	}
}

// WriteJSON sends a JSON response with the provided status code.
func WriteJSON[T any](w http.ResponseWriter, status int, payload Envelope[T]) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(payload)
}
