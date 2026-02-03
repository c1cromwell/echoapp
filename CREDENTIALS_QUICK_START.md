# Verifiable Credentials - Quick Start Guide

Fast-track guide to using the W3C Verifiable Credentials system with OIDC4VC support.

## 5-Minute Setup

### 1. Start the Service

```bash
cd /Users/thechadcromwell/Projects/echoapp
export ISSUER_DID=did:prism:cardano:issuer123
export ISSUER_PRIVATE_KEY_PATH=/path/to/private/key.pem
export VERIFIER_DID=did:prism:cardano:verifier456
export CARDANO_NETWORK_ID=testnet
export CARDANO_NODE_URL=http://localhost:8000

go run ./cmd/credentials/main.go
```

### 2. Check Health

```bash
curl http://localhost:8080/health
```

### 3. Issue a Credential

```bash
curl -X POST http://localhost:8080/api/v1/credentials \
  -H "Content-Type: application/json" \
  -d '{
    "subjectDid": "did:prism:cardano:alice",
    "credentialType": "ProofOfHumanity",
    "claims": {"email": "alice@example.com"},
    "verificationClaims": [
      {
        "type": "email_verification",
        "value": "verified",
        "verificationLevel": "high"
      }
    ],
    "preferredFormat": "json-ld"
  }'
```

Response:
```json
{
  "credentialId": "uuid-1234",
  "verifiableCredential": "{...credential...}",
  "format": "json-ld",
  "issuedAt": "2024-01-15T10:30:00Z",
  "expiresAt": "2025-01-15T10:30:00Z",
  "status": "issued"
}
```

### 4. Verify the Credential

```bash
curl -X POST http://localhost:8080/api/v1/credentials/verify \
  -H "Content-Type: application/json" \
  -d '{
    "credential": "{credential-json-from-above}",
    "format": "json-ld",
    "issuerDid": "did:prism:cardano:issuer123"
  }'
```

Response:
```json
{
  "credentialId": "uuid-1234",
  "isValid": true,
  "verifiedAt": "2024-01-15T10:31:00Z",
  "issuer": "did:prism:cardano:issuer123",
  "subject": "did:prism:cardano:alice",
  "credentialType": "ProofOfHumanity",
  "signatureValid": true,
  "notExpired": true,
  "notRevoked": true,
  "revocationStatus": "active",
  "errors": []
}
```

## Code Examples

### Go Integration

```go
package main

import (
	"context"
	"github.com/thechadcromwell/echoapp/pkg/credentials"
)

func main() {
	// Create service
	config := credentials.LoadConfig()
	service, _ := credentials.NewService(config)
	defer service.Close()

	ctx := context.Background()

	// Issue credential
	req := &credentials.CredentialIssuanceRequest{
		SubjectDID: "did:prism:cardano:user",
		CredentialType: credentials.ProofOfHumanity,
		Claims: map[string]interface{}{
			"email": "user@example.com",
		},
		PreferredFormat: credentials.JSONLDFormat,
	}

	resp, err := service.IssueCredential(ctx, req)
	if err != nil {
		panic(err)
	}

	println("Credential issued:", resp.CredentialID)

	// Verify credential
	verifyReq := &credentials.CredentialVerificationRequest{
		Credential: resp.VerifiableCredential,
		Format: credentials.JSONLDFormat,
		IssuerDID: config.IssuerConfig.IssuerDID,
	}

	result, _ := service.VerifyCredential(ctx, verifyReq)
	println("Valid:", result.IsValid)
}
```

### Batch Verification

```go
requests := []credentials.CredentialVerificationRequest{
	{Credential: cred1, Format: credentials.JSONLDFormat},
	{Credential: cred2, Format: credentials.JWTFormat},
	{Credential: cred3, Format: credentials.SDJWTFormat},
}

results, err := service.verifier.BatchVerify(ctx, requests)
for i, result := range results {
	println(fmt.Sprintf("Credential %d: %v", i, result.IsValid))
}
```

### Monitor Issuance Progress

```go
credentialID := "uuid-1234"

for {
	progress := service.GetIssuanceProgress(credentialID)
	if progress == nil {
		println("Issuance complete")
		break
	}

	println(fmt.Sprintf("Progress: %d%% - %s", progress.Progress, progress.CurrentStep))
	time.Sleep(1 * time.Second)
}
```

## OIDC4VC Flows

### Authorization Code Flow

**Step 1: Authorization Request**

```bash
curl -X GET http://localhost:8080/oauth/authorization \
  ?client_id=wallet \
  &redirect_uri=http://wallet:3000/callback \
  &response_type=code \
  &scope=openid \
  &state=random123 \
  &code_challenge=E9Mrozoa2owUednJR2Qjj9I2VkJN9Ps4O1RR3nMpMms \
  &code_challenge_method=S256
```

**Step 2: Token Request**

```bash
curl -X POST http://localhost:8080/oauth/token \
  -H "Content-Type: application/json" \
  -d '{
    "grant_type": "authorization_code",
    "code": "auth-code-from-step1",
    "client_id": "wallet",
    "redirect_uri": "http://wallet:3000/callback",
    "code_verifier": "E9Mrozoa2owUednJR2QjjE2VkJN9Ps4O1RR3nMpMms"
  }'
```

Response:
```json
{
  "access_token": "eyJ0eXA...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "c_nonce": "nonce123",
  "c_nonce_expires_in": 300
}
```

**Step 3: Credential Request**

```bash
curl -X POST http://localhost:8080/credential \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJ0eXA..." \
  -d '{
    "format": "json-ld+jwt",
    "credential_type": ["ProofOfHumanity"],
    "proof": {
      "proof_type": "jwt",
      "jwt": "eyJ0eXA..."
    }
  }'
```

### Pre-Authorized Code Flow

**Step 1: Get Pre-Authorized Code** (via QR or deep link)

```
openid-credential://?issuer=http://localhost:8080&pre-authorized_code=code123
```

**Step 2: Token Request**

```bash
curl -X POST http://localhost:8080/oauth/token \
  -H "Content-Type: application/json" \
  -d '{
    "grant_type": "urn:ietf:params:oauth:grant-type:pre-authorized_code",
    "pre-authorized_code": "code123"
  }'
```

**Step 3: Credential Request** (same as auth code flow)

### Presentation Flow (Verification)

**Step 1: Verifier Creates Request**

```bash
curl -X GET http://localhost:8080/verification/request \
  ?credential_type=ProofOfHumanity \
  &client_id=verifier \
  &redirect_uri=http://verifier:3000/result \
  &state=state123
```

**Step 2: Wallet Submits Presentation**

```bash
curl -X POST http://localhost:8080/verification/submit \
  -H "Content-Type: application/json" \
  -d '{
    "vp_token": "eyJ0eXA...",
    "presentation_submission": {
      "id": "sub123",
      "definition_id": "def123",
      "descriptor_map": [
        {
          "id": "input_id_1",
          "format": "json-ld",
          "path": "$"
        }
      ]
    },
    "state": "state123"
  }'
```

**Step 3: Check Verification Result**

```bash
curl http://localhost:8080/verification/pres_state123/status
```

## Credential Types

### Proof of Humanity

```json
{
  "type": ["VerifiableCredential", "ProofOfHumanity"],
  "credentialSubject": {
    "id": "did:prism:cardano:user",
    "claims": {
      "humanityProof": true
    }
  }
}
```

### KYC-Lite

```json
{
  "type": ["VerifiableCredential", "KYCLite"],
  "credentialSubject": {
    "id": "did:prism:cardano:user",
    "claims": {
      "firstName": "Alice",
      "lastName": "Smith",
      "verificationLevel": "intermediate"
    }
  }
}
```

### High-Assurance

```json
{
  "type": ["VerifiableCredential", "HighAssurance"],
  "credentialSubject": {
    "id": "did:prism:cardano:user",
    "claims": {
      "firstName": "Alice",
      "lastName": "Smith",
      "dateOfBirth": "1990-01-15",
      "verificationLevel": "high"
    }
  }
}
```

### Professional

```json
{
  "type": ["VerifiableCredential", "Professional"],
  "credentialSubject": {
    "id": "did:prism:cardano:user",
    "claims": {
      "profession": "Software Engineer",
      "employer": "Tech Company Inc",
      "credentials": ["AWS Certified", "Kubernetes CKA"]
    }
  }
}
```

## Common Tasks

### Revoke a Credential

```bash
curl -X POST http://localhost:8080/api/v1/credentials/cred-id/revoke \
  -H "Content-Type: application/json" \
  -d '{
    "issuerDid": "did:prism:cardano:issuer",
    "subjectDid": "did:prism:cardano:user",
    "reason": "User requested revocation"
  }'
```

### Check Revocation Status

```bash
curl http://localhost:8080/api/v1/revocation/status/cred-id
```

### Batch Check Revocation

```bash
curl -X POST http://localhost:8080/api/v1/revocation/batch-check \
  -H "Content-Type: application/json" \
  -d '{
    "credentialIds": ["cred1", "cred2", "cred3"]
  }'
```

### Convert Credential Format

```bash
curl -X POST http://localhost:8080/api/v1/credentials/cred-id/convert \
  -H "Content-Type: application/json" \
  -d '{
    "format": "jwt",
    "privateKey": "base64-encoded-key"
  }'
```

### Get Trust Score

```bash
curl http://localhost:8080/api/v1/credentials/cred-id/trust-score
```

### Get Issuance Progress

```bash
curl http://localhost:8080/api/v1/credentials/cred-id/progress
```

## Configuration Examples

### Development

```bash
export ISSUER_DID=did:prism:cardano:dev-issuer
export CARDANO_NETWORK_ID=testnet
export LOG_LEVEL=debug
export CRED_ISSUANCE_TIMEOUT_SECONDS=120
export OIDC4VC_ENABLE_PKCE=false
```

### Production

```bash
export ISSUER_DID=did:prism:cardano:prod-issuer
export ISSUER_PRIVATE_KEY_PATH=/secure/path/to/key.pem
export CARDANO_NETWORK_ID=mainnet
export CARDANO_API_KEY=production-key
export CARDANO_API_SECRET=production-secret
export LOG_LEVEL=info
export SERVER_TLS_ENABLED=true
export SERVER_CERT_FILE=/etc/tls/cert.pem
export SERVER_KEY_FILE=/etc/tls/key.pem
export OIDC4VC_ENABLE_PKCE=true
```

## Performance Benchmarks

### Issuance

```bash
# Single credential
Time: 5-10 seconds (mostly blockchain anchoring)

# Batch issue 10 credentials
Time: 50-100 seconds (parallel with 10 concurrent)
```

### Verification

```bash
# Cached (same credential verified again)
Time: 50-100ms

# Fresh (cache miss, blockchain check)
Time: 3-5 seconds

# Batch verify 100 credentials
Time: 5-10 seconds (with 20 concurrent)
```

### Revocation Check

```bash
# Cached
Time: < 50ms

# Fresh blockchain check
Time: 1-3 seconds

# Batch check 50 credentials
Time: 100-200ms (with intelligent caching)
```

## Troubleshooting

### "Invalid signature" Error

**Cause**: Issuer's private key doesn't match issuer DID

**Solution**:
```bash
# Ensure ISSUER_PRIVATE_KEY_PATH points to correct key
# Verify issuer DID matches public key
openssl pkey -in /path/to/key.pem -text -noout | grep "pub"
```

### "Timeout" During Issuance

**Cause**: Cardano node not responsive

**Solution**:
```bash
# Check Cardano node
curl http://localhost:8000/health

# Increase timeout
export CRED_ISSUANCE_TIMEOUT_SECONDS=120

# Check blockchain status
cardano-cli query tip --testnet-magic 1097911063
```

### "Credential Not Found"

**Cause**: Storage not configured or credential ID wrong

**Solution**:
```bash
# Check storage path exists
mkdir -p $(echo $CRED_STORAGE_PATH)

# List issued credentials
curl "http://localhost:8080/api/v1/credentials/subject/did:prism:cardano:user"
```

## Next Steps

1. **Integrate with DID Service**: Link credentials with DID system
2. **Add Database Support**: Switch from in-memory to PostgreSQL
3. **Implement Wallet**: Build mobile/web wallet for credentials
4. **Setup Monitoring**: Add Prometheus metrics and alerts
5. **Deploy**: Use Docker and Kubernetes manifests
6. **Testing**: Run integration and load tests

See CREDENTIALS_README.md for detailed documentation.
