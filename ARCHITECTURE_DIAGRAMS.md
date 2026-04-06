# ECHO Platform - Architecture Diagram

## High-Level System Architecture

```
┌────────────────────────────────────────────────────────────────┐
│                     iOS Application Layer                       │
│                  (SwiftUI + Presentation)                       │
└────────────────────────────────────────────────────────────────┘
                              ↓
┌────────────────────────────────────────────────────────────────┐
│                   Coordinator Layer (MVVM-C)                    │
│  ┌──────────────┬──────────────┬──────────────┬──────────────┐  │
│  │   Auth       │   Main       │   Settings   │   Chat       │  │
│  │ Coordinator  │ Coordinator  │ Coordinator  │ Coordinator  │  │
│  └──────────────┴──────────────┴──────────────┴──────────────┘  │
└────────────────────────────────────────────────────────────────┘
                              ↓
┌────────────────────────────────────────────────────────────────┐
│                   Use Case Layer (Business Logic)               │
│  ┌──────────────┬──────────────┬──────────────┬──────────────┐  │
│  │ Auth Use     │ Message Use  │ User Use     │ Token Use    │  │
│  │ Cases (7)    │ Cases (8)    │ Cases (8)    │ Cases (8)    │  │
│  └──────────────┴──────────────┴──────────────┴──────────────┘  │
└────────────────────────────────────────────────────────────────┘
                              ↓
┌────────────────────────────────────────────────────────────────┐
│                   Repository Layer (Data Abstraction)           │
│  ┌──────────────┬──────────────┬──────────────┬──────────────┐  │
│  │ Auth         │ User         │ Message      │ Token        │  │
│  │ Repository   │ Repository   │ Repository   │ Repository   │  │
│  └──────────────┴──────────────┴──────────────┴──────────────┘  │
└────────────────────────────────────────────────────────────────┘
         ↓                ↓                 ↓              ↓
    ┌─────────┐      ┌─────────┐      ┌─────────┐   ┌─────────┐
    │   API   │      │WebSocket│      │Database │   │Keychain │
    │ Client  │      │ Client  │      │ (Local) │   │(Secure) │
    └────┬────┘      └────┬────┘      └────┬────┘   └────┬────┘
         │                 │                 │             │
         └─────────────────┼─────────────────┴─────────────┘
                           ↓
            ┌──────────────────────────────┐
            │   Security & Encryption      │
            │ ┌──────────────────────────┐ │
            │ │ BiometricAuthManager     │ │
            │ │ KeychainManager          │ │
            │ │ KinnamiEncryption        │ │
            │ │ SecureEnclaveManager     │ │
            │ └──────────────────────────┘ │
            └──────────────────────────────┘
                           ↓
            ┌──────────────────────────────┐
            │   Backend Services (Phase 2) │
            │ ┌──────────────────────────┐ │
            │ │ Auth Service             │ │
            │ │ Trust Score Service      │ │
            │ │ Achievement Service      │ │
            │ │ Message Service          │ │
            │ │ Token Service            │ │
            │ └──────────────────────────┘ │
            └──────────────────────────────┘
```

## Component Interaction Flow

### Authentication Flow

```
┌────────────────┐
│  User Input    │ (Email, Password)
│  (Login View)  │
└────────┬───────┘
         │
         ↓
┌────────────────────────────────────────┐
│ AuthCoordinator                        │
│  - Navigate to login flow              │
└────────┬───────────────────────────────┘
         │
         ↓
┌────────────────────────────────────────┐
│ AuthenticateUseCase                    │
│  - Validate email/password             │
│  - Call repository                     │
└────────┬───────────────────────────────┘
         │
         ↓
┌────────────────────────────────────────┐
│ AuthRepository                         │
│  - Create API request                  │
│  - Call APIClient                      │
│  - Store tokens in Keychain            │
└────────┬───────────────────────────────┘
         │
         ↓
┌────────────────────────────────────────┐
│ APIClient + RequestInterceptor         │
│  - Add auth header                     │
│  - Call TLS 1.3+ endpoint              │
│  - Decrypt with Kinnami (if needed)    │
└────────┬───────────────────────────────┘
         │
         ↓
┌────────────────────────────────────────┐
│ Backend (Auth Service)                 │
│  - Verify credentials                  │
│  - Issue tokens                        │
│  - Return user data                    │
└────────┬───────────────────────────────┘
         │
         ↓
┌────────────────────────────────────────┐
│ KeychainManager                        │
│  - Store accessToken                   │
│  - Store refreshToken                  │
│  - Encrypt at rest                     │
└────────┬───────────────────────────────┘
         │
         ↓
┌────────────────────────────────────────┐
│ MainCoordinator                        │
│  - Navigate to main app                │
│  - Set isAuthenticated = true          │
└────────────────────────────────────────┘
```

### Message Encryption Flow

```
┌─────────────────┐
│  User Input     │ (Message Text)
│  (Chat View)    │
└────────┬────────┘
         │
         ↓
┌──────────────────────────────┐
│ SendMessageUseCase           │
│  - Get recipient's pubkey    │
│  - Call repository           │
└────────┬─────────────────────┘
         │
         ↓
┌──────────────────────────────────────────────┐
│ MessageRepository                            │
│  - Get recipient's public key                │
│  - Call KinnamiEncryption.encryptWithKeyAgreement()
└────────┬───────────────────────────────────┘
         │
         ↓
┌──────────────────────────────────────────────┐
│ KinnamiEncryption                            │
│  - Generate ephemeral key pair               │
│  - Perform ECDP-P256 key agreement           │
│  - Derive shared secret with HKDF            │
│  - Encrypt with AES-256-GCM                  │
│  - Return EncryptedMessage { nonce, cipher } │
└────────┬───────────────────────────────────┘
         │
         ↓
┌──────────────────────────────────────────────┐
│ SendMessageRequest                           │
│ {                                            │
│   conversationId: "conv-1",                  │
│   content: "Hello!",                         │
│   encryptedContent: "7K2h9...",              │
│   nonce: "abc123...",                        │
│   ephemeralPublicKey: "xyz789..."            │
│ }                                            │
└────────┬───────────────────────────────────┘
         │
         ↓
┌──────────────────────────────────────────────┐
│ APIClient                                    │
│  - Add auth header (from Keychain)           │
│  - POST to /messages/send                    │
│  - TLS 1.3+ encryption                       │
└────────┬───────────────────────────────────┘
         │
         ↓
┌──────────────────────────────────────────────┐
│ Backend (Message Service)                    │
│  - Validate signature                        │
│  - Store encrypted message                   │
│  - Relay to recipient via WebSocket          │
└────────┬───────────────────────────────────┘
         │
         ↓
┌──────────────────────────────────────────────┐
│ Recipient's Device                           │
│  - Receive encrypted message                 │
│  - Decrypt with own private key              │
│  - Display in ChatView                       │
└──────────────────────────────────────────────┘
```

### Real-Time Messaging Flow

```
┌──────────────────────────┐
│ WebSocketClient.connect()│
└────────┬─────────────────┘
         │
         ↓
┌──────────────────────────────────────┐
│ URLSessionWebSocket                  │
│  - Connect to wss://ws.echo.local    │
│  - TLS 1.3+ encryption               │
│  - Keep-alive heartbeat (30s)        │
└────────┬─────────────────────────────┘
         │
         ├─ Subscribe to channel      ┐
         │  /ws/messages              │
         │                            ├─ Control Protocol
         ├─ Send heartbeat ping (30s) │
         │                            │
         └─ Receive messages          ┘
                     │
                     ↓
         ┌──────────────────────┐
         │ Message received on  │
         │ WebSocket channel    │
         └──────┬───────────────┘
                │
                ├─ Text message
                │   └─ Encrypted content
                │       └─ Decrypt with KinnamiEncryption
                │
                ├─ Presence update
                │   └─ Update contact online status
                │
                └─ Notification
                    └─ Show banner
```

## Dependency Injection Graph

```
┌─────────────────────────────────────────────────────────┐
│                  DIContainer (Singleton)                 │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  Security Services:                                      │
│  ├─ SecureEnclaveManager → Singleton                    │
│  ├─ BiometricAuthManager → Factory                      │
│  ├─ KeychainManager → Singleton                         │
│  └─ KinnamiEncryption → Factory                         │
│                                                          │
│  Networking Services:                                    │
│  ├─ APIClient → Factory (configured)                    │
│  └─ WebSocketClient → Factory (configured)             │
│                                                          │
│  Storage Services:                                       │
│  ├─ LocalDatabase → Singleton                           │
│  ├─ SecureStorage → Singleton                           │
│  └─ CacheManager → Singleton                            │
│                                                          │
│  Repositories (dependent on services above):             │
│  ├─ AuthRepository                                       │
│  │   ├─ APIClient                                       │
│  │   ├─ KeychainManager                                 │
│  │   └─ SecureEnclaveManager                            │
│  │                                                       │
│  ├─ UserRepository                                       │
│  │   ├─ APIClient                                       │
│  │   └─ LocalDatabase                                   │
│  │                                                       │
│  ├─ MessageRepository                                    │
│  │   ├─ APIClient                                       │
│  │   ├─ WebSocketClient                                 │
│  │   ├─ KinnamiEncryption                               │
│  │   └─ LocalDatabase                                   │
│  │                                                       │
│  └─ TokenRepository                                      │
│      ├─ APIClient                                       │
│      └─ LocalDatabase                                   │
│                                                          │
│  Use Cases (dependent on repositories):                  │
│  ├─ AuthenticateUseCase(AuthRepository)                 │
│  ├─ SendMessageUseCase(MessageRepository)               │
│  ├─ GetBalanceUseCase(TokenRepository)                  │
│  └─ ... 27 more use cases ...                           │
│                                                          │
└─────────────────────────────────────────────────────────┘
         ↑
         │ Injected via
         │ DIContainer.resolve()
         │
    ┌────────────────────┐
    │  ViewModels        │
    │  Coordinators      │
    │  UseCases          │
    └────────────────────┘
```

## Data Flow Architecture

```
┌─────────────────┐
│   SwiftUI View  │
│  (Observable)   │
└────────┬────────┘
         │
         ↓
┌─────────────────────────────┐
│      ViewModel              │
│ - @Published State          │
│ - Input handlers            │
│ - Call UseCases             │
└────────┬────────────────────┘
         │
         ↓
┌─────────────────────────────┐
│      UseCase                │
│ - Validate input            │
│ - Call Repository           │
│ - Transform result          │
└────────┬────────────────────┘
         │
         ↓
┌─────────────────────────────┐
│      Repository             │
│ - Network or Local access   │
│ - Error handling            │
│ - Data transformation       │
└────────┬────────────────────┘
         │
     ┌───┼───┐
     │   │   │
     ↓   ↓   ↓
   API Web DB
   Req  Soc Ops
     │   │   │
     └───┼───┘
         │
         ↓
  Backend / Storage
```

## Security Layers

```
┌──────────────────────────────────────────────────────────┐
│                    Application Layer                      │
│  - SwiftUI Views                                         │
│  - Input validation                                      │
└──────────────────────────────────────────────────────────┘
         ↓
┌──────────────────────────────────────────────────────────┐
│              Business Logic Layer (UseCases)              │
│  - Authorization checks                                  │
│  - Data validation                                       │
└──────────────────────────────────────────────────────────┘
         ↓
┌──────────────────────────────────────────────────────────┐
│           Data Access Layer (Repositories)                │
│  - Request signing (if needed)                           │
│  - Response validation                                   │
└──────────────────────────────────────────────────────────┘
         ↓
┌──────────────────────────────────────────────────────────┐
│         Authentication & Authorization Layer              │
│  - BiometricAuthManager (Face/Touch ID)                  │
│  - KeychainManager (Encrypted storage)                   │
│  - Request Interceptor (Auth header injection)           │
└──────────────────────────────────────────────────────────┘
         ↓
┌──────────────────────────────────────────────────────────┐
│              Encryption Layer                             │
│  - KinnamiEncryption (E2E)                               │
│  - AES-256-GCM (Messages)                                │
│  - ECDP-P256 (Key Agreement)                             │
└──────────────────────────────────────────────────────────┘
         ↓
┌──────────────────────────────────────────────────────────┐
│           Transport Security Layer                        │
│  - URLSession (TLS 1.3+)                                 │
│  - WebSocket (WSS)                                       │
│  - Certificate validation                                │
└──────────────────────────────────────────────────────────┘
         ↓
┌──────────────────────────────────────────────────────────┐
│            Hardware Security Layer                        │
│  - Secure Enclave (Key storage)                          │
│  - Biometric sensor (Authentication)                     │
│  - Device Keychain (Encrypted credentials)               │
└──────────────────────────────────────────────────────────┘
```

## Model Relationships

```
User
├─ Profile
│  ├─ DIDs (decentralized identifiers)
│  ├─ Credentials (verifiable credentials)
│  ├─ TrustScore
│  │  └─ Level (Newcomer → Elite)
│  └─ Achievements
│     ├─ Badges (28 types)
│     └─ LevelBadges (Bronze → Platinum)
│
├─ Messaging
│  ├─ Conversations
│  │  ├─ Participants
│  │  ├─ Messages
│  │  │  ├─ Reactions
│  │  │  ├─ Attachments
│  │  │  └─ Encryption (Kinnami)
│  │  └─ UnreadCount
│  │
│  └─ Contacts
│     ├─ User data
│     ├─ OnlineStatus
│     └─ LastSeen
│
├─ Identity
│  ├─ DIDs
│  │  ├─ PublicKey
│  │  ├─ Proof
│  │  └─ VerificationMethod
│  │
│  └─ Verifications
│     ├─ Email
│     ├─ Phone
│     ├─ Identity
│     └─ Business
│
└─ Tokens
   ├─ Balance
   │  ├─ Available
   │  ├─ Frozen
   │  └─ Staked
   │
   ├─ Transactions
   │  ├─ Send
   │  ├─ Receive
   │  ├─ Stake
   │  └─ Rewards
   │
   ├─ TrustScore
   │  └─ Multiplier (1.0x → 2.0x)
   │
   └─ WebOfTrust
      ├─ Attestations (Vouch/Endorse/Verify)
      ├─ Confidence (1-10)
      └─ TrustBoost (up to 15 points)
```

## API Endpoint Topology

```
/auth                          Authentication
├─ POST /register
├─ POST /login
├─ POST /refresh
├─ POST /logout
├─ POST /verify-biometric
├─ POST /passkey/create
└─ POST /passkey/verify

/messages                      Messaging
├─ POST /send
├─ GET /conversations/{id}
├─ GET /conversations
├─ POST /conversations
├─ DELETE /{id}
├─ PUT /{id}
├─ PUT /{id}/read
├─ POST /{id}/reactions
└─ DELETE /{id}/reactions

/users                         User Management
├─ GET /profile
├─ PUT /profile
├─ GET /{id}
├─ GET /search
├─ POST /contacts/{id}
├─ DELETE /contacts/{id}
├─ GET /contacts
├─ POST /block/{id}
├─ DELETE /block/{id}
├─ POST /avatar
└─ DELETE /account

/identity                      Identity & Credentials
├─ POST /did/create
├─ GET /did/{did}
├─ PUT /did/update
├─ GET /did/list
├─ POST /verify
├─ GET /verifications
├─ POST /credentials/create
├─ POST /credentials/share
├─ GET /credentials/{id}/verify
└─ DELETE /credentials/{id}/revoke

/tokens                        Token & Rewards
├─ GET /balance
├─ GET /history
├─ POST /send
├─ POST /stake
├─ DELETE /unstake
├─ POST /rewards/claim
├─ GET /trust-score
└─ GET /achievements

/ws                            WebSocket Channels
├─ /messages                   Message streaming
├─ /notifications              Notifications
├─ /presence                   Online status
└─ /typing                     Typing indicators
```

---

This architecture provides:
- ✅ Clear separation of concerns
- ✅ Type-safe communication
- ✅ Enterprise-grade security
- ✅ Scalable design
- ✅ Easy testing
- ✅ High maintainability
