# Relay metadata review

**Reviewed:** 2026-11-24 · **Scope:** what Echo's server and iOS app learn, store, log or send to
third parties that could connect a person to a message. Message content is out of scope: it is
end-to-end encrypted, and this review does not re-examine the encryption.

**How it was done.** Two read-only passes over the code (Go backend; iOS app), then the fixes in
§3. Everything in §3 I changed and tested myself. Findings in §2 and §4 that I did not touch are
as reported by those passes; each carries a file reference so it can be checked, and the ones I
have not personally re-read are marked *(reported)*.

## 1. The question, and the short answer

*Can someone holding Echo's database, logs or network position say "A messaged B at time T"?*

Before this change: **yes, for every message, for 30 days**, including messages sent with "sealed
sender". After this change: **only for a message still waiting for an offline recipient**, and
then only because the app itself still names both parties in the conversation id it sends (§4.1).
The live server continues to see the sender of every message as it passes, because the app
connects with an authenticated socket (§4.2). Closing those two is client protocol work that needs
testing on devices; it is designed in §5 and not done.

Nothing here makes tracing impossible. See §6 for what no messenger of this design can stop.

## 2. What was found

Ranked by how much each one gives away.

| # | Finding | Where |
|---|---|---|
| 1 | The conversation id is `dm:<DID A>:<DID B>`. It travels in clear with every message, sealed or not, was stored with each queued message, and was sent to Apple inside each push | `ios/Echo/Sources/Features/Messaging/ConversationID.swift:6-9`; server stores what it is given |
| 2 | Queued messages were never deleted on delivery. Each row kept sender, recipient, conversation and timestamps for 30 days | `internal/database/postgres.go` `Dequeue` |
| 3 | The stored row held the whole envelope: `from`, `to`, conversation id, and the client's message id beside the sender | `internal/api/ws.go` `enqueueOffline` |
| 4 | "Sealed sender" hid the sender from the recipient's envelope only. The server issued the sealed token to an authenticated sender and checked it against the same sender | `internal/api/sealed_sender_handlers.go` |
| 5 | Group messages queued one row per offline member, each naming the poster | `internal/api/ws.go` group fan-out |
| 6 | Read receipts, reactions, edits, pins and anchoring confirmations were queued as rows naming both parties | `internal/api/ws.go` `PublishSignal` |
| 7 | A sealed message has no message id, so the id its integrity proof is filed under was built as `<conversation id>:<timestamp>` — both DIDs and the time, kept for good | `internal/api/ws.go` |
| 8 | The sender of every message was kept in two in-memory maps that were never emptied; with no chain configured one of them grew without limit | `internal/metagraph/anchoring.go`, `internal/services/relay` |
| 9 | Each device's Apple push token is stored against its DID; with #1, Apple's own records named both parties per push | `user_devices`; `internal/infra/apns.go` |
| 10 | Contact discovery uploaded contacts' phone numbers readable by the server. The private-matching library is not linked into the app, and the app silently used a test stand-in whose "blinded" values are `base64("mock:<phone>:<n>")` | `ios/Echo/Sources/Services/OPRF/` |
| 11 | Forwarding a message fell back to sending it **unencrypted** if encryption failed | `ios/Echo/Sources/Features/Messaging/MessageComposerLogic.swift` |
| 12 | The socket's access token was sent in the URL (`/ws?token=…`), where proxies and load balancers log it | `ios/Echo/Sources/Core/Networking/WebSocketURLBuilder.swift` |
| 13 | Logs: a username with its DID; a DID with its device label; a DID at phone verification | `internal/api/v3_handlers.go`, `identity_register_handlers.go`, `sms_recovery_handlers.go` |
| 14 | Reactions, edits, pins and reply/forward references are stored with the actor's DID, a message or conversation id and a time, with no expiry *(reported)* | `internal/database/message_ops_postgres.go` |
| 15 | Phone numbers: optional SMS recovery sends the raw number to the server (it must, to send an SMS) while the screen says only a hash is stored; the stored "hash" is an unsalted SHA-256, which for a phone number is reversible by trying every number *(reported)* | `ios/.../SMSOTPSetupView.swift`; `internal/api/sms_recovery_handlers.go` |
| 16 | Ratchet session state and a pre-key **private** key are stored in `UserDefaults`, in a type labelled "Keychain-backed" *(reported)* | `ios/Echo/Sources/Services/DoubleRatchetCoordinator.swift` (`DoubleRatchetStore`) |
| 17 | Link previews are fetched automatically for links in **received** messages, giving the device's IP address to any site a sender chooses *(reported)* | `ios/Echo/Sources/.../MessageLinkPreview.swift` |
| 18 | Many requests bypass the app's proxy setting by using `URLSession.shared`, so the SOCKS/Tor option does not cover them *(reported)* | 14 call sites, listed in the iOS pass |
| 19 | The push extension that would replace notification text on the device exists as a file but is in no build target *(reported)* | `ios/Echo/NotificationService/` |
| 20 | Certificate pinning code exists but no pins are configured, so it is inactive *(reported)* | `ios/.../CertificatePinner.swift` |
| 21 | Compliance (Comply) eDiscovery reads the message queue as a who-messaged-whom index and can send sender and time to the external Digital Evidence service *(reported)* | `internal/database/comply_251_postgres.go`, `internal/services/comply/rbac.go` |

Not found, and checked for: IP addresses stored in the database; IP addresses in any log line of
code that runs; message ids or DIDs logged on the message path; message previews or sender names
in pushes.

## 3. Fixed in this change

| # | Fix | Verified by |
|---|---|---|
| 2, 3 | On delivery, the row's sender, conversation and envelope are erased in the same statement that hands the message over. What remains names the recipient and the delivery time (delivery and read status need it) and is deleted after 24 hours | `TestQueuePrivacyPostgresIntegration`, against a real Postgres |
| 3, 5, 6 | A queued row never records the sender or the conversation: only the recipient. Applies to 1:1, group fan-out and signals | `TestQueuedMessageRecordsRecipientOnly` |
| — | Undelivered messages wait 14 days, not 30 (`ECHO_OFFLINE_RETENTION_DAYS`, 1–30) | same integration test |
| 6 | Anchoring confirmations are delivered live or not at all; never queued, never pushed | `TestAnchorConfirmationIsNeverQueued` |
| 7 | A message with no id gets a random one; the conversation is never part of it | `TestCommitmentIDNeverContainsTheConversation` |
| 8 | The sender recorded for a confirmation is forgotten when the confirmation is sent. With no chain configured, commitments are not collected at all | code; no dedicated test |
| 1, 9 | Pushes carry no conversation id: only a type, and for a missed call that call's random id | `TestAPNsPayloadCarriesNoConversationOrParty` |
| 10 | Without the real private-matching library, contact discovery is off. The test stand-in can no longer be reached from the app | compiles in the app build; no device test |
| 11 | A forward that cannot be encrypted is not sent | compiles in the app build; no device test |
| 12 | The app sends the socket token in the `Authorization` header. The server already accepted it there | `WebSocketURLBuilderTests` |
| 13 | The three log lines no longer contain a DID, username or device label | code |

One switch reverses the storage fixes for a deployment whose purpose is retention:
`ECHO_RETAIN_MESSAGE_METADATA=true`. It exists for Echo Comply (finding 21). It must not be set on
the consumer messaging gateway.

### Behaviour that changed as a result

- The server no longer knows the counterparty of a queued message, so it can no longer route a
  read receipt or reaction on someone's behalf from a REST call. The app already sends these as
  signals addressed to the peer; that path is unaffected. (The REST fallback had mostly stopped
  working already: the sender was being blanked once delivery was acknowledged.)
- A sender who is offline when their message is anchored does not receive the confirmation
  notice. The proof remains retrievable by message id.
- Messages for someone offline for more than 14 days are dropped.
- Comply eDiscovery over the consumer gateway's queue returns nothing useful unless the retention
  switch is on.

## 4. Still open

### 4.1 The conversation id names both parties (finding 1) — the largest remaining leak

The app builds `dm:<A>:<B>` and puts it on every message. The server now stores it only inside the
envelope of a message that is waiting for an offline recipient, and erases it on delivery; it no
longer goes to Apple. But while a message waits, the pair is in the database, and the live server
sees it on every message. Fix: §5.

### 4.2 The server sees the sender of every message as it passes (finding 4)

The socket is authenticated as the sender. "Sealed sender" as built cannot hide the sender from
the server. Fix: §5.

### 4.3 The rest

| Finding | What is needed | Size |
|---|---|---|
| 14 | Stop storing the actor's DID with reactions, edits, pins and references, or key them by a per-message pseudonym; add expiry | Medium: the features read these columns |
| 15 | Hash phone numbers with a server-held secret, not a bare SHA-256; correct the wording on the SMS screen (an auth screen: frozen design, text only) | Small on the server; touches lookups |
| 16 | Move ratchet state and the pre-key private key into the Keychain, with a migration | Medium; needs device testing |
| 17 | Do not fetch link previews for received messages unless the user asks | Small |
| 18 | Route every request through one proxy-aware session so the Tor/SOCKS option covers them all; add an onion address for the gateway (deploy stage) | Medium |
| 19 | Add the notification extension to the app target so the device, not the server, decides what a notification says | Xcode project change |
| 20 | Configure certificate pins | Small, but a wrong pin locks every user out; do it with a rollback plan |
| 21 | Decide how Comply obtains records without the consumer gateway keeping them | Product decision |
| — | In-memory rate-limit buckets keyed by DID or IP are never evicted *(reported)* | Small |
| — | `GET /v3/messages/{id}/status` has no check that the caller is a party to the message *(reported)*. Ids are random, which limits the exposure | Small |

## 5. Design for the two large items (not built)

Both change the wire protocol between two phones, so both sides must change together and the
result has to be tested on real devices. I have not done that and will not claim it works until
it has been.

**Opaque conversation ids.** Keep `dm:<A>:<B>` as the app's private, local name for a thread. On
the wire, send nothing that names the pair. For a sealed message the recipient learns the sender
from inside the decrypted payload and derives the local thread name itself, so the wire field can
be a constant. The server's validation already accepts any id beginning `dm:`.

**Sealed sender that the server cannot see through.**

1. Each account publishes a *delivery key* to its contacts inside encrypted messages, never to
   the server in the clear. The server stores only a hash of it.
2. To send, the app opens a second connection that carries **no** login, and presents the
   recipient's delivery key. The server checks the hash and accepts the message for that
   recipient. It never learns who sent it.
3. The sender's identity, signed, travels inside the encrypted payload.
4. Abuse control moves to what the server can still see: a rate limit per recipient and per
   connection; blocking is enforced by the recipient's device, which drops what it does not want.
5. First contact (no delivery key yet) uses the existing authenticated path, or a recipient can
   choose to accept unsealed first messages. That choice is the user's.

**Rotating mailbox identifiers** build on the same connection: the recipient polls for mail under
an identifier it changes on a schedule, so even the recipient side is not one long-lived name.

## 6. What this cannot stop

- **Watching the network.** Someone who can observe traffic at both ends can match a send to a
  receive by timing. Only a mixnet with delays and cover traffic resists that, at a real cost in
  speed and battery.
- **A compromised phone.** Software on the device reads messages before they are encrypted.
- **An order to change the code.** Reproducible builds make that detectable, not impossible.
- **The live server during delivery**, until §5 is built.
