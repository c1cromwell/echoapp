# Competitive Audit — SimpleX Chat

**Date:** 2026-06-26
**Scope:** Review SimpleX Chat (https://github.com/simplex-chat/simplex-chat) and identify features
ECHO messaging should add, filtered through ECHO's privacy-first, decentralized, `did:key` +
trust-tier + tokenomics thesis.
**Relationship to other docs:** complements the prior competitor audit
([`COMPETITIVE_AUDIT_2026-05.md`](COMPETITIVE_AUDIT_2026-05.md), which covered
WhatsApp/Telegram/Signal/XChat) and the messaging security review
([`MESSAGING_SR_REVIEW_2026-06.md`](MESSAGING_SR_REVIEW_2026-06.md)). SimpleX was the most important
omission from the first audit because it is the **purest privacy-first messenger** and sits at the
**opposite pole from ECHO on identity** — which is exactly what makes it worth studying.

> **Grounding note.** Every "ECHO already has / lacks" claim below is checked against current code, not
> the older roadmap. Several things assumed missing are in fact shipped (media transfer, groups,
> reactions, edits, pins, disappearing messages, a partial sealed-sender). The recommendations are the
> *remaining* genuine gaps.

---

## 1. SimpleX snapshot (2025–2026)

SimpleX's defining choice: **no user identifiers at all** — not a phone number, not a username, not
even a random account ID. Identity-free by construction.

- **Unidirectional pairwise queues (SMP).** Each connection uses two simplex (one-way) message
  queues, each with separate addresses for sender and recipient (+ an optional 3rd for iOS push). The
  effect SimpleX advertises: *"no metadata in common between your conversations with different
  contacts"* — a server can't reconstruct your social graph because there's no shared identifier
  across queues.
- **Disposable, non-cooperating relays.** SMP relays hold messages only transiently until retrieval,
  keep no user records, and **don't talk to each other**. Anyone can self-host an SMP/XFTP server and
  stay interoperable.
- **Private message routing (v6.0+, default).** A 2-hop routing scheme so the recipient's relay
  doesn't learn the sender's IP and the sender's relay doesn't learn the destination — transport-level
  anonymity per message. **Tor** is supported for all client↔server connections.
- **XFTP** — a separate end-to-end-encrypted, chunked file-transfer protocol (independent of the
  message relays).
- **Double Ratchet encryption** (Curve448) for forward secrecy + break-in recovery, with
  **post-quantum-hybrid key agreement on every ratchet step**. Layered with NaCl cryptobox at the
  queue level + extra server-to-recipient encryption; padding at multiple layers to defeat traffic
  analysis.
- **Connection model:** one-time **invitation links / QR codes**, plus optional long-term **contact
  addresses**; out-of-band verification of the contact.
- **Incognito mode:** a fresh random profile name per contact.
- **Decentralized groups:** only client devices know membership/messages; no group server. Join via
  group links/QR. Recent work targets large groups/communities and battery use.
- **Local data:** on-device DB + file encryption with a user passphrase; encrypted export/import.
  Multiple isolated profiles per install with transport isolation.
- **In progress (their roadmap):** short connection links, queue redundancy/rotation, broadcasts/feeds,
  programmable chat automations, and a *privacy-preserving identity server*.

**vs Signal:** Signal still uses the phone number as a global identifier (graph visible to Signal's
servers) and is centralized; SimpleX removes the identifier and allows self-hosted/federated relays.

---

## 2. The core tension — identity-erasing vs identity-forward

SimpleX and ECHO optimize **opposite** variables, and that's the whole reason this audit is useful:

| | **SimpleX** | **ECHO** |
|---|---|---|
| Identity | **None** — no identifiers, ever | **Verifiable** — `did:key`, `@username`, trust tiers, credentials |
| Core promise | **Metadata anonymity** (no social graph) | **Verifiable trust** (anti-scam, reputation, payments) |
| Anti-abuse | Hard (no identity to reason about) | Native (trust tier + on-chain evidence) |
| Economics | None | ECHO token, rewards, fee-split, in-chat pay |
| Transport | Disposable, federated, identity-free queues | Content-blind but **identity-routed** central relay |

**Implication for what we adopt:** ECHO should take SimpleX's **mechanisms** (forward secrecy,
metadata minimization, private routing, federation) — which are pure wins for *any* privacy product —
and **not** its **philosophy** (no-identifier routing), which would dismantle ECHO's differentiators
(trust tiers, verification, anti-scam, tokenomics). The audit leads with the crypto + metadata items
because they harden ECHO with **zero conflict** with the identity thesis, and explicitly fences off
the identity-erasing parts in §5.

---

## 3. Where ECHO already matches or leads (defend + surface — don't rebuild)

Verified in code:

- **Content-blind relay.** The server transports E2E payloads it can't read
  (`internal/services/relay/relay.go`, `internal/api/ws.go`).
- **Encrypted media pipeline** — chunked (256 KB), client-encrypted, content-addressed, trust-tier
  size gates, optional Merkle/Data-L1 anchoring (`internal/services/media/service.go`). This is ECHO's
  analog to SimpleX's XFTP and is **already done** — separate from the message relay, as SimpleX does.
- **Groups** with per-member encrypted key packages (`GroupKeySignal` in `internal/api/ws.go`;
  `migrations/022_groups.sql`).
- **Reactions, edit, delete (tombstones), pins, read receipts, disappearing messages** — durable,
  retention-aware (`migrations/018_message_ops.sql`, `migrations/011_reactions_discovery.sql`).
- **Partial sealed-sender** (`WO-219`): single-use, 5-minute delivery tokens strip the `From` field at
  delivery (`internal/services/messaging/sealed_sender.go`, `internal/api/ws.go`). *Caveat:* the relay
  still validates the token against the sender DID, so it's **masking, not architectural anonymity.**
- **did:key portable identity + trust tiers + integrity anchoring** — things SimpleX deliberately does
  not offer. This is ECHO's moat; surface it, don't trade it away.
- **Connection via QR / invite links** already exists (`echo://invite`, WO-222) — parity with
  SimpleX's invitation model, *plus* a verifiable identity behind it.

---

## 4. The real gaps & recommended additions (prioritized, on-thesis)

### Tier 1 — Cryptographic hardening (highest value, zero thesis conflict)

**A. Double Ratchet forward secrecy over Kinnami.**
Today `internal/crypto/kinnami.go` does a fresh *ephemeral* ECDH (P-256) **against the recipient's
static long-term key**, then AES-256-GCM. That gives per-message ephemerality on the sender side but
**no ratchet, no key evolution, and no post-compromise recovery** — compromising one static private
key retroactively decrypts *every* message to that user. SimpleX (like Signal) uses a **Double
Ratchet**: a DH ratchet + symmetric-key ratchet so keys evolve per message and a key compromise
self-heals ("break-in recovery").

- *Recommendation:* introduce a Double-Ratchet session layer over Kinnami's primitives — X3DH-style
  initial agreement (using the existing did:key/static keys as the identity keys), then a
  per-message ratchet. This is the **single highest-value** item: it's a real cryptographic gap in a
  product whose entire pitch is privacy.
- *Sequencing:* do this **together with** the open **device-decrypt fix** in
  `MESSAGING_SR_REVIEW_2026-06.md` — both touch the same key-management surface, and shipping a ratchet
  on top of broken device decryption would be wasted work.

**B. Post-quantum-hybrid key exchange.**
Neither ECHO nor SimpleX is "post-quantum" historically, but SimpleX now runs **PQ-hybrid key
agreement on every ratchet step**. ECHO's roadmap already names "post-quantum" as a Phase 7 ambition.

- *Recommendation:* add an **ML-KEM (Kyber) + X25519 hybrid** as an option in the new ratchet's DH
  step (hybrid, never PQ-only, to avoid betting on one primitive). Lands naturally once (A) exists.

### Tier 1 — Metadata privacy (turn "content-blind" into "metadata-blind")

**C. Harden sealed-sender toward default + minimize persisted routing metadata.**
The relay persists `sender_did → recipient_did` in `message_queue` (`migrations/002_messaging.sql`),
so the **social graph is reconstructable server-side** — the exact thing SimpleX's pairwise queues
prevent. ECHO can't go fully identifier-free (it needs identity for trust/APNs/anti-scam), but it can
move a long way:

- Make **sealed-sender the default** for trusted contacts rather than an opt-in token flow.
- **Minimize retention** of sender→recipient tuples (store only what offline delivery needs; drop the
  sender side as soon as delivered; avoid long-lived graph rows).
- Add **per-contact queue aliases / rotation** (a SimpleX idea that fits ECHO): route to a rotating
  per-conversation alias instead of the raw recipient DID where APNs allows, so the stored metadata
  isn't a stable global identifier.

**D. Private message routing / optional Tor.**
ECHO clients connect **directly over TLS to the central relay**, so the relay sees the client IP and
there is no transport anonymity. SimpleX defaults to 2-hop private routing + optional Tor.

- *Recommendation:* add **optional Tor / SOCKS transport** first (cheap, high-trust-tier-independent
  win), then a **2-hop private-routing** mode as relay federation (Tier 3) matures. Directly supports
  the **sovereign/edge-relay** direction in `NETWORK_STATE_OWNERSHIP_THESIS.md`.

### Tier 2 — Parity / quick wins

- **Voice messages** (recording) — present in SimpleX, **absent** in ECHO (ECHO has voice *calls* via
  WebRTC, but no recorded voice notes). Small, high-visibility.
- **Per-contact minimal-disclosure mode** — the on-thesis analog of SimpleX "incognito": instead of a
  *random* profile, present a **reduced-disclosure ECHO persona** per contact (ECHO already has
  personas/hidden personas). Keeps verifiability optional while giving privacy-conscious users
  per-contact separation.
- **Short connection links + queue rotation** — match SimpleX's shorter links and let users rotate a
  conversation's routing association without losing history.

### Tier 3 — Architecture bet

- **Relay federation / self-hosting.** ECHO is single-operator today; SimpleX's self-hostable,
  non-cooperating relay network is external validation of the model ECHO already calls its
  **load-bearing milestone** (`NETWORK_STATE_OWNERSHIP_THESIS.md`: community relay kit → federation).
  *Recommendation:* keep it where the roadmap has it (Phase 4+), but adopt SimpleX's design principles —
  **relays don't gossip, hold messages transiently, and are interoperable via an open relay protocol** —
  so federation doesn't reintroduce a server-side social graph or break integrity anchoring.

---

## 5. Explicitly DON'T adopt (conflicts with ECHO's thesis)

- **Full no-identifier routing.** Eliminating identity would kill trust tiers, verification, anti-scam
  evidence, payments, and tokenomics — i.e. everything that differentiates ECHO. Adopt metadata
  *minimization* (§4C), not identity *elimination*.
- **Random-only incognito as the model.** Use reduced-disclosure personas (§Tier 2) so privacy doesn't
  require throwing away verifiability.
- **Server-to-server gossip / shared membership state** in any future federation — it would
  reintroduce the social graph SimpleX avoids and could undercut ECHO's integrity-anchoring guarantees.

---

## 6. Roadmap integration (proposed work orders)

> Provisional IDs (`WO-SX*`) — final numbers assigned by Software Factory, mirroring how `WO-CA1…CA4`
> were stubbed by [`COMPETITIVE_AUDIT_2026-05.md`](COMPETITIVE_AUDIT_2026-05.md). Each should be
> mirrored as a "SimpleX Audit Additions" note in its phase doc.

| ID | SF WO | Phase | Title | Status (2026-05-29) |
|----|-------|-------|-------|---------------------|
| **WO-SX1** | **WO-314** | 3 | **Double Ratchet forward secrecy** over Kinnami | ✅ Completed (`e7797dc`, `3cfcc2d`) |
| **WO-SX2** | **WO-315** | 7 | **PQ-hybrid ratchet bootstrap** | ✅ Completed (bootstrap; per-step PQ future) |
| **WO-SX3** | **WO-317** | 5 | **Metadata minimization + sealed-sender default** | ✅ Completed (`e15e5e8`) |
| **WO-SX4** | **WO-319** | 5 | **Private routing / optional Tor transport** | ✅ Completed (SOCKS + settings UI; 2-hop routing backlog) |
| **WO-SX5** | **WO-316** | 6 | **Voice messages** (recorded notes) | ✅ Completed (`0259024`) |
| **WO-SX6** | **WO-318** | 5 | **Per-contact minimal-disclosure persona** | ✅ Completed (`9fcc7af`) |
| **WO-SX7** | **WO-320** | 4 | **Federation design principles for the relay kit** | 📋 Backlog |

**Sequencing / dependencies:**
- **WO-SX1 is the keystone** and is gated by the device-decrypt fix — do them as one effort.
- **WO-SX2** depends on SX1 (it's a step inside the new ratchet).
- **WO-SX3/SX4** (metadata + routing) are independent of the crypto track and can run in parallel in
  the Privacy phase; SX3 builds on the existing sealed-sender (WO-219).
- **WO-SX7** rides the existing federation milestone (Phase 4+) — adopt principles now so federation
  doesn't reintroduce a social graph.
- **No overlap with WO-CA1…CA4** (private AI, backups, multi-device sync, in-chat economy) — these are
  orthogonal privacy/crypto items.

---

## 7. Sources

- SimpleX Chat repository & README: <https://github.com/simplex-chat/simplex-chat>
- SimpleX whitepaper / protocol overview (SMP unidirectional queues, double ratchet, private routing):
  <https://simplex.chat/docs/protocol/> and <https://simplex.chat/blog/>
- SimpleX post-quantum double-ratchet: SimpleX blog, "post-quantum resistant" announcement.
- ECHO grounding (current code): `internal/services/relay/relay.go`, `internal/api/ws.go`,
  `internal/crypto/kinnami.go`, `internal/services/media/service.go`, `internal/services/messaging/sealed_sender.go`,
  `migrations/002_messaging.sql`, `migrations/018_message_ops.sql`, `migrations/022_groups.sql`.
- Related ECHO docs: [`COMPETITIVE_AUDIT_2026-05.md`](COMPETITIVE_AUDIT_2026-05.md),
  [`MESSAGING_SR_REVIEW_2026-06.md`](MESSAGING_SR_REVIEW_2026-06.md),
  [`NETWORK_STATE_OWNERSHIP_THESIS.md`](NETWORK_STATE_OWNERSHIP_THESIS.md).

---

## 8. Software Factory sync (2026-05-29)

Provisional `WO-SX*` IDs are now tracked in Software Factory. **Last synced:** 2026-05-29.

| Provisional | Software Factory | Status |
|-------------|------------------|--------|
| WO-SX1 | WO-314 | completed |
| WO-SX2 | WO-315 (child of WO-314) | completed |
| WO-SX3 | WO-317 | completed |
| WO-SX4 | WO-319 | completed |
| WO-SX5 | WO-316 | completed |
| WO-SX6 | WO-318 | completed |
| WO-SX7 | WO-320 | backlog |

**Commits on `main`:** `e7797dc` SX1 Go · `3cfcc2d` SX1 iOS · `3bc5e4a` SX2 Go PQ · `bed4e76` SX2 iOS hook · `e15e5e8` SX3 · `0259024` SX5 · `9fcc7af` SX6 · `671729f` SX4 transport · *(this push)* SX2 iOS ML-KEM + SX4 Privacy Hub UI.

**Out of scope (tracked on WO-315 / WO-319 descriptions):** per-ratchet-step PQ; 2-hop private routing; proxy on ancillary `URLSession.shared` clients.
