# Cross-Product Feature-Gap Review

> Status: **Accepted — 2026-06-12.** Whole-repo review across the three shipping products plus the
> Passport 4th product. Answers three founder decisions (Comply web portal, Passport agentic claims,
> x402/x401) and records the recommended changes. Execution artifacts: ADR 0005, ADR 0006, and
> WO-309 … WO-317 in the phase docs.

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

### Echo Messaging — ~25% feature-complete

| Area | Status | Where |
|------|--------|-------|
| E2EE (Kinnami: X25519 + ChaCha20-Poly1305) | ✅ Solid | `internal/crypto/kinnami.go`, `ios/.../KinnamiEncryption.swift` |
| 1:1 relay + offline queue + APNs | ✅ Works | `internal/services/relay/relay.go` |
| PSI contact discovery (OPRF) | ✅ Backend ~done; iOS client in flight | `internal/services/contacts/oprf.go` (WO-220/221) |
| did:key identity + trust tiers | ✅ ~80% | `pkg/didkey/`, Identity Metagraph |
| Reactions / typing / read receipts | ⚠️ Partial — backend fan-out incomplete | `internal/api/reactions_*`, WO-192/10 |
| Disappearing messages | ⚠️ UI only, backend not wired | `ios/.../ChatSettingsSheet.swift` |
| Edit / delete / forward / pin / reply | ⚠️ UI sketched, backend missing | WO-25/59/84 |
| **Group key distribution** | 🟨 Group create + encrypted group chat UI wired; server fan-out still partial | `ios/.../Groups/`, WO-207 |
| **Voice / video calls** | ❌ UI skeleton only; no WebRTC signaling | `ios/.../Calling/` |
| **Media / file relay** | 🟨 Local thread media + shared gallery from `ConversationThreadStore` | WO-237 |
| **Search / indexing** | ❌ Designed, not built | WO-3, WO-16 |
| **Multi-device message sync** | ❌ Not started | WO-CA3 |
| Voice notes | 🟩 DM + group capture wired (WO-194 / WO-SX5) | `VoiceNoteRecorder`, `GroupChatView` |

**Read:** the planning is thorough; the gap is **execution + sequencing**, not missing design.

### Echo Comply — audit backend only, no web surface

| Area | Status | Where |
|------|--------|-------|
| Encrypted audit trail → IPFS + Data L1 | ✅ | `internal/logging/` |
| Constellation Digital Evidence fingerprinting | ✅ | `internal/evidence/` |
| Auth audit logging | ✅ | `internal/auth/audit.go` |
| On-chain org-role credential model | ✅ Data model only | `EchoOrgRoleCredential` in `IdentityTypes.scala` |
| iOS Enterprise profile | ⚠️ Scaffold (`loadOrganization` TODO) | `ios/.../Enterprise/EnterpriseProfileView.swift` |
| iOS Evidence service | ⚠️ Mock | `ios/.../Evidence/EvidenceService.swift` |
| Retention / litigation hold / eDiscovery | ❌ Planned (Phase 7) | WO-250/251 |
| Compliance dashboard / reporting | ❌ Planned as iOS view | WO-252 |
| HIPAA / FOIA / law-firm logic | ❌ Planned | WO-253/254/262–264 |
| SSO / SAML / OIDC / SCIM | ❌ Planned | WO-285 |
| **Web / admin frontend** | ❌ **None anywhere in the repo** | — |

**Read:** Comply has the integrity/evidence backend but **no operator surface**, and what is planned
was scoped as iOS views. Comply's buyers (compliance officers, GC, IT admins) do desk work; an
iOS-only Comply is not enterprise-credible.

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

- **Finish messaging core before widening scope.** Group key distribution (WO-207), calls, media,
  search (WO-3/16), and multi-device sync (WO-CA3) are the foundation Comply and Passport both ride
  on. The highest near-term risk is scope-spreading onto a 25%-complete core.
- **There is no web client at all today.** The Comply portal is the repo's first web surface — stand
  up the shared web foundation (auth, design system, generated API client from `openapi.yaml`) so a
  future **web messaging client** (a real competitive gap vs Signal/WhatsApp/Telegram) can reuse it.
- **Privacy-preserving on-device AI (WO-CA1)** remains the strongest differentiator; keep it on the
  roadmap, gated behind messaging core + the bot framework (already sequenced in
  `COMPETITIVE_AUDIT_IMPLEMENTATION_PLAN.md`).

---

## Sequencing roadmap

1. **Messaging core first (gates everything).** Group key distribution (WO-207), calls, media,
   search (WO-3/16), multi-device sync (WO-CA3) — per
   `COMPETITIVE_AUDIT_IMPLEMENTATION_PLAN.md` Waves 0–1.
2. **Comply reporting wedge, in parallel (low coupling).** Ship **WO-309** (web shell) + **WO-313**
   (read-mostly reporting/analytics/auditing dashboard) first as the wedge, then **WO-310–312/314**
   (admin console, retention/hold config, eDiscovery, mobile companion).
3. **Passport x402 / agentic, after pay-in-chat (Wave B).** **WO-315** (x402 verifier) →
   **WO-316** (agentic GNAP) / **WO-317** (x402 merchant rail).

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
