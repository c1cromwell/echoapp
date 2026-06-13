# Echo Comply — Web Admin Portal

WO-309 scaffold. Next.js (App Router) + Supabase. The repo's first web surface, per
[`docs/adr/0005-comply-web-admin-portal.md`](../docs/adr/0005-comply-web-admin-portal.md).

## Architecture (load-bearing)

- **Supabase** = portal **operator auth** (SSO/SAML/OIDC + SCIM) and the **operator ↔ org
  membership** mapping (`organizations`, `org_members`). Nothing else.
- **Go Comply backend** (`COMPLY_API_BASE_URL`) = **system of record** for all compliance data
  (retention, holds, evidence, audit). The portal calls it **server-side only**, scoped by
  `X-Org-DID`. No compliance payloads or PII are ever stored in Supabase or sent to the browser.
- **Zero-PII invariant** (T0–T7): dashboards render hashes/CIDs/aggregate metrics only.

```
browser ──(session cookie)──> Next.js server ──(Bearer + X-Org-DID)──> Go Comply API ──> metagraph
   │                              │
   └── Supabase Auth (SSO/SCIM) ──┘   (operator identity + org membership only)
```

## Layout

| Path | Purpose | WO |
|------|---------|-----|
| `src/middleware.ts` + `src/lib/supabase/middleware.ts` | Session refresh + route guard | WO-309 |
| `src/app/login`, `src/app/auth/callback` | SSO/OIDC + email sign-in | WO-309 |
| `src/lib/auth/org.ts` | Resolve operator's orgDID + portal role | WO-309 |
| `src/app/(portal)/layout.tsx` | Authenticated shell + org switcher | WO-309 |
| `src/app/(portal)/dashboard` | Reporting/analytics/auditing (placeholder) | WO-313 |
| `src/lib/api/comply.ts` | Server-only Comply API client (orgDID-scoped) | WO-309 |
| `src/app/api/scim/v2/*` | SCIM 2.0 provisioning (auth-gated stub) | WO-309 → WO-310 |
| `supabase/migrations/0001_portal_orgs.sql` | `organizations` + `org_members` + RLS | WO-309 |

## Run

```bash
cd web
cp .env.example .env.local      # fill in Supabase + Comply API values
npm install
npm run generate-api            # (after Comply endpoints land in ../openapi.yaml)
npm run dev                     # http://localhost:3000
```

Apply the Supabase schema with `supabase db push` (or paste `supabase/migrations/0001_portal_orgs.sql`
into the SQL editor). Configure SSO/SAML/OIDC providers in the Supabase Auth dashboard per tenant.

## What's next (this WO and after)

- **WO-309 finish:** real SCIM provisioning into `org_members`; org switcher state; sign-out;
  per-tenant SSO config doc; CI (lint + type-check + zero-PII payload test).
- **WO-310–312:** org/seat/role console, retention & litigation-hold config, eDiscovery/matters.
- **WO-313:** populate the dashboard once `/comply/dashboard` (WO-252) is live.
- **WO-314:** responsive read-mostly mobile companion.
