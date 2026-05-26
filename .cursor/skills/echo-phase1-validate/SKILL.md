---
name: echo-phase1-validate
description: >-
  Phase 1 TestFlight and go/no-go validation for Echo. Use when preparing
  TestFlight builds, running validate-phase1, checking dev cluster health,
  or answering "is Phase 1 ready to ship?"
---

# Echo Phase 1 validate

Primary doc: [`docs/PHASE1_LAUNCH.md`](../../docs/PHASE1_LAUNCH.md). WO: **WO-230**.

## Gate order (run top to bottom)

```bash
make release-check                                    # Go build + race tests + vet + fmt
cd ios/Echo && DEVELOPER_DIR=/Applications/Xcode.app \
  swift build --target Echo --target EchoSecurityTests
make dev && make dev-status                           # cluster up
curl -s http://localhost:8000/health                  # → operational
make validate-phase1                                  # 6-step WO-230 (needs Docker + JDK 21 + Euclid)
```

Use MCP **`echo-local-dev`**: `run_release_check`, `cluster_status`, `health_backend`, `run_validate_phase1`.

## validate-phase1 steps (must all be `ok` before TestFlight)

| Step | What |
|------|------|
| 0 | Prerequisites (docker, jq, curl, go, jdk) |
| 1 | Local `did:key` derivation |
| 2 | Register DID with backend |
| 3 | Identity Metagraph L0/L1 reachable |
| 4 | Test message through relay |
| 5 | Merkle root submitted to Data L1 |
| 6 | Global L0 snapshot height increments |

Script: `scripts/validate-phase1.sh`. Steps 3/5 may **skip** when Euclid is down — fine for simulator-only; **all ok** required for ship.

## Device testing

Physical iPhone cannot use `localhost`. Set `API_URL` in Xcode build settings / `.env` per `PHASE1_LAUNCH.md` §1c (LAN IP of dev machine).

Optional Identity nodes for VC/trust tier:

```bash
make start-identity   # L0 :9600, L1 :9500
```

## iOS hard gates (CI)

| Target | Role |
|--------|------|
| `EchoSecurityTests` | WO-208/211/223/224 — must pass in Xcode CI |
| `EchoPhase3Tests` | Phase 3 signals — separate from broken `EchoTests` |

## Common blockers

| Symptom | Fix |
|---------|-----|
| Euclid dir missing | `cd metagraph && ./scripts/setup-euclid.sh` |
| Backend unhealthy | `make dev-logs`, check Postgres/Redis |
| Identity optional skip | `make start-identity` after `sbt assembly` |
| iOS build fails on Previews | Full **Xcode.app** required, not CLT only |

## Do not confuse with

- **Phase 2** onboarding gaps — see `echo-phase2-gaps`
- **Cardano/PRISM** — not Phase 1; ADR-0001

## Related

- [`CONTRIBUTING.md`](../../CONTRIBUTING.md) — first-time dev setup
- Skill: `echo-local-dev` MCP + `echo-ios-agent-vs-xcode`
