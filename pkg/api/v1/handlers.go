package v1

import (
	"encoding/json"
	"net/http"
	"time"
)

// UserListResponse represents a list of users
type UserListResponse struct {
	Data      []User `json:"data"`
	RequestID string `json:"request_id"`
	Timestamp string `json:"timestamp"`
}

// User represents a user resource
type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email,omitempty"`
}

// ProfileResponse represents a user profile
type ProfileResponse struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	RequestID string `json:"request_id"`
	Timestamp string `json:"timestamp"`
}

// Handler is the base handler type for v1 API
type Handler struct {
	requestID string
}

// NewHandler creates a new v1 handler
func NewHandler(requestID string) *Handler {
	return &Handler{
		requestID: requestID,
	}
}

// GetUsers handles GET /v1/users
func (h *Handler) GetUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeV1Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", h.requestID)
		return
	}

	users := []User{
		{ID: "user1", Name: "Alice", Email: "alice@example.com"},
		{ID: "user2", Name: "Bob", Email: "bob@example.com"},
		{ID: "user3", Name: "Charlie", Email: "charlie@example.com"},
	}

	response := UserListResponse{
		Data:      users,
		RequestID: h.requestID,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetProfile handles GET /v1/users/profile
func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeV1Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", h.requestID)
		return
	}

	response := ProfileResponse{
		UserID:    "user-123",
		Email:     "user@example.com",
		RequestID: h.requestID,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// CreateUser handles POST /v1/users
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeV1Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", h.requestID)
		return
	}

	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		writeV1Error(w, http.StatusBadRequest, "INVALID_PAYLOAD", "Failed to parse request body", h.requestID)
		return
	}

	// Generate ID for new user
	user.ID = "user-" + time.Now().Format("20060102150405")

	response := map[string]interface{}{
		"data":       user,
		"request_id": h.requestID,
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// writeV1Error writes a standardized v1 error response
func writeV1Error(w http.ResponseWriter, statusCode int, errorCode, message, requestID string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errResponse := map[string]interface{}{
		"code":        errorCode,
		"message":     message,
		"request_id":  requestID,
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
		"status_code": statusCode,
	}

	json.NewEncoder(w).Encode(errResponse)
}
