# Comply API client

`comply.ts` is the **server-only** client for the Go Comply backend, which is the system of
record for all compliance data. The portal is a thin client: it never stores compliance
payloads in Supabase and never calls this API from the browser.

## Generated types

The WO-309 scaffold ships hand-typed request/response shapes (e.g. `ComplyDashboardSummary`)
so the portal compiles before the backend endpoints exist. Once the WO-250+ Comply endpoints
are added to the root `openapi.yaml`:

```bash
npm run generate-api   # openapi-typescript ../openapi.yaml -> src/lib/api/schema.d.ts
```

Then replace the hand-typed shapes with `components["schemas"][...]` from `schema.d.ts`.

## Scoping & invariants

- Every call carries `X-Org-DID` (tenant scope) + a service `Authorization: Bearer` token.
- `cache: "no-store"` — compliance data is never statically cached.
- Responses are **zero-PII**: hashes, CIDs, and aggregate metrics only. If a payload would
  carry message content or PII, that is a backend bug, not something the portal renders.
