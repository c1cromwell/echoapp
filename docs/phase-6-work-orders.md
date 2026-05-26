# Phase 6: Calls & File Sharing

**Total Work Orders:** 31  
**Status Summary:** 31 Backlog  
**Last synced with Software Factory:** 2026-05-26

---

## Backlog (31)

### WO-5: Implement WebRTC Call Infrastructure with Noise Protocol Encryption

**Blueprint:** Voice and Video Calls with Screen Sharing

## Summary

Build the WebRTC call infrastructure for end-to-end encrypted voice and video calls. Establishes peer-to-peer connections with Noise Protocol encryption overlay, handles NAT traversal via STUN/TURN, falls back to relay when P2P fails, and supports up to 50 participants per call. Backend provides WebRTC signaling via the WebSocket relay.

## In Scope

- WebRTC `RTCPeerConnection` setup with STUN/TURN server configuration
- ICE candidate exchange through existing WebSocket relay (signaling channel)
- Noise Protocol encryption layer on top of WebRTC's DTLS-SRTP for additional security
- P2P direct connection when possible; automatic TURN relay fallback for restricted NAT
- Up to 50 participants using SFU (Selective Forwarding Unit) architecture for group calls
- Automatic quality adjustment: 1080p → 720p → 480p → audio-only based on network conditions
- Connection quality monitoring: RTT, packet loss, jitter sampled every second
- Graceful network interruption handling: retry with exponential backoff, reconnect mid-call
- Participant identity verification: validate caller's DID + trust tier before connecting

## Out of Scope

- Call UI controls (WO-19)
- Screen sharing (WO-31)
- Call recording (WO-43)
- Real-time transcription (WO-55)
- SFU server provisioning (DevOps work order)

## Requirements

Derived from the Voice and Video Calls blueprint.

**WebRTC Setup (iOS):**
```swift
// Core/Calling/WebRTCCallManager.swift
actor WebRTCCallManager {
    private var peerConnection: RTCPeerConnection?
    private let factory: RTCPeerConnectionFactory

    func initiateCall(to recipientDID: String, callType: CallType) async throws -> CallSession {
        // 1. Create peer connection with STUN/TURN config
        let config = RTCConfiguration()
        config.iceServers = [
            RTCIceServer(urlStrings: ["stun:stun.echo.app:3478"]),
            RTCIceServer(urlStrings: ["turn:turn.echo.app:3478"],
                         username: authToken, credential: authSecret)
        ]
        peerConnection = factory.peerConnection(with: config, constraints: nil, delegate: self)

        // 2. Add audio/video tracks
        let audioTrack = factory.audioTrack(withTrackId: "audio0")
        peerConnection?.add(audioTrack, streamIds: ["stream0"])
        if callType == .video {
            let videoTrack = factory.videoTrack(with: cameraSource, trackId: "video0")
            peerConnection?.add(videoTrack, streamIds: ["stream0"])
        }

        // 3. Exchange ICE candidates via WebSocket signaling
        let offer = try await peerConnection!.offer(for: RTCMediaConstraints(...)...)
        webSocket.send(WSMessage(type: .callOffer, payload: SignalingPayload(
            offer: offer.sdp, callerDID: myDID, recipientDID: recipientDID
        )))

        return CallSession(callId: UUID().uuidString, participants: [myDID, recipientDID])
    }
}
```

## Blueprints

- Voice and Video Calls with Screen Sharing — Defines WebRTC infrastructure, Noise Protocol encryption, P2P + relay fallback, up to 50 participants, quality adjustment, and call establishment flow

---

### WO-19: Build Voice and Video Call User Interface with Controls

**Blueprint:** Voice and Video Calls with Screen Sharing

## Summary

Build the voice and video call iOS user interface — call initiation from contact profile or conversation, in-call controls (mute, video toggle, camera flip, end call), participant gallery view for group calls, connection quality indicators, and verified caller ID display with trust tier badge.

## In Scope

- Call initiation: call button on contact profile and conversation header; voice/video selection
- Incoming call screen: caller name, avatar, trust tier badge, answer/decline buttons
- In-call controls: mute/unmute (microphone), video on/off, camera flip (front/back), speaker toggle, end call button
- Participant gallery view: grid layout for up to 50 participants, speaker highlighting, auto-layout adjustment
- Connection quality indicator: color-coded (green/yellow/red) showing signal strength
- Call timer display
- Verified caller ID: show trust tier badge (`TrustBadge`) and verification checkmark (`VerificationBadge`) in call UI
- Virtual background: apply background blur using `AVFoundation` segmentation (iOS 15+) or Vision framework
- Call quality fallback: graceful degradation to audio-only if video quality degrades severely
- PiP (Picture-in-Picture) support: iOS native `AVPictureInPictureController` for background call UI

## Out of Scope

- WebRTC connection establishment (WO-5)
- Screen sharing UI (WO-31)
- Call recording controls (WO-43)
- Transcription display (WO-55)

## Requirements

Derived from the Voice and Video Calls blueprint.

**Call UI Architecture:**
```swift
// Presentation/Features/Calls/ActiveCallView.swift
struct ActiveCallView: View {
    @ObservedObject var callManager: WebRTCCallManager
    let session: CallSession

    var body: some View {
        ZStack {
            // Video grid for all participants
            ParticipantGridView(participants: callManager.participants)

            // Control bar overlay
            VStack {
                Spacer()
                HStack(spacing: 40) {
                    CallButton(icon: "mic.slash", isActive: callManager.isMuted) {
                        callManager.toggleMute()
                    }
                    CallButton(icon: "video.slash", isActive: !callManager.isVideoEnabled) {
                        callManager.toggleVideo()
                    }
                    CallButton(icon: "camera.rotate", action: callManager.flipCamera)
                    CallButton(icon: "phone.down.fill", tint: .red, action: callManager.endCall)
                }
                .padding(.bottom, 48)
            }

            // Quality indicator (top-right)
            ConnectionQualityBadge(quality: callManager.connectionQuality)
        }
    }
}
```

## Blueprints

- Voice and Video Calls with Screen Sharing — Defines call controls (mute, video, camera, end), participant gallery (up to 50), quality indicators, virtual backgrounds, and verified caller ID display

---

### WO-21: Build File Chunking and IPFS Distribution System

**Blueprint:** Large File Sharing and Cloud Storage Integration

## Summary

Build the file chunking and IPFS distribution system for large files up to 2GB. Files are chunked into 256KB pieces, each encrypted with a chunk-specific derived key, and uploaded in parallel to IPFS. IPFS pinning ensures availability. Supports resume on interrupted uploads.

## In Scope

- Automatic file chunking: split file into 256KB pieces; last chunk may be smaller
- Per-chunk key derivation: `chunkKey = HKDF(fileKey, info: "chunk_{index}")` — unique key per chunk
- Parallel chunk upload to IPFS: `URLSession` background upload tasks, max 5 concurrent
- Per-chunk SHA-256 integrity hash stored in file manifest (encrypted metadata)
- Retry logic: up to 3 retries per failed chunk with exponential backoff
- IPFS pinning via Pinata API (primary) with web3.storage fallback
- Resume capability: track last successfully uploaded chunk index in SwiftData; resume from there on app relaunch
- IPFS content addressing: construct file manifest with list of chunk CIDs + reassembly instructions
- IPFS gateway fallback: if direct IPFS node unavailable, use gateway for retrieval
- Memory efficiency: load and encrypt one 256KB chunk at a time (no 2GB in RAM)

## Out of Scope

- File encryption key management (WO-9)
- Blockchain file integrity anchoring (WO-34)
- Filecoin long-term storage (WO-185)
- Virus scanning (WO-58)
- Cloud storage integration (WO-46)

## Requirements

Derived from the Large File Sharing and Cloud Storage Integration blueprint.

**Chunking and Upload:**
```swift
// Core/FileSharing/FileChunkingService.swift
actor FileChunkingService {
    let chunkSize = 256 * 1024  // 256KB

    func uploadToIPFS(encryptedFile: EncryptedFile, fileKey: SymmetricKey) async throws -> [String] {
        let chunks = split(encryptedFile.content, into: chunkSize)
        var chunkCIDs: [String] = Array(repeating: "", count: chunks.count)

        // Parallel upload (max 5 concurrent)
        await withTaskGroup(of: (Int, String).self) { group in
            for (index, chunk) in chunks.enumerated() {
                group.addTask {
                    // Per-chunk encryption
                    let chunkKey = try! HKDF<SHA256>.deriveKey(inputKeyMaterial: fileKey, info: Data("chunk_\(index)".utf8), outputByteCount: 32)
                    let encryptedChunk = try! AES.GCM.seal(chunk, using: chunkKey).combined!

                    // Upload to IPFS
                    let cid = try! await self.ipfs.add(encryptedChunk)
                    return (index, cid)
                }
            }
            for await (index, cid) in group { chunkCIDs[index] = cid }
        }
        return chunkCIDs
    }
}
```

## Blueprints

- Large File Sharing and Cloud Storage Integration — Defines 256KB chunks, per-chunk encryption, parallel upload, IPFS storage, pinning services, retry logic, resume capability, and 2GB file size limit

---

### WO-31: Implement Screen Sharing with Encrypted Content Transmission

**Blueprint:** Voice and Video Calls with Screen Sharing

## Summary

Implement iOS screen sharing during calls using ReplayKit broadcast extension. Screen content is captured, encrypted end-to-end, and transmitted as a supplementary video track in the WebRTC session. Annotation tools (drawing, pointer highlighting) overlay the shared screen for participants.

## In Scope

- iOS `RPBroadcastSampleHandler` extension for screen capture (captures entire screen or selected app)
- Screen sharing as WebRTC supplementary video track alongside camera video
- E2E encryption of screen content via Noise Protocol (same encryption as audio/video streams)
- Annotation overlay: `PKCanvasView` (PencilKit) for drawing, pointer dot indicator for highlighting
- Real-time annotation sync: broadcast annotation state via WebSocket data channel
- Screen sharing controls in call UI: start/stop sharing button, share selection (full screen vs. app window)
- Visual indicator for all participants: "Alice is sharing screen" banner
- Quality target: 15–30 FPS adaptive, auto-reduce to 10 FPS under poor network
- Recording prevention: best-effort warnings (iOS cannot prevent screenshots technically)

## Out of Scope

- WebRTC base call connection (WO-5)
- Call recording of screen content (WO-43)
- Screen sharing scheduling

## Requirements

Derived from the Voice and Video Calls blueprint.

**Screen Capture (iOS):**
```swift
// Core/Calling/ScreenSharingManager.swift
class ScreenSharingManager: NSObject, RPScreenRecorderDelegate {
    func startScreenShare(into peerConnection: RTCPeerConnection) async throws {
        // 1. Request screen capture permission via ReplayKit
        let recorder = RPScreenRecorder.shared()
        try await recorder.startCapture { sampleBuffer, bufferType, error in
            if bufferType == .video {
                // 2. Convert CMSampleBuffer → RTCVideoFrame
                // 3. Add to screen-share video track in peer connection
                self.screenVideoSource.capturer(didCapture: videoFrame)
            }
        }
    }
}
```

## Blueprints

- Voice and Video Calls with Screen Sharing — Defines screen sharing capabilities: full/window sharing, annotation tools, pointer highlighting, permission controls, and 15-30 FPS quality target

---

### WO-34: Implement Blockchain File Integrity Verification

**Blueprint:** Large File Sharing and Cloud Storage Integration

## Summary

Implement blockchain file integrity verification — computing a SHA-256 hash of the complete file content and anchoring it to the Constellation Data L1 within 30 seconds of upload. Recipients verify the file by comparing their local hash against the on-chain record. Tamper detection triggers an alert when hashes don't match.

## In Scope

- SHA-256 hash computation for complete file (streaming, memory-efficient for up to 2GB)
- Data L1 submission via Metagraph Gateway: `{type: "file_integrity", fileHash, fileId, uploadedAt}`
- For multiple related files: build Merkle tree → anchor root hash (same pattern as message commitments)
- Blockchain confirmation tracking: watch for metagraph snapshot confirming anchoring, update file record
- Recipient verification: `GET /v1/files/{fileId}/integrity-proof` → returns `{fileHash, snapshotHash, snapshotHeight, merkleProof?}`
- iOS verification flow: recompute SHA-256 on downloaded file, compare against on-chain hash, display `✓ Verified` or `⚠️ Tampered`
- Verification within 10 seconds for up to 2GB (streaming SHA-256 computation)

## Out of Scope

- File encryption (WO-9)
- IPFS storage (WO-21)
- Filecoin long-term storage (WO-185)

## Requirements

Derived from the Large File Sharing blueprint.

**File Integrity Submission:**
```go
type FileIntegritySubmission struct {
    Type       string  // "file_integrity"
    FileID     string  // UUID
    FileHash   []byte  // SHA-256 of encrypted file (32 bytes)
    FileSize   int64
    UploadedAt time.Time
    SchemaVersion int
}
// Submit to Data L1 via Metagraph Gateway after IPFS upload completes
// Recipient verifies: sha256(downloadedFile) == onChainHash
```

## Blueprints

- Large File Sharing and Cloud Storage Integration — Defines SHA-256 file hashing, Constellation blockchain anchoring, hash verification by recipients, tamper detection, and Merkle tree for multiple files

---

### WO-46: Build Cloud Storage Integration Layer

**Blueprint:** Large File Sharing and Cloud Storage Integration

## Summary

Build the cloud storage integration layer — OAuth 2.0 flows for Google Drive, Dropbox, and OneDrive that allow users to select files from their cloud accounts. Selected files are streamed-downloaded, encrypted client-side, and processed through the IPFS file sharing pipeline. No unencrypted data is stored locally.

## In Scope

- OAuth 2.0 authorization flows for Google Drive (v3 API), Dropbox (v2 API), OneDrive (v1 API)
- Secure token storage in Keychain with automatic refresh handling
- Unified cloud file picker UI: tabbed view per provider, browse folder hierarchy, search files, preview thumbnails
- File selection: stream-download selected file from cloud provider → pipe directly into file encryption pipeline (WO-9) → no unencrypted local disk copy
- Cloud metadata preservation: filename, creation date, file type encrypted alongside content
- Provider rate limit handling: exponential backoff + user notification on 429 responses
- OAuth revocation: `DELETE /v1/integrations/{provider}` — delete token from Keychain, call provider revocation endpoint
- Support files up to 2GB from each provider

## Out of Scope

- File encryption (WO-9)
- IPFS upload (WO-21)
- Local file picker (uses iOS `UIDocumentPickerViewController` for local files)

## Requirements

Derived from the Large File Sharing blueprint.

**OAuth Integration:**
```swift
// Core/FileSharing/CloudStorageIntegrationManager.swift
enum CloudProvider: String { case googleDrive, dropbox, oneDrive }

actor CloudStorageIntegrationManager {
    func authorize(provider: CloudProvider) async throws -> CloudToken {
        // Initiate OAuth 2.0 PKCE flow via ASWebAuthenticationSession
        let authURL = provider.authorizationURL(scopes: ["files.readonly"])
        let redirectURL = try await ASWebAuthenticationSession(url: authURL, callbackURLScheme: "echo").start()
        let token = try await exchangeCodeForToken(from: redirectURL, provider: provider)
        try keychain.store("cloud_token_\(provider.rawValue)", value: encode(token))
        return token
    }

    func streamFile(fileId: String, provider: CloudProvider) -> AsyncStream<Data> {
        // Stream-download file from cloud provider API
        // Pipe directly to FileEncryptionService (WO-9)
        // Never buffer full file to disk unencrypted
    }
}
```

## Blueprints

- Large File Sharing and Cloud Storage Integration — Defines Google Drive, Dropbox, and OneDrive integration with OAuth, file selection, automatic encryption, metadata preservation, and access revocation

---

### WO-58: Implement File Security Scanning and Threat Detection

**Blueprint:** Large File Sharing and Cloud Storage Integration

## Summary

Build the file security scanning system that checks uploaded files for malware and ransomware before IPFS distribution. The backend Media Service (port 8008) coordinates scans via multiple antivirus API integrations. Files with detected threats are blocked and quarantined.

## In Scope

- Backend Media Service (port 8008) scanning pipeline: run on each uploaded encrypted chunk set before IPFS pinning
- Multi-engine scanning: primary (VirusTotal API), secondary (Cloudmersive AV API) for redundancy
- Hash-based pre-check: compute SHA-256 of file, query VirusTotal hash lookup first (fast path, no upload needed if hash known)
- Full scan: upload decrypted sample to scanning service (NEVER uploaded to IPFS) for unknown files
- Result aggregation: majority vote across engines; any `malicious` verdict = block
- Malicious file handling: quarantine in temporary storage, notify sender with threat details, prevent IPFS pin
- Scan logging: `{fileId, scanEngines[], verdict, threatName?, timestamp}` in PostgreSQL (no file content)
- Scan timeout: max 60 seconds for files ≤ 100MB; large files (> 100MB) use hash-only fast path

## Out of Scope

- Antivirus engine development
- User-side scanning (backend only)
- Manual file review

## Requirements

Derived from the Large File Sharing blueprint.

**Scanning Pipeline:**
```go
// media/scanner.go
type FileScanResult struct {
    FileID    string
    Verdict   ScanVerdict  // .clean, .malicious, .suspicious, .unknown
    ThreatName string      // e.g., "Trojan.GenericKD.12345"
    Engines   []EngineResult
    ScannedAt time.Time
}

func (s *FileScanner) Scan(fileData []byte, fileHash []byte) (FileScanResult, error) {
    // 1. Fast path: VirusTotal hash lookup (no upload)
    if result, err := s.virustotal.LookupHash(fileHash); err == nil {
        return mapVTResult(result), nil
    }
    // 2. Full scan: upload to scanning API
    vtResult, _ := s.virustotal.ScanFile(fileData)
    cloudResult, _ := s.cloudmersive.ScanFile(fileData)
    return aggregateResults(vtResult, cloudResult), nil
}
```

## Blueprints

- Large File Sharing and Cloud Storage Integration — Defines decentralized virus scanning, multiple engine integration, malware/ransomware detection, blocking, quarantine, and scan result reporting

---

### WO-68: Build File Expiration and Cryptographic Deletion System

**Blueprint:** Large File Sharing and Cloud Storage Integration

## Summary

Implement file expiration and cryptographic deletion — when a file's expiration time is reached, its encryption key is destroyed (making the file permanently unreadable), the IPFS pins are removed, and a deletion event is anchored on the Constellation blockchain as proof. Notification is sent 24 hours before expiry.

## In Scope

- Expiration options: 1 hour, 1 day, 1 week, 1 month, 1 year, never (configurable at share time)
- `BGProcessingTask` scheduled at expiry time for iOS; backend also tracks expiry and initiates if iOS app is offline
- Cryptographic deletion: delete file encryption key from Keychain (iOS) and from backend Media Service key storage
- IPFS unpin: call Pinata and web3.storage unpin API for all chunk CIDs
- Blockchain deletion record: submit `{type: "file_deletion", fileId, fileHash, deletedAt}` to Data L1 as proof
- Pre-expiry notification: push notification 24 hours before expiry with "Extend" and "Confirm Delete" options
- Manual deletion: user-initiated at any time, same pipeline
- Deletion verification: confirm key is gone (cannot decrypt test chunk) before marking as deleted in database

## Out of Scope

- File encryption (WO-9)
- IPFS node provisioning
- Disappearing messages (separate feature)

## Requirements

Derived from the Large File Sharing blueprint.

**Deletion Flow:**
```swift
// Core/FileSharing/FileExpirationManager.swift
actor FileExpirationManager {
    func deleteFile(_ fileId: String) async throws {
        // 1. Destroy encryption key (iOS Keychain)
        try keychain.delete("file_key_\(fileId)")

        // 2. Unpin all chunk CIDs from IPFS
        let manifest = try await fetchFileManifest(fileId)
        for cid in manifest.chunkCIDs {
            try await ipfs.unpin(cid)
        }

        // 3. Anchor deletion proof on Data L1
        try await metagraph.submitDataL1(DeletionSubmission(
            type: "file_deletion",
            fileId: fileId,
            fileHash: manifest.fileHash,
            deletedAt: Date()
        ))

        // 4. Mark as deleted in local DB
        database.markFileDeleted(fileId)
    }
}
```

## Blueprints

- Large File Sharing and Cloud Storage Integration — Defines configurable expiration times, cryptographic key destruction, automatic deletion scheduling, blockchain deletion records, expiration notifications, and manual deletion

---

### WO-70: Implement Core Group Creation and Management System

**Blueprint:** Public and Private Groups with Verified Status Display

## Summary

Build the core group creation and management system — creating public and private groups with metadata, configuring privacy settings, managing invite links, and anchoring group metadata on Data L1. Trust tier determines maximum group size limits.

## In Scope

- Group creation: name (max 100 chars), description (max 500 chars), avatar (PNG/JPG/GIF, 5MB), category, up to 10 topic tags
- Public vs. private selection: public groups searchable; private groups invite-only
- Private group invite links: unique time-limited links (7-day default, configurable), `POST /v1/groups/{id}/invite-links`
- Trust tier group size limits: Tier 1 → 10 members, Tier 2 → 50, Tier 3 → 500, Tier 4 → 10,000, Tier 5 → 1,000,000
- Data L1 metadata anchoring: `{type: "group_metadata", groupId, adminDID, memberCountHash, createdAt}` — NO member list on-chain
- Group settings modification by owner: name, description, avatar, category, tags
- Group admin DID validation: admin must have a valid did:key DID registered on the Constellation Identity Metagraph
- Symmetric group key generation by creator (via `GroupKeyManager`) — distributed to founding members

## Out of Scope

- Verification requirements (WO-81)
- Group discovery/search (WO-101)
- Moderation (WO-131, WO-112)
- Permission structures (WO-123)

## Requirements

Derived from the Public and Private Groups with Verified Status Display blueprint.

**Group Creation API:**
```go
// POST /v1/groups
type CreateGroupRequest struct {
    Name        string   `json:"name"`        // Max 100 chars
    Description string   `json:"description"` // Max 500 chars
    IsPublic    bool     `json:"is_public"`
    Category    string   `json:"category"`
    Tags        []string `json:"tags"`        // Max 10 tags
    AdminDID    string   `json:"admin_did"`
}
// Response includes: groupId, inviteLink (for private groups), symmetricKeyEncrypted (for admin)
```

**Data L1 Anchoring:**
```go
type GroupMetadataSubmission struct {
    Type            string    // "group_metadata"
    GroupID         string    // UUID
    AdminDID        string    // Creator's DID
    MemberCountHash []byte    // H(memberCount || groupSalt) — privacy-preserving
    CreatedAt       time.Time
    SchemaVersion   int
    // NEVER: member list, group name, description (privacy)
}
```

## Blueprints

- Public and Private Groups with Verified Status Display — Defines group creation, public/private distinction, trust tier size limits, invite links, group key management, and Data L1 anchoring

---

### WO-77: Implement Core Channel Creation and Management System

**Blueprint:** Broadcast Channels and Community Features

## Summary

Build the core broadcast channel creation and management system — users create one-to-many communication channels with name, description, avatar, category, and topic tags. Channel metadata is stored on the backend and discoverable. Channels use the standard relay for content distribution to all subscribers.

## In Scope

- Channel creation: name (3–50 chars), description (max 500 chars), avatar (PNG/JPG/GIF, max 2MB), category, up to 10 tags
- Unique channel ID generation and metadata storage in backend PostgreSQL
- Channel management interface for owner: edit name, description, avatar, category, tags
- Input validation: unique name within category, no duplicate channels, content policy check
- Channel metadata endpoint: `GET /v1/channels/{id}` — returns all public metadata
- Channel list management: archive/unarchive channel, delete channel (with confirmation)
- Channel creation confirmation step: show summary before finalizing

## Out of Scope

- Privacy settings (WO-88)
- Content posting (WO-97)
- Subscriber management
- Blockchain anchoring

## Requirements

Derived from the Broadcast Channels and Community Features blueprint.

**Channel Data Model:**
```go
// POST /v1/channels
type CreateChannelRequest struct {
    Name        string   `json:"name"`        // 3–50 chars
    Description string   `json:"description"` // Max 500 chars
    Category    string   `json:"category"`
    Tags        []string `json:"tags"`        // Max 10
    IsPublic    bool     `json:"is_public"`
}

type Channel struct {
    ID          string    `db:"id"`
    OwnerDID    string    `db:"owner_did"`
    Name        string    `db:"name"`
    Description string    `db:"description"`
    AvatarURL   string    `db:"avatar_url"`
    Category    string    `db:"category"`
    Tags        []string  `db:"tags"`
    SubscriberCount int   `db:"subscriber_count"`
    CreatedAt   time.Time `db:"created_at"`
    IsPublic    bool      `db:"is_public"`
    IsArchived  bool      `db:"is_archived"`
}
```

## Blueprints

- Broadcast Channels and Community Features — Defines channel creation, naming, category selection, topic tagging, channel management interface, and metadata storage

---

### WO-79: Develop File Management and Organization Interface

**Blueprint:** Large File Sharing and Cloud Storage Integration

## Summary

Build the iOS file management and organization interface — a file browser for managing all shared files with folder organization, tagging, search, file preview, and permission management. Connects to the backend Media Service (port 8008) for file records and manages IPFS CID-based retrieval.

## In Scope

- `FilesTab` view: grid/list toggle showing shared files organized by folder
- Folder creation, rename, delete with drag-and-drop file organization
- File tagging: add custom tags for organization and searchability
- File search: by filename, file type, tag, or sharing date; results within 2 seconds for up to 10,000 files
- File preview: inline thumbnail for images, PDF first-page preview, generic icon for other types
- Sharing history per file: recipients, sharing dates, download count, current access status
- Permission management: revoke recipient access, modify expiration date, update sharing settings
- Sorting and filtering: by name, size, type, sharing date, expiration date; preferences persisted
- File download status tracking: queued, downloading, decrypting, ready

## Out of Scope

- File encryption (WO-9)
- IPFS upload (WO-21)
- Collaborative document editing (WO-188)

## Requirements

Derived from the Large File Sharing blueprint.

**File Model:**
```swift
struct SharedFile: Identifiable, Codable {
    let id: String            // UUID
    var displayName: String
    let mimeType: String
    let fileSize: Int64
    let uploadedAt: Date
    var expiresAt: Date?
    var tags: [String]
    var folderId: UUID?
    var recipients: [SharedFileRecipient]
    var downloadCount: Int
    var ipfsManifestCID: String  // CID of encrypted manifest
    var integrityHash: Data      // SHA-256 for verification (WO-34)
    var status: FileStatus       // .active, .expired, .deleted
}
```

## Blueprints

- Large File Sharing and Cloud Storage Integration — Defines file organization in folders, file search, tagging, preview, download history, file permission management, sorting/filtering

---

### WO-81: Build Group Verification Requirements and Badge System

**Blueprint:** Public and Private Groups with Verified Status Display

## Summary

Build the group verification requirements and badge system. Group creators configure the minimum trust tier required to join. The group badge level is derived from the percentage of Tier 3+ members in the group. Backend enforces verification requirements on join attempts.

## In Scope

- Minimum trust tier configuration during group creation and editable by admins: Open (Tier 1), Standard (Tier 2), Verified (Tier 3), High-Trust (Tier 4), Restricted (Tier 5)
- Join validation: backend checks `requesterTrustTier >= group.minimumRequiredTier` before allowing join
- Group badge calculation from member composition:
  - < 30% Tier 3+ → Unverified (gray)
  - 30–50% → Basic (bronze)
  - 50–75% → Verified (silver)
  - 75–90% → Trusted (gold)
  - 90%+ → Elite (blue)
- `calculateGroupBadge(members: [Member])` implementation (from blueprint)
- Real-time badge updates: recalculate badge when members join/leave or tier changes
- Badge display: in group list, group header, search results
- Tap badge → detail sheet showing verification breakdown percentages
- Admin exception: admin can manually approve users below the minimum tier (logged in audit trail)

## Out of Scope

- Group creation (WO-70)
- Participant verification display (WO-90)
- Permission enforcement (WO-123)

## Requirements

Derived from the Public and Private Groups blueprint.

**Badge Calculation:**
```swift
func calculateGroupBadge(members: [Member]) -> GroupBadge {
    let tier3Plus = members.filter { $0.trustTier >= 3 }.count
    let percentage = Double(tier3Plus) / Double(members.count)
    switch percentage {
    case 0.9...:       return .elite      // Blue ✦ 90%+
    case 0.75..<0.9:   return .trusted    // Gold ● 75-90%
    case 0.5..<0.75:   return .verified   // Silver ◑ 50-75%
    case 0.3..<0.5:    return .basic      // Bronze ◐ 30-50%
    default:           return .unverified // Gray ○ <30%
    }
}
```

## Blueprints

- Public and Private Groups with Verified Status Display — Defines minimum trust tier requirements, group badge levels (Unverified/Basic/Verified/Trusted/Elite), badge calculation formula, and enforcement

---

### WO-88: Build Channel Privacy Configuration and Access Control System

**Blueprint:** Broadcast Channels and Community Features

## Summary

Implement channel privacy configuration — public (open), private (invite-only), and semi-private (discoverable but approval-gated). Supports invite link generation with usage limits and expiration. Subscriber approval workflow for semi-private channels.

## In Scope

- Channel type configuration: `public`, `private`, `semi-private` (set on creation, modifiable by owner)
- Public channel: discoverable, joinable by anyone → `POST /v1/channels/{id}/subscribe` directly
- Private channel: invite link required → `POST /v1/channels/{id}/invite-links` generates time-limited URL
- Semi-private channel: subscribe triggers `RequestApproval` → admin reviews queue, approve/deny
- Invite link: configurable expiry (1h to 30 days), usage limit (1–1000 uses), stored in PostgreSQL
- Approval request queue: admins see `{requesterDID, requestedAt, requesterTrustTier, requesterBio}` list
- Approve/deny with optional message; approved users receive push notification
- Privacy change: owner can switch between types; existing subscribers keep access; pending requests cancelled on switch to private

## Out of Scope

- Channel creation (WO-77)
- Content encryption (standard E2E from messaging system)
- Moderation (WO-121)

## Requirements

Derived from the Broadcast Channels blueprint.

**Privacy Config API:**
```go
// PATCH /v1/channels/{id}/privacy
type ChannelPrivacyUpdate struct {
    PrivacyType string `json:"privacy_type"` // "public", "private", "semi_private"
}

// POST /v1/channels/{id}/invite-links
type InviteLinkRequest struct {
    ExpiresAfterHours int `json:"expires_after_hours"` // 1–720
    MaxUses           int `json:"max_uses"`             // 1–1000, 0 = unlimited
}

// POST /v1/channels/{id}/subscription-requests/{requestId}/approve
// POST /v1/channels/{id}/subscription-requests/{requestId}/deny
```

## Blueprints

- Broadcast Channels and Community Features — Defines public/private/semi-private channel types, invite link generation with expiry and limits, subscriber approval workflow

---

### WO-90: Develop Participant Verification Display and Role Management

**Blueprint:** Public and Private Groups with Verified Status Display

## Summary

Build the participant verification display and role management UI for group conversations. Each member row shows their trust tier badge, verification checkmark (Tier 4+), and role indicator (Owner/Admin/Moderator). Admins can promote/demote members through a role assignment interface.

## In Scope

- `GroupMemberRow` SwiftUI component: avatar + display name + `TrustBadge` + role badge + `VerificationBadge` (Tier 4+)
- Member list sortable by: trust tier, role, join date, display name
- Role indicators: Owner (👑), Admin (🛡️), Moderator (⚒️), Member (no icon)
- Tap member → full profile sheet showing trust tier, verification status, join date, role
- Role assignment: Owner can promote to Admin; Admin can promote to Moderator; Owner can demote any role
- Role change confirmation dialog with reason field
- Role changes logged with `{actorDID, targetDID, newRole, reason, timestamp}` in PostgreSQL (group audit log)
- Real-time status updates: trust tier badge refreshes within 30 seconds of tier change (WebSocket push)

## Out of Scope

- Permission enforcement logic (WO-123)
- Moderation actions (WO-131)
- Group badge calculation (WO-81)

## Requirements

Derived from the Public and Private Groups blueprint.

**Member Row:**
```swift
struct GroupMemberRow: View {
    let member: GroupMember
    var body: some View {
        HStack {
            Avatar(member.avatarURL)
            VStack(alignment: .leading) {
                HStack {
                    Text(member.displayName)
                    TrustBadge(tier: member.trustTier)
                    if member.trustTier >= 4 { VerificationBadge(status: .identityVerified) }
                    if let role = member.role, role != .member { RoleBadge(role: role) }
                }
                Text("Tier \(member.trustTier) · \(member.role?.displayName ?? "Member")")
                    .font(.caption).foregroundColor(.secondary)
            }
        }
    }
}
```

## Blueprints

- Public and Private Groups with Verified Status Display — Defines per-member trust tier badge, admin/moderator role indicators, verification checkmark, role assignment interface, and real-time status updates

---

### WO-97: Develop Multi-Format Content Posting and Distribution System

**Blueprint:** Broadcast Channels and Community Features

## Summary

Build the multi-format content posting pipeline for broadcast channels — channel creators post text, images, video, files, and polls. All content is E2E encrypted and distributed to subscribers via the relay's fan-out mechanism. Subscribers receive push notifications for new posts.

## In Scope

- Text post: max 10,000 characters with Markdown formatting (bold, italic, links)
- Image post: PNG/JPG/GIF/WebP, max 10MB, auto-generate thumbnail (JPEG, 200px max dimension)
- Video post: MP4/WebM/MOV, max 100MB, max 10 minutes, generate thumbnail from first frame
- File post: documents/PDFs/archives, max 25MB, virus scan before distribution (WO-58 pipeline)
- Poll post: up to 10 options, 1h–30 days voting window, configurable result visibility
- All content E2E encrypted: media uploaded to IPFS via Media Service (port 8008), content hash anchored on Data L1
- `POST /v1/channels/{id}/posts` — content type discriminated union, validates and routes appropriately
- Subscriber notification: Notification Service (port 8007) sends APNs push to all subscribers with channel name + "New post" (no content preview)
- Unique post IDs, timestamps, author DID in all posts

## Out of Scope

- Scheduled posting (WO-139)
- Content categorization/tagging (WO-172)
- Moderation filtering (WO-121)
- Channel analytics (WO-158)

## Requirements

Derived from the Broadcast Channels blueprint.

**Content Post Type:**
```go
type ChannelPost struct {
    PostID      string     `json:"post_id"`
    ChannelID   string     `json:"channel_id"`
    AuthorDID   string     `json:"author_did"`
    ContentType PostType   `json:"content_type"`  // text, image, video, file, poll
    EncryptedPayload Data  `json:"encrypted_payload"` // Kinnami encrypted
    MediaCID    string     `json:"media_cid,omitempty"` // IPFS CID for media
    Timestamp   time.Time  `json:"timestamp"`
}
```

## Blueprints

- Broadcast Channels and Community Features — Defines supported content types (text, image, video, file, poll), content distribution, subscriber notifications, and content validation

---

### WO-101: Implement Group Discovery and Search Functionality

**Blueprint:** Public and Private Groups with Verified Status Display

## Summary

Implement group discovery and search for public groups. Users can search by name, description keywords, category, and trust badge level. Results include group metadata and a join preview. Privacy is preserved by not logging search terms per user. Private groups are excluded from discovery.

## In Scope

- `GET /v1/groups/search?q={query}&category={cat}&minBadge={level}&maxMembers={n}&page={p}` backend endpoint
- Full-text search over group name, description, and tags in PostgreSQL (GIN text search index)
- Category browsing: `/v1/groups?category={cat}&page={p}` returns paginated group list (20 per page)
- Badge level filter: only show groups with badge ≥ specified level
- Group preview before joining: group name, avatar, badge, member count, category, brief description, minimum trust requirement
- Join validation: check `userTrustTier >= group.minimumRequiredTier` before allowing join; return 403 with required tier if not qualified
- iOS group discovery view: search bar, category scroll, result cards with `GroupBadge` indicator
- Privacy-preserving search: backend logs aggregate search counts only (no per-user query logs)
- Private groups: never returned in search results; accessible by invite link only

## Out of Scope

- Group creation (WO-70)
- Advanced ML recommendations
- Private group discovery

## Requirements

Derived from the Public and Private Groups blueprint.

**Search Request/Response:**
```swift
struct GroupSearchRequest {
    var query: String
    var category: Category?
    var minimumVerificationLevel: GroupBadge?
    var maxMemberCount: Int?
    var page: Int
}

struct GroupSearchResult {
    let groupId: String
    let name: String
    let description: String
    let memberCount: Int
    let verificationBadge: GroupBadge
    let category: Category
    let minimumTrustTier: Int    // Required to join
    let avatarURL: URL?
}
```

## Blueprints

- Public and Private Groups with Verified Status Display — Defines public group search (name/description/tags/category), badge filtering, member count filtering, join preview, and privacy-preserving search

---

### WO-108: Create Channel Discovery and Search Infrastructure

**Blueprint:** Broadcast Channels and Community Features

## Summary

Build the channel discovery and search infrastructure — full-text search by name/description/tags, category browsing, trending channel detection, and channel preview before subscribing. Public channels only; private channels are excluded from discovery.

## In Scope

- `GET /v1/channels/search?q={query}&category={cat}&minSubscribers={n}&page={p}` with full-text PostgreSQL search
- Trending channels: update hourly based on subscriber growth rate + post engagement; `GET /v1/channels/trending`
- Category browsing: `GET /v1/channels?category={cat}&page={p}` paginated (20 per page)
- Channel preview: `GET /v1/channels/{id}/preview` — recent 5 posts (encrypted, preview only accessible to subscribers), subscriber count, description
- Subscribe action: `POST /v1/channels/{id}/subscribe` with trust tier validation where applicable
- Fuzzy search with trigram index for typo tolerance
- Privacy: private channels excluded from all search/browse/trending
- Aggregate search analytics: query count per term (no per-user logging)

## Out of Scope

- Advanced ML recommendations
- Private channel discovery
- Channel quality scoring

## Requirements

Derived from the Broadcast Channels blueprint.

**Search Response:**
```go
type ChannelSearchResult struct {
    ChannelID       string `json:"channel_id"`
    Name            string `json:"name"`
    DescriptionPreview string `json:"description_preview"` // First 100 chars
    SubscriberCount int    `json:"subscriber_count"`
    Category        string `json:"category"`
    IsVerified      bool   `json:"is_verified"`  // Owner's trust tier 4+
    PostFrequency   string `json:"post_frequency"` // "daily", "weekly", "inactive"
}
```

## Blueprints

- Broadcast Channels and Community Features — Defines channel search, category-based discovery, recommendation engine, trending channels, channel preview, and subscription from discovery

---

### WO-112: Build Group Moderation and Evidence-Based Reporting System

**Blueprint:** Public and Private Groups with Verified Status Display

## Summary

Build the group moderation tools — automated message filtering by trust tier, temporary muting, permanent banning, spam detection, and the evidence-based reporting workflow. Moderation logs are stored locally and in encrypted IPFS audit storage for Organization tier groups.

## In Scope

- `ModerationSettings` per group: `minimumTierToPost`, `spamDetectionEnabled`, `mutedMembers`, `bannedMembers`
- Auto-filter: messages from users below `minimumTierToPost` are held for admin review (not delivered to group)
- Temporary mute: `ModerationAction.mute(duration: TimeInterval)` — DID-based, backend blocks relay for muted DID in this group
- Permanent ban: backend rejects all messages and join attempts from banned DID in group
- Evidence-based report: `{reporterDID, targetDID, description: max 500 chars, evidence: FileAttachment?}` submitted to group admins
- Moderation event log: `{eventId, timestamp, moderatorDID, targetDID, action, reason, evidenceHash?}`
- Spam detection: backend flags messages with >3 links, duplicate content within 60 seconds, or velocity abuse
- Appeal process: banned users submit `{did, groupId, appealReason}` to group owner; owner responds within 7 days

## Out of Scope

- Blockchain anchoring of moderation records (WO-156)
- Advanced AI content filtering
- Cross-group coordination

## Requirements

Derived from the Public and Private Groups blueprint.

**Moderation Data Models (from blueprint):**
```swift
struct ModerationSettings {
    var messageFilteringEnabled: Bool
    var minimumTierToPost: TrustTier
    var spamDetectionEnabled: Bool
    var mutedMembers: Set<String>   // DIDs
    var bannedMembers: Set<String>  // DIDs
}

struct ModerationEvent {
    let eventId: String
    let timestamp: Date
    let moderatorDID: String
    let targetDID: String
    let action: ModerationAction
    let reason: String
    let evidenceHash: Data?  // SHA-256 of evidence file if attached
}

enum ModerationAction {
    case mute(duration: TimeInterval)
    case unmute, ban, unban, deleteMessage, warnUser
}
```

## Blueprints

- Public and Private Groups with Verified Status Display — Defines moderation features, message filtering, muting/banning, evidence-based reporting, spam detection, and appeal process

---

### WO-121: Implement Channel Moderation and Content Management Tools

**Blueprint:** Broadcast Channels and Community Features

## Summary

Build channel moderation tools — admins can delete posts, mute/ban users, configure spam detection, and maintain a moderation log. Moderation actions include required reason selection. An appeal process allows affected users to contest decisions.

## In Scope

- `DELETE /v1/channels/{id}/posts/{postId}` with `reason` (required): deletes post, notifies author, logs action
- User mute: `POST /v1/channels/{id}/mutes/{userDID}` with `duration` and `reason` — muted users cannot post
- User ban: `POST /v1/channels/{id}/bans/{userDID}` with `reason` — removes from channel, prevents re-subscribe
- Spam detection rules: multiple links (>3 in a post), rapid consecutive posting (<30s), duplicate content within 1 hour → auto-flag for review queue
- Moderation log: `GET /v1/channels/{id}/moderation-log` (admin only) with action type, actor, target, reason, timestamp
- Appeal endpoint: `POST /v1/channels/{id}/appeals` — user submits appeal; admin must respond within 48 hours
- Bulk moderation: `DELETE /v1/channels/{id}/posts` with array of post IDs
- Transparency: public log of deleted posts count per day (anonymized, no content) accessible to subscribers

## Out of Scope

- AI-powered content analysis
- Community voting on moderation decisions
- Blockchain anchoring (Channels don't anchor moderation like Enterprise groups)

## Requirements

Derived from the Broadcast Channels blueprint.

**Moderation Log Entry:**
```go
type ChannelModerationEvent struct {
    EventID      string    `db:"event_id"`
    ChannelID    string    `db:"channel_id"`
    AdminDID     string    `db:"admin_did"`
    TargetType   string    `db:"target_type"` // "post" or "user"
    TargetID     string    `db:"target_id"`
    Action       string    `db:"action"`      // "delete", "mute", "ban"
    Reason       string    `db:"reason"`
    Duration     *int64    `db:"duration_seconds,omitempty"` // For mutes
    Timestamp    time.Time `db:"timestamp"`
}
```

## Blueprints

- Broadcast Channels and Community Features — Defines channel moderation: post deletion, user muting/banning, spam detection, moderation logs, and appeals process

---

### WO-123: Develop Permission Structure and Role-Based Access Control

**Blueprint:** Public and Private Groups with Verified Status Display

## Summary

Implement the group permission structure and role-based access control system. Four roles (Member/Moderator/Admin/Owner) combined with trust tier-based capabilities define what each user can do. Permission enforcement occurs on both the backend and iOS client. Custom permission configuration per group is supported.

## In Scope

- Permission matrix enforcement on backend: check role + trust tier before allowing each group action
- Backend: `PermissionService` validates `{userDID, groupId, action}` → `allow | deny | requireHigherTier`
- Default permission matrix (from blueprint): Member can send messages; Moderator can pin/mute; Admin can ban/invite; Owner can delete group
- Trust tier-based hybrid: media sharing (Tier 2+), member invitations (Tier 3+)
- Custom permission configuration: group admin can override defaults per role
- Permission inheritance: lower roles cannot exceed higher role capabilities (enforced on save)
- Permission delegation: temporary elevation for specific users (1h–7d), stored in PostgreSQL with expiry
- iOS UI: per-action permission check before showing action UI; clear error messages on denied actions

## Out of Scope

- Moderation tools (WO-131)
- Governance voting (WO-148)
- Group discovery (WO-101)

## Requirements

Derived from the Public and Private Groups blueprint.

**Permission Matrix (Default):**

| Permission | Member | Moderator | Admin | Owner |
|---|---|---|---|---|
| Send messages | ✓ | ✓ | ✓ | ✓ |
| Share media | Tier 2+ | ✓ | ✓ | ✓ |
| Pin messages | ✗ | ✓ | ✓ | ✓ |
| Mute members | ✗ | ✓ | ✓ | ✓ |
| Ban members | ✗ | ✗ | ✓ | ✓ |
| Invite members | Tier 3+ | ✓ | ✓ | ✓ |
| Edit group info | ✗ | ✗ | ✓ | ✓ |
| Delete group | ✗ | ✗ | ✗ | ✓ |

## Blueprints

- Public and Private Groups with Verified Status Display — Defines role permission matrix, trust tier-based hybrid permissions, custom configuration, permission inheritance, and delegation

---

### WO-130: Build Channel Reporting and Data Export System

**Blueprint:** Broadcast Channels and Community Features

## Summary

Build the channel data export and compliance reporting system. Channel creators can export analytics data as CSV, JSON, or PDF with custom date ranges and metric selection. Automated periodic reports can be scheduled. All exports are anonymized (no subscriber PII).

## In Scope

- Report generation: `POST /v1/channels/{id}/reports` — select metrics, date range, format (CSV/JSON/PDF)
- Automated reports: schedule daily/weekly/monthly reports to be generated and pushed via notification or email
- Compliance exports: anonymized aggregate subscriber data summaries (total count, tier distribution, geographic region as country-level only)
- Historical data retention: analytics stored for up to 2 years in PostgreSQL with automatic archival
- Report templates: save and reuse custom metric + date range combinations
- Large report handling: asynchronous generation for reports > 10MB, downloadable via signed URL when ready
- Report sharing: `POST /v1/channels/{id}/reports/{reportId}/share` — generate time-limited access link
- Data integrity verification: SHA-256 checksum included in all exported files

## Out of Scope

- Real-time analytics dashboard (WO-158)
- External BI tool integration

## Requirements

Derived from the Broadcast Channels blueprint.

**Export Request:**
```go
type ChannelReportRequest struct {
    Metrics   []string   `json:"metrics"`   // e.g., ["subscriber_count", "post_views", "engagement_rate"]
    DateRange DateRange  `json:"date_range"`
    Format    string     `json:"format"`    // "csv", "json", "pdf"
    Anonymized bool      `json:"anonymized"` // Always true for compliance exports
}
```

## Blueprints

- Broadcast Channels and Community Features — Defines data export (CSV/JSON/PDF), automated reporting, compliance reports, historical retention, report templates, and report sharing

---

### WO-131: Build Group Moderation and Evidence-Based Reporting System

**Blueprint:** Public and Private Groups with Verified Status Display

## Summary

Build the comprehensive group moderation system with blockchain-anchored audit trail. Extends WO-112 with immutable moderation records anchored on Data L1 (for Organization tier groups), a transparency dashboard showing moderation statistics, and a structured appeals workflow with time-bound response requirements.

## In Scope

- All capabilities from WO-112 (filtering, muting, banning, spam detection, evidence-based reports)
- Blockchain anchoring: for Organization tier groups, submit `H(moderationEvent || salt)` to Data L1 — proves moderation occurred without revealing details
- Moderation transparency dashboard: available to all group members — shows aggregate action counts, recent policy changes (not individual actions unless configured)
- 7-day appeal window for bans with structured appeal form and mandatory response deadline
- Automatic moderation log export: Organization tier can export encrypted IPFS-stored moderation logs for compliance
- Trust score impact recording: report-based penalties feed back to Trust Service for score adjustment

## Out of Scope

- AI-powered content filtering
- Cross-group moderation

## Requirements

Derived from the Public and Private Groups blueprint.

**Blockchain Anchoring for Moderation (Organization Tier):**
```go
// After each moderation action (Org tier groups only):
type ModerationAuditSubmission struct {
    Type          string  // "group_moderation"
    GroupID       string
    EventHash     []byte  // H(moderationEvent || salt) — privacy-preserving
    ActionType    string  // "mute", "ban", "delete_message"
    Timestamp     time.Time
    SchemaVersion int
    // NEVER: moderator DID, target DID, or reason in plaintext
}
```

**Note:** WO-112 covers all groups; this work order (WO-131) adds the blockchain anchoring and compliance-grade audit trail that only Organization tier groups require.

## Blueprints

- Public and Private Groups with Verified Status Display — Defines moderation logs, immutable audit trail, automatic spam detection, appeals process, and moderation transparency

---

### WO-139: Develop Scheduled Posting and Content Planning System

**Blueprint:** Broadcast Channels and Community Features

## Summary

Build the channel scheduled posting system — creator drafts a post, selects a future delivery time, and the system stores it locally and delivers it at the scheduled time via `BGProcessingTask`. Supports recurring post patterns, bulk scheduling via CSV, draft management, and timezone-aware delivery.

## In Scope

- Scheduling UI: date/time picker with presets (1h, 6h, 1d, 1w) and custom selection
- Timezone detection: show recipient's estimated timezone if available; convert display times accordingly
- Recurring patterns: daily, weekly, monthly; configurable end date or occurrence count
- Draft management: save incomplete posts locally in SwiftData, auto-save on background
- Scheduled post queue view: list all pending posts with preview, scheduled time, status (pending, processing, delivered, failed)
- Edit/reschedule before delivery: modify content or delivery time
- Cancel scheduled post: remove from SwiftData, cancel `BGTask`
- Bulk scheduling: import CSV with `{content, scheduled_for, recurrence?}` rows
- `BGProcessingTask` fires at scheduled time, calls standard `POST /v1/channels/{id}/posts` endpoint

## Out of Scope

- AI-powered optimal timing recommendations
- External calendar integration
- Content approval workflows

## Requirements

Derived from the Broadcast Channels blueprint.

**Scheduled Post Model:**
```swift
struct ScheduledChannelPost: Identifiable, Codable {
    let id: UUID
    let channelId: String
    var contentType: PostType
    var encryptedContent: Data       // Encrypted locally
    let scheduledFor: Date
    var isRecurring: Bool
    var recurrenceRule: RecurrenceRule?
    var status: ScheduledStatus      // .draft, .pending, .delivered, .failed
}
// Same BGProcessingTask architecture as scheduled messages (WO-65)
// On delivery: decrypt → POST /v1/channels/{channelId}/posts
```

## Blueprints

- Broadcast Channels and Community Features — Defines scheduled posting, preset intervals, recurring patterns, draft management, timezone handling, and bulk scheduling

---

### WO-148: Build Group Governance and Voting System

**Blueprint:** Public and Private Groups with Verified Status Display

## Summary

Build the group governance and voting system — community-driven proposals and voting on group policy, moderation decisions, and community direction. Vote results are blockchain-anchored on Data L1 for transparency and immutability. Supports yes/no, multiple choice, and configurable quorum requirements.

## In Scope

- Governance proposal creation: title, description, vote type (yes/no, multiple choice), voting period (1–30 days), quorum threshold (10–75% of eligible members)
- Eligible voters: members who meet minimum trust tier requirement for governance (configurable, default Tier 2+)
- Voting UI: proposal card with progress bars per option, participation percentage, time remaining
- Anonymous voting option (configurable per proposal): votes recorded without linking DID to choice
- Quorum check: if quorum not met by deadline, proposal fails automatically
- Result execution: approved proposals trigger automatic notifications; manual admin execution for rule changes
- Data L1 anchoring of results: `{type: "governance_vote", groupId, proposalId, result, voteCount, participationPct, closedAt}`
- Voting history: permanent record of all proposals with outcomes in group settings

## Out of Scope

- Cross-group governance
- Complex ranked-choice algorithms (yes/no and multiple choice only for v1)
- Legal binding of decisions

## Requirements

Derived from the Public and Private Groups blueprint.

**Governance Proposal:**
```swift
struct GovernanceProposal: Identifiable, Codable {
    let id: UUID
    let groupId: String
    let title: String
    let description: String
    let voteType: VoteType        // .yesNo, .multipleChoice(options: [String])
    let votingEndsAt: Date
    let quorumRequired: Double    // 0.1 to 0.75 (10% to 75%)
    let anonymousVoting: Bool
    var votes: [GovernanceVote]
    var status: ProposalStatus    // .active, .passed, .failed, .quorumNotMet

    var participationPct: Double { Double(votes.count) / Double(eligibleMemberCount) }
}
```

## Blueprints

- Public and Private Groups with Verified Status Display — Defines governance polls, voting weight configuration, quorum requirements, real-time results, blockchain anchoring, and voting history

---

### WO-150: Build Content Calendar and Planning Management System

**Blueprint:** Broadcast Channels and Community Features

## Summary

Build the content calendar and planning management system — a visual calendar interface for organizing scheduled posts, managing content series, and planning recurring campaigns. Builds on top of the scheduled posting system (WO-139) by adding calendar UI, series management, and template support.

## In Scope

- Calendar view: monthly/weekly/daily views of scheduled and published posts with iOS `EKEventStore`-style grid UI
- Drag-and-drop rescheduling: tap and drag a scheduled post to a new date/time slot
- Content series management: group related posts into named series with sequence tracking and cross-referencing
- Content series sequencing: auto-number posts in a series; subscribers see "Part 2 of 5" labels
- Draft management with version notes: save drafts with `{title, content, version, lastEditedAt, notes}`
- Template library: predefined post templates (announcement, poll, image post) with placeholder fields
- Bulk operations: select multiple scheduled posts → reschedule all, delete all, or move to different series
- Content performance planning: show historical engagement data inline in calendar for comparison when scheduling

## Out of Scope

- Individual post scheduling (WO-139)
- External calendar system integrations (iCal, Google Calendar)
- Team collaboration/multi-user workflows

## Requirements

Derived from the Broadcast Channels blueprint.

**Calendar Data Model:**
```swift
struct ContentCalendarEntry: Identifiable {
    let id: UUID
    let channelId: String
    var scheduledPost: ScheduledChannelPost
    var seriesName: String?        // Nil if standalone post
    var seriesSequence: Int?       // Position in series
    var engagementHint: Double?    // Historical avg engagement for this time slot
}
```

## Blueprints

- Broadcast Channels and Community Features — Defines content calendar interface, content series management, draft management, template library, and bulk scheduling

---

### WO-158: Implement Channel Analytics and Performance Tracking System

**Blueprint:** Broadcast Channels and Community Features

## Summary

Build the channel analytics and performance tracking dashboard — subscriber growth, post engagement metrics, content performance analysis, and peak activity patterns. All analytics are privacy-preserving (aggregated, anonymized — no individual subscriber tracking). Data retained for 12 months.

## In Scope

- Subscriber count chart: daily/weekly/monthly growth trend
- Post performance metrics: views (approximate, privacy-preserving), reactions per post, shares count
- Engagement rate: reactions / views per post, trend over time
- Top performing posts: ranked by engagement rate over selected date range
- Peak engagement times: hourly heatmap showing when subscribers are most active
- Subscriber tier distribution: breakdown of Tier 1–5 (anonymized aggregate)
- Export: `GET /v1/channels/{id}/analytics?from={date}&to={date}&format=csv`
- Real-time updates: key metrics updated within 5 minutes of user interactions
- 12-month data retention in PostgreSQL analytics tables
- Privacy: all metrics are aggregated server-side before returning — no individual DID tracking

## Out of Scope

- Individual subscriber identification
- External analytics platforms
- Predictive analytics

## Requirements

Derived from the Broadcast Channels blueprint.

**Analytics Response:**
```go
type ChannelAnalytics struct {
    SubscriberGrowth  []DailyMetric  // count per day
    PostEngagement    []PostMetric   // per post: views, reactions, engagement_rate
    PeakHours        [24]float64    // Activity index per hour of day (0-1)
    TierDistribution  map[int]int    // tier → approximate count
    TopPosts         []PostSummary  // Top 10 by engagement
    DateRange        DateRange
}
```

## Blueprints

- Broadcast Channels and Community Features — Defines subscriber count tracking, engagement metrics, content performance analytics, peak activity patterns, export, privacy-compliant reporting

---

### WO-168: Build Channel Roles and Governance Voting System

**Blueprint:** Broadcast Channels and Community Features

## Summary

Build channel roles (Owner/Admin/Moderator/Subscriber) and governance voting for channel policy decisions. Similar to group governance (WO-148) but scoped to broadcast channels. Governance polls use the same poll infrastructure from Message Reactions (WO-23). Results are optionally blockchain-anchored.

## In Scope

- Channel role hierarchy: Owner (full control), Administrator (moderation + settings), Moderator (content moderation), Subscriber (read + react)
- Role assignment: `POST /v1/channels/{id}/roles/{userDID}` — owner assigns admins, admins assign moderators
- Role permission matrix per role (similar to groups): configure which role can post, moderate, invite, change settings
- Governance poll creation: channel admin creates poll on policy/content direction using poll infrastructure (WO-23)
- Subscriber voting: eligible subscribers vote using existing poll voting mechanism
- Voting weight: configurable (equal = 1 person 1 vote; tenure-weighted; token-stake-weighted)
- Quorum requirements: configurable 10–75% of subscribers
- Results anchoring: optional Data L1 submission of `{type: "channel_governance", channelId, proposalId, result}`
- Voting history: archive of all governance decisions accessible in channel settings

## Out of Scope

- Liquid democracy / delegated voting
- Automated governance execution (admin reviews and manually implements)

## Requirements

Derived from the Broadcast Channels blueprint.

**Role API:**
```go
// POST /v1/channels/{id}/roles/{userDID}
type AssignRoleRequest struct {
    Role   string `json:"role"` // "administrator", "moderator"
    Reason string `json:"reason"`
}
```

## Blueprints

- Broadcast Channels and Community Features — Defines channel roles, role-based permissions, governance voting, voting weight configuration, quorum requirements, and voting transparency

---

### WO-175: Implement Subscriber Segmentation and Targeted Messaging System

**Blueprint:** Broadcast Channels and Community Features

## Summary

Build the subscriber segmentation and targeted messaging system — channel creators define segments (by engagement, subscription date, content interest, or manual assignment) and send targeted posts to specific segments. All analytics are privacy-preserving with aggregate-only reporting.

## In Scope

- Segment creation: `{segmentName, criteria: {subscriptionDateRange?, engagementLevel?, manualDIDs?}}`
- Automatic interest segmentation: track which content categories each subscriber engages with (react/reply to); group by top category (privacy: stored as aggregate pattern, not per-post history)
- Engagement levels: Active (≥3 interactions/week), Moderate (1–2/week), Inactive (<1/week) — auto-computed weekly
- Segment-based post: creator selects target segment when composing a post; only that segment receives it
- Segment analytics: size (subscriber count), engagement rate, growth — all aggregated, no individual tracking
- Segment preview: estimated reach before sending
- Manual segment assignment: creator can manually add/remove specific subscribers
- Maximum 20 segments per channel

## Out of Scope

- ML-based automatic segmentation
- External marketing automation (Mailchimp, etc.)
- Cross-channel subscriber sharing

## Requirements

Derived from the Broadcast Channels blueprint.

**Segment Model:**
```go
type SubscriberSegment struct {
    ID        string          `db:"id"`
    ChannelID string          `db:"channel_id"`
    Name      string          `db:"name"`
    Type      SegmentType     `db:"type"` // "automatic" or "manual"
    Criteria  SegmentCriteria `db:"criteria"`
    EstimatedSize int         `db:"estimated_size"`  // Updated hourly
}
```

## Blueprints

- Broadcast Channels and Community Features — Defines subscriber segmentation, segment-based messaging, interest-based segmentation, engagement-based segmentation, segment analytics, and privacy protection

---

### WO-185: Implement Filecoin Long-Term Storage Integration

**Assignee:** Chad Cromwell

**Blueprint:** Large File Sharing and Cloud Storage Integration

## Summary

Implement the Filecoin long-term storage integration as an optional tier of storage persistence for important files. Users can configure selected files to have Filecoin storage deals created, ensuring permanent availability beyond IPFS pinning TTL. Includes automated deal renewal, storage provider selection, and proof-of-storage verification.

## In Scope

- Filecoin storage deal creation via Estuary API or web3.storage Filecoin pipeline
- Automatic storage provider selection: rank providers by price, reputation score, uptime history
- Storage duration configuration: 6 months, 1 year, 2 years, permanent (with auto-renewal)
- Automated deal renewal: check deal expiry, initiate renewal 30 days before expiration
- Proof-of-storage verification: query Filecoin network for deal status and proof verification
- Retrieval verification: test retrieve first chunk from Filecoin to confirm accessibility
- Cost tracking: display deal cost (FIL equivalent), provider info, deal CID in file details
- Integration with IPFS: files already pinned on IPFS → create Filecoin deal pointing to same CID
- Fallback: if Filecoin deal creation fails, fall back to extended IPFS pinning

## Out of Scope

- IPFS short-term storage (WO-21)
- File encryption (WO-9)
- File expiration and deletion (WO-68)

## Requirements

Derived from the Large File Sharing blueprint.

**Filecoin Deal:**
```go
// backend/media/filecoin_client.go
type FilecoinDealRequest struct {
    FileCID      string  // IPFS CID of encrypted file manifest
    Duration     int64   // Storage duration in epochs (epochs ~30s, 518400 = 180 days)
    Replication  int     // Number of provider copies (default: 3)
    MaxPrice     string  // Max price per GiB/epoch (FIL)
}

func (c *FilecoinClient) CreateStorageDeal(cid string, durationDays int) (*FilecoinDeal, error) {
    // 1. Select providers by reputation and price
    providers := c.selectProviders(criteria: ProviderCriteria{MinUptime: 0.99, MaxPrice: maxPricePerGiB})
    // 2. Propose deal to each selected provider
    deal, err := c.estuary.MakeDeal(FilecoinDealRequest{
        FileCID:     cid,
        Duration:    int64(durationDays) * 2880,  // 2880 epochs/day
        Replication: len(providers),
    })
    return deal, err
}
```

## Blueprints

- Large File Sharing and Cloud Storage Integration — Defines Filecoin integration for long-term storage, automatic storage deals, deal duration configuration, retrieval verification, storage provider selection, and cost tracking

---

### WO-229: Implement Phase 4 Community Relay Node Registry on Data L1

**Type:** Build

**Blueprint:** Privacy-Preserving Blockchain Data Model

## Summary

Implement the Phase 4 community relay node registry on the Constellation Data L1, enabling community operators to register their relay nodes on-chain and clients to discover and rotate across them. Includes the on-chain registration schema, Scala validation, Go backend registration/discovery API, and iOS client-side relay node rotation.

## In Scope

- **Data L1 relay node entry schema:**
  ```plaintext
  RelayNodeEntry {
    nodeDID:        string     // T7
    endpointURL:    string     // T7 (WSS endpoint)
    echoStake:      uint64     // T7 (TokenLock amount — governance-set minimum)
    cloudProvider:  string     // T7: "aws", "digitalocean", "hetzner", "bare-metal"
    registeredAt:   timestamp  // T7
  }
  ```
- **Scala Data L1 validation:** Minimum ECHO stake required (governance-set); valid WSS endpoint format; cloud provider is not "aws" for community operators (AWS reserved for project-operated nodes); node DID signature required
- **Cloud diversity enforcement:** Registry governance tracks cloud provider distribution; reject community operator registration if > 60% of existing nodes are on the same provider
- **Go backend relay registry API:**
  - `POST /relay-nodes` — register node (requires staked ECHO + DID signature)
  - `GET /relay-nodes` — list active nodes with uptime/latency/cloud provider stats
  - `DELETE /relay-nodes/{nodeDID}` — deregister (operator or governance revocation)
- **iOS client node rotation:** On app launch, fetch active relay node list from backend; select 3 lowest-latency nodes; rotate connection across them per session; prefer non-AWS nodes in Phase 4
- **Slashing integration:** Node downtime > 1h → warning; > 4h → 1% stake slash; repeated violations → `DELETE /relay-nodes/{nodeDID}` (enforced by Metagraph L0)
- **Minimum activation threshold:** Federated relay mode activates when 5+ community relay nodes are registered and active

## Out of Scope

- Project-operated relay node deployment (WO for Phase 2 mainnet)
- ECHO staking mechanics (WO-127)
- Phase 1–3 relay infrastructure (existing WebSocket relay WOs)

## Requirements

From the Privacy-Preserving Blockchain Data Model blueprint:

**Relay Node Registry entry (T7 — fully public):**
| Field | Content |
|---|---|
| Node DID | Operator's DID |
| Endpoint URL | WSS relay endpoint |
| ECHO Stake | TokenLock amount |
| Cloud Provider | Must be non-AWS for community operators |
| Uptime (30d rolling) | Public metric |
| Traffic Served (30d) | Encrypted blob count (no content metadata) |

## Blueprints

- Privacy-Preserving Blockchain Data Model — Defines the Phase 4 relay node registry schema, on-chain data types, and cloud diversity requirement
- Data Layer — Defines relay node registration in Data L1 with minimum stake, cloud diversity thresholds, and slashing conditions

---

### WO-238: Build VIP Subscription Management System

**Type:** Build

**Blueprint:** ECHO Token Economics and Founder Allocation, Frontend

## Summary

Build the full VIP subscription management system in the iOS app — subscription purchase flow (monthly $9.99 / annual $99), `VIPSubscription` data model, `VIPFeatures` capability enforcement, `VIPBadge` animated component, and graceful downgrade handling. All subscription revenue flows to the community treasury. The free tier always retains full messaging, E2E encryption, blockchain anchoring, and token rewards — VIP adds convenience and status, never security.

## In Scope

- **`VIPSubscription` data model:**
  ```swift
  struct VIPSubscription {
      let status: VIPStatus               // active, pastDue, cancelled, expired
      let renewsAt: Date
      let monthlyPrice: Decimal           // $9.99
      let allowSpendID: String            // References on-chain AllowSpend approval
      let features: VIPFeatures
  }
  enum VIPStatus { case active, pastDue, cancelled, expired }
  struct VIPFeatures {
      let maxGroupSize: Int               // 100,000
      let cloudStorageGB: Int             // 20
      let dailyRewardCapECHO: Decimal     // 150 ECHO/day
      let maxScheduledMessageDays: Int    // 365
      let governanceBonusPct: Float       // 10% weight bonus
      let priorityRelay: Bool             // true
      let customThemes: Bool              // true
      let vipBadge: Bool                  // true
      let advancedBotCount: Int           // 10
  }
  ```
- **Subscription purchase flow:**
  - Settings → "Upgrade to VIP" → feature comparison screen (free vs. VIP capacity)
  - Monthly ($9.99/auto-renew) or annual ($99/yr, 10% discount) selection
  - `AllowSpend` authorization via Stargazer SDK: user approves time-limited ECHO or fiat allowance ("Allow ECHO app to charge up to $9.99/month — expires monthly")
  - Confirmation: VIP badge appears immediately after authorization
- **`VIPBadge` component:** Animated gold ring around user avatar; displayed inline in chats and profiles for VIP users
- **VIP feature enforcement:** Check `VIPSubscription.status` before unlocking capacity increases; real-time enforcement when subscription lapses
- **Graceful downgrade:** On subscription cancellation or expiry:
  - Groups the user created above 10K members → locked to new joins (not deleted) until VIP renewed or group size naturally drops below free-tier limit (10K)
  - Scheduled messages beyond 24h in the future → remain scheduled but cannot create new ones beyond 24h limit
  - Cloud storage above 2GB → read-only until VIP renewed or files deleted
  - Accumulated ECHO rewards and wallet balance are **never** affected by subscription status
- **Renewal management:** Past-due grace period (3 days); retry `SpendTransaction` against existing `AllowSpend`
- **Backend VIP status endpoint:** `GET /v1/subscription/status` → returns `{status, renewsAt, features}`; cached in Redis TTL: 5 minutes
- **Rate limit tier upgrade:** On active VIP, backend Contacts Service and Message Relay bump per-DID limits to VIP tier (WO-44)

## Out of Scope

- AllowSpend payment rails infrastructure (WO-206)
- App Store in-app purchase integration (VIP uses AllowSpend/ECHO native payment, not IAP)
- Bot framework and advanced bot management (covered by bot work orders)

## Requirements

From the Frontend blueprint (VIP Subscription Management section):

**Feature comparison: Free vs. VIP:**
| Feature | Free | VIP ($9.99/mo) |
|---|---|---|
| Max group size | 10,000 | 100,000 |
| Cloud storage | 2GB | 20GB |
| Daily reward cap | 100 ECHO | 150 ECHO |
| Disappearing messages | 24h max | 1 year max |
| Governance bonus | Standard | +10% weight |
| Advanced bots | 3 | 10 |

**Downgrade invariant:** Free tier always retains full messaging, E2E encryption, blockchain anchoring, and token rewards. VIP adds convenience and status, never security.

## Blueprints

- Frontend — Defines full `VIPSubscription` model, `VIPFeatures`, `VIPBadge` component, subscription flow, downgrade behavior, and VIP status management
- ECHO Token Economics and Founder Allocation — Defines `AllowSpend` primitive for time-limited subscription payment authorization

---
