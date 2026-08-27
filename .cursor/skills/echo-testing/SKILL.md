---
name: echo-testing
description: >-
  Echo launch and regression testing — automated headless gates plus Xcode/Mac
  manual E2E. Use when preparing TestFlight, running regression before release,
  validating Phase 1/2/3 flows, or when the user asks for testing checklists,
  E2E on simulator/device, or make regression.
---

# Echo testing (launch + regression)

| Doc | Use |
|-----|------|
| **[`docs/E2E_TESTING.md`](../../docs/E2E_TESTING.md)** | **Canonical** — daily gates, Week A/B, sign-off, TestFlight |
| [`docs/E2E_TWO_DEVICE.md`](../../docs/E2E_TWO_DEVICE.md) | Two-device connectivity, logging, offline, restore |
| [`docs/GO_LIVE_SEPT_1_2026.md`](../../docs/GO_LIVE_SEPT_1_2026.md) | Sept 1 calendar, corporate, competitive readiness |
| Old `E2E_QUICK_START.md` / `E2E_LAUNCH_AND_TESTING.md` | Redirect stubs → `E2E_TESTING.md` |

## Agent workflow (iOS testing help)

When the user struggles with iOS E2E, run **in order**:

1. `make dev` (or confirm `health_backend`)
2. `make ios-preflight` — or MCP `run_ios_preflight`
3. If iOS code changed: `make ios-preflight BUILD=1 TESTS=1`
4. If onboarding/API fails: MCP `smoke_ios_backend`
5. Point user to **E2E_TESTING.md §5** (Week A checklist) — agents cannot tap Simulator UI

Do **not** dump the full E2E doc — use §1 daily gates + §5 Week A.

**Do not** suggest redesigning onboarding/login from the design prototype — those iOS flows are frozen ([`ECHO_IOS_UI_IMPLEMENTATION_SPEC.md`](../../docs/ECHO_IOS_UI_IMPLEMENTATION_SPEC.md) §0).

## When to run what

| Cadence | Automated | Manual (Mac + Xcode) |
|---------|-----------|----------------------|
| Every PR / push | CI: Go + iOS SPM | — |
| Pre-merge / weekly | `make regression` | Spot-check `E2E_TESTING.md` §5 |
| Before Xcode session | `make ios-preflight BUILD=1` | `E2E_TESTING.md` §5 |
| Pre-TestFlight | `make regression-with-phase1` | `E2E_TESTING.md` §5–§10 + device |

## Commands

```bash
make ios-preflight              # backend + Xcode + scheme (agents run this)
make ios-preflight BUILD=1 TESTS=1
make regression
make regression-quick
make regression-with-phase1
```

## MCP `echo-local-dev`

| Tool | Use |
|------|-----|
| `run_ios_preflight` | Before user opens Xcode |
| `smoke_ios_backend` | SMS / OIDC / health for iOS |
| `run_regression` | Full headless gate |
| `run_release_check` | Go only |
| `run_validate_phase1` | WO-230 go/no-go |
| `run_ios_phase3_tests` | Phase 3 logic |
| `health_backend` / `cluster_status` | Stack up? |

Reload MCP after pulling (`tools/echo-local-dev-mcp/README.md`).

## iOS — agent limits

- Scheme name: **EchoApp** (not Echo)
- Simulator `API_URL`: `http://localhost:8000` (in EchoApp.xcscheme)
- Device: LAN IP via scheme env
- Agents **cannot**: Simulator UI, Face ID, two-device chat, TestFlight install

## Related skills

| Skill | Use |
|-------|-----|
| `echo-phase1-validate` | WO-230 go/no-go |
| `echo-ios-agent-vs-xcode` | Agent vs Xcode ownership |
| `echo-phase3-ios-wire` | Phase 3 two-client WS |
