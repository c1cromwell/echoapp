# EchoApp - Testing Guide

This guide explains how to test the EchoApp REST API framework with curl and other tools.

## Quick Start Testing

### 1. Start the Server

```bash
cd /Users/thechadcromwell/Projects/echoapp
go run main.go
```

Server will start on `http://localhost:8000`

### 2. Test Health Check (No Auth Required)

```bash
curl -i http://localhost:8000/health
```

Expected Response (200 OK):
```json
{
  "status": "operational",
  "timestamp": "2026-01-13T01:02:00Z",
  "version": "1.0.0",
  "uptime": "1.510031042s",
  "request_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### 3. Test Authentication

#### Without Token (Should Fail)
```bash
curl -i http://localhost:8000/v1/users
```

Expected Response (401 Unauthorized):
```json
{
  "code": "MISSING_AUTH",
  "message": "Authorization header required",
  "request_id": "",
  "timestamp": "2026-01-13T01:02:00Z",
  "status_code": 401
}
```

#### With Valid Token
```bash
curl -i -H "Authorization: Bearer test-token-xyz" http://localhost:8000/v1/users
```

Expected Response (200 OK):
```json
{
  "data": [
    {"id": "user1", "name": "Alice"},
    {"id": "user2", "name": "Bob"}
  ],
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "timestamp": "2026-01-13T01:02:00Z"
}
```

## Comprehensive Test Suite

### Authentication Tests

```bash
# Test 1: Missing Authorization header
curl -i http://localhost:8000/v1/users

# Test 2: Invalid Bearer format
curl -i -H "Authorization: Basic dXNlcjpwYXNz" http://localhost:8000/v1/users

# Test 3: Valid Bearer token
curl -i -H "Authorization: Bearer my-secret-token" http://localhost:8000/v1/users
```

### CORS Tests

```bash
# Test 1: Request with allowed origin
curl -i -H "Origin: http://localhost:3000" \
  -H "Authorization: Bearer token" \
  http://localhost:8000/v1/users

# Test 2: Request with disallowed origin (should fail)
curl -i -H "Origin: http://evil.com" \
  -H "Authorization: Bearer token" \
  http://localhost:8000/v1/users

# Test 3: Preflight request (OPTIONS)
curl -i -X OPTIONS \
  -H "Origin: http://localhost:3000" \
  -H "Access-Control-Request-Method: POST" \
  http://localhost:8000/v1/users
```

### V1 API Tests

```bash
# Get users
curl -H "Authorization: Bearer token" http://localhost:8000/v1/users

# Get user profile
curl -H "Authorization: Bearer token" http://localhost:8000/v1/users/profile

# Create user
curl -X POST \
  -H "Authorization: Bearer token" \
  -H "Content-Type: application/json" \
  -d '{"name":"John","email":"john@example.com"}' \
  http://localhost:8000/v1/users
```

### V2 API Tests (Enhanced)

```bash
# Get users with pagination
curl -H "Authorization: Bearer token" http://localhost:8000/v2/users

# Get detailed profile
curl -H "Authorization: Bearer token" http://localhost:8000/v2/users/profile

# Update user
curl -X PATCH \
  -H "Authorization: Bearer token" \
  -H "Content-Type: application/json" \
  -d '{"status":"inactive"}' \
  http://localhost:8000/v2/users/123

# Delete user
curl -X DELETE \
  -H "Authorization: Bearer token" \
  http://localhost:8000/v2/users/123
```

### Error Handling Tests

```bash
# Test 1: Invalid endpoint
curl -i -H "Authorization: Bearer token" http://localhost:8000/v1/invalid

# Test 2: Invalid HTTP method
curl -i -X DELETE \
  -H "Authorization: Bearer token" \
  http://localhost:8000/v1/users

# Test 3: Invalid JSON payload
curl -i -X POST \
  -H "Authorization: Bearer token" \
  -H "Content-Type: application/json" \
  -d 'invalid-json' \
  http://localhost:8000/v1/users
```

## Testing with Request IDs

All responses include X-Request-ID header and timestamp for request tracking:

```bash
# Send custom request ID
curl -H "X-Request-ID: my-custom-id-123" \
  -H "Authorization: Bearer token" \
  http://localhost:8000/v1/users

# View response headers including request ID
curl -i -H "Authorization: Bearer token" http://localhost:8000/v1/users
```

## Testing with Different Tools

### Using httpie

```bash
# Install: brew install httpie

# Test health check
http http://localhost:8000/health

# Test with auth
http http://localhost:8000/v1/users Authorization:"Bearer token"

# Test POST with JSON
http POST http://localhost:8000/v1/users \
  Authorization:"Bearer token" \
  name=John \
  email=john@example.com
```

### Using curl with verbose output

```bash
# Show request and response headers
curl -v -H "Authorization: Bearer token" http://localhost:8000/v1/users

# Show timing information
curl -w "@curl-format.txt" -o /dev/null -s \
  -H "Authorization: Bearer token" \
  http://localhost:8000/v1/users
```

### Using Postman

1. Create new request
2. Set method to GET
3. URL: `http://localhost:8000/v1/users`
4. Headers tab:
   - Key: `Authorization`
   - Value: `Bearer test-token`
5. Send

### Using curl with shell scripts

```bash
#!/bin/bash

API_BASE="http://localhost:8000"
AUTH_TOKEN="test-token-xyz"

# Function to make API request
api_request() {
  local method=$1
  local endpoint=$2
  local data=$3
  
  if [ -z "$data" ]; then
    curl -X "$method" \
      -H "Authorization: Bearer $AUTH_TOKEN" \
      "$API_BASE$endpoint"
  else
    curl -X "$method" \
      -H "Authorization: Bearer $AUTH_TOKEN" \
      -H "Content-Type: application/json" \
      -d "$data" \
      "$API_BASE$endpoint"
  fi
}

# Make requests
echo "Getting users..."
api_request "GET" "/v1/users"

echo -e "\n\nGetting profile..."
api_request "GET" "/v1/users/profile"
```

## Performance Testing

### Using Apache Bench

```bash
# Install: brew install httpd

# Test endpoint with 100 requests, 10 concurrent
ab -n 100 -c 10 \
  -H "Authorization: Bearer token" \
  http://localhost:8000/v1/users
```

### Using wrk

```bash
# Install: brew install wrk

# Test with 4 threads for 30 seconds
wrk -t4 -c100 -d30s \
  -H "Authorization: Bearer token" \
  http://localhost:8000/v1/users
```

## Debugging Tips

### Enable verbose logging

Set environment variable before starting server:
```bash
export LOG_LEVEL=debug
export ENVIRONMENT=development
go run main.go
```

### Inspect network traffic

```bash
# Using mitmproxy
mitmproxy -p 8081

# Configure curl to use proxy
curl -x localhost:8081 \
  -H "Authorization: Bearer token" \
  http://localhost:8000/v1/users
```

### Check server response headers

```bash
curl -i -H "Authorization: Bearer token" http://localhost:8000/v1/users
```

Key headers to look for:
- `X-Request-ID` - Unique request identifier
- `Access-Control-Allow-Origin` - CORS origin
- `Content-Type` - Response format (should be application/json)

## Troubleshooting

### Issue: Connection refused
- **Cause**: Server not running
- **Solution**: Start server with `go run main.go`

### Issue: 401 Unauthorized
- **Cause**: Missing or invalid Authorization header
- **Solution**: Add `-H "Authorization: Bearer <token>"` to request

### Issue: 403 Forbidden (CORS)
- **Cause**: Origin not in allowed list
- **Solution**: Request from allowed origin or update CORS config

### Issue: Invalid JSON in response
- **Cause**: Server error or invalid endpoint
- **Solution**: Check endpoint URL and use `-i` flag to see status code

## Integration Testing

Example test script that verifies all major features:

```bash
#!/bin/bash

set -e

API_URL="http://localhost:8000"
TOKEN="test-token-xyz"

echo "🧪 Running EchoApp Integration Tests"
echo ""

# Test 1: Health Check
echo "✓ Test 1: Health Check"
curl -s -f "$API_URL/health" > /dev/null

# Test 2: Missing Auth
echo "✓ Test 2: Missing Auth (should fail)"
curl -s -f "$API_URL/v1/users" 2>&1 | grep -q "MISSING_AUTH" || true

# Test 3: With Auth
echo "✓ Test 3: V1 Users endpoint"
curl -s -f -H "Authorization: Bearer $TOKEN" "$API_URL/v1/users" > /dev/null

# Test 4: V2 Users
echo "✓ Test 4: V2 Users endpoint with pagination"
curl -s -f -H "Authorization: Bearer $TOKEN" "$API_URL/v2/users" | grep -q "pagination" || true

# Test 5: Request ID tracking
echo "✓ Test 5: Request ID tracking"
RESPONSE=$(curl -s -H "Authorization: Bearer $TOKEN" "$API_URL/v1/users")
echo "$RESPONSE" | grep -q "request_id" || true

echo ""
echo "✅ All tests passed!"
```

Save as `test.sh` and run:
```bash
chmod +x test.sh
./test.sh
```
