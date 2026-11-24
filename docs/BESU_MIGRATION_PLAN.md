# Moving Echo's chain from Tessellation to Hyperledger Besu

**Written:** 2026-11-24 · **Status:** direction chosen by the owner on 2026-11-24; spike done (§2); stages B1–B5 not started and each needs approval at its gate

Supersedes the implementation stages of [`OWN_L0_PLAN.md`](OWN_L0_PLAN.md) (its §12 records why).
The goal of that plan is unchanged: Echo runs its own chain and depends on no outside network.

## 1. Why

Echo decided on 2026-11-17 to run its own validators. Tessellation is built to be run by
Constellation on Constellation's network; run alone, at a version Constellation itself does not run
in production (4.0.7), a three-node rehearsal showed three faults that need an operator or lose data
(`OWN_L0_PLAN.md` §12): joins that stall for good, a two-node outage that does not heal, and a
metagraph silently cut off from Global L0. Besu is built for the case Echo is now in — one
organisation running a small permissioned validator set — and has many independent operators.

## 2. Spike results (2026-11-24)

Code: [`spike/besu/`](../spike/besu/). Four Besu 26.11.2 validators in Docker, QBFT consensus, one
contract (`EchoAnchor.sol`, 54 lines) holding the Data L1 rules, a Go client using the same kind of
key the gateway already signs with (secp256k1).

| Question | Besu spike | Tessellation rehearsal (for comparison) |
|---|---|---|
| Time from start to a working network | About 25 seconds, four validators peered and producing blocks | About 10 minutes for 18 processes, nodes joined one at a time |
| Anchor a Merkle root | Final 1–2 seconds after submit. QBFT has no forks, so a block that exists is final | Final on Metagraph L0 in about 5 seconds; inclusion by Global L0 separate and later |
| Root from an unlisted key | Transaction reverted, root not stored | Rejected by Metagraph L0 |
| Proof checked with the chain off | Verifier checks **the validators' signatures** (3 of 4, quorum 3), the state proof and the stored entry. A tampered root, an untrusted validator set and a never-anchored root were each rejected | Verifier could not check signatures; it trusted the archived snapshot |
| One validator down | Anchoring continued | Continued |
| Below quorum (two of four down), then both restarted | Chain stopped, did not fork, and **resumed by itself 9 seconds after restart** | Did not recover; needed a whole-network rollback |
| A validator loses its disk | Rejoined and synced from nothing in 8 seconds, no action | Joins stalled about half the time until restarted |
| Every validator restarted in turn under load | 36 of 36 anchors accepted over 3 minutes, all four nodes at the same height | Not attempted |

### Limits of this evidence

- One machine, minutes of running, a chain 56 blocks long. Sync time on a chain months old, and
  behaviour across real network latency, are untested.
- The spike network is closed only by a static peer list. Node and account permissioning are not set up.
- Only the two Data L1 rules were ported. The six identity update types and everything about the
  ECHO token were not.
- I have run Besu for one afternoon. The comparison column comes from a week on Tessellation, so
  Besu has had less chance to show its faults.

## 3. Design

### Network

- Four validators, QBFT. Four is the smallest set that survives one failure (QBFT needs more than
  two thirds); three would stop on any single loss.
- Private network and firewall as designed for stage 2: no chain port public
  (`deploy/terraform/hetzner/gl0-nodes.tf` becomes the Besu node file).
- Zero gas price and zero base fee: there is no fee token, and the gateway is the only writer.
- Closed set: static peers plus Besu node permissioning. Validators are named in the genesis;
  changing them is a vote among validators.

### Contracts — the rules that are Scala today

| Today (Scala, `metagraph/modules/shared_data`) | On Besu |
|---|---|
| `MerkleRootUpdate`, `TrustCommitmentUpdate` (Data L1) | `EchoAnchor` — done in the spike |
| Six identity updates: VC issuance, trust-tier commitment, status list batch, org role credential, device key, username | `EchoIdentity`, one function each. **Change:** today these carry DIDs and usernames as text on-chain; on Besu they become 32-byte hashes, so the zero-PII rule holds for identity too. Needs a pass against `docs/data-classification.md` |
| Signer allowlists (`DATA_AUTHORIZED_SIGNERS`, `IDENTITY_SERVICE_PUBLIC_KEY_HEX`) | The contract's signer list. The gateway's existing secp256k1 key gives it an Ethereum address |
| PII pattern checks (email, phone, DID regexes) | Not needed: a `bytes32` field cannot hold them |
| Token locks, stake delegation, rewards, allow-spend, mint ban, founder revocation | See "The token" below |

Contracts stay small, have no upgrade proxy in v1 (a new version is a new address in gateway
config), and get a review before production genesis.

### The token — a decision for you (§6)

The ECHO ledger users see in v1 is already in Postgres (`internal/wallet/pg_store.go`); chain
submissions accompany it. v1 is value-free: no transfer, no redemption. So there are two ways:

- **A (recommended): v1 chain carries anchoring and identity only.** The value-free ledger stays in
  Postgres as it is today, and its state is anchored like everything else (a daily Merkle root of
  balances). A token contract comes with v1.1, when transfer is designed and counsel has ruled.
  Smallest migration; nothing a user sees changes.
- **B: port the token now** as a contract with transfers disabled, plus locks, delegation and
  rewards as contract functions. Adds an estimated one to two weeks and a contract that holds the
  whole supply and must be right.

### Gateway (Go)

- A chain interface in front of `internal/metagraph`, with a Besu implementation selected by
  `ECHO_CHAIN_NETWORK=besu` (the setting from stage 1; `sovereign` and `constellation` stay until
  Tessellation is removed).
- Anchoring service, proof store (migration 036) and the proof endpoint keep their shape. The proof
  bundle changes to header + state proof, and `cmd/verifyproof` gains signature checking.
- The finality tracker shrinks: finality is immediate, so it only archives headers.
- Several RPC endpoints per layer, tried in turn (the spike's load test did this), which removes
  the single-node dependency noted in stage 2.

### iOS

Correction to what I said on 2026-11-24: the app is affected. 31 Swift files reference the chain,
mostly the wallet (`Core/Stargazer/`): DAG addresses and Tessellation-format transaction signing.
Under token option A the wallet's on-chain signing is not used in v1 and can stay dormant behind
its existing flags; the visible change is that an account's address is no longer a DAG address.
Under option B the signing code is rewritten for Ethereum transactions. The keys are secp256k1 in
both systems, so no key changes; only the address format and the transaction encoding do. The
planned dependency on Constellation's iOS SDK for real-funds custody goes away.

### What carries over, what is retired

| Carries over | Retired |
|---|---|
| `ECHO_CHAIN_NETWORK`; Digital Evidence off by default | Scala metagraph (`metagraph/`), Euclid, hydra |
| Durable proofs (migration 036), proof endpoint, offline verifier (rewritten inside) | Vendored Tessellation binaries, the two forks, `make chain-supply-check` in its current form |
| Terraform private network and firewall; key ceremony pattern; Ansible skeleton and the "verify after deploy" idea | `deploy/sovereign/` launcher and rehearsal, Global L0 playbooks |
| OpenTimestamps daily checkpoint (stage 3 of the old plan) | DAG collateral plan; Constellation deploy path |
| `validate-phase1` (chain steps re-pointed) | — |

Supply chain: Besu is Apache-2.0. The same rule as stage 1 applies — pin a version, keep Echo's
own copy of the binaries with checksums, and be able to build from source.

## 4. Stages

Estimates are from reading the code and one afternoon with Besu. Treat them as guesses.

| Stage | What | Estimate | Gate |
|---|---|---|---|
| B0 | Spike | done | — |
| B1 | Chain interface; anchoring on Besu behind `ECHO_CHAIN_NETWORK=besu`; four-node local network replacing the hydra cluster in `make dev`; offline verifier with signature checks; `validate-phase1` chain steps on Besu | 4–6 days | `validate-phase1` passes on Besu, including a proof verified with the chain stopped |
| B2 | `EchoIdentity` contract; gateway identity client; data-classification review of the new on-chain shapes | 4–6 days | Identity steps of `validate-phase1` and the identity tests pass |
| B3 | Deploy: Terraform for four hosts, Ansible Besu role, key ceremony adapted, permissioning, monitoring, backup, restore drill, daily external checkpoint | 5–7 days | `validate-phase1` passes on staging; staging rebuilt from backup alone |
| B4 | Token, per your choice in §6: A = daily balance anchor; B = token contract and iOS signing | A: 2–3 days · B: 1–2 weeks | A: a balance proof verifies offline · B: token tests and a contract review |
| B5 | Remove Tessellation code and docs; 7-day soak with failure tests; launch review; production genesis | 7 days soak + 2 days | Your sign-off before genesis |

Total with option A: roughly three weeks of work plus the soak. B1 and B2 can overlap.

## 5. Risks and unknowns

| Risk | Mitigation |
|---|---|
| Besu has faults I have not met yet; one afternoon is thin evidence | B1's gate and the B5 soak, with the failure tests from the spike kept as a script (`spike/besu/failure-tests.sh`) and run again on staging |
| A contract bug is permanent on-chain | Keep contracts small; no token in v1 (option A); review before genesis; new version = new address |
| Chain growth (a block every 2 seconds when busy, every 30 when idle) | Measure disk use in the soak; set pruning and block period before genesis |
| Validator set changes are a vote, and a lost majority of keys halts the chain | Same key ceremony discipline as before; four keys in at least two locations |
| An ERC-20 style token is transferable by default (if option B) | Transfers disabled in the contract itself, not in the app; counsel reviews before any enablement |
| Schedule | Gates after B1 and B3 are the points to stop and reconsider |

## 6. Needed from you

| Decision | My recommendation |
|---|---|
| Token in v1: A (ledger stays in Postgres, anchored daily) or B (contract now) | A |
| Where the four validators run: four dedicated hosts, or fewer hosts shared with other workloads | Four small dedicated hosts; price them from the Hetzner Cloud Console, which I could not do reliably |
| The uncommitted stage 2 Tessellation work (launcher, Ansible, ceremony, Terraform) | Commit it once as a record, then retire it in B5. The Terraform and ceremony parts are reused |
| Whether a few weeks' delay to launch is acceptable | — |
