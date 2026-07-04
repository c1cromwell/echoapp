# Echo — three products, three launches

Echo ships as **three independently releasable products** from one monorepo. Each product has its own bundle ID / App Store listing (iOS), deploy surface (web or backend), and release tag train.

| Product | Customer | iOS | Backend / web | Release tag |
|---------|----------|-----|---------------|-------------|
| **Echo Messaging** | Consumers & teams | `com.echo.app` — scheme `EchoMessaging` | Gateway `:8000` + full stack (`make dev`) | `echo-messaging@v1.0.0` |
| **Echo Comply** | Enterprise compliance operators | `com.echo.comply` — scheme `EchoComplyCompanion` | Comply API `:8011` + Next.js portal | `comply-portal@v1.0.0`, `echo-comply-ios@v1.0.0` |
| **Echo Passport** | Credential wallet users | `com.echo.passport` — scheme `EchoPassport` | Gateway passport routes (`/credentials`, OIDC4VC) | `echo-passport@v1.0.0` |

Optional SDK-only drops (embedded passport in messaging) may use `passport-sdk@v*`.

---

## iOS — build flavors

All three apps share one Xcode target (`EchoApp`) with **product build configurations**:

| Scheme | Configuration | Bundle ID | Display name |
|--------|---------------|-----------|--------------|
| `EchoMessaging` | `Debug-Messaging` / `Release-Messaging` | `com.echo.app` | Echo |
| `EchoComplyCompanion` | `Debug-Comply` / `Release-Comply` | `com.echo.comply` | Echo Comply |
| `EchoPassport` | `Debug-Passport` / `Release-Passport` | `com.echo.passport` | Echo Passport |

Entry routing is compile-time via `ECHO_PRODUCT_*` flags in `Sources/App/EchoApp.swift`.

### Local run (Xcode)

1. Open `ios/Echo/EchoApp.xcodeproj`.
2. Select the product scheme (e.g. **Echo Comply Companion**).
3. Set run env vars on the scheme if needed:
   - Messaging / Passport: `API_URL=http://localhost:8000`
   - Comply companion: `API_URL=http://localhost:8000` (signed REST proxy) and optionally `COMPLY_API_URL=http://localhost:8011`
4. Archive with **Product → Archive** (uses `Release-*` configuration).

### CI

- **SPM library + tests:** `.github/workflows/ios-ci.yml`
- **Per-product Xcode build:** `.github/workflows/ios-products-ci.yml` (matrix over three schemes)

---

## CI/CD — product release tags

Create a tag with:

```bash
chmod +x scripts/release/product-tag.sh
./scripts/release/product-tag.sh echo-messaging 1.0.0
git push origin echo-messaging@v1.0.0
```

Supported products: `echo-messaging`, `comply-portal`, `echo-comply-ios`, `echo-passport`, `passport-sdk`.

Pushing a tag triggers `.github/workflows/release-products.yml`:

| Tag prefix | CI job |
|------------|--------|
| `echo-messaging@*` | `make release-check` + unsigned iOS archive (`EchoMessaging`) |
| `comply-portal@*` | `web/` lint + build + zero-PII script |
| `echo-comply-ios@*` | Unsigned iOS archive (`EchoComplyCompanion`) |
| `echo-passport@*` | Unsigned iOS archive (`EchoPassport`) |

Download the `.xcarchive` artifact from GitHub Actions, then sign and upload to App Store Connect in Xcode Organizer or `xcodebuild -exportArchive`.

---

## Echo Messaging — deploy

**Stack:** full Phase-1 cluster.

```bash
make dev          # metagraph + gateway :8000
make validate-phase1
```

**TestFlight:** archive `EchoMessaging` → App Store Connect app record for `com.echo.app`.

---

## Echo Comply — deploy

Comply is **two deployables** plus optional iOS companion:

### 1. Comply API (Go microservice)

**Comply-only** (no messaging gateway):

```bash
make dev-comply        # postgres :5433 + comply :8011
make dev-comply-stop
```

**Full stack** (messaging gateway + embedded Comply routes on `:8000`):

```bash
make dev               # metagraph + gateway :8000 (docker-compose.testnet.yml)
# Set COMPLY_SERVICE_TOKEN in gateway env for embedded /v3/comply/* routes
```

**Standalone Comply API** (portal + companion app direct calls):

```bash
make dev-comply        # postgres :5433 + comply :8011
# Often paired with make dev when testing Data L1 anchoring
```

Production: deploy `cmd/comply` container with env:

- `COMPLY_SERVICE_TOKEN` — rotate per tenant
- `DATABASE_*` — dedicated Postgres (never share with messaging if isolating tenants)
- `DATA_L1_URL` — tenant Constellation cluster for anchoring

### 2. Comply portal (Next.js)

Deploy `web/` with **tenant-scoped** env (see `web/.env.example`):

| Variable | Purpose |
|----------|---------|
| `NEXT_PUBLIC_SUPABASE_URL` | **Dedicated Supabase project per tenant tier** — portal auth + `org_members` only |
| `NEXT_PUBLIC_SUPABASE_ANON_KEY` | Browser-safe anon key |
| `SUPABASE_SERVICE_ROLE_KEY` | Server-only SCIM + admin reads |
| `COMPLY_API_BASE_URL` | **Dedicated Comply API** for this tenant (e.g. `https://comply.acme.example`) |
| `COMPLY_API_SERVICE_TOKEN` | Portal → Comply service credential |
| `SCIM_BEARER_TOKEN` | IdP → SCIM provisioning |

**Multi-tenant ops model:**

- **Starter / single org:** one Supabase project + one Comply API deploy + one `COMPLY_API_BASE_URL`.
- **Enterprise tier:** separate Supabase project **per customer** (zero cross-tenant auth), separate Comply Postgres + API deploy, optional dedicated Data L1 peer set.
- Portal never stores compliance PII — only operator identity and org DID routing.

Release: `./scripts/release/product-tag.sh comply-portal 1.0.0 && git push origin comply-portal@v1.0.0`

### 3. Echo Comply Companion (iOS)

Read-mostly org dashboard (`ComplyCompanionRootView`). Archive scheme `EchoComplyCompanion` → App Store listing `com.echo.comply`.

Release: `./scripts/release/product-tag.sh echo-comply-ios 1.0.0`

---

## Echo Passport — deploy

Passport uses **gateway route-based** APIs (`pkg/credentials`, OIDC4VC enrollment) — not a separate microservice.

**Standalone app:** archive `EchoPassport` → `com.echo.passport` on App Store.

**Embedded in messaging:** default Phase 2 plan — wallet enrollment inside `EchoMessaging` without a separate listing. Use `passport-sdk@v*` tags for library-only milestones.

Backend requirement: messaging gateway with Identity L0/L1 (`make start-identity`) for VC issuance paths.

---

## Quick reference

```bash
# Local stacks
make dev                 # Messaging metagraph + gateway :8000
make dev-comply          # Standalone Comply API :8011 (+ portal)
make start-comply        # Bare-metal comply on :8011

# Release tags
./scripts/release/product-tag.sh <product> <version>
git push origin <product>@v<version>

# iOS schemes (Xcode)
EchoMessaging | EchoComplyCompanion | EchoPassport
```

See also: [`docs/E2E_LAUNCH_AND_TESTING.md`](E2E_LAUNCH_AND_TESTING.md), [`web/README.md`](../web/README.md).
