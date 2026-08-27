# Signal vs Echo Messaging — Parity Matrix

**Date:** 2026-07-26  
**Scope:** Signal-class messaging table-stakes + high-visibility UX, while keeping Echo’s `did:key` identity, Glacial UX, and metagraph/rewards backend.  
**Signal trees reviewed (vendored):** [`libsignal/`](../libsignal/), [`Signal-Server/`](../Signal-Server/), [`Signal-iOS/`](../Signal-iOS/)  
**Related:** [`GO_LIVE_SEPT_1_2026.md`](GO_LIVE_SEPT_1_2026.md) · [`TELEGRAM_ECHO_PARITY.md`](TELEGRAM_ECHO_PARITY.md) · [`COMPETITIVE_AUDIT_SIMPLEX_2026-06.md`](COMPETITIVE_AUDIT_SIMPLEX_2026-06.md) · [`COMPETITIVE_AUDIT_IMPLEMENTATION_PLAN.md`](COMPETITIVE_AUDIT_IMPLEMENTATION_PLAN.md)

> **Not a clone plan.** Echo does **not** adopt AGPL `libsignal` as a drop-in, phone-primary registration, Stories, MobileCoin, or Signal’s UI. Echo evolves its own Kinnami + Double Ratchet stack and Glacial design.

---

## 1. Architecture contrast

| Layer | Signal | Echo |
|-------|--------|------|
| Client | `Signal-iOS` → SignalServiceKit → LibSignalClient | `ios/Echo` → Glacial UI → Kinnami / DoubleRatchet |
| Protocol | PQXDH + Triple Ratchet (SPQR) + Sender Keys + sealed sender (`libsignal/rust/protocol`) | Kinnami + Double Ratchet (WO-314) + PQ bootstrap (WO-315) + sealed tokens (WO-317) |
| Gateway | Signal-Server (REST + gRPC + WS); offline queues Dynamo/Redis/FDB | Go gateway `:8000` WS relay + `/v3/*` |
| Adjacent | SVR, Storage Service, Groups CDN, TURN/SFU | Constellation metagraph (Identity / Data / Currency), rewards/wallet |
| Identity | Phone → ACI/PNI; usernames optional | **`did:key` + passkeys + `@username`** (phone-free by design) |

```text
Signal:  iOS ── LibSignalClient ── Signal-Server ── SVR/CDN/TURN
Echo:    iOS ── Kinnami/Ratchet ── Go gateway ── Metagraph (trust/rewards/anchor)
```

---

## 2. What Echo already matches (defend — don’t rebuild)

| Signal capability | Echo | Evidence |
|-------------------|------|----------|
| E2E 1:1 messaging | Have | `internal/crypto/kinnami.go`, `KinnamiEncryption.swift`, WS relay |
| Double Ratchet FS | Have | `internal/crypto/ratchet.go`, `DoubleRatchet.swift`, WO-314 |
| PQ session bootstrap | Partial | `internal/crypto/pqhybrid.go`, `PQHybridCrypto.swift`, WO-315 |
| Sealed / UD-ish send | Partial | `internal/services/messaging/sealed_sender.go`, WO-317 |
| Tor / SOCKS | Partial | `TransportProxySettings.swift`, WO-319 |
| Typing / receipts / reactions | Have (E2E gate) | `ConversationSignalService.swift`, `internal/api/ws.go` |
| Edit / delete / pin / disappearing | Have | message ops + `migrations/018_message_ops.sql` |
| Voice notes | Have | `VoiceNoteRecorder.swift`, WO-316 |
| Encrypted media | Have | `internal/services/media/service.go` |
| Groups + key distribute | Partial MVP | `GroupKeyManager.swift`, WS `group_key` |
| Link new device | Have | WO-288, `DeviceLinkAPIClient.swift` |
| Usernames + QR / invite | Have | WO-222, `echo://invite` |
| Calls signaling | Partial | `internal/api/call_handlers.go`, WebRTC live vs stub |
| Phone-free identity | **Ahead** | ADR-0001, `pkg/didkey/` |

---

## 3. Gap matrix — implement (waves)

### Wave S1 — Multi-device history + cloud backup (P1)

| Gap | Signal reference | Echo today | Done when |
|-----|------------------|------------|-----------|
| Full history sync | Link’n’Sync + sync messages | `device_sync.go`, `DeviceHistorySyncService` (`history` / `history_chunk`), `/v3/sync/ack` + PG trim | Two-device seed/pull/revoke E2E on device; config sync follow-on |
| Secure cloud backup | Backups + SVR / AEP | `backup_handlers.go` + dedicated `message_backup_blob` (027); restore diagnostics on recovery | Round-trip across restart with `DATABASE_URL`; BackupView UX polish |
| Config sync | Storage Service | Partial | Privacy prefs / blocked / pins sync messages (follow-on) |

**Key paths:** `migrations/019_device_sync.sql`, `internal/api/sync_handlers.go`, `internal/api/backup_handlers.go`, `pkg/passport/sync.go`, `main.go` (SyncService wiring), `BackupView.swift`, `RecoveryService.swift`.

### Wave S2 — Crypto / metadata credibility (P0)

| Gap | Signal / libsignal | Echo action |
|-----|-------------------|-------------|
| Sealed-sender strength | UD + delivery certs (`Messages/UD/`, `sealed_sender.rs`) | Minimize `message_queue` graph retention; default sealed for trusted contacts; conversation delivery aliases |
| Group Sender Keys | `SenderKeyManager`, `group_cipher.rs` | Group ratchet + SKDM-style distribute; rekey on membership change |
| Safety numbers | `SignalUI/SafetyNumbers/` | Echo fingerprint compare + key-change banner (Glacial) |
| Per-step PQ | Triple Ratchet / SPQR | Wave S5 — evolve Echo ratchet (do not vendor AGPL libsignal) |

**Key paths:** `sealed_sender.go`, `migrations/002_messaging.sql`, group key packages, new safety-number UI under chat settings.

### Wave S3 — Groups & calls (P2)

| Gap | Signal | Echo action |
|-----|--------|-------------|
| Group Phase 3 signals | Full group typing/receipts/reactions | `GroupChatViewModel` + WS fanout |
| Group admin / invite links | GroupsV2 invite links | Invite links, roles, remove+rekey E2E |
| Calls polish | RingRTC + CallKit + TURN `/v2/calling/relays` | Require WebRTC in release; CallKit; stable TURN mint |
| Screen share / call links | Call Links | Post-S3 unless marketed |

### Wave S4 — High-visibility UX (P3)

| Feature | Signal locus | Echo action |
|---------|--------------|-------------|
| Polls E2E | `Interactions/Polls/` | Finish WO-23 (UI exists) |
| View-once media | `Attachments/V2/ViewOnce/` | Flag + burn-after-read |
| Message requests | `MessageRequest*` | First-contact gate for non-contacts |
| GIF lite | `GifPicker/` | GIF search (no Signal sticker CDN) |
| Chat wallpaper | `SignalUI/Wallpapers/` | Optional per-thread Echo token background |
| Screenshot notify | Settings + notify | Wire setting → peer WS/notify |
| Rich NSE | `SignalNSE/` | Notification Service Extension decrypt preview |

### Wave S5 — Later

- Per-ratchet-step PQ hybrid  
- 2-hop private routing / federation principles (WO-320)  
- Key Transparency optional (Echo metagraph anchoring is intentional analog)

---

## 4. Explicitly out of thesis

| Signal feature | Why not |
|----------------|---------|
| Phone-primary registration | Echo is `did:key` / passkey first |
| AGPL libsignal drop-in | License; unsupported outside Signal; replaces shipped crypto |
| Stories | Not Echo launch pillar |
| MobileCoin / donations | Echo earn/stake wallet instead |
| Full zkgroup stack | Revisit only if group privacy requirements demand |
| Identity-free routing (SimpleX) | Breaks trust tiers / anti-scam / payments |

---

## 5. Server capability analogs (not wire-compatible)

| Signal-Server | Echo counterpart |
|---------------|------------------|
| Account + devices | Identity + `/v1/login/link-device/*` + device list |
| Message fanout + offline queue | WS relay + message queue |
| Sealed / anonymous send | Sealed-sender tokens (harden in S2) |
| Prekeys | Session / ratchet bootstrap |
| Attachment CDN forms | Media service upload credentials |
| TURN credentials | Call ICE / TURN mint (S3) |
| SVR / encrypted backups | BIP-39-wrapped `/v3/backup/*` (S1) |

---

## 6. Delivery waves & Software Factory

Tickets: [`SIGNAL_PARITY_WORK_ORDERS.md`](SIGNAL_PARITY_WORK_ORDERS.md) — epic **WO-330** (child of WO-321), waves **WO-333** (S1), **WO-331** (S2), **WO-334** (S3), **WO-332** (S4).

| Wave | WO | Window | Focus | Implementation (2026-07-26) |
|------|-----|--------|-------|------------------------------|
| Now | WO-321… | → Sept 1 2026 | E2E + App Store P0 only | Go-live children |
| **S1** | **WO-333** | Sept–Oct | History sync + cloud backup | Isolated `message_backup_blob` (027); history chunking; `/v3/sync/ack` + PG trim; restore diagnostics after phrase recovery |
| **S2** | **WO-331** | Oct–Nov | Sealed-sender + Sender Keys + safety numbers | Redact `sender_did`; sender-key group path; safety-number UX |
| **S3** | **WO-334** | Nov–Dec | Group signals + calls polish | Group typing/reactions UI; `/v3/calls/relays` |
| **S4** | **WO-332** | Parallel / Q1 | View-once, requests, polls, GIF, NSE, wallpaper | Wired banner/GIF/view-once/wallpaper; NSE skeleton (Xcode target TBD) |
| **S5** | — | Later | Per-step PQ + federation | WO-320 backlog |

**Assignee:** Chad Cromwell  
**Sept 1:** Do **not** block soft launch on S1–S4; market only what ships ([`GO_LIVE_SEPT_1_2026.md`](GO_LIVE_SEPT_1_2026.md) §8.4).

---

## 7. Honest “comparable?” answer

- **Daily E2E chat:** Close once Double Ratchet + Phase 3 E2E are proven; still behind on per-step PQ, true sealed-sender metadata minimization, and Sender Keys.  
- **Device story:** Link-device exists; **Signal users expect restore/sync** — largest product gap (Wave S1).  
- **Identity/UX:** Echo stays phone-free + trust-tier differentiated — do not chase Signal’s phone model.  
- **Sept 1:** Stay on go-live scope; this document is the **post-launch** Signal-parity program.

---

## 8. Sources

- `libsignal/README.md`, `libsignal/rust/protocol/`, `libsignal/swift/`  
- `Signal-Server/README.md`, `service/.../controllers/`, `service/src/main/proto/org/signal/chat/`  
- `Signal-iOS/SignalServiceKit/`, `Signal/ConversationView/`, `Podfile` (LibSignalClient)  
- Echo: `internal/crypto/`, `internal/api/ws.go`, `ios/Echo/Sources/Services/`, competitive audits above  
