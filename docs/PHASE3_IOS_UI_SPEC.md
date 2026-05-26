# Phase 3 iOS UI Implementation Spec — Typing · Read Receipts · Reactions

**Purpose:** Step-by-step build spec for the iOS client layer of three Phase 3
launch-critical features. The **backend is implemented and tested**
(`internal/api/ws.go`, `internal/api/ws_signals_test.go`, `internal/api/v3_handlers.go`,
`internal/api/reactions_handler_test.go`). This spec is the client layer.

### Agent layer — landed (May 2026)

The following are **implemented and covered by `EchoPhase3Tests`** (run in Xcode CI or
`cd ios/Echo && swift test --filter EchoPhase3Tests`):

| File | Role |
|------|------|
| `Core/Networking/ConversationSignal.swift` | `WSEnvelope`, payloads, codec |
| `Core/Networking/WebSocketURLBuilder.swift` | `API_URL` → `wss://…/ws?token=` |
| `Core/Networking/ConversationSignalTransport.swift` | iOS WS transport bridge |
| `Services/ConversationSignalService.swift` | Send/receive typing, receipts, reactions |
| `Services/ReactionsAPI.swift` | `POST/GET /v3/messages/react(ions)` |
| `Features/Messaging/ConversationSignalLogic.swift` | Privacy merge, toggle, status advance |
| `Features/Messaging/ChatDetailViewModel.swift` | ViewModel ready for SwiftUI wiring |
| `Core/DI/Container.swift` | Registers `conversationSignalService`, `reactionsAPI` |
| `Phase3Tests/*` | Unit tests (codec, service, logic, ViewModel) |

**Still for Xcode:** SwiftUI views (`TypingIndicatorView`, `ReactionPickerView`,
`ReactionChipsView`), refactor `ChatView` to bind `ChatDetailViewModel`, add new
sources to `EchoApp.xcodeproj`, manual two-client checklist (Step 5 manual).

---

## Implementation readiness (May 2026)

| Area | Status | Notes |
|------|--------|-------|
| Backend WS envelope + relay rules | ✅ Shipped | Ephemeral types require non-empty `to` ≠ sender |
| Backend REST reactions | ✅ Shipped | `POST /v3/messages/react`, `GET /v3/messages/reactions` |
| iOS wire models + signal service | ❌ Not started | No `ConversationSignal.swift` yet |
| iOS WS auth token on connect | ❌ Not started | `WebSocketClient` still uses `wss://ws.echo.local`, no `?token=` |
| iOS chat UI wired to domain models | ⚠️ Partial | `ChatView` uses local `ChatMessage` + `MessageStatus`, not `Message` + `DeliveryStatus` |
| iOS ViewModel for open chat | ⚠️ Partial | `MessagingViewModel` exists but has no typing/receipt/reaction state |
| Privacy toggles | ✅ Models exist | See **Privacy flags** below — names differ from early draft |

**Can this be implemented with tests?** **Yes**, with a deliberate split:

1. **Agent / CI (no Xcode UI runtime)** — wire models, `ConversationSignalService`,
   reactions REST client, pure ViewModel logic, and XCTest coverage in a **dedicated
   SPM test target** (mirror `EchoSecurityTests`). These compile and run via
   `swift test --filter EchoPhase3Tests` on macOS CI.
2. **Xcode (simulator or device)** — SwiftUI surfaces, live WebSocket against
   `make dev`, two-account manual flows, animation polish, and wiring into
   `EchoApp.xcodeproj` + `Container.swift` lifecycle.

The headless Cursor agent **cannot** run the Xcode app or dial a real WebSocket to
your LAN backend; it **can** land Steps 0–3 logic + unit tests and leave Step 4–5
UI/integration for Xcode.

---

## Build ownership matrix

Use this when assigning work between the agent and Xcode.

| Step | Deliverable | Agent + unit tests | Xcode required |
|------|-------------|:------------------:|:--------------:|
| **0a** | `ConversationSignal.swift` encode/decode | ✅ | — |
| **0b** | `ConversationSignalService` send/receive dispatch | ✅ (mock delegate) | — |
| **0c** | `WebSocketClient.connect(accessToken:)` / URL from `API_URL` | ✅ URL + request shape tests | Live socket smoke |
| **1a** | Typing debounce + `peerIsTyping` ViewModel logic | ✅ | — |
| **1b** | `TypingIndicatorView` (Glacial) | Compile-only in SPM | Simulator visual QA |
| **2a** | Read-receipt send batching + `DeliveryStatus` advance | ✅ | — |
| **2b** | Checkmark UI on sent bubbles | — | ✅ Wire `SmartCheckmarkView` / `DeliveryStatus` |
| **3a** | `ReactionsAPI` + toggle logic | ✅ | — |
| **3b** | `ReactionPickerView`, `ReactionChipsView`, long-press | — | ✅ |
| **4** | DI (`Container.swift`), app foreground WS lifecycle | Partial (registration) | ✅ App target + schemes |
| **5 manual** | Two clients vs `make dev` | — | ✅ **Required** |

### Recommended new test target

Add `EchoPhase3Tests` in `ios/Echo/Package.swift` (same pattern as
`EchoSecurityTests`) so Phase 3 tests do not depend on fixing all pre-existing
`EchoTests` compile issues. CI hard-gate optional once stable.

---

## Codebase alignment (read before coding)

### ViewModel name

Use **`ChatDetailViewModel`** (`Features/Messaging/ChatDetailViewModel.swift`) — wired
for Phase 3 signals. The older **`MessagingViewModel`** remains for conversation list only.
Refactor `ChatView` in `MessagingScreens.swift` to accept `@Bindable var viewModel: ChatDetailViewModel`.

### Dual message/status types (must unify)

| Layer | Message type | Status type |
|-------|--------------|-------------|
| Domain | `Domain/Models/Models.swift` → `Message` | `Features/Evidence/DeliveryStatus.swift` |
| Presentation mock | `ChatMessage` in `MessagingScreens.swift` | `MessageStatus` in `MessageBubbles.swift` |

Phase 3 **must** use `Message`, `Reaction`, and `DeliveryStatus` in the wired chat path.
`MessageBubble` can adopt `DeliveryStatus` or map at the view boundary — but do not
maintain two parallel status enums long term.

### Privacy flags (not `SuppressTyping` / `SuppressReceipts`)

Use existing settings:

| Spec intent | Global (`EnhancedPrivacySettings`) | Per-persona (`PersonaPrivacySettings`) |
|-------------|--------------------------------------|----------------------------------------|
| Send/show typing | `typingIndicators` | `sendTypingIndicators` |
| Send/show read receipts | `readReceipts` | `sendReadReceipts` |

Honor **both** when a persona is active (persona overrides global if stricter).

### WebSocket identity = JWT subject (DID)

`ServeWS` maps connections with `UserIDExtractor` → JWT `sub` (the user's **DID**).
Ephemeral `to` / peer routing must use the peer's **`did:key`**, not an internal user UUID.

### Envelope mismatch (critical)

The server parses **`WSMessage` / `WSEnvelope`**, not `WSRelayMessage`:

- `WSRelayMessage` (`WebSocketClient.swift`) — used by `MessageRelayManager` for E2E blobs.
- **New signals** — JSON shape in [Backend contracts](#backend-contracts-already-shipped--build-the-client-to-match-exactly).

Do not send typing/receipt/reaction through `WSRelayMessage`.

### Reuse list

- `Core/Networking/WebSocketClient.swift` — `actor WebSocketClient`, `WebSocketDelegate`
- `Core/Networking/APIClient.swift` — `PasskeySigningInterceptor` (REST only; do not re-sign WS)
- `Domain/Models/Models.swift` — `Message`, `Reaction`, `Conversation`, `User`
- `Features/Evidence/DeliveryStatus.swift` — `Comparable` delivery lifecycle
- `Presentation/ViewModels/ViewModels.swift` — `MessagingViewModel` (extend)
- `Core/DI/Container.swift` — register new services

---

## Backend contracts (already shipped — build the client to match exactly)

### WebSocket envelope

The server (`internal/api/ws.go`) routes a single JSON envelope. **iOS must send this
exact shape**:

```jsonc
{ "type": "...", "from": "<senderDID>", "to": "<recipientDID>",
  "conversation_id": "<convID>", "payload": { ... }, "timestamp": "<RFC3339>" }
```

Rules enforced server-side:

- **Ephemeral signal types** = `typing`, `read_receipt`, `reaction`. Relayed **only to
  `to`**; dropped if `to` is empty or equals sender. Never broadcast, never persisted.
  Client **must always set `to`** = peer DID.
- `from` is overwritten by the server with the authenticated sender; clients may omit it.
- WS auth: `Authorization: Bearer <accessToken>` **or** `?token=<accessToken>` on `/ws`.

Signal payloads (snake_case field names):

| `type` | payload |
|--------|---------|
| `typing` | `{ "conversation_id": String, "state": "start" \| "stop" }` |
| `read_receipt` | `{ "conversation_id": String, "message_ids": [String], "read_at": RFC3339 }` |
| `reaction` | `{ "conversation_id": String, "message_id": String, "emoji": String }` (empty `emoji` = removed) |

### REST (reactions — durable source of truth)

Auth: signed request via `PasskeySigningInterceptor`. Base URL = `APIClient` host.

- `POST /v3/messages/react` → `{ "message_id": String, "emoji": String }`
  (empty `emoji` removes caller's reaction). Response:
  `{ "message_id": String, "reactions": [ { "emoji": String, "count": Int, "reactors": [DID] } ] }`
- `GET /v3/messages/reactions?message_id=<id>` → same response shape.

> REST is truth; WS `reaction` is latency optimization. Reconcile with GET when stale.

---

## Step 0 — Shared: conversation-signal envelope + transport

### 0a — Wire models (agent ✅)

**New file:** `Core/Networking/ConversationSignal.swift`

```swift
struct WSEnvelope<Payload: Codable>: Codable {
    var type: String
    var to: String
    var from: String?
    var conversationId: String?
    var payload: Payload
    var timestamp: String?
    enum CodingKeys: String, CodingKey {
        case type, to, from
        case conversationId = "conversation_id"
        case payload, timestamp
    }
}
struct TypingPayload: Codable { /* conversation_id, state */ }
struct ReadReceiptPayload: Codable { /* conversation_id, message_ids, read_at */ }
struct ReactionPayload: Codable { /* conversation_id, message_id, emoji */ }
```

**Unit tests:** round-trip JSON for each payload; assert snake_case keys via
`JSONEncoder` output string contains `"conversation_id"`, etc.

### 0b — Signal service (agent ✅)

**New file:** `Services/ConversationSignalService.swift`

Wraps `WebSocketClient`, conforms to `WebSocketDelegate`:

- `sendTyping(conversationId:peerDID:state:) async`
- `sendReadReceipt(conversationId:peerDID:messageIds:) async`
- `sendReaction(conversationId:peerDID:messageId:emoji:) async`
- Inbound: `AsyncStream` or Combine publishers — `typingEvents`, `readReceiptEvents`, `reactionEvents`
- Decode in `webSocketDidReceiveMessage`; ignore unknown `type` (forward-compat)

**Unit tests:** mock delegate receives synthetic JSON; verify dispatch by `type`;
unknown type ignored.

### 0c — WS connect with token (agent partial, Xcode smoke)

**Modify:** `WebSocketClient.swift`

- Add `connect(accessToken: String, delegate:)` building URL:
  `ws(s)://<API_HOST>/ws?token=<accessToken>` (derive host/scheme from `API_URL` / build setting).
- Read token from Keychain same path as `AuthenticationInterceptor` (`getAuthToken()`).
- Replace default `wss://ws.echo.local` for production paths.

**Agent tests:** URL construction given mock `API_URL`.

**Xcode smoke:** connect two simulators; confirm hub registers both DIDs (backend logs / integration test pattern in `test/integration/ws_test.go`).

---

## Step 1 — Typing indicators

### 1a — Logic (agent ✅)

In chat ViewModel (extended `MessagingViewModel` or new `ChatDetailViewModel`):

- Debounce ~**1.5s** on input change; send `state:"start"` on burst start;
  `state:"stop"` on clear, **4s idle**, or send.
- `@Published var peerIsTyping: Bool`; set from inbound `typing` for open `conversationId`;
  auto-clear after **6s** if `stop` missed.
- If `!privacy.typingIndicators` (and persona `sendTypingIndicators`), **do not send**;
  optionally hide inbound indicator when user disabled showing (product choice: usually still show peer's if they send).

### 1b — UI (Xcode ✅)

- **New:** `Features/Messaging/TypingIndicatorView.swift` — three animated dots, Glacial spacing.
- Insert above composer in refactored `ChatView` when `peerIsTyping`.

---

## Step 2 — Read receipts

### 2a — Logic (agent ✅)

- On peer messages visible (`onAppear` / scroll-to-bottom): batch unacknowledged peer
  `message.id`s → `sendReadReceipt(..., peerDID: originalSender.did)`.
- Local: mark read via repository / `markAsRead` (sets `isRead`).
- Inbound receipt: advance **own** messages' `deliveryStatus` with `max(current, .read)` —
  `DeliveryStatus` is `Comparable`; never regress.
- If `!readReceipts` / `!sendReadReceipts`, do not send; cap displayed status at `.delivered` for outbound.

### 2b — UI (Xcode ✅)

- Use `SmartCheckmarkView` / `DeliveryStatus.icon` on **sent** bubbles (not `MessageStatus` strings).

---

## Step 3 — Reactions

### 3a — API + logic (agent ✅)

**New or extend:** `Services/ReactionsAPI.swift` (or methods on repository using `APIClient`)

```swift
struct ReactionCount: Codable {
    let emoji: String
    let count: Int
    let reactors: [String]
}
```

- `react(messageId:emoji:)` → POST `/v3/messages/react`
- `removeReaction(messageId:)` → POST with `emoji:""`
- `reactions(messageId:)` → GET

Toggle: select emoji → POST; same emoji again → remove. Optimistic UI, replace with server response.
After successful REST, also `sendReaction(...)` over WS.

**Unit tests:** toggle state machine; mock API responses.

### 3b — UI (Xcode ✅)

- **New:** `ReactionPickerView` — long-press bubble → 👍 ❤️ 😂 😮 😢 🙏 + more
- **New:** `ReactionChipsView` — `emoji × count`, highlight if `reactors` contains current DID
- Inbound WS reaction → update chips; stale → GET reconcile

---

## Step 4 — Wiring & configuration

| Task | Owner |
|------|-------|
| Register `ConversationSignalService` singleton in `Container.swift` | Xcode app target |
| Inject into chat ViewModel with `MessagingService` / repository | Agent + Xcode |
| WS URL + REST URL from `API_URL` (`docs/PHASE1_LAUNCH.md` §1c) | Agent |
| Connect signal service on conversation open / foreground; teardown on background | Xcode lifecycle |
| Add new Swift files to **`EchoApp.xcodeproj`** (SPM library alone is not enough for app builds) | Xcode |

---

## Step 5 — Tests

### Unit / logic (agent ✅ — `EchoPhase3Tests`)

| Test | Assert |
|------|--------|
| Envelope encode/decode | snake_case keys, each payload type |
| `ConversationSignalService` send | JSON `type`, `to`, nested payload |
| `ConversationSignalService` receive | Dispatches by `type`; ignores unknown |
| Reaction toggle | Select → reselect removes |
| Read receipt status | Only advances `DeliveryStatus`; never `.sent` ← `.read` regression |
| Typing debounce helper | start once per burst; stop on idle (inject clock) |
| Privacy gates | No WS send when toggles off |

Run locally:

```bash
cd ios/Echo
swift test --filter EchoPhase3Tests
```

### Manual (Xcode + `make dev` — **not automatable in agent**)

Prerequisites: `API_URL` → LAN backend; two accounts signed in (simulator + device or two simulators if supported).

1. **Typing:** A types → B sees indicator; A stops/sends → clears; kill A mid-type → B auto-clears ~6s.
2. **Read receipts:** B opens chat → A sees read; disable B receipts → A stays delivered.
3. **Reactions:** A long-press → 👍 on B's message; toggle off; `GET /v3/messages/reactions` matches; B offline → reconcile on reopen.
4. **Privacy leak:** Third account C online, not in thread → receives **no** typing/receipt/reaction (server guard + client must set `to` correctly).

---

## Gotchas / decisions

- **Envelope mismatch:** server uses `WSEnvelope` shape above, *not* `WSRelayMessage`.
- **`to` is mandatory** for ephemeral signals — omitting silently drops (by design).
- **Best-effort:** signals not queued; use typing 6s timeout; reactions/receipts reconcile via REST.
- **1:1 first:** one `to` per signal. Group fan-out = one signal per peer (future).
- **Signing:** REST via interceptor; WS via bearer/query token only.
- **Two MessagingService types:** consolidate or inject repository into ViewModel before shipping.
- **CI:** full `EchoTests` may not compile; use `EchoPhase3Tests` for green gate.

---

## Suggested implementation order

1. Agent: Step 0a–0b + tests + `EchoPhase3Tests` target  
2. Agent: Step 0c URL/token + 3a ReactionsAPI + tests  
3. Agent: ViewModel logic (1a, 2a) + tests  
4. Xcode: Refactor `ChatView` → domain models + ViewModel  
5. Xcode: Steps 1b, 2b, 3b UI + Step 4 DI/lifecycle  
6. Xcode: Step 5 manual checklist against `make dev`
