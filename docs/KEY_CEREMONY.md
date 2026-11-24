# Key ceremony — Echo's own chain

Runbook for creating the keys of an Echo network (staging or production). Part of
[`OWN_L0_PLAN.md`](OWN_L0_PLAN.md) stage 2. Run it once per network.

**What is at stake.** Whoever holds the node keys can sign the chain's history. If every copy of
them is lost the network cannot be restarted; if they are stolen, history can be forged. The
allocation keys hold the whole ECHO supply.

## What gets created

| Key | Count | Used by | Where it lives after the ceremony |
|---|---|---|---|
| `gl0-1` … `gl0-3` | 3 | Global L0 nodes | On its host, plus both backups |
| `mg-1` … `mg-3` | 3 | Metagraph hosts (one key signs for all five layers on a host) | On its host, plus both backups |
| `echo-owner`, `identity-owner` | 2 | Sign each metagraph's genesis; needed again only to change a metagraph's owner | Backups only. The deploy copies them to the first metagraph host for the genesis and deletes them afterwards |
| `alloc-community-rewards` | 1 | The only allocation v1 spends from | Secrets store, plus both backups |
| `alloc-treasury`, `alloc-founders`, `alloc-team-advisors`, `alloc-ecosystem` | 4 | Hold undistributed ECHO | Backups only. Never on a server |
| Gateway signing key (`identity-service.pem`) | 1 | The gateway signs Data L1 and Identity L1 updates; the chain rejects any other signer | Kubernetes secret `IDENTITY_SERVICE_KEY_PEM`, plus both backups |

There is no separate fee key: under the environment Echo runs (`dev`, see the plan §10) snapshot
fees are zero. Every node address receives a nominal Global L0 balance at genesis in case that changes.

## Before you start

- A machine that will stay offline for the ceremony, with Java 21, Go, `openssl`, and this
  repository checked out at the commit being deployed.
- While still online, on that machine: `metagraph/vendor/fetch.sh` (downloads the Tessellation tools
  and checks them against the committed checksums) and `go build ./cmd/tessellation-id ./cmd/didkey`
  (so Go needs no network later). Then disconnect it.
- Two new encrypted USB drives, one per backup holder. The holders and locations are recorded in
  the plan §10: the owner and one family member, two separate places.

## Steps

1. Mount the first encrypted drive. Run:

   ```bash
   deploy/sovereign/ceremony/ceremony.sh staging /Volumes/<drive>/echo-staging
   ```

   It prints each key as `created`, writes the private files to the drive, and writes the public
   files to `deploy/sovereign/networks/staging/`. It stops if the ECHO genesis does not total
   exactly 1,000,000,000.
2. Check the output: `manifest.yml` lists six node ids and thirteen addresses; `gl0.priority` has
   3 lines, `mg.seedlist` 3, `gl0.seedlist` 6; `echo-genesis.csv` has five rows matching the
   allocation table in the plan §10.
3. Copy the whole secrets directory to the second encrypted drive. Compare the two copies:
   `diff -r <drive 1>/echo-staging <drive 2>/echo-staging` must print nothing.
4. Prove a backup can be read before trusting it: from the second drive, run the script again with
   the same arguments. Every key must print `kept`, and `git status` must show no change to the
   public files.
5. Hand the second drive to the second holder. Record the date and who holds what in the plan.
6. Commit `deploy/sovereign/networks/staging/`. Nothing in it is secret.
7. Reconnect only after both drives are unmounted. For the deploy, mount one drive on the operator's
   machine and pass its path as `echo_secrets_dir` (see [`deploy/ansible/README.md`](../deploy/ansible/README.md)).

## Rules

- A key file never enters this repository, a chat, an email or a cloud drive. The script refuses a
  secrets directory inside the repository.
- Production and staging have separate ceremonies and separate drives. Staging keys are never
  reused for production.
- The script never regenerates an existing key. Re-running it rebuilds the public files from the
  keys present, which is also how a node is added later (`GL0_NODES=4 ceremony.sh …`).
- The sample validator keys that ship in the Euclid checkout (`source/p12-files/`) are public
  knowledge. They are for the local cluster only and are not used by this deploy.

## If a key is lost or exposed

| Event | Effect | Action |
|---|---|---|
| One node key lost, backups intact | None | Restore the file from a backup |
| One node key exposed | That node can be impersonated; one of three cannot forge history alone | Generate a replacement key, update the seedlists on every node, restart nodes one at a time (`global-l0-join.yml`) |
| Two or more Global L0 keys exposed | History after the last external checkpoint can be rewritten | Stop the network; restart from the last checkpointed snapshot with new keys (`global-l0-rollback.yml`); tell users |
| Gateway signing key exposed | A third party who can also reach the private network could anchor false roots | Generate a new key, update `gateway_signer` in the manifest, redeploy the metagraph layers, update the Kubernetes secret |
| An allocation key exposed | That allocation can be moved | Move the balance to a new address at once. ECHO is non-transferable to users in v1, which limits where it can go |
| Every copy of the node keys lost | The network cannot be restarted | New genesis. Anchored history survives only in the archived snapshots and external checkpoints |
