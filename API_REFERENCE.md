# Identity Verification System - Complete API Reference

## Base URL
```
https://api.example.com/api/v1
```

## Authentication
All requests should include:
- `Authorization: Bearer {token}`
- `Content-Type: application/json`

## Response Format

### Success Response
```json
{
  "success": true,
  "data": {...},
  "tx_hash": "abc123...",
  "timestamp": "2024-01-15T10:30:00Z",
  "request_id": "req-12345"
}
```

### Error Response
```json
{
  "success": false,
  "message": "Error description",
  "code": "ERROR_CODE",
  "details": "Additional context",
  "request_id": "req-12345"
}
```

---

## SCHEMA MANAGEMENT ENDPOINTS

### Create Schema
**POST** `/schemas`

Create a new credential schema with version 1.

**Request Body:**
```json
{
  "name": "University Degree",
  "description": "Bachelor's degree credential schema",
  "version": 1,
  "created_by": "did:example:issuer123",
  "properties": {
    "degree_type": {
      "type": "string",
      "enum": ["Bachelor", "Master", "PhD"]
    },
    "university": {
      "type": "string"
    },
    "graduation_date": {
      "type": "string",
      "format": "date"
    },
    "gpa": {
      "type": "number",
      "minimum": 0,
      "maximum": 4.0
    }
  },
  "required": ["degree_type", "university", "graduation_date"],
  "type": ["VerifiableCredential", "UniversityDegree"],
  "context": [
    "https://www.w3.org/2018/credentials/v1",
    "https://www.w3.org/2018/credentials/examples/v1"
  ]
}
```

**Response (201):**
```json
{
  "success": true,
  "schema_id": "schema_1705316400000000000",
  "version": 1,
  "tx_hash": "tx_abc123...",
  "timestamp": "2024-01-15T10:30:00Z",
  "request_id": "req-12345"
}
```

---

### Get Schema
**GET** `/schemas/{schemaId}?version=1`

Retrieve a specific schema version.

**Parameters:**
- `schemaId` (path, required): Schema identifier
- `version` (query, optional): Version number (default: 1)

**Response (200):**
```json
{
  "success": true,
  "schema": {
    "schema_id": "schema_1705316400000000000",
    "name": "University Degree",
    "description": "Bachelor's degree credential schema",
    "version": 1,
    "created_at": "2024-01-15T10:30:00Z",
    "created_by": "did:example:issuer123",
    "content_hash": "abc123def456...",
    "properties": {...},
    "required": [...],
    "context": [...],
    "status": "active",
    "tx_hash": "tx_abc123...",
    "timestamp": "2024-01-15T10:30:00Z"
  },
  "request_id": "req-12345"
}
```

---

### Query Schemas
**GET** `/schemas?schema_id=...&created_by=...&status=active`

Query schemas with filters.

**Query Parameters:**
- `schema_id` (optional): Filter by schema ID
- `created_by` (optional): Filter by creator DID
- `status` (optional): Filter by status (active, deprecated, archived)
- `limit` (optional): Results per page (default: 100)
- `offset` (optional): Pagination offset (default: 0)

**Response (200):**
```json
{
  "success": true,
  "total": 45,
  "count": 10,
  "schemas": [...],
  "query_time": "45.2ms",
  "request_id": "req-12345"
}
```

---

### Update Schema
**PUT** `/schemas/{schemaId}`

Update a schema (creates new version).

**Request Body:**
```json
{
  "name": "University Degree",
  "description": "Updated: Bachelor's degree credential schema",
  "version": 2,
  "created_by": "did:example:issuer123",
  "properties": {...},
  "required": [...]
}
```

**Response (200):**
```json
{
  "success": true,
  "schema_id": "schema_1705316400000000000",
  "version": 2,
  "tx_hash": "tx_xyz789...",
  "timestamp": "2024-01-15T10:35:00Z",
  "request_id": "req-12345"
}
```

---

### Deprecate Schema
**POST** `/schemas/{schemaId}/deprecate`

Mark a schema as deprecated.

**Request Body:**
```json
{
  "reason": "Schema v1 superseded by v2"
}
```

**Response (200):**
```json
{
  "success": true,
  "schema_id": "schema_1705316400000000000",
  "action": "deprecated",
  "tx_hash": "tx_abc123...",
  "timestamp": "2024-01-15T10:40:00Z",
  "request_id": "req-12345"
}
```

---

### Get Schema Version History
**GET** `/schemas/{schemaId}/history`

Retrieve all versions of a schema.

**Response (200):**
```json
{
  "success": true,
  "schema_id": "schema_1705316400000000000",
  "version_count": 3,
  "versions": [
    {
      "version": 1,
      "schema_id": "schema_1705316400000000000",
      "content_hash": "abc123...",
      "created_at": "2024-01-15T10:30:00Z",
      "created_by": "did:example:issuer123"
    },
    {
      "version": 2,
      "schema_id": "schema_1705316400000000000",
      "content_hash": "xyz789...",
      "created_at": "2024-01-15T10:35:00Z",
      "created_by": "did:example:issuer123"
    }
  ],
  "request_id": "req-12345"
}
```

---

### Validate Credential Against Schema
**POST** `/schemas/{schemaId}/validate`

Validate a credential against its schema.

**Request Body:**
```json
{
  "credential": {
    "degree_type": "Bachelor",
    "university": "MIT",
    "graduation_date": "2023-06-15",
    "gpa": 3.85
  }
}
```

**Response (200) - Valid:**
```json
{
  "success": true,
  "valid": true,
  "errors": [],
  "request_id": "req-12345"
}
```

**Response (200) - Invalid:**
```json
{
  "success": true,
  "valid": false,
  "errors": [
    "required field missing: graduation_date",
    "unexpected field: major"
  ],
  "request_id": "req-12345"
}
```

---

## ISSUER MANAGEMENT ENDPOINTS

### Register Issuer
**POST** `/issuers`

Register a new issuer on-chain.

**Request Body:**
```json
{
  "issuer_did": "did:example:issuer123",
  "name": "MIT",
  "description": "Massachusetts Institute of Technology",
  "public_key": "-----BEGIN PUBLIC KEY-----\n...\n-----END PUBLIC KEY-----",
  "authority_level": "standard",
  "metadata": {
    "logo_url": "https://example.com/logo.png",
    "website": "https://mit.edu"
  }
}
```

**Response (201):**
```json
{
  "success": true,
  "issuer_did": "did:example:issuer123",
  "status": "pending",
  "authority_level": "standard",
  "tx_hash": "tx_abc123...",
  "registered_at": "2024-01-15T10:30:00Z",
  "request_id": "req-12345"
}
```

---

### Get Issuer
**GET** `/issuers/{issuerId}`

Retrieve issuer information.

**Response (200):**
```json
{
  "success": true,
  "issuer": {
    "issuer_did": "did:example:issuer123",
    "name": "MIT",
    "description": "Massachusetts Institute of Technology",
    "status": "verified",
    "authority_level": "standard",
    "registered_at": "2024-01-15T10:30:00Z",
    "verified_at": "2024-01-15T10:45:00Z",
    "public_key": "...",
    "metadata": {...},
    "credential_count": 1250,
    "rating": 4.8,
    "trust_score": 98.5,
    "tx_hash": "tx_abc123...",
    "timestamp": "2024-01-15T10:45:00Z"
  },
  "request_id": "req-12345"
}
```

---

### Verify Issuer
**POST** `/issuers/{issuerId}/verify`

Verify an issuer (moves from pending to verified).

**Request Body:**
```json
{
  "verification_url": "https://authority.example.com/issuer/verify/123",
  "authority_level": "standard"
}
```

**Response (200):**
```json
{
  "success": true,
  "issuer_did": "did:example:issuer123",
  "status": "verified",
  "tx_hash": "tx_xyz789...",
  "verified_at": "2024-01-15T10:45:00Z",
  "request_id": "req-12345"
}
```

---

### Suspend Issuer
**POST** `/issuers/{issuerId}/suspend`

Temporarily suspend an issuer.

**Request Body:**
```json
{
  "reason": "Compliance review in progress"
}
```

**Response (200):**
```json
{
  "success": true,
  "issuer_did": "did:example:issuer123",
  "status": "suspended",
  "reason": "Compliance review in progress",
  "tx_hash": "tx_abc123...",
  "request_id": "req-12345"
}
```

---

### Revoke Issuer
**POST** `/issuers/{issuerId}/revoke`

Permanently revoke an issuer.

**Request Body:**
```json
{
  "reason": "Credential fraud detected"
}
```

**Response (200):**
```json
{
  "success": true,
  "issuer_did": "did:example:issuer123",
  "status": "revoked",
  "reason": "Credential fraud detected",
  "tx_hash": "tx_abc123...",
  "request_id": "req-12345"
}
```

---

### Get Issuer Audit Trail
**GET** `/issuers/{issuerId}/audit`

Retrieve all audit entries for an issuer.

**Response (200):**
```json
{
  "success": true,
  "issuer_did": "did:example:issuer123",
  "audit_count": 5,
  "audit_trail": [
    {
      "id": "audit_001",
      "issuer_did": "did:example:issuer123",
      "action": "registered",
      "new_status": "pending",
      "timestamp": "2024-01-15T10:30:00Z",
      "tx_hash": "tx_abc123..."
    },
    {
      "id": "audit_002",
      "issuer_did": "did:example:issuer123",
      "action": "verified",
      "old_status": "pending",
      "new_status": "verified",
      "timestamp": "2024-01-15T10:45:00Z",
      "tx_hash": "tx_xyz789..."
    }
  ],
  "request_id": "req-12345"
}
```

---

### Get Issuer Credentials
**GET** `/issuers/{issuerId}/credentials`

Retrieve all credentials issued by an issuer.

**Response (200):**
```json
{
  "success": true,
  "issuer_did": "did:example:issuer123",
  "credential_count": 1250,
  "credentials": [...],
  "request_id": "req-12345"
}
```

---

### Update Issuer Metadata
**PUT** `/issuers/{issuerId}/metadata`

Update issuer metadata on-chain.

**Request Body:**
```json
{
  "logo_url": "https://new-domain.com/logo.png",
  "website": "https://new-website.edu",
  "support_email": "support@example.com"
}
```

**Response (200):**
```json
{
  "success": true,
  "issuer_did": "did:example:issuer123",
  "metadata": {...},
  "tx_hash": "tx_abc123...",
  "request_id": "req-12345"
}
```

---

## CREDENTIAL MANAGEMENT ENDPOINTS

### Get Credential Audit Trail
**GET** `/credentials/{credentialId}/audit`

Retrieve audit trail for a credential.

**Response (200):**
```json
{
  "success": true,
  "credential_id": "cred_123",
  "event_count": 3,
  "audit_trail": [
    {
      "id": "event_001",
      "credential_id": "cred_123",
      "event_type": "issued",
      "actor": "did:example:issuer123",
      "timestamp": "2024-01-15T10:30:00Z",
      "tx_hash": "tx_abc123..."
    },
    {
      "id": "event_002",
      "credential_id": "cred_123",
      "event_type": "verified",
      "actor": "did:example:verifier456",
      "timestamp": "2024-01-15T10:35:00Z",
      "tx_hash": "tx_xyz789..."
    }
  ],
  "last_event": {...},
  "request_id": "req-12345"
}
```

---

### Revoke Credential
**POST** `/credentials/{credentialId}/revoke`

Revoke a credential.

**Request Body:**
```json
{
  "reason": "No longer valid",
  "actor": "did:example:issuer123"
}
```

**Response (200):**
```json
{
  "success": true,
  "credential_id": "cred_123",
  "status": "revoked",
  "reason": "No longer valid",
  "tx_hash": "tx_abc123...",
  "timestamp": "2024-01-15T10:40:00Z",
  "request_id": "req-12345"
}
```

---

### Authorize App Access
**POST** `/credentials/{credentialId}/authorize-app`

Authorize an application to access a credential.

**Request Body:**
```json
{
  "app_id": "app_mobile_wallet",
  "app_name": "Mobile Credential Wallet",
  "permissions": ["read", "verify", "present"],
  "expires_in": 7776000
}
```

**Response (200):**
```json
{
  "success": true,
  "credential_id": "cred_123",
  "app_id": "app_mobile_wallet",
  "authorized": true,
  "permissions": ["read", "verify", "present"],
  "tx_hash": "tx_abc123...",
  "authorized_at": "2024-01-15T10:40:00Z",
  "expires_at": "2024-02-12T10:40:00Z",
  "request_id": "req-12345"
}
```

---

### Check App Access
**GET** `/credentials/{credentialId}/check-app-access?app_id=app_mobile_wallet`

Check if an app has access to a credential.

**Response (200):**
```json
{
  "success": true,
  "credential_id": "cred_123",
  "app_id": "app_mobile_wallet",
  "allowed": true,
  "request_id": "req-12345"
}
```

---

### Verify Credential Integrity
**POST** `/credentials/{credentialId}/verify`

Verify a credential hasn't been tampered with.

**Request Body:**
```json
{
  "content_hash": "abc123def456..."
}
```

**Response (200):**
```json
{
  "success": true,
  "credential_id": "cred_123",
  "valid": true,
  "request_id": "req-12345"
}
```

---

## TRUST LEVEL ENDPOINTS

### Get Current Trust Level
**GET** `/trust-level/{userId}`

Get a user's current trust level.

**Response (200):**
```json
{
  "success": true,
  "user_id": "user_12345",
  "level": "kyc-verified",
  "verification_method": "third_party_verification",
  "verified_by": "did:example:verifier123",
  "confidence": 0.90,
  "updated_at": "2024-01-15T10:30:00Z",
  "request_id": "req-12345"
}
```

---

### Update Trust Level
**PUT** `/trust-level/{userId}`

Update a user's trust level.

**Request Body:**
```json
{
  "level": "organization-verified",
  "verification_method": "organizational_verification",
  "verified_by": "did:example:org123",
  "confidence": 0.98,
  "reason": "Employee verification complete"
}
```

**Response (200):**
```json
{
  "success": true,
  "user_id": "user_12345",
  "level": "organization-verified",
  "verification_method": "organizational_verification",
  "confidence": 0.98,
  "tx_hash": "tx_abc123...",
  "updated_at": "2024-01-15T10:35:00Z",
  "request_id": "req-12345"
}
```

---

### Get Trust Level History
**GET** `/trust-level/{userId}/history`

Retrieve complete trust level history.

**Response (200):**
```json
{
  "success": true,
  "user_id": "user_12345",
  "current_level": "kyc-verified",
  "history_count": 3,
  "records": [
    {
      "level": "unverified",
      "verification_method": "self_certified",
      "confidence": 0.0,
      "updated_at": "2024-01-10T08:00:00Z"
    },
    {
      "level": "device-verified",
      "verification_method": "apple_digital_id",
      "confidence": 0.95,
      "updated_at": "2024-01-12T14:00:00Z"
    },
    {
      "level": "kyc-verified",
      "verification_method": "third_party_verification",
      "confidence": 0.90,
      "updated_at": "2024-01-15T10:30:00Z"
    }
  ],
  "request_id": "req-12345"
}
```

---

### Verify with Apple Digital ID
**POST** `/trust-level/{userId}/verify-apple`

Verify user with Apple Digital ID.

**Request Body:**
```json
{
  "apple_user_id": "apple_123abc...",
  "certification_details": {
    "device_id": "device_456def...",
    "verified_timestamp": "2024-01-15T10:30:00Z"
  }
}
```

**Response (200):**
```json
{
  "success": true,
  "user_id": "user_12345",
  "level": "device-verified",
  "verification_method": "apple_digital_id",
  "tx_hash": "tx_abc123...",
  "verified_at": "2024-01-15T10:30:00Z",
  "request_id": "req-12345"
}
```

---

### Verify with Third Party
**POST** `/trust-level/{userId}/verify-third-party`

Verify user with third-party verification service.

**Request Body:**
```json
{
  "verifier_did": "did:example:verifier123",
  "verification_data": {
    "kyc_level": "full",
    "document_type": "passport",
    "verified_timestamp": "2024-01-15T10:30:00Z"
  }
}
```

**Response (200):**
```json
{
  "success": true,
  "user_id": "user_12345",
  "level": "kyc-verified",
  "verification_method": "third_party_verification",
  "verifier": "did:example:verifier123",
  "tx_hash": "tx_abc123...",
  "verified_at": "2024-01-15T10:30:00Z",
  "request_id": "req-12345"
}
```

---

### Verify with Organization
**POST** `/trust-level/{userId}/verify-organization`

Verify user through organizational verification.

**Request Body:**
```json
{
  "organization_did": "did:example:employer123",
  "employee_id": "emp_12345"
}
```

**Response (200):**
```json
{
  "success": true,
  "user_id": "user_12345",
  "level": "organization-verified",
  "verification_method": "organizational_verification",
  "organization": "did:example:employer123",
  "tx_hash": "tx_abc123...",
  "verified_at": "2024-01-15T10:30:00Z",
  "request_id": "req-12345"
}
```

---

### Downgrade Trust Level
**POST** `/trust-level/{userId}/downgrade`

Downgrade a user's trust level.

**Request Body:**
```json
{
  "reason": "Failed compliance check",
  "actor": "did:example:admin123"
}
```

**Response (200):**
```json
{
  "success": true,
  "user_id": "user_12345",
  "action": "downgraded",
  "reason": "Failed compliance check",
  "tx_hash": "tx_abc123...",
  "downgraded_at": "2024-01-15T10:40:00Z",
  "request_id": "req-12345"
}
```

---

## TRANSACTION TRACKING ENDPOINTS

### Get Transaction Status
**GET** `/transactions/{txHash}`

Get the status of a specific transaction.

**Response (200):**
```json
{
  "success": true,
  "tx_hash": "tx_abc123...",
  "status": "confirmed",
  "operation": "register-issuer",
  "entity": "issuer",
  "created_at": "2024-01-15T10:30:00Z",
  "confirmed_at": "2024-01-15T10:31:00Z",
  "block_height": 8234567,
  "confirmations": 15,
  "request_id": "req-12345"
}
```

---

### Query Transactions
**GET** `/transactions?operation_type=register-issuer&status=confirmed`

Query transactions with filters.

**Query Parameters:**
- `tx_hash` (optional): Filter by transaction hash
- `status` (optional): Filter by status (pending, confirmed, failed, cancelled)
- `operation_type` (optional): Filter by operation type
- `related_entity` (optional): Filter by entity type
- `limit` (optional): Results per page
- `offset` (optional): Pagination offset

**Response (200):**
```json
{
  "success": true,
  "transaction_count": 42,
  "transactions": [...],
  "request_id": "req-12345"
}
```

---

### Get Pending Transactions
**GET** `/transactions/pending`

Get all pending transactions.

**Response (200):**
```json
{
  "success": true,
  "pending_count": 5,
  "transactions": [...],
  "request_id": "req-12345"
}
```

---

### Confirm Transaction
**POST** `/transactions/{txHash}/confirm`

Manually confirm a transaction.

**Request Body:**
```json
{
  "block_height": 8234567
}
```

**Response (200):**
```json
{
  "success": true,
  "tx_hash": "tx_abc123...",
  "status": "confirmed",
  "block_height": 8234567,
  "request_id": "req-12345"
}
```

---

### Retry Failed Transaction
**POST** `/transactions/{txHash}/retry`

Retry a failed transaction.

**Response (200):**
```json
{
  "success": true,
  "original_tx": "tx_abc123...",
  "new_tx_hash": "tx_xyz789...",
  "status": "pending",
  "request_id": "req-12345"
}
```

---

### Get Transaction Stats
**GET** `/transactions/stats`

Get transaction statistics.

**Response (200):**
```json
{
  "success": true,
  "stats": {
    "pending": 5,
    "confirmed": 1250,
    "failed": 3,
    "total": 1258,
    "oldest_pending": {
      "tx_hash": "tx_old...",
      "created_at": "2024-01-14T15:30:00Z"
    },
    "average_confirmation_time": "45.3s"
  },
  "request_id": "req-12345"
}
```

---

## HTTP Status Codes

| Code | Meaning |
|------|---------|
| 200 | OK - Request successful |
| 201 | Created - Resource created |
| 400 | Bad Request - Invalid input |
| 401 | Unauthorized - Authentication required |
| 403 | Forbidden - Access denied |
| 404 | Not Found - Resource doesn't exist |
| 409 | Conflict - Resource state conflict |
| 500 | Internal Server Error |
| 503 | Service Unavailable |

---

## Rate Limiting

- **Default:** 100 requests per minute per IP
- **Headers returned:**
  - `X-RateLimit-Limit`: Maximum requests allowed
  - `X-RateLimit-Remaining`: Requests remaining
  - `X-RateLimit-Reset`: Unix timestamp for limit reset

---

## Pagination

All list endpoints support pagination:
- `limit`: Results per page (default: 100, max: 1000)
- `offset`: Starting position (default: 0)

Response includes:
- `total`: Total number of items
- `count`: Items in current response
- `offset`: Current offset
- `limit`: Current limit

---

## Version Information

**API Version:** 1.0.0  
**Last Updated:** January 2024  
**Status:** Production Ready
