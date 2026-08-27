# Telegram vs Echo Messaging — Parity Matrix

**Date:** 2026-08-27  
**Scope:** Telegram-class daily UX (Saved Messages, drafts, forward, search, folders, scheduled send, durable channels) while keeping Echo’s `did:key` identity, Glacial UX, and content-blind gateway.  
**Related:** [`SIGNAL_ECHO_PARITY.md`](SIGNAL_ECHO_PARITY.md) · [`GO_LIVE_SEPT_1_2026.md`](GO_LIVE_SEPT_1_2026.md) · [`E2E_TESTING.md`](E2E_TESTING.md)

> **Not a clone plan.** Echo does **not** ship Telegram Stories, Stars, Nearby people, Mini Apps, or Secret Chat as a product. Multi-device history is the same program as Signal Wave S1 (WO-333).

---

## 1. Architecture contrast

| Layer | Telegram | Echo |
|-------|----------|------|
| Client | Telegram-iOS / TDLib | `ios/Echo` → Glacial UI → Kinnami / Double Ratchet |
| Transport | MTProto + datacenter cloud | Go gateway `:8000` WS relay + `/v3/*` (content-blind blobs) |
| Identity | Phone + username | **`did:key` + passkeys + `@username`** |
| Cloud copy | Server holds plaintext for sync | Phrase-encrypted `/v3/backup/*` + device `/v3/sync/*` (S1) |
| Channels | Broadcast + admin | `/v3/broadcasts/*` + `ChannelAdminView` |

```text
Telegram:  iOS ── MTProto ── Cloud DC (server-readable)
Echo:      iOS ── Kinnami/Ratchet ── Go gateway ── Metagraph (trust/rewards)
```

---

## 2. Code validation (2026-08-27)

Status is **what is in this repo**, not what TestFlight users have tapped. “Have” means sources + (usually) `EchoApp.xcodeproj` membership. “Xcode” means a human still has to wire Info.plist, a new target, entitlements, or a live two-device pass.

| Telegram capability | Echo | Evidence |
|---------------------|------|----------|
| 1:1 E2E chat | Have | Kinnami + WS relay (see Signal matrix) |
| Saved Messages | Have | `SavedMessagesStore.swift`, hub pin in `MessagesHubView` |
| Drafts | Have | `ComposerDraftStore` load/save in `ChatView` |
| Forward / multi-forward | Have | `ForwardMessageSheet` (single + multi-select) |
| Quote replies | Have | `replyToMessageId` / `ChatMessageEnvelope` |
| Link previews | Have | `MessageLinkPreview.swift`, `ChatInlineLinkPreview` |
| Global search | Have | `GlobalSearchSheet`, `KeywordSearchEngine`, `LocalMessageIndexer` |
| Chat folders | Have (sync API) | `/v3/chat-folders`, `ChatFolders.swift`, `ChatFoldersView` |
| Scheduled send | Partial | iOS `ScheduledMessageStore` + BGTask; **HTTP** `/v3/messages/schedule` (in-memory, WO-338) |
| Silent send | Partial | Gateway `ScheduleSilent` + WS `silent`; iOS composer not a first-class silent toggle |
| Broadcast channels | Partial | `/v3/broadcasts/*`, `ChannelsListView`, `ChannelAdminView` |
| Bots | Partial | `/v3/bots/` foundation (WO-11) |
| Typing / receipts / reactions | Have | Phase 3 — `ChatDetailViewModel` in `ChatView` |
| Voice notes | Have | `VoiceNoteRecorder` |
| Stickers / GIF | Partial | GIF search `GifSearchService`; no sticker packs |
| Voice rooms / live streams | Skip / later | Not a Sept 1 pillar |
| Secret chats | Skip | Echo E2E is default; do not clone Telegram Secret Chat UX |
| Stories / Stars / Nearby | Skip | Out of thesis |

SPM unit coverage: `ios/Echo/Phase3Tests/TelegramP1ParityTests.swift`.

---

## 3. Gap matrix — implement (waves)

### Wave T0 — Durable channels + admin (P0)

| Gap | Telegram | Echo today | Done when |
|-----|----------|------------|-----------|
| Durable broadcast | Channels | `/v3/broadcasts/*` (service + HTTP) | Two-device create/subscribe/post; Postgres soak before marketing “durable” |
| Admin | Owner/admin roles | `ChannelAdminView` + member/role/join-request routes | Approve/deny + role change in Xcode UI |

**WO-336.** Shared with FEATURE [Broadcast Channels](https://factory.8090.ai).

### Wave T1 — Daily UX (P1)

| Gap | Telegram | Echo today | Done when |
|-----|----------|------------|-----------|
| Saved Messages | Cloud “Saved” | Self-DID conversation | Two-device: save + reopen after kill |
| Drafts | Per-chat draft | `ComposerDraftStore` | Leave chat, reopen, text still there |
| Forward | Multi-chat forward | `ForwardMessageSheet` | Forward to Saved + another DM |
| Global search | Search chats + messages | `GlobalSearchSheet` | Live message indexed and found |
| Quotes | Reply quote | Envelope `replyToMessageId` | Peer sees quote bubble |
| Link previews | In-bubble preview | `MessageLinkPreview` | URL in composer shows card |

**WO-337** (`in_progress` — logic landed; two-device E2E + hub wiring remaining).

### Wave T2 — Folders, schedule, bots (P2)

| Gap | Telegram | Echo today | Done when |
|-----|----------|------------|-----------|
| Folders | Client + cloud | GET/PUT `/v3/chat-folders` | Two-device folder round-trip |
| Scheduled | Server queue | Local BGTask + `/v3/messages/schedule` | Second device lists pending; Postgres later |
| Silent | Send without notify | Gateway flags | Composer silent toggle + WS `silent` |
| Bots | Bot API | `/v3/bots/` | Token send + webhook smoke |

**WO-338.**

### Wave T3 — Later

Stickers (Echo packs, not Signal CDN), in-chat translate, chat import. **WO-339** backlog.

Multi-device **history restore** is **not** a Telegram-only ticket — it is Signal **S1 / WO-333**.

---

## 4. Explicitly out of thesis

| Telegram feature | Why not |
|------------------|---------|
| Secret Chat clone | Echo DMs are already E2E; a second “secret” mode confuses trust UX |
| Stories | Not a launch pillar (same as Signal Stories) |
| Stars / Mini Apps / payments-as-Telegram | Echo wallet/stake instead |
| Nearby people | Conflicts with privacy + anti-scam thesis |
| Phone-primary MTProto identity | ADR-0001 `did:key` |

---

## 5. Delivery waves & Software Factory

Tickets: [`TELEGRAM_PARITY_WORK_ORDERS.md`](TELEGRAM_PARITY_WORK_ORDERS.md) — epic **WO-335** (child of WO-321).

| Wave | WO | Window | Focus | Implementation (2026-08-27) |
|------|-----|--------|-------|------------------------------|
| Now | WO-321… | → Sept 1 2026 | E2E chat + App Store P0 | Do not block on T0–T3 |
| **T0** | **WO-336** | Post-Sept | Durable channels + admin | HTTP + admin views exist; soak + Xcode E2E |
| **T1** | **WO-337** | Parallel | Saved / drafts / forward / search | Sources in EchoApp; two-device E2E |
| **T2** | **WO-338** | Parallel | Folders + schedule HTTP + bots | `/v3/messages/schedule` mounted (in-memory); iOS still local BGTask |
| **T3** | **WO-339** | Later | Stickers / translate / import | Backlog |
| **S1** | **WO-333** | Shared | History sync + backup | Same as Signal Wave S1 |

**FEATURE:** Telegram-class Messaging UX (`3ddcb0c4-d216-47c6-84f2-3a45dd35dd73`)  
**Blueprint:** Telegram-class Messaging UX (`cc331999-bf7c-497a-ba5c-924efe52bc4e`)  
**Assignee:** Chad Cromwell  
**Sept 1:** Do **not** block soft launch; market only what two-device E2E proves.

---

## 6. iOS / Xcode remaining (not missing Swift files)

Almost every messaging `.swift` file under `Sources/Features/Messaging`, `Sources/Services`, and `Sources/Features/Calling` is **already** in `EchoApp.xcodeproj`. The remaining work is **targets, entitlements, and live wiring**, not “add file to compile”.

See the Xcode checklist in the agent reply / [`TELEGRAM_PARITY_WORK_ORDERS.md`](TELEGRAM_PARITY_WORK_ORDERS.md) §Xcode.

---

## 7. Gateway endpoints (Telegram-adjacent)

| Echo | Role |
|------|------|
| `GET/PUT /v3/chat-folders` | Folder JSON blob per DID |
| `POST/GET /v3/messages/schedule` | Create / list scheduled (opaque content; list omits body) |
| `GET/PATCH/DELETE /v3/messages/schedule/{id}` | Owner read / edit / cancel |
| `POST /v3/messages/schedule/{id}/send-now` | Immediate deliver via in-memory `MessagingService` |
| `/v3/broadcasts/*` | Channels |
| `/v3/backup/*` + `/v3/sync/*` | Cloud copy analog (S1) |
| `/v3/bots/*` | Bot foundation |

Scheduled queue is **in-memory** until a Postgres store lands (WO-338). iOS local BGTask remains the durable path on-device.

---

## 8. Honest “comparable?” answer

- **Daily chat UX:** Close on drafts, Saved Messages, forward, search, quotes once T1 two-device E2E is green.  
- **Cloud account:** Telegram users expect the cloud to remember everything; Echo will only match that after **S1** backup/sync.  
- **Channels:** API exists; do not market as Telegram-durable until soak.  
- **Sept 1:** Stay on go-live scope; this document is the **post-launch** Telegram-parity program.
