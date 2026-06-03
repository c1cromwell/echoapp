# Echo — Agent guide

Quick entry point for Cursor agents working in this repo. Read this first, then load the relevant **project skill** under `.cursor/skills/`.

## Start here

| Task | Read / use |
|------|------------|
| Any work in this repo | Skill: **`echo-repo-map`** |
| Auth, WebSocket, signed API | Skill: **`echo-auth-contracts`** |
| iOS implementation | Skill: **`echo-ios-agent-vs-xcode`** |
| Update work order status | Skill: **`echo-work-order-sync`** + MCP `software-factory-echo` |
| Chain / submission / PII review | Skill: **`echo-t0-t7-review`** + `docs/data-classification.md` |
| Local cluster / validate / release check | MCP **`echo-local-dev`** (see `.cursor/mcp.json`) |
| TestFlight / Phase 1 go-no-go | Skill: **`echo-phase1-validate`** + [`docs/E2E_LAUNCH_AND_TESTING.md`](docs/E2E_LAUNCH_AND_TESTING.md) |
| Launch & regression testing (full matrix) | Skill: **`echo-testing`** + [`docs/E2E_QUICK_START.md`](docs/E2E_QUICK_START.md) (start here) / [`docs/E2E_LAUNCH_AND_TESTING.md`](docs/E2E_LAUNCH_AND_TESTING.md) |
| Phase 2 onboarding / VC / contacts | Skill: **`echo-phase2-gaps`** |
| Phase 3 messaging UI (Xcode wiring) | Skill: **`echo-phase3-ios-wire`** |
| UI/UX design + full iOS feature map | [`docs/ux-spec.md`](docs/ux-spec.md) + [`docs/ECHO_IOS_UI_IMPLEMENTATION_SPEC.md`](docs/ECHO_IOS_UI_IMPLEMENTATION_SPEC.md) |
| Design prototype (React, **post-auth only**) | [`docs/Echo Design System Setup_latest/`](docs/Echo%20Design%20System%20Setup_latest/) — **not** onboarding/login |
| Metagraph Scala L1 validators | Skill: **`echo-metagraph-scala`** |

Planning history and MCP roadmap: [`docs/AGENT_TOOLING_RECOMMENDATIONS.md`](docs/AGENT_TOOLING_RECOMMENDATIONS.md).

## MCP servers

| Server | Purpose |
|--------|---------|
| `software-factory-echo` | Work orders, blueprints, requirements (user/global config) |
| `echo-local-dev` | `make dev-status`, `release-check`, `validate-phase1`, `metagraph-test`, iOS Phase 3 tests, `/health` |

Reload MCP in Cursor after pulling. Setup: [`tools/echo-local-dev-mcp/README.md`](tools/echo-local-dev-mcp/README.md).

## Architecture (one screen)

```text
echoapp/
├── cmd/                    CLI utilities (didkey, credentials)
├── internal/api/           Go HTTP + WebSocket gateway (:8000)
├── internal/auth/          JWT, passkey validation, rate limits
├── pkg/didkey/             did:key derivation (Phase 1 identity)
├── pkg/credentials/        VC 2.0 + OIDC4VC
├── metagraph/              Scala L1 validators (Identity, Data, Currency)
├── ios/Echo/               SwiftPM library + EchoApp.xcodeproj
├── scripts/validate-phase1.sh   WO-230 go/no-go
└── docs/                   Phase WOs, launch guides, gap audits, ADRs
```

**Phase 1 identity:** `did:key` only — **no Cardano / Atala PRISM** in Phase 1–2 ([ADR-0001](docs/adr/0001-phase1-identity-method.md)).

## Source-of-truth hierarchy

1. **Software Factory** (`software-factory-echo` MCP) — WO status and descriptions  
2. **Gap audits** — e.g. [`docs/PHASE2_GAP_AUDIT.md`](docs/PHASE2_GAP_AUDIT.md) — code vs WO reality  
3. **Feature specs** — e.g. [`docs/PHASE3_IOS_UI_SPEC.md`](docs/PHASE3_IOS_UI_SPEC.md)  
4. **`phase-*-work-orders.md`** — export; may drift; sync via `echo-work-order-sync`

Avoid loading [`docs/Echo_Combined_Requirements.md`](docs/Echo_Combined_Requirements.md) unless the user asks — use phase docs instead.

## Common commands

```bash
make release-check          # Go build + test + vet + fmt
make dev                    # Docker backend stack
make validate-phase1        # 6-step E2E (needs cluster + JDK 21)
make metagraph-test         # Scala L1 tests
cd ios/Echo && swift test --filter EchoPhase3Tests
```

Full setup: [`CONTRIBUTING.md`](CONTRIBUTING.md). TestFlight & E2E: [`docs/E2E_LAUNCH_AND_TESTING.md`](docs/E2E_LAUNCH_AND_TESTING.md).

## Agent constraints

- **Headless agents** cannot run Xcode UI or live WebSocket E2E against a LAN backend — land logic + unit tests; leave SwiftUI wiring to Xcode.
- **Frozen iOS onboarding & login** — `FirstRunCoordinator`, `GlacialLoginScreen`, and related auth views are the correct design. Do **not** redesign or map from `docs/Echo Design System Setup_latest/.../onboarding/*`. Post-auth UI may follow the prototype per [`docs/ECHO_IOS_UI_IMPLEMENTATION_SPEC.md`](docs/ECHO_IOS_UI_IMPLEMENTATION_SPEC.md) §0.
- **Do not re-implement** passkey signing on REST; use existing `PasskeySigningInterceptor`.
- **Ephemeral WS signals** use `WSEnvelope` shape with mandatory `to` (peer DID) — not `WSRelayMessage`.
- **Minimize scope** — match existing patterns; no Cardano code in Phase 1–2 paths.

## Project skills (`.cursor/skills/`)

| Skill | Status |
|-------|--------|
| `echo-repo-map` | ✅ Step 1 |
| `echo-auth-contracts` | ✅ Step 1 |
| `echo-ios-agent-vs-xcode` | ✅ Step 1 |
| `echo-work-order-sync` | ✅ Step 2 |
| `echo-t0-t7-review` | ✅ Step 2 |
| `echo-phase1-validate` | ✅ Step 4 |
| `echo-phase2-gaps` | ✅ Step 4 |
| `echo-phase3-ios-wire` | ✅ Step 4 |
| `echo-metagraph-scala` | ✅ Step 4 |
