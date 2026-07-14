# Contributing to Echo

Echo is a **monorepo** that ships three independently releasable products from one codebase. This guide is the end-to-end setup reference: prerequisites, clean databases, starting every service, and building each app.

| Product | iOS scheme | Bundle ID | Primary backend |
|---------|------------|-----------|-----------------|
| **Echo Messaging** | `EchoMessaging` | `com.echo.app` | Gateway `:8000` + Constellation metagraph |
| **Echo Comply** | `EchoComplyCompanion` | `com.echo.comply` | Comply API `:8011` + Next.js portal `:3000` |
| **Echo Passport** | `EchoPassport` | `com.echo.passport` | Gateway passport routes on `:8000` |

Deeper references:

| Doc | Use |
|-----|-----|
| [`docs/E2E_QUICK_START.md`](docs/E2E_QUICK_START.md) | Daily regression matrix, Xcode smoke paths, troubleshooting |
| [`docs/E2E_LAUNCH_AND_TESTING.md`](docs/E2E_LAUNCH_AND_TESTING.md) | TestFlight, feature matrix, launch sign-off |
| [`docs/PRODUCT_LAUNCH.md`](docs/PRODUCT_LAUNCH.md) | Release tags, CI/CD, production deploy |
| [`AGENTS.md`](AGENTS.md) | Cursor agent entry + MCP |

> **Working on the iOS app only?** Jump to **§4** — build and iterate with just Xcode, no backend required.

---

## 1. Prerequisites

| Tool | Version | Required for |
|------|---------|--------------|
| **macOS 14+** or Ubuntu 22.04+ | — | Dev + iOS |
| **Go** | 1.21+ (`go.mod`) | Gateway, Comply microservice, CLI |
| **Docker + Compose v2** | 8 GB+ RAM, 4+ CPUs | Postgres, Redis, NATS, MinIO, metagraph peers |
| **JDK** | **21** (Temurin) | Scala metagraph assembly + tests |
| **sbt** | 1.9+ | Metagraph build |
| **Scala** | 2.13.10 | Pinned in `metagraph/build.sbt` |
| **Xcode** | 15+ (16 recommended) | All three iOS apps |
| **Node.js** | 20+ | Comply web portal (`web/`) |
| **jq, yq, argc, Ansible, g8** | latest | Euclid `hydra` (via `setup-euclid.sh`) |

### macOS (one-time)

```bash
brew install --cask temurin@21 sbt coursier/formulas/coursier go jq yq argc ansible node
cs install scala:2.13.10 scalafmt giter8
brew install --cask docker && open -a Docker
xcode-select --install
export JAVA_HOME=$(/usr/libexec/java_home -v 21)
# Docker Desktop → Settings → Resources: Memory ≥ 8 GB, CPUs ≥ 4
```

### Verify

```bash
java -version    # 21.x
go version
docker compose version
node --version   # 20+
cd metagraph && ./scripts/setup-euclid.sh   # expect: ✓ All prerequisites found
```

---

## 2. Clone and bootstrap

```bash
git clone git@github.com:c1cromwell/echoapp.git && cd echoapp
cp .env.example .env
go mod download
cd metagraph && ./scripts/setup-euclid.sh && cd ..
```

`setup-euclid.sh` clones [`euclid-development-environment`](../euclid-development-environment) as a **sibling directory** (`../euclid-development-environment`). `make dev` expects it there.

### Recommended `.env` for local dev

Edit `.env` after copy. Values used by `make run` (foreground gateway) and optionally forwarded to Docker:

```bash
# Gateway
API_PORT=8000
ENVIRONMENT=development
DEV_MODE=true
ALLOW_DEV_OTP=true          # SMS OTP in X-Dev-OTP header (dev only)

# Embedded Comply routes on the messaging gateway (:8000/v3/comply/*)
COMPLY_SERVICE_TOKEN=dev-comply-token

# Passport / OIDC4VC wallet enrollment (Phase 2)
OIDC4VC_ENABLED=true

# Optional Phase 5+ features
# ECHO_PQ_ENABLED=true
# CLOUD_OAUTH_STUB=true
# ESTUARY_API_TOKEN=...
```

**Note:** `docker-compose.testnet.yml` (used by `make dev`) currently forwards `DEV_MODE` but not every `.env` key. For containerized gateway with Comply + OIDC4VC, either run `make run` against local Postgres/Redis or add the vars to `docker-compose.testnet.yml` under `echoapp.environment`.

---

## 3. Architecture (local dev)

```text
┌─────────────────────────────────────────────────────────────────────────┐
│  iOS apps (Xcode schemes)                                               │
│  EchoMessaging (:8000) │ EchoPassport (:8000) │ EchoComply (:8000+8011)│
└────────────┬───────────────────────────────┬────────────────────────────┘
             │                               │
             ▼                               ▼
┌────────────────────────────┐   ┌──────────────────────────┐
│  Messaging gateway :8000    │   │  Comply API :8011         │
│  (Go — main.go)             │   │  (cmd/comply)             │
│  Postgres echoapp :5432     │   │  Postgres echo_comply :5433│
│  Redis, NATS, MinIO :9700   │   │  (make dev-comply)        │
└────────────┬───────────────┘   └────────────┬─────────────┘
             │                                 │
             ▼                                 ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  Constellation / Euclid metagraph (host — hydra in sibling repo)        │
│  Global L0 :9000 │ Metagraph L0 :9200 │ Currency L1 :9300 │ Data :9400 │
│  Identity L0 :9600 │ Identity L1 :9500  (make start-identity)           │
└─────────────────────────────────────────────────────────────────────────┘

Comply web portal (:3000) ──server-side──▶ Comply API :8011
              └── Supabase (operator auth + org membership only)
```

**Port conflicts to avoid:** Global L0 uses **9000**. The legacy `docker-compose.yml` MinIO mapping also used 9000 — **`make dev` uses `docker-compose.testnet.yml` with MinIO on 9700/9701** instead.

---

## 4. iOS-only setup (mobile developers)

For contributors working on **only the iOS app** (`ios/Echo/`). You can build, preview, and iterate on UI with **just Xcode** — no backend, metagraph, Docker, or web stack. A backend is needed only for live end-to-end flows (§4d).

### 4a. Minimal prerequisites

| Tool | Needed for iOS-only? |
|------|----------------------|
| **macOS 14+**, **Xcode 15+** (16 recommended) | ✅ required |
| Go, JDK/sbt/Scala, Docker, Node, Euclid (`setup-euclid.sh`) | ❌ not needed |

### 4b. Open the project

```bash
git clone git@github.com:c1cromwell/echoapp.git && cd echoapp
open ios/Echo/EchoApp.xcodeproj
```

- Xcode **auto-resolves** the two Swift Package deps on first open — **MLKEMNativeSwift** and **`stasel/WebRTC`** (both public). Wait for "Resolving Packages" to finish, or File → Packages → Resolve Package Versions. No CocoaPods/Carthage.
- Scheme: **`EchoMessaging`** (not the legacy `EchoApp` alias). Pick an iPhone simulator (e.g. iPhone 17). **⌘B** build · **⌘R** run · **⌘U** test.
- `EchoOPRF.xcframework` (PSI) is **optional** — the app falls back to a mock when it's absent. Skip `make echooprf-ios` unless you're working on PSI.

### 4c. Work without a backend (UI-only)

The app **compiles** with no backend; at runtime, network calls degrade to empty/error states (it is not a designed offline mode). The productive backend-free loops:

- **SwiftUI Previews** — 44+ `#Preview`s under `ios/Echo/Sources`; the fastest loop for a single view.
- **Screen catalog** — `make screen-catalog`, then `open docs/screen_catalog/index.html`. Renders every screen to PNG via the simulator; Xcode only, no backend.
- **DEBUG-seeded Messages** — a DEBUG build auto-populates the Messages tab (4 DMs, 2 groups, 2 hidden folders) via device-local stores — see `ios/Echo/Sources/Services/DebugSeedData.swift`. No backend; seeds once, after first-run creates a local identity.

> First-run onboarding/login calls the backend. For pure UI work, stay in Previews / screen-catalog, or point at a backend (§4d).

### 4d. Point at a shared backend (live flows)

To exercise real login, messaging, wallet, etc., point the app at a backend. Resolution precedence (`ios/Echo/Sources/Core/Networking/EchoAPIBaseURL.swift`): **`ECHO_API_URL`** → **`API_URL`** → Info.plist `$(API_URL)` → fallback `https://api.echo.local` (placeholder). The `EchoMessaging` scheme presets `API_URL=http://localhost:8000`.

- **Shared/staging backend (recommended for mobile-only devs):** Xcode → Product → Scheme → Edit Scheme → Run → Arguments → Environment Variables → set `ECHO_API_URL=https://<your-team-staging>`.
  <!-- TODO: replace <your-team-staging> with the team's hosted dev/staging backend URL. -->
- **Physical device:** use a LAN-reachable URL, not `localhost`.
- **Run the whole stack yourself:** see §6 (Echo Messaging — full E2E setup) — requires Go + Docker + metagraph, beyond "iOS-only."

### 4e. Build & test gates (iOS-only)

> ⚠️ Do **not** run `make ios-preflight` without a backend — it curls `/health` and exits 1. Use a plain Xcode build instead.

```bash
# Build (simulator; no backend, no signing) — or just ⌘B in Xcode
cd ios/Echo
xcodebuild build -project EchoApp.xcodeproj -scheme EchoMessaging \
  -destination 'platform=iOS Simulator,name=iPhone 17' CODE_SIGNING_ALLOWED=NO

# Backend-free unit tests (crypto/auth + messaging logic)
swift test --filter EchoSecurityTests
swift test --filter EchoPhase3Tests
```

Prefer `EchoSecurityTests` / `EchoPhase3Tests`; the legacy `EchoTests` target has compile issues (see §13).

### 4f. Adding new Swift files

The project is a **classic checked-in `.xcodeproj`** (no XcodeGen). Register new files in the pbxproj:

```bash
python3 scripts/xcode-add-sources.py ios/Echo/EchoApp.xcodeproj/project.pbxproj <GROUPID>:NewFile.swift
```

Find `<GROUPID>` by grepping a sibling file's group in `project.pbxproj`. Idempotent; writes a `.bak`. (The SwiftPM `Package.swift` globs `Sources/` automatically — only the `.xcodeproj` needs the manual entry.)

### 4g. Design system & guardrails

- **Use design tokens**, never raw hex: `Color.echo*`, `Spacing`, `Typography`, `.glacialShadow()` — source of truth `ios/Echo/Sources/DesignSystem/`.
- **Onboarding & login are frozen** (`FirstRunCoordinator`, `GlacialLoginScreen`) — do not redesign ([`AGENTS.md`](AGENTS.md)).
- The **`echo-ios-design` skill** (`.cursor/skills/echo-ios-design/` for Cursor, `.agents/skills/echo-ios-design/` for Claude Code) encodes these rules + the screen-catalog loop, and delegates generic SwiftUI review to `swiftui-pro` / `swiftui-expert`.
- Xcode smoke path (with a backend): [`docs/E2E_QUICK_START.md` §7](docs/E2E_QUICK_START.md).

---

## 5. Clean slate (reset all data)

Run this when you need fresh Postgres, Redis, object storage, and chain state.

```bash
# 1. Stop everything
make dev-stop
make dev-comply-stop 2>/dev/null || true
make stop-identity 2>/dev/null || true
if [ -d ../euclid-development-environment ]; then
  cd ../euclid-development-environment && scripts/hydra stop && cd - >/dev/null
fi

# 2. Delete Docker volumes (destroys all DB + MinIO data)
docker compose -f docker-compose.testnet.yml down -v --remove-orphans
docker compose -f docker-compose.comply.yml down -v --remove-orphans
docker compose -f docker-compose.identity.yml down -v --remove-orphans 2>/dev/null || true

# 3. Optional: wipe identity service keys (regenerated on next make dev)
rm -rf .keys/identity-service.*

# 4. Optional: iOS simulator clean install
# Xcode → Device → Erase All Content and Settings (per simulator)
```

Migrations apply **automatically** on gateway and Comply startup (`database.Migrate` in `main.go` / `cmd/comply`). Manual apply:

```bash
DATABASE_HOST=localhost DATABASE_NAME=echoapp DATABASE_USER=echoapp \
  DATABASE_PASSWORD=echoapp_dev go run ./cmd/migrate
```

---

## 6. Echo Messaging — full E2E setup

**Goal:** Two simulators/devices can register, DM, and pass Phase 1 go/no-go.

### 5a. Start infrastructure

```bash
# Metagraph cluster + backend stack (Postgres, Redis, NATS, MinIO, gateway)
make dev
make dev-status          # all ✓ except Identity (until next step)

# Identity metagraph (VC issuance, trust tiers — required for validate-phase1 step 3)
make start-identity
curl -s http://localhost:9600/node/info | jq .state   # → "Ready"
curl -s http://localhost:9500/node/info | jq .state

# Go/no-go harness (6 steps)
make validate-phase1     # expect: Go/No-Go: GO
```

First `make dev` builds Euclid images and can take **15–30 minutes**. Subsequent starts are faster.

### 5b. Rebuild metagraph JARs (after Scala changes)

```bash
cd metagraph
export JAVA_HOME=$(/usr/libexec/java_home -v 21)   # macOS
sbt "sharedData/test" "identityL0/assembly" "identityL1/assembly" "dataL1/assembly"
cd ..

# Redeploy Data L1 + Metagraph L0 into hydra (see docs/E2E_QUICK_START.md §3b)
make start-identity      # Identity JARs
```

### 5c. Build and run iOS (Echo Messaging)

```bash
# Headless compile gate
make ios-preflight BUILD=1 TESTS=1

# Or Xcode
open ios/Echo/EchoApp.xcodeproj
# Scheme: EchoMessaging  (not legacy EchoApp)
# Run env: API_URL=http://localhost:8000  (pre-set in scheme)
# Physical device: API_URL=http://<Mac-LAN-IP>:8000
```

**Simulator smoke:** [`docs/E2E_QUICK_START.md` §7](docs/E2E_QUICK_START.md) — onboarding, two-client DM, Phase 3 signals.

**CI-equivalent build:**

```bash
cd ios/Echo
xcodebuild build -project EchoApp.xcodeproj -scheme EchoMessaging \
  -configuration Debug-Messaging -destination 'generic/platform=iOS' CODE_SIGNING_ALLOWED=NO
```

### 5d. Messaging health checks

```bash
curl -s http://localhost:8000/health | jq .status          # operational
curl -s http://localhost:9400/node/info | jq .state        # Data L1
curl -s http://localhost:8000/v3/pq/status | jq .          # PQ gate (if ECHO_PQ_ENABLED)
```

### 5e. Stop messaging stack

```bash
make dev-stop
make stop-identity
cd ../euclid-development-environment && scripts/hydra stop
```

---

## 7. Echo Comply — full E2E setup

Comply is **three surfaces**: embedded gateway routes (retention/holds for messaging users), standalone **Comply API** (`:8011`), and **web admin portal** (`web/`).

### 6a. Comply API only (isolated tenant DB)

Best for portal and operator workflow development without the full metagraph cluster.

```bash
make dev-comply
curl -s http://localhost:8011/health | jq .

# Postgres: localhost:5433, database echo_comply, user/password echoapp/echoapp_dev
make dev-comply-stop     # tear down
```

### 6b. Comply + Data L1 anchoring

For evidence anchoring and dashboard metrics that read chain state:

```bash
make dev                 # metagraph + Data L1 on :9400
make dev-comply          # Comply API on :8011 (separate Postgres on :5433)
```

Point Comply at Data L1 via `DATA_L1_URL=http://host.docker.internal:9400` (default in `docker-compose.comply.yml`).

### 6c. Comply web portal

```bash
# Terminal 1 — Comply API
make dev-comply

# Terminal 2 — portal
cd web
cp .env.example .env.local
# Fill: NEXT_PUBLIC_SUPABASE_* (dedicated Supabase project)
#       COMPLY_API_BASE_URL=http://localhost:8011
#       COMPLY_API_SERVICE_TOKEN=dev-service-token  (match Comply token policy)
npm install
npm run dev              # http://localhost:3000
```

Apply portal schema: `supabase db push` or run `web/supabase/migrations/0001_portal_orgs.sql` in Supabase SQL editor.

Architecture: [`web/README.md`](web/README.md) · ADR [`docs/adr/0005-comply-web-admin-portal.md`](docs/adr/0005-comply-web-admin-portal.md).

### 6d. Echo Comply Companion (iOS)

```bash
# Needs messaging gateway (signed REST) + Comply API
make dev                 # or at minimum gateway on :8000
make dev-comply          # :8011

open ios/Echo/EchoApp.xcodeproj
# Scheme: EchoComplyCompanion
# Env (pre-set): API_URL=http://localhost:8000, COMPLY_API_URL=http://localhost:8011
```

**CI-equivalent build:**

```bash
cd ios/Echo
xcodebuild build -project EchoApp.xcodeproj -scheme EchoComplyCompanion \
  -configuration Debug-Comply -destination 'generic/platform=iOS' CODE_SIGNING_ALLOWED=NO
```

### 6e. Embedded Comply on messaging gateway

Retention and hold APIs also mount on `:8000` when `COMPLY_SERVICE_TOKEN` is set and `COMPLY_EMBEDDED` is not `false` (default embedded). Use this path when testing Comply features **inside** the messaging app without the standalone microservice.

---

## 8. Echo Passport — full E2E setup

Passport uses **gateway routes** (`/credentials`, OIDC4VC, sync) — no separate microservice.

### 7a. Backend

```bash
make dev
make start-identity      # VC issuance + StatusList2021 paths

# Foreground gateway with Passport env (if not in Docker)
export OIDC4VC_ENABLED=true
export COMPLY_SERVICE_TOKEN=dev-comply-token   # optional; unrelated to Passport core
make run
```

Verify issuer metadata:

```bash
curl -s http://localhost:8000/.well-known/openid-credential-issuer | jq .
```

### 7b. iOS (Echo Passport)

```bash
open ios/Echo/EchoApp.xcodeproj
# Scheme: EchoPassport
# Env: API_URL=http://localhost:8000
```

**CI-equivalent build:**

```bash
cd ios/Echo
xcodebuild build -project EchoApp.xcodeproj -scheme EchoPassport \
  -configuration Debug-Passport -destination 'generic/platform=iOS' CODE_SIGNING_ALLOWED=NO
```

Passport can also run **embedded** inside Echo Messaging (default Phase 2 plan). Standalone `EchoPassport` is the separate App Store listing (`com.echo.passport`).

See [`docs/ECHO_PASSPORT_PLAN.md`](docs/ECHO_PASSPORT_PLAN.md) and ADR [`docs/adr/0003-echo-passport-custody-model.md`](docs/adr/0003-echo-passport-custody-model.md).

---

## 9. Quick start matrix (which commands when)

| I am working on… | Start | iOS scheme | Verify |
|------------------|-------|------------|--------|
| **Messaging DMs / chat** | `make dev` | `EchoMessaging` | `make validate-phase1`, E2E §7 |
| **Wallet / staking** | `make dev` | `EchoMessaging` | `./scripts/validate-wallet.sh` |
| **Comply portal / holds** | `make dev-comply` (+ `make dev` for L1) | — | `curl :8011/health`, portal login |
| **Comply iOS companion** | `make dev` + `make dev-comply` | `EchoComplyCompanion` | Companion dashboard loads |
| **Passport / credentials** | `make dev` + `make start-identity` | `EchoPassport` | OIDC4VC metadata + wallet enrollment |
| **Metagraph validators** | `make dev` | — | `make metagraph-test` |
| **Go API only (no chain)** | `make run` | any | `curl :8000/health` |

---

## 10. Builds and test gates (all products)

### Backend (every PR)

```bash
make release-check       # build + race tests + vet + gofmt
make metagraph-test      # Scala validators (JDK 21)
```

### iOS (every PR touching `ios/`)

```bash
make ios-preflight BUILD=1 TESTS=1
cd ios/Echo && swift build --target Echo
cd ios/Echo && swift test --filter EchoPhase3Tests    # messaging signals / PQ
cd ios/Echo && swift test --filter EchoSecurityTests  # crypto / auth
```

Per-product Xcode matrix (matches CI):

```bash
cd ios/Echo
for scheme in EchoMessaging EchoComplyCompanion EchoPassport; do
  xcodebuild build -project EchoApp.xcodeproj -scheme "$scheme" \
    -destination 'generic/platform=iOS' CODE_SIGNING_ALLOWED=NO
done
```

### Comply portal

```bash
cd web && npm run lint && npm run type-check && npm run build && npm run test:zero-pii
```

### Full regression (pre-TestFlight)

```bash
make regression-with-phase1
```

---

## 11. Daily workflow

```bash
make dev && make dev-status     # start of day
make start-identity             # if working on VC / Passport / trust tier

# Surface you changed:
make test                       # Go
make metagraph-test             # Scala
make ios-preflight BUILD=1      # iOS compile

open ios/Echo/EchoApp.xcodeproj # pick product scheme
```

Regression by milestone: [`docs/E2E_QUICK_START.md` §0](docs/E2E_QUICK_START.md).

---

## 12. Repository layout

```text
echoapp/
├── main.go                 # Messaging gateway (:8000)
├── cmd/comply/             # Comply microservice (:8011)
├── cmd/migrate/            # Manual SQL migrations
├── internal/api/           # HTTP + WebSocket handlers
├── internal/services/      # Domain services (media, comply, cloud, zk, …)
├── pkg/credentials/        # W3C VC 2.0, OIDC4VC (Passport)
├── metagraph/              # Scala Tessellation validators
├── migrations/             # Postgres schema (shared; auto-applied on boot)
├── ios/Echo/               # SwiftPM library + EchoApp.xcodeproj (3 schemes)
├── web/                    # Echo Comply Next.js portal
├── scripts/                # validate-phase1, regression, identity keys
├── docker-compose.testnet.yml   # make dev (messaging backend)
├── docker-compose.comply.yml    # make dev-comply
└── docker-compose.identity.yml  # make start-identity
```

---

## 13. Known gaps and doc drift (2026-05)

| Topic | Reality | Action |
|-------|---------|--------|
| `make dev` vs Comply sidecar | `make dev` uses `docker-compose.testnet.yml` — **does not** start `:8011` Comply container | Use `make dev-comply` for Comply API; embedded Comply routes need `COMPLY_SERVICE_TOKEN` in gateway env |
| `PRODUCT_LAUNCH.md` “make dev includes comply” | Accurate for `docker-compose.yml`, not testnet compose | This doc + testnet compose are source of truth for messaging dev |
| Legacy `EchoApp` scheme | Alias; product work uses `EchoMessaging` / `EchoComplyCompanion` / `EchoPassport` | CI builds all three via `ios-products-ci.yml` |
| Comply portal Supabase | Requires a real Supabase project (no in-repo local Supabase stack) | Create project; apply `web/supabase/migrations/` |
| Passport iOS module (WO-297) | Backend Wave A ✅; standalone Passport UI backlog items remain | See [`docs/ECHO_MESSAGING_LAUNCH_STATUS.md`](docs/ECHO_MESSAGING_LAUNCH_STATUS.md) |
| `EchoTests` target | Some legacy tests fail to compile | Prefer `EchoPhase3Tests`, `EchoSecurityTests`, `xcodebuild -scheme EchoMessaging test` |

Phase gap audits: [`docs/PHASE4_7_GAP_AUDIT.md`](docs/PHASE4_7_GAP_AUDIT.md).

---

## 14. Troubleshooting

| Symptom | Fix |
|---------|-----|
| `docker: Cannot connect` | Start Docker Desktop |
| Port 9000 in use | Global L0 — stop hydra or conflicting MinIO from `docker-compose.yml` |
| `make dev` timeout on L0 | `cd ../euclid-development-environment && scripts/hydra status` |
| Identity L0/L1 down | `make start-identity` after `make dev`; rebuild JARs if needed |
| Device cannot reach API | Set `API_URL=http://<Mac-LAN-IP>:8000` in Xcode scheme |
| Comply portal 401 | Align `COMPLY_API_SERVICE_TOKEN` in `web/.env.local` with Comply `COMPLY_SERVICE_TOKEN` |
| Migrations out of date | Restart gateway/comply (auto-migrate) or `make migrate` |
| Wrong JDK for sbt | `export JAVA_HOME=$(/usr/libexec/java_home -v 21)` |

Full table: [`docs/E2E_QUICK_START.md` §11](docs/E2E_QUICK_START.md).

---

## 15. Contributing code

1. Branch from `main`.
2. Run gates in §10 for surfaces you touched.
3. Follow [`docs/data-classification.md`](docs/data-classification.md) (T0–T7) for anything chain-adjacent or logged.
4. Phase 1–2 identity: **`did:key` only** — no Cardano/PRISM ([ADR-0001](docs/adr/0001-phase1-identity-method.md)).
5. Open PR; ensure CI green.

Release tags per product: [`docs/PRODUCT_LAUNCH.md`](docs/PRODUCT_LAUNCH.md).

**Agents:** read [`AGENTS.md`](AGENTS.md) and load the relevant skill under `.cursor/skills/`.
