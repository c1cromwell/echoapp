# iOS OpenAPI Alignment - Implementation Guide

**Purpose**: Exact code changes needed to align iOS implementation with OpenAPI spec

---

## 1. Update Endpoints.swift

### 1.1 AuthEndpoint (CRITICAL)

**OLD:**
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

**NEW:**
```swift
enum AuthEndpoint: APIEndpoint {
    case register
    case challenge  // NEW
    case verify     // NEW (was implicitly in login)
    case refreshToken
    case logout
    
    // REMOVED: login, verifyBiometric, createPasskey, verifyPasskey
    // These don't exist in OpenAPI spec
```

**Path mapping:**
```swift
var path: String {
    switch self {
    case .register:
        return "/auth/register"
    case .challenge:      // NEW
        return "/auth/challenge"
    case .verify:         // NEW
        return "/auth/verify"
    case .refreshToken:
        return "/auth/refresh"
    case .logout:
        return "/auth/logout"
    }
}
```

---

### 1.2 ConversationEndpoint (NEW - separate for clarity)

**Current**: Mixed with MessageEndpoint
**New**: Separate endpoint enum

```swift
enum ConversationEndpoint: APIEndpoint {
    case list(cursor: String? = nil, limit: Int = 20)
    case create
    case get(conversationId: String)
    case delete(conversationId: String)
    
    var path: String {
        switch self {
        case .list(let cursor, let limit):
            var params = "limit=\(limit)"
            if let cursor = cursor {
                params += "&cursor=\(cursor)"
            }
            return "/conversations?\(params)"
        case .create:
            return "/conversations"
        case .get(let id):
            return "/conversations/\(id)"
        case .delete(let id):
            return "/conversations/\(id)"
        }
    }
    
    var method: HTTPMethod {
        switch self {
        case .list, .get:
            return .get
        case .create:
            return .post
        case .delete:
            return .delete
        }
    }
}
```

---

### 1.3 MessageEndpoint (UPDATED)

**OLD:**
```swift
enum MessageEndpoint: APIEndpoint {
    case send
    case fetch(conversationId: String, limit: Int = 50, offset: Int = 0)
    case searchMessages(query: String)
    case markAsRead(messageId: String)
    case deleteMessage(id: String)
    case editMessage(id: String)
    case addReaction(messageId: String)
    case removeReaction(messageId: String)
```

**NEW:**
```swift
enum MessageEndpoint: APIEndpoint {
    case sendMessage(conversationId: String)  // RENAMED from send
    case getMessages(conversationId: String, cursor: String? = nil, limit: Int = 50)  // RENAMED from fetch
    case getMessageDetails(messageId: String)  // NEW
    case markAsRead(messageId: String)
    case addReaction(messageId: String, reactionType: String)  // UPDATED
    case removeReaction(messageId: String, reactionType: String)  // NEW
    
    // REMOVED: searchMessages, deleteMessage, editMessage
    // These aren't in OpenAPI spec
    
    var path: String {
        switch self {
        case .sendMessage(let conversationId):
            return "/conversations/\(conversationId)/messages"
        case .getMessages(let conversationId, let cursor, let limit):
            var params = "limit=\(limit)"
            if let cursor = cursor {
                params += "&cursor=\(cursor)"
            }
            return "/conversations/\(conversationId)/messages?\(params)"
        case .getMessageDetails(let messageId):
            return "/messages/\(messageId)"
        case .markAsRead(let messageId):
            return "/messages/\(messageId)/read"
        case .addReaction(let messageId, let reactionType):
            return "/messages/\(messageId)/reactions"
        case .removeReaction(let messageId, let reactionType):
            return "/messages/\(messageId)/reactions"
        }
    }
    
    var method: HTTPMethod {
        switch self {
        case .sendMessage, .addReaction, .markAsRead:
            return .post
        case .getMessages, .getMessageDetails:
            return .get
        case .removeReaction:
            return .delete
        }
    }
}
```

---

### 1.4 UserEndpoint (UPDATED - use DID, not ID)

**OLD:**
```swift
enum UserEndpoint: APIEndpoint {
    case getProfile
    case updateProfile
    case getUser(id: String)          // Should be did
    case searchUsers(query: String)
    case addContact(id: String)        // Should be did
    case removeContact(id: String)     // Should be did
    case blockUser(id: String)         // Should be did
    case unblockUser(id: String)       // Should be did
    case getContacts(limit: Int = 50, offset: Int = 0)  // Pagination
```

**NEW:**
```swift
enum UserEndpoint: APIEndpoint {
    case getProfile
    case updateProfile
    case getUser(did: String)    // CHANGED: id → did
    case searchUsers(query: String)
    case uploadAvatar            // NEW
    case updateAccount           // NEW (was unclear before)
    
    // REMOVED: contact-related endpoints (see ContactEndpoint below)
    
    var path: String {
        switch self {
        case .getProfile:
            return "/users/profile"
        case .updateProfile:
            return "/users/profile"
        case .getUser(let did):
            return "/users/\(did)"
        case .searchUsers(let query):
            return "/users/search?q=\(query)"
        case .uploadAvatar:
            return "/users/avatar"
        case .updateAccount:
            return "/users/account"
        }
    }
}
```

---

### 1.5 ContactEndpoint (NEW - separate)

```swift
enum ContactEndpoint: APIEndpoint {
    case list(limit: Int = 20, cursor: String? = nil)
    case add(did: String)
    case get(did: String)
    case remove(did: String)
    case block(did: String)
    case unblock(did: String)
    
    var path: String {
        switch self {
        case .list(let limit, let cursor):
            var params = "limit=\(limit)"
            if let cursor = cursor {
                params += "&cursor=\(cursor)"
            }
            return "/contacts?\(params)"
        case .add(let did):
            return "/contacts"
        case .get(let did):
            return "/contacts/\(did)"
        case .remove(let did):
            return "/contacts/\(did)"
        case .block(let did):
            return "/contacts/\(did)/block"
        case .unblock(let did):
            return "/contacts/\(did)"
        }
    }
    
    var method: HTTPMethod {
        switch self {
        case .list, .get:
            return .get
        case .add, .block:
            return .post
        case .remove:
            return .delete
        case .unblock:
            return .delete
        }
    }
}
```

---

### 1.6 IdentityEndpoint (UPDATED)

**OLD:**
```swift
enum IdentityEndpoint: APIEndpoint {
    case createDID
    case resolveDID(did: String)
    case updateDIDDocument       // Not in spec
    case listDIDs                // Not in spec
    case verifyIdentity
    case getVerifications
    case addVerification         // Not in spec
    case revokeVerification      // Not in spec
    case createCredential
    case shareCredential         // Not in spec
    case verifyCredential(id: String)
    case revokeCredential(id: String)
```

**NEW:**
```swift
enum IdentityEndpoint: APIEndpoint {
    case createDID
    case getDID(did: String)
    case submitVerification          // RENAMED from verifyIdentity
    case getVerifications
    case getCredentials              // NEW
    case createCredential
    case verifyCredential(credentialId: String)
    case revokeCredential(credentialId: String)
    
    // REMOVED: updateDIDDocument, listDIDs, addVerification, revokeVerification, shareCredential
    
    var path: String {
        switch self {
        case .createDID:
            return "/identity/did"
        case .getDID(let did):
            return "/identity/did/\(did)"
        case .submitVerification:
            return "/identity/verify"
        case .getVerifications:
            return "/identity/verifications"
        case .getCredentials:
            return "/identity/credentials"
        case .createCredential:
            return "/identity/credentials"
        case .verifyCredential(let credentialId):
            return "/identity/credentials/\(credentialId)/verify"
        case .revokeCredential(let credentialId):
            return "/identity/credentials/\(credentialId)/revoke"
        }
    }
}
```

---

### 1.7 TokenEndpoint (REMOVE - split into separate enums)

**OLD:**
```swift
enum TokenEndpoint: APIEndpoint {
    case getBalance
    case getTransactionHistory(limit: Int = 50, offset: Int = 0)
    case sendTokens
    case stakeTokens
    case unstakeTokens
    case claimRewards
    case getStakingInfo          // Not in spec
    case getTrusScore            // TYPO! + Wrong path
    case getAchievements         // Not in spec
}
```

**NEW - SPLIT INTO 3 ENUMS:**

#### TokenEndpoint (for /tokens/*)
```swift
enum TokenEndpoint: APIEndpoint {
    case getBalance
    case getHistory(limit: Int = 20, cursor: String? = nil)
    case sendTokens
    case stakeTokens
    case unstakeTokens
    
    var path: String {
        switch self {
        case .getBalance:
            return "/tokens/balance"
        case .getHistory(let limit, let cursor):
            var params = "limit=\(limit)"
            if let cursor = cursor {
                params += "&cursor=\(cursor)"
            }
            return "/tokens/history?\(params)"
        case .sendTokens:
            return "/tokens/send"
        case .stakeTokens:
            return "/tokens/stake"
        case .unstakeTokens:
            return "/tokens/unstake"
        }
    }
}
```

#### TrustEndpoint (for /trust/*)
```swift
enum TrustEndpoint: APIEndpoint {
    case getTrustScore(userId: String)      // FIXED TYPO
    case submitVerification
    case reportUser
    case getVerificationStatus
    
    var path: String {
        switch self {
        case .getTrustScore(let userId):
            return "/trust/\(userId)/score"
        case .submitVerification:
            return "/trust/verify"
        case .reportUser:
            return "/trust/report"
        case .getVerificationStatus:
            return "/trust/verification/status"
        }
    }
}
```

#### RewardsEndpoint (for /rewards/*)
```swift
enum RewardsEndpoint: APIEndpoint {
    case getBalance
    case getActivity(limit: Int = 20, cursor: String? = nil)
    case claimRewards
    case getReferralCode
    case stakeTokens
    case unstakeTokens
    
    var path: String {
        switch self {
        case .getBalance:
            return "/rewards/balance"
        case .getActivity(let limit, let cursor):
            var params = "limit=\(limit)"
            if let cursor = cursor {
                params += "&cursor=\(cursor)"
            }
            return "/rewards/activity?\(params)"
        case .claimRewards:
            return "/rewards/claim"
        case .getReferralCode:
            return "/rewards/referral"
        case .stakeTokens:
            return "/rewards/stake"
        case .unstakeTokens:
            return "/rewards/unstake"
        }
    }
}
```

---

## 2. Update ViewModels.swift - Service Protocols

### 2.1 AuthServiceProtocol (CRITICAL)

**OLD:**
```swift
public protocol AuthServiceProtocol {
    func requestOTP(phone: String) async throws -> OTPResponse
    func verifyOTP(phone: String, code: String) async throws -> AuthResponse
    func registerPasskey() async throws
    func authenticateWithPasskey() async throws -> AuthResponse
    func refreshToken() async throws -> String
}
```

**NEW:**
```swift
public protocol AuthServiceProtocol {
    // Registration
    func register(publicKey: String, displayName: String?, deviceInfo: DeviceInfo) async throws -> AuthResponse
    
    // Authentication (challenge-response flow)
    func requestChallenge(did: String) async throws -> ChallengeResponse
    func verifyChallenge(did: String, challenge: String, signature: String) async throws -> AuthResponse
    
    // Token management
    func refreshToken(refreshToken: String) async throws -> AuthResponse
    func logout() async throws
}

// NEW: Supporting types
public struct DeviceInfo: Codable {
    public let deviceId: String
    public let deviceName: String
    public let osVersion: String
    public let appVersion: String
}

public struct ChallengeResponse: Codable {
    public let challenge: String  // base64url
    public let expiresAt: Date
}

public struct AuthResponse: Codable {
    public let accessToken: String
    public let refreshToken: String
    public let expiresIn: Int
    public let user: UserProfile
}
```

---

### 2.2 Remove OTPResponse

**DELETE:**
```swift
public struct OTPResponse: Codable {
    public let expiresIn: Int
    public let phone: String
}
```

This doesn't exist in OpenAPI spec. OTP auth is not the flow - it's challenge-response with passkey/biometric.

---

### 2.3 MessagingServiceProtocol (UPDATE)

**OLD:**
```swift
public protocol MessagingServiceProtocol {
    func fetchConversations() async throws -> [ConversationModel]
    func fetchMessages(conversationId: String) async throws -> [MessageModel]
    func sendMessage(_ content: String, to conversationId: String) async throws
    func markAsRead(conversationId: String) async throws
}
```

**NEW:**
```swift
public protocol MessagingServiceProtocol {
    // Conversations
    func fetchConversations(cursor: String?, limit: Int) async throws -> FetchConversationsResponse
    func createConversation(participantDIDs: [String], name: String?) async throws -> Conversation
    func getConversation(conversationId: String) async throws -> Conversation
    func deleteConversation(conversationId: String) async throws
    
    // Messages
    func getMessages(conversationId: String, cursor: String?, limit: Int) async throws -> FetchMessagesResponse
    func sendMessage(conversationId: String, message: SendMessageRequest) async throws -> MessageAccepted
    func getMessageDetails(messageId: String) async throws -> Message
    func markAsRead(messageId: String) async throws
    func addReaction(messageId: String, reactionType: String) async throws
    func removeReaction(messageId: String, reactionType: String) async throws
}

// NEW: Supporting types matching OpenAPI
public struct FetchConversationsResponse: Codable {
    public let conversations: [Conversation]
    public let nextCursor: String?
}

public struct FetchMessagesResponse: Codable {
    public let messages: [Message]
    public let nextCursor: String?
}

public struct Conversation: Codable, Identifiable {
    public let id: String
    public let participants: [String]  // DIDs
    public let lastMessage: Message?
    public let unreadCount: Int
    public let updatedAt: Date
    public let isGroup: Bool
    public let groupName: String?
}

public struct SendMessageRequest: Codable {
    public let contentType: String
    public let encryptedPayload: EncryptedPayload
    public let signature: String  // base64url
    public let replyToId: String?
    public let expiresIn: Int?  // Disappearing message TTL in seconds
}

public struct Message: Codable, Identifiable {
    public let id: String
    public let conversationId: String
    public let senderDID: String  // CHANGED from senderId
    public let contentType: String
    public let encryptedPayload: EncryptedPayload
    public let signature: String
    public let timestamp: Date
    public let status: String  // "sending", "sent", "delivered", "read", "failed", "anchored"
    public let expiresAt: Date?
    public let reactions: [Reaction]?
}

public struct EncryptedPayload: Codable {
    public let version: Int
    public let ephemeralPublicKey: String  // base64url
    public let nonce: String  // base64url
    public let ciphertext: String  // base64url
    public let tag: String  // base64url
    public let commitment: String  // base64url
}

public struct Reaction: Codable {
    public let userId: String
    public let reactionType: String
    public let timestamp: Date
}

public struct MessageAccepted: Codable {
    public let messageId: String
    public let status: String  // "relayed" or "queued"
    public let timestamp: Date
}
```

---

### 2.4 ContactServiceProtocol (NEW)

**NEW:**
```swift
public protocol ContactServiceProtocol {
    func listContacts(limit: Int, cursor: String?) async throws -> FetchContactsResponse
    func addContact(did: String) async throws -> Contact
    func getContact(did: String) async throws -> Contact
    func removeContact(did: String) async throws
    func blockContact(did: String) async throws
    func unblockContact(did: String) async throws
}

public struct FetchContactsResponse: Codable {
    public let contacts: [Contact]
    public let nextCursor: String?
}

public struct Contact: Codable, Identifiable {
    public let id: String  // Same as DID
    public let did: String
    public let displayName: String
    public let username: String?
    public let avatarURL: URL?
    public let trustScore: Int
    public let trustLevel: String
    public let isBlocked: Boolean
    public let lastSeen: Date?
}
```

---

### 2.5 TrustServiceProtocol (NEW - separate from rewards)

**UPDATE existing:**
```swift
public protocol TrustServiceProtocol {
    func getTrustScore(userId: String) async throws -> TrustScoreResult
    func submitVerification(documents: [URL], selfie: URL) async throws
    func reportUser(did: String, reason: String, evidence: [URL]?) async throws
    func getVerifications() async throws -> [VerificationRecord]
}

public struct TrustScoreResult: Codable {
    public let userId: String
    public let score: Int
    public let level: String
    public let breakdown: TrustBreakdown
    public let snapshotHash: String?  // For on-chain verification
}

public struct TrustBreakdown: Codable {
    public let identity: Int  // 0-30
    public let behavior: Int  // 0-25
    public let network: Int  // 0-25
    public let activity: Int  // 0-20
}

public struct VerificationRecord: Codable {
    public let id: String
    public let status: String  // "pending", "approved", "rejected"
    public let type: String  // "identity", "phone", etc.
    public let submittedAt: Date
    public let reviewedAt: Date?
}
```

---

### 2.6 RewardsServiceProtocol (NEW - separate from tokens)

**NEW:**
```swift
public protocol RewardsServiceProtocol {
    func getBalance() async throws -> RewardsBalance
    func getActivity(limit: Int, cursor: String?) async throws -> FetchActivityResponse
    func claimRewards() async throws -> ClaimResponse
    func getReferralCode() async throws -> ReferralInfo
    func stakeTokens(amount: Double, period: Int) async throws -> StakeResponse
    func unstakeTokens(amount: Double) async throws -> UnstakeResponse
}

public struct RewardsBalance: Codable {
    public let balance: Double
    public let currency: String  // "ECHO"
    public let lastUpdated: Date
}

public struct RewardActivityRecord: Codable, Identifiable {
    public let id: String
    public let type: String  // "messaging", "transaction", "referral", "staking", "achievement"
    public let amount: Double
    public let description: String
    public let earnedAt: Date
}

public struct FetchActivityResponse: Codable {
    public let activities: [RewardActivityRecord]
    public let nextCursor: String?
}

public struct ClaimResponse: Codable {
    public let claimedAmount: Double
    public let newBalance: Double
    public let transactionId: String
}

public struct ReferralInfo: Codable {
    public let code: String
    public let referralCount: Int
    public let earnings: Double
}

public struct StakeResponse: Codable {
    public let stakeId: String
    public let amount: Double
    public let period: Int  // days
    public let apy: Double  // Annual percentage yield
    public let estimatedRewards: Double
}

public struct UnstakeResponse: Codable {
    public let unstakeId: String
    public let amount: Double
    public let releaseDate: Date
}
```

---

## 3. Update Models.swift

### Add/Update Core Models

```swift
// UPDATE: User model to include DID
public struct UserProfile: Codable, Identifiable {
    public let id: String  // Same as DID
    public let did: String  // ADDED
    public let phone: String?  // Optional
    public let email: String?  // ADDED
    public let displayName: String?
    public let username: String?
    public let avatarURL: URL?
    public let bio: String?
    public let trustScore: Int?  // ADDED
    public let trustLevel: String?  // ADDED
    public let isVerified: Boolean?  // ADDED
    public let createdAt: Date?  // ADDED
}

// NEW: Device info
public struct DeviceInfo: Codable {
    public let deviceId: String
    public let deviceName: String
    public let osVersion: String
    public let appVersion: String
}

// UPDATE: Message model with encryption
public struct MessageModel: Codable, Identifiable {
    public let id: String
    public let conversationId: String
    public let senderDID: String  // CHANGED from senderId
    public let contentType: String  // text, image, video, etc.
    public let encryptedPayload: EncryptedPayload  // ADDED
    public let signature: String  // ADDED
    public let status: MessageStatus
    public let createdAt: Date
    public let timestamp: Date?  // ADDED
    public let expiresAt: Date?  // ADDED (disappearing messages)
}
```

---

## 4. Implementation Checklist

### Phase 5 Endpoint Alignment Tasks

- [ ] Update `AuthEndpoint` enum
- [ ] Create `ConversationEndpoint` enum  
- [ ] Update `MessageEndpoint` enum
- [ ] Update `UserEndpoint` enum (use DID)
- [ ] Create `ContactEndpoint` enum
- [ ] Update `IdentityEndpoint` enum
- [ ] Split `TokenEndpoint` into TokenEndpoint, TrustEndpoint, RewardsEndpoint
- [ ] Fix typo in `getTrusScore` → `getTrustScore`
- [ ] Update all path strings to match OpenAPI exactly
- [ ] Add HTTPMethod enum support for POST/PUT/DELETE
- [ ] Update cursor-based pagination (remove offset)

### Phase 5 Service Protocol Tasks

- [ ] Rewrite `AuthServiceProtocol` for challenge-response
- [ ] Update `MessagingServiceProtocol` with full OpenAPI methods
- [ ] Create `ContactServiceProtocol`
- [ ] Create separate `TrustServiceProtocol`
- [ ] Create separate `RewardsServiceProtocol`
- [ ] Remove `OTPResponse` and OTP-based auth
- [ ] Add all supporting types (ChallengeResponse, etc.)
- [ ] Update `ConversationModel` with all required fields
- [ ] Update `Message` model with DID and encryption
- [ ] Add `EncryptedPayload` model

### Phase 5 ViewModel Tasks

- [ ] Update `AuthViewModel` auth flow
- [ ] Remove OTP-based screen views (or refactor)
- [ ] Update `MessagingViewModel` method signatures
- [ ] Create `ContactViewModel`
- [ ] Create `TrustViewModel` (separate from rewards)
- [ ] Create `RewardsViewModel`
- [ ] Update all mock services

---

## Summary

This guide provides exact code changes needed to align iOS with OpenAPI spec:

✅ **Changes defined for:**
- 7 Endpoint enums (AuthEndpoint, ConversationEndpoint, MessageEndpoint, UserEndpoint, ContactEndpoint, IdentityEndpoint, and split Token/Trust/Rewards)
- 6 Service protocols (Auth, Messaging, Contact, Trust, Rewards + splits)
- Core models with all required fields
- Supporting types for requests/responses

❌ **Removed/deprecated:**
- OTP-based authentication
- Login endpoint
- Passkey/biometric endpoints
- Various endpoints not in OpenAPI spec
- ID parameters (changed to DID)
- Offset-based pagination (changed to cursor)

**Next step**: Implement these changes in Phase 5 for full API integration.
