# iOS Implementation Spec (v4.1)

## Aligned To

| Document | Version |
|----------|---------|
| PRD | v2.4 |
| Data Layer Architecture | v3.3 |
| iOS Frontend Architecture | v3.2 |
| ECHO Tokenomics | v1.0 |
| Hidden Folders Spec | v1.0 |
| **DESIGN.md** | **v1.0 — "The Glacial Interface"** |

---

## 1. Project Structure

```
Echo/
├── App/
│   ├── EchoApp.swift                 # @main entry, tab bar, dependency injection
│   ├── AppCoordinator.swift          # MVVM-C root coordinator
│   └── DependencyContainer.swift     # Factory-based DI
│
├── Core/
│   ├── DesignSystem/                     # "The Glacial Interface" design system
│   │   ├── GlacialTheme.swift           # Color tokens, typography, shadows, animations
│   │   ├── Components/
│   │   │   ├── EchoLogo.swift           # Concentric ripple circle logo (Canvas)
│   │   │   ├── SecureThreadIndicator.swift # 2px pulsating encrypted connection line
│   │   │   ├── GlacialNavigationBar.swift  # Frosted glass nav bar with ghost border
│   │   │   ├── SignatureGradientButton.swift # Pill CTA with deep navy → sky blue gradient
│   │   │   └── GhostBorderCard.swift     # Surface-container card with 15% opacity border
│   │   └── Modifiers/
│   │       ├── GlacialShadow.swift       # Tinted ambient shadows (never grey)
│   │       ├── GhostBorder.swift         # outline-variant at 15% opacity
│   │       └── IcyBackground.swift       # Blurred gradient orbs atmosphere
│   │
│   ├── Crypto/
│   │   ├── SecureEnclaveManager.swift    # P-256 key management (biometric-gated)
│   │   ├── KinnamiEncryption.swift       # X25519 + ChaCha20-Poly1305 (E2E)
│   │   ├── AESEncryption.swift           # AES-256-GCM (local storage + hidden folders)
│   │   ├── KeyDerivation.swift           # HKDF-SHA256
│   │   ├── CommitmentGenerator.swift     # H(H(plaintext) || nonce)
│   │   └── HashUtils.swift               # SHA-256 for Digital Evidence fingerprinting
│   │
│   ├── Network/
│   │   ├── APIClient.swift               # REST client (URLSession, cert pinning)
│   │   ├── WebSocketClient.swift         # WSS relay connection
│   │   ├── WebSocketMessageTypes.swift   # message, queueDrain, confirmation, typing, etc.
│   │   └── CertificatePinning.swift      # TLS 1.3 pin management
│   │
│   ├── Identity/
│   │   ├── DIDManager.swift              # Cardano DID lifecycle
│   │   ├── PasskeyManager.swift          # WebAuthn passkey registration/auth
│   │   └── CredentialCache.swift         # Cached verifiable credentials
│   │
│   ├── Persistence/
│   │   ├── MainModelContainer.swift      # SwiftData container (conversations, messages)
│   │   ├── HiddenFolderStorage.swift     # Separate encrypted SwiftData store
│   │   ├── KeychainManager.swift         # Secure key storage
│   │   └── BackupExclusion.swift         # isExcludedFromBackup enforcement
│   │
│   └── Stargazer/                        # Constellation Stargazer SDK integration
│       ├── StargazerBridge.swift         # SDK wrapper — balance, locks, delegations
│       ├── TransactionBuilder.swift      # Construct v3 transactions (TokenLock, etc.)
│       ├── StargazerKeyManager.swift     # Constellation keypair (linked to DID)
│       └── WalletTypes.swift             # Balance, TokenLockPosition, DelegationPosition
│
├── Features/
│   ├── Auth/
│   │   ├── AuthCoordinator.swift         # Login ↔ SMS ↔ Onboarding routing
│   │   ├── LoginScreen.swift             # Glacial Interface login (passkey + SMS)
│   │   ├── LoginViewModel.swift          # Passkey assertion, SMS code, challenge-response
│   │   ├── SMSVerificationScreen.swift   # OTP entry with glacial styling
│   │   ├── OnboardingFlow.swift          # DID creation + Stargazer wallet setup
│   │   ├── PasskeyAssertionDelegate.swift # ASAuthorizationController delegate
│   │   └── AuthModels.swift              # AuthChallenge, PasskeyCredential, AuthError
│   │
│   ├── Messaging/
│   │   ├── ConversationListView.swift
│   │   ├── ConversationListViewModel.swift
│   │   ├── ChatView.swift
│   │   ├── ChatViewModel.swift
│   │   ├── MessageRelayManager.swift     # Send/receive via WebSocket relay
│   │   ├── AnchoringTracker.swift        # Track .anchored / .verified status
│   │   ├── GroupKeyManager.swift         # Group symmetric key distribution
│   │   ├── OfflineQueueManager.swift     # Local outbox for offline sends
│   │   └── Models/
│   │       ├── Message.swift             # SwiftData model
│   │       ├── Conversation.swift
│   │       └── DeliveryStatus.swift      # .sending → .sent → .delivered → .read → .anchored → .verified
│   │
│   ├── Wallet/                           # ★ NEW — Stargazer SDK native wallet
│   │   ├── WalletCoordinator.swift
│   │   ├── WalletTab.swift               # Main wallet view
│   │   ├── WalletViewModel.swift         # Balance, staking, delegation, rewards, vesting
│   │   ├── Components/
│   │   │   ├── BalanceCard.swift         # Total balance + USD equivalent
│   │   │   ├── BalanceBreakdown.swift    # Available / staked / delegated / pending
│   │   │   ├── DailyRewardsSection.swift # Reward progress bars with daily caps
│   │   │   ├── FounderVestingSection.swift # Founder-only vesting display
│   │   │   ├── RecentActivityList.swift  # Transaction history
│   │   │   └── WalletActionButton.swift  # Stake / Delegate / Swap / Bridge buttons
│   │   ├── Staking/
│   │   │   ├── StakingView.swift         # Amount picker + tier selector
│   │   │   ├── StakingViewModel.swift    # TokenLock transaction construction
│   │   │   └── StakingTier.swift         # Bronze/Silver/Gold/Platinum enum
│   │   ├── Delegation/
│   │   │   ├── ValidatorBrowserView.swift # List validators with metrics
│   │   │   ├── ValidatorBrowserViewModel.swift
│   │   │   ├── ValidatorDetailView.swift
│   │   │   └── DelegationViewModel.swift # StakeDelegation transaction
│   │   ├── Rewards/
│   │   │   ├── ClaimRewardsView.swift    # Review + claim via AtomicAction
│   │   │   └── ClaimRewardsViewModel.swift
│   │   ├── Swap/                         # Phase 3+ — PacaSwap integration
│   │   │   ├── SwapView.swift
│   │   │   └── SwapViewModel.swift
│   │   └── Bridge/                       # Phase 3+ — Base/Ink bridge
│   │       ├── BridgeView.swift
│   │       └── BridgeViewModel.swift
│   │
│   ├── HiddenFolders/                    # Biometric-protected secure vaults
│   │   ├── HiddenFolderManager.swift
│   │   ├── HiddenFolderKeyManager.swift
│   │   ├── HiddenMediaStore.swift
│   │   ├── MessageRouter.swift           # Route incoming messages to correct store
│   │   ├── DuressManager.swift           # Duress PIN + decoy folders
│   │   └── Models/
│   │       ├── HiddenFolder.swift
│   │       ├── HiddenMessage.swift
│   │       └── HiddenFolderRoute.swift
│   │
│   ├── Evidence/                         # Digital Evidence integration
│   │   ├── DigitalEvidenceBridge.swift   # SHA-256 fingerprint + API submission
│   │   ├── SmartCheckmarkView.swift      # ✓ badge on verified messages
│   │   └── EvidenceModels.swift
│   │
│   ├── Trust/
│   │   ├── TrustTierView.swift
│   │   ├── VerificationBadgeView.swift
│   │   └── TrustCirclesView.swift
│   │
│   └── Profile/
│       ├── ProfileView.swift
│       ├── ProfileViewModel.swift
│       ├── SettingsView.swift
│       └── AccountDeletionFlow.swift
│
├── Services/
│   ├── BackendAPIClient.swift            # REST API wrapper (all endpoints)
│   ├── NotificationService.swift         # APNs registration + handling
│   ├── SystemExclusions.swift            # Siri, Spotlight, Handoff exclusions
│   └── ScreenProtection.swift            # Screenshot/recording blocking
│
├── Resources/
│   ├── Assets.xcassets
│   ├── Localizable.strings
│   └── Info.plist
│
└── Tests/
    ├── CryptoTests/
    ├── WalletTests/
    ├── MessagingTests/
    ├── HiddenFolderTests/
    └── EvidenceTests/
```

---

## 2. App Entry + Tab Bar

```swift
// App/EchoApp.swift

import SwiftUI
import StargazerSDK

@main
struct EchoApp: App {
    @StateObject private var appState = AppState()
    
    var body: some Scene {
        WindowGroup {
            if appState.isAuthenticated {
                TabView(selection: $appState.selectedTab) {
                    MessagingTab()
                        .tabItem { Label("Messages", systemImage: "bubble.left.and.bubble.right") }
                        .tag(AppTab.messages)
                    
                    WalletTab()
                        .tabItem { Label("Wallet", systemImage: "wallet.pass") }
                        .tag(AppTab.wallet)
                    
                    ProfileTab()
                        .tabItem { Label("Me", systemImage: "person.circle") }
                        .tag(AppTab.profile)
                }
            } else {
                AuthCoordinator()  // Login → SMS → Onboarding (Glacial Interface)
            }
        }
    }
}

enum AppTab: String {
    case messages, wallet, profile
}
```

---

## 3. Onboarding (DID + Wallet Setup)

```swift
// Features/Auth/OnboardingFlow.swift

struct OnboardingFlow: View {
    @StateObject private var viewModel = OnboardingViewModel()
    
    var body: some View {
        NavigationStack {
            switch viewModel.step {
            case .welcome:
                WelcomeView(onContinue: { viewModel.step = .createIdentity })
                
            case .createIdentity:
                // 1. Create Secure Enclave P-256 key (biometric-gated)
                // 2. Generate Cardano DID from public key
                // 3. Register DID on Cardano (backend pays tx fee)
                CreateIdentityView(viewModel: viewModel)
                
            case .setupWallet:
                // 4. Initialize Stargazer SDK with Constellation keypair
                // 5. Link Constellation wallet address to Cardano DID
                // 6. Show recovery phrase (BIP-39 standard)
                // 7. Confirm recovery phrase (re-enter 3 random words)
                SetupWalletView(viewModel: viewModel)
                
            case .setPasskey:
                // 8. Register WebAuthn passkey for passwordless auth
                PasskeyRegistrationView(viewModel: viewModel)
                
            case .complete:
                // 9. Profile photo + display name (optional)
                // 10. Navigate to main app
                CompleteProfileView(viewModel: viewModel)
            }
        }
    }
}

@MainActor
class OnboardingViewModel: ObservableObject {
    @Published var step: OnboardingStep = .welcome
    
    private let secureEnclave = SecureEnclaveManager()
    private let stargazer = StargazerBridge()
    private let api = BackendAPIClient()
    
    /// Create DID + wallet in a single coordinated flow.
    func createIdentityAndWallet() async throws {
        // 1. Secure Enclave key (biometric prompt)
        let publicKey = try secureEnclave.createIdentityKey()
        
        // 2. Cardano DID
        let did = try await api.registerDID(publicKey: publicKey)
        
        // 3. Constellation wallet via Stargazer SDK
        let wallet = try await stargazer.createWallet()
        
        // 4. Link wallet to DID on backend
        try await api.linkWalletToDID(did: did, walletAddress: wallet.address)
        
        // 5. Store mapping locally
        UserDefaults.standard.set(did, forKey: "echo_did")
        
        step = .setupWallet
    }
}

enum OnboardingStep {
    case welcome, createIdentity, setupWallet, setPasskey, complete
}
```

---

## 4. Stargazer SDK Bridge

```swift
// Core/Stargazer/StargazerBridge.swift

import Foundation
import StargazerSDK

/// Wraps Constellation Stargazer SDK for ECHO-specific operations.
actor StargazerBridge {
    private var client: StargazerClient?
    
    // MARK: - Initialization
    
    func createWallet() async throws -> WalletInfo {
        let sdk = try StargazerClient.initialize()
        let wallet = try await sdk.createWallet()
        self.client = sdk
        return WalletInfo(address: wallet.address, publicKey: wallet.publicKey)
    }
    
    func importWallet(mnemonic: String) async throws -> WalletInfo {
        let sdk = try StargazerClient.initialize()
        let wallet = try await sdk.importWallet(mnemonic: mnemonic)
        self.client = sdk
        return WalletInfo(address: wallet.address, publicKey: wallet.publicKey)
    }
    
    // MARK: - Balance Queries
    
    func getBalance() async throws -> BalanceInfo {
        guard let client else { throw StargazerError.notInitialized }
        let balance = try await client.getBalance(token: "ECHO")
        return BalanceInfo(total: balance.total, available: balance.available)
    }
    
    func getTokenLocks() async throws -> [TokenLockPosition] {
        guard let client else { throw StargazerError.notInitialized }
        let locks = try await client.getTokenLocks(token: "ECHO")
        return locks.map { lock in
            TokenLockPosition(
                id: lock.id,
                amount: Decimal(lock.amount) / 100_000_000, // Convert from smallest unit
                tier: lock.metadata["tier"] ?? "unknown",
                lockedUntil: Date(timeIntervalSince1970: TimeInterval(lock.expiresAt)),
                vestingType: lock.metadata["vestingType"],
                originalAmount: Decimal(lock.metadata["originalAmount"].flatMap(Int64.init) ?? lock.amount) / 100_000_000,
                cliffDate: lock.metadata["cliffDate"].flatMap { ISO8601DateFormatter().date(from: $0) },
                cliffCompleted: lock.metadata["cliffCompleted"] == "true",
                vestedAmount: Decimal(lock.metadata["vestedAmount"].flatMap(Int64.init) ?? 0) / 100_000_000,
                withdrawableAmount: Decimal(lock.metadata["withdrawable"].flatMap(Int64.init) ?? 0) / 100_000_000,
                nextUnlockDate: lock.metadata["nextUnlockDate"].flatMap { ISO8601DateFormatter().date(from: $0) },
                nextUnlockAmount: Decimal(lock.metadata["nextUnlockAmount"].flatMap(Int64.init) ?? 0) / 100_000_000,
                delegatedTo: lock.metadata["delegatedTo"]
            )
        }
    }
    
    func getDelegations() async throws -> [DelegationPosition] {
        guard let client else { throw StargazerError.notInitialized }
        return try await client.getDelegations(token: "ECHO").map { del in
            DelegationPosition(
                id: del.id,
                stakeId: del.stakeId,
                validatorId: del.validatorId,
                amount: Decimal(del.amount) / 100_000_000,
                since: Date(timeIntervalSince1970: TimeInterval(del.createdAt))
            )
        }
    }
    
    // MARK: - Transactions (v3 Primitives)
    
    /// Stake ECHO via TokenLock — requires biometric.
    func submitTokenLock(amount: Decimal, tier: StakingTier) async throws -> String {
        guard let client else { throw StargazerError.notInitialized }
        let amountSmallest = NSDecimalNumber(decimal: amount * 100_000_000).int64Value
        let tx = try await client.buildTokenLock(
            token: "ECHO",
            amount: amountSmallest,
            metadata: ["tier": tier.rawValue, "durationDays": String(tier.durationDays)]
        )
        // Biometric prompt happens automatically via Stargazer SDK signing
        return try await client.signAndSubmit(tx)
    }
    
    /// Delegate staked ECHO to validator via StakeDelegation.
    func submitStakeDelegation(stakeId: String, validatorId: String) async throws -> String {
        guard let client else { throw StargazerError.notInitialized }
        let tx = try await client.buildStakeDelegation(
            stakeId: stakeId,
            validatorId: validatorId
        )
        return try await client.signAndSubmit(tx)
    }
    
    /// Unstake via WithdrawLock (14-day cooldown enforced on-chain).
    func submitWithdrawLock(stakeId: String, amount: Decimal) async throws -> String {
        guard let client else { throw StargazerError.notInitialized }
        let amountSmallest = NSDecimalNumber(decimal: amount * 100_000_000).int64Value
        let tx = try await client.buildWithdrawLock(stakeId: stakeId, amount: amountSmallest)
        return try await client.signAndSubmit(tx)
    }
    
    /// Claim rewards via AtomicAction (verify tier + claim + update cap).
    func submitRewardClaim(rewardTypes: [String]) async throws -> String {
        guard let client else { throw StargazerError.notInitialized }
        let tx = try await client.buildAtomicAction(actions: rewardTypes.map { type in
            StargazerAction(type: "claim_reward", params: ["rewardType": type])
        })
        return try await client.signAndSubmit(tx)
    }
}

// MARK: - Types

struct WalletInfo {
    let address: String
    let publicKey: String
}

struct BalanceInfo {
    let total: Decimal
    let available: Decimal
}

struct TokenLockPosition: Identifiable {
    let id: String
    let amount: Decimal
    let tier: String
    let lockedUntil: Date
    let vestingType: String?
    let originalAmount: Decimal
    let cliffDate: Date?
    let cliffCompleted: Bool
    let vestedAmount: Decimal
    let withdrawableAmount: Decimal
    let nextUnlockDate: Date?
    let nextUnlockAmount: Decimal
    let delegatedTo: String?
    
    var isFounderVesting: Bool { vestingType == "founder" }
}

struct DelegationPosition: Identifiable {
    let id: String
    let stakeId: String
    let validatorId: String
    let amount: Decimal
    let since: Date
}

enum StargazerError: Error {
    case notInitialized
    case transactionFailed(String)
}
```

---

## 5. Delivery Status (Updated)

```swift
// Features/Messaging/Models/DeliveryStatus.swift

/// Message delivery lifecycle.
/// Each status is progressive — a message moves forward through these states.
enum DeliveryStatus: String, Codable, Comparable {
    case sending        // Encrypting / queued locally (offline)
    case sent           // Accepted by relay server, recipient offline
    case delivered       // Delivered to recipient's device
    case read           // Recipient opened the message
    case failed         // Relay rejected or unrecoverable error
    case anchored       // Commitment included in finalized metagraph snapshot (all users)
    case verified       // Digital Evidence fingerprint anchored (Org tier + Smart Checkmark)
    
    /// Icon displayed next to message timestamp.
    var icon: String {
        switch self {
        case .sending:   return "arrow.up.circle"
        case .sent:      return "checkmark"
        case .delivered:  return "checkmark.circle"
        case .read:      return "eye"
        case .failed:    return "exclamationmark.circle"
        case .anchored:  return "link"           // 🔗 chain-link icon
        case .verified:  return "checkmark.seal" // ✓ Smart Checkmark
        }
    }
    
    /// Whether tapping the icon opens a verification URL.
    var hasVerificationURL: Bool {
        self == .verified
    }
}
```

---

## 6. Digital Evidence Bridge

```swift
// Features/Evidence/DigitalEvidenceBridge.swift

import CryptoKit

actor DigitalEvidenceBridge {
    private let api: BackendAPIClient
    
    init(api: BackendAPIClient) {
        self.api = api
    }
    
    /// Fingerprint media before E2E encryption. Returns Event ID for message metadata.
    func fingerprintMedia(_ data: Data, messageId: String) async throws -> EvidenceResult {
        let hash = SHA256.hash(data: data)
        let hashHex = hash.map { String(format: "%02x", $0) }.joined()
        
        let result = try await api.submitEvidenceFingerprint(
            contentHash: hashHex,
            sourceType: "media",
            messageId: messageId
        )
        return result
    }
    
    /// Get verification URL for a verified message.
    func verificationURL(eventId: String) -> URL? {
        URL(string: "https://digitalevidence.constellationnetwork.io/verify/\(eventId)")
    }
}

struct EvidenceResult: Codable {
    let eventId: String
    let verificationUrl: String
    let timestamp: Date
}
```

---

## 6. Design System: The Glacial Interface

All ECHO iOS views follow the Glacial Interface design system (DESIGN.md). Key rules enforced across the entire app:

### Visual Rules

| Rule | Implementation | Violation |
|------|---------------|-----------|
| **No 1px solid borders** | Use tonal shifts, ghost borders (15% opacity), or negative space | Never use SwiftUI `.border()` or `Divider()` at full opacity |
| **No pure black** | Use `Color.Echo.onSurface` (#131B2E) or `Color.Echo.deepNavy` (#0F172A) | Never use `Color.black` or `#000000` |
| **No grey shadows** | Tint all shadows with `Color.Echo.onSurface` at 4-8% opacity | Never use `.shadow(color: .gray, ...)` |
| **Signature gradient** | 135° from Deep Navy (#0F172A) → Sky Blue (#0EA5E9) for primary CTAs | Never use a flat color for primary buttons |
| **Ghost borders only** | `outline-variant` at 15% opacity — a "glint on glass" | Never use `outline` at 100% for card borders |
| **Spring animations** | High damping ratio (0.85) — mimics weight of premium glass | Never use `.linear` or fast `.easeIn` animations |
| **Inter font** | Steep editorial hierarchy: Display 56px, Headline 24px, Body 16px, Label 12px | Never use system font for visible UI text |
| **Frosted glass layers** | `.ultraThinMaterial` at 60% + ghost border for floating elements | Never use opaque backgrounds for overlays |
| **32px corner radius** | Standard for cards, buttons, input containers | Never use sharp corners (< 12px) on surface elements |

### Login Screen Specification

The login screen (LoginScreen.swift) implements the Glacial Interface with these components:

**Secure Thread Indicator:** 2px sky blue line at the very top of the screen, pulsating opacity 0.6→1.0 on a 2-second cycle. Indicates active encrypted connection. Present on all authenticated screens.

**Frosted Glass Header:** ECHO logo (concentric ripple circles drawn via Canvas) + "ECHO" wordmark in `primaryContainer` blue. Background: `.ultraThinMaterial` at 60% opacity. Bottom edge: ghost border (sky light at 15% opacity). Shadow: tinted ambient (on-surface at 4%, radius 32px).

**Icy Background Atmosphere:** Two large blurred circles — `primaryContainer` at 10% opacity (top-left, blur 120px) and `secondaryContainer` at 10% (bottom-right, blur 100px). Creates the frozen lake depth effect.

**Passkey Button (Primary CTA):** Full-width pill (32px radius) with signature gradient fill. Contains: fingerprint icon in frosted circle (white 10% fill + white 20% stroke border), "Login with Passkey" in bold white, "FaceID, TouchID, or PIN" in sky-light 70%, chevron right at white 50%. Deep glacial shadow (navy 15%, radius 32px). Press state: scale 0.95 with spring animation.

**Secure Alternative Divider:** Tappable HStack with two ghost lines (outlineVariant 20%) flanking "SECURE ALTERNATIVE" label (10px, bold, tracked 1.5, outline 60%). Chevron toggles SMS section.

**SMS Section (Expandable):** `surfaceContainerLow` background, 32px radius, ghost border 15%. Phone input: `surfaceContainerLowest` fill, no visible border (ghost border 10%). "Send Code" button: `surfaceContainerHighest` fill with ghost border 20%.

**Footer:** "New to ECHO?" in `onSurfaceVariant` + "Get Started" in bold `primaryContainer`.

**Authentication Flow:**
1. User taps passkey button → `ASAuthorizationController` presents Face ID / Touch ID
2. Secure Enclave signs challenge with P-256 key
3. Backend verifies signature against registered DID public key
4. Bearer token returned → stored in Keychain → navigate to Messages tab
5. Alternative: SMS flow → enter phone → receive OTP → verify → same token flow

---

## 7. Technology Stack (Updated)

| Component | Technology | Purpose |
|-----------|------------|---------|
| UI Framework | SwiftUI | Declarative UI |
| Architecture | MVVM-C | Separation of concerns |
| Language | Swift 5.9+ | Type safety, performance |
| Concurrency | Swift Concurrency (async/await, actors) | Thread-safe async |
| Security | CryptoKit, Security.framework | Encryption, Secure Enclave |
| Wallet | Constellation Stargazer SDK | TokenLock, StakeDelegation, balance, bridge |
| Networking | URLSession, WebSocket | API + real-time relay |
| Persistence | SwiftData, Keychain | Local storage (main + hidden folder stores) |
| DI | Factory pattern | Dependency injection |
| E2E Encryption | Kinnami (X25519 + ChaCha20-Poly1305) | Message encryption |
| Identity Signing | ECDSA P-256 (Secure Enclave) | DID signing, request signing |
| Storage Encryption | AES-256-GCM (HKDF-derived key) | Local data + hidden folders |
| Hidden Folders | AES-256-GCM (biometric-derived key) | Layer 2 storage encryption |
| Transport | TLS 1.3+ with certificate pinning | Network security |
| Push | APNs | Offline message notifications |
| Evidence | Constellation Digital Evidence API | Media fingerprinting, Smart Checkmark |

---

## 8. Implementation Priority

| Priority | Component | Effort | Phase |
|----------|-----------|--------|-------|
| **P0 — Phase 2 Core** | | | |
| P0 | Onboarding (DID + Stargazer wallet creation) | 2 weeks | Phase 2 |
| P0 | E2E encrypted messaging (Kinnami + relay) | 3 weeks | Phase 2 |
| P0 | WebSocket relay client + offline queue | 2 weeks | Phase 2 |
| P0 | Wallet tab — balance display, pending rewards | 1 week | Phase 2 |
| P0 | Wallet tab — staking (TokenLock) + tier selection | 1 week | Phase 2 |
| P0 | Wallet tab — delegation (StakeDelegation) + validator browser | 1 week | Phase 2 |
| P0 | Wallet tab — reward claiming (AtomicAction) | 1 week | Phase 2 |
| P0 | Wallet tab — founder vesting display | 3 days | Phase 2 |
| P0 | Anchoring tracker (.anchored delivery status) | 3 days | Phase 2 |
| P0 | Conversation list + chat UI | 3 weeks | Phase 2 |
| **P1 — Phase 2 Polish** | | | |
| P1 | Hidden folders (biometric Layer 2 encryption) | 3 weeks | Phase 2 |
| P1 | Group messaging (key distribution + fan-out) | 2 weeks | Phase 2 |
| P1 | Trust tier display + verification badges | 1 week | Phase 2 |
| P1 | Settings, profile, account management | 1 week | Phase 2 |
| P1 | Unstaking (WithdrawLock + cooldown UI) | 3 days | Phase 2 |
| **P2 — Phase 3** | | | |
| P2 | Digital Evidence media fingerprinting (VIP+) | 1 week | Phase 3 |
| P2 | Smart Checkmark (.verified status + Explorer link) | 3 days | Phase 3 |
| P2 | Sealed sender | 2 weeks | Phase 3 |
| P2 | PacaSwap swap integration in wallet | 2 weeks | Phase 3 |
| P2 | Base bridge integration in wallet | 1 week | Phase 3 |
| P2 | Multi-device sync | 3 weeks | Phase 3 |
| **P3 — Phase 4+** | | | |
| P3 | Ink bridge integration | 1 week | Phase 4 |
| P3 | Governance voting UI | 2 weeks | Phase 4 |
| P3 | AllowSpend approval management (marketplace) | 1 week | Phase 5 |

**Total Phase 2 effort:** ~16–20 engineering weeks (1 senior iOS developer)
**Total through Phase 3:** ~24–28 engineering weeks

---

*iOS Implementation Spec v4.0*
*Aligned to: PRD v2.4, Data Layer v3.3, Tokenomics v1.0, Hidden Folders v1.0*
*Status: Implementation-ready for Phase 2*
