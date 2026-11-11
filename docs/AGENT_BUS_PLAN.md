# Echo agent message bus — implementation plan

**Status: phases 1–7 implemented and unit-tested (9 Oct 2026).** Verb list: [`AGENT_BUS.md`](AGENT_BUS.md). Human review still open — see the table at the bottom. Grants, challenges, and audit rows live in process memory.

**Thesis:** the messaging graph stays; the app shell does not grow. Other agents call an encrypted identity and delivery layer. The only Echo UI for that path is enroll, recover, and permission review.

**Do not block messaging E2E on this plan.** Two-simulator → phone → TestFlight → two phones stays [`E2E_TESTING.md` §5](E2E_TESTING.md#5-messaging-tests--run-in-this-order).

Frozen: `FirstRunCoordinator`, `GlacialLoginScreen`. Phase 1–2 identity stays `did:key` (ADR-0001). No Cardano. Server stays ciphertext-only (T0–T7).

---

## What already is the bus

| Piece | Where | Keep |
|-------|--------|------|
| Content-blind relay | `sealed_text` / opaque WS in `internal/api/ws.go` | Yes |
| Hardware identity | Secure Enclave `did:key`, passkey-signed REST | Yes |
| Names | `@username` → key, `echo://invite?u=` | Yes |
| Edge search | `LocalMessageIndexer` + `EncryptedIndexStore` | Yes |
| Phrase backup | `/v3/backup/*` | Yes |
| Bot-shaped grants | `/v3/bots/*`, `internal/services/bots` | Replace, do not extend the marketplace UI |

## What this plan adds

| Phase | Outcome | Lane |
|-------|---------|------|
| 1 | Resolve a name or room, send a sealed blob | Headless |
| 2 | Scoped, expiring grant signed by the hardware key | Headless |
| 3 | Agent DID, audit of reads/sends, no plaintext relay | Headless |
| 4 | `needs_confirmation` for pay, forward, key change | Headless |
| 5 | Recipients, expiry, render hint; receipts optional | Headless |
| 6 | Temporary search grant over the local index | Headless |
| 7 | Grant-review screen; messaging recovery uses the Passport threshold | Thin UI |

Stop: new chat tabs, folders, sticker packs, channel-admin chrome, server full-text search, a vendor account.

---

## Phase 1 — Resolve and sealed send

**Add**

- `POST /v3/agent/resolve` — body `{ "name": "@bob" }` or `{ "room": "…" }` → `{ "did", "thread_id" }`. Reuse username search and `dm:{sorted-did}:{sorted-did}`. No conversation list.
- `POST /v3/agent/send` — sealed payload only (`ciphertext`, `to`). Reject a `text` field.
- Subscribe stays the existing WebSocket. Document the verb set in one page next to this file when the handlers land.

**Tests:** `go test ./internal/api/ -run AgentResolve`. Fixture: two registered `did:key` users, resolve `@bob`, send opaque bytes, assert the peer queue contains those bytes and the handler returns 400 if `text` is set.

**Done when:** those tests pass. No Simulator.

## Phase 2 — Scoped capability

**Add**

- Grant object: `verb` (`send` | `read_metadata` | `search`), `thread_id` or `*`, `exp`, holder agent id.
- `POST /v3/agent/grants` requires a passkey signature from the user key (`PasskeySigningInterceptor` shape). Response is the grant, not a copy of the identity key.
- Agent calls in phase 1 accept `Authorization: Grant <token>` and 401 when expired or the verb does not match.
- User JWT and `X-Bot-Token` stay for the existing app and WO-11. New agent clients cannot use them on `/v3/agent/*`.

**Tests:** expired grant rejected; `send` grant cannot call search; signature from a different key rejected.

**Done when:** `go test ./internal/api/ -run AgentGrant` passes.

## Phase 3 — Agent record and audit

**Add**

- Agent is a `did:key` plus a display name the user approved at grant time.
- Append-only audit row per read or send: agent DID, verb, thread id, time. No ciphertext, no plaintext.
- `internal/api/bot_relay_handlers.go` `handleBotRelayMessage`: stop accepting `text`. Ciphertext only. Same for any payment path that echoes a body.

**Tests:** audit row written on send; log line does not contain the payload; plaintext bot body returns 400.

**Done when:** `go test ./internal/api/ -run 'AgentAudit|BotRelay'` passes.

## Phase 4 — Confirmation gate

**Add**

- Verbs `pay`, `forward`, and `rotate_key` return `409` / `needs_confirmation` with a challenge id.
- `POST /v3/agent/confirm` completes only with a fresh passkey signature. The agent cannot confirm alone.
- Inbound message payload gains `untrusted: true` when the sender is not a contact. Agents that act on content must see that flag; the gateway does not strip it.

**Tests:** pay without confirm is not relayed; confirm with the user key relays once; replay of the challenge fails.

**Done when:** `go test ./internal/api/ -run AgentConfirm` passes.

## Phase 5 — Delivery object

**Add**

- Sealed send accepts optional `expires_at` and `render_hint` (`text` | `spoken` | `card` | `silent`). Stored as routing metadata beside the blob. The gateway does not choose a screen.
- Typing, delivered, and read stay `WSEnvelope` types and become optional: omit them unless the grant includes `emit_signals`.

**Tests:** hint and expiry round-trip on the queued envelope; a grant without `emit_signals` does not fan out typing.

**Done when:** `go test ./internal/api/ -run AgentDelivery` passes.

## Phase 6 — Search grant

**Add**

- No server search index.
- iOS: `KeywordSearchEngine` accepts a grant check (verb `search`, unexpired) before it returns hits. The index key never leaves `EncryptedIndexStore`.
- A headless Mac tool can call the same check only against a local store the user unlocked. It does not upload the index.

**Tests:** `swift test --filter AgentSearchGrant` (new Phase3 test): expired grant returns no hits; valid grant returns a fixture hit; store file is not copied to the result.

**Done when:** that test passes. Human check is not required.

## Phase 7 — Thin UI

**Add**

- One screen, **Grant review**: agent name, verb, room, expiry, revoke, and the confirm sheet for phase 4. No new tab in `MainTabView`.
- Recovery: keep the 24-word phrase. Add the ADR-0004 Shamir threshold (already in `pkg/passport/recovery/`) as an optional path for the **messaging** root. Device add and key rotation still require the hardware key or M-of-N. The agent can request them and cannot finish them.

**Tests:** SwiftUI grant list is an Xcode pass (agent cannot tap it). Threshold math stays the existing Passport tests; add one messaging test that reconstruction does not happen in the Go process.

**Done when:** grant review is in the EchoApp target, phrase restore still works on one Simulator, and a second device cannot rotate keys with only the agent grant.

---

## Order and non-goals

Ship 1 → 2 → 3 before any UI. Phases 4–6 can follow in parallel after 2. Phase 7 is last.

Not in this plan: car/watch/glasses clients, a natural-language parser inside the gateway (the caller resolves language, then calls `resolve`), Telegram parity waves T0–T3, Rewards/wallet changes.

## How we will know it is tested

| Gate | Command or person | 9 Oct 2026 |
|------|-------------------|-----------------|
| Phases 1–5 | `go test ./internal/api/ -run Agent` plus `BotRelay` | Passed |
| Phase 6 | `swift test --filter AgentSearchGrant` | Passed (4 tests) |
| Phase 6 iOS compile | `swift build` for the iOS Simulator triple | Passed. `KeywordSearchEngine` drops hits when a grant is present and invalid. |
| Phase 7 model | Same Swift filter (`reviewRevokeAndConfirmOnce`) | Passed. Revoke and one-shot confirm. |
| Phase 7 screen | Human, Simulator: Privacy → Agent grants, then phrase restore | Not run. The screen is in the EchoApp target and compiled; it opens on an empty local list and does not call the gateway or the device passkey. |
| Screen catalog PNG | `make screen-catalog` | Preview registered (`privacy` / `05-agent-grants`). Gallery not regenerated. |
| Existing chat | [`E2E_TESTING.md` §5](E2E_TESTING.md) stages 1–4 | Not run from this work. |
