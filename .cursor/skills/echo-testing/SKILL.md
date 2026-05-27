---
name: echo-testing
description: >-
  Echo launch and regression testing — automated headless gates plus Xcode/Mac
  manual E2E. Use when preparing TestFlight, running regression before release,
  validating Phase 1/2/3 flows, or when the user asks for testing checklists,
  E2E on simulator/device, or make regression.
---

# Echo testing (launch + regression)

**Canonical doc:** [`docs/TESTING.md`](../../docs/TESTING.md)

## When to run what

| Cadence | Automated | Manual (Mac + Xcode) |
|---------|-----------|----------------------|
| Every PR / push | CI: Go + iOS SPM (`EchoSecurityTests`, `EchoPhase3Tests`) | — |
| Pre-merge / weekly | `make regression` or `./scripts/run-regression.sh` | Spot-check one §5 flow |
| Pre-TestFlight / launch | `make regression-with-phase1` (if metagraph up) | **All** §5–7 checklists |
| Post-deploy regression | `curl …/health` + `make regression --quick` | TestFlight §7 on physical device |

## Automated regression (agent runs these)

```bash
# Default — Go release-check + targeted suites + iOS SPM (needs Xcode.app on Mac)
make regression

# Fast — Go race tests only
make regression-quick

# Include WO-230 validate-phase1 (Docker + Euclid + JDK 21)
make regression-with-phase1
```

Script: `scripts/run-regression.sh` (`--quick`, `--with-phase1`, `--ios-only`).

MCP **`echo-local-dev`**: `run_release_check`, `run_validate_phase1`, `run_ios_phase3_tests`, `health_backend`, `cluster_status`.

## iOS automated (Mac + Xcode required)

Agent **cannot** substitute for full Xcode.app (CLT-only fails `gomobile` / app archive).

```bash
cd ios/Echo
DEVELOPER_DIR=/Applications/Xcode.app/Contents/Developer \
  swift build --target Echo --target EchoSecurityTests --target EchoPhase3Tests
swift test --filter EchoSecurityTests
swift test --filter EchoPhase3Tests

# App target compile (no launch)
xcodebuild -project EchoApp.xcodeproj -scheme Echo \
  -destination 'platform=iOS Simulator,name=iPhone 16 Pro' \
  -configuration Debug build

# Live OPRF (WO-221) — once per machine / after mobile/echooprf changes
./scripts/build-echooprf-ios.sh
```

## Manual E2E — agent does NOT run these

Use checklists in `docs/TESTING.md`:

| Section | Scope |
|---------|--------|
| §5 | Simulator — onboarding, login, messaging, Phase 3 signals, WO-100 VC, WO-221 PSI |
| §6 | Physical iPhone — LAN `API_URL`, biometrics, two-device PSI |
| §7 | TestFlight build regression |

Prerequisites for manual runs:

```bash
make dev && curl -s http://localhost:8000/health   # backend
# Device: API_URL = http://<Mac-LAN-IP>:8000 (PHASE1_LAUNCH §1c)
# WO-100: OIDC4VC_ENABLED=true in .env
# WO-221: ./scripts/build-echooprf-ios.sh + embed EchoOPRF.xcframework
```

## Launch sign-off template

Copy from `docs/TESTING.md` §8. All automated gates green **and** required manual sections checked before enabling TestFlight external beta.

## Related skills & docs

| Resource | Use |
|----------|-----|
| `echo-phase1-validate` | WO-230 go/no-go, TestFlight prep |
| `echo-ios-agent-vs-xcode` | Agent vs Xcode ownership |
| `echo-phase3-ios-wire` | Phase 3 two-client WS E2E |
| `docs/PHASE1_LAUNCH.md` | Signing, TestFlight upload |
| `docs/metagraph-backend-e2e-testing.md` | Metagraph manual steps |
| `docs/PHASE3_IOS_UI_SPEC.md` | Phase 3 Step 5 detail |
