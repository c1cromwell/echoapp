package v2

import (
	"encoding/json"
	"net/http"
	"time"
)

// EnhancedUser represents a user resource with additional v2 fields
type EnhancedUser struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email,omitempty"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// UserListResponseV2 represents a paginated list of users
type UserListResponseV2 struct {
	Data       []EnhancedUser `json:"data"`
	Pagination Pagination     `json:"pagination"`
	RequestID  string         `json:"request_id"`
	Timestamp  string         `json:"timestamp"`
}

// Pagination represents pagination metadata
type Pagination struct {
	Total  int `json:"total"`
	Page   int `json:"page"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// EnhancedProfileResponse represents a user profile with additional v2 fields
type EnhancedProfileResponse struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	Phone     string `json:"phone,omitempty"`
	Verified  bool   `json:"verified"`
	LastLogin string `json:"last_login"`
	CreatedAt string `json:"created_at"`
	RequestID string `json:"request_id"`
	Timestamp string `json:"timestamp"`
}

// Handler is the base handler type for v2 API
type Handler struct {
	requestID string
}

// NewHandler creates a new v2 handler
func NewHandler(requestID string) *Handler {
	return &Handler{
		requestID: requestID,
	}
}

// GetUsers handles GET /v2/users with pagination and enhanced fields
func (h *Handler) GetUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeV2Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", h.requestID)
		return
	}

	users := []EnhancedUser{
		{
			ID:        "user1",
			Name:      "Alice",
			Email:     "alice@example.com",
			Status:    "active",
			CreatedAt: "2025-01-01T00:00:00Z",
			UpdatedAt: "2025-01-12T14:30:00Z",
		},
		{
			ID:        "user2",
			Name:      "Bob",
			Email:     "bob@example.com",
			Status:    "active",
			CreatedAt: "2025-01-02T00:00:00Z",
			UpdatedAt: "2025-01-10T10:15:00Z",
		},
		{
			ID:        "user3",
			Name:      "Charlie",
			Email:     "charlie@example.com",
			Status:    "inactive",
			CreatedAt: "2025-01-03T00:00:00Z",
			UpdatedAt: "2024-12-15T09:00:00Z",
		},
	}

	response := UserListResponseV2{
		Data: users,
		Pagination: Pagination{
			Total:  3,
			Page:   1,
			Limit:  10,
			Offset: 0,
		},
		RequestID: h.requestID,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetProfile handles GET /v2/users/profile with enhanced fields
func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeV2Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", h.requestID)
		return
	}

	response := EnhancedProfileResponse{
		UserID:    "user-123",
		Email:     "user@example.com",
		Phone:     "+1-555-0100",
		Verified:  true,
		LastLogin: time.Now().UTC().AddDate(0, 0, -7).Format(time.RFC3339),
		CreatedAt: "2024-01-01T00:00:00Z",
		RequestID: h.requestID,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// UpdateUser handles PATCH /v2/users/{id}
func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		writeV2Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only PATCH is allowed", h.requestID)
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		writeV2Error(w, http.StatusBadRequest, "INVALID_PAYLOAD", "Failed to parse request body", h.requestID)
		return
	}

	response := map[string]interface{}{
		"data": map[string]interface{}{
			"id":         "user-123",
			"updated":    updates,
			"status":     "success",
			"updated_at": time.Now().UTC().Format(time.RFC3339),
		},
		"request_id": h.requestID,
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// DeleteUser handles DELETE /v2/users/{id}
func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeV2Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only DELETE is allowed", h.requestID)
		return
	}

	response := map[string]interface{}{
		"data": map[string]interface{}{
			"id":         "user-123",
			"deleted_at": time.Now().UTC().Format(time.RFC3339),
		},
		"request_id": h.requestID,
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// writeV2Error writes a standardized v2 error response
func writeV2Error(w http.ResponseWriter, statusCode int, errorCode, message, requestID string) {
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
