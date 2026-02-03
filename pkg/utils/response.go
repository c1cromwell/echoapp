package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Response represents a standard API response
type Response struct {
	Status    string      `json:"status"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	Code      string      `json:"code,omitempty"`
	Message   string      `json:"message,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
	Timestamp string      `json:"timestamp"`
}

// SuccessResponse returns a success response
func SuccessResponse(data interface{}) Response {
	return Response{
		Status:    "success",
		Data:      data,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

// ErrorResponseWithCode returns an error response with a code
func ErrorResponseWithCode(code, message string) Response {
	return Response{
		Code:      code,
		Message:   message,
		Status:    "error",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

// WriteJSON writes a JSON response
func WriteJSON(w http.ResponseWriter, statusCode int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(data)
}

// GetStatusCodeMessage returns a message for an HTTP status code
func GetStatusCodeMessage(statusCode int) (string, string) {
	switch statusCode {
	case http.StatusOK:
		return "OK", "Request successful"
	case http.StatusCreated:
		return "CREATED", "Resource created successfully"
	case http.StatusAccepted:
		return "ACCEPTED", "Request accepted for processing"
	case http.StatusNoContent:
		return "NO_CONTENT", "Operation successful"
	case http.StatusBadRequest:
		return "BAD_REQUEST", "Invalid request parameters"
	case http.StatusUnauthorized:
		return "UNAUTHORIZED", "Authentication required"
	case http.StatusForbidden:
		return "FORBIDDEN", "Access denied"
	case http.StatusNotFound:
		return "NOT_FOUND", "Resource not found"
	case http.StatusConflict:
		return "CONFLICT", "Resource conflict"
	case http.StatusInternalServerError:
		return "INTERNAL_ERROR", "Internal server error"
	case http.StatusServiceUnavailable:
		return "SERVICE_UNAVAILABLE", "Service temporarily unavailable"
	default:
		return "UNKNOWN_ERROR", "Unknown error occurred"
	}
}

// FormatDuration formats a duration in a human-readable way
func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	} else if d < time.Hour {
		return fmt.Sprintf("%.1fm", d.Minutes())
	} else if d < 24*time.Hour {
		return fmt.Sprintf("%.1fh", d.Hours())
	}
	return fmt.Sprintf("%.1fd", d.Hours()/24)
}

// ToJSON converts a value to JSON bytes
func ToJSON(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// FromJSON unmarshals JSON bytes to a value
func FromJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// GinErrorResponse sends a Gin error response
func GinErrorResponse(c *gin.Context, statusCode int, message string, details string) {
	code, _ := GetStatusCodeMessage(statusCode)
	c.JSON(statusCode, gin.H{
		"success":    false,
		"message":    message,
		"code":       code,
		"details":    details,
		"request_id": c.GetString("request_id"),
	})
}
