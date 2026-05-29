# ADR 0002 — Identity Metagraph Deployment: standalone Compose, not Hydra

- **Status:** Accepted
- **Date:** 2026-05-29
- **Deciders:** Platform / Metagraph
- **Related work order:** WO-273, WO-274 (Identity / VC bring-up)
- **Blocks / unblocks:** WO-272, WO-276, validate-phase1 Step 3
- **Supersedes:** N/A

## Context

Echo runs a second, dedicated **Identity Metagraph** (L0 + L1) alongside the
Euclid currency/data metagraph. It anchors VC issuance, trust-tier commitments,
StatusList2021 revocation, and EchoOrgRoleCredential state (WO-272). The
Scala modules live in `metagraph/modules/identity_l0` and
`metagraph/modules/identity_l1`:

- `identity_l0` extends `CurrencyL0App` — a full, independent L0 metagraph with
  its own genesis.
- `identity_l1` extends `CurrencyL1App` — an L1 that peers with **its own** L0
  (`identity_l0`), not the currency metagraph-l0.

Local bring-up was ambiguous and non-functional:

1. `metagraph/euclid.json` and the vendored `euclid-development-environment/euclid.json`
   listed `identity_l0` / `identity_l1` as hydra `framework.modules`, `docker`
   containers, and `layers` — implying hydra would orchestrate them.
2. `docker-compose.identity.yml` also existed, implying a separate deployment.
3. The compose file crash-looped: it used an invalid L0 subcommand
   (`run-initial-validator`), had no genesis flow, was missing
   `CL_GLOBAL_L0_PEER_ID` / `CL_L0_PEER_ID`, pointed L1 at the wrong L0 host, and
   assigned no static IPs.

Investigation of the stock Euclid SDK (euclid `v0.19.0`, Tessellation
`4.0.0-rc.0`) showed hydra is hard-coded to a **single metagraph**:

- `scripts/hydra-operations/logs.sh` enumerates `valid_layers` as exactly
  `global-l0 dag-l1 metagraph-l0 currency-l1 data-l1` — no `identity-*`.
- `infra/metagraph-base-image/Dockerfile` only builds `currencyL0/currencyL1/dataL1`
  (`SHOULD_BUILD_*`); there is no identity build path.
- The local ansible start playbooks (`start/{metagraph-l0,currency-l1}/…`) are
  templated for one metagraph instance.

The `identity_*` entries in `euclid.json` were therefore **dead config** — hydra
never read or acted on them.

## Decision

**Deploy the Identity Metagraph as a standalone `docker-compose.identity.yml`
stack that joins the running Euclid cluster's network and peers with the live
Global L0. Do not attempt to have hydra manage it.** Remove the dead
`identity_*` entries from both `euclid.json` files so the Compose stack is the
single source of truth.

### Concretely

1. **euclid.json** (both copies) drop `identity_l0` / `identity_l1` from
   `framework.modules`, `docker.default_containers`, and `layers`. The Scala
   `sbt` projects (`identityL0`, `identityL1` in `metagraph/build.sbt`) are
   independent of these entries and still build.

2. **Boot flow** mirrors Euclid's own metagraph-l0 genesis + currency-l1
   initial-validator templates for Tessellation `4.0.0-rc.0`, implemented in
   `scripts/identity/{l0,l1}-entrypoint.sh`:
   - **L0:** `create-genesis genesis.csv` → `cl-wallet create-owner-signing-message`
     → `run-genesis genesis.snapshot --metagraph-owner-message ./owner-message
     --ip <static-ip>` (foreground).
   - **L1:** `run-initial-validator --ip <static-ip>`.

3. **Global L0 peering.** `CL_GLOBAL_L0_PEER_ID` is read at runtime from the live
   `http://172.50.0.10:9000/node/info` rather than derived from a key — the
   Euclid dev keys are regenerated independently of any host `p12`, so
   key-derived ids are unreliable.

4. **Dedicated owner/node key.** L0 generates its own `identity.p12`
   (`cl-keytool generate`) persisted in a shared Docker volume. Reusing the
   cluster's `token-key.p12` fails Global L0 registration with
   `AddressAlreadyInUse` because that key already owns the currency metagraph.
   Persisting the key keeps the Identity metagraph id stable across restarts.

5. **Static IPs** on the external `custom-network` (subnet `172.50.0.0/24`;
   Euclid nodes use `.10/.20/.30`): Identity L0 = `172.50.0.50:9600`,
   Identity L1 = `172.50.0.51:9500`. L1 dials L0 at `172.50.0.50:9600`.

6. **Tooling.** `make start-identity` checks the assembly JARs, the
   `metagraph-base-image:latest`, and Global L0 reachability before starting,
   then waits for both L0 and L1. `scripts/validate-phase1.sh` and the Go
   backend (`pkg/credentials/config.go`, `IDENTITY_L0_URL` / `IDENTITY_L1_URL`,
   defaults `:9600` / `:9500`) already target these endpoints.

## Consequences

### Positive

- Identity L0 boots to `Ready`, registers its owner on the Global L0, and
  anchors snapshots upstream; Identity L1 boots to `Ready` and aligns global +
  metagraph snapshots. (Verified live against the `make dev` cluster.)
- One unambiguous mechanism — Compose — for a feature hydra cannot express.
- No forks/patches of the vendored Euclid SDK, so `hydra` updates and
  `setup-euclid.sh` re-runs don't re-break Identity.
- Reuses the Euclid base image (JARs, `cl-wallet`/`cl-keytool`, `genesis.csv`)
  so there is no second image to build/maintain.

### Negative

- Identity bring-up depends on the Euclid cluster already running (Global L0 on
  `:9000`) and on the `custom-network` existing. `make start-identity` fails
  fast with guidance when either is missing.
- The boot logic duplicates Euclid's ansible genesis sequence in shell; if
  upstream changes the L0 genesis CLI surface, `scripts/identity/*.sh` must be
  updated to match.
- Single-node L1 logs benign "Not enough peers" consensus messages (only one
  L1 validator in dev); harmless and expected.

### Neutral

- The Identity metagraph id changes only when the shared volume is wiped
  (`docker compose -f docker-compose.identity.yml down -v`); otherwise stable.

## Alternatives considered

### Option B — Make hydra manage Identity

Rejected. Stock hydra (`v0.19.0`) supports a single metagraph and a fixed layer
set; there is no extension point for a second metagraph. Supporting it would
require forking the Euclid base-image Dockerfile, the ansible local playbooks,
and the layer enumeration — a large, fragile change that would re-break on every
SDK bump and `setup-euclid.sh` re-run.

### Option C — Run Identity as a separate full Euclid clone

Rejected. A second Euclid environment would duplicate the Global L0 / DAG L1 and
its own network, and would not share the same Global L0 the rest of the cluster
uses — defeating the point of anchoring Identity state to the same chain.

## Implementation status

- [x] `identity_*` removed from both `euclid.json` files; Compose is source of truth
- [x] `scripts/identity/{l0,l1}-entrypoint.sh` implement the genesis / validator flow
- [x] `docker-compose.identity.yml` rewritten: genesis flow, live Global L0 peer
      id, dedicated key, static IPs, shared volume
- [x] `make start-identity` JAR + base-image + Global-L0 preflight and L0/L1 waits
- [x] `validate-phase1.sh` Identity messaging corrected (no `hydra status`)
- [x] Identity L0 + L1 verified `Ready` live against the `make dev` cluster

## References

- ADR 0001 — Phase 1 Identity Method: `did:key`
- Euclid `v0.19.0` ansible templates (`start/metagraph-l0/genesis.ansible.yml`,
  `start/currency-l1/initial-validator.ansible.yml`) — reference boot sequence
- `docker-compose.identity.yml`, `scripts/identity/l0-entrypoint.sh`,
  `scripts/identity/l1-entrypoint.sh`
- WO-272 — VC issuance, trust tier commitments, StatusList2021 (Identity Metagraph)
- WO-273 / WO-274 — Identity / VC bring-up
