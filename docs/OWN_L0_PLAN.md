# Echo on its own Global L0 — independence plan

**Written:** 2026-11-16 · **Updated:** 2026-11-24 · **Status:** goal unchanged (Echo runs its own chain). **Implementation on Tessellation stopped at stage 2 on 2026-11-24**: the owner chose to move to Hyperledger Besu after the stage 2 findings (§12). Current plan: [`BESU_MIGRATION_PLAN.md`](BESU_MIGRATION_PLAN.md). Stages 3–4 below will not be carried out as written
**Goal:** Echo keeps working, with the same validators and the same user keys, if Constellation's
public network, company or hosted services stop being available.
**Related:** [`ZERO_TO_PRODUCTION.md`](ZERO_TO_PRODUCTION.md) §4 (current plan: integrationnet → mainnet),
[`V1_RELEASE.md`](V1_RELEASE.md), [ADR-0002](adr/0002-identity-metagraph-deployment.md).

## 1. What Echo depends on today

| Dependency | Where | If Constellation disappears |
|---|---|---|
| **Global L0 network** (their validators) | Planned production target: integrationnet, then mainnet (`metagraph/euclid.json` `deploy.network`) | The metagraph cannot finalise anything. Fatal. |
| **Mainnet admission** — seedlist registration, DAG collateral / staking, snapshot fees paid in DAG | `ZERO_TO_PRODUCTION.md` §4 | Launch is gated on a third party and on holding their token |
| **Tessellation source and release jars** (`global-l0.jar`, `cl-wallet.jar`, `cl-keytool.jar`) | Euclid's `infra/metagraph-ubuntu/Dockerfile` downloads from `github.com/Constellation-Labs/tessellation` | The cluster image can no longer be built |
| **`tessellation-sdk` 4.0.0-rc.0** | `metagraph/project/Dependencies.scala`, from Maven Central | Maven Central artefacts are permanent, but there are no fixes |
| **Euclid / hydra tooling and the g8 template** | `../euclid-development-environment` | Build and deploy scripts stop being maintained |
| **Hosted services** — block explorer (`be-*.constellationnetwork.io`), Digital Evidence (`internal/evidence/client.go`) | Verification steps in the deploy doc; evidence feature | Lose the public explorer; evidence anchoring stops |
| **Stargazer wallet SDK** | `ios/Echo/Sources/Core/Stargazer/` — a protocol and a backend-proxied bridge; the real SDK is not integrated | Nothing breaks today |

What is *not* a dependency: the Go gateway talks to the chain through plain HTTP URLs
(`GLOBAL_L0_URL`, `CURRENCY_L1_URL`, `DATA_L1_URL`, `IDENTITY_L*_URL`), and user wallet keys and
`DAG…` addresses are derived on the device. Neither changes when the network underneath changes.

Tessellation and Euclid are Apache-2.0 (checked 2026-11-16), which permits forking, modifying and
running a private network.

## 2. The key fact: nothing is on a public network yet

Echo has never produced a snapshot on integrationnet or mainnet. The cluster that passed
`make validate-phase1` (26/26) **is** a private network with its own Global L0. So:

- Launching on your own L0 needs **no migration** — there is no state to move.
- Launching on Constellation first and leaving later means moving token balances, anchored
  roots and identity state, and keeping old proofs verifiable after the old chain is gone.

That makes "now" materially cheaper than "later", which is unusual and is the main input to §5.

## 3. Phase A — supply-chain hedge (do regardless; about 2–4 days)

Cheap, reversible, and useful under every option.

1. **Mirror the source.** Fork `tessellation` (at the pinned tag), `euclid-development-environment`
   and the `currency` g8 template into Echo's own GitHub organisation. Tag what production uses.
2. **Vendor the binaries.** Store `global-l0.jar`, `cl-wallet.jar`, `cl-keytool.jar` and the
   `tessellation-sdk` artefacts in Echo's own artefact store, with checksums committed to this repo.
3. **Build from the mirror.** Point the image build at the fork and prove it once with network
   access to Constellation's GitHub blocked. Until that passes, the hedge is not real.
4. **Name the network in config.** One setting, `ECHO_CHAIN_NETWORK = sovereign | constellation`,
   selecting the URL set, fee handling and explorer links, so switching is a config change.
5. **Make proofs self-contained.** An anchor proof should carry the snapshot, its signatures and the
   validator key set, and every snapshot should be archived to Echo's own object storage. Then a
   proof verifies with no live chain at all. (Today the finality index is in memory and the proof
   stores only a transaction hash.)
6. **Stop relying on the hosted services for v1.** Put Digital Evidence behind a flag, off by default.
   Replace explorer links with Echo's own read API.
7. **Pin a stable Tessellation release** instead of `4.0.0-rc.0` before any production genesis.

## 4. Phase B — the sovereign network (about 2–3 weeks, including soak)

Same code and the same Scala validators; Echo runs the Global L0 too.

| Work | Notes |
|---|---|
| Global L0 cluster, 3 nodes, Echo-held keys | Can share the three existing metagraph hosts or use three more; separate hosts are safer |
| Seedlist of Echo's own peer IDs | Closed validator set at launch |
| Production genesis | Real ECHO allocation — today's cluster uses the template's 30,000 tokens |
| Metagraph L0, Currency L1, Data L1, Identity L0/L1 | Already in terraform and compose; Identity L1 already runs three nodes |
| Authorised-signer check on Data L1 | Already listed as open in `V1_RELEASE.md` |
| Key ceremony and custody | Validator keys, genesis owner key, snapshot-fee keys; offline backups |
| Snapshot archive, backup and rollback runbook | `hydra start-rollback` is the mechanism; needs a tested restore |
| Monitoring and alerting | Snapshot ordinal advancing on every layer, peer counts, disk |
| Upgrade procedure | Echo now owns Tessellation upgrades and security patches |
| Minimal explorer | Read-only API over snapshots; a UI can wait |
| **External checkpoint** | Publish each day's Global L0 snapshot hash somewhere Echo does not control (see §6) |

**Spikes to run before committing to dates — I have not verified these:**

- Multi-node Global L0 genesis outside hydra's local defaults (the local cluster runs one GL0 node).
- Which `CL_APP_ENV` a private production network should use; the local cluster runs `dev`, and the
  built-in environments carry Constellation-specific parameters.
- Collateral and snapshot-fee behaviour on a network where Echo sets the rules.
- Licences of any ancillary pieces Echo wants to reuse (explorer, wallet SDK).

## 5. Options

| | **A. Sovereign from day one** | **B. Hedge now, launch on Constellation, migrate when needed** | **C. Do nothing until forced** |
|---|---|---|---|
| Before v1 launch | Phase A + Phase B | Phase A only | Nothing |
| Extra time before launch | About 3–4 weeks | About 2–4 days | None |
| Launch gated on Constellation (seedlist, DAG collateral, fees) | No | Yes | Yes |
| Migration later | None | About 3–6 weeks, with a write freeze and user-visible risk (§7) | Same, under time pressure, possibly from a dead network |
| Public-network credibility at launch | None; mitigated by the external checkpoint | Yes, while Constellation lasts | Yes, while Constellation lasts |
| Ongoing cost | Run and patch the Global L0 | Their fees and rules | Their fees and rules |

Time figures are estimates from reading the code and tooling, not from having done it.

**Recommendation: A, provided the §4 spikes pass; otherwise B.** Reasons:

- v1's chain use is invisible to users and the token is deliberately value-free, so the thing a
  public network adds — outside validators and a market — is not something v1 uses.
- A removes three launch dependencies (seedlist approval, holding DAG for collateral, paying fees).
- A is the only option with no migration, and the configuration is the one already tested.
- If Constellation thrives, joining it later is additive: anchor Echo's snapshot hashes to it.

The honest cost of A: Echo is the only validator, so "on-chain" means "signed by Echo" until the
external checkpoint and, later, independent validators exist. Say that plainly in marketing.

Phase A is worth doing this week under every option.

## 6. Keeping a self-run chain honest

A chain where one company runs every node is tamper-*evident* only if someone else holds a copy of
its history. In rising order of effort:

1. Publish the daily Global L0 snapshot hash to a public, append-only place Echo cannot rewrite
   (a public blockchain transaction or a timestamping service).
2. Let anyone run a read-only node that follows and verifies the chain.
3. Admit independent validators to the seedlist — partners first, then open. This is the step that
   turns a hedge into a network, and it is business development, not engineering.

## 7. If Echo launches on Constellation first (option B) — the migration runbook

1. **Stand up the sovereign network in parallel** (Phase B) with an empty genesis.
2. **Dual-write anchors** to both networks for a soak period; compare roots daily.
3. **Freeze** token-affecting writes (stakes, claims). Anchoring continues on the new network.
4. **Export** final state from the old metagraph: balances, locks, anchored roots, identity state.
5. **Genesis with state** on the new network: balances via `genesis.csv`; data-application state
   needs a "genesis from export" path that does not exist yet (`genesis` returns empty state today).
6. **Flip** `ECHO_CHAIN_NETWORK`. Addresses and keys are unchanged, so no user action is needed.
7. **Keep old proofs verifiable** from the snapshot archive (Phase A item 5). Without that archive,
   every proof issued on the old network becomes unverifiable when the old network is gone.

Step 5 and step 7 are the hard parts, and step 7 only works if Phase A was done *before* launch.

## 8. Decision

**2026-11-17: option A — Echo launches on its own Global L0.** The implementation plan is §9.
No implementation starts until that plan is approved. Option B stays available as the fallback
until the production genesis (§9 stage 4); nothing before that point is hard to undo.

## 9. Implementation plan for option A

Five stages. Each ends in a gate with a pass test. Durations are estimates from reading the code
and tooling, not from having done this before.

| Stage | What | Estimate | Gate |
|---|---|---|---|
| 0 | Spikes: prove the unknowns | 2–3 days | Go / no-go on option A |
| 1 | Supply-chain independence | 3–4 days | Chain builds with no access to Constellation |
| 2 | Production-shaped network on staging | 5–7 days | `validate-phase1` passes against staging |
| 3 | Operations: monitoring, backup, checkpoint | 3–4 days, overlaps stage 2 | Restore drill passes |
| 4 | Soak, then production genesis | 7 days soak + 1 day | Your sign-off before genesis |

Total: about 3–4 weeks. Stage 1 can run in parallel with stage 0.

### What I checked in the tooling on 2026-11-17

- Euclid's **remote** deploy playbooks cover Metagraph L0, Currency L1 and Data L1 only. There is
  no Global L0 playbook; they expect an existing network and take its node from
  `euclid.json` `deploy.network.gl0_node`. The Global L0 deployment has to be written by Echo.
- The remote playbooks set `CL_APP_ENV` to the network name (`integrationnet` or `mainnet`).
  The local playbooks set `CL_APP_ENV=dev` and `CL_COLLATERAL=0` on every layer.
- The local cluster starts one Global L0 node with `run-genesis genesis.csv`. Nothing in the
  tooling starts a second Global L0 node.
- `deploy/terraform/hetzner/metagraph-nodes.tf` deliberately does not provision a Global L0 host.
- `configs/dag-staking-plan.json` plans 750,000 DAG of mainnet collateral. Option A removes that
  requirement.

### Stage 0 — spikes (go / no-go)

All run on the local machine. No production resources.

| # | Question | Pass test | If it fails |
|---|---|---|---|
| S1 | Can three Global L0 nodes form a cluster from Echo's own genesis? | Three nodes `Ready`; snapshots keep advancing with any one node stopped, including the genesis node; a stopped node rejoins | Launch with one Global L0 node plus a warm standby and a tested restore. Weaker, and I would ask you again before accepting it |
| S2 | Which `CL_APP_ENV` suits a private production network? | Read the Tessellation source for what each environment changes; pick one; the full cluster runs under it with Echo's seedlist enforced | Keep `dev` with an explicit seedlist and closed firewall, or carry a small patch in Echo's fork |
| S3 | How do collateral and snapshot fees behave on Echo's network? | The metagraph's snapshots are accepted for 24 hours with the collateral and fee settings chosen for production | Fund the fee wallet in the Global L0 genesis |
| S4 | Is there a stable Tessellation release to pin instead of `4.0.0-rc.0`? | Scala tests and `validate-phase1` pass on it | Stay on the release candidate, pinned in Echo's fork |
| S5 | Are the extras Echo wants to reuse (explorer, monitoring service) licensed for it? | Licence read and recorded here | Write the minimal read API instead; already planned |

**Gate:** S1–S3 pass, or their fallbacks are accepted by you. Otherwise revert to option B; the
stage 1 work is still used.

### Stage 1 — supply-chain independence

| Task | Where | Done when |
|---|---|---|
| Fork `tessellation`, `euclid-development-environment` and the `currency` template into Echo's GitHub organisation; tag the versions in use | GitHub (your account — see "Needed from you") | Forks exist with an `echo-prod-<version>` tag |
| Store `global-l0.jar`, `cl-wallet.jar`, `cl-keytool.jar` and the `tessellation-sdk` artefacts in Echo's storage; commit checksums | New `metagraph/vendor/SHA256SUMS`; release assets on the fork plus a copy in object storage | Checksums verified at build time |
| Point the image build at the fork | Fork of Euclid: `infra/metagraph-ubuntu/Dockerfile` (two hard-coded Constellation URLs today) | No `Constellation-Labs` URL left in the build path |
| Point Echo's setup at the fork | `metagraph/scripts/setup-euclid.sh`, `Makefile` | `make dev` clones and builds from Echo's fork |
| Guard against regression | New `make chain-supply-check` | Fails if a Constellation URL reappears, or a checksum differs |
| Name the network in config | Go config: `ECHO_CHAIN_NETWORK = sovereign \| constellation`; `deploy/k8s/base/configmap.yaml`; `.env.example` | One setting selects URLs, fee handling and explorer links |
| Turn off Constellation-hosted services | `internal/evidence/client.go` behind `DIGITAL_EVIDENCE_ENABLED`, default off; hide the iOS evidence entry points in Release | Gateway makes no request to `*.constellationnetwork.io` in a test run |
| Self-contained anchor proofs | New migration `036_anchor_proofs.sql`; snapshot archiver writing to object storage; `GET /v1/anchors/{root}/proof`; small offline verifier in `cmd/` | A proof verifies with the chain switched off |
| Persist the finality index | Same table; replaces the in-memory index (open item in `V1_RELEASE.md`) | Finality survives a node restart |

**Gate:** from a clean clone, the chain images build and `validate-phase1` passes with outbound
access to Constellation's GitHub blocked.

### Stage 2 — production-shaped network, on staging first

| Task | Where | Notes |
|---|---|---|
| Provision Global L0 hosts | New `deploy/terraform/hetzner/gl0-nodes.tf` | Recommended: three dedicated hosts, separate from the three metagraph hosts. Alternative: co-locate on the existing three and save the cost |
| Firewall | Same terraform | Peer ports open only between Echo's nodes; public ports only to the gateway and the read API |
| Global L0 playbooks: genesis, validator join, rollback | New `deploy/ansible/global-l0/` | Echo-written; Euclid has none for remote |
| Seedlist | Generated from Echo's node IDs, committed | Closed set at launch |
| Point the metagraph deploy at Echo's Global L0 | `metagraph/euclid.json` `deploy.network`; forked remote playbooks if the S2 answer needs it | Replaces `integrationnet` |
| Identity L0 / L1 on the same network | Existing compose and entrypoints, per ADR-0002 | Three Identity L1 nodes already work locally |
| Production genesis files | Global L0 `genesis.csv`; metagraph ECHO allocation | Replaces the template's 30,000 tokens. Needs your numbers |
| Authorised-signer check on Data L1 | `metagraph/modules/` validations | Open item in `V1_RELEASE.md`; without it any party that reaches the port can submit |
| Key ceremony | Runbook in `docs/` | One key per node, a genesis owner key, a fee key. Generated offline, two offline backups, loaded to hosts from the secrets store |
| Gateway config | `deploy/k8s` overlays | URLs for the sovereign network |
| Retire the Constellation path in docs | `ZERO_TO_PRODUCTION.md` §4, `configs/dag-staking-plan.json`, comment in `metagraph-nodes.tf` | Marked superseded, not deleted, so option B stays documented |

**Gate:** `make validate-phase1` passes 26/26 against the staging network, with three Global L0
nodes and the production environment setting.

### Stage 3 — operations

| Task | Done when |
|---|---|
| Monitoring and alerts: snapshot ordinal advancing on every layer, peer count, disk, clock drift | An alert fires in a test where one node is stopped |
| Backups: snapshot archive to object storage in a second location | Archive is current to within one hour |
| Restore drill using `hydra start-rollback` and the new Global L0 rollback playbook | Staging rebuilt from backup alone, state matches |
| External checkpoint: daily job publishing the Global L0 snapshot hash outside Echo's control | Seven consecutive daily checkpoints, each verifiable by a third party |
| Minimal explorer: read-only API over snapshots and anchored roots | Replaces every explorer link in the deploy docs |
| Upgrade runbook: how Echo takes a Tessellation patch from the fork to production | One upgrade rehearsed on staging |

### Stage 4 — soak and production genesis

1. Seven-day soak on staging with production topology and synthetic traffic.
2. Failure tests during the soak: stop one Global L0 node, stop the genesis node, fill a disk,
   restore from backup.
3. Launch review with you: results of the above, open risks, key custody confirmed.
4. **Production genesis.** This is the one step that is hard to undo: once users anchor data,
   restarting from a new genesis discards it. I will ask you before running it.

### Needed from you

| Decision or action | Needed by | My recommendation |
|---|---|---|
| GitHub organisation to hold the forks | Stage 1 start | The organisation that already owns this repo |
| Where vendored binaries live | Stage 1 | Release assets on the fork, plus a copy in the existing object storage |
| Three extra hosts for Global L0, or co-locate | Stage 2 start | Three extra. I have not priced them; I will confirm the cost before provisioning |
| ECHO genesis allocation | Stage 2 | Needs your tokenomics numbers; counsel-gated per the launch posture |
| Who holds the offline key backups | Stage 2 | Two people or two locations, not one |
| Where the daily checkpoint is published | Stage 3 | A Bitcoin timestamp through OpenTimestamps: no token to hold, no account. I will verify its current terms before building on it |
| Sign-off before production genesis | Stage 4 | — |

### Out of scope for v1

- Outside validators and an open seedlist (§6 step 3).
- An explorer user interface.
- Any bridge or anchor to Constellation's network.
- Making ECHO transferable; the value-free launch posture is unchanged.

### Risks

| Risk | Effect | Mitigation |
|---|---|---|
| Multi-node Global L0 does not work outside Constellation's own setup | Single point of failure, or option A fails | Spike S1 runs first; fallback is stated |
| Echo is now the only party patching Tessellation | A consensus bug or security fix is Echo's to apply | Pin a stable release; upgrade runbook; keep the fork able to pull upstream while upstream exists |
| Key loss or theft | Loss: network cannot be restarted. Theft: history can be forged | Offline ceremony, two backups, external checkpoint limits what a thief can rewrite |
| "On-chain" claims read as more than they are | Trust and possibly regulatory exposure | State plainly that Echo runs the validators; publish the checkpoint |
| Schedule: 3–4 weeks is an estimate | Launch slips | Gates after stages 0 and 2 give two early points to fall back to option B |

## 10. Stage 0 results (2026-11-17) — gate passed

Test cluster: `deploy/sovereign/gl0/` (three Global L0 containers on one machine, own network and
volumes, host ports 19000 / 19010 / 19020; it does not touch `make dev`). Tessellation 4.0.0-rc.0.

| # | Result | Evidence |
|---|---|---|
| S1 multi-node Global L0 | **Pass** | Three nodes `Ready` from Echo's own genesis; every snapshot carries three signatures. Validator stopped: snapshots paused about 90 seconds, then continued on two nodes. Genesis node stopped: same. Each stopped node rejoined in about 3–3.5 minutes and all three then reported the same snapshot hash. A fourth node with a key not on the seedlist was not admitted |
| S2 environment | **Answered: `dev`, with an explicit seedlist** | Read from the Tessellation source. The other three environments switch protocol features on at Constellation's own snapshot ordinals (in the millions), and `mainnet` adds a fixed 250,000 collateral and Constellation's own allowance list. A chain starting at ordinal 0 would run outdated rules under them. `dev` activates everything from ordinal 0. What `dev` relaxes: a seedlist is optional (Echo passes `--seedlist` and `--prioritySeedlist`, which are enforced in every environment), the metagraph owner message is optional at genesis, loopback joins are allowed, and logging is verbose |
| S3 collateral and fees | **Pass** | `dev` has a base snapshot fee of 0 and Echo sets collateral to 0. The existing local cluster has run three days under these settings and its Global L0 is still accepting the metagraph's snapshots |
| S4 stable release | **Partly verified** | 4.0.1–4.0.7 exist as release tags and on Maven Central. All Scala tests pass against 4.0.7 (0 failures). Not yet run live: the cluster rebuild on 4.0.7 and `validate-phase1` happen in stage 1 |
| S5 licences | **Done** | Tessellation, Euclid and the metagraph monitoring service are Apache-2.0. I did not find a public explorer repository to reuse; the minimal read API stays in the plan |

Limits of this evidence:

- One machine. Three hosts with real network latency are first exercised in stage 2.
- The metagraph was not attached to the three-node Global L0; that is stage 2.
- Two of three nodes down at once was not tested. Expect the network to stop until one returns.
- Follow-up for stage 2: a production log configuration (the `dev` default is verbose), and a
  longer-term option of adding an `echo` environment in the fork so `dev` is not the production label.

### Decisions recorded 2026-11-17

| Decision | Value |
|---|---|
| Forks | New repositories under the owner's GitHub account |
| Vendored binaries | Release assets on the Tessellation fork (primary), a copy in Echo's object storage, checksums in this repo |
| Global L0 hosts | Prove with Docker first; price dedicated hosts at the start of stage 2 |
| Offline key backups | Two holders (the owner and one family member), two separate locations in California |
| Daily checkpoint | Bitcoin timestamp through OpenTimestamps; terms to be re-checked when stage 3 starts |
| ECHO genesis allocation | See below |

### ECHO genesis allocation (recommended and adopted, subject to counsel)

Mint the documented fixed supply — 1,000,000,000 ECHO, 8 decimals (`configs/genesis.yaml`) — at
genesis, into five addresses that Echo holds, one per documented bucket. Nothing goes to an
individual at genesis.

| Bucket | ECHO | Genesis holder |
|---|---|---|
| Community rewards | 400,000,000 | Rewards emitter key (online; the only bucket v1 spends from) |
| Treasury | 220,000,000 | Offline reserve key |
| Founders | 180,000,000 | Offline reserve key, undistributed |
| Future team and advisors | 100,000,000 | Offline reserve key, undistributed |
| Ecosystem and partnerships | 100,000,000 | Offline reserve key, undistributed |

Why: it keeps the published "fixed supply, no minting after genesis" promise, and it keeps the v1
value-free posture, because personal founder and team allocations are the part that implies value
and is waiting on counsel. Splitting the reserve buckets later is an ordinary transfer, not a new
genesis. The addresses come from the stage 2 key ceremony, so the genesis file is written then.

## 11. Stage 1 results (2026-11-19) — gate passed

| Item | State | Evidence |
|---|---|---|
| Pin a stable Tessellation | Done: 4.0.7 | SDK, node binaries and all five `Main.scala` version strings agree; `make chain-supply-check` enforces it |
| Vendored binaries | Done locally | Built from source at tag `v4.0.7` (commit `a4b07efe`); checksums in `metagraph/vendor/SHA256SUMS-v4.0.7`; includes a Maven-layout archive of the SDK |
| Image build without Constellation | Done | Both chain images built with `--no-cache` and every GitHub host unreachable. `metagraph/euclid-overlay/` replaces Euclid's node Dockerfile; `hydra install` is no longer run |
| Regression guard | Done | `make chain-supply-check` |
| Live validation on 4.0.7 | Done | From-scratch `make dev` + `make start-identity`, then `make validate-phase1`: 26 passed, 0 failed |
| Network setting | Done | `ECHO_CHAIN_NETWORK` (`sovereign` default, `constellation`); any other value stops the gateway at startup |
| Constellation-hosted services off | Done | Digital Evidence needs `DIGITAL_EVIDENCE_ENABLED=true` as well as its API key. The two iOS evidence types have no callers, so nothing in the app reaches them |
| Durable proofs and finality | Done | Migration `036_anchor_proofs.sql`; proofs and root finality are in Postgres; a tracker archives the signed metagraph and Global L0 snapshots to object storage |
| Self-contained proof | Done, with limits below | `GET /v1/anchors/{root}/proof`; `cmd/verifyproof`. Verified live: with the chain stopped, the verifier confirmed the archived snapshot contains the root and its signer ids are trusted; a tampered bundle was rejected |
| Forks under the owner's GitHub account | Done | Public forks `c1cromwell/tessellation` and `c1cromwell/euclid-development-environment`, carrying the tags in use. Binaries published as release `echo-v4.0.7` on the Tessellation fork; `metagraph/vendor/fetch.sh` downloaded them into an empty directory and all five checksums matched |

Two facts found on the way:

- Constellation publishes no binaries for 4.0.1–4.0.7 — only source tags and the SDK — and its
  mainnet reported version 3.6.2 on 2026-11-19. Echo's binaries therefore have to be built from
  source; that is now scripted (`metagraph/vendor/build-from-source.sh`).
- Tessellation 4.0.7 changed two things Echo's code met: an assembly merge conflict
  (fixed in `metagraph/build.sbt`) and snapshot reads now need an `Accept` header (fixed in the
  finality tracker, with a regression test). Unit tests caught neither; the image build and the
  live run did.

Limits of the proof, stated plainly:

- The verifier checks that the archived snapshot contains the root and that its signers are on the
  trusted list. It does **not** verify the validators' signatures; that needs Tessellation's own
  serialization. Until that exists, the proof is only as strong as the integrity of the archive —
  which is what the stage 3 daily external checkpoint protects.
- The archived Global L0 snapshot only bounds the time; nothing yet links it to the metagraph snapshot.
- Proofs enter through real message traffic. The live test injected the validation run's root by
  hand; a message sent between two devices has not exercised this path.

Carried into stage 2: `METAGRAPH_L0_URL` in the k8s configmap (the tracker is idle without it);
an object-storage copy of the vendored binaries as a second mirror; the Euclid checkout ships
well-known sample validator keys (`source/p12-files/`), which the key ceremony replaces; one
`dagexplorer.io` link remains in `internal/wallet/service.go`.

## 12. Stage 2 progress (2026-11-24) — gate not yet met

The gate is `validate-phase1` 26/26 against **staging**. Staging hosts do not exist yet, so the
gate is open. Everything below was built and exercised on one machine.

### What was run

`make sovereign-up` starts the production shape in Docker: three Global L0 nodes and three
metagraph hosts, each host running Metagraph L0, Currency L1, Data L1, Identity L0 and Identity L1
(18 node processes). Every process is started by the launcher that the real hosts will use
(`deploy/sovereign/node/echo-node.sh`), with seedlists enforced and the signer allowlists set.

| Check | Result |
|---|---|
| All 18 nodes reach `Ready` from three geneses (Global L0, ECHO metagraph, Identity metagraph) | Pass |
| `make validate-phase1` against it | 26 passed, 0 failed |
| Update signed by an unlisted key (`TestSignerAllowlistRejectsUnknownSigner`) | Not finalized; Metagraph L0 logged the block as invalid, naming the signer check |
| One Global L0 node stopped for 5 minutes, from steady state | Global snapshots continued with 2 signers; both metagraphs stayed fully included (0–1 snapshots behind); an anchor submitted meanwhile finalized; the node was `Ready` again about 3 minutes after restart |
| Two Global L0 nodes down at once (by my mistake: I stopped a second while the first was rejoining) | **Network stalled and did not recover when both returned.** The survivor stopped producing snapshots and the returning nodes waited on it. Recovered with a whole-network restore from the newest snapshot; metagraph nodes kept running and their backlog was included within about 10 minutes |
| Identity L0 as a three-node cluster | Works. Before this it had only run as a single node |
| Global L0 keeps including each metagraph's snapshots after the deploy | **Failed once, for the ECHO metagraph, on the Ansible-deployed network.** Every node was `Ready` and the deploy reported success, but nothing after snapshot 34 was included. Recovered by hand; about 40 snapshots were discarded. See "A metagraph can be cut off" below |

### Built

| Plan item | State |
|---|---|
| Authorised-signer check on Data L1 | Done. `DATA_AUTHORIZED_SIGNERS`, enforced in Metagraph L0 `validateData`; every proof must be from a listed key. The launcher refuses to start Metagraph L0 or Identity L0 without the list |
| Provision Global L0 hosts; firewall | Written, `terraform validate` passes, **never applied**. `deploy/terraform/hetzner/gl0-nodes.tf`: private chain network `10.0.2.0/24`, fixed addresses, no chain port public, SSH only. `gl0_colocate=true` puts Global L0 on the metagraph hosts instead |
| Global L0 playbooks: genesis, join, rollback | Written in `deploy/ansible/` and run end to end against six systemd containers standing in for hosts, with keys from a throwaway ceremony: all 18 services `Ready` under a non-root account, both metagraph ids recorded, owner keys removed from the host after genesis, total supply read back as exactly 1,000,000,000 ECHO. One join stalled (see below) and was restarted automatically. The rollback playbook was run against the same hosts: it chose the node with the newest snapshot, restored there, rejoined the other two and removed the rollback setting. `verify.yml` (new) passes |
| Seedlist | Generated by the key ceremony script into `deploy/sovereign/networks/<network>/`; committed after the real ceremony |
| Point the metagraph deploy at Echo's Global L0 | Done differently from the plan: `hydra remote-deploy` accepts only `integrationnet` / `mainnet` and checks Constellation's seedlist, so Echo's own Ansible deploys the metagraph layers too. `euclid.json` `deploy.network` is no longer used for this |
| Identity L0 / L1 on the same network | Done in the rehearsal and in the playbooks |
| Production genesis files | The ceremony script writes the Global L0 genesis and the 1,000,000,000 ECHO genesis (five addresses, §10) and refuses any other total. Real files come from the real ceremony |
| Key ceremony | Script `deploy/sovereign/ceremony/ceremony.sh` and runbook `docs/KEY_CEREMONY.md`. Run once for a throwaway network; the real one is yours to run |
| Gateway config | `deploy/k8s/base/configmap.yaml` points at the fixed private addresses and sets `METAGRAPH_L0_URL` |
| Retire the Constellation path in docs | `ZERO_TO_PRODUCTION.md` §4–5, `configs/dag-staking-plan.json`, `metagraph-nodes.tf` marked superseded, not deleted |
| Carried from stage 1 | `METAGRAPH_L0_URL` set; `dagexplorer.io` link now empty unless `ECHO_CHAIN_NETWORK=constellation`; production log configuration (INFO, to journald). **Not done:** object-storage mirror of the vendored binaries |

### Found on the way

- **Seedlists have two jobs.** Global L0 accepts a metagraph's snapshot only when one of its
  signers is on Global L0's own seedlist, so the metagraph node ids must be on it. A metagraph node
  key could therefore also join Global L0 as a validator. Separately, metagraph nodes accept a
  global snapshot only when the majority of the *priority* list agrees; that list is the three
  Global L0 nodes. Both lists are now explicit in the launcher.
- **A metagraph genesis needs a running Global L0.** It cannot be made at the offline ceremony, so
  the deploy makes it on the first metagraph host and removes the owner key from the host afterwards.
- **A join can stall for good — a Tessellation 4.0.7 defect.** A joining node downloads the newest
  snapshot and then asks a peer for that exact consensus round. If the cluster has moved on by one
  round, the peer answers `409 Conflict`; the joiner retries the same stale round 20 times
  (`fetchOutcomeFromCluster`), then stays in `Observing` for ever. It happened on roughly half of
  the metagraph joins here. Only a restart clears it, so the launcher stops a node that is not
  `Ready` 7 minutes after joining and its supervisor starts it again. The proper fix is in Echo's
  Tessellation fork (download again on a conflict); not done.
- **A metagraph can be cut off from Global L0 while looking healthy. This is the most important
  finding of the stage.** On the Ansible-deployed network, as the second Metagraph L0 node became a
  validator, the two nodes produced different signed versions of snapshots 33–34 (one carried both
  signatures, one only its own). Global L0 accepted the second node's version of 34. That node then
  left and rejoined on its own, and the first node carried on from *its* version. From then on
  every snapshot named a parent Global L0 had never accepted, and Global L0 dropped them without a
  message at the production log level. All nodes stayed `Ready`; the metagraph's own ordinal kept
  rising; `validate-phase1` step 5 would still have passed, because it asks Metagraph L0, not Global L0.
  - What I know: the mechanism above, from the binary hashes on both sides. Restarting one node
    does not help. Stopping all three Metagraph L0 nodes and starting the first again (it resumes in
    rollback mode) does: it restarted from snapshot 35 and was included at once, and stayed in step
    while the other two rejoined.
  - What that costs: the snapshots made while cut off (36–75 here) are discarded. Roots anchored
    in that window would have to be submitted again; the gateway does not do that today.
  - What I do not know: why the two nodes diverged. It happened in one of two deployments (not in
    the Docker rehearsal). It followed a stalled join that the launcher had restarted, which may or
    may not be related. I have not reproduced it on purpose.
  - What is in place: `deploy/ansible/playbooks/verify.yml`, run at the end of `site.yml`, fails
    the deploy if either metagraph is more than 5 snapshots behind on Global L0, and
    `deploy/sovereign/stack/lag.sh` shows the same for the rehearsal. Continuous alerting on this
    lag is stage 3 and should be treated as required before any real traffic.
- **The tools need `CL_APP_ENV=dev` too.** Without it `create-genesis` assumes Constellation's
  testnet rules and fails with a misleading "No majority peers" error.
- **Hetzner's prices and orderable server types changed several times in 2026** and the sources I
  found disagree with each other. I removed the default Global L0 server type instead of guessing.
  The hosts are not priced.

### Not yet done, or not verified

- Nothing has run on real hosts: no Terraform apply, no real network latency, no host firewall
  (`ufw` cannot run in the test containers).
- The real key ceremony for staging.
- The gateway reaches one node per layer. If that node is down the gateway loses that layer.
- Two-of-three Global L0 loss needs an operator. Alerting on it is stage 3.
- `validate-phase1` was not re-run after the whole-network restore, and was not run against the
  Ansible-deployed network (its ports are not published to the gateway in the container test).
- `validate-phase1` step 5 checks finality on Metagraph L0 only. It should also require Global L0
  inclusion; until it does, a 26/26 result does not rule out the cut-off described above.
- Recovery of a cut-off metagraph is a manual procedure (in `deploy/ansible/README.md`), not a playbook.
- The gateway does not re-submit roots that were anchored in snapshots later discarded.

### Needed from you to reach the gate

| Item | Why |
|---|---|
| Choose: three dedicated Global L0 hosts, or co-locate on the three metagraph hosts | Three or six servers. Co-location saves three; a lost host then takes a Global L0 node and a metagraph node together |
| Server types and a cost you accept, from the Hetzner Cloud Console | I could not confirm current prices or availability |
| Run the staging key ceremony (`docs/KEY_CEREMONY.md`) | Keys must not pass through me |
| Hetzner API token at apply time, and a go-ahead to provision | Provisioning spends money |
