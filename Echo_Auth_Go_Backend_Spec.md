# Echo Authentication — Go Backend Implementation Spec

> **Version:** 2.0 · **Date:** March 2026 · **Status:** Ready for Development

---

## Table of Contents

1. [Overview](#1-overview)
2. [Project Structure](#2-project-structure)
3. [API Endpoints](#3-api-endpoints)
4. [Registration Flow](#4-registration-flow)
5. [Login Flow](#5-login-flow)
6. [Token Architecture](#6-token-architecture)
7. [Token Refresh with Rotation](#7-token-refresh-with-rotation)
8. [Step-Up Authentication](#8-step-up-authentication)
9. [Rate Limiting Strategy](#9-rate-limiting-strategy)
10. [Device Fingerprinting](#10-device-fingerprinting)
11. [Account Recovery](#11-account-recovery)
12. [Error Code Catalog](#12-error-code-catalog)
13. [Database Schema](#13-database-schema)
14. [Security Middleware](#14-security-middleware)
15. [Testing Requirements](#15-testing-requirements)

---

## 1. Overview

The Echo Go backend handles all server-side authentication logic: OTP verification, WebAuthn passkey validation, JWT issuance and rotation, DID creation on Cardano, device management, and rate limiting. The design principle is **simple yet secure** — phone + OTP for registration, biometric passkey for returning login, with progressive verification layered on top.

### Key Dependencies

| Dependency | Purpose | Version |
|---|---|---|
| `go-webauthn/webauthn` | WebAuthn/FIDO2 passkey support | v0.10+ |
| `golang-jwt/jwt/v5` | JWT creation and validation (ES256) | v5.2+ |
| `go-redis/redis/v9` | Rate limiting, OTP store, token blocklist | v9.5+ |
| `jackc/pgx/v5` | PostgreSQL driver | v5.5+ |
| `twilio/twilio-go` | SMS OTP delivery | Latest |
| Atala PRISM SDK (gRPC) | DID creation and resolution on Cardano | Latest |

---

## 2. Project Structure

```
echo-backend/
├── cmd/
│   └── server/
│       └── main.go                    # Entry point, wire dependencies
│
├── internal/
│   ├── auth/
│   │   ├── handler.go                 # HTTP handlers for all /auth/* routes
│   │   ├── service.go                 # Core business logic (AuthService interface)
│   │   ├── middleware.go              # JWT validation, device binding middleware
│   │   ├── otp.go                     # OTP generation, hashing, verification
│   │   ├── passkey.go                 # WebAuthn registration & assertion
│   │   ├── token.go                   # JWT creation, refresh token management
│   │   ├── device.go                  # Device fingerprinting & validation
│   │   ├── recovery.go               # Account recovery flows
│   │   ├── stepup.go                  # Step-up authentication logic
│   │   └── ratelimit.go              # Rate limiting middleware (Redis)
│   │
│   ├── identity/
│   │   ├── did.go                     # DID creation & resolution (Atala PRISM)
│   │   ├── verification.go           # Third-party verification (Prove, Daon, Alloy)
│   │   └── credential.go             # Verifiable credential issuance
│   │
│   ├── store/
│   │   ├── postgres.go               # User, session, device, audit persistence
│   │   └── redis.go                  # OTP cache, rate limits, token blocklist
│   │
│   └── model/
│       ├── user.go                    # User entity
│       ├── session.go                 # Session & refresh token entity
│       ├── device.go                  # Device entity
│       └── audit.go                   # Audit log entry entity
│
├── pkg/
│   ├── crypto/                        # Ed25519 helpers, hashing utilities
│   ├── cardano/                       # Cardano blockchain client wrapper
│   └── sms/                           # Twilio + silent push abstraction
│
├── config/
│   └── config.go                      # Environment-based configuration
│
└── migrations/
    └── *.sql                          # Database migration files
```

---

## 3. API Endpoints

All authentication endpoints live under `/v1/auth`. Every request must include a `X-Device-Info` header (JSON-encoded `DeviceInfo` struct) or the request is rejected with `400`.

### Registration Endpoints

| Method | Endpoint | Auth | Rate Limit | Description |
|---|---|---|---|---|
| `POST` | `/auth/register/phone` | None | 5/phone/hr | Send OTP to phone number |
| `POST` | `/auth/register/verify` | None | 5/session | Verify OTP, create account + DID |
| `POST` | `/auth/register/passkey` | Bearer (temp) | 3/min | Register passkey after OTP verify |
| `POST` | `/auth/register/credential` | None | 10/IP/hr | Register via verifiable credential |

### Login & Session Endpoints

| Method | Endpoint | Auth | Rate Limit | Description |
|---|---|---|---|---|
| `POST` | `/auth/login` | None | 10/IP/min | Login with passkey or DID signature |
| `POST` | `/auth/refresh` | Refresh Token | 30/hr | Rotate access + refresh tokens |
| `POST` | `/auth/logout` | Bearer | None | Invalidate current or all sessions |
| `POST` | `/auth/step-up` | Bearer | 5/hr | Elevate session for sensitive ops |

### Recovery Endpoints

| Method | Endpoint | Auth | Rate Limit | Description |
|---|---|---|---|---|
| `POST` | `/auth/recovery/initiate` | None | 3/phone/day | Start account recovery |
| `POST` | `/auth/recovery/verify` | Recovery Token | 5/session | Complete recovery flow |

### Device Management Endpoints

| Method | Endpoint | Auth | Rate Limit | Description |
|---|---|---|---|---|
| `GET` | `/auth/devices` | Bearer | 30/min | List registered devices |
| `DELETE` | `/auth/devices/:id` | Bearer + Step-Up | 10/hr | Revoke a device session |
| `GET` | `/auth/audit-log` | Bearer | 10/min | View login attempt history |

> **Note:** `/auth/step-up`, `/auth/devices`, and `/auth/audit-log` are **new endpoints** not in the original OpenAPI spec. They close critical security and usability gaps.

---

## 4. Registration Flow

### 4.1 Phone Registration — `POST /auth/register/phone`

**Request:**
```json
{
  "phone_number": "5551234567",
  "country_code": "+1",
  "device_info": { /* DeviceInfo struct */ }
}
```

**Service Logic:**

1. Validate phone number format and country code against E.164 pattern
2. Check rate limit: max 5 OTP requests per phone per hour (`rl:otp:{phone}` in Redis)
3. Check if phone is already registered — if so, **return identical success response** (prevent enumeration)
4. Generate 6-digit OTP via `crypto/rand`
5. Hash OTP with bcrypt (cost=10), store in Redis with key `otp:{verification_id}` and 10-min TTL
6. Send OTP via silent push notification (primary) or SMS via Twilio (fallback)
7. Return `verification_id` (UUID v4) and `retry_after: 60`

**Response (200):**
```json
{
  "verification_id": "550e8400-e29b-41d4-a716-446655440000",
  "expires_at": "2026-03-11T12:10:00Z",
  "retry_after": 60
}
```

> **Security:** The response is identical whether the phone is new or already registered. This prevents phone number enumeration attacks.

### 4.2 OTP Verification — `POST /auth/register/verify`

**Request:**
```json
{
  "verification_id": "550e8400-e29b-41d4-a716-446655440000",
  "code": "482916",
  "device_info": { /* DeviceInfo struct */ }
}
```

**Service Logic:**

1. Look up verification session by `verification_id` in Redis
2. If session not found or expired → return `AUTH_003` error
3. Increment attempt counter (`otp_attempts:{verification_id}`)
4. If attempts > 5 → invalidate session, return `AUTH_003`
5. Compare submitted code against bcrypt hash
6. If mismatch → return `AUTH_003` (generic, no hint)
7. **On success:**
   - Create user record in Postgres (`status = pending_passkey`)
   - Start DID generation via Atala PRISM SDK (async goroutine, non-blocking)
   - Generate temporary access token (JWT, 5-min TTL, `scope: passkey_registration`)
   - Generate WebAuthn challenge (random 32 bytes), store server-side with 5-min TTL
8. Return `AuthResponse` with temp tokens, user stub, and `passkey_challenge`

**Response (201):**
```json
{
  "access_token": "eyJhbG...",
  "refresh_token": null,
  "expires_at": "2026-03-11T12:05:00Z",
  "user": {
    "id": "usr_abc123",
    "did": "did:prism:cardano:pending",
    "status": "pending_passkey",
    "trust_score": 0
  },
  "passkey_challenge": "base64-encoded-32-bytes"
}
```

### 4.3 Passkey Registration — `POST /auth/register/passkey`

**Request:**
```json
{
  "attestation_response": {
    "id": "credential-id-base64",
    "raw_id": "raw-credential-id-base64",
    "response": {
      "client_data_json": "base64...",
      "attestation_object": "base64..."
    },
    "type": "public-key"
  },
  "device_info": { /* DeviceInfo struct */ }
}
```

**Service Logic:**

1. Validate temp access token (must have `scope: passkey_registration`)
2. Parse WebAuthn attestation response using `go-webauthn` library
3. Verify attestation:
   - **Production:** Require Apple App Attestation format, validate certificate chain
   - **Debug builds only:** Allow `none` attestation format
4. Extract public key (COSE format) from attestation object
5. Compute `device_id` hash from `DeviceInfo` fields
6. Store in Postgres `credentials` table:
   - `credential_id`, `public_key`, `sign_count: 0`, `device_id`, `aaguid`, `created_at`
7. Update DID document on Cardano with new public key (async)
8. Update user `status = active`, assign initial trust score = 5 (device-verified)
9. Issue full access token (15-min TTL) and refresh token (30-day TTL), both device-bound
10. Delete temp tokens and passkey challenge from Redis

**Response (201):**
```json
{
  "access_token": "eyJhbG...",
  "refresh_token": "rt_a1b2c3d4...",
  "expires_at": "2026-03-11T12:15:00Z",
  "user": {
    "id": "usr_abc123",
    "did": "did:prism:cardano:abc123def456",
    "display_name": null,
    "username": null,
    "trust_score": 5,
    "trust_tier": 0,
    "status": "active"
  }
}
```

---

## 5. Login Flow

### 5.1 Pre-Login Challenge — `GET /auth/login/challenge`

Before a passkey login, the client must request a challenge:

```json
// Response
{
  "challenge": "base64-encoded-32-bytes",
  "timeout": 300000,
  "rp_id": "echo.app"
}
```

The challenge is stored server-side in Redis with a 5-min TTL keyed to the client's session.

### 5.2 Passkey Login — `POST /auth/login`

**Request:**
```json
{
  "auth_type": "passkey",
  "credential": {
    "id": "credential-id-base64",
    "raw_id": "raw-credential-id-base64",
    "response": {
      "client_data_json": "base64...",
      "authenticator_data": "base64...",
      "signature": "base64..."
    },
    "type": "public-key"
  },
  "device_info": { /* DeviceInfo struct */ }
}
```

**Service Logic:**

1. Rate limit check: 10 attempts per IP per minute (`rl:login:{ip}` sliding window)
2. Retrieve challenge from server-side session — must match `clientDataJSON.challenge`
3. Look up stored credential by `credential_id` in Postgres
4. Verify assertion signature against stored public key using `go-webauthn`
5. Verify `sign_count` is strictly greater than stored value → detects cloned passkeys
6. Compute device fingerprint from request `DeviceInfo`
7. Compare against registered `device_id`:
   - **Match:** Proceed normally
   - **Mismatch:** Trigger new device flow → return `AUTH_007`, require step-up
8. Update `sign_count` and `last_login_at` in Postgres
9. Write audit log entry: `{ user_id, ip, device_id, timestamp, result: success }`
10. Issue access token (15-min) + refresh token (30-day), both bound to `device_id`

**Response (200):** Standard `AuthResponse`.

### 5.3 DID Signature Login — `POST /auth/login`

**Request:**
```json
{
  "auth_type": "did_signature",
  "did": "did:prism:cardano:abc123def456",
  "signature": "base64-ed25519-signature",
  "timestamp": "2026-03-11T12:00:00Z",
  "nonce": "random-unique-nonce",
  "device_info": { /* DeviceInfo struct */ }
}
```

**Service Logic:**

1. Rate limit check (same as passkey)
2. **Anti-replay validation:**
   - Verify `timestamp` is within 5-minute window of server time
   - Verify `nonce` has not been used before (check Redis set `used_nonces`, add with 10-min TTL)
3. Resolve DID document from Cardano (or local cache, 24hr TTL)
4. Extract public key from DID document's `authentication` method
5. Construct signed message: `echo:auth:{did}:{timestamp}:{nonce}`
6. Verify Ed25519 signature against public key
7. Proceed with device check and token issuance (same as passkey flow steps 6-10)

> **Security:** The mandatory nonce + timestamp window prevents replay attacks. This was **optional** in the original spec and is now **required**.

---

## 6. Token Architecture

### 6.1 Access Token (JWT)

**Format:** JWT signed with ES256 (ECDSA P-256)

```json
{
  "iss": "https://api.echo.app",
  "sub": "did:prism:cardano:abc123",
  "iat": 1710000000,
  "exp": 1710000900,
  "jti": "unique-token-id-uuid",
  "device_id": "sha256-device-fingerprint",
  "trust_tier": 2,
  "scope": "messaging payments",
  "elevated": false
}
```

| Claim | Purpose |
|---|---|
| `sub` | User's DID — primary identifier |
| `jti` | Unique token ID for revocation checking |
| `device_id` | Binds token to specific device |
| `trust_tier` | Current trust level (0–4) for feature gating |
| `scope` | Permitted API scopes based on trust tier |
| `elevated` | `true` when step-up auth completed (5-min TTL) |

**TTL:** 15 minutes. Stored **in-memory only** on the client (never persisted).

### 6.2 Refresh Token

**Format:** Opaque UUID v4 string (not a JWT)

| Property | Value |
|---|---|
| Format | `rt_{uuid_v4}` |
| TTL | 30 days |
| Storage (server) | Postgres `refresh_tokens` table |
| Storage (client) | iOS Keychain, `kSecAttrAccessibleWhenUnlockedThisDeviceOnly` |
| Rotation | Single-use — new token issued on every refresh |
| Device Binding | Bound to `device_id` hash at creation |

### 6.3 Signing Key Management

```go
// config/config.go

type AuthConfig struct {
    JWTPrivateKey      *ecdsa.PrivateKey  // ES256 signing key
    JWTPublicKey       *ecdsa.PublicKey   // Verification key
    KeyRotationPeriod  time.Duration      // 90 days
    OldKeysGracePeriod time.Duration      // 24 hours (accept old key during rotation)
}
```

Keys are rotated every 90 days. During rotation, the old key remains valid for 24 hours to allow in-flight tokens to complete. The `kid` (Key ID) header in the JWT identifies which key was used.

---

## 7. Token Refresh with Rotation

**Endpoint:** `POST /auth/refresh`

**Request:**
```json
{
  "refresh_token": "rt_a1b2c3d4-e5f6-7890-abcd-ef1234567890"
}
```

**Service Logic:**

1. Look up refresh token in Postgres by token hash
2. Validate:
   - Token exists and is not revoked
   - Token has not expired (30-day TTL from creation)
   - `device_id` in token record matches request `DeviceInfo` hash
3. **Immediately mark the old refresh token as `used`** (single-use enforcement)
4. **Replay detection:** If the token was already marked as `used`:
   - This means the token was stolen and both the attacker and legitimate client are using it
   - **Revoke ALL refresh tokens for this user** (nuclear option)
   - Return `AUTH_006` and force full re-authentication
   - Write critical audit log entry
5. Generate new access token (15-min TTL)
6. Generate new refresh token (30-day TTL), store in Postgres bound to same `device_id`
7. Return new `AuthResponse`

**Response (200):**
```json
{
  "access_token": "eyJhbG...(new)",
  "refresh_token": "rt_new-token-uuid...",
  "expires_at": "2026-03-11T12:15:00Z",
  "user": { /* current user profile */ }
}
```

> **Security:** Single-use refresh tokens with automatic replay detection. If a stolen token is used after the legitimate client already rotated it, ALL sessions are immediately revoked. This is the recommended IETF best practice (RFC 6749 Section 10.4).

---

## 8. Step-Up Authentication

Certain sensitive operations require elevated privileges. The step-up mechanism issues a short-lived elevated token without requiring full re-authentication.

### 8.1 Trigger Actions

| Action | Required Step-Up Method | Elevated Token TTL |
|---|---|---|
| Change phone number | Fresh OTP to current phone | 5 minutes |
| Revoke a device | Biometric passkey re-verify | 5 minutes |
| Send payment > 100 ECHO | Biometric passkey re-verify | 5 minutes |
| Export recovery phrase | OTP + biometric | 2 minutes |
| Delete account | OTP + biometric + 24hr cooling | N/A (queued) |

### 8.2 Step-Up Flow — `POST /auth/step-up`

**Request:**
```json
{
  "method": "passkey",
  "credential": { /* WebAuthn assertion */ },
  "action": "revoke_device"
}
```

**Service Logic:**

1. Validate current bearer token (must be authenticated)
2. Verify the requested `action` requires step-up
3. Perform verification based on `method`:
   - `passkey`: Full WebAuthn assertion verification (same as login)
   - `otp`: Send fresh OTP, verify code
   - `passkey+otp`: Both must succeed
4. Issue new access token with `elevated: true` and short TTL (2–5 min based on action)
5. Log step-up event in audit log

**Response (200):**
```json
{
  "elevated_token": "eyJhbG...",
  "expires_at": "2026-03-11T12:05:00Z",
  "action": "revoke_device"
}
```

### 8.3 Step-Up Middleware

```go
// internal/auth/middleware.go

func RequireStepUp(action string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            claims := getClaimsFromContext(r.Context())
            if !claims.Elevated {
                writeError(w, http.StatusForbidden, "AUTH_008", 
                    "Additional verification required")
                return
            }
            // Verify the elevated token was issued for this specific action
            if claims.ElevatedAction != action {
                writeError(w, http.StatusForbidden, "AUTH_008",
                    "Elevated token not valid for this action")
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}

// Usage in router:
// r.With(RequireStepUp("revoke_device")).Delete("/auth/devices/{id}", handler.RevokeDevice)
```

---

## 9. Rate Limiting Strategy

All rate limits are implemented with Redis sliding window counters. The middleware runs **before** handler logic to reject requests early and conserve resources.

### 9.1 Rate Limit Configuration

| Endpoint Category | Redis Key Pattern | Limit | Window | On Exceed |
|---|---|---|---|---|
| OTP send | `rl:otp:{phone_hash}` | 5 requests | 1 hour | Block phone 1hr |
| OTP verify | `rl:verify:{session_id}` | 5 attempts | Per session lifetime | Invalidate session |
| Login (per IP) | `rl:login:ip:{ip}` | 10 attempts | 1 minute | 429 + `Retry-After` header |
| Login (per account) | `rl:login:did:{did_hash}` | 20 attempts | 1 hour | Temp lock + audit alert |
| Token refresh | `rl:refresh:{did_hash}` | 30 requests | 1 hour | Revoke all sessions |
| General API (unverified) | `rl:api:{did_hash}` | 100 requests | 1 minute | 429 response |
| General API (verified) | `rl:api:{did_hash}` | 500 requests | 1 minute | 429 response |
| Step-up | `rl:stepup:{did_hash}` | 5 requests | 1 hour | 429 response |

### 9.2 Implementation

```go
// internal/auth/ratelimit.go

type RateLimiter struct {
    redis *redis.Client
}

type RateLimitConfig struct {
    Key     string
    Limit   int
    Window  time.Duration
    Penalty func(ctx context.Context) error  // optional escalation
}

func (rl *RateLimiter) Check(ctx context.Context, cfg RateLimitConfig) error {
    pipe := rl.redis.Pipeline()
    now := time.Now().UnixMilli()
    windowStart := now - cfg.Window.Milliseconds()

    // Remove expired entries
    pipe.ZRemRangeByScore(ctx, cfg.Key, "0", fmt.Sprintf("%d", windowStart))
    // Add current request
    pipe.ZAdd(ctx, cfg.Key, redis.Z{Score: float64(now), Member: fmt.Sprintf("%d", now)})
    // Count requests in window
    countCmd := pipe.ZCard(ctx, cfg.Key)
    // Set TTL on the key
    pipe.Expire(ctx, cfg.Key, cfg.Window)

    _, err := pipe.Exec(ctx)
    if err != nil {
        return fmt.Errorf("rate limit check failed: %w", err)
    }

    if countCmd.Val() > int64(cfg.Limit) {
        if cfg.Penalty != nil {
            _ = cfg.Penalty(ctx)
        }
        return ErrRateLimitExceeded
    }
    return nil
}
```

### 9.3 Response Headers

Every response includes rate limit headers:

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 87
X-RateLimit-Reset: 1710001200
Retry-After: 45           # Only on 429 responses
```

---

## 10. Device Fingerprinting

### 10.1 DeviceInfo Schema

```go
// internal/model/device.go

type DeviceInfo struct {
    DeviceID        string `json:"device_id"         validate:"required"`
    Platform        string `json:"platform"          validate:"required,oneof=ios"`
    OSVersion       string `json:"os_version"        validate:"required"`
    AppVersion      string `json:"app_version"       validate:"required,semver"`
    Model           string `json:"model"             validate:"required"`
    Locale          string `json:"locale"`
    Timezone        string `json:"timezone"`
    SecureEnclave   bool   `json:"secure_enclave"`
    BiometricType   string `json:"biometric_type"    validate:"oneof=face_id touch_id none"`
    JailbreakStatus bool   `json:"jailbreak_status"`
    PushToken       string `json:"push_token,omitempty"`
}

type DeviceRecord struct {
    ID              string    `db:"id"`               // UUID
    UserID          string    `db:"user_id"`
    DeviceHash      string    `db:"device_hash"`      // SHA-256 of stable fields
    FriendlyName    string    `db:"friendly_name"`    // "iPhone 15 Pro"
    Platform        string    `db:"platform"`
    OSVersion       string    `db:"os_version"`
    LastIP          string    `db:"last_ip"`
    LastLocation    string    `db:"last_location"`    // Geo from IP
    LastActiveAt    time.Time `db:"last_active_at"`
    CreatedAt       time.Time `db:"created_at"`
    CredentialID    string    `db:"credential_id"`    // FK to passkey
    IsCurrentDevice bool      `db:"-"`                // Computed at query time
}
```

### 10.2 Device Hash Computation

```go
func ComputeDeviceHash(info DeviceInfo) string {
    // Use stable fields only (exclude version numbers that change on update)
    data := fmt.Sprintf("%s:%s:%s", info.DeviceID, info.Platform, info.Model)
    hash := sha256.Sum256([]byte(data))
    return hex.EncodeToString(hash[:])
}
```

### 10.3 Jailbreak Rejection

```go
func (s *AuthService) ValidateDeviceIntegrity(info DeviceInfo) error {
    if info.JailbreakStatus {
        return &AuthError{Code: "AUTH_010", Message: "Device verification failed"}
    }
    if !info.SecureEnclave {
        return &AuthError{Code: "AUTH_010", Message: "Device verification failed"}
    }
    return nil
}
```

---

## 11. Account Recovery

### 11.1 Recovery Methods

| Method | Prerequisite | Security Model |
|---|---|---|
| Recovery Phrase | User saved 12-word BIP-39 phrase at onboarding | Full self-custody; phrase reconstructs DID key |
| Trusted Contacts | User designated 3+ trusted contacts | Social recovery with 2-of-3 Shamir threshold |
| Phone Re-verification | Same phone number still accessible | OTP to phone + fresh passkey registration |

### 11.2 Initiate Recovery — `POST /auth/recovery/initiate`

**Request:**
```json
{
  "recovery_method": "trusted_contacts",
  "identifier": "+15551234567",
  "device_info": { /* DeviceInfo */ }
}
```

**Service Logic:**

1. Rate limit: 3 attempts per phone per day
2. Identify user account by phone or DID
3. Verify the user has the requested recovery method configured
4. Based on method:
   - **`recovery_phrase`:** Return challenge to sign with derived key
   - **`trusted_contacts`:** Send recovery request to all trusted contacts
   - **`phone`:** Send OTP to registered phone number
5. Create recovery session in Postgres (24hr TTL)
6. Return `recovery_session_id` and `required_steps`

### 11.3 Trusted Contact Recovery (Detailed)

1. Server sends push notification to all trusted contacts (minimum 3)
2. Each contact opens notification → sees "Alex is trying to recover their account"
3. Contact authenticates with their own passkey
4. Contact confirms they know the requester (explicit button tap)
5. Server generates a recovery shard for each confirming contact (Shamir's Secret Sharing)
6. When 2-of-3 shards are collected → server reconstructs recovery authorization
7. User receives recovery token → can register new passkey on new device
8. All old device sessions are immediately revoked
9. DID document is updated with new device's public key (Cardano transaction)

### 11.4 Complete Recovery — `POST /auth/recovery/verify`

**Request (trusted contacts):**
```json
{
  "recovery_session_id": "uuid",
  "method": "trusted_contacts"
}
```

Server automatically completes when shard threshold is met. Response includes temp token for passkey registration (same flow as initial registration step 4.3).

---

## 12. Error Code Catalog

All error responses follow this format:

```json
{
  "error": {
    "code": "AUTH_003",
    "message": "That code is incorrect. Please try again.",
    "retry_after": null
  }
}
```

| Code | HTTP | Internal Meaning | User-Facing Message |
|---|---|---|---|
| `AUTH_001` | 400 | Invalid phone number format | Please enter a valid phone number. |
| `AUTH_002` | 429 | OTP rate limit exceeded | Too many attempts. Please try again later. |
| `AUTH_003` | 400 | Invalid or expired OTP | That code is incorrect. Please try again. |
| `AUTH_004` | 401 | Passkey verification failed | Authentication failed. Please try again. |
| `AUTH_005` | 401 | Access token expired | Your session has expired. Please sign in. |
| `AUTH_006` | 401 | Refresh token invalid or reused | Please sign in again. |
| `AUTH_007` | 403 | Unknown device detected | New device detected. Verify your identity. |
| `AUTH_008` | 403 | Step-up authentication required | Additional verification required. |
| `AUTH_009` | 423 | Account temporarily locked | Account locked. Try again in 1 hour. |
| `AUTH_010` | 400 | Passkey attestation or device integrity failure | Device verification failed. |
| `AUTH_011` | 400 | Recovery session expired or invalid | Recovery session expired. Please restart. |
| `AUTH_012` | 429 | Global rate limit exceeded | Too many requests. Please slow down. |

> **Security:** Error messages NEVER reveal whether a phone number is registered, whether a username exists, or any internal state. All messages are deliberately generic.

---

## 13. Database Schema

### 13.1 Users Table

```sql
CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    did             TEXT UNIQUE,
    phone_hash      TEXT UNIQUE NOT NULL,          -- SHA-256 of E.164 phone
    display_name    TEXT,
    username        TEXT UNIQUE,
    status          TEXT NOT NULL DEFAULT 'pending_passkey'
                    CHECK (status IN ('pending_passkey', 'active', 'suspended', 'deleted')),
    trust_score     INTEGER NOT NULL DEFAULT 0,
    trust_tier      INTEGER NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

### 13.2 Credentials Table (Passkeys)

```sql
CREATE TABLE credentials (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id),
    credential_id   TEXT UNIQUE NOT NULL,           -- WebAuthn credential ID
    public_key      BYTEA NOT NULL,                 -- COSE public key
    sign_count      BIGINT NOT NULL DEFAULT 0,
    device_id       TEXT NOT NULL,                   -- Device hash
    aaguid          TEXT,                            -- Authenticator identifier
    friendly_name   TEXT,                            -- "iPhone 15 Pro"
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_used_at    TIMESTAMPTZ
);

CREATE INDEX idx_credentials_user ON credentials(user_id);
CREATE INDEX idx_credentials_device ON credentials(device_id);
```

### 13.3 Refresh Tokens Table

```sql
CREATE TABLE refresh_tokens (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id),
    token_hash      TEXT UNIQUE NOT NULL,            -- SHA-256 of token
    device_id       TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'active'
                    CHECK (status IN ('active', 'used', 'revoked')),
    expires_at      TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    used_at         TIMESTAMPTZ                      -- Set on first use
);

CREATE INDEX idx_refresh_tokens_user ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_hash ON refresh_tokens(token_hash);
```

### 13.4 Audit Log Table

```sql
CREATE TABLE auth_audit_log (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID REFERENCES users(id),       -- Nullable for failed lookups
    event_type      TEXT NOT NULL,                    -- 'login', 'register', 'refresh', 'logout', 'step_up', 'recovery'
    result          TEXT NOT NULL,                    -- 'success', 'failed', 'blocked'
    ip_address      INET NOT NULL,
    device_id       TEXT,
    device_info     JSONB,
    error_code      TEXT,                             -- e.g. 'AUTH_004'
    metadata        JSONB,                            -- Additional context
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_audit_user ON auth_audit_log(user_id, created_at DESC);
CREATE INDEX idx_audit_ip ON auth_audit_log(ip_address, created_at DESC);
```

---

## 14. Security Middleware

### 14.1 Auth Middleware Stack

```go
// Applied to all authenticated routes
r.Use(
    middleware.RequestID,
    middleware.RealIP,
    middleware.Logger,
    auth.ExtractDeviceInfo,       // Parse X-Device-Info header
    auth.ValidateDeviceIntegrity, // Reject jailbroken devices
    auth.RateLimit,               // Check rate limits
    auth.ValidateJWT,             // Verify and decode access token
    auth.BindDevice,              // Verify token device_id matches request device
)
```

### 14.2 JWT Validation

```go
func (m *AuthMiddleware) ValidateJWT(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        tokenString := extractBearerToken(r)
        if tokenString == "" {
            writeError(w, 401, "AUTH_005", "Authentication required")
            return
        }

        claims, err := m.tokenService.ValidateAccessToken(tokenString)
        if err != nil {
            if errors.Is(err, jwt.ErrTokenExpired) {
                writeError(w, 401, "AUTH_005", "Session expired")
            } else {
                writeError(w, 401, "AUTH_004", "Authentication failed")
            }
            return
        }

        // Check blocklist (for force-revoked tokens)
        if m.tokenService.IsBlocklisted(r.Context(), claims.ID) {
            writeError(w, 401, "AUTH_006", "Please sign in again")
            return
        }

        ctx := context.WithValue(r.Context(), claimsKey, claims)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

---

## 15. Testing Requirements

| Category | Coverage Target | Tools | Key Scenarios |
|---|---|---|---|
| Unit Tests | 90% of `auth/` package | `testing`, `testify` | OTP hashing, JWT claims, rate limit logic, device hash |
| Integration Tests | All auth endpoints | `httptest`, testcontainers (Postgres, Redis) | Full registration flow, login flow, token rotation, replay detection |
| Security Tests | Auth attack surface | Custom test harness | Brute-force OTP, token replay, sign counter bypass, device spoofing |
| Load Tests | Rate limit validation | `k6` or `vegeta` | Verify rate limits hold under 1000 RPS |
| Penetration Test | Full auth surface | External auditor | Pre-launch (recommended: Trail of Bits or NCC Group) |

### Critical Test Scenarios

1. **OTP brute force:** 6th attempt on same session must fail regardless of correct code
2. **Token replay:** Using an already-rotated refresh token must revoke ALL user sessions
3. **Sign counter regression:** Login with sign_count <= stored must reject (cloned passkey)
4. **Device mismatch:** Valid passkey from wrong device must trigger step-up, not succeed
5. **Concurrent refresh:** Two simultaneous refresh requests — one must succeed, one must trigger replay detection
6. **Enumeration prevention:** Registering an existing phone returns same response shape and timing as new phone

---

*Last Updated: March 2026*
*Go Version: 1.22+*
*Target: Production deployment Phase 1*
