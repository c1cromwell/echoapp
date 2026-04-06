package integration

import (
	"net/http"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/testutil"
)

// Contract tests validate that API responses match the expected shape
// defined in openapi.yaml. When you update the spec or handlers,
// these tests catch drift between them.
//
// Each test validates:
//   - HTTP status code
//   - Content-Type header
//   - Required fields in the response body
//   - Field types (string, number, bool, array, object)

// assertFields checks that all required keys exist in a response map.
func assertFields(t *testing.T, endpoint string, body map[string]interface{}, required []string) {
	t.Helper()
	for _, key := range required {
		if _, ok := body[key]; !ok {
			t.Errorf("[%s] missing required field %q in response", endpoint, key)
		}
	}
}

// assertString checks a field is a non-empty string.
func assertString(t *testing.T, endpoint, field string, body map[string]interface{}) {
	t.Helper()
	val, ok := body[field]
	if !ok {
		t.Errorf("[%s] missing field %q", endpoint, field)
		return
	}
	if _, ok := val.(string); !ok {
		t.Errorf("[%s] field %q should be string, got %T", endpoint, field, val)
	}
}

// assertArray checks a field is an array.
func assertArray(t *testing.T, endpoint, field string, body map[string]interface{}) []interface{} {
	t.Helper()
	val, ok := body[field]
	if !ok {
		t.Errorf("[%s] missing field %q", endpoint, field)
		return nil
	}
	arr, ok := val.([]interface{})
	if !ok {
		t.Errorf("[%s] field %q should be array, got %T", endpoint, field, val)
		return nil
	}
	return arr
}

// assertObject checks a field is an object (map).
func assertObject(t *testing.T, endpoint, field string, body map[string]interface{}) map[string]interface{} {
	t.Helper()
	val, ok := body[field]
	if !ok {
		t.Errorf("[%s] missing field %q", endpoint, field)
		return nil
	}
	obj, ok := val.(map[string]interface{})
	if !ok {
		t.Errorf("[%s] field %q should be object, got %T", endpoint, field, val)
		return nil
	}
	return obj
}

// assertContentType checks the Content-Type header.
func assertContentType(t *testing.T, endpoint string, resp *http.Response, expected string) {
	t.Helper()
	ct := resp.Header.Get("Content-Type")
	if ct != expected {
		t.Errorf("[%s] Content-Type = %q, want %q", endpoint, ct, expected)
	}
}

// --- Contract: GET /health ---
// Expected shape from openapi.yaml:
//   { status: string, timestamp: string, version: string, uptime: string, request_id: string }

func TestContract_Health(t *testing.T) {
	ts, cleanup := testutil.StartTestServer(t)
	defer cleanup()

	endpoint := "GET /health"
	resp := ts.Get("/health", "")

	assertContentType(t, endpoint, resp, "application/json")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("[%s] status = %d, want 200", endpoint, resp.StatusCode)
	}

	var body map[string]interface{}
	ts.DecodeJSON(resp, &body)

	assertFields(t, endpoint, body, []string{"status", "timestamp", "version", "uptime", "request_id"})
	assertString(t, endpoint, "status", body)
	assertString(t, endpoint, "timestamp", body)
	assertString(t, endpoint, "version", body)
	assertString(t, endpoint, "uptime", body)
	assertString(t, endpoint, "request_id", body)
}

// --- Contract: GET /v1/users ---
// Expected shape:
//   { data: [{ id: string, name: string }], request_id: string, timestamp: string }

func TestContract_V1Users(t *testing.T) {
	ts, cleanup := testutil.StartTestServer(t)
	defer cleanup()

	endpoint := "GET /v1/users"
	resp := ts.Get("/v1/users", "testtoken1234")

	assertContentType(t, endpoint, resp, "application/json")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("[%s] status = %d, want 200", endpoint, resp.StatusCode)
	}

	var body map[string]interface{}
	ts.DecodeJSON(resp, &body)

	assertFields(t, endpoint, body, []string{"data", "request_id", "timestamp"})
	assertString(t, endpoint, "request_id", body)
	assertString(t, endpoint, "timestamp", body)

	data := assertArray(t, endpoint, "data", body)
	if len(data) == 0 {
		t.Fatalf("[%s] data array is empty", endpoint)
	}

	// Each user must have id and name
	for i, item := range data {
		user, ok := item.(map[string]interface{})
		if !ok {
			t.Errorf("[%s] data[%d] should be object", endpoint, i)
			continue
		}
		assertFields(t, endpoint, user, []string{"id", "name"})
		assertString(t, endpoint, "id", user)
		assertString(t, endpoint, "name", user)
	}
}

// --- Contract: GET /v1/users/profile ---
// Expected shape:
//   { user_id: string, email: string, request_id: string, timestamp: string }

func TestContract_V1Profile(t *testing.T) {
	ts, cleanup := testutil.StartTestServer(t)
	defer cleanup()

	endpoint := "GET /v1/users/profile"
	resp := ts.Get("/v1/users/profile", "testtoken1234")

	assertContentType(t, endpoint, resp, "application/json")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("[%s] status = %d, want 200", endpoint, resp.StatusCode)
	}

	var body map[string]interface{}
	ts.DecodeJSON(resp, &body)

	assertFields(t, endpoint, body, []string{"user_id", "email", "request_id", "timestamp"})
	assertString(t, endpoint, "user_id", body)
	assertString(t, endpoint, "email", body)
}

// --- Contract: GET /v2/users ---
// Expected shape (enhanced with pagination):
//   { data: [{ id, name, status, created_at }], pagination: { total, page, limit }, request_id, timestamp }

func TestContract_V2Users(t *testing.T) {
	ts, cleanup := testutil.StartTestServer(t)
	defer cleanup()

	endpoint := "GET /v2/users"
	resp := ts.Get("/v2/users", "testtoken1234")

	assertContentType(t, endpoint, resp, "application/json")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("[%s] status = %d, want 200", endpoint, resp.StatusCode)
	}

	var body map[string]interface{}
	ts.DecodeJSON(resp, &body)

	assertFields(t, endpoint, body, []string{"data", "pagination", "request_id", "timestamp"})

	// Validate pagination object
	pagination := assertObject(t, endpoint, "pagination", body)
	if pagination != nil {
		assertFields(t, endpoint+".pagination", pagination, []string{"total", "page", "limit"})
	}

	// Validate user objects have v2 fields
	data := assertArray(t, endpoint, "data", body)
	for i, item := range data {
		user, ok := item.(map[string]interface{})
		if !ok {
			t.Errorf("[%s] data[%d] should be object", endpoint, i)
			continue
		}
		assertFields(t, endpoint, user, []string{"id", "name", "status", "created_at"})
	}
}

// --- Contract: GET /v2/users/profile ---
// Expected shape (enhanced with extra fields):
//   { user_id, email, phone, verified, last_login, created_at, request_id, timestamp }

func TestContract_V2Profile(t *testing.T) {
	ts, cleanup := testutil.StartTestServer(t)
	defer cleanup()

	endpoint := "GET /v2/users/profile"
	resp := ts.Get("/v2/users/profile", "testtoken1234")

	assertContentType(t, endpoint, resp, "application/json")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("[%s] status = %d, want 200", endpoint, resp.StatusCode)
	}

	var body map[string]interface{}
	ts.DecodeJSON(resp, &body)

	assertFields(t, endpoint, body, []string{
		"user_id", "email", "phone", "verified",
		"last_login", "created_at", "request_id", "timestamp",
	})
	assertString(t, endpoint, "user_id", body)
	assertString(t, endpoint, "email", body)
	assertString(t, endpoint, "phone", body)
	assertString(t, endpoint, "last_login", body)
	assertString(t, endpoint, "created_at", body)

	// verified should be bool
	if _, ok := body["verified"].(bool); !ok {
		t.Errorf("[%s] 'verified' should be bool, got %T", endpoint, body["verified"])
	}
}

// --- Contract: Error responses ---
// All errors should have: { code, message, request_id, timestamp, status_code }

func TestContract_ErrorResponse_Unauthorized(t *testing.T) {
	ts, cleanup := testutil.StartTestServer(t)
	defer cleanup()

	endpoint := "GET /v1/users (no auth)"
	resp := ts.Get("/v1/users", "")

	assertContentType(t, endpoint, resp, "application/json")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("[%s] status = %d, want 401", endpoint, resp.StatusCode)
	}

	var body map[string]interface{}
	ts.DecodeJSON(resp, &body)

	assertFields(t, endpoint, body, []string{"code", "message", "request_id", "timestamp", "status_code"})
	assertString(t, endpoint, "code", body)
	assertString(t, endpoint, "message", body)
}

func TestContract_ErrorResponse_NotFound(t *testing.T) {
	ts, cleanup := testutil.StartTestServer(t)
	defer cleanup()

	endpoint := "GET /v1/nonexistent"
	resp := ts.Get("/v1/nonexistent", "testtoken1234")

	assertContentType(t, endpoint, resp, "application/json")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("[%s] status = %d, want 404", endpoint, resp.StatusCode)
	}

	var body map[string]interface{}
	ts.DecodeJSON(resp, &body)

	assertFields(t, endpoint, body, []string{"code", "message", "request_id", "timestamp", "status_code"})
}
