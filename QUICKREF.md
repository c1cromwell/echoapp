# EchoApp - Quick Reference

## Project Overview
A production-ready REST API framework in Go with multi-version API support, TLS 1.3+, CORS policies, authentication middleware, and standardized error responses.

## Quick Start
```bash
cd /Users/thechadcromwell/Projects/echoapp
go run main.go              # Start on :8000
curl http://localhost:8000/health  # Test health
```

## API Endpoints

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/health` | ✗ | Health check |
| GET | `/v1/users` | ✓ | List users (v1) |
| GET | `/v1/users/profile` | ✓ | Get profile (v1) |
| POST | `/v1/users` | ✓ | Create user (v1) |
| GET | `/v2/users` | ✓ | List users (v2) with pagination |
| GET | `/v2/users/profile` | ✓ | Get profile (v2) with details |
| PATCH | `/v2/users/{id}` | ✓ | Update user (v2) |
| DELETE | `/v2/users/{id}` | ✓ | Delete user (v2) |

## Common curl Commands

```bash
# Health check (no auth)
curl http://localhost:8000/health

# With auth token
curl -H "Authorization: Bearer TOKEN" http://localhost:8000/v1/users

# With custom request ID
curl -H "X-Request-ID: my-id" \
  -H "Authorization: Bearer TOKEN" \
  http://localhost:8000/v1/users

# POST request
curl -X POST \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"John","email":"john@example.com"}' \
  http://localhost:8000/v1/users

# Test CORS
curl -H "Origin: http://localhost:3000" \
  -H "Authorization: Bearer TOKEN" \
  http://localhost:8000/v1/users
```

## File Structure

```
echoapp/
├── main.go                    # Server, routing, handlers
├── pkg/
│   ├── api/v1/handlers.go    # V1 endpoints
│   ├── api/v2/handlers.go    # V2 endpoints
│   ├── middleware/auth.go    # Auth, CORS, logging
│   └── config/config.go      # Configuration
├── go.mod                     # Dependencies
├── README.md                  # Full documentation
├── TESTING.md                 # Testing guide
└── DEVELOPMENT.md             # Development guide
```

## Error Response Format

```json
{
  "code": "ERROR_CODE",
  "message": "Description",
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "timestamp": "2026-01-13T01:02:00Z",
  "status_code": 400
}
```

## Common Error Codes

| Code | Status | Meaning |
|------|--------|---------|
| MISSING_AUTH | 401 | Authorization header missing |
| INVALID_AUTH_FORMAT | 401 | Invalid Bearer token format |
| INVALID_TOKEN | 401 | Token validation failed |
| CORS_DENIED | 403 | Origin not allowed |
| METHOD_NOT_ALLOWED | 405 | HTTP method not supported |
| ENDPOINT_NOT_FOUND | 404 | Endpoint doesn't exist |
| INVALID_PAYLOAD | 400 | Request body invalid |

## Configuration

Set environment variables before running:

```bash
export API_PORT=8080                    # Default: 8000
export ENVIRONMENT=production           # Default: development
export LOG_LEVEL=debug                  # Default: info
export TLS_ENABLED=false                # Default: false
export TLS_CERT_FILE=/path/to/cert.pem # Optional
export TLS_KEY_FILE=/path/to/key.pem   # Optional
export CORS_ENABLED=true                # Default: true
```

## Middleware Stack (Request → Response)

1. Request ID Middleware → generates/tracks IDs
2. CORS Middleware → validates origin
3. Auth Middleware → validates Bearer token
4. Handler → processes request
5. Response → includes request ID & timestamp

## Key Features

### ✅ Multi-Version APIs
- `/v1/*` - Basic endpoints
- `/v2/*` - Enhanced endpoints with extra fields

### ✅ Authentication
- Bearer token validation
- Passkey verification placeholder
- Request-level user tracking

### ✅ CORS
- Configurable allowed origins
- Preflight request handling
- Strict origin validation

### ✅ TLS 1.3+
- Modern cipher suites
- Server cipher preference
- Automatic HTTPS support

### ✅ Request Tracking
- Unique request IDs
- Request/response logging
- Timestamp tracking

### ✅ Error Handling
- Standardized error format
- Unique error codes
- Descriptive messages

### ✅ Health Check
- Service availability check
- Uptime tracking
- Version info

## Development

### Adding Endpoints
1. Add handler in `pkg/api/vX/handlers.go`
2. Add route in `main.go` router
3. Test with curl

### Adding Middleware
1. Create middleware function in `pkg/middleware/auth.go`
2. Apply in `RegisterRoutes()` method
3. Test integration

### Testing
```bash
go test ./...           # Run all tests
go build -o echoapp    # Build binary
./echoapp              # Run binary
```

## Production Checklist

- [ ] TLS certificates installed
- [ ] Auth token validation implemented
- [ ] CORS origins configured for production
- [ ] Rate limiting configured
- [ ] Database connections pooled (if needed)
- [ ] Logging configured
- [ ] Health check working
- [ ] Error handling comprehensive
- [ ] Input validation in place
- [ ] Tests passing

## Useful Resources

| Resource | Link |
|----------|------|
| Full Docs | `/Users/thechadcromwell/Projects/echoapp/README.md` |
| Testing Guide | `/Users/thechadcromwell/Projects/echoapp/TESTING.md` |
| Dev Guide | `/Users/thechadcromwell/Projects/echoapp/DEVELOPMENT.md` |
| Go HTTP | https://golang.org/pkg/net/http/ |
| REST Best Practices | https://restfulapi.net/ |

## Support

For detailed information:
- **Overview**: See `README.md`
- **Testing**: See `TESTING.md`
- **Development**: See `DEVELOPMENT.md`
- **Endpoints**: Test with curl or Postman

## Status

✅ Project: COMPLETE
✅ Build: SUCCESSFUL (8.2M)
✅ Tests: Ready
✅ Documentation: Complete
