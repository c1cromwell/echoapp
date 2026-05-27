# Agent Tooling Recommendations — Echo

**Created:** 2026-05-26  
**Status:** Accepted (Steps 1–4 implemented 2026-05-26)  
**Purpose:** Record planning choices for Cursor skills, MCP services, and agent harnesses so future you (and agents) know what was decided and why.

---

## Executive summary

Echo is a **multi-stack, phase-gated** product (Go backend, Scala metagraph, iOS, Docker/Euclid, Software Factory work orders). The repo already has strong phase docs and the `software-factory-echo` MCP, but agents lacked repeatable harnesses for WO/doc ↔ code sync, local cluster validation, stack-specific patterns, and agent-vs-Xcode work splits.

**Decision:** Invest first in **project skills** (`.cursor/skills/`) and **`AGENTS.md`**, then **`echo-local-dev` MCP** (Step 3 in rollout). Defer custom mega-agents and duplicate work-order systems.

---

## What we already had (baseline)

| Asset | Role |
|-------|------|
| `docs/phase-{1–7}-work-orders.md` | Phase backlog (~278 WOs) |
| `docs/E2E_LAUNCH_AND_TESTING.md`, `PHASE2_GAP_AUDIT.md`, `PHASE3_IOS_UI_SPEC.md` | Launch / gap / feature specs |
| `docs/data-classification.md` + `.semgrep/t0_t7_rules.yaml` | T0–T7 privacy invariant |
| `docs/adr/0001-phase1-identity-method.md` | did:key; no Cardano Phase 1–2 |
| `software-factory-echo` MCP | WOs, blueprints, requirements |
| `Makefile` | `dev`, `validate-phase1`, `metagraph-test`, `release-check` |
| CI | Go (+ T0–T7), iOS, metagraph, cross-platform crypto |
| Isolated iOS test targets | `EchoSecurityTests`, `EchoPhase3Tests` |

---

## Observed friction (why tooling was needed)

1. **WO status drift** — Software Factory, `phase-*-work-orders.md`, and code/CHANGELOG disagree.
2. **Obsolete WO traps** — Cardano/PRISM WOs (180, 182, 132) vs did:key replacements (273, 274).
3. **Agent vs Xcode split** — Logic in SPM; SwiftUI + `EchoApp.xcodeproj` need Xcode (see Phase 3 pattern).
4. **Heavy local env** — `validate-phase1` needs Docker + Euclid + JDK 21; agents cannot fully sign off E2E headlessly.
5. **Doc volume** — `Echo_Combined_Requirements.md` is too large for default context; use gap audits and feature specs.
6. **Dual iOS types** — `Message`/`DeliveryStatus` vs mock `ChatMessage`/`MessageStatus`.

---

## Skills — tiered plan

### Tier 1 — Implemented (Step 1, 2026-05-26)

| Skill | Location | Trigger |
|-------|----------|---------|
| `echo-repo-map` | `.cursor/skills/echo-repo-map/` | Any multi-stack task |
| `echo-auth-contracts` | `.cursor/skills/echo-auth-contracts/` | Auth, WS, signed REST |
| `echo-ios-agent-vs-xcode` | `.cursor/skills/echo-ios-agent-vs-xcode/` | iOS features |

Also added: **`AGENTS.md`** at repo root (entry point for agents).

### Tier 1 — Implemented (Step 2, 2026-05-26)

| Skill | Location | Trigger |
|-------|----------|---------|
| `echo-work-order-sync` | `.cursor/skills/echo-work-order-sync/` | Update WOs, sync Software Factory after merges |
| `echo-t0-t7-review` | `.cursor/skills/echo-t0-t7-review/` | Go/Scala changes touching chain or submissions |

Phase doc headers (`docs/phase-*-work-orders.md`) synced with Software Factory same day.

### Tier 2 — Implemented (Step 4, 2026-05-26)

| Skill | Location | Trigger |
|-------|----------|---------|
| `echo-phase1-validate` | `.cursor/skills/echo-phase1-validate/` | TestFlight / Phase 1 go/no-go |
| `echo-phase2-gaps` | `.cursor/skills/echo-phase2-gaps/` | Onboarding, contacts, VC (see gap audit) |
| `echo-phase3-ios-wire` | `.cursor/skills/echo-phase3-ios-wire/` | Xcode wiring after agent lands messaging logic |
| `echo-metagraph-scala` | `.cursor/skills/echo-metagraph-scala/` | Scala L1 validator changes |

### Tier 3 — Optional (not yet implemented)

| Skill | When |
|-------|------|
| `echo-openapi-drift` | New REST handlers vs `openapi.yaml` |
| `echo-commit-style` | Commit message conventions |

---

## MCP services — tiered plan

### Tier 1 — Implemented (Step 3, 2026-05-26)

**`echo-local-dev` MCP** — [`tools/echo-local-dev-mcp/`](../tools/echo-local-dev-mcp/), registered in [`.cursor/mcp.json`](../.cursor/mcp.json):

| Tool | Wraps |
|------|--------|
| `cluster_status` | `make dev-status` |
| `run_release_check` | `make release-check` |
| `run_validate_phase1` | `make validate-phase1` |
| `run_metagraph_test` | `make metagraph-test` |
| `run_ios_phase3_tests` | `swift test --filter EchoPhase3Tests` |
| `health_backend` | `GET /health` |

### Tier 2 — Consider

- OpenAPI contract MCP (`openapi.yaml` ↔ handlers)
- Postgres read-only (dev) for reactions/DID/refresh tokens
- GitHub MCP for PR/CI (pairs with `babysit` skill)

### Explicitly deferred

- Full Euclid/hydra MCP (use `make dev` + docs)
- Auto-edit all 278 WOs without human review
- Duplicate Software Factory

---

## Agent harnesses

| Harness | Role |
|---------|------|
| WO reconciler | Agent + `echo-work-order-sync` + SF MCP after merge windows |
| Phase 1 gate | Agent + `echo-local-dev` MCP → GO/NO-GO report |
| PR babysitter | Existing Cursor **`babysit`** skill on CI failures |
| Phase 3 iOS wire | Human/Xcode + **`echo-phase3-ios-wire`** skill |
| Security reviewer | Agent + `echo-t0-t7-review` before chain-touching PRs |

---

## Doc hygiene decisions

| Source of truth | Use for |
|-----------------|---------|
| **Software Factory** | WO status, assignee, relationships |
| **Gap audits** (`PHASE2_GAP_AUDIT.md`, etc.) | What is actually built vs stub |
| **Feature specs** (`PHASE3_IOS_UI_SPEC.md`, etc.) | How to implement a feature |
| **ADRs** | Architecture decisions (did:key, no Cardano P1–2) |
| **`phase-*-work-orders.md`** | Human-readable export; may drift — sync periodically |

Do **not** load `Echo_Combined_Requirements.md` or full blueprints by default in agent sessions.

---

## Rollout schedule (accepted)

```text
Step 1 (done 2026-05-26): AGENTS.md + echo-repo-map + echo-auth-contracts + echo-ios-agent-vs-xcode
Step 2 (done 2026-05-26): echo-work-order-sync + echo-t0-t7-review; sync phase doc headers
Step 3 (done 2026-05-26): echo-local-dev MCP (tools/echo-local-dev-mcp + .cursor/mcp.json)
Step 4 (done 2026-05-26): echo-phase1-validate, echo-phase2-gaps, echo-phase3-ios-wire, echo-metagraph-scala
```

---

## Changelog

| Date | Change |
|------|--------|
| 2026-05-26 | Initial recommendations documented; Step 1 skills + AGENTS.md implemented |
| 2026-05-26 | Step 2: echo-work-order-sync + echo-t0-t7-review; phase WO headers synced |
| 2026-05-26 | Step 3: echo-local-dev MCP server + project `.cursor/mcp.json` |
| 2026-05-26 | Step 4: four phase-specific skills (validate, phase2 gaps, phase3 wire, metagraph) |
