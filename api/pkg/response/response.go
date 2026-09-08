// Package response provides helpers for writing HTTP JSON responses with a
// consistent format across the whole application.
package response

import (
	"encoding/json"
	"net/http"
)

// body is the standard JSON response structure.
//
//	success=true  → the Data field is populated
//	success=false → the Message / Errors fields are populated
type body struct {
	Success bool     `json:"success"`           // status: success or failure
	Data    any      `json:"data,omitempty"`    // payload data (on success)
	Message string   `json:"message,omitempty"` // short message (on failure)
	Errors  []string `json:"errors,omitempty"`  // detailed error list
}

// Success writes a success response with the given status code and data.
func Success(w http.ResponseWriter, status int, data any) {
	write(w, status, body{Success: true, Data: data})
}

// Error writes a failure response with the given status code and message.
func Error(w http.ResponseWriter, status int, message string) {
	write(w, status, body{Success: false, Message: message})
}

// ValidationError writes a 400 response specific to validation errors
// (standard message + error list).
func ValidationError(w http.ResponseWriter, err error) {
	write(w, http.StatusBadRequest, body{
		Success: false,
		Message: "validation failed",
		Errors:  []string{err.Error()},
	})
}

// write is an internal helper: sets the JSON header, writes the status, and
// encodes the body.
func write(w http.ResponseWriter, status int, b body) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(b)
}
