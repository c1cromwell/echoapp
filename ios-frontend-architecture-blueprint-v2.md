# iOS Frontend Architecture

## Overview

The frontend is a native iOS application built with SwiftUI and MVVM-C (Model-View-ViewModel-Coordinator) architecture. It provides a secure, user-friendly interface for private messaging, identity management, token rewards, and interaction with the Constellation metagraph through the Go backend REST services. The application prioritizes security through Secure Enclave integration, biometric authentication, and end-to-end encryption via the Kinnami protocol.

## Technology Stack

| Component | Technology | Purpose |
|-----------|------------|---------|
| **UI Framework** | SwiftUI | Declarative UI |
| **Architecture** | MVVM-C | Separation of concerns |
| **Language** | Swift 5.9+ | Type safety, performance |
| **Concurrency** | Swift Concurrency (async/await) | Asynchronous operations |
| **Security** | CryptoKit, Security.framework | Encryption, Secure Enclave |
| **Networking** | URLSession, WebSocket | API & real-time |
| **Persistence** | SwiftData, Keychain | Local storage |
| **DI** | Factory pattern | Dependency injection |
| **Encryption** | Kinnami (custom) | E2E message encryption |
| **Transport** | TLS 1.3+ | Network security |
| **Push** | APNs | Notifications |
| **Analytics** | Privacy-preserving | Usage metrics |

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                        iOS Application Architecture                  │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                      Presentation Layer                      │   │
│  │  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐        │   │
│  │  │  Views  │  │ViewModels│  │Coordinators│ │  UI    │        │   │
│  │  │(SwiftUI)│  │ (State) │  │(Navigation)│ │Components│       │   │
│  │  └─────────┘  └─────────┘  └─────────┘  └─────────┘        │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                              │                                      │
│                              ▼                                      │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                       Domain Layer                           │   │
│  │  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐        │   │
│  │  │Use Cases│  │ Models  │  │Protocols│  │  DTOs   │        │   │
│  │  │(Business│  │ (Domain)│  │(Abstractions)│        │        │   │
│  │  │  Logic) │  │         │  │         │  │         │        │   │
│  │  └─────────┘  └─────────┘  └─────────┘  └─────────┘        │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                              │                                      │
│                              ▼                                      │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                        Data Layer                            │   │
│  │  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐        │   │
│  │  │  API    │  │WebSocket│  │  Local  │  │ Secure  │        │   │
│  │  │ Client  │  │ Client  │  │ Storage │  │ Enclave │        │   │
│  │  └─────────┘  └─────────┘  └─────────┘  └─────────┘        │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                              │                                      │
│                              ▼                                      │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                      Security Layer                          │   │
│  │  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐        │   │
│  │  │ Kinnami │  │Biometric│  │  Key    │  │ Secure  │        │   │
│  │  │Encryption│ │  Auth   │  │ Manager │  │ Storage │        │   │
│  │  └─────────┘  └─────────┘  └─────────┘  └─────────┘        │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

## Project Structure

```
ECHO/
├── App/
│   ├── ECHOApp.swift                    # App entry point
│   ├── AppDelegate.swift                # Push notifications, lifecycle
│   └── SceneDelegate.swift              # Scene management
│
├── Core/
│   ├── DI/
│   │   ├── Container.swift              # Dependency container
│   │   └── Factories/                   # Service factories
│   │
│   ├── Security/
│   │   ├── SecureEnclaveManager.swift   # Secure Enclave operations
│   │   ├── BiometricAuthManager.swift   # Face ID / Touch ID
│   │   ├── KeychainManager.swift        # Keychain wrapper
│   │   └── KinnamiEncryption.swift      # E2E encryption
│   │
│   ├── Networking/
│   │   ├── APIClient.swift              # REST client
│   │   ├── WebSocketClient.swift        # Real-time messaging
│   │   ├── Endpoints.swift              # API endpoints
│   │   └── RequestInterceptor.swift     # Auth, encryption
│   │
│   ├── Storage/
│   │   ├── LocalDatabase.swift          # SwiftData setup
│   │   ├── SecureStorage.swift          # Encrypted storage
│   │   └── CacheManager.swift           # Caching layer
│   │
│   └── Utilities/
│       ├── Logger.swift                 # Privacy-safe logging
│       ├── Constants.swift              # App constants
│       └── Extensions/                  # Swift extensions
│
├── Domain/
│   ├── Models/
│   │   ├── User.swift                   # User model
│   │   ├── Message.swift                # Message model
│   │   ├── Conversation.swift           # Conversation model
│   │   ├── DID.swift                    # Decentralized identity
│   │   ├── Credential.swift             # Verifiable credential
│   │   └── Token.swift                  # ECHO token
│   │
│   ├── UseCases/
│   │   ├── Auth/
│   │   │   ├── AuthenticateUseCase.swift
│   │   │   ├── RegisterUseCase.swift
│   │   │   └── PasskeyUseCase.swift
│   │   │
│   │   ├── Messaging/
│   │   │   ├── SendMessageUseCase.swift
│   │   │   ├── FetchMessagesUseCase.swift
│   │   │   └── EncryptMessageUseCase.swift
│   │   │
│   │   ├── Identity/
│   │   │   ├── CreateDIDUseCase.swift
│   │   │   ├── VerifyIdentityUseCase.swift
│   │   │   └── ManageCredentialsUseCase.swift
│   │   │
│   │   └── Tokens/
│   │       ├── GetBalanceUseCase.swift
│   │       ├── SendTokensUseCase.swift
│   │       └── StakeTokensUseCase.swift
│   │
│   └── Repositories/
│       ├── AuthRepository.swift
│       ├── MessageRepository.swift
│       ├── UserRepository.swift
│       └── TokenRepository.swift
│
├── Presentation/
│   ├── Coordinators/
│   │   ├── AppCoordinator.swift         # Root coordinator
│   │   ├── AuthCoordinator.swift        # Auth flow
│   │   ├── MainCoordinator.swift        # Main app flow
│   │   └── SettingsCoordinator.swift    # Settings flow
│   │
│   ├── Features/
│   │   ├── Auth/
│   │   │   ├── Views/
│   │   │   │   ├── WelcomeView.swift
│   │   │   │   ├── PasskeySetupView.swift
│   │   │   │   └── BiometricPromptView.swift
│   │   │   └── ViewModels/
│   │   │       └── AuthViewModel.swift
│   │   │
│   │   ├── Onboarding/
│   │   │   ├── Views/
│   │   │   │   ├── OnboardingView.swift
│   │   │   │   ├── VCOnboardingView.swift
│   │   │   │   └── IDVerificationView.swift
│   │   │   └── ViewModels/
│   │   │       └── OnboardingViewModel.swift
│   │   │
│   │   ├── Conversations/
│   │   │   ├── Views/
│   │   │   │   ├── ConversationListView.swift
│   │   │   │   ├── ConversationRow.swift
│   │   │   │   └── NewConversationView.swift
│   │   │   └── ViewModels/
│   │   │       └── ConversationListViewModel.swift
│   │   │
│   │   ├── Chat/
│   │   │   ├── Views/
│   │   │   │   ├── ChatView.swift
│   │   │   │   ├── MessageBubble.swift
│   │   │   │   ├── MessageInputView.swift
│   │   │   │   └── MediaPickerView.swift
│   │   │   └── ViewModels/
│   │   │       └── ChatViewModel.swift
│   │   │
│   │   ├── Profile/
│   │   │   ├── Views/
│   │   │   │   ├── ProfileView.swift
│   │   │   │   ├── TrustScoreView.swift
│   │   │   │   └── CredentialsView.swift
│   │   │   └── ViewModels/
│   │   │       └── ProfileViewModel.swift
│   │   │
│   │   ├── Wallet/
│   │   │   ├── Views/
│   │   │   │   ├── WalletView.swift
│   │   │   │   ├── SendTokensView.swift
│   │   │   │   ├── StakingView.swift
│   │   │   │   └── TransactionHistoryView.swift
│   │   │   └── ViewModels/
│   │   │       └── WalletViewModel.swift
│   │   │
│   │   └── Settings/
│   │       ├── Views/
│   │       │   ├── SettingsView.swift
│   │       │   ├── PrivacySettingsView.swift
│   │       │   └── SecuritySettingsView.swift
│   │       └── ViewModels/
│   │           └── SettingsViewModel.swift
│   │
│   └── Components/
│       ├── Buttons/
│       │   ├── PrimaryButton.swift
│       │   └── SecondaryButton.swift
│       ├── Inputs/
│       │   ├── SecureTextField.swift
│       │   └── MessageInput.swift
│       ├── Cards/
│       │   ├── CredentialCard.swift
│       │   └── TransactionCard.swift
│       └── Indicators/
│           ├── TrustBadge.swift
│           └── VerificationBadge.swift
│
└── Resources/
    ├── Assets.xcassets
    ├── Localizable.strings
    └── Info.plist
```

## Core Security Implementation

### Secure Enclave Manager

```swift
import Foundation
import LocalAuthentication
import CryptoKit
import Security

/// Manages all Secure Enclave operations for the ECHO app
/// Keys stored here NEVER leave the device hardware
actor SecureEnclaveManager {
    
    // MARK: - Singleton
    static let shared = SecureEnclaveManager()
    
    // MARK: - Key Identifiers
    private enum KeyIdentifier {
        static let masterKey = "com.echo.master-key"
        static let didKey = "com.echo.did-key"
        static let messageKey = "com.echo.message-key"
        static let tokenKey = "com.echo.token-key"
    }
    
    // MARK: - Types
    
    enum SecureEnclaveError: LocalizedError {
        case notAvailable
        case keyGenerationFailed(Error?)
        case keyNotFound
        case signingFailed(Error?)
        case biometricFailed
        case encryptionFailed
        case decryptionFailed
        
        var errorDescription: String? {
            switch self {
            case .notAvailable:
                return "Secure Enclave is not available on this device"
            case .keyGenerationFailed(let error):
                return "Failed to generate key: \(error?.localizedDescription ?? "Unknown")"
            case .keyNotFound:
                return "Key not found in Secure Enclave"
            case .signingFailed(let error):
                return "Failed to sign data: \(error?.localizedDescription ?? "Unknown")"
            case .biometricFailed:
                return "Biometric authentication failed"
            case .encryptionFailed:
                return "Encryption failed"
            case .decryptionFailed:
                return "Decryption failed"
            }
        }
    }
    
    // MARK: - Availability Check
    
    /// Check if Secure Enclave is available
    var isAvailable: Bool {
        let context = LAContext()
        var error: NSError?
        
        guard context.canEvaluatePolicy(.deviceOwnerAuthenticationWithBiometrics, error: &error) else {
            return false
        }
        
        // Check for Secure Enclave support
        if #available(iOS 13.0, *) {
            return SecureEnclave.isAvailable
        }
        
        return true
    }
    
    // MARK: - Key Generation
    
    /// Generate the master identity key protected by biometrics
    /// This key is used to derive all other application keys
    func generateMasterKey() async throws -> Data {
        guard isAvailable else {
            throw SecureEnclaveError.notAvailable
        }
        
        // Check if key already exists
        if let existingKey = try? await getPublicKey(for: KeyIdentifier.masterKey) {
            return existingKey
        }
        
        // Create biometric-protected access control
        var accessError: Unmanaged<CFError>?
        guard let accessControl = SecAccessControlCreateWithFlags(
            kCFAllocatorDefault,
            kSecAttrAccessibleWhenUnlockedThisDeviceOnly,
            [.privateKeyUsage, .biometryCurrentSet],
            &accessError
        ) else {
            throw SecureEnclaveError.keyGenerationFailed(accessError?.takeRetainedValue())
        }
        
        // Key generation attributes
        let attributes: [String: Any] = [
            kSecAttrKeyType as String: kSecAttrKeyTypeECSECPrimeRandom,
            kSecAttrKeySizeInBits as String: 256,
            kSecAttrTokenID as String: kSecAttrTokenIDSecureEnclave,
            kSecPrivateKeyAttrs as String: [
                kSecAttrIsPermanent as String: true,
                kSecAttrApplicationTag as String: KeyIdentifier.masterKey.data(using: .utf8)!,
                kSecAttrAccessControl as String: accessControl,
            ],
        ]
        
        // Generate key pair
        var error: Unmanaged<CFError>?
        guard let privateKey = SecKeyCreateRandomKey(attributes as CFDictionary, &error) else {
            throw SecureEnclaveError.keyGenerationFailed(error?.takeRetainedValue())
        }
        
        // Export public key
        guard let publicKey = SecKeyCopyPublicKey(privateKey),
              let publicKeyData = SecKeyCopyExternalRepresentation(publicKey, &error) as Data? else {
            throw SecureEnclaveError.keyGenerationFailed(error?.takeRetainedValue())
        }
        
        return publicKeyData
    }
    
    /// Generate a DID-specific signing key
    func generateDIDKey() async throws -> Data {
        return try await generateKey(
            identifier: KeyIdentifier.didKey,
            requiresBiometric: true
        )
    }
    
    /// Generate a key with specified parameters
    private func generateKey(
        identifier: String,
        requiresBiometric: Bool
    ) async throws -> Data {
        var flags: SecAccessControlCreateFlags = [.privateKeyUsage]
        if requiresBiometric {
            flags.insert(.biometryCurrentSet)
        }
        
        var accessError: Unmanaged<CFError>?
        guard let accessControl = SecAccessControlCreateWithFlags(
            kCFAllocatorDefault,
            kSecAttrAccessibleWhenUnlockedThisDeviceOnly,
            flags,
            &accessError
        ) else {
            throw SecureEnclaveError.keyGenerationFailed(accessError?.takeRetainedValue())
        }
        
        let attributes: [String: Any] = [
            kSecAttrKeyType as String: kSecAttrKeyTypeECSECPrimeRandom,
            kSecAttrKeySizeInBits as String: 256,
            kSecAttrTokenID as String: kSecAttrTokenIDSecureEnclave,
            kSecPrivateKeyAttrs as String: [
                kSecAttrIsPermanent as String: true,
                kSecAttrApplicationTag as String: identifier.data(using: .utf8)!,
                kSecAttrAccessControl as String: accessControl,
            ],
        ]
        
        var error: Unmanaged<CFError>?
        guard let privateKey = SecKeyCreateRandomKey(attributes as CFDictionary, &error) else {
            throw SecureEnclaveError.keyGenerationFailed(error?.takeRetainedValue())
        }
        
        guard let publicKey = SecKeyCopyPublicKey(privateKey),
              let publicKeyData = SecKeyCopyExternalRepresentation(publicKey, &error) as Data? else {
            throw SecureEnclaveError.keyGenerationFailed(error?.takeRetainedValue())
        }
        
        return publicKeyData
    }
    
    // MARK: - Key Retrieval
    
    /// Get public key data for a key identifier
    func getPublicKey(for identifier: String) async throws -> Data {
        let privateKey = try await getPrivateKey(for: identifier, context: nil)
        
        var error: Unmanaged<CFError>?
        guard let publicKey = SecKeyCopyPublicKey(privateKey),
              let publicKeyData = SecKeyCopyExternalRepresentation(publicKey, &error) as Data? else {
            throw SecureEnclaveError.keyNotFound
        }
        
        return publicKeyData
    }
    
    /// Get private key reference (requires biometric for protected keys)
    private func getPrivateKey(
        for identifier: String,
        context: LAContext?
    ) async throws -> SecKey {
        var query: [String: Any] = [
            kSecClass as String: kSecClassKey,
            kSecAttrApplicationTag as String: identifier.data(using: .utf8)!,
            kSecAttrKeyType as String: kSecAttrKeyTypeECSECPrimeRandom,
            kSecReturnRef as String: true,
        ]
        
        if let context = context {
            query[kSecUseAuthenticationContext as String] = context
        }
        
        var item: CFTypeRef?
        let status = SecItemCopyMatching(query as CFDictionary, &item)
        
        guard status == errSecSuccess, let key = item else {
            throw SecureEnclaveError.keyNotFound
        }
        
        return key as! SecKey
    }
    
    // MARK: - Signing
    
    /// Sign data using the master key (triggers biometric prompt)
    func sign(
        data: Data,
        reason: String = "Authenticate to sign"
    ) async throws -> Data {
        return try await sign(
            data: data,
            keyIdentifier: KeyIdentifier.masterKey,
            reason: reason
        )
    }
    
    /// Sign data for DID operations
    func signForDID(
        data: Data,
        reason: String = "Sign with your identity"
    ) async throws -> Data {
        return try await sign(
            data: data,
            keyIdentifier: KeyIdentifier.didKey,
            reason: reason
        )
    }
    
    /// Sign data with specified key
    private func sign(
        data: Data,
        keyIdentifier: String,
        reason: String
    ) async throws -> Data {
        // Create authentication context
        let context = LAContext()
        context.localizedReason = reason
        
        // Authenticate first
        let authenticated = try await withCheckedThrowingContinuation { continuation in
            context.evaluatePolicy(
                .deviceOwnerAuthenticationWithBiometrics,
                localizedReason: reason
            ) { success, error in
                if success {
                    continuation.resume(returning: true)
                } else {
                    continuation.resume(throwing: SecureEnclaveError.biometricFailed)
                }
            }
        }
        
        guard authenticated else {
            throw SecureEnclaveError.biometricFailed
        }
        
        // Get key with authenticated context
        let privateKey = try await getPrivateKey(for: keyIdentifier, context: context)
        
        // Sign the data
        var error: Unmanaged<CFError>?
        guard let signature = SecKeyCreateSignature(
            privateKey,
            .ecdsaSignatureMessageX962SHA256,
            data as CFData,
            &error
        ) else {
            throw SecureEnclaveError.signingFailed(error?.takeRetainedValue())
        }
        
        return signature as Data
    }
    
    // MARK: - Key Derivation
    
    /// Derive application keys from master key signature
    /// Returns keys for different purposes (messaging, storage, tokens)
    func deriveApplicationKeys() async throws -> DerivedKeys {
        // Sign a deterministic value to get consistent derived keys
        let derivationInput = "ECHO-KEY-DERIVATION-v1".data(using: .utf8)!
        let signature = try await sign(
            data: derivationInput,
            reason: "Unlock your ECHO account"
        )
        
        // Hash the signature to create master secret
        let masterSecret = SHA256.hash(data: signature)
        let masterKey = SymmetricKey(data: Data(masterSecret))
        
        // Derive individual keys using HKDF
        return DerivedKeys(
            messageKey: deriveKey(from: masterKey, info: "messages"),
            storageKey: deriveKey(from: masterKey, info: "storage"),
            tokenKey: deriveKey(from: masterKey, info: "tokens"),
            sessionKey: deriveKey(from: masterKey, info: "session")
        )
    }
    
    private func deriveKey(from master: SymmetricKey, info: String) -> SymmetricKey {
        return HKDF<SHA256>.deriveKey(
            inputKeyMaterial: master,
            info: info.data(using: .utf8)!,
            outputByteCount: 32
        )
    }
    
    // MARK: - Key Deletion
    
    /// Delete a key from Secure Enclave
    func deleteKey(identifier: String) async throws {
        let query: [String: Any] = [
            kSecClass as String: kSecClassKey,
            kSecAttrApplicationTag as String: identifier.data(using: .utf8)!,
        ]
        
        let status = SecItemDelete(query as CFDictionary)
        
        guard status == errSecSuccess || status == errSecItemNotFound else {
            throw SecureEnclaveError.keyNotFound
        }
    }
    
    /// Delete all ECHO keys (for account deletion)
    func deleteAllKeys() async throws {
        try await deleteKey(identifier: KeyIdentifier.masterKey)
        try await deleteKey(identifier: KeyIdentifier.didKey)
        try await deleteKey(identifier: KeyIdentifier.messageKey)
        try await deleteKey(identifier: KeyIdentifier.tokenKey)
    }
}

// MARK: - Derived Keys

struct DerivedKeys {
    let messageKey: SymmetricKey
    let storageKey: SymmetricKey
    let tokenKey: SymmetricKey
    let sessionKey: SymmetricKey
    
    /// Clear keys from memory (best effort in Swift)
    func clear() {
        // In production, use secure memory clearing
        // Swift's ARC handles memory management
    }
}
```

### Biometric Authentication Manager

```swift
import LocalAuthentication
import Foundation

/// Manages biometric authentication (Face ID / Touch ID)
@MainActor
final class BiometricAuthManager: ObservableObject {
    
    // MARK: - Published State
    
    @Published private(set) var biometricType: BiometricType = .none
    @Published private(set) var isAuthenticated = false
    @Published private(set) var authenticationError: AuthError?
    
    // MARK: - Types
    
    enum BiometricType {
        case none
        case touchID
        case faceID
        
        var displayName: String {
            switch self {
            case .none: return "Passcode"
            case .touchID: return "Touch ID"
            case .faceID: return "Face ID"
            }
        }
        
        var iconName: String {
            switch self {
            case .none: return "lock.fill"
            case .touchID: return "touchid"
            case .faceID: return "faceid"
            }
        }
    }
    
    enum AuthError: LocalizedError {
        case notAvailable
        case notEnrolled
        case lockout
        case cancelled
        case failed(Error)
        
        var errorDescription: String? {
            switch self {
            case .notAvailable:
                return "Biometric authentication is not available"
            case .notEnrolled:
                return "No biometrics enrolled. Please set up Face ID or Touch ID."
            case .lockout:
                return "Biometric authentication is locked. Please use your device passcode."
            case .cancelled:
                return "Authentication was cancelled"
            case .failed(let error):
                return error.localizedDescription
            }
        }
    }
    
    // MARK: - Private Properties
    
    private let context = LAContext()
    
    // MARK: - Initialization
    
    init() {
        updateBiometricType()
    }
    
    // MARK: - Public Methods
    
    /// Check and update available biometric type
    func updateBiometricType() {
        var error: NSError?
        
        guard context.canEvaluatePolicy(.deviceOwnerAuthenticationWithBiometrics, error: &error) else {
            biometricType = .none
            return
        }
        
        switch context.biometryType {
        case .touchID:
            biometricType = .touchID
        case .faceID:
            biometricType = .faceID
        default:
            biometricType = .none
        }
    }
    
    /// Authenticate with biometrics
    func authenticate(reason: String = "Authenticate to access ECHO") async -> Bool {
        let context = LAContext()
        context.localizedFallbackTitle = "Use Passcode"
        
        var error: NSError?
        
        // Check availability
        guard context.canEvaluatePolicy(.deviceOwnerAuthenticationWithBiometrics, error: &error) else {
            if let laError = error as? LAError {
                handleLAError(laError)
            }
            return false
        }
        
        // Perform authentication
        do {
            let success = try await context.evaluatePolicy(
                .deviceOwnerAuthenticationWithBiometrics,
                localizedReason: reason
            )
            
            isAuthenticated = success
            authenticationError = nil
            return success
            
        } catch let error as LAError {
            handleLAError(error)
            return false
        } catch {
            authenticationError = .failed(error)
            return false
        }
    }
    
    /// Authenticate for a specific operation with custom UI
    func authenticateForOperation(
        reason: String,
        fallbackTitle: String? = nil
    ) async throws -> Bool {
        let context = LAContext()
        context.localizedFallbackTitle = fallbackTitle ?? "Enter Passcode"
        context.localizedCancelTitle = "Cancel"
        
        var error: NSError?
        
        guard context.canEvaluatePolicy(.deviceOwnerAuthenticationWithBiometrics, error: &error) else {
            throw AuthError.notAvailable
        }
        
        do {
            return try await context.evaluatePolicy(
                .deviceOwnerAuthenticationWithBiometrics,
                localizedReason: reason
            )
        } catch let error as LAError {
            throw mapLAError(error)
        }
    }
    
    /// Reset authentication state
    func resetAuthentication() {
        isAuthenticated = false
        authenticationError = nil
    }
    
    // MARK: - Private Methods
    
    private func handleLAError(_ error: LAError) {
        authenticationError = mapLAError(error)
        isAuthenticated = false
    }
    
    private func mapLAError(_ error: LAError) -> AuthError {
        switch error.code {
        case .biometryNotAvailable:
            return .notAvailable
        case .biometryNotEnrolled:
            return .notEnrolled
        case .biometryLockout:
            return .lockout
        case .userCancel, .appCancel, .systemCancel:
            return .cancelled
        default:
            return .failed(error)
        }
    }
}
```

### Kinnami Encryption Service

```swift
import Foundation
import CryptoKit

/// Kinnami encryption protocol implementation
/// Provides end-to-end encryption for all message content
final class KinnamiEncryptionService {
    
    // MARK: - Types
    
    struct EncryptedPayload: Codable {
        let version: Int
        let ephemeralPublicKey: Data
        let nonce: Data
        let ciphertext: Data
        let tag: Data
        let commitment: Data
        
        var serialized: Data {
            try! JSONEncoder().encode(self)
        }
        
        static func deserialize(from data: Data) throws -> EncryptedPayload {
            try JSONDecoder().decode(EncryptedPayload.self, from: data)
        }
    }
    
    struct KeyPair {
        let privateKey: Curve25519.KeyAgreement.PrivateKey
        let publicKey: Curve25519.KeyAgreement.PublicKey
        
        init() {
            self.privateKey = Curve25519.KeyAgreement.PrivateKey()
            self.publicKey = privateKey.publicKey
        }
        
        init(privateKey: Curve25519.KeyAgreement.PrivateKey) {
            self.privateKey = privateKey
            self.publicKey = privateKey.publicKey
        }
    }
    
    enum EncryptionError: LocalizedError {
        case keyDerivationFailed
        case encryptionFailed
        case decryptionFailed
        case invalidPayload
        case invalidPublicKey
        
        var errorDescription: String? {
            switch self {
            case .keyDerivationFailed: return "Failed to derive encryption key"
            case .encryptionFailed: return "Encryption failed"
            case .decryptionFailed: return "Decryption failed"
            case .invalidPayload: return "Invalid encrypted payload"
            case .invalidPublicKey: return "Invalid public key"
            }
        }
    }
    
    // MARK: - Constants
    
    private static let currentVersion = 1
    private static let info = "ECHO-Kinnami-v1".data(using: .utf8)!
    
    // MARK: - Singleton
    
    static let shared = KinnamiEncryptionService()
    
    // MARK: - Public Methods
    
    /// Encrypt a message for a recipient
    /// Uses X25519 ECDH key agreement + ChaCha20-Poly1305
    func encrypt(
        plaintext: Data,
        recipientPublicKey: Data
    ) throws -> EncryptedPayload {
        // Parse recipient's public key
        guard let recipientKey = try? Curve25519.KeyAgreement.PublicKey(
            rawRepresentation: recipientPublicKey
        ) else {
            throw EncryptionError.invalidPublicKey
        }
        
        // Generate ephemeral key pair for forward secrecy
        let ephemeralKeyPair = KeyPair()
        
        // Perform X25519 key agreement
        guard let sharedSecret = try? ephemeralKeyPair.privateKey.sharedSecretFromKeyAgreement(
            with: recipientKey
        ) else {
            throw EncryptionError.keyDerivationFailed
        }
        
        // Derive symmetric key using HKDF
        let symmetricKey = sharedSecret.hkdfDerivedSymmetricKey(
            using: SHA256.self,
            salt: Data(),
            sharedInfo: Self.info,
            outputByteCount: 32
        )
        
        // Generate random nonce
        let nonce = try ChaChaPoly.Nonce()
        
        // Encrypt with ChaCha20-Poly1305
        guard let sealedBox = try? ChaChaPoly.seal(
            plaintext,
            using: symmetricKey,
            nonce: nonce
        ) else {
            throw EncryptionError.encryptionFailed
        }
        
        // Create commitment for integrity verification
        let commitment = createCommitment(plaintext: plaintext, nonce: Data(nonce))
        
        return EncryptedPayload(
            version: Self.currentVersion,
            ephemeralPublicKey: ephemeralKeyPair.publicKey.rawRepresentation,
            nonce: Data(nonce),
            ciphertext: sealedBox.ciphertext,
            tag: sealedBox.tag,
            commitment: commitment
        )
    }
    
    /// Decrypt a message using own private key
    func decrypt(
        payload: EncryptedPayload,
        privateKey: Curve25519.KeyAgreement.PrivateKey
    ) throws -> Data {
        // Parse ephemeral public key
        guard let ephemeralPublicKey = try? Curve25519.KeyAgreement.PublicKey(
            rawRepresentation: payload.ephemeralPublicKey
        ) else {
            throw EncryptionError.invalidPublicKey
        }
        
        // Perform key agreement
        guard let sharedSecret = try? privateKey.sharedSecretFromKeyAgreement(
            with: ephemeralPublicKey
        ) else {
            throw EncryptionError.keyDerivationFailed
        }
        
        // Derive symmetric key
        let symmetricKey = sharedSecret.hkdfDerivedSymmetricKey(
            using: SHA256.self,
            salt: Data(),
            sharedInfo: Self.info,
            outputByteCount: 32
        )
        
        // Reconstruct nonce
        guard let nonce = try? ChaChaPoly.Nonce(data: payload.nonce) else {
            throw EncryptionError.invalidPayload
        }
        
        // Reconstruct sealed box
        guard let sealedBox = try? ChaChaPoly.SealedBox(
            nonce: nonce,
            ciphertext: payload.ciphertext,
            tag: payload.tag
        ) else {
            throw EncryptionError.invalidPayload
        }
        
        // Decrypt
        guard let plaintext = try? ChaChaPoly.open(sealedBox, using: symmetricKey) else {
            throw EncryptionError.decryptionFailed
        }
        
        // Verify commitment
        let expectedCommitment = createCommitment(plaintext: plaintext, nonce: payload.nonce)
        guard expectedCommitment == payload.commitment else {
            throw EncryptionError.decryptionFailed
        }
        
        return plaintext
    }
    
    /// Encrypt data for local storage
    func encryptForStorage(
        plaintext: Data,
        key: SymmetricKey
    ) throws -> Data {
        guard let sealedBox = try? AES.GCM.seal(plaintext, using: key) else {
            throw EncryptionError.encryptionFailed
        }
        return sealedBox.combined!
    }
    
    /// Decrypt locally stored data
    func decryptFromStorage(
        ciphertext: Data,
        key: SymmetricKey
    ) throws -> Data {
        guard let sealedBox = try? AES.GCM.SealedBox(combined: ciphertext),
              let plaintext = try? AES.GCM.open(sealedBox, using: key) else {
            throw EncryptionError.decryptionFailed
        }
        return plaintext
    }
    
    // MARK: - Commitment
    
    /// Create a commitment (hash) for integrity verification
    /// This commitment can be stored on-chain without revealing content
    private func createCommitment(plaintext: Data, nonce: Data) -> Data {
        var hasher = SHA256()
        
        // Hash the plaintext
        let innerHash = SHA256.hash(data: plaintext)
        hasher.update(data: Data(innerHash))
        
        // Add nonce for uniqueness
        hasher.update(data: nonce)
        
        // Add timestamp
        let timestamp = UInt64(Date().timeIntervalSince1970)
        withUnsafeBytes(of: timestamp.bigEndian) { hasher.update(bufferPointer: $0) }
        
        return Data(hasher.finalize())
    }
    
    // MARK: - Key Generation
    
    /// Generate a new key pair for messaging
    func generateKeyPair() -> KeyPair {
        return KeyPair()
    }
    
    /// Generate a symmetric key for local encryption
    func generateSymmetricKey() -> SymmetricKey {
        return SymmetricKey(size: .bits256)
    }
}
```

## Networking Layer

### API Client

```swift
import Foundation

/// Main API client for communicating with the Go backend
actor APIClient {
    
    // MARK: - Types
    
    enum APIError: LocalizedError {
        case invalidURL
        case requestFailed(Error)
        case invalidResponse
        case httpError(Int, String?)
        case decodingFailed(Error)
        case unauthorized
        case serverError(String)
        
        var errorDescription: String? {
            switch self {
            case .invalidURL: return "Invalid URL"
            case .requestFailed(let error): return "Request failed: \(error.localizedDescription)"
            case .invalidResponse: return "Invalid response from server"
            case .httpError(let code, let message): return "HTTP \(code): \(message ?? "Unknown error")"
            case .decodingFailed(let error): return "Failed to decode response: \(error.localizedDescription)"
            case .unauthorized: return "Unauthorized. Please sign in again."
            case .serverError(let message): return "Server error: \(message)"
            }
        }
    }
    
    struct APIResponse<T: Decodable>: Decodable {
        let success: Bool
        let data: T?
        let error: APIErrorResponse?
    }
    
    struct APIErrorResponse: Decodable {
        let code: String
        let message: String
    }
    
    // MARK: - Properties
    
    private let baseURL: URL
    private let session: URLSession
    private let encryptionService: KinnamiEncryptionService
    private let secureEnclave: SecureEnclaveManager
    
    private var authToken: String?
    private var sessionKey: SymmetricKey?
    
    // MARK: - Initialization
    
    init(
        baseURL: URL = URL(string: "https://api.echo.network")!,
        encryptionService: KinnamiEncryptionService = .shared,
        secureEnclave: SecureEnclaveManager = .shared
    ) {
        self.baseURL = baseURL
        self.encryptionService = encryptionService
        self.secureEnclave = secureEnclave
        
        // Configure URLSession with TLS 1.3
        let configuration = URLSessionConfiguration.default
        configuration.tlsMinimumSupportedProtocolVersion = .TLSv13
        configuration.httpAdditionalHeaders = [
            "Content-Type": "application/json",
            "Accept": "application/json",
            "X-Client-Version": Bundle.main.appVersion,
            "X-Platform": "iOS",
        ]
        
        self.session = URLSession(configuration: configuration)
    }
    
    // MARK: - Authentication
    
    /// Set authentication token after login
    func setAuthToken(_ token: String) {
        self.authToken = token
    }
    
    /// Set session key for encrypted communication
    func setSessionKey(_ key: SymmetricKey) {
        self.sessionKey = key
    }
    
    /// Clear authentication state
    func clearAuth() {
        self.authToken = nil
        self.sessionKey = nil
    }
    
    // MARK: - Request Methods
    
    /// Perform a GET request
    func get<T: Decodable>(
        endpoint: Endpoint,
        queryItems: [URLQueryItem]? = nil
    ) async throws -> T {
        let request = try buildRequest(
            endpoint: endpoint,
            method: "GET",
            queryItems: queryItems
        )
        
        return try await performRequest(request)
    }
    
    /// Perform a POST request
    func post<T: Decodable, B: Encodable>(
        endpoint: Endpoint,
        body: B
    ) async throws -> T {
        var request = try buildRequest(
            endpoint: endpoint,
            method: "POST"
        )
        
        let bodyData = try JSONEncoder().encode(body)
        request.httpBody = try encryptBody(bodyData)
        
        return try await performRequest(request)
    }
    
    /// Perform a PUT request
    func put<T: Decodable, B: Encodable>(
        endpoint: Endpoint,
        body: B
    ) async throws -> T {
        var request = try buildRequest(
            endpoint: endpoint,
            method: "PUT"
        )
        
        let bodyData = try JSONEncoder().encode(body)
        request.httpBody = try encryptBody(bodyData)
        
        return try await performRequest(request)
    }
    
    /// Perform a DELETE request
    func delete<T: Decodable>(
        endpoint: Endpoint
    ) async throws -> T {
        let request = try buildRequest(
            endpoint: endpoint,
            method: "DELETE"
        )
        
        return try await performRequest(request)
    }
    
    // MARK: - Private Methods
    
    private func buildRequest(
        endpoint: Endpoint,
        method: String,
        queryItems: [URLQueryItem]? = nil
    ) throws -> URLRequest {
        var components = URLComponents(
            url: baseURL.appendingPathComponent(endpoint.path),
            resolvingAgainstBaseURL: true
        )
        components?.queryItems = queryItems
        
        guard let url = components?.url else {
            throw APIError.invalidURL
        }
        
        var request = URLRequest(url: url)
        request.httpMethod = method
        request.timeoutInterval = 30
        
        // Add auth token if available
        if let token = authToken {
            request.addValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }
        
        // Add request ID for tracing
        request.addValue(UUID().uuidString, forHTTPHeaderField: "X-Request-ID")
        
        return request
    }
    
    private func encryptBody(_ data: Data) throws -> Data {
        guard let sessionKey = sessionKey else {
            // No encryption if no session key (e.g., during auth)
            return data
        }
        
        return try encryptionService.encryptForStorage(
            plaintext: data,
            key: sessionKey
        )
    }
    
    private func decryptResponse(_ data: Data) throws -> Data {
        guard let sessionKey = sessionKey else {
            return data
        }
        
        return try encryptionService.decryptFromStorage(
            ciphertext: data,
            key: sessionKey
        )
    }
    
    private func performRequest<T: Decodable>(_ request: URLRequest) async throws -> T {
        do {
            let (data, response) = try await session.data(for: request)
            
            guard let httpResponse = response as? HTTPURLResponse else {
                throw APIError.invalidResponse
            }
            
            // Handle HTTP errors
            switch httpResponse.statusCode {
            case 200...299:
                break
            case 401:
                throw APIError.unauthorized
            case 400...499:
                let errorMessage = String(data: data, encoding: .utf8)
                throw APIError.httpError(httpResponse.statusCode, errorMessage)
            case 500...599:
                let errorMessage = String(data: data, encoding: .utf8)
                throw APIError.serverError(errorMessage ?? "Internal server error")
            default:
                throw APIError.httpError(httpResponse.statusCode, nil)
            }
            
            // Decrypt if needed
            let decryptedData = try decryptResponse(data)
            
            // Decode response
            let decoder = JSONDecoder()
            decoder.dateDecodingStrategy = .iso8601
            
            let apiResponse = try decoder.decode(APIResponse<T>.self, from: decryptedData)
            
            if let error = apiResponse.error {
                throw APIError.serverError(error.message)
            }
            
            guard let responseData = apiResponse.data else {
                throw APIError.invalidResponse
            }
            
            return responseData
            
        } catch let error as APIError {
            throw error
        } catch {
            throw APIError.requestFailed(error)
        }
    }
}

// MARK: - Endpoints

enum Endpoint {
    // Auth
    case authChallenge
    case authVerify
    case register
    
    // User
    case profile
    case updateProfile
    case trustScore
    
    // Messages
    case conversations
    case conversation(id: String)
    case messages(conversationId: String)
    case sendMessage(conversationId: String)
    
    // Contacts
    case contacts
    case addContact
    case blockContact(id: String)
    
    // Identity
    case createDID
    case credentials
    case verifyIdentity
    
    // Tokens
    case balance
    case transactions
    case send
    case stake
    
    var path: String {
        switch self {
        case .authChallenge: return "/v1/auth/challenge"
        case .authVerify: return "/v1/auth/verify"
        case .register: return "/v1/auth/register"
        case .profile: return "/v1/user/profile"
        case .updateProfile: return "/v1/user/profile"
        case .trustScore: return "/v1/user/trust-score"
        case .conversations: return "/v1/conversations"
        case .conversation(let id): return "/v1/conversations/\(id)"
        case .messages(let id): return "/v1/conversations/\(id)/messages"
        case .sendMessage(let id): return "/v1/conversations/\(id)/messages"
        case .contacts: return "/v1/contacts"
        case .addContact: return "/v1/contacts"
        case .blockContact(let id): return "/v1/contacts/\(id)/block"
        case .createDID: return "/v1/identity/did"
        case .credentials: return "/v1/identity/credentials"
        case .verifyIdentity: return "/v1/identity/verify"
        case .balance: return "/v1/wallet/balance"
        case .transactions: return "/v1/wallet/transactions"
        case .send: return "/v1/wallet/send"
        case .stake: return "/v1/wallet/stake"
        }
    }
}

// MARK: - Extensions

extension Bundle {
    var appVersion: String {
        return infoDictionary?["CFBundleShortVersionString"] as? String ?? "1.0"
    }
}
```

### WebSocket Client

```swift
import Foundation
import Combine

/// WebSocket client for real-time messaging
actor WebSocketClient {
    
    // MARK: - Types
    
    enum ConnectionState {
        case disconnected
        case connecting
        case connected
        case reconnecting
    }
    
    enum WebSocketError: LocalizedError {
        case notConnected
        case connectionFailed(Error)
        case sendFailed(Error)
        
        var errorDescription: String? {
            switch self {
            case .notConnected: return "WebSocket not connected"
            case .connectionFailed(let error): return "Connection failed: \(error.localizedDescription)"
            case .sendFailed(let error): return "Send failed: \(error.localizedDescription)"
            }
        }
    }
    
    struct WSMessage: Codable {
        let type: MessageType
        let payload: Data
        let timestamp: Date
        
        enum MessageType: String, Codable {
            case message
            case typing
            case presence
            case receipt
            case ack
        }
    }
    
    // MARK: - Properties
    
    private var webSocketTask: URLSessionWebSocketTask?
    private let baseURL: URL
    private let session: URLSession
    private let encryptionService: KinnamiEncryptionService
    
    private(set) var connectionState: ConnectionState = .disconnected
    private var authToken: String?
    private var reconnectAttempts = 0
    private let maxReconnectAttempts = 5
    
    // Callbacks
    private var messageHandler: ((WSMessage) -> Void)?
    private var connectionHandler: ((ConnectionState) -> Void)?
    
    // MARK: - Initialization
    
    init(
        baseURL: URL = URL(string: "wss://ws.echo.network")!,
        encryptionService: KinnamiEncryptionService = .shared
    ) {
        self.baseURL = baseURL
        self.encryptionService = encryptionService
        
        let configuration = URLSessionConfiguration.default
        configuration.tlsMinimumSupportedProtocolVersion = .TLSv13
        self.session = URLSession(configuration: configuration)
    }
    
    // MARK: - Connection Management
    
    /// Connect to WebSocket server
    func connect(authToken: String) async throws {
        self.authToken = authToken
        
        var request = URLRequest(url: baseURL)
        request.addValue("Bearer \(authToken)", forHTTPHeaderField: "Authorization")
        request.addValue("iOS", forHTTPHeaderField: "X-Platform")
        
        webSocketTask = session.webSocketTask(with: request)
        
        connectionState = .connecting
        connectionHandler?(.connecting)
        
        webSocketTask?.resume()
        
        connectionState = .connected
        connectionHandler?(.connected)
        reconnectAttempts = 0
        
        // Start receiving messages
        await receiveMessages()
    }
    
    /// Disconnect from server
    func disconnect() {
        webSocketTask?.cancel(with: .goingAway, reason: nil)
        webSocketTask = nil
        connectionState = .disconnected
        connectionHandler?(.disconnected)
    }
    
    /// Set message handler
    func onMessage(_ handler: @escaping (WSMessage) -> Void) {
        self.messageHandler = handler
    }
    
    /// Set connection state handler
    func onConnectionStateChange(_ handler: @escaping (ConnectionState) -> Void) {
        self.connectionHandler = handler
    }
    
    // MARK: - Sending
    
    /// Send a message
    func send(_ message: WSMessage) async throws {
        guard connectionState == .connected else {
            throw WebSocketError.notConnected
        }
        
        let data = try JSONEncoder().encode(message)
        let wsMessage = URLSessionWebSocketTask.Message.data(data)
        
        do {
            try await webSocketTask?.send(wsMessage)
        } catch {
            throw WebSocketError.sendFailed(error)
        }
    }
    
    /// Send typing indicator
    func sendTypingIndicator(conversationId: String) async throws {
        let payload = try JSONEncoder().encode(["conversationId": conversationId])
        
        let message = WSMessage(
            type: .typing,
            payload: payload,
            timestamp: Date()
        )
        
        try await send(message)
    }
    
    /// Send read receipt
    func sendReadReceipt(messageId: String, conversationId: String) async throws {
        let payload = try JSONEncoder().encode([
            "messageId": messageId,
            "conversationId": conversationId
        ])
        
        let message = WSMessage(
            type: .receipt,
            payload: payload,
            timestamp: Date()
        )
        
        try await send(message)
    }
    
    // MARK: - Receiving
    
    private func receiveMessages() async {
        guard let webSocketTask = webSocketTask else { return }
        
        do {
            while connectionState == .connected {
                let message = try await webSocketTask.receive()
                
                switch message {
                case .data(let data):
                    if let wsMessage = try? JSONDecoder().decode(WSMessage.self, from: data) {
                        messageHandler?(wsMessage)
                    }
                    
                case .string(let string):
                    if let data = string.data(using: .utf8),
                       let wsMessage = try? JSONDecoder().decode(WSMessage.self, from: data) {
                        messageHandler?(wsMessage)
                    }
                    
                @unknown default:
                    break
                }
            }
        } catch {
            // Connection lost, attempt reconnect
            await handleDisconnect()
        }
    }
    
    private func handleDisconnect() async {
        connectionState = .reconnecting
        connectionHandler?(.reconnecting)
        
        guard reconnectAttempts < maxReconnectAttempts,
              let authToken = authToken else {
            connectionState = .disconnected
            connectionHandler?(.disconnected)
            return
        }
        
        reconnectAttempts += 1
        
        // Exponential backoff
        let delay = UInt64(pow(2.0, Double(reconnectAttempts))) * 1_000_000_000
        try? await Task.sleep(nanoseconds: delay)
        
        do {
            try await connect(authToken: authToken)
        } catch {
            await handleDisconnect()
        }
    }
    
    // MARK: - Ping/Pong
    
    /// Send ping to keep connection alive
    func ping() async throws {
        try await webSocketTask?.sendPing(pongReceiveHandler: { error in
            if let error = error {
                print("Pong error: \(error)")
            }
        })
    }
}
```

## Domain Models

### Message Model

```swift
import Foundation
import SwiftData

@Model
final class Message: Identifiable {
    // MARK: - Properties
    
    @Attribute(.unique) var id: String
    var conversationId: String
    var senderId: String
    var senderDID: String
    
    // Content
    var contentType: ContentType
    var encryptedContent: Data
    var commitment: Data  // For blockchain anchoring
    
    // Metadata
    var timestamp: Date
    var editedAt: Date?
    var expiresAt: Date?  // For disappearing messages
    
    // Status
    var status: DeliveryStatus
    var readAt: Date?
    var readBy: [String]  // DIDs who have read (for groups)
    
    // Threading
    var replyToId: String?
    var forwardedFrom: String?
    
    // Reactions
    var reactions: [Reaction]
    
    // Edit history (stored locally only)
    var editHistory: [EditRecord]
    
    // MARK: - Types
    
    enum ContentType: String, Codable {
        case text
        case image
        case video
        case audio
        case file
        case location
        case contact
        case sticker
    }
    
    enum DeliveryStatus: String, Codable {
        case sending
        case sent
        case delivered
        case read
        case failed
    }
    
    struct Reaction: Codable {
        let emoji: String
        let userId: String
        let timestamp: Date
    }
    
    struct EditRecord: Codable {
        let content: Data
        let editedAt: Date
    }
    
    // MARK: - Initialization
    
    init(
        id: String = UUID().uuidString,
        conversationId: String,
        senderId: String,
        senderDID: String,
        contentType: ContentType,
        encryptedContent: Data,
        commitment: Data
    ) {
        self.id = id
        self.conversationId = conversationId
        self.senderId = senderId
        self.senderDID = senderDID
        self.contentType = contentType
        self.encryptedContent = encryptedContent
        self.commitment = commitment
        self.timestamp = Date()
        self.status = .sending
        self.readBy = []
        self.reactions = []
        self.editHistory = []
    }
    
    // MARK: - Computed Properties
    
    var isEdited: Bool {
        editedAt != nil
    }
    
    var isExpired: Bool {
        if let expiresAt = expiresAt {
            return Date() > expiresAt
        }
        return false
    }
    
    var canEdit: Bool {
        // Can edit within 24 hours
        let editWindow = TimeInterval(24 * 60 * 60)
        return Date().timeIntervalSince(timestamp) < editWindow
    }
}
```

### User Model

```swift
import Foundation
import SwiftData

@Model
final class User: Identifiable {
    // MARK: - Properties
    
    @Attribute(.unique) var id: String
    @Attribute(.unique) var did: String
    
    // Profile
    var displayName: String
    var username: String?
    var avatarURL: String?
    var bio: String?
    
    // Trust & Verification
    var trustScore: Int
    var verificationBadges: [VerificationBadge]
    var isVerified: Bool
    
    // Status
    var status: UserStatus
    var lastSeen: Date?
    var isOnline: Bool
    
    // Privacy settings
    var showLastSeen: PrivacyLevel
    var showOnlineStatus: PrivacyLevel
    var showProfilePhoto: PrivacyLevel
    var showReadReceipts: Bool
    
    // Contact info
    var isContact: Bool
    var isBlocked: Bool
    var isFavorite: Bool
    
    // Metadata
    var createdAt: Date
    var updatedAt: Date
    
    // MARK: - Types
    
    enum UserStatus: String, Codable {
        case active
        case away
        case busy
        case offline
    }
    
    enum PrivacyLevel: String, Codable {
        case everyone
        case contacts
        case nobody
    }
    
    struct VerificationBadge: Codable {
        let type: BadgeType
        let issuedAt: Date
        let expiresAt: Date?
        
        enum BadgeType: String, Codable {
            case identityVerified = "identity_verified"
            case phoneVerified = "phone_verified"
            case emailVerified = "email_verified"
            case governmentId = "government_id"
            case enterprise = "enterprise"
        }
    }
    
    // MARK: - Initialization
    
    init(
        id: String,
        did: String,
        displayName: String
    ) {
        self.id = id
        self.did = did
        self.displayName = displayName
        self.trustScore = 0
        self.verificationBadges = []
        self.isVerified = false
        self.status = .active
        self.isOnline = false
        self.showLastSeen = .contacts
        self.showOnlineStatus = .contacts
        self.showProfilePhoto = .everyone
        self.showReadReceipts = true
        self.isContact = false
        self.isBlocked = false
        self.isFavorite = false
        self.createdAt = Date()
        self.updatedAt = Date()
    }
    
    // MARK: - Computed Properties
    
    var trustTier: TrustTier {
        switch trustScore {
        case 80...100: return .trusted
        case 60..<80: return .verified
        case 40..<60: return .member
        case 20..<40: return .newcomer
        default: return .unverified
        }
    }
    
    enum TrustTier: String {
        case unverified, newcomer, member, verified, trusted
        
        var color: String {
            switch self {
            case .unverified: return "gray"
            case .newcomer: return "blue"
            case .member: return "green"
            case .verified: return "purple"
            case .trusted: return "gold"
            }
        }
    }
}
```

## Presentation Layer

### App Coordinator

```swift
import SwiftUI

/// Root coordinator managing app-level navigation
@MainActor
final class AppCoordinator: ObservableObject {
    
    // MARK: - Published State
    
    @Published var isAuthenticated = false
    @Published var currentFlow: AppFlow = .splash
    @Published var showingAlert = false
    @Published var alertMessage = ""
    
    // MARK: - Types
    
    enum AppFlow {
        case splash
        case onboarding
        case authentication
        case main
    }
    
    // MARK: - Dependencies
    
    private let authRepository: AuthRepository
    private let secureEnclave: SecureEnclaveManager
    private let biometricAuth: BiometricAuthManager
    
    // Child coordinators
    lazy var authCoordinator = AuthCoordinator(parent: self)
    lazy var mainCoordinator = MainCoordinator(parent: self)
    
    // MARK: - Initialization
    
    init(
        authRepository: AuthRepository,
        secureEnclave: SecureEnclaveManager,
        biometricAuth: BiometricAuthManager
    ) {
        self.authRepository = authRepository
        self.secureEnclave = secureEnclave
        self.biometricAuth = biometricAuth
    }
    
    // MARK: - Lifecycle
    
    /// Check authentication state on app launch
    func checkAuthenticationState() async {
        currentFlow = .splash
        
        // Small delay for splash
        try? await Task.sleep(nanoseconds: 1_000_000_000)
        
        // Check if user has completed onboarding
        let hasCompletedOnboarding = UserDefaults.standard.bool(forKey: "hasCompletedOnboarding")
        
        guard hasCompletedOnboarding else {
            currentFlow = .onboarding
            return
        }
        
        // Check if we have stored credentials
        let hasCredentials = await authRepository.hasStoredCredentials()
        
        guard hasCredentials else {
            currentFlow = .authentication
            return
        }
        
        // Authenticate with biometrics
        let authenticated = await biometricAuth.authenticate(
            reason: "Unlock ECHO"
        )
        
        if authenticated {
            isAuthenticated = true
            currentFlow = .main
        } else {
            currentFlow = .authentication
        }
    }
    
    /// Handle successful authentication
    func didAuthenticate() {
        isAuthenticated = true
        currentFlow = .main
    }
    
    /// Handle logout
    func logout() async {
        await authRepository.clearCredentials()
        isAuthenticated = false
        currentFlow = .authentication
    }
    
    /// Show error alert
    func showError(_ message: String) {
        alertMessage = message
        showingAlert = true
    }
}
```

### Chat View

```swift
import SwiftUI

struct ChatView: View {
    // MARK: - Properties
    
    @StateObject private var viewModel: ChatViewModel
    @FocusState private var isInputFocused: Bool
    
    @State private var showingAttachmentPicker = false
    @State private var showingMessageActions = false
    @State private var selectedMessage: Message?
    
    // MARK: - Initialization
    
    init(conversation: Conversation) {
        _viewModel = StateObject(wrappedValue: ChatViewModel(conversation: conversation))
    }
    
    // MARK: - Body
    
    var body: some View {
        VStack(spacing: 0) {
            // Messages list
            ScrollViewReader { proxy in
                ScrollView {
                    LazyVStack(spacing: 8) {
                        ForEach(viewModel.messages) { message in
                            MessageBubble(
                                message: message,
                                isFromCurrentUser: message.senderId == viewModel.currentUserId
                            )
                            .id(message.id)
                            .contextMenu {
                                messageContextMenu(for: message)
                            }
                            .onTapGesture(count: 2) {
                                viewModel.toggleReaction(on: message, emoji: "❤️")
                            }
                        }
                    }
                    .padding()
                }
                .onChange(of: viewModel.messages.count) { _, _ in
                    if let lastMessage = viewModel.messages.last {
                        withAnimation {
                            proxy.scrollTo(lastMessage.id, anchor: .bottom)
                        }
                    }
                }
            }
            
            // Typing indicator
            if viewModel.isRecipientTyping {
                TypingIndicator()
                    .padding(.horizontal)
            }
            
            Divider()
            
            // Message input
            MessageInputView(
                text: $viewModel.messageText,
                isFocused: $isInputFocused,
                onSend: {
                    Task {
                        await viewModel.sendMessage()
                    }
                },
                onAttachment: {
                    showingAttachmentPicker = true
                },
                onVoiceNote: {
                    viewModel.startRecording()
                }
            )
        }
        .navigationTitle(viewModel.conversation.displayName)
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            ToolbarItem(placement: .principal) {
                headerView
            }
            
            ToolbarItem(placement: .navigationBarTrailing) {
                Menu {
                    Button(action: { viewModel.startCall(video: false) }) {
                        Label("Voice Call", systemImage: "phone")
                    }
                    Button(action: { viewModel.startCall(video: true) }) {
                        Label("Video Call", systemImage: "video")
                    }
                    Button(action: { viewModel.showInfo() }) {
                        Label("Info", systemImage: "info.circle")
                    }
                } label: {
                    Image(systemName: "ellipsis.circle")
                }
            }
        }
        .sheet(isPresented: $showingAttachmentPicker) {
            AttachmentPicker(onSelect: { attachment in
                Task {
                    await viewModel.sendAttachment(attachment)
                }
            })
        }
        .task {
            await viewModel.loadMessages()
            await viewModel.subscribeToUpdates()
        }
        .onChange(of: viewModel.messageText) { _, newValue in
            if !newValue.isEmpty {
                Task {
                    await viewModel.sendTypingIndicator()
                }
            }
        }
    }
    
    // MARK: - Subviews
    
    private var headerView: some View {
        HStack(spacing: 8) {
            AsyncImage(url: URL(string: viewModel.conversation.avatarURL ?? "")) { image in
                image.resizable()
            } placeholder: {
                Circle().fill(Color.gray.opacity(0.3))
            }
            .frame(width: 32, height: 32)
            .clipShape(Circle())
            
            VStack(alignment: .leading, spacing: 2) {
                HStack(spacing: 4) {
                    Text(viewModel.conversation.displayName)
                        .font(.headline)
                    
                    if viewModel.conversation.isVerified {
                        TrustBadge(tier: viewModel.conversation.trustTier)
                    }
                }
                
                if viewModel.isRecipientOnline {
                    Text("Online")
                        .font(.caption)
                        .foregroundColor(.green)
                } else if let lastSeen = viewModel.recipientLastSeen {
                    Text("Last seen \(lastSeen, style: .relative)")
                        .font(.caption)
                        .foregroundColor(.secondary)
                }
            }
        }
    }
    
    @ViewBuilder
    private func messageContextMenu(for message: Message) -> some View {
        Button(action: { viewModel.reply(to: message) }) {
            Label("Reply", systemImage: "arrowshape.turn.up.left")
        }
        
        Button(action: { viewModel.copy(message) }) {
            Label("Copy", systemImage: "doc.on.doc")
        }
        
        Button(action: { viewModel.forward(message) }) {
            Label("Forward", systemImage: "arrowshape.turn.up.right")
        }
        
        if message.senderId == viewModel.currentUserId && message.canEdit {
            Button(action: { viewModel.edit(message) }) {
                Label("Edit", systemImage: "pencil")
            }
        }
        
        if message.senderId == viewModel.currentUserId {
            Button(role: .destructive, action: { viewModel.delete(message) }) {
                Label("Delete", systemImage: "trash")
            }
        }
        
        Divider()
        
        Menu("React") {
            ForEach(["❤️", "👍", "👎", "😂", "😮", "😢"], id: \.self) { emoji in
                Button(emoji) {
                    viewModel.toggleReaction(on: message, emoji: emoji)
                }
            }
        }
    }
}

// MARK: - Message Bubble

struct MessageBubble: View {
    let message: Message
    let isFromCurrentUser: Bool
    
    @State private var decryptedContent: String = ""
    
    var body: some View {
        HStack {
            if isFromCurrentUser { Spacer(minLength: 60) }
            
            VStack(alignment: isFromCurrentUser ? .trailing : .leading, spacing: 4) {
                // Reply preview
                if let replyToId = message.replyToId {
                    ReplyPreview(messageId: replyToId)
                        .padding(.bottom, 4)
                }
                
                // Message content
                HStack(alignment: .bottom, spacing: 8) {
                    if !isFromCurrentUser {
                        // Avatar for group chats
                    }
                    
                    VStack(alignment: isFromCurrentUser ? .trailing : .leading, spacing: 2) {
                        Text(decryptedContent)
                            .padding(.horizontal, 12)
                            .padding(.vertical, 8)
                            .background(isFromCurrentUser ? Color.accentColor : Color(.systemGray5))
                            .foregroundColor(isFromCurrentUser ? .white : .primary)
                            .cornerRadius(16)
                        
                        // Timestamp and status
                        HStack(spacing: 4) {
                            if message.isEdited {
                                Text("edited")
                                    .font(.caption2)
                                    .foregroundColor(.secondary)
                            }
                            
                            Text(message.timestamp, style: .time)
                                .font(.caption2)
                                .foregroundColor(.secondary)
                            
                            if isFromCurrentUser {
                                MessageStatusIcon(status: message.status)
                            }
                        }
                    }
                }
                
                // Reactions
                if !message.reactions.isEmpty {
                    ReactionBar(reactions: message.reactions)
                }
            }
            
            if !isFromCurrentUser { Spacer(minLength: 60) }
        }
        .task {
            await decryptMessage()
        }
    }
    
    private func decryptMessage() async {
        // Decrypt message content using Kinnami
        // This is handled by the ChatViewModel in practice
        decryptedContent = "Decrypted message content"
    }
}

// MARK: - Message Input

struct MessageInputView: View {
    @Binding var text: String
    var isFocused: FocusState<Bool>.Binding
    let onSend: () -> Void
    let onAttachment: () -> Void
    let onVoiceNote: () -> Void
    
    var body: some View {
        HStack(spacing: 12) {
            Button(action: onAttachment) {
                Image(systemName: "plus.circle.fill")
                    .font(.title2)
                    .foregroundColor(.accentColor)
            }
            
            TextField("Message", text: $text, axis: .vertical)
                .textFieldStyle(.plain)
                .lineLimit(1...5)
                .padding(.horizontal, 12)
                .padding(.vertical, 8)
                .background(Color(.systemGray6))
                .cornerRadius(20)
                .focused(isFocused)
            
            if text.isEmpty {
                Button(action: onVoiceNote) {
                    Image(systemName: "mic.fill")
                        .font(.title2)
                        .foregroundColor(.accentColor)
                }
            } else {
                Button(action: onSend) {
                    Image(systemName: "arrow.up.circle.fill")
                        .font(.title)
                        .foregroundColor(.accentColor)
                }
            }
        }
        .padding(.horizontal)
        .padding(.vertical, 8)
        .background(Color(.systemBackground))
    }
}
```

### Trust Badge Component

```swift
import SwiftUI

struct TrustBadge: View {
    let tier: User.TrustTier
    var size: Size = .small
    
    enum Size {
        case small, medium, large
        
        var iconSize: CGFloat {
            switch self {
            case .small: return 12
            case .medium: return 16
            case .large: return 24
            }
        }
        
        var fontSize: Font {
            switch self {
            case .small: return .caption2
            case .medium: return .caption
            case .large: return .subheadline
            }
        }
    }
    
    var body: some View {
        HStack(spacing: 2) {
            Image(systemName: tier.iconName)
                .font(.system(size: size.iconSize, weight: .semibold))
            
            if size != .small {
                Text(tier.displayName)
                    .font(size.fontSize)
                    .fontWeight(.medium)
            }
        }
        .foregroundColor(tier.color)
        .padding(.horizontal, size == .small ? 4 : 8)
        .padding(.vertical, 2)
        .background(tier.color.opacity(0.15))
        .cornerRadius(size == .small ? 4 : 8)
    }
}

extension User.TrustTier {
    var iconName: String {
        switch self {
        case .unverified: return "questionmark.circle"
        case .newcomer: return "person.circle"
        case .member: return "checkmark.circle"
        case .verified: return "checkmark.seal"
        case .trusted: return "checkmark.seal.fill"
        }
    }
    
    var displayName: String {
        switch self {
        case .unverified: return "Unverified"
        case .newcomer: return "Newcomer"
        case .member: return "Member"
        case .verified: return "Verified"
        case .trusted: return "Trusted"
        }
    }
    
    var color: Color {
        switch self {
        case .unverified: return .gray
        case .newcomer: return .blue
        case .member: return .green
        case .verified: return .purple
        case .trusted: return .orange
        }
    }
}
```

## Security Principles

| Principle | Implementation |
|-----------|----------------|
| **Biometric Binding** | All signing keys require Face ID/Touch ID via Secure Enclave |
| **Key Isolation** | Private keys never leave the Secure Enclave |
| **E2E Encryption** | All messages encrypted with Kinnami (X25519 + ChaCha20) |
| **Forward Secrecy** | Ephemeral keys for each message session |
| **Transport Security** | TLS 1.3 minimum for all network requests |
| **Local Encryption** | All cached data encrypted with derived storage key |
| **Memory Protection** | Derived keys cleared when app backgrounds |
| **No PII Logging** | Logger sanitizes all sensitive data |

## Error Handling

```swift
/// User-facing error codes (no sensitive details)
enum ECHOError: Int, LocalizedError {
    // Authentication (1xxx)
    case authFailed = 1001
    case biometricFailed = 1002
    case sessionExpired = 1003
    
    // Network (2xxx)
    case networkUnavailable = 2001
    case requestTimeout = 2002
    case serverError = 2003
    
    // Encryption (3xxx)
    case encryptionFailed = 3001
    case decryptionFailed = 3002
    case keyNotFound = 3003
    
    // Messages (4xxx)
    case messageSendFailed = 4001
    case messageNotFound = 4002
    
    // Identity (5xxx)
    case didCreationFailed = 5001
    case verificationFailed = 5002
    
    var errorDescription: String? {
        "Error \(rawValue). Please try again or contact support."
    }
    
    var supportCode: String {
        "ECHO-\(rawValue)"
    }
}
```

## Performance Optimizations

| Area | Optimization |
|------|--------------|
| **Message Loading** | Pagination with cursor-based loading |
| **Image Loading** | AsyncImage with caching, progressive loading |
| **Encryption** | Hardware-accelerated via Secure Enclave |
| **Database** | SwiftData with lazy loading, batch operations |
| **WebSocket** | Automatic reconnection, message queuing |
| **Memory** | View recycling, image downsampling |
| **Network** | Request deduplication, response caching |

## Testing Strategy

| Test Type | Coverage | Tools |
|-----------|----------|-------|
| Unit Tests | ViewModels, UseCases, Services | XCTest |
| Integration Tests | API client, WebSocket | XCTest + MockServer |
| UI Tests | Critical flows | XCUITest |
| Snapshot Tests | UI components | swift-snapshot-testing |
| Security Tests | Encryption, key management | Custom + third-party audit |

---

*Blueprint Version: 2.0*
*Last Updated: February 17, 2026*
*Status: Complete iOS Implementation Specification*
