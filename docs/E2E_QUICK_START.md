# Echo — Developer quick start & daily regression

**One doc for:** first-time setup, full rebuilds (Go + Scala + iOS), and **what to run each day** at each phase/milestone.

| Doc | Use |
|-----|-----|
| **This file** | Setup, rebuild, daily/weekly regression by milestone |
| [`E2E_LAUNCH_AND_TESTING.md`](E2E_LAUNCH_AND_TESTING.md) | Full launch procedures, feature delivery matrix, TestFlight sign-off |
| [`WEEK_A_B_LAUNCH.md`](WEEK_A_B_LAUNCH.md) | Week A messaging → Week B contacts sprint |
| [`TESTFLIGHT_WEEK_A_B.md`](TESTFLIGHT_WEEK_A_B.md) | Tester script A1–A10 / B1–B6 |

**Agents:** skill **`echo-testing`** · MCP **`echo-local-dev`** (`cluster_status`, `run_ios_preflight`, `run_regression`, `run_validate_phase1`).

**Frozen UX:** validate shipped `FirstRunCoordinator` / `GlacialLoginScreen` only — do not redesign from the React prototype ([`ux-spec.md`](ux-spec.md), [`ECHO_IOS_UI_IMPLEMENTATION_SPEC.md`](ECHO_IOS_UI_IMPLEMENTATION_SPEC.md) §0).

---

## 0. Daily regression by milestone

Run the row for your current focus **before** opening Xcode or uploading TestFlight.

| Milestone | When | Automated (run in order) | Manual (you / Xcode) | Pass if |
|-----------|------|--------------------------|----------------------|---------|
| **Stack health** | Every morning | `make dev` → `make dev-status` | — | All endpoints ✓; backend `operational` |
| **Backend PR** | Every Go change | `make release-check` | — | Build + race tests + vet + fmt green |
| **Metagraph PR** | Every Scala change | `make metagraph-test` | — | `sharedData` + identity/data L1 tests green |
| **Phase 1 go/no-go** | Metagraph/auth changes; pre-TestFlight | `make regression-with-phase1` | — | `validate-phase1` → **GO** (all steps `ok`) |
| **iOS compile gate** | Every iOS change | `make ios-preflight BUILD=1 TESTS=1` | — | Zero `FAIL` lines |
| **Week A — messaging** | DM / chat / Phase 3 UI | `make phase3-signals-proof` + `make regression` | [`TESTFLIGHT_WEEK_A_B.md`](TESTFLIGHT_WEEK_A_B.md) **A1–A10** | Two clients DM; typing/reactions/history |
| **Week B — contacts** | After Week A green | `make regression` | **B1–B6** in same tester doc | Invite, QR, block, link-device |
| **Pre-TestFlight** | Before archive | `make regression-with-phase1` | Quick start §7 simulator smoke + [`E2E_LAUNCH_AND_TESTING.md` §9](E2E_LAUNCH_AND_TESTING.md#9-launch-checklist--sign-off) | Sign-off template complete |

**PR bar (all milestones):**

```bash
make fmt && make lint && make vet && make test
cd metagraph && sbt scalafmtCheckAll && sbt test && cd ..
make validate-phase1    # or document skips in PR if simulator-only
```

---

## 1. Prerequisites

| Tool | Version | Why |
|------|---------|-----|
| macOS 14+ or Ubuntu 22.04+ | — | Phase 1 tested on both |
| **JDK** | **21** (Temurin) | Tessellation 4.0.0-rc.0 |
| sbt | 1.9+ | Metagraph Scala |
| Scala | 2.13.10 | Pinned in `metagraph/build.sbt` |
| Go | 1.21+ | Backend |
| Docker + Compose v2 | **8 GB+ RAM, 4+ CPUs** | Postgres, Redis, Euclid cluster |
| Xcode | 15+ (macOS) | iOS (`ios/Echo/`) |
| `jq`, `yq`, `argc`, Ansible 2.16+, g8 | latest | Euclid `hydra` |

`metagraph/scripts/setup-euclid.sh` checks every prerequisite and reports what is missing.

### macOS (one-time install)

```bash
brew install --cask temurin@21 sbt coursier/formulas/coursier go jq yq argc ansible
cs install scala:2.13.10 scalafmt giter8
brew install --cask docker && open -a Docker
xcode-select --install
# Docker Desktop → Settings → Resources: Memory ≥ 8 GB, CPUs ≥ 4
```

### Ubuntu (one-time install)

```bash
sudo apt update && sudo apt install -y curl git jq build-essential python3-pip
sudo pip3 install --user 'ansible>=2.16'
curl -s "https://get.sdkman.io" | bash && source "$HOME/.sdkman/bin/sdkman-init.sh"
sdk install java 21.0.5-tem && sdk install sbt 1.9.9 && sdk install scala 2.13.10
curl -fsSL https://get.docker.com | sudo sh && sudo usermod -aG docker "$USER"
# Go: https://go.dev/dl/ — match go.mod toolchain
```

### Verify toolchain

```bash
java -version    # 21.x
sbt --version
scala -version   # 2.13.10
go version
docker version && docker compose version
```

macOS JDK pin: `export JAVA_HOME=$(/usr/libexec/java_home -v 21)`

---

## 2. One-time clone & bootstrap

```bash
git clone git@github.com:c1cromwell/echoapp.git && cd echoapp
cp .env.example .env
go mod download
cd metagraph && ./scripts/setup-euclid.sh && cd ..
```

Expect `✓ All prerequisites found` from `setup-euclid.sh`.

---

## 3. Full rebuild (backend + metagraph + iOS)

Use after pulling large changes, before `validate-phase1`, or when JARs/containers look stale.

### 3a. Go backend

```bash
# Local binary
go build -o echoapp main.go

# Docker API (testnet stack)
make testnet-up          # rebuilds echoapp image; needs hydra cluster up
# or full stack:
make dev                 # Euclid + backend + identity keys
```

### 3b. Scala metagraph (assembly JARs)

`sbt compile` alone does **not** produce runnable nodes — you need **assembly**:

```bash
cd metagraph
export JAVA_HOME=$(/usr/libexec/java_home -v 21)   # macOS; omit on Linux if JAVA_HOME set
sbt "sharedData/test" "identityL0/assembly" "identityL1/assembly" "dataL1/assembly" "currencyL0/assembly"
cd ..
```

**Deploy into hydra** (Euclid cluster — Data L1 + Metagraph L0):

```bash
EUCLID=../euclid-development-environment
cp metagraph/modules/data_l1/target/scala-2.13/*assembly*.jar "$EUCLID/infra/shared/jars/data-l1.jar"
cp metagraph/modules/l0/target/scala-2.13/*assembly*.jar "$EUCLID/infra/shared/jars/metagraph-l0.jar"
cd "$EUCLID" && scripts/hydra stop
for n in metagraph-node-1 metagraph-node-2 metagraph-node-3; do
  docker cp "$EUCLID/infra/shared/jars/data-l1.jar" "$n:/code/data-l1/data-l1.jar"
  docker cp "$EUCLID/infra/shared/jars/metagraph-l0.jar" "$n:/code/metagraph-l0/metagraph-l0.jar"
done
scripts/hydra start-genesis
cd -
```

Identity metagraph (separate from hydra):

```bash
make start-identity    # uses identity_l0/l1 assembly JARs
```

Or rebuild images from source: `cd "$EUCLID" && scripts/hydra build --no_cache` then `start-genesis`.

### 3c. iOS

```bash
make ios-preflight BUILD=1 TESTS=1

# Optional: live PSI framework (WO-221)
make echooprf-ios
# Then in Xcode: embed ios/Echo/Libraries/EchoOPRF.xcframework (Embed & Sign)
```

Manual compile gate:

```bash
cd ios/Echo
xcodebuild -project EchoApp.xcodeproj -scheme EchoApp \
  -sdk iphonesimulator \
  -destination 'generic/platform=iOS Simulator' \
  build
swift test --filter EchoSecurityTests
swift test --filter EchoPhase3Tests
```

---

## 4. First boot (clean machine → GO)

```bash
# 1. Metagraph JARs (§3b)
cd metagraph && sbt assembly && cd ..

# 2. Full cluster + backend
make dev

# 3. Identity L0/L1 (Step 3 of validate-phase1)
make start-identity
curl -s http://localhost:9600/node/info | jq .state   # → "Ready"
curl -s http://localhost:9500/node/info | jq .state

# 4. WO-230 go/no-go
make validate-phase1
```

Expected: **Go/No-Go: GO** with all six steps `ok` (Steps 3 and 5 need identity + Euclid Data L1).

**Stop when done:**

```bash
make dev-stop
make stop-identity
cd ../euclid-development-environment && scripts/hydra stop
```

---

## 5. Daily workflow

```bash
make dev && make dev-status    # start of day (idempotent)

# Surface you touched today:
make run                       # foreground Go API
make test                      # Go suite
make metagraph-test            # Scala validators
open ios/Echo/EchoApp.xcodeproj   # scheme EchoApp; simulator API_URL=http://localhost:8000
```

**Agent / CI gate before Xcode:**

```bash
make ios-preflight             # health + Xcode + scheme
make ios-preflight BUILD=1 TESTS=1
```

---

## 6. Agent vs you (iOS)

| Who | What | Command |
|-----|------|---------|
| **Agent / CI** | Go + iOS unit tests, backend health, compile | `make regression` · `make ios-preflight` |
| **You (Xcode)** | Face ID, onboarding taps, two-client chat, TestFlight | §7 below (~15 min) |

Agents **cannot** drive Simulator UI or real biometrics.

---

## 7. Xcode smoke path (~15 min)

### Setup (once per machine)

1. Open **`ios/Echo/EchoApp.xcodeproj`** (scheme **`EchoApp`**).
2. **Simulator:** `API_URL=http://localhost:8000` (in scheme).
3. **Physical iPhone:** Scheme → Run → Environment → `API_URL=http://<Mac-LAN-IP>:8000` (`ipconfig getifaddr en0`).
4. Simulator: **Features → Face ID → Enrolled**.

### New user (canonical flow)

| # | Action | Pass if |
|---|--------|---------|
| 1 | Cold launch | Welcome screen |
| 2 | Set up → username | Availability OK |
| 3 | Continue → Face ID | Enrollment completes |
| 4 | Recovery → **SMS backup** | OTP (`DEV_MODE=true` → `X-Dev-OTP` header) |
| 5 | Main app / Messages | Tab bar visible |
| 6 | Settings → Privacy → discovery ON | Toggle saves |
| 7 | Me → **Find contacts on ECHO** → Scan | No crash |
| 8 | Add match → **Message** | DM thread opens |

### Returning user

Force-quit → relaunch → Face ID login → app unlocks.

### Two-client relay (Week A core)

1. User A: new conversation → search **@username** of B → send message.
2. User B: open same thread → sees message.
3. Both need completed onboarding + same `API_URL`.

Full Week A/B steps: [`TESTFLIGHT_WEEK_A_B.md`](TESTFLIGHT_WEEK_A_B.md).

### Phase 3 signals (two clients)

**Headless:** `make phase3-signals-proof`

| Signal | Pass if |
|--------|---------|
| Typing | Recipient sees indicator; clears on send |
| Read receipts | Open chat → sender checkmarks → read |
| Reactions | Long-press 👍 toggles; matches REST |

---

## 8. Running tests (reference)

| Surface | Command |
|---------|---------|
| Go (unit + integration) | `make test` |
| Go release gate | `make release-check` |
| Go endpoint smoke | `make test-endpoints` (needs `make run`) |
| Scala validators | `cd metagraph && sbt 'sharedData/test'` |
| Scala wired L1 | `sbt 'identityL1/test' 'dataL1/test' 'l1/test'` |
| Phase 1 go/no-go | `make validate-phase1` |
| Full regression | `make regression` · `make regression-with-phase1` |
| iOS SPM | `make ios-preflight BUILD=1 TESTS=1` |
| T0–T7 | `semgrep --config .semgrep/t0_t7_rules.yaml --error .` |

---

## 9. Repository layout

```
echoapp/
├── cmd/           # Go CLIs (didkey, migrate, …)
├── internal/      # api/ auth/ metagraph/ services/ …
├── pkg/           # didkey/ credentials/ …
├── metagraph/     # Scala Tessellation 4.x (shared_data + L0/L1)
├── ios/Echo/      # SwiftPM + EchoApp.xcodeproj
├── scripts/       # validate-phase1.sh, run-regression.sh, …
└── Makefile       # make help
```

Architecture map: [`AGENTS.md`](../AGENTS.md).

---

## 10. Submitting Scala L1 validation changes

1. **Rule** in `metagraph/modules/shared_data/.../validations/` — pure `Either[String, Unit]`.
2. **Wire** in `Main.dispatch` (`identity_l1`, `l1`, or `data_l1`).
3. **Pure test** in `*ValidationsSpec.scala`.
4. **Wired test** in module `MainSpec.scala` (WO-277).
5. `sbt scalafmtAll && sbt test` green before push.

New `IdentityUpdate` / `EchoUpdate` variants: update encoders in `IdentityTypes.scala` / `Types.scala` and every `dispatch` match (`-Werror`).

---

## 11. Troubleshooting

| Symptom | Fix |
|---------|-----|
| `docker: Cannot connect` | Start Docker Desktop |
| `sbt assembly` slow first run | Normal — Coursier cache warms up |
| Wrong JDK | `export JAVA_HOME=$(/usr/libexec/java_home -v 21)` |
| Port 9000–9602 in use | `lsof -nP -iTCP -sTCP:LISTEN \| grep ':9'` |
| Identity L0/L1 down | `make start-identity` after `make dev` |
| `IDENTITY_SERVICE_DID` unset | `scripts/identity/ensure-identity-service-key.sh` via `make dev` |
| Device API errors | `API_URL` = Mac LAN IP, not `localhost` |
| `validate-phase1` Step 5 fails | Rebuild Data L1 JAR + redeploy to hydra (§3b) |
| `xcodebuild` fails | Full Xcode.app; `sudo xcode-select -s /Applications/Xcode.app/Contents/Developer` |
| New Swift file missing in target | Add file to **EchoApp** target in Xcode |

Env vars: [E2E_LAUNCH_AND_TESTING.md §3e](E2E_LAUNCH_AND_TESTING.md#3e-environment-variable-reference).

---

## 12. Minimal launch sign-off

Before TestFlight upload:

- [ ] `make regression` green
- [ ] `make ios-preflight BUILD=1` green
- [ ] §7 smoke on **simulator**
- [ ] §7 smoke on **device** (LAN `API_URL`)
- [ ] `make validate-phase1` all steps `ok`

**Full sign-off:** [`E2E_LAUNCH_AND_TESTING.md` §9](E2E_LAUNCH_AND_TESTING.md#9-launch-checklist--sign-off).

---

## Related

| Resource | Use |
|----------|-----|
| [`E2E_LAUNCH_AND_TESTING.md`](E2E_LAUNCH_AND_TESTING.md) | Tiers, env vars, TestFlight, feature matrix, sign-off |
| [`metagraph-backend-e2e-testing.md`](metagraph-backend-e2e-testing.md) | Hydra ports, node URLs |
| Skill **`echo-ios-agent-vs-xcode`** | Agent vs Xcode ownership |
| Skill **`echo-phase1-validate`** | WO-230 detail |
