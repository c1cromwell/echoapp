# ECHO Messaging — 25% → 100% Completion Plan

> Status: **Active plan — 2026-06-12.** The "what to implement next" roadmap for the messaging
> product. Derived from `docs/CROSS_PRODUCT_GAP_REVIEW.md`,
> `docs/COMPETITIVE_AUDIT_IMPLEMENTATION_PLAN.md`, `docs/PHASE2_GAP_AUDIT.md`, and the phase-3 work
> orders. Every wave reuses existing WOs — this sequences and gates them; it does not invent new scope.

## Where we are (~25%)

Solid foundation, thin surface:
- ✅ E2EE (Kinnami: X25519 + ChaCha20-Poly1305) — `internal/crypto/kinnami.go`, `ios/.../KinnamiEncryption.swift`
- ✅ 1:1 relay + offline queue + APNs — `internal/services/relay/relay.go`
- ✅ PSI contact discovery backend — `internal/services/contacts/` (iOS client in flight, WO-221)
- ✅ did:key identity + trust tiers (~80%)
- ⚠️ Reactions / typing / receipts — backend fan-out incomplete (WO-192/10)
- ⚠️ Edit/delete/forward/pin/reply/disappearing — UI sketched, backend not wired
- ❌ Group key distribution (no group UI), calls, media relay, search, multi-device sync

## Definition of "100%"

Two bars, both required:
1. **Competitor parity** (Signal/WhatsApp/Telegram table stakes): reliable 1:1 + group chat with
   reactions/typing/receipts/edit/delete/reply/forward/pin; voice+video calls; media & voice notes;
   search; multi-device with history; encrypted backup/restore; disappearing messages.
2. **Echo differentiators intact**: E2EE-by-design, did:key identity, metagraph integrity anchoring,
   trust tiers, **zero-PII / no server-readable content**, on-device privacy AI (not server plaintext).

"100%" = both bars green end-to-end per `docs/E2E_LAUNCH_AND_TESTING.md`, on real devices, with the
T0–T7 Semgrep invariant green.

---

## Waves (critical path top to bottom)

### M0 — Close in-flight signals (the current 1:1 core)  · **gates everything**
Finish what's already started so the 1:1 experience is reliable before adding surface area.
- **WOs:** WO-192 (typing + read receipts), WO-10 (reactions real-time sync), WO-48 (offline
  support/sync), WO-57 (push + analytics).
- **Work:** complete relay WebSocket fan-out for `typing`/`receipt`/`reaction` events
  (`internal/services/relay/relay.go`, `internal/api/ws.go`); wire iOS `ChatDetailViewModel` signal
  handlers; durable read-state.
- **Gate:** two-client typing + read receipts + reactions pass `E2E_LAUNCH_AND_TESTING.md` §6.4.
- **Est:** 2–4 wks (mostly iOS + E2E).
- **Progress (2026-06-12):**
  - ✅ **Backend signal fan-out done + tested.** `SignalPublisher` + `Hub.PublishSignal`
    (`internal/api/ws.go`); durable read receipts (`MarkRead`/`GetMessageMeta` in
    `internal/database/{database,postgres}.go` + migration `017_message_read_receipts.sql`);
    `handleMessageReceipt` now honors `receiptType` and serves `GET /v3/messages/{id}/status` for
    offline reconnect sync (WO-48); server-authoritative reaction fan-out in `handleMessageReact`
    (WO-10). Hub wired via `V3Handlers.Signals` (`main.go`). 6 new tests in
    `internal/api/message_signals_test.go`, all green. Typing relay was already complete.
  - ✅ **WO-57 offline push done + tested.** `OfflineNotifier` interface + `Hub.SetOfflineNotifier`
    (`internal/api/ws.go`): a directed message to an offline recipient now fires a content-blind push
    (conversation id + sender only). Reaction fan-out (`handleMessageReact`) pushes when the peer is
    offline. Adapter over `notification.Service` in `internal/api/offline_notifier.go`, wired in
    `main.go`. Ephemeral signals never push. 4 new tests in `internal/api/offline_push_test.go`
    (green, stable under `-race`).
  - ✅ **iOS durable-receipt wiring done (pending Xcode build).** New `MessageReceiptsAPI`
    (`ios/.../Services/MessageReceiptsAPI.swift`: `markRead`/`markDelivered`/`status`) registered in
    the pbxproj + DI container. `ChatDetailViewModel` now persists read receipts via REST
    (`persistReadReceipts`) alongside the live WS signal, and `reconcileReceiptsOnOpen()` pulls
    `GET .../status` for own sent messages on open/reconnect (WO-48). Wired in `MessagingScreens`.
    2 new unit tests in `Phase3Tests/ChatDetailViewModelTests.swift` (+ `MockMessageReceiptsAPIClient`).
    The existing WS typing/receipt/reaction handling was already complete and correct. **All edited
    Swift parses clean headlessly; founder confirms the full build + Phase3Tests in Xcode.**
  - ⏳ **Remaining:** two-device E2E per §6.4 (the real gate, run on devices).

### M1 — Message operations (edit / delete / reply / forward / pin / disappearing)
The per-message actions whose UI exists but backend doesn't.
- **WOs:** WO-25 (edit w/ immutable history), WO-84 (synchronized delete), WO-59 (pin/forward/reply),
  disappearing-message backend (timer enforcement on relay + client).
- **Work:** relay verbs + commitment updates for edit/delete; pinned-message store (max 5);
  reply/forward references; disappearing-timer enforcement server + client expiry.
- **Gate:** edit/delete propagate to both clients; disappearing messages vanish on schedule on both
  ends and in offline-queue replay; immutable edit history verifiable.
- **Est:** 3–5 wks. **Depends:** M0.
- **Progress (2026-06-12) — backend done + tested (hybrid model):**
  - ✅ **Edit (WO-25), hybrid.** Non-retained conversations = relay-only edit fan-out (clients hold
    history); retained (Comply / litigation hold) = server-side **immutable versions**
    (`message_edits`) + `GET /v3/messages/{id}/history` for eDiscovery. `POST /v3/messages/{id}/edit`.
  - ✅ **Delete (WO-84).** `POST /v3/messages/{id}/delete` tombstones + fans out `delete`; under
    retention, edit history is preserved (litigation hold), else purged.
  - ✅ **Pins (WO-59).** `POST /v3/messages/{id}/pin|unpin` (max 5/conv, 409 over limit),
    `GET /v3/conversations/{id}/pins`.
  - ✅ **Disappearing.** Per-conversation TTL `GET|POST /v3/conversations/{id}/disappearing` + synced
    `disappearing_config` signal; relay `ExpiresAt`/`PurgeExpired` already enforce expiry.
  - ✅ **Retention gate.** `POST /v3/conversations/{id}/retention` (Comply sets it). New
    `MessageOpsStore` across the `DB` interface + MemoryDB + PostgresDB + migration `018_message_ops.sql`.
    6 new tests in `internal/api/message_ops_test.go` (green under `-race`); WS payloads
    `EditSignal`/`DeleteSignal`/`PinSignal`/`DisappearingSignal` in `ws.go`.
  - ✅ **iOS M1 wiring done (pending Xcode build).** New `MessageOpsAPI` client
    (`ios/.../Services/MessageOpsAPI.swift`: edit/delete/pin/unpin/disappearing) registered in
    pbxproj + DI. `ChatDetailViewModel` now: `applyEdit` encrypts + persists via REST; `deleteMessage`,
    `togglePin` (max 5, optimistic w/ rollback), `setDisappearing`; and consumes the new inbound
    signals (`edit`/`delete`/`pin`/`disappearing_config`) added to `ConversationSignal` codec +
    service. `MessagingScreens` actions call the VM. 8 new `Phase3Tests` (+ `MockMessageOpsAPIClient`);
    all edited Swift parses clean headlessly.
  - ⏳ **Remaining:** reply/forward are client-side payload metadata; **Xcode full build + Phase3Tests**
    (founder gate); two-device E2E.

### M2 — Groups (key distribution + service + UI)  · **largest single gap**
Groups are modeled but have no key distribution and no UI.
- **WOs:** WO-207 (E2EE canonical incl. group key distribution), groups service
  (`internal/services/groups/`), new group iOS UI.
- **Work (backend):** implement `GroupKeyManager.distributeGroupKey()` (sender-key / per-member
  sealed AES-256-GCM), member add/remove rekey, `/v3/groups/*` handlers, relay group fan-out (NATS).
- **Work (iOS):** GroupCreate / GroupDetail / GroupChat views; roles/permissions from `groups/models.go`.
- **Gate:** create group → all members decrypt; remove member → rekey, removed device cannot decrypt
  new messages; group fan-out delivers offline members on reconnect.
- **Est:** 5–8 wks. **Depends:** M0 (relay/signals). Highest risk — staff first after M0.
- **Progress (2026-05-29) — M2 started (backend + iOS scaffolding, pending Xcode E2E):**
  - ✅ **Group key distribute REST + WS.** `POST /v3/groups/key/distribute` fans opaque per-member
    `group_key` signals via `SignalPublisher` (`internal/api/group_key_handlers.go`); `GroupKeySignal`
    in `ws.go`. Member add/remove return `requires_rekey: true`. Tests in
    `internal/api/group_key_handlers_test.go`.
  - ✅ **Group text fan-out.** WS `type:text` with `conversation_id: group:{id}` and empty `to` routes
    to all group members (`routeGroupText` in `ws.go`); `GroupService.GroupMemberDIDs` wired in `main.go`.
    Tests in `internal/api/ws_group_fanout_test.go`.
  - ✅ **iOS key distribution.** `GroupKeyManager.distributeGroupKey` / `rotateAndDistribute`;
    `GroupKeyDistributionService`; `GroupsAPIClient`; `group_key` codec in `ConversationSignal.swift`.
  - ✅ **iOS group UI (minimal).** `GroupCreateView` + member picker, `GroupChatView`, navigation from
    `MessagesTabView` / `ChatDestinationView`; DI in `Container.swift`; pbxproj entries.
  - ✅ **Group ciphertext on wire.** `TextMessagePayload.group_ciphertext` + `sendGroupTextMessage`;
    `GroupChatViewModel` encrypt/decrypt round-trip.
  - ✅ **GroupDetailView + rekey.** Member list, add/remove (admin); `rekeyForMembers` on
    `requires_rekey`; `GroupDetailSheet` from group chat toolbar.
  - ✅ **Offline WS replay.** Hub `wsOfflineQueue` enqueues undelivered group text + directed signals;
    `flushOffline` on reconnect. Tests in `ws_offline_queue_test.go`.
  - ✅ **E2E checklist.** `docs/E2E_LAUNCH_AND_TESTING.md` §6.9 (two-device manual gate).
  - ⏳ **Remaining:** Xcode full build + two-device §6.9 sign-off on real devices/simulators.

### M3 — Multi-device + history sync + backup  · **Signal parity**
Phone-free + restore. Already sequenced in `COMPETITIVE_AUDIT_IMPLEMENTATION_PLAN.md` Wave 1.
- **WOs:** WO-288 (new-device QR + recovery login), WO-273 (controller on metagraph), WO-73
  (cross-device encrypted search index sync), WO-CA3 (did:key-scoped message history sync),
  WO-64 (encrypted backup), WO-CA2 (consumer cloud backup).
- **Work:** device controller doc + per-device keys; sync namespace scoped by controller DID;
  ciphertext-only sync bundles (reuse relay/offline-queue patterns + sync cursor / device epoch);
  user-phrase-encrypted backup export/import + optional cloud tier.
- **Gate:** add iPad after iPhone → full history appears after unlock; revoke device stops new sync;
  fresh-install restore from cloud backup + 24 words; server operators cannot decrypt any blob.
- **Est:** 8–14 wks (CA3 is the long pole). **Depends:** M0; benefits from M2 for group history.
- **Decisions (2026-06-13):** first slice = WO-CA3 history-sync **server foundation**; sync model =
  **per-device addressed streams** (revoke closes/purges the stream; pairwise-ECDH-wrapped ciphertext).
  Device-link foundation (WO-288/273: `/identity/devices*`, on-chain `DeviceKeyRecord`, iOS
  `DeviceLinkAPIClient`) already exists.
- **Progress (2026-06-13) — M3a server foundation done + tested:**
  - ✅ **Content-blind `DeviceSyncStore` (WO-CA3).** Per-`(controllerDID, targetDeviceID)` append-only
    streams with monotonic seq; `AppendSyncEntry`/`PullSyncEntries`/`SyncHead`/`RevokeDeviceStream`
    across the `DB` interface + MemoryDB + PostgresDB + migration `019_device_sync.sql`. Server stores
    opaque ciphertext only.
  - ✅ **Endpoints** (`internal/api/sync_handlers.go`, controller-DID scoped from the access token):
    `POST /v3/sync/push`, `GET /v3/sync/pull?device_id=&after=&limit=`, `GET /v3/sync/head`,
    `POST /v3/sync/revoke`. Revoke closes + purges the stream and rejects further push/pull (403).
  - ✅ 5 tests (`internal/api/sync_handlers_test.go`): cursor paging, per-device isolation,
    cross-account scoping, revoke-stops-sync, head. Green under `-race`.
  - ⏳ **Remaining (M3b/M3c):** iOS client (primary seeds history to a newly-linked device; new device
    pulls + decrypts on unlock; device-key wrapping); encrypted backup (WO-64/CA2) reusing
    `pkg/storage/encblob`; WO-73 search-index sync deferred to M6.
  - ✅ **M3b.1 API client + device id (2026-05-29).** `DeviceSyncAPIClient` (`/v3/sync/*`),
    `DeviceIdentityStore` (keychain `dev-*` id), `SyncCursorStore`; DI wired.
  - ✅ **M3b.2 Bundle + crypto (2026-05-29).** `HistorySyncBundle` + builder/merger;
    `DeviceSyncCrypto` (P-256 ECDH wrap/unwrap); `ConversationThreadStore` export/merge helpers;
    `Phase3Tests/DeviceSyncTests.swift`.
  - ✅ **M3b.3 Orchestration + device-link wiring (2026-05-29).** `DeviceHistorySyncService`
    (seed/pull/revoke); primary polls after QR link and pushes wrapped bundle; linked device
    assigns sync stream id + pulls on login/Messages tab; E2E §6.10.
  - ✅ **M3b.4 Revoke on device removal (2026-05-29).** `DeviceManagementViewModel` maps identity
    pubkeys by device label and calls `POST /v3/sync/revoke` when removing a linked device.
  - ✅ **M3c Encrypted backup (2026-05-29).** `BackupCrypto` (phrase + HKDF + AES-GCM),
    `MessageBackupService` (local `.enc` + cloud `/v3/backup/*`), `BackupView` wiring;
    server relay via `encblob`; `Phase3Tests/BackupCryptoTests.swift`; E2E §6.11.
  - ✅ **M3 finish (2026-05-29).** `BackupScheduler` (daily/weekly/manual, wifi-only, foreground);
    `BackupSessionKeyStore` (keychain session key after opt-in); local restore + auto-backup in
    `BackupView`; `revokeAllRemoteSyncStreams()` on logout-all; `MessagesTabView` runs scheduler.
  - ⏳ **M3 sign-off only:** Xcode two-device §6.10–6.11; WO-73 search-index sync deferred M6.

### M4 — Calls (voice + video, WebRTC)
UI skeleton exists; signaling foundation landed (real WebRTC + CallKit still pending).
- **WOs:** WO-196 (call history + notifications) + WebRTC signaling (new).
- **Work:** call signaling over the WebSocket hub (`internal/api/ws.go`): offer/answer/ICE; TURN/STUN
  config; iOS `CallViewModel`/`CallState` real peer-connection wiring (`react-native-webrtc` /
  CallKit); 1:1 first, then group calls.
- **Gate:** 1:1 voice + video connect on real devices across NAT; call history + missed-call badges;
  E2EE media (DTLS-SRTP) verified.
- **Est:** 5–8 wks. **Depends:** M0. Independent of M2/M3 — parallelizable with separate staffing.
- **Progress (2026-05-29) — M4a foundation:**
  - ✅ **Backend:** `call_signal` WS routing via `deliverOrQueue`; `GET /v3/calls/ice-servers`;
    `ws_call_signal_test.go` + ICE handler tests.
  - ✅ **iOS:** `CallSignal`/`CallSignalCodec`, `CallSignalingService`, `CallICEAPIClient`,
    `WebRTCCallSession` stub, `CallHistoryStore` (WO-196), `CallViewModel` rewired;
    `ConversationSignalService` decodes `call_signal`; `ContactDetailView` voice/video sheets;
    `Phase3Tests/CallSignalCodecTests.swift`; E2E §6.12.
  - ⏳ **M4 remaining:** WebRTC.framework peer connection, CallKit, TURN credentials, group calls,
    missed-call push specialization; Xcode §6.12 sign-off.

### M5 — Media & files (relay pipeline + voice notes)
Encryption service exists; transport pipeline foundation in progress.
- **WOs:** WO-237 (offline-queue IPFS overflow), WO-194 (voice notes), media message type.
- **Work:** encrypted media upload/download via relay + IPFS/Storj overflow for large blobs
  (reuse `pkg/storage/encblob/` from Passport); image/video/file message types + thumbnails;
  `AVAudioRecorder` + Opus voice notes with waveform/playback.
- **Gate:** send/receive image, video, file, and voice note 1:1 and in a group; large file overflows
  to IPFS; media is client-encrypted (server stores ciphertext + CID only).
- **Est:** 4–6 wks. **Depends:** M2 (for group media), M3 (encblob sync overlap).
- **Progress (2026-05-29) — M5a foundation:**
  - ✅ **Backend:** `GET /v3/media/{id}/chunks/{index}` chunk download; `RetrieveChunk` in media service;
    WO-237 overflow queue pins to `encblob` at depth ≥1000 + `overflow_manifest` on reconnect;
    `media_handlers_test.go`.
  - ✅ **iOS:** `MediaAPIClient`, `MediaMessageCrypto`, `MediaMessageService`, `MediaAttachmentRef`
    in `TextMessagePayload`; `ChatDetailViewModel.sendMedia` / `sendVoiceNote`; `VoiceNoteRecorder`
    (AAC); DI + `Phase3Tests/MediaMessageWireTests.swift`.
  - ⏳ **M5 remaining:** PhotosPicker/attachment UI, thumbnails, group media path, overflow manifest
    fetch on iOS, Opus codec, waveform playback, Xcode §6.13 sign-off.

### M6 — Search & advanced messaging
Designed, unbuilt.
- **WOs:** WO-3 (local encrypted index), WO-16 (keyword search), WO-29 (advanced filters),
  WO-197 (conversation search), WO-198/54 (archive + folders), WO-23 (polls), screenshot alerts.
- **Work:** device-local encrypted inverted index + tokenizer (HKDF-derived index key); fuzzy/boolean
  search + ranking; archive folders; zero-knowledge polls; screenshot-notification WS event.
- **Gate:** search returns ranked local results with no plaintext index on server; cross-device index
  sync (via M3/WO-73); poll create/vote/close E2E.
- **Est:** 5–8 wks. **Depends:** M0 (data), M3/WO-73 (index sync).

### M7 — Privacy-preserving on-device AI  · **differentiator, last**
- **WOs:** WO-CA1 (private AI), precursors WO-2.5/2.6 (on-device translation + thread summaries).
- **Work:** on-device model runtime (Core ML / Apple Intelligence) for smart replies, summaries,
  translation; per-message consent; AI-invocation audit log; **no server plaintext** ever.
- **Gate:** summarize/translate/suggest fully on-device by default; zero default cloud plaintext;
  opt-in rate tracked; any server-assist gated behind ZK/confidential-compute.
- **Est:** 12+ wks after a bot-framework MVP. **Depends:** M0–M6 stable.

---

## Critical path & parallelization

```
M0 (gates all)
 ├─ M1  message ops
 ├─ M2  groups  ───────────────┐  (highest risk — staff first after M0)
 ├─ M4  calls   (independent)  │
 └─ M3  multi-device ──────────┼─ M5 media ── M6 search ── M7 AI
                               │
                  (M3 history benefits from M2 groups)
```

- **Single track:** M0 → M2 → M3 → M5 → M6 → M7, with M1 folded into M0/M2 and M4 slotted wherever.
- **Three tracks (staff permitting):** (A) M0→M1→M2 groups [iOS+relay], (B) M3 multi-device+backup
  [backend+iOS], (C) M4 calls [WebRTC]. Converge before M5/M6.

## Rough effort to 100%

| Wave | Est | Cumulative (single track) |
|------|-----|---------------------------|
| M0 | 2–4 wks | ~1 mo |
| M1 | 3–5 wks | ~2 mo |
| M2 groups | 5–8 wks | ~4 mo |
| M3 multi-device + backup | 8–14 wks | ~7 mo |
| M4 calls | 5–8 wks | (parallel) |
| M5 media | 4–6 wks | ~8.5 mo |
| M6 search | 5–8 wks | ~10 mo |
| M7 AI | 12+ wks | ~13+ mo |

≈ **10 months to parity (M0–M6)** single-track; ~5–6 months with three parallel tracks. M7 (AI) is
the differentiator tail beyond parity.

## Invariant guardrails (every wave)

- **No server-readable content.** Relay stays content-blind; media/search indexes/backups are
  client-encrypted (server holds ciphertext + CID only). T0–T7 Semgrep stays green.
- **Group rekey on membership change** — removed members must not decrypt subsequent messages.
- **AI on-device by default** — no Grok-style server plaintext reads.

## Start here (next two sprints)

1. **M0**: finish WO-192/10 relay fan-out + iOS signal handlers → pass E2E §6.4.
2. In parallel, spike **M2** `GroupKeyManager.distributeGroupKey()` (the long pole) so group work is
   unblocked the moment M0 lands.

## Related

- `docs/COMPETITIVE_AUDIT_IMPLEMENTATION_PLAN.md` — overlapping Wave 0–4 sequencing + dependency graph
- `docs/PHASE2_GAP_AUDIT.md` — current truth table
- `docs/phase-3-work-orders.md` — messaging WO specs
- `docs/E2E_LAUNCH_AND_TESTING.md` — E2E gates per wave
