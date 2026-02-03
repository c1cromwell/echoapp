# EchoApp - Development Guide

A comprehensive guide for developing, testing, and extending the EchoApp REST API framework.

## Project Structure

```
echoapp/
├── main.go                          # Entry point, server setup, core handlers
├── go.mod                           # Go module dependencies
├── go.sum                           # Dependency checksums
├── README.md                        # Project overview and quick start
├── TESTING.md                       # Testing guide and examples
├── DEVELOPMENT.md                   # This file
└── pkg/
    ├── api/                         # API handlers by version
    │   ├── v1/
    │   │   └── handlers.go         # V1 API handlers and types
    │   └── v2/
    │       └── handlers.go         # V2 API handlers with enhancements
    ├── middleware/                  # HTTP middleware
    │   └── auth.go                 # Auth, CORS, logging middleware
    └── config/                      # Configuration management
        └── config.go               # Config structures and loading
```

## Development Workflow

### 1. Setup Development Environment

```bash
# Clone/navigate to project
cd /Users/thechadcromwell/Projects/echoapp

# Install dependencies
go mod download
go mod tidy

# Build executable
go build -o echoapp main.go

# Or run directly
go run main.go
```

### 2. Code Organization Principles

**Main Package (main.go)**
- Server initialization and lifecycle
- HTTP router setup
- Top-level handler implementations
- Middleware chain setup
- Graceful shutdown logic

**API Packages (pkg/api/vX/)**
- Version-specific handler implementations
- Type definitions for request/response
- Version-specific error handling
- Version-specific middleware if needed

**Middleware Package (pkg/middleware/)**
- Reusable HTTP middleware
- Cross-cutting concerns (auth, logging, CORS)
- Shared utility functions
- Middleware composition helpers

**Config Package (pkg/config/)**
- Configuration structures
- Environment variable loading
- Configuration validation
- Sensible defaults

### 3. Adding New Endpoints

#### Step 1: Define Handler in Version Package

```go
// pkg/api/v1/handlers.go

func (h *Handler) GetUserByID(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        writeV1Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", h.requestID)
        return
    }

    // Extract ID from URL path
    id := r.URL.Path[len("/v1/users/"):]

    response := UserResponse{
        ID:        id,
        Name:      "User " + id,
        Email:     id + "@example.com",
        RequestID: h.requestID,
        Timestamp: time.Now().UTC().Format(time.RFC3339),
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(response)
}
```

#### Step 2: Add Route in Main Router

```go
// main.go - in v1Handler function

case "/v1/users/:id":
    s.v1GetUserByID(w, r)

// Or better: add to main routing switch
case r.URL.Path:
    // Use regex or path splitting for parameter extraction
```

#### Step 3: Test the Endpoint

```bash
curl -H "Authorization: Bearer token" http://localhost:8000/v1/users/123
```

### 4. Adding New Middleware

#### Step 1: Create Middleware Function

```go
// pkg/middleware/auth.go

func RateLimitMiddleware(maxRequests int, window time.Duration) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Check rate limit
            if isRateLimited(getRequestID(r), maxRequests, window) {
                writeErrorResponse(w, http.StatusTooManyRequests, "RATE_LIMITED", 
                    "Too many requests", getRequestID(r))
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

#### Step 2: Apply Middleware to Router

```go
// main.go - in RegisterRoutes method
corsHandler := s.corsMiddleware(
    s.authMiddleware(
        rateLimitMiddleware(100, time.Minute)(
            s.requestIDMiddleware(handler)
        )
    )
)
```

### 5. Adding New Error Codes

Error codes are defined in handler files and returned as JSON responses. To add new ones:

```go
// Define in handler or separate constants file
const (
    ErrorUserNotFound = "USER_NOT_FOUND"
    ErrorInvalidEmail = "INVALID_EMAIL"
    ErrorDuplicateUser = "DUPLICATE_USER"
)

// Use in handler
writeV1Error(w, http.StatusNotFound, ErrorUserNotFound, 
    "User not found with ID: " + id, h.requestID)
```

### 6. Testing Guidelines

#### Unit Tests

```go
// handlers_test.go
package v1

import (
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestGetUsers(t *testing.T) {
    handler := NewHandler("test-request-id")
    
    req := httptest.NewRequest("GET", "/v1/users", nil)
    w := httptest.NewRecorder()
    
    handler.GetUsers(w, req)
    
    if w.Code != http.StatusOK {
        t.Errorf("Expected 200, got %d", w.Code)
    }
}
```

#### Integration Tests

```bash
#!/bin/bash
# Run server in background, test endpoints, verify responses

go run main.go &
PID=$!
sleep 2

# Test endpoints
curl -H "Authorization: Bearer token" http://localhost:8000/v1/users

kill $PID
```

#### Load Tests

```bash
# Using Apache Bench
ab -n 1000 -c 50 \
  -H "Authorization: Bearer token" \
  http://localhost:8000/v1/users
```

### 7. Configuration Management

#### Loading Configuration

```go
// main.go
config := config.LoadConfig()

server := NewAPIServer(APIConfig{
    Port:            config.Server.Port,
    AllowedOrigins:  config.CORS.AllowedOrigins,
    RequireAuthToken: config.Auth.RequireAuth,
})
```

#### Environment Variables

```bash
# Create .env file
API_PORT=8080
ENVIRONMENT=development
LOG_LEVEL=debug
TLS_ENABLED=false
CORS_ENABLED=true

# Load in shell
export $(cat .env | xargs)
go run main.go
```

### 8. Debugging

#### Enable Verbose Logging

```go
// In middleware or handlers
log.Printf("[%s] Debug: %+v", requestID, someData)
```

#### Use Network Inspector

```bash
# Monitor network traffic
curl -v -H "Authorization: Bearer token" http://localhost:8000/v1/users
```

#### Print Request/Response Details

```go
// In middleware
func debugMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Printf("Request: %s %s", r.Method, r.URL.Path)
        log.Printf("Headers: %+v", r.Header)
        
        next.ServeHTTP(w, r)
        
        log.Printf("Response sent to %s", r.RemoteAddr)
    })
}
```

### 9. Performance Optimization

#### Connection Pooling

Configure in main.go:
```go
server := &http.Server{
    Addr:           ":" + config.Server.Port,
    Handler:        mux,
    ReadTimeout:    config.Server.ReadTimeout,
    WriteTimeout:   config.Server.WriteTimeout,
    MaxHeaderBytes: config.Server.MaxHeaderBytes,
}
```

#### Response Caching

Add cache middleware:
```go
func cacheMiddleware(ttl time.Duration) func(http.Handler) http.Handler {
    cache := make(map[string]CachedResponse)
    
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Check cache, return if hit
            // Otherwise call next and cache response
            next.ServeHTTP(w, r)
        })
    }
}
```

### 10. Security Best Practices

#### Input Validation

```go
func validateInput(input string) error {
    if len(input) == 0 {
        return errors.New("input cannot be empty")
    }
    if len(input) > 1024 {
        return errors.New("input too long")
    }
    return nil
}
```

#### Secure Token Validation

Replace placeholder in middleware/auth.go:
```go
func ValidateToken(token string) (string, bool) {
    // TODO: Implement actual passkey verification
    // This should validate against:
    // - Token signature
    // - Token expiration
    // - User authorization
    
    // Example:
    // 1. Parse JWT token
    // 2. Verify signature with public key
    // 3. Check expiration
    // 4. Verify user exists and is active
    
    return extractUserID(token), true
}
```

#### Prevent Common Attacks

- **SQL Injection**: Use parameterized queries
- **XSS**: Validate and escape user input
- **CSRF**: Use CSRF tokens for state-changing operations
- **Rate Limiting**: Implement per-user rate limiting
- **DDoS**: Use reverse proxy with rate limiting

### 11. Building for Production

```bash
# Build optimized binary
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o echoapp-prod main.go

# Generate TLS certificates
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes

# Run with TLS
TLS_CERT_FILE=cert.pem TLS_KEY_FILE=key.pem API_PORT=8443 ./echoapp-prod
```

### 12. Deployment Checklist

- [ ] All tests passing
- [ ] Error handling comprehensive
- [ ] Input validation in place
- [ ] TLS certificates installed
- [ ] CORS origins configured for production
- [ ] Auth token validation implemented
- [ ] Rate limiting configured
- [ ] Logging enabled
- [ ] Health check endpoint working
- [ ] Database connections pooled (if applicable)
- [ ] Environment variables configured
- [ ] Graceful shutdown implemented

### 13. Common Tasks

#### Add a new API version

1. Create `pkg/api/v3/handlers.go`
2. Define v3-specific handlers and types
3. Add v3 routing in main.go
4. Update documentation

#### Add database integration

1. Create `pkg/db/database.go`
2. Implement connection pooling
3. Add query functions
4. Update handlers to use database

#### Add middleware chain

Update RegisterRoutes in main.go:
```go
corsHandler := s.corsMiddleware(
    s.authMiddleware(
        loggingMiddleware(
            s.requestIDMiddleware(handler)
        )
    )
)
```

#### Change error response format

Update `APIError` struct in main.go and all `writeError` calls across packages.

## Resources

- [Go HTTP Package](https://golang.org/pkg/net/http/)
- [REST API Best Practices](https://restfulapi.net/)
- [CORS Explained](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS)
- [TLS 1.3](https://www.rfc-editor.org/rfc/rfc8446)
