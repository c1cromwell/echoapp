# Phase 3: Messaging Core

**Total Work Orders:** 37  
**Status Summary:** 24 Completed, 13 Backlog  
**Last synced with Software Factory:** 2026-05-29

---

## Backlog (36)

### WO-3: Implement Local Message Indexing Engine with Privacy-Preserving Metadata

**Blueprint:** Advanced Message Search and Archive System

## Summary

Build the local message indexing engine that creates an encrypted, queryable search index from decrypted message content on the user's device. All indexing occurs locally — no message content is transmitted externally. The index is incremental (new messages indexed on receipt), encrypted at rest, and serves as the foundation for all search features.

## In Scope

- Local inverted index built from decrypted message plaintext (tokenize → stem → store)
- Automatic index update on every received/sent message (within 5 seconds of receipt)
- Index storage: encrypted with AES-256-GCM using HKDF-derived key from Secure Enclave
- Incremental indexing: only index new/edited messages; no full rebuild needed
- Index size budget: maximum 10% of available device storage, evict oldest entries when exceeded
- Index schema: `{term → [(messageId, conversationId, timestamp, fieldType)]}` inverted index
- Privacy metadata extraction: index sender DID, conversation ID, timestamp, content type — but not raw contact names
- Index corruption recovery: detect via integrity hash on index file, trigger rebuild on corruption
- Index backup to user-controlled location (IPFS, iCloud Keychain-protected)

## Out of Scope

- Search UI (WO-16)
- Semantic/NLP indexing (WO-41)
- Cross-device index sync (WO-73)
- Cloud-based indexing

## Requirements

Derived from the Advanced Message Search and Archive System blueprint.

**Index Architecture:**
```swift
// Core/Search/LocalMessageIndexer.swift
actor LocalMessageIndexer {
    private let encryptedIndex: EncryptedIndexStore
    private let tokenizer: MessageTokenizer

    func indexMessage(_ message: Message) async {
        // 1. Tokenize plaintext: lowercase, remove stopwords, stem
        let tokens = tokenizer.tokenize(message.plaintext)
        // 2. Add to inverted index with position and metadata
        for token in tokens {
            await encryptedIndex.addEntry(IndexEntry(
                term: token,
                messageId: message.id,
                conversationId: message.conversationId,
                timestamp: message.timestamp,
                senderDID: message.senderDID,
                contentType: message.contentType
            ))
        }
    }

    // Called on app launch to catch up on any unindexed messages
    func indexPendingMessages() async {
        let unindexed = await database.fetchUnindexedMessages(since: lastIndexedAt)
        for message in unindexed { await indexMessage(message) }
    }
}
```

## Blueprints

- Advanced Message Search and Archive System — Defines local indexing architecture, privacy-preserving metadata extraction, index encryption, size limits, and backup requirements

---

### WO-4: Implement Core Message Encryption and Delivery System

**Blueprint:** Blockchain-Anchored Messaging with Provable Integrity

## Summary

Build the core Go backend message relay service (Message Relay, port 8002) that transports end-to-end encrypted message blobs between clients via WebSocket. The relay is content-blind — it validates P-256 signatures and routes encrypted blobs but cannot read, decrypt, or modify message content. This includes WebSocket session management, online delivery, offline queueing up to 1000 messages per recipient, and APNs push notifications. Overflow beyond 1000 messages is handled by WO-237.

## In Scope

- WebSocket server with sticky load balancing for per-user session management
- Incoming message validation: verify P-256 ECDSA signature against sender's cached DID public key
- Online delivery: forward encrypted blob via recipient's WebSocket connection if connected
- Offline queueing: store encrypted blobs in Redis (fast) with PostgreSQL fallback (durable)
- APNs push notification on offline queue insertion (conversation ID only, no content)
- Queue depth: up to 1000 messages per recipient in standard queue; trigger overflow path (WO-237) when exceeded
- Retention policy: 30 days for 1:1 chats, 7 days for large groups (100+ members)
- Commitment hash extraction from message and forwarding to `AnchoringBatcher` (WO-15)
- `queueDrain` message delivery on recipient WebSocket reconnect
- NATS pub/sub integration for cross-pod group message fan-out
- Message delivery ACKs and status tracking (sent → delivered)
- `EncryptedPayload` deserialization for signature validation (commitment field in payload per WO-207 canonical spec)

## Out of Scope

- Message encryption/decryption (done on device — canonical spec in WO-207)
- Commitment anchoring batch submission (WO-15)
- APNs server certificate management
- Group key distribution (covered by Groups work orders)
- Queue overflow IPFS backup when > 1000 messages (WO-237)

## Requirements

From the Blockchain-Anchored Messaging and Backend blueprints. The relay receives the canonical `EncryptedPayload` (defined in End-to-End Message Encryption blueprint, implemented in WO-207) — the `commitment` field is extracted and forwarded to `AnchoringBatcher`.

**Message Relay Flow (Go):**
```go
func (s *RelayService) HandleMessage(senderConn *WSConnection, msg IncomingMessage) error {
    if !s.verifySignature(msg.Payload, msg.Signature, msg.SenderDID) {
        return ErrInvalidSignature
    }
    if err := s.rateLimiter.Check(msg.SenderDID, msg.Type); err != nil {
        return err
    }
    // Extract commitment from EncryptedPayload and forward to batcher
    s.anchoringBatcher.AddCommitment(msg.MessageID, msg.CommitmentHash)
    recipientConn, online := s.connections.Get(msg.RecipientDID)
    if online {
        recipientConn.Send(msg.EncryptedPayload)
        return s.updateStatus(msg.MessageID, StatusDelivered)
    }
    // Check queue depth before enqueuing — overflow handled by WO-237
    if s.offlineQueue.Depth(msg.RecipientDID) >= 1000 {
        return s.overflowToIPFS(msg)  // WO-237
    }
    return s.offlineQueue.Enqueue(msg.RecipientDID, msg.EncryptedPayload, msg.MessageID)
}
```

## Blueprints

- Blockchain-Anchored Messaging with Provable Integrity — Defines relay architecture, content-blind transport, offline queue design, delivery statuses, and APNs integration
- Backend — Specifies Message Relay service (port 8002), NATS pub/sub fan-out, Redis/PostgreSQL storage, and sticky WebSocket load balancing
- End-to-End Message Encryption and Commitment — Defines canonical `EncryptedPayload` struct the relay receives (commitment field extracted for anchoring)

---

### WO-10: Implement Emoji Reaction System with Real-time Synchronization

**Blueprint:** Message Reactions, Polls, and Interactive Elements

## Summary

Build the emoji reaction system — users react to messages with standard Unicode emojis. Reactions are transmitted as separate encrypted message blobs (type `.reaction`) via the relay, generating their own commitment hashes anchored in Merkle batches. Reaction counts are aggregated and display in real-time via WebSocket.

## In Scope

- Emoji picker UI: scrollable grid of Unicode emoji categories, recently used section, search
- Multiple reactions per user per message (e.g., user can add 👍 and ❤️ to the same message)
- Tap existing reaction to add/remove (toggle behavior for user's own reactions)
- Reaction count aggregation display: `👍 3 ❤️ 2` below message bubble
- Tap reaction count to see who reacted with each emoji (shows trust tier badge per reactor)
- Real-time sync: reaction message relayed via WebSocket to all participants within 2 seconds
- Backend: relay reaction blob as `MessageType.reaction` with `parentMessageId` — same relay pipeline as text
- Commitment hash generated for each reaction (anchored in standard Merkle batch)
- Reaction notifications: optional push notification when user's message receives a reaction

## Out of Scope

- Custom emoji creation (WO-60 NFT emojis)
- Reaction-based token rewards
- Rich media reactions (WO-50)

## Requirements

Derived from the Message Reactions, Polls, and Interactive Elements blueprint.

**Reaction Message Type:**
```swift
// Message content type for reactions
enum MessageContentType: String, Codable {
    // ...
    case reaction  // Separate message object linked to parent
}

struct ReactionPayload: Codable {
    let parentMessageId: String   // Message being reacted to
    let emoji: String             // Unicode emoji string e.g. "👍"
    let action: ReactionAction    // .add or .remove
}

// Displayed in chat:
struct MessageReactionSummary {
    let emoji: String
    let count: Int
    let reactorDIDs: [String]   // Shown when user taps count
    let currentUserReacted: Bool
}
```

## Blueprints

- Message Reactions, Polls, and Interactive Elements — Defines emoji library, multiple reactions, reaction counts, real-time sync, and encryption requirements

---

### WO-16: Build Keyword Search Engine with Fuzzy Matching and Result Ranking

**Blueprint:** Advanced Message Search and Archive System

## Summary

Build the keyword search engine against the local encrypted message index — full-text search with fuzzy matching (≤2 character differences), partial word matching, boolean operators, and result ranking by relevance + recency. Results are highlighted with matched terms shown in context. Built on top of the local index from WO-3.

## In Scope

- Full-text keyword search against local inverted index (WO-3)
- Case-insensitive matching
- Fuzzy matching: Levenshtein distance ≤ 2 for query terms ≥ 5 characters
- Partial word matching: match from 3+ character prefix
- Boolean operators: AND (default), OR, NOT, phrase search (`"exact phrase"`)
- Result ranking: relevance score (term frequency × inverse document frequency) × recency decay
- Keyword highlighting in results: show matched term with 50 chars context on each side
- Search history: last 100 queries stored locally in SwiftData
- Performance: results within 2 seconds for index of up to 100,000 messages
- Search suggestions based on past history (debounced, 300ms delay)

## Out of Scope

- Advanced filters (WO-29)
- Semantic/NLP search (WO-41)
- Search result sharing (WO-83)
- Cloud search

## Requirements

Derived from the Advanced Message Search and Archive System blueprint.

**Search Implementation:**
```swift
// Core/Search/KeywordSearchEngine.swift
actor KeywordSearchEngine {
    private let index: LocalMessageIndexer

    func search(query: String, limit: Int = 50) async -> [SearchResult] {
        let parsed = parseQuery(query)  // Handle AND/OR/NOT/phrases

        var candidates: [[String: SearchResult]] = []
        for term in parsed.requiredTerms {
            let exactMatches = await index.lookup(term: term.lowercased())
            let fuzzyMatches = await index.fuzzyLookup(term: term, maxDistance: 2)
            candidates.append(merge(exactMatches, fuzzyMatches))
        }

        // Intersect/union based on boolean logic
        let results = applyBooleanLogic(candidates, parsed.logic)

        // Rank: TF-IDF × recency decay
        return results
            .sorted { rankScore($0) > rankScore($1) }
            .prefix(limit)
            .map { $0.withHighlighting(query: query, contextChars: 50) }
    }

    private func rankScore(_ result: SearchResult) -> Double {
        let tfidf = result.termFrequency / log(Double(totalDocs) / Double(result.docFrequency))
        let recencyDecay = exp(-0.001 * Date().timeIntervalSince(result.message.timestamp))
        return tfidf * recencyDecay
    }
}
```

## Blueprints

- Advanced Message Search and Archive System — Defines keyword search, fuzzy matching, result ranking, boolean operators, and highlighting requirements

---

### WO-17: Build Identity Verification System with Digital ID and Third-Party Integration

**Blueprint:** Frontend

## Summary

Implement the iOS identity verification UI flows — Apple Digital ID integration for iOS 17+ devices, third-party IDV service integration (Prove, Daon, 1Kosmos, Darwinium), government ID document scanning, and selfie-based liveness detection. Successful verification issues a high-assurance Verifiable Credential on Cardano and triggers the 100 ECHO reward.

## In Scope

- Apple Digital ID API integration (`AuthenticationServices` framework, iOS 17+) with graceful fallback when unavailable
- Third-party IDV service coordination: Prove (device verification), Daon (liveness), 1Kosmos (biometric), Darwinium (fraud assessment)
- `VerifyIdentityUseCase` and `ZKProofUseCase` in `Domain/UseCases/Identity/`
- Government ID document scanning view: camera capture, image quality check, document type selection (passport, driver's license, national ID)
- Selfie capture view with liveness detection guidance
- Verification progress view with status (submitted, processing, approved, failed)
- Error handling and retry flow with clear user guidance on failure reasons
- Trust tier elevation display after successful verification
- 100 ECHO reward notification display after verification
- Zero-knowledge proof generation for credential presentation (Phase 3+ via Midnight SDK)

## Out of Scope

- Backend IDV service coordination (Go backend work orders)
- Cardano credential issuance (backend work orders)
- ECHO token distribution (Rewards backend work order)
- Verification service API keys and backend setup

## Requirements

Derived from the Frontend and In-App High-Assurance Identity Verification blueprints.

**Apple Digital ID Flow (iOS 17+):**
```swift
import AuthenticationServices

// Domain/UseCases/Identity/VerifyIdentityUseCase.swift
struct VerifyIdentityUseCase {
    func verifyWithAppleDigitalID() async throws -> VerificationResult {
        // 1. Check ASAuthorizationController availability
        // 2. Request Digital ID credentials via ASAuthorizationDigitalIDProvider
        // 3. On success, submit credential reference to backend
        // 4. Backend coordinates with Apple, issues Cardano VC
        return .success(method: .appleDigitalID)
    }

    func verifyWithDocumentScan(documentType: DocumentType) async throws -> VerificationResult {
        // 1. Open camera for ID capture
        // 2. Process image quality checks locally
        // 3. Submit to third-party IDV (never to ECHO backend directly)
        // 4. IDV provider returns pass/fail reference UUID
        // 5. Backend receives callback from IDV provider
        return .success(method: .documentScan)
    }
}

enum DocumentType { case passport, driversLicense, nationalID }
enum VerificationMethod { case appleDigitalID, documentScan, darwiniumFraudAssessment }
```

**IDV Service Selection Logic:**
```swift
struct IDVServiceSelector {
    func selectService(for device: DeviceInfo) -> IDVService {
        if device.supportsAppleDigitalID && device.osVersion >= 17 {
            return .appleDigitalID
        }
        // Fallback priority: Daon → Prove → 1Kosmos
        if device.country == .us { return .prove }  // US: Prove for device verification
        return .daon                                  // International: Daon
    }
}

// Darwinium fraud assessment runs in parallel for all verification paths
```

**ZK Proof (Phase 3+ — Midnight SDK):**
```swift
// Domain/UseCases/Identity/ZKProofUseCase.swift
struct ZKProofUseCase {
    // "Prove I'm Tier 3+ without revealing my score"
    func proveMinimumTrustTier(_ tier: Int) async throws -> ZKProof { ... }
    // "Prove I'm 18+ without revealing birthdate"
    func proveAgeOver18() async throws -> ZKProof { ... }
}
```

## Blueprints

- Frontend — Defines identity verification use cases, Apple Digital ID integration, third-party IDV coordination, and ZK proof generation
- In-App High-Assurance Identity Verification and Reward — Specifies the verification flow, NIST 800-63-3 IAL2 compliance, third-party service integration, and 100 ECHO reward distribution

---

### WO-23: Build Poll Creation and Voting System with Zero-Knowledge Privacy

**Blueprint:** Message Reactions, Polls, and Interactive Elements

## Summary

Build the poll creation and anonymous voting system. Polls are sent as encrypted message blobs in conversations. Votes are encrypted and submitted anonymously; results are tallied using ZK proofs that prove vote integrity without revealing individual voter identities. Results update in real-time.

## In Scope

- Poll creation UI: question text, 2–10 options, optional time limit (1h to 30 days)
- Poll message type: `MessageContentType.poll` with serialized `PollPayload`
- E2E encrypted poll distribution (same relay pipeline as messages)
- Anonymous vote submission: vote encrypted before relay; ZK proof proves valid vote without revealing choice
- Real-time result updates via WebSocket when new votes arrive
- Vote change support: user can change vote before poll closes (previous vote invalidated via ZK proof)
- Result visualization: progress bars per option with count and percentage
- Poll expiry: automatic closure at configured time; closed polls show final results
- Results export: CSV/JSON export of final aggregated results (no individual voter data)
- Trust-score-based voting weight (optional, per blueprint): high-trust users can have more weight in community polls

## Out of Scope

- Governance smart contract integration (separate governance work orders)
- Ranked choice or matrix question types (v1 only: multiple choice)

## Requirements

Derived from the Message Reactions, Polls, and Interactive Elements blueprint.

**Poll Data Model:**
```swift
struct PollPayload: Codable {
    let pollId: UUID
    let question: String
    let options: [PollOption]   // 2–10 options
    let closesAt: Date?         // nil = open indefinitely
    let allowVoteChange: Bool
    let weightedByTrustScore: Bool  // Optional trust-weighted voting

    struct PollOption: Codable {
        let id: UUID
        let text: String
        var encryptedVoteCount: Data  // ZK-aggregated count, updated on each vote
    }
}

// Vote submission (anonymous):
struct PollVotePayload: Codable {
    let pollId: UUID
    let optionId: UUID
    let zkProof: Data?         // ZK proof of valid vote (Phase 3+)
    // Voter identity not included in payload; relay cannot link vote to voter DID
}
```

## Blueprints

- Message Reactions, Polls, and Interactive Elements — Defines multiple choice polls, anonymous voting, ZK proof vote tallying, time limits, vote change, result visualization, and export

---

### WO-25: Implement Message Editing with Immutable History Tracking

**Blueprint:** Blockchain-Anchored Messaging with Provable Integrity

## Summary

Implement message editing on both backend and iOS with the constraint that the original on-chain commitment is immutable. Edits generate new commitment hashes anchored in subsequent Merkle batches. Full local edit history is maintained on-device. The blockchain always records "a message existed" while the edit history is device-local.

## In Scope

- Backend: `PUT /v1/messages/{messageId}` endpoint accepting edited encrypted payload + new commitment + P-256 signature; validate 24-hour edit window; relay encrypted edit to recipient; add new commitment to anchoring batch
- iOS: edit message UI with 24-hour countdown indicator; local edit history storage in SwiftData; "edited" timestamp indicator in chat bubble; edit history view accessible via long-press
- Edit notification to recipients: special WebSocket message type `messageEdit` with message ID, new encrypted payload, edit timestamp, edit count
- Original commitment preservation: on-chain Merkle root for original message UNCHANGED; new commitment is generated for the edited version and anchored in next batch
- Edit count and "last edited at" timestamp display

## Out of Scope

- Message deletion
- Bulk editing
- Group admin edit override permissions
- Edit history cross-device sync (local only)

## Requirements

Derived from the Blockchain-Anchored Messaging blueprint.

**Edit Immutability Principle:**
```
Original on-chain Merkle root → UNCHANGED (proves original message existed)
Edit creates NEW commitment  → Anchored in next 5-min batch (proves edit existed)
Local edit history           → Stored in SwiftData (device-local only)
Edit metadata                → NOT stored on-chain (privacy by design)
```

**Backend Edit Endpoint:**
```go
// PUT /v1/messages/:messageId
type EditMessageRequest struct {
    EncryptedPayload []byte `json:"encrypted_payload"`  // Re-encrypted edited content
    NewCommitment    []byte `json:"new_commitment"`      // H(H(newPlaintext) || nonce)
    Signature        []byte `json:"signature"`           // P-256 sig over new payload
    EditTimestamp    int64  `json:"edit_timestamp"`
}

func (h *MessageHandler) EditMessage(messageID string, req EditMessageRequest) error {
    // 1. Verify sender is original message author
    // 2. Verify 24-hour edit window not expired
    // 3. Verify P-256 signature
    // 4. Relay edited encrypted payload to recipient
    // 5. Add new commitment to AnchoringBatcher
    // 6. Return edit confirmation with edit count
}
```

**iOS Edit History:**
```swift
struct MessageEdit: Codable {
    let editedAt: Date
    let editCount: Int
    // NOTE: previous plaintext NOT stored after display (privacy)
    // Only timestamp and count retained for UI
}

// Chat bubble display:
// - "edited" indicator shows if editCount > 0
// - Tap "edited" → shows edit history: [original send time, edit timestamps]
// - Original commitment hash available for verification (anchored on-chain)
```

## Blueprints

- Blockchain-Anchored Messaging with Provable Integrity — Defines edit window (24h), original commitment immutability, new commitment generation per edit, local edit history storage, and edit notification behavior

---

### WO-26: Integrate Third-Party Identity Verification Services

**Blueprint:** Decentralized Identity and Authentication

## Summary

Build the third-party identity verification services integration on the backend — coordinates with Prove, Daon, 1Kosmos, and Darwinium via their APIs, orchestrates the verification callback flow, triggers VC issuance via the credential service, and initiates the 100 ECHO reward distribution. This is the backend orchestration layer for all in-app identity verification flows.

## In Scope

- IDV provider coordinator: routes to appropriate provider based on verification type and device capabilities
  - Prove: device trust verification (US primary)
  - Daon: liveness detection and face matching
  - 1Kosmos: biometric verification
  - Darwinium: fraud risk assessment (always runs in parallel)
- Webhook callback handling: `POST /v1/identity/verification/callback` — validates HMAC, maps reference UUID to user DID, determines tier
- Verification session management: create session with `{sessionId, userDID, provider, startedAt, expiresAt}` stored in Redis (15-minute expiry)
- Verification result processing: on pass → trigger VC issuance (WO-132) + trust tier elevation (WO-37) + 100 ECHO reward (WO-184)
- Retry tracking: max 5 attempts per 24h (WO-159 rate limit enforcement)
- Credential type mapping: Prove/Darwinium → Proof of Humanity; Daon + doc scan → KYC-Lite; all + high confidence → High-Assurance
- Audit log: `{sessionId, provider, result, referenceUUID, userDID_hash, timestamp}` — no PII

## Out of Scope

- Document image processing (done by IDV provider SDK on device)
- iOS verification UI (WO-104, WO-113, WO-17)
- VC issuance (WO-132)
- Trust tier updates (WO-37)

## Requirements

Derived from the Decentralized Identity and Authentication blueprint.

```go
type VerificationOrchestrator struct {
    prove      *prove.Client      // US device verification
    daon       *daon.Client       // Liveness
    oneKosmos  *onekosmos.Client  // Biometric
    darwinium  *darwinium.Client  // Fraud assessment (always)
}
```

## Blueprints

- Decentralized Identity and Authentication — Defines third-party verification service integration, provider selection, callback handling, and credential issuance coordination

---

### WO-28: Develop Comprehensive Messaging System with Advanced Features

**Blueprint:** Frontend

## Summary

Build the comprehensive iOS messaging UI system — all message send/receive UI, the `MessageRelayManager` actor for E2E encrypted relay operations, `AnchoringTracker` for on-chain status updates, and `GroupKeyManager` for group symmetric key lifecycle. This work order covers the complete messaging experience from composing a message to displaying its anchored status.

## In Scope

- `MessageRelayManager` actor: full send pipeline (encrypt → sign → relay) and receive pipeline (verify signature → decrypt → display)
- `AnchoringTracker` class: track pending message commitments, receive `confirmation` WebSocket events, update message status to `.anchored`
- `GroupKeyManager` actor: generate group symmetric keys, encrypt keys per-member, rotate on membership changes
- `WebSocketRelay` handling for `message`, `queueDrain`, `typing`, `presence`, `receipt`, `ack`, `confirmation`, `groupKey` message types
- `DeliveryStatus` enum: `sending`, `sent`, `delivered`, `read`, `failed`, `anchored`, `verified`
- Chain-link icon 🔗 in message UI for `.anchored` status; Smart Checkmark ✓ badge for `.verified` (Organization tier)
- Chat view: message bubbles, edit indicator, pin indicator, reply quote, reaction summary
- Typing indicator: real-time display with 5-second timeout
- Read receipts: delivery and read status per message
- Audio messages: Opus codec recording up to 5 minutes, waveform visualization, playback with speed control (0.75x–1.5x)
- Message editing: 24-hour window, "edited" indicator, local edit history
- Message pinning: up to 10 per conversation, local-only storage
- Message forwarding: new E2E encryption for destination, optional attribution

## Out of Scope

- Backend relay service (Go backend work orders)
- On-chain Merkle anchoring logic (Go backend WO-15)
- Disappearing messages UI (Phase 4 work orders)
- Hidden folders (Phase 4 work orders)
- Push notification implementation (WO-57)
- Group creation and management UI (Groups work orders)

## Requirements

Derived from the Frontend and Blockchain-Anchored Messaging blueprints.

**Message Send Pipeline:**
```swift
actor MessageRelayManager {
    func sendMessage(
        plaintext: Data,
        contentType: Message.ContentType,
        recipientPublicKey: Data,
        conversationId: String
    ) async throws -> Message {
        // 1. E2E encrypt (X25519 + ChaCha20-Poly1305)
        let encryptedPayload = try encryption.encrypt(plaintext: plaintext, recipientPublicKey: recipientPublicKey)
        // 2. Sign encrypted payload with Secure Enclave (P-256)
        let signature = try await secureEnclave.sign(data: encryptedPayload.serialized, reason: "Send message")
        // 3. Submit to relay via WebSocket
        let messageId = UUID().uuidString
        let response = try await webSocket.sendMessage(SendMessageRequest(messageId: messageId, conversationId: conversationId, contentType: contentType, encryptedPayload: encryptedPayload, signature: signature))
        // 4. Track commitment for anchoring
        anchoringTracker.track(messageId: messageId, commitment: encryptedPayload.commitment)
        return Message(id: messageId, status: response.status == "relayed" ? .delivered : .sent)
    }
}
```

**Delivery Status Display:**
```swift
enum DeliveryStatus: String, Codable {
    case sending    // ⏳ Encrypting / queued locally
    case sent       // ✓  Accepted by relay
    case delivered  // ✓✓ Delivered to recipient device
    case read       // ✓✓ (blue) Recipient opened message
    case failed     // ❌ Relay rejected
    case anchored   // 🔗 In finalized metagraph snapshot
    case verified   // ✓  Digital Evidence fingerprint (Org tier)
}
```

**AnchoringTracker:**
```swift
@MainActor final class AnchoringTracker: ObservableObject {
    @Published private(set) var pendingAnchors: [String: PendingAnchor] = [:]

    func confirmAnchoring(messageId: String, snapshotHash: String, snapshotHeight: Int, merkleProof: [Data]?) {
        pendingAnchors.removeValue(forKey: messageId)
        NotificationCenter.default.post(name: .messageAnchored, object: nil, userInfo: ["messageId": messageId, "snapshotHash": snapshotHash])
    }
}
```

## Blueprints

- Frontend — Defines `MessageRelayManager`, `AnchoringTracker`, `GroupKeyManager`, `DeliveryStatus`, WebSocket message types, and Digital Evidence bridge
- Blockchain-Anchored Messaging with Provable Integrity — Defines the commitment anchoring flow, delivery status semantics, and Merkle proof verification

---

### WO-29: Develop Advanced Search Filters for Date, Sender, and Content Type

**Blueprint:** Advanced Message Search and Archive System

## Summary

Build the advanced search filter UI and filter logic — date range, sender identity, conversation context, and content type filters that narrow keyword and semantic search results. Filters are composable (combine multiple filters), apply in real-time, and can be saved as named presets. Built on top of the local index (WO-3) and keyword search (WO-16).

## In Scope

- Date range filter: calendar picker with presets (last 7 days, 30 days, 90 days, 1 year, all time)
- Sender filter: contact search with autocomplete showing trust tier badge, include/exclude toggle
- Conversation filter: search within a specific conversation context
- Content type filter: text, images, files, links, voice notes, reactions
- Trust score filter: filter by sender's minimum trust tier (Tier 1–5)
- Multi-filter combination with AND logic (all filters applied simultaneously)
- Filter preset saving: save + name up to 10 filter combinations for quick reuse
- Real-time result update: re-execute search with new filter within 1 second of filter change
- Filter state persists within a search session; reset on new search or explicit clear

## Out of Scope

- Basic keyword search (WO-16)
- Semantic search (WO-41)
- Archive management (WO-54)

## Requirements

Derived from the Advanced Message Search and Archive System blueprint.

**Filter Model:**
```swift
struct SearchFilters: Equatable {
    var dateRange: DateInterval?
    var senderDIDs: Set<String>     // Empty = all senders
    var conversationId: String?
    var contentTypes: Set<MessageContentType>  // Empty = all types
    var minimumTrustTier: Int       // 1–5, default 1
    var includeArchived: Bool
}

// Preset storage:
struct SearchFilterPreset: Identifiable, Codable {
    let id: UUID
    var name: String
    var filters: SearchFilters
}
// Stored in SwiftData, max 10 presets
```

**Filter Application:**
```swift
func applyFilters(_ results: [SearchResult], filters: SearchFilters) -> [SearchResult] {
    results.filter { result in
        let message = result.message
        let passesDates = filters.dateRange == nil || filters.dateRange!.contains(message.timestamp)
        let passesSender = filters.senderDIDs.isEmpty || filters.senderDIDs.contains(message.senderDID)
        let passesType = filters.contentTypes.isEmpty || filters.contentTypes.contains(message.contentType)
        let passesTier = message.senderTrustTier >= filters.minimumTrustTier
        return passesDates && passesSender && passesType && passesTier
    }
}
```

## Blueprints

- Advanced Message Search and Archive System — Defines date range, sender, conversation, content type, and trust score filters; filter combination; preset saving; and real-time updates

---

### WO-36: Develop Interactive Button System for Quick Actions

**Blueprint:** Message Reactions, Polls, and Interactive Elements

## Summary

Build the interactive button message type that lets users embed up to 6 action buttons in a message. Button actions include: quick reply text, calendar event creation, and external URL deep links. All button interaction data is E2E encrypted. Creators see aggregated click stats.

## In Scope

- `MessageContentType.interactive` with `InteractivePayload` containing buttons array
- Button configuration: label (max 50 chars), action type (`.quickReply`, `.calendarEvent`, `.externalURL`)
- Quick reply: tap button → automatically send predefined text response as new message in conversation
- Calendar event: tap → open iOS calendar picker; confirm → create event in device Calendar via `EventKit`
- External URL: tap → confirmation prompt → open URL in `SFSafariViewController`
- Click tracking: when user taps button, send `{type: "button_interaction", buttonId, messageId}` via WebSocket
- Aggregated click stats for message creator (no individual identity linked to clicks)
- Button UI: rendered below message bubble, styled as tappable cards

## Out of Scope

- Payment processing or e-commerce buttons
- Complex conditional button workflows
- Bot-generated button messages (Bot Framework work orders)

## Requirements

Derived from the Message Reactions, Polls, and Interactive Elements blueprint.

**Interactive Payload:**
```swift
struct InteractivePayload: Codable {
    let buttons: [InteractiveButton]  // 1–6 buttons

    struct InteractiveButton: Codable, Identifiable {
        let id: UUID
        let label: String           // Max 50 chars
        let action: ButtonAction
        var clickCount: Int         // Aggregated (not per-user)

        enum ButtonAction: Codable {
            case quickReply(text: String)
            case calendarEvent(title: String, dateHint: Date?)
            case externalURL(url: URL)
        }
    }
}
```

## Blueprints

- Message Reactions, Polls, and Interactive Elements — Defines 1–6 buttons, quick response/calendar/URL actions, click tracking, aggregated stats, and E2E encryption requirements

---

### WO-41: Implement Local Semantic Search with Privacy-Preserving NLP Processing

**Blueprint:** Advanced Message Search and Archive System

## Summary

Build the on-device semantic search engine using CoreML and Natural Language framework for concept-based message retrieval. Users find messages by meaning ("find messages where Alice mentioned the meeting") not just exact keywords. All NLP processing runs locally — no content leaves the device. Built on top of the local index from WO-3 and integrated with the keyword search UI.

## In Scope

- On-device NLP using `NaturalLanguage.framework` for intent recognition and synonym expansion
- Sentence embedding using a CoreML model (BERT-based, quantized for iOS) for semantic similarity
- Semantic search: compute embedding for query, find messages with high cosine similarity scores
- Intent recognition: classify query intent (information lookup, file search, conversation review) to adjust ranking
- Synonym expansion: expand query terms using `NLEmbedding` synonym API
- Context-aware results: consider surrounding messages when scoring relevance
- Related message suggestions: show semantically similar messages in "You might also want" section
- Query refinement recommendations: suggest alternative search terms based on semantic analysis
- Performance target: results within 5 seconds for 100K message histories
- On-device model storage: CoreML model bundled with app, ~50MB

## Out of Scope

- External AI API calls (requirement: no content leaves device)
- Cross-device model sync
- Cloud-based NLP processing
- Search result sharing (WO-83)

## Requirements

Derived from the Advanced Message Search and Archive System blueprint.

**Semantic Search Pipeline:**
```swift
// Core/Search/SemanticSearchEngine.swift
actor SemanticSearchEngine {
    private let embedding: CoreMLEmbedding   // BERT-based, on-device
    private let nlpProcessor: NLProcessor

    func search(query: String, limit: Int = 20) async -> [SearchResult] {
        // 1. Expand query with synonyms
        let synonyms = nlpProcessor.getSynonyms(query)
        let expandedQuery = query + " " + synonyms.joined(separator: " ")

        // 2. Generate query embedding
        let queryEmbedding = await embedding.encode(expandedQuery)

        // 3. Compare against pre-indexed message embeddings
        let candidates = await messageEmbeddingIndex.findNearest(
            to: queryEmbedding,
            topK: limit * 2,  // Retrieve 2x for re-ranking
            threshold: 0.6    // Minimum cosine similarity
        )

        // 4. Re-rank considering context and recency
        return candidates
            .sorted { semanticScore($0, queryEmbedding) > semanticScore($1, queryEmbedding) }
            .prefix(limit)
            .map { SearchResult(message: $0.message, score: $0.similarity) }
    }
}
```

## Blueprints

- Advanced Message Search and Archive System — Defines semantic search with local NLP, concept-based matching, intent recognition, synonym matching, context-aware results, and privacy-preserving processing

---

### WO-48: Implement Offline Support and Data Synchronization System

**Blueprint:** Frontend

## Summary

Implement the offline message queue and data synchronization system that allows the app to function when the relay server is unavailable. The `OfflineQueueManager` stores outbound encrypted messages locally and `MessageRelayManager` automatically drains the queue on reconnect. The relay server also drains queued inbound messages from the server-side offline queue on reconnect.

## In Scope

- `OfflineQueueManager` — local outbox for encrypted outbound messages, stored in SwiftData (encrypted)
- `MessageRelayManager.drainOfflineQueue()` — automatic drain on WebSocket reconnect
- WebSocket reconnect logic with exponential backoff
- Handling of incoming `queueDrain` WebSocket message type (server draining its offline queue for this user)
- `WebSocketRelay` reconnect handler that calls `drainOfflineQueue()` on connection restore
- Local cache management with size limits and cleanup of old entries
- `CacheManager` setup for conversation list, contact list, and user profile caching
- Offline mode indicator in UI (network status, sync progress)
- Conflict resolution: server state wins for already-delivered messages; local outbox wins for pending sends

## Out of Scope

- Backend offline queue storage (Go backend work order)
- Real-time messaging send/receive (WO-28)
- Background app refresh scheduling (iOS system background tasks - separate concern)
- Cross-device synchronization (separate work order)

## Requirements

Derived from the Frontend blueprint.

**Offline Queue Flow:**
```swift
// Core/Relay/OfflineQueueManager.swift
actor OfflineQueueManager {
    private let database: LocalDatabase

    // Queue an encrypted message when relay is unavailable
    func enqueue(_ request: SendMessageRequest) throws {
        // Serialize and store in SwiftData (encrypted at rest)
        try database.saveQueuedMessage(request)
    }

    // Retrieve all pending messages for drain attempt
    func dequeueAll() -> [SendMessageRequest] {
        return database.fetchAllQueuedMessages()
    }

    func remove(_ messageId: String) {
        database.deleteQueuedMessage(messageId)
    }
}
```

**Reconnect + Drain Flow:**
```swift
// Core/Relay/MessageRelayManager.swift
actor MessageRelayManager {
    private let offlineQueue: OfflineQueueManager

    func drainOfflineQueue() async {
        let queuedMessages = offlineQueue.dequeueAll()
        for request in queuedMessages {
            do {
                _ = try await webSocket.sendMessage(request)
                offlineQueue.remove(request.messageId)
            } catch {
                // Leave in queue, will retry on next reconnect
                try? offlineQueue.enqueue(request)
            }
        }
    }
}

// WebSocket reconnect handler
private func handleMessage(_ wsMessage: WSMessage) async {
    switch wsMessage.type {
    case .message, .queueDrain:
        // .queueDrain = server delivering messages that arrived while offline
        await messageRelayManager.handleIncomingMessage(wsMessage.payload)
    case .confirmation:
        await anchoringTracker.confirmAnchoring(...)
    default: break
    }
}
```

**WebSocket Reconnect Policy:**
```swift
// Exponential backoff: 1s, 2s, 4s, 8s, 16s, 32s (max), then repeat at 32s
// On successful reconnect:
// 1. Re-authenticate (sign challenge with Secure Enclave)
// 2. Drain local outbox (MessageRelayManager.drainOfflineQueue)
// 3. Server automatically drains its offline queue for this user
```

**Local Cache:**
- SwiftData models: `Message`, `Conversation`, `User`, `GroupKey`
- All fields encrypted at rest via `SecureStorage` (AES-256-GCM, HKDF-derived key)
- Size limit: 1000 messages per conversation cached locally; older evicted
- Cache invalidated on snapshot confirmation events (for token balance, trust tier)

## Blueprints

- Frontend — Defines `OfflineQueueManager`, `MessageRelayManager.drainOfflineQueue()`, WebSocket reconnect behavior, and `queueDrain` message type
- Backend — Specifies server-side offline queue (Redis/PostgreSQL), 30-day retention for 1:1, 7-day for large groups

---

### WO-50: Create Rich Media Reaction System with Voice and Visual Responses

**Blueprint:** Message Reactions, Polls, and Interactive Elements

## Summary

Build the rich media reaction system — users can respond to messages with voice note reactions (up to 30 seconds), photo reactions, or short video clip reactions (up to 15 seconds). All rich media reactions are compressed, E2E encrypted, and relayed via the standard message pipeline as `MessageContentType.reaction` with a media attachment.

## In Scope

- Voice reaction recording: `AVAudioRecorder`, Opus/AAC codec, max 30 seconds, waveform visualization during record
- Photo reaction: camera capture or gallery picker, auto-downscale to max 1080px, JPEG compression
- Video clip reaction: `AVCaptureSession`, max 15 seconds, H.264 compression, 720p max resolution
- All rich media encrypted with Kinnami before transmission (same pipeline as media messages)
- Media stored in IPFS after relay upload (via Media Service port 8008)
- Playback: inline in reaction row; video auto-muted, tap to unmute/play with volume
- Delete own reaction: send `{type: "reaction_delete", reactionMessageId}` WebSocket message
- Sync to all conversation participants within 5 seconds
- Notification on rich media reaction: "Alice reacted with a voice note"

## Out of Scope

- Standard emoji reactions (WO-10)
- NFT emoji reactions (WO-60)
- Advanced video editing/filters

## Requirements

Derived from the Message Reactions, Polls, and Interactive Elements blueprint.

**Rich Reaction Message:**
```swift
struct RichReactionPayload: Codable {
    let parentMessageId: String
    let mediaType: RichReactionMediaType    // .voice, .photo, .video
    let mediaURL: URL                       // IPFS CID URL after upload
    let durationSeconds: Double?            // For voice/video
    let thumbnailData: Data?                // For video: first frame thumbnail
}

enum RichReactionMediaType: String, Codable { case voice, photo, video }
// Transmitted as: MessageContentType.reaction with RichReactionPayload
// Same commitment hash anchoring as standard reactions
```

## Blueprints

- Message Reactions, Polls, and Interactive Elements — Defines voice notes (30s), photo, video clips (15s) as reactions, compression, E2E encryption, playback, and deletion

---

### WO-54: Build Message Archive System with Custom Folders and Search Integration

**Blueprint:** Advanced Message Search and Archive System

## Summary

Build the message archive system with custom folders and automatic archiving rules. Users can create named archive folders with hierarchical organization, move messages and conversations manually, and define automatic rules that archive content based on inactivity, content type, or sender trust score.

## In Scope

- Custom archive folder creation (up to 5 levels deep, unlimited folders)
- Manual archiving: drag-and-drop or context menu to move messages/conversations to archive folders
- Auto-archiving rules: trigger based on conversation inactivity (configurable 30d–2yr), content type, or sender trust tier (e.g., archive all messages from Tier 1 senders)
- Archive browsing view: folder tree, sortable by date/sender/relevance
- Archive search integration: archived messages included in search results with archive location indicator
- Archive restoration: move archived content back to active conversations
- Archive deletion: secure removal with user confirmation; optionally secure-wipe (zero-fill before delete)
- Full search capability across archived content using existing local index (WO-3)

## Out of Scope

- Secure backup (WO-64)
- Cross-device archive sync (WO-73)
- Archive analytics (WO-92)
- Bulk operations beyond folder-level

## Requirements

Derived from the Advanced Message Search and Archive System blueprint.

**Archive Folder Model:**
```swift
struct ArchiveFolder: Identifiable, Codable {
    let id: UUID
    var name: String
    var parentFolderId: UUID?       // nil = root level
    var createdAt: Date
    var autoArchiveRules: [AutoArchiveRule]
}

struct AutoArchiveRule: Codable {
    var type: RuleType
    var inactiveDays: Int?          // For .inactivity rule
    var contentTypes: [MessageContentType]?  // For .contentType rule
    var maxTrustTier: Int?          // For .trustTier rule (archive if tier ≤ value)

    enum RuleType: String, Codable {
        case inactivity, contentType, trustTier
    }
}
```

**Auto-Archive Evaluation (daily background job):**
```swift
actor AutoArchiveEngine {
    func evaluateRules() async {
        let folders = await database.fetchFoldersWithRules()
        for folder in folders {
            for rule in folder.autoArchiveRules {
                let eligibleItems = await database.findItemsMatchingRule(rule)
                for item in eligibleItems {
                    await archiveItem(item, to: folder.id)
                }
            }
        }
    }
}
```

## Blueprints

- Advanced Message Search and Archive System — Defines custom archive folders, automatic archiving by inactivity/content type/trust score, archive browsing, search integration, and secure deletion

---

### WO-57: Build Push Notification and Analytics System

**Blueprint:** Frontend

## Summary

Implement APNs push notification integration and the full notification management system — including digest mode, Do Not Disturb scheduling with time-zone auto-adjustment, lock screen preview controls (default: hidden for privacy), and five notification categories. Also covers privacy-preserving analytics collection.

## In Scope

- APNs device token registration in `AppDelegate.swift`
- Push notification permission request flow with clear description of notification types
- Handling of incoming push notifications: `message` (wake-up with conversation ID only), `transaction_confirmation`, `reward_update`, `anchoringConfirmation`
- In-app notification display for foreground notifications (custom banner overlay)
- Notification badge management: app icon badge count, conversation-level unread counts
- Per-conversation and global notification muting
- `UNUserNotificationCenter` configuration with notification categories and actions
- **Digest mode:** real-time (default), hourly batch, daily summary — user selects per-conversation or globally
- **Do Not Disturb scheduling:** user defines DND time windows; automatic time-zone adjustment when user travels
- **Lock screen preview controls:** show full preview / show sender only / hide all (default: **hidden** for privacy); configurable globally and per-conversation
- **5 notification categories:** messages, transactions, rewards, governance, system — each independently configurable
- Privacy-preserving analytics: opt-in consent flow, anonymized event tracking (no DID, no message content, no IP)
- Analytics event batching and transmission to backend (anonymized aggregates only)

## Out of Scope

- Backend Notification Service implementation (port 8007, Go backend work order)
- APNs server-side integration (backend work order)
- Marketing or promotional push notifications
- Analytics dashboard or reporting

## Requirements

From the Frontend blueprint (User Management Features — Notification Management section):

**Notification Settings model:**
```swift
struct NotificationSettings: Codable {
    var globalEnabled: Bool
    var digestMode: DigestMode            // NEW: real-time, hourly, daily
    var showPreview: PreviewSetting       // .always, .senderOnly, .never (default: .never)
    var dndEnabled: Bool
    var dndSchedule: DNDSchedule?         // Time window + auto time-zone adjustment
    var categorySettings: CategorySettings
    var conversationSettings: [String: ConversationNotificationSetting]

    enum DigestMode: String, Codable {    // NEW
        case realTime   // Default: push immediately
        case hourlyBatch  // Bundle notifications and deliver hourly
        case dailySummary // Single daily digest
    }

    struct CategorySettings: Codable {   // NEW: per-category on/off
        var messages: Bool = true
        var transactions: Bool = true
        var rewards: Bool = true
        var governance: Bool = true
        var system: Bool = true
    }

    struct DNDSchedule: Codable {
        var startTime: DateComponents    // e.g., 22:00
        var endTime: DateComponents      // e.g., 08:00
        var autoTimeZoneAdjustment: Bool // true: adjusts when user changes time zone
    }

    struct ConversationNotificationSetting: Codable {
        var muted: Bool
        var mutedUntil: Date?
        var digestMode: DigestMode?      // nil = inherit global setting
        var customSound: String?
    }
}
```

**APNs Push Payload (content-blind):**
```json
{
  "aps": { "content-available": 1, "badge": 3, "sound": "default" },
  "type": "message",
  "conversation_id": "conv_abc123",
  "category": "messages"
  // NO: sender DID, message content, preview text
}
```

**Lock screen preview default = hidden:**
Per the Frontend blueprint: lock screen preview controls default to **hidden** for privacy. Users must explicitly opt in to show message previews on the lock screen.

## Blueprints

- Frontend — Defines notification management with digest mode, DND scheduling with time-zone auto-adjustment, lock screen preview controls (default: hidden), and five notification categories
- Backend — Specifies that backend Notification Service (port 8007) sends only conversation IDs in push payloads, never message content

---

### WO-59: Build Message Enhancement Features (Pinning, Forwarding, Reactions, Replies)

**Blueprint:** Blockchain-Anchored Messaging with Provable Integrity

## Summary

Implement message pinning, forwarding, emoji reactions, and message replies. Each of these creates separate message objects that are individually E2E encrypted and anchored in Merkle batches. Forwarding metadata is stored locally only (not on-chain, for privacy). Pinning is local-only (device-local, not synced to metagraph).

## In Scope

- **Pinning**: iOS — pin up to 10 messages per conversation, store locally in SwiftData, dedicated pinned messages view accessible from conversation header; Backend — no sync needed (local-only per blueprint)
- **Forwarding**: iOS — select message(s), choose destination conversation(s), re-encrypt with destination conversation keys, send with optional `isForwarded` flag; forwarding metadata stored LOCAL ONLY; restrictions: cannot forward disappearing messages or hidden folder messages
- **Reactions**: iOS — emoji picker with standard Unicode + custom reactions, `ReactionMessage` type with `parentMessageId` reference, multiple reactions per user per message, real-time aggregation display; Backend — relay reaction as separate encrypted message blob; add reaction commitment to anchoring batch
- **Replies**: iOS — `ReplyMessage` type with `quotedMessageId`, quote preview in chat bubble, reply thread navigation; Backend — relay as separate encrypted message, add commitment to batch

## Out of Scope

- Custom emoji creation (Phase 7, NFT Emoji work order)
- Advanced thread branching
- Group admin moderation of reactions

## Requirements

Derived from the Blockchain-Anchored Messaging blueprint.

**Privacy by Design — On-Chain Storage:**
```
Reactions:  Each reaction generates its own commitment, anchored in Merkle batch
Replies:    Each reply anchored independently in Merkle batch
Pinning:    LOCAL ONLY — not synced to metagraph
Forwarding: Forwarding metadata LOCAL ONLY — not on-chain
```

**Message Types (shared enum):**
```swift
enum MessageContentType: String, Codable {
    case text, image, video, audio, file
    case reaction(parentId: String, emoji: String)
    case reply(quotedId: String)
    case forward(originalId: String?, attribution: Bool)
    case edited(originalId: String, editCount: Int)
}
```

**Forwarding:**
```swift
// Each forwarded message is FULLY RE-ENCRYPTED for the destination conversation
// New key agreement with destination recipient's public key
// New commitment generated and anchored independently
// Forward attribution is optional ("Forwarded from [name]" label)
// Backend cannot link forwarded message to original (privacy by design)
```

**Reactions (Backend relay):**
```go
// Reactions are transmitted as separate message objects with a parentMessageId field
// Backend routes reaction blob to all conversation participants
// Relay cannot read reaction content (encrypted)
// Commitment generated: H(H(reactionData) || nonce) → added to anchoring batch
```

## Blueprints

- Blockchain-Anchored Messaging with Provable Integrity — Defines reactions/replies as separate anchored message objects, forwarding metadata as local-only, pinning as local-only, and forwarding re-encryption requirements

---

### WO-64: Create Secure Backup System for Encrypted Archive Data

**Blueprint:** Advanced Message Search and Archive System

## Summary

Build the encrypted backup system for message archives and search indexes. All backup data is encrypted with AES-256-GCM using user-controlled keys before leaving the device. Supports IPFS/Filecoin for decentralized backup, major cloud providers with client-side encryption, and hardware devices. Backup integrity is verified with cryptographic hashes on restore.

## In Scope

- Backup encryption: AES-256-GCM, user-controlled master key (stored in Secure Enclave)
- Backup packages include: encrypted message archive, encrypted search index, settings
- Cloud storage integration: iCloud Drive, Google Drive, Dropbox — all client-side encrypted (provider sees only ciphertext)
- IPFS/Filecoin backup: encrypt and pin to IPFS, create Filecoin storage deal for long-term
- Hardware device backup: export encrypted package to Documents directory for AirDrop/USB transfer
- Backup scheduling: daily, weekly, monthly with configurable retention (keep last N backups)
- Backup verification on restore: SHA-256 integrity hash checked before restoring
- Backup size: up to 10GB per package with progress indicators and pause/resume
- Selective restore: restore full backup or specific folders/conversations

## Out of Scope

- Real-time synchronization (WO-73)
- Enterprise backup solutions
- Backup sharing between users

## Requirements

Derived from the Advanced Message Search and Archive System blueprint.

**Backup Package Structure:**
```swift
struct BackupPackage {
    let version: Int
    let createdAt: Date
    let integrityHash: Data        // SHA-256 of encrypted payload
    let encryptedPayload: Data     // AES-256-GCM encrypted, user key
    // Payload contents (decrypted):
    //   - messages: [EncryptedMessage] (already E2E encrypted)
    //   - archiveFolders: [ArchiveFolder]
    //   - searchIndexSnapshot: Data (locally encrypted index)
    //   - settings: UserSettings
}

// Encryption:
// userKey = SecureEnclaveManager.deriveBackupKey()  (HKDF from Secure Enclave signature)
// encrypted = AES256GCM.encrypt(payload: serialize(contents), key: userKey)
// integrityHash = SHA256(encrypted)
```

## Blueprints

- Advanced Message Search and Archive System — Defines backup encryption, hardware/cloud/IPFS storage, backup scheduling, verification, and selective restoration requirements

---

### WO-73: Implement Cross-Device Search Index Synchronization with Encryption

**Blueprint:** Advanced Message Search and Archive System

## Summary

Implement cross-device encrypted search index synchronization — allowing users to search their complete message history from any registered device. Index deltas are E2E encrypted and synced through the backend. Each device maintains a full local index copy; delta sync minimizes bandwidth. Offline search works from the local copy.

## In Scope

- Index delta generation: after each indexing batch, compute diff from last sync checkpoint
- E2E encrypted delta transmission: encrypt delta with Kinnami before sending to backend
- Backend sync endpoint: `POST /v1/search/index-sync` — store encrypted deltas, serve to other devices
- Device-side delta apply: merge incoming deltas into local index with conflict resolution (last-write-wins per term)
- Selective device participation: user can include/exclude specific devices from index sync (up to 10 devices)
- Offline support: local index is always available; sync happens opportunistically on connection
- Bandwidth optimization: only changed index entries transmitted (delta sync), estimated 80%+ reduction vs. full index
- Sync conflict resolution: timestamp-based (newer indexed-at wins), user notified of significant conflicts

## Out of Scope

- Message content synchronization (messages sync through relay server)
- Device-to-device direct sync (goes through backend)
- Enterprise device management

## Requirements

Derived from the Advanced Message Search and Archive System blueprint.

**Delta Sync Protocol:**
```swift
// Core/Search/IndexSyncManager.swift
actor IndexSyncManager {
    private let localIndex: LocalMessageIndexer
    private let encryptor: KinnamiEncryptionService

    func syncToServer() async throws {
        let delta = await localIndex.computeDeltaSince(lastSyncCheckpoint)
        guard !delta.isEmpty else { return }

        // Encrypt delta before sending to server
        let encryptedDelta = try encryptor.encrypt(
            plaintext: serialize(delta),
            recipientPublicKey: serverSyncKey  // Server stores encrypted, can't read
        )
        try await backendAPI.post("/v1/search/index-sync", body: IndexSyncRequest(
            deviceId: currentDeviceId,
            encryptedDelta: encryptedDelta,
            checkpointHash: delta.endCheckpoint
        ))
        lastSyncCheckpoint = delta.endCheckpoint
    }

    func pullFromServer() async throws {
        let deltas = try await backendAPI.get("/v1/search/index-sync?since=\(lastSyncCheckpoint)")
        for encryptedDelta in deltas {
            let delta = try encryptor.decrypt(encryptedDelta.payload)
            await localIndex.applyDelta(deserialize(delta))
        }
    }
}
```

## Blueprints

- Advanced Message Search and Archive System — Defines encrypted index synchronization, cross-device search, offline support, bandwidth optimization, conflict resolution, and up to 10 device support

---

### WO-83: Develop Search Result Sharing with Secure Links and Access Controls

**Blueprint:** Advanced Message Search and Archive System

## Summary

Build the secure message link sharing system — generate cryptographically secure links to specific messages or conversations, with access controls (allowed recipients), expiration, view count limits, and password protection. Links are served by the backend and validated on access.

## In Scope

- `POST /v1/search/shares` — create a secure link for a message or conversation range; returns `shareId` + `shareURL`
- Access token embedded in URL: `echomsg://share/{shareId}?token={hmac-signed-token}`
- Access controls: specify allowed recipient DIDs; backend validates DID on access attempt
- Expiration: 1 hour to 30 days, automatic deactivation after expiry
- View count limit: 1–100 views; backend decrements counter and rejects at 0
- Optional password protection: HMAC-SHA256 hash of password stored; checked on access
- Link revocation: `DELETE /v1/search/shares/{shareId}` — immediate deactivation
- Access logging: record access attempts with timestamp and accessor DID
- iOS: share sheet integration to distribute the `echomsg://` deep link
- Read-only shared view: show message(s) with context, original sender's trust tier badge

## Out of Scope

- Public links without access controls (all links require DID verification or password)
- Bulk link generation
- External sharing platforms

## Requirements

Derived from the Advanced Message Search and Archive System blueprint.

**Share Link Model:**
```go
// backend: search/share_model.go
type MessageShareLink struct {
    ShareID         string    `db:"share_id"`        // UUID
    OwnerDID        string    `db:"owner_did"`
    MessageIDs      []string  `db:"message_ids"`     // Messages to share
    AllowedDIDs     []string  `db:"allowed_dids"`    // Empty = password-only
    PasswordHash    string    `db:"password_hash,omitempty"`  // HMAC-SHA256
    ExpiresAt       time.Time `db:"expires_at"`
    MaxViews        int       `db:"max_views"`        // 0 = unlimited
    ViewCount       int       `db:"view_count"`
    CreatedAt       time.Time `db:"created_at"`
    IsRevoked       bool      `db:"is_revoked"`
}
```

## Blueprints

- Advanced Message Search and Archive System — Defines secure link generation, conversation links, access control, expiration, view count limiting, password protection, link revocation, and access tracking

---

### WO-84: Develop Synchronized Message Deletion System

**Blueprint:** Disappearing Messages with Cryptographic Verification

## Summary

Implement the client-side message deletion logic that fires when a disappearing message's timer expires. Per the blueprint, deletion is entirely independent per-device — there are no server-side deletion broadcast commands. Each device deletes its own local copy when its own timer fires. This work order builds the `deleteMessageLocally` function and the `BGTaskScheduler` integration for background deletion.

## In Scope

- `deleteMessageLocally(_:)` function: delete from SwiftData → delete encryption keys from Keychain (via `DisappearingKeyManager`) → delete cached media → clear from memory → preserve commitment hash
- `BGTaskScheduler` registration for background deletion (`BGProcessingTask` for expired messages)
- On-app-launch check: scan SwiftData for any messages past their `expiresAt` and delete immediately
- `DisappearingMessageManager` actor coordinating timer callbacks, foreground deletion, and background deletion
- Soft deletion record: mark message as "deleted, commitment preserved" in SwiftData (audit log, local only)
- Failed deletion retry: if deletion fails (e.g., Keychain locked when device is off), retry on next app foreground
- Deletion verification check: confirm key is gone from Keychain after deletion attempt

## Out of Scope

- Server-side deletion broadcast (not in blueprint — each device deletes independently)
- Syncing deletion confirmation across devices (no sync required, blueprint is clear)
- Legal hold infrastructure (separate Enterprise/Compliance work order)
- Forwarding restriction enforcement (WO-59)

## Requirements

Derived from the Disappearing Messages blueprint.

**Architecture Correction — No Synchronized Deletion:**
```
⚠️ The blueprint explicitly states:
"Independent client-side deletion (no server coordination required)"
"Each device's timer triggers local deletion"

There are NO deletion broadcast commands over the network.
Each device independently deletes when its own timer expires.
If a recipient is offline when the timer expires, their device
will delete when they next open the app (on-launch check).
```

**Deletion Sequence:**
```swift
// Core/Messaging/DisappearingMessageManager.swift
actor DisappearingMessageManager {
    func deleteMessageLocally(_ messageId: String) async {
        // 1. Delete plaintext from SwiftData
        try? await database.deleteMessage(messageId)
        // 2. Destroy encryption keys (via DisappearingKeyManager)
        await keyManager.destroyKey(messageId: messageId)
        // 3. Delete any cached media
        try? mediaCache.deleteMedia(for: messageId)
        // 4. Clear from in-memory message cache
        messageCache.removeValue(forKey: messageId)
        // 5. Write soft-deletion audit record (local only)
        database.markAsDeleted(messageId: messageId, deletedAt: Date())
        // NOTE: commitment hash is PRESERVED — do NOT delete from anchoringInfo table
    }
}
```

**On-Launch Cleanup:**
```swift
// AppDelegate or @main struct lifecycle
func applicationDidBecomeActive() {
    Task {
        let expiredIds = await database.fetchExpiredMessageIds()
        for id in expiredIds {
            await disappearingManager.deleteMessageLocally(id)
        }
    }
}
```

## Blueprints

- Disappearing Messages with Cryptographic Verification — Defines client-side-only deletion architecture, deletion sequence, what gets deleted vs. preserved, failed deletion handling, and the explicit rejection of server-coordinated deletion

---

### WO-92: Build Privacy-Preserving Search Analytics and User Insights Dashboard

**Blueprint:** Advanced Message Search and Archive System

## Summary

Build the privacy-preserving search analytics dashboard — an on-device analytics system that surfaces insights about a user's search behavior and communication patterns without transmitting any data externally. Helps users discover communication trends and optimize their search strategies.

## In Scope

- Local analytics storage: SwiftData records of search events (query hash, timestamp, result count, clicked result index — no plaintext query stored)
- Search frequency tracking: daily/weekly/monthly search counts
- Popular search terms: aggregate top N terms (store hashed + stemmed, not raw queries — privacy)
- Search trend visualization: line chart of search volume over time (Apple Charts / custom)
- Query response time tracking: P50/P95/P99 latency per search engine type (keyword, semantic, filter)
- Result relevance feedback: implicit (did user open a result? did they refine the search?) to score quality
- Analytics dashboard view: accessible from search settings, charts with customizable date range
- Export: CSV/JSON export of aggregated analytics data (no message content, no raw queries)

## Out of Scope

- External analytics service (Mixpanel, Firebase, etc.) — privacy requirement: local only
- Cross-user analytics comparisons
- Real-time analytics streaming

## Requirements

Derived from the Advanced Message Search and Archive System blueprint.

**Privacy-Safe Analytics Event:**
```swift
struct SearchAnalyticsEvent: Codable {
    let eventType: SearchEventType
    let timestamp: Date
    let queryTokenCount: Int     // How many tokens in query (not the actual query)
    let engineType: SearchEngine // .keyword, .semantic, .filter
    let resultCount: Int
    let durationMs: Int
    let clickedResultRank: Int?  // nil = no result clicked
    // NEVER: raw query text, message content, sender names
}

enum SearchEventType: String, Codable {
    case querySubmitted, resultClicked, queryRefined, noResultsFound
}
```

## Blueprints

- Advanced Message Search and Archive System — Defines search frequency tracking, popular terms, trend analysis, performance metrics, quality metrics, and privacy-preserving local analytics

---

### WO-103: Implement Cryptographic Message Security System

**Blueprint:** Verified Financial Institution Integration

## Summary

Implement the cryptographic message security system for banking communications — E2E encryption via Kinnami (X25519 + ChaCha20-Poly1305), institutional ECDSA P-256 signatures on every banking message, signature verification by customer, and key management including rotation and revocation for institutional DID keys.

## In Scope

- Banking messages encrypted with Kinnami (same as all messages) — platform cannot read content
- Institutional P-256 signature on every message from bank: sign message hash with institutional DID private key
- iOS signature verification: when customer receives bank message, verify P-256 signature against institutional DID public key from Cardano; display `✓ Verified from [Bank Name]` or `⚠️ Invalid Signature`
- Institutional key management: key generation for institutional DID, key rotation (requires multi-sig), key revocation on Cardano
- Immutable signature audit: `{messageId, institutionDID, signatureHash, timestamp}` anchored on Data L1
- Signature revocation validation: check institutional DID credential revocation registry before accepting signed messages

## Out of Scope

- Customer-side key management (handled by Secure Enclave, WO-7/WO-18)
- Message transport (standard relay pipeline)
- Biometric transaction auth (WO-128)

## Requirements

Derived from the Verified Financial Institution Integration blueprint.

**Message Authentication Flow:**
```
Bank → Signs message with institutional P-256 key
     → Encrypts with customer's X25519 public key (Kinnami)
     → Sends to ECHO backend (cannot read content)
Customer → Decrypts with own private key
         → Verifies P-256 signature against institution's DID on Cardano
         → Shows verification checkmark
```

## Blueprints

- Verified Financial Institution Integration — Defines E2E encryption, cryptographic signing, signature verification, key management, and immutable audit trails for banking communications

---

### WO-104: Implement Government ID Document Scanning and Validation System

**Blueprint:** In-App High-Assurance Identity Verification and Reward

## Summary

Build the iOS government ID document scanning UI for the In-App High-Assurance Identity Verification flow. The user captures their driver's license, passport, or national ID using the iOS camera. Image quality is validated locally before the image is submitted directly to the third-party IDV provider (Stripe Identity or Sumsub). ECHO's backend never receives or stores document images.

## In Scope

- Document type selection UI: driver's license, passport, national ID card
- iOS camera interface with live guidance overlay (position document in frame, improve lighting)
- Local image quality checks before submission: minimum resolution (720p), blur detection, glare detection, complete document visible
- Document capture: front/back photo for driver's license, photo page for passport
- Submit captured images directly to IDV provider SDK (Stripe Identity SDK or Sumsub SDK) — ECHO backend is bypassed entirely
- Progress indicators and clear error messages for quality failures ("Document is blurry, please retake")
- Image data cleared from memory immediately after SDK submission
- Privacy disclosure shown before scanning: "Your ID will be processed by [IDV provider]. Echo never stores your document."

## Out of Scope

- Selfie/liveness check (WO-113)
- Third-party IDV API integration (WO-120, backend)
- Apple Digital ID (WO-126)
- OCR processing (done by IDV provider, not ECHO)

## Requirements

Derived from the In-App High-Assurance Identity Verification blueprint.

**Privacy Architecture:**
```
User's Document → IDV Provider SDK (Stripe Identity / Sumsub)
                      ↓
                 IDV Provider processes, verifies, DELETES images
                      ↓
                 Callback to ECHO backend: {pass/fail, reference_uuid} only
                 ECHO NEVER RECEIVES the document image
```

**iOS Camera View:**
```swift
// Presentation/Features/Verification/DocumentScanView.swift
struct DocumentScanView: View {
    let documentType: DocumentType
    @ObservedObject var viewModel: VerificationViewModel

    var body: some View {
        ZStack {
            CameraPreview()
            DocumentGuideOverlay(documentType: documentType)
            VStack {
                Spacer()
                Text(viewModel.captureGuidance)  // "Move closer", "Hold steady"
                Button("Capture") { viewModel.captureDocument() }
                    .disabled(!viewModel.imageQualityAcceptable)
            }
        }
    }
}
```

## Blueprints

- In-App High-Assurance Identity Verification and Reward — Defines document type support, image quality requirements, privacy architecture (IDV provider handles PII), and NIST 800-63-3 IAL2 compliance

---

### WO-113: Build Selfie-Based Liveness Detection and Face Matching System

**Blueprint:** In-App High-Assurance Identity Verification and Reward

## Summary

Build the iOS selfie capture and liveness detection UI for the In-App High-Assurance Identity Verification flow. The selfie is submitted directly to the IDV provider SDK (Daon or 1Kosmos) for liveness detection, anti-spoofing, and face matching against the previously scanned document. ECHO's backend never receives or processes biometric images.

## In Scope

- Selfie camera interface with positioning guides (center face in frame, consistent lighting)
- Real-time quality feedback: face detected, adequate lighting, proximity guidance
- In-app Daon or 1Kosmos SDK integration for liveness challenge and anti-spoofing
- Pass liveness result to IDV provider for face matching against document photo
- Anti-spoofing measures embedded in IDV provider SDK (photo, video, mask detection)
- Clear status display during verification: "Analyzing…", "Verification complete"
- Image data cleared from memory after SDK submission
- Privacy disclosure: "Your selfie will be processed by [IDV provider] and deleted after verification."

## Out of Scope

- Document scanning (WO-104)
- Face matching algorithm (handled by IDV provider, not ECHO)
- Trust score elevation (WO-181)
- Verifiable credential issuance (WO-132)

## Requirements

Derived from the In-App High-Assurance Identity Verification blueprint.

**Privacy Architecture:**
```
User's Selfie → IDV Provider SDK (Daon / 1Kosmos)
                    ↓
               SDK performs: liveness check + face match + anti-spoof
                    ↓
               SDK deletes images after processing
                    ↓
               Callback to ECHO backend: {pass/fail, confidence, reference_uuid}
               ECHO NEVER RECEIVES the selfie image
```

**iOS Selfie Capture:**
```swift
// Presentation/Features/Verification/LivenessCheckView.swift
struct LivenessCheckView: View {
    @ObservedObject var viewModel: VerificationViewModel

    var body: some View {
        VStack {
            // IDV provider SDK provides the camera/liveness UI natively
            IDVProviderLivenessView(
                provider: viewModel.selectedIDVProvider,  // .daon or .oneKosmos
                onComplete: { result in
                    viewModel.handleLivenessResult(result)
                }
            )
            Text(viewModel.livenessStatus)  // "Hold still...", "Almost done..."
        }
    }
}
```

## Blueprints

- In-App High-Assurance Identity Verification and Reward — Defines selfie capture, liveness detection, face matching, anti-spoofing requirements, and privacy architecture (IDV provider processes all biometric data)

---

### WO-120: Integrate Third-Party Identity Verification Service with NIST Compliance

**Blueprint:** In-App High-Assurance Identity Verification and Reward

## Summary

Build the backend integration with third-party NIST 800-63-3 IAL2-compliant identity verification providers. The backend acts as a coordinator: it never receives raw images or PII — the iOS app submits documents directly to the IDV SDK, and the IDV provider sends a callback to the backend with a reference UUID and pass/fail result. The backend then maps the UUID to the user's DID, issues a credential, and triggers the trust tier elevation.

## In Scope

- IDV provider webhook/callback endpoint: `POST /v1/identity/verification/callback`
- Callback payload processing: `{referenceUUID, result: "pass"|"fail", confidence, documentType, ageOver18}`
- Reference UUID to user DID mapping (stored temporarily during active verification sessions)
- Verification result routing: on pass → trigger VC issuance (WO-132) and trust tier elevation (WO-37, WO-181)
- Multi-provider support: Stripe Identity (primary), Sumsub (fallback), configurable per region
- GDPR/CCPA compliance: no PII stored, only reference UUID + outcome + timestamp
- Callback signature verification (IDV provider HMAC signature)
- Retry logic for transient IDV provider failures
- Compliance audit log: verification request count, outcome, provider — no PII

## Out of Scope

- Document image processing (done by IDV provider's SDK)
- Selfie/liveness processing (done by IDV provider's SDK)
- Apple Digital ID (WO-126)
- VC issuance (WO-132)
- Trust tier management (WO-37, WO-181)

## Requirements

Derived from the In-App High-Assurance Identity Verification blueprint.

**Backend Never Sees PII:**
```
iOS App → [Document images + selfie] → IDV Provider SDK
                                             ↓
                                    IDV Provider processes, verifies, DELETES images
                                             ↓
IDV Provider → Callback to ECHO Backend: {reference_uuid, pass/fail, confidence_score, document_type, age_over_18}
                ↓
ECHO Backend: map reference_uuid → userDID → trigger VC issuance + trust tier elevation
```

**Callback Handler:**
```go
// POST /v1/identity/verification/callback
type IDVCallbackPayload struct {
    ReferenceUUID   string  `json:"reference_uuid"`
    Result          string  `json:"result"`          // "pass" or "fail"
    ConfidenceScore float64 `json:"confidence_score"` // 0.0–1.0
    DocumentType    string  `json:"document_type"`    // "passport", "drivers_license"
    AgeOver18       bool    `json:"age_over_18"`
    // NO PII: no name, DOB, address, document number
}

func (h *IdentityHandler) HandleIDVCallback(payload IDVCallbackPayload) error {
    // 1. Verify HMAC signature from IDV provider
    // 2. Map reference_uuid to userDID (temporary session record)
    userDID := h.sessionStore.GetDIDForReference(payload.ReferenceUUID)
    // 3. Determine trust tier from result
    if payload.Result == "pass" {
        tier := 4  // High-assurance → Tier 4
        h.credentialService.IssueHighAssuranceCredential(userDID, payload.DocumentType)
        h.trustService.ElevateTier(userDID, tier)
        h.rewardsService.QueueVerificationReward(userDID, "high_assurance")
    }
    return nil
}
```

## Blueprints

- In-App High-Assurance Identity Verification and Reward — Defines the privacy architecture, NIST 800-63-3 IAL2 compliance, provider callback model, and post-verification actions

---

### WO-126: Implement Apple Digital ID Integration for iOS Devices

**Blueprint:** In-App High-Assurance Identity Verification and Reward

## Summary

Implement Apple Digital ID integration as the alternative high-assurance verification path for iOS 17+ users. Apple's Digital ID API provides privacy-preserving government-level identity verification without exposing personal data to ECHO. On devices where Apple Digital ID is not available, the flow falls back to document scanning (WO-104/WO-113).

## In Scope

- iOS 17+ Apple Digital ID API integration via `AuthenticationServices` framework
- `ASAuthorizationController` with `ASAuthorizationDigitalIDProvider` for credential request
- Request construction: specify required attributes (age verification, not full identity data)
- Callback handling: receive attribute tokens from Apple, submit reference to ECHO backend
- Backend endpoint: `POST /v1/identity/apple-digital-id/verify` receives Apple-signed assertion
- Backend forwards to Apple for server-side verification, receives `{verified, attributes, referenceId}`
- On success: issue High-Assurance VC (WO-132) and trigger trust tier elevation (Tier 4)
- Device capability check: detect iOS 17+, display "Apple Digital ID Available" option if supported
- Graceful fallback: if Apple Digital ID unavailable (older iOS or no enrolled ID), route to document scan

## Out of Scope

- Government ID document scanning (WO-104)
- Selfie liveness detection (WO-113)
- Stripe Identity / Sumsub integration (WO-120)
- Android support

## Requirements

Derived from the In-App High-Assurance Identity Verification blueprint.

**iOS Integration:**
```swift
// Domain/UseCases/Identity/VerifyWithAppleDigitalIDUseCase.swift
struct VerifyWithAppleDigitalIDUseCase {
    func execute() async throws -> VerificationResult {
        guard isAppleDigitalIDAvailable() else {
            throw VerificationError.appleDigitalIDUnavailable
        }

        // Request only what's needed (privacy-preserving)
        let provider = ASAuthorizationDigitalIDProvider()
        let request = provider.createRequest()
        request.requestedElements = [.ageOver18, .stateID]  // Minimum necessary

        let controller = ASAuthorizationController(authorizationRequests: [request])
        let authorization = try await controller.performRequests()

        // Submit assertion to ECHO backend (not raw identity data)
        let result = try await apiClient.post("/v1/identity/apple-digital-id/verify",
            body: AppleDigitalIDRequest(assertion: authorization.credential.assertion))

        return .success(method: .appleDigitalID, tier: 4)
    }

    private func isAppleDigitalIDAvailable() -> Bool {
        // Check iOS 17+ and that user has enrolled an Apple Digital ID
        guard #available(iOS 17.0, *) else { return false }
        return ASAuthorizationDigitalIDProvider.isAvailable
    }
}
```

## Blueprints

- In-App High-Assurance Identity Verification and Reward — Defines Apple Digital ID as a privacy-preserving verification method (iOS 17+), automatic VC creation, no personal data storage

---

### WO-149: Develop Deletion Policies and Selective Preservation System

**Blueprint:** Disappearing Messages with Cryptographic Verification

**Purpose**: Implement flexible deletion policies that support different retention periods and selective message preservation based on user trust levels and compliance requirements, enabling verified users to access advanced features while maintaining audit capabilities for legal and business purposes.

**Requirements**:
- Configure trust level-based deletion policies that allow verified users longer retention periods and access to advanced preservation features
- Implement selective message preservation that allows users to mark specific disappearing messages for extended retention or permanent preservation
- Enforce deletion policy compliance automatically based on user trust level and message classification
- Support policy customization for enterprise users with specific compliance requirements and retention schedules
- Handle policy exceptions for legal hold scenarios where messages must be preserved beyond normal expiration times
- Provide policy auditing capabilities that track all preservation decisions and policy enforcement actions
- Ensure policy enforcement respects user consent and provides clear notifications about preservation decisions
- Support graduated policy tiers that offer different preservation options based on user verification status and trust score
- Implement policy compliance reporting that demonstrates adherence to retention requirements for regulatory purposes
- Provide policy transparency by clearly displaying current retention policies and preservation options to users

**Out of Scope**:
- Trust score calculation or verification processes
- Legal hold management interfaces
- Message deletion execution logic
- Compliance reporting user interfaces

---

### WO-172: Create Content Categorization and Archive Management System

**Blueprint:** Broadcast Channels and Community Features

## Summary

Build the channel content categorization and archive management system — creators organize posts into custom hierarchical categories and topic tags, subscribers browse historical content through an archive with full-text search and date/type filters.

## In Scope

- Category management: creator creates/edits/deletes custom categories (hierarchical, up to 3 levels deep) from channel settings
- Tag management: creator adds custom tags to posts; subscribers browse by tag
- Archive view: chronological list of all published posts with category/tag filter, date range filter, content type filter
- Full-text archive search: search post titles and (where unencrypted preview available) descriptions
- Retention policies: channel creator sets auto-deletion schedule (30d to permanent) with subscriber notification 7 days before deletion
- Bulk categorization: select multiple posts → assign category/tags in one action
- Archive export: `GET /v1/channels/{id}/posts/export` → JSON/CSV of post metadata (not encrypted content)
- Compliance deletion: when post is deleted, removed from all archive indexes + blockchain deletion record submitted

## Out of Scope

- AI/ML auto-categorization
- Semantic search (iOS Advanced Message Search — separate feature)
- External CMS integrations

## Requirements

Derived from the Broadcast Channels blueprint.

**Category Model:**
```go
type ContentCategory struct {
    ID       string  `db:"id"`
    ChannelID string `db:"channel_id"`
    Name     string  `db:"name"`
    ParentID *string `db:"parent_id"`  // Hierarchical
    SortOrder int    `db:"sort_order"`
}
```

## Blueprints

- Broadcast Channels and Community Features — Defines content categorization, topic tagging, archive functionality, retention policies, bulk operations, and archive export

---

### WO-192: Implement Real-time Messaging Features with Typing Indicators

**Assignee:** Chad Cromwell

**Blueprint:** Dynamic Trust Network and Social Verification

## Summary

Implement real-time typing indicators and read receipts in the WebSocket relay protocol. Typing indicators broadcast over WebSocket as separate non-encrypted status messages (not as message blobs). Read receipts are per-message delivery confirmations. Both respect user privacy settings and can be disabled per conversation.

## In Scope

- WebSocket message types for real-time presence: `typing` (user is typing), `presence` (online/offline), `receipt` (delivery/read status)
- Typing indicator: iOS `WebSocketRelay` sends `{type: "typing", conversationId, senderDID}` on each keystroke, backend fans out to all participants
- Typing timeout: 5-second inactivity timer on iOS; automatic `{type: "typing_stopped"}` sent after timeout
- Multiple users typing: iOS displays "Alice, Bob are typing…" for group conversations
- Read receipts: iOS sends `{type: "receipt", messageId, status: "read"}` when message is opened; backend relays to sender
- Delivery receipt: backend sends `{type: "receipt", messageId, status: "delivered"}` to sender when relay delivers to recipient device
- Privacy controls: typing indicators and read receipts can be disabled globally or per conversation
- Status events are NOT encrypted (they are metadata, not content) and do not generate commitment hashes

## Out of Scope

- Audio message recording/playback (WO-194)
- Message search (WO-197)
- Group messaging beyond typing indicators

## Requirements

Derived from the Dynamic Trust Network blueprint.

**WebSocket Typing Event:**
```swift
// Core/Relay/TypingIndicatorManager.swift
actor TypingIndicatorManager {
    private var typingTimers: [String: Task<Void, Never>] = [:]  // keyed by conversationId

    func userStartedTyping(conversationId: String) async {
        webSocket.send(WSMessage(type: .typing, conversationId: conversationId))
        typingTimers[conversationId]?.cancel()
        typingTimers[conversationId] = Task {
            try? await Task.sleep(nanoseconds: 5_000_000_000)  // 5 seconds
            webSocket.send(WSMessage(type: .typingStopped, conversationId: conversationId))
        }
    }
}
```

**Read Receipt:**
```swift
// Send when message is opened/scrolled into view
func markAsRead(messageId: String) {
    webSocket.send(WSMessage(type: .receipt, payload: ReceiptPayload(
        messageId: messageId, status: "read", timestamp: Date()
    )))
}
```

## Blueprints

- Dynamic Trust Network and Social Verification — Defines typing indicators (5s timeout, multiple users display), read receipts, delivery tracking, and privacy controls
- Frontend — Defines `WSMessage.MessageType` enum including `.typing`, `.presence`, `.receipt`

---

### WO-194: Build Audio Messaging System with Voice Notes

**Assignee:** Chad Cromwell

**Blueprint:** Dynamic Trust Network and Social Verification

## Summary

Build the audio messaging (voice notes) system — recording up to 5 minutes with Opus codec compression, end-to-end encryption matching standard messages, waveform visualization, playback with variable speed, and optional local transcription for search integration. Voice notes use the same `MessageRelayManager` pipeline as text messages.

## In Scope

- iOS audio recording with `AVAudioRecorder` using Opus codec (or AAC with `AVFoundation`)
- Real-time waveform visualization during recording (amplitude sampling)
- Maximum recording duration: 5 minutes with visual countdown
- Pause/resume recording; cancel with swipe gesture
- E2E encryption of audio data (same Kinnami pipeline as text messages)
- Waveform visualization during playback with seek control
- Playback speed control: 0.75x, 1x, 1.25x, 1.5x
- Background audio playback with `AVAudioSession` category `.playback`
- Optional local transcription using iOS `SFSpeechRecognizer` (on-device, no cloud, privacy-preserving)
- Transcribed text stored locally in SwiftData alongside audio message for search indexing
- Transcription search integration: voice notes appear in message search results when text matches

## Out of Scope

- Real-time voice calling (WO-5, WO-19)
- Cloud-based transcription (privacy requirement: local only)
- Multi-language transcription in v1 (English only initially)
- Voice message backup to cloud

## Requirements

Derived from the Dynamic Trust Network blueprint.

**Audio Recording and Encoding:**
```swift
// Presentation/Features/Chat/VoiceMessageRecorder.swift
class VoiceMessageRecorder: ObservableObject {
    @Published var isRecording = false
    @Published var duration: TimeInterval = 0
    @Published var waveformData: [Float] = []

    func startRecording() {
        // AVAudioRecorder with Opus/AAC codec
        let settings: [String: Any] = [
            AVFormatIDKey: Int(kAudioFormatMPEG4AAC),  // AAC for AVFoundation compat
            AVSampleRateKey: 16000,    // 16kHz for voice
            AVNumberOfChannelsKey: 1,  // Mono
            AVEncoderBitRateKey: 32000 // 32kbps for voice quality
        ]
        // Start recording, sample amplitude for waveform
    }

    func stopAndEncrypt() async throws -> Data {
        let audioData = try audioRecorder.stopAndGetData()
        // Encrypt with Kinnami (same pipeline as text messages)
        return try encryption.encrypt(plaintext: audioData, recipientPublicKey: recipientKey)
    }
}
```

**Transcription (Local, Privacy-Preserving):**
```swift
func transcribeLocally(_ audioURL: URL) async -> String? {
    guard SFSpeechRecognizer.authorizationStatus() == .authorized else { return nil }
    let recognizer = SFSpeechRecognizer(locale: .current)
    let request = SFSpeechURLRecognitionRequest(url: audioURL)
    request.requiresOnDeviceRecognition = true  // MUST be on-device
    let result = try? await recognizer.recognitionTask(with: request).result
    return result?.bestTranscription.formattedString
}
```

## Blueprints

- Dynamic Trust Network and Social Verification — Defines audio messages: Opus codec, up to 5 minutes, E2E encryption, waveform, playback speed, optional transcription

---

### WO-196: Implement Call History and Notification Badge System

**Assignee:** Chad Cromwell

**Blueprint:** Dynamic Trust Network and Social Verification

## Summary

Build the call history tracking system and iOS notification badge management. Call history records all voice and video calls with type, duration, timestamp, and participants. Notification badges show unread counts on the app icon and per-conversation; missed call badges appear on the calls tab and contacts. All badge state is managed via SwiftData locally, with badge counts synced to the backend.

## In Scope

- Call history data model: `CallRecord {callId, callType, duration, participants, timestamp, missedByCurrentUser}`
- Call history list view: filterable by call type and contact, searchable
- Missed call badges per contact: show count of missed calls since last viewed
- App icon badge count: sum of unread messages + missed calls (via `UNUserNotificationCenter`)
- Conversation-level badge counts: unread message counts per conversation in conversation list
- Badge count clearing: automatic on entering conversation or call history view
- Muted conversations: do NOT contribute to badge counts
- Archived conversations: do NOT show badges
- Real-time badge updates via WebSocket: `receipt` and `ack` messages trigger badge recalculation
- Badge state persisted in SwiftData, survives app restarts

## Out of Scope

- Voice/video call implementation (WO-5, WO-19, WO-31)
- Call recording (WO-43)
- Advanced call analytics

## Requirements

Derived from the Dynamic Trust Network blueprint.

**Call Record Model:**
```swift
// Domain/Models/CallRecord.swift
struct CallRecord: Identifiable, Codable {
    let id: String
    let callType: CallType     // .voice, .video
    let participants: [String] // DIDs
    let startedAt: Date
    let duration: TimeInterval
    let missedByCurrentUser: Bool
    let initiatorDID: String
}
enum CallType: String, Codable { case voice, video }
```

**Badge Management:**
```swift
// Core/Notifications/BadgeManager.swift
@MainActor class BadgeManager: ObservableObject {
    @Published var totalBadgeCount: Int = 0

    func recalculate() async {
        let unread = await database.totalUnreadMessages()
        let missed = await database.totalMissedCalls()
        totalBadgeCount = unread + missed
        UNUserNotificationCenter.current().setBadgeCount(totalBadgeCount)
    }

    func clearConversationBadge(conversationId: String) async {
        await database.markConversationRead(conversationId)
        await recalculate()
    }
}
```

## Blueprints

- Dynamic Trust Network and Social Verification — Defines call history, missed call indicators, notification badges, real-time badge updates, muted conversation handling

---

### WO-197: Build Message and Conversation Search System

**Assignee:** Chad Cromwell

**Blueprint:** Dynamic Trust Network and Social Verification

## Summary

Implement in-app message and conversation search — keyword search across locally-stored decrypted messages, conversation filtering by contact name or last message, fuzzy matching for typos, and search result highlighting. All search is performed on-device against the local SwiftData store (encrypted messages are decrypted for local indexing). This is the basic in-app search; the Advanced Message Search feature (separate work orders) adds semantic search and cross-device index sync.

## In Scope

- iOS full-text search over local SwiftData message store using `NSPredicate` or SwiftData `#Predicate`
- Fuzzy matching for typos: use `NSString.compare` with `.diacriticInsensitiveSearch` and edit distance tolerance
- Search filters: sender DID, date range, content type (text, image, audio, file)
- Case-insensitive, partial-word matching with search result highlighting
- Conversation search: filter conversation list by contact name or last message content
- Search results ranked by: exact match > fuzzy match, recency weighting
- Offline support: search works entirely from local SwiftData store
- Respect blocked users: exclude blocked user messages from search results
- Respect archived conversations: include when user explicitly searches archives

## Out of Scope

- Semantic/NLP search (Advanced Message Search System — WO-41)
- Cross-device index synchronization (Advanced Message Search System — WO-73)
- Message search analytics (Advanced Message Search System — WO-92)
- Search result sharing with secure links (Advanced Message Search System — WO-83)

## Requirements

Derived from the Dynamic Trust Network blueprint.

**In-App Search:**
```swift
// Presentation/Features/Search/ConversationSearchViewModel.swift
@MainActor class ConversationSearchViewModel: ObservableObject {
    @Published var searchText = ""
    @Published var results: [SearchResult] = []

    func search(query: String) async {
        guard !query.isEmpty else { results = []; return }
        let messages = await database.searchMessages(
            query: query,
            filters: SearchFilters(
                excludeBlockedUsers: true,
                includeArchived: false
            )
        )
        results = messages.map { SearchResult(message: $0, highlight: highlight(query, in: $0.plaintext)) }
    }
}

struct SearchFilters {
    var senderDID: String?
    var dateRange: DateInterval?
    var contentType: MessageContentType?
    var excludeBlockedUsers: Bool
    var includeArchived: Bool
}
```

## Blueprints

- Dynamic Trust Network and Social Verification — Defines message search with keyword, sender, date range, and media type filters, plus conversation search and privacy-controlled results

---

### WO-198: Implement Conversation Archive and Management System

**Assignee:** Chad Cromwell

**Blueprint:** Dynamic Trust Network and Social Verification

## Summary

Build the conversation archiving system — users can manually archive conversations to remove them from the main list while keeping them accessible. Archived conversations suppress notifications and badges. Auto-archive by inactivity is configurable. Archive state is stored in SwiftData locally with backend sync to the Contacts Service.

## In Scope

- Archive/unarchive actions from conversation swipe menu and long-press context menu
- Separate "Archived" view accessible from conversation list (swipe or tab)
- `ConversationArchiveSettings` per conversation: manual archived state + auto-archive threshold
- Auto-archive rule: archive conversations with no new messages for N days (user-configurable: never, 30 days, 90 days, 1 year)
- Archived conversations: no notification badges, no push notifications, not in main conversation list
- Option to unarchive on new message receipt (configurable per conversation)
- Bulk archive: select multiple conversations and archive in one action
- Archive search: include archived conversations when user explicitly opens archive view and searches
- Archive state synced to Contacts Service (port 8005): `PATCH /v1/conversations/{id}/archive`

## Out of Scope

- Conversation deletion (separate feature)
- Advanced conversation categorization beyond archive/active
- Archive analytics (Advanced Message Search — WO-92)

## Requirements

Derived from the Dynamic Trust Network blueprint.

**Archive Data Model:**
```swift
struct ConversationSettings: Codable {
    let conversationId: String
    var isArchived: Bool
    var archivedAt: Date?
    var autoArchiveAfterDays: Int?   // nil = never auto-archive
    var unarchiveOnNewMessage: Bool  // true = unarchive when new message arrives
    var isMuted: Bool
    var mutedUntil: Date?
}
```

**Archive Logic:**
```swift
// Archive a conversation:
// 1. Set isArchived = true in SwiftData
// 2. Sync to backend: PATCH /v1/conversations/{id}/archive
// 3. Remove from main conversation list
// 4. Stop badge counts from this conversation
// 5. Suppress push notifications (backend Notification Service checks archive status)

// Auto-archive job (runs daily):
// 1. Fetch all non-archived conversations
// 2. For conversations with auto-archive enabled: check lastMessageAt
// 3. If inactive > threshold: call archiveConversation()
```

## Blueprints

- Dynamic Trust Network and Social Verification — Defines archive/unarchive functionality, auto-archive by inactivity, no badges for archived conversations, and archive search support

---

### WO-207: Implement End-to-End Message Encryption and Commitment

**Type:** Build

**Blueprint:** End-to-End Message Encryption and Commitment

## Summary

Implement the canonical `EncryptedPayload` structure including the commitment hash field (as specified in the E2E Encryption blueprint), and implement `GroupKeyManager.distributeGroupKey()` and `encryptGroupMessage()` for per-member group key package distribution. The encryption relay pipeline (WO-4) and Merkle batching (WO-15) cover transport and anchoring. This work order ensures the iOS encryption layer conforms exactly to the blueprint specification.

## In Scope

- **Canonical `EncryptedPayload` struct** with all six fields: `senderEphemeralPublicKey`, `ciphertext`, `authTag`, `nonce`, `commitment`, `schemaVersion`
- **Commitment hash in payload:** `commitment = SHA256(SHA256(plaintext) + commitmentNonce)` — double-hash with per-message random nonce; commitment stored in `EncryptedPayload` and extracted by relay for Merkle batching
- **`encrypt(plaintext:recipientPublicKey:)` function** conforming to blueprint spec — ephemeral X25519 key generation, HKDF-SHA256 key derivation with salt `"echo-message-salt-v1"` and info `"echo-message-encryption"`, ChaCha20-Poly1305 seal
- **`GroupKeyManager.distributeGroupKey(groupKey:members:)`** — encrypt group AES-256 key individually for each member using standard 1:1 X25519+ChaCha20 encryption; return `[(did, encryptedKeyPackage)]` array
- **`GroupKeyManager.encryptGroupMessage(plaintext:groupId:)`** — fetch current group key, AES-256-GCM seal
- **Key rotation trigger** — on membership change (add/remove), admin generates new group key and calls `distributeGroupKey` to all current members
- **Schema version `1`** in all outgoing `EncryptedPayload` instances; relay and backend validate `schemaVersion` field

## Out of Scope

- WebSocket relay transport (WO-4)
- Merkle batch anchoring backend (WO-15)
- Sealed sender Phase 3 (WO-219)
- Phase 3 client-side Merkle proof verification (separate work order)

## Requirements

From the End-to-End Message Encryption and Commitment blueprint:

**Canonical `EncryptedPayload`:**
```swift
struct EncryptedPayload: Codable {
    let senderEphemeralPublicKey: Data   // Recipient uses for X25519 key agreement
    let ciphertext: Data                  // ChaCha20-Poly1305 ciphertext
    let authTag: Data                     // AEAD integrity tag (16 bytes)
    let nonce: Data                       // 12-byte random nonce
    let commitment: Data                  // H(H(plaintext) || nonce) — anchored on-chain
    let schemaVersion: Int                // Current: 1
}
```

**Key Derivation (HKDF-SHA256):**
```swift
HKDF<SHA256>.deriveKey(
    inputKeyMaterial: sharedSecret,
    salt: Data("echo-message-salt-v1".utf8),
    info: Data("echo-message-encryption".utf8),
    outputByteCount: 32
)
```

**Commitment Hash Design:** Double-hash prevents content exposure; random `commitmentNonce` prevents dictionary attacks. After plaintext deletion, commitment becomes permanently unverifiable — on-chain Merkle root proves "message existed at timestamp T" without revealing content.

## Blueprints

- End-to-End Message Encryption and Commitment — Defines canonical `EncryptedPayload`, X25519+ChaCha20 encryption spec, commitment hash formula, `GroupKeyManager` API, and key rotation semantics

---

### WO-237: Implement Offline Message Queue Overflow Backup to IPFS

**Type:** Build

**Blueprint:** Backend

## Summary

Implement the offline message queue overflow backup mechanism in the Go Message Relay service (port 8002): when a recipient's offline queue exceeds 1000 messages, overflow encrypted blobs are pinned to IPFS/Storj and the relay stores only the CID in queue metadata. On reconnect, the relay provides CIDs for the recipient to retrieve overflow messages directly from IPFS, maintaining the content-blind relay model.

## In Scope

- **Queue depth check:** Before enqueuing a new offline message, check recipient's current queue depth in Redis. If depth ≥ 1000, trigger overflow path instead of standard queue insertion
- **IPFS overflow pin:** Submit encrypted message blob to IPFS via Media Service (port 8008); receive CID; store `{recipientDID, CID, timestamp, expiresAt}` in `overflow_queue` PostgreSQL table
- **Queue metadata entry:** For overflow messages, enqueue `{type: "overflow", cid: "QmXxx...", retrievalURL: "https://ipfs.echo.app/ipfs/QmXxx..."}` instead of the blob itself
- **On-reconnect delivery:** When recipient reconnects via WebSocket, relay drains standard queue first (blobs), then sends overflow CID list: `{type: "overflow_manifest", cids: ["QmXxx...", ...]}`
- **Recipient IPFS retrieval:** iOS client receives `overflow_manifest`, fetches each CID from IPFS gateway, decrypts (same E2E key — message is the same opaque blob), processes normally
- **TTL and cleanup:** Overflow IPFS pins expire at same TTL as queue retention (30 days for 1:1, 7 days for large groups); scheduled cleanup job unpins expired blobs and deletes PostgreSQL records
- **Content-blind model preserved:** Relay stores only CID — the blob pinned to IPFS is the same opaque E2E encrypted ciphertext; relay cannot read it whether it's in Redis or on IPFS

## Out of Scope

- Standard offline queue (existing in WO-4)
- IPFS/Storj infrastructure (WO-33)
- iOS message decryption (WO-28)

## Requirements

From the Backend blueprint:

**Overflow backup (updated spec):**
> When a recipient's queue exceeds 1000 messages, overflow E2E encrypted blobs are pinned to IPFS/Storj. The relay stores only the CID in queue metadata. On reconnect, relay provides CIDs for the recipient to retrieve overflow messages directly from IPFS. Content-blind model preserved — backup is the same opaque encrypted blob.

```go
type OfflineQueueEntry struct {
    Type    string    // "blob" or "overflow"
    Payload []byte    // E2E encrypted blob (if type=="blob")
    CID     string    // IPFS CID (if type=="overflow")
    QueuedAt time.Time
    ExpiresAt time.Time
}
```

## Blueprints

- Backend — Defines overflow message backup to IPFS when queue exceeds 1000 messages, CID-based queue metadata, and on-reconnect delivery of overflow manifest

---

---

## Competitive Audit Additions (2026-05-26)

Proposed from `docs/COMPETITIVE_AUDIT_2026-05.md` (Tier 1). Provisional IDs — final WO numbers assigned by Software Factory.

### WO-CA3: did:key-scoped multi-device message sync
**Source:** competitive audit (Signal multi-device sync). **Extends:** WO-73 (cross-device search-index sync).
Sync message history across a user's registered devices keyed by `did:key` (controller pattern), E2E re-encrypted per device; no plaintext server copy.

---

## SimpleX Audit Additions (2026-05-29)

From [`docs/COMPETITIVE_AUDIT_SIMPLEX_2026-06.md`](COMPETITIVE_AUDIT_SIMPLEX_2026-06.md). Synced to Software Factory 2026-05-29.

### WO-314: SimpleX SX1 — Double Ratchet forward secrecy over Kinnami
**Status:** ✅ Completed · **Commits:** `e7797dc` (Go), `3cfcc2d` (iOS)  
**Evidence:** `internal/crypto/ratchet.go`, `ios/Echo/Sources/Core/Security/DoubleRatchet.swift`, `DoubleRatchetCoordinator.swift`
