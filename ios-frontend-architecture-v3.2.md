# iOS Frontend Architecture (v3.2)

## Changelog

| Version | Date | Changes |
|---------|------|---------|
| 3.2 | March 7, 2026 | Added ECHO Wallet tab architecture built on Stargazer SDK. Added WalletTab, WalletViewModel, BalanceCard, StakingFlow, DelegationFlow, FounderVestingSection components. Resolved wallet decision: native wallet replaces "rewards page" concept. Added founder vesting display. Added token genesis linkage to DID profile. |
| 3.1 | March 7, 2026 | Aligned with Constellation ecosystem. Added: Digital Evidence media fingerprinting for VIP+ users, Smart Checkmark rendering for Org-tier messages, "verified" delivery status for Digital Evidence-anchored messages. Added token management decision note (Stargazer vs in-app). Updated staking references to use Tessellation v3 primitives (TokenLock, StakeDelegation). |
| 3.0 | February 23, 2026 | Aligned with relay architecture decision. Added: MessageRelayManager, offline queue drain on reconnect, sealed sender support (Phase 3), group key management, on-chain anchoring confirmation handling, "anchored" delivery status. Updated encryption spec to canonical table. Removed P2P/libp2p references. Updated WebSocket to handle queue_drain and confirmation message types. |
| 2.0 | February 17, 2026 | Complete iOS implementation specification |

---

## Overview

The frontend is a native iOS application built with SwiftUI and MVVM-C (Model-View-ViewModel-Coordinator) architecture. It provides a secure, user-friendly interface for private messaging, identity management, token rewards, and interaction with the Constellation metagraph through the Go backend relay services.

**Messaging model:** The app sends and receives E2E encrypted message blobs via a stateless WebSocket relay server. The relay cannot read, modify, or forge messages. All message content is encrypted on-device before transmission and decrypted on-device after receipt. The app verifies sender signatures and commitment hashes locally — no trust in the relay is required for content authenticity.

**Security model:** Private keys and passkeys live exclusively in the iOS Secure Enclave and are never extractable. All signing operations require biometric authentication (Face ID / Touch ID). Derived keys handle encryption for messaging (Curve25519), local storage (AES-256-GCM), and session encryption.

## Technology Stack

| Component | Technology | Purpose |
|-----------|------------|---------|
| **UI Framework** | SwiftUI | Declarative UI |
| **Architecture** | MVVM-C | Separation of concerns |
| **Language** | Swift 5.9+ | Type safety, performance |
| **Concurrency** | Swift Concurrency (async/await) | Asynchronous operations |
| **Security** | CryptoKit, Security.framework | Encryption, Secure Enclave |
| **Networking** | URLSession, WebSocket | API & real-time relay |
| **Persistence** | SwiftData, Keychain | Local storage |
| **DI** | Factory pattern | Dependency injection |
| **E2E Encryption** | Kinnami (X25519 + ChaCha20-Poly1305) | Message encryption |
| **Identity Signing** | ECDSA P-256 (Secure Enclave) | DID signing, request signing |
| **Storage Encryption** | AES-256-GCM (HKDF-derived key) | Local data encryption |
| **Transport** | TLS 1.3+ with certificate pinning | Network security |
| **Push** | APNs | Offline message notifications |
| **Wallet** | Constellation Stargazer SDK | Token management, staking (TokenLock), delegation (StakeDelegation), swaps, bridges |
| **Analytics** | Privacy-preserving (no PII) | Usage metrics |

## Encryption Specification (Canonical)

All iOS crypto operations follow this spec (shared with backend and data layer docs):

| Purpose | Algorithm | Key Type | Library |
|---------|-----------|----------|---------|
| Identity/DID signing | ECDSA P-256 | Secure Enclave hardware key | Security.framework |
| Message key agreement | X25519 ECDH | Ephemeral Curve25519 | CryptoKit |
| Message encryption | ChaCha20-Poly1305 | Derived symmetric (256-bit) | CryptoKit |
| Sealed sender envelope (Phase 3) | AES-256-GCM | Derived from recipient identity key | CryptoKit |
| Local storage encryption | AES-256-GCM | Derived from master key via HKDF | CryptoKit |
| Key derivation | HKDF-SHA256 | From Secure Enclave signature | CryptoKit |
| Hash commitments | SHA-256 | N/A | CryptoKit |
| Transport | TLS 1.3 | Certificate-based (pinned) | URLSession |

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                        iOS Application Architecture                  │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                      Presentation Layer                      │   │
│  │  ┌─────────┐  ┌──────────┐  ┌───────────┐  ┌────────────┐  │   │
│  │  │  Views  │  │ViewModels│  │Coordinators│  │UI Components│ │   │
│  │  │(SwiftUI)│  │ (State) │  │(Navigation)│  │             │  │   │
│  │  └─────────┘  └──────────┘  └───────────┘  └────────────┘  │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                              │                                      │
│                              ▼                                      │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                       Domain Layer                           │   │
│  │  ┌─────────┐  ┌─────────┐  ┌──────────┐  ┌─────────────┐  │   │
│  │  │Use Cases│  │ Models  │  │ Protocols│  │ Group Key   │  │   │
│  │  │(Business│  │ (Domain)│  │(Abstract.)│  │ Manager     │  │   │
│  │  │  Logic) │  │         │  │          │  │             │  │   │
│  │  └─────────┘  └─────────┘  └──────────┘  └─────────────┘  │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                              │                                      │
│                              ▼                                      │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                        Data Layer                            │   │
│  │  ┌─────────┐  ┌──────────┐  ┌──────────┐  ┌────────────┐  │   │
│  │  │  API    │  │ WebSocket│  │  Local   │  │  Secure    │  │   │
│  │  │ Client  │  │  Relay   │  │ Storage  │  │  Enclave   │  │   │
│  │  │         │  │  Client  │  │          │  │            │  │   │
│  │  └─────────┘  └──────────┘  └──────────┘  └────────────┘  │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                              │                                      │
│                              ▼                                      │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                      Security Layer                          │   │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌────────────┐ │   │
│  │  │ Kinnami  │  │Biometric │  │  Key     │  │  Secure    │ │   │
│  │  │Encryption│  │  Auth    │  │ Manager  │  │  Storage   │ │   │
│  │  └──────────┘  └──────────┘  └──────────┘  └────────────┘ │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                              │                                      │
│                              ▼                                      │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                 Relay Server (Content-Blind)                  │   │
│  │  Server sees: encrypted blobs, recipient DID, timestamps     │   │
│  │  Server CANNOT: read, decrypt, modify, or forge messages     │   │
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
│   │   └── KinnamiEncryption.swift      # E2E encryption (X25519 + ChaCha20)
│   │
│   ├── Networking/
│   │   ├── APIClient.swift              # REST client
│   │   ├── WebSocketRelay.swift         # Real-time message relay + queue drain
│   │   ├── Endpoints.swift              # API endpoints
│   │   ├── RequestInterceptor.swift     # Auth, encryption
│   │   └── CertificatePinner.swift      # TLS certificate pinning
│   │
│   ├── Relay/
│   │   ├── MessageRelayManager.swift    # Coordinates send/receive via relay
│   │   ├── OfflineQueueManager.swift    # Local outbox for offline sends
│   │   ├── SealedSenderService.swift    # Phase 3: sender-anonymous envelopes
│   │   └── AnchoringTracker.swift       # Tracks commitment → on-chain confirmation
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
│   │   ├── Message.swift                # Message model (with .anchored status)
│   │   ├── Conversation.swift           # Conversation model
│   │   ├── DID.swift                    # Decentralized identity
│   │   ├── Credential.swift             # Verifiable credential
│   │   ├── Token.swift                  # ECHO token
│   │   └── GroupKey.swift               # Group encryption key management
│   │
│   ├── UseCases/
│   │   ├── Auth/
│   │   │   ├── AuthenticateUseCase.swift
│   │   │   ├── RegisterUseCase.swift
│   │   │   └── PasskeyUseCase.swift
│   │   │
│   │   ├── Messaging/
│   │   │   ├── SendMessageUseCase.swift       # Encrypt → sign → relay
│   │   │   ├── ReceiveMessageUseCase.swift    # Decrypt → verify → display
│   │   │   ├── FetchMessagesUseCase.swift
│   │   │   ├── EncryptMessageUseCase.swift
│   │   │   └── VerifyAnchoringUseCase.swift   # Verify Merkle proof (Phase 3)
│   │   │
│   │   ├── Groups/
│   │   │   ├── CreateGroupUseCase.swift
│   │   │   ├── ManageGroupKeyUseCase.swift    # Key rotation on member change
│   │   │   └── GroupFanOutUseCase.swift
│   │   │
│   │   ├── Identity/
│   │   │   ├── CreateDIDUseCase.swift
│   │   │   ├── VerifyIdentityUseCase.swift
│   │   │   └── ManageCredentialsUseCase.swift
│   │   │
│   │   └── Tokens/
│   │       ├── GetBalanceUseCase.swift
│   │       ├── SendTokensUseCase.swift
│   │       ├── StakeTokensUseCase.swift
│   │       └── ClaimRewardsUseCase.swift
│   │
│   └── Repositories/
│       ├── AuthRepository.swift
│       ├── MessageRepository.swift        # Uses MessageRelayManager
│       ├── UserRepository.swift
│       ├── TokenRepository.swift
│       └── GroupRepository.swift
│
├── Presentation/
│   ├── Coordinators/
│   │   ├── AppCoordinator.swift
│   │   ├── AuthCoordinator.swift
│   │   ├── MainCoordinator.swift
│   │   └── SettingsCoordinator.swift
│   │
│   ├── Features/
│   │   ├── Auth/                          # Login, passkey setup, biometric
│   │   ├── Onboarding/                    # Onboarding, VC setup, ID verification
│   │   ├── Conversations/                 # Conversation list, new conversation
│   │   ├── Chat/                          # Chat view, message bubbles, input
│   │   ├── Profile/                       # Profile, trust score, credentials
│   │   ├── Wallet/                        # Balance, send, stake, history
│   │   ├── Groups/                        # Group management, member list
│   │   └── Settings/                      # Privacy, security, notifications
│   │
│   └── Components/
│       ├── Buttons/
│       ├── Inputs/
│       ├── Cards/
│       └── Indicators/
│           ├── TrustBadge.swift
│           ├── VerificationBadge.swift
│           └── AnchorStatusIndicator.swift  # Shows pending/anchored status
│
└── Resources/
    ├── Assets.xcassets
    ├── Localizable.strings
    └── Info.plist
```

## New Components (v3.0)

### MessageRelayManager

Coordinates all message send/receive operations through the WebSocket relay:

```swift
/// Manages message relay through the stateless WebSocket server.
/// The relay server transports E2E encrypted blobs it cannot read.
actor MessageRelayManager {
    
    private let webSocket: WebSocketRelay
    private let encryption: KinnamiEncryptionService
    private let secureEnclave: SecureEnclaveManager
    private let offlineQueue: OfflineQueueManager
    private let anchoringTracker: AnchoringTracker
    
    // MARK: - Send Flow
    
    /// Full message send pipeline: encrypt → commit → sign → relay
    func sendMessage(
        plaintext: Data,
        contentType: Message.ContentType,
        recipientPublicKey: Data,
        conversationId: String
    ) async throws -> Message {
        // 1. E2E encrypt with Kinnami (X25519 + ChaCha20-Poly1305)
        let encryptedPayload = try encryption.encrypt(
            plaintext: plaintext,
            recipientPublicKey: recipientPublicKey
        )
        
        // 2. Sign the encrypted payload with Secure Enclave (P-256)
        let signature = try await secureEnclave.sign(
            data: encryptedPayload.serialized,
            reason: "Send message"
        )
        
        // 3. Submit to relay via WebSocket
        let messageId = UUID().uuidString
        let request = SendMessageRequest(
            messageId: messageId,
            conversationId: conversationId,
            contentType: contentType,
            encryptedPayload: encryptedPayload,
            signature: signature
        )
        
        do {
            let response = try await webSocket.sendMessage(request)
            
            // 4. Track commitment for on-chain anchoring confirmation
            anchoringTracker.track(
                messageId: messageId,
                commitment: encryptedPayload.commitment
            )
            
            return Message(
                id: messageId,
                status: response.status == "relayed" ? .delivered : .sent
            )
        } catch {
            // 5. If relay unavailable, queue locally for retry
            try offlineQueue.enqueue(request)
            return Message(id: messageId, status: .sending)
        }
    }
    
    // MARK: - Receive Flow
    
    /// Process incoming encrypted message from relay
    func receiveMessage(
        encryptedPayload: KinnamiEncryptionService.EncryptedPayload,
        senderDID: String,
        senderPublicKey: Data,
        signature: Data
    ) async throws -> Data {
        // 1. Verify sender signature (P-256)
        let isValid = try await verifySenderSignature(
            payload: encryptedPayload.serialized,
            signature: signature,
            senderPublicKey: senderPublicKey
        )
        guard isValid else {
            throw MessageError.invalidSignature
        }
        
        // 2. Decrypt with own private key (Kinnami)
        let privateKey = try await getMessagingPrivateKey()
        let plaintext = try encryption.decrypt(
            payload: encryptedPayload,
            privateKey: privateKey
        )
        
        // 3. Verify commitment integrity
        // (Already done inside decrypt — commitment check is built in)
        
        return plaintext
    }
    
    // MARK: - Offline Queue Drain
    
    /// Called on WebSocket reconnect — drain any queued outbound messages
    func drainOfflineQueue() async {
        let queuedMessages = offlineQueue.dequeueAll()
        for request in queuedMessages {
            do {
                _ = try await webSocket.sendMessage(request)
            } catch {
                // Re-queue if still failing
                try? offlineQueue.enqueue(request)
            }
        }
    }
}
```

### AnchoringTracker

Tracks message commitments and updates delivery status when on-chain confirmation arrives:

```swift
/// Tracks message commitment hashes and updates status when
/// the metagraph confirms anchoring in a finalized snapshot.
@MainActor
final class AnchoringTracker: ObservableObject {
    
    /// Messages pending on-chain anchoring
    @Published private(set) var pendingAnchors: [String: PendingAnchor] = [:]
    
    struct PendingAnchor {
        let messageId: String
        let commitment: Data
        let submittedAt: Date
    }
    
    /// Track a new message commitment
    func track(messageId: String, commitment: Data) {
        pendingAnchors[messageId] = PendingAnchor(
            messageId: messageId,
            commitment: commitment,
            submittedAt: Date()
        )
    }
    
    /// Called when WebSocket receives a confirmation from the relay
    /// (type: "message_anchored" with snapshotHash and optional Merkle proof)
    func confirmAnchoring(
        messageId: String,
        snapshotHash: String,
        snapshotHeight: Int,
        merkleProof: [Data]?
    ) {
        pendingAnchors.removeValue(forKey: messageId)
        
        // Phase 3+: Verify Merkle proof locally
        if let proof = merkleProof {
            // verifyMerkleInclusion(commitment, proof, snapshotHash)
        }
        
        // Update message delivery status to .anchored
        NotificationCenter.default.post(
            name: .messageAnchored,
            object: nil,
            userInfo: [
                "messageId": messageId,
                "snapshotHash": snapshotHash,
                "snapshotHeight": snapshotHeight
            ]
        )
    }
}
```

### GroupKeyManager

Manages group symmetric key lifecycle:

```swift
/// Manages group encryption keys.
/// Group keys are symmetric (AES-256) and distributed to members
/// via individually encrypted E2E messages.
actor GroupKeyManager {
    
    private let encryption: KinnamiEncryptionService
    private let keychain: KeychainManager
    
    struct GroupKeyInfo {
        let groupId: String
        let key: SymmetricKey
        let version: Int
        let receivedAt: Date
    }
    
    /// Generate a new group key (called by group admin)
    func generateGroupKey(groupId: String) -> GroupKeyInfo {
        let key = encryption.generateSymmetricKey()
        let version = (getLatestKeyVersion(groupId: groupId) ?? 0) + 1
        let info = GroupKeyInfo(
            groupId: groupId, key: key,
            version: version, receivedAt: Date()
        )
        storeGroupKey(info)
        return info
    }
    
    /// Encrypt group key for each member (admin distributes via relay)
    func encryptGroupKeyForMembers(
        groupKey: SymmetricKey,
        memberPublicKeys: [(did: String, publicKey: Data)]
    ) throws -> [(did: String, encryptedKey: Data)] {
        return try memberPublicKeys.map { member in
            let keyData = groupKey.withUnsafeBytes { Data($0) }
            let encrypted = try encryption.encrypt(
                plaintext: keyData,
                recipientPublicKey: member.publicKey
            )
            return (did: member.did, encryptedKey: encrypted.serialized)
        }
    }
    
    /// Decrypt a received group key (member receives via relay)
    func receiveGroupKey(
        encryptedKey: Data,
        groupId: String,
        version: Int,
        privateKey: Curve25519.KeyAgreement.PrivateKey
    ) throws -> GroupKeyInfo {
        let payload = try KinnamiEncryptionService.EncryptedPayload
            .deserialize(from: encryptedKey)
        let keyData = try encryption.decrypt(payload: payload, privateKey: privateKey)
        let key = SymmetricKey(data: keyData)
        let info = GroupKeyInfo(
            groupId: groupId, key: key,
            version: version, receivedAt: Date()
        )
        storeGroupKey(info)
        return info
    }
    
    /// Encrypt a group message with the current group key
    func encryptForGroup(plaintext: Data, groupId: String) throws -> Data {
        guard let keyInfo = getLatestKey(groupId: groupId) else {
            throw GroupError.noGroupKey
        }
        return try encryption.encryptForStorage(plaintext: plaintext, key: keyInfo.key)
    }
    
    /// Decrypt a group message
    func decryptFromGroup(ciphertext: Data, groupId: String, keyVersion: Int) throws -> Data {
        guard let keyInfo = getKey(groupId: groupId, version: keyVersion) else {
            throw GroupError.noGroupKey
        }
        return try encryption.decryptFromStorage(ciphertext: ciphertext, key: keyInfo.key)
    }
    
    // MARK: - Private storage methods
    private func storeGroupKey(_ info: GroupKeyInfo) { /* Keychain storage */ }
    private func getLatestKey(groupId: String) -> GroupKeyInfo? { return nil }
    private func getKey(groupId: String, version: Int) -> GroupKeyInfo? { return nil }
    private func getLatestKeyVersion(groupId: String) -> Int? { return nil }
}
```

## Updated WebSocket Relay Client

Key changes from v2: handles `queue_drain` (offline messages on reconnect), `confirmation` (on-chain anchoring), and `group_key` message types.

```swift
/// Updated WSMessage types for relay model
struct WSMessage: Codable {
    let type: MessageType
    let payload: Data
    let timestamp: Date
    
    enum MessageType: String, Codable {
        case message            // E2E encrypted message blob
        case typing             // Typing indicator
        case presence           // Online/offline status
        case receipt            // Read/delivery receipt
        case ack                // Server acknowledgement
        case queueDrain         // Offline queue delivery on reconnect
        case confirmation       // On-chain anchoring confirmation
        case groupKey           // Group key distribution
    }
}
```

On reconnect, the relay server automatically drains the offline queue:

```swift
// In WebSocketRelay.connect():
// After successful reconnect, server sends type=queueDrain messages
// for all messages received while this client was offline.
// The app processes these through the normal receiveMessage flow.

private func handleMessage(_ wsMessage: WSMessage) async {
    switch wsMessage.type {
    case .message, .queueDrain:
        // Decrypt and display (same flow for live and queued messages)
        await messageRelayManager.handleIncomingMessage(wsMessage.payload)
        
    case .confirmation:
        // On-chain anchoring confirmed
        let confirmation = try? JSONDecoder().decode(
            WSConfirmation.self, from: wsMessage.payload
        )
        if let conf = confirmation {
            await anchoringTracker.confirmAnchoring(
                messageId: conf.referenceId,
                snapshotHash: conf.snapshotHash,
                snapshotHeight: conf.snapshotHeight,
                merkleProof: conf.merkleProof
            )
        }
        
    case .groupKey:
        // New group key received (member add/remove triggered rotation)
        await groupKeyManager.handleKeyDistribution(wsMessage.payload)
        
    case .typing:
        // Update typing indicator UI
        break
        
    case .presence:
        // Update contact online status
        break
        
    case .receipt:
        // Update message delivery/read status
        break
        
    case .ack:
        break
    }
}
```

## Updated Message Delivery Status

```swift
enum DeliveryStatus: String, Codable {
    case sending        // Encrypting / queued locally (offline)
    case sent           // Accepted by relay, recipient offline (queued on server)
    case delivered       // Delivered to recipient's device
    case read           // Recipient opened the message
    case failed         // Relay rejected or unrecoverable error
    case anchored       // Commitment included in finalized metagraph snapshot
    case verified       // Digital Evidence fingerprint anchored (Org tier + Smart Checkmark)
}
```

The `anchored` status is displayed as a subtle chain-link icon (🔗) next to the message timestamp, indicating blockchain-verified integrity via ECHO's Merkle root pipeline. This is available for all users.

The `verified` status is displayed as a Smart Checkmark badge (✓) next to the message — indicating the message has been individually fingerprinted via Constellation's Digital Evidence API with a public verification URL. Tapping the checkmark opens the Digital Evidence Explorer verification page in Safari. This is available for Organization tier senders and optional for VIP users who fingerprint media.

## New Components (v3.1)

### DigitalEvidenceBridge

Handles client-side Digital Evidence interactions: media fingerprinting before E2E encryption, Smart Checkmark rendering, and verification URL management.

```swift
// DigitalEvidenceBridge.swift

import CryptoKit

actor DigitalEvidenceBridge {
    private let backendAPI: BackendAPIClient
    
    /// Fingerprint media before E2E encryption (VIP+ users, optional).
    /// Computes SHA-256 of raw media data, submits to backend which forwards to Digital Evidence API.
    func fingerprintMedia(_ mediaData: Data, messageId: String) async throws -> EvidenceResult? {
        let hash = SHA256.hash(data: mediaData)
        let hashHex = hash.compactMap { String(format: "%02x", $0) }.joined()
        
        let result = try await backendAPI.submitEvidenceFingerprint(
            contentHash: hashHex,
            messageId: messageId,
            metadata: ["type": "media", "source": "echo_ios"]
        )
        
        return EvidenceResult(
            eventId: result.eventId,
            verificationURL: result.verificationUrl,
            timestamp: result.timestamp
        )
    }
    
    /// Render Smart Checkmark for verified messages.
    func verificationURL(for message: Message) -> URL? {
        guard let eventId = message.evidenceEventId else { return nil }
        return URL(string: "https://digitalevidence.constellationnetwork.io/verify/\(eventId)")
    }
}

struct EvidenceResult: Codable {
    let eventId: String
    let verificationURL: String
    let timestamp: Date
}
```

### Token Management Decision

**RESOLVED: Build native ECHO Wallet inside iOS app using Stargazer SDK.**

A "rewards page" implies gamification points. A "wallet" implies real ownership. For ECHO's "all users are owners" thesis, the wallet framing is essential.

### WalletTab (Stargazer SDK Integration)

The ECHO iOS app adds a Wallet tab as a primary navigation destination alongside Messages and Profile. Built on the Constellation Stargazer Wallet SDK for key management, transaction signing, and L0 token operations.

```swift
// EchoWallet/WalletTab.swift

import SwiftUI
import StargazerSDK

struct WalletTab: View {
    @StateObject private var viewModel = WalletViewModel()
    
    var body: some View {
        NavigationStack {
            ScrollView {
                BalanceCard(
                    total: viewModel.totalBalance,
                    usdValue: viewModel.usdValue
                )
                
                BalanceBreakdown(
                    available: viewModel.available,
                    staked: viewModel.staked,
                    delegatedTo: viewModel.delegatedValidator,
                    pending: viewModel.pendingRewards
                )
                
                HStack(spacing: 12) {
                    WalletActionButton(icon: "lock.fill", label: "Stake") {
                        viewModel.showStaking = true
                    }
                    WalletActionButton(icon: "arrow.up.right", label: "Delegate") {
                        viewModel.showDelegation = true
                    }
                    WalletActionButton(icon: "arrow.left.arrow.right", label: "Swap") {
                        viewModel.showSwap = true
                    }
                    WalletActionButton(icon: "link", label: "Bridge") {
                        viewModel.showBridge = true
                    }
                }
                
                DailyRewardsSection(rewards: viewModel.dailyRewards)
                
                // Founder vesting — only visible if DID has a founder TokenLock
                if let vesting = viewModel.founderVesting {
                    FounderVestingSection(vesting: vesting)
                }
                
                RecentActivityList(activity: viewModel.recentActivity)
            }
            .navigationTitle("ECHO Wallet")
            .task { await viewModel.loadWallet() }
        }
    }
}
```

### WalletViewModel

```swift
// EchoWallet/WalletViewModel.swift

@MainActor
class WalletViewModel: ObservableObject {
    private let stargazer: StargazerClient
    private let backendAPI: BackendAPIClient
    
    @Published var totalBalance: Decimal = 0
    @Published var available: Decimal = 0
    @Published var staked: Decimal = 0
    @Published var pendingRewards: Decimal = 0
    @Published var delegatedValidator: ValidatorInfo?
    @Published var founderVesting: VestingInfo?
    @Published var dailyRewards: DailyRewards = .empty
    @Published var recentActivity: [WalletActivity] = []
    
    func loadWallet() async {
        // Balance from Stargazer SDK (reads metagraph state)
        let balance = try? await stargazer.getBalance(token: .echo)
        self.totalBalance = balance?.total ?? 0
        self.available = balance?.available ?? 0
        
        // TokenLock positions (staking)
        let locks = try? await stargazer.getTokenLocks(token: .echo)
        self.staked = locks?.reduce(0) { $0 + $1.amount } ?? 0
        
        // StakeDelegation positions
        let delegations = try? await stargazer.getDelegations(token: .echo)
        self.delegatedValidator = delegations?.first?.validator
        
        // Pending rewards from backend cache
        let rewards = try? await backendAPI.getPendingRewards()
        self.pendingRewards = rewards?.total ?? 0
        self.dailyRewards = rewards?.daily ?? .empty
        
        // Founder vesting TokenLock (has cliff/vest metadata)
        if let founderLock = locks?.first(where: { $0.isFounderVesting }) {
            self.founderVesting = VestingInfo(from: founderLock)
        }
    }
    
    /// Claim rewards via AtomicAction (verify tier + claim + update cap)
    func claimRewards() async throws {
        try await stargazer.submitAtomicAction([
            .verifyTrustTier(did: currentDID),
            .claimRewards(did: currentDID, types: dailyRewards.claimableTypes),
            .updateDailyCap(did: currentDID)
        ])
        await loadWallet()
    }
    
    /// Stake ECHO via TokenLock
    func stakeEcho(amount: Decimal, tier: StakingTier) async throws {
        try await stargazer.submitTokenLock(TokenLockRequest(
            token: .echo, amount: amount,
            tier: tier.rawValue, duration: tier.durationDays
        ))
        await loadWallet()
    }
    
    /// Delegate staked ECHO to validator via StakeDelegation
    func delegateToValidator(_ validatorId: String, stakeId: String) async throws {
        try await stargazer.submitStakeDelegation(StakeDelegationRequest(
            stakeId: stakeId, validatorId: validatorId
        ))
        await loadWallet()
    }
    
    /// Withdraw vested founder tokens via WithdrawLock (14-day cooldown)
    func withdrawVestedTokens(amount: Decimal) async throws {
        guard let vesting = founderVesting, amount <= vesting.withdrawable else {
            throw WalletError.insufficientVestedBalance
        }
        try await stargazer.submitWithdrawLock(WithdrawLockRequest(amount: amount))
        await loadWallet()
    }
}
```

### FounderVestingSection

Visible only for founders (detected by founder TokenLock type on DID). All vesting data is on-chain and publicly verifiable.

```swift
// EchoWallet/FounderVestingSection.swift

struct FounderVestingSection: View {
    let vesting: VestingInfo
    
    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            Text("Founder Allocation")
                .font(.headline)
            
            HStack {
                VStack(alignment: .leading) {
                    Label("Allocated", systemImage: "banknote")
                    Text(vesting.totalAllocated.formatted())
                        .font(.title3.bold())
                }
                Spacer()
                VStack(alignment: .trailing) {
                    Label("Vested", systemImage: "checkmark.circle")
                    Text(vesting.vested.formatted())
                        .font(.title3.bold())
                        .foregroundStyle(.green)
                }
            }
            
            ProgressView(value: vesting.vestingProgress)
                .tint(.green)
            
            Text("\(vesting.vestingPercentage)% vested")
                .font(.caption)
                .foregroundStyle(.secondary)
            
            if vesting.cliffCompleted {
                LabeledContent("Next unlock", value: vesting.nextUnlockAmount.formatted())
                LabeledContent("Unlock date", value: vesting.nextUnlockDate.formatted(date: .abbreviated, time: .omitted))
            } else {
                LabeledContent("Cliff date", value: vesting.cliffDate.formatted(date: .abbreviated, time: .omitted))
                Text("No tokens vest until cliff completes")
                    .font(.caption)
                    .foregroundStyle(.orange)
            }
            
            if vesting.withdrawable > 0 {
                Button("Withdraw \(vesting.withdrawable.formatted()) ECHO") {
                    // Triggers WithdrawLock flow
                }
                .buttonStyle(.borderedProminent)
            }
            
            Link("View on DAG Explorer", destination: vesting.explorerURL)
                .font(.caption)
            
            Text("Founder vesting is on-chain and publicly verifiable by all ECHO users.")
                .font(.caption2)
                .foregroundStyle(.secondary)
        }
        .padding()
        .background(.ultraThinMaterial, in: RoundedRectangle(cornerRadius: 12))
    }
}

struct VestingInfo {
    let totalAllocated: Decimal
    let vested: Decimal
    let locked: Decimal
    let nextUnlockAmount: Decimal
    let nextUnlockDate: Date
    let cliffCompleted: Bool
    let cliffDate: Date
    let withdrawable: Decimal
    let explorerURL: URL
    
    var vestingProgress: Double {
        Double(truncating: (vested / totalAllocated) as NSDecimalNumber)
    }
    var vestingPercentage: String {
        String(format: "%.1f", vestingProgress * 100)
    }
    
    init(from lock: TokenLockPosition) {
        self.totalAllocated = lock.originalAmount
        self.vested = lock.vestedAmount
        self.locked = lock.lockedAmount
        self.nextUnlockAmount = lock.nextUnlockAmount
        self.nextUnlockDate = lock.nextUnlockDate
        self.cliffCompleted = lock.cliffCompleted
        self.cliffDate = lock.cliffDate
        self.withdrawable = lock.withdrawableAmount
        self.explorerURL = URL(string: "https://dagexplorer.io/address/\(lock.walletAddress)")!
    }
}
```

### StakingTier

```swift
enum StakingTier: String, CaseIterable, Identifiable {
    case bronze, silver, gold, platinum
    
    var id: String { rawValue }
    
    var durationDays: Int {
        switch self {
        case .bronze: return 30
        case .silver: return 90
        case .gold: return 180
        case .platinum: return 365
        }
    }
    
    var aprPercentage: Double {
        switch self {
        case .bronze: return 5.0
        case .silver: return 8.0
        case .gold: return 12.0
        case .platinum: return 15.0
        }
    }
    
    var displayName: String {
        switch self {
        case .bronze: return "Bronze — 30 days, 5% APR"
        case .silver: return "Silver — 90 days, 8% APR"
        case .gold: return "Gold — 180 days, 12% APR"
        case .platinum: return "Platinum — 365 days, 15% APR"
        }
    }
}
```

## Security Principles

| Principle | Implementation |
|-----------|----------------|
| **Biometric Binding** | All signing keys require Face ID/Touch ID via Secure Enclave |
| **Key Isolation** | Private keys never leave the Secure Enclave |
| **E2E Encryption** | All messages encrypted with Kinnami (X25519 + ChaCha20-Poly1305) |
| **Content-Blind Relay** | Relay server transports opaque ciphertext; cannot read or modify |
| **Client-Side Verification** | Recipient verifies sender signature + commitment hash locally |
| **Forward Secrecy** | Ephemeral keys for each message session |
| **Transport Security** | TLS 1.3 minimum with certificate pinning for all connections |
| **Local Encryption** | All cached data encrypted with HKDF-derived storage key |
| **Memory Protection** | Derived keys cleared when app backgrounds |
| **No PII Logging** | Logger sanitizes all sensitive data |
| **Offline Resilience** | Local outbox queues encrypted messages when relay unavailable |

## Performance Optimizations

| Area | Optimization |
|------|--------------|
| **Message Loading** | Pagination with cursor-based loading |
| **Image Loading** | AsyncImage with caching, progressive loading |
| **Encryption** | Hardware-accelerated via Secure Enclave |
| **Database** | SwiftData with lazy loading, batch operations |
| **WebSocket** | Automatic reconnection with exponential backoff + offline queue drain |
| **Memory** | View recycling, image downsampling |
| **Network** | Request deduplication, response caching |
| **Group Messages** | Group symmetric key avoids per-recipient re-encryption for large groups |
| **Anchoring** | Background tracking; non-blocking UI updates on confirmation |

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
    case relayUnavailable = 2004    // WebSocket relay disconnected
    
    // Encryption (3xxx)
    case encryptionFailed = 3001
    case decryptionFailed = 3002
    case keyNotFound = 3003
    case invalidSignature = 3004   // Sender signature verification failed
    case commitmentMismatch = 3005 // Commitment hash doesn't match content
    
    // Messages (4xxx)
    case messageSendFailed = 4001
    case messageNotFound = 4002
    case rateLimitExceeded = 4003  // Relay rate limit hit
    case messageQueued = 4004      // Not an error; message queued for offline send
    
    // Identity (5xxx)
    case didCreationFailed = 5001
    case verificationFailed = 5002
    
    // Groups (6xxx)
    case groupKeyMissing = 6001
    case groupKeyRotationFailed = 6002
    case notGroupAdmin = 6003
    
    var errorDescription: String? {
        "Error \(rawValue). Please try again or contact support."
    }
    
    var supportCode: String {
        "ECHO-\(rawValue)"
    }
}
```

## Testing Strategy

| Test Type | Coverage | Tools |
|-----------|----------|-------|
| Unit Tests | ViewModels, UseCases, Services, RelayManager, GroupKeyManager | XCTest |
| Integration Tests | API client, WebSocket relay, offline queue drain | XCTest + MockServer |
| UI Tests | Critical flows (send message, verify identity, claim reward) | XCUITest |
| Snapshot Tests | UI components, anchoring indicator, trust badge | swift-snapshot-testing |
| Security Tests | Encryption, key management, signature verification, sealed sender | Custom + third-party audit |
| Relay Tests | Offline queue, reconnect drain, rate limit handling | XCTest + MockRelay |

---

*Blueprint Version: 3.0*
*Last Updated: February 23, 2026*
*Status: Aligned with relay architecture decision and Data Layer v3.0*

**Note:** This document covers the structural and architectural changes for v3.0. All existing v2.0 implementation code (SecureEnclaveManager, BiometricAuthManager, KinnamiEncryptionService, APIClient, domain models, presentation layer views) remains valid and is carried forward unchanged except where noted in the changelog.
