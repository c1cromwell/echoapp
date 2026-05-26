---
name: echo-auth-contracts
description: >-
  Echo authentication and realtime contracts: passkey-signed REST, JWT WebSocket,
  ephemeral signal envelopes. Use when implementing or reviewing auth, login,
  WebSocket, typing indicators, read receipts, reactions, or API middleware.
---

# Echo auth & realtime contracts

## REST — passkey signing (WO-1)

All protected REST requests use **`PasskeySigningInterceptor`** (`ios/Echo/Sources/Core/Networking/APIClient.swift`, `internal/api/passkey_auth.go`).

Headers (do **not** re-implement in feature code):

- `X-Sender-DID` — user's `did:key`
- `X-Signature` — ECDSA P-256 over canonical string
- `X-Timestamp` — Unix seconds

Canonical string (must match server):

```text
{METHOD}\n{PATH}\n{TIMESTAMP}\n{hex(SHA256(body))}
```

Public paths skip signing (see `PasskeySigningInterceptor.publicPaths` and `router.go` `publicPaths`).

## REST — session JWT (WO-287)

- Issue/refresh/revoke: `internal/auth/token.go`, `internal/auth/service.go`
- ES256 access tokens; refresh rotation with reuse detection
- iOS: `AuthenticationInterceptor` adds `Authorization: Bearer`

## WebSocket — connect

Server: `internal/api/ws.go` → `ServeWS` on **`/ws`**

Auth (either):

- Header: `Authorization: Bearer <accessToken>`
- Query: `?token=<accessToken>`

iOS URL builder: `WebSocketURLBuilder.webSocketURL(apiBaseURL:accessToken:)`  
Identity for routing: JWT `sub` = user **DID** (`router.UserIDExtractor`).

## WebSocket — ephemeral signals (Phase 3)

**Not** `WSRelayMessage`. Use `WSEnvelope` (`ios/Echo/Sources/Core/Networking/ConversationSignal.swift`):

```json
{
  "type": "typing|read_receipt|reaction",
  "to": "<peer DID>",
  "conversation_id": "<id>",
  "payload": { ... },
  "timestamp": "<RFC3339>"
}
```

Rules (`internal/api/ws.go`):

- `to` is **required**; empty or self → dropped (not broadcast)
- Never persisted; best-effort relay
- `from` overwritten by server

Payload field names are **snake_case** (see `PHASE3_IOS_UI_SPEC.md`).

## REST — reactions (durable)

- `POST /v3/messages/react` — body `{ "message_id", "emoji" }` (empty emoji = remove)
- `GET /v3/messages/reactions?message_id=`
- iOS: `ReactionsAPI.swift`; server: `internal/api/v3_handlers.go`

WS `reaction` signal is latency optimization; REST is source of truth.

## Identity registration (public)

- `POST /identity/register`, `GET /identity/{did}`, device add/token flows
- `pkg/didkey/` — P-256 → `did:key`

## Checklist for new endpoints

1. Public or passkey-protected? Update both Go router and iOS public path lists if public.
2. WS or REST? WS uses JWT only; REST uses passkey interceptor.
3. Any chain submission? Run **`echo-t0-t7-review`** (Step 2 skill).
