# Frontend Architecture

## Overview

The frontend is a native iOS application built with SwiftUI and MVVM-C (Model-View-ViewModel-Coordinator) architecture. It provides a secure, user-friendly interface for private messaging, identity management, token rewards, and interaction with the Constellation metagraph through the Go backend relay services.

**Messaging model:** The app sends and receives E2E encrypted message blobs via a stateless WebSocket relay server. The relay cannot read, modify, or forge messages. All message content is encrypted on-device before transmission and decrypted on-device after receipt. The app verifies sender signatures and commitment hashes locally—no trust in the relay is required for content authenticity.

**Security model:** Private keys and passkeys live exclusively in the iOS Secure Enclave and are never extractable. All signing operations require biometric authentication (Face ID / Touch ID). Derived keys handle encryption for messaging (Curve25519), local storage (AES-256-GCM), and session encryption.

## Technology Stack

| Component | Technology | Purpose |
| --- | --- | --- |
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
| **ZK Proofs** | Midnight SDK (Phase 3+) | Zero-knowledge credential verification |
| **Storage Encryption** | AES-256-GCM (HKDF-derived key) | Local data encryption |
| **Transport** | TLS 1.3+ with certificate pinning | Network security |
| **Push** | APNs | Offline message notifications |
| **Wallet SDK** | Stargazer SDK | Native Constellation L0 token wallet |
| **DEX Integration** | PacaSwap SDK (Phase 3+) | On-chain ECHO/DAG and ECHO/USDC swaps |
| **Bridge Integration** | Base & Ink bridges (Phase 3+) | Cross-chain token transfers |
| **Analytics** | Privacy-preserving (no PII) | Usage metrics |

## Encryption Specification (Canonical)

All iOS crypto operations follow this spec (shared with Backend and Data Layer blueprints):

| Purpose | Algorithm | Key Type | Library |
| --- | --- | --- | --- |
| Identity/DID signing | ECDSA P-256 | Secure Enclave hardware key | Security.framework |
| Message key agreement | X25519 ECDH | Ephemeral Curve25519 | CryptoKit |
| Message encryption | ChaCha20-Poly1305 | Derived symmetric (256-bit) | CryptoKit |
| Sealed sender envelope (Phase 3) | AES-256-GCM | Derived from recipient identity key | CryptoKit |
| Local storage encryption | AES-256-GCM | Derived from master key via HKDF | CryptoKit |
| Key derivation | HKDF-SHA256 | From Secure Enclave signature | CryptoKit |
| Hash commitments | SHA-256 | N/A | CryptoKit |
| Contact discovery hashing | Argon2id | Per-user salt (device-local) | Swift Argon2 |
| Transport | TLS 1.3 | Certificate-based (pinned) | URLSession |

## Architecture Overview

```
┌───────────────────────────────────────────────────────┐
│              iOS Application Architecture              │
├───────────────────────────────────────────────────────┤
│                                                       │
│  ┌─────────────────────────────────────────────────┐ │
│  │            Presentation Layer                    │ │
│  │  ┌─────┐  ┌──────┐  ┌──────┐  ┌──────────────┐ │ │
│  │  │Views│  │VMs   │  │Coords│  │UI Components │ │ │
│  │  └─────┘  └──────┘  └──────┘  └──────────────┘ │ │
│  └─────────────────────────────────────────────────┘ │
│                        │                              │
│                        ▼                              │
│  ┌─────────────────────────────────────────────────┐ │
│  │            Domain Layer                          │ │
│  │  ┌─────────┐  ┌──────┐  ┌─────────────────────┐ │ │
│  │  │UseCases │  │Models│  │Group Key Manager    │ │ │
│  │  └─────────┘  └──────┘  └─────────────────────┘ │ │
│  └─────────────────────────────────────────────────┘ │
│                        │                              │
│                        ▼                              │
│  ┌─────────────────────────────────────────────────┐ │
│  │            Data Layer                            │ │
│  │  ┌─────┐  ┌────────┐  ┌──────┐  ┌────────────┐ │ │
│  │  │API  │  │WebSocket│  │Local │  │Secure      │ │ │
│  │  │     │  │Relay   │  │Store │  │Enclave     │ │ │
│  │  └─────┘  └────────┘  └──────┘  └────────────┘ │ │
│  └─────────────────────────────────────────────────┘ │
│                        │                              │
│                        ▼                              │
│  ┌─────────────────────────────────────────────────┐ │
│  │         Relay Server (Content-Blind)             │ │
│  │  Server sees: encrypted blobs, recipient DID     │ │
│  │  Server CANNOT: read, decrypt, modify messages   │ │
│  └─────────────────────────────────────────────────┘ │
│                                                       │
└───────────────────────────────────────────────────────┘
```

## Project Structure

```
ECHO/
├── App/
│   ├── ECHOApp.swift              # App entry point
│   ├── AppDelegate.swift          # Push notifications, lifecycle
│   └── SceneDelegate.swift        # Scene management
│
├── Core/
│   ├── DI/
│   │   ├── Container.swift        # Dependency container
│   │   └── Factories/             # Service factories
│   │
│   ├── Security/
│   │   ├── SecureEnclaveManager.swift    # Secure Enclave ops
│   │   ├── BiometricAuthManager.swift    # Face ID / Touch ID
│   │   ├── KeychainManager.swift         # Keychain wrapper
│   │   └── KinnamiEncryption.swift       # E2E (X25519 + ChaCha20)
│   │
│   ├── Networking/
│   │   ├── APIClient.swift               # REST client
│   │   ├── WebSocketRelay.swift          # Real-time relay + queue
│   │   ├── Endpoints.swift               # API endpoints
│   │   ├── RequestInterceptor.swift      # Auth, encryption
│   │   └── CertificatePinner.swift       # TLS pinning
│   │
│   ├── Relay/
│   │   ├── MessageRelayManager.swift     # Send/receive via relay
│   │   ├── OfflineQueueManager.swift     # Local outbox for offline
│   │   ├── SealedSenderService.swift     # Phase 3: sender-anonymous
│   │   └── AnchoringTracker.swift        # Track commitment → on-chain
│   │
│   ├── Storage/
│   │   ├── LocalDatabase.swift           # SwiftData setup
│   │   ├── SecureStorage.swift           # Encrypted storage
│   │   └── CacheManager.swift            # Caching layer
│   │
│   └── Utilities/
│       ├── Logger.swift                  # Privacy-safe logging
│       ├── Constants.swift               # App constants
│       └── Extensions/                   # Swift extensions
│
├── Domain/
│   ├── Models/
│   │   ├── User.swift                    # User model
│   │   ├── Message.swift                 # Message (with .anchored status)
│   │   ├── Conversation.swift            # Conversation model
│   │   ├── DID.swift                     # Decentralized identity
│   │   ├── Credential.swift              # Verifiable credential
│   │   ├── Token.swift                   # ECHO token (native Constellation L0)
│   │   ├── GroupKey.swift                # Group encryption key
│   │   ├── Proposal.swift                # Governance proposal
│   │   └── VotingPower.swift             # Trust-tier weighted governance power
│   │
│   ├── UseCases/
│   │   ├── Auth/
│   │   │   ├── AuthenticateUseCase.swift
│   │   │   ├── RegisterUseCase.swift
│   │   │   └── PasskeyUseCase.swift
│   │   │
│   │   ├── Messaging/
│   │   │   ├── SendMessageUseCase.swift        # Encrypt → sign → relay
│   │   │   ├── ReceiveMessageUseCase.swift     # Decrypt → verify
│   │   │   ├── EncryptMessageUseCase.swift
│   │   │   └── VerifyAnchoringUseCase.swift    # Merkle proof (Phase 3)
│   │   │
│   │   ├── Groups/
│   │   │   ├── CreateGroupUseCase.swift
│   │   │   ├── ManageGroupKeyUseCase.swift     # Key rotation
│   │   │   └── GroupFanOutUseCase.swift
│   │   │
│   │   ├── Contacts/
│   │   │   ├── ContactDiscoveryUseCase.swift    # Argon2id hash + server match
│   │   │   ├── QRContactExchangeUseCase.swift   # Generate/scan QR with DID
│   │   │   ├── InviteLinkUseCase.swift          # Generate referral links
│   │   │   └── UsernameSearchUseCase.swift      # Public handle search
│   │   │
│   │   ├── Identity/
│   │   │   ├── CreateDIDUseCase.swift
│   │   │   ├── VerifyIdentityUseCase.swift
│   │   │   ├── ManageCredentialsUseCase.swift
│   │   │   └── ZKProofUseCase.swift            # Midnight ZK proofs (Phase 3+)
│   │   │
│   │   ├── Tokens/
│   │   │   ├── GetBalanceUseCase.swift
│   │   │   ├── SendTokensUseCase.swift
│   │   │   ├── StakeTokensUseCase.swift
│   │   │   ├── ClaimRewardsUseCase.swift
│   │   │   ├── SwapTokensUseCase.swift         # PacaSwap DEX (Phase 3+)
│   │   │   └── BridgeTokensUseCase.swift       # Cross-chain bridge (Phase 3+)
│   │   │
│   │   └── Governance/
│   │       ├── GetProposalsUseCase.swift       # Fetch active proposals
│   │       ├── CalculateVotingPowerUseCase.swift  # Trust-tier weighted
│   │       └── VoteOnProposalUseCase.swift     # Submit governance vote
│   │
│   └── Repositories/
│       ├── AuthRepository.swift
│       ├── MessageRepository.swift       # Uses MessageRelayManager
│       ├── UserRepository.swift
│       ├── TokenRepository.swift
│       ├── GroupRepository.swift
│       └── GovernanceRepository.swift    # Proposal management
│
├── Presentation/
│   ├── Coordinators/
│   │   ├── AppCoordinator.swift
│   │   ├── AuthCoordinator.swift
│   │   ├── MainCoordinator.swift
│   │   └── SettingsCoordinator.swift
│   │
│   ├── Features/
│   │   ├── Auth/                   # Login, passkey setup
│   │   ├── Onboarding/             # Username + passkey (zero PII)
│   │   ├── Conversations/          # Conversation list
│   │   ├── Chat/                   # Chat view, message bubbles
│   │   ├── Profile/                # Profile, trust score
│   │   ├── Wallet/                 # Balance, stake, delegate, swap, bridge
│   │   ├── Contacts/              # Contact discovery, QR exchange, invites
│   │   ├── PhoneVerification/     # Optional Tier 2 upgrade (not onboarding)
│   │   ├── Groups/                 # Group management
│   │   ├── Governance/             # Proposals, voting (Phase 4+)
│   │   └── Settings/               # Privacy, security
│   │
│   └── Components/
│       ├── Buttons/
│       ├── Inputs/
│       ├── Cards/
│       └── Indicators/
│           ├── TrustBadge.swift
│           ├── VerificationBadge.swift
│           └── AnchorStatusIndicator.swift  # Shows anchored status
│
└── Resources/
    ├── Assets.xcassets
    ├── Localizable.strings
    └── Info.plist
```

## Core Components

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
            
            // 4. Track commitment for on-chain anchoring
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
        encryptedPayload: EncryptedPayload,
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
        guard isValid else { throw MessageError.invalidSignature }
        
        // 2. Decrypt with own private key (Kinnami)
        let privateKey = try await getMessagingPrivateKey()
        let plaintext = try encryption.decrypt(
            payload: encryptedPayload,
            privateKey: privateKey
        )
        
        return plaintext
    }
    
    /// Called on WebSocket reconnect — drain queued outbound messages
    func drainOfflineQueue() async {
        let queuedMessages = offlineQueue.dequeueAll()
        for request in queuedMessages {
            do {
                _ = try await webSocket.sendMessage(request)
            } catch {
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
    
    @Published private(set) var pendingAnchors: [String: PendingAnchor] = [:]
    
    struct PendingAnchor {
        let messageId: String
        let commitment: Data
        let submittedAt: Date
    }
    
    func track(messageId: String, commitment: Data) {
        pendingAnchors[messageId] = PendingAnchor(
            messageId: messageId,
            commitment: commitment,
            submittedAt: Date()
        )
    }
    
    /// Called when WebSocket receives confirmation from relay
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
    
    /// Encrypt a group message with the current group key
    func encryptForGroup(plaintext: Data, groupId: String) throws -> Data {
        guard let keyInfo = getLatestKey(groupId: groupId) else {
            throw GroupError.noGroupKey
        }
        return try encryption.encryptForStorage(plaintext: plaintext, key: keyInfo.key)
    }
    
    private func storeGroupKey(_ info: GroupKeyInfo) { /* Keychain */ }
    private func getLatestKey(groupId: String) -> GroupKeyInfo? { return nil }
    private func getLatestKeyVersion(groupId: String) -> Int? { return nil }
}
```

### WebSocket Relay Client

Handles real-time message transport with offline queue drain and anchoring confirmations:

```swift
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

// On reconnect, relay server automatically drains offline queue:
private func handleMessage(_ wsMessage: WSMessage) async {
    switch wsMessage.type {
    case .message, .queueDrain:
        await messageRelayManager.handleIncomingMessage(wsMessage.payload)
        
    case .confirmation:
        let conf = try? JSONDecoder().decode(WSConfirmation.self, from: wsMessage.payload)
        if let conf = conf {
            await anchoringTracker.confirmAnchoring(
                messageId: conf.referenceId,
                snapshotHash: conf.snapshotHash,
                snapshotHeight: conf.snapshotHeight,
                merkleProof: conf.merkleProof
            )
        }
        
    case .groupKey:
        await groupKeyManager.handleKeyDistribution(wsMessage.payload)
        
    case .typing, .presence, .receipt, .ack:
        // Update UI accordingly
        break
    }
}
```

### Message Delivery Status

```swift
enum DeliveryStatus: String, Codable {
    case sending        // Encrypting / queued locally (offline)
    case sent           // Accepted by relay, recipient offline
    case delivered      // Delivered to recipient's device
    case read           // Recipient opened the message
    case failed         // Relay rejected or unrecoverable error
    case anchored       // Commitment in finalized metagraph snapshot
    case verified       // Digital Evidence fingerprint (Org tier)
}
```

The `anchored` status displays a chain-link icon (🔗) next to the message timestamp, indicating blockchain-verified integrity via ECHO's Merkle root pipeline (available for all users).

The `verified` status displays a Smart Checkmark badge (✓) — indicating the message was individually fingerprinted via Constellation's Digital Evidence API with a public verification URL (Organization tier senders, optional for VIP users who fingerprint media).

### Digital Evidence Bridge (v3.1)

Handles client-side Digital Evidence interactions for VIP/Organization tier users:

```swift
actor DigitalEvidenceBridge {
    private let backendAPI: BackendAPIClient
    
    /// Fingerprint media before E2E encryption (VIP+ users, optional).
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
    
    /// Get verification URL for Smart Checkmark rendering
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

## Core Features

**Device-Based Authentication**: Users authenticate using passkeys stored in the iOS Secure Enclave, which never leave the device. All signing operations require biometric authentication (Face ID / Touch ID). Keys are generated inside the Secure Enclave using `SecKeyCreateRandomKey` with `kSecAttrTokenIDSecureEnclave` — the private key material is never accessible to any software, including ECHO.

**Key Hierarchy**: The Secure Enclave master key derives purpose-specific keys via HKDF-SHA256 with unique context strings: "echo-did-signing" (DID operations), "echo-msg-encryption" (message key agreement), "echo-storage-encryption" (local database encryption), and "echo-wallet-signing" (token transactions). Each derived key is cryptographically independent.

**Background Key Purging**: When the app transitions to background, all derived key material is cleared from application memory. Returning to the app requires biometric authentication to re-derive the storage decryption key. This ensures a memory dump of a backgrounded app reveals no usable key material.

**Biometric Lockout Policy**: 5 consecutive biometric failures → device passcode fallback. 10 total failures → 15-minute lockout. Implemented in `BiometricAuthManager.swift`.

**Recovery**: During initial account setup, a 24-word BIP-39 mnemonic recovery phrase is generated from the Secure Enclave public parameters + user passphrase. Displayed once, never stored on any server. On a new device, entering the phrase generates a new Secure Enclave key pair and submits a DID document update to Cardano (key rotation, same DID identifier).

**Identity Verification**: The app integrates with Apple's Digital ID API (if available) or third-party verification services (Prove, Daon, 1Kosmos, Darwinium) for credential verification. Users can upload passport or license scans and complete selfie verification to establish verifiable credentials on the Cardano identity layer. Phase 3+ adds optional zero-knowledge (ZK) credential verification via Midnight blockchain, allowing users to prove trust tier eligibility or KYC compliance without revealing credential details ("Prove I'm Tier 3+ without revealing my score" or "Prove I'm 18+ without revealing birthdate").

**Message Relay Architecture**: The app sends/receives E2E encrypted blobs via WebSocket to a stateless relay server. The relay cannot read, decrypt, or modify message content. All encryption/decryption occurs on-device. The app verifies sender signatures locally—no trust in the relay is required for authenticity.

**Offline Support**: When the relay is unavailable or the device is offline, messages are queued locally in OfflineQueueManager. On reconnect, the app automatically drains its outbound queue and the relay server drains any incoming messages that arrived while offline.

**On-Chain Anchoring**: Message commitments are tracked by AnchoringTracker. When the backend confirms a commitment was included in a finalized metagraph snapshot, the message status updates to `.anchored` with a chain-link icon displayed in the UI. Phase 3 adds client-side Merkle proof verification.

**Group Messaging**: GroupKeyManager handles symmetric key lifecycle. Group admins generate and distribute keys to members via individually encrypted E2E messages. On member add/remove, keys are rotated. For large groups (100+ members), group symmetric keys avoid per-recipient re-encryption.

**Digital Evidence Integration (VIP/Org Tier)**: VIP users can optionally fingerprint media before E2E encryption. Organization tier messages automatically receive Smart Checkmark badges indicating individual-event Digital Evidence anchoring with public verification URLs.

**ECHO Wallet (Stargazer SDK)**: ECHO includes a native decentralized wallet built on the Constellation Stargazer Wallet SDK. The wallet is a primary tab in the iOS app alongside Messaging and Profile. This replaces the concept of a "rewards page" with true asset ownership. ECHO is a native Constellation Network Metagraph L1 token deployed on the public Hypergraph mainnet, ensuring full interoperability with the Constellation ecosystem.

**Why a wallet, not a rewards page:** A rewards page implies gamification points inside someone else's app. A wallet implies real assets the user owns, controls, and can use across the Constellation ecosystem. For a project whose core value proposition is "all users are owners," the wallet framing is essential.

**Wallet Features:**

* Balance display: available, staked (TokenLock), delegated, pending rewards, USD equivalent (ECHO price from PacaSwap TWAP oracle)
* Staking: lock ECHO via TokenLock, choose tier (Bronze 30d/5%, Silver 90d/8%, Gold 180d/12%, Platinum 365d/15%)
* Delegation: browse validators (uptime, commission, delegated stake, APR estimate), delegate via StakeDelegation, switch validators instantly (no cooldown)
* Rewards: claim pending rewards via AtomicAction (atomic: verify trust tier + claim rewards + update daily cap), daily cap progress bar, trust tier multiplier display
* Swap (Phase 3+): ECHO ↔ DAG and ECHO ↔ USDC via PacaSwap DEX integration (constant product AMM with 0.3% fees)
* Bridge (Phase 3+): ECHO → Base (for Aerodrome DeFi), ECHO → Ink (for Kraken exchange access)
* Founder vesting display (founders only): total allocated (CEO 100M ECHO / 10%, co-founders 20M ECHO / 2% each), vested amount, locked amount, next unlock date (monthly after 1-year cliff), cliff completion status, withdrawable balance with 14-day cooldown, "View on DAG Explorer" link for public verification
* Transaction history: all staking, delegation, reward, swap, and bridge activity
* Liquidity provider interface (Phase 3+): add/remove liquidity to ECHO/DAG and ECHO/USDC pools, view LP token balance, stake LP tokens for liquidity mining rewards

**External wallet compatibility:** Users can also view and manage ECHO in standalone Stargazer wallet or D'Cent hardware wallet. The ECHO iOS wallet and Stargazer share the same underlying Constellation keypair.

**Privacy-Preserving Contact Discovery**: Users can find contacts through four mechanisms: (1) Phone number matching — contacts' phone numbers are hashed on-device using Argon2id with a per-user salt before server transmission; the server matches hashes and returns encrypted DID references without ever seeing raw numbers. (2) QR code DID exchange — in-person contact sharing with zero server involvement. (3) Username search — optional public handles discoverable via search. (4) Invite links — referral links that track the 50 ECHO reward chain (max 3 tiers). Contact discovery is opt-in; users who decline are discoverable only via QR code, username, or direct DID share.

**Push Notifications**: Users receive APNs push notifications for offline messages (no content exposed—only conversation ID wake-up signals), transaction confirmations, and reward updates.

**ZK Proof Generation (Phase 3+)**: The `ZKProofUseCase.swift` generates zero-knowledge proofs on-device for privacy-preserving verification via Midnight. Supported proof types: trust tier threshold ("Prove I'm Tier 3+ without revealing my score"), age verification ("Prove I'm 18+ without revealing my birthdate"), credential validity ("Prove my credential is valid without revealing its content"), and balance threshold ("Prove I hold enough ECHO for staking without revealing my exact balance"). Private inputs never leave the device during proof generation. Target proof generation time: under 5 seconds on modern iPhone hardware.

**Analytics**: The app collects anonymized usage analytics with explicit user consent (no PII).

### Wallet Components (Stargazer SDK)

The ECHO app adds a "Wallet" tab alongside Messaging and Profile:

```
┌──────────────────────────────────────────────────┐
│  Tab Bar:  💬 Messages  |  👛 Wallet  |  👤 Me (+ Governance in Phase 4)  │
└──────────────────────────────────────────────────┘
```

**WalletTab SwiftUI View:**

```swift
import SwiftUI
import StargazerSDK  // Constellation Stargazer Wallet SDK

struct WalletTab: View {
    @StateObject private var viewModel = WalletViewModel()
    
    var body: some View {
        NavigationStack {
            ScrollView {
                BalanceCard(
                    balance: viewModel.totalBalance,
                    usdValue: viewModel.usdValue
                )
                
                BalanceBreakdown(
                    available: viewModel.available,
                    staked: viewModel.staked,
                    delegatedTo: viewModel.delegatedValidator,
                    pending: viewModel.pendingRewards
                )
                
                ActionButtons(
                    onStake: { viewModel.showStaking = true },
                    onDelegate: { viewModel.showDelegation = true },
                    onSwap: { viewModel.showSwap = true },
                    onBridge: { viewModel.showBridge = true }
                )
                
                DailyRewardsSection(rewards: viewModel.dailyRewards)
                
                // Founder section — only visible if user has founder TokenLock
                if let vesting = viewModel.founderVesting {
                    FounderVestingSection(vesting: vesting)
                }
                
                RecentActivityList(activity: viewModel.recentActivity)
            }
            .navigationTitle("ECHO Wallet")
        }
    }
}
```

**WalletViewModel:**

```swift
@MainActor
class WalletViewModel: ObservableObject {
    private let stargazer: StargazerClient  // Stargazer SDK
    private let backendAPI: BackendAPIClient
    private let metagraphQuery: MetagraphQueryClient
    
    @Published var totalBalance: Decimal = 0
    @Published var available: Decimal = 0
    @Published var staked: Decimal = 0
    @Published var pendingRewards: Decimal = 0
    @Published var delegatedValidator: ValidatorInfo?
    @Published var founderVesting: VestingInfo?  // nil for non-founders
    @Published var dailyRewards: DailyRewards = .empty
    @Published var recentActivity: [WalletActivity] = []
    
    func loadWallet() async {
        // 1. Query balance from Stargazer SDK (reads metagraph state)
        let balance = try? await stargazer.getBalance(token: .echo)
        self.totalBalance = balance?.total ?? 0
        self.available = balance?.available ?? 0
        
        // 2. Query TokenLock positions (staking)
        let locks = try? await stargazer.getTokenLocks(token: .echo)
        self.staked = locks?.reduce(0) { $0 + $1.amount } ?? 0
        
        // 3. Query StakeDelegation positions
        let delegations = try? await stargazer.getDelegations(token: .echo)
        self.delegatedValidator = delegations?.first?.validator
        
        // 4. Query pending rewards from backend cache
        let rewards = try? await backendAPI.getPendingRewards()
        self.pendingRewards = rewards?.total ?? 0
        self.dailyRewards = rewards?.daily ?? .empty
        
        // 5. Check for founder vesting TokenLock (special type with cliff/vest metadata)
        if let founderLock = locks?.first(where: { $0.isFounderVesting }) {
            let vestingProgress = Float(founderLock.vestedAmount) / Float(founderLock.originalAmount)
            self.founderVesting = VestingInfo(
                totalAllocated: founderLock.originalAmount,
                vested: founderLock.vestedAmount,
                locked: founderLock.lockedAmount,
                nextUnlockAmount: founderLock.nextUnlockAmount,
                nextUnlockDate: founderLock.nextUnlockDate,
                cliffCompleted: founderLock.cliffCompleted,
                cliffDate: founderLock.cliffDate,
                withdrawable: founderLock.withdrawableAmount,
                vestingProgress: vestingProgress
            )
        }
        
        // 6. Query USD value from PacaSwap TWAP oracle (Phase 3+)
        if let echoPrice = try? await getPacaSwapPrice() {
            self.usdValue = totalBalance * echoPrice
        }
    }
    
    // Claim rewards via AtomicAction (verify tier + claim + update cap)
    // AtomicAction ensures all steps succeed or all fail (no partial claims)
    func claimRewards() async throws {
        try await stargazer.submitAtomicAction([
            .verifyTrustTier(did: currentDID),
            .claimRewards(did: currentDID, types: dailyRewards.claimableTypes),
            .recordNetworkActivity(did: currentDID)
        ])
        await loadWallet()
    }
    
    // Stake ECHO via TokenLock (native Tessellation v3 primitive)
    func stakeEcho(amount: Decimal, tier: StakingTier) async throws {
        try await stargazer.submitTokenLock(TokenLockRequest(
            token: .echo,
            amount: amount,
            tier: tier.rawValue,
            duration: tier.durationDays
        ))
        await loadWallet()
    }
    
    // Delegate staked ECHO via StakeDelegation (native v3 primitive)
    // Instant validator switching - no cooldown to change delegation
    func delegateToValidator(_ validatorId: String, stakeId: String) async throws {
        try await stargazer.submitStakeDelegation(StakeDelegationRequest(
            stakeId: stakeId,
            validatorId: validatorId
        ))
        await loadWallet()
    }
    
    // Withdraw vested founder tokens via WithdrawLock (14-day cooldown)
    func withdrawVestedTokens(amount: Decimal) async throws {
        guard let vesting = founderVesting, amount <= vesting.withdrawable else {
            throw WalletError.insufficientVestedBalance
        }
        try await stargazer.submitWithdrawLock(WithdrawLockRequest(
            amount: amount
            // 14-day cooldown enforced by Currency L1 validation
        ))
        await loadWallet()
    }
    
    // Query ECHO price from PacaSwap TWAP oracle (Phase 3+)
    private func getPacaSwapPrice() async throws -> Decimal {
        let twapPrice = try await metagraphQuery.getTWAPPrice(
            tokenA: "ECHO",
            tokenB: "DAG",
            windowSeconds: 600  // 10-minute TWAP
        )
        let dagUsdPrice = try await getDagUsdPrice()
        return twapPrice * dagUsdPrice
    }
    
    // Swap ECHO tokens via PacaSwap DEX (Phase 3+)
    func swapTokens(
        inputToken: String,
        outputToken: String,
        inputAmount: Decimal,
        minOutputAmount: Decimal
    ) async throws {
        try await stargazer.submitAtomicAction([
            .swapExactInput(
                inputToken: inputToken,
                outputToken: outputToken,
                inputAmount: inputAmount,
                minOutputAmount: minOutputAmount,
                slippageTolerance: 0.05  // 5% default
            )
        ])
        await loadWallet()
    }
}

struct VestingInfo {
    let totalAllocated: Decimal       // CEO: 100M, Co-founders: 20M each
    let vested: Decimal                // Amount unlocked and available
    let locked: Decimal                // Amount still in 4-year vesting
    let nextUnlockAmount: Decimal      // 1/36th monthly after cliff
    let nextUnlockDate: Date           // Next monthly unlock date
    let cliffCompleted: Bool           // 1-year cliff status
    let cliffDate: Date                // 12 months from genesis
    let withdrawable: Decimal          // Vested but not yet withdrawn
    let vestingProgress: Float         // 0.0 to 1.0 (for progress bar)
}

struct DailyRewards {
    let messaging: Decimal             // Earned today from messaging
    let currentAutoScaledRate: Decimal  // Current per-message rate (auto-scales with network activity)
    let referrals: Decimal             // Referral bonuses (50 ECHO per verified referral)
    let staking: Decimal               // Auto-distributed staking rewards (5-15% APY by tier)
    let total: Decimal                 // Total earned today
    let claimableTypes: [String]       // Reward types ready to claim
    let trustTierRewardMultiplier: Float // Reward scale: 1.0x (Tier 1) to 3.0x (Tier 5)
    let networkDailyBudget: Decimal    // Today's emission budget (Year 1 ≈ 219,178 ECHO/day)
    let networkDistributedToday: Decimal // Total distributed across all users today
    
    static var empty: DailyRewards {
        DailyRewards(
            messaging: 0, currentAutoScaledRate: 0.1,
            referrals: 0,
            staking: 0, total: 0,
            claimableTypes: [],
            trustTierRewardMultiplier: 1.0,
            networkDailyBudget: 219178,
            networkDistributedToday: 0
        )
    }
}
```

**Staking Flow:**

```
User taps [Stake] →
  ├── Select amount (slider + input)
  ├── Select tier:
  │   ├── Bronze (30 days, 5% APR)
  │   ├── Silver (90 days, 8% APR)
  │   ├── Gold (180 days, 12% APR)
  │   └── Platinum (365 days, 15% APR)
  ├── Review: "Lock 8,000 ECHO for 180 days at 12% APR"
  ├── Biometric confirmation (Secure Enclave signs transaction)
  └── Stargazer SDK → TokenLock transaction → Currency L1
```

**Delegation Flow:**

```
User taps [Delegate] →
  ├── Validator Browser:
  │   ├── List of active L1 validators
  │   ├── Per validator: uptime %, commission %, total delegated, APR
  │   ├── Sort by: APR, uptime, commission, total delegated
  │   └── Filter: Currency L1, Data L1, both
  ├── Select validator → "Delegate 8,000 staked ECHO to Validator #7"
  ├── Biometric confirmation
  └── Stargazer SDK → StakeDelegation transaction → Currency L1
```

**Swap Flow (Phase 3+ — PacaSw**ap** DEX**):

```
User taps [Swap] →
  ├── Select pair: ECHO/DAG or ECHO/USDC
  ├── Enter input amount
  ├── See live quote:
  │   ├── Exchange rate (from constant product AMM: x*y=k)
  │   ├── Price impact % (larger trades = higher impact)
  │   ├── Estimated output amount
  │   ├── 0.3% swap fee breakdown
  │   ├── Minimum received (with 5% slippage protection)
  │   └── Route: "ECHO → DAG (direct)" or "ECHO → DAG → USDC (2-hop)"
  ├── Biometric confirmation
  └── Stargazer SDK → AtomicAction swap → Currency L1
      ├── Success: tokens swapped atomically (both sides execute or neither)
      └── Failure: no tokens moved, error displayed
```

**Bridge Flow (Phase 3+ — Base/I**nk** Cross-Chain**):

```
User taps [Bridge] →
  ├── Select destination chain:
  │   ├── Base (for Aerodrome DeFi, treasury operations)
  │   └── Ink (for Kraken exchange access)
  ├── Enter ECHO amount to bridge
  ├── See bridge details:
  │   ├── Bridge fee (e.g., 0.1%)
  │   ├── Estimated time (~1-2 minutes for finality)
  │   ├── Destination address (user's wallet on target chain)
  │   └── Total to receive (after fees)
  ├── Biometric confirmation
  └── Bridge transaction initiated
      ├── Status: "Locking ECHO..." → "Waiting for confirmation..." → "Minting on Base/Ink..."
      └── Complete: Tokens available on destination chain
      └── View on explorer: Link to Base/Ink block explorer
```

## Messaging Features

* **Message Editing with History** - Edit messages within 24 hours with full edit history
* **Message Pinning** - Pin important messages for quick reference
* **Message Forwarding** - Forward messages to other conversations with optional attribution
* **Typing Indicators** - Real-time typing status for active conversations
* **Read Receipts** - Track message delivery and read status
* **Audio Messages** - Record and send voice notes with playback
* **Message Reactions** - React to messages with emojis and custom reactions
* **Message Replies** - Quote and reply to specific messages
* **Disappearing Messages** - Auto-delete messages after specified time
* **Hidden Folders** - Biometrically protected folders for sensitive messages
* **On-Chain Anchoring** - Chain-link icon for blockchain-verified integrity (all users)
* **Smart Checkmark** - Verified badge for Digital Evidence-anchored messages (Org tier)

## User Management Features

* **User Profiles** - Display user information, avatars, status, and verification badges
* **Contact Management** - Organize contacts, add to favorites, create custom groups
* **Contact Blocking** - Block users to prevent messaging and visibility
* **Privacy Settings** - Control who can see last seen, online status, profile picture
* **Notification Management** - Configure notifications per conversation or globally:
  - Per-conversation settings: mute, mentions-only, all notifications
  - Digest mode: real-time (default), hourly batch, daily summary
  - Do Not Disturb scheduling with automatic time-zone adjustment
  - Lock screen preview controls: show/hide message content (default: hidden for privacy)
  - Notification categories: messages, transactions, rewards, governance, system

## Security Principles

| Principle | Implementation |
| --- | --- |
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
| --- | --- |
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
    case relayUnavailable = 2004
    
    // Encryption (3xxx)
    case encryptionFailed = 3001
    case decryptionFailed = 3002
    case keyNotFound = 3003
    case invalidSignature = 3004
    case commitmentMismatch = 3005
    
    // Messages (4xxx)
    case messageSendFailed = 4001
    case messageNotFound = 4002
    case rateLimitExceeded = 4003
    case messageQueued = 4004
    
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
| --- | --- | --- |
| Unit Tests | ViewModels, UseCases, Services, RelayManager, GroupKeyManager | XCTest |
| Integration Tests | API client, WebSocket relay, offline queue drain | XCTest + MockServer |
| UI Tests | Critical flows (send message, verify identity, claim reward) | XCUITest |
| Snapshot Tests | UI components, anchoring indicator, trust badge | swift-snapshot-testing |
| Security Tests | Encryption, key management, signature verification, sealed sender | Custom + third-party audit |
| Relay Tests | Offline queue, reconnect drain, rate limit handling | XCTest + MockRelay |

### Governance Features (Phase 4+)

Trust-tier weighted governance enables ECHO token holders to vote on protocol upgrades, treasury allocation, and ecosystem decisions. Governance weight is calculated as `StakedECHO × TrustTierMultiplier` to prevent plutocracy while rewarding verified, active community members.

```swift
actor GovernanceManager {
    private let stargazer: StargazerClient
    private let backendAPI: BackendAPIClient
    
    struct Proposal {
        let id: String
        let title: String
        let description: String
        let proposalType: ProposalType
        let votingEndsAt: Date
        let votesFor: Decimal
        let votesAgainst: Decimal
        let quorumRequired: Decimal  // 20% of staked tokens
        let approvalThreshold: Float  // 50% or 67% (supermajority)
        let status: ProposalStatus
    }
    
    enum ProposalType {
        case protocolUpgrade          // 67% supermajority required
        case treasuryAllocation       // Simple majority (50%)
        case validatorAdmission       // Simple majority
        case emergencyAction          // 75% supermajority
    }
    
    enum ProposalStatus {
        case active, passed, rejected, executed
    }
    
    /// Calculate user's governance voting power
    func calculateVotingPower() async throws -> VotingPower {
        // 1. Get total staked ECHO (including founder vesting locks)
        let locks = try await stargazer.getTokenLocks(token: .echo)
        let totalStaked = locks.reduce(0) { $0 + $1.amount }
        
        // 2. Get trust tier from backend
        let trustTier = try await backendAPI.getTrustTier()
        
        // 3. Apply trust tier multiplier
        let govMultiplier = trustTier.governanceMultiplier
        let effectiveWeight = totalStaked * Decimal(govMultiplier)
        
        return VotingPower(
            stakedAmount: totalStaked,
            trustTier: trustTier.level,
            governanceMultiplier: govMultiplier,
            effectiveWeight: effectiveWeight
        )
    }
    
    /// Submit vote on active proposal
    func voteOnProposal(
        proposalId: String,
        vote: VoteChoice,
        votingPower: VotingPower
    ) async throws {
        // Verify eligibility: must be Tier 2+ and have staked ECHO
        guard votingPower.trustTier >= 2, votingPower.stakedAmount > 0 else {
            throw GovernanceError.ineligibleToVote
        }
        
        // Submit vote via AtomicAction (verify stake + record vote)
        try await stargazer.submitAtomicAction([
            .verifyStake(did: currentDID, minAmount: 0),
            .submitVote(
                proposalId: proposalId,
                vote: vote,
                weight: votingPower.effectiveWeight
            )
        ])
    }
}

struct VotingPower {
    let stakedAmount: Decimal
    let trustTier: Int  // 1-5
    let governanceMultiplier: Float  // Governance scale: 0.0 (Tier 1) to 2.0 (Tier 5)
    // Note: This is DIFFERENT from the reward multiplier (1.0x-3.0x)
    // Governance: 0.0/0.5/1.0/1.5/2.0 — Rewards: 1.0/1.2/1.5/2.0/3.0
    let effectiveWeight: Decimal  // StakedECHO × governanceMultiplier
}

enum VoteChoice: String {
    case `for`, against, abstain
}
```

**Governance UI:**

```swift
struct GovernanceTab: View {
    @StateObject private var viewModel = GovernanceViewModel()
    
    var body: some View {
        NavigationStack {
            ScrollView {
                // User's voting power card
                VotingPowerCard(power: viewModel.votingPower)
                
                // Active proposals
                ForEach(viewModel.activeProposals) { proposal in
                    ProposalCard(
                        proposal: proposal,
                        onVote: { choice in
                            await viewModel.vote(
                                proposalId: proposal.id,
                                choice: choice
                            )
                        }
                    )
                }
                
                // Executed proposals (history)
                Section("Past Decisions") {
                    ForEach(viewModel.executedProposals) { proposal in
                        ProposalHistoryRow(proposal: proposal)
                    }
                }
            }
            .navigationTitle("Governance")
        }
    }
}
```

**Trust Tier Governance Multipliers:**

| Trust Tier | Multiplier | Governance Weight Formula |
| --- | --- | --- |
| Tier 1 (Unverified) | 0.0x | 0 (cannot vote) |
| Tier 2 (Newcomer) | 0.5x | StakedECHO × 0.5 |
| Tier 3 (Member) | 1.0x | StakedECHO × 1.0 |
| Tier 4 (Verified) | 1.5x | StakedECHO × 1.5 |
| Tier 5 (Trusted) | 2.0x | StakedECHO × 2.0 |

This ensures that:

* A whale who buys 50M ECHO but never verifies (Tier 1) gets **zero** governance power
* The CEO's 100M staked ECHO at Tier 5 = 200M effective weight
* 10,000 Tier 5 community members × 10K ECHO each = 200M effective weight (community can outvote CEO)
* Economic commitment (staking) + verified participation (trust tier) both matter for governance