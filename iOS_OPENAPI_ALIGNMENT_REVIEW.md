# iOS Codebase & OpenAPI Spec Alignment Review

**Date:** February 23, 2026
**Status:** ⚠️ REVIEW REQUIRED - Several alignments needed

---

## Executive Summary

The iOS implementation (Phase 4) creates a comprehensive UI layer and ViewModels, but we need to **align the Endpoint definitions and Service protocols** with the OpenAPI specification. The OpenAPI spec defines the actual API contract; our Endpoints and Services must match exactly.

**Key Findings:**
- ✅ Architecture is solid (MVVM-C, protocol-based services)
- ✅ UI components are complete and accessible
- ⚠️ **Endpoint definitions have inconsistencies with OpenAPI spec**
- ⚠️ **Service protocols need refinement to match API contract**
- ❌ **Authentication flow endpoints don't match OpenAPI**
- ❌ **Missing some OpenAPI endpoints in Endpoints enum**

---

## 1. Authentication Endpoints Misalignment

### Current Implementation (Endpoints.swift)
```swift
enum AuthEndpoint: APIEndpoint {
    case register
    case login
    case refreshToken
    case logout
    case verifyBiometric
    case createPasskey
    case verifyPasskey
```

### OpenAPI Spec
```yaml
/auth/register:
  POST - Register new account (publicKey, deviceInfo)

/auth/challenge:
  POST - Request authentication challenge (did)

/auth/verify:
  POST - Verify signed challenge (did, challenge, signature)

/auth/refresh:
  POST - Refresh access token (refreshToken)

/auth/logout:
  POST - Logout and invalidate tokens
```

### Issue
- ❌ No `/auth/challenge` endpoint (required for challenge-response auth)
- ❌ No `/auth/verify` endpoint (actual verification endpoint)
- ❌ `login` endpoint doesn't exist in OpenAPI spec
- ❌ `verifyBiometric`, `createPasskey`, `verifyPasskey` are custom, not in spec
- ✅ `refreshToken`, `logout` are correct

### Required Fix
Replace with:
```swift
enum AuthEndpoint: APIEndpoint {
    case register
    case challenge
    case verify
    case refreshToken
    case logout
```

---

## 2. Conversation Endpoints

### Current Implementation
```swift
case fetch(conversationId: String, limit: Int = 50, offset: Int = 0)
case fetchConversations(limit: Int = 20, offset: Int = 0)
case createConversation
case getConversation(id: String)
```

### OpenAPI Spec
```yaml
GET /conversations - List conversations (with cursor pagination)
POST /conversations - Create conversation
GET /conversations/{conversationId} - Get conversation details
DELETE /conversations/{conversationId} - Leave conversation
```

### Issue
- ⚠️ Using `offset` pagination instead of OpenAPI's `cursor` pagination
- ✅ Path structure is mostly correct

### Required Fix
Update pagination to use cursor-based instead of offset:
```swift
case fetchConversations(cursor: String? = nil, limit: Int = 20)
case fetch(conversationId: String, cursor: String? = nil, limit: Int = 50)
```

---

## 3. Message Endpoints

### Current Implementation
```swift
case send
case fetch(conversationId: String)
case markAsRead(messageId: String)
case deleteMessage(id: String)
case editMessage(id: String)
case addReaction(messageId: String)
case removeReaction(messageId: String)
```

### OpenAPI Spec
```yaml
GET /conversations/{conversationId}/messages - Fetch messages
POST /conversations/{conversationId}/messages - Send message
GET /messages/{messageId} - Get message details
POST /messages/{messageId}/read - Mark as read
POST /messages/{messageId}/reactions - Add/remove reactions
```

### Issue
- ⚠️ `send` path is `/messages/send` but spec shows `POST /conversations/{conversationId}/messages`
- ❌ Missing message details endpoint `/messages/{messageId}`
- ❌ `deleteMessage` and `editMessage` not in OpenAPI spec
- ❌ Reactions endpoint structure unclear

### Required Fix
Realign message endpoints:
```swift
case sendMessage(conversationId: String)
case getMessages(conversationId: String, cursor: String? = nil)
case getMessageDetails(messageId: String)
case markAsRead(messageId: String)
case addReaction(messageId: String)
case removeReaction(messageId: String)
```

---

## 4. User & Contact Endpoints

### Current Implementation
```swift
case getProfile
case updateProfile
case getUser(id: String)
case searchUsers(query: String)
case addContact(id: String)
case removeContact(id: String)
case blockUser(id: String)
case unblockUser(id: String)
case getContacts(limit: Int = 50, offset: Int = 0)
```

### OpenAPI Spec
```yaml
GET /users/profile - Get user profile
PUT /users/profile - Update profile
GET /users/{did} - Get user by DID
GET /users/search - Search users
POST /users/avatar - Upload avatar
PUT /users/account - Update account settings
GET /contacts - List contacts
POST /contacts - Add contact
GET /contacts/{did} - Get contact details
DELETE /contacts/{did} - Remove contact
POST /contacts/{did}/block - Block user
```

### Issue
- ⚠️ Using `id` instead of `did` (Cardano DID format)
- ⚠️ Avatar upload should be separate endpoint
- ⚠️ Contact endpoints use `/contacts` path, not under `/users/`
- ❌ Contact structure in OpenAPI is different

### Required Fix
```swift
case getProfile
case updateProfile
case getUser(did: String)  // DID, not ID
case searchUsers(query: String)
case uploadAvatar
case updateAccount
// Separate ContactEndpoint enum:
enum ContactEndpoint: APIEndpoint {
    case getContacts(limit: Int = 20, cursor: String? = nil)
    case addContact(did: String)
    case getContact(did: String)
    case removeContact(did: String)
    case blockContact(did: String)
}
```

---

## 5. Identity Endpoints

### Current Implementation
```swift
case createDID
case resolveDID(did: String)
case updateDIDDocument
case listDIDs
case verifyIdentity
case getVerifications
case addVerification
case revokeVerification
case createCredential
case shareCredential
case verifyCredential(id: String)
case revokeCredential(id: String)
```

### OpenAPI Spec
```yaml
POST /identity/did - Create DID
GET /identity/did/{did} - Get DID details
POST /identity/verify - Submit verification
GET /identity/verifications - List verifications
POST /identity/credentials - Create credential
GET /identity/credentials - List credentials
POST /identity/credentials/{credentialId}/verify - Verify credential
POST /identity/credentials/{credentialId}/revoke - Revoke credential
```

### Issue
- ⚠️ Some endpoints don't match OpenAPI paths
- ❌ `updateDIDDocument`, `listDIDs` not in spec
- ❌ `shareCredential` not in spec

### Required Fix
```swift
case createDID
case getDID(did: String)
case submitVerification
case getVerifications
case getCredentials
case createCredential
case verifyCredential(credentialId: String)
case revokeCredential(credentialId: String)
```

---

## 6. Token Endpoints

### Current Implementation
```swift
case getBalance
case getTransactionHistory(limit: Int = 50, offset: Int = 0)
case sendTokens
case stakeTokens
case unstakeTokens
case claimRewards
case getStakingInfo
case getTrusScore  // NOTE: Typo! "TrusScore"
case getAchievements
```

### OpenAPI Spec
```yaml
GET /tokens/balance - Get token balance
GET /tokens/history - Get transaction history
POST /tokens/send - Send tokens
POST /tokens/stake - Stake tokens
POST /tokens/unstake - Unstake tokens
GET /trust/{userId}/score - Get trust score (DIFFERENT PATH)
POST /trust/report - Report user (DIFFERENT PATH)
GET /rewards/balance - Get rewards balance (DIFFERENT PATH)
GET /rewards/activity - Get activity
POST /rewards/claim - Claim rewards
GET /rewards/referral - Get referral code
```

### Issue
- ⚠️ Trust and Rewards endpoints are in **separate `/trust` and `/rewards` paths**, not `/tokens`
- ❌ Typo: `getTrusScore` should be `getTrustScore`
- ❌ `getStakingInfo` not in spec
- ❌ `getAchievements` not in spec

### Required Fix
Create separate endpoint enums:
```swift
// TokenEndpoint for /tokens/*
enum TokenEndpoint: APIEndpoint {
    case getBalance
    case getHistory(limit: Int = 20, cursor: String? = nil)
    case sendTokens
    case stakeTokens
    case unstakeTokens
}

// TrustEndpoint for /trust/*
enum TrustEndpoint: APIEndpoint {
    case getTrustScore(userId: String)
    case submitVerification
    case reportUser
    case getVerificationStatus
}

// RewardsEndpoint for /rewards/*
enum RewardsEndpoint: APIEndpoint {
    case getBalance
    case getActivity
    case claimRewards
    case getReferralCode
    case stakeTokens
    case unstakeTokens
}
```

---

## 7. Missing Endpoint Groups

### OpenAPI Has But Endpoints Enum Doesn't
- ❌ **Groups**: `/groups/*` endpoints not implemented
- ❌ **WebSocket**: Real-time messaging spec not implemented
- ❌ **Device Management**: Device info endpoints

### From OpenAPI Spec (Partial)
```yaml
/groups - Create group
/groups/{groupId} - Get group details
/groups/{groupId}/members - Manage members
/groups/{groupId}/keys - Group key rotation
```

---

## 8. Service Protocols Alignment

### Current ViewModels (ViewModels.swift)
```swift
public protocol AuthServiceProtocol {
    func requestOTP(phone: String) async throws -> OTPResponse
    func verifyOTP(phone: String, code: String) async throws -> AuthResponse
    func registerPasskey() async throws
    func authenticateWithPasskey() async throws -> AuthResponse
    func refreshToken() async throws -> String
}
```

### Issue
- ❌ OTP-based auth is **not in OpenAPI spec**
- ✅ Passkey/biometric in spec, but implementation unclear
- ❌ Should match OpenAPI: challenge → verify → access token

### Required Fix
```swift
public protocol AuthServiceProtocol {
    func register(publicKey: String, displayName: String?, deviceInfo: DeviceInfo) async throws -> AuthResponse
    func requestChallenge(did: String) async throws -> ChallengeResponse
    func verifyChallenge(did: String, challenge: String, signature: String) async throws -> AuthResponse
    func refreshToken(refreshToken: String) async throws -> AuthResponse
    func logout() async throws
}
```

---

## 9. Request/Response Model Alignment

### Current Models (incomplete)
- ✅ Basic User, Contact, Message models exist
- ❌ Models don't include OpenAPI-specified fields:
  - Missing `did` (Decentralized Identifier)
  - Missing `snapshotHash` (blockchain verification)
  - Missing encryption metadata (`encryptedPayload`)
  - Missing `expiresAt` (disappearing messages)

### Required Models
Need to create/update:
```swift
// From OpenAPI spec
struct AuthResponse {
    let accessToken: String
    let refreshToken: String
    let expiresIn: Int
    let user: UserProfile
}

struct ChallengeResponse {
    let challenge: String  // base64url
    let expiresAt: Date
}

struct Message {
    let id: String
    let conversationId: String
    let senderDID: String  // DID, not ID
    let contentType: String  // enum
    let encryptedPayload: EncryptedPayload  // Kinnami format
    let signature: String  // base64url
    let status: MessageStatus  // enum
    let expiresAt: Date?  // nullable
}

struct EncryptedPayload {
    let version: Int
    let ephemeralPublicKey: String  // base64url
    let nonce: String  // base64url
    let ciphertext: String  // base64url
    let tag: String  // base64url
    let commitment: String  // base64url
}
```

---

## 10. HTTP Methods Misalignment

### Current Implementation (check paths)
Need to verify:
- GET vs POST usage
- PUT vs PATCH vs POST for updates
- DELETE for removal

### OpenAPI Clearly Specifies
```yaml
GET /users/profile - retrieve
PUT /users/profile - update
POST /users/avatar - upload
DELETE /contacts/{did} - remove
POST /messages/{messageId}/reactions - add (not PATCH)
```

---

## Priority Alignment Items (High → Low)

### 🔴 CRITICAL (Block Phase 5)
1. **Fix AuthEndpoint** - Challenge/verify flow doesn't match spec
2. **Separate Token/Trust/Rewards** - Currently mixed in TokenEndpoint
3. **Update Message Endpoint** - Use conversation-based path
4. **Update Service Protocols** - AuthServiceProtocol flow is wrong

### 🟠 HIGH (Important for API Integration)
5. **Change ID to DID** - Use Decentralized Identifiers everywhere
6. **Add pagination cursor** - Replace offset with cursor
7. **Add missing models** - EncryptedPayload, ChallengeResponse, etc.
8. **HTTP methods** - Verify GET/POST/PUT/DELETE usage

### 🟡 MEDIUM (Polish)
9. **Fix typo** - `getTrusScore` → `getTrustScore`
10. **Add Group endpoints** - For group messaging
11. **WebSocket spec** - Real-time messaging
12. **Device Management** - Device info endpoints

### 🟢 LOW (Future)
13. **Disappearing messages** - `expiresAt` field
14. **Message reactions** - Full implementation
15. **On-chain anchoring** - `snapshotHash` verification

---

## Recommended Action Plan

### Phase 5: API Integration (Updated)

**Sprint 1: Fix Core Authentication**
- [ ] Update `AuthEndpoint` to match challenge/verify flow
- [ ] Update `AuthServiceProtocol` with correct methods
- [ ] Update `AuthViewModel` to call challenge endpoint first
- [ ] Add `ChallengeResponse` model
- [ ] Remove OTP-based authentication (not in spec)

**Sprint 2: Reorganize Endpoints**
- [ ] Separate `TokenEndpoint`, `TrustEndpoint`, `RewardsEndpoint`
- [ ] Update all message endpoints to use conversation paths
- [ ] Change all `id` parameters to `did`
- [ ] Implement cursor-based pagination

**Sprint 3: Update Service Protocols**
- [ ] Rewrite `AuthServiceProtocol`
- [ ] Create `ContactServiceProtocol`
- [ ] Create `TrustServiceProtocol` (separate from tokens)
- [ ] Create `RewardsServiceProtocol` (separate from tokens)

**Sprint 4: Implement Real Services**
- [ ] Implement `AuthService` calling real endpoints
- [ ] Implement `MessagingService` with Kinnami encryption
- [ ] Implement `TrustService` with verification flow
- [ ] Implement `RewardsService`

**Sprint 5: Testing & Integration**
- [ ] Update all tests with new service methods
- [ ] Integration tests against OpenAPI spec
- [ ] Test with mock server or staging environment

---

## File Changes Summary

### Files to Update
1. **`ios/Echo/Sources/Core/Networking/Endpoints.swift`** (CRITICAL)
   - Rewrite endpoint definitions
   - Add missing endpoint groups
   - Fix paths to match OpenAPI exactly

2. **`ios/Echo/Sources/Presentation/ViewModels/ViewModels.swift`** (CRITICAL)
   - Update `AuthServiceProtocol`
   - Add `ContactServiceProtocol`
   - Separate `TrustServiceProtocol`
   - Separate `RewardsServiceProtocol`

3. **`ios/Echo/Sources/Domain/Models/Models.swift`** (HIGH)
   - Add missing models (ChallengeResponse, EncryptedPayload, etc.)
   - Update User model with `did` field
   - Update Message model with encryption fields

4. **`ios/Echo/Sources/Presentation/Screens/**`** (MEDIUM)
   - Update ViewModels to match new service protocols
   - Remove OTP-based flow from AuthScreens
   - Add challenge-response visualization

---

## Questions for Clarification

1. **DID Format**: Should we validate DID format (e.g., `did:prism:...`)?
2. **Encryption**: Should Kinnami encryption be in a separate service or in MessagingService?
3. **Staging vs Production**: How do we handle API base URL switching?
4. **Token Management**: Should we cache tokens or always use keychain?
5. **WebSocket**: Should Phase 5 include real-time messaging WebSocket?

---

## Conclusion

The iOS implementation has **excellent architecture and UI**, but the **endpoint and service definitions need realignment with the OpenAPI specification**. This is critical before Phase 5 API integration to ensure the iOS app correctly implements the ECHO backend contract.

**Next Step**: Review this alignment document and decide which items to include in Phase 5 sprint planning.
