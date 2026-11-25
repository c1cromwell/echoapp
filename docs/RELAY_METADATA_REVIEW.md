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

Before this work: **yes, for every message, for 30 days**, including messages sent with "sealed
sender". Now, for a 1:1 message between two people who have already exchanged one message each
way: **no — not from the database, the logs, or the message as it passes.** The message arrives on
a connection with no login, addressed by a key only the recipient's contacts hold, carrying no
sender and no conversation (§5).

Three things qualify that, and they matter:

- **It is built and unit-tested, not yet proven between two phones.** §5.4 is the test to run.
- **The first message to a new contact still shows the server who sent it**, because the sender
  has no delivery key for them yet. Every message after the reply does not.
- **Typing indicators, read receipts, reactions, edits and calls still travel on the logged-in
  socket**, naming both people to the live server (not stored). And the sealed request comes from
  the same network address as the logged-in socket, so a server set on it could match the two.
  §4.2.

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
| 8 | The sender of every message was kept in two in-memory maps that were never emptied; with no chain configured one of them grew without limit | `internal/ledger/anchoring.go`, `internal/services/relay` |
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
| 9 | Pushes carry no conversation id: only a type, and for a missed call that call's random id | `TestAPNsPayloadCarriesNoConversationOrParty` |
| 10 | Without the real private-matching library, contact discovery is off. The test stand-in can no longer be reached from the app | compiles in the app build; no device test |
| 11 | A forward that cannot be encrypted is not sent | compiles in the app build; no device test |
| 12 | The app sends the socket token in the `Authorization` header. The server already accepted it there | `WebSocketURLBuilderTests` |
| 13 | The three log lines no longer contain a DID, username or device label | code |
| 15 (server half) | A phone number registered for recovery is stored as HMAC-SHA256 under a secret the database does not hold (`PHONE_HASH_SECRET`, required in production), not as a bare hash. This protects against a copy of the database; it does not protect against someone holding the database and the secret. The raw number still has to reach the server to send the SMS | `TestPhoneStoreKey` |
| 15 (app half) | The SMS recovery screen now says the number goes to the text-message provider once and that only a scrambled form is kept. Text only; the frozen design is unchanged | read |
| 17 | A link in a message someone else sent is shown as its domain only; the linked site is not contacted unless the user opts in. Previews for the user's own links use a session with no cookies that honours the proxy setting | `LinkPreviewPolicyTests` |

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

### 4.1 Closed by §5, pending the two-phone test

The conversation id on the wire (finding 1) and the server seeing the sender of each message
(finding 4), for 1:1 text messages after first contact.

### 4.2 What sealed delivery does not cover

| What | Why it is still visible | What would close it |
|---|---|---|
| The first message to a new contact | No delivery key for them yet, so it goes over the logged-in socket (sealed for the recipient, with the constant conversation id) | A way to fetch a contact's delivery key privately, e.g. inside the contact-exchange QR or invite link |
| Typing, read receipts, reactions, edits, deletes, pins, polls, call set-up | They are sent on the logged-in socket with `from` and `to`. Not stored, but the live server sees the pair | Send them through sealed delivery too. They need a signature inside, because the server no longer vouches for the sender |
| Group messages | Group fan-out is by the server, which knows the membership | A different design (per-member sealed delivery); not started |
| Network address | The sealed request and the logged-in socket come from the same address | A relay or Tor in front of sealed delivery. The app's SOCKS setting already applies to it |
| Timing | A send and a delivery a moment apart can be matched by anyone watching both | Delays and cover traffic (§6) |
| Anchoring confirmation | A sealed message is anchored with no sender recorded, so the server has no one to notify | The app fetches the proof by message id when it wants it |

### 4.3 The rest

| Finding | What is needed | Size |
|---|---|---|
| 14 | Stop storing the actor's DID with reactions, edits, pins and references, or key them by a per-message pseudonym; add expiry | Medium: the features read these columns |
| 16 | Move ratchet state and the pre-key private key into the Keychain, with a migration | Medium; needs device testing |
| 18 | Route every request through one proxy-aware session so the Tor/SOCKS option covers them all; add an onion address for the gateway (deploy stage) | Medium |
| 19 | Add the notification extension to the app target so the device, not the server, decides what a notification says | Xcode project change |
| 20 | Configure certificate pins | Small, but a wrong pin locks every user out; do it with a rollback plan |
| 21 | Decide how Comply obtains records without the consumer gateway keeping them | Product decision |
| — | In-memory rate-limit buckets keyed by DID or IP are never evicted *(reported)* | Small |
| — | `GET /v3/messages/{id}/status` has no check that the caller is a party to the message. Since the server no longer records who sent a message, it cannot make that check; the random message id, known only to the two devices, is the permission. The response no longer has a conversation to reveal | By design |

## 5. Sealed delivery and opaque conversation ids (built 2026-11-24)

### 5.1 How it works

1. **Delivery key.** Each device makes 32 random bytes and keeps them in the Keychain. It tells
   the server only the SHA-256 (`PUT /v3/messages/delivery-key`). The server stores
   hash → account, the two most recent per account.
2. **Handing it out.** The key travels inside every sealed message, within the encryption, so the
   recipient can reply. The receiving device starts using it only after the message it came in
   has been authenticated as from that person; otherwise anyone could redirect replies.
3. **Sending.** With the recipient's key, the app posts to `POST /v1/sealed/deliver` on a
   separate connection with no login, no cookie and no device header. The server hashes the key,
   finds the recipient, and delivers. It records nothing about the caller.
4. **No conversation on the wire.** A sealed message carries the constant `dm:sealed`. The
   recipient's device reads the sender from inside the encryption and works out the thread
   itself. `dm:<A>:<B>` is now only each device's private name for a thread.
5. **First contact.** With no key yet, the message goes over the logged-in socket, sealed for the
   recipient and carrying the sender's key. The reply comes back sealed, and from then on both
   directions are.
6. **A replaced key.** If the server answers "unknown key", the app forgets that key and uses the
   first-contact path once; the reply brings the new one.

### 5.2 Abuse and blocking

The server cannot limit or block a sender it cannot see, so:

- The rate limit is per recipient mailbox.
- A delivery key is a permission. Only people the user has messaged hold it.
- Blocking someone drops their messages on the device, and replaces the delivery key with every
  earlier key revoked at once, so the copy they hold stops working. Other contacts get "unknown
  key", use the first-contact path once, and pick up the new key.
- The old server-side block still applies to everything sent over the logged-in socket.

### 5.3 Verified by

| What | Test |
|---|---|
| A message reaches the recipient with no sender and no conversation in the envelope or the stored row | `TestSealedDeliver_ReachesRecipientWithoutNamingTheSender` |
| A login, header or context identity presented by the caller never reaches storage | `TestSealedDeliver_IgnoresAnyIdentityTheCallerPresents` |
| Works through the real middleware with no token; registering a key still needs a login | `TestSealedDeliver_ThroughTheRealHandlerChain` |
| Unknown key, short key, missing or oversized payload are refused and deliver nothing | `TestSealedDeliver_Refusals` |
| Another account cannot register someone's key; rotation keeps two; revoke-others keeps one | `TestDeliveryKeyRegistration`, `TestDeliveryKeyRevokeOthers`, `TestDeliveryKeysPostgresIntegration` (real Postgres) |
| A sealed message is anchored with no sender | `TestSealedDeliver_AnchorsWithoutASender` |
| The app's request contains no DID, token or conversation, and only a content-type header | `SealedDeliveryTests.testDeliveryRequestCarriesNoIdentity` |
| The recipient derives the same thread name the sender uses locally | `testRecipientDerivesTheThreadLocally` |
| A key from a message is unused until that message is authenticated | `testPeerKeyIsUsedOnlyAfterTheMessageIsAuthenticated` |
| The app's key hash equals the server's | `testKeyHashMatchesTheServer` |

### 5.4 Not verified: two phones

Nothing above runs two real apps against each other. Before relying on this, with two devices
(or one device and a simulator) signed in as A and B against a test backend:

1. A sends B a first message. B receives it in the right thread. (First-contact path.)
2. B replies. A receives it in the same thread. In the gateway's request log the reply is a
   `POST /v1/sealed/deliver` with no user, and no `sealed_text` frame from B appears on B's socket.
3. A sends again. It is also a `POST /v1/sealed/deliver`.
4. Put B in airplane mode, send from A, bring B back: the message arrives, and the
   `message_queue` row for it had an empty sender and conversation while it waited.
5. B blocks A. A's next message does not appear on B. B unblocks; messages flow again after one
   exchange.
6. Delete and reinstall B (new delivery key). A's next message still arrives (unknown-key
   fallback), and the one after that is sealed again.
7. Replies, forwards and media each arrive in the right thread, and delivery ticks still update.

If step 2 or 3 shows a `sealed_text` frame on the socket instead, the key exchange did not
complete; the message was still delivered, but not privately.

## 6. What this cannot stop

- **Watching the network.** Someone who can observe traffic at both ends can match a send to a
  receive by timing. Only a mixnet with delays and cover traffic resists that, at a real cost in
  speed and battery.
- **A compromised phone.** Software on the device reads messages before they are encrypted.
- **An order to change the code.** Reproducible builds make that detectable, not impossible.
- **The live server for what §4.2 lists**: first contact, signals, groups.
