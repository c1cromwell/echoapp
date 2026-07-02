# Cross-Product Feature-Gap Review

> Status: **Accepted — 2026-06-12** (decisions). **Maturity tables refreshed 2026-05-29** — see
> [`ECHO_MESSAGING_LAUNCH_STATUS.md`](ECHO_MESSAGING_LAUNCH_STATUS.md) and
> [`PHASE4_7_GAP_AUDIT.md`](PHASE4_7_GAP_AUDIT.md) for current ship state.

## Context

The founder asked for a whole-repo feature-gap review across **Echo Protocol / Echo Message**
(messaging), **Echo Comply** (compliance), and **Echo Passport** (credential wallet), with three
explicit decisions:

1. Should Comply add a **web admin portal** for business, plus a **reporting / analytics / auditing**
   page (web + mobile) showing how a business is doing on the app?
2. Should Passport include **agentic claims** and add the **x402 / x401** specs?
3. "Honestly, any changes I should make?"

This document records the findings and the answers. It is the durable reference that the new ADRs
and work orders link back to.

---

## Per-product maturity

### Echo Messaging — MVP code complete (2026-05-29)

| Area | Status | Where |
|------|--------|-------|
| E2EE + SX1 ratchet | ✅ | `internal/crypto/kinnami.go`, WO-314 |
| 1:1 + groups + signals | ✅ | Relay, `ConversationSignalService`, Phase 3 WOs |
| PSI + contacts | ✅ | WO-220/221 |
| Search / archive (local) | ✅ | WO-3, 16, 29, 54 |
| Disappearing + sealed sender | ✅ | WO-38/75, WO-219 |
| Calls + voice notes | ✅ | WO-5/19, WO-316 |
| Wallet / staking | ✅ | `/v3/wallet/*`, `validate-wallet.sh` |
| **Message anchoring (Data L1)** | ✅ | WO-15 — `AnchoringService`, merkle-proof API |
| **Channels / large files** | backlog | Phase 6 WOs |
| **Device E2E / App Store** | go-live | WO-233, WO-238 |

**Read:** messaging MVP is code-complete; remaining work is go-live + WO-15 integrity pipeline.

### Echo Comply — service + web portal shipped; verticals backlog

| Area | Status | Where |
|------|--------|-------|
| Comply service (retention, hold, eDiscovery API) | ✅ | `internal/services/comply/`, WO-250–252 |
| Web admin portal + dashboards | ✅ | `web/`, WO-308–313 |
| eDiscovery / matter UI | 🔄 | WO-309 in progress |
| HIPAA / FOIA / law-firm / FINRA verticals | backlog | WO-253–307 |
| Org lifecycle (SSO, billing, org DID) | backlog | WO-281–286 |
| iOS Comply context coordinator | ✅ | WO-289 |

**Read:** operator surface exists; regulated vertical packs and enterprise org onboarding are the Comply program.

### Echo Passport — mid Wave A

| Area | Status | Where |
|------|--------|-------|
| Holder model + aggregation API | ✅ Wave A | `pkg/passport/service.go`, `internal/api/passport_handlers.go` |
| Client-encrypted credential sync | ✅ Wave A | `pkg/passport/sync.go`, `pkg/storage/encblob/` |
| Selective disclosure (SD-JWT) | ✅ Complete | `pkg/passport/disclosure/sdjwt.go` |
| Shamir M-of-N recovery | ✅ Wave A | `pkg/passport/recovery/` |
| AllowSpend / SpendTransaction primitives | ✅ | `internal/metagraph/transactions.go` |
| iOS Passport module | ❌ Backlog | WO-297 |
| Metagraph `passportCredentialRefs` state | ❌ Backlog | WO-298 |
| Pay-in-chat (P2P) | ❌ Backlog (Wave B) | WO-299/300 |
| Verifier marketplace + external rails | ❌ Backlog (Wave C) | WO-301/302 |
| **Agentic / GNAP presentation + pay** | ❌ Backlog (Wave D) | WO-307 |
| **x402 / x401** | ❌ No references anywhere | — |

**Read:** the holder core is real; agentic is designed but deferred to Wave D; there is no
HTTP-native payment interop layer today.

---

## The three decisions, answered

### 1. Comply web admin portal + reporting/analytics/auditing → **YES (web-primary)**

Comply's entire value (retention, litigation hold, eDiscovery, FOIA, matter management, audit
reporting) is regulated **desk work**. The backend primitives already exist; the missing layer is the
surface that lets a business **operate and see** them. Decision: build a **web-primary** admin portal
with a reporting/analytics/auditing dashboard, and a **responsive, read-mostly** mobile/iOS companion
of the dashboard (an exec can check posture on a phone; configuration stays on web). Recorded in
**ADR 0005**; delivered by **WO-309 … WO-314**.

### 2. Passport agentic claims → **YES, formalize via GNAP (the Wave D path)**

"Agentic claims" = an AI agent / smart glasses **presents a credential and initiates a payment on the
user's behalf**. Already designed as Wave D / WO-307 with the correct guardrail (per-action
confirmation; agent never holds standing authority). Decision: formalize now as **ADR 0006** + a
concrete `pkg/agent/consent/` GNAP grant model (**WO-316**), so Wave B/C build toward it instead of
leaving it as prose.

### 3. x402 / x401 → **Adopt x402; do NOT invent x401**

- **x402** is a real, growing standard (Coinbase + Cloudflare + Solana "x402 Foundation"; ~69k
  agents, ~$50M cumulative volume by Apr 2026) for HTTP-native agent/stablecoin payments using the
  `402 Payment Required` status code. It is the natural **wire protocol** for two Passport surfaces
  that are otherwise proprietary: **verifier-pays-per-presentation** (Wave C) and **agent/merchant
  payment** (Wave D). Decision: adopt it as transport **on top of** the existing on-chain `AllowSpend`
  consent — x402 is *how the charge is requested and paid over HTTP*; AllowSpend remains the
  single-use, biometric-gated *on-chain authorization*. Delivered by **WO-315** (verifier) and
  **WO-317** (merchant rail).
- **x401 is not a standard.** It reads as an intuition that 402 (payment) needs a 401 (authorization)
  companion. That role is real but is filled by **GNAP** (already in the plan) and emerging agent-auth
  work (OAuth AAP, AP2, NIST agent-identity). Decision: implement the authorization half as the GNAP
  grant from decision #2; **do not invent an "x401" spec.** Both calls recorded in **ADR 0006**.

**How they compose (x402 + AllowSpend + GNAP):**

```
verifier/agent → HTTP 402 Payment Required (x402, CAIP-2 network, price, token)
   client → mint single-use, TTL-bounded, biometric-gated AllowSpend
   client → SpendTransaction on Currency L1
   client → return x402 payment proof  (agent flow scoped by a GNAP grant,
                                         fresh per-action confirmation each payment)
```

---

## Other changes worth making

- **Finish messaging go-live before widening scope.** Device E2E, TestFlight, and App Store P0s (WO-233, WO-238) gate launch. Optional parallel: WO-15 message anchoring for integrity claims.
- **Comply verticals + org lifecycle** are the next product track (WO-309 → WO-281–286 → vertical wedge).
- **Passport iOS (WO-297)** then Wave B pay-in-chat (WO-298–300).
- **Privacy-preserving on-device AI (WO-CA1)** remains the strongest differentiator; keep it on the
  roadmap, gated behind messaging core + the bot framework (already sequenced in
  `COMPETITIVE_AUDIT_IMPLEMENTATION_PLAN.md`).

---

## Sequencing roadmap

1. **Messaging go-live.** Two-client E2E, TestFlight, App Store checklist — [`ECHO_MESSAGING_LAUNCH_STATUS.md`](ECHO_MESSAGING_LAUNCH_STATUS.md).
2. **Comply program.** Finish **WO-309** → org lifecycle **WO-281–286** → one vertical wedge (law firm or FINRA).
3. **Passport.** **WO-297** iOS module → Wave B **WO-298–300** (pay-in-chat) → x402/agentic per ADR 0006.

---

## Invariant guardrails (apply to every new surface)

- **Zero readable PII / no server-held key** (T0–T7, Semgrep). Comply dashboards render
  hashes/CIDs/aggregate metrics, never message content. Passport x402 carries payment proofs +
  **tokenized** instrument references, never a PAN or PII.
- **Per-action confirmation only** for agent payments — no standing/unlimited approvals, ever.

---

## Related artifacts

- `docs/adr/0005-comply-web-admin-portal.md` — Comply web portal decision
- `docs/adr/0006-passport-x402-agentic.md` — x402 + agentic + "no x401" decision
- `docs/phase-7-work-orders.md` — WO-309 … WO-314 (Comply), WO-316 (agentic)
- `docs/phase-5-work-orders.md` — WO-315, WO-317 (Passport x402)
- `docs/ECHO_PASSPORT_PLAN.md` — updated Wave C/D for x402
- `docs/COMPETITIVE_AUDIT_IMPLEMENTATION_PLAN.md` — messaging-core sequencing

## Sources (x402 / x401 research, 2026-06)

- Coinbase — Introducing x402: <https://www.coinbase.com/developer-platform/discover/launches/x402>
- x402.org standard: <https://www.x402.org/>
- Cloudflare — launching the x402 Foundation: <https://blog.cloudflare.com/x402/>
- Crossmint — agentic payments protocols compared (MPP/ACP/AP2/x402): <https://www.crossmint.com/learn/agentic-payments-protocols-compared>
- x401: no formal specification found — authorization role mapped to GNAP / emerging agent-auth.
