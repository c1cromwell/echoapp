# Running Echo's chain

**For:** whoever operates the four validators. **Status (2026-11-24):** every procedure here has
been run against the local four-validator network. None has been run on real hosts yet; the
first run on staging is the test of this document (plan stage B4, `docs/BESU_MIGRATION_PLAN.md` §11).

The chain is four Hyperledger Besu validators (QBFT). It keeps running with one down, stops with
two down, and loses nothing either way. Only the gateway writes to it.

## 1. What you hold, and how bad it is to lose each thing

| Thing | Where it lives | If it leaks | If it is lost |
|---|---|---|---|
| A validator's key (`nodeN/key`) | That host, `/etc/echo-besu/key`, plus the ceremony backup | One leaked key cannot forge a block (three of four must sign). Three can rewrite history; checkpoints (§6) are what would expose that | Restore it from the ceremony backup. Without a backup, that validator is gone and must be voted out and replaced |
| Contract owner key (`owner.key`) | Offline | The holder can change which address may write to the contracts | No one can ever change the writer. Existing data is unaffected |
| Reserved key (`reserved.key`) | Offline | Nothing today: the reserve contract has no function that moves anything | Nothing today. Keep it anyway; a future contract may recognise the address |
| Gateway signing key | The gateway's secret store | The holder can anchor roots and write identity records until the owner key names a new writer | Name a new writer with the owner key |
| `genesis.json`, `validators.txt` | Public | Not secret. Publish them | Rebuild from the ceremony backup |
| Chain data | Each host, `/var/lib/echo-besu` | Not secret: hashes only | A validator with no data copies it from the others. All four lost at once: restore from backup (§5) |

## 2. Key ceremony (once per network; you run it)

On a machine you control, with Docker and Go, ideally disconnected from the network once the
Besu image is pulled:

```bash
deploy/besu/ops/ceremony.sh /path/to/ceremony
```

It refuses to write into a directory that is not empty. It prints the validator addresses, the
contract owner's address and the reserved address.

Then, before anything else:

1. Copy the whole directory to two encrypted backups kept in two places.
2. Check a copy: `cd <copy> && shasum -a 256 -c MANIFEST.sha256`.
3. Publish `genesis.json` and `validators.txt`. Anyone verifying a proof needs the validator list
   from somewhere other than Echo's servers.

Staging and production are separate ceremonies with separate keys. Give production its own chain
id: edit `chainId` in `deploy/besu/qbft-config.json` for the production ceremony (local and
staging use 20261).

## 3. First deployment

1. **Hosts.** In `deploy/terraform/hetzner`: set `besu_enabled = true`, set `admin_cidrs` to your
   own address, confirm `besu_server_type` and its price in the Hetzner console, then
   `terraform apply`. Outputs give the public and private addresses.
2. **Validators.** Copy `deploy/besu/ansible/inventory.example.ini` to `inventory.ini`, fill in the
   addresses and `ceremony_dir`, then:
   ```bash
   cd deploy/besu/ansible && ansible-playbook -i inventory.ini site.yml
   ```
   It installs one host at a time and sets the host firewall: the peer port is open to the other
   three validators only, JSON-RPC and metrics to the gateway subnet only.
3. **Contracts.** From a machine on the private network, with the owner key present only for this step:
   ```bash
   go run ./cmd/echochain deploy -rpc http://10.0.2.11:8545 -owner owner.key -signer <gateway address> -name EchoAnchor
   go run ./cmd/echochain deploy -rpc http://10.0.2.11:8545 -owner owner.key -signer <gateway address> -name EchoIdentity
   go run ./cmd/echochain deploy -rpc http://10.0.2.11:8545 -owner owner.key -name EchoReserve -reserved <reserved address>
   ```
   Each prints the contract's address. Take `owner.key` back offline.
4. **Gateway.** Set `ECHO_CHAIN_NETWORK=besu`, `BESU_RPC_URLS` (Terraform output `besu_rpc_urls`),
   `BESU_ANCHOR_CONTRACT`, `BESU_IDENTITY_CONTRACT`, `BESU_VALIDATORS` (from `validators.txt`).
   Apply database migrations (`make migrate`).
5. **Check.** `scripts/validate-phase1.sh` against the deployment, then the failure tests.

## 4. Routine

| When | What |
|---|---|
| Daily | Back up one validator, a different one each day (§5), and publish a checkpoint (§6) |
| On any alert | `deploy/besu/monitoring/alerts.yml` says what each one means |
| Before every Besu upgrade | Back up, then upgrade one validator at a time, waiting for each to rejoin and catch up before the next. Never two at once |
| Quarterly | Run the restore drill on staging (§5) |

## 5. Backup and restore

A backup is one validator's chain data, taken while that validator is stopped for a few seconds.
The other three keep the chain running.

```bash
/opt/echo-besu/backup.sh echo-besu /var/lib/echo-besu /backups/besu-$(date -u +%F).tar.gz
```

Copy the archive and its `.sha256` off the host. One validator at a time, never two.

**One validator lost its data.** Start it empty; it copies the chain from the others. Nothing to restore.

**A validator's host is gone.** New host, same private address, run Ansible (it installs the same
key from the ceremony backup). It copies the chain from the others.

**All four lost.** New hosts, run Ansible, stop the validators, then on each:

```bash
/opt/echo-besu/restore.sh echo-besu /var/lib/echo-besu /backups/<latest for this validator>.tar.gz
docker compose -f /opt/echo-besu/docker-compose.yml up -d
```

`restore.sh` refuses an archive that does not match its checksum. Backups of different validators
need not be from the same moment: in the drill each was taken at a different height, and on start
they agreed on the longest history any of them had. (Backups days apart have not been tried.)
Anything anchored after the newest backup is gone from the chain; the gateway's database still has
those roots and they can be anchored again. This is why backups are daily.

The local rehearsal of exactly this is `make chain-restore-drill`.

## 6. Checkpoints

Echo runs every validator, so the validators could agree to replace history, and a proof checked
only against Echo's chain would not notice. A checkpoint published where Echo cannot edit it would.

```bash
go run ./cmd/echochain checkpoint -rpc http://10.0.2.11:8545 -out checkpoint-$(date -u +%F).json
```

Publish the file somewhere outside Echo's control that keeps history: a public Git repository on
a host you do not run, a timestamping service, a newspaper-of-record style mailing list. More
than one place is better. **Where to publish is a decision still open (§8).**

Anyone, later:

```bash
go run ./cmd/echochain checkpoint-verify -in checkpoint-2026-12-01.json -validators <validators.txt contents>
go run ./cmd/echochain checkpoint-verify -in checkpoint-2026-12-01.json -validators <...> -rpc <a validator>
```

The first checks the validators' signatures offline. The second also asks the chain for that
block and exits 3 if the chain now holds a different one: history was rewritten after the
checkpoint was published.

## 7. Changing the validator set

Validators are changed by vote among the current validators (`qbft_proposeValidatorVote` on more
than half of them), not by editing a file. After any change: update `validators.txt`, publish the
new list with the block number it took effect at, update `BESU_VALIDATORS` on the gateway, and add
or remove the node in the Ansible inventory so the allowlist and firewall follow. Proofs from
before the change verify against the old list; keep every published list.

This has not been rehearsed, locally or anywhere. Rehearse it on staging before it is needed.

## 8. Open

| What | Why it matters |
|---|---|
| Where checkpoints are published | Until they are published outside Echo, they protect nothing |
| Where backups are stored, and who can read them | Off the validators' hosts, in a different account |
| Who is paged | The alert rules exist; nothing routes them yet |
| A fourth site | With three sites, two validators share one; losing that site stops the chain until one returns |
| Validator-set change rehearsal | §7 |
| JSON-RPC has no authentication | It is reachable only from the gateway subnet. Anything on that subnet can read the chain and submit transactions, though only the gateway's key can write to the contracts |
