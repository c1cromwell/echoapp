---
name: echo-metagraph-scala
description: >-
  Echo Constellation metagraph Scala L1 validators and tests. Use when editing
  metagraph/modules, Identity/Data L1 validation logic, sbt tests, or WO-272/277/274.
---

# Echo metagraph (Scala)

Path: `metagraph/modules/`. Identity method: **`did:key`** only on Phase 1 L1 ([ADR-0001](../../docs/adr/0001-phase1-identity-method.md)).

## Build & test

```bash
make metagraph-test          # sharedData + identityL0 + identityL1 (JDK 21 preferred)
make metagraph-verify-skeleton   # static WO-276 check (jq, no sbt)
cd metagraph && sbt test     # full suite (slower)
```

MCP: `run_metagraph_test`.

JDK 21: Makefile auto-detects Homebrew `openjdk@21`. Set `JAVA_HOME` if sbt fails.

## Module map

| Module | Role |
|--------|------|
| `shared_data` | Shared types + **`IdentityValidations.scala`** |
| `identity_l0` | Identity L0 node |
| `identity_l1` | Identity L1 — DID register, VC, StatusList2021 submissions |
| `data_l0` / `data_l1` | Data layer Merkle / commitments |
| `currency_l0` / `currency_l1` | ECHO token L1 |

Validators live under `modules/shared_data/src/main/scala/com/echo/shared_data/validations/`.

## IdentityValidations rules (Phase 1)

File: `IdentityValidations.scala`

- **DIDs:** must be `did:key:z...` (`DidKeyPrefix`)
- **Usernames:** `^[A-Za-z0-9_]{3,30}$`; on-chain key = lowercased
- **Trust commitments:** 64-char hex (32 bytes)
- **StatusList2021:** bit vector length 131072
- **Authorized sender:** Identity Service DID (`IDENTITY_SERVICE_DID`) for Phase 1
- Pure functions: `Either[String, Unit]` — L1 app enforces sender

Tests: `IdentityValidationsSpec.scala`, `ValidationsSpec.scala`.

## Local cluster (not managed by this skill)

Metagraph runtime via Euclid **`hydra`** in sibling `../euclid-development-environment`:

```bash
make dev              # hydra start-genesis + backend
make dev-status       # endpoint reachability
make start-identity   # custom Identity L0/L1 Docker (needs sbt assembly JARs)
```

Ports: Global L0 `:9000`, Metagraph L0 `:9200`, Currency `:9300`, Data `:9400`, Identity L1 `:9500`, Identity L0 `:9600`.

## T0–T7 before chain changes

Run skill **`echo-t0-t7-review`** on any new L1 payload fields:

- No PII in `IdentityUpdate` case classes
- Merkle roots = 32-byte SHA-256
- Trust tier on-chain = hash commitment, not raw integer

## Common WOs

| WO | Topic |
|----|-------|
| 272 | Identity L1 validation deploy |
| 277 | Scala unit tests for L1 |
| 274 | VC 2.0 + StatusList2021 (Go issues; Scala validates) |
| 276 | Metagraph skeleton verify |

## Do not implement

- Cardano on-chain schemas (WO-20, WO-37) — blocked; Constellation + did:key path
- Cardano trust scoring (WO-12) — blocked

## Related

- `metagraph/scripts/setup-euclid.sh`, `verify-identity-skeleton.sh`
- Go publishers: `pkg/credentials/identity_l1_publish.go`
- Skill: `echo-t0-t7-review`, MCP `echo-local-dev`
