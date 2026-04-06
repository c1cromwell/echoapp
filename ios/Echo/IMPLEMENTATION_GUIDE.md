# ECHO iOS Frontend Architecture - Implementation Guide

## Overview

This document provides a comprehensive guide to the ECHO iOS frontend architecture implementation. The codebase follows MVVM-C (Model-View-ViewModel-Coordinator) pattern with clean architecture principles and production-ready security standards.

## Project Structure

```
ios/Echo/Sources/
├── Core/
│   ├── DI/
│   │   └── Container.swift              # Dependency injection container
│   ├── Security/
│   │   ├── BiometricAuthManager.swift   # Face ID / Touch ID integration
│   │   ├── KeychainManager.swift         # Encrypted credential storage
│   │   └── KinnamiEncryption.swift      # E2E message encryption
│   ├── Networking/
│   │   ├── APIClient.swift              # REST API client with interceptors
│   │   ├── WebSocketClient.swift        # Real-time messaging
│   │   └── Endpoints.swift              # API endpoint definitions
│   └── Storage/
│       └── LocalDatabase.swift          # SwiftData persistence layer
│
├── Domain/
│   ├── Models/
│   │   └── Models.swift                 # Domain models (User, Message, etc.)
│   ├── Repositories/
│   │   └── Repositories.swift           # Repository protocols & implementations
│   └── UseCases/
│       └── UseCases.swift               # Business logic use cases
│
└── Presentation/
    ├── Coordinators/
    │   └── Coordinators.swift           # Navigation coordinators
    ├── Components/
    │   └── Components.swift             # Reusable UI components
    └── Features/
        ├── Auth/                        # Authentication feature
        ├── Onboarding/                  # Onboarding flow
        ├── Conversations/               # Conversations list
        ├── Chat/                        # Chat messaging
        ├── Profile/                     # User profile
        └── Wallet/                      # Token wallet

Tests/
└── EchoTests.swift                      # Comprehensive test suite
```

## Core Layer

### Dependency Injection Container

**File:** `Core/DI/Container.swift`

The DI container manages all service instances and their lifecycles:

```swift
// Usage
let apiClient = DIContainer.shared.resolveAPIClient()
let keychain = DIContainer.shared.resolveKeychain()
let authRepository = DIContainer.shared.resolveAuthRepository()
```

**Features:**
- Service factory registration
- Singleton pattern for shared services
- Protocol-based dependency resolution
- Lazy initialization of dependencies

### Security Managers

#### BiometricAuthManager
**File:** `Core/Security/BiometricAuthManager.swift`

Provides Face ID / Touch ID authentication:

```swift
let bioManager = BiometricAuthManager()
let authenticated = await bioManager.authenticate(reason: "Verify your identity")
```

**Features:**
- Face ID, Touch ID, and Optic ID support
- Fallback to device passcode
- Device passcode verification
- Context invalidation for security

#### KeychainManager
**File:** `Core/Security/KeychainManager.swift`

Secure wrapper around iOS Keychain:

```swift
// Store token
try await keychain.storeAuthToken("secure-token")

// Retrieve token
let token = try await keychain.getAuthToken()

// Clear credentials
try await keychain.clearAuthCredentials()
```

**Features:**
- AES encryption with `kSecAttrAccessibleWhenUnlockedThisDeviceOnly`
- Dedicated services for auth, encryption, tokens
- Automatic credential cleanup
- Type-safe storage operations

#### KinnamiEncryption
**File:** `Core/Security/KinnamiEncryption.swift`

End-to-End encryption using Kinnami protocol:

```swift
// Generate key pair
let (privateKey, publicKey) = encryption.generateEphemeralKeyPair()

// Encrypt with key agreement
let encrypted = try await encryption.encryptWithKeyAgreement(
    plaintext: "Secret message",
    recipientPublicKeyData: recipientPublicKey
)

// Decrypt message
let decrypted = try await encryption.decryptWithKeyAgreement(
    encryptedMessage: encrypted,
    ourPrivateKey: privateKey
)
```

**Features:**
- ECDP-P256 key agreement
- AES-256-GCM encryption
- HKDF key derivation with SHA256
- Message signing and verification
- Authenticated encryption with nonce

### Networking Layer

#### APIClient
**File:** `Core/Networking/APIClient.swift`

RESTful API client with request/response handling:

```swift
// GET request
let user: User = try await apiClient.get(endpoint: UserEndpoint.getProfile)

// POST with body
let response = try await apiClient.post(
    endpoint: AuthEndpoint.login,
    body: loginRequest
)

// File upload
let result = try await apiClient.upload(
    endpoint: UserEndpoint.uploadAvatar,
    data: imageData,
    filename: "avatar.jpg",
    mimeType: "image/jpeg"
)
```

**Features:**
- TLS 1.3+ enforcement
- Request interceptor chain
- Authentication header injection
- Automatic error handling
- JSON encoding/decoding
- Response status validation
- URLSession configuration with timeouts

#### WebSocketClient
**File:** `Core/Networking/WebSocketClient.swift`

Real-time messaging over WebSocket:

```swift
// Connect
await wsClient.connect(delegate: self)

// Send message
try await wsClient.send(text: "Hello!")

// Receive (via delegate)
func webSocketDidReceiveMessage(_ client: WebSocketClient, message: String) {
    // Handle message
}

// Disconnect
await wsClient.disconnect()
```

**Features:**
- Automatic reconnection with exponential backoff
- Message queueing when disconnected
- Heartbeat mechanism every 30 seconds
- TLS 1.3+ encryption
- Control message protocol (ping/pong, subscribe/unsubscribe)

#### Endpoints
**File:** `Core/Networking/Endpoints.swift`

Comprehensive API endpoint definitions:

- **AuthEndpoint**: register, login, refreshToken, verifyBiometric, createPasskey
- **MessageEndpoint**: send, fetch, createConversation, reactions, search
- **UserEndpoint**: profile, contacts, search, block/unblock
- **IdentityEndpoint**: DID management, credentials, verifications
- **TokenEndpoint**: balance, transactions, staking, trust score
- **WebSocketEndpoint**: messaging, notifications, presence, typing

### Storage Layer

#### LocalDatabase
**File:** `Core/Storage/LocalDatabase.swift`

SwiftData persistence with type-safe models:

```swift
// Save
let user = LocalUser(id: "1", email: "test@echo.local", ...)
try await database.saveUser(user)

// Retrieve
let user = try await database.getUser(id: "1")

// Delete
try await database.deleteMessage(id: "msg-1")

// Clear all
try await database.clearAllData()
```

**Models:**
- `LocalUser` - User profiles
- `LocalMessage` - Message history
- `LocalConversation` - Conversation metadata
- `LocalContact` - Contact list
- `LocalDID` - Decentralized identifiers
- `LocalCredential` - VC credentials
- `LocalToken` - Token balance data
- `LocalAchievement` - Achievement tracking

**Features:**
- Thread-safe actor-based operations
- FetchDescriptor-based queries with filtering
- Pagination support with limit/offset
- Relationship preservation
- Automatic timestamp handling
- Cloud sync ready with CloudKit config

## Domain Layer

### Models
**File:** `Domain/Models/Models.swift`

Comprehensive domain models covering:

- **User** - User profile with trust score
- **Message** - End-to-end encrypted messages
- **Conversation** - Multi-user conversations
- **Contact** - Contact with online status
- **DID** - Decentralized identifiers
- **Credential** - Verifiable credentials
- **Token** - ECHO token balance and transactions
- **Achievement** - Gamification badges
- **TrustScore** - Trust level scoring (1-100)
- **WebOfTrustAttestation** - Peer attestations
- **Settings** - App configuration

All models implement:
- `Codable` for API serialization
- `Identifiable` for list rendering
- `Equatable` for comparisons
- Proper validation and defaults

### Repositories
**File:** `Domain/Repositories/Repositories.swift`

Protocol-based repository pattern:

```swift
// Repository protocols
protocol AuthRepository { ... }
protocol UserRepository { ... }
protocol MessageRepository { ... }
protocol TokenRepository { ... }

// Concrete implementations
actor ConcreteAuthRepository: AuthRepository { ... }
actor ConcreteUserRepository: UserRepository { ... }
actor ConcreteMessageRepository: MessageRepository { ... }
actor ConcreteTokenRepository: TokenRepository { ... }
```

**Responsibilities:**
- Network communication via APIClient
- WebSocket message streaming
- Local storage via LocalDatabase
- E2E encryption/decryption
- Error handling and retry logic
- Data transformation

### Use Cases
**File:** `Domain/UseCases/UseCases.swift`

20+ business logic implementations:

```swift
// Authentication
AuthenticateUseCase, RegisterUseCase, CreatePasskeyUseCase

// Messaging
SendMessageUseCase, FetchMessagesUseCase, EditMessageUseCase,
DeleteMessageUseCase, MarkAsReadUseCase, AddReactionUseCase

// User Management
GetProfileUseCase, UpdateProfileUseCase, SearchUsersUseCase,
AddContactUseCase, GetContactsUseCase, UploadAvatarUseCase

// Identity
CreateDIDUseCase, VerifyIdentityUseCase

// Tokens & Rewards
GetBalanceUseCase, SendTokensUseCase, StakeTokensUseCase,
GetTransactionHistoryUseCase, GetTrustScoreUseCase,
GetAchievementsUseCase
```

## Presentation Layer

### Coordinators
**File:** `Presentation/Coordinators/Coordinators.swift`

Navigation management with MVVM-C pattern:

```swift
// App-level coordinator
let appCoordinator = AppCoordinator()
appCoordinator.start()
appCoordinator.navigate(to: .main)

// Feature coordinators
let authCoordinator = AuthCoordinator()
authCoordinator.showOnboarding()

let chatCoordinator = ChatCoordinator(conversationId: "123")
chatCoordinator.start()
```

**Coordinators:**
- **AppCoordinator** - Root navigation, authentication state
- **AuthCoordinator** - Login/Register flows
- **MainCoordinator** - Tab navigation, main app flows
- **SettingsCoordinator** - Settings and account management
- **ChatCoordinator** - Conversation-specific navigation
- **DeepLinkHandler** - URL scheme handling

### Reusable Components
**File:** `Presentation/Components/Components.swift`

20+ production-ready UI components:

```swift
// Buttons
PrimaryButton(title: "Login", action: { ... })
SecondaryButton(title: "Cancel", action: { ... })

// Input
TextInputField(label: "Email", placeholder: "user@echo.local", text: $email)

// Displays
TrustBadge(level: .verified)
VerificationBadge(type: .email, status: .verified)
AchievementCard(achievement: achievement)

// Messages
MessageBubble(message: message, isCurrentUser: true)

// Lists
UserListItem(user: user, isOnline: true, onTap: { ... })
ConversationCell(conversation: conversation)

// States
LoadingIndicator()
EmptyStateView(icon: "envelope.open", title: "No Messages")
ErrorBanner(message: "Something went wrong")
```

### Feature Modules

Each feature module includes:
- **Views** - SwiftUI views for the feature
- **ViewModels** - State management and business logic
- **Navigation** - Feature-specific routing

## Security Architecture

### Key Security Features

1. **Secure Enclave Integration**
   - Master key storage in hardware
   - Biometric-gated key access
   - ECDSA-P256 key generation

2. **Biometric Authentication**
   - Face ID and Touch ID support
   - Fallback to device passcode
   - Cached authentication for UX

3. **Keychain Protection**
   - All tokens encrypted at rest
   - `WhenUnlockedThisDeviceOnly` access level
   - Automatic cleanup on logout

4. **E2E Encryption**
   - Kinnami protocol implementation
   - AES-256-GCM for message encryption
   - Ephemeral key agreement per session
   - Signature verification for authenticity

5. **Network Security**
   - TLS 1.3+ enforced
   - Certificate pinning ready
   - Request signing support
   - Encrypted WebSocket for real-time messaging

6. **Privacy-Preserving Logging**
   - No sensitive data in logs
   - Structured logging with privacy categories
   - Audit trail for security events

## Testing

### Test Coverage

**File:** `Tests/EchoTests.swift`

Comprehensive test suite with:

- **Unit Tests**: 30+ tests for individual components
- **Integration Tests**: Auth flow, message encryption
- **Performance Tests**: Encryption, key generation benchmarks
- **Mock Objects**: Mock repositories for testing

### Running Tests

```bash
# Run all tests
xcodebuild test -scheme Echo

# Run specific test class
xcodebuild test -scheme Echo -only EchoTests/BiometricAuthManagerTests

# With coverage
xcodebuild test -scheme Echo -enableCodeCoverage YES
```

## Architecture Patterns

### MVVM-C
- **Model**: Domain models
- **View**: SwiftUI components
- **ViewModel**: State, input handlers, business logic
- **Coordinator**: Navigation between features

### Dependency Injection
- DIContainer manages all service creation
- Protocol-based abstractions
- Factory pattern for flexible initialization

### Repository Pattern
- Repository protocols define data access contracts
- Concrete implementations handle API/Storage/WebSocket
- Repositories decouple domain from infrastructure

### Use Case Pattern
- Single Responsibility Principle
- Pure business logic independent of UI
- Easy to test and reuse

## Configuration

### API Base URL
Edit `Core/Networking/APIClient.swift`:

```swift
static let `default` = APIConfiguration(
    baseURL: URL(string: "https://api.echo.local")!,
    ...
)
```

### WebSocket URL
Edit `Core/Networking/WebSocketClient.swift`:

```swift
static let `default` = WebSocketConfiguration(
    baseURL: URL(string: "wss://ws.echo.local")!,
    ...
)
```

### Security Settings
Edit `BiometricAuthManager.swift` for biometric policies
Edit `KeychainManager.swift` for storage access levels

## Performance Optimization

1. **Lazy Loading**: View models load data on demand
2. **Caching**: CacheManager for frequently accessed data
3. **Background Operations**: Image processing, file uploads
4. **Memory Management**: Proper cleanup in deinit blocks
5. **Network Optimization**: Request batching, compression

## Future Enhancements

- [ ] Feature flags for A/B testing
- [ ] Offline mode with sync queue
- [ ] Analytics integration
- [ ] Push notifications
- [ ] Voice/Video calling
- [ ] Document sharing
- [ ] Advanced search with Spotlight integration
- [ ] Widget support
- [ ] Watch app companion
- [ ] Custom keyboard

## Dependencies

### Framework Dependencies
- SwiftUI (iOS 15+)
- Combine for reactive programming
- SwiftData for persistence
- CryptoKit for cryptography
- Security framework for Keychain
- LocalAuthentication for biometrics

### No External Package Dependencies
The implementation uses only Apple frameworks to minimize dependencies and improve security.

## Code Quality

- **Swift 5.9+** with strict syntax
- **Async/await** for concurrency
- **Actor** for thread safety
- **Type-safe** throughout
- **Error handling** with typed errors
- **Protocol-oriented** design
- **SwiftUI** best practices
- **Memory safety** guarantees

## Deployment

### Build Configurations
- Debug: Full logging, optimization disabled
- Release: Production-optimized, logging disabled

### Version Management
- Semantic versioning (MAJOR.MINOR.PATCH)
- Changelog tracking
- Build number auto-increment

## Support & Troubleshooting

### Common Issues

**Biometric Not Available**
- Check device has Face/Touch ID capability
- Verify biometric enrollment in Settings
- Test with simulator (simulates biometrics)

**Keychain Access Denied**
- Verify app has Keychain sharing entitlements
- Check device locked status
- Verify Keychain Sharing capability enabled

**WebSocket Connection Fails**
- Check WebSocket URL configuration
- Verify TLS certificate validity
- Check network connectivity
- Verify firewall doesn't block WebSocket

## Documentation

See additional documentation:
- [iOS Blueprint Architecture](ios-frontend-architecture-blueprint-v2.md)
- [Security Implementation Guide](../SECURITY_GUIDE.md)
- [API Reference](../API_REFERENCE.md)
- [Testing Guide](../TESTING.md)

## Contributors

ECHO iOS Team

## License

See LICENSE file in repository root.
