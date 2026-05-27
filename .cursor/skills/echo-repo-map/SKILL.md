---
name: echo-repo-map
description: >-
  Maps the Echo monorepo stacks, docs, Makefile targets, and phase sources of truth.
  Use when starting any Echo task, exploring the codebase, or when the user mentions
  phases, metagraph, backend, iOS, work orders, or TestFlight.
---

# Echo repo map

## Stacks

| Stack | Path | Build / test |
|-------|------|--------------|
| Go API | `internal/api/`, `internal/auth/`, `pkg/` | `make build`, `make test`, `go test ./...` |
| Scala metagraph | `metagraph/modules/` | `make metagraph-test`, `cd metagraph && sbt test` |
| iOS | `ios/Echo/Sources/` | Xcode or `cd ios/Echo && swift build` |
| iOS tests | `ios/Echo/Tests/`, `SecurityTests/`, `Phase3Tests/` | `swift test --filter EchoPhase3Tests` |

Entry: `main.go` (backend), `ios/Echo/EchoApp.xcodeproj` (app), `ios/Echo/Package.swift` (SPM library).

## Phase docs (read the relevant one only)

| Phase | WO list | Actionable extras |
|-------|---------|-------------------|
| 1 | `docs/phase-1-work-orders.md` | `docs/E2E_LAUNCH_AND_TESTING.md`, `docs/data-classification.md` |
| 2 | `docs/phase-2-work-orders.md` | `docs/PHASE2_GAP_AUDIT.md` |
| 3 | `docs/phase-3-work-orders.md` | `docs/PHASE3_IOS_UI_SPEC.md` |
| 4–7 | `docs/phase-4-work-orders.md` … `phase-7` | Blueprint refs in WOs |

**ADR:** `docs/adr/0001-phase1-identity-method.md` — Phase 1–2 use **`did:key`**, not Cardano/PRISM.

## Makefile (common)

| Target | Purpose |
|--------|---------|
| `make dev` | Backend Docker stack (:8000) |
| `make dev-status` | Health of stack components |
| `make validate-phase1` | WO-230 six-step go/no-go |
| `make metagraph-test` | Scala validator tests |
| `make release-check` | Pre-ship Go gate |
| `make start-identity` | Identity L0/L1 (optional) |

## Software Factory MCP

Use `software-factory-echo` for WO status: `whoami` → `list_work_orders` → `read_work_order` → `edit_work_orders`.

Do not treat `docs/phase-*-work-orders.md` status headers as authoritative if they conflict with Software Factory or gap audits.

## Obsolete patterns (do not implement)

- `did:prism:cardano:`, Atala PRISM, Cardano anchoring — **Phase 1–2**
- WO-180, WO-182 as written — superseded by WO-273, WO-274
- `WSRelayMessage` for typing/read_receipt/reaction — use `WSEnvelope` in `ConversationSignal.swift`

## Agent entry

Repo root [`AGENTS.md`](../../AGENTS.md). Tooling roadmap: [`docs/AGENT_TOOLING_RECOMMENDATIONS.md`](../../docs/AGENT_TOOLING_RECOMMENDATIONS.md).
