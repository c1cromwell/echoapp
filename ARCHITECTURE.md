# EchoApp Architecture Guide

## Project Structure

```
echoapp/
├── cmd/                          # Command-line entry points
│   ├── credentials/              # Credentials service
│   │   └── main.go              # Service initialization
│   ├── did/                      # DID service
│   │   └── main.go              # Service initialization
│   └── cardanoidentity/          # Cardano identity service
│       └── main.go              # Service initialization
│
├── internal/                     # Internal packages (not importable)
│   └── crypto/
│       └── ed25519.go           # Cryptographic utilities
│
├── pkg/                          # Public packages
│   ├── api/                      # HTTP API handlers
│   │   ├── handlers/
│   │   │   └── identity.go      # Identity HTTP handlers
│   │   ├── v1/
│   │   │   └── handlers.go      # V1 API handlers
│   │   └── v2/
│   │       └── handlers.go      # V2 API handlers
│   │
│   ├── cardano/                  # Blockchain integration
│   │   ├── client.go            # Cardano client
│   │   ├── types.go             # Type definitions
│   │   ├── metadata.go          # On-chain metadata
│   │   ├── operations.go        # Client operations
│   │   └── ...
│   │
│   ├── credentials/              # Credential management
│   │   ├── service.go           # Business logic
│   │   ├── config.go            # Configuration
│   │   ├── issuer.go            # Issuer operations
│   │   ├── verifier.go          # Verifier operations
│   │   └── oidc4vc/             # OIDC4VC flows
│   │
│   ├── identity/                 # Identity management
│   │   ├── service.go           # Identity service
│   │   ├── handlers.go          # HTTP handlers
│   │   ├── cache/               # Caching layer
│   │   │   └── stubs.go         # Cache implementation
│   │   ├── trust/               # Trust level management
│   │   │   └── stubs.go         # Trust service
│   │   └── vc/                  # Verifiable credentials
│   │       └── stubs.go         # VC service
│   │
│   ├── did/                      # DID operations
│   │   ├── service.go           # DID service
│   │   ├── config.go            # Configuration
│   │   ├── handlers.go          # HTTP handlers
│   │   └── ...
│   │
│   ├── middleware/               # HTTP middleware
│   │   ├── auth.go              # Authentication
│   │   └── cors.go              # CORS handling
│   │
│   ├── config/                   # Configuration
│   │   └── config.go            # Config structures
│   │
│   └── utils/                    # Utilities
│       ├── response.go          # Response helpers
│       └── errors.go            # Error handling
│
├── go.mod                        # Module definition
├── go.sum                        # Dependency checksums
└── main.go                       # Primary entry point
```

## Core Services

### 1. Credentials Service (`cmd/credentials/main.go`)
**Purpose**: Manage verifiable credentials

**Key Components**:
- Issues and verifies credentials
- Implements OIDC4VC flows
- Manages credential storage
- Handles revocation

**Endpoints**:
- `POST /credentials` - Issue credential
- `GET /credentials/:id` - Retrieve credential
- `DELETE /credentials/:id` - Revoke credential
- `GET /health` - Health check
- `GET /ready` - Readiness probe

**Dependencies**:
- pkg/credentials
- pkg/credentials/oidc4vc

### 2. DID Service (`cmd/did/main.go`)
**Purpose**: Manage Decentralized Identifiers (DIDs)

**Key Components**:
- Creates and manages DIDs
- Resolves DID documents
- Manages public keys
- Handles DID updates

**Endpoints**:
- `POST /did` - Create DID
- `GET /did/:id` - Resolve DID
- `PUT /did/:id` - Update DID
- `GET /health` - Health check
- `GET /ready` - Readiness probe

**Dependencies**:
- pkg/did
- internal/crypto

### 3. Cardano Identity Service (`cmd/cardanoidentity/main.go`)
**Purpose**: Manage identity on Cardano blockchain

**Key Components**:
- Stores identities on blockchain
- Manages trust levels
- Tracks audit trails
- Handles credential storage

**Endpoints**:
- `POST /identities` - Create identity
- `GET /identities/:id` - Get identity
- `PUT /trust-level/:userID` - Update trust level
- `GET /credentials/user/:userID` - Get user credentials
- `GET /health` - Health check
- `GET /ready` - Readiness probe

**Dependencies**:
- pkg/cardano
- pkg/identity
- pkg/middleware

## Package Details

### pkg/cardano - Blockchain Integration

**Client** (`client.go`)
```go
type Client struct {
    baseURL            string
    timeout            time.Duration
    config             ClientConfig
    logger             *log.Logger
    networkConnected   bool
    credentialCache    *Cache
    trustLevelCache    *Cache
    auditTrailCache    *Cache
}
```

**Methods**:
- `Health(ctx context.Context)` - Check service health
- `GetCredential(ctx context.Context, credentialID string)` - Retrieve credential
- `GetUserCredentials(ctx context.Context, userID string)` - Get user's credentials
- `StoreCredential(ctx context.Context, ...)` - Store credential
- `GetTrustLevel(ctx context.Context, userID string)` - Get trust level
- `UpdateTrustLevel(ctx context.Context, ...)` - Update trust level
- `GetAuditTrail(ctx context.Context, userID string)` - Get audit trail

**Cache** (`client.go`)
```go
type Cache struct {
    mu      map[string]time.Time // expiration times
    data    map[string]interface{}
    ttl     time.Duration
}
```

**Operations** (`operations.go`)
- Credential queries with caching
- Credential verification
- Credential revocation
- Trust level history
- Cache invalidation

### pkg/identity - Identity Management

**Service** (`service.go`)
```go
type Service struct {
    cardanoClient        *cardano.Client
    trustLevelService    *trust.TrustLevelService
    vcService            *vc.VerifiableCredentialService
    identityCache        *cache.IdentityCache
    logger               *log.Logger
}
```

**Methods**:
- `GetTrustLevel(ctx context.Context, userID string)` - Retrieve trust level
- `UpdateTrustLevel(ctx context.Context, ...)` - Update trust level
- `StoreCredential(ctx context.Context, ...)` - Store credential
- `GetCredential(ctx context.Context, credentialID string)` - Retrieve credential
- `GetUserCredentials(ctx context.Context, userID string)` - Get user credentials
- `GetStorageHealth(ctx context.Context)` - Check storage health

**Cache** (`cache/stubs.go`)
```go
type IdentityCache struct {
    mu    sync.RWMutex
    cache map[string]*CacheEntry
    ttl   time.Duration
}
```

**Methods**:
- `Set(ctx context.Context, key string, value interface{})` - Cache value
- `Get(ctx context.Context, key string)` - Retrieve from cache
- `GetTrustLevel(userID string)` - Get trust level from cache
- `GetTrustLevelWithFallback(userID string, fallback func())` - Get with fallback
- `GetUserCredentials(userID string)` - Get user credentials from cache
- `GetUserCredentialsWithFallback(userID string, fallback func())` - Get with fallback
- `InvalidateTrustLevel(userID string)` - Invalidate cache entry
- `GetMetrics()` - Get cache metrics
- `Clear()` - Clear all cache entries

### pkg/middleware - HTTP Middleware

**CORS Middleware** (`cors.go`)
```go
type CORSPolicy struct {
    AllowedOrigins []string
    AllowedMethods []string
    AllowedHeaders []string
    ExposedHeaders []string
    MaxAge         int
}

func CORSMiddleware(policy *CORSPolicy) func(http.Handler) http.Handler
```

**Authentication Middleware** (`auth.go`)
```go
func AuthMiddleware(skipPaths []string) func(http.Handler) http.Handler
func LoggingMiddleware(next http.Handler) http.Handler
func RequestIDMiddleware(next http.Handler) http.Handler
```

### internal/crypto - Cryptographic Utilities

**Ed25519** (`ed25519.go`)
```go
type Ed25519KeyPair struct {
    PrivateKey ed25519.PrivateKey
    PublicKey  ed25519.PublicKey
}

type CryptoUtils struct{}

// Methods
func (kp *Ed25519KeyPair) Sign(message []byte) ([]byte, error)
func (kp *Ed25519KeyPair) Verify(message, signature []byte) bool
func (c *CryptoUtils) GenerateKey() (*Ed25519KeyPair, error)
func (c *CryptoUtils) SignMessage(privateKey ed25519.PrivateKey, message []byte) ([]byte, error)
func (c *CryptoUtils) VerifySignature(publicKey ed25519.PublicKey, message, signature []byte) bool
```

## Data Types

### Credential
```go
type Credential struct {
    ID             string                 // Unique ID
    CredentialID   string                 // Blockchain ID
    UserID         string                 // Owner
    SchemaID       string                 // Schema reference
    CredentialType string                 // Type
    Issuer         string                 // Issuer DID
    IssuedAt       time.Time             // Issue date
    ExpiresAt      *time.Time            // Expiration date
    ContentHash    string                 // Content hash
    TxHash         string                 // Blockchain tx
    Status         string                 // active/revoked/expired
    Data           map[string]interface{} // Credential data
    Metadata       map[string]interface{} // Additional metadata
    Timestamp      time.Time              // Record timestamp
}
```

### TrustLevel
```go
type TrustLevel struct {
    UserID    string    // User identifier
    Level     string    // Trust level
    TxHash    string    // Blockchain reference
    UpdatedAt time.Time // Last update
    UpdatedBy string    // Updated by
    Reason    string    // Reason for update
    Timestamp time.Time // Record timestamp
}
```

### AuditEntry
```go
type AuditEntry struct {
    ID            string    // Audit entry ID
    UserID        string    // User identifier
    Action        string    // Action performed
    Details       string    // Action details
    ActorID       string    // Who performed action
    PreviousLevel string    // Previous value
    NewLevel      string    // New value
    Reason        string    // Reason for action
    Timestamp     time.Time // When it happened
}
```

## Server Configuration

Each service uses a standard configuration pattern:

```go
type ServerConfig struct {
    Host             string        // Server host
    Port             int           // Server port
    TLSEnabled       bool          // TLS enabled
    TLSCertPath      string        // Certificate path
    TLSKeyPath       string        // Key path
    ReadTimeout      time.Duration // Read timeout
    WriteTimeout     time.Duration // Write timeout
    ShutdownTimeout  time.Duration // Graceful shutdown
}
```

## Error Handling

**ErrorResponse** (`pkg/utils/errors.go`)
```go
type ErrorResponse struct {
    Error ErrorDetails
    Code  string
    Message string
    RequestID string
    Timestamp string
    StatusCode int
}
```

**Common Error Codes**:
- `INVALID_REQUEST` - 400 Bad Request
- `UNAUTHORIZED` - 401 Unauthorized
- `FORBIDDEN` - 403 Forbidden
- `NOT_FOUND` - 404 Not Found
- `CONFLICT` - 409 Conflict
- `INTERNAL_ERROR` - 500 Internal Server Error
- `SERVICE_UNAVAILABLE` - 503 Service Unavailable

## Request/Response Pattern

All HTTP handlers follow this pattern:

```go
func (h *Handlers) handleRequest(c *gin.Context) {
    // 1. Extract request ID
    requestID := getOrCreateRequestID(c)
    
    // 2. Parse request
    var req RequestStruct
    if err := c.BindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    // 3. Business logic
    result, err := h.service.DoSomething(c.Request.Context(), ...)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    // 4. Return response
    c.JSON(http.StatusOK, ResponseStruct{
        Data:      result,
        RequestID: requestID,
        Timestamp: time.Now(),
    })
}
```

## Deployment Considerations

1. **TLS Configuration**: All services support TLS 1.3+
2. **Health Checks**: `/health` and `/ready` endpoints
3. **Graceful Shutdown**: Signal handling with configurable timeout
4. **Logging**: Structured logging with request IDs
5. **Request Tracing**: X-Request-ID header tracking
6. **CORS Policy**: Configurable origin validation
7. **Rate Limiting**: Ready for middleware integration
8. **Caching**: TTL-based in-memory cache

## Testing Strategy

1. **Unit Tests**: Test individual packages
2. **Integration Tests**: Test service interactions
3. **API Tests**: Validate HTTP endpoints
4. **Crypto Tests**: Verify cryptographic operations
5. **Cache Tests**: Validate cache behavior
6. **Load Tests**: Performance testing

## Future Enhancements

1. Database persistence (currently in-memory cache)
2. Distributed caching (Redis integration)
3. Message queues (async operations)
4. GraphQL API layer
5. Swagger/OpenAPI documentation
6. Metrics and monitoring (Prometheus)
7. Distributed tracing (Jaeger)
8. Rate limiting (Token bucket)
9. Database migrations
10. Configuration management (Viper)
