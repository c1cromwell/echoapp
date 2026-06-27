# Phase 5: Hidden Folders & Privacy Features

**Total Work Orders:** 46  
**Status Summary:** 14 Completed, 32 Backlog  
**Last synced with Software Factory:** 2026-05-29

---

## Backlog (43)

### WO-7: Implement Biometric Authentication System for Hidden Folders

**Blueprint:** Hidden Folders with Biometric Protection

## Summary

Implement the biometric authentication gate for hidden folder access using iOS `LocalAuthentication` framework. This `BiometricAuthManager` wrapper handles Face ID, Touch ID, and fallback PIN with progressive lockout after repeated failures. It is called before any hidden folder content is decrypted or displayed.

## In Scope

- `BiometricAuthManager` using `LAContext` with `.deviceOwnerAuthenticationWithBiometrics` policy
- Face ID and Touch ID support with system-provided authentication UI (no custom biometric capture)
- Fallback PIN: `LAPolicy.deviceOwnerAuthentication` includes PIN fallback when biometrics fail
- Progressive lockout:
  - 3 failed attempts → 1-minute lockout
  - 5 failed attempts → 5-minute lockout
  - 10 failed attempts → 30-minute lockout
- Lockout state persisted in SwiftData (survives app restarts)
- Authentication callback: `onSuccess(method: .faceID | .touchID | .pin)` and `onFailure(error: LAError)`
- Biometric re-enrollment handling: detect when biometric database changes and require key rotation (WO-18)
- Used by: `HiddenFolderManager` before decryption, folder list display, and moving conversations in/out of hidden folders

## Out of Scope

- Custom biometric template capture (uses iOS system biometrics)
- Biometric-derived encryption keys (WO-18)
- Hidden folder management UI (WO-30)

## Requirements

Derived from the Hidden Folders blueprint.

**BiometricAuthManager:**
```swift
// Core/Security/BiometricAuthManager.swift
actor BiometricAuthManager {
    private let context = LAContext()
    private var failedAttempts = 0
    private var lockedUntil: Date?

    func authenticate(reason: String) async throws -> AuthMethod {
        // Check lockout
        if let lockedUntil = lockedUntil, Date() < lockedUntil {
            throw BiometricError.lockedOut(until: lockedUntil)
        }

        // Try biometrics first, fall back to PIN if needed
        let policy: LAPolicy = .deviceOwnerAuthentication  // includes PIN fallback
        let success = try await context.evaluatePolicy(policy, localizedReason: reason)

        if success {
            failedAttempts = 0
            return context.biometryType == .faceID ? .faceID : .touchID
        }
        throw BiometricError.failed
    }

    func recordFailure() {
        failedAttempts += 1
        switch failedAttempts {
        case 3: lockedUntil = Date().addingTimeInterval(60)
        case 5: lockedUntil = Date().addingTimeInterval(300)
        case 10: lockedUntil = Date().addingTimeInterval(1800)
        default: break
        }
    }
}
```

## Blueprints

- Hidden Folders with Biometric Protection — Defines Face ID/Touch ID support, PIN fallback, failed attempt tracking, lockout policy, Secure Enclave integration, and biometric re-enrollment handling

---

### WO-18: Implement Biometric-Derived Encryption Key Management

**Blueprint:** Hidden Folders with Biometric Protection

## Summary

Implement the biometric-derived encryption key management system for hidden folders. Each hidden folder gets a unique AES-256 key derived from the user's Secure Enclave key and folder ID using HKDF. Keys are stored in the Keychain with biometric access control — they can only be retrieved after successful biometric authentication. Keys are rotated when biometric data changes.

## In Scope

- `HiddenFolderKeyManager` using `SecKeyCreateRandomKey` in Secure Enclave for master key
- Per-folder derived key: `HKDF-SHA256(masterKey, salt: folderId)` — unique AES-256-GCM key per folder
- Keychain storage with `SecAccessControl`: `.biometryCurrentSet` flag requiring biometric authentication for each Keychain access
- Key rotation on biometric database change: detect via `LAContext.evaluatedPolicyDomainState` comparison; destroy old keys and derive new ones from updated Secure Enclave key
- Secure key destruction: Keychain delete + memory zero-fill on biometric removal
- Key escrow prevention: keys stored in Keychain with `.thisDeviceOnly` attribute, no iCloud sync
- Key retrieval requires `BiometricAuthManager` authentication first (called by `HiddenFolderManager`)

## Out of Scope

- Biometric authentication UI (WO-7)
- Hidden folder management (WO-30)
- Cross-device key sync (by design: hidden folders are device-local)

## Requirements

Derived from the Hidden Folders blueprint.

**Key Derivation:**
```swift
// Core/Security/HiddenFolderKeyManager.swift
actor HiddenFolderKeyManager {
    // Master key: P-256 hardware key in Secure Enclave (never extractable)
    // Derived keys: HKDF from Secure Enclave signature, stored in Keychain

    func keyForFolder(_ folderId: String) async throws -> SymmetricKey {
        // 1. Require biometric auth (LAContext already evaluated by caller)
        // 2. Look up derived key from Keychain (requires biometric)
        let keychainKey = "hidden_folder_key_\(folderId)"
        if let existing = try? keychain.loadBiometricProtected(keychainKey) {
            return SymmetricKey(data: existing)
        }

        // 3. Derive new key from Secure Enclave signature + folderId
        let signature = try await secureEnclave.sign(
            data: Data(folderId.utf8),
            reason: "Derive encryption key for hidden folder"
        )
        let derivedKey = try HKDF<SHA256>.deriveKey(
            inputKeyMaterial: SymmetricKey(data: signature),
            info: Data("hidden_folder_\(folderId)".utf8),
            outputByteCount: 32
        )

        // 4. Store in Keychain with biometric access control
        try keychain.storeBiometricProtected(keychainKey, value: derivedKey.withUnsafeBytes { Data($0) })
        return derivedKey
    }

    func destroyKey(_ folderId: String) {
        keychain.delete("hidden_folder_key_\(folderId)")  // Also memory-zeros
    }
}
```

## Blueprints

- Hidden Folders with Biometric Protection — Defines biometric-derived key generation, PBKDF2/HKDF derivation, unique key per folder, Secure Enclave storage, key rotation, escrow prevention, and secure destruction

---

### WO-30: Create Hidden Folder Management System

**Blueprint:** Hidden Folders with Biometric Protection

## Summary

Build the hidden folder management system — create, rename, and delete folders that are completely invisible in the main chat interface. Folder metadata is stored in an encrypted, local-only SwiftData partition. Up to 20 hidden folders per user. Secure deletion cryptographically wipes all folder content and keys.

## In Scope

- `HiddenFolderManager` actor: create, rename, delete folders; list folders (requires biometric auth)
- Encrypted metadata storage: `HiddenFolder {id, name, createdAt, lastAccessedAt, notificationSettings, backupEnabled}` in SwiftData, encrypted partition separate from main message store
- Folder naming: up to 50 character user-defined names
- Folder visibility: hidden folders never appear in main conversation list, search results, or backups of main store
- Maximum 20 hidden folders per user (enforced on creation)
- Secure folder deletion: destroy biometric-derived key (`HiddenFolderKeyManager.destroyKey`), then delete all messages in folder from SwiftData, zero-fill metadata
- Folder recovery: recovery phrase (BIP39) used for backup restoration (WO-69)
- No cloud sync: `.thisDeviceOnly` on all Keychain items, no iCloud Drive backup of hidden folder partition

## Out of Scope

- Biometric authentication gate (WO-7)
- Encryption key management (WO-18)
- Conversation moving (WO-42)
- Notification management for hidden folders (WO-61)

## Requirements

Derived from the Hidden Folders blueprint.

**Hidden Folder Model:**
```swift
struct HiddenFolder: Identifiable, Codable {
    let id: UUID
    var name: String           // Max 50 chars
    let createdAt: Date
    var lastAccessedAt: Date
    var notificationSetting: HiddenFolderNotification  // .silent, .suppress, .show
    var backupEnabled: Bool
    var conversationCount: Int  // Computed from linked conversations

    // Stored in SwiftData with NSPersistentCloudKitContainer disabled (local only)
    // Encrypted partition: separate from main message store
}
```

**Secure Deletion:**
```swift
func deleteFolder(_ folderId: UUID) async throws {
    // 1. Require biometric auth
    // 2. Destroy encryption key (key manager)
    await keyManager.destroyKey(folderId.uuidString)
    // 3. Delete all messages in folder from SwiftData
    try await database.deleteAllMessages(inFolder: folderId)
    // 4. Delete folder metadata record
    try await database.deleteFolder(folderId)
    // 5. Verify deletion complete
}
```

## Blueprints

- Hidden Folders with Biometric Protection — Defines folder creation, naming, metadata encryption, local-only storage, secure deletion with cryptographic wiping, and recovery options

---

### WO-38: Build Disappearing Messages with Cryptographic Deletion Verification

**Blueprint:** Blockchain-Anchored Messaging with Provable Integrity

**Purpose**: Enable users to send messages that automatically delete after predetermined time periods while maintaining blockchain-anchored proof that conversations occurred, supporting legal and business verification needs without revealing content.

**Requirements**:
- Users must be able to set disappearing message timers from preset intervals ranging from 10 seconds to 7 days
- Premium users must have access to custom timing options beyond preset intervals
- Messages must display countdown timers showing remaining time before deletion
- System must use time-locked smart contracts to enforce deletion timing
- Messages must be cryptographically verified before deletion to ensure integrity
- Blockchain-anchored deletion records must be created as immutable proof that conversations occurred
- Messages must automatically delete from all devices (sender and recipient) when timer expires
- System must prevent screenshots or copying of disappearing messages where technically possible
- Deletion verification must be available to prove message existence without revealing content

**Out of Scope**:
- Manual message deletion outside of timer expiration
- Recovery of deleted disappearing messages
- Disappearing message forwarding or reactions

---

### WO-42: Implement Conversation Moving and Encryption System

**Blueprint:** Hidden Folders with Biometric Protection

## Summary

Implement the conversation moving system that re-encrypts an existing conversation with the target hidden folder's biometric-derived key and removes it from the main conversation list. The move is atomic — either fully completes or fully reverts. All conversation history, media, and metadata are preserved.

## In Scope

- `MoveConversationToHiddenFolder(conversationId, folderId)` — atomic operation
- Step 1: Require biometric auth and retrieve folder's biometric-derived AES-256 key
- Step 2: Re-encrypt all messages with folder key (original E2E encryption remains; add folder key layer)
- Step 3: Move conversation record from main SwiftData context to hidden folder context
- Step 4: Remove from main conversation list (no longer visible in main UI)
- Step 5: Update notification settings to folder configuration
- Atomic operation: use SwiftData `withTransaction`; rollback on any failure
- Move reversal: decrypt with folder key, restore to main context, delete from folder
- Integrity verification: checksum validation before and after move

## Out of Scope

- Folder creation (WO-30)
- Encryption key management (WO-18)
- Biometric auth (WO-7)

## Requirements

Derived from the Hidden Folders blueprint.

**Move Flow:**
```swift
// Core/HiddenFolders/ConversationMoveManager.swift
actor ConversationMoveManager {
    func moveToHiddenFolder(_ conversationId: String, folder: HiddenFolder) async throws {
        // 1. Get folder key (requires biometric, done by caller before this call)
        let folderKey = try await keyManager.keyForFolder(folder.id.uuidString)

        // 2. In a single atomic transaction:
        try await database.performAtomicTransaction {
            // Re-encrypt messages with folder key (on top of existing E2E encryption)
            let messages = try database.fetchMessages(for: conversationId)
            for message in messages {
                let reEncrypted = try AES256GCM.encrypt(message.storedBlob, key: folderKey)
                try database.updateMessageBlob(message.id, blob: reEncrypted, isHiddenEncrypted: true)
            }
            // Move conversation to hidden folder context
            try database.moveConversation(conversationId, to: folder.id)
            // Remove from main list
            try database.setConversationHidden(conversationId, true)
        }
    }
}
```

## Blueprints

- Hidden Folders with Biometric Protection — Defines conversation moving, AES-256-GCM biometric-derived re-encryption, main list removal, metadata preservation, move reversal, and atomic operation requirement

---

### WO-47: Implement Hidden Folders with Biometric Protection

**Blueprint:** Blockchain-Anchored Messaging with Provable Integrity

**Purpose**: Provide users with secure, locally-stored hidden folders for organizing sensitive messages, protected by biometric authentication to ensure only authorized access while maintaining complete privacy.

**Requirements**:
- Users must be able to create multiple hidden folders for organizing sensitive messages
- Hidden folders must be protected by biometric authentication using Face ID or Touch ID where available
- System must provide PIN code fallback authentication when biometric authentication is unavailable
- Hidden folder contents must be stored locally only, with no synchronization to metagraph or cloud services
- All hidden folder data must be encrypted using iOS Data Protection encryption at rest
- System must log access attempts to hidden folders for security auditing
- Users must be able to move messages into and out of hidden folders
- Hidden folders must be completely invisible in normal message views when locked
- Authentication must be required each time a user attempts to access hidden folders

**Out of Scope**:
- Cloud backup or sync of hidden folder contents
- Sharing hidden folder access with other users
- Hidden folder access on non-iOS platforms

---

### WO-51: Implement Enhanced Multi-Layer Encryption for Hidden Messages

**Blueprint:** Hidden Folders with Biometric Protection

## Summary

Implement the enhanced multi-layer encryption for messages stored in hidden folders. Messages in hidden folders have two encryption layers: the standard Kinnami E2E layer (from the messaging system) and an additional AES-256-GCM layer using the biometric-derived folder key. This ensures even if the Kinnami layer were compromised, hidden folder content requires biometric auth to access.

## In Scope

- Second-layer AES-256-GCM encryption applied to stored message blobs using `HiddenFolderKeyManager.keyForFolder`
- AES-GCM nonce: 12 bytes random, prepended to ciphertext
- Key ratcheting: every 100 messages within a hidden folder, derive a new subkey via HKDF from folder key + message sequence number
- Message authentication codes (MAC): AES-GCM's built-in AEAD authentication tag verifies message integrity on decrypt
- Replay attack prevention: message sequence numbers in metadata; reject out-of-order decryption attempts
- Encrypted metadata layer: wrap timestamp, sender DID, delivery status with same folder key
- Decryption gate: call `BiometricAuthManager.authenticate()` before any bulk decryption session begins
- Session unlock: once biometric authenticated, decrypt messages within session without re-prompting (session TTL: 15 minutes idle)

## Out of Scope

- Kinnami E2E layer (exists in messaging system, not modified here)
- Biometric authentication prompts (WO-7)
- Key generation/management (WO-18)

## Requirements

Derived from the Hidden Folders blueprint.

**Multi-Layer Encryption:**
```
Message at rest in hidden folder:
  [AES-256-GCM(folderKey, nonce) [Kinnami ChaCha20-Poly1305(recipientKey) [plaintext]]]
  
Layer 1 (outer): AES-256-GCM with biometric-derived folder key
Layer 2 (inner): Existing Kinnami encryption (X25519 + ChaCha20-Poly1305)

Decryption flow:
1. Biometric auth → retrieve folder key from Keychain
2. AES-256-GCM decrypt outer layer → Kinnami ciphertext
3. Kinnami decrypt → plaintext (requires recipient private key, already done by messaging layer)
```

**Key Ratcheting:**
```swift
func derivedSubkey(for messageSequence: Int) throws -> SymmetricKey {
    let epoch = messageSequence / 100  // New subkey every 100 messages
    let info = Data("hidden_ratchet_\(folderId)_\(epoch)".utf8)
    return try HKDF<SHA256>.deriveKey(inputKeyMaterial: folderKey, info: info, outputByteCount: 32)
}
```

## Blueprints

- Hidden Folders with Biometric Protection — Defines multi-layer encryption, AES-256-GCM + Noise Protocol combination, key ratcheting, forward secrecy, MAC verification, and replay attack prevention

---

### WO-56: Implement Silent Message Core Infrastructure with Notification Suppression

**Blueprint:** Silent and Scheduled Private Chats

## Summary

Build the silent message core infrastructure — notification suppression flags embedded in the encrypted message payload, per-conversation and per-message silent mode toggles, and backend enforcement that respects silent flags when routing notifications. Silent messages appear in the conversation thread only when actively opened.

## In Scope

- `MessageMetadata.isSilent` boolean flag embedded in encrypted payload (relay cannot read it)
- iOS: silent message composer toggle (per-message) and conversation-level silent mode toggle
- `UNNotificationServiceExtension`: when notification received, decrypt metadata, check `isSilent`; if true, drop push notification (not shown)
- Badge suppression: silent messages do NOT increment app badge count or conversation badge
- Typing indicator suppression: when in silent mode for conversation, typing indicators are NOT sent
- Read receipt suppression: silent mode conversations suppress read receipts
- Per-conversation silent mode duration: 1h, 8h, 24h, or "until disabled" presets
- Silent message visual indicator in chat: subtle icon (🔇) to distinguish silent messages
- Backend Notification Service: receives `{conversationId, isSilent}` flag (never the content), does not push APNs if silent

## Out of Scope

- Message scheduling (WO-65)
- Trust score rate limiting (WO-87)
- Blockchain anchoring of scheduled messages (WO-96)

## Requirements

Derived from the Silent and Scheduled Private Chats blueprint.

**Silent Flag Architecture:**
```
Silent flag is encrypted INSIDE the message payload.
Relay server CANNOT read it.
Recipient's device decrypts payload, then checks isSilent.
iOS Notification Service Extension intercepts APNs delivery,
  decrypts minimal metadata, suppresses if isSilent == true.

Note: The backend pre-screens at the APNs level:
  Message Relay knows per-conversation silent status from user settings
  (stored in Contacts Service) and skips APNs push accordingly.
```

**iOS Silent Toggle:**
```swift
// Presentation/Features/Chat/SilentMessageToggle.swift
struct SilentMessageComposerView: View {
    @Binding var isSilent: Bool
    var body: some View {
        Toggle(isOn: $isSilent) {
            Label("Silent", systemImage: "bell.slash")
        }
        .toggleStyle(.button)
    }
}
```

## Blueprints

- Silent and Scheduled Private Chats — Defines silent mode per-conversation and per-message, notification/badge/typing/receipt suppression, encrypted suppression flags, and relay node privacy

---

### WO-61: Create Hidden Folder Notification Management System

**Blueprint:** Hidden Folders with Biometric Protection

## Summary

Build the notification management system for hidden folder conversations. When a hidden folder is locked, incoming message notifications show no preview content. When unlocked, full notification behavior can be configured per folder. All notification settings are folder-specific and override global settings.

## In Scope

- `HiddenFolderNotificationManager`: configure per-folder notification behavior
- Notification modes per folder: `.silent` (appear only when folder is unlocked), `.suppress` (no notifications at all), `.showGeneric` (notify with generic "New message" text, no sender/content)
- Backend: Notification Service (port 8007) marks messages from hidden conversation IDs with a `hidden_folder` flag in APNs payload
- iOS: `UNNotificationServiceExtension` — when `hidden_folder` flag present and folder is locked, replace notification content with generic text
- When folder is unlocked (biometric authenticated): show full notification with sender/preview if `.silent` mode is active
- Badge management: hidden folder badges only shown when folder is unlocked (not in global app badge count by default)
- Custom notification sounds per folder (up to 10 custom sounds loaded from app bundle)
- Do Not Disturb time windows per folder: suppress all notifications during specified hours

## Out of Scope

- General notification system (WO-57)
- Biometric authentication (WO-7)
- Encryption of message content (WO-51)

## Requirements

Derived from the Hidden Folders blueprint.

**Notification Privacy (iOS Notification Service Extension):**
```swift
// NotificationService.swift (UNNotificationServiceExtension)
class NotificationServiceExtension: UNNotificationServiceExtension {
    override func didReceive(_ request: UNNotificationRequest, withContentHandler handler: @escaping (UNNotificationContent) -> Void) {
        let content = request.content.mutableCopy() as! UNMutableNotificationContent
        guard content.userInfo["hidden_folder"] as? Bool == true else {
            handler(content)  // Not a hidden folder message, show normally
            return
        }
        // Hidden folder message: check if folder is currently unlocked
        if HiddenFolderSessionManager.shared.isUnlocked(content.userInfo["folder_id"] as? String) {
            handler(content)  // Folder unlocked: show full notification
        } else {
            // Folder locked: show generic notification
            content.title = "New Message"
            content.body = "Tap to view"
            content.subtitle = ""
            handler(content)
        }
    }
}
```

## Blueprints

- Hidden Folders with Biometric Protection — Defines silent notifications, notification suppression, locked/unlocked folder notification behaviors, custom sounds, and badge management

---

### WO-65: Build Message Scheduling System with Time-Locked Encryption

**Blueprint:** Silent and Scheduled Private Chats

## Summary

Build the message scheduling system — compose a message now, store it encrypted locally, and deliver it via the relay at the specified future time (up to 30 days). Scheduling uses `BGTaskScheduler` for background delivery. There are no smart contracts involved — delivery is triggered by the sender's device timer (same architecture as disappearing messages: client-side control). Supports cross-timezone scheduling and recurring patterns.

## In Scope

- Scheduling UI: date/time picker with presets (1h, 4h, 8h, 24h, 1 week), custom date/time, timezone detection
- Recipient timezone display and automatic conversion using `TimeZone` and `Calendar`
- Scheduled message stored locally in SwiftData (encrypted with AES-256-GCM, Secure Enclave-derived key)
- `BGProcessingTask` registration to fire at scheduled time and transmit message
- On timer: load from SwiftData, decrypt, re-encrypt for recipient (standard Kinnami pipeline), transmit via relay
- Message editing before delivery: update stored encrypted message
- Message cancellation: delete from SwiftData and cancel `BGTask`
- Recurring patterns: daily, weekly, monthly — creates a new `BGTask` per scheduled occurrence
- Delivery confirmation: push `{type: "receipt", status: "scheduled_delivered"}` back to sender

## Out of Scope

- Silent message functionality (WO-56)
- Trust score rate limiting (WO-87)
- Blockchain anchoring of delivery proof (WO-96)
- Smart contracts for scheduling (not used — client-side only)

## Requirements

Derived from the Silent and Scheduled Private Chats blueprint.

**Scheduling Architecture (Client-Side):**
```
⚠️ Scheduled messages are NOT time-locked via smart contracts.
Per the blueprint: "messages are encrypted and stored locally on the sender's
device until the scheduled delivery time."

Delivery trigger: iOS BGProcessingTask fires at scheduled time
→ Load from SwiftData → Encrypt for recipient → Transmit via relay
→ Same relay pipeline as instant messages
```

**Scheduled Message Model:**
```swift
struct ScheduledMessage: Identifiable, Codable {
    let id: UUID
    let conversationId: String
    let recipientDID: String
    var encryptedContent: Data      // AES-256-GCM with local device key
    let scheduledFor: Date
    let timezone: TimeZone
    var isRecurring: Bool
    var recurrenceRule: RecurrenceRule?
    var bgTaskIdentifier: String?   // Registered BGProcessingTask ID
    var status: ScheduledStatus     // .pending, .delivered, .cancelled, .failed
}
```

## Blueprints

- Silent and Scheduled Private Chats — Defines message scheduling up to 30 days, local storage until delivery, cross-timezone handling, recurring scheduling, edit/cancel, and delivery confirmation

---

### WO-66: Implement Time-Sensitive Encryption Key Management System

**Blueprint:** Disappearing Messages with Cryptographic Verification

## Summary

Implement the encryption key destruction mechanism for disappearing messages. Disappearing messages use the standard Kinnami encryption (X25519 + ChaCha20-Poly1305) — there are no special "time-sensitive" keys. What makes disappearing messages ephemeral is that the encryption keys are cryptographically destroyed from the Keychain and memory when the expiration timer fires. This work order builds that key lifecycle management component.

## In Scope

- `DisappearingKeyManager` — wraps `KeychainManager` with expiry tracking; stores `(messageId, encryptionKey, expiresAt)` tuples
- Scheduled key destruction at expiration time via `BGTaskScheduler` (background) or timer (foreground)
- Key destruction: overwrite key material in Keychain with zeros, then delete the Keychain item
- In-memory key clearing: call `withUnsafeMutableBytes { bytes.initialize(repeating: 0) }` on key data before releasing
- Expiry persistence: store expiry timestamps in SwiftData (survives app restart)
- On app launch: check for any expired keys and destroy them immediately (handle offline case)
- VIP user support: accept custom expiry durations beyond the preset intervals
- Key inventory query: `isKeyAvailable(messageId:)` returns false after expiry (for proof generation logic)

## Out of Scope

- Message content encryption/decryption (standard Kinnami, WO-4/WO-28)
- Countdown timer UI (WO-75)
- Message deletion from database (WO-84)
- Smart contract key release (not used — deletion is client-side, see WO-105)

## Requirements

Derived from the Disappearing Messages blueprint.

**Key Lifecycle:**
```
T=0:      Message created, standard Kinnami keys generated (X25519 ephemeral)
          Derived symmetric key stored in Keychain with expiry tag
T=expiry: Timer fires →
          1. Overwrite key data with zeros (secure memory wipe)
          2. Delete from Keychain
          3. Mark messageId as "key destroyed" in SwiftData
T=expiry+ Application can no longer decrypt the message (key gone)
          On-chain Merkle root persists forever (proof of existence)
```

**Key Manager Implementation:**
```swift
// Core/Security/DisappearingKeyManager.swift
actor DisappearingKeyManager {
    private let keychain: KeychainManager

    func storeKey(_ key: SymmetricKey, messageId: String, expiresAt: Date) throws {
        let keyData = key.withUnsafeBytes { Data($0) }
        try keychain.store(key: "disappearing_\(messageId)", value: keyData, expiresAt: expiresAt)
        scheduleKeyDestruction(messageId: messageId, at: expiresAt)
    }

    func destroyKey(messageId: String) {
        // 1. Overwrite with zeros (secure wipe)
        if var keyData = try? keychain.load("disappearing_\(messageId)") {
            keyData.withUnsafeMutableBytes { $0.initialize(repeating: 0) }
        }
        // 2. Delete from Keychain
        keychain.delete("disappearing_\(messageId)")
        // 3. Mark as destroyed in SwiftData
        markKeyDestroyed(messageId: messageId)
    }

    func isKeyAvailable(_ messageId: String) -> Bool {
        keychain.exists("disappearing_\(messageId)") && !isDestroyed(messageId: messageId)
    }
}
```

## Blueprints

- Disappearing Messages with Cryptographic Verification — Defines the key destruction sequence, what gets deleted vs. preserved, secure deletion requirements, and client-side independence principle

---

### WO-67: Implement Silent Notifications and Scheduled Message System

**Blueprint:** Blockchain-Anchored Messaging with Provable Integrity

**Purpose**: Provide users with flexible notification control and message scheduling capabilities, allowing them to manage conversation interruptions and send messages at optimal times while maintaining reliable delivery.

**Requirements**:
- Users must be able to mute notifications for specific conversations with duration options from 1 hour to forever
- Muted conversations must support silent message delivery without generating notifications
- Users must be able to schedule messages to be sent up to 30 days in the future
- Scheduled messages must be queued locally and sent automatically at the specified time using background task execution
- System must support offline message queuing for scheduled messages when device is offline at send time
- Users must be able to edit or cancel scheduled messages before they are sent
- Mute settings must be stored locally and persist across app restarts
- Scheduled messages must maintain end-to-end encryption and be processed through normal message delivery pipeline
- System must handle timezone changes and daylight saving time adjustments for scheduled messages

**Out of Scope**:
- Recurring or repeating scheduled messages
- Global notification settings management
- Integration with external calendar systems

---

### WO-69: Implement Secure Backup and Recovery System for Hidden Folders

**Blueprint:** Hidden Folders with Biometric Protection

## Summary

Build the hidden folder backup and recovery system using BIP39 recovery phrases. Users can create encrypted backups of hidden folder contents that can be restored on the same or a new device. Backups require biometric authentication plus the recovery phrase. Local-only storage (no cloud upload).

## In Scope

- BIP39 12-word recovery phrase generation (`BIP39.generateMnemonic()` using BIP39 wordlist)
- Recovery phrase display: show during setup, require user to verify 3 random words
- Backup encryption: `AES-256-GCM(key: PBKDF2(phrase, salt, iterations: 100_000), data: backupPayload)`
- Backup creation: requires biometric auth + recovery phrase confirmation; encrypts all hidden folder data
- Backup payload: folder metadata, encrypted messages, key derivation hints (no keys stored in backup — re-derived from phrase + biometric)
- Backup storage: local app Documents directory, `.excludedFromBackup` attribute on file (no iCloud sync)
- Backup restoration: requires biometric auth + recovery phrase; verify phrase, re-derive keys, decrypt and restore
- Backup scheduling: daily, weekly, monthly automatic backups
- Secure backup deletion: zero-fill backup file before removing

## Out of Scope

- Cloud backup storage (local only by design)
- Cross-device sync of live hidden folder data

## Requirements

Derived from the Hidden Folders blueprint.

**Backup Encryption:**
```swift
// Core/HiddenFolders/HiddenFolderBackupManager.swift
struct HiddenFolderBackupManager {
    func createBackup(folders: [HiddenFolder], phrase: String) async throws -> URL {
        // 1. Derive backup key from recovery phrase (PBKDF2)
        let phraseData = Data(phrase.utf8)
        let salt = generateSalt()  // 16 bytes random
        let backupKey = try PBKDF2.deriveKey(password: phraseData, salt: salt, iterations: 100_000, keySize: 32)

        // 2. Serialize and encrypt backup payload
        let payload = BackupPayload(folders: folders, exportedAt: Date())
        let encrypted = try AES256GCM.encrypt(serialize(payload), key: SymmetricKey(data: backupKey))

        // 3. Write to local file (excluded from iCloud backup)
        let url = backupDirectory.appendingPathComponent("hidden_backup_\(Date().ISO8601Format()).enc")
        try Data(salt + encrypted).write(to: url)
        try (url as NSURL).setResourceValue(true, forKey: .isExcludedFromBackupKey)
        return url
    }
}
```

## Blueprints

- Hidden Folders with Biometric Protection — Defines biometric-protected backup, BIP39 recovery phrase, AES-256-GCM backup encryption, multi-device restoration, integrity verification, and local-only storage

---

### WO-72: Implement Core Persona Data Model and Management System

**Blueprint:** Multiple Personas with Selective Visibility

## Summary

Build the core persona data model and management system — the foundational layer for Multiple Personas. Users create up to 5 personas, each with independent profile data (name, avatar, bio, category) while sharing the master DID and trust score. Persona management CRUD operations are backed by SwiftData locally and synced to the backend.

## In Scope

- `Persona` data model: `{id, masterDID, displayName, avatarURL, bio, category, createdAt, trustTier (inherited), customVerificationBadges[]}`
- Persona creation UI: name (2–50 chars), optional avatar, optional bio, category selection (Professional, Personal, Family, Gaming, Custom)
- Persona limit enforcement: max 5 per account (enforced both client and backend)
- Persona editing: update name, avatar, bio, category
- Persona deletion: confirmation dialog + 7-day grace period (soft delete, conversation history archived)
- Persona list management interface: shows count, allows reorder, quick switch
- Backend: `POST /v1/personas`, `GET /v1/personas`, `PATCH /v1/personas/{id}`, `DELETE /v1/personas/{id}` on Identity Service
- Audit log: all persona operations timestamped in local SwiftData for security review

## Out of Scope

- Selective visibility controls (WO-82)
- Persona-specific conversation isolation (WO-91)
- Trust scoring per persona (shared from master DID, see WO-111)
- Blockchain anchoring (WO-122)

## Requirements

Derived from the Multiple Personas blueprint.

**Persona Data Model:**
```swift
// Domain/Models/Persona.swift
struct Persona: Identifiable, Codable {
    let id: UUID
    let masterDID: String          // Links to user's primary DID
    var displayName: String        // 2–50 chars, unique per user
    var avatarURL: URL?
    var bio: String?               // Optional, max 200 chars
    var category: PersonaCategory  // .professional, .personal, .family, .gaming, .custom(String)
    let createdAt: Date
    var isActive: Bool             // Can be deactivated without deletion
    // Trust score inherited from master DID (not stored per persona)
    // Verification badges are persona-specific (separate from master trust tier)
}

enum PersonaCategory: Codable {
    case professional, personal, family, gaming
    case custom(String)  // User-defined label
}
```

## Blueprints

- Multiple Personas with Selective Visibility — Defines persona creation (up to 5), display name/avatar/bio/category, master DID linkage, editing, deletion with grace period, and trust score inheritance

---

### WO-75: Build Disappearing Message Configuration and Timer System

**Blueprint:** Disappearing Messages with Cryptographic Verification

## Summary

Build the disappearing message configuration UI — expiration time pickers at both conversation and message levels — and the `CountdownTimer` SwiftUI component. The combined Frontend blueprint adds an important tier-based constraint: **free tier users are capped at 24 hours maximum** disappearing message duration; **VIP users can set up to 1 year**.

## In Scope

- Per-conversation disappearing messages toggle with expiration picker (settings drawer)
- Per-message expiration override via long-press context menu
- **Free tier preset intervals:** 10 seconds, 1 minute, 5 minutes, 1 hour, 1 day (24h max for free)
- **VIP preset intervals:** All free presets + 7 days, 30 days, 90 days, 365 days (1 year max)
- Custom duration picker: VIP only — any value from 10 seconds to 365 days; validated with tier check
- `CountdownTimer` SwiftUI component: displays formatted remaining time, ticks every second in foreground, fires `onExpire` callback when elapsed
- Expiration time embedded in encrypted message metadata (recipient reads `expiresAt` from decrypted payload)
- Conversation header indicator showing active disappearing message setting
- Warning banner in disappearing conversations: "⚠️ Screenshots cannot be prevented on iOS. Only send disappearing messages to trusted contacts."
- **Tier gate enforcement:** if free user attempts to set > 24h, show VIP upsell prompt
- Timer format options: exact countdown (`5:23`) or relative (`in 5 minutes`)
- Accessibility: VoiceOver announcements for timer state

## Out of Scope

- Key destruction (WO-66)
- Message deletion from SwiftData (WO-84)
- Trust score-based restrictions on minimum timer durations (WO-143)
- Backend timer state (timers are fully client-side, no server sync)
- VIP subscription purchase flow (WO-238)

## Requirements

From the combined Frontend blueprint (Messaging Features section):

**Tier-based disappearing message constraint (NEW):**
> "Disappearing Messages — Auto-delete messages after specified time (**up to 1 year for VIP, 24 hours max for free tier**)"

**Expiration Config UI:**
```swift
enum DisappearingInterval: Int, CaseIterable {
    // Free tier + VIP
    case tenSeconds = 10
    case oneMinute = 60
    case fiveMinutes = 300
    case oneHour = 3600
    case oneDay = 86400        // 24h — FREE TIER MAXIMUM

    // VIP only
    case sevenDays = 604800
    case thirtyDays = 2592000
    case ninetyDays = 7776000
    case oneYear = 31536000   // 365 days — VIP MAXIMUM

    case custom = -1           // VIP only — validated 10s min, 1yr max

    var requiresVIP: Bool {
        switch self {
        case .sevenDays, .thirtyDays, .ninetyDays, .oneYear, .custom: return true
        default: return false
        }
    }
}

// Tier gate: show VIP upsell if free user selects VIP-only interval
func validateSelection(_ interval: DisappearingInterval, isVIP: Bool) -> Bool {
    if interval.requiresVIP && !isVIP {
        // Show VIP upsell: "Upgrade to VIP for up to 1 year disappearing messages"
        return false
    }
    return true
}
```

**CountdownTimer Component:**
```swift
struct CountdownTimer: View {
    let expiresAt: Date
    let onExpire: (Bool) -> Void
    @State private var timeRemaining: TimeInterval = 0

    var body: some View {
        Text(formatTime(timeRemaining))
            .font(.caption).foregroundColor(.secondary)
            .onAppear { startTimer() }
    }
    private func startTimer() {
        timeRemaining = max(0, expiresAt.timeIntervalSinceNow)
        Timer.scheduledTimer(withTimeInterval: 1.0, repeats: true) { timer in
            timeRemaining = max(0, expiresAt.timeIntervalSinceNow)
            if timeRemaining <= 0 { timer.invalidate(); onExpire(true) }
        }
    }
}
```

## Blueprints

- Disappearing Messages with Cryptographic Verification — Defines preset intervals, per-conversation vs. per-message settings, `CountdownTimer`, and screenshot warning
- Frontend — Specifies VIP vs. free tier disappearing message duration limits (VIP: 1 year max; Free: 24 hours max)

---

### WO-76: Implement Secure Local Storage for Scheduled Messages

**Blueprint:** Silent and Scheduled Private Chats

## Summary

Build the secure local storage system for pending scheduled messages. Messages are encrypted with AES-256-GCM using a device-specific key before being written to SwiftData. Automatic cleanup occurs within 24 hours of successful delivery. Supports backup and recovery for device migration.

## In Scope

- `ScheduledMessageStore` actor wrapping SwiftData model container
- Per-message AES-256-GCM encryption with HKDF-derived device-specific key (from Secure Enclave, not biometric)
- Tamper detection: SHA-256 integrity hash stored alongside encrypted content; verified on read
- Automatic cleanup: delete from SwiftData + cancel `BGTask` within 24 hours of delivery confirmation
- Backup support: scheduled messages exported as encrypted package (same mechanism as archive backup, WO-64)
- Recovery: restore from backup on device migration or reinstall, re-register `BGTask` for future messages
- Storage optimization: compress message content before encryption if > 1KB
- Capacity alerting: warn user when scheduled message storage exceeds 100MB (rarely reached for text/small media)
- Audit log of operations: create, edit, cancel, deliver events stored locally

## Out of Scope

- Scheduling logic and BGTask registration (WO-65)
- Time-locked smart contracts (not used)
- Blockchain anchoring (WO-96)

## Requirements

Derived from the Silent and Scheduled Private Chats blueprint.

**Local Storage Architecture:**
```swift
// Core/ScheduledMessages/ScheduledMessageStore.swift
@ModelActor actor ScheduledMessageStore {
    private let encryptor: AES256GCMEncryptor

    func save(_ message: ScheduledMessage) async throws {
        let compressed = try zstdCompress(serialize(message.plainContent))
        let encrypted = try encryptor.encrypt(compressed)  // HKDF from Secure Enclave
        let hash = sha256(encrypted)

        let record = ScheduledMessageRecord(
            id: message.id,
            encryptedContent: encrypted,
            integrityHash: hash,
            scheduledFor: message.scheduledFor,
            status: .pending
        )
        modelContext.insert(record)
        try modelContext.save()
    }

    func loadAndVerify(id: UUID) async throws -> ScheduledMessage {
        let record = try fetch(id: id)
        guard sha256(record.encryptedContent) == record.integrityHash else {
            throw StorageError.integrityViolation
        }
        let compressed = try encryptor.decrypt(record.encryptedContent)
        return try deserialize(zstdDecompress(compressed))
    }
}
```

## Blueprints

- Silent and Scheduled Private Chats — Defines local encryption, secure storage, automatic cleanup, backup support, recovery options, storage optimization, and audit logging for scheduled messages

---

### WO-78: Create Access Logging and Security Auditing System

**Blueprint:** Hidden Folders with Biometric Protection

## Summary

Build the access logging and security auditing system for hidden folders — recording every authentication attempt, folder access event, and security anomaly in an encrypted local log. Users can review their access history to detect unauthorized access. The system flags suspicious patterns (failed attempts, unusual timing).

## In Scope

- `HiddenFolderAuditLog` SwiftData model: `{eventId, timestamp, eventType, authMethod, folderId, success, deviceInfo, duration}`
- Event types: `folderUnlocked`, `folderLocked`, `authFailed`, `folderCreated`, `folderDeleted`, `conversationMoved`, `backupCreated`, `backupRestored`
- Encrypted log storage: AES-256-GCM with device-specific key (not biometric-derived, so accessible for audit without biometric)
- Log retention: 90 days automatic purge of old entries
- User audit view: list of access events with timestamp, authentication method, and outcome
- Suspicious pattern detection: flag if 3+ failed attempts within 10 minutes; flag if access at unusual hour (user-defined quiet hours)
- Optional location logging: GPS coordinates on access events (user opt-in, requires location permission)
- Log tampering detection: chain-hash integrity linking each log entry to prevent deletion of individual records

## Out of Scope

- Real-time security alerting (local analysis only)
- Log export or sharing
- Network-based threat detection

## Requirements

Derived from the Hidden Folders blueprint.

**Audit Log Entry:**
```swift
// Domain/Models/HiddenFolderAuditEntry.swift
struct HiddenFolderAuditEntry: Identifiable, Codable {
    let id: UUID
    let timestamp: Date
    let eventType: AuditEventType
    let authMethod: BiometricMethod?  // .faceID, .touchID, .pin
    let folderId: UUID?
    let success: Bool
    let sessionDuration: TimeInterval?
    let appVersion: String
    let osVersion: String
    let previousEntryHash: Data  // Chain-hash for tamper detection
}

// Suspicious pattern detection:
func detectSuspiciousPatterns(recentEntries: [HiddenFolderAuditEntry]) -> [SecurityAlert] {
    let recentFailures = recentEntries.filter { !$0.success && Date().timeIntervalSince($0.timestamp) < 600 }
    if recentFailures.count >= 3 {
        return [SecurityAlert(type: .repeatedAuthFailures, count: recentFailures.count)]
    }
    return []
}
```

## Blueprints

- Hidden Folders with Biometric Protection — Defines access timestamp logging, biometric method logging, failed attempt logging, access duration, optional location logging, log encryption, 90-day retention, and security pattern detection

---

### WO-82: Build Selective Visibility and Permission Control System

**Blueprint:** Multiple Personas with Selective Visibility

## Summary

Build the selective visibility and permission control system that governs which contacts can see each persona. Contacts cannot discover personas they haven't been granted access to. Permissions are managed through an explicit grant/revoke interface and enforced both client-side and on the backend.

## In Scope

- `PersonaPermission` data model: `{personaId, grantedToDID, grantedAt, revokedAt?}`
- Permission grant UI: for each persona, show contact list with grant/revoke toggles
- Contact-persona access matrix view: who can see which of my personas
- Backend enforcement: Identity Service returns persona data only to DIDs in permission list
- Server-side zero-knowledge property: non-permitted contacts receive no indication that additional personas exist
- Permission revocation: immediate effect — revoked contact's persona view disappears on next profile fetch
- Permission audit log: `{action: grant|revoke, personaId, contactDID, timestamp}` stored locally and synced to backend
- Permission sync: stored in PostgreSQL Contacts Service, cached in Redis per persona

## Out of Scope

- Persona data model (WO-72)
- Conversation isolation (WO-91)
- Blockchain anchoring of permissions (WO-122)

## Requirements

Derived from the Multiple Personas blueprint.

**Permission Model:**
```swift
struct PersonaPermission: Codable {
    let personaId: UUID
    let grantedToDID: String
    let grantedAt: Date
    var revokedAt: Date?

    var isActive: Bool { revokedAt == nil }
}

// Backend enforcement:
// GET /v1/profile/{did} with ?personaId={id}
// Returns 404 if caller's DID not in permission list (not 403, to avoid confirming persona existence)
```

**Permission Check (Backend):**
```go
func (s *PersonaService) GetPersonaForCaller(personaID, requesterDID string) (*Persona, error) {
    if !s.db.HasPermission(personaID, requesterDID) {
        return nil, ErrNotFound  // Return 404, not 403 (don't reveal persona exists)
    }
    return s.db.GetPersona(personaID)
}
```

## Blueprints

- Multiple Personas with Selective Visibility — Defines permission granting/revoking, access matrix, permission enforcement, persona discovery prevention, and permission auditing

---

### WO-87: Integrate Trust Score Limitations for Silent and Scheduled Messages

**Blueprint:** Silent and Scheduled Private Chats

## Summary

Implement trust score-based rate limits for silent and scheduled messaging to prevent abuse. Low-trust users face daily frequency limits; violations result in progressive feature suspension. Limits are enforced at the backend Message Relay and visible to users through a quota display in settings.

## In Scope

- Backend rate limit enforcement per DID per day: query Trust Service (port 8003) tier before message relay
- Silent message rate limit: Tier 1 (≤10/day per recipient), Tier 2+ (50/day per recipient), Tier 3+ (200/day), Tier 4+ (unlimited)
- Scheduled message rate limit: Tier 1 (5/day total), Tier 2 (20/day), Tier 3 (100/day), Tier 4+ (unlimited)
- Rate limit storage in Redis: `{DID}:silent_count:{YYYY-MM-DD}` counters with daily TTL
- HTTP 429 response when limit exceeded with `{quota, used, resets_at}` payload
- Progressive feature suspension on repeated violations: 24h → 7d → 30d suspension
- iOS quota display: "Silent messages used today: 3/10" in per-conversation settings
- Limit transparency: clear explanation of why limited and when quota resets

## Out of Scope

- Trust score computation (WO-181)
- Silent message core (WO-56)
- Message scheduling core (WO-65)

## Requirements

Derived from the Silent and Scheduled Private Chats blueprint.

**Rate Limit Matrix:**

| Trust Tier | Silent/Day (per recipient) | Scheduled/Day (total) |
|---|---|---|
| Tier 1 (Unverified) | 10 | 5 |
| Tier 2 (Newcomer) | 50 | 20 |
| Tier 3 (Member) | 200 | 100 |
| Tier 4+ (Verified) | Unlimited | Unlimited |

**Backend Rate Check:**
```go
func (s *RelayService) CheckSilentMessageLimit(senderDID, recipientDID string) error {
    tier := s.trustService.GetTier(senderDID)
    limit := silentMessageLimit(tier)
    key := fmt.Sprintf("%s:silent:%s:%s", senderDID, recipientDID, today())
    count, _ := s.redis.Incr(ctx, key).Result()
    s.redis.Expire(ctx, key, 24*time.Hour)
    if count > int64(limit) { return ErrRateLimitExceeded }
    return nil
}
```

## Blueprints

- Silent and Scheduled Private Chats — Defines trust score-based rate limiting, daily frequency limits, limit escalation on abuse, user transparency, and rate limiting enforcement

---

### WO-96: Implement Blockchain Anchoring for Scheduled Message Delivery Verification

**Blueprint:** Silent and Scheduled Private Chats

## Summary

Implement blockchain anchoring for scheduled message delivery verification. After a scheduled message is delivered via the relay, its delivery commitment hash is submitted to the Constellation Data L1 — creating an immutable, cryptographically verifiable proof that the message was delivered at the specific intended timestamp.

## In Scope

- After scheduled message relay transmission: submit delivery event to `AnchoringBatcher` with delivery timestamp metadata
- Delivery commitment: `H(H(messageContent) || deliveryTimestamp || nonce)` — standard message commitment but with delivery timestamp embedded in metadata
- Data L1 submission includes: delivery timestamp, scheduled time, sender commitment hash (not content)
- Timestamp verification API: `GET /v1/scheduled-messages/{messageId}/delivery-proof` — returns `{snapshotHash, snapshotHeight, merkleProof, deliveryTimestamp}`
- Proof export: serialized JSON proof showing delivery time vs. scheduled time gap (proves on-time delivery)
- Compliance recording: store `{messageId, scheduledFor, deliveredAt, txHash}` in PostgreSQL for audit queries

## Out of Scope

- Message scheduling core (WO-65)
- General message anchoring (WO-15 — this is specifically for scheduled message timing verification)
- Smart contract automation

## Requirements

Derived from the Silent and Scheduled Private Chats blueprint.

**Delivery Proof Structure:**
```go
type ScheduledDeliveryProof struct {
    MessageID       string    `json:"message_id"`
    ScheduledFor    time.Time `json:"scheduled_for"`
    ActualDelivery  time.Time `json:"actual_delivery"`
    TimingDeltaMs   int64     `json:"timing_delta_ms"` // Actual vs. scheduled
    CommitmentHash  []byte    `json:"commitment_hash"`
    SnapshotHash    string    `json:"snapshot_hash"`
    SnapshotHeight  int       `json:"snapshot_height"`
    MerkleProof     [][]byte  `json:"merkle_proof"`
    VerificationURL string    `json:"verification_url"` // DAG Explorer link
}
// Serialized to JSON for export; shareable as delivery proof
```

## Blueprints

- Silent and Scheduled Private Chats — Defines blockchain anchoring of delivery timestamps, timestamp verification, integrity proof generation, proof sharing, and compliance recording

---

### WO-102: Develop Persona Privacy Settings and Contact Management

**Blueprint:** Multiple Personas with Selective Visibility

## Summary

Build per-persona privacy settings and persona-aware contact management. Each persona can have independent privacy controls (last seen, online status, profile picture, status message) that override the global settings when interacting as that persona. Contact lists show which contacts know about each persona.

## In Scope

- Per-persona `PrivacySettings` struct (same fields as global `PrivacySettings` but stored per `personaId`)
- Per-persona privacy settings UI in persona editing screen
- Per-persona blocking: `PersonaBlockList {personaId, blockedDID}` — blocked contacts cannot interact with that specific persona
- Contact list filtered by persona: "Who knows about this persona?" view showing contacts with permission
- Contact categorization: contacts tagged by which personas they know (auto-derived from permission grants)
- Automatic persona suggestion: based on contact's known persona + recent conversation pattern
- Privacy settings sync: stored in PostgreSQL Identity Service, cached in Redis
- Settings apply immediately: update backend, invalidate cached profile for affected contacts

## Out of Scope

- Global privacy settings (WO-39)
- Persona creation (WO-72)
- Permission granting UI (WO-82)

## Requirements

Derived from the Multiple Personas blueprint.

**Per-Persona Privacy Settings:**
```swift
// Each Persona has its own PrivacySettings override
struct PersonaPrivacySettings: Codable {
    let personaId: UUID
    var showLastSeen: VisibilityLevel
    var showOnlineStatus: VisibilityLevel
    var showProfilePicture: VisibilityLevel
    var showStatusMessage: VisibilityLevel
    var allowGroupInvites: VisibilityLevel
    var allowCalls: VisibilityLevel
    // When nil, falls back to global PrivacySettings
}

// Per-persona notification preferences:
struct PersonaNotificationSettings: Codable {
    let personaId: UUID
    var messageNotifications: Bool
    var callNotifications: Bool
    var groupActivityNotifications: Bool
}
```

## Blueprints

- Multiple Personas with Selective Visibility — Defines per-persona privacy settings, notification preferences, contact management with persona awareness, per-persona blocking, and immediate settings enforcement

---

### WO-125: Develop Screenshot Prevention and Security Monitoring System

**Blueprint:** Disappearing Messages with Cryptographic Verification

**Purpose**: Implement device-level security measures to detect and prevent screenshots and message forwarding of disappearing messages, while providing user notifications about protection limitations and potential security bypasses to maintain transparency about privacy boundaries.

**Requirements**:
- Implement screenshot detection that identifies when users attempt to capture disappearing message content
- Provide screenshot prevention mechanisms that block or obscure content during capture attempts where technically feasible
- Generate screenshot notification alerts to all conversation participants when screenshot attempts are detected
- Implement forwarding prevention that blocks copy, share, and forward actions on disappearing message content
- Detect forwarding attempts and notify conversation participants when users try to share disappearing message content
- Provide clear user awareness notifications about protection limitations and potential bypass methods
- Support multiple device platforms with appropriate security measures for each operating system
- Handle protection limitation scenarios gracefully by informing users when security measures may be ineffective
- Maintain security event logging that records all prevention attempts and bypass notifications
- Ensure security measures don't interfere with legitimate accessibility features or assistive technologies

**Out of Scope**:
- Root/jailbreak detection
- Third-party app monitoring
- Hardware-level security enforcement
- Legal enforcement of screenshot policies

---

### WO-134: Create Blockchain-Based Audit Trail System

**Blueprint:** Verified Financial Institution Integration

## Summary

Build the blockchain-based audit trail system for financial institution communications — anchoring interaction event hashes on the Constellation Data L1 for tamper-proof record keeping. Records message exchanges, transaction verifications, and document access events without exposing sensitive personal or financial data.

## In Scope

- Interaction event anchoring: submit `H(eventData || salt)` to Data L1 for each: message sent/received, transaction authorization, document accessed/signed, fraud alert responded to
- Event data includes: `{timestamp, institutionDID, customerDID_hash, eventType, interactionId}` — hash committed, not plaintext
- Data L1 submission: `{type: "financial_audit", interactionHash, eventType, institutionDID, timestamp}` via Metagraph Gateway
- Retention policy: 7-year minimum for financial communications (enforced via IPFS pinning TTL)
- Audit trail query: `GET /v1/financial/audit?institutionDID={did}&from={date}&to={date}` — returns interaction records with on-chain references
- Regulatory access: authorized regulator can query audit trail with time-limited access token issued by platform
- Verification: regulators verify audit completeness by cross-referencing Data L1 snapshot records

## Out of Scope

- Raw message content in audit trail (privacy)
- Specific regulatory reporting formats
- External regulatory system integration

## Requirements

```go
type FinancialAuditEvent struct {
    EventID         string    // UUID
    EventType       string    // "message", "transaction_auth", "document_access"
    InteractionHash []byte    // H(eventDetails || salt)
    InstitutionDID  string
    CustomerDIDHash []byte    // H(customerDID || salt) — privacy-preserving
    Timestamp       time.Time
    SnapshotHash    string    // Set after Data L1 anchoring
}
```

## Blueprints

- Verified Financial Institution Integration — Defines blockchain audit trail, regulatory compliance recording, interaction logging, and verification capabilities

---

### WO-135: Implement Screenshot and Forwarding Prevention System

**Blueprint:** Disappearing Messages with Cryptographic Verification

**Purpose**: Provide technical safeguards against unauthorized content capture and sharing of disappearing messages, while maintaining user awareness of protection limitations and potential bypasses to ensure informed usage of the disappearing message feature.

**Requirements**:
- Detect screenshot attempts on disappearing messages and prevent capture when technically possible through platform-specific APIs
- Block message forwarding functionality for disappearing messages within the application interface
- Generate real-time notifications to all conversation participants when screenshot attempts are detected
- Display forwarding prevention notifications when users attempt to forward disappearing message content
- Provide clear user education about protection limitations including potential bypasses through external cameras or screen recording
- Implement platform-specific prevention measures for iOS, Android, and desktop environments with graceful degradation
- Log all prevention events and bypass attempts for security monitoring and user awareness
- Support accessibility features while maintaining security measures for users with assistive technologies
- Ensure prevention measures work consistently across different device orientations and display configurations
- Provide user settings to control notification preferences for screenshot and forwarding prevention alerts

**Out of Scope**:
- Prevention of external recording devices or cameras pointed at screens
- Integration with trust scoring system
- Message deletion functionality
- Cryptographic proof generation

---

### WO-143: Build Trust Score-Based Disappearing Message Restrictions

**Blueprint:** Disappearing Messages with Cryptographic Verification

**Purpose**: Integrate disappearing message functionality with the existing trust scoring system to prevent abuse and harassment by restricting very short expiration timeframes for users with low trust scores, while maintaining transparency about restrictions and providing appeal mechanisms.

**Requirements**:
- Enforce minimum expiration timeframes based on user trust score levels, preventing users with low trust scores from setting very short disappearing times
- Define trust score thresholds that determine available expiration options, with higher trust scores unlocking shorter timeframe options
- Implement restriction escalation that increases minimum timeframes for users who repeatedly abuse disappearing message features
- Provide transparent restriction notifications that explain why certain expiration options are unavailable to users
- Support restriction appeals process that allows users to request review of their disappearing message limitations
- Monitor restriction enforcement effectiveness and generate reports on abuse prevention metrics
- Ensure restriction transparency by clearly displaying available expiration options based on current trust score
- Handle trust score updates in real-time, immediately adjusting available expiration options when scores change
- Provide graduated restriction levels rather than binary restrictions, offering multiple tiers of access based on trust score ranges
- Log all restriction enforcement events for audit purposes and pattern analysis of potential abuse

**Out of Scope**:
- Trust score calculation algorithms
- Trust score display interfaces
- Message deletion functionality
- Encryption key management

---

### WO-162: Implement Immutable Audit Trail and Blockchain Recording System

**Blueprint:** Verified Financial Institution Integration

## Summary

Build the comprehensive immutable audit trail and blockchain recording system for financial institution communications — extending WO-134 with ZK proof privacy preservation for customer data, SOX/FDIC compliance-grade records, and automated regulatory report generation.

## In Scope

- All capabilities from WO-134 (interaction event hashing, Data L1 anchoring)
- ZK proof privacy layer: customer interactions represented as ZK commitments — regulators can verify interaction count and timing without seeing participant identities
- SOX compliance records: specific event types required for SOX Section 302/906 financial communication attestations
- FDIC compliance recording: message delivery confirmations and customer response records in auditable format
- Automated compliance report generation: `POST /v1/financial/compliance-reports` — generates PDF reports for specified period meeting examination requirements
- Audit completeness verification: regulators can cryptographically verify no gaps in audit trail using Merkle proofs from Data L1 snapshots
- Cross-reference capability: link audit trail entries to specific transaction IDs and message IDs for investigation support
- Minimum retention: 7 years for all financial communications

## Out of Scope

- WO-134 base functionality (this extends it)
- External regulatory system submission integrations

## Requirements

Derived from the Verified Financial Institution Integration blueprint.

```
ZK Audit Proof: Proves "institution sent X messages in period Y to customers in segment Z"
               without revealing: customer DIDs, message content, individual identities
               Used for: regulatory examination compliance
```

## Blueprints

- Verified Financial Institution Integration — Defines immutable audit trails, regulatory compliance recording, ZK proofs for privacy, and regulatory examination support

---

### WO-183: Build Zero-Knowledge Proof Integration System

**Assignee:** Chad Cromwell

**Blueprint:** Decentralized Identity and Authentication

## Summary

Implement the ZK proof infrastructure connecting all components of the Zero-Knowledge Proofs and Midnight Integration blueprint. This work order covers the overall ZK system architecture, proof type registry, and the Phase 1–2 fallback verification that operates before Midnight is available. The dedicated implementation work orders are: WO-212 (Midnight PoC/infrastructure), WO-235 (iOS `ZKProofUseCase`), and WO-236 (Go backend verification service).

## In Scope

- **Proof type registry:** Central registry of supported ZK claim types, expected Midnight circuit names, public signal schemas, and caching TTLs
- **Phase 1–2 fallback verification (non-ZK):** Before Midnight is available, implement hash-commitment verification: `H(score || nonce)` compared against Cardano trust commitment UTXO datum; `H(birthYear || nonce)` for age claims
- **Challenge nonce service:** `GET /zk/challenge` endpoint generates single-use cryptographic nonce (TTL: 5 minutes); included in proof public signals to prevent replay attacks; `POST /zk/nonce/consume` marks nonce as used
- **Proof result caching:** Redis-based boolean result cache per `{DID, claimType, threshold}`:
  - Trust tier: 5-minute TTL
  - Age verification: 1-hour TTL
  - KYC compliance: 1-hour TTL
  - Group membership: 5-minute TTL
  - Balance threshold: 5-minute TTL
- **ZK audit log:** Log ZK verification results (type, DID_hash, result, timestamp) — no underlying credential data stored

## Out of Scope

- Midnight testnet PoC and Phase 3 evaluation (WO-212)
- iOS `ZKProofUseCase` with Midnight SDK (WO-235)
- Go backend Midnight client and `/zk/verify/:claimType` endpoints (WO-236)
- ZK circuit deployment on Midnight (WO-212)

## Requirements

From the Zero-Knowledge Proofs and Midnight Integration blueprint:

**5 ZK Claim Types:**
| Claim | Circuit | Phase | Private Input |
|---|---|---|---|
| Trust tier minimum | `TrustTierVerifier` | Phase 3–4 | Exact trust score |
| Age verification | `AgeVerifier` | Phase 3–4 | Birthdate from gov ID credential |
| KYC compliance | `KYCVerifier` | Phase 4 | Passport data, IDV provider details |
| Group membership | `GroupMembershipProver` | Phase 4 | Full group list |
| Balance threshold | `BalanceThresholdProver` | Phase 4 | Exact ECHO balance |

**Midnight Architecture:**
```plaintext
Cardano (source of truth) ←→ Midnight Bridge ←→ Midnight ZK Layer
                                                    ↑
                                         iOS generates proof on-device
                                         (private inputs never leave device)
                                                    ↓
                                         Go backend forwards to Midnight
                                         Caches boolean result only
```

**Graceful Degradation (Phase 1–2 and Midnight outage):**
Hash-commitment fallback: compare `H(score || nonce)` against on-chain trust commitment. Slightly weaker privacy (reveals tier confirmation) but maintains system availability.

## Blueprints

- Zero-Knowledge Proofs and Midnight Integration — Defines full Midnight architecture, 5 Compact DSL ZK circuits, iOS proof generation, Go verification flow, graceful degradation, and phase rollout
- Decentralized Identity and Authentication — Defines trust tier and credential structures that ZK proofs verify against

---

### WO-212: Implement Zero-Knowledge Proofs and Midnight Integration

**Type:** Build

**Blueprint:** Zero-Knowledge Proofs and Midnight Integration

## Summary

Execute the Phase 3 Midnight testnet evaluation and PoC, acquire NIGHT/DUST tokens, and deploy the ZK trust tier and age verification proofs to Midnight mainnet at Phase 3–4 transition. The full iOS Midnight SDK integration and Go backend verification service are covered by dedicated work orders. This work order covers the infrastructure, PoC validation, and Phase 4 readiness gate.

## In Scope

- **Phase 3 evaluation gate:** Monitor Midnight mainnet stability; document go/no-go criteria for ECHO integration (uptime %, proof generation time < 5s on iPhone 12, DUST burn rate at expected ZK operation volume)
- **NIGHT token acquisition:** Acquire minimal NIGHT holdings from Midnight chain; set up DUST generation for ZK computation fees
- **Compact contract deployment (testnet):** Deploy `TrustTierVerifier` and `AgeVerifier` contracts on Midnight testnet using Compact DSL (TypeScript-based); validate end-to-end flow with iOS PoC device
- **Cardano ↔ Midnight bridge testing:** Validate that Midnight contracts can read Cardano trust tier UTXO datums via the native IOG-built bridge; document any latency or reliability issues
- **Phase 4 mainnet contract deployment:** Deploy production ZK contracts (`TrustTierVerifier`, `AgeVerifier`, and Phase 4+ `KYCVerifier`, `GroupMembershipProver`, `BalanceThresholdProver`) to Midnight mainnet
- **Graceful degradation setup:** Configure backend Trust Service fallback to hash-commitment verification (`H(score || nonce)` comparison) when Midnight is unavailable

## Out of Scope

- iOS `ZKProofUseCase` Midnight SDK integration (separate work order)
- Go backend ZK verification endpoint (separate work order)
- Full KYC/group/balance proofs (Phase 4 — included in mainnet contract deployment above but not the full backend/iOS integration)

## Requirements

From the Zero-Knowledge Proofs and Midnight Integration blueprint:

**Phase Rollout:**
| Phase | Action |
|---|---|
| Phase 3 | Evaluate Midnight mainnet stability; run ZK trust tier PoC on Midnight testnet |
| Phase 3–4 | Deploy ZK trust tier + age proofs to Midnight mainnet |
| Phase 4 | Add KYC compliance, group membership, balance threshold proofs |

**Midnight Token Model:** NIGHT (governance/staking) + DUST (renewable, non-tradable, pays for ZK computation). ECHO does not need large NIGHT holdings — ZK calls consume DUST generated from minimal NIGHT at predictable rate.

**Graceful Degradation:** If Midnight unavailable, backend falls back to `H(score || nonce)` hash-commitment verification — slightly weaker privacy but maintains system availability.

## Blueprints

- Zero-Knowledge Proofs and Midnight Integration — Defines Midnight architecture, NIGHT/DUST token model, Compact DSL contracts, Cardano bridge, graceful degradation, and phase rollout timeline

---

### WO-218: Implement GDPR Account Deletion and Cryptographic Key Erasure

**Type:** Build

**Blueprint:** Privacy Architecture and Secure Data Handling

## Summary

Implement GDPR right-to-erasure compliance through cryptographic key deletion rather than data deletion. Deleting the keys makes all encrypted content permanently unreadable — functionally equivalent to deletion even if encrypted ciphertext persists in caches or offline queues. Includes DID deactivation on Cardano and token handling before account removal.

## In Scope

- Account deletion iOS flow: Settings → Delete Account → confirm identity (biometric) → choose token disposition (burn or transfer) → initiate deletion
- Backend deletion sequence:
  1. Delete all Keychain items and local SwiftData on device (including derived storage key)
  2. Backend wipes offline message queue (Redis/PostgreSQL) for the user's DID — all ciphertext destroyed
  3. DID document marked as `deactivated` on Cardano — no new messages can be sent to this DID
  4. On-chain Merkle roots remain (contain no personal data — only opaque hashes)
  5. Token balance: execute user's choice (burn via `TokenBurner` or transfer to specified wallet)
  6. Backend deletes account record from PostgreSQL (DID-to-account mapping)
- Deletion confirmation: provide user with deletion receipt showing steps completed
- Deletion audit log (anonymized): `{deletionId, timestamp, stepsCompleted[]}` — no DID in log
- 30-day grace period option: soft-delete with reactivation possible before permanent erasure

## Out of Scope

- Legal hold exemptions (enterprise compliance work orders)
- Data export (separate GDPR data portability feature)
- Third-party credential revocation (handled by WO-182 credential revocation)

## Requirements

From the Primary Architecture and Secure Data Handling blueprint:

**Erasure Process:**
1. User requests account deletion
2. Backend deletes all Keychain and local storage on device (including derived keys)
3. Backend wipes offline message queue (ephemeral ciphertext destroyed)
4. DID document is marked as deactivated on Cardano
5. On-chain Merkle roots remain (they contain no personal data — only opaque hashes)
6. Token balance is either burned or transferred before deletion (user choice)

**GDPR Note:** On-chain Merkle roots that persist prove "messages existed at timestamps" without revealing content or identity — they are not personal data under GDPR's definition.

## Blueprints

- Primary Architecture and Secure Data Handling — Defines GDPR right-to-erasure via cryptographic key deletion, the six-step erasure process, and the legal basis for retaining on-chain commitments post-deletion

---

### WO-219: Implement Phase 3 Sealed Sender Protocol

**Type:** Build

**Blueprint:** Privacy Architecture and Secure Data Handling

## Summary

Implement the Phase 3 sealed sender protocol — hiding the sender's DID from the relay server by encrypting it inside the E2E message envelope. The relay sees only the recipient DID and an encrypted delivery token (proving the sender is registered, without revealing identity). This is a significant privacy upgrade that breaks traffic analysis.

## In Scope

- Outer envelope structure (visible to relay): `{recipientDID, encryptedDeliveryToken, ciphertext}`
- `encryptedDeliveryToken`: proves sender is a registered ECHO user without revealing their DID; encrypted with recipient's identity key
- Inner envelope (decrypted by recipient only): `{senderDID, messageContent, commitmentHash, ECDSASignature}`
- iOS `SealedSenderService` in `Core/Relay/`: construct sealed envelopes on send, unwrap on receive
- Backend relay update: accept sealed sender messages; route by recipient DID only; cannot determine sender
- Delivery token generation: short-lived (5-minute) opaque token proving sender registration, signed by backend
- Backward compatibility: support both standard (Phase 1–2) and sealed sender formats via message type flag
- Migration: sealed sender enabled by default for all new messages once both client and server are updated

## Out of Scope

- Phase 4 federated relay (separate work order)
- Phase 4 optional P2P direct connection
- ZK proofs for identity claims (WO-183)

## Requirements

From the Primary Architecture and Secure Data Handling blueprint:

**Sealed Sender Implementation (Phase 3):**
```plaintext
Outer envelope (visible to relay):
  - Recipient DID
  - Encrypted delivery token (proves sender is registered, without revealing identity)
  - E2E ciphertext blob

Inner envelope (decrypted by recipient only):
  - Sender DID
  - Message content
  - Commitment hash
  - ECDSA signature
```

**Transport Privacy Roadmap:**
| Phase | Protection | Server Sees |
|---|---|---|
| 1–2 | Baseline TLS 1.3 | Sender DID, recipient DID, timestamp, blob size |
| 3 | Sealed sender | Recipient DID, timestamp, blob size (sender hidden) |
| 4 | Federated relay | Each operator sees only its fraction |

## Blueprints

- Primary Architecture and Secure Data Handling — Defines the sealed sender outer/inner envelope structure, delivery token mechanism, and Phase 3 transport privacy upgrade

---

### WO-227: Implement Phase 3 Client-Side Merkle Proof Verification

**Type:** Build

**Blueprint:** End-to-End Message Encryption and Commitment

## Summary

Implement Phase 3 trustless client-side Merkle proof verification in the iOS `AnchoringTracker`. When the backend sends an anchoring confirmation with Merkle proof siblings, the iOS client independently verifies the commitment is included in the on-chain Merkle root — without trusting the relay server. This removes the relay from the integrity verification trust chain.

## In Scope

- **`verifyMerkleProof(commitment:siblings:onChainRoot:) -> Bool`** in `Core/Relay/AnchoringTracker.swift` — walk the Merkle path by hashing commitment with each sibling, compare computed root to on-chain root
- **Receive Merkle proof in anchoring confirmation:** Backend WebSocket `confirmation` message (Phase 3+) includes `merkleProof: [Data]?` (sibling hashes, leaf to root); Phase 1–2 omits this field (trust backend confirmation)
- **Trust model transition:** Phase 1–2: display `.anchored` when backend confirms. Phase 3+: display `.anchored` only after `verifyMerkleProof` returns `true`; display `.anchorVerificationFailed` if false (integrity violation alert)
- **Backend Merkle proof endpoint:** `GET /v1/messages/{messageId}/merkle-proof` → returns `{commitment, siblings, snapshotHash, snapshotHeight}` for manual re-verification after the fact
- **Integrity violation alert:** If Phase 3 verification fails, show `UIAlertController` warning: "Message integrity could not be verified — the relay may have tampered with this message"
- **Commitment local storage:** Store each message's `commitmentNonce` in SwiftData alongside the message so the commitment can be recomputed for Phase 3 verification

## Out of Scope

- Merkle batch submission (WO-15, backend)
- Standard anchoring confirmation (Phase 1–2, already in WO-15 and `AnchoringTracker`)
- Sealed sender (WO-219)

## Requirements

From the End-to-End Message Encryption and Commitment blueprint:

**Phase 3 client-side Merkle proof verification:**
```swift
func verifyMerkleProof(
    commitment: Data,
    siblingHashes: [Data],
    onChainRoot: Data
) -> Bool {
    var computed = commitment
    for sibling in siblingHashes {
        computed = Data(SHA256.hash(data: computed + sibling))
    }
    return computed == onChainRoot
}
```

**Merkle proof structure:**
```swift
struct MerkleProof {
    let commitment: Data        // User's message commitment hash
    let siblings: [Data]        // Sibling hashes from leaf to root
    let snapshotHash: String    // On-chain snapshot hash
    let snapshotHeight: Int     // On-chain block height
}
```

## Blueprints

- End-to-End Message Encryption and Commitment — Defines Phase 3 client-side Merkle proof verification, `verifyMerkleProof` function, and trust model transition from server-confirmation to trustless verification

---

### WO-233: Execute Phase 3 App Store Security Gates and Open Source Release

**Type:** Build

**Blueprint:** Production Launch, Infrastructure, and Deployment

## Summary

Complete all four security gate requirements before App Store submission, prepare the Apple privacy manifest (`PrivacyInfo.xcprivacy`), and execute the Apache 2.0 open source publication of all ECHO code (iOS app, Go backend, Scala metagraph validation). This is the Phase 3 launch gate work order.

## In Scope

- **Security gate 1 — E2E encryption audit:** Engage third-party cryptographic security firm to review Kinnami (X25519 + ChaCha20-Poly1305) implementation, commitment hash design, and Secure Enclave integration. Resolve all critical/high findings before submission
- **Security gate 2 — Scala L1 code review:** Blockchain security firm audits all Currency L1 and Data L1 Scala validation logic for consensus-layer vulnerabilities
- **Security gate 3 — Go backend penetration test:** External pen tester runs OWASP-scope test against all 10 Go microservices and WebSocket relay; all critical/high findings resolved
- **Security gate 4 — DAG staking verification:** Confirm 750K+ DAG staked to 3 L0 nodes; verifiable on DAG Explorer before App Store submission
- **Apple privacy manifest (`PrivacyInfo.xcprivacy`):** Declare all API usage (Contacts permission, camera, biometrics, push notifications, local network); confirm zero third-party SDKs collecting PII; complete App Privacy Report review
- **App Store submission checklist:** Push notification entitlements (APNs), VIP in-app purchase entitlements (Phase 5 prep), Secure Enclave entitlements
- **Apache 2.0 open source publication:** Publish iOS app (Swift/SwiftUI), Go backend (10 microservices), Scala metagraph L1 validation to GitHub under Apache 2.0. Pre-publish review: remove production credentials, private keys, treasury multi-sig config, unpatched security vulnerabilities
- **Phase 3 beta rollout:** Closed alpha (100–500 users) → TestFlight public beta (1K–10K) → App Store soft launch (10K–100K)

## Out of Scope

- Phase 3 sealed sender implementation (WO-219)
- Phase 3 Midnight PoC (WO-212)
- Ongoing security monitoring post-launch (WO-210)

## Requirements

From the Production Launch blueprint:

**Security gates before App Store:**
| Gate | Requirement |
|---|---|
| E2E encryption audit | Third-party cryptographic review of Kinnami stack |
| Scala L1 code review | Blockchain security firm audit |
| Backend penetration test | Go backend + relay OWASP scope |
| DAG staking | 750K DAG verified on DAG Explorer |

**Open source:** Full codebase under Apache 2.0 at App Store launch. ECHO's value proposition is "no company owns your account" — open source is cryptographic proof of this.

## Blueprints

- Production Launch, Infrastructure, and Deployment — Defines Phase 3 security gate requirements, App Store submission prerequisites, beta rollout stages, and open source publication strategy

---

### WO-235: Build iOS ZKProofUseCase with Midnight SDK Integration

**Type:** Build

**Blueprint:** Zero-Knowledge Proofs and Midnight Integration

## Summary

Build the iOS `ZKProofUseCase` actor with Midnight SDK integration for on-device ZK proof generation across all 5 claim types (trust tier minimum, age verification, KYC compliance, group membership, balance threshold). Private inputs never leave the device — only proof bytes and public signals are transmitted to the backend for on-chain Midnight verification.

## In Scope

- **`ZKProofUseCase` actor** in `Domain/UseCases/Identity/` with Midnight iOS SDK (`MidnightClient`)
- **5 proof generation methods:**
  - `proveTrustTierMinimum(minimumTier: Int)` — reads local Cardano trust tier datum, generates Compact circuit proof
  - `proveAgeThreshold(ageThreshold: Int)` — reads birthdate from local credential cache, proves age ≥ threshold
  - `proveKYCCompliance()` — proves KYC passed from approved issuer (Stripe Identity or Sumsub DID in approved set)
  - `proveGroupMembership(groupId: String)` — proves membership without revealing all groups
  - `proveBalanceThreshold(minimumECHO: Decimal)` — proves ECHO balance ≥ threshold without revealing exact amount
- **`ZKProof` struct:** `proofBytes: Data`, `publicSignals: [String: Any]`, `claimType: ZKClaimType`, `generatedAt: Date`; private inputs NOT stored
- **`ZKClaimType` enum:** `.trustTierMinimum`, `.ageVerification`, `.kycCompliance`, `.groupMembership`, `.balanceThreshold`
- **Proof submission:** `submitAndVerify(_ proof: ZKProof)` → `POST /zk/verify/:claimType` backend endpoint → returns `Bool`
- **Target performance:** < 5 seconds proof generation on iPhone 12 or newer
- **Nonce/replay prevention:** Backend provides challenge nonce per proof request; included in proof public signals

## Out of Scope

- Go backend ZK verification service (separate work order)
- Midnight contract deployment (WO-212)
- Phase 3 PoC (WO-212 covers evaluation/testnet)

## Requirements

From the Zero-Knowledge Proofs and Midnight Integration blueprint:

**`ZKProofUseCase`:**
```swift
actor ZKProofUseCase {
    private let midnightSDK: MidnightClient

    func proveTrustTierMinimum(minimumTier: Int) async throws -> ZKProof {
        let trustDatum = try await cardanoIdentity.getTrustTierDatum()
        let proof = try await midnightSDK.generateProof(
            circuit: "TrustTierVerifier",
            privateInputs: ["userDID": currentDID, "minTier": minimumTier],
            publicSignals: ["minimumTier": minimumTier]
        )
        return ZKProof(proofBytes: proof.bytes, publicSignals: proof.publicSignals, ...)
    }
}
```

Private inputs (score, birthdate, balance) are **never stored** — used only during generation.

## Blueprints

- Zero-Knowledge Proofs and Midnight Integration — Defines all 5 ZK claim types, iOS `ZKProofUseCase` implementation, `ZKProof` struct, Midnight SDK integration, and proof submission to backend

---

### WO-236: Implement Go Backend ZK Verification Service with Midnight and Graceful Degradation

**Type:** Build

**Blueprint:** Zero-Knowledge Proofs and Midnight Integration

## Summary

Build the Go backend ZK verification service in the Trust Service (port 8003) that forwards ZK proof bytes to Midnight for on-chain verification, caches boolean results with configurable TTL, and gracefully degrades to hash-commitment verification when Midnight is unavailable. All 5 claim types are supported.

## In Scope

- **5 verification endpoints:** `POST /zk/verify/trust-tier`, `/zk/verify/age`, `/zk/verify/credential`, `/zk/verify/group-membership`, `/zk/verify/balance-threshold`
- **Midnight client (`MidnightClient`):** Forward `{CircuitName, ProofBytes, PublicSignals, SubjectDID}` to Midnight node for on-chain verification; receive boolean `Valid` result
- **Result caching:** Cache boolean result in Redis: `key = "zk:{subjectDID}:{claimType}:{threshold}", value = {valid, verifiedAt, expiresAt}`, configurable TTL (default: 5 minutes for trust tier, 1 hour for age/KYC)
- **Graceful degradation:** If Midnight is unavailable or times out (> 10s), fall back to hash-commitment verification: retrieve `H(score || nonce)` from Cardano, compare against expected commitment for the claimed tier. Log degradation event as circuit-breaker metric
- **Nonce/replay prevention:** Each verification request requires a challenge nonce from `GET /zk/challenge`; backend includes nonce in Midnight verification call; used nonces rejected within 5-minute window
- **Backend never caches underlying credentials:** Only the boolean `valid` result is cached; score, birthdate, credential content are never accessible to backend

## Out of Scope

- iOS proof generation (separate work order)
- Midnight contract deployment (WO-212)
- ZK proof generation infrastructure (WO-212 handles PoC and contract deployment)

## Requirements

From the Zero-Knowledge Proofs and Midnight Integration blueprint:

**Go verification flow:**
```go
func (s *TrustService) VerifyZKProof(ctx context.Context, req ZKVerifyRequest) (*ZKVerifyResult, error) {
    midnightResult, err := s.midnightClient.VerifyProof(ctx, MidnightVerifyRequest{
        CircuitName:   req.ClaimType,
        ProofBytes:    req.ProofBytes,
        PublicSignals: req.PublicSignals,
        SubjectDID:    req.SubjectDID,
    })
    if err != nil {
        return s.fallbackHashVerification(ctx, req)  // Graceful degradation
    }
    // Cache boolean result — NEVER the underlying credential data
    cacheKey := fmt.Sprintf("zk:%s:%s:%s", req.SubjectDID, req.ClaimType, req.Threshold)
    s.cache.Set(cacheKey, midnightResult.Valid, s.config.ZKCacheTTL)
    return &ZKVerifyResult{Valid: midnightResult.Valid, ...}, nil
}
```

## Blueprints

- Zero-Knowledge Proofs and Midnight Integration — Defines the Go backend ZK verification flow, Midnight client integration, boolean caching, graceful degradation pattern, and nonce/replay prevention

---

### WO-248: Build Data Sovereignty Layer iOS Client and Opt-In Settings

**Type:** Build

**Blueprint:** Data Sovereignty Layer

## Summary

Build the iOS Data Sovereignty Layer client — the on-device behavioral statistics computation pipeline, Midnight ZK anonymization proof generation, opt-in settings, and contribution submission. Users who opt in earn ECHO rewards proportional to the value their anonymized data generates. Message content is **never** included; only metadata patterns (message counts, response latency distributions, feature usage booleans).

## In Scope

- **`DataSovereigntySettings` UI** in Settings → Privacy → Data Contribution:
  - Master opt-in toggle (default: OFF — requires explicit consent)
  - Per-category toggles: communication patterns, topic frequency, trust network patterns, feature usage
  - `AnonymizationLevel` selector: Standard (ε=1.0) or Enhanced (ε=0.1, lower payment, stronger privacy)
  - Minimum payment threshold configuration
  - Contribution history and earnings display
- **On-device statistics computation** from local SwiftData — never transmit raw data:
  - Messages per day (count only), response latency distribution (histogram buckets, no timestamps)
  - Keyword category frequencies (hashed with per-user salt — not the actual keywords)
  - Trust tier distribution of contacts (count breakdown only, no contact identities)
  - Feature usage boolean flags
- **Differential privacy noise** applied on-device before any submission (using Laplace/Gaussian mechanism for each statistic)
- **Midnight ZK proof generation** via `ZKProofUseCase`: `proveAnonymization(did:)` — generates proof that the data package cannot be linked to the contributor's DID; private inputs (DID) never transmitted
- **Contribution submission:** `POST /datasov/contribute` with `{dataPackage, zkProof, anonymizationLevel}` — server verifies ZK proof before accepting
- **Earnings display:** show pending contribution earnings, historical payouts, current contribution weight

## Out of Scope

- Backend Data Sovereignty Service (WO-240)
- Data buyer query API (WO-240)
- Midnight ZK infrastructure (WO-212, WO-235)
- Fee distribution (WO-240)

## Requirements

From the Data Sovereignty Layer foundation blueprint:

**REQ-DSL-001:** All behavioral statistics shall be computed on the user's device from local data. Raw data shall never be transmitted.

**REQ-DSL-002:** Every contribution shall include a Midnight ZK proof demonstrating the data cannot be linked to the contributor's DID.

**REQ-DSL-005:** Data contribution shall default to OFF. Users must explicitly opt in via Settings → Privacy → Data Contribution.

**REQ-DSL-006:** Users shall be able to revoke opt-in at any time. Future contributions stop immediately upon revocation.

**Data Classification:** Only T7 (public) pattern metadata allowed in contributions. Any T0–T4 data (message content, DIDs, contact lists) must be rejected by the on-device computation layer before packaging.

```swift
struct DataSovereigntySettings {
    var isOptedIn: Bool = false
    var contributionCategories: Set<DataCategory>
    var minimumPaymentThreshold: Decimal
    var anonymizationLevel: AnonymizationLevel
}

enum AnonymizationLevel {
    case standard   // Differential privacy ε = 1.0
    case enhanced   // Differential privacy ε = 0.1
}
```

## Blueprints

- Data Sovereignty Layer — Defines on-device computation architecture, ZK proof requirement, opt-in controls, differential privacy levels, data category constraints, and T0–T4 prohibition

---

### WO-249: Implement Data Sovereignty Service Backend with Query API and Fee Distribution

**Type:** Build

**Blueprint:** Data Sovereignty Layer

## Summary

Implement the Go backend Data Sovereignty Service that receives anonymized contributions, verifies ZK anonymization proofs via Midnight, applies server-side differential privacy, aggregates data into the community pool, executes data buyer queries with privacy budget enforcement, and distributes query fees (70% contributors / 30% Privacy Commons Treasury) monthly.

## In Scope

- **Contribution processing endpoint:** `POST /datasov/contribute` — verifies ZK proof via Midnight, applies server-side differential privacy noise, aggregates into community pool (no individual extraction possible after aggregation), records `{contributorDIDHash, timestamp, dataCategory, weight}` in PostgreSQL
- **Data buyer query API:** `POST /datasov/query` — accepts `{queryType, dataCategories, timeRange, minSampleSize, maxQueryDepth}` with API key; enforces minimum 10,000 contributor sample size; tracks and enforces differential privacy epsilon budget per buyer
- **Privacy budget accounting:** track epsilon consumed per dataset per buyer; block queries when budget exhausted
- **Monthly fee distribution:** calculate each contributor's share proportional to data weight; distribute 70% to contributors via Currency L1 `SpendTransaction`; route 30% to Privacy Commons Treasury wallet; enforce minimum 10 ECHO payout threshold
- **Aggregate pool management:** immutable once aggregated — individual contributions cannot be extracted; zstd-compressed, AES-256-GCM encrypted in PostgreSQL
- **Contributor hash index:** `sha256(contributorDID + salt)` — cannot be reversed; salt rotated monthly

## Out of Scope

- iOS client (WO-239)
- Midnight ZK infrastructure (WO-212, WO-236)
- Privacy Commons Treasury management (WO-253)

## Requirements

From the Data Sovereignty Layer foundation blueprint:

**REQ-DSL-003:** Server-side differential privacy noise shall be applied to all contributions and queries. Privacy budget epsilon shall be tracked per dataset.

**REQ-DSL-004:** No query result shall contain statistics derived from fewer than 10,000 contributors.

**REQ-DSL-007:** 70% of query fees shall be distributed to contributing users proportional to data weight. 30% shall flow to Privacy Commons Treasury. Distribution at least monthly.

**REQ-DSL-008:** Message content, sender/recipient DIDs, and contact identities shall be prohibited in any contribution. The Data Sovereignty Service shall validate all contributions against T0–T4 classification rules.

```go
type DataSovereigntyService struct {
    midnightClient MidnightClient
    privacyEngine  DiffPrivEngine   // Differential privacy noise
    queryEngine    AggregateEngine  // Aggregate query execution
    feeDistributor FeeDistributor   // 70/30 payment distribution
}
```

## Blueprints

- Data Sovereignty Layer — Defines full backend service architecture, contribution processing pipeline, query API with privacy budget enforcement, fee distribution formula (70/30), and data classification prohibitions

---

### WO-255: Implement Portable Identity Export API and Protocol Developer API

**Type:** Build

**Blueprint:** Portable Social Graph and Protocol Layer

## Summary

Implement the portable identity export API and the Protocol Developer API that allows any third-party application to verify ECHO users' trust tiers and credentials without ECHO's involvement. The DID and trust tier are anchored on Cardano and verifiable by anyone. This is the foundation of the ECHO Protocol network effect.

## In Scope

- **Portable identity export:** `GET /identity/export` — returns a complete portable identity package:
  ```json
  {
    "did": "did:prism:cardano:abc123",
    "didDocument": { ... },
    "credentials": [{ "type": "TrustTierAttestation", "tier": 4, "cardanoTxHash": "...", "verificationURL": "..." }],
    "trustAttestation": { "tier": 4, "cardanoRef": "utxo:abc123#0", "verifiableByAnyone": true }
  }
  ```
  Export includes everything needed to prove identity and trust tier in any DID-compatible app without an ECHO account
- **Protocol Developer API (rate-limited):**
  - `GET /protocol/identity/resolve/:did` — DID Document (public Cardano data); 100 requests/minute unauthenticated
  - `GET /protocol/identity/tier/:did` — trust tier + verification URL; requires developer API key
  - `POST /protocol/identity/verify` — verify a credential presentation; 10/minute per API key
  - `GET /protocol/contacts/mutual/:did1/:did2` — boolean mutual connection check (privacy-preserving); returns `{mutualConnection: true/false}` without revealing contact lists
- **Developer API key management:** Foundation governance grant issues developer API keys; `POST /protocol/keys/request` for developers
- **iOS identity export UI:** Settings → Account → Export Portable Identity → shows export package summary + "Download" button + explanation of portability

## Out of Scope

- DID-preserving account deletion (WO-256)
- Trust tier computation (WO-181)
- Cardano DID anchoring (WO-37, WO-180)

## Requirements

**REQ-GRAPH-001:** DID shall be resolvable by any W3C DID-compatible resolver without ECHO's involvement.
**REQ-GRAPH-002:** Any third party shall be able to verify a user's trust tier by querying Cardano directly.
**REQ-GRAPH-003:** Users shall be able to export complete portable identity via `GET /identity/export` at any time.
**REQ-GRAPH-005:** Developers who register for a Foundation API key shall be able to resolve DIDs, verify trust tiers, and check credential validity at documented rate limits.
**REQ-GRAPH-006:** Portable social graph shall be designed so users can migrate to a competing application with zero data loss and zero ECHO cooperation required.

## Blueprints

- Portable Social Graph and Protocol Layer — Defines portable identity export format, Protocol Developer API endpoints, rate limits, and zero-lock-in design principles

---

### WO-256: Implement DID-Preserving Account Deletion with Identity Metagraph Deactivation

**Type:** Build

**Blueprint:** Portable Social Graph and Protocol Layer

## Summary

Implement DID-preserving account deletion for `did:key` — when a user deletes their ECHO account, all active credentials and trust tier records are **revoked** on the Constellation Identity Metagraph, but the `did:key` DID itself (a public key fingerprint) remains mathematically valid. The user loses platform access, but the DID string is portable: any DID-compatible application can still resolve the public key from it. If the user retains their recovery phrase (WO-234), they can re-register the same DID on another compatible platform.

## In Scope

- **Account deletion flow (iOS):**
  1. Settings → Delete Account → confirm identity (biometric)
  2. Choose token disposition (burn or transfer to wallet address)
  3. Multi-step confirmation: "Your account will be deleted. All credentials will be revoked on the Identity Metagraph, but your `did:key` identifier remains valid — you can use your recovery phrase to re-register on any compatible platform."
  4. Progress screen: local key deletion → queue wipe → Identity Metagraph credential revocation → token disposal
- **Identity Metagraph deactivation:** submit StatusList2021 revocation for all active credentials linked to this DID (sets all relevant bit positions to `1`); submit trust tier commitment expiry/revocation to Identity Metagraph
- **Account record deletion:** remove DID-to-account mapping from PostgreSQL, wipe offline message queue (Redis/PostgreSQL), revoke auth tokens
- **Local key deletion:** delete Secure Enclave keys and Keychain data — without the private key, the DID can no longer sign requests, ending platform access
- **Portability preservation:** the `did:key` DID string encodes the public key — it remains resolvable by any W3C DID-compatible resolver; if the user has their recovery phrase, they can derive the same key pair and re-register on a compatible application
- **Backend endpoint:** `DELETE /v1/account` with `{tokenDisposition, biometricSignature}` — orchestrates the full deletion sequence

## Out of Scope

- GDPR cryptographic erasure (WO-218 — covers key destruction for privacy compliance)
- Portable identity export (WO-255)
- Recovery phrase generation (WO-234)

## Requirements

**REQ-GRAPH-004:** When a user deletes their ECHO account, their DID remains valid and verifiable. The DID is deactivated (all platform records revoked on the Identity Metagraph) but not destroyed — the user retains the ability to recover their identity on another platform.

**Design note for `did:key`:** Unlike `did:prism:cardano:` (which had an on-chain UTXO record that could be set to "deactivated"), a `did:key` DID is a mathematical transformation of the public key with no on-chain record. "Deactivation" on ECHO means revoking all Identity Metagraph records (credentials + trust tier). The DID string remains resoluble by any DID resolver globally — it is not erased from any ledger because it was never written to one.

**Deactivated ≠ Destroyed:**
- **Deactivated:** All ECHO Identity Metagraph credentials revoked; trust tier commitment expired; platform account deleted; device keys deleted
- **Portable:** `did:key:z6Mk...` still resolves to the public key; recovery phrase allows re-registration on another platform
- **Destroyed:** Secure Enclave private key deleted — device-level signing gone; only reversible via recovery phrase

## Blueprints

- Portable Social Graph and Protocol Layer — Defines REQ-GRAPH-004 (DID deactivation on account deletion, not destruction) and the portability preservation requirement

---

### WO-257: Implement Post-Quantum Cryptography Mode iOS Client (Kyber-768 + Dilithium3)

**Type:** Build

**Blueprint:** Post-Quantum Cryptography Mode

## Summary

Implement the Post-Quantum Cryptography Mode iOS client — Kyber-768 key generation, Dilithium3 signing, hybrid X25519+Kyber key agreement, Keychain storage for oversized PQ keys (too large for Secure Enclave), Settings toggle, and DID Document update with new PQ public keys. PQ Mode is additive: hybrid X25519+Kyber protects against both quantum and classical attacks.

## In Scope

- **`PostQuantumKeyPair` struct:** `kyberPublicKey` (1,184 bytes), `kyberPrivateKey`, `dilithiumPublicKey` (1,952 bytes), `dilithiumPrivateKey` — all stored in iOS Keychain (not Secure Enclave — too large); `kSecAttrAccessibleWhenUnlockedThisDeviceOnly`
- **`PQEncryptionService` actor:** hybrid key agreement combining X25519 ECDH (classical) + Kyber-768 encapsulation (post-quantum); `HKDF-SHA3-256(ss1 || ss2 || context)` for final key derivation; Kyber ciphertext included in message envelope
- **Dual signing:** every outbound message signed with ECDSA P-256 (Secure Enclave) + Dilithium3; recipients verify either signature independently
- **Settings → Security → Post-Quantum Mode:** opt-in toggle; activation requires biometric; clearly explains the trade-off (1.15KB message size increase, ~8ms latency)
- **DID Document update on activation:** `POST /identity/pq/activate` triggers Cardano DID Document update adding `#key-kyber` (Kyber768EncryptionKey2024) and `#key-dil3` (Dilithium3VerificationKey2024) + `pqMode: true, pqActivatedAt: timestamp`; completes within 30 seconds
- **Backward compatibility:** check recipient's DID Document for Kyber public key before sending PQ message; fall back to standard X25519 if recipient doesn't support PQ Mode
- **Kyber-768 library:** use NIST FIPS 203 (CRYSTALS-Kyber) implementation via Swift package dependency

## Out of Scope

- Backend PQ routing (WO-258)
- Organizational PQ policy enforcement (WO-259)
- Secure Enclave key management for P-256 (WO-223 — unchanged)

## Requirements

**REQ-PQ-001:** PQ Mode activatable via Settings → Security → Post-Quantum Mode; requires biometric.
**REQ-PQ-002:** Activating PQ Mode triggers Cardano DID Document update with Kyber-768 and Dilithium3 public keys within 30 seconds.
**REQ-PQ-003:** PQ messages use hybrid X25519 + Kyber-768 scheme. Neither leg alone sufficient to decrypt.
**REQ-PQ-004:** PQ Mode users can still send/receive standard messages with non-PQ users (backward compatible).
**REQ-PQ-006:** PQ Mode adds < 10ms to message send latency on iPhone 12 or newer.

**Algorithm selection rationale:** Kyber-768 = NIST FIPS 203 Level 3 (quantum security equivalent to AES-192); Dilithium3 = NIST FIPS 204 Level 3. Hybrid approach provides security against both quantum and classical attacks.

## Blueprints

- Post-Quantum Cryptography Mode — Defines full hybrid key agreement protocol, iOS implementation with `PQEncryptionService`, key storage requirements, DID Document extension, backward compatibility, and performance targets

---

### WO-258: Implement PQ Mode Backend Routing, DID Document Extension, and Dual Signature Verification

**Type:** Build

**Blueprint:** Post-Quantum Cryptography Mode

## Summary

Implement the backend PQ Mode routing and DID Document extension support — the Message Relay detects PQ capability from sender/recipient DID Documents and routes accordingly; PQ-encrypted envelopes include the Kyber ciphertext; DID Document update endpoint adds Kyber and Dilithium public keys to Cardano; standard-mode fallback for non-PQ recipients.

## In Scope

- **DID Document PQ extension:** `POST /identity/pq/activate` — update user's Cardano DID Document to add:
  ```json
  { "id": "#key-kyber", "type": "Kyber768EncryptionKey2024", "publicKeyHex": "..." }
  { "id": "#key-dil3",  "type": "Dilithium3VerificationKey2024", "publicKeyHex": "..." }
  ```
  Plus `pqMode: true, pqActivatedAt: timestamp` in DID Document metadata
- **PQ capability detection:** backend caches DID Document PQ capability (Redis, TTL: 60s); `hasPQMode(did)` returns `true` if `#key-kyber` present in DID Document
- **PQ envelope routing:** Message Relay extracts sender's PQ key capability from cached DID Document; if both sender and recipient support PQ, route PQ-enhanced envelope (includes Kyber ciphertext field); if recipient doesn't support PQ, sender falls back to standard mode
- **PQ envelope schema extension:** `EncryptedPayload` extended with optional `kyberCiphertext: Data?` field; schema version 2 for PQ-enabled messages; relay stores `isPQMode: bool` in queue metadata for PQ-capable offline recipients
- **Dual signature verification:** backend validates either P-256 (Secure Enclave) or Dilithium3 signature as sufficient; dual-signed messages log both verifications
- **Backward compatibility:** messages without `kyberCiphertext` field use standard mode; PQ-mode backend gracefully handles mixed PQ/non-PQ conversations

## Out of Scope

- iOS PQ client (WO-257)
- Organizational policy enforcement (WO-259)

## Requirements

**REQ-PQ-002:** DID Document update with Kyber/Dilithium keys within 30 seconds.
**REQ-PQ-003:** Neither key agreement leg alone sufficient to decrypt PQ messages.
**REQ-PQ-004:** Standard-mode users continue to work normally alongside PQ-mode users.

**Envelope schema (version 2):**
```go
type PQEncryptedPayload struct {
    SenderEphemeralX25519PublicKey []byte  // Standard
    KyberCiphertext                []byte  // PQ addition: 1,088 bytes
    Ciphertext                     []byte  // ChaCha20-Poly1305 (unchanged)
    Commitment                     []byte  // SHA-256 (unchanged)
    SchemaVersion                  int     // 2 for PQ messages
}
```

## Blueprints

- Post-Quantum Cryptography Mode — Defines DID Document extension with Kyber/Dilithium keys, hybrid routing, backward compatibility, and PQ envelope schema

---

### WO-259: Implement PQ Mode Organizational Policy Enforcement for ECHO Comply

**Type:** Build

**Blueprint:** ECHO Comply — Enterprise Compliance Messaging, Post-Quantum Cryptography Mode

## Summary

Implement PQ Mode organizational policy enforcement for ECHO Comply — administrators can mandate Post-Quantum encryption for all users in their organization. The policy is anchored to Data L1 for audit trail. Non-PQ messages from PQ-enforced org users are rejected by the backend (HTTP 403). This is particularly critical for healthcare (7+ year record retention under quantum attack window) and legal (eDiscovery holds).

## In Scope

- **`CompliancePolicy.requirePQMode` field:** organization administrators enable PQ Mode enforcement via `POST /comply/policy/pq-enforcement` with `{orgDID, requirePQMode: true, effectiveAt}`
- **Data L1 policy anchor:** PQ enforcement activation anchors `{pq_enforcement: active, orgDID, effectiveAt}` to Data L1 for audit trail
- **Backend enforcement:** Message Relay rejects standard-mode messages from PQ-enforced org users:
  ```
  HTTP 403 Forbidden
  { "error": "pq_mode_required", "message": "Your organization requires Post-Quantum mode. Please enable it in Settings → Security → Post-Quantum Mode." }
  ```
- **Grace period:** 30-day grace period after policy activation where non-PQ messages are **warned** but not rejected; enforcement begins after grace period
- **iOS compliance indicator:** compliance admin dashboard shows PQ adoption rate per organization (% of users with PQ Mode active)
- **User activation prompt:** when non-PQ user sends a message in a PQ-enforced org, iOS shows prompt: "Your organization requires Post-Quantum mode. Tap to enable in Settings."
- **PQ adoption monitoring:** `GET /comply/policy/pq-adoption` returns `{orgDID, totalUsers, pqEnabledUsers, pqAdoptionRate, gracePeriodEndsAt}`

## Out of Scope

- iOS PQ client (WO-257)
- Backend PQ routing (WO-258)
- Non-compliance PQ enforcement for consumer users (consumer PQ is opt-in only)

## Requirements

**REQ-PQ-005:** ECHO Comply administrators shall be able to enforce PQ Mode for all organizational users. Enforcement anchors a compliance policy on Data L1.

**Healthcare justification:** Healthcare records retained 7+ years are already within the quantum computing attack window projected by CISA. HIPAA covered entities must be PQ-ready.

**Legal justification:** Communications under eDiscovery holds may remain relevant for decades — well within post-quantum attack horizon.

## Blueprints

- Post-Quantum Cryptography Mode — Defines REQ-PQ-005 organizational enforcement, backend rejection behavior, and Data L1 policy anchoring
- ECHO Comply — Enterprise Compliance Messaging — Defines organizational policy enforcement infrastructure and compliance dashboard

---

### WO-260: Implement Privacy Commons Treasury On-Chain Management and Disbursement Tracking

**Type:** Build

**Blueprint:** Privacy Commons Treasury

## Summary

Implement the Privacy Commons Treasury on-chain management system — tracking all inflows and outflows with `privacy_commons_disbursement` events on Data L1, managing the three program pools (Legal Defense, Journalist Access, Research Grants), integrating with the AI Treasury CFO Agent for minimum reserve monitoring, and providing public balance transparency on DAG Explorer.

## In Scope

- **On-chain treasury structure:** Privacy Commons Treasury DID with three `TokenLock` sub-pools:
  - `legal_defense_reserve`: emergency legal funding
  - `journalist_access_pool`: subsidized ECHO Comply access
  - `research_grants_pool`: open-source privacy research grants
  - `operating_reserve`: minimum 3-month program funding buffer
- **Data L1 disbursement events:** every inflow and outflow anchors `{privacy_commons_disbursement, category: "legal_defense"|"journalist_access"|"research_grant"|"operating_reserve", amount, approvalMethod, timestamp}` — recipient identity NEVER included
- **Fee routing from Data Sovereignty Layer:** auto-receive 30% of all Data Sovereignty query fees (WO-240); route to appropriate sub-pool based on governance-set allocation
- **Revenue share routing:** receive governance-set % (starting 5%) of annual platform surplus from main treasury; `SpendTransaction` from main treasury to Privacy Commons DID
- **Community donation support:** Privacy Commons Treasury DID is published; any ECHO holder can donate via `SpendTransaction` to treasury DID
- **CFO AI Agent integration:** monitor minimum 3-month reserve level; flag to governance notification channel if reserve falls below threshold (per REQ-PCT-005)
- **Public DAG Explorer transparency:** all balances, inflows, and outflows publicly visible; Treasury DID published in Foundation governance documents
- **Governance enforcement:** disbursements under $10K require 3-of-5 Community Board multi-sig; over $50K require full governance vote; implemented via multi-sig `AtomicAction` bundles

## Out of Scope

- Anonymous journalist access credential issuance (WO-261)
- Legal defense case management (legal processes, not code)
- Research grant selection process (governance, not code)

## Requirements

**REQ-PCT-001:** All Privacy Commons Treasury inflows and outflows recorded on Data L1 with category labels. Individual recipient identities never recorded.
**REQ-PCT-003:** Emergency legal defense cases receive initial funding decision within 24 hours of verified application.
**REQ-PCT-005:** Community governance maintains minimum 3-month funding reserve. AI Treasury CFO Agent flags if below threshold.

## Blueprints

- Privacy Commons Treasury — Defines funding sources (30% DSL fees, 5%+ platform revenue, donations), program pool structure, on-chain transparency requirements, disbursement governance thresholds, and minimum reserve policy

---

### WO-261: Build Anonymous Journalist and Activist Access Credential System

**Type:** Build

**Blueprint:** Privacy Commons Treasury

## Summary

Build the anonymous journalist and activist access credential issuance system — the Privacy Commons Treasury funds ECHO Comply Professional access for journalists and activists in surveillance-risk environments, but the recipient's identity is **never stored** in any ECHO system. Access is granted via anonymous Cardano credentials, funded by the treasury paying the DID activation cost without knowing who the DID belongs to.

## In Scope

- **Anonymous access application flow:** qualified recipients (verified by CPJ, RSF, EFF, or equivalent press freedom org) apply via anonymous application form; ECHO receives only: organization voucher reference, geography category (country withheld), mission category
- **Anonymous Cardano credential issuance:** Privacy Commons Treasury pays DID activation cost on Cardano on behalf of anonymous recipient; issued `JournalistAccessCredential` with `{type: "JournalistAccess", issuerDID: "did:prism:cardano:privacy-commons-treasury", tier: "professional", expiresAt}` — recipient's DID is the credential subject but is never stored by ECHO
- **ECHO Comply Professional access:** upon credential issuance, backend maps the anonymous DID to Professional tier without storing the holder's identity; access includes Duress PIN, Hidden Folders, Post-Quantum Mode, full compliance features
- **Press freedom org integration:** integration with CPJ, RSF, EFF API (or manual voucher upload) to verify eligibility without ECHO performing verification directly; `POST /comply/journalist-access/vouch` with `{voucherReference, issuingOrgDID}`
- **Access renewal:** credentials expire after 1 year; renewal flow through same anonymous process; no identity memory between renewal cycles
- **Treasury payment:** `SpendTransaction` from journalist_access_pool to Cardano DID activation fee; transparent on DAG Explorer as treasury outflow but no recipient linkage

## Out of Scope

- Privacy Commons Treasury pool management (WO-260)
- Legal defense case management
- Research grant disbursement

## Requirements

**REQ-PCT-002:** Journalist and activist access grants shall be issued via anonymous Cardano credentials. Treasury funds DID activation costs without storing recipient identity in any ECHO system.

**Access level:** Full ECHO Comply Professional tier: Duress PIN, Hidden Folders, Post-Quantum Mode, and all compliance features — at no cost to recipient.

**Privacy guarantee:** ECHO cannot identify which journalist or activist is using a given DID. The vouching organization only confirms eligibility; they do not share identity with ECHO.

## Blueprints

- Privacy Commons Treasury — Defines anonymous journalist/activist access credential issuance, eligibility criteria, press freedom org vouching process, anonymous Cardano credential structure, and REQ-PCT-002 identity non-storage requirement

---

---

## Competitive Audit Additions (2026-05-26)

Proposed from `docs/COMPETITIVE_AUDIT_2026-05.md` (Tier 1). Provisional IDs — final WO numbers assigned by Software Factory.

### WO-CA2: Consumer secure encrypted backups
**Source:** competitive audit (Signal Secure Backups). **Extends:** WO-64; reuses BIP-39 recovery (WO-234).
User-held recovery key (the 24-word phrase) encrypts a cloud backup of chats + opt-in media tiers, with cross-platform restore. The server never holds the key.

---

## Echo Passport — Verifier, External Rails & More Credentials — Wave C (4)

> New product. Full plan: `docs/ECHO_PASSPORT_PLAN.md`. All Backlog. Makes the Passport useful
> *outside* Echo and turns the verifier side into revenue. **Holder side stays feeless**;
> monetize the verifier/enterprise side and external payment rails (existing 0.5–1.5% rail fee).

### WO-301: Echo Passport — Verifier / Relying-Party API + ECHO Metering (Wave C)

**Status:** 📋 Backlog · **Depends:** WO-295 (selective disclosure), WO-118 (trusted issuers)
**Blueprint:** Decentralized Identity and Authentication

## Summary

Let a relying party request a presentation and pay per verification in ECHO. Verifier fee splits
70% holder / 30% Privacy Commons Treasury (mirrors the existing Data-Sovereignty split).

## In Scope

- `pkg/passport/verifier/`: relying-party API (`POST /v1/verify/request`), per-presentation metering + fee split.
- Reuse OIDC4VCI verifier flow + StatusList2021 revocation; verifier identity via `trusted_issuers`.

## Out of Scope

- Verifier marketplace UI/discovery (WO-305, Phase 7). External payment rails (WO-302).

**Acceptance Criteria:**
- A verifier receives only disclosed claims; fee split lands 70/30 on-chain.
- Holder consents per presentation; no silent disclosure.

### WO-302: Echo Passport — External Payment-Rail Adapter (Wave C)

**Status:** 📋 Backlog · **Depends:** WO-299 (consent), licensed rail partner
**Blueprint:** ECHO Tokenomics, Founder Allocation, and Token Launch

## Summary

Merchant payments via a licensed rail (card-network tokenization / Open Banking PISP partner).
Echo orchestrates consent + presentation; the partner moves the money. Echo never holds a raw PAN.

## In Scope

- `pkg/payments/rails/`: adapter interface + one reference adapter.
- Tokenized payment-instrument references (network token / Open Banking mandate id) held encrypted in Layer 2; the VC attests "verified Visa ending 1234", never the PAN.
- Apply the existing 0.5–1.5% rail fee → community treasury.

## Out of Scope

- Echo acting as money transmitter (out of scope by design — partner-routed). P2P native (WO-299).

**Acceptance Criteria:**
- No raw PAN/account number in any VC, blob, or log (Semgrep fixtures that *should* trip).
- Merchant payment completes via partner with per-action biometric consent.
- **Open item:** confirm licensed partner + jurisdiction (US-first vs EU/eIDAS) before build.

### WO-303: Echo Passport — Recovery Hardening (Wave C)

**Status:** 📋 Backlog · **Depends:** WO-296 (Shamir recovery)
**Blueprint:** Universal Onboarding and Identity Creation

## Summary

Harden recovery for higher-value credentials: guardian bonds, audited Shamir implementation, and
account-abstraction-style device rotation.

## In Scope

- Optional guardian stake/bond (anti-collusion); device add/rotate/revoke flows.
- Third-party audit hooks for the Shamir implementation.

## Out of Scope

- Base recovery (WO-296). Citizenship-tier recovery policy (Wave D).

**Acceptance Criteria:**
- Device rotation revokes old shares; quorum recompute verified.
- Still no server-held key under any hardened path.

### WO-304: Echo Passport — Legal-Document & Employment/Income Credentials (Wave C)

**Status:** 📋 Backlog · **Depends:** WO-293, WO-118 (trusted issuers)
**Blueprint:** Decentralized Identity and Authentication

## Summary

Add credential domains beyond identity: deeds/wills as **notarized-hash VCs** (issuer-attested
hash + reference, never the legal instrument), plus verified job/income/housing from accredited issuers.

## In Scope

- Credential type definitions + schemas for legal-document (hash) and employment/income VCs.
- Issuer onboarding via `trusted_issuers`; reuse issuance/verification pipeline.

## Out of Scope

- Echo issuing government ID or being the legal authority (always accredited third-party issuers).

**Acceptance Criteria:**
- Legal-doc VC stores only a hash + issuer attestation; raw document never on server/chain.
- Employment/income VCs verify against accredited issuer signatures.

---

# Echo Passport — x402 Payment Interop (ADR 0006)

> Added 2026-06-12 per `docs/CROSS_PRODUCT_GAP_REVIEW.md` and `docs/adr/0006-passport-x402-agentic.md`.
> Adopt **x402** (HTTP `402 Payment Required`) as the payment wire protocol for Passport's verifier and
> merchant surfaces, layered on the existing on-chain **AllowSpend / SpendTransaction** consent +
> settlement. x402 = how a charge is requested/proven over HTTP; AllowSpend = the single-use,
> biometric-gated authorization. No raw PAN/PII on the wire (tokenized instrument references only).

### WO-315: Echo Passport — x402 Facilitator + Verifier Pays-Per-Presentation (Wave C)

**Status:** 📋 Backlog · **Depends:** WO-301/302 (verifier/rails), AllowSpend primitives, WO-295 (disclosure)
**Blueprint:** Decentralized Identity and Authentication

## Summary

Expose the verifier-pays-per-presentation marketplace over **x402**: a relying party requests a
presentation, receives `402 Payment Required` (CAIP-2 network, price, accepted token), and pays in
ECHO/stablecoin; the payment is backed by a single-use `AllowSpend` and settled as a
`SpendTransaction` on Currency L1. Extends `pkg/passport/verifier/` and `/v1/verify/request`.

## In Scope

- x402 facilitator: emit/validate `402` payment requirements; verify x402 payment proofs.
- Bridge x402 ⇄ `AllowSpend`/`SpendTransaction` (replay-safe, single-use mapping).
- Per-presentation metering + fee split (70% holder / 30% Privacy Commons Treasury, per plan).

## Out of Scope

- Standing approvals; raw PAN/PII on the wire (tokenized references only).

**Acceptance Criteria:**
- A verifier completes an x402 round-trip and receives a selective-disclosure presentation.
- Payment settles as a single-use `SpendTransaction` on Currency L1; the fee split is applied.
- Replaying an x402 proof cannot double-spend the backing `AllowSpend`.

### WO-317: Echo Passport — x402 ↔ External Merchant Rail Adapter (Wave C/D)

**Status:** 📋 Backlog · **Depends:** WO-315, WO-302 (licensed rail / PISP partner)
**Blueprint:** Decentralized Identity and Authentication

## Summary

Adapter (`pkg/payments/rails/`) that bridges an x402 payment to a licensed external rail (card-network
tokenization / Open Banking PISP partner). Echo orchestrates consent and passes a presentation + a
**tokenized** instrument reference; the partner moves the money. Echo is never the money transmitter.

## In Scope

- Rail adapter interface + 1 reference adapter; consent → x402 → rail call mapping.
- Tokenized payment-instrument references (never a PAN), per ADR 0003.

## Out of Scope

- Echo holding funds or acting as money transmitter (licensing remains a deferred open item, WO-302).

**Acceptance Criteria:**
- A merchant payment completes via the licensed rail with only a tokenized instrument reference.
- No raw PAN/account number enters any VC, blob, chain payload, or x402 message (T0–T7 gate green).

---

## SimpleX Audit Additions (2026-05-29)

From [`docs/COMPETITIVE_AUDIT_SIMPLEX_2026-06.md`](COMPETITIVE_AUDIT_SIMPLEX_2026-06.md). Synced to Software Factory 2026-05-29.

### WO-317: SimpleX SX3 — Metadata minimization + sealed-sender default
**Status:** ✅ Completed · **Commit:** `e15e5e8`  
**Evidence:** `SealedSenderPolicy.swift`, `ConversationQueueAliasStore.swift`, `internal/api/ws.go` scrub

### WO-319: SimpleX SX4 — Optional Tor/SOCKS transport proxy
**Status:** ✅ Completed  
**Evidence:** `TransportProxySettings.swift`, `APIClient`, `WebSocketClient`, `TransportProxySettingsView` in `PrivacyHubView`  
**Remaining (future):** 2-hop routing; proxy on ancillary `URLSession.shared` clients

### WO-318: SimpleX SX6 — Per-contact minimal-disclosure persona
**Status:** ✅ Completed · **Commit:** `9fcc7af`  
**Evidence:** `ContactPersonaStore.swift`, `ContactDetailView`, `ContactThreadHelper`
