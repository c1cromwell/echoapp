# Echo agent bus

Agents call `/v3/agent/*`. The user key stays on the device. The server accepts ciphertext and routing metadata.

Agent calls use `Authorization: Grant <token>`. A user JWT or `X-Bot-Token` on these routes returns `AGENT_GRANT_REQUIRED`. Creating a grant, confirming a challenge, and storing recovery share metadata use the existing passkey headers (`X-Sender-DID`, `X-Signature`, `X-Timestamp`).

| Verb | Route | Who calls it |
|------|--------|----------------|
| Issue grant | `POST /v3/agent/grants` | User passkey. Body: `agent_did`, `agent_name`, `verb`, `thread_id` or `*`, `expires_at`, optional `emit_signals`. |
| List / revoke | `GET /v3/agent/grants`, `DELETE /v3/agent/grants/{id}` | User passkey. List responses omit the token. |
| Resolve | `POST /v3/agent/resolve` | Grant `send` or `read_metadata`. Body `{ "name": "@bob" }` or `{ "room": "dm:…" }`. Returns `did` and `thread_id`. |
| Send | `POST /v3/agent/send` | Grant `send`. Body: `to`, `ciphertext`. Optional `expires_at`, `render_hint` (`text`, `spoken`, `card`, `silent`). A `text` field is `400 PLAINTEXT_REFUSED`. |
| Confirm | `POST /v3/agent/confirm` | User passkey only. Body `{ "challenge_id" }`. |
| Signal | `POST /v3/agent/signal` | Grant `send`. Typing is published only when the grant has `emit_signals`. |
| Search | `POST /v3/agent/search` | Grant `search` still returns `400 SEARCH_EDGE_ONLY`. Search runs on the device. |
| Recovery metadata | `POST /v3/agent/recovery/shares` | User passkey. `index` and `holder_did` only. `share_bytes`, `secret`, and `key` are refused. |
| Rotate | `POST /v3/agent/rotate-key` | Grant `rotate_key` opens a challenge. The generation does not change until the user confirms. |

`pay`, `forward`, and `rotate_key` do not deliver. They return `409` with `needs_confirmation` and `challenge_id`. Confirming the same challenge twice returns `409 CHALLENGE_REPLAY`.

A sealed send is published as WebSocket type `sealed_text`. The payload carries ciphertext, expiry, render hint, thread id, and `untrusted`. `untrusted` is true when the recipient has not added the sender as a contact. The HTTP response is `queued`, `thread_id`, and `untrusted`. It does not echo the blob.

Audit rows are agent DID, verb, thread id, and time. They do not include the payload.

The iOS grant list is **Privacy → Agent grants**. It is not a tab.
