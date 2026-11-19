package response

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse represents a standard structured error payload.
type ErrorResponse struct {
	Error   string            `json:"error"`
	Code    int               `json:"code"`
	Details map[string]string `json:"details,omitempty"`
}

// JSON sends a JSON response with status code.
func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// Error sends a JSON formatted error response.
func Error(w http.ResponseWriter, status int, message string, details map[string]string) {
	JSON(w, status, ErrorResponse{
		Error:   message,
		Code:    status,
		Details: details,
	})
}
