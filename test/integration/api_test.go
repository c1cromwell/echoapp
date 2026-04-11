package integration

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/testutil"
)

// TestHealthEndpoint verifies the /health endpoint responds
// without authentication and returns operational status.
func TestHealthEndpoint(t *testing.T) {
	ts, cleanup := testutil.StartTestServer(t)
	defer cleanup()

	resp := ts.Get("/health", "") // no auth needed
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	ts.DecodeJSON(resp, &body)

	if body["status"] != "operational" {
		t.Errorf("expected status 'operational', got %v", body["status"])
	}
	if body["version"] != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %v", body["version"])
	}
	if body["request_id"] == nil || body["request_id"] == "" {
		t.Error("expected request_id to be set")
	}
}

// TestUnauthorizedRequest verifies that endpoints reject
// requests without an Authorization header.
func TestUnauthorizedRequest(t *testing.T) {
	ts, cleanup := testutil.StartTestServer(t)
	defer cleanup()

	resp := ts.Get("/v1/users", "") // no token
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	ts.DecodeJSON(resp, &body)

	if body["code"] != "MISSING_AUTH" {
		t.Errorf("expected error code 'MISSING_AUTH', got %v", body["code"])
	}
}

// TestInvalidAuthFormat verifies that a malformed Authorization
// header is rejected with the correct error code.
func TestInvalidAuthFormat(t *testing.T) {
	ts, cleanup := testutil.StartTestServer(t)
	defer cleanup()

	resp := ts.Do("GET", "/v1/users", "", nil)
	resp.Body.Close()

	// Try with bad format (not "Bearer <token>")
	req, _ := http.NewRequest("GET", ts.BaseURL+"/v1/users", nil)
	req.Header.Set("Authorization", "Basic abc123")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	ts.DecodeJSON(resp, &body)

	if body["code"] != "INVALID_AUTH_FORMAT" {
		t.Errorf("expected 'INVALID_AUTH_FORMAT', got %v", body["code"])
	}
}

// TestV1GetUsers verifies the GET /v1/users endpoint
// returns the expected user list.
func TestV1GetUsers(t *testing.T) {
	ts, cleanup := testutil.StartTestServer(t)
	defer cleanup()

	token := ts.IssueTestToken("did:echo:testuser")
	resp := ts.Get("/v1/users", token)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	ts.DecodeJSON(resp, &body)

	data, ok := body["data"].([]interface{})
	if !ok {
		t.Fatal("expected 'data' to be an array")
	}
	if len(data) != 2 {
		t.Fatalf("expected 2 users, got %d", len(data))
	}

	user1 := data[0].(map[string]interface{})
	if user1["name"] != "Alice" {
		t.Errorf("expected first user 'Alice', got %v", user1["name"])
	}
}

// TestV1GetProfile verifies the GET /v1/users/profile endpoint
// returns profile data with the user ID extracted from the token.
func TestV1GetProfile(t *testing.T) {
	ts, cleanup := testutil.StartTestServer(t)
	defer cleanup()

	token := ts.IssueTestToken("did:echo:testuser")
	resp := ts.Get("/v1/users/profile", token)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	ts.DecodeJSON(resp, &body)

	expectedUserID := "did:echo:testuser"
	if body["user_id"] != expectedUserID {
		t.Errorf("expected user_id '%s', got %v", expectedUserID, body["user_id"])
	}
	if body["email"] != "user@example.com" {
		t.Errorf("expected email 'user@example.com', got %v", body["email"])
	}
}

// TestV2GetUsers verifies the v2 users endpoint includes
// pagination metadata and enhanced user fields.
func TestV2GetUsers(t *testing.T) {
	ts, cleanup := testutil.StartTestServer(t)
	defer cleanup()

	token := ts.IssueTestToken("did:echo:testuser")
	resp := ts.Get("/v2/users", token)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	ts.DecodeJSON(resp, &body)

	// V2 should include pagination
	pagination, ok := body["pagination"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'pagination' object in v2 response")
	}
	if pagination["total"] != float64(2) {
		t.Errorf("expected total 2, got %v", pagination["total"])
	}

	// V2 users should have status and created_at fields
	data := body["data"].([]interface{})
	user1 := data[0].(map[string]interface{})
	if user1["status"] != "active" {
		t.Errorf("expected status 'active', got %v", user1["status"])
	}
	if user1["created_at"] == nil {
		t.Error("expected created_at field in v2 user")
	}
}

// TestV2GetProfile verifies the v2 profile endpoint includes
// enhanced fields like phone, verified, and last_login.
func TestV2GetProfile(t *testing.T) {
	ts, cleanup := testutil.StartTestServer(t)
	defer cleanup()

	token := ts.IssueTestToken("did:echo:testuser")
	resp := ts.Get("/v2/users/profile", token)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	ts.DecodeJSON(resp, &body)

	if body["phone"] != "+1-555-0100" {
		t.Errorf("expected phone '+1-555-0100', got %v", body["phone"])
	}
	if body["verified"] != true {
		t.Errorf("expected verified true, got %v", body["verified"])
	}
	if body["last_login"] == nil {
		t.Error("expected last_login in v2 profile")
	}
}

// TestNotFoundEndpoint verifies that unknown paths return 404.
func TestNotFoundEndpoint(t *testing.T) {
	ts, cleanup := testutil.StartTestServer(t)
	defer cleanup()

	resp := ts.Get("/v1/nonexistent", ts.IssueTestToken("did:echo:testuser"))
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

// TestMethodNotAllowed verifies that POST to a GET-only
// endpoint returns 405.
func TestMethodNotAllowed(t *testing.T) {
	ts, cleanup := testutil.StartTestServer(t)
	defer cleanup()

	resp := ts.Post("/v1/users", ts.IssueTestToken("did:echo:testuser"), map[string]string{"name": "Charlie"})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		var body map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&body)
		t.Fatalf("expected 405, got %d (body: %v)", resp.StatusCode, body)
	}
}

// TestRequestIDPropagation verifies that a client-provided
// X-Request-ID is echoed back in the response.
func TestRequestIDPropagation(t *testing.T) {
	ts, cleanup := testutil.StartTestServer(t)
	defer cleanup()

	req, _ := http.NewRequest("GET", ts.BaseURL+"/health", nil)
	customID := "my-custom-request-id-123"
	req.Header.Set("X-Request-ID", customID)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	gotID := resp.Header.Get("X-Request-ID")
	if gotID != customID {
		t.Errorf("expected X-Request-ID '%s', got '%s'", customID, gotID)
	}
}
