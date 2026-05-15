// Package response defines standard HTTP response envelopes.
package response

import (
	"encoding/json"
	"net/http"
)

// Envelope defines the standard API response shape.
type Envelope struct {
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
	Meta    any    `json:"meta,omitempty"`
}

// OK returns a successful response envelope.
func OK(data any, message string) Envelope {
	return Envelope{
		Message: message,
		Data:    data,
	}
}

// WriteJSON sends a JSON response with the provided status code.
func WriteJSON(w http.ResponseWriter, status int, payload Envelope) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(payload)
}
