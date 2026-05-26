# Phase 3 iOS UI Implementation Spec — Typing · Read Receipts · Reactions

**Purpose:** A portable, step-by-step build spec for the iOS UI of the three Phase 3
launch-critical features. The **backend is already implemented and tested** (see
"Backend contracts" below); this spec is the client layer. Build and verify it in
Xcode (simulator + device) — it cannot be run in the headless agent environment.

**Target files live under** `ios/Echo/Sources/`. Existing pieces you'll reuse:
- `Core/Networking/WebSocketClient.swift` — `actor WebSocketClient`, `WebSocketDelegate`.
- `Core/Networking/APIClient.swift` — REST client; the `PasskeySigningInterceptor`
  already signs every request with `X-Sender-DID` + `X-Signature` + `X-Timestamp`
  (do **not** re-implement signing).
- `Domain/Models/Models.swift` — `Message` (has `reactions: [Reaction]?`, `isRead`,
  `deliveryStatus`), `Reaction`, `Conversation`, `User`.
- `Features/Evidence/DeliveryStatus.swift` — `DeliveryStatus` enum (icons + labels).
- `Services/MessagingService.swift`, `Presentation/Screens/Messaging/MessagingScreens.swift`.
- `Core/DI/Container.swift` — dependency wiring.

---

## Backend contracts (already shipped — build the client to match exactly)

### WebSocket envelope
The server (`internal/api/ws.go`) routes a single JSON envelope. **iOS must send this
exact shape** (the existing `WSRelayMessage` struct is a *different* shape and is not
what the server parses for these signals):

```jsonc
{ "type": "...", "from": "<senderDID>", "to": "<recipientDID>",
  "conversation_id": "<convID>", "payload": { ... }, "timestamp": "<RFC3339>" }
```

Rules enforced server-side:
- **Ephemeral signal types** = `typing`, `read_receipt`, `reaction`. They are relayed
  **only to `to`** and are **dropped if `to` is empty or equals the sender** — never
  broadcast, never persisted. So the client **must always set `to`** = the peer's DID.
- `from` is overwritten by the server with the authenticated sender; clients may omit it.
- WS auth: connect with `Authorization: Bearer <accessToken>` **or** `?token=<accessToken>`.

Signal payloads (match field names exactly — they are snake_case):

| `type` | payload |
|---|---|
| `typing` | `{ "conversation_id": String, "state": "start" \| "stop" }` |
| `read_receipt` | `{ "conversation_id": String, "message_ids": [String], "read_at": RFC3339 }` |
| `reaction` | `{ "conversation_id": String, "message_id": String, "emoji": String }` (empty `emoji` = removed) |

### REST (reactions — the durable source of truth)
Auth: standard signed request (interceptor handles it). Base path joins `APIClient.baseURL`.
- `POST /v3/messages/react` → body `{ "message_id": String, "emoji": String }`
  (empty `emoji` removes the caller's reaction). Response:
  `{ "message_id": String, "reactions": [ { "emoji": String, "count": Int, "reactors": [DID] } ] }`
- `GET /v3/messages/reactions?message_id=<id>` → same response shape.

> Design note: the reaction REST call is the truth; the `reaction` WS signal is a
> latency optimization. After applying a WS reaction, optionally reconcile by GET.

---

## Step 0 — Shared: conversation-signal envelope + transport

**New file:** `Core/Networking/ConversationSignal.swift`

1. Define the wire envelope matching the server:
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
   struct TypingPayload: Codable { let conversationId: String; let state: String
       enum CodingKeys: String, CodingKey { case conversationId = "conversation_id"; case state } }
   struct ReadReceiptPayload: Codable { let conversationId: String; let messageIds: [String]; let readAt: String
       enum CodingKeys: String, CodingKey { case conversationId = "conversation_id"; case messageIds = "message_ids"; case readAt = "read_at" } }
   struct ReactionPayload: Codable { let conversationId: String; let messageId: String; let emoji: String
       enum CodingKeys: String, CodingKey { case conversationId = "conversation_id"; case messageId = "message_id"; case emoji } }
   ```
2. **New service** `Services/ConversationSignalService.swift` — wraps `WebSocketClient`,
   conforms to `WebSocketDelegate`, exposes:
   - `func sendTyping(conversationId:peerDID:state:) async`
   - `func sendReadReceipt(conversationId:peerDID:messageIds:) async`
   - `func sendReaction(conversationId:peerDID:messageId:emoji:) async`
   - inbound `AsyncStream`/Combine publishers: `typingEvents`, `readReceiptEvents`, `reactionEvents`.
   Send by JSON-encoding `WSEnvelope` → `webSocketClient.send(text:)`. Receive in
   `webSocketDidReceiveMessage(_:message:)`: decode the envelope's `type`, then decode the
   matching payload and publish (ignore unknown types — forward-compat).
3. **WS connect must carry the access token.** `WebSocketConfiguration.baseURL` currently
   has no token. Add a `connect(token:)` path that appends `?token=<accessToken>` (read
   the token from Keychain the same way `PasskeySigningInterceptor` reads the DID) and
   points at the real backend (`API_URL` host, `ws(s)://…/ws`), not `wss://ws.echo.local`.

---

## Step 1 — Typing indicators

1. **Send** (`Features/Messaging/.../ChatViewModel`): on the message input's text-change,
   debounce ~1.5s. Send `state:"start"` on first keystroke of a burst; send `state:"stop"`
   when input clears or after 4s idle, and on send. Always pass `peerDID` = the 1:1
   conversation's other participant.
2. **Receive/state:** ViewModel holds `@Published var peerIsTyping: Bool`. On a `typing`
   event for the open conversation, set true with `state=="start"`, false on `"stop"`;
   also auto-clear after a 6s safety timeout (in case `stop` is missed — signals are best-effort).
3. **Render:** a `TypingIndicatorView` (three animated dots) shown as a bubble at the bottom
   of the message list when `peerIsTyping`. Match the Glacial design system spacing/colors.
4. **Privacy honor:** respect the existing `SuppressTyping` / privacy settings — if the user
   disabled typing indicators, don't send (and optionally don't show).

## Step 2 — Read receipts

1. **Send:** when the chat is open and messages from the peer become visible (`onAppear`
   of the last message, or scroll-to-bottom), collect the peer-authored `message.id`s not
   yet acknowledged and `sendReadReceipt(...)` to the peer (`to` = original sender).
   Mark them locally read via `MessagingService.markAsRead(_:)` (already sets `isRead`/`readAt`).
2. **Receive:** on a `read_receipt` event, update the matching outgoing messages'
   `deliveryStatus` to `.read` (`DeliveryStatus` is `Comparable` — only advance, never regress).
3. **Render:** message status checkmark next to your own sent messages using
   `DeliveryStatus`’s existing icon (sent → delivered → read). Read = the "double check"
   style already defined in `DeliveryStatus.swift`.
4. **Privacy honor:** respect `SuppressReceipts` — if read receipts are disabled, don't send
   them (and show the peer only up to `delivered`).

## Step 3 — Reactions

1. **Reaction API client** (`Services/MessagingService.swift` or a small `ReactionsAPI`):
   - `func react(messageId:emoji:) async throws -> [ReactionCount]` → `POST /v3/messages/react`.
   - `func removeReaction(messageId:) async throws -> [ReactionCount]` → POST with `emoji:""`.
   - `func reactions(messageId:) async throws -> [ReactionCount]` → GET.
   Define `struct ReactionCount: Codable { let emoji: String; let count: Int; let reactors: [String] }`.
2. **Reaction picker UI:** long-press a message bubble → `ReactionPickerView` (a horizontal
   row of common emoji: 👍 ❤️ 😂 😮 😢 🙏 + "more"). Tapping calls `react(...)`; tapping the
   already-selected emoji calls `removeReaction(...)` (toggle). Optimistically update, then
   replace with the server response.
3. **Live update:** after a successful `react`/`removeReaction`, also
   `sendReaction(...peerDID:emoji:)` over WS so the peer updates instantly. On an inbound
   `reaction` event, update that message's reaction chips; if state looks stale, GET to reconcile.
4. **Render:** a `ReactionChipsView` under the bubble showing `emoji ×count`, highlighting
   chips the current user contributed to (use `reactors` containing your DID).

---

## Step 4 — Wiring & configuration

- **DI** (`Core/DI/Container.swift`): register `ConversationSignalService` as a singleton;
  inject into the chat ViewModel alongside `MessagingService`.
- **Base URLs:** WS at `ws(s)://<API_HOST>/ws`; REST at the existing `APIClient.baseURL`.
  Drive both from the `API_URL` build setting (see `docs/PHASE1_LAUNCH.md §1c`).
- **Lifecycle:** connect the signal service when a conversation opens (or app foreground),
  disconnect/teardown on background; rely on the client's existing reconnect/backoff.

## Step 5 — Tests

**Unit (XCTest, `ios/Echo/Tests`):**
- Envelope encode/decode round-trips for each signal type; snake_case keys verified.
- `ConversationSignalService` send builds the correct JSON; receive dispatches by `type`
  and ignores unknown types.
- Reaction toggle logic (select → reselect removes); read-receipt only advances
  `DeliveryStatus`.

**Manual (simulator + 2 devices/clients against `make dev`):**
1. Point `API_URL` at the LAN backend; sign in two accounts.
2. Typing: A types → B sees the indicator; A stops/sends → it clears; verify it auto-clears
   if B kills A's app mid-type.
3. Read receipts: B opens the chat → A's messages flip to read (double-check); with B's
   receipts disabled, A stays at delivered.
4. Reactions: A long-presses B's message, picks 👍 → both see the chip; A re-taps 👍 → removed;
   `GET /v3/messages/reactions` reflects the final state; kill B's socket, react, reopen →
   B reconciles via GET.
5. **Privacy/leak check:** with a 3rd account C online but not in the conversation, confirm
   C never receives typing/receipt/reaction signals (server already guards this — verify the
   client doesn't broadcast).

## Gotchas / decisions
- **Envelope mismatch:** the server parses the `WSMessage` shape above, *not* the existing
  `WSRelayMessage`. Send the `WSEnvelope` JSON for these signals.
- **`to` is mandatory** for ephemeral signals — omitting it silently drops the signal (by design).
- **Best-effort delivery:** signals aren't queued/persisted; always have a fallback
  (typing auto-clear timeout; reactions/receipts reconcile via REST/`isRead`).
- **1:1 first:** the relay targets a single `to`. Group typing/receipts/reactions need
  per-participant fan-out (send one signal per peer, or a future server-side conversation
  fan-out) — out of scope for the launch shortlist.
- **Signing:** REST reaction calls go through the existing interceptor (do not add headers);
  WS uses the bearer/query token, not request signing.
```
