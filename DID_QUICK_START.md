# DID System - Quick Start Guide

## Quick Implementation Examples

### 1. Initialize the DID Service

```go
import (
	"github.com/thechadcromwell/echoapp/pkg/did"
	"github.com/thechadcromwell/echoapp/internal/crypto"
)

// Load configuration
config, _ := did.LoadConfig("")

// Initialize components
atalaClient := did.NewAtalaClient(&config.AtalaPRISM)
cache := did.NewCache(&config.Cache)
repo := did.NewInMemoryRepository()
resolver := did.NewResolver(atalaClient, cache, repo, &config.DID)
deviceManager := did.NewDeviceManager(repo, cache, &config.DID)
cryptoUtils := crypto.NewCryptoUtils()

// Initialize main service
service := did.NewService(
	atalaClient,
	resolver,
	deviceManager,
	repo,
	cache,
	&config.DID,
	cryptoUtils,
)
```

### 2. Create a New DID

```go
ctx := context.Background()

req := &did.DIDCreationRequest{
	UserID:   "user-123",
	DeviceID: "device-456",
	PublicKey: "base64-encoded-public-key", // or leave empty for auto-generation
	DeviceName: "iPhone 15",
}

resp, err := service.CreateDID(ctx, req)
if err != nil {
	log.Printf("Error: %v", err)
	return
}

fmt.Printf("DID Created: %s\n", resp.DID)
fmt.Printf("Transaction Hash: %s\n", resp.TransactionHash)
```

### 3. Resolve a DID

```go
did := "did:prism:cardano:uuid"

// Simple resolution (uses cache)
document, err := service.ResolveDID(ctx, did)
if err != nil {
	log.Printf("Error: %v", err)
	return
}

// With metadata
document, metadata, err := service.ResolveDIDWithMetadata(ctx, did)
if err != nil {
	log.Printf("Error: %v", err)
	return
}

fmt.Printf("Cache Valid: %v\n", metadata.CacheValid)
fmt.Printf("Resolution Time: %v\n", time.Since(metadata.ResolutionTimestamp))
```

### 4. Register a New Device

```go
did := "did:prism:cardano:uuid"

// Initiate device registration
pending, err := service.InitiateDeviceRegistration(did)
if err != nil {
	log.Printf("Error: %v", err)
	return
}

// Generate QR code
qrData, qrCode, err := service.GenerateQRCodeForDeviceRegistration(did)
if err != nil {
	log.Printf("Error: %v", err)
	return
}

// Client scans QR code, sends challenge and public key back
device, err := service.CompleteDeviceRegistration(
	pending.DeviceID,
	pending.Challenge,
	"public-key-base64",
	"iPhone 15 Pro",
)
if err != nil {
	log.Printf("Error: %v", err)
	return
}

fmt.Printf("Device Registered: %s\n", device.DeviceID)
```

### 5. Generate Keys

```go
crypto := crypto.NewCryptoUtils()

// Generate Ed25519 key pair
publicKey, privateKey, err := crypto.GenerateKeyPair()
if err != nil {
	log.Printf("Error: %v", err)
	return
}

fmt.Printf("Public Key: %s\n", publicKey)
fmt.Printf("Private Key: %s\n", privateKey)

// Sign a message
message := []byte("Hello, World!")
signature, err := crypto.SignMessage(message, privateKey)
if err != nil {
	log.Printf("Error: %v", err)
	return
}

// Verify signature
valid, err := crypto.VerifySignature(message, signature, publicKey)
if err != nil {
	log.Printf("Error: %v", err)
	return
}

fmt.Printf("Signature Valid: %v\n", valid)
```

## REST API Examples

### Create DID

```bash
curl -X POST http://localhost:8080/v1/dids \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "user-123",
    "deviceId": "device-456",
    "deviceName": "iPhone 15",
    "publicKey": "AAABBB..." 
  }'
```

### Resolve DID

```bash
# Without metadata
curl http://localhost:8080/v1/dids/did:prism:cardano:uuid

# With metadata
curl "http://localhost:8080/v1/dids/did:prism:cardano:uuid?metadata=true"
```

### Register Device

```bash
curl -X POST http://localhost:8080/v1/dids/did:prism:cardano:uuid/devices/register/initiate

# Get QR Code
curl -X POST http://localhost:8080/v1/dids/did:prism:cardano:uuid/devices/register/qrcode

# Complete Registration
curl -X POST http://localhost:8080/v1/devices/register/complete \
  -H "Content-Type: application/json" \
  -d '{
    "deviceId": "device-123",
    "challenge": "challenge-string",
    "publicKey": "AAABBB...",
    "deviceName": "iPhone 15"
  }'
```

### List Devices

```bash
curl http://localhost:8080/v1/dids/did:prism:cardano:uuid/devices
```

### Cache Management

```bash
# Invalidate cache for DID
curl -X POST http://localhost:8080/v1/cache/invalidate/did:prism:cardano:uuid

# Clear all cache
curl -X POST http://localhost:8080/v1/cache/clear

# Get cache stats
curl http://localhost:8080/v1/cache/stats
```

## Error Handling

```go
import "errors"

_, err := service.ResolveDID(ctx, did)
if err != nil {
	// Check for specific error
	if did.IsDIDError(err, did.ErrCodeDIDNotFound) {
		log.Println("DID not found")
	} else if did.IsDIDError(err, did.ErrCodeTimeout) {
		log.Println("Request timed out")
	} else {
		log.Printf("Error: %v", err)
	}
}
```

## Configuration Examples

### Development

```bash
export DID_ATALA_PRISM_ENDPOINT=https://prism-testnet.atalaprism.io
export DID_CARDANO_NETWORK_ID=testnet
export DID_LOGGING_LEVEL=debug
export DID_SERVER_PORT=8080
```

### Production

```bash
export DID_ATALA_PRISM_ENDPOINT=https://prism.atalaprism.io
export DID_ATALA_PRISM_API_KEY=production-key
export DID_ATALA_PRISM_API_SECRET=production-secret
export DID_CARDANO_NETWORK_ID=mainnet
export DID_LOGGING_LEVEL=info
export DID_SERVER_HOST=0.0.0.0
export DID_SERVER_PORT=8080
export DID_SERVER_TLS_ENABLED=true
export DID_SERVER_CERT_FILE=/etc/tls/cert.pem
export DID_SERVER_KEY_FILE=/etc/tls/key.pem
```

## Common Patterns

### Getting DID by User ID

```go
mapping, err := service.GetDIDMappingByUserID(ctx, "user-123")
if err != nil {
	log.Printf("Error: %v", err)
	return
}

fmt.Printf("User's DID: %s\n", mapping.DID)
```

### Cache Invalidation on Update

```go
// Update DID document
err := service.UpdateDID(ctx, did, updatedDocument)
if err != nil {
	log.Printf("Error: %v", err)
	return
}

// Cache is automatically invalidated by service.UpdateDID()
```

### Health Checks

```go
healthy, err := service.Health(ctx)
if !healthy || err != nil {
	log.Printf("Service unhealthy: %v", err)
	return
}
```

### Concurrent DID Resolution

```go
dids := []string{
	"did:prism:cardano:uuid1",
	"did:prism:cardano:uuid2",
	"did:prism:cardano:uuid3",
}

results, errors := resolver.ResolveMultiple(ctx, dids)

for did, doc := range results {
	fmt.Printf("Resolved %s\n", did)
}

for did, err := range errors {
	fmt.Printf("Failed to resolve %s: %v\n", did, err)
}
```

## File Structure

```
echoapp/
├── cmd/
│   └── did/
│       └── main.go                  # Application entry point
├── pkg/
│   └── did/
│       ├── models.go                # Data structures
│       ├── errors.go                # Error handling
│       ├── config.go                # Configuration
│       ├── cache.go                 # Caching layer
│       ├── repository.go            # Data storage
│       ├── atala_client.go          # Atala PRISM client
│       ├── resolver.go              # DID resolution
│       ├── multidevice.go           # Device registration
│       ├── service.go               # Core business logic
│       └── handlers.go              # HTTP handlers
├── internal/
│   └── crypto/
│       └── ed25519.go               # Cryptographic utilities
├── go.mod                           # Go module definition
└── DID_SYSTEM_README.md            # Full documentation
```

## Troubleshooting

### "DID not found"
- Check that the DID format is correct
- Verify Cardano blockchain connectivity
- Check cache expiration

### "Atala PRISM connection failed"
- Verify API endpoint and credentials
- Check network connectivity
- Review API rate limits

### "Timeout"
- Increase timeout configuration
- Check Cardano blockchain status
- Reduce concurrent requests

### "Cache full"
- Increase MaxSize in cache config
- Reduce cache TTL
- Monitor cache hit rate

## Performance Tips

1. **Enable Caching**: TTL of 24 hours recommended
2. **Use Connection Pooling**: Default pool size of 10
3. **Batch Operations**: Resolve multiple DIDs concurrently
4. **Monitor Metrics**: Track cache hit rate and latency
5. **Load Testing**: Test with expected concurrent users

## Security Checklist

- [ ] All API calls use HTTPS
- [ ] API keys stored in environment variables
- [ ] Private keys stored in Secure Enclave
- [ ] CORS configured for your domain
- [ ] Rate limiting enabled
- [ ] Audit logging enabled
- [ ] TLS certificates valid
- [ ] API key rotation schedule
