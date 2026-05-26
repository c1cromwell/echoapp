# Phase 7: Advanced Platform Features

**Total Work Orders:** 86  
**Status Summary:** 86 Backlog  
**Last synced with Software Factory:** 2026-05-26

---

## Backlog (86)

### WO-11: Implement Bot SDK Core Framework with Messaging and API Access

**Blueprint:** Decentralized Bot Framework and Automation

## Summary

Build the Bot SDK core framework — a Go SDK library (published to GitHub) that developers use to create bots for the ECHO platform. Provides secure access to messaging, payment, and file APIs with rate limiting, permission enforcement, and privacy-preserving monitoring.

## In Scope

- Go SDK package with `EchoBot` struct and builder pattern for bot configuration
- Messaging API: `bot.SendMessage(recipientDID, content)`, `bot.OnMessage(handler)` — max 100 messages/minute per bot (rate limited by backend)
- Payment API: `bot.RequestPayment(userDID, amountECHO, reason)` → triggers user authorization prompt; max amount = user-defined limit
- File API: `bot.UploadFile(data, filename)`, `bot.GetFile(fileID)` — max 50MB per file
- Blockchain API: `bot.ReadChainState(query)` — read-only access to metagraph state
- Permission declaration: bots declare required permissions in manifest; backend validates against user grants before each operation
- Standard error codes: `BotError{code, message, retryAfter}` for all API failures
- Privacy-safe SDK logging: log bot-level metrics (call counts, error rates, latency) with no user data
- SDK documentation and code examples in README

## Out of Scope

- Smart contract deployment (WO-24)
- Bot marketplace (WO-52)
- Trust scoring (WO-40)
- Specific bot type implementations (WO-186, WO-189, WO-191)

## Requirements

Derived from the Decentralized Bot Framework and Automation blueprint.

**SDK Interface:**
```go
// echo-bot-sdk/bot.go
type EchoBot struct {
    BotDID       string
    Permissions  []BotPermission
    apiClient    *ECHOAPIClient
}

func (b *EchoBot) SendMessage(recipientDID string, content BotMessageContent) error
func (b *EchoBot) OnMessage(handler MessageHandler) error
func (b *EchoBot) RequestPayment(userDID string, amount Decimal, reason string) (*PaymentRequest, error)
func (b *EchoBot) UploadFile(data []byte, filename string) (string, error) // returns fileID
func (b *EchoBot) ReadChainState(query ChainQuery) (interface{}, error)   // read-only

// Rate limiting enforced by backend; SDK returns BotError{code: "RATE_LIMIT_EXCEEDED"}
// Permission enforcement: each API call validated against user's granted permissions
```

## Blueprints

- Decentralized Bot Framework and Automation — Defines SDK messaging/payment/file/blockchain APIs, permission management, rate limiting, error handling, and logging requirements

---

### WO-40: Create Bot Trust Scoring and Verification System

**Blueprint:** Decentralized Bot Framework and Automation

## Summary

Build the bot trust scoring system — computes a 0–100 trust score for each bot based on user feedback (40%), reliability metrics (30%), and security audit results (30%). Scores displayed with color-coded indicators and updated in real-time as new data is collected.

## In Scope

- Bot trust score computation: `score = (userFeedback × 0.40) + (reliabilityScore × 0.30) + (auditScore × 0.30)`
- User feedback: weighted average of star ratings (1–5), converted to 0–100
- Reliability metrics: uptime percentage (target 99%+), P50/P95 response time, error rate over 30-day rolling window
- Security audit score: 100 for clean audit, deduct per finding (Critical: −50, High: −20, Medium: −10, Low: −5)
- Trust score display: color-coded `TrustBadge` in bot profile (green >80, yellow 60–80, red <60)
- Score history: `{timestamp, score, changeReason}` for transparency on significant changes (>10 points)
- Real-time updates: score recomputed on each new rating or metric update; cached in Redis (5-minute TTL)
- Appeals: `POST /v1/bots/{id}/trust-score/appeals` — developer contests score with evidence; 48-hour review

## Out of Scope

- Bot marketplace UI (WO-52)
- Security audit execution (WO-74)
- Permission management (WO-63)

## Requirements

Derived from the Decentralized Bot Framework blueprint.

**Score Computation:**
```go
type BotTrustScore struct {
    BotDID           string
    TotalScore       int      // 0–100
    UserFeedbackScore float64  // 0–100 (avg star rating * 20)
    ReliabilityScore  float64  // 0–100 (uptime * 100, penalized for errors)
    AuditScore        float64  // 100 - sum(finding penalties)
    ComputedAt        time.Time
}
```

## Blueprints

- Decentralized Bot Framework and Automation — Defines trust score calculation from user feedback, reliability metrics, and security audit results, with color-coded display and appeals process

---

### WO-43: Build Call Recording System with Consent Management

**Blueprint:** Voice and Video Calls with Screen Sharing

## Summary

Build the call recording system with explicit multi-party consent management. All participants must consent before recording starts. Recordings are encrypted and stored locally or in cloud storage. Enterprise users have additional compliance recording capabilities. Recording metadata is logged for audit purposes.

## In Scope

- Consent flow: `POST /v1/calls/{callId}/recording/start` triggers consent request to all participants; recording only starts when all accept
- Visual recording indicator: prominent red "REC" indicator visible to all participants during recording
- Recording controls: start, pause, resume, stop with real-time sync across participants
- Recording capture: `AVAudioRecorder` for audio + `RPScreenRecorder` for video (if screen sharing active)
- Encrypted local storage: AES-256-GCM with Secure Enclave-derived key, stored in app Documents directory
- Recording metadata: `{callId, participants (DIDs), duration, recordedAt, encryptedFileURL, consentConfirmations}`
- Recording management: view recordings list, share encrypted recording (export with key), delete
- Enterprise compliance mode (Org tier): auto-recording on call start with legal notification banner, configurable retention policy
- Access control: recordings shareable only by the recording initiator, E2E encrypted on export

## Out of Scope

- Cloud storage backend provisioning
- Post-processing or video editing
- Transcription of recordings (WO-55)

## Requirements

Derived from the Voice and Video Calls blueprint.

**Consent Flow:**
```swift
// When recording is initiated, server sends consent request to ALL participants:
// WSMessage(type: .recordingConsentRequest, callId: callId, requesterDID: initiatorDID)
// Each participant shows: "Alice wants to record this call. Do you consent?"
// Recording begins ONLY when ALL participants respond "Yes"
// If any participant declines: recording request is rejected

// If participant joins after recording is already in progress:
// They see: "This call is being recorded" with option to leave
```

## Blueprints

- Voice and Video Calls with Screen Sharing — Defines recording consent, controls, recording storage, encrypted content, metadata, and enterprise compliance recording

---

### WO-52: Develop Decentralized Bot Marketplace with Discovery and Management

**Blueprint:** Decentralized Bot Framework and Automation

## Summary

Build the decentralized bot marketplace — browse, search, and install bots with transparent trust scores, reviews, and permission summaries. Users install bots through an explicit permission granting flow. Installed bots are manageable (modify permissions, monitor usage, uninstall).

## In Scope

- Bot listing: name, description, capabilities, trust score badge, star rating, install count, developer DID/profile
- Search: text query + category filter (Productivity, Trading, Customer Service, Entertainment, AI, Utilities) + trust score range + rating filter
- Sort options: trust score, popularity, recently updated
- Bot installation flow: show permission manifest (from on-chain registration) → user explicitly approves each permission → install confirmed
- Installed bots management: view all installed bots, current permissions, last activity, usage stats
- User reviews: star rating + text review (spam-filtered), verified user badge (Tier 3+), helpful/unhelpful votes
- Developer profile: developer DID, published bots, verification status, contact link
- Uninstall: revoke all permissions, remove bot from contact list

## Out of Scope

- Bot trust score computation (WO-40)
- SDK (WO-11)
- Security auditing (WO-74)

## Requirements

Derived from the Decentralized Bot Framework blueprint.

**Bot Listing:**
```swift
struct BotListing: Identifiable {
    let botDID: String
    let name: String
    let description: String
    let category: BotCategory
    let trustScore: Int         // 0–100
    let starRating: Double      // 1.0–5.0
    let installCount: Int
    let permissions: [BotPermission]  // Declared permissions
    let developerDID: String
}
```

## Blueprints

- Decentralized Bot Framework and Automation — Defines bot marketplace with discovery, trust scores, reviews, installation flow, permission display, and management interface

---

### WO-55: Implement Real-Time Transcription with Privacy-Preserving Processing

**Blueprint:** Voice and Video Calls with Screen Sharing

## Summary

Implement real-time call transcription using iOS `SFSpeechRecognizer` with `requiresOnDeviceRecognition = true`. All transcription processing occurs on-device — no audio is sent to external services. Transcripts include speaker labels, timestamps, and confidence scores. Users who enable transcription see live text alongside the call.

## In Scope

- `SFSpeechRecognizer` with `requiresOnDeviceRecognition = true` (mandatory for privacy)
- Streaming recognition: `SFSpeechAudioBufferRecognitionRequest` fed from WebRTC audio buffers
- Speaker identification: label each utterance by participant DID (maps to display name)
- Transcript display: scrolling text overlay in call UI, toggled per-user
- Real-time updates: transcript view updates within 3 seconds of speech
- Speaker labels in output: `[Alice]: Let's schedule for Thursday`
- Confidence indicators: low-confidence segments shown in lighter color
- Transcript export: TXT, PDF export after call with full timestamped transcript
- Accessibility: large text, high contrast mode for transcript view
- Transcript search: store in local SwiftData post-call for message search integration

## Out of Scope

- Cloud transcription APIs (privacy requirement)
- Language translation
- External note-taking integrations
- Sentiment analysis

## Requirements

Derived from the Voice and Video Calls blueprint.

**On-Device Transcription:**
```swift
// Core/Calling/CallTranscriptionManager.swift
class CallTranscriptionManager {
    func startTranscription(for audioTrack: RTCAudioTrack) async throws {
        let recognizer = SFSpeechRecognizer(locale: .current)!
        let request = SFSpeechAudioBufferRecognitionRequest()
        request.requiresOnDeviceRecognition = true  // MUST BE TRUE — privacy requirement
        request.shouldReportPartialResults = true

        recognitionTask = recognizer.recognitionTask(with: request) { [weak self] result, error in
            if let result = result {
                let transcript = TranscriptSegment(
                    text: result.bestTranscription.formattedString,
                    timestamp: Date(),
                    isFinal: result.isFinal
                )
                self?.transcriptPublisher.send(transcript)
            }
        }

        // Feed WebRTC audio into SFSpeech request buffer
        audioTrack.add(audioSink { pcmBuffer in
            request.appendAudioPCMBuffer(pcmBuffer)
        })
    }
}
```

## Blueprints

- Voice and Video Calls with Screen Sharing — Defines real-time transcription, local device processing, multiple language support, speaker identification, transcript search, and export

---

### WO-60: Implement NFT Emoji Collection and Trading System

**Blueprint:** Message Reactions, Polls, and Interactive Elements

## Summary

Build the NFT emoji collection and peer-to-peer trading system. Users collect Cardano-blockchain-verified NFT emojis, use them as unique reactions, and trade them with other users. Ownership is verified on-chain before allowing usage. Trades are recorded as Cardano transactions with automatic ownership transfer.

## In Scope

- NFT emoji gallery view: grid of user's owned emojis with name, rarity level, acquisition date
- NFT emoji usage in reactions: verified against Cardano on-chain ownership before allowing in emoji picker
- Rarity levels: Common, Uncommon, Rare, Epic, Legendary — distinct visual indicator per rarity
- Peer-to-peer trade initiation: propose trade offer (my NFT → your NFT) via in-app trading interface
- Trade confirmation: both parties confirm; Cardano transaction atomically transfers ownership
- NFT emoji marketplace: browse and purchase available emojis from other users or the ECHO store
- Cardano NFT standard: CIP-25 NFT metadata standard on Cardano
- Ownership sync: on app launch, query Cardano for current ownership state of all NFT emoji wallet
- Visual indicator: NFT reactions display a small sparkle/badge to distinguish from standard emoji

## Out of Scope

- NFT emoji minting (done by ECHO platform team, not user-created)
- Auction/bidding mechanics (P2P trade only)
- External NFT marketplaces (Cardano only)

## Requirements

Derived from the Message Reactions, Polls, and Interactive Elements blueprint.

**NFT Emoji Ownership Check:**
```swift
// Before showing NFT emoji in picker:
func isOwnedByCurrentUser(_ nftEmojiId: String) async -> Bool {
    // Query Cardano wallet for NFT ownership
    let assets = try? await cardanoWallet.getAssets(policyId: echoNFTEmojiPolicyId)
    return assets?.contains { $0.assetId == nftEmojiId } ?? false
}

// Trade transaction (Cardano):
// 1. User A proposes: send A's NFT, receive B's NFT
// 2. User B confirms
// 3. Smart contract (Plutus) atomically: transfer A's NFT to B, B's NFT to A
// 4. Both parties receive ownership confirmation
```

## Blueprints

- Message Reactions, Polls, and Interactive Elements — Defines NFT emoji collection, blockchain verification, trading, marketplace, rarity levels, and unique reaction display

---

### WO-62: Build Call Scheduling System with Calendar Integration

**Blueprint:** Voice and Video Calls with Screen Sharing

## Summary

Build the call scheduling system with calendar integration — users schedule one-time and recurring calls with participants, sync to iOS Calendar, set reminders, and receive automatic call initiation at the scheduled time. Timezone-aware scheduling with conflict detection.

## In Scope

- Call scheduling UI: date/time picker, participant selection by contact DID, call type (voice/video), duration, optional agenda/notes
- iOS Calendar integration: create `EKEvent` in user's calendar via `EventKit` framework; two-way sync (edit in Calendar updates ECHO schedule)
- Reminder notifications: configurable intervals — 15 minutes, 1 hour, 24 hours before scheduled call
- Automatic call initiation at scheduled time via `BGProcessingTask`: fire `WebRTCCallManager.initiateCall` at scheduled time
- One-click join: participants receive APNs push 2 minutes before call → tap to join auto-initiates WebRTC
- Timezone handling: store in UTC, display in each participant's local timezone via `TimeZone.current`
- Conflict detection: check against user's existing calls/meetings; suggest alternative times
- Recurring calls: daily, weekly, monthly patterns with end date
- Agenda sharing: optional encrypted agenda attached to call invitation, visible to all participants

## Out of Scope

- Google Calendar / Outlook integration (iOS native EventKit only for v1)
- Conference room booking
- External video conferencing bridge

## Requirements

Derived from the Voice and Video Calls blueprint.

```swift
struct ScheduledCall: Identifiable, Codable {
    let id: UUID
    let callType: CallType
    let participants: [String]  // DIDs
    let scheduledFor: Date
    let timezone: TimeZone
    var agenda: Data?           // Encrypted
    var isRecurring: Bool
    var recurrenceRule: RecurrenceRule?
    var ekEventIdentifier: String?  // iOS Calendar event ID
}
```

## Blueprints

- Voice and Video Calls with Screen Sharing — Defines call scheduling, calendar integration, reminder notifications, automatic initiation, timezone handling, and participant invitations

---

### WO-63: Implement Granular Permission Management System for Bot Access Control

**Blueprint:** Decentralized Bot Framework and Automation

## Summary

Build the granular bot permission management system — users control exactly which data and capabilities each installed bot can access. Permissions are granted explicitly during installation and revocable at any time. Backend enforces permission boundaries on every bot API call.

## In Scope

- Permission types: `messaging.read`, `messaging.send`, `files.read`, `files.write`, `payments.view`, `payments.initiate`, `contacts.read`, `blockchain.read`, `blockchain.execute`
- Permission granting UI: each permission shown with plain-language explanation and risk level (low/medium/high)
- Per-permission toggle during install and in bot settings afterwards
- Permission revocation: `DELETE /v1/bots/{id}/permissions/{permissionType}` — immediate effect (<5 seconds)
- Backend enforcement: every bot API call checks permission grant; 403 if not granted
- Access transparency: `GET /v1/bots/{id}/access-log` — data type accessed + frequency over last 30 days (no content)
- Permission audit log: `{timestamp, action: grant|revoke, permission, userDID, botDID}`
- Permission history: view all changes over time, restore previous permission set

## Out of Scope

- Bot SDK permission checks (enforced by backend, not SDK)
- Marketplace integration (WO-52)

## Requirements

Derived from the Decentralized Bot Framework blueprint.

**Permission Enforcement:**
```go
func (s *BotPermissionService) Authorize(botDID, userDID, permission string) error {
    granted, err := s.db.HasPermission(botDID, userDID, permission)
    if err != nil || !granted {
        return BotError{Code: "PERMISSION_DENIED", Permission: permission}
    }
    return nil
}
// Called by each bot API endpoint before processing request
```

## Blueprints

- Decentralized Bot Framework and Automation — Defines granular permission types, user-controlled grants, revocation, API-level enforcement, and access transparency

---

### WO-71: Implement Call Metadata Blockchain Anchoring with Zero-Knowledge Proofs

**Blueprint:** Voice and Video Calls with Screen Sharing

## Summary

After a call ends, anchor privacy-preserving call metadata on the Constellation Data L1 layer. Only non-sensitive metadata is anchored: participant count hash, duration, quality metrics, and call timestamp. Participant identities are protected via ZK proofs (Phase 3+). This creates a tamper-evident audit trail for compliance use cases without exposing call content.

## In Scope

- Post-call metadata submission to Data L1 via Metagraph Gateway (within 5 minutes of call end)
- Metadata schema: `{callId, participantCountHash, durationSeconds, avgLatencyMs, packetLossPercent, completedAt}`
- Participant privacy: participant DIDs hashed `H(DID || callSalt)` — not raw DIDs on-chain
- Quality metrics: RTT, packet loss, jitter percentiles from WebRTC stats API
- Phase 3+: ZK proof of caller trust tier (prove tier ≥ N without revealing actual tier or DID)
- Backend endpoint: `POST /v1/calls/{callId}/anchor` — called by call termination handler
- Local storage of anchoring result: `{callId, snapshotHash, snapshotHeight, anchoredAt}` in PostgreSQL for compliance queries

## Out of Scope

- Recording actual call audio/video content
- External compliance monitoring integrations
- Real-time anchoring during calls (post-call only)

## Requirements

Derived from the Voice and Video Calls blueprint.

**Call Metadata Submission:**
```go
// backend/calls/anchoring.go
type CallMetadataSubmission struct {
    Type                 string  // "call_metadata"
    CallID               string  // UUID
    ParticipantCountHash []byte  // H(count || callSalt) — privacy-preserving
    DurationSeconds      int
    AvgLatencyMs         float64
    PacketLossPercent    float64
    CompletedAt          time.Time
    SchemaVersion        int     // Current: 1
    // NEVER: participant DIDs, call content, conversation links
}

func (h *CallAnchoringService) AnchorCallMetadata(callID string, stats CallStats) error {
    countSalt := generateSalt()
    submission := CallMetadataSubmission{
        Type:                 "call_metadata",
        CallID:               callID,
        ParticipantCountHash: sha256(append([]byte(strconv.Itoa(stats.ParticipantCount)), countSalt...)),
        DurationSeconds:      int(stats.Duration.Seconds()),
        AvgLatencyMs:         stats.AvgLatencyMs,
        PacketLossPercent:    stats.PacketLossPercent,
        CompletedAt:          time.Now(),
    }
    txHash, err := h.metagraph.SubmitDataL1(submission)
    if err == nil { h.db.StoreAnchorResult(callID, txHash) }
    return err
}
```

## Blueprints

- Voice and Video Calls with Screen Sharing — Defines call metadata anchoring on blockchain, participant privacy via ZK proofs, quality metrics recording, and compliance audit trail

---

### WO-74: Build Bot Security Auditing and Compliance Verification System

**Blueprint:** Decentralized Bot Framework and Automation

## Summary

Build the automated bot security auditing system that runs before each bot marketplace listing. Scans bot code for common vulnerabilities, validates permission usage, and generates a compliance report. Audit results feed directly into the bot trust score (WO-40).

## In Scope

- Automated code scanning pipeline using industry-standard tools (Semgrep, Gosec for Go bots)
- Vulnerability detection: SQL injection, XSS, buffer overflows, unauthorized data access patterns, hardcoded secrets
- Permission compliance check: verify bot only requests permissions declared in manifest; flag undeclared access attempts
- Data access validation: confirm bot encrypts user data, doesn't log sensitive fields, handles permission denials correctly
- Policy compliance: content guidelines, acceptable use, privacy requirements checklist
- Severity classification: Critical (−50 pts), High (−20 pts), Medium (−10 pts), Low (−5 pts) — fed to trust score (WO-40)
- Audit report: `{botDID, version, findings[], overallResult: approved|rejected|approved_with_warnings, auditedAt}`
- Audit history: version-tracked records for all audits; re-audit required on each version update
- Audit endpoint: `POST /v1/admin/bots/{id}/audit` triggers pipeline; webhook callback when complete

## Out of Scope

- Manual human code review
- Bot marketplace display of audit results (WO-52)
- Trust score computation (WO-40 consumes audit output)

## Requirements

Derived from the Decentralized Bot Framework blueprint.

**Audit Report:**
```go
type BotAuditReport struct {
    BotDID       string
    CodeVersion  string
    Findings     []AuditFinding
    OverallResult string  // "approved", "rejected", "approved_with_warnings"
    AuditedAt    time.Time
}

type AuditFinding struct {
    Severity    string  // "critical", "high", "medium", "low"
    Category    string  // "sql_injection", "unauthorized_access", etc.
    Description string
    LineNumber   int
}
```

## Blueprints

- Decentralized Bot Framework and Automation — Defines bot security audit before marketplace listing, vulnerability scanning, permission verification, data access validation, policy compliance, and audit reports

---

### WO-80: Build Enterprise Onboarding Portal with Document Submission

**Blueprint:** Enterprise Organization Profiles with Verified Status

## Summary

Build the enterprise onboarding portal for organizations seeking verified enterprise status — a web-based portal where organizations submit business registration certificates, regulatory licenses, and executive authorization documents. Tracks verification progress through a multi-stage review workflow.

## In Scope

- Organization registration: name, contact info, business type selection (bank, corporation, government, non-profit), jurisdiction
- Document upload: PDF/JPG/PNG, max 10MB per file; required document types: business registration certificate, regulatory license, executive authorization letter, compliance certifications
- Document types by tier: Basic Enterprise (business registration), Regulated Entity (+ banking license, AML certification), Government Agency (+ government ID documents)
- Submission status tracking: `pending → under_review → verified → rejected`; email notifications at each transition
- Status dashboard: verification progress per stage, estimated completion, next required action
- Encrypted document storage: AES-256-GCM, encrypted at rest, access logged
- Communication thread: organization can message verification team; threaded conversation with file attachments
- Onboarding completion: on approval → trigger enterprise profile creation (WO-98) + institutional DID creation (WO-86)

## Out of Scope

- Actual regulatory database verification (WO-152 handles the automated checks)
- Enterprise profile management UI (WO-98)
- Payment processing for verification fees

## Requirements

Derived from the Enterprise Organization Profiles blueprint.

```go
type EnterpriseOnboardingApplication struct {
    ApplicationID   string
    OrganizationName string
    BusinessType    string
    Jurisdiction    string
    Documents       []OnboardingDocument
    Status          ApplicationStatus  // pending, under_review, verified, rejected
    SubmittedAt     time.Time
    ReviewNotes     string
}
```

## Blueprints

- Enterprise Organization Profiles with Verified Status — Defines enterprise onboarding portal, documentation requirements, multi-stage review, status tracking, and compliance auditing

---

### WO-85: Create Revenue Sharing and Monetization Infrastructure for Bot Developers

**Blueprint:** Decentralized Bot Framework and Automation

## Summary

Build the bot developer revenue sharing infrastructure — ECHO token payment processing for bot usage, subscription billing, transaction fee collection, and developer payout automation. All payments flow through Constellation Currency L1 using `AllowSpend` and `SpendTransaction` v3 primitives.

## In Scope

- One-time bot payment: `SpendTransaction` from user to bot developer wallet when user pays for bot usage
- Subscription billing: `AllowSpend` authorization for recurring monthly/quarterly/annual charges; automated renewal
- Transaction fee collection: for bots facilitating trades, collect configurable fee (0.1%–5%) as `FeeTransaction`
- Platform revenue split: 20% platform fee, 80% to developer via automatic split in `SpendTransaction`
- Developer revenue dashboard: total earnings, breakdown by payment type, user acquisition count, payout history
- Automatic payouts: batch settled daily at midnight UTC to developer wallets
- Subscription management: user can cancel → immediate cancellation, prorated refund for unused period
- Compliance recording: `{transactionType, amount, userDID_hash, botDID, timestamp}` — anonymized, stored for 7 years

## Out of Scope

- Marketplace payment UI (WO-52)
- ECHO token infrastructure (Data Layer work orders)
- Tax filing services

## Requirements

Derived from the Decentralized Bot Framework blueprint.

**Payment Split:**
```go
// When user pays for bot usage, atomic split via Currency L1:
// AtomicAction: [
//   SpendTransaction(from: userDID, to: developerWallet, amount: price * 0.80),
//   FeeTransaction(from: userDID, to: platformTreasury, amount: price * 0.20)
// ]
```

## Blueprints

- Decentralized Bot Framework and Automation — Defines ECHO token payments, subscription models, transaction fees, revenue sharing, developer dashboard, and compliance recording

---

### WO-86: Implement Institutional DID Management System

**Blueprint:** Verified Financial Institution Integration

## Summary

Implement the institutional DID management system for financial institutions — creating institutional DIDs with enhanced verification requirements, multi-signature executive authorization, regulatory compliance documentation, and institutional DID resolution that distinguishes them from individual user DIDs.

## In Scope

- Institutional DID format: `did:prism:cardano:institution-{id}` — distinct from user DIDs (`did:prism:cardano:{user-id}`)
- DID document fields: `institutionName`, `regulatoryLicenseHash`, `jurisdictions[]`, `executivePublicKeys[]` (multi-sig)
- Multi-signature activation: DID requires signatures from N-of-M designated C-level executives before activation (default: 2-of-3)
- Regulatory compliance documentation upload: PDF/image hashes stored in DID document on Cardano
- DID resolution endpoint: `GET /v1/identity/institutional/{did}` — returns full institutional DID document + verification status
- Institutional DID type flag: resolver distinguishes `institutional` vs `individual` DID type
- Audit trail: all DID create/update/deactivate operations logged in Cardano transaction metadata
- DID recovery: re-issue requires all M executives to re-sign

## Out of Scope

- Individual user DID management (WO-180)
- Specific regulatory validation logic (WO-152)

## Requirements

Derived from the Verified Financial Institution Integration blueprint.

**Institutional DID Document:**
```json
{
  "@context": "https://www.w3.org/ns/did/v1",
  "id": "did:prism:cardano:institution-jpmorgan-001",
  "type": "InstitutionalDID",
  "institutionName": "JPMorgan Chase",
  "regulatoryLicenseHash": "<sha256-of-license-doc>",
  "jurisdictions": ["US", "EU"],
  "executivePublicKeys": ["<key1>", "<key2>", "<key3>"],
  "signaturesRequired": 2
}
```

## Blueprints

- Verified Financial Institution Integration — Defines institutional DID creation with regulatory compliance documentation and multi-signature authorization requirements

---

### WO-89: Implement Multi-Stage Business Verification System

**Blueprint:** Enterprise Organization Profiles with Verified Status

## Summary

Build the multi-stage business verification engine — automated validation of incorporation certificates, regulatory licenses (FDIC, AML, banking), and executive multi-signature authorization. Assigns verification tier, records decision on Cardano, and creates institutional DID (WO-86) on approval.

## In Scope

- Stage 1: Business registration validation — verify business name against incorporation certificate; check business address validity; confirm active legal entity status
- Stage 2: Regulatory compliance verification — FDIC database query for banking institutions; state/federal license database check; AML certification validation
- Stage 3: Executive authorization — collect P-256 digital signatures from N designated C-level executives (email-invited); verify identity via Daon/1Kosmos for each signer
- Tier classification: `{basicEnterprise: business_registration_only, regulatedEntity: + regulatory_license, governmentAgency: + government_authority_docs}`
- Blockchain recording: verification decision anchored on Cardano `{institutionDID, verificationTier, verifiedAt, verificationHash}` — immutable and portable
- Automated workflow with manual escalation: automated checks where possible; failed checks → manual review queue
- Status API: `GET /v1/enterprise/applications/{id}/status` — per-stage status with notes

## Out of Scope

- Enterprise profile UI (WO-98)
- Employee account management (WO-107)

## Requirements

Derived from the Enterprise Organization Profiles blueprint.

```go
type VerificationStage string
const (
    StageBusinessRegistration VerificationStage = "business_registration"
    StageRegulatoryCompliance VerificationStage = "regulatory_compliance"
    StageExecutiveAuthorization VerificationStage = "executive_authorization"
)
type VerificationTier string
const (
    TierBasicEnterprise  VerificationTier = "basic_enterprise"
    TierRegulatedEntity  VerificationTier = "regulated_entity"
    TierGovernmentAgency VerificationTier = "government_agency"
)
```

## Blueprints

- Enterprise Organization Profiles with Verified Status — Defines multi-stage verification process, regulatory compliance requirements, executive authorization, tier classification, and blockchain recording

---

### WO-91: Implement Persona-Aware Conversation and Messaging System

**Blueprint:** Multiple Personas with Selective Visibility

## Summary

Build the persona-aware conversation and messaging system. When a user initiates a conversation or joins a group, they select which persona to present. Messages are sent and received under the chosen persona. Conversation histories are completely isolated per persona — messages from Persona A cannot be seen when viewing Persona B's conversations.

## In Scope

- Persona selector sheet: shown when starting a new conversation (if user has multiple personas)
- `personaId` embedded in message metadata (encrypted, not visible to relay)
- Separate conversation contexts in SwiftData per persona: conversations linked to a `personaId`
- Persona-specific notification routing: push notifications labeled with persona name
- Auto-suggest persona: backend suggests most-used persona for a given contact based on conversation history
- Persona switching mid-conversation: confirmation sheet → notify contact that persona has changed
- Separate per-persona: read receipts, typing indicators, call history, file storage, reactions
- Conversation archival preserves persona isolation even after persona deletion

## Out of Scope

- Persona creation (WO-72)
- Selective visibility permissions (WO-82)
- Blockchain anchoring (WO-122)

## Requirements

Derived from the Multiple Personas blueprint.

**Conversation Isolation:**
```swift
// Domain/Models/Conversation.swift
struct Conversation: Identifiable, Codable {
    let id: String
    let personaId: UUID         // Which persona this conversation belongs to
    let participantDIDs: [String]
    var lastMessage: Message?
    // Messages in this conversation are ONLY visible when viewing this personaId context
}

// Conversation list is filtered by current active persona:
func conversations(for personaId: UUID) async -> [Conversation] {
    return await database.fetchConversations(personaId: personaId)
}
```

**Message Metadata:**
```swift
struct MessageMetadata: Codable {
    let senderPersonaId: UUID    // Embedded in E2E encrypted payload (relay cannot read)
    let recipientPersonaId: UUID? // nil for master DID conversations
}
```

## Blueprints

- Multiple Personas with Selective Visibility — Defines separate conversation threads per persona, persona selection on conversation initiation, isolated message history, persona-specific features, and conversation archival isolation

---

### WO-93: Build Secure Banking API Integration Layer

**Blueprint:** Verified Financial Institution Integration

## Summary

Build the secure banking API integration layer — PCI DSS and SOC 2 Type II compliant API endpoints that allow financial institutions to send transaction alerts, fraud notifications, and customer service messages through the ECHO platform using their institutional DID for authentication.

## In Scope

- `POST /v1/financial/messages` — bank sends E2E encrypted message to customer's DID; requires institutional DID bearer token
- `POST /v1/financial/alerts` — automated fraud/transaction alerts with structured payload
- Institutional DID authentication: JWT-like token signed with institutional DID P-256 key; validated by Identity Service
- API key management: generate/rotate/revoke API keys per institution; scoped to specific message types
- PCI DSS compliance controls: no raw card data in API payloads; TLS 1.3 mandatory; all payloads encrypted
- SOC 2 Type II controls: comprehensive access logging (all requests, responses with redacted PII), monitoring, 90-day log retention
- Rate limiting: per-institution rate limits (configurable, default 1000 messages/minute)
- Payload validation: enforce required fields, message format, encryption standards

## Out of Scope

- Specific bank core system integrations
- Message content encryption (standard Kinnami pipeline, WO-103)

## Requirements

Derived from the Verified Financial Institution Integration blueprint.

**Alert Payload:**
```go
type BankAlertRequest struct {
    CustomerDID   string    `json:"customer_did"`
    AlertType     string    `json:"alert_type"`  // "transaction_verify", "fraud_alert", "account_security"
    EncryptedBody []byte    `json:"encrypted_body"`  // Kinnami encrypted
    BankDID       string    `json:"bank_did"`
    Signature     []byte    `json:"signature"`    // P-256 signature
    Nonce         string    `json:"nonce"`         // Anti-replay
}
```

## Blueprints

- Verified Financial Institution Integration — Defines PCI DSS/SOC 2 compliance requirements, API authentication, rate limiting, and audit logging

---

### WO-95: Implement ECHO Token Balance Display with Real-time Updates

**Blueprint:** User Rewards Tracker on Profile

## Summary

Build the ECHO token balance display component in the iOS Wallet tab — shows real-time total, available, staked, delegated, and pending reward balances via Stargazer SDK. Also renders the `DailyRewards` section showing today's auto-scaled reward rate, network-level budget consumption, and trust tier reward multiplier (1.0×–3.0×).

## In Scope

- `BalanceCard` SwiftUI component: total balance (large text) + USD equivalent (PacaSwap TWAP rate)
- `BalanceBreakdown` component: available (spendable), staked (in TokenLock), delegated (in StakeDelegation), pending rewards
- Real-time balance: fetch from Stargazer SDK every 30 seconds + on WebSocket snapshot confirmation
- Balance verification badge: green checkmark if confirmed in metagraph snapshot; yellow if cache (>30s old)
- Offline mode: show last-known balance with "Last updated X minutes ago" and warning icon
- Privacy mode toggle: mask balance as `**** ECHO` for screen privacy
- Balance history: last 30-day sparkline chart in wallet tab
- **`DailyRewards` section** showing current auto-scaled reward state:
  - `currentAutoScaledRate`: today's effective per-message rate (starts at 0.1 ECHO, decays with volume)
  - `networkDailyBudget`: 219,178 ECHO/day in Year 1 (80M ÷ 365)
  - `networkDistributedToday`: how much of today's budget has been distributed across all users
  - `trustTierRewardMultiplier`: user's reward scale multiplier (1.0× to 3.0× — NOTE: reward scale, distinct from governance scale 0.0–2.0×)
- Founder vesting section: visible only if user has founder `TokenLock` with cliff/vesting metadata

## Out of Scope

- Earnings breakdown charts (WO-106)
- Transaction history (WO-124)
- Staking/delegation UI actions (separate wallet WOs)

## Requirements

From the Frontend blueprint (canonical `DailyRewards` struct):

```swift
struct DailyRewards {
    let messaging: Decimal              // Earned today from messaging
    let currentAutoScaledRate: Decimal  // Current per-message rate (auto-scales with network volume)
    let referrals: Decimal              // Referral bonuses today
    let staking: Decimal                // Staking rewards today
    let total: Decimal                  // Total earned today
    let claimableTypes: [String]        // Reward types ready to claim
    let trustTierRewardMultiplier: Float // 1.0x (Tier 1) → 3.0x (Tier 5) — REWARD scale
    let networkDailyBudget: Decimal     // Today's total emission budget (Year 1 ≈ 219,178 ECHO/day)
    let networkDistributedToday: Decimal // Total distributed across all users today

    static var empty: DailyRewards {
        DailyRewards(
            messaging: 0, currentAutoScaledRate: 0.1,
            referrals: 0, staking: 0, total: 0,
            claimableTypes: [],
            trustTierRewardMultiplier: 1.0,
            networkDailyBudget: 219178,
            networkDistributedToday: 0
        )
    }
}
```

**Balance components:**
```swift
struct BalanceCard: View {
    let balance: Decimal
    let usdValue: Decimal
    let verificationStatus: BalanceVerificationStatus  // .verified, .cached, .offline
}
struct BalanceBreakdown: View {
    let available: Decimal
    let staked: Decimal
    let delegatedTo: ValidatorInfo?
    let pending: Decimal
}
```

**Network budget display:** Show "X ECHO distributed today of Y ECHO daily budget" to give users context on network-level activity and their share of rewards.

## Blueprints

- Frontend — Defines `DailyRewards` struct with auto-scaled rate fields, `networkDailyBudget`, `networkDistributedToday`, and `trustTierRewardMultiplier` (1.0×–3.0× reward scale)
- User Rewards Tracker on Profile — Defines ECHO balance display with real-time updates and balance history

---

### WO-98: Create Enterprise Profile Management with Verification Badges

**Blueprint:** Enterprise Organization Profiles with Verified Status

## Summary

Build the enterprise profile display and management UI — showing verified organizational information with tier-specific badges, employee hierarchy, branded communication channel customization, and customer-facing profile discovery. All on iOS; backend serves profile data from Identity Service.

## In Scope

- `EnterpriseProfileView`: organization name, logo, industry, verification tier badge, business address
- Verification tier badges: Basic Enterprise (bronze shield), Regulated Entity (silver shield + license icon), Government Agency (blue star)
- Organizational hierarchy: owner → admins → employees with role labels and individual verification status
- Branded channel customization: custom logo upload (max 2MB PNG/JPG), color scheme, standardized message template library
- Profile management dashboard (admin only): update business info, manage employee accounts, view verification history
- Customer-facing discovery: search verified organizations by name, industry, or verification tier; tap to view profile and official channels
- Anti-phishing protection: unverified senders attempting to impersonate → `⚠️ Unverified — this may be a phishing attempt`

## Out of Scope

- Verification processing (WO-89)
- RBAC enforcement (WO-107)
- Cryptographic signatures (WO-119)

## Requirements

Derived from the Enterprise Organization Profiles blueprint.

```swift
struct EnterpriseProfile {
    let institutionalDID: String
    let organizationName: String
    let logoURL: URL?
    let verificationTier: VerificationTier  // basicEnterprise, regulatedEntity, governmentAgency
    let complianceLevel: String
    let industry: String
    let employees: [EnterpriseEmployee]
    let communicationChannels: [EnterpriseChannel]
}
```

## Blueprints

- Enterprise Organization Profiles with Verified Status — Defines enterprise profile UI, verification badges, organizational hierarchy, branded channels, and customer-facing discovery

---

### WO-106: Build Earnings Breakdown Dashboard with Interactive Charts

**Blueprint:** User Rewards Tracker on Profile

## Summary

Build the earnings breakdown dashboard — interactive charts showing ECHO token earnings by source (messaging rewards, referrals, staking yields) with daily/weekly/monthly time frames. Includes trend comparison and optimization recommendations. Part of the Wallet tab.

## In Scope

- `EarningsBreakdownView`: pie/bar chart showing percentage contribution per source
- Sources: `messaging` (0.1 ECHO per message, capped by trust tier), `referrals` (50 ECHO per verified referral), `staking` (5–15% APY), `verification_reward` (100 ECHO one-time)
- Time period selector: 7 days, 30 days, 90 days, 1 year
- Interactive line chart: daily earnings per source using Apple Charts (`Swift Charts` framework)
- Peak earnings days: highlight top 3 days with annotation
- Period comparison: current vs. previous period delta (`+12% vs. last month`)
- Optimization recommendations: if messaging earnings < expected, suggest "Complete Tier 3 verification to increase your messaging reward cap"
- CSV export: `GET /v1/rewards/earnings?from={date}&to={date}&format=csv`

## Out of Scope

- Transaction history detail (WO-124)
- Achievement milestones (WO-116)
- Goal setting (WO-142)

## Requirements

Derived from the User Rewards Tracker blueprint.

```swift
struct DailyEarnings: Codable {
    let date: Date
    let messaging: Decimal
    let referrals: Decimal
    let staking: Decimal
    let total: Decimal
}
```

## Blueprints

- User Rewards Tracker on Profile — Defines earnings breakdown by source, daily/weekly/monthly trends, peak activity, export, and optimization recommendations
- Frontend — Specifies `DailyRewards` struct and trust tier multiplier display in wallet

---

### WO-107: Build Role-Based Access Control System for Enterprise Users

**Blueprint:** Enterprise Organization Profiles with Verified Status

## Summary

Build the role-based access control system for enterprise users — defining roles (Owner/Admin/Representative/Viewer), managing employee linking to enterprise DID, enforcement of privilege levels per role, and employee provisioning/deprovisioning with comprehensive audit logging.

## In Scope

- Role definitions: `Owner` (full admin + delete), `Administrator` (user management + settings), `CustomerServiceRep` (messaging only), `ComplianceOfficer` (audit + reports), `Viewer` (read-only)
- Employee account linking: employee submits `{employeeDID, enterpriseDID, role}` → admin approves; employee's DID document updated with enterprise affiliation credential on Cardano
- Permission enforcement: backend Identity Service checks employee role before allowing each enterprise API action
- Role management UI (admin only): list employees, assign/change roles, remove employees
- Verification level requirements: `CustomerServiceRep` requires Tier 2+; `ComplianceOfficer` requires Tier 3+; `Administrator` requires Tier 3+
- Employee provisioning: on hire, admin invites employee DID; employee accepts → role activated
- Employee deprovisioning: on departure, admin removes → Cardano credential revoked; access removed immediately
- Audit log: `{actorDID, targetEmployeeDID, action: assign_role|revoke_role|remove_employee, oldRole, newRole, timestamp}` — immutable

## Out of Scope

- Active Directory / SAML integration (future enterprise feature)
- Cryptographic signatures (WO-119)
- Communication channel enforcement (WO-119)

## Requirements

Derived from the Enterprise Organization Profiles blueprint.

```go
type EnterpriseRoleAssignment struct {
    EmployeeDID    string
    EnterpriseDID  string
    Role           EnterpriseRole
    AssignedAt     time.Time
    AssignedBy     string  // Admin DID
    IsActive       bool
}
```

## Blueprints

- Enterprise Organization Profiles with Verified Status — Defines organizational hierarchy, role-based access controls, employee linking, provisioning workflows, and audit logging

---

### WO-110: Create Fraud Alert Channel Management System

**Blueprint:** Verified Financial Institution Integration

## Summary

Build the fraud alert channel management system — financial institutions create dedicated fraud alert channels, customers opt-in, and banks send automated alerts via the API. Customers confirm or reject alerts using their DID P-256 signature + Face ID/Touch ID, creating an immutable authorization record.

## In Scope

- Fraud alert channel creation: bank creates channel linked to institutional DID; `POST /v1/financial/channels`
- Customer opt-in management: `POST /v1/financial/channels/{channelId}/subscribe` — customer subscribes with their DID
- Alert types: `transaction_verification` (confirm/deny transaction), `account_security` (suspicious activity), `fraud_alert` (unauthorized activity reported)
- Automated alert delivery: bank calls `POST /v1/financial/alerts` → encrypted alert relayed to subscribed customer via standard relay
- Customer response: iOS shows alert with "Confirm" / "Deny" buttons; biometric auth + DID P-256 sign response → submit to bank via API
- Alert history: `GET /v1/financial/alerts?institutionDID={did}&status={status}` — delivery status, customer response, timestamps
- Alert delivery confirmation to bank: after customer signs response, webhook notification to bank with `{alertId, responseType, signatureHash}`
- Response time tracking: `{alertSent, alertDelivered, customerResponded}` timestamps

## Out of Scope

- Fraud detection algorithms
- Bank fraud monitoring system integrations
- Customer notification preferences (uses global notification settings)

## Requirements

Derived from the Verified Financial Institution Integration blueprint.

```go
type FraudAlertCustomerResponse struct {
    AlertID    string
    CustomerDID string
    Response   string  // "confirmed" or "denied"
    Signature  []byte  // Customer's P-256 DID signature
    Timestamp  time.Time
    // Creates immutable proof: customer authorized or rejected this transaction
}
```

## Blueprints

- Verified Financial Institution Integration — Defines fraud alert channels, customer opt-in management, automated alerts, cryptographic customer confirmation, and alert audit trail

---

### WO-111: Build Persona Trust Scoring and Verification System

**Blueprint:** Multiple Personas with Selective Visibility

## Summary

Integrate personas with the trust scoring system and build the persona-specific verification badge system. The master identity's trust score applies to all personas. Each persona can earn additional context-specific verification badges (professional credentials, gaming achievements, community badges) that are displayed independently.

## In Scope

- Master DID trust score inheritance: all personas display the same trust score from the master identity
- Per-persona `VerificationBadge` collection: `{personaId, badgeType, earnedAt, credentialRef}`
- Badge types: `professionalCredential` (from employer/organization), `gamingAchievement` (from gaming platforms), `communityMembership` (from group admins)
- Badge earning: organization or group admin issues a signed badge credential linked to a specific persona
- Badge display: in persona profile view and in conversations as separate from trust tier badge
- Badge independence: a badge on Professional persona is NOT visible on Gaming persona (by design)
- Badge credential portability: user can request to transfer a badge to another persona if the issuer allows
- Credential verification: badges are W3C VC-signed by the issuer DID and stored on Cardano

## Out of Scope

- Master trust score calculation (WO-181)
- Persona creation (WO-72)
- ZK-proof-based trust verification (WO-183)

## Requirements

Derived from the Multiple Personas blueprint.

**Persona Badge System:**
```swift
struct PersonaVerificationBadge: Identifiable, Codable {
    let id: UUID
    let personaId: UUID
    let badgeType: PersonaBadgeType
    let issuerDID: String        // Organization, gaming platform, or group admin DID
    let credentialRef: String    // Cardano VC reference
    let earnedAt: Date
    var isDisplayed: Bool

    enum PersonaBadgeType: Codable {
        case professionalCredential(title: String, organization: String)
        case gamingAchievement(game: String, achievement: String)
        case communityMembership(groupName: String)
    }
}
// Trust score displayed for all personas: same score from master DID
// TrustBadge shows: "Trust: [score] (from your main identity)"
// PersonaVerificationBadge shows: context-specific earned credentials
```

## Blueprints

- Multiple Personas with Selective Visibility — Defines master identity trust score inheritance, per-persona verification badges, badge independence, professional/gaming/community credentials, and badge portability

---

### WO-116: Create Achievement Milestone System with Badge Unlocking

**Blueprint:** User Rewards Tracker on Profile

## Summary

Build the achievement milestone system — predefined milestones unlock badges and bonus multipliers as users progress on the platform. Achievements are tracked locally and verified against backend + blockchain records to prevent tampering.

## In Scope

- Achievement definitions: "First 1000 Messages" (+0.1x multiplier), "Trusted Verifier" (Tier 3+, +0.2x), "Super Referrer" (10 verified referrals, +0.3x), "Active Staker" (1000+ ECHO staked, +0.2x), "Governance Participant" (5+ votes, +0.1x)
- Progress tracking: `{achievementId, currentProgress, targetProgress, completedAt?}` stored locally + synced to backend
- Badge unlock: `AchievementUnlockView` with celebration animation on completion; badge stored in `UserProfile.achievements`
- Multiplier activation: on unlock, update earning multiplier in Trust Service; applied to all future reward calculations
- Progress bars per achievement in rewards tab
- Achievement history: list of unlocked badges with dates and multipliers earned
- Cryptographic verification: achievement record anchored on Data L1 `{type: "achievement", userDID_hash, achievementId, unlockedAt}` — prevents retroactive manipulation
- Achievement categories: messaging, verification, referral, community — shown in separate sections

## Out of Scope

- Custom user-defined achievements
- Leaderboards (WO-147)
- Trust score calculation (WO-181)

## Requirements

Derived from the User Rewards Tracker blueprint.

```swift
struct Achievement: Identifiable, Codable {
    let id: String                 // "first_1000_messages"
    let name: String
    let category: AchievementCategory
    var currentProgress: Int
    let targetProgress: Int
    var completedAt: Date?
    let bonusMultiplier: Float     // 0.1, 0.2, 0.3, etc.
    var isUnlocked: Bool { completedAt != nil }
}
```

## Blueprints

- User Rewards Tracker on Profile — Defines achievement milestones, badge unlocking, bonus multipliers, progress tracking, and blockchain verification

---

### WO-117: Build Customer Service Channel with Trust Score Integration

**Blueprint:** Verified Financial Institution Integration

## Summary

Build the customer service channel for verified financial institutions — verified bank representatives can directly message customers through dedicated channels showing trust tier badges. Trust score-based routing prioritizes high-tier customers to premium support. ZK proofs confirm customer identity without revealing personal financial data.

## In Scope

- Bank creates customer service channel linked to institutional DID: `POST /v1/financial/channels?type=customer_service`
- Bank representative account linking: representative DID linked to institutional DID as authorized agent
- Representative verification display in chat: `VerificationBadge` showing bank institution name + role (Customer Service)
- Trust score-based service levels: Tier 4 customers → premium support channel (faster SLA); Tier 1–3 → standard channel
- ZK proof for customer identity: customer proves "I am a customer of this bank" without revealing account number (Midnight SDK, Phase 3+)
- Response time tracking: `{inquiryReceivedAt, firstResponseAt}` — logged per interaction; SLA monitoring
- Service quality metrics: per-representative response times, conversation resolution rates
- Interaction history: full conversation logs per customer-institution pair, accessible to both parties

## Out of Scope

- Bank representative training systems
- External bank CRM integration
- Customer satisfaction surveys

## Requirements

Derived from the Verified Financial Institution Integration blueprint.

```swift
// iOS: Customer service conversation shows institution verification badge
struct InstitutionalMessageSenderBadge: View {
    let institutionDID: String
    let institutionName: String
    let representativeRole: String
    var body: some View {
        HStack {
            Image(systemName: "building.columns.fill").foregroundColor(.blue)
            Text("\(institutionName) · \(representativeRole)")
            VerificationBadge(status: .institutionVerified)
        }
    }
}
```

## Blueprints

- Verified Financial Institution Integration — Defines customer service channels, verified representative display, trust score prioritization, ZK proof for customer identity, and service quality metrics

---

### WO-119: Implement Secure Enterprise Communication Channels with Compliance Recording

**Blueprint:** Enterprise Organization Profiles with Verified Status

## Summary

Build verified enterprise communication channels — dedicated channels for enterprises to communicate with customers with E2E encryption, institutional P-256 signatures on every message, blockchain-anchored compliance records, and anti-phishing trust indicators. Satisfies FDIC and SOX regulatory examination requirements.

## In Scope

- Enterprise communication channel with institutional branding: organization name, logo, verification tier badge displayed in channel header
- E2E encrypted messages: same Kinnami pipeline as all messages; bank cannot read customer replies, platform cannot read either party
- Institutional P-256 digital signature on every outbound message from enterprise: `sign(messageHash, institutionalKey)`
- Customer signature verification indicator: `✓ Authenticated by [Organization]` in message bubble
- Official announcement messages: required for financial disclosures and legal notifications — flagged with blockchain anchoring in compliance mode
- Compliance recording: `{messageHash, institutionDID, recipientDID_hash, timestamp}` anchored on Data L1 for each compliance-mode message
- Message template system: predefined approved templates with `{templateId, content, requiresSignature: true}` — approved by compliance officers
- Communication audit dashboard: message history, signature verification rate, compliance record count, regulatory readiness export

## Out of Scope

- External messaging platform integrations
- Marketing campaign features

## Requirements

Derived from the Enterprise Organization Profiles blueprint.

```swift
struct EnterpriseMessage {
    let institutionDID: String
    let encryptedContent: Data          // Kinnami E2E
    let institutionalSignature: Data    // P-256 sig
    let isComplianceMode: Bool          // Anchored on Data L1
    let templateId: String?             // For standardized messages
}
```

## Blueprints

- Enterprise Organization Profiles with Verified Status — Defines secure enterprise communication channels, cryptographic signatures, compliance recording, blockchain anchoring, and anti-fraud indicators

---

### WO-122: Implement Blockchain Anchoring and Cryptographic Privacy for Personas

**Blueprint:** Multiple Personas with Selective Visibility

## Summary

Implement blockchain anchoring for persona integrity and cryptographic privacy using ZK proofs. Per-persona messages are anchored in the standard Merkle batch pipeline. Persona existence is proven with privacy-preserving commitment hashes on Data L1 — contacts cannot discover personas they haven't been granted access to.

## In Scope

- Per-persona message commitment anchoring: messages sent as any persona flow through the standard `AnchoringBatcher` with `personaId` in the commitment metadata (NOT visible on-chain)
- Privacy-preserving persona registration on Data L1: submit `H(personaId || masterDID || nonce)` — proves persona exists without revealing which persona or who owns it
- ZK-proof-based persona existence proof: user can prove "I have a verified professional persona" without revealing persona details (Phase 3+ via Midnight SDK)
- Persona operation audit trail on Cardano: persona creation/deletion/modification events recorded as Cardano transaction metadata (only `{masterDID, operationType, timestamp}` — no persona details)
- Persona-specific encryption keys: separate Kinnami key agreement per persona (so conversations as Persona A and Persona B are cryptographically separate)

## Out of Scope

- Persona creation UI (WO-72)
- Permission granting (WO-82)
- Message relay infrastructure (WO-4)
- ZK proof generation infrastructure (WO-183)

## Requirements

Derived from the Multiple Personas blueprint.

**Privacy-Preserving Persona Registration:**
```go
// Submit to Data L1 on persona creation:
type PersonaRegistrationCommitment struct {
    Type       string // "persona_registration"
    Commitment []byte // H(personaId || masterDID || salt) — 32 bytes
    SchemaVersion int
    // NEVER: personaId in plaintext, displayName, avatar, bio
}
// Data L1 validators: verify structure, authorized sender DID
// Network observers see: "a persona was registered" — not which persona or whose
```

**Persona-Specific E2E Keys:**
```swift
// Each persona gets a distinct X25519 key pair for message encryption
// Stored in Keychain keyed by personaId
// Key agreement for Persona A uses Persona A's keypair
// Key agreement for Persona B uses Persona B's keypair
// Backend only sees: encrypted blobs + which personaId they belong to (encrypted in payload)
```

## Blueprints

- Multiple Personas with Selective Visibility — Defines per-persona message anchoring, ZK proofs for privacy, persona existence privacy, and immutable audit trails

---

### WO-124: Develop Transaction History with Cryptographic Proof Verification

**Blueprint:** User Rewards Tracker on Profile

## Summary

Build the ECHO token transaction history view with cryptographic proof verification. Users see all reward distributions, staking events, and transfer transactions with their metagraph snapshot reference for independent verification. Supports CSV/JSON/PDF export for tax reporting.

## In Scope

- Transaction list: `{txId, type, amount, source, timestamp, snapshotHash?, status}`
- Transaction types: `messaging_reward`, `referral_reward`, `staking_reward`, `verification_reward`, `token_transfer`, `stake_lock`, `stake_unlock`
- Per-transaction proof verification: tap transaction → shows `{snapshotHash, snapshotHeight, DAG Explorer link}` — user can independently verify on-chain
- Verification status badge: `✓ Verified` (anchored in finalized snapshot) or `⏳ Pending`
- Filtering: by date range, amount range, transaction type
- Search by transaction ID
- Export: CSV (all fields + tx hash), JSON, PDF (formatted report with totals by category)
- Tax summary: annual total by reward type, exportable for tax reporting
- Backend: `GET /v1/rewards/transactions?page={p}&type={type}&from={date}&to={date}` — paginated, indexed

## Out of Scope

- Balance display (WO-95)
- Earnings charts (WO-106)
- External wallet transaction import

## Requirements

Derived from the User Rewards Tracker blueprint.

```swift
struct RewardTransaction: Identifiable, Codable {
    let id: String          // Metagraph transaction ID
    let type: RewardType
    let amount: Decimal     // ECHO
    let source: String
    let timestamp: Date
    let snapshotHash: String?  // Nil if pending
    let isVerified: Bool
}
```

## Blueprints

- User Rewards Tracker on Profile — Defines transaction history with cryptographic proofs, export for tax reporting, and blockchain verification

---

### WO-127: Develop Token Staking System with Dynamic APY Management

**Blueprint:** ECHO Token Reward System and Incentive Economy

## Summary

Implement the ECHO token staking system using Tessellation v3 `TokenLock` and `StakeDelegation` primitives on the Constellation Currency L1. Users lock ECHO for 30–365 days to earn staking APY. Delegation to validators via `StakeDelegation` routes staking rewards through validator performance. Unstaking uses `WithdrawLock` with 14-day cooldown.

## In Scope

- Staking tiers via `TokenLock`:
  - Bronze: 30 days, 5% APY
  - Silver: 90 days, 8% APY
  - Gold: 180 days, 12% APY
  - Platinum: 365 days, 15% APY
- Minimum stake: 100 ECHO (enforced by Currency L1 validator)
- `StakeDelegation` to validators: choose validator → instantly delegate staked ECHO; no cooldown to change validator
- `WithdrawLock` request: 14-day cooldown before staked ECHO is released (enforced by Currency L1)
- Daily staking reward distribution: auto-calculated and distributed from User Rewards pool (Currency L1 automated)
- Staking position query: `GET /v1/staking/positions` — lists all `TokenLock` positions with amount, tier, unlockAt, accruedRewards
- iOS staking flow: tap [Stake] → amount + tier selection → review → biometric confirm → Stargazer SDK `submitTokenLock`

## Out of Scope

- Validator node operation (WO-138/WO-153)
- Cross-chain staking
- Liquid staking derivatives

## Requirements

Derived from the ECHO Token Reward System and Data Layer blueprints.

**Staking via Stargazer SDK (iOS):**
```swift
try await stargazer.submitTokenLock(TokenLockRequest(
    token: .echo,
    amount: amount,
    tier: tier.rawValue,
    duration: tier.durationDays  // 30, 90, 180, or 365
))
// Currency L1 enforces lock period; WithdrawLock triggers 14-day cooldown
```

## Blueprints

- Data Layer — Defines `TokenLock`, `StakeDelegation`, `WithdrawLock` Tessellation v3 primitives and staking mechanics
- Frontend — Specifies staking UI flow, tier selection, and delegation browser in WalletTab

---

### WO-128: Implement Transaction Verification with Biometric Authentication

**Blueprint:** Verified Financial Institution Integration

## Summary

Implement the transaction verification system with biometric authentication — customers authorize financial institution requests (payments, account changes, fraud responses) using iOS Face ID/Touch ID + DID P-256 signature. Creates an immutable authorization proof anchored on Cardano as evidence for dispute prevention.

## In Scope

- Transaction authorization UI: when bank sends authorization request, iOS shows `{transactionDescription, amount, merchantName}` → "Authorize with Face ID" button
- Biometric prompt via `LAContext.evaluatePolicy(.deviceOwnerAuthenticationWithBiometrics)` → success triggers P-256 signing
- Authorization signature: `DID.sign(SHA256(transactionId + nonce + timestamp), reason: "Authorize transaction")`
- Immutable proof creation: `{transactionId, customerDID, signature, biometricVerified: true, timestamp}` anchored on Cardano via Identity Service
- Anti-replay nonce: server-generated nonce per authorization request; single-use, 5-minute expiry
- Multiple transaction types: `{payment_authorization, account_change_approval, fraud_response, document_signing}`
- Authorization proof export: customer can request `{transactionId, proof}` for dispute documentation
- Dispute prevention: authorization proof with `biometricVerified: true` constitutes non-repudiation evidence

## Out of Scope

- Biometric hardware management (iOS system handles)
- Transaction processing (bank's core system)
- Risk assessment

## Requirements

Derived from the Verified Financial Institution Integration blueprint.

```swift
struct TransactionAuthorizationRequest {
    let requestId: UUID
    let institutionDID: String
    let transactionDescription: String
    let amountECHO: Decimal?
    let nonce: String        // Server-generated, single-use, 5-min expiry
    let expiresAt: Date
}
// Customer authorizes: biometric auth → sign(requestId + nonce + timestamp) → submit to backend
// Backend anchors authorization proof on Cardano
```

## Blueprints

- Verified Financial Institution Integration — Defines biometric authentication, DID signature combination, immutable authorization proof, anti-replay nonces, and dispute prevention

---

### WO-133: Integrate Trust Score Display with Earning Impact Projections

**Blueprint:** User Rewards Tracker on Profile

## Summary

Build the trust score integration section in the rewards tracker — showing how the user's current trust tier directly impacts their earning potential, projected monthly income calculation, and actionable steps to increase earnings through trust tier improvement.

## In Scope

- Trust score display: current tier (1–5), trust score number (0–100), trust tier label ("Member", "Verified", etc.)
- Earning multiplier display: `"Your tier 3 gives you a 1.0x rewards multiplier"` with comparison to Tier 4 (`"Upgrade to Tier 4 for 1.5x multiplier"`)
- Projected monthly income: `currentMonthlyEarnings × (targetTierMultiplier / currentTierMultiplier)` — displayed as "Estimated monthly at Tier 4: X ECHO"
- Improvement recommendations: `[{step: "Complete government ID verification", impact: "+0.5x multiplier", estimatedEarnings: "80 more ECHO/month"}]`
- Trust score timeline: sparkline showing score over last 90 days
- Real-time update: when trust tier changes, earning projections refresh within 60 seconds (via WebSocket push)
- Verification progress: if Tier 3 → show "Complete X to reach Tier 4" with progress indicator

## Out of Scope

- Trust score calculation (WO-181)
- Verification processing (WO-120)
- Achievement system (WO-116)

## Requirements

Derived from the User Rewards Tracker blueprint.

```swift
struct TrustEarningProjection {
    let currentTier: Int          // 1–5
    let currentMultiplier: Float  // 0.5, 1.0, 1.5, 2.0
    let currentMonthlyEarnings: Decimal
    let projectedAtNextTier: Decimal
    let recommendedSteps: [EarningStep]
}
```

## Blueprints

- User Rewards Tracker on Profile — Defines trust score integration with earning impact display, projected income, and optimization recommendations

---

### WO-137: Create Compliance and Legal Discovery Support System

**Blueprint:** Disappearing Messages with Cryptographic Verification

**Purpose**: Implement compliance features that support legal discovery requirements while preserving user privacy, enabling organizations to meet regulatory obligations through cryptographic evidence preservation and legal hold capabilities without compromising the disappearing message functionality.

**Requirements**:
- Implement legal hold support that preserves cryptographic evidence and metadata for messages under legal discovery while still allowing content deletion
- Provide compliance recording system that maintains audit trails of all disappearing message activities for regulatory reporting
- Enable evidence preservation mechanism that retains blockchain hashes, delivery confirmations, and participant lists for legal purposes
- Support legal discovery requests by generating comprehensive reports of communication patterns without revealing deleted content
- Ensure regulatory compliance with data retention laws while respecting user privacy through selective evidence preservation
- Maintain detailed audit trail records that include message creation, expiration, deletion, and proof generation events
- Provide legal proof generation that creates court-admissible evidence of communication occurrence and delivery
- Support compliance policy configuration that allows organizations to set retention requirements based on regulatory needs
- Enable compliance reporting that generates periodic summaries of disappearing message usage and retention activities
- Ensure privacy preservation by maintaining separation between compliance data and actual message content

**Out of Scope**:
- Legal consultation or advice features
- Integration with specific legal case management systems
- Automated legal hold triggers based on external events
- Content preservation beyond cryptographic evidence

---

### WO-138: Implement Validator Reward Distribution and Performance Management

**Blueprint:** ECHO Token Reward System and Incentive Economy

## Summary

Build the Constellation metagraph L1 validator reward distribution system — performance tracking, epoch-based reward calculation (every 24 hours), performance multipliers, and slashing for misbehavior. Validator rewards come from the 250M ECHO Validator Rewards pool, transitioning to transaction fees in later years.

## In Scope

- Validator registry: minimum 100,000 ECHO stake; track `{validatorDID, stakeAmount, publicKey, endpoint, status}`
- Performance metrics (equally weighted, 25% each): block production rate, block validity rate, uptime %, response latency P95
- Performance score: weighted average (0–100)
- Reward multipliers: score > 90 → 1.5×, score > 95 → 2.0×
- 24-hour epoch reward distribution: base reward per block × performance multiplier + transaction fee share proportional to blocks produced
- Reward schedule: Years 1–2: 25M ECHO/yr; Years 3–5: 50M ECHO/yr; Years 6–10: 25M ECHO/yr
- Slashing (Currency L1 enforced): 1% for >1h offline, 10% for >24h offline, 5% per invalid block, 50% for double signing, 25% for >7 days inactive
- Validator analytics: `GET /v1/validators/{did}/performance` — uptime, score, rewards, slashing history
- Scala/Euclid SDK: Currency L1 validation logic for slashing enforcement and reward eligibility

## Out of Scope

- User token staking (WO-127)
- Validator node networking and software

## Requirements

```go
type ValidatorPerformanceSnapshot struct {
    ValidatorDID    string
    EpochTimestamp  time.Time
    BlocksProduced  int
    BlockValidity   float64  // % valid blocks
    UptimePct       float64
    LatencyP95Ms    float64
    PerformanceScore float64  // Composite 0–100
    BaseReward      Decimal
    PerformanceMultiplier float64
    SlashAmount     Decimal  // 0 if no violation
    NetReward       Decimal
}
```

## Blueprints

- Data Layer — Defines Constellation metagraph Currency L1, validator staking, performance validation, and epoch reward distribution

---

### WO-140: Implement Trust Score Elevation and Financial Feature Access Control

**Blueprint:** In-App High-Assurance Identity Verification and Reward

## Summary

Implement the trust score elevation flow and financial feature access gate for users who complete high-assurance verification. After reaching Trust Tier 4 (high-assurance), users gain access to regulated financial features: payment rails, Verified Financial Institution messaging channels, and premium group creation. This work order handles the gating logic, not the underlying financial services.

## In Scope

- Trust tier gate middleware for financial feature endpoints: check `trustTier >= 4` via Trust Service (port 8003) before allowing access
- On verification success: trigger Trust Service to refresh tier from Cardano (invalidate Redis cache)
- iOS: unlock UI elements for Tier 4+ features (payment rail buttons, premium channel access) after tier confirmation
- Trust tier display on user profile: prominently show current tier, verification method, and expiry
- Handle tier revocation: when credential expires or is revoked, re-check tier, disable financial features, notify user via push
- Financial feature access log for compliance: `{userDID, featureAccessed, trustTierAtTime, timestamp}`

## Out of Scope

- Payment processing implementation (Verified Financial Institution work orders)
- Identity verification processing (WO-104, WO-113, WO-120)
- VC issuance (WO-132)
- Trust score calculation (WO-181)
- ECHO token rewards (WO-151)

## Requirements

Derived from the In-App High-Assurance Identity Verification blueprint.

**Feature Access by Tier:**
```
Tier 1: Basic messaging only
Tier 2: Messaging + basic rewards
Tier 3: Full rewards, group creation (≤500 members), file sharing (≤100MB)
Tier 4 (High-Assurance): All above + payment rails, financial institution channels, premium groups (≤10K members), enhanced rewards multiplier (1.5x)
Tier 5: All features, governance voting, max rewards multiplier (2.0x)
```

**Tier Gate Middleware (Go):**
```go
func RequireMinTrustTier(minTier int) Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            userDID := r.Context().Value("userDID").(string)
            tier, err := trustService.GetTier(userDID)
            if err != nil || tier < minTier {
                http.Error(w, `{"code":"INSUFFICIENT_TRUST_TIER","required":`+strconv.Itoa(minTier)+`}`, http.StatusForbidden)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}

// Apply to financial endpoints:
router.POST("/v1/payments/initiate", RequireMinTrustTier(4), paymentsHandler)
router.POST("/v1/financial-institution/channels", RequireMinTrustTier(4), institutionHandler)
```

## Blueprints

- In-App High-Assurance Identity Verification and Reward — Defines trust score elevation, financial feature access, Tier 4 feature unlock, and revocation handling
- Decentralized Identity and Authentication — Specifies trust tier feature access table

---

### WO-141: Implement Permission Structures and Access Control System

**Blueprint:** Public and Private Groups with Verified Status Display

## Summary

Build the extended permission management UI — allowing group owners to customize the default permission matrix per role, configure trust tier thresholds for specific actions, set up temporary permission delegations, and view a complete permission audit log. This is the admin interface layer on top of the permission enforcement system from WO-123.

## In Scope

- Permission configuration UI: toggle matrix per role showing each permission and its current state
- Verification level threshold configuration: set minimum trust tier for each capability (e.g., "File sharing requires Tier 2+")
- Permission delegation: `{targetDID, permission, expiresAt}` temporary grant; auto-revoked when expired
- Permission delegation management: list active delegations, revoke before expiry
- Permission audit log view: sortable/filterable list of permission changes with actor, target, change, timestamp
- Immediate effect: permission changes applied to backend within 1 second of save
- Permission inheritance visualization: show inheritance chain (Owner includes all Admin permissions, etc.)
- Bulk permission update: apply same permission change to multiple members

## Out of Scope

- Permission enforcement (WO-123 handles the actual enforcement)
- Governance voting (WO-148)
- Moderation tools (WO-112, WO-131)

## Requirements

Derived from the Public and Private Groups blueprint.

**Permission Delegation:**
```swift
struct PermissionDelegation: Identifiable, Codable {
    let id: UUID
    let grantedToDID: String
    let permission: GroupPermission    // .postMedia, .inviteMembers, etc.
    let grantedByDID: String           // Admin who granted
    let expiresAt: Date
    var isActive: Bool { Date() < expiresAt }
}
// Stored in PostgreSQL, auto-expired on each permission check
// iOS: refresh from backend every 5 minutes; real-time via WebSocket on change
```

## Blueprints

- Public and Private Groups with Verified Status Display — Defines custom permission configuration, role-based permissions, verification level thresholds, permission delegation, and permission auditing

---

### WO-142: Build Earning Goals System with Progress Tracking

**Blueprint:** User Rewards Tracker on Profile

## Summary

Build the earning goals system — users set ECHO token earning targets (daily/weekly/monthly) and track progress toward them. Color-coded status (on-track/behind/at-risk) with personalized recommendations to help achieve goals. Note: WO-160 covers the same feature; this work order (WO-142) is the earlier implementation pass.

## In Scope

- Goal creation: select period (daily/weekly/monthly), target ECHO amount, optional deadline
- Progress display: progress bar, `{earned}/{target} ECHO`, percentage complete, time remaining
- Color-coded status: green (on-track if current velocity × remaining time ≥ remaining target), yellow (slightly behind), red (unlikely to achieve at current pace)
- Personalized recommendations: `"Send 20 more messages today to hit your daily goal"`, `"Refer 1 verified friend to reach your monthly goal"`
- Goal achievement: celebration animation + unlock notification + option to share achievement card
- Goal modification: adjust target or deadline; show impact on achievability
- Goal history: past goals with completed/missed status, actual vs. target, duration

## Out of Scope

- Automated goal suggestions (future enhancement)
- Social challenges with other users
- Integration with external task management tools

## Requirements

Derived from the User Rewards Tracker blueprint.

```swift
struct EarningGoal: Identifiable, Codable {
    let id: UUID
    var period: GoalPeriod       // .daily, .weekly, .monthly
    var targetECHO: Decimal
    var deadline: Date
    var currentEarnings: Decimal
    var status: GoalStatus       // .onTrack, .behind, .atRisk, .completed, .missed

    var progressPct: Double { Double(currentEarnings / targetECHO) }
}
```

## Blueprints

- User Rewards Tracker on Profile — Defines earning goals, progress tracking, color-coded status, personalized recommendations, and goal history

---

### WO-145: Create Premium Badge System and Verification Status Display

**Blueprint:** In-App High-Assurance Identity Verification and Reward

## Summary

Build the "Identity Verified" premium badge system and verification status display. Users who complete high-assurance verification (Tier 4) receive a distinctive verification badge displayed on their profile, in conversations, and in groups. The badge is backed by the user's trust tier from Cardano and is automatically removed when credentials expire or the tier drops.

## In Scope

- `VerificationBadge` SwiftUI component for display across all surfaces (profile, chat header, group member list)
- Badge levels: standard verification badge (Tier 3), premium "Identity Verified" badge (Tier 4+)
- Badge tap → detail sheet showing verification method used, verification date, credential expiration date
- No personal data exposed in badge detail view (only: verification method, date, expiry — not document type or name)
- Automatic badge removal: when trust tier drops below required level (credential expiry or revocation)
- Backend: trust tier ≥ 4 in cached trust score determines badge eligibility
- Badge state synchronization: refresh when trust tier WebSocket push is received
- `VerificationBadge` and `TrustBadge` components in `Presentation/Components/Indicators/`

## Out of Scope

- Verification processing (WO-104, WO-113, WO-120)
- Trust score calculation (WO-181)
- Financial feature access (WO-140)
- ECHO token rewards (WO-151)

## Requirements

Derived from the In-App High-Assurance Identity Verification blueprint.

**Badge Component:**
```swift
// Presentation/Components/Indicators/VerificationBadge.swift
struct VerificationBadge: View {
    let verificationStatus: VerificationStatus
    let onTap: () -> Void

    var body: some View {
        Button(action: onTap) {
            HStack(spacing: 2) {
                Image(systemName: verificationStatus.iconName)
                    .foregroundColor(verificationStatus.color)
                    .font(.system(size: 12, weight: .semibold))
            }
        }
    }
}

enum VerificationStatus {
    case unverified          // No badge
    case basicVerified       // Gray shield: Tier 2
    case memberVerified      // Bronze shield: Tier 3
    case identityVerified    // Blue checkmark: Tier 4 (high-assurance)
    case trustedVerified     // Gold shield: Tier 5

    var iconName: String {
        switch self {
        case .identityVerified: return "checkmark.seal.fill"
        case .trustedVerified: return "star.seal.fill"
        default: return "shield.fill"
        }
    }
}
```

**Badge Detail Sheet:**
```swift
// Shows when user taps verification badge
struct VerificationDetailSheet: View {
    let user: User
    var body: some View {
        VStack {
            Image(systemName: "checkmark.seal.fill").foregroundColor(.blue)
            Text("Identity Verified")
            Text("Verified via: \(user.verificationMethod)")
            Text("Verified: \(user.verifiedAt?.formatted() ?? "—")")
            Text("Valid until: \(user.credentialExpiresAt?.formatted() ?? "—")")
            // NO: user's name, document type, DOB, address
        }
    }
}
```

## Blueprints

- In-App High-Assurance Identity Verification and Reward — Defines premium badge display, verification status management, badge revocation on credential expiry, and verification detail visibility

---

### WO-146: Build Anti-Gaming Detection and Prevention System

**Blueprint:** ECHO Token Reward System and Incentive Economy

**Purpose**: This work order implements comprehensive anti-gaming measures that prevent Sybil attacks, spam farming, and reward manipulation while maintaining legitimate user experience. Completing this protects the token economy from abuse and ensures rewards are distributed fairly to authentic network participants.

**Requirements**:
- Create Data L1 validation layer that tracks user activity patterns including messages per hour/day, unique recipients, account age, device fingerprints, and IP addresses to calculate Sybil risk scores from 0-100
- Implement spam detection system that monitors rapid message sending, repetitive content patterns, and messages to many unique recipients, flagging accounts with spam scores >70 for reduced reward rates
- Build Sybil attack prevention that identifies linked accounts through device fingerprints and IP addresses, applies collective earning caps across linked accounts, and requires additional verification for high-risk accounts (score >80)
- Create trust score requirement enforcement where payment rewards require minimum trust score of 50, referral bonuses require both users above trust score 30, with automatic validation before reward distribution
- Implement progressive penalty system that reduces reward multipliers for accounts flagged with suspicious patterns: 50% reduction for spam flags, 75% reduction for Sybil flags, temporary suspension for multiple violations
- Build anti-gaming analytics dashboard that displays flagged accounts, risk score distributions, penalty applications, and gaming attempt statistics with real-time monitoring and alert capabilities

**Out of Scope**:
- Trust score calculation algorithms (integrates with existing trust system)
- Device fingerprinting and IP tracking infrastructure
- Content analysis for spam detection beyond pattern matching
- Manual review processes for flagged accounts

---

### WO-147: Implement Anonymized Leaderboards with Zero-Knowledge Ranking

**Blueprint:** User Rewards Tracker on Profile

## Summary

Build the anonymized leaderboards with ZK-proof-based ranking verification. Regional and trust-tier leaderboards show relative earnings positions without revealing actual amounts. Users can prove their rank position using ZK proofs. Note: WO-166 covers the same feature; this work order implements the earlier version.

## In Scope

- Regional leaderboard: top 100 earners in user's geographic region, anonymized display names (`Earner#1247`)
- Trust tier leaderboard: separate rankings per trust tier (1–5), so users compete within their tier
- ZK ranking proof: user can generate proof of their rank position without revealing earnings amount (Phase 3+ via Midnight SDK)
- Ranking updates: recalculate every 5 minutes; publish snapshot to backend
- Time period filters: daily, weekly, monthly, all-time
- Privacy: actual earnings amounts never displayed, only relative ranking percentile (`"Top 8%"`)
- Ranking audit log: all ranking calculations logged for transparency (no user data, just aggregate)
- Tied rankings: shared positions with clear tie-breaking rules (earlier achievement date)

## Out of Scope

- Individual earning amount disclosure (privacy requirement)
- Direct user-to-user comparison outside leaderboards
- Leaderboard prizes

## Requirements

Derived from the User Rewards Tracker blueprint.

```swift
struct LeaderboardEntry {
    let rank: Int
    let anonymizedName: String   // "Earner#1247" — no DID exposure
    let trustTier: Int
    let rankPercentile: Double   // 0.08 = top 8%
    let isCurrentUser: Bool      // Highlight user's own row
    // NEVER: actual ECHO amount, DID, display name
}
```

## Blueprints

- User Rewards Tracker on Profile — Defines anonymized regional/tier leaderboards, ZK ranking proofs, real-time updates, and privacy protection

---

### WO-151: Build ECHO Token Reward Distribution System for Identity Verification

**Blueprint:** In-App High-Assurance Identity Verification and Reward

## Summary

Build the ECHO token reward distribution for the In-App High-Assurance Identity Verification feature — crediting exactly 100 ECHO tokens to a user's account immediately upon successful high-assurance verification. Duplicate prevention ensures each user receives the reward only once per verification type. Integrates with the Rewards Service (port 8004) and Currency L1 via `AtomicAction`.

## In Scope

- Reward trigger: called by Identity Service after successful IDV callback and VC issuance
- Idempotency: check PostgreSQL reward history before submitting — prevent duplicate rewards per `(DID, verificationMethod)`
- 100 ECHO reward submitted via `AtomicAction` bundle to Currency L1: `[verifyTrustTier(did), claimVerificationReward(did, 100, "high_assurance"), updateDailyCap(did)]`
- Reward processing target: within 1 block cycle (~5 seconds) of verification completion
- WebSocket push confirmation to user: `{type: "reward_credited", amount: 100, txHash, verificationMethod}`
- In-app reward notification: "🎉 You've earned 100 ECHO for completing identity verification!"
- Reward transaction history record in PostgreSQL for audit and user viewing in Wallet tab

## Out of Scope

- Verification processing (WO-104, WO-113, WO-120)
- Currency L1 implementation (WO-8)
- Wallet balance display (iOS Wallet work orders)
- Ongoing messaging/referral rewards (separate Rewards work orders)

## Requirements

Derived from the In-App High-Assurance Identity Verification blueprint.

**Reward Transaction:**
```go
// Rewards Service POST /v1/rewards/identity-verification (internal)
type IdentityVerificationRewardRequest struct {
    UserDID           string `json:"user_did"`
    VerificationMethod string `json:"verification_method"` // "high_assurance", "kyc_lite"
    IDVReferenceUUID  string `json:"idv_reference_uuid"`   // For idempotency key
}

func (s *RewardsService) DistributeIdentityReward(req IdentityVerificationRewardRequest) error {
    // 1. Idempotency check
    if s.db.HasReward(req.UserDID, req.VerificationMethod) { return ErrAlreadyRewarded }

    // 2. AtomicAction: verify tier + claim + update cap
    rewardAmount := int64(100_000_000_000_000_000) // 100 ECHO in smallest units
    txHash, err := s.metagraph.SubmitAtomicAction([]AtomicAction{
        {Type: "verify_trust_tier", DID: req.UserDID},
        {Type: "claim_verification_reward", DID: req.UserDID, Amount: rewardAmount},
        {Type: "update_daily_cap", DID: req.UserDID},
    })
    if err != nil { return s.queueRetry(req) }

    // 3. Record and notify
    s.db.RecordReward(req.UserDID, req.VerificationMethod, txHash)
    s.websocket.Push(req.UserDID, RewardNotification{Amount: 100, TxHash: txHash})
    return nil
}
```

## Blueprints

- In-App High-Assurance Identity Verification and Reward — Defines 100 ECHO reward amount, automatic distribution, reward timing, and notification requirement
- Data Layer — Specifies `AtomicAction` primitive for atomic multi-step Currency L1 operations

---

### WO-152: Implement Regulatory Compliance Verification System

**Blueprint:** Verified Financial Institution Integration

## Summary

Build the regulatory compliance verification system for financial institutions — validating FDIC registration, AML compliance, KYC documentation, and banking license status during institutional DID registration and through continuous monitoring. Compliance status is recorded on Cardano as part of the institutional DID document.

## In Scope

- FDIC compliance verification: query FDIC public API for institution registration status and validity
- AML compliance verification: verify institution has current AML program certification documents (hash stored in DID document)
- KYC compliance: verify institution has current KYC procedures documentation on file
- Banking license validation: verify license number against state/federal regulatory databases
- Continuous monitoring: daily automated status checks via regulatory APIs; flag changes immediately
- Compliance status UTXO on Cardano: `{institutionDID, complianceStatus: active|suspended|revoked, lastVerified, licenseExpiry}`
- Compliance change alerts: webhook notification to platform admins when status changes
- Compliance audit trail: all verification events timestamped in Cardano transaction metadata
- Compliance report generation: `GET /v1/financial/institutions/{did}/compliance-report` — PDF report for regulatory examinations

## Out of Scope

- Legal interpretation of regulatory requirements
- Customer-facing compliance UI

## Requirements

Derived from the Verified Financial Institution Integration blueprint.

```go
type InstitutionalComplianceStatus struct {
    InstitutionDID     string
    FDICStatus         string    // "active", "suspended", "not_registered"
    AMLCertified       bool
    KYCCompliant       bool
    LicenseStatus      string
    LastVerifiedAt     time.Time
    NextCheckAt        time.Time
    ComplianceScore    int       // 0–100 composite
}
```

## Blueprints

- Verified Financial Institution Integration — Defines FDIC compliance verification, AML/KYC requirements, regulatory database integration, continuous monitoring, and compliance reporting

---

### WO-153: Implement Validator Reward Distribution and Performance Management System

**Blueprint:** ECHO Token Reward System and Incentive Economy

## Summary

Implement the enhanced validator reward distribution and performance management system (Phase 6 version) — same validator reward mechanics as WO-138 but extended with the full validator registry management API, stake update capabilities, and a performance dashboard for network health monitoring.

## In Scope

- All capabilities from WO-138 (validator registry, performance tracking, reward distribution, slashing)
- Validator registry management API: `POST /v1/validators/register`, `PATCH /v1/validators/{did}/stake`, `GET /v1/validators/{did}`, `DELETE /v1/validators/{did}`
- Performance history API: `GET /v1/validators/{did}/performance?from={date}&to={date}` — paginated epoch history
- Reward claim API: `POST /v1/validators/{did}/rewards/claim` — triggers AtomicAction on Currency L1
- Network health dashboard: total active validators, avg performance score, total staked, daily rewards distributed, slashing events
- Stake update: validators can increase/decrease stake (minimum 100K ECHO) with lock period enforcement

## Out of Scope

- Validator node software
- Consensus algorithm implementation
- User staking (WO-127)

## Requirements

Derived from the ECHO Token Reward System blueprint.

**Note:** WO-138 and WO-153 cover the same domain. WO-138 is Phase 6 first pass; this WO-153 adds the management APIs and monitoring dashboard layer on top.

```go
// Validator registry entry
type ValidatorEntry struct {
    ValidatorDID    string
    StakeAmount     Decimal
    RegistrationDate time.Time
    Status          ValidatorStatus  // active, suspended, removed
    TotalBlocksProduced int
    CurrentPerformanceScore float64
}
```

## Blueprints

- Data Layer — Specifies Constellation metagraph Currency L1, validator staking, performance validation, and epoch reward distribution mechanics

---

### WO-154: Create Wallet Integration with DeFi Protocol Connections

**Blueprint:** User Rewards Tracker on Profile

## Summary

Build the wallet integration section showing external DeFi opportunities for ECHO tokens — PacaSwap DEX swaps, cross-chain bridge to Base/Ink, and validator delegation APY information. Powered by Stargazer SDK within the iOS Wallet tab.

## In Scope

- PacaSwap swap interface: ECHO/DAG and ECHO/USDC pairs, live quote (exchange rate, price impact, min received, 0.3% fee), biometric confirmation → AtomicAction swap via Stargazer SDK
- Bridge interface: ECHO → Base (Aerodrome DeFi access), ECHO → Ink (Kraken exchange access); show fee, estimated time, destination address
- Validator delegation browser: list of L1 validators with uptime %, commission %, delegated stake, APR estimate; delegate via `StakeDelegation` → Stargazer SDK
- USD value of balance: derived from PacaSwap TWAP oracle (10-min window × DAG/USD price)
- Founder vesting section (founders only): vested/locked/withdrawable breakdown, next unlock date, `WithdrawLock` action

## Out of Scope

- MetaMask/WalletConnect integration (ECHO is native Constellation, not EVM)
- Custom DeFi protocol development
- Fiat on-ramp

## Requirements

Derived from the User Rewards Tracker and Frontend blueprints.

**Note:** ECHO is a native Constellation Network Metagraph L1 token — not EVM-compatible. External wallet connectivity uses Stargazer SDK (Constellation-native) and the Constellation bridges to Base/Ink for cross-chain access.

```swift
// From Frontend blueprint:
// Swap: stargazer.submitAtomicAction([.swapExactInput(...)])
// Bridge: bridge transaction via Stargazer SDK
// Delegation: stargazer.submitStakeDelegation(...)
```

## Blueprints

- User Rewards Tracker on Profile — Defines wallet integration with DeFi protocol connections, staking, liquidity provision, and conversion rates
- Frontend — Defines complete WalletTab with swap, bridge, delegation, and founder vesting flows

---

### WO-155: Implement Token Burning and Deflationary Mechanisms

**Blueprint:** ECHO Token Reward System and Incentive Economy

## Summary

Implement the ECHO token burning and deflationary mechanism — automatically burning 10% of collected transaction fees from message anchoring, payments, file storage, and smart contract execution. Burn events are recorded on Currency L1 as immutable supply reduction events. Governance can adjust the burn rate via on-chain vote.

## In Scope

- Automatic burn trigger: 10% of all platform fee transactions routed to burn address (unspendable wallet) via `FeeTransaction`
- Fee sources that trigger burning: message anchoring fees, payment transaction fees (0.1% of value), file storage fees (0.01 ECHO/GB/month), smart contract execution fees
- Burn transaction: `FeeTransaction` to null/burn address on Currency L1 — permanently removes from circulating supply
- Total supply reduction tracking: metagraph state tracks `totalBurned` counter, updated on each burn event
- Burn analytics: `GET /v1/tokens/burn-stats` — total burned, burn rate over time, effective inflation rate, projected supply
- Burn event API: `GET /v1/tokens/burn-events` — paginated list of burn transactions with `{amount, source, txHash, timestamp}`
- Governance-controlled burn rate: on-chain proposal to increase burn % above 10%; requires 20% quorum, 50% approval

## Out of Scope

- Fee collection infrastructure (handled by payment and messaging systems)
- Governance voting implementation (WO-177)

## Requirements

Derived from the ECHO Token Reward System blueprint.

```go
// Burn process:
// 1. Fee collected: 1 ECHO messaging anchoring fee
// 2. FeeTransaction: 0.1 ECHO → burn address (permanently removed)
//                   0.9 ECHO → validator rewards pool + platform treasury
// Currency L1 validators verify: correct 10% burn ratio, valid burn address, authorized sender
type BurnEvent struct {
    TxHash      string
    Amount      Decimal    // ECHO burned
    Source      string     // "messaging_fee", "payment_fee", "storage_fee"
    Timestamp   time.Time
    CumulativeBurned Decimal
}
```

## Blueprints

- Data Layer — Defines `FeeTransaction` v3 primitive for automated fee payments and burn mechanics

---

### WO-157: Build Secure Document Exchange System

**Blueprint:** Verified Financial Institution Integration

## Summary

Build the secure financial document exchange system — E2E encrypted document transmission between banks and customers, institutional P-256 digital signatures on all documents, document hash anchoring on Data L1 for tamper-proof verification, and configurable retention policies meeting regulatory requirements.

## In Scope

- Document E2E encryption: encrypt with customer's Kinnami key before transmission (same as message encryption)
- Institutional document signing: sign `SHA256(documentContent)` with institutional DID P-256 key
- Customer signature verification: verify institutional signature on receipt; display `✓ Authentic from [Bank]`
- Document hash anchoring: `{type: "document_integrity", documentHash, institutionDID, customerDID_hash, timestamp}` on Data L1
- Document retention policies: configurable by type (account statements: 7yr, loan docs: 10yr, correspondence: 3yr) — IPFS pinning TTL enforced
- Secure document deletion: destroy encryption key + IPFS unpin + anchor deletion event on Data L1
- Audit trail: document lifecycle events logged `{uploaded, accessed, signed, deleted, timestamp, actorDID}`
- Document access log: institution and customer can request full access history per document

## Out of Scope

- Document content analysis or OCR
- Document format conversion
- DMS integration

## Requirements

```go
type FinancialDocument struct {
    DocumentID      string
    DocumentType    string   // "account_statement", "loan_doc", "compliance_notice"
    InstitutionDID  string
    CustomerDID     string
    EncryptedContent []byte  // Kinnami encrypted
    Signature       []byte   // Institutional P-256 signature
    ContentHash     []byte   // SHA-256 for verification
    RetentionYears  int
    ExpiresAt       time.Time
}
```

## Blueprints

- Verified Financial Institution Integration — Defines secure document exchange with E2E encryption, digital signing, blockchain anchoring, compliance recording, and retention policies

---

### WO-160: Build Earning Goals System with Progress Tracking

**Blueprint:** User Rewards Tracker on Profile

## Summary

Build the enhanced earning goals system (Phase 6 implementation) — extending the foundation from WO-142 with broader goal timeframes (quarterly, yearly), feasibility scoring based on historical patterns, and more sophisticated goal recommendations. Users set ECHO earning targets and receive data-driven guidance.

## In Scope

- Goal creation with extended timeframes: daily, weekly, monthly, quarterly, yearly
- Target range validation: minimum 100 ECHO, maximum 1,000,000 ECHO
- Feasibility score (0–100%): computed weekly based on `historicalDailyAvg × remainingDays vs. remainingTarget`
- Personalized recommendation engine: `{action, estimatedImpact}` — "Complete Tier 3 verification (+0.5x multiplier, +45 ECHO/month)", "Refer 3 more friends (+150 ECHO in referral bonuses)"
- Progress bar with velocity indicator: "At current pace, you'll reach your goal in X days"
- Goal achievement: celebrate with `ConfettiView`, suggest next goal, record in achievement history
- Goal modification: update target or deadline; recalculate feasibility and velocity
- Goal archive: list of all past goals with `{period, targetECHO, achievedECHO, status, completionDate?}`

## Out of Scope

- Automated goal creation (future ML feature)
- Social challenges
- Leaderboard integration

## Requirements

Derived from the User Rewards Tracker blueprint.

```swift
struct EarningGoal: Identifiable, Codable {
    let id: UUID
    var period: GoalPeriod        // .daily, .weekly, .monthly, .quarterly, .yearly
    var targetECHO: Decimal       // 100 to 1_000_000
    var deadline: Date
    var currentEarnings: Decimal
    var feasibilityScore: Double  // 0.0 to 1.0, updated weekly
    var status: GoalStatus
}
```

## Blueprints

- User Rewards Tracker on Profile — Defines earning goals with target amounts, timeframes, progress tracking, recommendations, goal achievement notifications, and history

---

### WO-161: Build Advanced Anti-Gaming and Security Monitoring System

**Blueprint:** ECHO Token Reward System and Incentive Economy

## Summary

Build the advanced anti-gaming and security monitoring system for the reward economy — Sybil risk scoring, spam pattern detection, progressive reward decay enforcement, and trust score gating for reward eligibility. Protects the 400M ECHO user rewards pool from manipulation.

## In Scope

- Sybil risk scoring: detect accounts from same device fingerprint, IP cluster, or behavioral signature; apply reduced reward rates (50%) for high-risk account groups
- Spam detection: flag accounts with rapid-send patterns (>10 messages/minute), messages to >100 unique recipients/day, or highly repetitive content → apply progressive earning suspension (1h, 24h, 7d)
- Progressive reward decay: Currency L1 validation rule — after 1000 messages/day, each additional 100 messages reduces multiplier by 10%; daily reset at midnight UTC
- Trust score gating: minimum Tier 2 for messaging rewards, Tier 3 for payment rewards, Tier 3 for referral bonuses (enforced in Currency L1 Scala validation)
- Linked account detection: multiple DIDs with same device fingerprint → collective daily cap across all linked accounts
- Security monitoring dashboard: real-time suspicious pattern alerts, anti-gaming rule trigger events, violation audit log
- Anti-gaming metrics: all stored in metagraph Data L1 state (per-DID velocity counters, linked account graphs)

## Out of Scope

- Trust score calculation (WO-181)
- Account suspension/banning (separate moderation)
- Identity verification (WO-120)

## Requirements

```go
type AntiGamingCheck struct {
    UserDID        string
    SybilRiskScore float64    // 0.0 to 1.0 (1.0 = highest risk)
    MessageVelocity int       // Messages in last minute
    DailyCount     int        // Messages today
    IsSpamFlagged  bool
    EffectiveMultiplier float64  // After all deductions
}
```

## Blueprints

- Data Layer — Defines anti-gaming Currency L1 validation rules, daily cap enforcement, velocity checks, and trust tier gating for reward eligibility

---

### WO-163: Develop Channel Monetization and Payment Processing System

**Blueprint:** Broadcast Channels and Community Features

## Summary

Build the channel monetization system — ECHO token subscription tiers (free, basic, premium), premium content access gates, and donation functionality. All payments are ECHO token transactions on the Constellation Currency L1 using `AllowSpend` and `SpendTransaction` v3 primitives.

## In Scope

- Subscription tier configuration by channel creator: `{tierName, echoPrice, benefits[]}`
- Free tier: all public content; Basic tier (paid): + some premium posts; Premium tier (paid): all content
- Premium content posts: creator marks individual posts as "Basic+" or "Premium+" — access gate enforced by backend
- Subscriber upgrade/downgrade/cancel: managed via `AllowSpend` authorization for recurring payments (Phase 5 `AllowSpend` primitive)
- One-time donations: `SpendTransaction` from subscriber to creator wallet with optional message
- Revenue tracking: `GET /v1/channels/{id}/revenue` — earnings by source (subscriptions, donations), payout schedule
- Sponsored content disclosure: `{isSponsored: true, sponsor: "Brand Name"}` field on posts, required by creator
- Payment escrow: subscription fees collected into escrow at month start; released to creator at month end minus platform fee

## Out of Scope

- Fiat currency payments
- Tax calculation
- Affiliate marketing

## Requirements

Derived from the Broadcast Channels blueprint.

**Payment Primitives:**
```
Subscription: AllowSpend (Phase 5) → monthly SpendTransaction deduction
Donation: immediate SpendTransaction from subscriber to creator
Platform fee: 20% retained via FeeTransaction
All transactions: Constellation Currency L1, ECHO token
```

## Blueprints

- Broadcast Channels and Community Features — Defines subscription fee configuration, premium content tiers, donation functionality, revenue tracking, and sponsored content disclosure

---

### WO-166: Implement Anonymized Leaderboards with Zero-Knowledge Ranking

**Blueprint:** User Rewards Tracker on Profile

## Summary

Build the Phase 7 anonymized leaderboards with ZK-proof-based ranking verification — regional and trust tier rankings with privacy-preserving earning comparison. Users see their percentile ranking and can prove their position using ZK proofs. Extension of WO-147 with enhanced ZK proof capabilities.

## In Scope

- Same core leaderboard features as WO-147 (regional + tier leaderboards, anonymized names, time period filters)
- Enhanced ZK proof (Midnight SDK, Phase 3+): user generates proof of rank position `"I am in top 10%"` without revealing earnings amount or identity
- ZK proof sharing: user can share the proof with contacts as a verifiable claim (`{rank_percentile, period, zkProof}`)
- Enhanced ranking analytics: distance to next tier, ranking momentum (position change over last 7 days), estimated time to advance
- `"Top X%"` percentile display instead of exact rank number (more privacy-preserving)
- Smooth position transition animations (Core Animation)

## Out of Scope

- Actual earning amounts (privacy)
- Profile linking from leaderboard
- Prizes for top positions

## Requirements

Derived from the User Rewards Tracker blueprint.

```swift
struct AnonymizedLeaderboardEntry {
    let rank: Int
    let anonymizedId: String    // "Earner#5829"
    let tier: Int
    let percentile: Double      // 0.08 = top 8%
    let isCurrentUser: Bool
    let positionDelta: Int?     // +3 or -1 vs. last period
    // NEVER: DID, actual earnings, real name
}
```

## Blueprints

- User Rewards Tracker on Profile — Defines anonymized regional/tier leaderboards, ZK ranking proofs, real-time updates, ranking privacy, and percentile display

---

### WO-167: Implement ECHO Token Reward Distribution for Identity Verification

**Blueprint:** In-App High-Assurance Identity Verification and Reward

**Purpose**: Automatically distribute ECHO token rewards to users who successfully complete identity verification, providing direct incentive for strengthening the network's trust layer and encouraging user participation in the verification process.

**Requirements**:
- System must automatically distribute 100 ECHO tokens to user accounts immediately upon successful identity verification completion
- Reward distribution must be triggered only once per user, preventing duplicate rewards for re-verification or credential renewal
- Token distribution must be recorded with timestamp, verification method used, and transaction ID for auditing purposes
- System must handle reward distribution failures gracefully with retry mechanism and error logging
- Reward history must be viewable by users in their account showing verification date and ECHO amount received
- Reward verification must confirm successful token transfer before marking verification process as complete
- System must track total ECHO rewards distributed for identity verification across all users for reporting
- Reward compliance must ensure token distribution follows applicable regulations and platform token policies

**Out of Scope**:
- Identity verification processing and validation logic
- Trust score calculation and elevation
- Premium badge issuance and display
- ECHO token wallet management and general token functionality

---

### WO-170: Implement Token Burning and Deflationary Mechanism System

**Blueprint:** ECHO Token Reward System and Incentive Economy

## Summary

Implement the extended token burning and deflationary mechanism system (Phase 6) — building on WO-155's foundation with governance integration for burn rate voting (adjustable from 5% to 20%), enhanced burn rate monitoring, and public burn verification API for third-party auditing.

## In Scope

- All capabilities from WO-155 (10% automatic burn on fees via `FeeTransaction`)
- Extended fee burn sources: premium subscription payments (10% burned)
- Governance-controlled burn rate: on-chain vote to adjust burn percentage within 5%–20% range; requires 20% quorum, 50% approval; new rate takes effect on next governance epoch
- Burn rate anomaly alerting: if actual burn rate deviates >1% from configured rate, alert platform operators
- Third-party verification API: `GET /v1/tokens/burn-events/{txHash}/proof` → returns burn proof with metagraph snapshot reference for independent auditing
- Effective inflation rate calculation: `(rewardsDistributed - burned) / totalSupply × 100` — published daily
- Burn forecast: project next 12-month burn at current rates

## Out of Scope

- WO-155 scope (first implementation pass)
- Governance voting infrastructure (WO-177)

## Requirements

Extended from WO-155 with governance burn rate control.

```go
type BurnRateConfig struct {
    CurrentBurnPct     float64   // Governance-set, 5-20%
    LastUpdated        time.Time
    GovernanceProposalId string  // Last governance proposal that changed this
}
```

## Blueprints

- Data Layer — Defines `FeeTransaction` primitive and token economics

---

### WO-171: Integrate External Wallet Connections with DeFi Protocol Support

**Blueprint:** User Rewards Tracker on Profile

## Summary

Integrate ECHO token with external DeFi protocols accessible via the Stargazer SDK's cross-chain bridges. This work order handles the DeFi opportunity display and interaction flows for Phase 3+ features: PacaSwap liquidity provision (LP tokens), cross-chain bridge to Base/Ink for broader DeFi access, and external wallet compatibility (Stargazer + D'Cent hardware wallet).

## In Scope

- Liquidity provider interface: add/remove liquidity to ECHO/DAG and ECHO/USDC pools; view LP token balance; stake LP tokens for liquidity mining rewards
- LP position display: current pool share %, impermanent loss estimate, accrued fees
- Bridge to Base: ECHO → Base for Aerodrome DeFi (lending, yield, treasury BTC accumulation)
- Bridge to Ink: ECHO → Ink for Kraken exchange listing access
- D'Cent hardware wallet display: link D'Cent hardware wallet DID to show ECHO balance and initiate transactions
- USD value from PacaSwap 10-minute TWAP oracle × DAG/USD price feed

## Out of Scope

- MetaMask/WalletConnect (EVM wallets) — ECHO is Constellation-native
- Fiat on-ramp
- Custom DeFi protocol development

## Requirements

Derived from the User Rewards Tracker and Frontend blueprints.

```swift
// LP position display
struct LiquidityPosition {
    let poolPair: String          // "ECHO/DAG" or "ECHO/USDC"
    let lpTokenBalance: Decimal
    let poolSharePercent: Double
    let accruedFees: Decimal
    let impermanentLossEstimate: Double  // Percentage
}
// Stargazer SDK: stargazer.addLiquidity(pool: .echoDag, amount: inputECHO)
```

## Blueprints

- User Rewards Tracker on Profile — Defines external wallet integration, DeFi protocol connections, staking, liquidity provision
- Frontend — Defines PacaSwap DEX integration, cross-chain bridges, LP interface, and D'Cent hardware wallet support

---

### WO-173: Build Reward Analytics Dashboard with Distribution Insights

**Blueprint:** User Rewards Tracker on Profile

## Summary

Build the internal reward analytics dashboard for platform administrators — tracking reward distribution efficiency, user engagement correlation, tokenomics health metrics, and optimization insights. All analytics are aggregate/anonymized; no individual user data exposed. Used for platform operations and tokenomics tuning.

## In Scope

- Total rewards distributed per day/week/month (aggregate, by type)
- Reward pool utilization rate: `claimedRewards / maxPossibleRewards` per period
- User engagement correlation: DAU × average rewards earned per user
- Peak earning patterns: hourly/daily earning velocity heatmap
- Trust tier distribution of active earners: breakdown by tier showing reward claim rates
- Tokenomics health metrics: reward inflation rate, token velocity, daily burn rate (if applicable)
- Optimization recommendations: flag if messaging reward cap is too high/low based on actual claim rates
- Data export: CSV/JSON with aggregate metrics, no user-level data
- Analytics retention: 2 years in PostgreSQL analytics tables

## Out of Scope

- Individual user earnings tracking
- User-facing analytics (WO-106, WO-158 are user-facing)
- Predictive ML modeling

## Requirements

Derived from the User Rewards Tracker blueprint.

```go
type RewardDistributionReport struct {
    Period      DateRange
    TotalDistributed Decimal
    ByType      map[string]Decimal  // "messaging", "referral", "staking"
    UtilizationRate float64          // 0.0 to 1.0
    ActiveEarners int               // Unique earners in period
    AvgEarningsPerUser Decimal
}
```

## Blueprints

- User Rewards Tracker on Profile — Defines reward distribution analytics, engagement correlation, tokenomics optimization, export capabilities, and privacy-preserving aggregation

---

### WO-174: Develop Cross-Chain Bridge for Constellation-Cardano Token Interoperability

**Blueprint:** ECHO Token Reward System and Incentive Economy

## Summary

Build the cross-chain bridge between Constellation Metagraph (ECHO native) and Cardano (ECHO wrapped) to enable ECHO to be used across both ecosystems. Lock ECHO on Constellation → mint equivalent wrapped ECHO on Cardano; reverse to bridge back. Fixed total supply maintained across both chains.

## In Scope

- Constellation-side bridge contract: `CrossChainBridge` on Currency L1 — locks ECHO tokens when bridge-out initiated, releases when bridge-in confirmed
- Cardano-side integration: mint Cardano native ECHO asset (via Plutus script) when Constellation lock confirmed; burn on bridge-back
- 1:1 peg enforcement: Cardano minted amount = Constellation locked amount; total supply remains 1B at all times
- Cryptographic proof verification: Cardano bridge validates Constellation lock proof (metagraph snapshot signature) before minting
- Bridge fee: 0.1% of bridged amount (`FeeTransaction` to bridge operator + security fund)
- Bridge status tracking: `GET /v1/bridge/transactions/{id}` — shows `locking → locked → minting → complete` or reverse
- Emergency pause: bridge operators can pause for security incidents
- iOS bridge UI: from WalletTab → [Bridge] → select Base or Ink (Constellation bridges, easier path) vs. Cardano (this work order)

## Out of Scope

- Cardano smart contract development (requires Plutus developers)
- DeFi protocol integration on Cardano
- Bridge insurance/loss protection

## Requirements

```go
type BridgeTransaction struct {
    TxID          string
    SourceChain   string  // "constellation" or "cardano"
    TargetChain   string
    Amount        Decimal
    Status        string  // "locking", "locked", "minting", "complete"
    Proof         []byte  // Metagraph snapshot sig (constellation→cardano)
    Fee           Decimal // 0.1% of amount
}
```

## Blueprints

- Data Layer — Defines cross-chain bridge concepts and token interoperability with Cardano

---

### WO-176: Implement Blockchain Verification with Smart Contract Integration

**Blueprint:** User Rewards Tracker on Profile

## Summary

Implement blockchain verification for the reward calculation system — using Constellation Currency L1 `AtomicAction` bundles for atomic reward claiming (verify tier + claim + update cap in one indivisible transaction), and Data L1 anchoring of reward calculation audit events for transparency. Users can verify their rewards independently on the DAG Explorer.

## In Scope

- `AtomicAction` bundle for reward claims: ensures claim is atomic — all steps succeed or none execute (prevents partial claims and race conditions)
- Reward calculation transparency: reward calculation inputs and multipliers stored in reward claim transaction metadata (visible on-chain)
- Claim verification endpoint: `GET /v1/rewards/verify/{txHash}` — fetch on-chain claim record and return `{amount, tier, multiplier, snapshotHash, verificationURL}`
- DAG Explorer link per reward transaction (displayed in transaction history, WO-124)
- Anti-gaming on-chain enforcement: Currency L1 validators reject claims with: amount > daily cap, incorrect multiplier, duplicate nonce (idempotency)
- Reward recovery: if database shows claim pending but chain shows confirmed, backend reconciles from metagraph snapshot state

## Out of Scope

- Reward calculation algorithm (WO-181, Trust Service)
- Smart contract development (uses existing Tessellation v3 primitives)
- Gas fee optimization (platform absorbs all fees)

## Requirements

Derived from the User Rewards Tracker blueprint.

```go
// AtomicAction bundle for reward claim:
// [verifyTrustTier(did), claimReward(did, amount, type), updateDailyCap(did)]
// Currency L1 validates ALL three or rejects ALL three
// On-chain record: txHash + claim details visible on DAG Explorer
```

## Blueprints

- User Rewards Tracker on Profile — Defines blockchain recording for reward calculations, smart contract automation, calculation transparency, independent verification, and audit trail
- Data Layer — Specifies `AtomicAction` primitive for atomic multi-step Currency L1 operations

---

### WO-177: Build Decentralized Governance Voting System for Protocol Management

**Blueprint:** ECHO Token Reward System and Incentive Economy

## Summary

Build the decentralized governance voting system for ECHO protocol management — trust-tier-weighted token voting, 7-day voting periods, 20% quorum, and automatic parameter updates on proposal passage. Governance decisions (reward rate changes, burn rate adjustments, validator parameter changes) are recorded on Data L1.

## In Scope

- Governance proposal creation: `{proposalType: "reward_rate_change"|"burn_rate_adjustment"|"validator_parameter"|"ecosystem_priority", title, description, proposedValue}`
- Voting power: `StakedECHO × trustTierMultiplier` (Tier 1: 0×, Tier 2: 0.5×, Tier 3: 1.0×, Tier 4: 1.5×, Tier 5: 2.0×) — prevents plutocracy
- 7-day voting period with automatic finalization
- Quorum: 20% of total eligible voting power must participate; otherwise proposal fails
- Approval threshold: 50% for standard proposals; 67% supermajority for protocol upgrades; 75% for emergency
- Automatic parameter updates: on proposal passage, Currency L1 or Data L1 validators apply new parameter in next epoch
- Data L1 anchoring of all votes and outcomes
- Voting history: permanent record accessible via `GET /v1/governance/proposals`

## Out of Scope

- Identity verification for voters (trust tier from existing Trust Service)
- Content moderation of proposals
- External governance platforms

## Requirements

Derived from Data Layer and Frontend blueprints.

**Voting power formula (from Frontend blueprint):**
```swift
effectiveWeight = stakedECHO × trustTierMultiplier
// CEO (100M staked, Tier 5): 200M weight
// 10,000 Tier-5 community members × 10K staked each: 200M weight
// Community can outvote founders if sufficiently organized and verified
```

## Blueprints

- Data Layer — Defines governance vote rules on Data L1 (one-vote-per-DID, active proposal check, minimum stake requirement)
- Frontend — Specifies governance voting UI with trust-tier-weighted voting power and proposal management

---

### WO-178: Develop Comprehensive Tax Reporting Tools with Multi-Format Export

**Blueprint:** User Rewards Tracker on Profile

**Purpose**: Provide users with comprehensive tax reporting capabilities that simplify compliance with local tax regulations through automated calculations, standardized reporting formats, and secure documentation management.

**Requirements**:
- Export transaction data in standard tax reporting formats (IRS Form 8949, Schedule D, CSV) with all required fields populated automatically
- Calculate taxable income from reward distributions based on configurable tax jurisdictions (US, EU, UK, Canada) with current tax year rules
- Generate tax documentation including annual summaries, quarterly reports, and transaction-level details with cryptographic verification
- Ensure tax compliance by implementing validation rules for different tax jurisdictions and providing compliance status indicators
- Maintain tax auditing capabilities with complete audit trails of all tax calculations and report generations for regulatory review
- Provide tax calculation history showing year-over-year comparisons and tax liability estimates based on current earning patterns
- Support tax data encryption and secure storage with access controls ensuring only authorized users can access tax information
- Offer tax support resources including links to relevant tax guidance, calculation explanations, and professional tax advisor referrals

**Out of Scope**:
- Professional tax advice or legal guidance
- Direct filing with tax authorities
- Tax optimization strategies or recommendations

---

### WO-184: Implement ECHO Reward Coordination System

**Assignee:** Chad Cromwell

**Blueprint:** Decentralized Identity and Authentication

## Summary

Build the Rewards Service (port 8004) that distributes ECHO token rewards to users who complete identity verification. Uses `AtomicAction` bundles to atomically verify tier + claim reward + update daily cap on the Currency L1 layer. Prevents duplicate rewards per DID per verification type and tracks reward history for audit.

## In Scope

- `POST /v1/rewards/verification` — receive IDV completion signal from Identity Service, validate eligibility, queue reward submission
- Reward eligibility check: verify DID has not already received reward for this verification type (idempotent, dedup via PostgreSQL)
- Reward amounts: 100 ECHO for High-Assurance, adjustable for KYC-Lite and Proof of Humanity
- `AtomicAction` bundle construction: `[verifyTrustTier(did), claimVerificationReward(did, amount, type), updateDailyCap(did)]`
- Submission to Currency L1 via Metagraph Gateway (port 8006)
- Target delivery: within 1 block cycle (~5 seconds) of verification completion
- WebSocket push to user after Currency L1 confirmation: transaction hash + new balance
- Reward transaction history storage in PostgreSQL for audit

## Out of Scope

- Identity verification processing (IDV work orders)
- Currency L1 layer implementation (WO-8)
- Wallet balance display (iOS Wallet work orders)
- Ongoing messaging/activity rewards (separate rewards work orders)

## Requirements

Derived from the Decentralized Identity and Authentication blueprint.

**Reward Transaction Structure:**
```go
type VerificationRewardTransaction struct {
    TransactionType     string    // "verification_reward"
    UserDID             string    // "did:prism:cardano:abc123"
    RewardAmount        int64     // 100000000000000000 = 100 ECHO (18 decimal places)
    VerificationType    string    // "high_assurance", "kyc_lite", "proof_of_humanity"
    VerificationTimestamp time.Time
    IssuerDID           string    // Verification service DID
    Signature           []byte    // Backend P-256 signature
    Nonce               int64     // Prevents replay attacks
}
```

**AtomicAction Bundle (prevents partial execution):**
```go
func (s *RewardsService) SubmitVerificationReward(did string, verType string) error {
    // 1. Dedup check: prevent double-rewards
    if alreadyRewarded := s.db.HasReward(did, verType); alreadyRewarded {
        return ErrAlreadyRewarded
    }

    // 2. Build AtomicAction: all steps succeed or all fail
    bundle := AtomicActionBundle{
        Actions: []AtomicAction{
            {Type: "verify_trust_tier", DID: did},
            {Type: "claim_verification_reward", DID: did, Amount: rewardAmount(verType), VerType: verType},
            {Type: "update_daily_cap", DID: did},
        },
    }

    // 3. Submit to Currency L1 via Metagraph Gateway
    txHash, err := s.metagraph.SubmitAtomicAction(bundle)
    if err != nil { return s.queueRetry(did, verType) }

    // 4. Record reward and push WebSocket confirmation
    s.db.RecordReward(did, verType, txHash)
    s.websocket.PushRewardConfirmation(did, txHash, rewardAmount(verType))
    return nil
}
```

## Blueprints

- Decentralized Identity and Authentication — Defines verification reward flow, 100 ECHO amount, `AtomicAction` requirement, and reward notification
- Data Layer — Specifies `AtomicAction` primitive for atomic multi-step Currency L1 operations
- Backend — Defines Rewards Service (port 8004) and Metagraph Gateway integration pattern

---

### WO-186: Implement Rule-Based Bot Engine with Condition Evaluation and Action Execution

**Assignee:** Chad Cromwell

**Blueprint:** Decentralized Bot Framework and Automation

## Summary

Build the rule-based bot engine — a conditions-and-actions automation framework that developers use to build simple bots without AI. Rules define triggers (time, message, data) and actions (send message, initiate payment, call API). The engine evaluates rules and executes actions via the Bot SDK.

## In Scope

- Rule data model: `{ruleId, botDID, condition: Condition, action: Action, isEnabled, createdAt}`
- Condition types: `timeBased(cron: "0 9 * * 1-5")`, `messageReceived(pattern: String)`, `balanceThreshold(amount: Decimal)`, `webhookTriggered(url: String)`
- Action types: `sendMessage(recipientDID, content)`, `initiatePayment(userDID, amount)`, `callWebhook(url, payload)`, `chainToRule(ruleId)`
- Rule chaining: action result can trigger another rule (up to 10-rule chain depth)
- Rule scheduling: cron-based scheduling via Go `robfig/cron` library
- Rule testing: `POST /v1/bots/{id}/rules/{ruleId}/test` — dry run with sample input, returns expected action
- Rule auditing: all executions logged `{ruleId, triggeredAt, input, output, success}`
- Permission enforcement: each action checked against user's granted bot permissions before execution
- Error handling: `BotError{code, message, retryAfter}` with exponential backoff retry

## Out of Scope

- AI/NLP processing (WO-189)
- Trading-specific logic (WO-191)
- Bot SDK implementation (WO-11)

## Requirements

Derived from the Decentralized Bot Framework blueprint.

```go
type BotRule struct {
    RuleID    string
    BotDID    string
    Condition Condition
    Action    Action
    IsEnabled bool
}
```

## Blueprints

- Decentralized Bot Framework and Automation — Defines rule-based bots, condition evaluation, action execution, rule chaining, scheduling, testing, and auditing

---

### WO-188: Build Collaborative Document Editing System

**Assignee:** Chad Cromwell

**Blueprint:** Large File Sharing and Cloud Storage Integration

## Summary

Build real-time collaborative document editing by integrating a decentralized office suite (CRDT-based synchronization) into the ECHO file sharing infrastructure. Multiple users edit shared documents simultaneously with conflict resolution, version history, and offline editing. All document content remains E2E encrypted.

## In Scope

- CRDT-based real-time sync: use Yjs or equivalent CRDT library for conflict-free concurrent edits
- Document types: text documents (markdown), simple spreadsheets, rich text
- Real-time change delivery: document state deltas transmitted via WebSocket relay to all active collaborators
- Conflict resolution: CRDT algorithm resolves simultaneous edits deterministically — no manual merge required
- Version history: snapshot full document state on every 10th edit; store snapshots in encrypted IPFS (via WO-21 pipeline)
- Change tracking: per-user edit attribution (`[Alice added: "paragraph text"]`)
- Comment and annotation system: threaded comments on specific document ranges; separate from document content
- Permission model: document owner can grant `{view, comment, edit}` permissions per collaborator DID
- Offline editing: Yjs offline mode — changes queued locally, applied when WebSocket reconnects

## Out of Scope

- File encryption (WO-9)
- IPFS storage (WO-21)
- File management UI (WO-79)
- External office suite hosting

## Requirements

Derived from the Large File Sharing blueprint.

```swift
// Collaborative document session
struct DocumentSession {
    let documentId: String
    let encryptedYjsState: Data  // Current CRDT state, E2E encrypted
    let activeCollaborators: [String]  // DIDs
    let permissions: [String: EditPermission]  // DID → view|comment|edit
}
```

## Blueprints

- Large File Sharing and Cloud Storage Integration — Defines collaborative document editing: real-time sync, conflict resolution, version history, comments, permission-based access, and offline editing

---

### WO-189: Build AI-Powered Assistant Bot Framework with Privacy-Preserving NLP

**Assignee:** Chad Cromwell

**Blueprint:** Decentralized Bot Framework and Automation

## Summary

Build the AI-powered assistant bot framework using on-device NLP — no user data sent to external AI services. Developers build conversational bots using local CoreML/ONNX model inference. Handles intent recognition, entity extraction, context management, and response generation entirely on-device.

## In Scope

- On-device NLP pipeline: `SFNLTagger` for entity recognition, CoreML ONNX model for intent classification
- Local model loading: bundle quantized BERT/DistilBERT model (~50MB) with app; model versioning in manifest
- Intent recognition: classify user input into developer-defined intents (`{intent, confidence, entities}`)
- Entity extraction: extract named entities (date, amount, contact name) from user messages
- Context management: maintain conversation state `{sessionId, intentHistory[], extractedEntities{}}` in memory
- Response generation: template-based responses filled with extracted entities (`"Your balance is {amount} ECHO"`)
- Fallback handling: when confidence < 0.6, respond with "I didn't understand — could you rephrase?"
- Model update: developer pushes new model version via marketplace update
- All NLP runs on bot host device — no data transmitted to external AI API

## Out of Scope

- Training new models (developers provide pre-trained models)
- External AI API integration (privacy requirement)
- Voice/image processing
- Cloud NLP services

## Requirements

Derived from the Decentralized Bot Framework blueprint.

```swift
// Bot SDK NLP component (on-device):
struct BotNLPPipeline {
    func analyze(_ input: String, context: ConversationContext) -> NLPResult {
        let intent = intentClassifier.classify(input)   // CoreML ONNX model
        let entities = NLTagger.tag(input)              // SFNLTagger on-device
        return NLPResult(intent: intent, entities: entities, confidence: intent.confidence)
    }
}
```

## Blueprints

- Decentralized Bot Framework and Automation — Defines AI-powered assistants with local processing, zero-knowledge privacy, intent recognition, entity extraction, model management

---

### WO-191: Create Trading Bot Engine with Secure Transaction Authorization

**Assignee:** Chad Cromwell

**Blueprint:** Decentralized Bot Framework and Automation

## Summary

Build the trading bot engine — transaction authorization framework, portfolio tracking, risk management, and audit trail for bots that automate ECHO token transactions. All transactions require explicit per-trade user authorization via `AllowSpend` approval, ensuring users remain in control.

## In Scope

- Transaction authorization: every trade requires `AllowSpend` from user (`POST /v1/bots/{id}/authorization/allow-spend`) with amount limit and expiry
- Order execution: `SpendTransaction` via Metagraph Gateway using authorized `AllowSpend` → Currency L1
- Portfolio tracking: track bot's managed position `{token, amount, avgCostBasis}` in PostgreSQL
- Risk management limits: user-configurable max position size, max loss per day, stop-loss level
- Stop-loss/take-profit: bot monitors price (from PacaSwap TWAP oracle), triggers automatic sell when threshold reached
- Performance tracking: `{totalReturn, sharpeRatio, maxDrawdown, winRate}` over rolling 30-day period
- Audit trail: all trade decisions and executions logged `{timestamp, decision, amount, price, txHash}`
- Compliance recording: aggregate trade stats stored 7 years; no PII in records

## Out of Scope

- Specific exchange integrations (bot developer's responsibility)
- Trading strategy algorithms (bot developer's responsibility)
- Tax calculations

## Requirements

Derived from the Decentralized Bot Framework blueprint.

```go
// Trading bot requires explicit AllowSpend before any trade:
type TradeAuthorization struct {
    BotDID     string
    UserDID    string
    MaxAmount  Decimal      // ECHO tokens
    ExpiresAt  time.Time
    // Stored on Currency L1 via AllowSpend primitive
}
```

## Blueprints

- Decentralized Bot Framework and Automation — Defines trading bots with transaction authorization, portfolio management, risk management, performance tracking, and compliance recording

---

### WO-193: Implement Bot Analytics and Monitoring System with Privacy-Preserving Metrics

**Assignee:** Chad Cromwell

**Blueprint:** Decentralized Bot Framework and Automation

## Summary

Build the bot analytics and monitoring service — privacy-preserving metrics collection for bot developers. Tracks usage, performance, and errors without collecting user-identifiable data. Developers access their bot's metrics via dashboard and export API.

## In Scope

- Analytics event collection per bot: `{eventType: "message_processed"|"payment_initiated"|"error", timestamp, durationMs, success}` — NO user DIDs
- Aggregated metrics: daily active sessions, messages processed per day, P50/P95 response times, error rate per error type
- Error tracking: categorized error log with error code and frequency (no user context)
- Developer analytics dashboard: line charts for usage trends, bar charts for error breakdown, response time percentiles
- Performance alerting: notify developer via webhook if error rate > 5% or P95 > 5 seconds
- Anonymized engagement: session count, avg session duration — no individual user tracking
- Analytics export: `GET /v1/bots/{id}/analytics?from={date}&to={date}&format=csv`
- Historical retention: 12 months
- No external analytics platform (Mixpanel, Firebase) — all in-house, privacy-first

## Out of Scope

- Individual user behavior analysis
- Cross-bot benchmarking

## Requirements

Derived from the Decentralized Bot Framework blueprint.

```go
type BotAnalyticsEvent struct {
    BotDID       string
    EventType    string    // "message_processed", "rule_executed", "payment_initiated", "error"
    DurationMs   int
    Success      bool
    ErrorCode    string    // empty if success
    Timestamp    time.Time
    // NEVER: userDID, message content, payment amounts
}
```

## Blueprints

- Decentralized Bot Framework and Automation — Defines bot analytics with usage tracking, performance metrics, error tracking, anonymized reporting, and export capability

---

### WO-195: Build Bot Governance System with Community Voting and Policy Management

**Assignee:** Chad Cromwell

**Blueprint:** Decentralized Bot Framework and Automation

## Summary

Build the bot governance system — community-driven voting on bot policies, security standards, and marketplace guidelines. Governance proposals are submitted by community members, voted on by eligible users, and results are anchored on the Constellation Data L1 for transparency and immutability.

## In Scope

- Governance proposal creation: `{proposalType: "policy_change"|"security_standard"|"marketplace_guideline", title, description, options}`
- Eligible voters: Tier 3+ users with verified accounts and at least 100 ECHO staked
- Voting weight: `stakedECHO × trustTierMultiplier` (same formula as platform governance in Frontend blueprint)
- Quorum requirements: 5% of eligible voters for standard proposals; 10% for major policy changes
- Results anchoring: `{type: "bot_governance_vote", proposalId, result, voteCount, participationPct, closedAt}` on Data L1
- Voting transparency: public vote totals (not individual votes) visible to all
- Implementation tracking: approved proposals logged with status (pending_implementation, implemented)
- Integration: approved policies automatically update bot marketplace listing requirements

## Out of Scope

- Frontend governance UI (separate work orders)
- Complex liquid democracy voting
- Legal compliance for decisions

## Requirements

Derived from the Decentralized Bot Framework blueprint.

```go
type BotGovernanceProposal struct {
    ProposalID   string
    ProposalType string
    Title        string
    Description  string
    Options      []string
    VotingEndsAt time.Time
    QuorumPct    float64
    Status       ProposalStatus  // active, passed, failed, implemented
}
```

## Blueprints

- Decentralized Bot Framework and Automation — Defines governance voting for bot policies, security standards, marketplace guidelines, voting transparency, blockchain anchoring, and community participation

---

### WO-216: Implement AI Burn Agent and BTC Reserve Accumulation System

**Type:** Build

**Blueprint:** ECHO Token Economics and Founder Allocation

## Summary

Build the Phase 5 deflationary mechanisms: an AI Burn Agent that uses 30% of annual treasury surplus to buy ECHO from the ECHO/DAG pool via atomic swap and permanently burns it, and a BTC Reserve Agent that converts remaining surplus to Bitcoin via the Base bridge for cold storage accumulation.

## In Scope

- AI Burn Agent: automated agent that calculates 30% of annual treasury surplus → executes `AtomicAction` swap on PacaSwap (ECHO/DAG pool) → burns received ECHO via `TokenBurner` logic on Currency L1
- Burn schedule: calculated annually based on treasury surplus, executed in monthly tranches
- BTC Reserve Agent: convert surplus to USDC via ECHO/USDC pool → bridge to Base via Constellation-Base bridge → acquire BTC on Aerodrome/Base DeFi → transfer to treasury cold storage wallet
- Burn event transparency: each burn anchored on Data L1 `{type: "ai_burn", amount, txHash, treasurySurplusUsed}`
- Treasury surplus calculation: `annualRevenue - annualOperatingExpenses - 70%_reserve = surplusForBurn`
- Agent monitoring dashboard: burn rate, BTC reserve balance, cumulative burned supply, effective inflation rate
- Governance controls: DAO can adjust burn percentage (within 20–40% range) and BTC allocation

## Out of Scope

- 10% fee burn (WO-155, WO-170 — existing burn mechanisms)
- PacaSwap DEX infrastructure
- Base bridge infrastructure

## Requirements

From the ECHO Token Economics and Founder Allocation blueprint:

**Phase 5 Deflationary Mechanisms:**
- AI Burn Agent: buys ECHO from ECHO/DAG pool via atomic swap → burns. 30% of annual treasury surplus.
- Transaction fee burning: 10% of all fees permanently removed via TokenBurner logic (separate WO-155)
- BTC reserve: AI BTC Reserve Agent converts surplus to Bitcoin via Base bridge → cold storage

## Blueprints

- ECHO Token Economics and Founder Allocation — Defines Phase 5 AI Burn Agent, BTC reserve accumulation, and deflationary mechanism design

---

### WO-250: Implement ECHO Comply Service (Port 8010) with Retention Policies and Digital Evidence Fingerprinting

**Type:** Build

**Blueprint:** ECHO Comply — Enterprise Compliance Messaging

## Summary

Build the ECHO Comply Service — a new Go microservice (port 8010) that enforces configurable message retention policies for enterprise compliance customers. All Organization-tier messages receive automated integrity coverage. Retention policies are anchored to the Constellation Data L1, creating court-admissible records of when policies were activated. This is the foundation for ECHO Comply Phase 1 revenue.

## In Scope

- **New microservice: Comply Service (port 8010)** in the Go backend service architecture
- **Retention policy management:** `POST /comply/retention/policy` with types: `permanent` (never deleted), `time_limited` (deleted after N years), `litigation_hold` (deletion blocked while active)
- **Data L1 policy anchoring:** each policy activation submits `{compliance_retention: policyType, scope, orgDID, effectiveDate}` to Data L1 within 5 seconds
- **Enforcement layer:** Comply Service intercepts all deletion, expiry, and device-clearing requests for messages under active retention; blocks them and returns 403 `retention_policy_active`
- **Scope enforcement:** policies apply to ALL message types including hidden folder messages, group messages, and media attachments
- **Policy API:** `GET /comply/retention/policy` — list all active retention policies with Data L1 anchor references, scope, and effective dates
- **Organization-tier Digital Evidence fingerprinting:** every message sent by org-tier users automatically submits to Constellation Digital Evidence API; `eventID` and `verificationURL` embedded in message envelope; Smart Checkmark badge (✓) displayed in iOS
- **Comply Service pricing metadata:** associate `compliant_tier` (Starter/Professional/Enterprise) and `seats` count with each org DID

## Out of Scope

- Litigation hold (WO-251)
- eDiscovery export (WO-251)
- Compliance dashboard (WO-252)
- HIPAA clinical routing (WO-253)
- FOIA classification (WO-254)

## Requirements

From the ECHO Comply — Enterprise Compliance Messaging foundation blueprint:

**REQ-COMPLY-001.1:** Every message shall produce a commitment hash anchored on Data L1 every 5 minutes or 1000 messages.
**REQ-COMPLY-001.2:** Organization-tier messages shall be individually fingerprinted via Constellation Digital Evidence API.
**REQ-COMPLY-001.3:** `verificationURL` shall be publicly accessible without an ECHO account.
**REQ-COMPLY-002.2:** Retention policy activation shall anchor a `compliance_retention` record to Data L1 with policy type, scope, orgDID, and effective date.
**REQ-COMPLY-002.3:** The backend Comply Service (port 8010) shall enforce retention policies — disappearing messages, device clearing, and deletion requests all blocked.

**NFR-COMPLY-004:** 99.9%+ uptime SLA for Professional and Enterprise tiers.
**NFR-COMPLY-005:** 100% of Organization-tier messages have a Data L1 anchor record.

## Blueprints

- ECHO Comply — Enterprise Compliance Messaging — Defines REQ-COMPLY-001 (tamper-evident integrity), REQ-COMPLY-002 (retention policies), Digital Evidence fingerprinting, and Comply Service port 8010

---

### WO-251: Build Litigation Hold Activation and eDiscovery Export System

**Type:** Build

**Blueprint:** ECHO Comply — Enterprise Compliance Messaging

## Summary

Build the litigation hold activation system and eDiscovery export pipeline. Litigation holds must activate within 5 seconds, immediately disabling disappearing messages and activating Digital Evidence fingerprinting for all custodian communications. eDiscovery exports include full Merkle proof references and Digital Evidence event IDs, with an export checksum anchored to Data L1 for tamper-evidence.

## In Scope

- **Litigation hold activation:** `POST /comply/litigation/hold` with `{matterID, custodianDIDs[], scope}`; completes within 5 seconds:
  1. Disable disappearing messages for all custodian conversations
  2. Activate `litigation_hold` retention policy (permanent)
  3. Enable Digital Evidence fingerprinting for all custodian messages
  4. Anchor `{litigation_hold: active, matterID, custodianCount, activatedAt, activatedByDID}` to Data L1 within 30 seconds
  5. Send in-app notification to each custodian: "This conversation is under legal hold. Disappearing messages have been disabled."
- **Hold release:** `PUT /comply/litigation/hold/:matterID/release` — anchors `{litigation_hold: released, matterID, releasedAt}` to Data L1
- **Hold status query:** `GET /comply/litigation/hold/:matterID` — returns hold status, custodian count, Data L1 anchor reference
- **eDiscovery export:** `POST /comply/ediscovery/export` with `{matterID, dateRange, custodianSet, keywordFilters?}`:
  - Export package: encrypted message blobs + Merkle proof references + Digital Evidence event IDs + sender/recipient DID pairs + timestamps
  - Export checksum anchored to Data L1: `{exportID, queryHash, messageCount, requesterDID, exportTimestamp}`
  - Export status polling: `GET /comply/ediscovery/export/:exportID` — `pending | processing | ready | delivered`
  - Human-readable cover sheet with Data L1 anchor reference and independent verification instructions
  - Performance: up to 100K messages within 30 minutes

## Out of Scope

- Comply Service base infrastructure (WO-250)
- Compliance dashboard (WO-252)
- Healthcare/FOIA segment-specific features

## Requirements

From the ECHO Comply foundation blueprint:

**REQ-COMPLY-003.1:** `POST /comply/litigation/hold` activation within 5 seconds.
**REQ-COMPLY-003.2:** On hold: disappearing messages disabled; custodian messages retain permanently; Digital Evidence activated.
**REQ-COMPLY-003.3:** `litigation_hold` Data L1 anchor within 30 seconds.
**REQ-COMPLY-003.4:** In-app custodian notification.
**REQ-COMPLY-004.2:** Export package with Merkle proofs, Digital Evidence event IDs, DID pairs, timestamps.
**REQ-COMPLY-004.3:** Export checksum anchored to Data L1 for tamper verification.

**NFR-COMPLY-002:** Litigation hold activation within 5 seconds.
**NFR-COMPLY-003:** 100K message exports complete within 30 minutes.

## Blueprints

- ECHO Comply — Enterprise Compliance Messaging — Defines REQ-COMPLY-003 (litigation hold) and REQ-COMPLY-004 (eDiscovery export) with complete acceptance criteria

---

### WO-252: Build Compliance Dashboard and Audit Report Generator

**Type:** Build

**Blueprint:** ECHO Comply — Enterprise Compliance Messaging

## Summary

Build the ECHO Comply compliance dashboard and structured audit report generator. The dashboard shows real-time integrity coverage, active holds, and retention status. The audit report generates PDF/JSON documents with Data L1 anchor references suitable for regulatory submissions, OCR investigations, and court filings — all derived from on-chain records with no PII exposed.

## In Scope

- **Compliance dashboard endpoint:** `GET /comply/dashboard` returns:
  - Digital Evidence fingerprint coverage rate (% of org messages with individual DE fingerprint)
  - Number of active retention policies (count by type: permanent, time_limited, litigation_hold)
  - Number of active litigation holds
  - Pending eDiscovery export count
  - Metagraph anchor health status (last successful Data L1 anchor timestamp + latency)
- **Audit report generation:** `GET /comply/audit/report` — structured compliance audit report in JSON and PDF:
  - All compliance events in time range with Data L1 anchor references
  - Integrity coverage statistics
  - Retention policy history
  - Litigation hold history
  - Export history
  - SUITABLE FOR: OCR investigations, FOIA audits, court submissions
- **iOS Comply Dashboard view:** organization administrators see compliance status, coverage metrics, active holds, pending exports; all data sourced from Data L1 records and relay metadata
- **Zero PII in dashboard:** all metrics are aggregate counts and coverage rates — no message content, no DID linkage in reports

## Out of Scope

- Comply Service base (WO-250)
- Litigation hold and export (WO-251)
- HIPAA/FOIA segment-specific dashboards

## Requirements

**REQ-COMPLY-005.1:** Dashboard returns DE coverage rate, active retention count, active holds count, pending export count, anchor health.
**REQ-COMPLY-005.2:** Audit report in JSON + PDF suitable for regulatory examinations; includes Data L1 anchor references.
**REQ-COMPLY-005.3:** All dashboard data derived from Data L1 records and relay metadata — no PII or message content.

**NFR-COMPLY-001:** DE fingerprint generated and `eventID` returned within 2 seconds of message send.

## Blueprints

- ECHO Comply — Enterprise Compliance Messaging — Defines REQ-COMPLY-005 (compliance dashboard and audit report) with complete acceptance criteria

---

### WO-253: Implement Healthcare HIPAA Clinical Routing, 6-Year Retention, and FHIR Export

**Type:** Build

**Blueprint:** ECHO Comply — Enterprise Compliance Messaging

## Summary

Implement the Healthcare (HIPAA) segment-specific features on top of the ECHO Comply foundation: role-based clinical routing with 5-minute escalation, HIPAA 6-year retention enforcement, HL7 FHIR-compatible JSON export for EHR integration, BAA workflow, and 24-hour breach detection alerting.

## In Scope

- **Role-based clinical routing configuration:** Organization admins define routing roles (attending physician, charge nurse, unit coordinator, on-call specialist) with trust tier verification requirements; `POST /comply/healthcare/routing/rules`
- **On-call escalation:** incoming patient alert routes to on-call cardiologist (verified badge required); if no response within 5 minutes → escalates to attending physician; all routing events are Digital Evidence fingerprinted and retained permanently
- **HIPAA 6-year retention:** healthcare organization registration automatically sets 6-year minimum retention for all ePHI communications; overrides any shorter retention setting
- **HL7 FHIR-compatible export:** eDiscovery exports for healthcare org include optional FHIR-JSON format for integration with EHR systems; `exportFormat: "fhir_r4"` parameter on `POST /comply/ediscovery/export`
- **MFA enforcement:** Secure Enclave biometric required for all healthcare users (non-negotiable); backend rejects authentication tokens not signed by Secure Enclave P-256 key
- **Breach detection alerting:** monitoring for unusual access patterns to ePHI conversations; 24-hour alert to compliance dashboard; `GET /comply/healthcare/breach/alerts`
- **BAA workflow:** `POST /comply/healthcare/baa/sign` initiates HIPAA Business Associate Agreement digital signing flow; BAA status tracked in org metadata

## Out of Scope

- Base Comply Service (WO-250)
- General litigation hold and eDiscovery (WO-251)
- FOIA-specific features (WO-254)
- Full EHR integration (out of scope for ECHO)

## Requirements

From ECHO Comply foundation blueprint — Healthcare (HIPAA) segment requirements:
- Minimum retention: 6 years for all ePHI
- Encryption: AES-256 + E2E always on, no opt-out
- Access controls: role-based with verified badge requirement
- MFA enforcement: Secure Enclave biometric (non-negotiable)
- BAA: HIPAA Business Associate Agreement included for all healthcare contracts
- Breach reporting: 24-hour incident detection alerting
- Export format: HL7 FHIR-compatible JSON

## Blueprints

- ECHO Comply — Enterprise Compliance Messaging — Defines Healthcare (HIPAA) segment requirements including clinical routing, 6-year retention, FHIR export, BAA, and breach reporting

---

### WO-254: Implement FOIA Auto-Classification, Permanent Retention, and Request Deadline Tracking

**Type:** Build

**Blueprint:** ECHO Comply — Enterprise Compliance Messaging

## Summary

Implement the Local Government (FOIA) segment features on top of ECHO Comply: permanent retention for official government communications, keyword-triggered auto-classification for FOIA-scope messages, personal vs. official conversation marking, NARA-compatible archive export, and FOIA request tracking with statutory deadline alerts.

## In Scope

- **Permanent retention enforcement:** all communications by public officials in their official capacity are automatically set to `permanent` retention; government organization registration activates this policy
- **Official vs. personal conversation marking:** users can mark a conversation as "personal" to exclude from FOIA scope; personal marking anchors a `{foia_exclusion: personal, conversationID_hash, markedAt}` record to Data L1 for audit trail; admins can override personal markings
- **FOIA keyword-triggered auto-classification:** configurable keyword list triggers automatic FOIA-scope retention; `POST /comply/foia/keywords` to manage keyword rules; matches activate permanent retention + Digital Evidence fingerprinting for conversation
- **FOIA request management:** `POST /comply/foia/request` to log incoming FOIA request with `{requestID, receivedAt, responseDueDate, subject}`; `GET /comply/foia/requests` lists all pending with statutory deadline countdown
- **Deadline alerts:** automated notifications 30 days, 7 days, and 1 day before FOIA response deadline
- **NARA-compatible archive export:** eDiscovery exports for government orgs include optional NARA metadata schema; `exportFormat: "nara_m0408"` parameter
- **Personal communication privacy:** personal conversations excluded from all FOIA exports and requests; exclusion is user-controlled with audit trail

## Out of Scope

- Base Comply Service (WO-250)
- General litigation hold and eDiscovery (WO-251)
- HIPAA-specific features (WO-253)

## Requirements

From ECHO Comply foundation blueprint — Local Government (FOIA) segment requirements:
- Retention: permanent for all official government communications
- Scope: any communication by a public official in official capacity
- Auto-classification: keyword-triggered retention for FOIA-triggering terms
- Export format: NARA-compatible archive with metadata schema
- Response deadline: FOIA request tracking with statutory deadline alerts
- Personal communications: users can mark a conversation "personal" to exclude from FOIA scope

## Blueprints

- ECHO Comply — Enterprise Compliance Messaging — Defines Local Government (FOIA) segment requirements including permanent retention, auto-classification, NARA export, and FOIA request tracking

---

### WO-262: Implement Law Firm Matter Organization and Auto-Litigation Hold on Matter Creation

**Type:** Build

**Blueprint:** ECHO Comply — Enterprise Compliance Messaging, ECHO Comply — Law Firms (Chain-of-Custody)

## Summary

Implement the law firm matter management system for ECHO Comply — every conversation can be assigned to a client matter ID, and all matter-assigned conversations receive automatic permanent retention. When a new matter is created, a litigation hold activates immediately for all assigned custodians. The compliance dashboard gains a matter-centric view showing all communications organized by matter number.

## In Scope

- **Matter management data model:** `Matter` entity with `matterID`, `orgDID`, `supervisingPartnerDID`, `assignedCustodians[]`, `retentionStatus`, `holdStatus`, `createdAt`, `dataL1Anchor`; stored in PostgreSQL under the Comply Service
- **Conversation-to-matter assignment:** users assign any conversation to a matter ID at creation or anytime via conversation settings; all messages in a matter-assigned conversation automatically inherit permanent retention policy
- **Comply Service API:**
  - `POST /comply/matters` — create matter; accepts `{matterID, custodians[], supervisingPartnerDID, notes}`; triggers auto-hold (see below)
  - `GET /comply/matters` — list all active matters with message count, date range, retention status, hold status
  - `PATCH /comply/matters/:matterID/assign` — assign conversation to a matter
  - `GET /comply/matters/:matterID` — detail view with full custodian list, retention policy, hold reference
- **Auto-litigation hold on matter creation:** `POST /comply/matters` immediately calls `ComplyService.activateLitigationHold(matterID, custodians[])`:
  - Disables disappearing messages for all matter-assigned custodian conversations
  - Activates permanent retention for all custodian messages
  - Anchors `litigation_hold` marker with `status: active` on Data L1 within 30 seconds
  - Notifies all custodians via in-app push: "You have been added as a custodian on Matter [ID]. Disappearing messages have been disabled for your conversations."
- **Compliance dashboard matter view:** `GET /comply/dashboard` extended with `mattersOverview` section; per-matter row shows message count, date range, retention status, hold status, Data L1 anchor
- **Matter-searchable eDiscovery:** matter IDs filterable in `POST /comply/ediscovery/export` via `{matterID}` parameter

## Out of Scope

- Ethical wall enforcement (separate WO)
- Attorney-client privilege designation (separate WO)
- Chain-of-custody export format (separate WO)
- Litigation hold release on matter closure (matter closure flow is Phase 2+)

## Requirements

### REQ-COC-001: Matter-Based Message Organization

**User Story:** As an attorney, I want all communications automatically organized by client matter number, so that I can retrieve a complete communication record for any matter without manual sorting.

**Acceptance Criteria:**
- AC-COC-001.1: Users shall be able to assign any conversation to a client matter ID at conversation creation or any time thereafter.
- AC-COC-001.2: All messages in a matter-assigned conversation shall automatically receive the matter's retention policy (permanent unless the matter has a specific retention period).
- AC-COC-001.3: The compliance dashboard shall display all active matters with message count, date range, retention status, and hold status.
- AC-COC-001.4: Matter IDs shall be searchable and filterable in the eDiscovery export interface.

### REQ-COC-002: Automatic Litigation Hold on Matter Creation

**User Story:** As a supervising partner, I want litigation hold to activate automatically when a matter is created, so that no attorney can accidentally delete communications that may be subject to discovery obligations.

**Acceptance Criteria:**
- AC-COC-002.1: When a matter is created, a litigation hold shall automatically activate for all assigned custodians (attorneys, paralegals, and staff assigned to the matter).
- AC-COC-002.2: Hold activation shall: disable disappearing messages for all matter conversations, enforce permanent retention, activate Digital Evidence fingerprinting for all custodian messages.
- AC-COC-002.3: A `litigation_hold` marker with `status: active` shall be anchored to the Data L1 within 30 seconds of matter creation.
- AC-COC-002.4: Hold release shall require explicit action by the supervising partner or designated compliance officer. Release anchors a `litigation_hold` status update (`released`) to the Data L1.

## Blueprints

- ECHO Comply — Law Firms (Chain-of-Custody) — Defines REQ-COC-001 and REQ-COC-002 with full matter organization and auto-hold requirements
- ECHO Comply — Enterprise Compliance Messaging (Foundation) — Defines litigation hold API and Data L1 anchoring schema shared across all Comply tiers

---

### WO-263: Implement Ethical Wall Enforcement and Conflict Detection for Law Firm Accounts

**Type:** Build

**Blueprint:** ECHO Comply — Enterprise Compliance Messaging, ECHO Comply — Law Firms (Chain-of-Custody)

## Summary

Implement automated ethical wall enforcement for ECHO Comply law firm accounts — when a new matter is created, the system checks for conflicts against all existing matters and automatically blocks cross-matter communication when a conflict is detected. General Counsel override is logged on Data L1. Active ethical walls are visible in the admin console.

## In Scope

- **Conflict detection on matter creation:** when `POST /comply/matters` is called, the Comply Service checks the org's conflict database (admin-defined list or API integration); if a conflict is detected for any assigned custodian, an ethical wall is immediately activated
- **Ethical wall enforcement:** backend enforces at the Message Relay layer — any message attempt between matter groups covered by an active wall returns `HTTP 403: {"error": "ethical_wall_active", "message": "Ethical wall active — conflict of interest detected for [matter reference]."}`
- **Conflict database management:**
  - `POST /comply/conflicts` — admin registers a conflict between two matter groups with `{matter1ID, matter2ID, conflictReason, reportedByDID}`
  - `GET /comply/conflicts` — list all active conflicts with affected matter pairs
  - Admin can import conflicts from CSV or register via ECHO Comply admin console
- **Ethical wall data model:** `EthicalWall` entity with `wallID`, `conflictingMatterIDs[]`, `activatedAt`, `activatedByDID`, `status`, `dataL1Anchor`
- **Data L1 anchor on wall activation:** `{type: "ethical_wall", wallID, conflictingMatterIDs, orgDID, activatedAt}` anchored to Data L1 within 30 seconds
- **General Counsel override:**
  - `POST /comply/ethical-walls/:wallID/override` — requires General Counsel DID authentication + reason; anchors override event on Data L1 with `overriddenByDID`, `overrideReason`, `timestamp`
  - Override logged in compliance dashboard
- **Admin console display:** `GET /comply/ethical-walls` returns all active walls with conflicting matter pairs, activation dates, and override history
- **Enforcement timing:** wall takes effect within 60 seconds of conflict detection per NFR-COC-003

## Out of Scope

- Matter organization and auto-hold (separate WO)
- Privilege designation (separate WO)

## Requirements

### REQ-COC-003: Ethical Wall Enforcement

**User Story:** As an IT director at a litigation firm, I want automatic ethical walls between conflicting matters, so that attorneys on opposing sides of a conflict cannot inadvertently communicate or access each other's client information.

**Acceptance Criteria:**
- AC-COC-003.1: When a new matter is created, the system shall check for conflict status against all existing matters in the org's conflict database (via admin-defined conflict list or API integration).
- AC-COC-003.2: If a conflict is detected, an ethical wall shall automatically prevent any communication between the conflicting matter groups. Attempts to message across the wall shall return a clear error: "Ethical wall active — conflict of interest detected for [matter reference]."
- AC-COC-003.3: Ethical wall overrides shall require explicit approval from the firm's General Counsel, logged on the Data L1.
- AC-COC-003.4: The admin console shall display all active ethical walls with the conflicting matter pairs and activation dates.

**NFR-COC-003:** Ethical wall enforcement shall take effect within 60 seconds of conflict detection.

## Blueprints

- ECHO Comply — Law Firms (Chain-of-Custody) — Defines REQ-COC-003 with full ethical wall enforcement requirements
- ECHO Comply — Enterprise Compliance Messaging (Foundation) — Defines Comply Service architecture and Data L1 anchoring infrastructure

---

### WO-264: Implement Attorney-Client Privilege Designation and Automated Privilege Log Generation

**Type:** Build

**Blueprint:** ECHO Comply — Enterprise Compliance Messaging, ECHO Comply — Law Firms (Chain-of-Custody)

## Summary

Implement attorney-client privilege designation for individual messages and entire conversations within ECHO Comply law firm accounts. Privileged messages are excluded from eDiscovery exports by default. The system auto-generates an FRCP-compliant privilege log listing all privileged communications for a matter and date range, ready for production in US federal litigation.

## In Scope

- **Privilege designation on messages and conversations:**
  - iOS UI: long-press any message → "Mark as Privileged — Attorney-Client" action; conversation settings → "Designate Privileged" toggle
  - Visual indicator: "Privileged — AC" label displayed inline on message bubble (visible to conversation participants only)
  - Designation anchored on Data L1: `{type: "privilege_designation", messageID, conversationID, designatingAttorneyDID, privilegeBasis: "attorney_client" | "work_product", timestamp}`
- **Comply Service privilege tracking:** `MessagePrivilege` entity with `messageID`, `conversationID`, `orgDID`, `designatingDID`, `privilegeBasis`, `designatedAt`, `dataL1Anchor`; stored in PostgreSQL
- **eDiscovery export exclusion:** privileged messages excluded from `POST /comply/ediscovery/export` by default; export UI shows `{excludedPrivilegedCount, totalMessages}`; export requester must explicitly set `includePrivileged: true` with a `privilegeLogEntry` to override exclusion
- **Privilege log generation:**
  - `GET /comply/audit/privilege-log?matterID=&dateRange=&format=frcp` returns FRCP-compliant privilege log with fields: date, author DID, recipient DIDs, privilege basis (attorney-client / work product), description field (populated from message metadata)
  - Export formats: JSON, CSV, PDF
  - Compatible with US federal litigation privilege log requirements (FRCP Rule 26(b)(5))
- **Privilege review workflow:** the compliance dashboard shows a "Privilege Review" section listing all privilege-designated messages by matter, with ability to revoke designation

## Out of Scope

- Matter organization (separate WO)
- Ethical walls (separate WO)
- Chain-of-custody export (separate WO)

## Requirements

### REQ-COC-004: Attorney-Client Privilege Designation

**User Story:** As an attorney, I want to mark communications as attorney-client privileged so they are excluded from discovery productions unless a privilege review has been completed.

**Acceptance Criteria:**
- AC-COC-004.1: Users shall be able to mark individual messages or entire conversations as privileged via a visible "Privileged — AC" indicator.
- AC-COC-004.2: Privileged messages shall be excluded from eDiscovery exports by default. The export interface shall show a count of excluded privileged messages and require explicit inclusion with a privilege log entry.
- AC-COC-004.3: Privilege designations shall be logged on the Data L1 with the designating attorney's DID and timestamp.

### REQ-COC-006: Privilege Log Generation

**User Story:** As a paralegal conducting document review, I want to automatically generate a privilege log from all privilege-designated communications, so that I can produce the required privilege log during discovery without manual compilation.

**Acceptance Criteria:**
- AC-COC-006.1: `GET /comply/audit/privilege-log` shall generate a privilege log containing: date, author DID, recipient DIDs, privilege basis (attorney-client / work product), and description — for all privilege-designated messages in the specified matter and date range.
- AC-COC-006.2: The privilege log shall export in standard formats used in US federal litigation (FRCP-compliant privilege log format).

## Blueprints

- ECHO Comply — Law Firms (Chain-of-Custody) — Defines REQ-COC-004 and REQ-COC-006 with privilege designation and privilege log requirements
- ECHO Comply — Enterprise Compliance Messaging (Foundation) — Defines Comply Service API structure and eDiscovery export schema

---

### WO-265: Implement Law Firm Sequenced Chain-of-Custody Export with EDRM Format

**Type:** Build

**Blueprint:** ECHO Comply — Enterprise Compliance Messaging, ECHO Comply — Law Firms (Chain-of-Custody)

## Summary

Implement the law firm-specific sequenced chain-of-custody eDiscovery export — each message in the export contains its own Digital Evidence `eventID`, the preceding message's `eventID` (forming a cryptographic chain), and the batch Merkle root, creating an unbroken chain of custody. The export is compatible with Relativity, Everlaw, and Logikcull via EDRM XML and CCSF metadata formats, and includes a plain-language cover sheet explaining the verification methodology.

## In Scope

- **Sequenced chain-of-custody packaging in `POST /comply/ediscovery/export`:** when `complianceTier: "legal"` is set in the request, the export engine builds a sequenced message chain:
  ```go
  type ChainOfCustodyMessage struct {
      MessageID          string  // Unique message identifier
      Timestamp          time.Time
      SenderDID          string
      RecipientDIDs      []string
      MatterID           string
      DEEventID          string  // Digital Evidence eventID for this message
      PrecedingEventID   string  // eventID of the immediately preceding matter message
      MerkleRootRef      string  // Data L1 Merkle root anchoring this message batch
      VerificationURL    string  // Public DE verification URL
      IsPrivileged       bool    // Attorney-client privilege flag
      EncryptedBlob      []byte  // E2E encrypted message content (only participants can decrypt)
  }
  ```
- **EDRM XML export format:** export package includes `EDRM XML` envelope with per-message metadata fields compatible with Relativity, Everlaw, and Logikcull (no custom transformation required per NFR-COC-002)
- **CCSF metadata format:** alternative export format for courts using CCSF metadata schema
- **Plain-language cover sheet:** export includes a human-readable cover sheet in PDF explaining: what the chain-of-custody is, how to verify integrity using the `verificationURL` without an ECHO account, and how the Merkle root proves the export has not been altered. Language suitable for a judge unfamiliar with blockchain technology.
- **Chain integrity verification:** each `PrecedingEventID` creates a linked chain; the export package includes a `chainIntegrityReport` showing: total messages in chain, unbroken links count, any gaps (with timestamps), and the overall chain validity status
- **Data L1 export checksum:** export checksum anchored to Data L1 as in standard eDiscovery (REQ-COMPLY-004); law firm exports include additional `chainHash = SHA256(all_eventIDs_in_order)` in the checksum record
- **Privilege exclusion integration:** law firm exports exclude privilege-designated messages by default; privilege log appended to cover sheet; `excludedCount` in the export summary

## Out of Scope

- Matter organization and auto-hold (separate WO)
- Ethical walls (separate WO)
- Privilege designation logic (separate WO — this WO consumes it)
- Standard HIPAA and FOIA export formats (not EDRM)

## Requirements

### REQ-COC-005: Sequenced Chain-of-Custody Export

**User Story:** As a litigation attorney preparing for trial, I want to produce a court-admissible communication record with a cryptographic chain of custody showing every message in sequence with integrity proofs, so that opposing counsel and the court cannot challenge the authenticity of the records.

**Acceptance Criteria:**
- AC-COC-005.1: eDiscovery exports for law firm matters shall include a sequenced chain of custody: each message's Digital Evidence `eventID`, the preceding message's `eventID`, and the metagraph Merkle root anchoring all messages in the batch — forming an unbroken cryptographic chain.
- AC-COC-005.2: The export cover sheet shall explain the chain-of-custody verification methodology in plain language suitable for a judge unfamiliar with blockchain technology.
- AC-COC-005.3: All exports shall include a `verificationURL` for each Digital Evidence event, accessible by any court officer or opposing counsel without an ECHO account.
- AC-COC-005.4: The export format shall be compatible with major eDiscovery review platforms (Relativity, Everlaw, Logikcull) via standard EDRM XML or CCSF metadata format.

**NFR-COC-001:** Every message in an ECHO Comply law firm account shall have an unbroken Digital Evidence chain. Any gap in the chain shall trigger an immediate alert to the firm's compliance officer.

**NFR-COC-002:** Exports shall be compatible with Relativity, Everlaw, and Logikcull without custom transformation.

## Blueprints

- ECHO Comply — Law Firms (Chain-of-Custody) — Defines REQ-COC-005 with full chain-of-custody export requirements
- ECHO Comply — Enterprise Compliance Messaging (Foundation) — Defines eDiscovery export API, Digital Evidence integration, and Data L1 export checksum schema

---

### WO-266: Implement FOIA Personal vs. Official Communication Designation and Admin Override

**Type:** Build

**Blueprint:** ECHO Comply — Enterprise Compliance Messaging, ECHO Comply — Local Government (FOIA)

## Summary

Implement the FOIA personal vs. official communication designation system — public officials can mark conversations as personal (not government business) to exclude them from FOIA scope, subject to administrator review and override. All designation decisions are logged with the DID and timestamp. Contested conversations are preserved in their original state pending legal review.

## In Scope

- **Per-conversation personal/official toggle:**
  - iOS UI: conversation settings → "Personal Communication — Not Government Business" toggle with a confirmation prompt explaining FOIA obligations
  - Visual indicator: conversations marked personal show a "Personal" badge in the compliance dashboard view
  - Default: all conversations on a government-registered ECHO Comply account are classified as `official_record` unless explicitly marked personal
- **Designation audit log:** every designation event anchored on Data L1: `{type: "foia_designation", conversationID, orgDID, designatingDID, designation: "personal" | "official", timestamp}`; stored in `FoiaDesignation` PostgreSQL entity under Comply Service
- **Admin review and override:**
  - `GET /comply/foia/designations?status=personal` — returns all personal-designated conversations for admin review
  - `POST /comply/foia/designations/:conversationID/override` — admin overrides personal designation to official; anchors override on Data L1 with `overriddenByDID`, `overrideReason`, `timestamp`
  - Admin override requires Organization admin DID authentication
- **Contested conversation preservation:** if a designation is under review (admin has queried it but not yet overridden), the conversation is locked to its current state — no deletion or modification permitted; `status: "under_review"` shown in admin dashboard
- **Automatic official record enforcement:** keyword detection for government-business terms (configurable by admin via `POST /comply/foia/keywords`); messages containing flagged keywords automatically classified as official records even if the user has not explicitly designated them; auto-classification logged on Data L1
- **eDiscovery export integration:** `POST /comply/ediscovery/export` accepts `{foiaRequestRef, excludePersonal: true}` by default; personal-designated conversations excluded; report shows `{excludedPersonalCount, disputedConversationCount}`

## Out of Scope

- Elected official device policy and SCIM provisioning (separate WO)
- NARA-compatible export format (covered by `POST /comply/ediscovery/export` with `format: "nara"` parameter — can be added to WO-254's scope)
- FOIA statutory deadline tracking (covered by WO-254)

## Requirements

### REQ-FOIA-002: Personal vs. Official Communication Designation

**User Story:** As a public official, I want to be able to mark clearly personal conversations (not government business) as outside FOIA scope, so that my private communications are not inadvertently captured in records responses.

**Acceptance Criteria:**
- AC-FOIA-002.1: Users shall be able to mark any conversation as "Personal — not government business" via a visible toggle in the conversation settings.
- AC-FOIA-002.2: Personal designation shall be logged with the DID and timestamp of who designated it and when.
- AC-FOIA-002.3: The agency administrator shall be able to review and override personal designations. Any override shall be logged on the Data L1.
- AC-FOIA-002.4: Contested designations (official vs. personal) shall be preserved in their original state pending legal review — the system shall not delete any contested conversation.

## Blueprints

- ECHO Comply — Local Government (FOIA) — Defines REQ-FOIA-002 with full personal/official designation requirements
- ECHO Comply — Enterprise Compliance Messaging (Foundation) — Defines Comply Service API structure and Data L1 anchoring schema

---

### WO-267: Implement Elected Official Device Policy and SCIM/Active Directory Provisioning for FOIA

**Type:** Build

**Blueprint:** ECHO Comply — Enterprise Compliance Messaging, ECHO Comply — Local Government (FOIA)

## Summary

Implement the elected official device policy system for ECHO Comply local government accounts — administrators provision all elected and appointed officials via bulk CSV upload or SCIM integration with Active Directory / Google Workspace, officials receive a mandatory onboarding notification explaining their records obligations, and the admin console tracks each official's activation status.

## In Scope

- **Bulk official provisioning:**
  - `POST /comply/officials/provision/csv` — accepts CSV with columns: `officialName, email, role, department, startDate`; backend creates ECHO Comply accounts and sends onboarding invitations
  - `POST /comply/officials/provision/scim` — SCIM 2.0 endpoint for Active Directory and Google Workspace integration; syncs users on provisioning/deprovisioning events
  - `POST /comply/officials/provision/scim/deactivate` — deprovisioning endpoint; deactivated accounts are archived (records preserved) not deleted
- **Mandatory onboarding notification:** upon account provisioning, each official receives an in-app notification (and email fallback) with the text: "You have been provisioned an ECHO Comply account for official government communications. Under [applicable statute — configurable], all official communications must be retained and may be subject to public records requests. Your messages will be automatically preserved."
- **Official status tracking:** `GET /comply/officials` returns per-official status: `active` (logged in within 30 days), `never_activated` (provisioned but never logged in), `inactive` (logged in but no activity > 30 days)
- **Admin console status dashboard:** admin can see all provisioned officials with current status, last login date, and number of official-record communications; filter by department, role, activation status
- **SCIM provisioning data model:** `Official` entity with `officialDID`, `orgDID`, `role`, `department`, `provisionedAt`, `activatedAt`, `lastLoginAt`, `status`, `scimExternalID`
- **Account deactivation on official departure:** when official is deprovisioned via SCIM, account is deactivated within 60 seconds; all historical records preserved indefinitely per permanent retention policy; transition notifications sent to admin

## Out of Scope

- Personal/official designation (separate WO)
- FOIA export generation (covered by WO-254)
- Statutory deadline tracking (covered by WO-254)

## Requirements

### REQ-FOIA-005: Elected Official Device Policy

**User Story:** As an IT administrator for a municipality, I want to enforce ECHO Comply usage for official communications across all elected and appointed officials, so that government business does not migrate to unpreserved personal apps.

**Acceptance Criteria:**
- AC-FOIA-005.1: Organization administrators shall be able to provision ECHO Comply accounts for all officials via bulk CSV upload or SCIM integration with the government's directory (Active Directory, Google Workspace).
- AC-FOIA-005.2: Once provisioned, officials shall receive a mandatory onboarding notification explaining their obligations under the applicable records law.
- AC-FOIA-005.3: The admin console shall show each official's account status: active, never activated, or inactive (last login > 30 days).

## Blueprints

- ECHO Comply — Local Government (FOIA) — Defines REQ-FOIA-005 with full device policy and SCIM provisioning requirements
- ECHO Comply — Enterprise Compliance Messaging (Foundation) — Defines organization provisioning infrastructure and Comply Service admin APIs

---

### WO-268: Implement HIPAA MFA Session Enforcement, OCR Audit Reporting, and 24-Hour Breach Detection

**Type:** Build

**Blueprint:** ECHO Comply — Enterprise Compliance Messaging, ECHO Comply — Healthcare (HIPAA)

## Summary

Implement HIPAA-specific MFA session enforcement, OCR-ready audit report generation, and 24-hour breach detection alerting for ECHO Comply Healthcare accounts. Session timeout (15 minutes of inactivity, biometric re-auth required) and biometric-only enforcement (no fallback to PIN or password for ePHI access) are non-negotiable. The audit report exports in HL7 FHIR-compatible JSON for EHR integration.

## In Scope

- **HIPAA session timeout enforcement:**
  - All Healthcare tier users: session automatically locks after 15 minutes of inactivity (configurable by admin, default 15 minutes, minimum 5 minutes, maximum 30 minutes)
  - On timeout: app enters locked state; user must re-authenticate with Face ID / Touch ID to continue
  - Timeout enforced at the iOS application layer (`SceneDelegate.sceneDidEnterBackground`) and by backend session tokens (server-side session TTL matches client timeout setting)
  - Backend endpoint: `POST /comply/healthcare/policy/session-timeout` — admin sets org-wide session timeout
- **Biometric-only enforcement:**
  - For Healthcare tier org DIDs: device passcode is the ONLY permitted fallback (PIN codes and password fallbacks disabled); enforced in `BiometricAuthManager.swift`
  - Admin cannot disable biometric requirement — attempting to set `biometricRequired: false` for Healthcare tier returns `HTTP 400: {"error": "healthcare_biometric_required", "message": "Biometric authentication cannot be disabled for HIPAA-covered organizations."}`
  - Backend validates Healthcare tier org DID on every authenticated request and enforces stricter token validation (shorter JWT TTL matching session timeout)
- **HL7 FHIR-compatible OCR audit report:**
  - `GET /comply/audit/report?format=hipaa_fhir` returns full audit report in HL7 FHIR-compatible JSON
  - Report includes: all active retention policies with Data L1 anchors, all active litigation holds, message count, fingerprint coverage rate, breach detection events, session timeout events (non-PII)
  - PDF export option for OCR production submissions
- **24-hour breach detection alert:**
  - The Comply Service monitors Digital Evidence fingerprint coverage for Healthcare org DIDs on a rolling 60-minute cadence
  - If any message was sent but no fingerprint received within 2 seconds (NFR-HIPAA-002), a "Preservation Gap" alert fires to:
    1. Compliance dashboard: `{alertType: "fingerprint_gap", orgDID, affectedMessageCount, detectedAt}`
    2. Admin email address (configurable via `POST /comply/healthcare/policy/breach-contact`)
  - Alert must fire within 60 minutes of gap detection to support HIPAA 24-hour breach notification timeline
  - `GET /comply/healthcare/breach-alerts` returns all historical gap alerts for auditing

## Out of Scope

- 6-year retention and clinical routing (covered by WO-253)
- HIPAA BAA (legal contract — not a code WO)
- Digital Evidence fingerprinting infrastructure (covered by WO-250)

## Requirements

### REQ-HIPAA-005: MFA Enforcement

**User Story:** As a HIPAA security officer, I want multi-factor authentication enforced for all users accessing ePHI, so that I comply with the HIPAA Security Rule requirement for entity authentication controls.

**Acceptance Criteria:**
- AC-HIPAA-005.1: All Healthcare tier users shall authenticate via Secure Enclave biometric (Face ID / Touch ID). This satisfies HIPAA's "something you are" MFA requirement.
- AC-HIPAA-005.2: The organization administrator shall not be able to disable biometric authentication. Device passcode is the only permitted fallback.
- AC-HIPAA-005.3: Session timeout shall be enforced at a configurable interval (default: 15 minutes of inactivity). Users must re-authenticate with biometrics after timeout.

### REQ-HIPAA-006: Audit Trail and OCR Reporting

**User Story:** As a compliance officer under an OCR audit, I want to produce a complete, independently verifiable audit trail of all communications, so that I can demonstrate compliance without relying on ECHO's cooperation.

**Acceptance Criteria:**
- AC-HIPAA-006.1: All messages shall be fingerprinted via the Digital Evidence API. The `verificationURL` shall be publicly accessible for independent verification.
- AC-HIPAA-006.2: `GET /comply/audit/report` shall generate an OCR-ready compliance report including: all active retention policies with Data L1 anchors, all active litigation holds, message count, fingerprint coverage rate, and breach detection events.
- AC-HIPAA-006.3: The audit report shall export in HL7 FHIR-compatible JSON format for integration with hospital EHR systems.
- AC-HIPAA-006.4: A 24-hour breach detection alert shall be configured: if the Comply Service detects any gap in Digital Evidence fingerprint coverage (message sent but fingerprint not received), an alert shall fire to the compliance dashboard and the admin email address.

**NFR-HIPAA-001:** 99.9% uptime with contractual SLA. Downtime notifications within 15 minutes of outage detection.
**NFR-HIPAA-002:** System must detect and alert on fingerprint coverage gaps within 60 minutes to support the 24-hour HIPAA breach notification timeline.

## Blueprints

- ECHO Comply — Healthcare (HIPAA) — Defines REQ-HIPAA-005 and REQ-HIPAA-006 with MFA enforcement, session timeout, OCR reporting, and breach detection requirements
- ECHO Comply — Enterprise Compliance Messaging (Foundation) — Defines Comply Service API and Digital Evidence fingerprinting infrastructure

---

### WO-269: Implement ECHO Comply Organization Channel Compliance Integration

**Type:** Build

**Blueprint:** Broadcast Channels and Community Features, ECHO Comply — Enterprise Compliance Messaging

## Summary

Implement the ECHO Comply organization channel compliance integration — when an ECHO Comply organization administrator creates a broadcast channel, the Comply Service automatically registers it under the org's active retention policy, enables Digital Evidence fingerprinting for all posts, adds the channel to the organization's eDiscovery scope, and propagates litigation holds to channel content.

## In Scope

- **`OnOrgChannelCreated` compliance hook:** when `POST /channels` is called by an Organization-tier DID, the backend automatically:
  1. Calls `ComplyService.OnOrgChannelCreated(channelID, orgDID)` (Go)
  2. Registers the channel under the org's active retention policy
  3. Activates Digital Evidence fingerprinting for all posts in that channel
  4. Adds the channel to the org's eDiscovery scope
  5. Anchors a `compliance_retention` record on Data L1: `{type: "compliance_retention", orgDID, channelID, policyType: orgPolicy.type, effectiveAt}`
  ```go
  func (cs *ComplyService) OnOrgChannelCreated(channelID, orgDID string) error {
      policy, err := cs.getActiveRetentionPolicy(orgDID)
      if err != nil { return err }
      return cs.registerChannelCompliance(channelID, orgDID, policy)
  }
  ```
- **Per-post Digital Evidence fingerprinting:** every post to a Comply-registered org channel triggers `MediaService.submitFingerprint` before E2E encryption; `eventID` and `verificationURL` embedded in post envelope; Smart Checkmark (✓) badge shown on all org channel posts
- **Channel eDiscovery scope inclusion:** `POST /comply/ediscovery/export` includes org channel posts when `includeChannels: true` (default: true for org-tier exports); per-post metadata includes channel ID, post timestamp, Digital Evidence event ID
- **Litigation hold propagation to channels:**
  - When a hold activates via `POST /comply/litigation/hold`, the Comply Service checks if any affected custodian is a channel admin or member; if yes, the channel is flagged as hold-covered
  - Hold-covered channel behavior: posts cannot be deleted or edited while hold is active; scheduled posts are suspended pending admin review
  - Channel admin receives in-app notification: "This channel is under legal hold (Matter [ID]). All content is subject to preservation. Scheduled posts have been suspended."
- **Compliance dashboard channel section:** `GET /comply/dashboard` extended with `channelsOverview` section showing per-channel retention status, fingerprint coverage rate, hold status
- **Disappearing content blocked on hold-covered channels:** if a channel post has an expiration timer and a hold activates before the timer expires, the timer is suspended (post not deleted); on hold release, admin is notified and must consciously re-enable timers

## Out of Scope

- Core channel creation, publishing, and moderation (covered by Broadcast Channels WOs in Phase 5)
- Litigation hold core infrastructure (covered by WO-251)
- Digital Evidence fingerprinting API (covered by WO-250)

## Requirements

From Broadcast Channels blueprint (ECHO Comply Org Channel section):

- All messages sent on Comply org channels shall be retained per the org's active retention policy. Disappearing content is blocked when a litigation hold is active.
- Digital Evidence fingerprinting is automatic for all Comply org channel posts. Smart Checkmark badge visible on all org channel posts.
- Channel posts are included in eDiscovery exports when the channel is covered by the export's scope.
- When a litigation hold activates covering a channel custodian, channel posts cannot be deleted or edited; scheduled posts are suspended pending admin review.

## Blueprints

- Broadcast Channels and Community Features — Defines ECHO Comply organization channel compliance behavior: auto-registration, litigation hold propagation, eDiscovery scope inclusion
- ECHO Comply — Enterprise Compliance Messaging (Foundation) — Defines Comply Service architecture, Digital Evidence integration, and Data L1 anchoring infrastructure

---

### WO-281: Implement ECHO Comply Organization Lifecycle and Org DID Minting

**Type:** Build

**Blueprint:** Backend, ECHO Comply — Enterprise Compliance Messaging

## Summary

Implement the ECHO Comply organization entity lifecycle in the Comply Service (`services/comply/internal/organization/`). Organization creation is split into three phases to avoid blocking the signup UX on Identity Metagraph latency. The org's `did:key` is derived from a platform KMS-managed keypair — it is a peer identity independent of any admin's personal DID, persisting through admin turnover.

## In Scope

- **`organizations` table migration:**
  ```sql
  CREATE TABLE organizations (
      id UUID PRIMARY KEY,
      org_did TEXT UNIQUE,                            -- NULL until async DID mint completes
      legal_name TEXT NOT NULL,
      display_name TEXT NOT NULL,
      primary_color TEXT NOT NULL DEFAULT '#4F46E5',
      logo_url TEXT,
      tier TEXT NOT NULL CHECK (tier IN ('starter', 'professional', 'enterprise')),
      seat_count_paid INTEGER NOT NULL DEFAULT 10,
      seat_count_used INTEGER NOT NULL DEFAULT 0,
      status TEXT NOT NULL CHECK (status IN ('active', 'enterprise_grace', 'suspended', 'terminated')),
      enterprise_grace_started_at TIMESTAMPTZ,
      enterprise_grace_expires_at TIMESTAMPTZ,
      domain TEXT, domain_verified_at TIMESTAMPTZ,
      stripe_customer_id TEXT, stripe_subscription_id TEXT,
      created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      created_by_did TEXT NOT NULL, terminated_at TIMESTAMPTZ
  );
  ```
- **3-phase org creation:**
  - **Phase 1 (sync, ~100ms):** `INSERT organizations (org_did = NULL)`, `INSERT baa_signatures`, `INSERT memberships (admin, current_credential_id = NULL)`, `INSERT status_lists (empty 16KB bit vector)` → return `CreateOrgOutput` immediately
  - **Phase 2 (async, 5–15s):** derive org `did:key` from KMS-managed org keypair (no chain transaction); register org VC on Constellation Identity Metagraph; `UPDATE organizations SET org_did = 'did:key:...'`; emit `OrganizationProvisioned` event
  - **Phase 3 (async, follows Phase 2):** issue Owner credential for admin's membership; allocate StatusList bit; sign VC with org's Ed25519 key; publish initial StatusList to Identity Metagraph
- **Org DID architecture:** org `did:key` derived from a KMS-managed org keypair (not the admin's Secure Enclave key); org DID is a peer identity that signs credentials independently and persists through admin turnover
- **Org lifecycle endpoints:**
  - `POST /v1/comply/organizations` — CreateOrg (Phase 1 sync response)
  - `GET /v1/comply/organizations/{id}` — GetOrg
  - `PATCH /v1/comply/organizations/{id}` — UpdateOrg (display name, logo, color)
- **iOS polling:** app polls `GET /v1/comply/organizations/{id}` until `org_did` is non-null (credential available within 15–30s)

## Out of Scope

- BAA signing (WO-281 — Implement ECHO Comply BAA Signature Capture)
- Member invitation and credential issuance (WO-282)
- StatusList management (WO-283)
- SSO configuration (WO-284)
- Seat enforcement (WO-285)

## Requirements

From Backend blueprint (Comply Service Architecture — Organization Creation: 3-Phase Flow):

**Key design decision:** Org DID is a peer identity, not a sub-resource of the admin's DID. The relationship is expressed as a Verifiable Credential (admin holds an `Owner-role` credential issued by the org DID), not as a parent-child DID hierarchy.

**Phase 2 async invariant:** `org_did` remains NULL and admin credential is not issued until the Identity Metagraph confirms VC registration. iOS shows a "Setting up your organization…" screen during this window.

## Blueprints

- Backend — Defines the Comply Service 3-phase org creation flow, `organizations` table schema, org DID architecture, and endpoint surface
- ECHO Comply — Enterprise Compliance Messaging (Foundation) — Defines the overall Comply Service product scope and compliance requirements

---

### WO-282: Implement ECHO Comply BAA Signature Capture and JWS Verification

**Type:** Build

**Blueprint:** Backend, ECHO Comply — Enterprise Compliance Messaging

## Summary

Implement the ECHO Comply BAA (Business Associate Agreement) signature capture and lifecycle in the Comply Service (`services/comply/internal/baa/`). BAA acceptance is a cryptographically-signed JWS artifact — not a checkbox log entry — so the admin's physical device participation is provable indefinitely from the `did:key` public key embedded in their DID identifier.

## In Scope

- **`baa_signatures` table migration:**
  ```sql
  CREATE TABLE baa_signatures (
      id UUID PRIMARY KEY,
      organization_id UUID NOT NULL REFERENCES organizations(id),
      signer_did TEXT NOT NULL,
      signer_legal_name TEXT NOT NULL,
      signer_role TEXT NOT NULL,
      baa_version TEXT NOT NULL,           -- versioned template (e.g. "v1.2")
      baa_document_hash TEXT NOT NULL,     -- SHA-256 of signed PDF
      signed_assertion TEXT NOT NULL,      -- JWS: signed by signer's Secure Enclave key
      pdf_storage_url TEXT NOT NULL,
      signed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      superseded_at TIMESTAMPTZ,
      return_or_destroy_window_ends_at TIMESTAMPTZ,  -- 30 days after cancellation
      purge_completed_at TIMESTAMPTZ,
      purge_attestation_hash TEXT          -- SHA-256 of destruction attestation
  );
  ```
- **JWS verification on BAA acceptance:** the backend verifies the JWS signed assertion:
  1. Signature valid against the admin's `did:key` public key (extracted directly from the DID — no chain lookup)
  2. `document_hash` matches SHA-256 of the canonical BAA version presented
  3. `signed_at` within 5 minutes (replay prevention)
  4. `signer_did` matches authenticated admin's DID
- **PDF generation:** generate a signed BAA PDF from the canonical versioned template + the admin's legal name, role, signature timestamp, and `signed_assertion` reference; store to Storj/S3; record URL in `baa_signatures.pdf_storage_url`
- **BAA versioning:** BAA template is versioned (`v1.0`, `v1.1`, etc.); new org signups always get the current version; existing orgs are prompted to re-sign on version change
- **Purge lifecycle:** 30 days after org cancellation, scheduled job triggers return-or-destroy procedure for ePHI data (HIPAA requirement); `purge_attestation_hash` records SHA-256 of the destruction attestation
- **Endpoints:**
  - `POST /v1/comply/organizations/{id}/baa` — SignBAA (accepts JWS assertion from iOS)
  - `GET /v1/comply/organizations/{id}/baa/current` — GetCurrentBAA
  - `GET /v1/comply/organizations/{id}/baa/{sigID}/pdf` — DownloadBAAPDF (authenticated; URL-signed S3 redirect)
- **iOS BAA signing flow:** app presents the BAA document, admin enters their legal name and role, taps "I agree and sign" → Secure Enclave signs `{signer_did, org_name, baa_version, document_hash, signed_at, nonce}` → submit JWS to `POST /v1/comply/organizations/{id}/baa`

## Out of Scope

- Org creation (WO-281)
- Member invitation/acceptance (WO-282)
- HIPAA data retention policies (WO-250)

## Requirements

From Backend blueprint (Comply Service Architecture — BAA JWS Verification section):

**Verification invariant:** If a customer disputes signing the BAA, the platform produces the JWS, proves the admin's `did:key`-bound Secure Enclave key signed it (proof of physical device participation), and proves the document hash matches the canonical BAA. The `did:key` public key embedded in the DID is the cryptographic anchor — no chain dependency for future verification.

## Blueprints

- Backend — Defines the BAA JWS verification flow, `baa_signatures` schema, and purge lifecycle
- ECHO Comply — Enterprise Compliance Messaging (Foundation) — Defines HIPAA BAA requirement as mandatory for all healthcare tier contracts

---

### WO-283: Implement ECHO Comply Member Invitation, Acceptance, and Org Role Credential Issuance

**Type:** Build

**Blueprint:** Backend, ECHO Comply — Enterprise Compliance Messaging, Universal Onboarding and Identity Creation

## Summary

Implement the ECHO Comply member invitation, acceptance, and org role credential issuance flow in the Comply Service (`services/comply/internal/membership/` and `services/comply/internal/invitation/`). When an invitation is accepted, the backend issues an `EchoOrgRoleCredential` (W3C VC 2.0) and the user's org context appears in the iOS context switcher. The flow handles both new users (Branch A: first-run + invite) and existing ECHO users (Branch B: auth + invite).

## In Scope

- **`memberships` and `invitations` table migrations** (see Backend blueprint schema)
- **Invitation creation:** `POST /v1/comply/organizations/{id}/members/invite` — generate 32-byte cryptographically random base64url token; 7-day expiry; store to `invitations` table
- **Pre-check endpoint:** `GET /v1/comply/invitations/{token}/resolve` — unauthenticated; returns `{ user_status: "new" | "existing", org_name, org_primary_color, org_logo_url, role, suggested_display_name, custom_message, role_assigned_by_display_name, expires_at }`; determines branch by checking if invited email has a registered DID
- **Branch A (new user):** invitation token flows through standard ECHO registration (Path A or B); after DID creation, `POST /v1/comply/invitations/{token}/accept` is called; backend creates membership + issues `EchoOrgRoleCredential`
- **Branch B (existing user):** biometric auth on existing personal DID → transparency screen → `POST /v1/comply/invitations/{token}/accept` → new org context added
- **Transparency screen:** displayed in both branches before acceptance — clearly shows what the org will and will NOT see (no personal messages, no personal trust tier)
- **`EchoOrgRoleCredential` issuance on acceptance:** signed by org's Ed25519 key; subject is user's `did:key`; fields include `orgDID`, `role`, `orgScopedDisplayName`, `department`; StatusList2021 bit allocated; VC delivered to iOS wallet
- **Idempotency:** re-accepting same (DID, token) returns existing membership ID; different DID on same token returns 409 (token hijacking prevention)
- **Member management endpoints:** `GET /v1/comply/organizations/{id}/members`, `POST /v1/comply/organizations/{id}/members/{did}/revoke`, `PATCH /v1/comply/organizations/{id}/members/{did}` (role change → new VC)
- **Invitation revocation (pre-accept):** `DELETE /v1/comply/invitations/{token}`
- **iOS `InvitationCoordinator`:** resolves token → branches to `newUser(ResolvedInvitation)` or `existingUser(ResolvedInvitation)` → presents branded org welcome screen → transparency screen → accept CTA

## Out of Scope

- Org creation (WO-281)
- BAA signing (WO-282)
- StatusList2021 batch publication infrastructure (WO-283 — this WO calls into that service for bit allocation)
- SSO-based provisioning (WO-284)

## Requirements

From Backend blueprint (Comply Service — membership and invitation packages) and Universal Onboarding blueprint (ECHO Comply Invited Member Flow):

**Transparency screen invariant:** Both branches MUST show the transparency screen before acceptance. This is the core trust mechanism — users must see exactly what the org will and will not see before joining.

**VC issuance timing:** `EchoOrgRoleCredential` must be available in the iOS wallet within 15 seconds of acceptance (subject to Identity Metagraph finality).

## Blueprints

- Backend — Defines the `membership` and `invitation` package architecture, database schemas, `EchoOrgRoleCredential` issuance flow, and full endpoint surface
- Universal Onboarding and Identity Creation — Defines the iOS invitation flow, Branch A/B pre-check, transparency screen UX, and `InvitationCoordinator`
- ECHO Comply — Enterprise Compliance Messaging (Foundation) — Defines the org role credential and member management requirements

---

### WO-284: Implement ECHO Comply Org-Scoped StatusList2021 and Member Credential Lifecycle

**Type:** Build

**Blueprint:** Backend, ECHO Comply — Enterprise Compliance Messaging

## Summary

Implement the org-scoped StatusList2021 credential lifecycle in the Comply Service (`services/comply/internal/credential/` and `services/comply/internal/credential/status_list/`). Each organization maintains its own 131,072-bit revocation vector. Changes batch-publish to the Constellation Identity Metagraph every 5 minutes. Guest credentials have a 90-day hard expiry with auto-revocation on the next publication cycle.

## In Scope

- **`credentials` and `status_lists` table migrations** (see Backend blueprint schema)
- **Credential issuance on membership acceptance:** called by WO-283 (invitation acceptance); allocates next available bit from `status_lists.next_available_bit` (monotonically increasing, never reused); signs `EchoOrgRoleCredential` VC with org's Ed25519 key; stores JWS in `credentials.payload_jws`
- **PostgreSQL advisory lock on bit allocation:** `SELECT pg_advisory_xact_lock(org_id_hash, list_index)` before `next_available_bit` increment — prevents concurrent issuances from claiming the same bit position
- **5-minute batch publication loop:** background goroutine queries `status_lists WHERE pending_publication = TRUE`; signs current bit vector as StatusList2021 VC; publishes to Constellation Identity Metagraph; updates `last_published_at`, `last_publication_tx_hash`, `pending_publication = FALSE`
- **Credential revocation:** `POST /v1/comply/organizations/{id}/members/{did}/revoke` flips bit in `status_lists.bit_vector`, sets `pending_publication = TRUE`; revocation propagates within 5 minutes + metagraph finality
- **Guest 90-day expiry:** Guest credentials have `expires_at = accepted_at + 90 days`; daily job queries `credentials WHERE role = 'guest' AND expires_at < NOW() AND revocation_status = 0` → marks revoked + sets `pending_publication = TRUE`; renewal reminder sent at 14 days before expiry
- **StatusList2021 public endpoint (unauthenticated):** `GET /v1/comply/organizations/{id}/status-list/{listIdx}` — returns the current StatusList2021 VC; verifiers fetch without logging their identity (privacy-preserving revocation checking)
- **Credential refresh:** `POST /v1/comply/organizations/{id}/members/{did}/credential/refresh` — revokes current VC, issues new one (used on role change)
- **New StatusList when current fills:** when `next_available_bit = capacity (131072)`, provision a new `status_lists` row with `list_index + 1`

## Out of Scope

- Identity Service StatusList21 for non-org credentials (WO-274)
- Org creation and DID minting (WO-281)
- Invitation/acceptance flow (WO-283)

## Requirements

From Backend blueprint (Comply Service Architecture — StatusList2021 Publication section):

**Bit reuse invariant:** Bits are NEVER reused after revocation. `next_available_bit` is monotonically increasing. Reusing a revoked bit would produce a misleading audit trail (a briefly-active appearance during the publish cycle).

**Revocation propagation SLA:** P99 revocation propagation ≤ 5 minutes (1 publication cycle) + metagraph snapshot finality (~15s target).

**Privacy invariant for StatusList endpoint:** The endpoint is unauthenticated because verifiers fetching a credential status list must not reveal which credential they are checking. A verifier fetches the entire vector and checks the bit locally.

## Blueprints

- Backend — Defines StatusList2021 publication batching, advisory lock bit allocation, Guest expiry, `credentials` and `status_lists` schemas, and the credential lifecycle endpoint surface
- ECHO Comply — Enterprise Compliance Messaging (Foundation) — Defines the compliance VC requirements

---

### WO-285: Implement ECHO Comply SSO Configuration Management (SAML, OIDC, SCIM)

**Type:** Build

**Blueprint:** Backend, ECHO Comply — Enterprise Compliance Messaging

## Summary

Implement the ECHO Comply SSO configuration management in the Comply Service (`services/comply/internal/sso/`). Supports SAML, OIDC, and SCIM with KMS-encrypted secrets. All credentials (SAML certificates, OIDC client secrets, SCIM bearer tokens) are encrypted at rest using the platform KMS — storing them in plaintext is a HIPAA violation.

## In Scope

- **`sso_configs` table migration** (see Backend blueprint schema) — stores KMS-encrypted secrets; no plaintext credentials ever in PostgreSQL
- **SAML configuration:** `PUT /v1/comply/organizations/{id}/sso/saml` — accepts entity ID, SSO URL, X.509 certificate; encrypts certificate with KMS before storage; validates SAML metadata URL is reachable
- **OIDC configuration:** `PUT /v1/comply/organizations/{id}/sso/oidc` — accepts issuer URL, client ID, client secret; encrypts client secret with KMS; validates OIDC discovery endpoint (`/.well-known/openid-configuration`)
- **SCIM provisioning:** `PUT /v1/comply/organizations/{id}/sso/scim` — accepts SCIM bearer token; encrypts with KMS; configures SCIM 2.0 endpoint for automated member provisioning/deprovisioning (aligns with FOIA SCIM WO-267 pattern)
- **Test round-trip:** `POST /v1/comply/organizations/{id}/sso/test` — performs a test SSO authentication against the configured IDP and reports `last_test_status: "success" | "failure"`
- **Certificate expiry alerts:** background job checks `cert_expires_at` for all SAML configs; sends compliance dashboard alert + admin email at 60 days before expiry; sets `last_test_status = 'cert_expiry_warning'`
- **Domain verification:** `POST /v1/comply/organizations/{id}/domain/verify` — DNS TXT record verification to confirm org owns the claimed domain; sets `domain_verified_at` on success
- **Enforcement modes:**
  - `required` — all org users must authenticate via SSO; non-SSO access blocked
  - `optional_for_guests` — guests can use personal ECHO DID; members/admins require SSO
  - `optional` — SSO available but not enforced
- **SSO + DID integration:** on successful SSO auth, backend either links the SSO identity to an existing `did:key` or triggers Branch B of the invitation acceptance flow for new org members

## Out of Scope

- FOIA-specific elected official SCIM provisioning (WO-267 — covers the FOIA government-specific use case)
- Org creation (WO-281)
- Seat enforcement (WO-285)

## Requirements

From Backend blueprint (Comply Service Architecture — `sso_configs` schema and SSO endpoint surface):

**Security invariant:** SCIM bearer tokens in plaintext are a HIPAA violation per the Backend blueprint. All secrets stored in `sso_configs` MUST be KMS-encrypted before INSERT/UPDATE.

**Enforcement mode default:** `optional` — org can deploy without forcing SSO immediately, then enforce after validating the configuration with the test round-trip.

## Blueprints

- Backend — Defines `sso_configs` schema, SSO endpoint surface, enforcement modes, KMS encryption requirement, cert expiry alerting, and domain verification
- ECHO Comply — Enterprise Compliance Messaging (Foundation) — Defines SSO/SCIM as required for Professional and Enterprise tier org accounts

---

### WO-286: Implement ECHO Comply Seat Enforcement and Stripe Billing Integration

**Type:** Build

**Blueprint:** Backend, ECHO Comply — Enterprise Compliance Messaging

## Summary

Implement ECHO Comply seat enforcement and Stripe billing integration in the Comply Service (`services/comply/internal/seats/`). The Starter tier has a hard cap at 10 seats. Professional tiers have soft triggers at 200 and 500 seats that move orgs toward Enterprise. When an org exceeds 500 seats without an Enterprise contract, they enter a 30-day grace period before suspension.

## In Scope

- **Seat cap enforcement at invitation time:** before issuing an invitation, check `seat_count_used < seat_count_paid`; return `HTTP 402` with `{ "error": "seat_limit_reached", "tier": "starter", "current": 10, "limit": 10 }` if hard-capped
- **Soft trigger notifications:**
  - Professional ≥50 seats: internal sales notification (no user-facing block)
  - Professional ≥200 seats: "Enterprise eligibility" email to admin
  - Professional ≥500 seats: transition org to `enterprise_grace` status; 30-day grace window; admin receives "Upgrade to Enterprise" prompt
- **Enterprise grace period:**
  - Full Professional access continues during grace
  - `enterprise_grace_started_at` and `enterprise_grace_expires_at` set at transition
  - Grace expiry: daily reconciliation job — `SELECT orgs WHERE status = 'enterprise_grace' AND enterprise_grace_expires_at < NOW()` → `SET status = 'suspended'` → notify admin
- **Org suspension behavior:** suspended orgs cannot send/receive messages; admin sees suspension reason on Comply dashboard; re-activation on Enterprise contract signed
- **Stripe webhook sync:** `subscription.updated` event → `UPDATE organizations SET seat_count_paid = new_quantity`; if `seat_count_paid < seat_count_used` post-downgrade: surface warning in admin console but do NOT auto-revoke members (no surprise disruption)
- **`seat_count_used` tracking:** incremented on membership acceptance; decremented on membership revocation; atomic with the membership transaction

## Out of Scope

- Stripe payment processing and checkout (handled by Stripe; this WO handles webhook-driven seat count sync only)
- Org creation (WO-281)
- Member invitation enforcement is a gate check in WO-283 that calls this service

## Requirements

From Backend blueprint (Comply Service Architecture — Seat Enforcement and Enterprise Grace Period):

**Downgrade invariant:** When `seat_count_paid` drops below `seat_count_used` (e.g., Stripe payment fails and plan downgrades), the platform surfaces a warning but does NOT automatically revoke member credentials. Sudden mass-revocation would violate HIPAA messaging continuity requirements.

**Starter hard cap:** exactly 10 seats; no grace, no soft trigger. Invitations are blocked before issuance with a clear error message directing the admin to upgrade.

## Blueprints

- Backend — Defines tier caps, grace period logic, daily reconciliation job, Stripe webhook sync, and seat count tracking invariants
- ECHO Comply — Enterprise Compliance Messaging (Foundation) — Defines pricing tiers (Starter $30/seat/10 min, Professional $50/seat/50 min, Enterprise $80–100/seat/500 min)

---

### WO-289: Build ECHO Comply iOS Context Coordinator, Composer Guardrails, and Credential Card

**Type:** Build

**Blueprint:** ECHO Comply — Enterprise Compliance Messaging, Frontend

## Summary

Implement the ECHO Comply iOS UX foundation: the `ContextCoordinator` singleton that drives all context-dependent rendering for Personal vs. Organization contexts, the `ComposerGuardrailsService` that prevents cross-context message leakage, and `CredentialCardView` that renders org membership credentials from VC display metadata. These components activate only for users with valid org membership VCs; standard ECHO Message users never see them. WO-281–286 cover the backend org lifecycle; this WO covers the iOS layer that makes it usable.

## In Scope

**ContextCoordinator (`@Observable @MainActor` singleton):**
- `loadAvailableContexts()` — reads all org membership VCs from credential store, verifies each against StatusList2021 bit vector on Constellation Identity Metagraph (not revoked), builds `OrganizationContext` from VC display metadata (color, logo, role)
- `switchTo(_ context: AppContext)` — updates `current`, resets `messagesSinceLastSwitch` to 0, triggers haptic feedback + crossfade animation, async-loads context-scoped conversation list
- Injected into SwiftUI environment at app launch as the single source of truth for active context
- Header band recolors to org brand color on switch; message list filters to org-scoped conversations; compose FAB restricts contact autocomplete to org members

**ComposerGuardrailsService (AppState-owned singleton):**
- Post-switch message banner: first 10 messages after a context switch show a dismissable banner — "Sending as [Org Name] — tap to switch context." Resets counter on next switch.
- Cross-context autocomplete filter: in an org context, contact autocomplete returns only org members; personal contacts return no results
- Clipboard origin guard: on paste, checks whether clipboard content was copied while a different context was active; if so, prompts — "This content was copied in your [Personal / Other Org] context. Are you sure you want to paste it here?" Fires once per clipboard content change per session.

**CredentialCardView:**
- Reads all visual treatment from VC display metadata (no hard-coded theme tokens)
- Displays: org logo, org display name, user's org-scoped display name, role badge
- Live revocation status indicator: performs StatusList2021 check against bit vector published to Constellation Identity Metagraph; shows valid/revoked state

**InvitationCoordinator:**
- `resolve(token:)` calls `GET /v1/comply/invitations/{token}/resolve` → determines Branch A (new user) vs Branch B (existing user)
- Branch A: org welcome screen (org branding) → standard registration → transparency screen → display name → `POST /v1/comply/invitations/{token}/accept`
- Branch B: org welcome screen → biometric auth → transparency screen → display name → `POST /v1/comply/invitations/{token}/accept`
- Both branches show transparency screen before acceptance

## Out of Scope

- Backend org lifecycle, BAA, member credentialing (WO-281–284)
- SSO configuration management (WO-285)
- Seat enforcement and Stripe billing (WO-286)
- Compliance-specific features (retention policies, litigation hold, eDiscovery — WO-250–269)

## Requirements

**Central architectural decision:** The user's `did:key` identifier is singular across all contexts. Context is runtime state, not a separate identity. A user with Personal + Mercy Health contexts has one DID. This is "Welcome back, Alex — you've been added to Mercy Health," not "Create a new Mercy Health account."

```swift
@MainActor @Observable
final class ContextCoordinator {
    private(set) var availableContexts: [AppContext] = [.personal]
    private(set) var current: AppContext = .personal
    private(set) var messagesSinceLastSwitch: Int = 0

    func loadAvailableContexts() async {
        // Read all valid org membership VCs from credential store
        // Verify each against StatusList2021 (not revoked)
        // Build OrganizationContext from VC display metadata (color, logo, role)
    }

    func switchTo(_ context: AppContext) async {
        current = context
        messagesSinceLastSwitch = 0
        // Haptic feedback + crossfade animation
        // Async: load context-scoped conversation list from local storage
    }
}

struct OrganizationContext: Hashable {
    let orgDID: String
    let displayName: String
    let primaryColor: Color
    let logoURL: URL?
    let role: OrgRole
    let orgScopedDisplayName: String
    let credentialID: String
    let revocationIndex: UInt32    // StatusList2021 bit position for live check
}
```

**Transparency Screen (required for both invitation branches):**
```
What [Org] WILL see in your work context:
  ✓ Messages you send to [Org] colleagues
  ✓ Your work display name
  ✓ Your work profile photo

What [Org] will NOT see:
  ✗ Your personal messages or conversations
  ✗ Who you message outside this organization
  ✗ Your personal trust tier or credential details
  ✗ Any data from your Personal context

Your personal ECHO identity remains entirely private.
```

**ComposerGuardrailsService — three layered checks:**
1. **Post-switch banner**: First 10 messages after context switch show dismissable banner. Resets on next switch.
2. **Cross-context autocomplete**: Org context → org members only. Personal contacts: no results.
3. **Clipboard origin guard**: Fires once per clipboard content change per session when content was copied in a different context.

## Blueprints

- Frontend — defines ContextCoordinator, ComposerGuardrailsService, CredentialCardView, InvitationCoordinator, OrgBrandingResolver, and the single-DID multi-context architecture
- ECHO Comply — Enterprise Compliance Messaging — defines org membership VC lifecycle, invitation flow branches, transparency screen content, and StatusList2021 revocation model

---

### WO-290: Build iOS Governance Tab with Trust-Tier-Weighted Voting

**Type:** Build

**Blueprint:** ECHO Token Economics and Founder Allocation, Frontend

## Summary

Implement the iOS Governance tab (Phase 4+ conditional) — `GovernanceManager` actor for voting power calculation and proposal submission, `GovernanceTab` SwiftUI view, `VotingPowerCard`, `ProposalCard`, and vote submission via `AtomicAction`. Governance weight is `StakedECHO × TrustTierMultiplier`, requiring Tier 2+ and staked ECHO to vote. The tab is hidden behind a feature flag and activates when the governance DAO goes live. The on-chain governance contract is WO-177; this WO is the iOS frontend.

## In Scope

- `GovernanceManager` actor: `fetchActiveProposals()`, `fetchExecutedProposals()`, `calculateVotingPower()` (reads staked ECHO from Stargazer SDK + trust tier from backend → computes effective weight), `voteOnProposal()` via `AtomicAction([.verifyStake, .submitVote])`
- `GovernanceTab` SwiftUI view: `VotingPowerCard` at top, `ForEach(activeProposals)` ProposalCard list, "Past Decisions" executed proposals section
- `VotingPowerCard`: staked amount, trust tier badge, governance multiplier, effective weight
- `ProposalCard`: title, proposal type badge, voting ends countdown, for/against progress bars, quorum indicator, user's vote status; "Vote For / Against / Abstain" buttons
- `ProposalHistoryRow` for executed proposals (result, final vote counts, execution status)
- Vote submission: biometric confirmation required; `AtomicAction` bundle ensures verifyStake + submitVote are atomic
- Tier 1 eligibility gate: Governance tab visible but vote buttons disabled with "Verify your identity to reach Tier 2 and unlock voting" CTA
- Feature flag: tab not rendered until `governanceDAOActive` remote config flag is enabled (Phase 4 launch)
- `GovernanceRepository` for caching proposals locally; poll for updates every 60 seconds when tab is visible

## Out of Scope

- On-chain governance voting contract and proposal submission (WO-177)
- Trust tier computation (WO-181, WO-49)
- Wallet staking and delegation (WO-127)
- Governance DAO legal entity formation (non-engineering)
- Proposal creation UI (governance admin only, not consumer-facing in Phase 4)

## Requirements

```swift
actor GovernanceManager {
    func calculateVotingPower() async throws -> VotingPower {
        // 1. Get total staked ECHO (including founder vesting locks)
        let locks = try await stargazer.getTokenLocks(token: .echo)
        let totalStaked = locks.reduce(0) { $0 + $1.amount }
        // 2. Get trust tier from backend
        let trustTier = try await backendAPI.getTrustTier()
        // 3. Apply trust tier multiplier
        let govMultiplier = trustTier.governanceMultiplier
        let effectiveWeight = totalStaked * Decimal(govMultiplier)
        return VotingPower(stakedAmount: totalStaked, trustTier: trustTier.level,
                           governanceMultiplier: govMultiplier, effectiveWeight: effectiveWeight)
    }

    func voteOnProposal(proposalId: String, vote: VoteChoice, votingPower: VotingPower) async throws {
        // Verify eligibility: must be Tier 2+ and have staked ECHO
        guard votingPower.trustTier >= 2, votingPower.stakedAmount > 0 else {
            throw GovernanceError.ineligibleToVote
        }
        try await stargazer.submitAtomicAction([
            .verifyStake(did: currentDID, minAmount: 0),
            .submitVote(proposalId: proposalId, vote: vote, weight: votingPower.effectiveWeight)
        ])
    }
}

struct VotingPower {
    let stakedAmount: Decimal
    let trustTier: Int       // 1–5
    let governanceMultiplier: Float  // 0.0 / 0.5 / 1.0 / 1.5 / 2.0
    // Note: DIFFERENT from reward multiplier (1.0x–3.0x)
    let effectiveWeight: Decimal     // StakedECHO × governanceMultiplier
}

enum VoteChoice: String { case `for`, against, abstain }
```

**Trust Tier Governance Multipliers:**

| Trust Tier | Multiplier | Governance Weight |
|---|---|---|
| Tier 1 (Unverified) | 0.0× | 0 — cannot vote |
| Tier 2 (Newcomer) | 0.5× | StakedECHO × 0.5 |
| Tier 3 (Member) | 1.0× | StakedECHO × 1.0 |
| Tier 4 (Verified) | 1.5× | StakedECHO × 1.5 |
| Tier 5 (Trusted) | 2.0× | StakedECHO × 2.0 |

**Proposal Types and Approval Thresholds:**

| Type | Threshold |
|---|---|
| Protocol upgrade | 67% supermajority |
| Treasury allocation | Simple majority (50%) |
| Validator admission | Simple majority |
| Emergency action | 75% supermajority |

**Quorum:** 20% of total staked tokens must participate for a vote to be valid.

## Blueprints

- Frontend — defines GovernanceManager, GovernanceTab, VotingPowerCard, ProposalCard, VotingPower struct, trust-tier governance multipliers, and feature flag gating
- ECHO Token Economics and Founder Allocation — defines governance weight formula, proposal types, quorum requirement, and approval thresholds

---

### WO-291: Implement AI Treasury Agent System (CFO, Stablecoin Manager, FeeTransaction Automation)

**Type:** Build

**Blueprint:** ECHO Token Economics and Founder Allocation, Production Launch, Infrastructure, and Deployment

## Summary

Implement the remaining Phase 5+ AI treasury agents not covered by WO-216 (Burn Agent + BTC Reserve): the CFO Agent that monitors DAG reserves and automates snapshot fee payments via the Tessellation v3 `FeeTransaction` primitive, the Stablecoin Manager Agent that converts surplus treasury ECHO to stablecoins via the Base bridge, and the Compliance and Reporting Agents that produce regulatory documentation and quarterly community financial reports. Prerequisites: governance DAO operational, 500K+ MAU.

## In Scope

**CFO Agent:**
- Continuously monitors DAG reserve balance (PagerDuty alert at <30 days runway, critical at <7 days)
- Automates `FeeTransaction` submissions from the treasury DAG reserves for Constellation metagraph snapshot fees — eliminates manual fee management
- Calculates projected DAG depletion rate based on rolling 30-day snapshot frequency
- Surfaces reserve health to treasury dashboard; triggers governance notification when annual token emission budget exceeds 90% / 99% thresholds

**Stablecoin Manager Agent:**
- Converts governance-approved portion of treasury ECHO surplus to USDC stablecoins (operational reserve target: 50M USDC equivalent) via Base bridge
- Executes periodic rebalancing toward governance-set stablecoin ratio
- Uses `AtomicAction` for bridge + conversion operations; all treasury movements publicly visible on DAG Explorer

**Compliance Agent:**
- Automated generation of regulatory documentation: on-chain transaction summaries, audit logs with cryptographic proof references, DAO governance decision records
- Formats output suitable for Foundation board review and external legal/audit use

**Reporting Agent:**
- Quarterly community financial reports: treasury inflows/outflows, token emission vs. budget, staking stats, reward distribution summary
- Annual token emission budget monitoring feed for governance notifications
- Treasury dashboard data feeds consumed by admin UI (links to WO-216's combined treasury operations dashboard)

**Integration:**
- All 4 agents integrate with the existing treasury operations dashboard alongside the Burn Agent and BTC Reserve Agent (WO-216)
- Shared `TreasuryAgentCoordinator` service manages agent lifecycle, health checks, and inter-agent communication

## Out of Scope

- AI Burn Agent and BTC Reserve Accumulation Agent (WO-216)
- On-chain governance voting mechanism (WO-177)
- PacaSwap DEX integration for token swaps (WO-215)
- DAO LLC legal entity formation (legal/operational)
- VIP subscription and organizational billing (WO-286)

## Requirements

**Phase 5 — Community Economy prerequisites (from Production Launch blueprint):**
- AI treasury agents deployed (CFO, Burn, BTC Reserve, Stablecoin Manager, Compliance, Reporting)
- FeeTransaction automation active — CFO agent manages DAG reserves
- Prerequisites: 500K+ MAU, stable governance DAO

**Monitoring thresholds enforced by CFO Agent:**

| Metric | Warning Threshold | Critical Threshold | Alert Channel |
|---|---|---|---|
| DAG snapshot fee reserves | < 30 days runway | < 7 days runway | Slack + PagerDuty |
| Annual token emission budget | > 90% of annual cap | > 99% of annual cap | Governance notification |

**FeeTransaction primitive (Tessellation v3):**

| Primitive | ECHO Operation |
|---|---|
| `FeeTransaction` | Automated metagraph snapshot fee payment from ECHO treasury DAG reserves. Eliminates manual snapshot fee management. |

**Treasury allocation reference (from token economics blueprint):**
```
Treasury (220M ECHO):
  - 80M → PacaSwap liquidity seeding (Phase 2, WO-215)
  - 50M → Operational reserve (stablecoins via bridge) ← Stablecoin Manager
  - 90M → Treasury multi-sig (3-of-5 founders → DAO governance)
```

**Stablecoin Manager target:** Maintain 50M USDC equivalent in operational reserve. Bridge via Base (Phase 3+). Rebalance quarterly or when reserve drops below 30M USDC equivalent.

## Blueprints

- Production Launch, Infrastructure, and Deployment — defines Phase 5 agent deployment requirements, DAG reserve monitoring thresholds, and FeeTransaction automation requirement
- ECHO Token Economics and Founder Allocation — defines treasury allocation, FeeTransaction primitive usage, stablecoin reserve targets, and Phase 5 deflationary mechanisms

---
