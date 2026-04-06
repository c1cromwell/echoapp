# Hidden Folders with Biometric Protection — Implementation Specification

## Document Info

| Field | Value |
|-------|-------|
| Feature | Hidden Folders with Biometric Protection |
| Version | 1.0 |
| Date | March 1, 2026 |
| Status | Implementation-ready |
| Phase | Phase 2 (Core Build) |
| Dependencies | Secure Enclave integration, Kinnami encryption, WebSocket relay, SwiftData |

---

## 1. Overview

Hidden Folders provide biometrically-protected secure vaults for sensitive one-on-one conversations. Conversations inside hidden folders are completely invisible in the main chat interface and require Face ID / Touch ID to access. The feature adds a second encryption layer bound to biometric authentication on top of ECHO's existing E2E encryption, creating defense-in-depth against device compromise, shoulder surfing, and coerced access.

### 1.1 Design Principles

- **Invisible by default.** No visual indicator in the main interface that hidden folders exist. An observer looking at the phone sees a normal messaging app.
- **Biometric-bound keys.** The additional encryption layer uses keys derived from Secure Enclave operations that require biometric authentication. Without the authorized user's face or fingerprint, hidden folder data is cryptographic noise.
- **Device-local only.** Hidden folder metadata and the biometric-derived key material never leave the device. No cloud sync, no relay server awareness, no on-chain footprint.
- **Plausible deniability.** A duress mechanism allows the user to present a decoy folder under coercion while the real hidden folder remains invisible.
- **Composable with existing features.** Hidden folders work alongside disappearing messages, message anchoring, reactions, and read receipts — each interaction is specified below.

### 1.2 Scope

| In Scope | Out of Scope (v1) |
|----------|-------------------|
| 1:1 conversations only | Group conversations (security complexity of shared group key + per-device biometric binding is prohibitive in v1) |
| Text, image, video, audio, file, location, contact, sticker messages | Voice/video calls from within hidden folders (calls use separate WebRTC infrastructure) |
| iOS (Face ID, Touch ID) | Android (deferred to Android launch phase) |
| Up to 10 hidden folders per user | Unlimited folders |
| Device-local storage | Cross-device sync of hidden folder state |

---

## 2. Gap Analysis (vs. Original Feature Description)

The original feature description establishes the concept well but has the following gaps that this spec resolves:

| # | Gap | Impact | Resolution |
|---|-----|--------|-----------|
| 1 | References "Noise Protocol" — ECHO uses Kinnami (X25519 + ChaCha20-Poly1305) | Encryption mismatch with all other ECHO docs | Aligned to ECHO's canonical encryption spec (Section 3) |
| 2 | No relay server interaction spec | Unclear whether relay knows about hidden folders | Specified: relay is completely unaware (Section 4) |
| 3 | No on-chain anchoring interaction | Unclear if hidden messages get Merkle root commitments | Specified: commitments are included normally; the chain sees opaque hashes regardless (Section 4.3) |
| 4 | No duress/coercion protection | User can be forced to unlock biometric | Added: Duress PIN and decoy folder mechanism (Section 5.5) |
| 5 | No lockout policy | Unlimited biometric attempts = brute force risk | Added: Progressive lockout with secure wipe option (Section 5.4) |
| 6 | No biometric reset handling | If user re-enrolls Face ID, hidden folder keys could be lost | Added: Recovery mechanism via Secure Enclave key persistence (Section 5.6) |
| 7 | No screenshot/screen recording protection | "Ultra-secure" claim undermined if screenshots work | Added: UIScreen capture detection + content redaction (Section 7.3) |
| 8 | No search specification | Could hidden folder content leak via search? | Added: Hidden folder content excluded from global search; folder-internal search only when unlocked (Section 7.4) |
| 9 | No SwiftData model changes | No implementation guidance for persistence | Added: Full data model (Section 8) |
| 10 | No API/relay surface changes | Unclear what (if anything) changes server-side | Added: Nothing changes server-side (Section 4.1) — this is entirely client-side |
| 11 | No notification security detail | Notification content could leak hidden conversations | Added: Notification suppression and redaction rules (Section 7.1) |
| 12 | No disappearing messages interaction | Two privacy features that must compose correctly | Added: Composition rules (Section 6.2) |
| 13 | No iCloud/backup exclusion | Hidden folder data could leak through device backups | Added: Backup exclusion via NSFileProtection + excludedFromBackup (Section 8.3) |
| 14 | No migration spec | Moving conversations in/out of hidden folders | Added: Migration flow with re-encryption (Section 6.1) |
| 15 | No forward/export restrictions | Messages could be forwarded out of hidden folders, defeating purpose | Added: Forward and export blocking (Section 7.5) |
| 16 | No entry point UX | How does the user access hidden folders? | Added: Gesture-based access (Section 7.2) |
| 17 | No media/attachment storage spec | Images, videos, files in hidden folders need Layer 2 encryption too — stored separately from main media cache | Added: Hidden media store with encrypted file wrappers (Section 8.4) |
| 18 | No Handoff/Continuity/Widget exclusion | Hidden conversation state could leak via iOS Handoff, Continuity, or home screen widgets | Added: System integration exclusions (Section 7.6) |
| 19 | "Different access requirements" per folder not specified | Original doc mentions per-folder security levels but provides no mechanism | Added: Per-folder security tiers (Section 5.7) |

---

## 3. Encryption Architecture

### 3.1 Key Hierarchy

Hidden folders add a second encryption layer on top of ECHO's standard E2E encryption. The two layers are independent — compromising one does not compromise the other.

```
Layer 1 (Standard ECHO E2E — already exists):
  Sender device → X25519 key agreement → ChaCha20-Poly1305 → Encrypted blob
  → Relay server (sees only ciphertext) → Recipient device → Decrypt

Layer 2 (Hidden Folder — new, device-local):
  Decrypted plaintext from Layer 1
  → AES-256-GCM with biometric-derived key
  → Encrypted at rest in hidden folder storage
  → Requires biometric to derive decryption key
```

**Important:** Layer 2 encryption is a **storage encryption** layer, not a transport layer. Messages are transported using the standard ECHO relay pipeline (Layer 1). Layer 2 encrypts the *decrypted local copy* before writing it to the hidden folder's SwiftData store. This means the relay server, the metagraph, and the standard message flow are completely unaware that hidden folders exist.

### 3.2 Biometric-Derived Key Generation

```
┌─────────────────────────────────────────────────────┐
│ Secure Enclave                                       │
│                                                      │
│  Master Identity Key (P-256, non-extractable)        │
│         │                                            │
│         │ Sign(folder_id || "hidden_folder_key" ||   │
│         │       creation_timestamp)                  │
│         │ [requires biometric authentication]         │
│         ▼                                            │
│  Biometric-Gated Signature                           │
│         │                                            │
│         │ HKDF-SHA256(signature,                     │
│         │   salt = folder_id,                        │
│         │   info = "echo-hidden-folder-v1")          │
│         ▼                                            │
│  Folder Encryption Key (AES-256, symmetric)          │
│         │                                            │
│         │ Stored in Keychain with                    │
│         │ kSecAccessControlBiometryCurrentSet        │
│         ▼                                            │
│  Used for AES-256-GCM encryption of all              │
│  hidden folder content on this device                │
└─────────────────────────────────────────────────────┘
```

**Key derivation steps:**

1. The Secure Enclave holds the user's master identity key (P-256), which was created during ECHO registration and requires biometric authentication for any signing operation.
2. When creating a hidden folder, the app requests a signature over `folder_id || "hidden_folder_key" || creation_timestamp` from the Secure Enclave. This triggers Face ID / Touch ID.
3. The signature output (which is deterministic for the same input and key) is fed through HKDF-SHA256 to derive a 256-bit symmetric key.
4. This symmetric key is stored in the iOS Keychain with `kSecAccessControlBiometryCurrentSet` access control, meaning it can only be read when biometric authentication succeeds.
5. Every subsequent access to the hidden folder retrieves this key from the Keychain, which triggers biometric authentication automatically.

**Why this is secure:** The Secure Enclave key is non-extractable. The derived symmetric key in the Keychain requires biometric to read. Even if an attacker extracts the full filesystem, they get AES-256-GCM ciphertext with no key. Even if they extract the Keychain (requires jailbreak + kernel exploit), the entry requires biometric and is protected by the Secure Enclave's hardware boundary.

### 3.3 Encryption Spec (Aligned to ECHO Canonical Table)

| Purpose | Algorithm | Key Type | Library |
|---------|-----------|----------|---------|
| Hidden folder storage encryption | AES-256-GCM | Biometric-derived symmetric (256-bit) via HKDF | CryptoKit |
| Key derivation for folder key | HKDF-SHA256 | From Secure Enclave signature | CryptoKit |
| Biometric gating | Secure Enclave P-256 sign | Hardware-bound, non-extractable | Security.framework |
| Keychain protection | kSecAccessControlBiometryCurrentSet | System-managed | Security.framework |

**Note:** The original feature description references the "Noise Protocol." ECHO does not use Noise Protocol — it uses Kinnami (X25519 + ChaCha20-Poly1305) for E2E message encryption. This spec aligns hidden folder encryption with ECHO's canonical encryption table from the Data Layer Architecture.

---

## 4. Relay and On-Chain Interaction

### 4.1 Relay Server: Completely Unaware

The relay server has **zero knowledge** of hidden folders. From the relay's perspective, messages to/from hidden folder contacts are identical to any other E2E encrypted message. The relay sees the same `EncryptedMessage` payload structure, the same sender/recipient DIDs, the same WebSocket message types.

Hidden folders are a purely client-side organizational and storage-encryption feature. No API changes, no new WebSocket message types, no new relay behavior.

**What this means architecturally:**
- No changes to `openapi.yaml`
- No changes to Go backend services
- No changes to `MessageRelayManager` (messages are relayed normally)
- The only changes are in the iOS client's local storage, UI, and key management layers

### 4.2 Message Flow (Hidden Folder Conversation)

```
SENDING from hidden folder:
1. User unlocks hidden folder (biometric)
2. User composes message
3. Standard ECHO encryption: X25519 + ChaCha20-Poly1305 (Layer 1)
4. Message sent via WebSocket relay (identical to any message)
5. Relay delivers to recipient (recipient may or may not have this conversation hidden)
6. Local copy of sent message: decrypt Layer 1 → re-encrypt with folder key (Layer 2) → store

RECEIVING into hidden folder:
1. WebSocket receives encrypted message (normal flow)
2. MessageRelayManager decrypts Layer 1 (standard Kinnami decryption)
3. App checks: is this conversation in a hidden folder?
   - YES → encrypt plaintext with folder key (Layer 2) → store in hidden folder DB → suppress notification (or show redacted)
   - NO → store normally in main conversation list
```

### 4.3 On-Chain Anchoring: Unchanged

Message commitment hashes (used for Merkle root anchoring on the metagraph Data L1) are computed from the Layer 1 encrypted payload, not from the decrypted plaintext. Since hidden folders only add Layer 2 encryption *after* Layer 1 decryption and *only for local storage*, the anchoring pipeline is completely unaffected.

- Commitment = H(H(plaintext) || nonce) — computed during Layer 1 encryption on the sender's device
- The same commitment is used whether the conversation is hidden or not
- The metagraph sees the same opaque hash regardless

### 4.4 Recipient Independence

Sender and recipient hidden folder states are independent. If Alice hides her conversation with Bob, Bob's view of the conversation is unaffected. Bob may have the same conversation in his main chat list, in a different hidden folder, or not be using ECHO at all. The hidden folder is a local organizational decision, not a protocol-level concept.

---

## 5. Security Features

### 5.1 Access Control

| Access Attempt | Behavior |
|---------------|----------|
| Face ID / Touch ID success | Folder unlocks; folder key retrieved from Keychain; content decrypted and displayed |
| Face ID / Touch ID failure | Folder remains locked; see lockout policy (5.4) |
| Device passcode only (no biometric) | Folder remains locked. Passcode fallback is explicitly disabled for hidden folders. |
| Device unlocked but app backgrounded | Folder auto-locks after configurable timeout (default: 30 seconds) |
| App terminated | Folder locked; derived keys cleared from memory |
| Device restarted | Folder locked; Keychain entry requires biometric after restart |

### 5.2 Auto-Lock Behavior

Hidden folders auto-lock when:
- App enters background (after configurable delay: 0s, 30s, 1min, 5min)
- Screen locks
- App is terminated or crashes
- Inactivity timer expires (configurable: 1min, 5min, 15min, never-while-open)
- User manually locks (lock button in hidden folder toolbar)

On lock: all decrypted content is cleared from memory. The SwiftData managed object context for the hidden folder is reset. The folder key reference is released. The UI transitions to the main chat interface with no visual indication that a hidden folder was open.

### 5.3 Memory Protection

```swift
// When hidden folder is locked, ensure no decrypted content remains in memory
func lockHiddenFolder(_ folderId: String) {
    // 1. Clear all decrypted messages from the view model
    hiddenFolderViewModel.clearDecryptedContent()
    
    // 2. Reset the SwiftData context for hidden folder store
    hiddenFolderModelContainer.mainContext.reset()
    
    // 3. Release the folder key reference
    folderKeyCache.removeValue(forKey: folderId)
    
    // 4. Overwrite any temporary plaintext buffers
    SecureMemory.zeroize(&temporaryPlaintextBuffer)
    
    // 5. Transition UI to main interface
    navigationState = .mainConversationList
}
```

### 5.4 Lockout Policy

| Failed Attempts | Action |
|----------------|--------|
| 1–3 | Standard retry prompt |
| 4–5 | 30-second cooldown between attempts |
| 6–8 | 5-minute cooldown; warning that further failures will trigger protective measures |
| 9 | Final warning: "One more failure will permanently delete this hidden folder's contents" |
| 10 | **Secure wipe**: Folder key deleted from Keychain; folder content becomes irrecoverable ciphertext. The encrypted data files remain on disk (to avoid forensic detection of deletion) but can never be decrypted. |

**User-configurable:** The wipe threshold (default 10) can be adjusted from 5–25. Users can also disable the wipe feature entirely (at their own risk — this weakens protection against brute force).

**Counter reset:** The failure counter resets to 0 after a successful biometric authentication.

### 5.5 Duress Protection (Decoy Folder)

For users facing coercion (someone forcing them to unlock their phone and hidden folders), ECHO provides a decoy mechanism:

**Setup (optional, configured in hidden folder settings):**
1. User creates a "decoy folder" containing benign conversations (or the app auto-generates plausible placeholder conversations)
2. User sets a "duress PIN" (separate from device passcode) — a 4–8 digit code

**Under coercion:**
1. When prompted for biometric to unlock hidden folders, user enters the duress PIN instead (the unlock prompt offers a "Use PIN" alternative)
2. The duress PIN unlocks the decoy folder, which looks identical to a real hidden folder
3. The real hidden folder remains invisible and locked
4. Optionally: entering the duress PIN silently triggers a "panic" action (configurable): send a silent alert to a designated emergency contact, or begin a secure wipe countdown

**Key property:** An observer cannot distinguish between a successful biometric unlock (showing real folders) and a duress PIN unlock (showing decoy folders). The UI is identical.

### 5.6 Biometric Re-enrollment Handling

If the user changes their Face ID or Touch ID enrollment (adds a new face, re-enrolls fingerprints), iOS invalidates Keychain entries protected with `kSecAccessControlBiometryCurrentSet`. This means the hidden folder key becomes inaccessible.

**Mitigation:**

1. During hidden folder creation, the app also generates a **recovery key** — a 24-word BIP-39-style mnemonic that the user is required to write down (same UX pattern as crypto wallet seed phrases).
2. The recovery key encrypts a backup of the folder encryption key using AES-256-GCM.
3. This encrypted backup is stored locally (not in the Keychain — in the app's encrypted data directory).
4. If biometric re-enrollment invalidates the Keychain entry, the app detects this on next folder access attempt and prompts: "Your biometric enrollment has changed. Enter your recovery phrase to restore access to hidden folders."
5. The recovery phrase decrypts the backup of the folder key, re-enrolls it in the Keychain with the new biometric, and the folder is accessible again.

**If recovery phrase is lost:** The folder content is irrecoverable. This is by design — there is no server-side backup, no backdoor, no "contact support" option. The security model requires that only the authorized user can access hidden folders.

### 5.7 Per-Folder Security Tiers

The original feature description mentions "different access requirements" per folder. This is implemented as three security tiers, selectable when creating each folder:

| Tier | Access Method | Auto-Lock | Screenshot Protection | Wipe on Lockout | Use Case |
|------|--------------|-----------|----------------------|-----------------|----------|
| **Standard** | Face ID / Touch ID | 5 min inactive | On | Disabled | General private conversations — casual privacy needs |
| **Elevated** | Face ID / Touch ID | 30 sec inactive | On | 15 attempts | Financial, legal, or professional discussions |
| **Maximum** | Face ID / Touch ID + confirmation prompt ("Are you sure?") | Immediate on background | On + auto-lock on screenshot attempt | 10 attempts | Highest sensitivity — whistleblowing, safety-critical |

Each folder's tier determines its default settings. Users can further customize individual settings (auto-lock timeout, wipe threshold) within the bounds of their chosen tier — customization can only make settings *stricter* than the tier default, never more permissive.

**Tier enforcement:** The tier is stored as part of the `HiddenFolder` model and cannot be downgraded after creation. A folder created at Maximum tier cannot be changed to Standard. Users can upgrade a folder's tier (Standard → Elevated → Maximum) but not downgrade. To downgrade, the user must move conversations out, delete the folder, and create a new one at the lower tier.

---

## 6. Feature Interactions

### 6.1 Moving Conversations In/Out of Hidden Folders

**Moving INTO a hidden folder:**
1. User long-presses a conversation in the main chat list
2. Selects "Move to Hidden Folder" → biometric prompt
3. On success: all existing messages in the conversation are decrypted (Layer 1), re-encrypted with the folder key (Layer 2), and moved to the hidden folder's separate SwiftData store
4. The conversation is removed from the main chat list
5. Future incoming messages for this conversation are automatically routed to the hidden folder

**Moving OUT of a hidden folder:**
1. User opens hidden folder (biometric) → long-presses conversation → "Move to Main Chats"
2. Messages are decrypted (Layer 2) and stored in the standard SwiftData store (encrypted at rest with the standard storage encryption key)
3. The conversation appears in the main chat list
4. Future messages route to the main chat list

**Re-encryption during migration is a background operation.** For large conversations, a progress indicator shows "Securing conversation..." with the option to continue using the app while migration completes. Messages are migrated in reverse chronological order (newest first) so the conversation is usable during migration.

### 6.2 Disappearing Messages

Disappearing messages and hidden folders compose as follows:

| Scenario | Behavior |
|----------|----------|
| Disappearing message timer active in a hidden conversation | Timer runs normally. Message deleted from hidden folder store when timer expires. Layer 2 encryption of the local copy does not affect the timer. |
| Message expires while hidden folder is locked | The app's background expiry daemon operates on encrypted metadata (message ID + expiry timestamp, stored outside the encrypted content). It deletes the Layer 2 ciphertext without needing to decrypt. |
| On-chain Merkle root after disappearing message expires | Unchanged — the Merkle root commitment persists on-chain (it's an opaque hash). The individual commitment becomes unverifiable after the message is deleted locally, same as non-hidden disappearing messages. |

### 6.3 Read Receipts

| Scenario | Behavior |
|----------|----------|
| User reads a message in a hidden folder | Read receipt sent normally via relay (the relay doesn't know the conversation is hidden) |
| User has read receipts disabled | No receipt sent (same as non-hidden conversations) |
| Hidden folder is locked and message arrives | Message is queued in encrypted form. No read receipt sent until user opens hidden folder and views the message. |

### 6.4 Reactions

Reactions on hidden folder messages work identically to normal messages. The reaction is sent via the standard relay pipeline. The local copy of the reaction is stored in the hidden folder's Layer 2 encrypted store.

### 6.5 Message Editing

Message edits in hidden folder conversations follow the standard edit flow (re-encrypt with Layer 1, relay, re-encrypt local copy with Layer 2). Edit history is stored in the hidden folder store.

### 6.6 Message Search

See Section 7.4.

---

## 7. User Experience

### 7.1 Notification Behavior

| Setting | Behavior | Default |
|---------|----------|---------|
| **Full suppression** | No notification of any kind. The message silently enters the hidden folder. The user sees it only when they open the folder. | ✅ Default |
| **Redacted notification** | A notification appears: "New message" with no sender name, no preview, no conversation identifier. Tapping it opens the app to the main chat list (not the hidden folder). | Optional |
| **Unlocked-only notification** | Notification appears only if the hidden folder is currently unlocked. If locked, notification is silently suppressed. | Optional |

**Badge count:** Hidden folder messages are excluded from the unread message badge count on the app icon by default. User can optionally include them (this reveals that unread messages exist somewhere not visible in the main list).

**Lock screen:** Hidden folder notifications never appear on the lock screen regardless of setting.

### 7.2 Entry Point and Access UX

**Accessing hidden folders (gesture-based, no visible button):**

1. From the main conversation list, user performs a specific gesture: **pull down past the search bar and hold for 2 seconds** (similar to how iOS reveals the Spotlight search, but with a deliberate hold)
2. This triggers the biometric prompt: "Authenticate to access hidden folders"
3. On success: the hidden folder browser appears (showing all hidden folders)
4. On failure: nothing happens. The app returns to the main chat list with no error message and no indication that a hidden folder exists.

**Alternative access (configurable):** User can optionally set a secret tap pattern (e.g., triple-tap on the navigation bar title) as an alternative entry gesture. This is configurable in Settings → Privacy → Hidden Folders (which itself requires biometric to access).

**First-time setup:** The first time a user moves a conversation to a hidden folder, the app walks them through: biometric enrollment verification → recovery phrase generation → recovery phrase confirmation (re-enter 3 random words) → duress PIN setup (optional) → folder created.

### 7.3 Screenshot and Screen Recording Protection

When a hidden folder is unlocked and visible:

```swift
// Prevent screenshots and screen recording
func applyHiddenFolderProtection(to window: UIWindow) {
    // 1. Add a secure text field overlay (iOS screenshot protection technique)
    let secureField = UITextField()
    secureField.isSecureTextEntry = true
    window.addSubview(secureField)
    secureField.centerYAnchor.constraint(equalTo: window.centerYAnchor).isActive = true
    secureField.centerXAnchor.constraint(equalTo: window.centerXAnchor).isActive = true
    window.layer.superlayer?.addSublayer(secureField.layer)
    secureField.layer.sublayers?.first?.addSublayer(window.layer)
    
    // 2. Detect screen capture and recording
    NotificationCenter.default.addObserver(
        forName: UIScreen.capturedDidChangeNotification,
        object: nil,
        queue: .main
    ) { [weak self] _ in
        if UIScreen.main.isCaptured {
            // Screen is being recorded — redact content
            self?.showScreenRecordingBlocker()
        }
    }
    
    // 3. Detect screenshot
    NotificationCenter.default.addObserver(
        forName: UIApplication.userDidTakeScreenshotNotification,
        object: nil,
        queue: .main
    ) { [weak self] _ in
        // Log screenshot attempt (local only, no server notification)
        self?.logSecurityEvent(.screenshotAttempt)
        // Optionally: immediately lock the hidden folder
        self?.lockAllHiddenFolders()
    }
}
```

**Behavior:**
- Screenshots of hidden folder content produce a blank/black image (using the secure text field overlay technique)
- Screen recording: when detected, the hidden folder content is replaced with a "Content Protected" overlay
- AirPlay mirroring: content is redacted while hidden folder is visible
- Screenshot attempt optionally triggers auto-lock (user-configurable)

### 7.4 Search

| Context | Behavior |
|---------|----------|
| Global search (main chat list search bar) | Hidden folder messages are **never** included in results, regardless of whether the folder is locked or unlocked |
| Hidden folder internal search | Available only when the folder is unlocked. Searches decrypted content within that folder only. Search index is encrypted with the folder key and stored alongside folder data. |
| Spotlight / iOS system search | Hidden folder content is never indexed for Spotlight. The `CSSearchableItem` API is not used for hidden folder messages. |
| Siri suggestions | Hidden folder contacts and conversations are excluded from Siri suggestions and share sheets. |

### 7.5 Forward and Export Restrictions

| Action | Behavior |
|--------|----------|
| Forward message from hidden folder to a non-hidden conversation | **Blocked.** User sees: "Messages in hidden folders cannot be forwarded." |
| Forward message from hidden folder to another hidden folder conversation | **Allowed.** (Both are protected by biometric) |
| Copy message text | **Allowed** (user has already authenticated biometrically; clipboard is their responsibility). Optionally: auto-clear clipboard after 60 seconds. |
| Share sheet | **Blocked.** The iOS share sheet is disabled for hidden folder content. |
| Export chat history | **Blocked** for hidden folder conversations. |
| AirDrop | **Blocked.** |

### 7.6 System Integration Exclusions

Hidden folder data must be excluded from every iOS system integration that could leak conversation content or contact associations:

| iOS Feature | Exclusion Method | Leak Vector if Not Excluded |
|------------|-----------------|---------------------------|
| **Handoff / Continuity** | Do not register `NSUserActivity` for hidden folder screens | Mac or iPad could show "Continue conversation with [contact]" |
| **Siri Suggestions** | Do not donate `INInteraction` or `INSendMessageIntent` for hidden contacts | Siri could suggest "Message [hidden contact]" on lock screen |
| **Spotlight** | Do not index with `CSSearchableItem` | System search could surface hidden conversation content |
| **Home Screen Widgets** | Widget data provider returns zero items for hidden folder conversations; recent messages widget excludes hidden conversations | Widget could show "[hidden contact]: message preview" |
| **App Shortcuts** | Do not register Shortcuts for hidden folder actions | Shortcuts app could expose hidden contact names |
| **Share Sheet suggestions** | Exclude hidden contacts from `INRelevantShortcut` donations | Share sheet's suggested contacts row could reveal hidden contacts |
| **CallKit / Recent Calls** | If voice/video calls are added (Phase 3+), calls to hidden folder contacts must not appear in the system Phone app's Recents | Recents list would reveal contact association |
| **Contact Suggestions (keyboard)** | Do not donate hidden contact names via `UITextInputMode` predictions | Keyboard could suggest hidden contact name while typing |

```swift
// HiddenFolderSystemExclusions.swift

struct HiddenFolderSystemExclusions {
    
    /// Call on every app launch and after any hidden folder route change
    static func enforceExclusions(hiddenContactDIDs: Set<String>) {
        // 1. Invalidate any accidentally donated Siri interactions
        INInteraction.delete(with: hiddenContactDIDs.map { "conversation_\($0)" })
        
        // 2. Remove any accidentally indexed Spotlight items
        let searchableIndex = CSSearchableIndex.default()
        searchableIndex.deleteSearchableItems(
            withIdentifiers: hiddenContactDIDs.map { "message_\($0)" }
        )
        
        // 3. Ensure widget timeline excludes hidden conversations
        WidgetCenter.shared.reloadAllTimelines()
    }
}
```

### 7.6 Hidden Folder Management UI

The hidden folder browser (accessed via Section 7.2 gesture) shows:

- List of hidden folders with user-assigned names (e.g., "Personal", "Finance", "Legal")
- Each folder shows: folder name, conversation count, last activity timestamp
- Swipe actions: rename folder, delete folder (with confirmation + biometric)
- Settings gear: access hidden folder settings (auto-lock timeout, notification behavior, duress PIN, lockout threshold, screenshot protection toggle)
- "+" button: create new folder (up to 10)

**Deleting a hidden folder:**
1. User swipes to delete → biometric prompt
2. Confirmation: "This will permanently delete all messages in this folder. This cannot be undone."
3. On confirm: folder key is deleted from Keychain. Encrypted data is overwritten with random bytes, then deleted from disk. Conversations previously in this folder do NOT return to the main chat list — they are gone.

---

## 8. Data Model

### 8.1 SwiftData Models

```swift
import Foundation
import SwiftData

// MARK: - Hidden Folder

@Model
final class HiddenFolder: Identifiable {
    @Attribute(.unique) var id: String
    var name: String
    var createdAt: Date
    var lastAccessedAt: Date
    var autoLockTimeout: TimeInterval  // seconds (default: 30)
    var notificationMode: NotificationMode
    var screenshotProtection: Bool
    var wipeLockoutThreshold: Int  // default: 10
    var isDecoy: Bool  // true = duress decoy folder
    var securityTier: SecurityTier  // Standard, Elevated, Maximum
    var mediaQuotaBytes: Int64  // default: 2GB (2_147_483_648)
    var mediaUsedBytes: Int64  // current media storage usage
    
    // Encrypted metadata (AES-256-GCM with folder key)
    var encryptedConversationIds: Data  // List of conversation IDs in this folder
    
    enum NotificationMode: String, Codable {
        case suppressed      // No notifications at all (default)
        case redacted        // "New message" with no details
        case unlockedOnly    // Show only when folder is unlocked
    }
    
    enum SecurityTier: String, Codable {
        case standard    // 5 min auto-lock, no wipe
        case elevated    // 30 sec auto-lock, wipe at 15
        case maximum     // Immediate auto-lock, wipe at 10, auto-lock on screenshot
    }
    
    init(id: String = UUID().uuidString, name: String, tier: SecurityTier = .elevated) {
        self.id = id
        self.name = name
        self.createdAt = Date()
        self.lastAccessedAt = Date()
        self.notificationMode = .suppressed
        self.screenshotProtection = true
        self.isDecoy = false
        self.securityTier = tier
        self.mediaQuotaBytes = 2_147_483_648
        self.mediaUsedBytes = 0
        self.encryptedConversationIds = Data()
        
        // Apply tier defaults
        switch tier {
        case .standard:
            self.autoLockTimeout = 300    // 5 min
            self.wipeLockoutThreshold = 0 // Disabled
        case .elevated:
            self.autoLockTimeout = 30     // 30 sec
            self.wipeLockoutThreshold = 15
        case .maximum:
            self.autoLockTimeout = 0      // Immediate
            self.wipeLockoutThreshold = 10
        }
    }
}

// MARK: - Hidden Message (Layer 2 encrypted)

@Model
final class HiddenMessage: Identifiable {
    @Attribute(.unique) var id: String  // Same ID as the original Message
    var folderId: String
    var conversationId: String
    var timestamp: Date
    var expiresAt: Date?  // Disappearing message expiry (stored unencrypted for background deletion)
    
    // All content fields are Layer 2 encrypted (AES-256-GCM with folder key)
    var encryptedContent: Data      // Encrypted: plaintext content + content type + sender info
    var encryptedMetadata: Data     // Encrypted: reactions, read status, edit history
    
    // Unencrypted operational fields (no sensitive content)
    var contentType: String         // "text", "image", etc. (needed for UI placeholder before decrypt)
    var isRead: Bool
    var deliveryStatus: String      // DeliveryStatus raw value
    
    init(id: String, folderId: String, conversationId: String) {
        self.id = id
        self.folderId = folderId
        self.conversationId = conversationId
        self.timestamp = Date()
        self.encryptedContent = Data()
        self.encryptedMetadata = Data()
        self.contentType = "text"
        self.isRead = false
        self.deliveryStatus = "delivered"
    }
}

// MARK: - Hidden Folder Routing Rule

@Model
final class HiddenFolderRoute {
    @Attribute(.unique) var conversationId: String
    var folderId: String
    
    init(conversationId: String, folderId: String) {
        self.conversationId = conversationId
        self.folderId = folderId
    }
}
```

### 8.2 Separate SwiftData Store

Hidden folder data uses a **separate SwiftData ModelContainer** with its own SQLite database file, stored in a directory protected with `NSFileProtectionCompleteUnlessOpen`:

```swift
// HiddenFolderStorage.swift

final class HiddenFolderStorage {
    private let containerURL: URL
    private var container: ModelContainer?
    
    init() {
        // Store in app's Library/Application Support (not Documents — avoids iCloud sync)
        let appSupport = FileManager.default.urls(for: .applicationSupportDirectory, in: .userDomainMask).first!
        containerURL = appSupport.appendingPathComponent("hidden_store", isDirectory: true)
        
        // Create directory with maximum file protection
        try? FileManager.default.createDirectory(at: containerURL, withIntermediateDirectories: true)
        
        // Set file protection: only accessible when device unlocked
        try? (containerURL as NSURL).setResourceValue(
            URLFileProtection.completeUnlessOpen,
            forKey: .fileProtectionKey
        )
    }
    
    /// Open the hidden folder database (requires biometric to access folder key)
    func open() throws -> ModelContainer {
        let config = ModelConfiguration(
            url: containerURL.appendingPathComponent("hidden.sqlite")
        )
        let container = try ModelContainer(
            for: HiddenFolder.self, HiddenMessage.self, HiddenFolderRoute.self,
            configurations: config
        )
        self.container = container
        return container
    }
    
    /// Close and clear all references
    func close() {
        container?.mainContext.reset()
        container = nil
    }
}
```

### 8.3 Backup Exclusion

```swift
// Exclude hidden folder data from ALL backup mechanisms
func excludeFromBackups() throws {
    var url = containerURL
    var resourceValues = URLResourceValues()
    resourceValues.isExcludedFromBackup = true
    try url.setResourceValues(resourceValues)
}
```

Hidden folder data is excluded from:
- iCloud Backup
- iTunes/Finder local backups
- Any third-party backup tool that respects `isExcludedFromBackup`

**Consequence:** If the user gets a new device and restores from backup, hidden folders are empty. The user must use the recovery phrase to re-derive the folder key, but the message content is gone (it existed only on the original device). This is by design — the tradeoff is security over convenience.

### 8.4 Media and Attachment Storage

Media files (images, videos, audio, documents) in hidden folder conversations require Layer 2 encryption just like text messages. They are stored in a separate encrypted media directory, not in the app's standard media cache.

```swift
// HiddenMediaStore.swift

final class HiddenMediaStore {
    private let mediaDirectory: URL
    
    init() {
        let appSupport = FileManager.default.urls(
            for: .applicationSupportDirectory, in: .userDomainMask
        ).first!
        mediaDirectory = appSupport.appendingPathComponent("hidden_media", isDirectory: true)
        
        try? FileManager.default.createDirectory(at: mediaDirectory, withIntermediateDirectories: true)
        
        // File protection + backup exclusion
        try? (mediaDirectory as NSURL).setResourceValue(
            URLFileProtection.completeUnlessOpen,
            forKey: .fileProtectionKey
        )
        var url = mediaDirectory
        var rv = URLResourceValues()
        rv.isExcludedFromBackup = true
        try? url.setResourceValues(rv)
    }
    
    /// Encrypt and store a media file
    func store(data: Data, messageId: String, folderKey: SymmetricKey) throws -> URL {
        let nonce = AES.GCM.Nonce()
        let sealed = try AES.GCM.seal(data, using: folderKey, nonce: nonce)
        
        let filename = SHA256.hash(data: Data(messageId.utf8)).compactMap {
            String(format: "%02x", $0)
        }.joined()
        
        let fileURL = mediaDirectory.appendingPathComponent(filename)
        try sealed.combined!.write(to: fileURL)
        return fileURL
    }
    
    /// Decrypt and return media data
    func retrieve(messageId: String, folderKey: SymmetricKey) throws -> Data {
        let filename = SHA256.hash(data: Data(messageId.utf8)).compactMap {
            String(format: "%02x", $0)
        }.joined()
        
        let fileURL = mediaDirectory.appendingPathComponent(filename)
        let encrypted = try Data(contentsOf: fileURL)
        let box = try AES.GCM.SealedBox(combined: encrypted)
        return try AES.GCM.open(box, using: folderKey)
    }
}
```

**Key behaviors:**

| Behavior | Detail |
|----------|--------|
| Thumbnail generation | Thumbnails are generated from decrypted media, encrypted with the folder key, and stored alongside the full media. Never stored in the system's thumbnail cache. |
| Photo library isolation | Hidden folder images/videos are never saved to the iOS Photos library unless the user explicitly exports (which is blocked by default — see Section 7.5). |
| Temporary files | Any temporary decrypted files (e.g., for video playback) are written to a `tmp/` subdirectory with `NSFileProtectionComplete`, auto-deleted when the folder locks, and overwritten with zeros before deletion. |
| Storage quota | Hidden media counts toward an optional per-folder storage limit (default: 2GB per folder, configurable). Users receive a warning at 80% capacity. |
| Cache eviction | Unlike the main media cache, hidden media is never evicted by the system. iOS may reclaim app cache space, but hidden media is stored in Application Support (not Caches), so it is preserved. |

---

## 9. Routing Logic

The message routing layer sits in `ReceiveMessageUseCase` and determines whether an incoming message should be stored in the main conversation list or a hidden folder:

```swift
// MessageRouter.swift

actor MessageRouter {
    private let hiddenFolderRoutes: [String: String]  // conversationId → folderId
    private let hiddenFolderKeyManager: HiddenFolderKeyManager
    private let hiddenFolderStorage: HiddenFolderStorage
    
    /// Route a decrypted message to the correct store
    func routeMessage(_ message: DecryptedMessage) async throws {
        guard let folderId = hiddenFolderRoutes[message.conversationId] else {
            // Normal conversation — store in main SwiftData
            try await storeInMainDatabase(message)
            return
        }
        
        // Hidden folder conversation — encrypt with Layer 2 and store
        let folderKey = try await hiddenFolderKeyManager.getFolderKey(folderId: folderId)
        let encryptedContent = try encryptForHiddenStorage(message, key: folderKey)
        try await storeInHiddenDatabase(encryptedContent, folderId: folderId)
        
        // Handle notification based on folder settings
        let folder = try await getFolder(folderId)
        switch folder.notificationMode {
        case .suppressed:
            break  // No notification
        case .redacted:
            NotificationManager.showRedacted()  // "New message"
        case .unlockedOnly:
            if hiddenFolderKeyManager.isFolderUnlocked(folderId) {
                NotificationManager.showNormal(message)
            }
        }
    }
}
```

The `HiddenFolderRoute` table is loaded into memory on app launch. It maps conversation IDs to folder IDs. This table is itself stored in the hidden folder database (Section 8.2), but a minimal encrypted index (just conversation ID hashes → folder IDs) is kept in the app's standard encrypted storage so routing decisions can be made without unlocking any hidden folder.

```swift
// Routing index: stored in standard encrypted storage
// Contains only H(conversationId) → folderId mappings
// No message content, no contact names, no conversation metadata
struct RoutingIndex: Codable {
    var routes: [String: String]  // H(conversationId) → folderId
}
```

---

## 10. Testing Strategy

| Test Type | Coverage | Tools |
|-----------|----------|-------|
| **Unit Tests** | Key derivation, HKDF, AES-256-GCM encrypt/decrypt, routing logic, lockout counter, auto-lock timer, security tier default enforcement, media quota calculation | XCTest |
| **Integration Tests** | Full message flow: receive → route → Layer 2 encrypt → store → unlock → decrypt → display. Media flow: receive attachment → encrypt → store in hidden media directory → decrypt → display. Migration flow: move conversation in/out with full re-encryption. | XCTest + MockRelay |
| **Security Tests** | Keychain access control validation, biometric re-enrollment recovery, duress PIN decoy isolation, memory clearing verification, backup exclusion, system integration exclusion (verify no Siri/Spotlight/Widget leaks), clipboard auto-clear, temporary file cleanup | Custom security test suite |
| **UI Tests** | Hidden folder access gesture, folder creation flow (including tier selection), migration in/out, notification suppression, screenshot blocking, forward/export blocking, storage quota warnings | XCUITest |
| **Penetration Testing** | Filesystem analysis after lockout wipe, memory dump analysis after folder lock, backup restoration verification, forensic detection of hidden folder existence, media cache analysis for leaked thumbnails | External security audit |
| **Edge Cases** | App crash during migration (partial state recovery), biometric re-enrollment mid-session, disappearing message expiry while folder locked, 10th lockout attempt, low storage during migration, folder at storage quota receiving new media, simultaneous messages to hidden and non-hidden conversations | XCTest |
| **Performance Tests** | Migration of conversation with 10K+ messages and 500+ media files, folder unlock latency with 50+ conversations, Layer 2 encrypt/decrypt throughput for video files | XCTest (measure blocks) |

---

## 11. Risks and Mitigations

| Risk | Severity | Mitigation |
|------|----------|-----------|
| User forgets recovery phrase and re-enrolls biometric → loses all hidden folder content | High | Mandatory recovery phrase confirmation during setup; periodic reminder to verify phrase; option to reveal phrase (behind biometric) in settings |
| Lockout wipe triggered accidentally (child playing with phone) | Medium | Default threshold is 10 (high); user-configurable; clear warning at attempt 9; wipe requires 3 consecutive failures after warning |
| Forensic analysis detects hidden folder database file on filesystem | Low | File exists regardless of whether hidden folders are in use (created on app install with empty data). An observer cannot determine if any conversations are hidden. |
| iOS update changes Secure Enclave / Keychain behavior | Medium | Recovery phrase serves as fallback independent of iOS-specific APIs; monitor Apple developer releases for breaking changes |
| Duress PIN entry accidentally triggered | Low | Duress PIN is optional and disabled by default; requires deliberate setup; duress actions are configurable (alert only, wipe, etc.) |
| Hidden contact leaked via iOS system integration (Siri, Handoff, Widget, Spotlight) | Medium | Comprehensive system exclusion layer (Section 7.6) runs on every app launch and after every route change; proactive deletion of any accidentally donated data |
| Large media migration (moving conversation with many photos/videos into hidden folder) causes performance degradation | Medium | Background migration with progress UI; media migrated in batches; app remains usable during migration; per-folder storage quota prevents unbounded growth |
| Clipboard exfiltration after copying hidden folder message text | Low | Optional auto-clear clipboard after 60 seconds (user-configurable); UIPasteboard.general expiration date set to 60 seconds |
| App crash during migration leaves conversation in inconsistent state (partially in main, partially in hidden) | Medium | Migration uses transactional writes: messages are written to the destination store first, then deleted from the source in a second pass. Crash recovery on next launch detects partial migration and completes it. |
| Hidden folder storage fills device | Low | Per-folder storage quota (default 2GB); 80% warning; option to delete media while keeping text messages |

---

## 12. Implementation Priority

| Priority | Component | Effort | Phase |
|----------|-----------|--------|-------|
| P0 | Biometric-derived key generation + Keychain storage | 1 week | Phase 2 |
| P0 | Layer 2 AES-256-GCM encryption/decryption | 1 week | Phase 2 |
| P0 | Hidden folder SwiftData store (separate database) | 1 week | Phase 2 |
| P0 | Message routing (incoming messages → correct store) | 1 week | Phase 2 |
| P0 | Basic UI (access gesture, folder list, conversation view) | 2 weeks | Phase 2 |
| P0 | Hidden media store (encrypted attachment storage) | 1 week | Phase 2 |
| P1 | Per-folder security tiers (Standard / Elevated / Maximum) | 3 days | Phase 2 |
| P1 | Conversation migration (move in/out of hidden folders) | 1 week | Phase 2 |
| P1 | Notification suppression and redaction | 3 days | Phase 2 |
| P1 | Auto-lock and memory clearing | 3 days | Phase 2 |
| P1 | Screenshot and screen recording protection | 3 days | Phase 2 |
| P1 | Recovery phrase generation and backup | 1 week | Phase 2 |
| P1 | System integration exclusions (Handoff, Siri, Spotlight, Widgets) | 3 days | Phase 2 |
| P2 | Lockout policy and secure wipe | 3 days | Phase 2 |
| P2 | Duress PIN and decoy folders | 1 week | Phase 3 |
| P2 | Search exclusion and folder-internal search | 3 days | Phase 3 |
| P2 | Forward/export restrictions | 2 days | Phase 3 |

**Total estimated effort:** ~10–12 engineering weeks (1 senior iOS developer)

---

*Hidden Folders Implementation Specification v1.0*
*March 1, 2026*
*Status: Implementation-ready for Phase 2*
