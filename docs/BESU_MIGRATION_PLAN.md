# Moving Echo's chain from Tessellation to Hyperledger Besu

**Written:** 2026-11-24 · **Status:** decisions final (§0); stages B1 and B2 built and their gates met locally (§7, §9); B3–B5 not started

Supersedes the implementation stages of [`OWN_L0_PLAN.md`](OWN_L0_PLAN.md) (its §12 records why).
The goal of that plan is unchanged: Echo runs its own chain and depends on no outside network.

## 0. Decisions (owner, 2026-11-24) — final

| Decision | Detail |
|---|---|
| Echo's chain is Hyperledger Besu | Private network, QBFT consensus, four validators run by Echo. Tessellation and Constellation are no longer part of Echo |
| The ECHO token is on-chain from genesis | A token contract deployed at genesis holds the full 1,000,000,000 supply. Reason given: there are no users yet, so there is nothing to migrate later and no period in which balances exist only in a database. This replaces my recommendation (ledger in Postgres for v1) |
| Tessellation and Constellation code is removed | Done on 2026-11-24 for everything that could go without breaking the gateway (§8). The rest is replaced stage by stage and listed there |

### Revised the same day: the ECHO token is on hold

After weighing the legal read (a non-transferable token is low risk; the risk sits in founder and
team allocations, investment vocabulary, any hint of future value, paid VIP tied to tokens, and
Apple's rule against crypto for tasks), the owner decided:

| Decision | Detail |
|---|---|
| No ECHO token for users, for now | No mention and no use of an ECHO token in the app, the API, store metadata or public documents. It may return later as an additional feature |
| The supply is still created at genesis, in one place | 1,000,000,000 units minted once to a single address called **reserved**. No "founders", "team", "treasury" or other labelled allocations. Nothing can move out of it: the contract has no transfer and no mint function |
| VIP is a paid set of features and nothing else | No token benefit, multiplier, allowance or balance is attached to VIP |
| Constellation Digital Evidence stays, for Comply | Comply is a separate product that may need a public chain. It remains off by default in the messaging gateway |

This replaces the token design described under "The token" in §3 and shrinks stage B3 (§4). The
five-way allocation table in `OWN_L0_PLAN.md` §10 is superseded.

What "on-chain from genesis" committed Echo to under the earlier decision (kept for the record;
the points about a token contract holding user balances no longer apply):

- The token contract must be right before production genesis. It holds the whole supply, and a
  deployed contract cannot be edited. It gets its own review (stage B3 gate).
- The launch posture is unchanged and is enforced **in the contract**: no transfer between users,
  no redemption, no minting after genesis. Enabling transfer later is a deliberate contract
  change after counsel has ruled, not a setting.
- Whether a deployed token changes how regulators see a value-free leaderboard is a question for
  counsel. I have flagged it; I cannot answer it.
- The iOS wallet's signing code is rewritten for this chain (stage B3), and every account gets an
  address on it.

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

### The token — on-chain from genesis (decided, §0)

- One contract holds the supply: 1,000,000,000 ECHO, 8 decimals, minted once in the constructor
  to the five allocation addresses (community rewards 400M, treasury 220M, founders 180M, team and
  advisors 100M, ecosystem 100M). There is no mint function.
- No transfer between users in v1. Balances change only through the functions Echo's rules need:
  reward payout from the community-rewards allocation, lock and unlock for staking tiers,
  delegation, the VIP allowance. Each is a contract function with the rule from
  `Validations.scala` (git history) behind it.
- The gateway's Postgres ledger becomes a cache of the chain, as it was designed to be; the chain
  is the record.
- Who signs: rewards are paid by the gateway's key; a user's own actions (lock, delegate) are
  signed on the device with the account key and relayed by the gateway.

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
| B0 | Spike | done 2026-11-24 | — |
| B1 | Anchoring on Besu: contract, gateway client with fail-over, durable chain proofs, offline verifier with signature checks, four-validator local network in `make dev`, `validate-phase1` on Besu | done 2026-11-24 (§7) | Met locally |
| B2 | Identity on-chain: `EchoIdentity` contract, records as hashes; gateway and credentials package on it; data-classification review; group and governance records as digests | done 2026-11-24 (§9) | Met locally |
| B3 | Token on-chain from genesis: token contract (fixed supply, no transfer, no mint), locks, delegation, rewards, VIP allowance; gateway wallet and rewards on the contract; iOS wallet signing and addresses; contract review | 2–3 weeks | Token tests pass; supply and allocations read back exactly; independent review of the token contract done |
| B4 | Deploy: Terraform for four hosts, Ansible Besu role, key ceremony, permissioning, monitoring, backup, restore drill, daily external checkpoint | 5–7 days | `validate-phase1` and `failure-tests.sh` pass on staging; staging rebuilt from backup alone |
| B5 | Remove the last Tessellation-era code and documents; 7-day soak with failure tests; launch review; production genesis | 7 days soak + 2 days | Your sign-off before genesis |

Total remaining: roughly five to six weeks of work plus the soak. B2 and B3 can overlap; B3 is the long one and the one with the most risk.

## 5. Risks and unknowns

| Risk | Mitigation |
|---|---|
| Besu has faults I have not met yet; one afternoon is thin evidence | B1's gate and the B5 soak, with the failure tests from the spike kept as a script (`spike/besu/failure-tests.sh`) and run again on staging |
| A contract bug is permanent on-chain | Keep contracts small; no token in v1 (option A); review before genesis; new version = new address |
| Chain growth (a block every 2 seconds when busy, every 30 when idle) | Measure disk use in the soak; set pruning and block period before genesis |
| Validator set changes are a vote, and a lost majority of keys halts the chain | Same key ceremony discipline as before; four keys in at least two locations |
| An ERC-20 style token is transferable by default (if option B) | Transfers disabled in the contract itself, not in the app; counsel reviews before any enablement |
| Several gateway replicas share one signing key (staging runs 2, production 3), and each numbers its own transactions | Handled in the client since B2: a writer whose transaction number was taken by another re-signs with a fresh one. Tested with three writers at once (18 writes, each landed once). Under heavy contention writes get slower, not lost; one key per replica remains the better design if write volume grows, and is a B4 option |
| The first write after every validator restarts at once took more than 30 seconds, once | Inclusion timeout raised to 90 seconds; measure properly in the B5 soak |
| Schedule | Gates after B1 and B3 are the points to stop and reconsider |

## 6. Needed from you

| Item | Needed by | Note |
|---|---|---|
| Counsel's view on a deployed, non-transferable token under the value-free launch posture | Before B3's gate | One question; the contract design does not depend on the answer, the launch does |
| Where the four validators run, and server types from the Hetzner Cloud Console | B4 | I could not price Hetzner reliably |
| Who reviews the token contract | B3 gate | It should not be only me: I wrote it |
| Run the key ceremony | B4 | Keys must not pass through me |

## 7. Stage B1 results (2026-11-24) — gate met locally

Gate: `validate-phase1` passes on Besu, including a proof verified with the chain stopped.

| Check | Result |
|---|---|
| `make dev` from nothing | Four validators, contract deployed, backend anchoring, in under two minutes |
| `make validate-phase1` | 26 passed, 0 failed, 1 skipped. The skip is step 3 (identity commitment), which has no on-chain form until B2 |
| Step 5 in detail | Root anchored through the gateway, final 2 seconds after submit; proof exported from a validator the gateway does not use first; verified offline with signatures from 3 of 4 validators; the same proof refused against other validators |
| Proof with the chain stopped | The gateway's own stored proof (Postgres) verified with zero validators running |
| Live tests (`make chain-test`) | Anchor, prove, verify, duplicate root refused, unlisted signer refused by the contract, trust commitment, fail-over past a dead endpoint, and a full batch through the gateway's code path |
| Failure tests through the gateway (`deploy/besu/local/failure-tests.sh`) | Pass, on the fourth run: one validator down; the gateway's preferred validator down; below quorum the chain stops without forking and resumes 8 seconds after restart; a validator that lost its disk resyncs in 9 seconds; 36 of 36 anchors accepted while every validator is restarted in turn. The first three runs each lost 1 anchor of 36 — see "Two faults" below |
| Offline verifier tests | A recorded real bundle passes; 12 kinds of tampering and wrong trust anchors are each rejected (`cmd/verifyproof`, `internal/besu`) |

What exists now:

| Piece | Where |
|---|---|
| Contract and pinned compiler; artefacts committed so builds need no Solidity tools | `chain/`, `internal/besu/contract/`, `make contracts-check` |
| Chain client: several endpoints tried in turn, writes return when final | `internal/besu/` |
| Offline proof: block header with validator signatures + state proof + storage entry | `internal/besu/proof.go`, `cmd/verifyproof` |
| Durable proofs: stored with the root when it is anchored, completed later if that step fails | `internal/metagraph/chain_anchor.go`, migration 037 |
| Local chain | `deploy/besu/local/` (`make dev`, `make chain-up`, `make dev-reset`) |
| Failure tests driven through the gateway | `deploy/besu/local/failure-tests.sh` |
| Operator tool: keys, deploy, validators, proof export | `cmd/echochain` |

Two faults the failure tests found in my client, both fixed:

- A validator can accept a transaction and restart before passing it on; the transaction then
  exists nowhere. The client now signs once and resubmits the same transaction every few seconds
  until it is in a block.
- A validator that has just restarted answers before it has caught up, and gave out a stale
  transaction number. The client now uses whichever validator is furthest ahead for every call.

One clean run after the fixes is the evidence so far; the B5 soak repeats these tests for days.

Changed behaviour to know about:

- `ECHO_CHAIN_NETWORK` accepts only `besu`. A value left over from before stops startup.
- If `BESU_RPC_URLS` is set, a bad contract address or a missing signing key stops startup. Unset
  means no chain and no anchoring, for development only. The Kubernetes config carries
  placeholders that stop startup until B4 fills them in.
- Group metadata and governance-vote anchors are refused (`ErrUnsupportedUpdate`). They carried a
  DID in clear text and the previous chain rejected them too; they return as hash commitments in B2.
- Identity, currency and wallet chain calls have nothing to talk to until B2 and B3. The features
  behave as they do when those endpoints are unset.

Not verified: more than minutes of running; more than one machine; node and account
permissioning (the local network is closed only by a static peer list).

## 8. Removal of Tessellation and Constellation (2026-11-24)

Recorded in git first (commit `f06352b`), then removed:

- The Scala metagraph, Euclid and hydra tooling, vendored Tessellation scripts and checksums.
- The stage 2 launcher, rehearsal, Ansible playbooks, key ceremony script and runbook.
- The Constellation deploy path (hydra inventory, DAG staking plan) and the metagraph CI workflow.
- In the gateway: the finality tracker and snapshot evidence, the Tessellation proof verifier,
  the `sovereign` / `constellation` network values, the Constellation explorer link.

Still in the repository, because deleting it today would break the build or a feature. Each goes
with the stage that replaces it:

| What | Goes in |
|---|---|
| Tessellation currency client and transaction types in `internal/metagraph` (rewards, locks, delegation) | B3 |
| Wallet DAG addresses and client-signed Tessellation transactions (`internal/wallet`), iOS `Core/Stargazer/` | B3 |
| Constellation Digital Evidence client (`internal/evidence`, off by default) and its iOS types | Your decision: the Comply service uses it (§9) |
| The package name `internal/metagraph` itself | B5 (rename only) |
| Mentions in older documents, `openapi.yaml`, the local-dev MCP server, `Makefile.did`, the tokenomics compose file, the `echo-metagraph-scala` skill | B5 |

On your machine, not in git, and left for you to delete: the `euclid-development-environment`
clones (in this folder and beside it), build output under `metagraph/`, stray `io/` class files,
`spike/besu` leftovers, and the old Docker images and volumes (about 20 GB).

## 9. Stage B2 results (2026-11-24) — gate met locally

Gate: step 3 of `validate-phase1` and the identity tests pass, with no skipped step.

| Check | Result |
|---|---|
| `make validate-phase1` | 29 passed, 0 failed, 0 skipped |
| Step 3 in detail | A trust-tier commitment anchored through the gateway; the script recomputes `H(tier ‖ nonce)` itself, reads the commitment back from a validator the gateway does not prefer, and searches that block for the DID's bytes: not present |
| Live identity tests (`make chain-test`) | Device key, trust tier, username, credential issuance, revocation list, digest attestation, and the whole gateway path. Refused, as they should be: a duplicate device key or credential, a replayed commitment, a username owned by another DID (in any letter case), an older revocation list replacing a newer one, any write from an unlisted key |
| Several writers sharing the gateway's key | Three clients writing at once: 18 of 18 writes landed, each exactly once |

What changed:

| Piece | Where |
|---|---|
| Identity contract: device keys, trust tiers, usernames, credential issuance, revocation lists, digest attestations. Every identifier is a `bytes32` | `chain/contracts/EchoIdentity.sol` |
| How identifiers are hashed, with domain separation, so a third party can recompute them | `internal/besu/ids.go` |
| Gateway identity path | `internal/metagraph/chain_identity.go`; `BESU_IDENTITY_CONTRACT` |
| Credentials package: issuance records and revocation lists go through an interface to the chain; its direct HTTP calls to the old chain are gone | `pkg/credentials` (`ChainPublisher`) |
| Group metadata and governance votes: anchored as one digest of the record | `BesuAnchor.attest` |
| Policy document | `docs/data-classification.md` |
| Removed: Tessellation signing and data/identity submission code, `cmd/tessellation-id`, the proof-id step of the key script | — |

Differences from the old chain that you should know:

- **Less is on-chain than before.** DIDs, usernames, device keys, credential ids and types were
  written as text; they are hashes now. The device label (free text, e.g. "Alice's iPhone") was
  written; it is not sent at all. The revocation list's full bit vector was written; only its
  hash is, and the gateway serves the list.
- **A hash does not hide a guessable value.** Usernames can be found by hashing candidates. Their
  ownership is meant to be public; do not treat a username as private. The same goes for a vote
  digest when the voter and proposal are known.
- **The timestamps the gateway used to send are gone.** The block's time is the record's time.
- **Organisation role credentials** have no function of their own: the gateway never submitted
  one separately. They are covered as credential issuances with their type hashed.

Not done in B2:

- **Offline proofs for identity records.** A root can be proven offline (B1); an identity record
  can only be read from a validator today. Extending the proof to identity records is small and
  belongs with B4.
- **Constellation Digital Evidence client.** I said it was self-contained. It is not: the Comply
  service's logic uses it. Removing it changes a Comply feature, so it is your decision, not a
  clean-up. It stays off by default.
- **A fault I fixed on the way, for the record:** two test packages signing with the same key at
  once made a write fail with "nonce too low". That was the multi-replica risk from B1 appearing
  for real; the client now re-signs (see §5).
- The in-memory revocation-list publisher (used only without Postgres) restarts its count at zero
  after a restart, and the chain will then refuse its lists as stale. With Postgres the count is
  durable.
