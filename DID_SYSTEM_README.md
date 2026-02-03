# DID Management System - Atala PRISM Integration

A comprehensive, production-ready Go-based DID (Decentralized Identifier) management system implementing Atala PRISM infrastructure for self-sovereign identity capabilities. This system enables users to maintain complete control over their identity data while establishing verifiable credentials on the Cardano blockchain.

## Features

### Core DID Management
- **DID Generation**: Unique DIDs following format `did:prism:cardano:<unique-identifier>`
- **Blockchain Anchoring**: DID documents anchored to Cardano blockchain via Atala PRISM
- **DID Resolution**: Query and retrieve authoritative DID documents from blockchain
- **DID Updates**: Modify DID documents with full audit trail

### Performance & Efficiency
- **Fast Generation**: Complete DID generation and anchoring within 30 seconds
- **Quick Resolution**: DID resolution within 2 seconds using cached data
- **24-Hour Caching**: Local caching with automatic expiration and invalidation
- **Connection Pooling**: HTTP connection pooling for optimal throughput

### Multi-Device Support
- **Secondary Device Registration**: Add multiple devices with separate public keys
- **QR Code Registration**: Secure device registration flow via QR codes
- **Secure Enclave Integration**: Support for iOS Secure Enclave storage
- **Device Management**: Activate/deactivate devices without DID regeneration

### Security & Compliance
- **Ed25519 Cryptography**: High-performance elliptic curve cryptography
- **W3C Compliance**: DID documents follow W3C DID specification 1.0
- **KERI Standards**: Key Event Receipt Infrastructure for scalability
- **Concurrent Access**: Thread-safe operations using sync.RWMutex and sync.Map

## Architecture

### Package Structure

```
pkg/did/
├── models.go           # W3C DID document structures
├── errors.go           # Custom error types and handling
├── config.go           # Configuration management
├── cache.go            # High-performance caching with TTL
├── repository.go       # DID-to-account mapping storage
├── atala_client.go     # Atala PRISM API client
├── resolver.go         # DID resolution with caching
├── multidevice.go      # Multi-device registration
├── service.go          # Core DID service logic
└── handlers.go         # HTTP endpoints (Gin framework)

internal/crypto/
└── ed25519.go          # Cryptographic utilities

cmd/did/
└── main.go             # Application entry point
```

### Component Interactions

```
HTTP Request
    ↓
Handlers (handlers.go)
    ↓
Service (service.go)
    ├─→ Resolver (resolver.go)
    │   ├─→ Cache (cache.go)
    │   └─→ AtalaClient (atala_client.go)
    │
    ├─→ DeviceManager (multidevice.go)
    │   └─→ Repository (repository.go)
    │
    └─→ CryptoUtils (internal/crypto/ed25519.go)
```

## API Endpoints

### DID Operations
- `POST /v1/dids` - Create new DID
- `GET /v1/dids/:did` - Resolve DID document
- `PUT /v1/dids/:did` - Update DID document
- `GET /v1/dids/:did/mapping` - Get DID mapping info

### Device Management
- `GET /v1/dids/:did/devices` - List all devices
- `POST /v1/dids/:did/devices` - Register new device
- `DELETE /v1/dids/:did/devices/:deviceId` - Unregister device

### Device Registration Flow
- `POST /v1/dids/:did/devices/register/initiate` - Start device registration
- `POST /v1/dids/:did/devices/register/qrcode` - Generate QR code
- `POST /v1/devices/register/complete` - Complete registration

### Verification & Health
- `POST /v1/dids/verify` - Verify DID document
- `GET /v1/health` - Health check
- `GET /v1/ready` - Readiness probe

### Cache Management
- `POST /v1/cache/invalidate/:did` - Invalidate DID cache
- `POST /v1/cache/clear` - Clear all cache
- `GET /v1/cache/stats` - Get cache statistics

## Configuration

### Environment Variables

```bash
# Atala PRISM
DID_ATALA_PRISM_ENDPOINT=https://prism.atalaprism.io
DID_ATALA_PRISM_API_KEY=your-api-key
DID_ATALA_PRISM_API_SECRET=your-api-secret

# Cardano Network
DID_CARDANO_NETWORK_ID=testnet
DID_CARDANO_NODE_URL=https://cardano-testnet-node.example.com

# Server
DID_SERVER_PORT=8080
DID_SERVER_HOST=0.0.0.0

# Logging
DID_LOGGING_LEVEL=info
DID_LOGGING_FORMAT=json
```

### Configuration Structure

```go
Config{
    AtalaPRISM: {
        Endpoint: "https://prism.atalaprism.io",
        APIKey: "...",
        APISecret: "...",
        Timeout: 30 * time.Second,
        MaxRetries: 3,
        ConnectionPool: 10,
    },
    Cardano: {
        NetworkID: "testnet",
        NodeURL: "...",
        ConfirmationThreshold: 6,
    },
    DID: {
        GenerationTimeout: 30 * time.Second,
        ResolutionTimeout: 2 * time.Second,
        SupportedKeyTypes: []string{"Ed25519VerificationKey2018"},
    },
    Cache: {
        Enabled: true,
        TTL: 24 * time.Hour,
        MaxSize: 10000,
        CleanupInterval: 1 * time.Hour,
    },
}
```

## Data Models

### DID Document (W3C Compliant)

```go
DIDDocument{
    ID: "did:prism:cardano:uuid",
    Context: [
        "https://www.w3.org/ns/did/v1",
        "https://w3id.org/security/suites/ed25519-2018/v1",
    ],
    PublicKey: [{
        ID: "did:prism:cardano:uuid#key-1",
        Type: "Ed25519VerificationKey2018",
        Controller: "did:prism:cardano:uuid",
        PublicKeyBase64: "...",
    }],
    Authentication: [{ ... }],
    AssertionMethod: [{ ... }],
    Service: [{
        ID: "did:prism:cardano:uuid#inbox",
        Type: "DIDCommMessaging",
        ServiceEndpoint: "https://...",
    }],
    Created: time.Time,
    Updated: time.Time,
}
```

### DID Mapping

```go
DIDMapping{
    DID: "did:prism:cardano:uuid",
    UserID: "user-id",
    AccountID: "account-id",
    CreatedAt: time.Time,
    UpdatedAt: time.Time,
    IsActive: true,
    PrimaryDevice: "device-id",
    Devices: []{
        DeviceID: "device-id",
        DeviceName: "iPhone 15",
        PublicKey: "...",
        CreatedAt: time.Time,
        IsSecureEnclave: true,
    },
}
```

## Key Algorithms & Standards

### Cryptography
- **Key Type**: Ed25519 (Edwards-curve Digital Signature Algorithm)
- **Key Size**: 32 bytes public key, 64 bytes private key
- **Encoding**: Base64 and Multibase formats supported

### DID Format
- **Method**: `prism`
- **Network**: `cardano`
- **Identifier**: UUIDv4

### Blockchain
- **Network**: Cardano (Mainnet/Testnet)
- **Confirmation Threshold**: 6 blocks
- **Transaction Type**: Smart contract interaction

## Performance Characteristics

| Operation | Target | Typical | Notes |
|-----------|--------|---------|-------|
| DID Generation | < 30s | 5-10s | Includes anchoring |
| DID Resolution | < 2s | 200-500ms | With cache hit |
| Device Registration | < 5s | 2-3s | QR code generation |
| Cache Lookup | < 10ms | < 5ms | In-memory access |
| Concurrent Requests | 100+ | Tested | Connection pooling |

## Error Handling

The system provides detailed error handling with specific error codes:

```go
ErrCodeInvalidDID           // Invalid DID format
ErrCodeDIDNotFound          // DID does not exist
ErrCodeDIDAlreadyExists     // DID already registered
ErrCodeGenerationFailed     // DID generation failed
ErrCodeAnchoringFailed      // Blockchain anchoring failed
ErrCodeResolutionFailed     // DID resolution failed
ErrCodeAtalaPRISMError      // Atala PRISM API error
ErrCodeBlockchainError      // Blockchain operation error
ErrCodeTimeout              // Operation timeout
```

## Security Considerations

1. **Private Key Storage**: Private keys should be stored in Secure Enclave or HSM
2. **API Key Management**: Use environment variables and secrets management
3. **HTTPS Only**: All external API calls use HTTPS
4. **Rate Limiting**: Implement rate limiting for API endpoints
5. **Audit Logging**: All DID operations are logged
6. **CORS Configuration**: Restrict cross-origin requests

## Deployment

### Docker

```dockerfile
FROM golang:1.21-alpine
WORKDIR /app
COPY . .
RUN go build -o did-service ./cmd/did
EXPOSE 8080
CMD ["./did-service"]
```

### Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: did-service
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: did-service
        image: did-service:latest
        ports:
        - containerPort: 8080
        env:
        - name: DID_ATALA_PRISM_ENDPOINT
          valueFrom:
            secretKeyRef:
              name: did-config
              key: atala-endpoint
```

## Development

### Prerequisites
- Go 1.21+
- Atala PRISM API credentials
- Cardano node access (testnet)

### Building

```bash
# Build the application
go build -o did-service ./cmd/did

# Run tests
go test ./...

# Generate documentation
godoc -http :6060
```

### Running Locally

```bash
# Set configuration
export DID_ATALA_PRISM_ENDPOINT=https://prism.atalaprism.io
export DID_CARDANO_NETWORK_ID=testnet

# Run server
go run ./cmd/did/main.go
```

## Testing

```bash
# Unit tests
go test ./pkg/did ./internal/crypto -v

# Integration tests
go test ./tests/integration -v

# Benchmark performance
go test ./pkg/did -bench=. -benchmem
```

## Monitoring & Logging

### Metrics
- DID generation time
- Resolution cache hit rate
- Atala PRISM API latency
- Blockchain confirmation time
- Device registration success rate

### Logging
- All DID operations logged at INFO level
- Errors logged at ERROR level
- Debug information at DEBUG level
- Structured JSON logging format

### Health Checks
- `/health` - Basic health status
- `/ready` - Readiness for traffic
- Service dependency checks

## Performance Tuning

### Cache Optimization
- Adjust TTL based on DID update frequency
- Monitor cache hit rate
- Tune MaxSize for memory constraints
- Adjust CleanupInterval for frequency

### Connection Pooling
- MaxIdleConnsPerHost: 10 (tunable)
- IdleConnTimeout: 90 seconds
- MaxConnections: 25 (database)

### Timeout Configuration
- Generation: 30 seconds
- Resolution: 2 seconds  
- Blockchain: 60 seconds
- API calls: 30 seconds

## License

[Your License Here]

## Support

For issues, questions, or contributions, please refer to the project repository.

## References

- [W3C DID Specification](https://www.w3.org/TR/did-core/)
- [Atala PRISM Documentation](https://atalaprism.io/)
- [Cardano Documentation](https://docs.cardano.org/)
- [KERI Standards](https://github.com/WebOfTrust/ietf-keri)
