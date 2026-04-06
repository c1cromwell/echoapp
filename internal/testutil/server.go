package testutil

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/thechadcromwell/echoapp/internal/api"
)

// TestServer wraps a real Echo API router for integration testing.
// It starts on a random port with no TLS and provides helper
// methods for making authenticated HTTP requests.
type TestServer struct {
	BaseURL  string
	Server   *http.Server
	listener net.Listener
	t        *testing.T
}

// StartTestServer creates and starts a test server on a random port
// using the shared api.Router — the same code path as production.
// Call cleanup() when done to shut it down.
func StartTestServer(t *testing.T) (ts *TestServer, cleanup func()) {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}

	port := listener.Addr().(*net.TCPAddr).Port
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)

	// Use the shared router — same as production
	router := api.NewRouter([]string{"*"})

	mux := http.NewServeMux()
	mux.Handle("/", router.Handler())

	server := &http.Server{Handler: mux}

	ts = &TestServer{
		BaseURL:  baseURL,
		Server:   server,
		listener: listener,
		t:        t,
	}

	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			t.Logf("test server error: %v", err)
		}
	}()

	waitForServer(t, baseURL, 3*time.Second)

	cleanup = func() {
		server.Close()
	}

	return ts, cleanup
}

// waitForServer polls the health endpoint until the server is ready.
func waitForServer(t *testing.T, baseURL string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(baseURL + "/health")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("test server did not become ready within %v", timeout)
}

// Get makes an authenticated GET request and returns the response.
func (ts *TestServer) Get(path, token string) *http.Response {
	ts.t.Helper()
	return ts.Do("GET", path, token, nil)
}

// Post makes an authenticated POST request with a JSON body.
func (ts *TestServer) Post(path, token string, body interface{}) *http.Response {
	ts.t.Helper()
	return ts.Do("POST", path, token, body)
}

// Do makes an HTTP request with optional auth and JSON body.
func (ts *TestServer) Do(method, path, token string, body interface{}) *http.Response {
	ts.t.Helper()

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			ts.t.Fatalf("failed to marshal body: %v", err)
		}
		bodyReader = strings.NewReader(string(data))
	}

	req, err := http.NewRequest(method, ts.BaseURL+path, bodyReader)
	if err != nil {
		ts.t.Fatalf("failed to create request: %v", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		ts.t.Fatalf("%s %s failed: %v", method, path, err)
	}
	return resp
}

// DecodeJSON reads and decodes a JSON response body into target.
func (ts *TestServer) DecodeJSON(resp *http.Response, target interface{}) {
	ts.t.Helper()
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		ts.t.Fatalf("failed to decode JSON response: %v", err)
	}
}
