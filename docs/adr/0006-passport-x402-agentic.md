# ADR 0006 — Echo Passport Payments & Agentic Claims: adopt x402 over AllowSpend; GNAP, not "x401"

- **Status:** Accepted
- **Date:** 2026-06-12
- **Deciders:** Product / Platform / Identity / Payments
- **Related work order:** WO-315 (x402 verifier), WO-316 (agentic GNAP), WO-317 (x402 merchant rail);
  builds on WO-299 (consent), WO-307 (agentic), WO-301/302 (verifier/rails)
- **Blocks / unblocks:** WO-315/317 (Wave C), WO-316 (Wave D)
- **Supersedes:** N/A (extends ADR 0003 custody model and `docs/ECHO_PASSPORT_PLAN.md` Waves C/D)

## Context

Echo Passport (ADR 0003, `docs/ECHO_PASSPORT_PLAN.md`) already has the on-chain payment-consent
primitives: Tessellation v3 `AllowSpend` / `SpendTransaction` (`internal/metagraph/transactions.go`),
single-use and biometric-gated, plus a planned GNAP-style agentic flow (Wave D / WO-307). What it does
**not** have is a standard **HTTP-native payment wire protocol** for the two surfaces where Echo
charges or pays across a network boundary: the **verifier-pays-per-presentation** marketplace (Wave C)
and **agent / merchant payments** (Wave D). Those were specified as proprietary endpoints.

The founder asked whether to add the **x402** and **x401** specs and **agentic claims**.

Research (2026-06, see `docs/CROSS_PRODUCT_GAP_REVIEW.md`):
- **x402** is a real and growing open standard (Coinbase; Cloudflare + Solana "x402 Foundation"; ~69k
  active agents, ~165M transactions, ~$50M cumulative volume by April 2026). It revives the HTTP
  `402 Payment Required` status code: a server answers a request with `402` + payment requirements
  (CAIP-2 network id, price, accepted token); the client returns a signed payment payload in an HTTP
  header. It is a **wire/transport** standard for machine-to-machine and agent payments — not a
  custody or consent model.
- **x401** is **not** an established specification. The name reads as an intuition that `402` (payment)
  needs a `401` (authorization) companion for agents. That authorization role is real, but it is
  served by **GNAP** (already in Echo's plan) and emerging agent-authorization work (OAuth Agent
  Authorization Profile, AP2, NIST AI-agent identity initiative, Feb 2026), not by an "x401" spec.

## Decision

**Adopt x402 as the HTTP payment wire protocol for Passport's verifier and agent/merchant surfaces.
Keep `AllowSpend` / `SpendTransaction` as the on-chain consent + settlement layer. Formalize agentic
claims via a GNAP grant. Do not invent an "x401" spec — GNAP fills the authorization role.**

### How the layers compose

```
1. resource/verifier/agent  →  HTTP 402 Payment Required        (x402: CAIP-2 net, price, token)
2. client                   →  mint single-use, TTL-bounded,    (AllowSpend — biometric-gated)
                               amount-capped AllowSpend
3. client                   →  SpendTransaction on Currency L1   (on-chain settlement)
4. client                   →  return x402 payment proof header  (x402: completes the HTTP exchange)
```

- **x402 = how a charge is requested and a payment is proven over HTTP.** Echo speaks x402 on the
  wire so verifiers, agents, and (via a licensed rail) merchants in the broader ecosystem can
  transact with Passport without a bespoke integration.
- **AllowSpend = the consent + authorization that backs the x402 payment.** Every x402 payment is
  backed by a fresh, single-use, time-boxed, biometric-gated `AllowSpend`; settlement is a
  `SpendTransaction` on Currency L1. x402 does **not** replace or weaken this — it wraps it.
- **GNAP = the agent authorization grant (the "x401" role).** For agentic flows, a GNAP grant scopes
  *what* an agent may present and spend (least-privilege, revocable). Every agent-initiated payment
  still requires a fresh **per-action confirmation** (push to phone / glasses tap); there is no
  standing or unlimited approval path — ever. This is the guardrail from WO-307, now bound to x402.

### Agentic claims

An AI agent / smart glasses may **present a credential** (SD-JWT selective disclosure from WO-295) and
**initiate a payment** on the user's behalf, under a GNAP grant + per-action confirmation. The agent
never holds keys or standing authority. Implemented in `pkg/agent/consent/` (WO-316).

### Scope of delivery

- **WO-315 (Wave C):** x402 facilitator + verifier-pays-per-presentation over HTTP 402, bridged to
  `AllowSpend`/`SpendTransaction`; extends `pkg/passport/verifier/` and `/v1/verify/request`.
- **WO-316 (Wave D):** agentic claims — GNAP grant model in `pkg/agent/consent/`; agent/glasses
  presents a VC + initiates an x402-backed payment under per-action confirmation (formalizes WO-307).
- **WO-317 (Wave C/D):** x402 ↔ external merchant rail adapter (`pkg/payments/rails/`); Echo
  orchestrates consent and passes a presentation + **tokenized** instrument reference; the licensed
  partner moves the money. Never a raw PAN.

### Invariants preserved

- **No raw PAN / no PII on the wire.** x402 carries a payment proof + tokenized instrument reference,
  consistent with ADR 0003 ("verified Visa ending 1234", never the PAN). T0–T7 Semgrep gate applies.
- **Echo is never the money transmitter.** P2P stays native/self-custodied; merchant flows route
  through the licensed rail (WO-317). Money-transmitter / PISP licensing remains the existing deferred
  open item (ADR 0003; WO-302).

## Consequences

### Positive
- Ecosystem interop: Passport can transact with the growing x402 agent/merchant economy without
  bespoke integrations, while keeping Echo's stronger on-chain consent model underneath.
- Removes a proprietary HTTP payment surface that would otherwise have to be invented and maintained.
- Agentic claims become a concrete, scoped, revocable capability instead of prose.

### Negative
- New external dependency surface (x402 facilitator, CAIP-2 network handling, token/stablecoin
  settlement assumptions) to track as the standard evolves.
- Two payment framings to keep coherent (HTTP x402 wire vs on-chain AllowSpend); the bridge (WO-315)
  must make the mapping unambiguous and replay-safe.

### Neutral
- "x401" is explicitly **not** adopted; the authorization layer is GNAP. If a real 401-companion
  standard emerges later, it can be evaluated as an alternative to the GNAP grant without changing the
  AllowSpend settlement core.

## Alternatives considered

### Option B — Proprietary payments only (no x402)
Rejected. Keeping only AllowSpend + bespoke HTTP endpoints forgoes interop with the agent-payment
ecosystem and forces Echo to invent and maintain a wire protocol that x402 already standardizes.

### Option C — Invent an "x401" authorization spec to pair with x402
Rejected. No such standard exists; inventing one duplicates GNAP and the active OAuth/NIST agent-auth
work, and creates a non-interoperable surface. Use GNAP now; evaluate a real standard if one lands.

### Option D — Defer x402 to a later evaluation spike
Considered. Rejected in favor of deciding now, because x402 maps cleanly onto already-planned Wave C/D
surfaces and adopting it as transport (not custody) is low-risk and reversible at the wire layer.

## Implementation status

- [ ] WO-315 — x402 facilitator + verifier pays-per-presentation (Wave C)
- [ ] WO-316 — agentic claims: GNAP grant + x402-backed per-action-confirmed payment (Wave D)
- [ ] WO-317 — x402 ↔ external merchant rail adapter, tokenized instrument ref (Wave C/D)
