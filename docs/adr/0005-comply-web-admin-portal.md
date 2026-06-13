# ADR 0005 — Echo Comply Web Admin Portal: web-primary operator surface, read-mostly mobile companion

- **Status:** Accepted
- **Date:** 2026-06-12
- **Deciders:** Product / Platform / Compliance
- **Related work order:** WO-309 … WO-314 (Comply web portal); builds on WO-250 … WO-289 (Comply backend)
- **Blocks / unblocks:** WO-313 (reporting dashboard — the wedge), WO-310/311/312 (admin console)
- **Supersedes:** N/A (refines the Phase 7 Comply plan, which assumed iOS-only surfaces)

## Context

Echo Comply (`docs/Echo_Combined_Requirements.md`, Phase 7 WO-250–289) targets regulated businesses —
healthcare (HIPAA), local government (FOIA), and law firms (chain-of-custody). The integrity backend
already exists: encrypted audit trail to IPFS + Data L1 (`internal/logging/`), Constellation Digital
Evidence fingerprinting (`internal/evidence/`), auth audit (`internal/auth/audit.go`), and the
on-chain `EchoOrgRoleCredential` model (`metagraph/.../IdentityTypes.scala`).

What is missing is the **operator surface**. Today the only Comply UI is an iOS `EnterpriseProfileView`
scaffold (`loadOrganization` is a TODO) and a mock `EvidenceService`. The full Phase 7 feature set
(retention policies, litigation hold, eDiscovery export, matter/ethical-wall/privilege management,
compliance dashboard, SSO/SAML/SCIM) was scoped as **iOS views**.

That is the wrong surface. Comply's actual users — compliance officers, general counsel, IT admins,
records officers — do this work at a desk: configuring retention policies, placing litigation holds,
running eDiscovery exports, reviewing audit reports for FOIA/court submission, provisioning seats. An
iOS-only Comply is not credible to an enterprise buyer and does not match the job. See
`docs/CROSS_PRODUCT_GAP_REVIEW.md`.

## Decision

**Echo Comply ships a web-primary admin + reporting/analytics/auditing portal as a first-class
product surface. Mobile (responsive web + iOS) is a read-mostly companion of the dashboard;
configuration and privileged operations live on web.**

The portal is a **thin client over the existing/planned Comply REST APIs** — it does not duplicate
compliance logic, retention enforcement, or evidence generation, all of which stay in the Go backend
and on-chain. All data is scoped by `orgDID` (multi-tenant).

### Scope (delivered by WO-309 … WO-314)

- **Portal shell + auth (WO-309):** the repo's first web build target; SSO/SAML/OIDC sign-in; SCIM
  provisioning; a TypeScript API client generated from `openapi.yaml`.
- **Org / seat / role admin (WO-310):** organization lifecycle, seat enforcement, role assignment over
  `EchoOrgRoleCredential` (owner/admin/moderator/member) + the Phase 7 org APIs.
- **Retention & litigation-hold config (WO-311):** UI over WO-250/251 (permanent / time-limited /
  litigation-hold policies; custodian scoping).
- **eDiscovery + matter management (WO-312):** export requests, matter creation, ethical-wall conflict
  view, attorney-client-privilege designation (wires WO-251/262–264).
- **Reporting / analytics / auditing dashboard (WO-313):** the wedge. Digital-Evidence coverage rate,
  active retention-policy count, litigation-hold count, pending-export count, metagraph anchor health,
  segment reports (HIPAA ePHI access / FOIA request tracking / law-firm privilege logs), and audit
  reports exportable as PDF/JSON.
- **Mobile companion (WO-314):** responsive read-mostly rendering of the WO-313 dashboard for iOS /
  small screens.

### Zero-PII invariant (non-negotiable)

The portal renders only **hashes, CIDs, and aggregate metrics** — never message content or readable
PII. Reporting payloads are classified and must pass the T0–T7 Semgrep gate
(`docs/data-classification.md`). The dashboard's "how the business is doing" view is operational and
compliance posture (coverage, holds, exports, anchor health, deadlines), **not** message analytics.

## Consequences

### Positive
- Comply becomes operable and demonstrable to enterprise buyers; matches the actual compliance job.
- The web shell (WO-309: auth, design system, generated API client) is reusable foundation for a
  future **web messaging client** — the repo gains its first web surface deliberately.
- Reuses the existing evidence/audit backend and on-chain org model; the portal is additive plumbing.

### Negative
- Introduces a new build target / deploy surface and a web stack choice the repo did not previously
  have (note existing `package.json` / `tsconfig.json` / `openapi.yaml` for codegen).
- Web is a new attack surface for a compliance product; auth (SSO/SCIM) and the zero-PII gate must be
  hard requirements, not follow-ups.

### Neutral
- Phase 7 Comply work orders that assumed iOS dashboard views (WO-252) are reframed: the dashboard is
  web-primary (WO-313) with an iOS companion (WO-314); the backend APIs are unchanged.

## Alternatives considered

### Option B — Keep Comply iOS-only (status quo plan)
Rejected. Compliance configuration and eDiscovery are desk work; an iOS-only console is not credible
to the buyer and is awkward for the task. Mobile remains valuable as a read-mostly posture view, not
as the primary surface.

### Option C — Reporting dashboard only, no admin console
Rejected as the end state, but **adopted as the first wave**: WO-313 (read-mostly dashboard) is the
low-coupling wedge that ships in parallel with messaging core, ahead of the heavier admin console
(WO-310–312). The full portal is still the destination.

## Implementation status

- [ ] WO-309 — web portal shell + SSO/SAML/OIDC + SCIM + generated TS API client
- [ ] WO-310 — org / seat / role admin console
- [ ] WO-311 — retention-policy & litigation-hold configuration UI
- [ ] WO-312 — eDiscovery export + matter / ethical-wall / privilege management UI
- [ ] WO-313 — reporting / analytics / auditing dashboard (the wedge)
- [ ] WO-314 — responsive read-mostly mobile/iOS companion
